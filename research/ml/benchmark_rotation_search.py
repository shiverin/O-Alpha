#!/usr/bin/env python3
"""Run checkpointed challenger tests for benchmark-funded rotation."""

from __future__ import annotations

import argparse
import json
from dataclasses import asdict, dataclass
from pathlib import Path
from typing import Any

import lightgbm as lgb

from backtest_benchmark_rotation import (
    ExecutionConfig,
    RotationPolicy,
    load_bars,
    load_features,
    load_signals,
    load_yaml,
    model_path,
    run_rotation,
    write_json,
)
from artifact_manifest import command_line, research_status_accepted, write_manifest


@dataclass(frozen=True)
class FoldSpec:
    name: str
    train_window: str
    test_window: str
    panel_dir: Path
    metadata_path: Path


@dataclass(frozen=True)
class VariantSpec:
    name: str
    policy: RotationPolicy
    note: str


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--root", default="reports/batches/2026-06-02_generalized_equity_meta")
    parser.add_argument("--feature-spec", default="research/ml/feature_spec.yaml")
    parser.add_argument("--benchmark", default="SPY")
    parser.add_argument("--initial-cash", type=float, default=100000.0)
    parser.add_argument("--execution-mode", choices=["next_open", "close_to_close"], default="next_open")
    parser.add_argument("--spread-bps", type=float, default=2.0)
    parser.add_argument("--slippage-bps", type=float, default=1.0)
    parser.add_argument("--commission-per-share", type=float, default=0.0)
    parser.add_argument("--min-commission", type=float, default=0.0)
    parser.add_argument("--sec-fees-bps-sell", type=float, default=0.0)
    parser.add_argument("--out-dir", default="")
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    root = Path(args.root)
    out_dir = Path(args.out_dir) if args.out_dir else root / "rotation_policy_search"
    out_dir.mkdir(parents=True, exist_ok=True)
    feature_names = load_yaml(args.feature_spec)["features"]
    folds = default_folds(root)
    variants = default_variants()
    execution = ExecutionConfig(
        mode=args.execution_mode,
        spread_bps=args.spread_bps,
        slippage_bps=args.slippage_bps,
        commission_per_share=args.commission_per_share,
        min_commission=args.min_commission,
        sec_fees_bps_sell=args.sec_fees_bps_sell,
    )

    rows: list[dict[str, Any]] = []
    for variant in variants:
        fold_results = [
            evaluate_fold(
                fold=fold,
                variant=variant,
                feature_names=feature_names,
                benchmark=args.benchmark.upper(),
                initial_cash=args.initial_cash,
                execution=execution,
            )
            for fold in folds
        ]
        aggregate = aggregate_variant(variant, fold_results)
        rows.append(aggregate)
        write_json(out_dir / f"{variant.name}.json", aggregate)

    champion = next(row for row in rows if row["variant"] == "checkpoint_default_ml_all")
    for row in rows:
        row["checkpoint_decision"] = checkpoint_decision(row, champion)

    promoted = [row for row in rows if row["checkpoint_decision"] in {"promote", "provisional_checkpoint"}]
    if promoted:
        current = sorted(
            promoted,
            key=lambda row: (row["compounded"]["excess_vs_benchmark"], row["compounded"]["strategy_return"]),
            reverse=True,
        )[0]
    else:
        current = champion

    summary = {
        "previous_checkpoint": champion,
        "current_checkpoint": current,
        "variants": sorted(
            rows,
            key=lambda row: (row["compounded"]["excess_vs_benchmark"], row["compounded"]["strategy_return"]),
            reverse=True,
        ),
        "promotion_gate": {
            "must_improve_compounded_strategy_return_vs_checkpoint": True,
            "must_improve_compounded_excess_vs_checkpoint": True,
            "must_beat_compounded_benchmark": True,
            "must_not_worsen_checkpoint_max_drawdown_by_more_than": 0.02,
            "must_have_at_least_alpha_trades_for_research_checkpoint": 1,
            "full_promotion_requires_at_least_alpha_trades": 3,
            "full_promotion_requires_folds_beating_benchmark": 2,
        },
        "manifest": {
            "backtest_command": command_line(),
            "benchmark": args.benchmark.upper(),
            "cost_model": asdict(execution),
        },
    }
    write_json(out_dir / "checkpoint_summary.json", summary)
    write_json(out_dir / "current_checkpoint.json", current)
    write_json(
        out_dir / "rejected_variants.json",
        {"variants": [row for row in rows if row["checkpoint_decision"].startswith("reject_")]},
    )
    write_markdown(out_dir / "checkpoint_summary.md", summary)
    write_manifest(out_dir / "checkpoint_manifest.json", summary["manifest"])
    print(
        f"checkpoint={current['variant']} "
        f"return={current['compounded']['strategy_return']*100:.2f}% "
        f"benchmark={current['compounded']['benchmark_return']*100:.2f}% "
        f"excess={current['compounded']['excess_vs_benchmark']*100:.2f}% "
        f"decision={current['checkpoint_decision']}"
    )


def default_folds(root: Path) -> list[FoldSpec]:
    return [
        FoldSpec(
            name="2023_oos",
            train_window="2020-07-27 to 2022-12-31",
            test_window="2023-01-01 to 2023-12-31",
            panel_dir=root / "walkforward/test_2023_21_126",
            metadata_path=root / "walkforward/train_2020_2022_21_126/artifact/metadata.json",
        ),
        FoldSpec(
            name="2024_oos",
            train_window="2020-07-27 to 2023-12-31",
            test_window="2024-01-01 to 2024-12-31",
            panel_dir=root / "walkforward/test_2024_21_126",
            metadata_path=root / "walkforward/train_2020_2023_21_126/artifact/metadata.json",
        ),
        FoldSpec(
            name="2025_2026_oos",
            train_window="2020-07-27 to 2024-12-31",
            test_window="2025-01-01 to 2026-06-01",
            panel_dir=root / "oos_2025_2026_panel",
            metadata_path=root / "train_21_126/artifact/metadata.json",
        ),
    ]


def default_variants() -> list[VariantSpec]:
    return [
        VariantSpec("checkpoint_default_ml_all", RotationPolicy(), "Existing checkpoint: model threshold, all assets."),
        VariantSpec(
            "model_status_guard",
            RotationPolicy(require_candidate_model=True),
            "Disable alpha allocation when the trained model artifact is rejected.",
        ),
        VariantSpec(
            "model_status_guard_threshold_floor_046",
            RotationPolicy(require_candidate_model=True, threshold_floor=0.46),
            "Use only accepted model artifacts and require an absolute 0.46 score floor.",
        ),
        VariantSpec(
            "model_status_guard_threshold_floor_047",
            RotationPolicy(require_candidate_model=True, threshold_floor=0.47),
            "Use only accepted model artifacts and require an absolute 0.47 score floor.",
        ),
        VariantSpec(
            "model_status_guard_stocks",
            RotationPolicy(require_candidate_model=True, candidate_universe="stocks"),
            "Use only accepted model artifacts and rotate only into single-name equities.",
        ),
        VariantSpec("stocks_only", RotationPolicy(candidate_universe="stocks"), "Only rotate into single-name equities."),
        VariantSpec("etfs_only", RotationPolicy(candidate_universe="etfs"), "Only rotate into non-benchmark ETFs."),
        VariantSpec("threshold_floor_046", RotationPolicy(threshold_floor=0.46), "Require a minimum absolute model score."),
        VariantSpec("threshold_floor_047", RotationPolicy(threshold_floor=0.47), "Stricter minimum model score."),
        VariantSpec("threshold_floor_048", RotationPolicy(threshold_floor=0.48), "Very strict minimum model score."),
        VariantSpec(
            "threshold_floor_047_stocks",
            RotationPolicy(threshold_floor=0.47, candidate_universe="stocks"),
            "Strict model score and single-name-only alpha.",
        ),
        VariantSpec(
            "rs_positive",
            RotationPolicy(min_relative_strength_21=0.0),
            "Require positive 21-day relative strength versus SPY.",
        ),
        VariantSpec(
            "rs_positive_stocks",
            RotationPolicy(candidate_universe="stocks", min_relative_strength_21=0.0),
            "Single-name-only alpha with positive 21-day relative strength.",
        ),
        VariantSpec(
            "strong_rs_5pct",
            RotationPolicy(min_relative_strength_21=0.05),
            "Require at least 5 percentage points of 21-day relative strength.",
        ),
        VariantSpec(
            "positive_21d_stock",
            RotationPolicy(candidate_universe="stocks", min_log_ret_21=0.0),
            "Single-name alpha must have positive 21-day return.",
        ),
        VariantSpec(
            "low_vol_35",
            RotationPolicy(max_close_to_close_vol_20=0.35),
            "Avoid candidates with annualized 20-day close-to-close vol above 35%.",
        ),
        VariantSpec(
            "stocks_low_vol_35_rs_positive",
            RotationPolicy(
                candidate_universe="stocks",
                min_relative_strength_21=0.0,
                max_close_to_close_vol_20=0.35,
            ),
            "Single-name alpha with positive relative strength and moderate realized vol.",
        ),
        VariantSpec("stop_loss_8pct", RotationPolicy(stop_loss_pct=0.08), "Exit alpha on 8% drawdown from entry."),
        VariantSpec(
            "stop_loss_10_take_profit_25",
            RotationPolicy(stop_loss_pct=0.10, take_profit_pct=0.25),
            "Exit alpha on 10% loss or 25% gain.",
        ),
        VariantSpec("max_hold_126", RotationPolicy(max_hold_bars=126), "Force alpha back to benchmark after roughly six months."),
        VariantSpec(
            "prob_plus_momentum",
            RotationPolicy(selection_mode="probability_plus_momentum", selection_momentum_weight=0.25),
            "Break same-day ties by blending probability with 21-day relative strength.",
        ),
    ]


def evaluate_fold(
    fold: FoldSpec,
    variant: VariantSpec,
    feature_names: list[str],
    benchmark: str,
    initial_cash: float,
    execution: ExecutionConfig,
) -> dict[str, Any]:
    metadata = json.loads(fold.metadata_path.read_text(encoding="utf-8"))
    threshold = float(metadata.get("thresholds", {}).get("enter_long", 0.5))
    model = lgb.Booster(model_file=str(model_path(fold.metadata_path, metadata)))
    result = run_rotation(
        bars=load_bars(fold.panel_dir / "bars.csv"),
        signals=load_signals(fold.panel_dir / "signals.csv"),
        features=load_features(fold.panel_dir / "dataset/features.csv", feature_names),
        feature_names=feature_names,
        model=model,
        threshold=threshold,
        benchmark=benchmark,
        initial_cash=initial_cash,
        use_ml=True,
        policy=variant.policy,
        allow_alpha=not variant.policy.require_candidate_model or research_status_accepted(metadata.get("status")),
        execution=execution,
        calibration=metadata.get("calibration", {"method": "none"}),
    )
    summary = result["summary"]
    return {
        "fold": fold.name,
        "train_window": fold.train_window,
        "test_window": fold.test_window,
        "model_status": metadata.get("status", "unknown"),
        "model_status_reason": metadata.get("status_reason", "unknown"),
        "summary": summary,
        "trades": result["trades"],
        "decisions": result["decisions"],
    }


def aggregate_variant(variant: VariantSpec, fold_results: list[dict[str, Any]]) -> dict[str, Any]:
    strategy_growth = 1.0
    benchmark_growth = 1.0
    max_drawdown = 0.0
    benchmark_max_drawdown = 0.0
    total_trades = 0
    folds_with_alpha_trades = 0
    folds_beating_benchmark = 0
    for result in fold_results:
        summary = result["summary"]
        strategy_growth *= 1.0 + summary["total_return"]
        benchmark_growth *= 1.0 + summary["benchmark_return"]
        max_drawdown = max(max_drawdown, summary["max_drawdown"])
        benchmark_max_drawdown = max(benchmark_max_drawdown, summary["benchmark_max_drawdown"])
        total_trades += summary["num_trades"]
        if summary["num_trades"] > 0:
            folds_with_alpha_trades += 1
        if summary["excess_return_vs_benchmark"] > 0:
            folds_beating_benchmark += 1
    strategy_return = strategy_growth - 1.0
    benchmark_return = benchmark_growth - 1.0
    return {
        "variant": variant.name,
        "note": variant.note,
        "policy": asdict(variant.policy),
        "folds": fold_results,
        "compounded": {
            "strategy_return": strategy_return,
            "benchmark_return": benchmark_return,
            "excess_vs_benchmark": strategy_return - benchmark_return,
            "max_fold_drawdown": max_drawdown,
            "benchmark_max_fold_drawdown": benchmark_max_drawdown,
            "total_trades": total_trades,
            "folds_with_alpha_trades": folds_with_alpha_trades,
            "folds_beating_benchmark": folds_beating_benchmark,
        },
    }


def checkpoint_decision(row: dict[str, Any], champion: dict[str, Any]) -> str:
    if row["variant"] == champion["variant"]:
        return "checkpoint"
    metrics = row["compounded"]
    champion_metrics = champion["compounded"]
    if metrics["strategy_return"] <= champion_metrics["strategy_return"]:
        return "reject_lower_return"
    if metrics["excess_vs_benchmark"] <= champion_metrics["excess_vs_benchmark"]:
        return "reject_lower_excess"
    if metrics["strategy_return"] <= metrics["benchmark_return"]:
        return "reject_under_benchmark"
    if metrics["max_fold_drawdown"] > champion_metrics["max_fold_drawdown"] + 0.02:
        return "reject_drawdown_regression"
    if metrics["total_trades"] < 1:
        return "reject_too_few_trades"
    if metrics["total_trades"] >= 3 and metrics["folds_beating_benchmark"] >= 2:
        return "promote"
    return "provisional_checkpoint"


def write_markdown(path: Path, summary: dict[str, Any]) -> None:
    current = summary["current_checkpoint"]
    previous = summary["previous_checkpoint"]
    lines = [
        "# Benchmark Rotation Checkpoint Search\n\n",
        "The search compares challenger policies against the previous ML rotation checkpoint on the same chronological folds. ",
        "Earlier 2023 and 2024 artifacts remain diagnostic because their models are rejected for insufficient effective training events.\n\n",
        "## Checkpoint\n\n",
        f"- Previous checkpoint: `{previous['variant']}`\n",
        f"- Current checkpoint: `{current['variant']}`\n",
        f"- Decision: `{current['checkpoint_decision']}`\n",
        f"- Compounded strategy return: {current['compounded']['strategy_return']*100:.2f}%\n",
        f"- Compounded benchmark return: {current['compounded']['benchmark_return']*100:.2f}%\n",
        f"- Compounded excess: {current['compounded']['excess_vs_benchmark']*100:.2f}%\n",
        f"- Total alpha trades: {current['compounded']['total_trades']}\n\n",
        "## Leaderboard\n\n",
        "| Variant | Decision | Return | Benchmark | Excess | Max Fold DD | Trades | Folds Beat Benchmark |\n",
        "|---|---|---:|---:|---:|---:|---:|---:|\n",
    ]
    for row in summary["variants"]:
        metrics = row["compounded"]
        lines.append(
            f"| `{row['variant']}` | `{row['checkpoint_decision']}` | "
            f"{metrics['strategy_return']*100:.2f}% | {metrics['benchmark_return']*100:.2f}% | "
            f"{metrics['excess_vs_benchmark']*100:.2f}% | {metrics['max_fold_drawdown']*100:.2f}% | "
            f"{metrics['total_trades']} | {metrics['folds_beating_benchmark']} |\n"
        )
    lines.append("\n## Fold Detail\n\n")
    for row in summary["variants"][:8]:
        lines.append(f"### {row['variant']}\n\n")
        lines.append("| Fold | Model Status | Return | Benchmark | Excess | Trades |\n")
        lines.append("|---|---|---:|---:|---:|---:|\n")
        for fold in row["folds"]:
            s = fold["summary"]
            lines.append(
                f"| {fold['fold']} | {fold['model_status']} / {fold['model_status_reason']} | "
                f"{s['total_return']*100:.2f}% | {s['benchmark_return']*100:.2f}% | "
                f"{s['excess_return_vs_benchmark']*100:.2f}% | {s['num_trades']} |\n"
            )
        lines.append("\n")
    path.write_text("".join(lines), encoding="utf-8")


if __name__ == "__main__":
    main()
