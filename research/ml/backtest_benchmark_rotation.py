#!/usr/bin/env python3
"""Backtest benchmark-funded alpha rotation.

The portfolio is always invested in either:

- the benchmark symbol, e.g. SPY; or
- one alpha symbol selected from accepted base BUY signals.

Signals are decided at the close of bar t and, by default, executed at the next
bar open with a simple spread/slippage cost model. A close-to-close mode remains
available only as a diagnostic ablation.
"""

from __future__ import annotations

import argparse
import csv
import json
import math
from dataclasses import asdict, dataclass
from pathlib import Path
from typing import Any

import lightgbm as lgb
import numpy as np
import pandas as pd
import yaml

from artifact_manifest import command_line, research_status_accepted, write_manifest

TRADING_DAYS_PER_YEAR = 252.0
BENCHMARK_PROXY_SYMBOLS = {"SPY", "VOO"}
ETF_SYMBOLS = {
    "DIA",
    "IWM",
    "QQQ",
    "SMH",
    "SPY",
    "VTI",
    "VOO",
    "XLB",
    "XLC",
    "XLE",
    "XLF",
    "XLI",
    "XLK",
    "XLP",
    "XLRE",
    "XLU",
    "XLV",
    "XLY",
}


@dataclass(frozen=True)
class RotationPolicy:
    require_candidate_model: bool = False
    threshold_floor: float | None = None
    candidate_universe: str = "all"
    min_relative_strength_21: float | None = None
    min_log_ret_21: float | None = None
    max_close_to_close_vol_20: float | None = None
    max_hold_bars: int = 0
    stop_loss_pct: float = 0.0
    take_profit_pct: float = 0.0
    selection_mode: str = "probability"
    selection_momentum_weight: float = 0.0


@dataclass(frozen=True)
class ExecutionConfig:
    mode: str = "next_open"
    spread_bps: float = 2.0
    slippage_bps: float = 1.0
    commission_per_share: float = 0.0
    min_commission: float = 0.0
    sec_fees_bps_sell: float = 0.0


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--bars-csv", required=True)
    parser.add_argument("--signals-csv", required=True)
    parser.add_argument("--features-csv", required=True)
    parser.add_argument("--metadata", required=True)
    parser.add_argument("--feature-spec", default="research/ml/feature_spec.yaml")
    parser.add_argument("--benchmark", default="SPY")
    parser.add_argument("--initial-cash", type=float, default=100000)
    parser.add_argument("--mode", choices=["ml", "base"], default="ml")
    parser.add_argument("--require-candidate-model", action="store_true")
    parser.add_argument("--threshold-floor", type=float)
    parser.add_argument("--candidate-universe", choices=["all", "stocks", "etfs"], default="all")
    parser.add_argument("--min-relative-strength-21", type=float)
    parser.add_argument("--min-log-ret-21", type=float)
    parser.add_argument("--max-close-to-close-vol-20", type=float)
    parser.add_argument("--max-hold-bars", type=int, default=0)
    parser.add_argument("--stop-loss-pct", type=float, default=0.0)
    parser.add_argument("--take-profit-pct", type=float, default=0.0)
    parser.add_argument("--selection-mode", choices=["probability", "probability_plus_momentum"], default="probability")
    parser.add_argument("--selection-momentum-weight", type=float, default=0.0)
    parser.add_argument("--execution-mode", choices=["next_open", "close_to_close"], default="next_open")
    parser.add_argument("--spread-bps", type=float, default=2.0)
    parser.add_argument("--slippage-bps", type=float, default=1.0)
    parser.add_argument("--commission-per-share", type=float, default=0.0)
    parser.add_argument("--min-commission", type=float, default=0.0)
    parser.add_argument("--sec-fees-bps-sell", type=float, default=0.0)
    parser.add_argument("--out-dir", required=True)
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    metadata_path = Path(args.metadata)
    metadata = json.loads(metadata_path.read_text(encoding="utf-8"))
    feature_spec = load_yaml(args.feature_spec)
    threshold = float(metadata.get("thresholds", {}).get("enter_long", 0.5))

    bars = load_bars(args.bars_csv)
    signals = load_signals(args.signals_csv)
    features = load_features(args.features_csv, feature_spec["features"])
    model = lgb.Booster(model_file=str(model_path(metadata_path, metadata)))
    policy = RotationPolicy(
        require_candidate_model=args.require_candidate_model,
        threshold_floor=args.threshold_floor,
        candidate_universe=args.candidate_universe,
        min_relative_strength_21=args.min_relative_strength_21,
        min_log_ret_21=args.min_log_ret_21,
        max_close_to_close_vol_20=args.max_close_to_close_vol_20,
        max_hold_bars=args.max_hold_bars,
        stop_loss_pct=args.stop_loss_pct,
        take_profit_pct=args.take_profit_pct,
        selection_mode=args.selection_mode,
        selection_momentum_weight=args.selection_momentum_weight,
    )
    execution = ExecutionConfig(
        mode=args.execution_mode,
        spread_bps=args.spread_bps,
        slippage_bps=args.slippage_bps,
        commission_per_share=args.commission_per_share,
        min_commission=args.min_commission,
        sec_fees_bps_sell=args.sec_fees_bps_sell,
    )

    result = run_rotation(
        bars=bars,
        signals=signals,
        features=features,
        feature_names=feature_spec["features"],
        model=model,
        threshold=threshold,
        benchmark=args.benchmark.upper(),
        initial_cash=args.initial_cash,
        use_ml=args.mode == "ml",
        policy=policy,
        allow_alpha=not policy.require_candidate_model or research_status_accepted(metadata.get("status")),
        execution=execution,
        calibration=metadata.get("calibration", {"method": "none"}),
    )

    out_dir = Path(args.out_dir)
    out_dir.mkdir(parents=True, exist_ok=True)
    result["manifest"] = {
        "backtest_command": command_line(),
        "metadata": str(metadata_path),
        "benchmark": args.benchmark.upper(),
        "cost_model": asdict(execution),
        "artifact_status": metadata.get("status"),
        "symbols": sorted(bars["symbol"].unique().tolist()),
    }
    write_json(out_dir / f"benchmark_rotation_{args.mode}.json", result)
    write_markdown(out_dir / f"benchmark_rotation_{args.mode}.md", result)
    write_csv(out_dir / f"benchmark_rotation_{args.mode}_equity.csv", result["equity_curve"])
    write_csv(out_dir / f"benchmark_rotation_{args.mode}_trades.csv", result["trades"])
    write_manifest(out_dir / f"benchmark_rotation_{args.mode}_manifest.json", result["manifest"])

    summary = result["summary"]
    print(
        f"{args.mode} rotation return={summary['total_return']*100:.2f}% "
        f"sharpe={summary['sharpe']:.3f} maxDD={summary['max_drawdown']*100:.2f}% "
        f"trades={summary['num_trades']} benchmark={summary['benchmark_return']*100:.2f}% "
        f"execution={summary['execution_mode']} costs=${summary['total_cost']:.2f}"
    )


def load_yaml(path: str | Path) -> dict[str, Any]:
    with open(path, "r", encoding="utf-8") as handle:
        return yaml.safe_load(handle)


def model_path(metadata_path: Path, metadata: dict[str, Any]) -> Path:
    uri = str(metadata.get("artifact_uri", ".")).strip() or "."
    path = Path(uri)
    if not path.is_absolute():
        path = metadata_path.parent / path
    if path.is_file():
        return path
    return path / "model.txt"


def load_bars(path: str | Path) -> pd.DataFrame:
    df = pd.read_csv(path)
    df["symbol"] = df["symbol"].str.upper().str.strip()
    df["time"] = pd.to_datetime(df["time"], utc=True)
    for column in ["open", "high", "low", "close", "volume"]:
        df[column] = pd.to_numeric(df[column], errors="coerce")
    return df.sort_values(["time", "symbol"]).reset_index(drop=True)


def load_signals(path: str | Path) -> pd.DataFrame:
    df = pd.read_csv(path)
    df["symbol"] = df["symbol"].str.upper().str.strip()
    df["time"] = pd.to_datetime(df["time"], utc=True)
    df["signal"] = df["signal"].str.upper().str.strip()
    return df.sort_values(["time", "symbol"]).reset_index(drop=True)


def load_features(path: str | Path, feature_names: list[str]) -> pd.DataFrame:
    df = pd.read_csv(path)
    df["symbol"] = df["symbol"].str.upper().str.strip()
    df["event_time"] = pd.to_datetime(df["event_time"], utc=True)
    for name in feature_names:
        if name not in df.columns:
            df[name] = 0.0
        df[name] = pd.to_numeric(df[name], errors="coerce").replace([np.inf, -np.inf], np.nan).fillna(0.0)
    return df.sort_values(["event_time", "symbol"]).reset_index(drop=True)


def run_rotation(
    bars: pd.DataFrame,
    signals: pd.DataFrame,
    features: pd.DataFrame,
    feature_names: list[str],
    model: lgb.Booster,
    threshold: float,
    benchmark: str,
    initial_cash: float,
    use_ml: bool,
    policy: RotationPolicy | None = None,
    allow_alpha: bool = True,
    execution: ExecutionConfig | None = None,
    calibration: dict[str, Any] | None = None,
) -> dict[str, Any]:
    execution = execution or ExecutionConfig()
    if execution.mode == "close_to_close":
        result = run_rotation_close_to_close(
            bars=bars,
            signals=signals,
            features=features,
            feature_names=feature_names,
            model=model,
            threshold=threshold,
            benchmark=benchmark,
            initial_cash=initial_cash,
            use_ml=use_ml,
            policy=policy,
            allow_alpha=allow_alpha,
            calibration=calibration,
        )
    elif execution.mode == "next_open":
        result = run_rotation_next_open(
            bars=bars,
            signals=signals,
            features=features,
            feature_names=feature_names,
            model=model,
            threshold=threshold,
            benchmark=benchmark,
            initial_cash=initial_cash,
            use_ml=use_ml,
            policy=policy,
            allow_alpha=allow_alpha,
            execution=execution,
            calibration=calibration,
        )
    else:
        raise ValueError(f"unsupported execution mode {execution.mode!r}")
    result["summary"]["execution_mode"] = execution.mode
    result["summary"]["cost_model"] = asdict(execution)
    return result


def run_rotation_close_to_close(
    bars: pd.DataFrame,
    signals: pd.DataFrame,
    features: pd.DataFrame,
    feature_names: list[str],
    model: lgb.Booster,
    threshold: float,
    benchmark: str,
    initial_cash: float,
    use_ml: bool,
    policy: RotationPolicy | None = None,
    allow_alpha: bool = True,
    calibration: dict[str, Any] | None = None,
) -> dict[str, Any]:
    policy = policy or RotationPolicy()
    effective_threshold = max(threshold, policy.threshold_floor or 0.0)
    close = bars.pivot(index="time", columns="symbol", values="close").sort_index()
    close = close.dropna(axis=1, how="all")
    if benchmark not in close.columns:
        raise ValueError(f"benchmark {benchmark} missing from bars")
    close = close.ffill()
    returns = close.pct_change().fillna(0.0)
    times = list(close.index)

    signal_map = {
        (row.symbol, row.time): row.signal
        for row in signals[["symbol", "time", "signal"]].itertuples(index=False)
    }
    candidate_probs = score_candidates(signals, features, feature_names, model, benchmark, policy, calibration) if allow_alpha else {}

    equity = float(initial_cash)
    benchmark_equity = float(initial_cash)
    holding = benchmark
    entry_symbol = benchmark
    entry_time = times[0]
    entry_index = 0
    entry_equity = equity
    trades: list[dict[str, Any]] = []
    curve: list[dict[str, Any]] = []
    decisions: list[dict[str, Any]] = []

    for i, current_time in enumerate(times):
        if i > 0:
            holding_return = safe_return(returns, current_time, holding)
            benchmark_return = safe_return(returns, current_time, benchmark)
            equity *= 1.0 + holding_return
            benchmark_equity *= 1.0 + benchmark_return
        curve.append(
            {
                "time": current_time.isoformat(),
                "equity": equity,
                "benchmark_equity": benchmark_equity,
                "holding": holding,
            }
        )

        if i >= len(times) - 1:
            continue

        next_holding = holding
        reason = "hold"
        probability = None
        if holding != benchmark:
            trade_return = equity / entry_equity - 1.0 if entry_equity > 0 else 0.0
            hold_bars = i - entry_index
            policy_exit = ""
            if policy.stop_loss_pct > 0 and trade_return <= -policy.stop_loss_pct:
                policy_exit = "stop_loss"
            elif policy.take_profit_pct > 0 and trade_return >= policy.take_profit_pct:
                policy_exit = "take_profit"
            elif policy.max_hold_bars > 0 and hold_bars >= policy.max_hold_bars:
                policy_exit = "max_hold"
            if policy_exit:
                close_trade(trades, entry_symbol, entry_time, current_time, entry_equity, equity, policy_exit)
                next_holding = benchmark
                entry_symbol = benchmark
                entry_time = current_time
                entry_index = i
                entry_equity = equity
                reason = f"alpha_{policy_exit}_to_benchmark"

        if next_holding != benchmark and signal_map.get((holding, current_time)) == "SELL":
            close_trade(trades, entry_symbol, entry_time, current_time, entry_equity, equity, "base_sell")
            next_holding = benchmark
            entry_symbol = benchmark
            entry_time = current_time
            entry_index = i
            entry_equity = equity
            reason = "alpha_sell_to_benchmark"

        if next_holding == benchmark:
            candidates = candidate_probs.get(current_time, [])
            if not use_ml:
                candidates = [{**candidate, "probability": 1.0, "score": 1.0} for candidate in candidates]
            accepted = [candidate for candidate in candidates if candidate["probability"] >= effective_threshold or not use_ml]
            if accepted:
                accepted.sort(key=lambda item: (item["score"], item["probability"], item["symbol"]), reverse=True)
                selected = accepted[0]
                next_holding = selected["symbol"]
                probability = selected["probability"] if use_ml else None
                entry_symbol = next_holding
                entry_time = current_time
                entry_index = i
                entry_equity = equity
                reason = "accepted_alpha_buy" if use_ml else "base_alpha_buy"

        if next_holding != holding or reason != "hold":
            decisions.append(
                {
                    "time": current_time.isoformat(),
                    "from": holding,
                    "to": next_holding,
                    "reason": reason,
                    "probability": probability,
                    "equity": equity,
                    "policy": asdict(policy),
                }
            )
        holding = next_holding

    if holding != benchmark:
        close_trade(trades, entry_symbol, entry_time, times[-1], entry_equity, equity, "end_of_test")

    summary = summarize(curve, trades, initial_cash, benchmark)
    summary["benchmark"] = benchmark
    summary["excess_return_vs_benchmark"] = summary["total_return"] - summary["benchmark_return"]
    summary["mode"] = "ml" if use_ml else "base"
    summary["threshold"] = threshold
    summary["effective_threshold"] = effective_threshold
    summary["policy"] = asdict(policy)
    summary["alpha_enabled"] = allow_alpha
    summary["total_cost"] = 0.0
    summary["turnover"] = 0.0
    return {
        "summary": summary,
        "decisions": decisions,
        "trades": trades,
        "equity_curve": curve,
    }


def run_rotation_next_open(
    bars: pd.DataFrame,
    signals: pd.DataFrame,
    features: pd.DataFrame,
    feature_names: list[str],
    model: lgb.Booster,
    threshold: float,
    benchmark: str,
    initial_cash: float,
    use_ml: bool,
    policy: RotationPolicy | None,
    allow_alpha: bool,
    execution: ExecutionConfig,
    calibration: dict[str, Any] | None,
) -> dict[str, Any]:
    policy = policy or RotationPolicy()
    effective_threshold = max(threshold, policy.threshold_floor or 0.0)
    open_ = bars.pivot(index="time", columns="symbol", values="open").sort_index().dropna(axis=1, how="all").ffill()
    close = bars.pivot(index="time", columns="symbol", values="close").sort_index().dropna(axis=1, how="all").ffill()
    if benchmark not in close.columns or benchmark not in open_.columns:
        raise ValueError(f"benchmark {benchmark} missing from bars")
    times = list(close.index)

    signal_map = {
        (row.symbol, row.time): row.signal
        for row in signals[["symbol", "time", "signal"]].itertuples(index=False)
    }
    candidate_probs = score_candidates(signals, features, feature_names, model, benchmark, policy, calibration) if allow_alpha else {}

    equity = float(initial_cash)
    benchmark_equity = float(initial_cash)
    holding = benchmark
    entry_symbol = benchmark
    entry_time = times[0]
    entry_index = 0
    entry_equity = equity
    pending: dict[str, Any] | None = None
    total_cost = 0.0
    turnover_notional = 0.0
    trades: list[dict[str, Any]] = []
    curve: list[dict[str, Any]] = [
        {
            "time": times[0].isoformat(),
            "equity": equity,
            "benchmark_equity": benchmark_equity,
            "holding": holding,
        }
    ]
    decisions: list[dict[str, Any]] = []

    for i in range(1, len(times)):
        current_time = times[i]
        previous_time = times[i - 1]

        equity *= price_ratio(close, open_, previous_time, current_time, holding)
        benchmark_equity *= price_ratio(close, open_, previous_time, current_time, benchmark)

        if pending is not None and pending["to"] != holding:
            from_symbol = holding
            to_symbol = str(pending["to"])
            sell_price = safe_price(open_, current_time, from_symbol)
            buy_price = safe_price(open_, current_time, to_symbol)
            sell_cost = trade_cost(equity, sell_price, "sell", execution)
            buy_notional = max(0.0, equity - sell_cost)
            buy_cost = trade_cost(buy_notional, buy_price, "buy", execution)
            fill_cost = sell_cost + buy_cost
            total_cost += fill_cost
            turnover_notional += equity + buy_notional
            equity = max(0.0, equity - fill_cost)

            if from_symbol != benchmark:
                close_trade(trades, entry_symbol, entry_time, current_time, entry_equity, equity, pending["reason"])
            holding = to_symbol
            if holding != benchmark:
                entry_symbol = holding
                entry_time = current_time
                entry_index = i
                entry_equity = equity
            else:
                entry_symbol = benchmark
                entry_time = current_time
                entry_index = i
                entry_equity = equity
            decisions.append(
                {
                    "decision_time": pending["decision_time"].isoformat(),
                    "execution_time": current_time.isoformat(),
                    "from": from_symbol,
                    "to": to_symbol,
                    "reason": pending["reason"],
                    "probability": pending.get("probability"),
                    "equity_after_cost": equity,
                    "execution_cost": fill_cost,
                    "policy": asdict(policy),
                }
            )
            pending = None

        equity *= price_ratio(open_, close, current_time, current_time, holding)
        benchmark_equity *= price_ratio(open_, close, current_time, current_time, benchmark)
        curve.append(
            {
                "time": current_time.isoformat(),
                "equity": equity,
                "benchmark_equity": benchmark_equity,
                "holding": holding,
            }
        )

        if i >= len(times) - 1:
            continue

        next_holding = holding
        reason = "hold"
        probability = None
        if holding != benchmark:
            trade_return = equity / entry_equity - 1.0 if entry_equity > 0 else 0.0
            hold_bars = i - entry_index
            policy_exit = ""
            if policy.stop_loss_pct > 0 and trade_return <= -policy.stop_loss_pct:
                policy_exit = "stop_loss"
            elif policy.take_profit_pct > 0 and trade_return >= policy.take_profit_pct:
                policy_exit = "take_profit"
            elif policy.max_hold_bars > 0 and hold_bars >= policy.max_hold_bars:
                policy_exit = "max_hold"
            if policy_exit:
                next_holding = benchmark
                reason = f"alpha_{policy_exit}_to_benchmark"

        if next_holding != benchmark and signal_map.get((holding, current_time)) == "SELL":
            next_holding = benchmark
            reason = "alpha_sell_to_benchmark"

        if next_holding == benchmark:
            candidates = candidate_probs.get(current_time, [])
            if not use_ml:
                candidates = [{**candidate, "probability": 1.0, "score": 1.0} for candidate in candidates]
            accepted = [candidate for candidate in candidates if candidate["probability"] >= effective_threshold or not use_ml]
            if accepted:
                accepted.sort(key=lambda item: (item["score"], item["probability"], item["symbol"]), reverse=True)
                selected = accepted[0]
                next_holding = selected["symbol"]
                probability = selected["probability"] if use_ml else None
                reason = "accepted_alpha_buy" if use_ml else "base_alpha_buy"

        if next_holding != holding:
            pending = {
                "decision_time": current_time,
                "to": next_holding,
                "reason": reason,
                "probability": probability,
            }

    if holding != benchmark:
        close_trade(trades, entry_symbol, entry_time, times[-1], entry_equity, equity, "end_of_test")

    summary = summarize(curve, trades, initial_cash, benchmark)
    summary["benchmark"] = benchmark
    summary["excess_return_vs_benchmark"] = summary["total_return"] - summary["benchmark_return"]
    summary["mode"] = "ml" if use_ml else "base"
    summary["threshold"] = threshold
    summary["effective_threshold"] = effective_threshold
    summary["policy"] = asdict(policy)
    summary["alpha_enabled"] = allow_alpha
    summary["total_cost"] = float(total_cost)
    summary["turnover"] = float(turnover_notional / initial_cash) if initial_cash > 0 else 0.0
    return {
        "summary": summary,
        "decisions": decisions,
        "trades": trades,
        "equity_curve": curve,
    }


def score_candidates(
    signals: pd.DataFrame,
    features: pd.DataFrame,
    feature_names: list[str],
    model: lgb.Booster,
    benchmark: str,
    policy: RotationPolicy,
    calibration: dict[str, Any] | None = None,
) -> dict[pd.Timestamp, list[dict[str, Any]]]:
    excluded = set(BENCHMARK_PROXY_SYMBOLS)
    excluded.add(benchmark)
    buys = signals[(signals["signal"] == "BUY") & (~signals["symbol"].isin(excluded))][["symbol", "time"]].copy()
    merged = buys.merge(
        features[["symbol", "event_time", *feature_names]],
        left_on=["symbol", "time"],
        right_on=["symbol", "event_time"],
        how="inner",
    )
    if merged.empty:
        return {}
    raw_probs = model.predict(merged[feature_names].astype(float), num_iteration=model.best_iteration)
    probs = apply_calibration(raw_probs, calibration)
    out: dict[pd.Timestamp, list[dict[str, Any]]] = {}
    for row, prob, raw_prob in zip(merged.itertuples(index=False), probs, raw_probs):
        symbol = str(row.symbol)
        if not universe_allows(symbol, policy.candidate_universe):
            continue
        rel_strength = float(getattr(row, "relative_strength_vs_spy_21", 0.0))
        log_ret_21 = float(getattr(row, "log_ret_21", 0.0))
        vol_20 = float(getattr(row, "close_to_close_vol_20", 0.0))
        if policy.min_relative_strength_21 is not None and rel_strength < policy.min_relative_strength_21:
            continue
        if policy.min_log_ret_21 is not None and log_ret_21 < policy.min_log_ret_21:
            continue
        if policy.max_close_to_close_vol_20 is not None and vol_20 > policy.max_close_to_close_vol_20:
            continue
        probability = float(prob)
        score = probability
        if policy.selection_mode == "probability_plus_momentum":
            score = probability + policy.selection_momentum_weight * rel_strength
        out.setdefault(row.time, []).append(
            {
                "symbol": symbol,
                "probability": probability,
                "raw_probability": float(raw_prob),
                "score": float(score),
                "relative_strength_vs_spy_21": rel_strength,
                "log_ret_21": log_ret_21,
                "close_to_close_vol_20": vol_20,
            }
        )
    return out


def universe_allows(symbol: str, candidate_universe: str) -> bool:
    is_etf = symbol in ETF_SYMBOLS
    if candidate_universe == "stocks":
        return not is_etf
    if candidate_universe == "etfs":
        return is_etf
    return True


def apply_calibration(probabilities: np.ndarray, calibration: dict[str, Any] | None) -> np.ndarray:
    probs = np.clip(np.asarray(probabilities, dtype=float), 1e-9, 1 - 1e-9)
    calibration = calibration or {"method": "none"}
    method = str(calibration.get("method", "none")).strip().lower()
    if method in {"", "none"}:
        return probs
    if method in {"platt", "logistic"}:
        intercept = float(calibration.get("intercept", 0.0))
        slope = float(calibration.get("slope", 1.0))
        logits = np.log(probs / (1.0 - probs))
        return np.clip(1.0 / (1.0 + np.exp(-(intercept + slope * logits))), 1e-9, 1 - 1e-9)
    if method in {"histogram", "isotonic"}:
        out = probs.copy()
        for bin_ in calibration.get("bins", []) or []:
            lower = float(bin_.get("lower", 0.0))
            upper = float(bin_.get("upper", 1.0))
            calibrated = float(bin_.get("calibrated", 0.5))
            mask = (probs >= lower) & (probs <= upper)
            out[mask] = calibrated
        return np.clip(out, 1e-9, 1 - 1e-9)
    return probs


def safe_return(returns: pd.DataFrame, time: pd.Timestamp, symbol: str) -> float:
    if symbol not in returns.columns:
        return 0.0
    value = returns.at[time, symbol]
    if pd.isna(value) or not math.isfinite(float(value)):
        return 0.0
    return float(value)


def safe_price(prices: pd.DataFrame, time: pd.Timestamp, symbol: str) -> float:
    if symbol not in prices.columns:
        return 0.0
    value = prices.at[time, symbol]
    if pd.isna(value) or not math.isfinite(float(value)) or float(value) <= 0:
        return 0.0
    return float(value)


def price_ratio(from_prices: pd.DataFrame, to_prices: pd.DataFrame, from_time: pd.Timestamp, to_time: pd.Timestamp, symbol: str) -> float:
    start = safe_price(from_prices, from_time, symbol)
    end = safe_price(to_prices, to_time, symbol)
    if start <= 0 or end <= 0:
        return 1.0
    return end / start


def trade_cost(notional: float, price: float, side: str, execution: ExecutionConfig) -> float:
    if notional <= 0 or price <= 0:
        return 0.0
    half_spread_bps = max(0.0, execution.spread_bps) / 2.0
    slippage_bps = max(0.0, execution.slippage_bps)
    cost = notional * (half_spread_bps + slippage_bps) / 10_000.0
    if execution.commission_per_share > 0:
        commission = execution.commission_per_share * (notional / price)
        if execution.min_commission > 0:
            commission = max(commission, execution.min_commission)
        cost += commission
    if side == "sell" and execution.sec_fees_bps_sell > 0:
        cost += notional * execution.sec_fees_bps_sell / 10_000.0
    return float(cost)


def close_trade(
    trades: list[dict[str, Any]],
    symbol: str,
    entry_time: pd.Timestamp,
    exit_time: pd.Timestamp,
    entry_equity: float,
    exit_equity: float,
    reason: str,
) -> None:
    if symbol == "":
        return
    trades.append(
        {
            "symbol": symbol,
            "entry_time": entry_time.isoformat(),
            "exit_time": exit_time.isoformat(),
            "entry_equity": entry_equity,
            "exit_equity": exit_equity,
            "return": exit_equity / entry_equity - 1.0 if entry_equity > 0 else 0.0,
            "reason": reason,
        }
    )


def summarize(
    curve: list[dict[str, Any]],
    trades: list[dict[str, Any]],
    initial_cash: float,
    benchmark_symbol: str,
) -> dict[str, Any]:
    equity = np.array([float(row["equity"]) for row in curve])
    benchmark = np.array([float(row["benchmark_equity"]) for row in curve])
    if len(equity) < 2:
        return {}
    returns = equity[1:] / equity[:-1] - 1.0
    benchmark_returns = benchmark[1:] / benchmark[:-1] - 1.0
    total_return = equity[-1] / initial_cash - 1.0
    benchmark_return = benchmark[-1] / initial_cash - 1.0
    annualized_return = (1.0 + total_return) ** (TRADING_DAYS_PER_YEAR / max(1, len(returns))) - 1.0
    benchmark_annualized_return = (
        (1.0 + benchmark_return) ** (TRADING_DAYS_PER_YEAR / max(1, len(benchmark_returns))) - 1.0
    )
    volatility = float(np.std(returns)) * math.sqrt(TRADING_DAYS_PER_YEAR)
    benchmark_volatility = float(np.std(benchmark_returns)) * math.sqrt(TRADING_DAYS_PER_YEAR)
    sharpe = annualized_return / volatility if volatility > 0 else 0.0
    benchmark_sharpe = benchmark_annualized_return / benchmark_volatility if benchmark_volatility > 0 else 0.0
    downside = returns[returns < 0]
    benchmark_downside = benchmark_returns[benchmark_returns < 0]
    downside_vol = float(np.std(downside)) * math.sqrt(TRADING_DAYS_PER_YEAR) if len(downside) else 0.0
    benchmark_downside_vol = (
        float(np.std(benchmark_downside)) * math.sqrt(TRADING_DAYS_PER_YEAR)
        if len(benchmark_downside)
        else 0.0
    )
    sortino = annualized_return / downside_vol if downside_vol > 0 else 0.0
    benchmark_sortino = (
        benchmark_annualized_return / benchmark_downside_vol if benchmark_downside_vol > 0 else 0.0
    )
    drawdown = max_drawdown(equity)
    benchmark_drawdown = max_drawdown(benchmark)
    return {
        "start_equity": initial_cash,
        "final_equity": float(equity[-1]),
        "total_return": float(total_return),
        "benchmark_return": float(benchmark_return),
        "annualized_return": float(annualized_return),
        "benchmark_annualized_return": float(benchmark_annualized_return),
        "volatility": float(volatility),
        "benchmark_volatility": float(benchmark_volatility),
        "sharpe": float(sharpe),
        "benchmark_sharpe": float(benchmark_sharpe),
        "sortino": float(sortino),
        "benchmark_sortino": float(benchmark_sortino),
        "max_drawdown": float(drawdown),
        "benchmark_max_drawdown": float(benchmark_drawdown),
        "num_trades": len([trade for trade in trades if trade["symbol"] != benchmark_symbol]),
    }


def max_drawdown(equity: np.ndarray) -> float:
    peak = np.maximum.accumulate(equity)
    dd = equity / peak - 1.0
    return float(abs(np.min(dd))) if len(dd) else 0.0


def write_json(path: Path, value: dict[str, Any]) -> None:
    path.write_text(json.dumps(value, indent=2) + "\n", encoding="utf-8")


def write_csv(path: Path, rows: list[dict[str, Any]]) -> None:
    if not rows:
        path.write_text("", encoding="utf-8")
        return
    with path.open("w", newline="", encoding="utf-8") as handle:
        writer = csv.DictWriter(handle, fieldnames=list(rows[0].keys()))
        writer.writeheader()
        writer.writerows(rows)


def write_markdown(path: Path, result: dict[str, Any]) -> None:
    s = result["summary"]
    lines = [
        "# Benchmark-Funded Alpha Rotation\n\n",
        f"- Mode: `{s['mode']}`\n",
        f"- Benchmark: `{s['benchmark']}`\n",
        f"- Threshold: `{s['threshold']:.12f}`\n",
        f"- Effective threshold: `{s['effective_threshold']:.12f}`\n",
        f"- Candidate universe: `{s['policy']['candidate_universe']}`\n\n",
        f"- Execution mode: `{s.get('execution_mode', 'unknown')}`\n",
        f"- Total explicit cost: `${s.get('total_cost', 0.0):.2f}`\n",
        f"- Turnover: `{s.get('turnover', 0.0):.3f}`\n\n",
        "| Metric | Strategy | Benchmark |\n",
        "|---|---:|---:|\n",
        f"| Total return | {s['total_return']*100:.2f}% | {s['benchmark_return']*100:.2f}% |\n",
        f"| Excess return | {s['excess_return_vs_benchmark']*100:.2f}% | 0.00% |\n",
        f"| Annualized return | {s['annualized_return']*100:.2f}% | {s['benchmark_annualized_return']*100:.2f}% |\n",
        f"| Sharpe | {s['sharpe']:.3f} | {s['benchmark_sharpe']:.3f} |\n",
        f"| Sortino | {s['sortino']:.3f} | {s['benchmark_sortino']:.3f} |\n",
        f"| Max drawdown | {s['max_drawdown']*100:.2f}% | {s['benchmark_max_drawdown']*100:.2f}% |\n",
        f"| Trades | {s['num_trades']} |  |\n",
    ]
    path.write_text("".join(lines), encoding="utf-8")


if __name__ == "__main__":
    main()
