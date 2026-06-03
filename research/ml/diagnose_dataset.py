#!/usr/bin/env python3
"""Diagnose ML meta-label dataset liveness and label breadth."""

from __future__ import annotations

import argparse
import json
from pathlib import Path
from typing import Any

import numpy as np
import pandas as pd

from artifact_manifest import command_line, git_sha, write_manifest


KEY_COLUMNS = {"symbol", "event_time"}


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--dataset-dir", help="directory containing features.csv and labeled_events.csv")
    parser.add_argument("--features-csv")
    parser.add_argument("--labeled-events-csv")
    parser.add_argument("--out-dir")
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    features_path, labels_path = resolve_inputs(args)
    out_dir = Path(args.out_dir) if args.out_dir else features_path.parent / "diagnostics"
    out_dir.mkdir(parents=True, exist_ok=True)

    features = load_features(features_path)
    labels = load_labels(labels_path)
    feature_stats = feature_liveness(features)
    label_stats = label_breadth(labels)
    overlap = overlap_report(features, labels)

    feature_stats.to_csv(out_dir / "feature_stats.csv", index=False)
    label_stats.to_csv(out_dir / "label_stats.csv", index=False)
    write_manifest(
        out_dir / "overlap_report.json",
        {
            "generated_by": command_line(),
            "git_sha": git_sha(),
            "features_csv": str(features_path),
            "labeled_events_csv": str(labels_path),
            **overlap,
        },
    )
    print(
        f"diagnostics written to {out_dir}: "
        f"{len(feature_stats)} features, {len(label_stats)} symbol label rows, "
        f"overlap={overlap['matched_label_rows']}/{overlap['label_rows']}"
    )


def resolve_inputs(args: argparse.Namespace) -> tuple[Path, Path]:
    if args.dataset_dir:
        dataset_dir = Path(args.dataset_dir)
        features = dataset_dir / "features.csv"
        labels = dataset_dir / "labeled_events.csv"
    else:
        if not args.features_csv or not args.labeled_events_csv:
            raise ValueError("provide --dataset-dir or both --features-csv and --labeled-events-csv")
        features = Path(args.features_csv)
        labels = Path(args.labeled_events_csv)
    if not features.exists():
        raise FileNotFoundError(features)
    if not labels.exists():
        raise FileNotFoundError(labels)
    return features, labels


def load_features(path: Path) -> pd.DataFrame:
    df = pd.read_csv(path)
    required = {"symbol", "event_time"}
    if missing := required - set(df.columns):
        raise ValueError(f"features CSV missing required columns: {sorted(missing)}")
    df["symbol"] = df["symbol"].astype(str).str.upper().str.strip()
    df["event_time"] = pd.to_datetime(df["event_time"], utc=True)
    return df


def load_labels(path: Path) -> pd.DataFrame:
    df = pd.read_csv(path)
    required = {"symbol", "event_time", "label"}
    if missing := required - set(df.columns):
        raise ValueError(f"labeled events CSV missing required columns: {sorted(missing)}")
    df["symbol"] = df["symbol"].astype(str).str.upper().str.strip()
    df["event_time"] = pd.to_datetime(df["event_time"], utc=True)
    df["label"] = pd.to_numeric(df["label"], errors="coerce")
    if "sample_weight" in df.columns:
        df["sample_weight"] = pd.to_numeric(df["sample_weight"], errors="coerce")
    return df


def feature_liveness(features: pd.DataFrame) -> pd.DataFrame:
    rows: list[dict[str, Any]] = []
    feature_columns = [column for column in features.columns if column not in KEY_COLUMNS]
    for column in feature_columns:
        values = pd.to_numeric(features[column], errors="coerce").replace([np.inf, -np.inf], np.nan)
        finite = values.dropna()
        zero_count = int((finite == 0).sum())
        nonzero_mask = values.fillna(0.0) != 0
        nonzero_symbols = int(features.loc[nonzero_mask, "symbol"].nunique())
        unique_rounded = int(finite.round(12).nunique()) if len(finite) else 0
        rows.append(
            {
                "feature": column,
                "rows": int(len(values)),
                "finite_rows": int(len(finite)),
                "missing_rate": float(values.isna().mean()) if len(values) else 0.0,
                "zero_rate": float(zero_count / len(finite)) if len(finite) else 0.0,
                "constant": bool(unique_rounded <= 1),
                "unique_rounded_values": unique_rounded,
                "mean": float(finite.mean()) if len(finite) else 0.0,
                "std": float(finite.std(ddof=0)) if len(finite) else 0.0,
                "min": float(finite.min()) if len(finite) else 0.0,
                "max": float(finite.max()) if len(finite) else 0.0,
                "nonzero_symbols": nonzero_symbols,
            }
        )
    return pd.DataFrame(rows).sort_values(["constant", "zero_rate", "feature"], ascending=[False, False, True])


def label_breadth(labels: pd.DataFrame) -> pd.DataFrame:
    rows: list[dict[str, Any]] = []
    for symbol, group in labels.groupby("symbol", sort=True):
        positives = int((group["label"] == 1).sum())
        negatives = int((group["label"] == 0).sum())
        row = {
            "symbol": symbol,
            "events": int(len(group)),
            "positive_labels": positives,
            "negative_labels": negatives,
            "positive_rate": float(positives / len(group)) if len(group) else 0.0,
            "first_event": group["event_time"].min().isoformat() if len(group) else "",
            "last_event": group["event_time"].max().isoformat() if len(group) else "",
            "avg_sample_weight": float(group["sample_weight"].mean()) if "sample_weight" in group.columns else 0.0,
        }
        if "barrier" in group.columns:
            counts = group["barrier"].astype(str).value_counts()
            for barrier, count in counts.items():
                row[f"barrier_{barrier}"] = int(count)
        rows.append(row)
    return pd.DataFrame(rows).sort_values(["events", "symbol"], ascending=[False, True])


def overlap_report(features: pd.DataFrame, labels: pd.DataFrame) -> dict[str, Any]:
    feature_keys = features[["symbol", "event_time"]].copy()
    label_keys = labels[["symbol", "event_time"]].copy()
    merged = label_keys.merge(feature_keys.drop_duplicates(), on=["symbol", "event_time"], how="left", indicator=True)
    matched = int((merged["_merge"] == "both").sum())
    label_rows = int(len(labels))
    duplicate_feature_keys = int(feature_keys.duplicated(["symbol", "event_time"]).sum())
    duplicate_label_keys = int(label_keys.duplicated(["symbol", "event_time"]).sum())
    return {
        "feature_rows": int(len(features)),
        "label_rows": label_rows,
        "matched_label_rows": matched,
        "unmatched_label_rows": int(label_rows - matched),
        "overlap_fraction": float(matched / label_rows) if label_rows else 0.0,
        "feature_symbols": sorted(features["symbol"].dropna().unique().tolist()),
        "label_symbols": sorted(labels["symbol"].dropna().unique().tolist()),
        "duplicate_feature_keys": duplicate_feature_keys,
        "duplicate_label_keys": duplicate_label_keys,
    }


if __name__ == "__main__":
    main()
