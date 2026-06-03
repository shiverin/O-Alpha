#!/usr/bin/env python3
"""Train the Stage 1 ML meta-label model.

Inputs are deliberately explicit:

- point-in-time feature rows keyed by symbol/event_time;
- either pre-labeled event rows or bars + primary signals for triple-barrier
  labeling;
- YAML feature and label configs that mirror the Go runtime defaults.

The script writes the artifact layout consumed by backend/internal/ml:
model.txt, metadata.json, feature_spec.json, label_config.json,
training_report.json, calibration.json, and parity_fixture.csv.
"""

from __future__ import annotations

import argparse
import itertools
import json
import math
from pathlib import Path
from typing import Any

import lightgbm as lgb
import numpy as np
import pandas as pd
import yaml
from sklearn.metrics import log_loss, roc_auc_score

from artifact_manifest import command_line, file_sha256, git_sha, write_manifest


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--features-csv", required=True)
    parser.add_argument("--labeled-events-csv")
    parser.add_argument("--bars-csv")
    parser.add_argument("--signals-csv")
    parser.add_argument("--feature-spec", default="research/ml/feature_spec.yaml")
    parser.add_argument("--label-config", default="research/ml/label_config.yaml")
    parser.add_argument("--out-dir", required=True)
    parser.add_argument("--model-name", default="ml_meta_label")
    parser.add_argument("--strategy-scope", default="base_strategy")
    parser.add_argument("--valid-fraction", type=float, default=0.20)
    parser.add_argument("--class-balance", choices=["none", "scale_pos_weight"], default="scale_pos_weight")
    parser.add_argument("--validation-mode", choices=["tail", "cpcv", "both"], default="both")
    parser.add_argument("--cpcv-groups", type=int, default=6)
    parser.add_argument("--cpcv-test-groups", type=int, default=2)
    parser.add_argument("--max-cpcv-splits", type=int, default=30)
    parser.add_argument("--embargo-pct", type=float, default=0.01)
    parser.add_argument("--min-train-events", type=int, default=200)
    parser.add_argument("--min-class-count", type=int, default=25)
    parser.add_argument("--min-validation-class-count", type=int, default=10)
    parser.add_argument("--min-data-in-leaf", type=int, default=0, help="0 uses an adaptive value")
    parser.add_argument("--num-leaves", type=int, default=31)
    parser.add_argument("--max-depth", type=int, default=5)
    parser.add_argument("--seed", type=int, default=42)
    parser.add_argument("--export-manifest", default="")
    parser.add_argument("--diagnostics-dir", default="")
    parser.add_argument("--export-command", default="")
    parser.add_argument("--backtest-command", default="")
    parser.add_argument("--benchmark", default="VOO")
    parser.add_argument("--cost-spread-bps", type=float, default=2.0)
    parser.add_argument("--cost-slippage-bps", type=float, default=1.0)
    return parser.parse_args()


def load_yaml(path: str | Path) -> dict[str, Any]:
    with open(path, "r", encoding="utf-8") as handle:
        return yaml.safe_load(handle)


def load_features(path: str | Path, feature_spec: dict[str, Any]) -> pd.DataFrame:
    df = pd.read_csv(path)
    required = {"symbol", "event_time"}
    missing = required - set(df.columns)
    if missing:
        raise ValueError(f"features CSV missing required columns: {sorted(missing)}")
    for name in feature_spec["features"]:
        if name not in df.columns:
            df[name] = 0.0
    df["event_time"] = pd.to_datetime(df["event_time"], utc=True)
    return df


def load_labeled_events(path: str | Path) -> pd.DataFrame:
    df = pd.read_csv(path)
    required = {"symbol", "event_time", "label", "label_end_time"}
    missing = required - set(df.columns)
    if missing:
        raise ValueError(f"labeled events CSV missing required columns: {sorted(missing)}")
    df["event_time"] = pd.to_datetime(df["event_time"], utc=True)
    df["label_end_time"] = pd.to_datetime(df["label_end_time"], utc=True)
    if "sample_weight" not in df.columns:
        df["sample_weight"] = 1.0
    return df


def build_triple_barrier_labels(
    bars_csv: str | Path,
    signals_csv: str | Path,
    cfg: dict[str, Any],
) -> pd.DataFrame:
    bars = pd.read_csv(bars_csv)
    signals = pd.read_csv(signals_csv)
    required_bars = {"symbol", "time", "open", "close"}
    required_signals = {"symbol", "time", "signal"}
    if missing := required_bars - set(bars.columns):
        raise ValueError(f"bars CSV missing required columns: {sorted(missing)}")
    if missing := required_signals - set(signals.columns):
        raise ValueError(f"signals CSV missing required columns: {sorted(missing)}")

    bars["time"] = pd.to_datetime(bars["time"], utc=True)
    signals["time"] = pd.to_datetime(signals["time"], utc=True)
    bars = bars.sort_values(["symbol", "time"])
    signals = signals.sort_values(["symbol", "time"])

    horizon = int(cfg.get("horizon_bars", 5))
    pt_mult = float(cfg.get("profit_take_vol_mult", 1.5))
    sl_mult = float(cfg.get("stop_loss_vol_mult", 1.0))
    vol_lookback = int(cfg.get("vol_lookback", 20))
    spacing = int(cfg.get("min_event_spacing_bars", 1))
    use_next_open = bool(cfg.get("use_next_open_for_entry", False))
    vertical_label = int(cfg.get("vertical_barrier_label", 0))

    rows: list[dict[str, Any]] = []
    for symbol, symbol_bars in bars.groupby("symbol", sort=False):
        symbol_bars = symbol_bars.reset_index(drop=True)
        symbol_signals = signals[signals["symbol"] == symbol].set_index("time")
        closes = symbol_bars["close"].astype(float)
        log_returns = np.log(closes / closes.shift(1))
        realized_vol = log_returns.rolling(vol_lookback).std().fillna(0.0)
        last_event_idx = -spacing - 1

        for i, bar in symbol_bars.iterrows():
            if bar["time"] not in symbol_signals.index:
                continue
            signal = symbol_signals.loc[bar["time"], "signal"]
            if isinstance(signal, pd.Series):
                signal = signal.iloc[-1]
            if not is_buy_signal(signal):
                continue
            if i - last_event_idx < spacing:
                continue
            if i + 1 >= len(symbol_bars):
                continue

            entry_idx = i
            entry_price = float(bar["close"])
            if use_next_open:
                entry_idx = i + 1
                entry_price = float(symbol_bars.loc[entry_idx, "open"])
            if entry_price <= 0:
                continue

            vol = float(realized_vol.iloc[i])
            if vol <= 0:
                continue
            profit_barrier = pt_mult * vol
            stop_barrier = -sl_mult * vol
            end_idx = min(len(symbol_bars) - 1, i + horizon)
            label = vertical_label
            barrier = "vertical"
            path_ret = 0.0

            for j in range(entry_idx + 1, end_idx + 1):
                path_ret = float(symbol_bars.loc[j, "close"]) / entry_price - 1.0
                if path_ret >= profit_barrier:
                    end_idx = j
                    label = 1
                    barrier = "profit_take"
                    break
                if path_ret <= stop_barrier:
                    end_idx = j
                    label = 0
                    barrier = "stop_loss"
                    break

            rows.append(
                {
                    "symbol": symbol,
                    "event_time": bar["time"],
                    "label": label,
                    "label_end_time": symbol_bars.loc[end_idx, "time"],
                    "sample_weight": abs(path_ret),
                    "side": 1,
                    "barrier": barrier,
                    "event_return": path_ret,
                }
            )
            last_event_idx = i

    return pd.DataFrame(rows)


def is_buy_signal(signal: Any) -> bool:
    if isinstance(signal, str):
        return signal.strip().upper() in {"BUY", "LONG", "1"}
    return int(signal) == 1


def merge_features_labels(features: pd.DataFrame, labels: pd.DataFrame) -> pd.DataFrame:
    merged = features.merge(labels, on=["symbol", "event_time"], how="inner")
    if merged.empty:
        raise ValueError("no feature rows matched labeled events")
    merged = merged.sort_values("event_time").reset_index(drop=True)
    return merged


def purged_tail_split(df: pd.DataFrame, valid_fraction: float) -> tuple[pd.DataFrame, pd.DataFrame]:
    valid_fraction = min(max(valid_fraction, 0.05), 0.50)
    split_idx = max(1, int(len(df) * (1.0 - valid_fraction)))
    validation = df.iloc[split_idx:].copy()
    validation_start = validation["event_time"].min()
    train = df.iloc[:split_idx].copy()
    train = train[train["label_end_time"] < validation_start]
    if train.empty or validation.empty:
        raise ValueError("purged split produced empty train or validation set")
    return train, validation


def train_model(
    train: pd.DataFrame,
    validation: pd.DataFrame,
    features: list[str],
    args: argparse.Namespace,
) -> tuple[lgb.Booster, dict[str, Any], np.ndarray]:
    params = training_params(train, args, seed=args.seed)
    dtrain = lgb.Dataset(
        train[features],
        label=train["label"],
        weight=normalized_weights(train),
        feature_name=features,
    )
    dvalid = lgb.Dataset(
        validation[features],
        label=validation["label"],
        weight=normalized_weights(validation),
        feature_name=features,
        reference=dtrain,
    )
    booster = lgb.train(
        params,
        dtrain,
        num_boost_round=2000,
        valid_sets=[dtrain, dvalid],
        valid_names=["train", "validation"],
        callbacks=[lgb.early_stopping(50), lgb.log_evaluation(50)],
    )
    probabilities = booster.predict(validation[features], num_iteration=booster.best_iteration)
    metrics = classification_metrics(validation["label"].to_numpy(), probabilities)
    return booster, metrics | {"training_config": params}, probabilities


def training_params(train: pd.DataFrame, args: argparse.Namespace, seed: int) -> dict[str, Any]:
    min_data_in_leaf = adaptive_min_data_in_leaf(len(train), int(args.min_data_in_leaf))
    params: dict[str, Any] = {
        "objective": "binary",
        "metric": ["auc", "binary_logloss"],
        "num_leaves": max(2, int(args.num_leaves)),
        "max_depth": int(args.max_depth),
        "learning_rate": 0.03,
        "min_data_in_leaf": min_data_in_leaf,
        "min_data_in_bin": max(1, min(15, min_data_in_leaf)),
        "min_sum_hessian_in_leaf": 1e-3,
        "feature_fraction": 0.8,
        "bagging_fraction": 0.8,
        "bagging_freq": 1,
        "lambda_l1": 0.0,
        "lambda_l2": 5.0,
        "force_col_wise": True,
        "verbosity": -1,
        "seed": seed,
    }
    positives = int((train["label"] == 1).sum())
    negatives = int((train["label"] == 0).sum())
    if args.class_balance == "scale_pos_weight" and positives > 0 and negatives > 0:
        params["scale_pos_weight"] = negatives / positives
    return params


def adaptive_min_data_in_leaf(train_rows: int, override: int) -> int:
    if override > 0:
        return override
    if train_rows <= 0:
        return 1
    # Keep this small enough to permit splits on research-sized datasets, but
    # let it grow once the event sample becomes production-sized.
    if train_rows < 200:
        base = max(2, train_rows // 10)
    else:
        base = max(10, train_rows // 50)
    return max(1, min(100, base, max(1, train_rows // 2)))


def normalized_weights(df: pd.DataFrame) -> pd.Series:
    weights = pd.to_numeric(df["sample_weight"], errors="coerce").replace([np.inf, -np.inf], np.nan).fillna(1.0)
    weights = weights.clip(lower=1e-6)
    mean = float(weights.mean())
    if mean <= 0 or not math.isfinite(mean):
        return pd.Series(np.ones(len(df)), index=df.index)
    return weights / mean


def classification_metrics(labels: np.ndarray, probabilities: np.ndarray) -> dict[str, Any]:
    metrics: dict[str, Any] = {"logloss": float(log_loss(labels, probabilities, labels=[0, 1]))}
    if len(set(labels.tolist())) > 1:
        metrics["auc"] = float(roc_auc_score(labels, probabilities))
    else:
        metrics["auc"] = None
    return metrics


def cpcv_report(
    dataset: pd.DataFrame,
    features: list[str],
    args: argparse.Namespace,
) -> dict[str, Any]:
    if args.validation_mode == "tail":
        return {"enabled": False, "reason": "validation_mode_tail"}

    splits = make_cpcv_splits(
        dataset,
        n_groups=args.cpcv_groups,
        n_test_groups=args.cpcv_test_groups,
        embargo_pct=args_embargo_pct(args),
        max_splits=args.max_cpcv_splits,
    )
    rows: list[dict[str, Any]] = []
    for fold, (train_idx, test_idx, test_groups) in enumerate(splits):
        train = dataset.iloc[train_idx].copy()
        test = dataset.iloc[test_idx].copy()
        if train.empty or test.empty:
            continue
        if len(set(train["label"].astype(int).tolist())) < 2 or len(set(test["label"].astype(int).tolist())) < 2:
            rows.append(
                {
                    "fold": fold,
                    "test_groups": test_groups,
                    "num_train": int(len(train)),
                    "num_test": int(len(test)),
                    "skipped": True,
                    "reason": "single_class_train_or_test",
                }
            )
            continue
        params = training_params(train, args, seed=args.seed + fold)
        booster = lgb.train(
            params,
            lgb.Dataset(
                train[features],
                label=train["label"],
                weight=normalized_weights(train),
                feature_name=features,
            ),
            num_boost_round=500,
            valid_sets=[],
            callbacks=[],
        )
        probabilities = booster.predict(test[features])
        metrics = classification_metrics(test["label"].to_numpy(), probabilities)
        rows.append(
            {
                "fold": fold,
                "test_groups": test_groups,
                "num_train": int(len(train)),
                "num_test": int(len(test)),
                "skipped": False,
                "auc": metrics.get("auc"),
                "logloss": metrics.get("logloss"),
                "positive_rate_test": float(test["label"].mean()),
            }
        )

    completed = [row for row in rows if not row.get("skipped")]
    aucs = [row["auc"] for row in completed if row.get("auc") is not None]
    loglosses = [row["logloss"] for row in completed if row.get("logloss") is not None]
    return {
        "enabled": True,
        "n_groups": args.cpcv_groups,
        "n_test_groups": args.cpcv_test_groups,
        "embargo_pct": args_embargo_pct(args),
        "num_splits": len(rows),
        "num_completed": len(completed),
        "auc_mean": float(np.mean(aucs)) if aucs else None,
        "auc_std": float(np.std(aucs)) if aucs else None,
        "logloss_mean": float(np.mean(loglosses)) if loglosses else None,
        "logloss_std": float(np.std(loglosses)) if loglosses else None,
        "folds": rows,
    }


def make_cpcv_splits(
    df: pd.DataFrame,
    n_groups: int,
    n_test_groups: int,
    embargo_pct: float,
    max_splits: int,
) -> list[tuple[np.ndarray, np.ndarray, list[int]]]:
    if "label_end_time" not in df.columns:
        raise ValueError("CPCV requires label_end_time for purging")
    n = len(df)
    n_groups = max(2, min(n_groups, n))
    n_test_groups = max(1, min(n_test_groups, n_groups - 1))
    groups = np.array_split(np.arange(n), n_groups)
    combos = list(itertools.combinations(range(n_groups), n_test_groups))
    if max_splits > 0 and len(combos) > max_splits:
        pick = np.linspace(0, len(combos) - 1, max_splits, dtype=int)
        combos = [combos[i] for i in pick]

    event_times = df["event_time"].reset_index(drop=True)
    label_end_times = df["label_end_time"].reset_index(drop=True)
    embargo_rows = int(math.ceil(n * max(0.0, embargo_pct)))
    splits: list[tuple[np.ndarray, np.ndarray, list[int]]] = []
    for combo in combos:
        test_idx = np.concatenate([groups[i] for i in combo])
        train_mask = np.ones(n, dtype=bool)
        train_mask[test_idx] = False
        for group_id in combo:
            group_idx = groups[group_id]
            test_start = event_times.iloc[group_idx[0]]
            test_end = label_end_times.iloc[group_idx].max()
            overlap = (event_times <= test_end) & (label_end_times >= test_start)
            train_mask[overlap.to_numpy()] = False
            embargo_start = int(group_idx[-1]) + 1
            embargo_end = min(n, embargo_start + embargo_rows)
            if embargo_start < embargo_end:
                train_mask[embargo_start:embargo_end] = False
        train_idx = np.flatnonzero(train_mask)
        splits.append((train_idx, np.sort(test_idx), list(combo)))
    return splits


def args_embargo_pct(args: argparse.Namespace) -> float:
    return float(getattr(args, "embargo_pct", 0.01))


def write_artifacts(
    out_dir: Path,
    args: argparse.Namespace,
    feature_spec: dict[str, Any],
    label_config: dict[str, Any],
    booster: lgb.Booster,
    train: pd.DataFrame,
    validation: pd.DataFrame,
    metrics: dict[str, Any],
    probabilities: np.ndarray,
    cpcv: dict[str, Any],
) -> None:
    out_dir.mkdir(parents=True, exist_ok=True)
    booster.save_model(str(out_dir / "model.txt"))
    write_json(out_dir / "feature_spec.json", feature_spec)
    write_json(out_dir / "label_config.json", label_config)
    write_json(out_dir / "calibration.json", {"method": "none"})

    training_config = metrics["training_config"]
    model_diagnostics = model_structure_diagnostics(booster)
    probability_diagnostics = validation_probability_diagnostics(probabilities)
    threshold_selection = select_operating_threshold(validation["label"].to_numpy(), probabilities)
    warnings = training_warnings(train, validation, training_config, model_diagnostics, probability_diagnostics, cpcv, args)
    status, status_reason = artifact_status(warnings, model_diagnostics, cpcv, train, validation, threshold_selection, args)
    manifest = artifact_manifest(args, feature_spec, label_config, train, validation, cpcv, status)
    thresholds = {
        "enter_long": threshold_selection["threshold"],
        "reduce": max(0.0, threshold_selection["threshold"] - 0.05),
        "bet_sizing_slope": 2.0,
        "pass_through_exits": True,
    }
    report = {
        "model_name": args.model_name,
        "strategy_scope": args.strategy_scope,
        "num_train": int(len(train)),
        "num_validation": int(len(validation)),
        "auc": metrics.get("auc"),
        "logloss": metrics.get("logloss"),
        "best_iteration": int(booster.best_iteration or booster.current_iteration()),
        "feature_importance": dict(zip(feature_spec["features"], booster.feature_importance().tolist())),
        "model_diagnostics": model_diagnostics,
        "probability_diagnostics": probability_diagnostics,
        "threshold_selection": threshold_selection,
        "thresholds": thresholds,
        "warnings": warnings,
        "status": status,
        "status_reason": status_reason,
        "cpcv": cpcv,
    }
    write_json(out_dir / "training_report.json", report)
    write_manifest(out_dir / "manifest.json", manifest)

    fixture = validation[["symbol", "event_time", *feature_spec["features"]]].copy()
    fixture["expected_probability"] = probabilities
    fixture.head(500).to_csv(out_dir / "parity_fixture.csv", index=False)

    metadata = {
        "model_name": args.model_name,
        "model_type": "lightgbm_binary",
        "strategy_scope": args.strategy_scope,
        "artifact_uri": ".",
        "feature_spec": feature_spec,
        "label_config": label_config,
        "training_config": training_config,
        "train_start": to_iso(train["event_time"].min()),
        "train_end": to_iso(train["event_time"].max()),
        "validation_start": to_iso(validation["event_time"].min()),
        "validation_end": to_iso(validation["event_time"].max()),
        "auc": metrics.get("auc"),
        "logloss": metrics.get("logloss"),
        "cpcv_auc_mean": cpcv.get("auc_mean"),
        "cpcv_logloss_mean": cpcv.get("logloss_mean"),
        "model_diagnostics": model_diagnostics,
        "probability_diagnostics": probability_diagnostics,
        "threshold_selection": threshold_selection,
        "thresholds": thresholds,
        "manifest": manifest,
        "leaves_parity_passed": False,
        "status": status,
        "status_reason": status_reason,
        "warnings": warnings,
        "created_at": pd.Timestamp.utcnow().isoformat(),
    }
    write_json(out_dir / "metadata.json", metadata)


def artifact_manifest(
    args: argparse.Namespace,
    feature_spec: dict[str, Any],
    label_config: dict[str, Any],
    train: pd.DataFrame,
    validation: pd.DataFrame,
    cpcv: dict[str, Any],
    status: str,
) -> dict[str, Any]:
    symbols = sorted(set(train["symbol"].astype(str).str.upper()) | set(validation["symbol"].astype(str).str.upper()))
    export_manifest = load_json_optional(args.export_manifest)
    export_command = args.export_command or str(export_manifest.get("export_command", ""))
    return {
        "artifact_id": f"{args.model_name}_{pd.Timestamp.utcnow().isoformat()}",
        "git_sha": git_sha(),
        "export_command": export_command,
        "train_command": command_line(),
        "backtest_command": args.backtest_command,
        "symbols": symbols,
        "context_symbols": feature_spec.get("context_symbols", []),
        "benchmark": args.benchmark,
        "export_manifest": args.export_manifest,
        "export_manifest_sha256": file_sha256(args.export_manifest),
        "export_dataset": export_manifest.get("dataset", {}),
        "export_symbol_count": export_manifest.get("training_symbols"),
        "feature_spec_sha256": file_sha256(args.feature_spec),
        "label_config_sha256": file_sha256(args.label_config),
        "diagnostics_dir": args.diagnostics_dir,
        "data_snapshot": {
            "train_start": to_iso(train["event_time"].min()),
            "train_end": to_iso(train["event_time"].max()),
            "validation_start": to_iso(validation["event_time"].min()),
            "validation_end": to_iso(validation["event_time"].max()),
            "features_csv": args.features_csv,
            "labeled_events_csv": args.labeled_events_csv,
            "bars_csv": args.bars_csv,
            "signals_csv": args.signals_csv,
        },
        "cost_model": {
            "spread_bps": args.cost_spread_bps,
            "slippage_bps": args.cost_slippage_bps,
        },
        "folds": cpcv.get("folds", []),
        "artifact_status": status,
        "feature_count": len(feature_spec.get("features", [])),
        "label_horizon_bars": label_config.get("horizon_bars"),
    }


def load_json_optional(path: str | Path | None) -> dict[str, Any]:
    if not path:
        return {}
    p = Path(path)
    if not p.exists():
        return {}
    with p.open("r", encoding="utf-8") as handle:
        value = json.load(handle)
    if isinstance(value, dict):
        return value
    return {}


def model_structure_diagnostics(booster: lgb.Booster) -> dict[str, Any]:
    model = booster.dump_model()
    trees = model.get("tree_info", [])
    leaf_counts = [int(tree.get("num_leaves", 0)) for tree in trees]
    split_count = int(sum(max(0, count - 1) for count in leaf_counts))
    importances = booster.feature_importance()
    return {
        "num_trees": len(trees),
        "leaf_counts": leaf_counts,
        "total_leaves": int(sum(leaf_counts)),
        "split_count": split_count,
        "nonzero_feature_importance_count": int((importances > 0).sum()),
        "single_leaf_model": bool(split_count == 0),
    }


def validation_probability_diagnostics(probabilities: np.ndarray) -> dict[str, Any]:
    if len(probabilities) == 0:
        return {
            "count": 0,
            "unique_rounded_probability_count": 0,
            "constant_predictions": True,
        }
    rounded = np.round(probabilities.astype(float), 12)
    return {
        "count": int(len(probabilities)),
        "min": float(np.min(probabilities)),
        "max": float(np.max(probabilities)),
        "mean": float(np.mean(probabilities)),
        "std": float(np.std(probabilities)),
        "unique_rounded_probability_count": int(len(set(rounded.tolist()))),
        "constant_predictions": bool(len(set(rounded.tolist())) <= 1),
    }


def select_operating_threshold(labels: np.ndarray, probabilities: np.ndarray) -> dict[str, Any]:
    labels = labels.astype(int)
    probabilities = probabilities.astype(float)
    if len(labels) == 0 or len(probabilities) == 0:
        return {
            "method": "validation_precision_lift",
            "threshold": 0.5,
            "base_positive_rate": None,
            "accepted_count": 0,
            "coverage": 0.0,
            "precision": None,
            "recall": None,
            "f1": None,
            "precision_lift": None,
            "reason": "empty_validation",
        }

    base_rate = float(labels.mean())
    positives = int(labels.sum())
    min_accept = max(5, int(math.ceil(len(labels) * 0.05)))
    best: dict[str, Any] | None = None
    for threshold in sorted(set(np.round(probabilities, 12).tolist())):
        accepted = probabilities >= threshold
        accepted_count = int(accepted.sum())
        if accepted_count < min_accept:
            continue
        tp = int(labels[accepted].sum())
        precision = float(tp / accepted_count) if accepted_count else 0.0
        recall = float(tp / positives) if positives else 0.0
        f1 = float((2 * precision * recall) / (precision + recall)) if precision + recall > 0 else 0.0
        coverage = float(accepted_count / len(labels))
        precision_lift = precision - base_rate
        if precision_lift <= 0:
            continue
        objective = precision_lift * math.sqrt(max(coverage, 1e-12))
        row = {
            "method": "validation_precision_lift",
            "threshold": float(threshold),
            "base_positive_rate": base_rate,
            "accepted_count": accepted_count,
            "coverage": coverage,
            "precision": precision,
            "recall": recall,
            "f1": f1,
            "precision_lift": precision_lift,
            "objective": objective,
            "reason": "best_precision_lift_with_min_coverage",
        }
        if best is None or (row["objective"], row["precision"], row["coverage"]) > (
            best["objective"],
            best["precision"],
            best["coverage"],
        ):
            best = row
    if best is not None:
        return best

    return {
        "method": "validation_precision_lift",
        "threshold": 0.5,
        "base_positive_rate": base_rate,
        "accepted_count": 0,
        "coverage": 0.0,
        "precision": None,
        "recall": None,
        "f1": None,
        "precision_lift": None,
        "reason": "no_threshold_beats_validation_base_rate",
    }


def training_warnings(
    train: pd.DataFrame,
    validation: pd.DataFrame,
    training_config: dict[str, Any],
    model_diagnostics: dict[str, Any],
    probability_diagnostics: dict[str, Any],
    cpcv: dict[str, Any],
    args: argparse.Namespace,
) -> list[str]:
    warnings: list[str] = []
    train_pos = int((train["label"] == 1).sum())
    train_neg = int((train["label"] == 0).sum())
    valid_pos = int((validation["label"] == 1).sum())
    valid_neg = int((validation["label"] == 0).sum())
    min_data = int(training_config.get("min_data_in_leaf", 0) or 0)
    if len(train) < args.min_train_events:
        warnings.append(f"train event count {len(train)} below research minimum {args.min_train_events}")
    if train_pos < args.min_class_count or train_neg < args.min_class_count:
        warnings.append(f"train class counts too small: positive={train_pos} negative={train_neg}")
    if valid_pos < args.min_validation_class_count or valid_neg < args.min_validation_class_count:
        warnings.append(f"validation class counts too small: positive={valid_pos} negative={valid_neg}")
    if min_data > 0 and len(train) <= min_data:
        warnings.append(f"min_data_in_leaf {min_data} is >= train events {len(train)}, preventing splits")
    if model_diagnostics.get("single_leaf_model"):
        warnings.append("LightGBM produced a single-leaf constant model")
    if probability_diagnostics.get("constant_predictions"):
        warnings.append("validation probabilities are constant")
    cpcv_auc = cpcv.get("auc_mean")
    if cpcv.get("enabled") and cpcv_auc is not None and float(cpcv_auc) < 0.5:
        warnings.append(f"CPCV AUC mean {float(cpcv_auc):.4f} is below 0.5; do not promote without broader validation")
    return warnings


def artifact_status(
    warnings: list[str],
    model_diagnostics: dict[str, Any],
    cpcv: dict[str, Any],
    train: pd.DataFrame,
    validation: pd.DataFrame,
    threshold_selection: dict[str, Any],
    args: argparse.Namespace,
) -> tuple[str, str]:
    train_pos = int((train["label"] == 1).sum())
    train_neg = int((train["label"] == 0).sum())
    valid_pos = int((validation["label"] == 1).sum())
    valid_neg = int((validation["label"] == 0).sum())
    if len(train) < args.min_train_events:
        return "rejected", "insufficient_train_events"
    if train_pos < args.min_class_count or train_neg < args.min_class_count:
        return "rejected", "insufficient_train_class_balance"
    if valid_pos < args.min_validation_class_count or valid_neg < args.min_validation_class_count:
        return "rejected", "insufficient_validation_class_balance"
    if model_diagnostics.get("single_leaf_model"):
        return "rejected", "constant_single_leaf_model"
    if cpcv.get("enabled") and int(cpcv.get("num_completed") or 0) == 0:
        return "rejected", "no_completed_cpcv_folds"
    cpcv_auc = cpcv.get("auc_mean")
    if cpcv.get("enabled") and cpcv_auc is not None and float(cpcv_auc) < 0.5:
        return "rejected", "cpcv_auc_below_random"
    if threshold_selection.get("reason") == "no_threshold_beats_validation_base_rate":
        return "rejected", "no_profitable_operating_threshold"
    if any("constant" in warning.lower() for warning in warnings):
        return "rejected", "constant_predictions"
    return "candidate", "passes_structural_training_gates"


def write_json(path: Path, value: dict[str, Any]) -> None:
    with open(path, "w", encoding="utf-8") as handle:
        json.dump(value, handle, indent=2, default=to_iso)
        handle.write("\n")


def to_iso(value: Any) -> Any:
    if value is None or pd.isna(value):
        return None
    if isinstance(value, np.generic):
        return value.item()
    return pd.Timestamp(value).isoformat()


def main() -> None:
    args = parse_args()
    feature_spec = load_yaml(args.feature_spec)
    label_config = load_yaml(args.label_config)
    features = load_features(args.features_csv, feature_spec)

    if args.labeled_events_csv:
        labels = load_labeled_events(args.labeled_events_csv)
    elif args.bars_csv and args.signals_csv:
        labels = build_triple_barrier_labels(args.bars_csv, args.signals_csv, label_config)
    else:
        raise ValueError("provide --labeled-events-csv or both --bars-csv and --signals-csv")

    dataset = merge_features_labels(features, labels)
    cpcv = cpcv_report(dataset, feature_spec["features"], args)
    train, validation = purged_tail_split(dataset, args.valid_fraction)
    booster, metrics, probabilities = train_model(train, validation, feature_spec["features"], args)
    write_artifacts(Path(args.out_dir), args, feature_spec, label_config, booster, train, validation, metrics, probabilities, cpcv)


if __name__ == "__main__":
    main()
