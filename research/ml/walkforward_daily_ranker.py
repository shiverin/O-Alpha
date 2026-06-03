#!/usr/bin/env python3
"""Walk-forward validation for daily LightGBM ranker sleeves.

Each fold trains a fresh model using only bars before the test year, then
backtests a benchmark-funded active sleeve over that OOS year. This is still a
Python research pre-screen; official alpha promotion remains in the Go harness.
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
from backtest_daily_ranker_sleeve import (
    RankerConfig,
    backtest_ranker_sleeve,
    build_daily_feature_frame,
    feature_importance_rows,
    promotion_status,
    train_ranker,
    universe_allows,
)
from portfolio_sleeve import write_csv, write_json


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--bars-csv", required=True)
    parser.add_argument("--benchmark", default="VOO")
    parser.add_argument("--candidate-universe", choices=["stocks", "all", "etfs"], default="stocks")
    parser.add_argument("--start-year", type=int, default=2021)
    parser.add_argument("--end-year", type=int, default=0, help="inclusive; default uses max year in bars")
    parser.add_argument("--min-history-bars", type=int, default=252)
    parser.add_argument("--validation-fraction", type=float, default=0.25)
    parser.add_argument("--relevance-bins", type=int, default=5)
    parser.add_argument("--objective", choices=["lambdarank", "regression"], default="lambdarank")
    parser.add_argument("--seed", type=int, default=17)
    parser.add_argument("--num-boost-round", type=int, default=400)
    parser.add_argument("--early-stopping-rounds", type=int, default=40)
    parser.add_argument("--initial-cash", type=float, default=100000.0)
    parser.add_argument("--spread-bps", type=float, default=2.0)
    parser.add_argument("--slippage-bps", type=float, default=1.0)
    parser.add_argument("--commission-per-share", type=float, default=0.0)
    parser.add_argument("--min-commission", type=float, default=0.0)
    parser.add_argument("--sec-fees-bps-sell", type=float, default=0.0)
    parser.add_argument("--rebalance-band", type=float, default=0.025)
    parser.add_argument("--variants", default="stocks_h21_s10_top3_reb42,stocks_h21_s15_top3_reb42,stocks_h42_s15_top5_reb42")
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
    variants = selected_variants(args.variants, args.candidate_universe)
    folds = annual_folds(bars, args.start_year, args.end_year)
    if not folds:
        raise ValueError("no annual folds available")

    out_dir = Path(args.out_dir)
    out_dir.mkdir(parents=True, exist_ok=True)
    fold_rows: list[dict[str, Any]] = []
    variant_rows: list[dict[str, Any]] = []
    for variant in variants:
        rows = []
        for fold in folds:
            row = run_fold(
                fold=fold,
                variant=variant,
                benchmark=benchmark,
                args=args,
                execution=execution,
                out_dir=out_dir,
            )
            rows.append(row)
            fold_rows.append(row)
        variant_rows.append(aggregate_variant(variant, rows))

    variant_rows.sort(
        key=lambda row: (
            row["decision_rank"],
            row["compounded_excess_return"],
            row["folds_beating_benchmark"],
        ),
        reverse=True,
    )
    report = {
        "summary": {
            "benchmark": benchmark,
            "status": "python_prescreen_only",
            "promotion_note": "official alpha promotion still requires cmd/alpha-research DSR/PBO gate",
            "fold_count": len(folds),
            "variant_count": len(variant_rows),
            "champion": variant_rows[0]["variant"] if variant_rows else "",
            "champion_decision": variant_rows[0]["decision"] if variant_rows else "",
            "folds": [{"name": fold["name"], "test_start": fold["test_start"].isoformat(), "test_end": fold["test_end"].isoformat()} for fold in folds],
        },
        "variants": variant_rows,
        "fold_results": fold_rows,
        "manifest": {
            "command": command_line(),
            "bars_csv": args.bars_csv,
            "bars_csv_sha256": file_sha256(args.bars_csv),
            "benchmark": benchmark,
            "cost_model": asdict(execution),
            "variants": variants,
        },
    }
    write_json(out_dir / "daily_ranker_walkforward.json", report)
    write_csv(out_dir / "daily_ranker_walkforward_variants.csv", variant_rows)
    write_csv(out_dir / "daily_ranker_walkforward_folds.csv", fold_rows)
    write_markdown(out_dir / "daily_ranker_walkforward.md", report)
    write_manifest(out_dir / "daily_ranker_walkforward_manifest.json", report["manifest"])

    champion = variant_rows[0]
    print(
        f"champion={champion['variant']} decision={champion['decision']} "
        f"return={champion['compounded_strategy_return']*100:.2f}% "
        f"benchmark={champion['compounded_benchmark_return']*100:.2f}% "
        f"excess={champion['compounded_excess_return']*100:.2f}% "
        f"folds_beating={champion['folds_beating_benchmark']}/{champion['fold_count']} "
        f"validation_positive={champion['folds_validation_positive']}/{champion['fold_count']}"
    )


def selected_variants(raw: str, default_universe: str) -> list[dict[str, Any]]:
    presets = {
        "stocks_h21_s10_top3_reb42": {
            "variant": "stocks_h21_s10_top3_reb42",
            "candidate_universe": "stocks",
            "horizon_bars": 21,
            "sleeve_fraction": 0.10,
            "top_k": 3,
            "max_name_weight": 0.04,
            "rebalance_every": 42,
            "min_score_z": 0.0,
            "max_candidate_vol": 0.0,
            "max_benchmark_vol": 0.0,
            "high_vol_scale": 1.0,
            "max_benchmark_drawdown": 0.0,
            "drawdown_scale": 1.0,
        },
        "stocks_h21_s15_top3_reb42": {
            "variant": "stocks_h21_s15_top3_reb42",
            "candidate_universe": "stocks",
            "horizon_bars": 21,
            "sleeve_fraction": 0.15,
            "top_k": 3,
            "max_name_weight": 0.05,
            "rebalance_every": 42,
            "min_score_z": 0.0,
            "max_candidate_vol": 0.0,
            "max_benchmark_vol": 0.0,
            "high_vol_scale": 1.0,
            "max_benchmark_drawdown": 0.0,
            "drawdown_scale": 1.0,
        },
        "stocks_h21_s10_top3_reb42_z05": {
            "variant": "stocks_h21_s10_top3_reb42_z05",
            "candidate_universe": "stocks",
            "horizon_bars": 21,
            "sleeve_fraction": 0.10,
            "top_k": 3,
            "max_name_weight": 0.04,
            "rebalance_every": 42,
            "min_score_z": 0.5,
            "max_candidate_vol": 0.0,
            "max_benchmark_vol": 0.0,
            "high_vol_scale": 1.0,
            "max_benchmark_drawdown": 0.0,
            "drawdown_scale": 1.0,
        },
        "stocks_h21_s15_top3_reb42_z05": {
            "variant": "stocks_h21_s15_top3_reb42_z05",
            "candidate_universe": "stocks",
            "horizon_bars": 21,
            "sleeve_fraction": 0.15,
            "top_k": 3,
            "max_name_weight": 0.05,
            "rebalance_every": 42,
            "min_score_z": 0.5,
            "max_candidate_vol": 0.0,
            "max_benchmark_vol": 0.0,
            "high_vol_scale": 1.0,
            "max_benchmark_drawdown": 0.0,
            "drawdown_scale": 1.0,
        },
        "stocks_h21_s15_top3_reb42_z10": {
            "variant": "stocks_h21_s15_top3_reb42_z10",
            "candidate_universe": "stocks",
            "horizon_bars": 21,
            "sleeve_fraction": 0.15,
            "top_k": 3,
            "max_name_weight": 0.05,
            "rebalance_every": 42,
            "min_score_z": 1.0,
            "max_candidate_vol": 0.0,
            "max_benchmark_vol": 0.0,
            "high_vol_scale": 1.0,
            "max_benchmark_drawdown": 0.0,
            "drawdown_scale": 1.0,
        },
        "stocks_h21_s15_top3_reb42_z15": {
            "variant": "stocks_h21_s15_top3_reb42_z15",
            "candidate_universe": "stocks",
            "horizon_bars": 21,
            "sleeve_fraction": 0.15,
            "top_k": 3,
            "max_name_weight": 0.05,
            "rebalance_every": 42,
            "min_score_z": 1.5,
            "max_candidate_vol": 0.0,
            "max_benchmark_vol": 0.0,
            "high_vol_scale": 1.0,
            "max_benchmark_drawdown": 0.0,
            "drawdown_scale": 1.0,
        },
        "stocks_h21_s15_top3_reb42_risk": {
            "variant": "stocks_h21_s15_top3_reb42_risk",
            "candidate_universe": "stocks",
            "horizon_bars": 21,
            "sleeve_fraction": 0.15,
            "top_k": 3,
            "max_name_weight": 0.05,
            "rebalance_every": 42,
            "min_score_z": 0.0,
            "max_candidate_vol": 0.0,
            "max_benchmark_vol": 0.28,
            "high_vol_scale": 0.5,
            "max_benchmark_drawdown": 0.08,
            "drawdown_scale": 0.5,
        },
        "stocks_h21_s15_top3_reb42_riskoff": {
            "variant": "stocks_h21_s15_top3_reb42_riskoff",
            "candidate_universe": "stocks",
            "horizon_bars": 21,
            "sleeve_fraction": 0.15,
            "top_k": 3,
            "max_name_weight": 0.05,
            "rebalance_every": 42,
            "min_score_z": 0.0,
            "max_candidate_vol": 0.0,
            "max_benchmark_vol": 0.25,
            "high_vol_scale": 0.0,
            "max_benchmark_drawdown": 0.05,
            "drawdown_scale": 0.0,
        },
        "stocks_h21_s15_top3_reb42_z05_risk": {
            "variant": "stocks_h21_s15_top3_reb42_z05_risk",
            "candidate_universe": "stocks",
            "horizon_bars": 21,
            "sleeve_fraction": 0.15,
            "top_k": 3,
            "max_name_weight": 0.05,
            "rebalance_every": 42,
            "min_score_z": 0.5,
            "max_candidate_vol": 0.0,
            "max_benchmark_vol": 0.28,
            "high_vol_scale": 0.5,
            "max_benchmark_drawdown": 0.08,
            "drawdown_scale": 0.5,
        },
        "stocks_h21_s15_top3_reb42_z10_riskoff": {
            "variant": "stocks_h21_s15_top3_reb42_z10_riskoff",
            "candidate_universe": "stocks",
            "horizon_bars": 21,
            "sleeve_fraction": 0.15,
            "top_k": 3,
            "max_name_weight": 0.05,
            "rebalance_every": 42,
            "min_score_z": 1.0,
            "max_candidate_vol": 0.0,
            "max_benchmark_vol": 0.25,
            "high_vol_scale": 0.0,
            "max_benchmark_drawdown": 0.05,
            "drawdown_scale": 0.0,
        },
        "stocks_h21_s15_top3_reb42_z10_vol30": {
            "variant": "stocks_h21_s15_top3_reb42_z10_vol30",
            "candidate_universe": "stocks",
            "horizon_bars": 21,
            "sleeve_fraction": 0.15,
            "top_k": 3,
            "max_name_weight": 0.05,
            "rebalance_every": 42,
            "min_score_z": 1.0,
            "max_candidate_vol": 0.30,
            "max_benchmark_vol": 0.0,
            "high_vol_scale": 1.0,
            "max_benchmark_drawdown": 0.0,
            "drawdown_scale": 1.0,
        },
        "stocks_h21_s15_top3_reb42_z10_vol35": {
            "variant": "stocks_h21_s15_top3_reb42_z10_vol35",
            "candidate_universe": "stocks",
            "horizon_bars": 21,
            "sleeve_fraction": 0.15,
            "top_k": 3,
            "max_name_weight": 0.05,
            "rebalance_every": 42,
            "min_score_z": 1.0,
            "max_candidate_vol": 0.35,
            "max_benchmark_vol": 0.0,
            "high_vol_scale": 1.0,
            "max_benchmark_drawdown": 0.0,
            "drawdown_scale": 1.0,
        },
        "stocks_h21_s15_top3_reb42_z10_vol35_riskoff": {
            "variant": "stocks_h21_s15_top3_reb42_z10_vol35_riskoff",
            "candidate_universe": "stocks",
            "horizon_bars": 21,
            "sleeve_fraction": 0.15,
            "top_k": 3,
            "max_name_weight": 0.05,
            "rebalance_every": 42,
            "min_score_z": 1.0,
            "max_candidate_vol": 0.35,
            "max_benchmark_vol": 0.25,
            "high_vol_scale": 0.0,
            "max_benchmark_drawdown": 0.05,
            "drawdown_scale": 0.0,
        },
        "stocks_h42_s15_top5_reb42": {
            "variant": "stocks_h42_s15_top5_reb42",
            "candidate_universe": "stocks",
            "horizon_bars": 42,
            "sleeve_fraction": 0.15,
            "top_k": 5,
            "max_name_weight": 0.04,
            "rebalance_every": 42,
            "min_score_z": 0.0,
            "max_candidate_vol": 0.0,
            "max_benchmark_vol": 0.0,
            "high_vol_scale": 1.0,
            "max_benchmark_drawdown": 0.0,
            "drawdown_scale": 1.0,
        },
        "stocks_h42_s10_top3_reb42_z10": {
            "variant": "stocks_h42_s10_top3_reb42_z10",
            "candidate_universe": "stocks",
            "horizon_bars": 42,
            "sleeve_fraction": 0.10,
            "top_k": 3,
            "max_name_weight": 0.04,
            "rebalance_every": 42,
            "min_score_z": 1.0,
            "max_candidate_vol": 0.0,
            "max_benchmark_vol": 0.0,
            "high_vol_scale": 1.0,
            "max_benchmark_drawdown": 0.0,
            "drawdown_scale": 1.0,
        },
        "stocks_h42_s15_top3_reb42_z10": {
            "variant": "stocks_h42_s15_top3_reb42_z10",
            "candidate_universe": "stocks",
            "horizon_bars": 42,
            "sleeve_fraction": 0.15,
            "top_k": 3,
            "max_name_weight": 0.05,
            "rebalance_every": 42,
            "min_score_z": 1.0,
            "max_candidate_vol": 0.0,
            "max_benchmark_vol": 0.0,
            "high_vol_scale": 1.0,
            "max_benchmark_drawdown": 0.0,
            "drawdown_scale": 1.0,
        },
        "stocks_h63_s10_top3_reb63_z10": {
            "variant": "stocks_h63_s10_top3_reb63_z10",
            "candidate_universe": "stocks",
            "horizon_bars": 63,
            "sleeve_fraction": 0.10,
            "top_k": 3,
            "max_name_weight": 0.04,
            "rebalance_every": 63,
            "min_score_z": 1.0,
            "max_candidate_vol": 0.0,
            "max_benchmark_vol": 0.0,
            "high_vol_scale": 1.0,
            "max_benchmark_drawdown": 0.0,
            "drawdown_scale": 1.0,
        },
        "stocks_h63_s15_top3_reb63_z10": {
            "variant": "stocks_h63_s15_top3_reb63_z10",
            "candidate_universe": "stocks",
            "horizon_bars": 63,
            "sleeve_fraction": 0.15,
            "top_k": 3,
            "max_name_weight": 0.05,
            "rebalance_every": 63,
            "min_score_z": 1.0,
            "max_candidate_vol": 0.0,
            "max_benchmark_vol": 0.0,
            "high_vol_scale": 1.0,
            "max_benchmark_drawdown": 0.0,
            "drawdown_scale": 1.0,
        },
        "stocks_h126_s10_top3_reb63_z10": {
            "variant": "stocks_h126_s10_top3_reb63_z10",
            "candidate_universe": "stocks",
            "horizon_bars": 126,
            "sleeve_fraction": 0.10,
            "top_k": 3,
            "max_name_weight": 0.04,
            "rebalance_every": 63,
            "min_score_z": 1.0,
            "max_candidate_vol": 0.0,
            "max_benchmark_vol": 0.0,
            "high_vol_scale": 1.0,
            "max_benchmark_drawdown": 0.0,
            "drawdown_scale": 1.0,
        },
        "stocks_h126_s15_top3_reb63_z10": {
            "variant": "stocks_h126_s15_top3_reb63_z10",
            "candidate_universe": "stocks",
            "horizon_bars": 126,
            "sleeve_fraction": 0.15,
            "top_k": 3,
            "max_name_weight": 0.05,
            "rebalance_every": 63,
            "min_score_z": 1.0,
            "max_candidate_vol": 0.0,
            "max_benchmark_vol": 0.0,
            "high_vol_scale": 1.0,
            "max_benchmark_drawdown": 0.0,
            "drawdown_scale": 1.0,
        },
        "etfs_h21_s15_top3_reb42_z10": {
            "variant": "etfs_h21_s15_top3_reb42_z10",
            "candidate_universe": "etfs",
            "horizon_bars": 21,
            "sleeve_fraction": 0.15,
            "top_k": 3,
            "max_name_weight": 0.05,
            "rebalance_every": 42,
            "min_score_z": 1.0,
            "max_candidate_vol": 0.0,
            "max_benchmark_vol": 0.0,
            "high_vol_scale": 1.0,
            "max_benchmark_drawdown": 0.0,
            "drawdown_scale": 1.0,
        },
        "etfs_h42_s15_top3_reb42_z10": {
            "variant": "etfs_h42_s15_top3_reb42_z10",
            "candidate_universe": "etfs",
            "horizon_bars": 42,
            "sleeve_fraction": 0.15,
            "top_k": 3,
            "max_name_weight": 0.05,
            "rebalance_every": 42,
            "min_score_z": 1.0,
            "max_candidate_vol": 0.0,
            "max_benchmark_vol": 0.0,
            "high_vol_scale": 1.0,
            "max_benchmark_drawdown": 0.0,
            "drawdown_scale": 1.0,
        },
        "etfs_h63_s15_top3_reb63_z10": {
            "variant": "etfs_h63_s15_top3_reb63_z10",
            "candidate_universe": "etfs",
            "horizon_bars": 63,
            "sleeve_fraction": 0.15,
            "top_k": 3,
            "max_name_weight": 0.05,
            "rebalance_every": 63,
            "min_score_z": 1.0,
            "max_candidate_vol": 0.0,
            "max_benchmark_vol": 0.0,
            "high_vol_scale": 1.0,
            "max_benchmark_drawdown": 0.0,
            "drawdown_scale": 1.0,
        },
        "etfs_h126_s15_top3_reb63_z10": {
            "variant": "etfs_h126_s15_top3_reb63_z10",
            "candidate_universe": "etfs",
            "horizon_bars": 126,
            "sleeve_fraction": 0.15,
            "top_k": 3,
            "max_name_weight": 0.05,
            "rebalance_every": 63,
            "min_score_z": 1.0,
            "max_candidate_vol": 0.0,
            "max_benchmark_vol": 0.0,
            "high_vol_scale": 1.0,
            "max_benchmark_drawdown": 0.0,
            "drawdown_scale": 1.0,
        },
        "all_h21_s10_top5_reb42": {
            "variant": "all_h21_s10_top5_reb42",
            "candidate_universe": "all",
            "horizon_bars": 21,
            "sleeve_fraction": 0.10,
            "top_k": 5,
            "max_name_weight": 0.03,
            "rebalance_every": 42,
            "min_score_z": 0.0,
            "max_candidate_vol": 0.0,
            "max_benchmark_vol": 0.0,
            "high_vol_scale": 1.0,
            "max_benchmark_drawdown": 0.0,
            "drawdown_scale": 1.0,
        },
    }
    names = [part.strip() for part in raw.split(",") if part.strip()]
    if not names:
        names = ["stocks_h21_s10_top3_reb42"]
    variants = []
    for name in names:
        if name not in presets:
            raise ValueError(f"unknown variant preset {name!r}")
        variant = dict(presets[name])
        if variant["candidate_universe"] == "default":
            variant["candidate_universe"] = default_universe
        variants.append(variant)
    return variants


def annual_folds(bars: pd.DataFrame, start_year: int, end_year: int) -> list[dict[str, Any]]:
    max_year = int(bars["time"].dt.year.max())
    if end_year <= 0:
        end_year = max_year
    folds = []
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


def run_fold(
    *,
    fold: dict[str, Any],
    variant: dict[str, Any],
    benchmark: str,
    args: argparse.Namespace,
    execution: ExecutionConfig,
    out_dir: Path,
) -> dict[str, Any]:
    config = RankerConfig(
        horizon_bars=int(variant["horizon_bars"]),
        min_history_bars=args.min_history_bars,
        validation_fraction=args.validation_fraction,
        relevance_bins=args.relevance_bins,
        objective=args.objective,
        seed=args.seed,
        num_boost_round=args.num_boost_round,
        early_stopping_rounds=args.early_stopping_rounds,
    )
    feature_frame, open_panel, close_panel = build_daily_feature_frame(
        fold["history_bars"],
        fold["test_bars"],
        benchmark,
        config,
    )
    train_rows = feature_frame[
        (feature_frame["source"] == "train")
        & feature_frame["label_excess"].notna()
        & feature_frame["symbol"].map(lambda symbol: universe_allows(symbol, variant["candidate_universe"], benchmark))
    ].copy()
    test_rows = feature_frame[
        (feature_frame["source"] == "test")
        & feature_frame["symbol"].map(lambda symbol: universe_allows(symbol, variant["candidate_universe"], benchmark))
    ].copy()
    if train_rows.empty or test_rows.empty:
        raise ValueError(f"fold {fold['name']} variant {variant['variant']} has empty train/test rows")
    model, training_report = train_ranker(train_rows, config)
    result = backtest_ranker_sleeve(
        model=model,
        test_rows=test_rows,
        open_panel=open_panel.loc[open_panel.index.isin(fold["test_bars"]["time"].unique())],
        close_panel=close_panel.loc[close_panel.index.isin(fold["test_bars"]["time"].unique())],
        benchmark=benchmark,
        initial_cash=args.initial_cash,
        execution=execution,
        sleeve_fraction=float(variant["sleeve_fraction"]),
        top_k=int(variant["top_k"]),
        max_name_weight=float(variant["max_name_weight"]),
        rebalance_every=int(variant["rebalance_every"]),
        rebalance_band=args.rebalance_band,
        min_score_z=float(variant.get("min_score_z", 0.0)),
        max_candidate_vol=float(variant.get("max_candidate_vol", 0.0)),
        max_benchmark_vol=float(variant.get("max_benchmark_vol", 0.0)),
        high_vol_scale=float(variant.get("high_vol_scale", 1.0)),
        max_benchmark_drawdown=float(variant.get("max_benchmark_drawdown", 0.0)),
        drawdown_scale=float(variant.get("drawdown_scale", 1.0)),
    )
    summary = result["summary"]
    status = promotion_status(
        {
            **summary,
            "candidate_universe": variant["candidate_universe"],
            "horizon_bars": variant["horizon_bars"],
            "sleeve_fraction": variant["sleeve_fraction"],
            "top_k": variant["top_k"],
            "max_name_weight": variant["max_name_weight"],
            "rebalance_every": variant["rebalance_every"],
        },
        training_report,
    )
    fold_dir = out_dir / "fold_artifacts" / variant["variant"] / fold["name"]
    fold_dir.mkdir(parents=True, exist_ok=True)
    model_path = fold_dir / "model.txt"
    model.save_model(str(model_path))
    write_json(fold_dir / "training_report.json", training_report)
    write_csv(fold_dir / "feature_importance.csv", feature_importance_rows(model))
    write_csv(fold_dir / "equity_curve.csv", result["equity_curve"])
    write_csv(fold_dir / "orders.csv", result["orders"])
    write_csv(fold_dir / "decisions.csv", result["decisions"])
    write_csv(fold_dir / "selections.csv", result["selections"])
    write_json(
        fold_dir / "summary.json",
        {
            "fold": fold["name"],
            "variant": variant,
            "summary": summary,
            "training": training_report,
            "status": status,
            "model_file": str(model_path),
        },
    )
    validation = training_report.get("validation_metrics", {})
    return {
        "variant": variant["variant"],
        "fold": fold["name"],
        "candidate_universe": variant["candidate_universe"],
        "horizon_bars": variant["horizon_bars"],
        "sleeve_fraction": variant["sleeve_fraction"],
        "top_k": variant["top_k"],
        "max_name_weight": variant["max_name_weight"],
        "rebalance_every": variant["rebalance_every"],
        "min_score_z": variant.get("min_score_z", 0.0),
        "max_candidate_vol": variant.get("max_candidate_vol", 0.0),
        "max_benchmark_vol": variant.get("max_benchmark_vol", 0.0),
        "high_vol_scale": variant.get("high_vol_scale", 1.0),
        "max_benchmark_drawdown": variant.get("max_benchmark_drawdown", 0.0),
        "drawdown_scale": variant.get("drawdown_scale", 1.0),
        "test_start": fold["test_start"].isoformat(),
        "test_end": fold["test_end"].isoformat(),
        "total_return": summary["total_return"],
        "benchmark_return": summary["benchmark_return"],
        "excess_return": summary["excess_return_vs_benchmark"],
        "sharpe": summary["sharpe"],
        "benchmark_sharpe": summary["benchmark_sharpe"],
        "max_drawdown": summary["max_drawdown"],
        "benchmark_max_drawdown": summary["benchmark_max_drawdown"],
        "turnover": summary["turnover"],
        "total_cost": summary["total_cost"],
        "num_orders": summary["num_orders"],
        "num_alpha_symbols_traded": summary["num_alpha_symbols_traded"],
        "mean_selected_forward_excess": summary["mean_selected_forward_excess"],
        "average_sleeve_scale": summary.get("average_sleeve_scale", 1.0),
        "validation_rank_ic": validation.get("mean_rank_ic", 0.0),
        "validation_top_minus_universe": validation.get("mean_top_minus_universe", 0.0),
        "status": status,
        "model_file": str(model_path),
    }


def aggregate_variant(variant: dict[str, Any], rows: list[dict[str, Any]]) -> dict[str, Any]:
    strategy_growth = 1.0
    benchmark_growth = 1.0
    folds_beating = 0
    folds_candidate = 0
    folds_validation_positive = 0
    max_drawdown = 0.0
    benchmark_max_drawdown = 0.0
    total_turnover = 0.0
    min_symbols = 10**9
    for row in rows:
        strategy_growth *= 1.0 + finite_float(row["total_return"])
        benchmark_growth *= 1.0 + finite_float(row["benchmark_return"])
        if finite_float(row["excess_return"]) > 0:
            folds_beating += 1
        if row["status"] == "candidate":
            folds_candidate += 1
        if finite_float(row["validation_rank_ic"]) > 0 and finite_float(row["validation_top_minus_universe"]) > 0:
            folds_validation_positive += 1
        max_drawdown = max(max_drawdown, finite_float(row["max_drawdown"]))
        benchmark_max_drawdown = max(benchmark_max_drawdown, finite_float(row["benchmark_max_drawdown"]))
        total_turnover += finite_float(row["turnover"])
        min_symbols = min(min_symbols, int(row["num_alpha_symbols_traded"]))
    fold_count = max(1, len(rows))
    out = {
        "variant": variant["variant"],
        "candidate_universe": variant["candidate_universe"],
        "horizon_bars": variant["horizon_bars"],
        "sleeve_fraction": variant["sleeve_fraction"],
        "top_k": variant["top_k"],
        "max_name_weight": variant["max_name_weight"],
        "rebalance_every": variant["rebalance_every"],
        "min_score_z": variant.get("min_score_z", 0.0),
        "max_candidate_vol": variant.get("max_candidate_vol", 0.0),
        "max_benchmark_vol": variant.get("max_benchmark_vol", 0.0),
        "high_vol_scale": variant.get("high_vol_scale", 1.0),
        "max_benchmark_drawdown": variant.get("max_benchmark_drawdown", 0.0),
        "drawdown_scale": variant.get("drawdown_scale", 1.0),
        "fold_count": len(rows),
        "compounded_strategy_return": strategy_growth - 1.0,
        "compounded_benchmark_return": benchmark_growth - 1.0,
        "compounded_excess_return": strategy_growth - benchmark_growth,
        "folds_beating_benchmark": folds_beating,
        "folds_candidate": folds_candidate,
        "folds_validation_positive": folds_validation_positive,
        "max_fold_drawdown": max_drawdown,
        "benchmark_max_fold_drawdown": benchmark_max_drawdown,
        "mean_turnover": total_turnover / fold_count,
        "min_alpha_symbols_per_fold": 0 if min_symbols == 10**9 else min_symbols,
    }
    out["decision"] = checkpoint_decision(out)
    out["decision_rank"] = decision_rank(out["decision"], out)
    return out


def checkpoint_decision(row: dict[str, Any]) -> str:
    if row["compounded_strategy_return"] <= row["compounded_benchmark_return"]:
        return "reject_under_benchmark"
    if row["folds_beating_benchmark"] < row["fold_count"]:
        return "reject_weak_fold_repeatability"
    if row["folds_validation_positive"] < row["fold_count"]:
        return "research_only_weak_validation"
    if row["mean_turnover"] > 8.0:
        return "research_only_high_turnover"
    if row["max_fold_drawdown"] > row["benchmark_max_fold_drawdown"] + 0.02:
        return "research_only_drawdown_regression"
    if row["min_alpha_symbols_per_fold"] < 5:
        return "research_only_insufficient_breadth"
    return "candidate_all_folds"


def decision_rank(decision: str, row: dict[str, Any]) -> float:
    base = {
        "candidate_all_folds": 5.0,
        "research_only_weak_validation": 4.0,
        "research_only_drawdown_regression": 3.0,
        "research_only_high_turnover": 3.0,
        "reject_weak_fold_repeatability": 2.0,
        "research_only_insufficient_breadth": 1.0,
        "reject_under_benchmark": 0.0,
    }.get(decision, 0.0)
    return base + 0.01 * row["folds_beating_benchmark"] + 0.001 * row["folds_validation_positive"]


def finite_float(value: Any, default: float = 0.0) -> float:
    try:
        out = float(value)
    except Exception:
        return default
    if out != out or out in (float("inf"), float("-inf")):
        return default
    return out


def write_markdown(path: Path, report: dict[str, Any]) -> None:
    s = report["summary"]
    lines = [
        "# Daily Ranker Walk-Forward\n\n",
        f"- Benchmark: `{s['benchmark']}`\n",
        f"- Status: `{s['status']}`\n",
        f"- Promotion note: {s['promotion_note']}\n",
        f"- Folds: `{s['fold_count']}`\n",
        f"- Variants: `{s['variant_count']}`\n",
        f"- Champion: `{s['champion']}`\n",
        f"- Champion decision: `{s['champion_decision']}`\n\n",
        "| Variant | Decision | Return | Benchmark | Excess | Folds Beat | Validation Positive | Candidate Folds | Max DD | Benchmark Max DD | Turnover |\n",
        "|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|\n",
    ]
    for row in report["variants"]:
        lines.append(
            f"| {row['variant']} | {row['decision']} | {row['compounded_strategy_return']*100:.2f}% | "
            f"{row['compounded_benchmark_return']*100:.2f}% | {row['compounded_excess_return']*100:.2f}% | "
            f"{row['folds_beating_benchmark']}/{row['fold_count']} | "
            f"{row['folds_validation_positive']}/{row['fold_count']} | "
            f"{row['folds_candidate']}/{row['fold_count']} | "
            f"{row['max_fold_drawdown']*100:.2f}% | {row['benchmark_max_fold_drawdown']*100:.2f}% | "
            f"{row['mean_turnover']:.3f} |\n"
        )
    path.write_text("".join(lines), encoding="utf-8")


if __name__ == "__main__":
    main()
