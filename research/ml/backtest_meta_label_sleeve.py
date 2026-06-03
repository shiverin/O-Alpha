#!/usr/bin/env python3
"""Backtest the existing meta-label model as a benchmark-core active sleeve."""

from __future__ import annotations

import argparse
import json
from dataclasses import asdict
from pathlib import Path
from typing import Any

import lightgbm as lgb

from artifact_manifest import command_line, research_status_accepted, write_manifest
from backtest_benchmark_rotation import (
    ExecutionConfig,
    RotationPolicy,
    load_bars,
    load_features,
    load_signals,
    load_yaml,
    model_path,
    safe_price,
    score_candidates,
)
from portfolio_sleeve import cap_weight_budget, complete_with_benchmark, simulate_target_weights, write_csv, write_json


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--bars-csv", required=True)
    parser.add_argument("--signals-csv", required=True)
    parser.add_argument("--features-csv", required=True)
    parser.add_argument("--metadata", required=True)
    parser.add_argument("--feature-spec", default="research/ml/feature_spec_live_core.yaml")
    parser.add_argument("--benchmark", default="VOO")
    parser.add_argument("--initial-cash", type=float, default=100000.0)
    parser.add_argument("--require-candidate-model", action="store_true")
    parser.add_argument("--threshold-floor", type=float)
    parser.add_argument("--candidate-universe", choices=["all", "stocks", "etfs"], default="all")
    parser.add_argument("--min-relative-strength-21", type=float)
    parser.add_argument("--min-log-ret-21", type=float)
    parser.add_argument("--max-close-to-close-vol-20", type=float)
    parser.add_argument("--max-hold-bars", type=int, default=126)
    parser.add_argument("--stop-loss-pct", type=float, default=0.0)
    parser.add_argument("--take-profit-pct", type=float, default=0.0)
    parser.add_argument("--selection-mode", choices=["probability", "probability_plus_momentum"], default="probability")
    parser.add_argument("--selection-momentum-weight", type=float, default=0.0)
    parser.add_argument("--sleeve-fraction", type=float, default=0.30)
    parser.add_argument("--top-k", type=int, default=3)
    parser.add_argument("--max-name-weight", type=float, default=0.10)
    parser.add_argument("--rebalance-every", type=int, default=21)
    parser.add_argument("--rebalance-band", type=float, default=0.005)
    parser.add_argument("--spread-bps", type=float, default=2.0)
    parser.add_argument("--slippage-bps", type=float, default=1.0)
    parser.add_argument("--commission-per-share", type=float, default=0.0)
    parser.add_argument("--min-commission", type=float, default=0.0)
    parser.add_argument("--sec-fees-bps-sell", type=float, default=0.0)
    parser.add_argument("--out-dir", required=True)
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    metadata_path = Path(args.metadata)
    metadata = json.loads(metadata_path.read_text(encoding="utf-8"))
    feature_names = load_yaml(args.feature_spec)["features"]
    bars = load_bars(args.bars_csv)
    signals = load_signals(args.signals_csv)
    features = load_features(args.features_csv, feature_names)
    model = lgb.Booster(model_file=str(model_path(metadata_path, metadata)))
    threshold = float(metadata.get("thresholds", {}).get("enter_long", 0.5))
    policy = RotationPolicy(
        require_candidate_model=args.require_candidate_model,
        threshold_floor=args.threshold_floor,
        candidate_universe=args.candidate_universe,
        min_relative_strength_21=args.min_relative_strength_21,
        min_log_ret_21=args.min_log_ret_21,
        max_close_to_close_vol_20=args.max_close_to_close_vol_20,
        max_hold_bars=args.max_hold_bars,
        stop_loss_pct=args.stop_loss_pct,
        take_profit_pct=args.take_profit_pct,
        selection_mode=args.selection_mode,
        selection_momentum_weight=args.selection_momentum_weight,
    )
    execution = ExecutionConfig(
        mode="next_open",
        spread_bps=args.spread_bps,
        slippage_bps=args.slippage_bps,
        commission_per_share=args.commission_per_share,
        min_commission=args.min_commission,
        sec_fees_bps_sell=args.sec_fees_bps_sell,
    )
    allow_alpha = not policy.require_candidate_model or research_status_accepted(metadata.get("status"))
    result = run_meta_label_sleeve(
        bars=bars,
        signals=signals,
        features=features,
        feature_names=feature_names,
        model=model,
        threshold=threshold,
        benchmark=args.benchmark.upper(),
        initial_cash=args.initial_cash,
        policy=policy,
        execution=execution,
        calibration=metadata.get("calibration", {"method": "none"}),
        allow_alpha=allow_alpha,
        sleeve_fraction=args.sleeve_fraction,
        top_k=args.top_k,
        max_name_weight=args.max_name_weight,
        rebalance_every=args.rebalance_every,
        rebalance_band=args.rebalance_band,
    )
    result["manifest"] = {
        "backtest_command": command_line(),
        "metadata": str(metadata_path),
        "benchmark": args.benchmark.upper(),
        "artifact_status": metadata.get("status"),
        "cost_model": asdict(execution),
        "policy": asdict(policy),
        "sleeve_fraction": args.sleeve_fraction,
        "top_k": args.top_k,
        "max_name_weight": args.max_name_weight,
        "rebalance_every": args.rebalance_every,
        "rebalance_band": args.rebalance_band,
        "symbols": sorted(bars["symbol"].unique().tolist()),
    }

    out_dir = Path(args.out_dir)
    out_dir.mkdir(parents=True, exist_ok=True)
    write_json(out_dir / "meta_label_sleeve.json", result)
    write_csv(out_dir / "meta_label_sleeve_equity.csv", result["equity_curve"])
    write_csv(out_dir / "meta_label_sleeve_orders.csv", result["orders"])
    write_csv(out_dir / "meta_label_sleeve_decisions.csv", result["decisions"])
    write_markdown(out_dir / "meta_label_sleeve.md", result)
    write_manifest(out_dir / "meta_label_sleeve_manifest.json", result["manifest"])

    summary = result["summary"]
    print(
        f"meta-label sleeve return={summary['total_return']*100:.2f}% "
        f"benchmark={summary['benchmark_return']*100:.2f}% "
        f"excess={summary['excess_return_vs_benchmark']*100:.2f}% "
        f"sharpe={summary['sharpe']:.3f} maxDD={summary['max_drawdown']*100:.2f}% "
        f"orders={summary['num_orders']} alpha_symbols={summary['num_alpha_symbols_traded']} "
        f"costs=${summary['total_cost']:.2f}"
    )


def run_meta_label_sleeve(
    *,
    bars,
    signals,
    features,
    feature_names: list[str],
    model: lgb.Booster,
    threshold: float,
    benchmark: str,
    initial_cash: float,
    policy: RotationPolicy,
    execution: ExecutionConfig,
    calibration: dict[str, Any] | None,
    allow_alpha: bool,
    sleeve_fraction: float,
    top_k: int,
    max_name_weight: float,
    rebalance_every: int,
    rebalance_band: float,
) -> dict[str, Any]:
    effective_threshold = max(threshold, policy.threshold_floor or 0.0)
    open_ = bars.pivot(index="time", columns="symbol", values="open").sort_index().dropna(axis=1, how="all").ffill()
    close = bars.pivot(index="time", columns="symbol", values="close").sort_index().dropna(axis=1, how="all").ffill()
    candidate_probs = (
        score_candidates(signals, features, feature_names, model, benchmark, policy, calibration) if allow_alpha else {}
    )
    signal_map = {
        (row.symbol, row.time): row.signal
        for row in signals[["symbol", "time", "signal"]].itertuples(index=False)
    }
    times = list(close.index)
    active: dict[str, dict[str, Any]] = {}
    target_by_time: dict[Any, dict[str, float]] = {}
    position_events: list[dict[str, Any]] = []
    last_target: dict[str, float] = {benchmark: 1.0}

    for i, current_time in enumerate(times[:-1]):
        changed = False
        for symbol in list(active):
            entry = active[symbol]
            entry_price = float(entry.get("entry_price", 0.0))
            current_price = safe_price(close, current_time, symbol)
            trade_return = current_price / entry_price - 1.0 if entry_price > 0 and current_price > 0 else 0.0
            hold_bars = i - int(entry.get("entry_index", i))
            exit_reason = ""
            if signal_map.get((symbol, current_time)) == "SELL":
                exit_reason = "base_sell"
            elif policy.stop_loss_pct > 0 and trade_return <= -policy.stop_loss_pct:
                exit_reason = "stop_loss"
            elif policy.take_profit_pct > 0 and trade_return >= policy.take_profit_pct:
                exit_reason = "take_profit"
            elif policy.max_hold_bars > 0 and hold_bars >= policy.max_hold_bars:
                exit_reason = "max_hold"
            if exit_reason:
                position_events.append(close_position_event(symbol, entry, current_time, current_price, exit_reason))
                del active[symbol]
                changed = True

        accepted = [
            candidate
            for candidate in candidate_probs.get(current_time, [])
            if candidate["probability"] >= effective_threshold
        ]
        accepted.sort(key=lambda item: (item["score"], item["probability"], item["symbol"]), reverse=True)
        for candidate in accepted:
            symbol = candidate["symbol"]
            if symbol == benchmark:
                continue
            if symbol not in active:
                active[symbol] = {
                    "entry_time": current_time,
                    "entry_index": i,
                    "entry_price": safe_price(close, current_time, symbol),
                    "probability": candidate["probability"],
                    "score": candidate["score"],
                    "close_to_close_vol_20": candidate.get("close_to_close_vol_20", 0.0),
                }
                changed = True
            else:
                active[symbol].update(
                    {
                        "probability": candidate["probability"],
                        "score": candidate["score"],
                        "close_to_close_vol_20": candidate.get("close_to_close_vol_20", 0.0),
                    }
                )

        ranked = sorted(active.items(), key=lambda item: (item[1].get("score", 0.0), item[0]), reverse=True)
        for symbol, entry in ranked[top_k:]:
            current_price = safe_price(close, current_time, symbol)
            position_events.append(close_position_event(symbol, entry, current_time, current_price, "ranked_out"))
            del active[symbol]
            changed = True
        selected = dict(ranked[:top_k])
        raw_weights = {
            symbol: max(0.01, float(entry.get("score", 0.0))) / max(0.05, float(entry.get("close_to_close_vol_20", 0.0)))
            for symbol, entry in selected.items()
        }
        active_weights = cap_weight_budget(raw_weights, sleeve_fraction, max_name_weight)
        target = complete_with_benchmark(active_weights, benchmark)
        scheduled_rebalance = rebalance_every > 0 and i % rebalance_every == 0 and bool(active)
        if changed or scheduled_rebalance or material_target_change(target, last_target):
            target_by_time[current_time] = target
            last_target = target

    for symbol, entry in active.items():
        current_time = times[-1]
        current_price = safe_price(close, current_time, symbol)
        position_events.append(close_position_event(symbol, entry, current_time, current_price, "end_of_test"))

    result = simulate_target_weights(
        open_=open_,
        close=close,
        benchmark=benchmark,
        target_by_decision_time=target_by_time,
        initial_cash=initial_cash,
        execution=execution,
        rebalance_band=rebalance_band,
    )
    result["position_events"] = position_events
    result["summary"].update(
        {
            "strategy": "meta_label_active_sleeve",
            "threshold": threshold,
            "effective_threshold": effective_threshold,
            "alpha_enabled": allow_alpha,
            "sleeve_fraction": sleeve_fraction,
            "top_k": top_k,
            "max_name_weight": max_name_weight,
            "rebalance_every": rebalance_every,
            "rebalance_band": rebalance_band,
            "policy": asdict(policy),
            "candidate_decision_dates": len(candidate_probs),
            "position_events": len(position_events),
        }
    )
    return result


def material_target_change(left: dict[str, float], right: dict[str, float]) -> bool:
    symbols = set(left) | set(right)
    return any(abs(float(left.get(symbol, 0.0)) - float(right.get(symbol, 0.0))) > 1e-6 for symbol in symbols)


def close_position_event(symbol: str, entry: dict[str, Any], exit_time, exit_price: float, reason: str) -> dict[str, Any]:
    entry_price = float(entry.get("entry_price", 0.0))
    return {
        "symbol": symbol,
        "entry_time": entry.get("entry_time").isoformat() if entry.get("entry_time") is not None else "",
        "exit_time": exit_time.isoformat(),
        "entry_price": entry_price,
        "exit_price": exit_price,
        "return": exit_price / entry_price - 1.0 if entry_price > 0 and exit_price > 0 else 0.0,
        "entry_probability": float(entry.get("probability", 0.0)),
        "entry_score": float(entry.get("score", 0.0)),
        "reason": reason,
    }


def write_markdown(path: Path, result: dict[str, Any]) -> None:
    s = result["summary"]
    lines = [
        "# Meta-Label Active Sleeve Backtest\n\n",
        f"- Benchmark core: `{s['benchmark']}`\n",
        f"- Active sleeve: `{s['sleeve_fraction']:.2f}`\n",
        f"- Top-k: `{s['top_k']}`\n",
        f"- Max name weight: `{s['max_name_weight']:.2f}`\n",
        f"- Effective threshold: `{s['effective_threshold']:.12f}`\n",
        f"- Candidate decision dates: `{s['candidate_decision_dates']}`\n",
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
        f"| Orders | {s['num_orders']} |  |\n",
        f"| Alpha symbols traded | {s['num_alpha_symbols_traded']} |  |\n",
    ]
    path.write_text("".join(lines), encoding="utf-8")


if __name__ == "__main__":
    main()
