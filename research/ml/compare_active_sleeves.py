#!/usr/bin/env python3
"""Costed next-open comparison for benchmark-funded active sleeves.

This is a Python pre-screen for allocation ideas. It mirrors the current Go
ranked-sleeve family configs closely enough to catch weak candidates before
spending official `cmd/alpha-research` runs. Promotion still belongs to the Go
harness.
"""

from __future__ import annotations

import argparse
import json
from dataclasses import asdict
from pathlib import Path
from typing import Any

import pandas as pd

from artifact_manifest import command_line, file_sha256, write_manifest
from backtest_benchmark_rotation import ExecutionConfig, load_bars
from backtest_composite_momentum_sleeve import run_composite_momentum_sleeve
from portfolio_sleeve import write_csv, write_json


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--bars-csv", required=True)
    parser.add_argument("--benchmark", default="VOO")
    parser.add_argument("--families", default="benchmark_ranked_sleeve,sector_ranked_sleeve")
    parser.add_argument("--start-year", type=int, default=2017)
    parser.add_argument("--end-year", type=int, default=0, help="inclusive; default uses max year in bars")
    parser.add_argument("--initial-cash", type=float, default=100000.0)
    parser.add_argument("--spread-bps", type=float, default=2.0)
    parser.add_argument("--slippage-bps", type=float, default=1.0)
    parser.add_argument("--commission-per-share", type=float, default=0.0)
    parser.add_argument("--min-commission", type=float, default=0.0)
    parser.add_argument("--sec-fees-bps-sell", type=float, default=0.0)
    parser.add_argument("--rebalance-band", type=float, default=0.025)
    parser.add_argument("--out-dir", required=True)
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    benchmark = args.benchmark.upper().strip()
    bars = load_bars(args.bars_csv)
    execution = ExecutionConfig(
        mode="next_open",
        spread_bps=args.spread_bps,
        slippage_bps=args.slippage_bps,
        commission_per_share=args.commission_per_share,
        min_commission=args.min_commission,
        sec_fees_bps_sell=args.sec_fees_bps_sell,
    )
    variants = selected_variants(parse_families(args.families), benchmark)
    folds = annual_folds(bars, args.start_year, args.end_year)
    if not folds:
        raise ValueError("no annual folds available from requested date range")

    rows: list[dict[str, Any]] = []
    fold_rows: list[dict[str, Any]] = []
    for variant in variants:
        fold_results = []
        for fold in folds:
            result = run_composite_momentum_sleeve(
                history_bars=fold["history_bars"],
                test_bars=fold["test_bars"],
                benchmark=benchmark,
                legs=variant["legs"],
                initial_cash=args.initial_cash,
                rebalance_every=variant["rebalance_every"],
                rebalance_band=args.rebalance_band,
                global_max_name_weight=variant["global_max_name_weight"],
                execution=execution,
            )
            summary = result["summary"]
            fold_result = {
                "variant": variant["name"],
                "family": variant["family"],
                "fold": fold["name"],
                "test_start": fold["test_start"].isoformat(),
                "test_end": fold["test_end"].isoformat(),
                "strategy_return": summary["total_return"],
                "benchmark_return": summary["benchmark_return"],
                "excess_return": summary["excess_return_vs_benchmark"],
                "sharpe": summary["sharpe"],
                "benchmark_sharpe": summary["benchmark_sharpe"],
                "max_drawdown": summary["max_drawdown"],
                "benchmark_max_drawdown": summary["benchmark_max_drawdown"],
                "turnover": summary["turnover"],
                "total_cost": summary["total_cost"],
                "num_rebalances": summary["num_rebalances"],
                "num_orders": summary["num_orders"],
                "num_alpha_symbols_traded": summary["num_alpha_symbols_traded"],
            }
            fold_results.append(fold_result)
            fold_rows.append(fold_result)
        rows.append(aggregate_variant(variant, fold_results))

    rows.sort(
        key=lambda row: (
            row["decision_rank"],
            row["compounded_excess_return"],
            row["compounded_strategy_return"],
        ),
        reverse=True,
    )
    report = {
        "summary": {
            "benchmark": benchmark,
            "families": sorted(parse_families(args.families)),
            "folds": [{"name": fold["name"], "test_start": fold["test_start"].isoformat(), "test_end": fold["test_end"].isoformat()} for fold in folds],
            "variant_count": len(rows),
            "champion": rows[0]["variant"] if rows else "",
            "champion_decision": rows[0]["decision"] if rows else "",
            "status": "python_prescreen_only",
            "promotion_note": "official alpha promotion still requires cmd/alpha-research DSR/PBO gate",
        },
        "variants": rows,
        "fold_results": fold_rows,
        "manifest": {
            "command": command_line(),
            "bars_csv": args.bars_csv,
            "bars_csv_sha256": file_sha256(args.bars_csv),
            "benchmark": benchmark,
            "cost_model": asdict(execution),
            "rebalance_band_half_l1": args.rebalance_band,
            "start_year": args.start_year,
            "end_year": args.end_year or int(bars["time"].dt.year.max()),
        },
    }
    out_dir = Path(args.out_dir)
    out_dir.mkdir(parents=True, exist_ok=True)
    write_json(out_dir / "active_sleeve_comparison.json", report)
    write_csv(out_dir / "active_sleeve_comparison_variants.csv", rows)
    write_csv(out_dir / "active_sleeve_comparison_folds.csv", fold_rows)
    write_markdown(out_dir / "active_sleeve_comparison.md", report)
    write_manifest(out_dir / "active_sleeve_comparison_manifest.json", report["manifest"])

    champion = rows[0]
    print(
        f"champion={champion['variant']} decision={champion['decision']} "
        f"return={champion['compounded_strategy_return']*100:.2f}% "
        f"benchmark={champion['compounded_benchmark_return']*100:.2f}% "
        f"excess={champion['compounded_excess_return']*100:.2f}% "
        f"folds_beating={champion['folds_beating_benchmark']}/{champion['fold_count']}"
    )


def parse_families(raw: str) -> set[str]:
    values = {part.strip().lower() for part in raw.split(",") if part.strip()}
    if not values or "all" in values:
        return {"benchmark_ranked_sleeve", "sector_ranked_sleeve"}
    return values


def selected_variants(families: set[str], benchmark: str) -> list[dict[str, Any]]:
    variants: list[dict[str, Any]] = []
    if "benchmark_ranked_sleeve" in families:
        variants.extend(
            [
                benchmark_ranked_variant(benchmark, "benchmark_ranked_sleeve_checkpoint", 21, 189, 0.30, 5, 0.08, 0.03),
                benchmark_ranked_variant(benchmark, "benchmark_ranked_sleeve_conservative", 42, 126, 0.20, 3, 0.08, 0.02),
                benchmark_ranked_variant(benchmark, "benchmark_ranked_sleeve_medium", 21, 189, 0.30, 5, 0.08, 0.04),
                benchmark_ranked_variant(benchmark, "benchmark_ranked_sleeve_slow", 63, 252, 0.25, 5, 0.10, 0.02),
            ]
        )
    if "sector_ranked_sleeve" in families:
        variants.extend(
            [
                sector_ranked_variant(benchmark, "sector_ranked_sleeve_checkpoint", 21, 189, 0.30, 3, 0.10, 0.01),
                sector_ranked_variant(benchmark, "sector_ranked_sleeve_conservative", 42, 126, 0.20, 3, 0.08, 0.00),
                sector_ranked_variant(benchmark, "sector_ranked_sleeve_medium", 21, 189, 0.30, 3, 0.10, 0.02),
                sector_ranked_variant(benchmark, "sector_ranked_sleeve_slow", 63, 252, 0.25, 4, 0.08, -0.01),
            ]
        )
    if not variants:
        raise ValueError(f"unsupported families: {sorted(families)}")
    return variants


def benchmark_ranked_variant(
    benchmark: str,
    name: str,
    rebalance_every: int,
    lookback: int,
    sleeve: float,
    top_k: int,
    max_name: float,
    min_relative_momentum: float,
) -> dict[str, Any]:
    return {
        "name": name,
        "family": "benchmark_ranked_sleeve",
        "benchmark": benchmark,
        "rebalance_every": rebalance_every,
        "global_max_name_weight": max_name,
        "legs": [
            {
                "name": "risk_budgeted_stocks",
                "candidate_universe": "stocks",
                "rank_mode": "vol_adjusted_momentum",
                "weight_mode": "risk_adjusted_edge",
                "lookback_bars": lookback,
                "sleeve_fraction": sleeve,
                "top_k": top_k,
                "max_name_weight": max_name,
                "min_relative_momentum": min_relative_momentum,
                "max_vol_20": 0.45,
                "edge_exponent": 2.0,
                "vol_floor": 0.10,
            }
        ],
    }


def sector_ranked_variant(
    benchmark: str,
    name: str,
    rebalance_every: int,
    lookback: int,
    sleeve: float,
    top_k: int,
    max_name: float,
    min_relative_momentum: float,
) -> dict[str, Any]:
    return {
        "name": name,
        "family": "sector_ranked_sleeve",
        "benchmark": benchmark,
        "rebalance_every": rebalance_every,
        "global_max_name_weight": max_name,
        "legs": [
            {
                "name": "risk_budgeted_sector_etfs",
                "candidate_symbols": ["DIA", "IWM", "QQQ", "SMH", "VTI", "XLB", "XLE", "XLF", "XLI", "XLK", "XLP", "XLU", "XLV", "XLY"],
                "candidate_universe": "etfs",
                "rank_mode": "vol_adjusted_momentum",
                "weight_mode": "risk_adjusted_edge",
                "lookback_bars": lookback,
                "sleeve_fraction": sleeve,
                "top_k": top_k,
                "max_name_weight": max_name,
                "min_relative_momentum": min_relative_momentum,
                "max_vol_20": 0.35,
                "edge_exponent": 2.0,
                "vol_floor": 0.08,
            }
        ],
    }


def annual_folds(bars: pd.DataFrame, start_year: int, end_year: int) -> list[dict[str, Any]]:
    bars = bars.sort_values(["time", "symbol"]).copy()
    max_year = int(bars["time"].dt.year.max())
    if end_year <= 0:
        end_year = max_year
    folds: list[dict[str, Any]] = []
    for year in range(start_year, end_year + 1):
        start = pd.Timestamp(year=year, month=1, day=1, tz="UTC")
        end = pd.Timestamp(year=year + 1, month=1, day=1, tz="UTC")
        history = bars[bars["time"] < start].copy()
        test = bars[(bars["time"] >= start) & (bars["time"] < end)].copy()
        if history.empty or test.empty or test["time"].nunique() < 30:
            continue
        folds.append(
            {
                "name": str(year),
                "test_start": test["time"].min(),
                "test_end": test["time"].max(),
                "history_bars": history,
                "test_bars": test,
            }
        )
    return folds


def aggregate_variant(variant: dict[str, Any], folds: list[dict[str, Any]]) -> dict[str, Any]:
    strategy_growth = 1.0
    benchmark_growth = 1.0
    folds_beating = 0
    max_drawdown = 0.0
    benchmark_max_drawdown = 0.0
    total_turnover = 0.0
    total_cost = 0.0
    min_alpha_symbols = 10**9
    for fold in folds:
        strategy_return = finite_float(fold["strategy_return"])
        benchmark_return = finite_float(fold["benchmark_return"])
        excess_return = finite_float(fold["excess_return"])
        strategy_growth *= 1.0 + strategy_return
        benchmark_growth *= 1.0 + benchmark_return
        if excess_return > 0:
            folds_beating += 1
        max_drawdown = max(max_drawdown, finite_float(fold["max_drawdown"]))
        benchmark_max_drawdown = max(benchmark_max_drawdown, finite_float(fold["benchmark_max_drawdown"]))
        total_turnover += finite_float(fold["turnover"])
        total_cost += finite_float(fold["total_cost"])
        min_alpha_symbols = min(min_alpha_symbols, fold["num_alpha_symbols_traded"])
    fold_count = max(1, len(folds))
    row = {
        "variant": variant["name"],
        "family": variant["family"],
        "fold_count": len(folds),
        "compounded_strategy_return": strategy_growth - 1.0,
        "compounded_benchmark_return": benchmark_growth - 1.0,
        "compounded_excess_return": strategy_growth - benchmark_growth,
        "folds_beating_benchmark": folds_beating,
        "max_fold_drawdown": max_drawdown,
        "benchmark_max_fold_drawdown": benchmark_max_drawdown,
        "mean_turnover": total_turnover / fold_count,
        "total_cost": total_cost,
        "min_alpha_symbols_per_fold": 0 if min_alpha_symbols == 10**9 else min_alpha_symbols,
        "rebalance_every": variant["rebalance_every"],
        "global_max_name_weight": variant["global_max_name_weight"],
        "legs": json.dumps(variant["legs"], sort_keys=True),
    }
    row["decision"] = checkpoint_decision(row)
    row["decision_rank"] = decision_rank(row["decision"], row)
    return row


def finite_float(value: Any, default: float = 0.0) -> float:
    try:
        out = float(value)
    except Exception:
        return default
    if out != out or out in (float("inf"), float("-inf")):
        return default
    return out


def checkpoint_decision(row: dict[str, Any]) -> str:
    if row["compounded_strategy_return"] <= row["compounded_benchmark_return"]:
        return "reject_under_benchmark"
    if row["min_alpha_symbols_per_fold"] < 5:
        return "reject_insufficient_breadth"
    if row["max_fold_drawdown"] > row["benchmark_max_fold_drawdown"] + 0.03:
        return "reject_drawdown_regression"
    if row["mean_turnover"] > 8.0:
        return "reject_high_turnover"
    if row["folds_beating_benchmark"] == row["fold_count"]:
        return "candidate_all_folds"
    if row["folds_beating_benchmark"] >= max(1, row["fold_count"] - 1):
        return "research_checkpoint_near_all_folds"
    return "reject_weak_fold_repeatability"


def decision_rank(decision: str, row: dict[str, Any]) -> float:
    base = {
        "candidate_all_folds": 5.0,
        "research_checkpoint_near_all_folds": 4.0,
        "reject_weak_fold_repeatability": 3.0,
        "reject_drawdown_regression": 2.0,
        "reject_high_turnover": 2.0,
        "reject_insufficient_breadth": 1.0,
        "reject_under_benchmark": 0.0,
    }.get(decision, 0.0)
    return base + 0.01 * float(row["folds_beating_benchmark"])


def write_markdown(path: Path, report: dict[str, Any]) -> None:
    s = report["summary"]
    lines = [
        "# Costed Active-Sleeve Comparison\n\n",
        f"- Benchmark: `{s['benchmark']}`\n",
        f"- Status: `{s['status']}`\n",
        f"- Promotion note: {s['promotion_note']}\n",
        f"- Variant count: `{s['variant_count']}`\n",
        f"- Champion: `{s['champion']}`\n",
        f"- Champion decision: `{s['champion_decision']}`\n\n",
        "| Variant | Decision | Return | Benchmark | Excess | Folds Beating | Max DD | Benchmark Max DD | Mean Turnover |\n",
        "|---|---|---:|---:|---:|---:|---:|---:|---:|\n",
    ]
    for row in report["variants"]:
        lines.append(
            f"| {row['variant']} | {row['decision']} | {row['compounded_strategy_return']*100:.2f}% | "
            f"{row['compounded_benchmark_return']*100:.2f}% | {row['compounded_excess_return']*100:.2f}% | "
            f"{row['folds_beating_benchmark']}/{row['fold_count']} | {row['max_fold_drawdown']*100:.2f}% | "
            f"{row['benchmark_max_fold_drawdown']*100:.2f}% | {row['mean_turnover']:.3f} |\n"
        )
    path.write_text("".join(lines), encoding="utf-8")


if __name__ == "__main__":
    main()
