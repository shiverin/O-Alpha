#!/usr/bin/env python3
"""Portfolio helpers for benchmark-core active sleeve research.

The simulator assumes signals are decided at the close of bar t and target
weights are executed at the next bar open. It keeps the benchmark as the
residual allocation and applies explicit spread/slippage costs to every
rebalance order.
"""

from __future__ import annotations

import csv
import json
import math
from dataclasses import asdict
from pathlib import Path
from typing import Any

import numpy as np
import pandas as pd

from backtest_benchmark_rotation import ExecutionConfig, safe_price, trade_cost

TRADING_DAYS_PER_YEAR = 252.0


def cap_weight_budget(raw_weights: dict[str, float], budget: float, max_weight: float) -> dict[str, float]:
    """Normalize positive raw weights into a capped long-only budget."""
    budget = max(0.0, float(budget))
    max_weight = max(0.0, float(max_weight))
    positive = {symbol: max(0.0, float(weight)) for symbol, weight in raw_weights.items() if weight > 0}
    if budget <= 0 or max_weight <= 0 or not positive:
        return {}

    capped: dict[str, float] = {}
    remaining_symbols = set(positive)
    remaining_budget = min(budget, max_weight * len(positive))

    while remaining_symbols and remaining_budget > 1e-12:
        total_raw = sum(positive[symbol] for symbol in remaining_symbols)
        if total_raw <= 0:
            equal = remaining_budget / len(remaining_symbols)
            for symbol in list(remaining_symbols):
                allocation = min(max_weight, equal)
                capped[symbol] = capped.get(symbol, 0.0) + allocation
                remaining_budget -= allocation
                remaining_symbols.remove(symbol)
            break

        progressed = False
        for symbol in list(remaining_symbols):
            proposed = remaining_budget * positive[symbol] / total_raw
            if proposed >= max_weight:
                capped[symbol] = max_weight
                remaining_budget -= max_weight
                remaining_symbols.remove(symbol)
                progressed = True
        if not progressed:
            for symbol in list(remaining_symbols):
                capped[symbol] = remaining_budget * positive[symbol] / total_raw
            break

    return {symbol: float(weight) for symbol, weight in capped.items() if weight > 1e-9}


def complete_with_benchmark(weights: dict[str, float], benchmark: str) -> dict[str, float]:
    cleaned = {symbol: max(0.0, float(weight)) for symbol, weight in weights.items() if weight > 1e-9}
    active_total = sum(weight for symbol, weight in cleaned.items() if symbol != benchmark)
    benchmark_weight = max(0.0, 1.0 - active_total)
    cleaned[benchmark] = cleaned.get(benchmark, 0.0) + benchmark_weight
    total = sum(cleaned.values())
    if total <= 0:
        return {benchmark: 1.0}
    return {symbol: weight / total for symbol, weight in cleaned.items() if weight / total > 1e-9}


def simulate_target_weights(
    open_: pd.DataFrame,
    close: pd.DataFrame,
    benchmark: str,
    target_by_decision_time: dict[pd.Timestamp, dict[str, float]],
    initial_cash: float,
    execution: ExecutionConfig,
    rebalance_band: float = 0.0,
) -> dict[str, Any]:
    """Run a next-open, target-weight portfolio backtest."""
    open_ = open_.sort_index().ffill()
    close = close.sort_index().ffill()
    common_times = close.index.intersection(open_.index).sort_values()
    close = close.loc[common_times]
    open_ = open_.loc[common_times]
    if benchmark not in close.columns or benchmark not in open_.columns:
        raise ValueError(f"benchmark {benchmark} missing from price panel")
    times = list(common_times)
    if len(times) < 2:
        raise ValueError("portfolio backtest needs at least two bars")

    equity = float(initial_cash)
    benchmark_equity = float(initial_cash)
    weights: dict[str, float] = {benchmark: 1.0}
    pending_target: dict[str, float] | None = None
    pending_decision_time: pd.Timestamp | None = None
    total_cost = 0.0
    turnover_notional = 0.0
    orders: list[dict[str, Any]] = []
    decisions: list[dict[str, Any]] = []
    curve: list[dict[str, Any]] = [
        {
            "time": times[0].isoformat(),
            "equity": equity,
            "benchmark_equity": benchmark_equity,
            "benchmark_weight": 1.0,
            "active_weight": 0.0,
            "active_symbols": "",
            "weights": json.dumps(weights, sort_keys=True),
        }
    ]

    for i in range(1, len(times)):
        current_time = times[i]
        previous_time = times[i - 1]

        equity, weights = apply_weighted_price_move(equity, weights, close, open_, previous_time, current_time)
        benchmark_equity *= benchmark_price_ratio(close, open_, previous_time, current_time, benchmark)

        if pending_target is not None:
            target = complete_with_benchmark(pending_target, benchmark)
            distance = weight_distance(weights, target)
            if distance > rebalance_band:
                fill = rebalance_at_open(
                    time=current_time,
                    equity=equity,
                    weights=weights,
                    target=target,
                    open_prices=open_,
                    execution=execution,
                )
                equity = fill["equity"]
                weights = fill["weights"]
                total_cost += fill["cost"]
                turnover_notional += fill["turnover_notional"]
                orders.extend(fill["orders"])
                decisions.append(
                    {
                        "decision_time": pending_decision_time.isoformat() if pending_decision_time is not None else "",
                        "execution_time": current_time.isoformat(),
                        "active_symbols": active_symbols(weights, benchmark),
                        "benchmark_weight": weights.get(benchmark, 0.0),
                        "active_weight": 1.0 - weights.get(benchmark, 0.0),
                        "turnover_distance": distance,
                        "execution_cost": fill["cost"],
                        "target_weights": json.dumps(target, sort_keys=True),
                    }
                )
            pending_target = None
            pending_decision_time = None

        equity, weights = apply_weighted_price_move(equity, weights, open_, close, current_time, current_time)
        benchmark_equity *= benchmark_price_ratio(open_, close, current_time, current_time, benchmark)
        curve.append(
            {
                "time": current_time.isoformat(),
                "equity": equity,
                "benchmark_equity": benchmark_equity,
                "benchmark_weight": weights.get(benchmark, 0.0),
                "active_weight": 1.0 - weights.get(benchmark, 0.0),
                "active_symbols": active_symbols(weights, benchmark),
                "weights": json.dumps(weights, sort_keys=True),
            }
        )

        if i < len(times) - 1 and current_time in target_by_decision_time:
            pending_target = target_by_decision_time[current_time]
            pending_decision_time = current_time

    summary = summarize_portfolio(
        curve=curve,
        orders=orders,
        decisions=decisions,
        initial_cash=initial_cash,
        benchmark=benchmark,
        total_cost=total_cost,
        turnover_notional=turnover_notional,
        execution=execution,
    )
    return {
        "summary": summary,
        "decisions": decisions,
        "orders": orders,
        "equity_curve": curve,
    }


def apply_weighted_price_move(
    equity: float,
    weights: dict[str, float],
    from_prices: pd.DataFrame,
    to_prices: pd.DataFrame,
    from_time: pd.Timestamp,
    to_time: pd.Timestamp,
) -> tuple[float, dict[str, float]]:
    gross_return = 0.0
    moved: dict[str, float] = {}
    for symbol, weight in weights.items():
        ratio = symbol_price_ratio(from_prices, to_prices, from_time, to_time, symbol)
        contribution = weight * ratio
        gross_return += contribution
        moved[symbol] = contribution
    if gross_return <= 0 or not math.isfinite(gross_return):
        return equity, weights
    new_weights = {
        symbol: contribution / gross_return
        for symbol, contribution in moved.items()
        if contribution / gross_return > 1e-9
    }
    return equity * gross_return, new_weights


def rebalance_at_open(
    time: pd.Timestamp,
    equity: float,
    weights: dict[str, float],
    target: dict[str, float],
    open_prices: pd.DataFrame,
    execution: ExecutionConfig,
) -> dict[str, Any]:
    symbols = sorted(set(weights) | set(target))
    orders: list[dict[str, Any]] = []
    total_cost = 0.0
    turnover_notional = 0.0
    for symbol in symbols:
        current_weight = float(weights.get(symbol, 0.0))
        target_weight = float(target.get(symbol, 0.0))
        delta = target_weight - current_weight
        if abs(delta) <= 1e-8:
            continue
        price = safe_price(open_prices, time, symbol)
        side = "buy" if delta > 0 else "sell"
        notional = abs(delta) * equity
        cost = trade_cost(notional, price, side, execution)
        total_cost += cost
        turnover_notional += notional
        orders.append(
            {
                "time": time.isoformat(),
                "symbol": symbol,
                "side": side,
                "current_weight": current_weight,
                "target_weight": target_weight,
                "delta_weight": delta,
                "notional": notional,
                "price": price,
                "cost": cost,
            }
        )
    equity = max(0.0, equity - total_cost)
    return {"equity": equity, "weights": target, "cost": total_cost, "turnover_notional": turnover_notional, "orders": orders}


def benchmark_price_ratio(
    from_prices: pd.DataFrame,
    to_prices: pd.DataFrame,
    from_time: pd.Timestamp,
    to_time: pd.Timestamp,
    benchmark: str,
) -> float:
    return symbol_price_ratio(from_prices, to_prices, from_time, to_time, benchmark)


def symbol_price_ratio(
    from_prices: pd.DataFrame,
    to_prices: pd.DataFrame,
    from_time: pd.Timestamp,
    to_time: pd.Timestamp,
    symbol: str,
) -> float:
    start = safe_price(from_prices, from_time, symbol)
    end = safe_price(to_prices, to_time, symbol)
    if start <= 0 or end <= 0:
        return 1.0
    return end / start


def weight_distance(current: dict[str, float], target: dict[str, float]) -> float:
    return 0.5 * sum(abs(float(target.get(symbol, 0.0)) - float(current.get(symbol, 0.0))) for symbol in set(current) | set(target))


def active_symbols(weights: dict[str, float], benchmark: str) -> str:
    return ";".join(symbol for symbol, weight in sorted(weights.items()) if symbol != benchmark and weight > 1e-6)


def summarize_portfolio(
    curve: list[dict[str, Any]],
    orders: list[dict[str, Any]],
    decisions: list[dict[str, Any]],
    initial_cash: float,
    benchmark: str,
    total_cost: float,
    turnover_notional: float,
    execution: ExecutionConfig,
) -> dict[str, Any]:
    equity = np.array([float(row["equity"]) for row in curve])
    benchmark_equity = np.array([float(row["benchmark_equity"]) for row in curve])
    returns = equity[1:] / equity[:-1] - 1.0
    benchmark_returns = benchmark_equity[1:] / benchmark_equity[:-1] - 1.0
    total_return = equity[-1] / initial_cash - 1.0
    benchmark_return = benchmark_equity[-1] / initial_cash - 1.0
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
        float(np.std(benchmark_downside)) * math.sqrt(TRADING_DAYS_PER_YEAR) if len(benchmark_downside) else 0.0
    )
    sortino = annualized_return / downside_vol if downside_vol > 0 else 0.0
    benchmark_sortino = benchmark_annualized_return / benchmark_downside_vol if benchmark_downside_vol > 0 else 0.0
    active_weights = np.array([float(row["active_weight"]) for row in curve])
    alpha_symbols = {order["symbol"] for order in orders if order["symbol"] != benchmark}
    return {
        "benchmark": benchmark,
        "start_equity": initial_cash,
        "final_equity": float(equity[-1]),
        "total_return": float(total_return),
        "benchmark_return": float(benchmark_return),
        "excess_return_vs_benchmark": float(total_return - benchmark_return),
        "annualized_return": float(annualized_return),
        "benchmark_annualized_return": float(benchmark_annualized_return),
        "volatility": float(volatility),
        "benchmark_volatility": float(benchmark_volatility),
        "sharpe": float(sharpe),
        "benchmark_sharpe": float(benchmark_sharpe),
        "sortino": float(sortino),
        "benchmark_sortino": float(benchmark_sortino),
        "max_drawdown": max_drawdown(equity),
        "benchmark_max_drawdown": max_drawdown(benchmark_equity),
        "total_cost": float(total_cost),
        "turnover": float(turnover_notional / initial_cash) if initial_cash > 0 else 0.0,
        "num_rebalances": len(decisions),
        "num_orders": len(orders),
        "num_alpha_symbols_traded": len(alpha_symbols),
        "average_active_weight": float(np.mean(active_weights)) if len(active_weights) else 0.0,
        "max_active_weight": float(np.max(active_weights)) if len(active_weights) else 0.0,
        "execution_mode": execution.mode,
        "cost_model": asdict(execution),
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
