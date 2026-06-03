#!/usr/bin/env python3
"""Backtest a benchmark-core cross-sectional momentum active sleeve."""

from __future__ import annotations

import argparse
import json
from dataclasses import asdict
from pathlib import Path
from typing import Any

import numpy as np
import pandas as pd

from artifact_manifest import command_line, file_sha256, write_manifest
from backtest_benchmark_rotation import BENCHMARK_PROXY_SYMBOLS, ETF_SYMBOLS, ExecutionConfig, load_bars
from portfolio_sleeve import cap_weight_budget, complete_with_benchmark, simulate_target_weights, write_csv, write_json


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--history-bars-csv", required=True)
    parser.add_argument("--test-bars-csv", required=True)
    parser.add_argument("--benchmark", default="VOO")
    parser.add_argument("--candidate-universe", choices=["all", "stocks", "etfs"], default="all")
    parser.add_argument("--initial-cash", type=float, default=100000.0)
    parser.add_argument("--lookback-bars", type=int, default=21)
    parser.add_argument("--sleeve-fraction", type=float, default=0.30)
    parser.add_argument("--top-k", type=int, default=3)
    parser.add_argument("--max-name-weight", type=float, default=0.10)
    parser.add_argument("--rebalance-every", type=int, default=21)
    parser.add_argument("--rebalance-band", type=float, default=0.025)
    parser.add_argument("--allocation-mode", choices=["equal", "score_over_vol"], default="equal")
    parser.add_argument("--require-positive-momentum", action="store_true", default=True)
    parser.add_argument("--allow-negative-momentum", dest="require_positive_momentum", action="store_false")
    parser.add_argument("--min-relative-momentum", type=float, default=0.0)
    parser.add_argument("--spread-bps", type=float, default=2.0)
    parser.add_argument("--slippage-bps", type=float, default=1.0)
    parser.add_argument("--commission-per-share", type=float, default=0.0)
    parser.add_argument("--min-commission", type=float, default=0.0)
    parser.add_argument("--sec-fees-bps-sell", type=float, default=0.0)
    parser.add_argument("--out-dir", required=True)
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    benchmark = args.benchmark.upper()
    history_bars = load_bars(args.history_bars_csv)
    test_bars = load_bars(args.test_bars_csv)
    execution = ExecutionConfig(
        mode="next_open",
        spread_bps=args.spread_bps,
        slippage_bps=args.slippage_bps,
        commission_per_share=args.commission_per_share,
        min_commission=args.min_commission,
        sec_fees_bps_sell=args.sec_fees_bps_sell,
    )
    result = run_momentum_sleeve(
        history_bars=history_bars,
        test_bars=test_bars,
        benchmark=benchmark,
        candidate_universe=args.candidate_universe,
        initial_cash=args.initial_cash,
        lookback_bars=args.lookback_bars,
        sleeve_fraction=args.sleeve_fraction,
        top_k=args.top_k,
        max_name_weight=args.max_name_weight,
        rebalance_every=args.rebalance_every,
        rebalance_band=args.rebalance_band,
        allocation_mode=args.allocation_mode,
        require_positive_momentum=args.require_positive_momentum,
        min_relative_momentum=args.min_relative_momentum,
        execution=execution,
    )
    manifest = {
        "command": command_line(),
        "history_bars_csv": args.history_bars_csv,
        "test_bars_csv": args.test_bars_csv,
        "history_bars_sha256": file_sha256(args.history_bars_csv),
        "test_bars_sha256": file_sha256(args.test_bars_csv),
        "benchmark": benchmark,
        "candidate_universe": args.candidate_universe,
        "lookback_bars": args.lookback_bars,
        "portfolio_config": {
            "sleeve_fraction": args.sleeve_fraction,
            "top_k": args.top_k,
            "max_name_weight": args.max_name_weight,
            "rebalance_every": args.rebalance_every,
            "rebalance_band": args.rebalance_band,
            "allocation_mode": args.allocation_mode,
            "require_positive_momentum": args.require_positive_momentum,
            "min_relative_momentum": args.min_relative_momentum,
        },
        "cost_model": asdict(execution),
        "status": promotion_status(result["summary"]),
    }
    result["manifest"] = manifest
    out_dir = Path(args.out_dir)
    out_dir.mkdir(parents=True, exist_ok=True)
    write_json(out_dir / "momentum_sleeve.json", result)
    write_csv(out_dir / "momentum_sleeve_equity.csv", result["equity_curve"])
    write_csv(out_dir / "momentum_sleeve_orders.csv", result["orders"])
    write_csv(out_dir / "momentum_sleeve_decisions.csv", result["decisions"])
    write_csv(out_dir / "momentum_sleeve_selections.csv", result["selections"])
    write_markdown(out_dir / "momentum_sleeve.md", result)
    write_manifest(out_dir / "momentum_sleeve_manifest.json", manifest)
    s = result["summary"]
    print(
        f"momentum sleeve lookback={args.lookback_bars} top_k={args.top_k} sleeve={args.sleeve_fraction:.2f} "
        f"return={s['total_return']*100:.2f}% benchmark={s['benchmark_return']*100:.2f}% "
        f"excess={s['excess_return_vs_benchmark']*100:.2f}% sharpe={s['sharpe']:.3f} "
        f"maxDD={s['max_drawdown']*100:.2f}% alpha_symbols={s['num_alpha_symbols_traded']} "
        f"status={manifest['status']}"
    )


def run_momentum_sleeve(
    *,
    history_bars: pd.DataFrame,
    test_bars: pd.DataFrame,
    benchmark: str,
    candidate_universe: str,
    initial_cash: float,
    lookback_bars: int,
    sleeve_fraction: float,
    top_k: int,
    max_name_weight: float,
    rebalance_every: int,
    rebalance_band: float,
    allocation_mode: str,
    require_positive_momentum: bool,
    min_relative_momentum: float,
    execution: ExecutionConfig,
) -> dict[str, Any]:
    all_bars = pd.concat([history_bars, test_bars], ignore_index=True).drop_duplicates(["time", "symbol"], keep="last")
    open_panel = all_bars.pivot(index="time", columns="symbol", values="open").sort_index().dropna(axis=1, how="all").ffill()
    close_panel = all_bars.pivot(index="time", columns="symbol", values="close").sort_index().dropna(axis=1, how="all").ffill()
    test_times = pd.DatetimeIndex(sorted(test_bars["time"].unique()))
    test_open = open_panel.loc[open_panel.index.isin(test_times)]
    test_close = close_panel.loc[close_panel.index.isin(test_times)]
    if benchmark not in close_panel.columns:
        raise ValueError(f"benchmark {benchmark} missing from bars")

    target_by_time: dict[pd.Timestamp, dict[str, float]] = {}
    selections: list[dict[str, Any]] = []
    decision_times = list(test_times[:: max(1, rebalance_every)])
    for time in decision_times:
        if time not in close_panel.index:
            continue
        momentum = np.log(close_panel.loc[time] / close_panel.shift(lookback_bars).loc[time])
        benchmark_momentum = float(momentum.get(benchmark, 0.0))
        relative_momentum = momentum - benchmark_momentum
        vol_20 = np.log(close_panel / close_panel.shift(1)).rolling(20, min_periods=5).std().loc[time] * np.sqrt(252.0)
        candidates = [
            {
                "symbol": symbol,
                "relative_momentum": float(score),
                "absolute_momentum": float(momentum.get(symbol, 0.0)),
                "vol_20": float(vol_20.get(symbol, 0.0)) if symbol in vol_20.index else 0.0,
            }
            for symbol, score in relative_momentum.dropna().items()
            if universe_allows(symbol, candidate_universe, benchmark)
        ]
        if require_positive_momentum:
            candidates = [candidate for candidate in candidates if candidate["relative_momentum"] > 0]
        if min_relative_momentum > 0:
            candidates = [candidate for candidate in candidates if candidate["relative_momentum"] >= min_relative_momentum]
        candidates.sort(key=lambda row: (row["relative_momentum"], row["symbol"]), reverse=True)
        selected = candidates[:top_k]
        if allocation_mode == "score_over_vol":
            raw_weights = {
                row["symbol"]: max(1e-6, row["relative_momentum"]) / max(0.05, row["vol_20"])
                for row in selected
            }
        else:
            raw_weights = {row["symbol"]: 1.0 for row in selected}
        active_weights = cap_weight_budget(raw_weights, sleeve_fraction, max_name_weight)
        target_by_time[time] = complete_with_benchmark(active_weights, benchmark)
        for rank, row in enumerate(selected, start=1):
            selections.append(
                {
                    "decision_time": time.isoformat(),
                    "rank": rank,
                    "symbol": row["symbol"],
                    "relative_momentum": row["relative_momentum"],
                    "absolute_momentum": row["absolute_momentum"],
                    "vol_20": row["vol_20"],
                    "target_weight": float(active_weights.get(row["symbol"], 0.0)),
                }
            )

    result = simulate_target_weights(
        open_=test_open,
        close=test_close,
        benchmark=benchmark,
        target_by_decision_time=target_by_time,
        initial_cash=initial_cash,
        execution=execution,
        rebalance_band=rebalance_band,
    )
    result["selections"] = selections
    result["summary"].update(
        {
            "strategy": "cross_sectional_momentum_active_sleeve",
            "candidate_universe": candidate_universe,
            "lookback_bars": lookback_bars,
            "sleeve_fraction": sleeve_fraction,
            "top_k": top_k,
            "max_name_weight": max_name_weight,
            "rebalance_every": rebalance_every,
            "rebalance_band": rebalance_band,
            "allocation_mode": allocation_mode,
            "require_positive_momentum": require_positive_momentum,
            "min_relative_momentum": min_relative_momentum,
            "selection_dates": len(target_by_time),
            "selection_rows": len(selections),
            "average_selected_relative_momentum": float(np.mean([row["relative_momentum"] for row in selections]))
            if selections
            else 0.0,
        }
    )
    return result


def universe_allows(symbol: str, candidate_universe: str, benchmark: str) -> bool:
    symbol = symbol.upper()
    excluded = set(BENCHMARK_PROXY_SYMBOLS)
    excluded.add(benchmark.upper())
    if symbol in excluded:
        return False
    is_etf = symbol in ETF_SYMBOLS
    if candidate_universe == "stocks":
        return not is_etf
    if candidate_universe == "etfs":
        return is_etf
    return True


def promotion_status(summary: dict[str, Any]) -> str:
    if summary["excess_return_vs_benchmark"] <= 0:
        return "rejected_under_benchmark"
    if summary["num_alpha_symbols_traded"] < 5:
        return "research_only_insufficient_breadth"
    if summary["max_drawdown"] > summary["benchmark_max_drawdown"] + 0.03:
        return "research_only_drawdown_regression"
    if summary["turnover"] > 8.0:
        return "research_only_high_turnover"
    return "candidate"


def write_markdown(path: Path, result: dict[str, Any]) -> None:
    s = result["summary"]
    lines = [
        "# Cross-Sectional Momentum Active Sleeve\n\n",
        f"- Benchmark core: `{s['benchmark']}`\n",
        f"- Candidate universe: `{s['candidate_universe']}`\n",
        f"- Lookback bars: `{s['lookback_bars']}`\n",
        f"- Active sleeve: `{s['sleeve_fraction']:.2f}`\n",
        f"- Top-k: `{s['top_k']}`\n",
        f"- Max name weight: `{s['max_name_weight']:.2f}`\n",
        f"- Rebalance every: `{s['rebalance_every']}` bars\n",
        f"- Allocation mode: `{s['allocation_mode']}`\n",
        f"- Minimum relative momentum: `{s['min_relative_momentum']:.4f}`\n",
        f"- Explicit costs: `${s['total_cost']:.2f}`\n\n",
        "| Metric | Strategy | Benchmark |\n",
        "|---|---:|---:|\n",
        f"| Total return | {s['total_return']*100:.2f}% | {s['benchmark_return']*100:.2f}% |\n",
        f"| Excess return | {s['excess_return_vs_benchmark']*100:.2f}% | 0.00% |\n",
        f"| Annualized return | {s['annualized_return']*100:.2f}% | {s['benchmark_annualized_return']*100:.2f}% |\n",
        f"| Sharpe | {s['sharpe']:.3f} | {s['benchmark_sharpe']:.3f} |\n",
        f"| Sortino | {s['sortino']:.3f} | {s['benchmark_sortino']:.3f} |\n",
        f"| Max drawdown | {s['max_drawdown']*100:.2f}% | {s['benchmark_max_drawdown']*100:.2f}% |\n",
        f"| Turnover | {s['turnover']:.3f} |  |\n",
        f"| Rebalances | {s['num_rebalances']} |  |\n",
        f"| Alpha symbols traded | {s['num_alpha_symbols_traded']} |  |\n",
    ]
    path.write_text("".join(lines), encoding="utf-8")


if __name__ == "__main__":
    main()
