#!/usr/bin/env python3
"""Export deterministic Python feature rows for Go feature parity checks."""

from __future__ import annotations

import argparse
from pathlib import Path
from typing import Any

import pandas as pd

from artifact_manifest import command_line, file_sha256, git_sha, write_manifest
from build_meta_dataset import load_yaml


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--features-csv", required=True)
    parser.add_argument("--feature-spec", required=True)
    parser.add_argument("--labeled-events-csv")
    parser.add_argument("--out", required=True)
    parser.add_argument("--max-rows", type=int, default=500)
    parser.add_argument("--symbols", default="", help="optional comma-separated symbol filter")
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    feature_spec = load_yaml(args.feature_spec)
    feature_names = list(feature_spec.get("features", []))
    if not feature_names:
        raise ValueError("feature spec must contain features")

    features = load_features(args.features_csv, feature_names)
    if args.labeled_events_csv:
        labels = load_labels(args.labeled_events_csv)
        features = labels[["symbol", "event_time"]].merge(features, on=["symbol", "event_time"], how="inner")
    symbols = parse_symbols(args.symbols)
    if symbols:
        features = features[features["symbol"].isin(symbols)]
    if features.empty:
        raise ValueError("no feature rows selected for fixture")

    selected = deterministic_sample(features, max(1, args.max_rows))
    out_path = Path(args.out)
    out_path.parent.mkdir(parents=True, exist_ok=True)
    selected[["symbol", "event_time", *feature_names]].to_csv(out_path, index=False)

    manifest = {
        "artifact_id": f"feature_fixture_{pd.Timestamp.utcnow().isoformat()}",
        "generated_by": command_line(),
        "git_sha": git_sha(),
        "features_csv": args.features_csv,
        "features_csv_sha256": file_sha256(args.features_csv),
        "feature_spec": args.feature_spec,
        "feature_spec_sha256": file_sha256(args.feature_spec),
        "labeled_events_csv": args.labeled_events_csv,
        "labeled_events_csv_sha256": file_sha256(args.labeled_events_csv),
        "fixture_csv": str(out_path),
        "selected_rows": int(len(selected)),
        "symbols": sorted(selected["symbol"].unique().tolist()),
        "feature_count": len(feature_names),
    }
    write_manifest(out_path.with_suffix(".manifest.json"), manifest)
    print(f"feature fixture written: {out_path} rows={len(selected)} features={len(feature_names)}")


def load_features(path: str | Path, feature_names: list[str]) -> pd.DataFrame:
    df = pd.read_csv(path)
    required = {"symbol", "event_time", *feature_names}
    if missing := required - set(df.columns):
        raise ValueError(f"features CSV missing columns: {sorted(missing)}")
    df["symbol"] = df["symbol"].astype(str).str.upper().str.strip()
    df["event_time"] = pd.to_datetime(df["event_time"], utc=True)
    return df.sort_values(["symbol", "event_time"]).reset_index(drop=True)


def load_labels(path: str | Path) -> pd.DataFrame:
    df = pd.read_csv(path)
    required = {"symbol", "event_time"}
    if missing := required - set(df.columns):
        raise ValueError(f"labeled events CSV missing columns: {sorted(missing)}")
    df["symbol"] = df["symbol"].astype(str).str.upper().str.strip()
    df["event_time"] = pd.to_datetime(df["event_time"], utc=True)
    return df.sort_values(["symbol", "event_time"]).drop_duplicates(["symbol", "event_time"])


def deterministic_sample(df: pd.DataFrame, max_rows: int) -> pd.DataFrame:
    df = df.sort_values(["event_time", "symbol"]).reset_index(drop=True)
    if len(df) <= max_rows:
        return df
    positions = sorted(set(round(i * (len(df) - 1) / (max_rows - 1)) for i in range(max_rows)))
    return df.iloc[positions].reset_index(drop=True)


def parse_symbols(value: str) -> set[str]:
    return {part.strip().upper() for part in value.split(",") if part.strip()}


if __name__ == "__main__":
    main()
