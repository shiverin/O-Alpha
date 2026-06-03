#!/usr/bin/env python3
"""Train and backtest a daily LightGBM ranker active sleeve.

The ranker is a portfolio-level research path: it scores every liquid candidate
daily, keeps VOO/SPY as the benchmark core, and allocates only a capped active
sleeve to the top-ranked names. Training uses only historical daily bars and a
chronological purged validation split.
"""

from __future__ import annotations

import argparse
import json
import math
from dataclasses import asdict, dataclass
from pathlib import Path
from typing import Any

import lightgbm as lgb
import numpy as np
import pandas as pd

from artifact_manifest import command_line, file_sha256, write_manifest
from backtest_benchmark_rotation import BENCHMARK_PROXY_SYMBOLS, ETF_SYMBOLS, ExecutionConfig, load_bars
from portfolio_sleeve import cap_weight_budget, complete_with_benchmark, simulate_target_weights, write_csv, write_json


FEATURE_NAMES = [
    "log_ret_1",
    "log_ret_5",
    "log_ret_10",
    "log_ret_21",
    "log_ret_63",
    "log_ret_126",
    "log_ret_252",
    "excess_log_ret_5",
    "excess_log_ret_21",
    "excess_log_ret_63",
    "excess_log_ret_126",
    "vol_20",
    "vol_63",
    "downside_vol_20",
    "return_to_vol_21",
    "return_to_vol_63",
    "beta_63",
    "corr_63",
    "residual_log_ret_21",
    "residual_log_ret_63",
    "distance_to_63d_high",
    "distance_to_63d_low",
    "ma_20_50",
    "ma_50_200",
    "volume_z_20",
    "dollar_volume_z_20",
    "amihud_20",
    "gap_pct",
    "intraday_ret",
    "benchmark_log_ret_21",
    "benchmark_vol_20",
]


@dataclass(frozen=True)
class RankerConfig:
    horizon_bars: int
    min_history_bars: int
    validation_fraction: float
    relevance_bins: int
    objective: str
    seed: int
    num_boost_round: int
    early_stopping_rounds: int


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--train-bars-csv", required=True)
    parser.add_argument("--test-bars-csv", required=True)
    parser.add_argument("--benchmark", default="VOO")
    parser.add_argument("--candidate-universe", choices=["all", "stocks", "etfs"], default="all")
    parser.add_argument("--initial-cash", type=float, default=100000.0)
    parser.add_argument("--horizon-bars", type=int, default=21)
    parser.add_argument("--min-history-bars", type=int, default=252)
    parser.add_argument("--validation-fraction", type=float, default=0.25)
    parser.add_argument("--relevance-bins", type=int, default=5)
    parser.add_argument("--objective", choices=["lambdarank", "regression"], default="lambdarank")
    parser.add_argument("--seed", type=int, default=17)
    parser.add_argument("--num-boost-round", type=int, default=400)
    parser.add_argument("--early-stopping-rounds", type=int, default=40)
    parser.add_argument("--sleeve-fraction", type=float, default=0.30)
    parser.add_argument("--top-k", type=int, default=3)
    parser.add_argument("--max-name-weight", type=float, default=0.10)
    parser.add_argument("--rebalance-every", type=int, default=21)
    parser.add_argument("--rebalance-band", type=float, default=0.025)
    parser.add_argument("--min-score-z", type=float, default=0.0)
    parser.add_argument("--max-candidate-vol", type=float, default=0.0)
    parser.add_argument("--max-benchmark-vol", type=float, default=0.0)
    parser.add_argument("--high-vol-scale", type=float, default=1.0)
    parser.add_argument("--max-benchmark-drawdown", type=float, default=0.0)
    parser.add_argument("--drawdown-scale", type=float, default=1.0)
    parser.add_argument("--spread-bps", type=float, default=2.0)
    parser.add_argument("--slippage-bps", type=float, default=1.0)
    parser.add_argument("--commission-per-share", type=float, default=0.0)
    parser.add_argument("--min-commission", type=float, default=0.0)
    parser.add_argument("--sec-fees-bps-sell", type=float, default=0.0)
    parser.add_argument("--out-dir", required=True)
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    benchmark = args.benchmark.upper()
    train_bars = load_bars(args.train_bars_csv)
    test_bars = load_bars(args.test_bars_csv)
    config = RankerConfig(
        horizon_bars=args.horizon_bars,
        min_history_bars=args.min_history_bars,
        validation_fraction=args.validation_fraction,
        relevance_bins=args.relevance_bins,
        objective=args.objective,
        seed=args.seed,
        num_boost_round=args.num_boost_round,
        early_stopping_rounds=args.early_stopping_rounds,
    )
    execution = ExecutionConfig(
        mode="next_open",
        spread_bps=args.spread_bps,
        slippage_bps=args.slippage_bps,
        commission_per_share=args.commission_per_share,
        min_commission=args.min_commission,
        sec_fees_bps_sell=args.sec_fees_bps_sell,
    )

    feature_frame, open_panel, close_panel = build_daily_feature_frame(train_bars, test_bars, benchmark, config)
    train_rows = feature_frame[
        (feature_frame["source"] == "train")
        & feature_frame["label_excess"].notna()
        & feature_frame["symbol"].map(lambda symbol: universe_allows(symbol, args.candidate_universe, benchmark))
    ].copy()
    test_rows = feature_frame[
        (feature_frame["source"] == "test")
        & feature_frame["symbol"].map(lambda symbol: universe_allows(symbol, args.candidate_universe, benchmark))
    ].copy()
    if train_rows.empty or test_rows.empty:
        raise ValueError("daily ranker needs non-empty train and test rows after universe filtering")

    model, training_report = train_ranker(train_rows, config)
    result = backtest_ranker_sleeve(
        model=model,
        test_rows=test_rows,
        open_panel=open_panel.loc[open_panel.index.isin(test_bars["time"].unique())],
        close_panel=close_panel.loc[close_panel.index.isin(test_bars["time"].unique())],
        benchmark=benchmark,
        initial_cash=args.initial_cash,
        execution=execution,
        sleeve_fraction=args.sleeve_fraction,
        top_k=args.top_k,
        max_name_weight=args.max_name_weight,
        rebalance_every=args.rebalance_every,
        rebalance_band=args.rebalance_band,
        min_score_z=args.min_score_z,
        max_candidate_vol=args.max_candidate_vol,
        max_benchmark_vol=args.max_benchmark_vol,
        high_vol_scale=args.high_vol_scale,
        max_benchmark_drawdown=args.max_benchmark_drawdown,
        drawdown_scale=args.drawdown_scale,
    )
    result["summary"].update(
        {
            "strategy": "daily_lightgbm_ranker_active_sleeve",
            "candidate_universe": args.candidate_universe,
            "horizon_bars": args.horizon_bars,
            "min_history_bars": args.min_history_bars,
            "sleeve_fraction": args.sleeve_fraction,
            "top_k": args.top_k,
            "max_name_weight": args.max_name_weight,
            "rebalance_every": args.rebalance_every,
            "rebalance_band": args.rebalance_band,
            "training": training_report,
        }
    )

    out_dir = Path(args.out_dir)
    out_dir.mkdir(parents=True, exist_ok=True)
    model_path = out_dir / "daily_ranker_model.txt"
    model.save_model(str(model_path))
    feature_importance = feature_importance_rows(model)
    manifest = {
        "command": command_line(),
        "train_bars_csv": args.train_bars_csv,
        "test_bars_csv": args.test_bars_csv,
        "train_bars_sha256": file_sha256(args.train_bars_csv),
        "test_bars_sha256": file_sha256(args.test_bars_csv),
        "benchmark": benchmark,
        "candidate_universe": args.candidate_universe,
        "features": FEATURE_NAMES,
        "ranker_config": asdict(config),
        "portfolio_config": {
            "sleeve_fraction": args.sleeve_fraction,
            "top_k": args.top_k,
            "max_name_weight": args.max_name_weight,
            "rebalance_every": args.rebalance_every,
            "rebalance_band": args.rebalance_band,
            "min_score_z": args.min_score_z,
            "max_candidate_vol": args.max_candidate_vol,
            "max_benchmark_vol": args.max_benchmark_vol,
            "high_vol_scale": args.high_vol_scale,
            "max_benchmark_drawdown": args.max_benchmark_drawdown,
            "drawdown_scale": args.drawdown_scale,
        },
        "cost_model": asdict(execution),
        "model_file": str(model_path),
        "status": promotion_status(result["summary"], training_report),
    }
    result["manifest"] = manifest
    write_json(out_dir / "daily_ranker_sleeve.json", result)
    write_json(out_dir / "training_report.json", training_report)
    write_json(out_dir / "feature_spec.json", {"features": FEATURE_NAMES})
    write_csv(out_dir / "daily_ranker_sleeve_equity.csv", result["equity_curve"])
    write_csv(out_dir / "daily_ranker_sleeve_orders.csv", result["orders"])
    write_csv(out_dir / "daily_ranker_sleeve_decisions.csv", result["decisions"])
    write_csv(out_dir / "daily_ranker_sleeve_selections.csv", result["selections"])
    write_csv(out_dir / "feature_importance.csv", feature_importance)
    write_markdown(out_dir / "daily_ranker_sleeve.md", result)
    write_manifest(out_dir / "daily_ranker_sleeve_manifest.json", manifest)

    summary = result["summary"]
    print(
        f"daily ranker sleeve return={summary['total_return']*100:.2f}% "
        f"benchmark={summary['benchmark_return']*100:.2f}% "
        f"excess={summary['excess_return_vs_benchmark']*100:.2f}% "
        f"sharpe={summary['sharpe']:.3f} maxDD={summary['max_drawdown']*100:.2f}% "
        f"orders={summary['num_orders']} alpha_symbols={summary['num_alpha_symbols_traded']} "
        f"status={manifest['status']}"
    )


def build_daily_feature_frame(
    train_bars: pd.DataFrame,
    test_bars: pd.DataFrame,
    benchmark: str,
    config: RankerConfig,
) -> tuple[pd.DataFrame, pd.DataFrame, pd.DataFrame]:
    train = train_bars.copy()
    test = test_bars.copy()
    train["source"] = "train"
    test["source"] = "test"
    bars = pd.concat([train, test], ignore_index=True).sort_values(["time", "symbol"])
    open_panel = bars.pivot(index="time", columns="symbol", values="open").sort_index().dropna(axis=1, how="all").ffill()
    high_panel = bars.pivot(index="time", columns="symbol", values="high").sort_index().dropna(axis=1, how="all").ffill()
    low_panel = bars.pivot(index="time", columns="symbol", values="low").sort_index().dropna(axis=1, how="all").ffill()
    close_panel = bars.pivot(index="time", columns="symbol", values="close").sort_index().dropna(axis=1, how="all").ffill()
    volume_panel = bars.pivot(index="time", columns="symbol", values="volume").sort_index().dropna(axis=1, how="all").ffill()
    source_by_time = bars.groupby("time")["source"].last().to_dict()
    if benchmark not in close_panel.columns or benchmark not in open_panel.columns:
        raise ValueError(f"benchmark {benchmark} missing from bars")

    benchmark_log_close = np.log(close_panel[benchmark])
    benchmark_ret_1 = benchmark_log_close.diff()
    benchmark_log_ret_21 = benchmark_log_close - benchmark_log_close.shift(21)
    benchmark_log_ret_63 = benchmark_log_close - benchmark_log_close.shift(63)
    benchmark_log_ret_126 = benchmark_log_close - benchmark_log_close.shift(126)
    benchmark_vol_20 = benchmark_ret_1.rolling(20, min_periods=5).std() * math.sqrt(252.0)
    benchmark_future = np.log(close_panel[benchmark].shift(-config.horizon_bars) / open_panel[benchmark].shift(-1))
    rows: list[pd.DataFrame] = []

    for symbol in close_panel.columns:
        close = close_panel[symbol]
        open_ = open_panel[symbol]
        high = high_panel[symbol]
        low = low_panel[symbol]
        volume = volume_panel[symbol]
        log_close = np.log(close)
        ret_1 = log_close.diff()
        pct_ret_1 = close.pct_change()
        vol_20 = ret_1.rolling(20, min_periods=5).std() * math.sqrt(252.0)
        vol_63 = ret_1.rolling(63, min_periods=20).std() * math.sqrt(252.0)
        downside_vol_20 = ret_1.where(ret_1 < 0).rolling(20, min_periods=5).std() * math.sqrt(252.0)
        beta_63 = ret_1.rolling(63, min_periods=20).cov(benchmark_ret_1) / benchmark_ret_1.rolling(63, min_periods=20).var()
        corr_63 = ret_1.rolling(63, min_periods=20).corr(benchmark_ret_1)
        log_ret_5 = log_close - log_close.shift(5)
        log_ret_21 = log_close - log_close.shift(21)
        log_ret_63 = log_close - log_close.shift(63)
        log_ret_126 = log_close - log_close.shift(126)
        rolling_high_63 = close.rolling(63, min_periods=20).max()
        rolling_low_63 = close.rolling(63, min_periods=20).min()
        ma_20 = close.rolling(20, min_periods=5).mean()
        ma_50 = close.rolling(50, min_periods=20).mean()
        ma_200 = close.rolling(200, min_periods=60).mean()
        dollar_volume = close * volume
        log_dollar_volume = np.log(dollar_volume.replace(0, np.nan))
        amihud = (pct_ret_1.abs() / dollar_volume.replace(0, np.nan)).rolling(20, min_periods=5).mean()
        future_symbol = np.log(close.shift(-config.horizon_bars) / open_.shift(-1))
        frame = pd.DataFrame(
            {
                "symbol": symbol,
                "time": close.index,
                "source": [source_by_time.get(time, "") for time in close.index],
                "bar_number": np.arange(len(close.index)),
                "log_ret_1": ret_1,
                "log_ret_5": log_ret_5,
                "log_ret_10": log_close - log_close.shift(10),
                "log_ret_21": log_ret_21,
                "log_ret_63": log_ret_63,
                "log_ret_126": log_ret_126,
                "log_ret_252": log_close - log_close.shift(252),
                "excess_log_ret_5": log_ret_5 - (benchmark_log_close - benchmark_log_close.shift(5)),
                "excess_log_ret_21": log_ret_21 - benchmark_log_ret_21,
                "excess_log_ret_63": log_ret_63 - benchmark_log_ret_63,
                "excess_log_ret_126": log_ret_126 - benchmark_log_ret_126,
                "vol_20": vol_20,
                "vol_63": vol_63,
                "downside_vol_20": downside_vol_20,
                "return_to_vol_21": log_ret_21 / vol_20.replace(0, np.nan),
                "return_to_vol_63": log_ret_63 / vol_63.replace(0, np.nan),
                "beta_63": beta_63,
                "corr_63": corr_63,
                "residual_log_ret_21": log_ret_21 - beta_63 * benchmark_log_ret_21,
                "residual_log_ret_63": log_ret_63 - beta_63 * benchmark_log_ret_63,
                "distance_to_63d_high": close / rolling_high_63 - 1.0,
                "distance_to_63d_low": close / rolling_low_63 - 1.0,
                "ma_20_50": ma_20 / ma_50 - 1.0,
                "ma_50_200": ma_50 / ma_200 - 1.0,
                "volume_z_20": (volume - volume.rolling(20, min_periods=5).mean())
                / volume.rolling(20, min_periods=5).std().replace(0, np.nan),
                "dollar_volume_z_20": (log_dollar_volume - log_dollar_volume.rolling(20, min_periods=5).mean())
                / log_dollar_volume.rolling(20, min_periods=5).std().replace(0, np.nan),
                "amihud_20": amihud,
                "gap_pct": open_ / close.shift(1) - 1.0,
                "intraday_ret": close / open_ - 1.0,
                "benchmark_log_ret_21": benchmark_log_ret_21,
                "benchmark_vol_20": benchmark_vol_20,
                "label_excess": future_symbol - benchmark_future,
                "future_return": future_symbol,
                "future_benchmark_return": benchmark_future,
                "open": open_,
                "high": high,
                "low": low,
                "close": close,
            }
        )
        rows.append(frame)

    feature_frame = pd.concat(rows, ignore_index=True)
    feature_frame = feature_frame[feature_frame["bar_number"] >= config.min_history_bars].copy()
    for name in FEATURE_NAMES:
        feature_frame[name] = (
            pd.to_numeric(feature_frame[name], errors="coerce").replace([np.inf, -np.inf], np.nan).fillna(0.0)
        )
    feature_frame["label_excess"] = pd.to_numeric(feature_frame["label_excess"], errors="coerce")
    feature_frame["time"] = pd.to_datetime(feature_frame["time"], utc=True)
    return feature_frame.sort_values(["time", "symbol"]).reset_index(drop=True), open_panel, close_panel


def train_ranker(train_rows: pd.DataFrame, config: RankerConfig) -> tuple[lgb.Booster, dict[str, Any]]:
    train_rows = add_relevance(train_rows, config.relevance_bins)
    unique_times = sorted(train_rows["time"].unique())
    validation_count = max(1, int(len(unique_times) * config.validation_fraction))
    validation_times = set(unique_times[-validation_count:])
    validation_start_index = len(unique_times) - validation_count
    purge_cutoff_index = max(0, validation_start_index - config.horizon_bars)
    purged_train_times = set(unique_times[:purge_cutoff_index])

    fit_rows = train_rows[train_rows["time"].isin(purged_train_times)].copy()
    valid_rows = train_rows[train_rows["time"].isin(validation_times)].copy()
    if fit_rows.empty or valid_rows.empty:
        raise ValueError("purged chronological validation split produced empty train or validation rows")

    fit_rows = fit_rows.sort_values(["time", "symbol"]).reset_index(drop=True)
    valid_rows = valid_rows.sort_values(["time", "symbol"]).reset_index(drop=True)
    params = lightgbm_params(config)
    train_set = make_dataset(fit_rows, config)
    valid_set = make_dataset(valid_rows, config)
    model = lgb.train(
        params,
        train_set,
        num_boost_round=config.num_boost_round,
        valid_sets=[valid_set],
        valid_names=["valid"],
        early_stopping_rounds=config.early_stopping_rounds,
        verbose_eval=False,
    )
    fit_scores = predict(model, fit_rows)
    valid_scores = predict(model, valid_rows)
    report = {
        "objective": config.objective,
        "train_rows": int(len(fit_rows)),
        "validation_rows": int(len(valid_rows)),
        "train_dates": int(fit_rows["time"].nunique()),
        "validation_dates": int(valid_rows["time"].nunique()),
        "purged_dates": int(validation_start_index - purge_cutoff_index),
        "best_iteration": int(model.best_iteration or model.current_iteration()),
        "best_score": model.best_score,
        "train_metrics": ranking_metrics(fit_rows, fit_scores),
        "validation_metrics": ranking_metrics(valid_rows, valid_scores),
    }
    return model, report


def lightgbm_params(config: RankerConfig) -> dict[str, Any]:
    common = {
        "learning_rate": 0.03,
        "num_leaves": 15,
        "min_data_in_leaf": 40,
        "feature_fraction": 0.85,
        "bagging_fraction": 0.85,
        "bagging_freq": 1,
        "lambda_l1": 0.0,
        "lambda_l2": 2.0,
        "seed": config.seed,
        "feature_pre_filter": False,
        "verbosity": -1,
        "num_threads": 0,
    }
    if config.objective == "lambdarank":
        common.update({"objective": "lambdarank", "metric": "ndcg", "ndcg_eval_at": [1, 3, 5]})
    else:
        common.update({"objective": "regression", "metric": "l2"})
    return common


def make_dataset(rows: pd.DataFrame, config: RankerConfig) -> lgb.Dataset:
    label_column = "relevance" if config.objective == "lambdarank" else "label_excess"
    dataset = lgb.Dataset(rows[FEATURE_NAMES].astype(float), label=rows[label_column].astype(float), free_raw_data=False)
    if config.objective == "lambdarank":
        dataset.set_group(rows.groupby("time", sort=False).size().astype(int).tolist())
    return dataset


def add_relevance(rows: pd.DataFrame, bins: int) -> pd.DataFrame:
    rows = rows.copy()
    bins = max(2, int(bins))
    ranks = rows.groupby("time")["label_excess"].transform(lambda values: values.rank(method="first", pct=True))
    rows["relevance"] = np.floor(ranks * bins).clip(0, bins - 1).astype(int)
    return rows


def predict(model: lgb.Booster, rows: pd.DataFrame) -> np.ndarray:
    return model.predict(rows[FEATURE_NAMES].astype(float), num_iteration=model.best_iteration)


def ranking_metrics(rows: pd.DataFrame, scores: np.ndarray, top_k: int = 3) -> dict[str, float]:
    scored = rows[["time", "symbol", "label_excess"]].copy()
    scored["score"] = scores
    top_excess: list[float] = []
    universe_excess: list[float] = []
    top1_hits: list[float] = []
    rank_ic: list[float] = []
    for _, group in scored.groupby("time"):
        group = group.dropna(subset=["label_excess"])
        if len(group) < max(2, top_k):
            continue
        ranked = group.sort_values("score", ascending=False)
        top = ranked.head(top_k)
        top_excess.append(float(top["label_excess"].mean()))
        universe_excess.append(float(group["label_excess"].mean()))
        top1_hits.append(float(ranked.iloc[0]["label_excess"] > 0))
        score_rank = group["score"].rank()
        label_rank = group["label_excess"].rank()
        corr = score_rank.corr(label_rank)
        if corr == corr and math.isfinite(float(corr)):
            rank_ic.append(float(corr))
    return {
        "mean_top_excess": float(np.mean(top_excess)) if top_excess else 0.0,
        "mean_universe_excess": float(np.mean(universe_excess)) if universe_excess else 0.0,
        "mean_top_minus_universe": float(np.mean(np.array(top_excess) - np.array(universe_excess))) if top_excess else 0.0,
        "top1_hit_rate": float(np.mean(top1_hits)) if top1_hits else 0.0,
        "mean_rank_ic": float(np.mean(rank_ic)) if rank_ic else 0.0,
        "dates": float(len(top_excess)),
    }


def backtest_ranker_sleeve(
    *,
    model: lgb.Booster,
    test_rows: pd.DataFrame,
    open_panel: pd.DataFrame,
    close_panel: pd.DataFrame,
    benchmark: str,
    initial_cash: float,
    execution: ExecutionConfig,
    sleeve_fraction: float,
    top_k: int,
    max_name_weight: float,
    rebalance_every: int,
    rebalance_band: float,
    min_score_z: float = 0.0,
    max_candidate_vol: float = 0.0,
    max_benchmark_vol: float = 0.0,
    high_vol_scale: float = 1.0,
    max_benchmark_drawdown: float = 0.0,
    drawdown_scale: float = 1.0,
) -> dict[str, Any]:
    scored = test_rows.copy()
    scored["score"] = predict(model, scored)
    scored = scored.sort_values(["time", "symbol"]).reset_index(drop=True)
    close_times = list(close_panel.index)
    decision_times = set(close_times[:: max(1, rebalance_every)])
    target_by_time: dict[pd.Timestamp, dict[str, float]] = {}
    selections: list[dict[str, Any]] = []
    for time, group in scored.groupby("time"):
        if time not in decision_times:
            continue
        if time not in close_panel.index:
            continue
        group = group.copy()
        score_std = float(group["score"].std(ddof=0))
        score_mean = float(group["score"].mean())
        if score_std > 0:
            group["score_z"] = (group["score"] - score_mean) / score_std
        else:
            group["score_z"] = 0.0
        eligible = group
        if min_score_z > 0:
            eligible = eligible[eligible["score_z"] >= min_score_z]
        if max_candidate_vol > 0:
            eligible = eligible[pd.to_numeric(eligible["vol_20"], errors="coerce").fillna(0.0) <= max_candidate_vol]
        ranked = eligible.sort_values("score", ascending=False).head(top_k).copy()
        if ranked.empty:
            target_by_time[time] = {benchmark: 1.0}
            continue
        risk = benchmark_risk_state(
            close_panel=close_panel,
            benchmark=benchmark,
            time=time,
            max_benchmark_vol=max_benchmark_vol,
            high_vol_scale=high_vol_scale,
            max_benchmark_drawdown=max_benchmark_drawdown,
            drawdown_scale=drawdown_scale,
        )
        effective_sleeve = max(0.0, min(1.0, sleeve_fraction * risk["scale"]))
        effective_max_name = max(0.0, min(max_name_weight, effective_sleeve))
        raw = {}
        for rank, row in enumerate(ranked.itertuples(index=False), start=1):
            rank_weight = float(top_k - rank + 1)
            inverse_vol = 1.0 / max(0.05, float(getattr(row, "vol_20", 0.0)))
            raw[row.symbol] = rank_weight * inverse_vol
        active_weights = cap_weight_budget(raw, effective_sleeve, effective_max_name)
        target_by_time[time] = complete_with_benchmark(active_weights, benchmark)
        for rank, row in enumerate(ranked.itertuples(index=False), start=1):
            selections.append(
                {
                    "decision_time": time.isoformat(),
                    "rank": rank,
                    "symbol": row.symbol,
                    "score": float(row.score),
                    "score_z": float(getattr(row, "score_z", 0.0)),
                    "target_weight": float(active_weights.get(row.symbol, 0.0)),
                    "vol_20": float(getattr(row, "vol_20", 0.0)),
                    "sleeve_scale": risk["scale"],
                    "benchmark_vol_20": risk["benchmark_vol_20"],
                    "benchmark_drawdown": risk["benchmark_drawdown"],
                    "risk_reasons": ";".join(risk["reasons"]),
                    "label_excess": None if pd.isna(row.label_excess) else float(row.label_excess),
                }
            )
    result = simulate_target_weights(
        open_=open_panel,
        close=close_panel,
        benchmark=benchmark,
        target_by_decision_time=target_by_time,
        initial_cash=initial_cash,
        execution=execution,
        rebalance_band=rebalance_band,
    )
    result["selections"] = selections
    result["summary"].update(
        {
            "selection_dates": len(target_by_time),
            "selection_rows": len(selections),
            "mean_selected_forward_excess": mean_selection_excess(selections),
            "min_score_z": min_score_z,
            "max_candidate_vol": max_candidate_vol,
            "max_benchmark_vol": max_benchmark_vol,
            "high_vol_scale": high_vol_scale,
            "max_benchmark_drawdown": max_benchmark_drawdown,
            "drawdown_scale": drawdown_scale,
            "average_sleeve_scale": mean_selection_value(selections, "sleeve_scale", 1.0),
        }
    )
    return result


def benchmark_risk_state(
    *,
    close_panel: pd.DataFrame,
    benchmark: str,
    time: pd.Timestamp,
    max_benchmark_vol: float,
    high_vol_scale: float,
    max_benchmark_drawdown: float,
    drawdown_scale: float,
) -> dict[str, Any]:
    scale = 1.0
    reasons: list[str] = []
    vol_20 = 0.0
    drawdown = 0.0
    if benchmark not in close_panel.columns or time not in close_panel.index:
        return {
            "scale": scale,
            "benchmark_vol_20": vol_20,
            "benchmark_drawdown": drawdown,
            "reasons": reasons,
        }
    index = close_panel.index.get_loc(time)
    if isinstance(index, slice):
        index = index.stop - 1
    if not isinstance(index, int):
        index = int(index)
    benchmark_close = pd.to_numeric(close_panel[benchmark].iloc[: index + 1], errors="coerce").dropna()
    if len(benchmark_close) >= 2:
        returns = np.log(benchmark_close / benchmark_close.shift(1)).dropna()
        if len(returns) >= 2:
            vol_20 = float(returns.tail(20).std(ddof=0) * math.sqrt(252.0))
        high = float(benchmark_close.tail(63).max())
        last = float(benchmark_close.iloc[-1])
        if high > 0 and last > 0:
            drawdown = last / high - 1.0
    if max_benchmark_vol > 0 and vol_20 > max_benchmark_vol:
        scale *= max(0.0, min(1.0, high_vol_scale))
        reasons.append("high_benchmark_vol")
    if max_benchmark_drawdown > 0 and drawdown < -max_benchmark_drawdown:
        scale *= max(0.0, min(1.0, drawdown_scale))
        reasons.append("benchmark_drawdown")
    return {
        "scale": scale,
        "benchmark_vol_20": vol_20,
        "benchmark_drawdown": drawdown,
        "reasons": reasons,
    }


def mean_selection_excess(selections: list[dict[str, Any]]) -> float:
    values = [float(row["label_excess"]) for row in selections if row.get("label_excess") is not None]
    return float(np.mean(values)) if values else 0.0


def mean_selection_value(selections: list[dict[str, Any]], key: str, default: float) -> float:
    values = [float(row[key]) for row in selections if row.get(key) is not None]
    return float(np.mean(values)) if values else default


def universe_allows(symbol: str, candidate_universe: str, benchmark: str) -> bool:
    symbol = symbol.upper()
    excluded = set(BENCHMARK_PROXY_SYMBOLS)
    excluded.add(benchmark.upper())
    if symbol in excluded:
        return False
    is_etf = symbol in ETF_SYMBOLS
    if candidate_universe == "stocks":
        return not is_etf
    if candidate_universe == "etfs":
        return is_etf
    return True


def feature_importance_rows(model: lgb.Booster) -> list[dict[str, Any]]:
    gain = model.feature_importance(importance_type="gain")
    split = model.feature_importance(importance_type="split")
    rows = [
        {"feature": name, "gain": float(g), "split": int(s)}
        for name, g, s in zip(FEATURE_NAMES, gain, split)
    ]
    return sorted(rows, key=lambda row: (row["gain"], row["split"]), reverse=True)


def promotion_status(summary: dict[str, Any], training_report: dict[str, Any]) -> str:
    validation = training_report.get("validation_metrics", {})
    if summary["excess_return_vs_benchmark"] <= 0:
        return "rejected_under_benchmark"
    if summary["excess_return_vs_benchmark"] < 0.02:
        return "research_only_low_excess"
    if summary["num_alpha_symbols_traded"] < 5:
        return "research_only_insufficient_breadth"
    if validation.get("mean_top_minus_universe", 0.0) <= 0 or validation.get("mean_rank_ic", 0.0) <= 0:
        return "research_only_weak_validation"
    if summary["max_drawdown"] > summary["benchmark_max_drawdown"] + 0.02:
        return "research_only_drawdown_regression"
    if summary["turnover"] > 8.0:
        return "research_only_high_turnover"
    return "candidate"


def write_markdown(path: Path, result: dict[str, Any]) -> None:
    s = result["summary"]
    validation = s["training"]["validation_metrics"]
    lines = [
        "# Daily LightGBM Ranker Active Sleeve\n\n",
        f"- Benchmark core: `{s['benchmark']}`\n",
        f"- Candidate universe: `{s['candidate_universe']}`\n",
        f"- Horizon bars: `{s['horizon_bars']}`\n",
        f"- Active sleeve: `{s['sleeve_fraction']:.2f}`\n",
        f"- Top-k: `{s['top_k']}`\n",
        f"- Max name weight: `{s['max_name_weight']:.2f}`\n",
        f"- Rebalance every: `{s['rebalance_every']}` bars\n",
        f"- Explicit costs: `${s['total_cost']:.2f}`\n",
        f"- Validation rank IC: `{validation['mean_rank_ic']:.4f}`\n",
        f"- Validation top-k excess over universe: `{validation['mean_top_minus_universe']:.6f}`\n\n",
        "| Metric | Strategy | Benchmark |\n",
        "|---|---:|---:|\n",
        f"| Total return | {s['total_return']*100:.2f}% | {s['benchmark_return']*100:.2f}% |\n",
        f"| Excess return | {s['excess_return_vs_benchmark']*100:.2f}% | 0.00% |\n",
        f"| Annualized return | {s['annualized_return']*100:.2f}% | {s['benchmark_annualized_return']*100:.2f}% |\n",
        f"| Sharpe | {s['sharpe']:.3f} | {s['benchmark_sharpe']:.3f} |\n",
        f"| Sortino | {s['sortino']:.3f} | {s['benchmark_sortino']:.3f} |\n",
        f"| Max drawdown | {s['max_drawdown']*100:.2f}% | {s['benchmark_max_drawdown']*100:.2f}% |\n",
        f"| Turnover | {s['turnover']:.3f} |  |\n",
        f"| Rebalances | {s['num_rebalances']} |  |\n",
        f"| Alpha symbols traded | {s['num_alpha_symbols_traded']} |  |\n",
        f"| Mean selected forward excess | {s['mean_selected_forward_excess']*100:.2f}% |  |\n",
    ]
    path.write_text("".join(lines), encoding="utf-8")


if __name__ == "__main__":
    main()
