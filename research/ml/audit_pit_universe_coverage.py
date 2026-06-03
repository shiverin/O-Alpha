#!/usr/bin/env python3
"""Audit price coverage for a point-in-time constituent universe."""

from __future__ import annotations

import argparse
import json
from pathlib import Path
from typing import Any

import pandas as pd

from artifact_manifest import command_line, file_sha256, git_sha, write_manifest
from portfolio_sleeve import write_csv, write_json


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--pit-universe", required=True)
    parser.add_argument("--bars-csv", required=True)
    parser.add_argument("--out-dir", required=True)
    parser.add_argument("--benchmark", default="VOO")
    parser.add_argument("--from-date", default="")
    parser.add_argument("--to-date", default="")
    parser.add_argument("--min-active-symbols", type=int, default=50)
    parser.add_argument("--min-coverage-ratio", type=float, default=0.95)
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    pit = load_json(args.pit_universe)
    intervals = normalize_intervals(pit.get("intervals", []), args.from_date, args.to_date)
    bars = load_bars(args.bars_csv, args.from_date, args.to_date)
    if bars.empty:
        raise ValueError("bars CSV has no rows after date filters")
    if not intervals:
        raise ValueError("PIT universe has no intervals after date filters")

    trading_days = sorted(pd.to_datetime(bars["date"].unique(), utc=True))
    dates_by_symbol = {
        symbol: set(frame["date"].unique())
        for symbol, frame in bars.groupby("symbol", sort=False)
    }
    bars_symbols = set(dates_by_symbol)
    pit_symbols = sorted({row["symbol"] for row in intervals})
    missing_symbols = [symbol for symbol in pit_symbols if symbol not in bars_symbols]
    observed_symbols = [symbol for symbol in pit_symbols if symbol in bars_symbols]

    daily_rows = []
    active_member_days = 0
    covered_member_days = 0
    for day in trading_days:
        active = active_symbols_on(day, intervals)
        covered = {symbol for symbol in active if symbol in dates_by_symbol and day.date().isoformat() in dates_by_symbol[symbol]}
        active_member_days += len(active)
        covered_member_days += len(covered)
        daily_rows.append(
            {
                "date": day.date().isoformat(),
                "active_symbols": len(active),
                "covered_symbols": len(covered),
                "missing_symbols": max(0, len(active) - len(covered)),
                "coverage_ratio": (len(covered) / len(active)) if active else 1.0,
            }
        )

    yearly_rows = aggregate_yearly(daily_rows)
    coverage_ratio = covered_member_days / active_member_days if active_member_days else 0.0
    min_active_symbols = min((row["active_symbols"] for row in daily_rows), default=0)
    min_covered_symbols = min((row["covered_symbols"] for row in daily_rows), default=0)
    status = "passed"
    reasons = []
    if coverage_ratio < args.min_coverage_ratio:
        status = "failed"
        reasons.append(
            f"coverage_ratio {coverage_ratio:.4f} below threshold {args.min_coverage_ratio:.4f}"
        )
    if min_covered_symbols < args.min_active_symbols:
        status = "failed"
        reasons.append(
            f"min_covered_symbols {min_covered_symbols} below threshold {args.min_active_symbols}"
        )
    if args.benchmark.upper() not in bars_symbols:
        status = "failed"
        reasons.append(f"benchmark {args.benchmark.upper()} missing from bars")

    report = {
        "summary": {
            "status": status,
            "reasons": reasons,
            "pit_symbol_count": len(pit_symbols),
            "bars_symbol_count": len(bars_symbols),
            "observed_pit_symbol_count": len(observed_symbols),
            "missing_pit_symbol_count": len(missing_symbols),
            "active_member_days": active_member_days,
            "covered_member_days": covered_member_days,
            "coverage_ratio": coverage_ratio,
            "trading_days": len(trading_days),
            "min_active_symbols": min_active_symbols,
            "min_covered_symbols": min_covered_symbols,
            "benchmark": args.benchmark.upper(),
        },
        "missing_symbols": missing_symbols,
        "yearly": yearly_rows,
        "daily": daily_rows,
        "manifest": {
            "command": command_line(),
            "git_sha": git_sha(),
            "pit_universe": args.pit_universe,
            "pit_universe_sha256": file_sha256(args.pit_universe),
            "bars_csv": args.bars_csv,
            "bars_csv_sha256": file_sha256(args.bars_csv),
            "from_date": args.from_date,
            "to_date": args.to_date,
            "min_active_symbols": args.min_active_symbols,
            "min_coverage_ratio": args.min_coverage_ratio,
        },
    }

    out_dir = Path(args.out_dir)
    out_dir.mkdir(parents=True, exist_ok=True)
    write_json(out_dir / "pit_coverage_report.json", report)
    write_csv(out_dir / "pit_coverage_daily.csv", daily_rows)
    write_csv(out_dir / "pit_coverage_yearly.csv", yearly_rows)
    write_csv(out_dir / "pit_coverage_missing_symbols.csv", [{"symbol": symbol} for symbol in missing_symbols])
    write_markdown(out_dir / "pit_coverage_report.md", report)
    write_manifest(out_dir / "pit_coverage_manifest.json", report["manifest"])
    s = report["summary"]
    print(
        f"pit coverage {status}: coverage={s['coverage_ratio']:.4f} "
        f"covered_min={s['min_covered_symbols']} missing_symbols={s['missing_pit_symbol_count']}"
    )


def load_json(path: str) -> dict[str, Any]:
    return json.loads(Path(path).read_text(encoding="utf-8"))


def load_bars(path: str, from_date: str, to_date: str) -> pd.DataFrame:
    bars = pd.read_csv(path, parse_dates=["time"])
    bars["symbol"] = bars["symbol"].astype(str).str.upper().str.strip()
    bars["date"] = bars["time"].dt.date.astype(str)
    if from_date:
        bars = bars[bars["date"] >= normalize_date(from_date)]
    if to_date:
        bars = bars[bars["date"] <= normalize_date(to_date)]
    bars = bars[bars["close"].astype(float) > 0].copy()
    return bars


def normalize_intervals(raw_intervals: list[dict[str, Any]], from_date: str, to_date: str) -> list[dict[str, str]]:
    out = []
    min_date = normalize_date(from_date)
    max_date = normalize_date(to_date)
    for row in raw_intervals:
        symbol = str(row.get("symbol", "")).upper().strip()
        start = normalize_date(row.get("start", ""))
        end = normalize_date(row.get("end", ""))
        if not symbol or not start:
            continue
        if max_date and start > max_date:
            continue
        if min_date and end and end < min_date:
            continue
        if min_date and start < min_date:
            start = min_date
        if max_date and (not end or end > max_date):
            end = max_date
        out.append({"symbol": symbol, "start": start, "end": end})
    return out


def active_symbols_on(day: pd.Timestamp, intervals: list[dict[str, str]]) -> set[str]:
    value = day.date().isoformat()
    active = set()
    for row in intervals:
        if row["start"] > value:
            continue
        if row.get("end") and row["end"] < value:
            continue
        active.add(row["symbol"])
    return active


def aggregate_yearly(daily_rows: list[dict[str, Any]]) -> list[dict[str, Any]]:
    grouped: dict[str, list[dict[str, Any]]] = {}
    for row in daily_rows:
        grouped.setdefault(row["date"][:4], []).append(row)
    out = []
    for year, rows in sorted(grouped.items()):
        active_days = sum(int(row["active_symbols"]) for row in rows)
        covered_days = sum(int(row["covered_symbols"]) for row in rows)
        out.append(
            {
                "year": year,
                "trading_days": len(rows),
                "mean_active_symbols": sum(float(row["active_symbols"]) for row in rows) / len(rows),
                "mean_covered_symbols": sum(float(row["covered_symbols"]) for row in rows) / len(rows),
                "min_active_symbols": min(int(row["active_symbols"]) for row in rows),
                "min_covered_symbols": min(int(row["covered_symbols"]) for row in rows),
                "coverage_ratio": covered_days / active_days if active_days else 0.0,
            }
        )
    return out


def normalize_date(value: Any) -> str:
    raw = str(value or "").strip()
    if not raw:
        return ""
    parsed = pd.to_datetime(raw, utc=True, errors="coerce")
    if pd.isna(parsed):
        return ""
    return parsed.date().isoformat()


def write_markdown(path: Path, report: dict[str, Any]) -> None:
    s = report["summary"]
    reasons = ", ".join(s["reasons"]) if s["reasons"] else "none"
    lines = [
        "# PIT Universe Coverage Report\n\n",
        f"- Status: `{s['status']}`\n",
        f"- Reasons: {reasons}\n",
        f"- Benchmark: `{s['benchmark']}`\n",
        f"- PIT symbols: `{s['pit_symbol_count']}`\n",
        f"- Bars symbols: `{s['bars_symbol_count']}`\n",
        f"- Missing PIT symbols: `{s['missing_pit_symbol_count']}`\n",
        f"- Trading days: `{s['trading_days']}`\n",
        f"- Active member-days: `{s['active_member_days']}`\n",
        f"- Covered member-days: `{s['covered_member_days']}`\n",
        f"- Coverage ratio: `{s['coverage_ratio']:.4f}`\n",
        f"- Minimum active symbols: `{s['min_active_symbols']}`\n",
        f"- Minimum covered symbols: `{s['min_covered_symbols']}`\n\n",
        "## Yearly Coverage\n\n",
        "| Year | Trading Days | Mean Active | Mean Covered | Min Active | Min Covered | Coverage |\n",
        "|---|---:|---:|---:|---:|---:|---:|\n",
    ]
    for row in report["yearly"]:
        lines.append(
            f"| {row['year']} | {row['trading_days']} | {row['mean_active_symbols']:.1f} | "
            f"{row['mean_covered_symbols']:.1f} | {row['min_active_symbols']} | "
            f"{row['min_covered_symbols']} | {row['coverage_ratio']:.4f} |\n"
        )
    if report["missing_symbols"]:
        preview = ", ".join(report["missing_symbols"][:50])
        if len(report["missing_symbols"]) > 50:
            preview += ", ..."
        lines.extend(["\n## Missing Symbols\n\n", preview + "\n"])
    path.write_text("".join(lines), encoding="utf-8")


if __name__ == "__main__":
    main()
