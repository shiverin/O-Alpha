#!/usr/bin/env python3
"""Backtest a composite benchmark-core momentum sleeve."""

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
    parser.add_argument("--legs-json", required=True)
    parser.add_argument("--initial-cash", type=float, default=100000.0)
    parser.add_argument("--rebalance-every", type=int, default=21)
    parser.add_argument("--rebalance-band", type=float, default=0.025)
    parser.add_argument("--global-max-name-weight", type=float, default=0.30)
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
    legs = parse_legs(args.legs_json)
    execution = ExecutionConfig(
        mode="next_open",
        spread_bps=args.spread_bps,
        slippage_bps=args.slippage_bps,
        commission_per_share=args.commission_per_share,
        min_commission=args.min_commission,
        sec_fees_bps_sell=args.sec_fees_bps_sell,
    )
    result = run_composite_momentum_sleeve(
        history_bars=history_bars,
        test_bars=test_bars,
        benchmark=benchmark,
        legs=legs,
        initial_cash=args.initial_cash,
        rebalance_every=args.rebalance_every,
        rebalance_band=args.rebalance_band,
        global_max_name_weight=args.global_max_name_weight,
        execution=execution,
    )
    manifest = {
        "command": command_line(),
        "history_bars_csv": args.history_bars_csv,
        "test_bars_csv": args.test_bars_csv,
        "history_bars_sha256": file_sha256(args.history_bars_csv),
        "test_bars_sha256": file_sha256(args.test_bars_csv),
        "benchmark": benchmark,
        "legs": legs,
        "portfolio_config": {
            "rebalance_every": args.rebalance_every,
            "rebalance_band": args.rebalance_band,
            "global_max_name_weight": args.global_max_name_weight,
        },
        "cost_model": asdict(execution),
        "status": promotion_status(result["summary"]),
    }
    result["manifest"] = manifest
    out_dir = Path(args.out_dir)
    out_dir.mkdir(parents=True, exist_ok=True)
    write_json(out_dir / "composite_momentum_sleeve.json", result)
    write_csv(out_dir / "composite_momentum_sleeve_equity.csv", result["equity_curve"])
    write_csv(out_dir / "composite_momentum_sleeve_orders.csv", result["orders"])
    write_csv(out_dir / "composite_momentum_sleeve_decisions.csv", result["decisions"])
    write_csv(out_dir / "composite_momentum_sleeve_selections.csv", result["selections"])
    write_markdown(out_dir / "composite_momentum_sleeve.md", result)
    write_manifest(out_dir / "composite_momentum_sleeve_manifest.json", manifest)
    s = result["summary"]
    print(
        f"composite momentum return={s['total_return']*100:.2f}% "
        f"benchmark={s['benchmark_return']*100:.2f}% excess={s['excess_return_vs_benchmark']*100:.2f}% "
        f"sharpe={s['sharpe']:.3f} maxDD={s['max_drawdown']*100:.2f}% "
        f"alpha_symbols={s['num_alpha_symbols_traded']} status={manifest['status']}"
    )


def parse_legs(raw: str) -> list[dict[str, Any]]:
    value = json.loads(raw)
    if not isinstance(value, list) or not value:
        raise ValueError("--legs-json must be a non-empty JSON array")
    legs = []
    for index, item in enumerate(value):
        if not isinstance(item, dict):
            raise ValueError("each leg must be a JSON object")
        sleeve = float(item.get("sleeve_fraction", 0.0))
        if sleeve < 0:
            raise ValueError("leg sleeve_fraction must be non-negative")
        legs.append(
            {
                "name": str(item.get("name", f"leg_{index+1}")),
                "candidate_universe": str(item.get("candidate_universe", "all")),
                "candidate_symbols": parse_symbols(item.get("candidate_symbols", [])),
                "lookback_bars": int(item.get("lookback_bars", 21)),
                "sleeve_fraction": sleeve,
                "top_k": int(item.get("top_k", 1)),
                "max_name_weight": float(item.get("max_name_weight", sleeve)),
                "rank_mode": str(item.get("rank_mode", "relative_momentum")),
                "weight_mode": str(item.get("weight_mode", item.get("allocation_mode", "equal"))),
                "allocation_mode": str(item.get("allocation_mode", "equal")),
                "min_relative_momentum": float(item.get("min_relative_momentum", 0.0)),
                "max_vol_20": float(item.get("max_vol_20", 0.0)),
                "edge_exponent": float(item.get("edge_exponent", 1.0)),
                "vol_floor": float(item.get("vol_floor", 1e-6)),
            }
        )
    total_sleeve = sum(leg["sleeve_fraction"] for leg in legs)
    if total_sleeve > 1.0:
        raise ValueError("combined sleeve_fraction cannot exceed 1.0")
    return legs


def run_composite_momentum_sleeve(
    *,
    history_bars: pd.DataFrame,
    test_bars: pd.DataFrame,
    benchmark: str,
    legs: list[dict[str, Any]],
    initial_cash: float,
    rebalance_every: int,
    rebalance_band: float,
    global_max_name_weight: float,
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
    log_returns_1 = np.log(close_panel / close_panel.shift(1))
    vol_20_panel = log_returns_1.rolling(20, min_periods=5).std() * np.sqrt(252.0)

    for time in decision_times:
        combined_weights: dict[str, float] = {}
        for leg in legs:
            momentum = np.log(close_panel.loc[time] / close_panel.shift(int(leg["lookback_bars"])).loc[time])
            benchmark_momentum = float(momentum.get(benchmark, 0.0))
            relative_momentum = momentum - benchmark_momentum
            vol_20 = vol_20_panel.loc[time]
            candidate_symbols = set(leg.get("candidate_symbols") or [])
            candidates = [
                {
                    "symbol": symbol,
                    "score": leg_score(
                        str(leg.get("rank_mode", "relative_momentum")),
                        float(score),
                        float(vol_20.get(symbol, 0.0)) if symbol in vol_20.index else 0.0,
                    ),
                    "relative_momentum": float(score),
                    "absolute_momentum": float(momentum.get(symbol, 0.0)),
                    "vol_20": float(vol_20.get(symbol, 0.0)) if symbol in vol_20.index else 0.0,
                }
                for symbol, score in relative_momentum.dropna().items()
                if (not candidate_symbols or symbol in candidate_symbols)
                if universe_allows(symbol, leg["candidate_universe"], benchmark)
                and float(score) >= float(leg["min_relative_momentum"])
                and (
                    float(leg.get("max_vol_20", 0.0)) <= 0
                    or float(vol_20.get(symbol, 0.0)) <= float(leg.get("max_vol_20", 0.0))
                )
            ]
            candidates.sort(key=lambda row: (row["score"], row["symbol"]), reverse=True)
            selected = candidates[: int(leg["top_k"])]
            raw_weights = leg_raw_weights(selected, leg)
            leg_weights = cap_weight_budget(raw_weights, leg["sleeve_fraction"], leg["max_name_weight"])
            for symbol, weight in leg_weights.items():
                combined_weights[symbol] = combined_weights.get(symbol, 0.0) + weight
            for rank, row in enumerate(selected, start=1):
                selections.append(
                    {
                        "decision_time": time.isoformat(),
                        "leg": leg["name"],
                        "rank": rank,
                        "symbol": row["symbol"],
                        "score": row["score"],
                        "relative_momentum": row["relative_momentum"],
                        "absolute_momentum": row["absolute_momentum"],
                        "vol_20": row["vol_20"],
                        "target_weight": float(leg_weights.get(row["symbol"], 0.0)),
                    }
                )
        combined_weights = {
            symbol: min(global_max_name_weight, weight)
            for symbol, weight in combined_weights.items()
            if weight > 1e-9
        }
        target_by_time[time] = complete_with_benchmark(combined_weights, benchmark)

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
            "strategy": "composite_momentum_active_sleeve",
            "legs": legs,
            "total_sleeve_fraction": sum(leg["sleeve_fraction"] for leg in legs),
            "rebalance_every": rebalance_every,
            "rebalance_band": rebalance_band,
            "global_max_name_weight": global_max_name_weight,
            "selection_dates": len(target_by_time),
            "selection_rows": len(selections),
        }
    )
    return result


def parse_symbols(raw: Any) -> list[str]:
    if raw is None:
        return []
    if isinstance(raw, str):
        return [part.strip().upper() for part in raw.split(",") if part.strip()]
    if isinstance(raw, list):
        return [str(part).strip().upper() for part in raw if str(part).strip()]
    raise ValueError("candidate_symbols must be a list or comma-separated string")


def leg_score(rank_mode: str, relative_momentum: float, vol_20: float) -> float:
    mode = rank_mode.strip().lower()
    if mode in {"low_vol", "low_volatility"}:
        return -vol_20
    if mode in {"mean_reversion", "relative_reversal"}:
        return -relative_momentum
    if mode == "vol_adjusted_momentum":
        return relative_momentum / vol_20 if vol_20 > 1e-9 else relative_momentum
    return relative_momentum


def leg_raw_weights(selected: list[dict[str, Any]], leg: dict[str, Any]) -> dict[str, float]:
    mode = str(leg.get("weight_mode") or leg.get("allocation_mode") or "equal").strip().lower()
    if mode in {"risk_adjusted_edge", "edge_over_vol"}:
        edge_exponent = max(1e-9, float(leg.get("edge_exponent", 1.0)))
        vol_floor = max(1e-9, float(leg.get("vol_floor", 1e-6)))
        threshold = float(leg.get("min_relative_momentum", 0.0))
        return {
            row["symbol"]: (max(0.0, row["relative_momentum"] - threshold) ** edge_exponent)
            / max(vol_floor, row["vol_20"])
            for row in selected
            if row["relative_momentum"] > threshold
        }
    if mode in {"score", "score_weighted"}:
        edge_exponent = max(1e-9, float(leg.get("edge_exponent", 1.0)))
        vol_floor = max(1e-9, float(leg.get("vol_floor", 1e-6)))
        return {
            row["symbol"]: (max(0.0, row["score"]) ** edge_exponent) / max(vol_floor, row["vol_20"])
            for row in selected
            if row["score"] > 0
        }
    if mode == "score_over_vol":
        return {
            row["symbol"]: max(1e-6, row["relative_momentum"]) / max(0.05, row["vol_20"])
            for row in selected
        }
    return {row["symbol"]: 1.0 for row in selected}


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
        "# Composite Momentum Active Sleeve\n\n",
        f"- Benchmark core: `{s['benchmark']}`\n",
        f"- Total active sleeve: `{s['total_sleeve_fraction']:.2f}`\n",
        f"- Rebalance every: `{s['rebalance_every']}` bars\n",
        f"- Global max name weight: `{s['global_max_name_weight']:.2f}`\n",
        f"- Explicit costs: `${s['total_cost']:.2f}`\n\n",
        "## Legs\n\n",
    ]
    for leg in s["legs"]:
        lines.append(
            f"- `{leg['name']}`: universe `{leg['candidate_universe']}`, lookback `{leg['lookback_bars']}`, "
            f"sleeve `{leg['sleeve_fraction']:.2f}`, top-k `{leg['top_k']}`, "
            f"min rel momentum `{leg['min_relative_momentum']:.4f}`, max vol `{leg.get('max_vol_20', 0.0):.2f}`\n"
        )
    lines.extend(
        [
            "\n| Metric | Strategy | Benchmark |\n",
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
    )
    path.write_text("".join(lines), encoding="utf-8")


if __name__ == "__main__":
    main()
