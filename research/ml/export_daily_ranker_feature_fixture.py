#!/usr/bin/env python3
"""Export daily ranker feature rows for Go parity checks."""

from __future__ import annotations

import argparse
from dataclasses import asdict
from pathlib import Path

import lightgbm as lgb
import pandas as pd

from artifact_manifest import command_line, file_sha256, git_sha, write_manifest
from backtest_benchmark_rotation import load_bars
from backtest_daily_ranker_sleeve import FEATURE_NAMES, RankerConfig, build_daily_feature_frame, universe_allows


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--bars-csv", required=True)
    parser.add_argument("--benchmark", default="VOO")
    parser.add_argument("--candidate-universe", choices=["all", "stocks", "etfs"], default="stocks")
    parser.add_argument("--horizon-bars", type=int, default=63)
    parser.add_argument("--min-history-bars", type=int, default=252)
    parser.add_argument("--max-rows", type=int, default=500)
    parser.add_argument("--model", default="", help="optional LightGBM model.txt; adds expected_score")
    parser.add_argument("--symbols", default="", help="optional comma-separated symbol filter")
    parser.add_argument("--out", required=True)
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    benchmark = args.benchmark.upper().strip()
    bars = load_bars(args.bars_csv)
    empty = bars.iloc[0:0].copy()
    config = RankerConfig(
        horizon_bars=args.horizon_bars,
        min_history_bars=args.min_history_bars,
        validation_fraction=0.25,
        relevance_bins=5,
        objective="lambdarank",
        seed=17,
        num_boost_round=400,
        early_stopping_rounds=40,
    )
    feature_frame, _, _ = build_daily_feature_frame(bars, empty, benchmark, config)
    selected = feature_frame[
        feature_frame["symbol"].map(lambda symbol: universe_allows(symbol, args.candidate_universe, benchmark))
    ].copy()
    symbols = parse_symbols(args.symbols)
    if symbols:
        selected = selected[selected["symbol"].isin(symbols)]
    if selected.empty:
        raise ValueError("no daily ranker feature rows selected")

    selected = deterministic_sample(selected.sort_values(["time", "symbol"]).reset_index(drop=True), args.max_rows)
    out_columns = ["symbol", "time", *FEATURE_NAMES]
    if args.model:
        model = lgb.Booster(model_file=args.model)
        selected["expected_score"] = model.predict(selected[FEATURE_NAMES].astype(float), num_iteration=model.best_iteration)
        out_columns.append("expected_score")
    out = selected[out_columns].rename(columns={"time": "event_time"})
    out_path = Path(args.out)
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out.to_csv(out_path, index=False)

    manifest = {
        "artifact_id": f"daily_ranker_feature_fixture_{pd.Timestamp.utcnow().isoformat()}",
        "generated_by": command_line(),
        "git_sha": git_sha(),
        "bars_csv": args.bars_csv,
        "bars_csv_sha256": file_sha256(args.bars_csv),
        "fixture_csv": str(out_path),
        "benchmark": benchmark,
        "candidate_universe": args.candidate_universe,
        "ranker_config": asdict(config),
        "features": FEATURE_NAMES,
        "feature_count": len(FEATURE_NAMES),
        "model": args.model,
        "model_sha256": file_sha256(args.model),
        "has_expected_score": bool(args.model),
        "selected_rows": int(len(out)),
        "symbols": sorted(out["symbol"].unique().tolist()),
    }
    write_manifest(out_path.with_suffix(".manifest.json"), manifest)
    print(f"daily ranker feature fixture written: {out_path} rows={len(out)} features={len(FEATURE_NAMES)}")


def deterministic_sample(df: pd.DataFrame, max_rows: int) -> pd.DataFrame:
    max_rows = max(1, int(max_rows))
    if len(df) <= max_rows:
        return df
    positions = sorted(set(round(i * (len(df) - 1) / (max_rows - 1)) for i in range(max_rows)))
    return df.iloc[positions].reset_index(drop=True)


def parse_symbols(value: str) -> set[str]:
    return {part.strip().upper() for part in value.split(",") if part.strip()}


if __name__ == "__main__":
    main()
