#!/usr/bin/env python3
"""Grid-search benchmark-core momentum sleeve variants across walk-forward folds."""

from __future__ import annotations

import argparse
import json
from dataclasses import asdict
from pathlib import Path
from typing import Any

from artifact_manifest import command_line, write_manifest
from backtest_benchmark_rotation import ExecutionConfig, load_bars
from backtest_momentum_sleeve import run_momentum_sleeve
from portfolio_sleeve import write_json


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--root", default="reports/batches/2026-06-02_generalized_equity_meta")
    parser.add_argument("--benchmark", default="VOO")
    parser.add_argument("--initial-cash", type=float, default=100000.0)
    parser.add_argument("--lookbacks", default="21,42,63,126,189,252")
    parser.add_argument("--top-ks", default="1,3,5,10")
    parser.add_argument("--sleeves", default="0.10,0.15,0.20,0.30")
    parser.add_argument("--universes", default="all,stocks,etfs")
    parser.add_argument("--allocation-modes", default="equal,score_over_vol")
    parser.add_argument("--min-relative-momentums", default="0.0,0.02,0.05,0.10")
    parser.add_argument("--rebalance-every", type=int, default=21)
    parser.add_argument("--rebalance-band", type=float, default=0.025)
    parser.add_argument("--spread-bps", type=float, default=2.0)
    parser.add_argument("--slippage-bps", type=float, default=1.0)
    parser.add_argument("--commission-per-share", type=float, default=0.0)
    parser.add_argument("--min-commission", type=float, default=0.0)
    parser.add_argument("--sec-fees-bps-sell", type=float, default=0.0)
    parser.add_argument("--out-dir", required=True)
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    root = Path(args.root)
    benchmark = args.benchmark.upper()
    execution = ExecutionConfig(
        mode="next_open",
        spread_bps=args.spread_bps,
        slippage_bps=args.slippage_bps,
        commission_per_share=args.commission_per_share,
        min_commission=args.min_commission,
        sec_fees_bps_sell=args.sec_fees_bps_sell,
    )
    folds = load_default_folds(root)
    variants = build_variants(args)
    rows: list[dict[str, Any]] = []
    for variant in variants:
        fold_results = []
        for fold in folds:
            result = run_momentum_sleeve(
                history_bars=fold["history_bars"],
                test_bars=fold["test_bars"],
                benchmark=benchmark,
                candidate_universe=variant["candidate_universe"],
                initial_cash=args.initial_cash,
                lookback_bars=variant["lookback_bars"],
                sleeve_fraction=variant["sleeve_fraction"],
                top_k=variant["top_k"],
                max_name_weight=variant["max_name_weight"],
                rebalance_every=args.rebalance_every,
                rebalance_band=args.rebalance_band,
                allocation_mode=variant["allocation_mode"],
                require_positive_momentum=True,
                min_relative_momentum=variant["min_relative_momentum"],
                execution=execution,
            )
            fold_results.append(
                {
                    "fold": fold["name"],
                    "test_window": fold["test_window"],
                    "summary": result["summary"],
                }
            )
        rows.append(aggregate_variant(variant, fold_results))

    rows.sort(
        key=lambda row: (
            row["decision_rank"],
            row["compounded"]["excess_vs_benchmark"],
            row["compounded"]["strategy_return"],
        ),
        reverse=True,
    )
    champion = rows[0]
    summary = {
        "current_checkpoint": champion,
        "variants": rows,
        "promotion_gate": {
            "must_beat_compounded_benchmark": True,
            "prefer_all_folds_beating_benchmark": True,
            "require_at_least_alpha_symbols": 5,
            "reject_if_max_fold_drawdown_regresses_more_than": 0.03,
            "reject_if_turnover_above": 8.0,
        },
        "manifest": {
            "command": command_line(),
            "benchmark": benchmark,
            "cost_model": asdict(execution),
            "folds": [{"name": fold["name"], "test_window": fold["test_window"]} for fold in folds],
        },
    }
    out_dir = Path(args.out_dir)
    out_dir.mkdir(parents=True, exist_ok=True)
    write_json(out_dir / "momentum_sleeve_search.json", summary)
    write_markdown(out_dir / "momentum_sleeve_search.md", summary)
    write_manifest(out_dir / "momentum_sleeve_search_manifest.json", summary["manifest"])
    c = champion["compounded"]
    print(
        f"checkpoint={champion['variant']} return={c['strategy_return']*100:.2f}% "
        f"benchmark={c['benchmark_return']*100:.2f}% excess={c['excess_vs_benchmark']*100:.2f}% "
        f"folds_beating={c['folds_beating_benchmark']} decision={champion['decision']}"
    )


def load_default_folds(root: Path) -> list[dict[str, Any]]:
    specs = [
        (
            "2023",
            "2023-01-01 to 2023-12-31",
            root / "walkforward/train_2020_2022_21_126/bars.csv",
            root / "walkforward/test_2023_21_126/bars.csv",
        ),
        (
            "2024",
            "2024-01-01 to 2024-12-31",
            root / "walkforward/train_2020_2023_21_126/bars.csv",
            root / "walkforward/test_2024_21_126/bars.csv",
        ),
        (
            "2025_2026",
            "2025-01-01 to 2026-06-01",
            root / "train_21_126/bars.csv",
            root / "oos_2025_2026_panel/bars.csv",
        ),
    ]
    return [
        {
            "name": name,
            "test_window": window,
            "history_bars": load_bars(history),
            "test_bars": load_bars(test),
        }
        for name, window, history, test in specs
    ]


def build_variants(args: argparse.Namespace) -> list[dict[str, Any]]:
    lookbacks = parse_ints(args.lookbacks)
    top_ks = parse_ints(args.top_ks)
    sleeves = parse_floats(args.sleeves)
    universes = parse_strings(args.universes)
    allocation_modes = parse_strings(args.allocation_modes)
    min_relative_momentums = parse_floats(args.min_relative_momentums)
    variants = []
    for lookback in lookbacks:
        for top_k in top_ks:
            for sleeve in sleeves:
                for universe in universes:
                    for allocation_mode in allocation_modes:
                        for min_relative_momentum in min_relative_momentums:
                            max_name_weight = min(0.30, max(0.03, sleeve / max(1, top_k)))
                            variants.append(
                                {
                                    "variant": (
                                        f"{universe}_l{lookback}_top{top_k}_s{int(sleeve*100)}_"
                                        f"m{int(min_relative_momentum*100)}_{allocation_mode}"
                                    ),
                                    "candidate_universe": universe,
                                    "lookback_bars": lookback,
                                    "top_k": top_k,
                                    "sleeve_fraction": sleeve,
                                    "max_name_weight": max_name_weight,
                                    "allocation_mode": allocation_mode,
                                    "min_relative_momentum": min_relative_momentum,
                                }
                            )
    return variants


def aggregate_variant(variant: dict[str, Any], fold_results: list[dict[str, Any]]) -> dict[str, Any]:
    strategy_growth = 1.0
    benchmark_growth = 1.0
    max_drawdown = 0.0
    benchmark_max_drawdown = 0.0
    total_turnover = 0.0
    total_cost = 0.0
    total_symbols = set()
    folds_beating = 0
    for result in fold_results:
        summary = result["summary"]
        strategy_growth *= 1.0 + summary["total_return"]
        benchmark_growth *= 1.0 + summary["benchmark_return"]
        max_drawdown = max(max_drawdown, summary["max_drawdown"])
        benchmark_max_drawdown = max(benchmark_max_drawdown, summary["benchmark_max_drawdown"])
        total_turnover += summary["turnover"]
        total_cost += summary["total_cost"]
        if summary["excess_return_vs_benchmark"] > 0:
            folds_beating += 1
        total_symbols.add((result["fold"], summary["num_alpha_symbols_traded"]))
    compounded = {
        "strategy_return": strategy_growth - 1.0,
        "benchmark_return": benchmark_growth - 1.0,
        "excess_vs_benchmark": strategy_growth - benchmark_growth,
        "max_fold_drawdown": max_drawdown,
        "benchmark_max_fold_drawdown": benchmark_max_drawdown,
        "folds_beating_benchmark": folds_beating,
        "mean_turnover": total_turnover / max(1, len(fold_results)),
        "total_cost": total_cost,
        "min_alpha_symbols_per_fold": min(result["summary"]["num_alpha_symbols_traded"] for result in fold_results),
    }
    decision = checkpoint_decision(compounded)
    return {
        **variant,
        "folds": fold_results,
        "compounded": compounded,
        "decision": decision,
        "decision_rank": decision_rank(decision, compounded),
    }


def checkpoint_decision(metrics: dict[str, Any]) -> str:
    if metrics["strategy_return"] <= metrics["benchmark_return"]:
        return "reject_under_benchmark"
    if metrics["min_alpha_symbols_per_fold"] < 5:
        return "reject_insufficient_breadth"
    if metrics["max_fold_drawdown"] > metrics["benchmark_max_fold_drawdown"] + 0.03:
        return "reject_drawdown_regression"
    if metrics["mean_turnover"] > 8.0:
        return "reject_high_turnover"
    if metrics["folds_beating_benchmark"] == 3:
        return "candidate_all_folds"
    if metrics["folds_beating_benchmark"] == 2:
        return "research_checkpoint_two_folds"
    return "reject_weak_fold_repeatability"


def decision_rank(decision: str, metrics: dict[str, Any]) -> float:
    rank = {
        "candidate_all_folds": 5.0,
        "research_checkpoint_two_folds": 4.0,
        "reject_weak_fold_repeatability": 3.0,
        "reject_drawdown_regression": 2.0,
        "reject_high_turnover": 2.0,
        "reject_insufficient_breadth": 1.0,
        "reject_under_benchmark": 0.0,
    }.get(decision, 0.0)
    return rank + 0.01 * metrics["folds_beating_benchmark"]


def parse_ints(raw: str) -> list[int]:
    return [int(part.strip()) for part in raw.split(",") if part.strip()]


def parse_floats(raw: str) -> list[float]:
    return [float(part.strip()) for part in raw.split(",") if part.strip()]


def parse_strings(raw: str) -> list[str]:
    return [part.strip() for part in raw.split(",") if part.strip()]


def write_markdown(path: Path, summary: dict[str, Any]) -> None:
    champion = summary["current_checkpoint"]
    c = champion["compounded"]
    lines = [
        "# Momentum Sleeve Walk-Forward Search\n\n",
        f"- Current checkpoint: `{champion['variant']}`\n",
        f"- Decision: `{champion['decision']}`\n",
        f"- Compounded strategy return: `{c['strategy_return']*100:.2f}%`\n",
        f"- Compounded benchmark return: `{c['benchmark_return']*100:.2f}%`\n",
        f"- Compounded excess: `{c['excess_vs_benchmark']*100:.2f}%`\n",
        f"- Folds beating benchmark: `{c['folds_beating_benchmark']}`\n",
        f"- Max fold drawdown: `{c['max_fold_drawdown']*100:.2f}%`\n",
        f"- Mean turnover: `{c['mean_turnover']:.3f}`\n\n",
        "## Top Variants\n\n",
        "| Variant | Decision | Return | Benchmark | Excess | Folds | Max DD | Turnover |\n",
        "|---|---|---:|---:|---:|---:|---:|---:|\n",
    ]
    for row in summary["variants"][:20]:
        m = row["compounded"]
        lines.append(
            f"| `{row['variant']}` | `{row['decision']}` | {m['strategy_return']*100:.2f}% | "
            f"{m['benchmark_return']*100:.2f}% | {m['excess_vs_benchmark']*100:.2f}% | "
            f"{m['folds_beating_benchmark']} | {m['max_fold_drawdown']*100:.2f}% | {m['mean_turnover']:.3f} |\n"
        )
    path.write_text("".join(lines), encoding="utf-8")


if __name__ == "__main__":
    main()
