#!/usr/bin/env python3
"""Build a point-in-time constituent universe manifest.

This is a data-prep utility, not a backtester. It converts common historical
index-membership CSV shapes into the JSON manifest consumed by the Go daily
ranker.
"""

from __future__ import annotations

import argparse
import csv
import json
import re
from pathlib import Path
from typing import Any

import pandas as pd

from artifact_manifest import command_line, file_sha256, git_sha, write_manifest


SYMBOL_COLUMNS = ["symbol", "ticker", "constituent", "security", "code"]
START_COLUMNS = ["start", "start_date", "first_date", "from", "date_added", "added", "entry_date"]
END_COLUMNS = ["end", "end_date", "last_date", "to", "date_removed", "removed", "exit_date"]
DATE_COLUMNS = ["date", "as_of", "asof", "timestamp"]
LIST_COLUMNS = ["symbols", "tickers", "constituents", "components", "members"]


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--input", required=True, help="historical constituent CSV")
    parser.add_argument("--format", choices=["auto", "interval", "snapshot"], default="auto")
    parser.add_argument("--out", required=True, help="output JSON manifest path")
    parser.add_argument("--source-name", default="")
    parser.add_argument("--source-url", default="")
    parser.add_argument("--min-date", default="", help="optional inclusive YYYY-MM-DD filter")
    parser.add_argument("--max-date", default="", help="optional inclusive YYYY-MM-DD filter")
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    input_path = Path(args.input)
    rows = read_rows(input_path)
    if not rows:
        raise ValueError("input CSV is empty")
    mode = args.format
    if mode == "auto":
        mode = infer_format(rows[0])
    if mode == "interval":
        intervals = parse_interval_rows(rows)
    else:
        intervals = parse_snapshot_rows(rows)
    intervals = filter_intervals(intervals, args.min_date, args.max_date)
    if not intervals:
        raise ValueError("no intervals produced after filters")

    symbols = sorted({row["symbol"] for row in intervals})
    manifest = {
        "version": "pit_constituents_v1",
        "generated_at": pd.Timestamp.utcnow().isoformat(),
        "source": {
            "name": args.source_name,
            "url": args.source_url,
            "input": str(input_path),
            "input_sha256": file_sha256(input_path),
            "format": mode,
            "command": command_line(),
            "git_sha": git_sha(),
        },
        "symbols": symbols,
        "symbol_count": len(symbols),
        "interval_count": len(intervals),
        "date_range": {
            "start": min(row["start"] for row in intervals),
            "end": max(row.get("end") or row["start"] for row in intervals),
        },
        "intervals": intervals,
    }

    out_path = Path(args.out)
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(manifest, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    write_manifest(out_path.with_suffix(".manifest.json"), manifest["source"])
    print(
        f"pit universe written: {out_path} "
        f"symbols={manifest['symbol_count']} intervals={manifest['interval_count']}"
    )


def read_rows(path: Path) -> list[dict[str, str]]:
    with path.open(newline="", encoding="utf-8-sig") as handle:
        return list(csv.DictReader(handle))


def infer_format(first_row: dict[str, str]) -> str:
    columns = {normalize_column(column) for column in first_row}
    if any(column in columns for column in START_COLUMNS) and any(column in columns for column in SYMBOL_COLUMNS):
        return "interval"
    if any(column in columns for column in DATE_COLUMNS):
        return "snapshot"
    raise ValueError(f"could not infer constituent CSV format from columns: {sorted(first_row)}")


def parse_interval_rows(rows: list[dict[str, str]]) -> list[dict[str, str]]:
    symbol_col = find_column(rows[0], SYMBOL_COLUMNS)
    start_col = find_column(rows[0], START_COLUMNS)
    end_col = find_column(rows[0], END_COLUMNS, required=False)
    intervals: list[dict[str, str]] = []
    for row in rows:
        symbol = normalize_symbol(row.get(symbol_col, ""))
        start = normalize_date(row.get(start_col, ""))
        if not symbol or not start:
            continue
        out = {"symbol": symbol, "start": start}
        end = normalize_date(row.get(end_col, "")) if end_col else ""
        if end:
            out["end"] = end
        intervals.append(out)
    return sorted(intervals, key=lambda item: (item["symbol"], item["start"], item.get("end", "")))


def parse_snapshot_rows(rows: list[dict[str, str]]) -> list[dict[str, str]]:
    date_col = find_column(rows[0], DATE_COLUMNS)
    list_col = find_column(rows[0], LIST_COLUMNS, required=False)
    snapshots: list[tuple[str, set[str]]] = []
    for row in rows:
        as_of = normalize_date(row.get(date_col, ""))
        if not as_of:
            continue
        if list_col:
            symbols = parse_symbol_list(row.get(list_col, ""))
        else:
            symbols = {
                normalize_symbol(value)
                for column, value in row.items()
                if column != date_col and normalize_symbol(value)
            }
        if symbols:
            snapshots.append((as_of, symbols))
    snapshots.sort(key=lambda item: item[0])
    active_start: dict[str, str] = {}
    previous_date = ""
    intervals: list[dict[str, str]] = []
    for as_of, symbols in snapshots:
        current = set(symbols)
        previous = set(active_start)
        for symbol in sorted(current - previous):
            active_start[symbol] = as_of
        for symbol in sorted(previous - current):
            start = active_start.pop(symbol)
            out = {"symbol": symbol, "start": start}
            if previous_date:
                out["end"] = previous_date
            intervals.append(out)
        previous_date = as_of
    for symbol, start in sorted(active_start.items()):
        intervals.append({"symbol": symbol, "start": start})
    return sorted(intervals, key=lambda item: (item["symbol"], item["start"], item.get("end", "")))


def filter_intervals(intervals: list[dict[str, str]], min_date: str, max_date: str) -> list[dict[str, str]]:
    min_date = normalize_date(min_date)
    max_date = normalize_date(max_date)
    out = []
    for interval in intervals:
        start = interval["start"]
        end = interval.get("end", "")
        if max_date and start > max_date:
            continue
        if min_date and end and end < min_date:
            continue
        clipped = dict(interval)
        if min_date and clipped["start"] < min_date:
            clipped["start"] = min_date
        if max_date and (not clipped.get("end") or clipped["end"] > max_date):
            clipped["end"] = max_date
        out.append(clipped)
    return out


def find_column(row: dict[str, str], candidates: list[str], required: bool = True) -> str:
    by_normalized = {normalize_column(column): column for column in row}
    for candidate in candidates:
        if candidate in by_normalized:
            return by_normalized[candidate]
    if required:
        raise ValueError(f"missing one of columns {candidates}; got {sorted(row)}")
    return ""


def normalize_column(value: str) -> str:
    return re.sub(r"[^a-z0-9]+", "_", value.strip().lower()).strip("_")


def normalize_symbol(value: Any) -> str:
    symbol = str(value).strip().upper()
    symbol = symbol.replace(".", "-")
    if not symbol or symbol in {"NAN", "NONE", "NULL"}:
        return ""
    return symbol


def parse_symbol_list(value: str) -> set[str]:
    raw = str(value).strip()
    if not raw:
        return set()
    raw = raw.strip("[]{}")
    parts = re.split(r"[,;|\s]+", raw)
    return {symbol for symbol in (normalize_symbol(part.strip("'\"")) for part in parts) if symbol}


def normalize_date(value: Any) -> str:
    raw = str(value).strip()
    if not raw or raw.upper() in {"NAN", "NAT", "NONE", "NULL"}:
        return ""
    parsed = pd.to_datetime(raw, utc=True, errors="coerce")
    if pd.isna(parsed):
        return ""
    return parsed.date().isoformat()


if __name__ == "__main__":
    main()
