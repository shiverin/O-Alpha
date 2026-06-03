#!/usr/bin/env python3
"""Build ML meta-label training features and triple-barrier labels.

The output is intentionally compatible with train_meta_label.py:

- features.csv: symbol,event_time plus the configured feature columns;
- labeled_events.csv: symbol,event_time,label,label_end_time,sample_weight.

Inputs must be point-in-time historical bars and deterministic base-strategy
signals. Optional HMM posterior rows can be supplied; otherwise a small
Gaussian-HMM estimator derives low/medium/high volatility state probabilities
from each symbol's historical returns.

By default labels are generated only for long entry signals. That matches the
current Go MLMetaLabelStrategy runtime, which filters BUY entries and passes
SELL signals through as exits.
"""

from __future__ import annotations

import argparse
import json
import math
from pathlib import Path
from typing import Any

import numpy as np
import pandas as pd
import yaml

from artifact_manifest import command_line, file_sha256, git_sha, write_manifest

TRADING_DAYS_PER_YEAR = 252.0


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--bars-csv", required=True)
    parser.add_argument("--signals-csv", required=True)
    parser.add_argument("--hmm-posteriors-csv")
    parser.add_argument("--feature-spec", default="research/ml/feature_spec.yaml")
    parser.add_argument("--label-config", default="research/ml/label_config.yaml")
    parser.add_argument("--out-dir", required=True)
    parser.add_argument("--hmm-states", type=int, default=3)
    parser.add_argument("--hmm-iterations", type=int, default=50)
    parser.add_argument(
        "--estimate-hmm-posteriors",
        action="store_true",
        help="estimate research HMM posteriors when no posterior CSV is supplied; off by default to match Go runtime zeros",
    )
    parser.add_argument("--fracdiff-d", type=float, default=0.5)
    parser.add_argument("--fracdiff-width", type=int, default=64)
    parser.add_argument(
        "--label-sides",
        default=None,
        choices=["long", "short", "long_short"],
        help="which signal sides become meta-label events; default comes from label config or long",
    )
    return parser.parse_args()


def load_yaml(path: str | Path) -> dict[str, Any]:
    with open(path, "r", encoding="utf-8") as handle:
        return yaml.safe_load(handle)


def load_bars(path: str | Path) -> pd.DataFrame:
    bars = pd.read_csv(path)
    required = {"symbol", "time", "open", "high", "low", "close", "volume"}
    if missing := required - set(bars.columns):
        raise ValueError(f"bars CSV missing required columns: {sorted(missing)}")
    bars["symbol"] = bars["symbol"].str.upper().str.strip()
    bars["time"] = pd.to_datetime(bars["time"], utc=True)
    bars = bars.sort_values(["symbol", "time"]).reset_index(drop=True)
    for column in ["open", "high", "low", "close", "volume"]:
        bars[column] = pd.to_numeric(bars[column], errors="coerce").fillna(0.0)
    return bars


def load_signals(path: str | Path) -> pd.DataFrame:
    signals = pd.read_csv(path)
    required = {"symbol", "time", "signal"}
    if missing := required - set(signals.columns):
        raise ValueError(f"signals CSV missing required columns: {sorted(missing)}")
    signals["symbol"] = signals["symbol"].str.upper().str.strip()
    signals["time"] = pd.to_datetime(signals["time"], utc=True)
    signals["side"] = signals["signal"].map(signal_side).fillna(0).astype(int)
    return signals.sort_values(["symbol", "time"]).reset_index(drop=True)


def signal_side(value: Any) -> int:
    if isinstance(value, str):
        normalized = value.strip().upper()
        if normalized in {"BUY", "LONG", "1"}:
            return 1
        if normalized in {"SELL", "SHORT", "-1"}:
            return -1
        return 0
    try:
        numeric = int(value)
    except Exception:
        return 0
    if numeric > 0:
        return 1
    if numeric < 0:
        return -1
    return 0


def build_features(
    bars: pd.DataFrame,
    signals: pd.DataFrame,
    feature_names: list[str],
    hmm_posteriors: pd.DataFrame | None,
    estimate_hmm_posteriors_flag: bool,
    hmm_states: int,
    hmm_iterations: int,
    fracdiff_d: float,
    fracdiff_width: int,
) -> pd.DataFrame:
    merged = bars.merge(signals, on=["symbol", "time"], how="left", suffixes=("", "_signal"))
    merged["side"] = merged["side"].fillna(0).astype(int)
    if "alpha_score" not in merged.columns:
        merged["alpha_score"] = merged["side"].astype(float)
    if "confidence" not in merged.columns:
        merged["confidence"] = merged["side"].abs().astype(float)

    hmm = hmm_posteriors
    if hmm is None and estimate_hmm_posteriors_flag:
        hmm = estimate_hmm_posteriors(bars, hmm_states, hmm_iterations)
    if hmm is not None:
        merged = merged.merge(hmm, on=["symbol", "time"], how="left")
    for column in regime_columns():
        if column not in merged.columns:
            merged[column] = 0.0
        merged[column] = merged[column].fillna(0.0)

    context = build_context_tables(bars)
    frames: list[pd.DataFrame] = []
    for _, group in merged.groupby("symbol", sort=False):
        group = group.sort_values("time").reset_index(drop=True)
        frames.append(features_for_symbol(group, context, feature_names, fracdiff_d, fracdiff_width))
    features = pd.concat(frames, ignore_index=True)
    features = features.rename(columns={"time": "event_time"})
    return features[["symbol", "event_time", *feature_names]]


def features_for_symbol(
    group: pd.DataFrame,
    context: dict[str, pd.DataFrame],
    feature_names: list[str],
    fracdiff_d: float,
    fracdiff_width: int,
) -> pd.DataFrame:
    close = group["close"].astype(float)
    open_ = group["open"].astype(float)
    high = group["high"].astype(float)
    low = group["low"].astype(float)
    volume = group["volume"].astype(float)
    log_close = np.log(close.where(close > 0))
    log_ret = log_close.diff()

    out = group[["symbol", "time"]].copy()
    rolling = {
        "ret_1": log_close.diff(1),
        "ret_2": log_close.diff(2),
        "ret_5": log_close.diff(5),
        "ret_10": log_close.diff(10),
        "ret_21": log_close.diff(21),
        "ret_63": log_close.diff(63),
    }
    vwap_proxy = (close * volume).rolling(20, min_periods=1).sum() / volume.rolling(20, min_periods=1).sum()
    atr = true_range(high, low, close).rolling(14, min_periods=1).mean() / close
    fast_ma = close.rolling(20).mean()
    slow_ma = close.rolling(50).mean()

    base_values = {
        "fracdiff_close_d0_5": fractional_difference(close, fracdiff_d, fracdiff_width),
        "fracdiff_log_close_d0_5": fractional_difference(log_close, fracdiff_d, fracdiff_width),
        "log_ret_1": rolling["ret_1"],
        "log_ret_2": rolling["ret_2"],
        "log_ret_5": rolling["ret_5"],
        "log_ret_10": rolling["ret_10"],
        "log_ret_21": rolling["ret_21"],
        "rolling_ret_21": rolling["ret_21"],
        "rolling_ret_63": rolling["ret_63"],
        "distance_to_20d_high": close / high.rolling(20, min_periods=1).max() - 1.0,
        "distance_to_20d_low": close / low.where(low > 0).rolling(20, min_periods=1).min() - 1.0,
        "close_to_vwap_proxy": close / vwap_proxy - 1.0,
        "close_to_close_vol_20": rolling_std_population(log_ret, 20) * math.sqrt(TRADING_DAYS_PER_YEAR),
        "ewma_vol_20": ewma_vol_go(log_ret, 20),
        "parkinson_vol_20": parkinson_vol(high, low, 20),
        "garman_klass_vol_20": garman_klass_vol(open_, high, low, close, 20),
        "rogers_satchell_vol_20": rogers_satchell_vol(open_, high, low, close, 20),
        "atr_pct_14": atr,
        "ma_fast_minus_slow_pct": (fast_ma - slow_ma) / slow_ma,
        "ma_fast_slope": fast_ma / fast_ma.shift(1) - 1.0,
        "ma_slow_slope": slow_ma / slow_ma.shift(1) - 1.0,
        "kalman_residual": optional_column(group, "kalman_residual", "residual"),
        "kalman_zscore": optional_column(group, "kalman_zscore", "zscore", "z_score"),
        "ensemble_score": optional_column(group, "ensemble_score", "alpha_score", "signal_score"),
        "ensemble_confidence": optional_column(group, "ensemble_confidence", "confidence"),
        "hmm_regime_probability_low": group["hmm_regime_probability_low"],
        "hmm_regime_probability_medium": group["hmm_regime_probability_medium"],
        "hmm_regime_probability_high": group["hmm_regime_probability_high"],
        "volume_z_20": zscore(volume, 20),
        "dollar_volume_z_20": zscore(close * volume, 20),
        "order_book_imbalance": optional_column(group, "order_book_imbalance", "book_imbalance", "l2_imbalance"),
        "signed_volume_imbalance_20": signed_volume_imbalance(open_, close, volume, 20),
        "amihud_illiquidity_20": (log_ret.abs() / (close * volume).replace(0, np.nan)).rolling(20, min_periods=1).mean(),
        "high_low_spread_proxy": (high - low) / ((high + low) / 2.0),
        "turnover_proxy": volume / volume.rolling(20, min_periods=1).mean() - 1.0,
        "gap_pct": open_ / close.shift(1) - 1.0,
    }

    for symbol in ["SPY", "QQQ", "IWM"]:
        ctx = context.get(symbol)
        if ctx is not None:
            aligned = align_context(group["time"], ctx)
            base_values[f"{symbol.lower()}_ret_1"] = np.log(aligned["close"] / aligned["close"].shift(1))
            base_values[f"{symbol.lower()}_ret_5"] = np.log(aligned["close"] / aligned["close"].shift(5))
            if symbol == "SPY":
                spy_ret = np.log(aligned["close"] / aligned["close"].shift(1))
                base_values["spy_vol_20"] = rolling_std_population(spy_ret, 20) * math.sqrt(TRADING_DAYS_PER_YEAR)
                base_values["relative_strength_vs_spy_21"] = rolling["ret_21"] - np.log(aligned["close"] / aligned["close"].shift(21))
                beta = rolling_beta(log_ret, spy_ret, 63)
                base_values["beta_to_spy_63"] = beta
                base_values["residual_ret_vs_spy_21"] = rolling["ret_21"] - beta * np.log(aligned["close"] / aligned["close"].shift(21))
        else:
            base_values[f"{symbol.lower()}_ret_1"] = 0.0
            base_values[f"{symbol.lower()}_ret_5"] = 0.0
            if symbol == "SPY":
                base_values["spy_vol_20"] = 0.0
                base_values["relative_strength_vs_spy_21"] = 0.0
                base_values["beta_to_spy_63"] = 0.0
                base_values["residual_ret_vs_spy_21"] = 0.0

    sector_symbol = str(group.get("sector_symbol", pd.Series([""])).iloc[0] or "").upper()
    sector = context.get(sector_symbol)
    if sector is not None:
        aligned = align_context(group["time"], sector)
        base_values["sector_etf_ret_1"] = np.log(aligned["close"] / aligned["close"].shift(1))
        base_values["sector_etf_ret_5"] = np.log(aligned["close"] / aligned["close"].shift(5))
    else:
        base_values["sector_etf_ret_1"] = 0.0
        base_values["sector_etf_ret_5"] = 0.0

    for name in feature_names:
        value = base_values.get(name, 0.0)
        out[name] = clean_series(value, len(out))
    return out


def build_labels(
    bars: pd.DataFrame,
    signals: pd.DataFrame,
    label_cfg: dict[str, Any],
    allowed_sides: set[int],
) -> pd.DataFrame:
    merged = bars.merge(signals[["symbol", "time", "side"]], on=["symbol", "time"], how="left")
    merged["side"] = merged["side"].fillna(0).astype(int)
    horizon = int(label_cfg.get("horizon_bars", 5))
    pt_mult = float(label_cfg.get("profit_take_vol_mult", 1.5))
    sl_mult = float(label_cfg.get("stop_loss_vol_mult", 1.0))
    vol_lookback = int(label_cfg.get("vol_lookback", 20))
    spacing = int(label_cfg.get("min_event_spacing_bars", 1))
    use_next_open = bool(label_cfg.get("use_next_open_for_entry", False))
    vertical_label = int(label_cfg.get("vertical_barrier_label", 0))

    rows: list[dict[str, Any]] = []
    for symbol, group in merged.groupby("symbol", sort=False):
        group = group.sort_values("time").reset_index(drop=True)
        close = group["close"].astype(float)
        realized_vol = np.log(close / close.shift(1)).rolling(vol_lookback).std().fillna(0.0)
        last_event_idx = -spacing - 1
        for i, row in group.iterrows():
            side = int(row["side"])
            if side == 0 or side not in allowed_sides or i - last_event_idx < spacing or i + 1 >= len(group):
                continue
            entry_idx = i + 1 if use_next_open else i
            entry_price = float(group.loc[entry_idx, "open" if use_next_open else "close"])
            if entry_price <= 0:
                continue
            vol = float(realized_vol.iloc[i])
            if vol <= 0:
                continue
            end_idx = min(len(group) - 1, i + horizon)
            label = vertical_label
            barrier = "vertical"
            path_ret = side * (float(group.loc[end_idx, "close"]) / entry_price - 1.0)
            for j in range(entry_idx + 1, end_idx + 1):
                path_ret = side * (float(group.loc[j, "close"]) / entry_price - 1.0)
                if path_ret >= pt_mult * vol:
                    end_idx = j
                    label = 1
                    barrier = "profit_take"
                    break
                if path_ret <= -sl_mult * vol:
                    end_idx = j
                    label = 0
                    barrier = "stop_loss"
                    break
            rows.append(
                {
                    "symbol": symbol,
                    "event_time": row["time"],
                    "label": label,
                    "label_end_time": group.loc[end_idx, "time"],
                    "sample_weight": abs(path_ret),
                    "side": side,
                    "barrier": barrier,
                    "event_return": path_ret,
                }
            )
            last_event_idx = i
    columns = ["symbol", "event_time", "label", "label_end_time", "sample_weight", "side", "barrier", "event_return"]
    return pd.DataFrame(rows, columns=columns)


def resolve_label_sides(args: argparse.Namespace, label_cfg: dict[str, Any]) -> set[int]:
    value = args.label_sides
    if value is None:
        value = str(label_cfg.get("label_sides", "long"))
    value = str(value).strip().lower()
    if value == "long":
        return {1}
    if value == "short":
        return {-1}
    if value == "long_short":
        return {-1, 1}
    raise ValueError(f"unsupported label_sides={value!r}")


def estimate_hmm_posteriors(bars: pd.DataFrame, n_states: int, n_iter: int) -> pd.DataFrame:
    n_states = max(2, min(n_states, 5))
    frames: list[pd.DataFrame] = []
    for symbol, group in bars.groupby("symbol", sort=False):
        group = group.sort_values("time").reset_index(drop=True)
        returns = np.log(group["close"].astype(float) / group["close"].astype(float).shift(1)).fillna(0.0).to_numpy()
        probs, variances = gaussian_hmm_posteriors(returns, n_states, n_iter)
        order = np.argsort(variances)
        low = probs[:, order[0]]
        high = probs[:, order[-1]]
        if n_states >= 3:
            medium = probs[:, order[1:-1]].sum(axis=1)
        else:
            medium = 1.0 - low - high
        frames.append(
            pd.DataFrame(
                {
                    "symbol": symbol,
                    "time": group["time"],
                    "hmm_regime_probability_low": low,
                    "hmm_regime_probability_medium": medium,
                    "hmm_regime_probability_high": high,
                }
            )
        )
    return pd.concat(frames, ignore_index=True)


def gaussian_hmm_posteriors(values: np.ndarray, n_states: int, n_iter: int) -> tuple[np.ndarray, np.ndarray]:
    x = np.nan_to_num(values.astype(float), nan=0.0, posinf=0.0, neginf=0.0)
    n = len(x)
    if n == 0:
        return np.zeros((0, n_states)), np.ones(n_states)
    quantiles = np.linspace(0.1, 0.9, n_states)
    means = np.quantile(x, quantiles)
    variances = np.full(n_states, max(np.var(x), 1e-8))
    transition = np.full((n_states, n_states), 0.05 / max(1, n_states - 1))
    np.fill_diagonal(transition, 0.95)
    initial = np.full(n_states, 1.0 / n_states)

    gamma = np.full((n, n_states), 1.0 / n_states)
    xi_sum = np.zeros_like(transition)
    for _ in range(max(1, n_iter)):
        emissions = gaussian_pdf(x[:, None], means[None, :], variances[None, :])
        alpha = np.zeros((n, n_states))
        scales = np.zeros(n)
        alpha[0] = initial * emissions[0]
        scales[0] = max(alpha[0].sum(), 1e-300)
        alpha[0] /= scales[0]
        for t in range(1, n):
            alpha[t] = emissions[t] * (alpha[t - 1] @ transition)
            scales[t] = max(alpha[t].sum(), 1e-300)
            alpha[t] /= scales[t]

        beta = np.ones((n, n_states))
        for t in range(n - 2, -1, -1):
            beta[t] = transition @ (emissions[t + 1] * beta[t + 1])
            beta[t] /= max(beta[t].sum(), 1e-300)

        gamma = alpha * beta
        gamma /= np.maximum(gamma.sum(axis=1, keepdims=True), 1e-300)
        xi_sum.fill(0.0)
        for t in range(n - 1):
            xi = alpha[t, :, None] * transition * emissions[t + 1, None, :] * beta[t + 1, None, :]
            xi /= max(xi.sum(), 1e-300)
            xi_sum += xi

        initial = gamma[0]
        transition = xi_sum / np.maximum(xi_sum.sum(axis=1, keepdims=True), 1e-300)
        weights = np.maximum(gamma.sum(axis=0), 1e-300)
        means = (gamma * x[:, None]).sum(axis=0) / weights
        variances = (gamma * (x[:, None] - means[None, :]) ** 2).sum(axis=0) / weights
        variances = np.maximum(variances, 1e-10)

    return gamma, variances


def gaussian_pdf(x: np.ndarray, mean: np.ndarray, variance: np.ndarray) -> np.ndarray:
    return np.exp(-0.5 * ((x - mean) ** 2) / variance) / np.sqrt(2 * np.pi * variance)


def load_hmm_posteriors(path: str | Path | None) -> pd.DataFrame | None:
    if not path:
        return None
    posteriors = pd.read_csv(path)
    required = {"symbol", "time", *regime_columns()}
    if missing := required - set(posteriors.columns):
        raise ValueError(f"HMM posterior CSV missing required columns: {sorted(missing)}")
    posteriors["symbol"] = posteriors["symbol"].str.upper().str.strip()
    posteriors["time"] = pd.to_datetime(posteriors["time"], utc=True)
    return posteriors[["symbol", "time", *regime_columns()]]


def regime_columns() -> list[str]:
    return [
        "hmm_regime_probability_low",
        "hmm_regime_probability_medium",
        "hmm_regime_probability_high",
    ]


def build_context_tables(bars: pd.DataFrame) -> dict[str, pd.DataFrame]:
    return {symbol: group.sort_values("time").copy() for symbol, group in bars.groupby("symbol", sort=False)}


def align_context(times: pd.Series, ctx: pd.DataFrame) -> pd.DataFrame:
    left = pd.DataFrame({"time": times, "_order": np.arange(len(times))}).sort_values("time")
    right = ctx[["time", "close"]].sort_values("time")
    return pd.merge_asof(left, right, on="time", direction="backward").sort_values("_order").reset_index(drop=True)


def fractional_difference(series: pd.Series, d: float, width: int) -> pd.Series:
    weights = fractional_weights(d, width)
    values = series.astype(float).to_numpy()
    out = np.zeros(len(values))
    for i in range(len(values)):
        n = min(i + 1, len(weights))
        window = values[i - n + 1 : i + 1][::-1]
        out[i] = float(np.nansum(weights[:n] * window))
    return pd.Series(out, index=series.index)


def fractional_weights(d: float, width: int) -> np.ndarray:
    width = max(1, width)
    weights = np.ones(width)
    for k in range(1, width):
        weights[k] = -weights[k - 1] * (d - k + 1) / k
    return weights


def true_range(high: pd.Series, low: pd.Series, close: pd.Series) -> pd.Series:
    prev_close = close.shift(1)
    return pd.concat([(high - low), (high - prev_close).abs(), (low - prev_close).abs()], axis=1).max(axis=1)


def parkinson_vol(high: pd.Series, low: pd.Series, window: int) -> pd.Series:
    value = np.log(high / low) ** 2
    return np.sqrt(value.rolling(window, min_periods=1).mean() / (4 * math.log(2))) * math.sqrt(TRADING_DAYS_PER_YEAR)


def garman_klass_vol(open_: pd.Series, high: pd.Series, low: pd.Series, close: pd.Series, window: int) -> pd.Series:
    value = 0.5 * np.log(high / low) ** 2 - (2 * math.log(2) - 1) * np.log(close / open_) ** 2
    return np.sqrt(value.clip(lower=0).rolling(window, min_periods=1).mean()) * math.sqrt(TRADING_DAYS_PER_YEAR)


def rogers_satchell_vol(open_: pd.Series, high: pd.Series, low: pd.Series, close: pd.Series, window: int) -> pd.Series:
    value = np.log(high / close) * np.log(high / open_) + np.log(low / close) * np.log(low / open_)
    return np.sqrt(value.clip(lower=0).rolling(window, min_periods=1).mean()) * math.sqrt(TRADING_DAYS_PER_YEAR)


def optional_column(frame: pd.DataFrame, *names: str) -> pd.Series:
    for name in names:
        if name in frame.columns:
            return pd.to_numeric(frame[name], errors="coerce").fillna(0.0)
    return pd.Series(np.zeros(len(frame)), index=frame.index)


def zscore(series: pd.Series, window: int) -> pd.Series:
    mean = series.rolling(window, min_periods=2).mean()
    std = series.rolling(window, min_periods=2).std(ddof=0)
    return (series - mean) / std


def signed_volume_imbalance(open_: pd.Series, close: pd.Series, volume: pd.Series, window: int) -> pd.Series:
    direction = np.sign(close - open_)
    close_direction = np.sign(close - close.shift(1))
    direction = direction.where(direction != 0, close_direction).fillna(0.0)
    signed = direction * volume
    return signed.rolling(window, min_periods=1).sum() / volume.rolling(window, min_periods=1).sum()


def rolling_beta(left: pd.Series, right: pd.Series, window: int) -> pd.Series:
    mean_left = left.rolling(window, min_periods=2).mean()
    mean_right = right.rolling(window, min_periods=2).mean()
    mean_product = (left * right).rolling(window, min_periods=2).mean()
    covariance = mean_product - mean_left * mean_right
    variance = right.rolling(window, min_periods=2).var(ddof=0)
    return covariance / variance


def rolling_std_population(series: pd.Series, window: int) -> pd.Series:
    return series.rolling(window, min_periods=2).std(ddof=0)


def ewma_vol_go(log_ret: pd.Series, window: int) -> pd.Series:
    values = pd.to_numeric(log_ret, errors="coerce").fillna(0.0).to_numpy()
    out = np.zeros(len(values))
    decay = 1.0 - (2.0 / (window + 1.0))
    for i in range(len(values)):
        start = max(1, i - window + 1)
        returns = values[start : i + 1]
        if len(returns) == 0:
            continue
        weights = decay ** np.arange(len(returns) - 1, -1, -1)
        variance = float(np.sum(weights * returns * returns) / np.sum(weights))
        out[i] = math.sqrt(max(0.0, variance)) * math.sqrt(TRADING_DAYS_PER_YEAR)
    return pd.Series(out, index=log_ret.index)


def clean_series(value: Any, length: int) -> pd.Series:
    if isinstance(value, pd.Series):
        series = value
    else:
        series = pd.Series(np.full(length, value))
    return pd.to_numeric(series, errors="coerce").replace([np.inf, -np.inf], np.nan).fillna(0.0)


def write_report(
    path: Path,
    features: pd.DataFrame,
    labels: pd.DataFrame,
    allowed_sides: set[int],
    args: argparse.Namespace,
) -> None:
    by_side = {}
    if not labels.empty:
        for side, group in labels.groupby("side", sort=True):
            by_side[str(int(side))] = {
                "events": int(len(group)),
                "positive_labels": int((group["label"] == 1).sum()),
                "negative_labels": int((group["label"] == 0).sum()),
            }
    payload = {
        "generated_by": command_line(),
        "git_sha": git_sha(),
        "bars_csv": args.bars_csv,
        "signals_csv": args.signals_csv,
        "hmm_posteriors_csv": args.hmm_posteriors_csv,
        "feature_spec_sha256": file_sha256(args.feature_spec),
        "label_config_sha256": file_sha256(args.label_config),
        "label_sides": sorted(allowed_sides),
        "num_feature_rows": int(len(features)),
        "num_labeled_events": int(len(labels)),
        "positive_labels": int((labels["label"] == 1).sum()) if not labels.empty else 0,
        "negative_labels": int((labels["label"] == 0).sum()) if not labels.empty else 0,
        "labels_by_side": by_side,
        "symbols": sorted(features["symbol"].unique().tolist()) if not features.empty else [],
    }
    with open(path, "w", encoding="utf-8") as handle:
        json.dump(payload, handle, indent=2)
        handle.write("\n")


def main() -> None:
    args = parse_args()
    feature_spec = load_yaml(args.feature_spec)
    label_config = load_yaml(args.label_config)
    allowed_sides = resolve_label_sides(args, label_config)
    bars = load_bars(args.bars_csv)
    signals = load_signals(args.signals_csv)
    hmm_posteriors = load_hmm_posteriors(args.hmm_posteriors_csv)

    features = build_features(
        bars,
        signals,
        feature_spec["features"],
        hmm_posteriors,
        args.estimate_hmm_posteriors,
        args.hmm_states,
        args.hmm_iterations,
        args.fracdiff_d,
        args.fracdiff_width,
    )
    labels = build_labels(bars, signals, label_config, allowed_sides)

    out_dir = Path(args.out_dir)
    out_dir.mkdir(parents=True, exist_ok=True)
    features.to_csv(out_dir / "features.csv", index=False)
    labels.to_csv(out_dir / "labeled_events.csv", index=False)
    write_report(out_dir / "dataset_report.json", features, labels, allowed_sides, args)
    write_manifest(
        out_dir / "manifest.json",
        {
            "artifact_id": f"dataset_{pd.Timestamp.utcnow().isoformat()}",
            "git_sha": git_sha(),
            "dataset_command": command_line(),
            "bars_csv": args.bars_csv,
            "signals_csv": args.signals_csv,
            "feature_spec_sha256": file_sha256(args.feature_spec),
            "label_config_sha256": file_sha256(args.label_config),
            "symbols": sorted(features["symbol"].unique().tolist()) if not features.empty else [],
            "label_sides": sorted(allowed_sides),
            "feature_rows": int(len(features)),
            "labeled_events": int(len(labels)),
        },
    )
    print(f"wrote {len(features)} feature rows and {len(labels)} labels to {out_dir}")


if __name__ == "__main__":
    main()
