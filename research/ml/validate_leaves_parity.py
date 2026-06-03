#!/usr/bin/env python3
"""Validate Python-vs-Go prediction parity for a LightGBM artifact.

Expected workflow:

1. Train with train_meta_label.py. It writes parity_fixture.csv with
   expected_probability from Python LightGBM.
2. Run the Go validator:
   go run ./cmd/validate-leaves-parity \
     --metadata /path/to/metadata.json \
     --fixture /path/to/parity_fixture.csv \
     --out /path/to/go_predictions.csv
3. Run this script to compare expected_probability to go_probability.
"""

from __future__ import annotations

import argparse
import json
from pathlib import Path

import numpy as np
import pandas as pd


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--fixture", required=True)
    parser.add_argument("--go-predictions", required=True)
    parser.add_argument("--metadata")
    parser.add_argument("--tolerance", type=float, default=1e-6)
    parser.add_argument("--mark-promoted", action="store_true")
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    fixture = pd.read_csv(args.fixture)
    go_predictions = pd.read_csv(args.go_predictions)
    if "expected_probability" not in fixture.columns:
        raise ValueError("fixture missing expected_probability column")
    if "go_probability" not in go_predictions.columns:
        raise ValueError("Go predictions missing go_probability column")
    if len(fixture) != len(go_predictions):
        raise ValueError(f"row count mismatch: fixture={len(fixture)} go={len(go_predictions)}")

    errors = np.abs(fixture["expected_probability"].to_numpy() - go_predictions["go_probability"].to_numpy())
    max_abs_error = float(errors.max()) if len(errors) else 0.0
    passed = max_abs_error <= args.tolerance
    print(f"max_abs_error={max_abs_error:.12g} tolerance={args.tolerance:.12g} passed={passed}")
    if args.metadata:
        update_metadata(Path(args.metadata), max_abs_error, passed, args.mark_promoted and passed)
    if not passed:
        raise SystemExit(1)


def update_metadata(path: Path, max_abs_error: float, passed: bool, promote: bool) -> None:
    metadata = json.loads(path.read_text(encoding="utf-8"))
    metadata["leaves_parity_max_abs_error"] = max_abs_error
    metadata["leaves_parity_passed"] = passed
    if promote:
        metadata["status"] = "promoted"
    path.write_text(json.dumps(metadata, indent=2) + "\n", encoding="utf-8")


if __name__ == "__main__":
    main()
