#!/usr/bin/env python3
"""Audit concentration and attribution for ranker walk-forward artifacts."""

from __future__ import annotations

import argparse
import json
from pathlib import Path
from typing import Any

import pandas as pd

from artifact_manifest import command_line, file_sha256, git_sha, write_manifest
from portfolio_sleeve import write_csv, write_json


MEGA_CAP_WINNER_COHORT = {"AAPL", "AMZN", "AVGO", "GOOG", "GOOGL", "LLY", "META", "MSFT", "NVDA", "TSLA"}


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--fold-artifacts-root", required=True)
    parser.add_argument("--variant", required=True)
    parser.add_argument("--out-dir", required=True)
    parser.add_argument("--mega-cap-symbols", default=",".join(sorted(MEGA_CAP_WINNER_COHORT)))
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    root = Path(args.fold_artifacts_root) / args.variant
    if not root.exists():
        raise FileNotFoundError(f"variant artifact directory not found: {root}")
    mega_caps = parse_symbol_set(args.mega_cap_symbols)
    selections = load_fold_csvs(root, "selections.csv")
    decisions = load_fold_csvs(root, "decisions.csv")
    orders = load_fold_csvs(root, "orders.csv")
    summaries = load_fold_summaries(root)
    if selections.empty:
        raise ValueError("selection artifacts are empty")

    symbol_rows = symbol_attribution(selections, orders, mega_caps)
    fold_rows = fold_attribution(selections, decisions, summaries, mega_caps)
    decision_rows = decision_attribution(decisions, mega_caps)
    summary = aggregate_summary(selections, decisions, orders, summaries, symbol_rows, fold_rows, mega_caps)
    report = {
        "summary": summary,
        "symbol_attribution": symbol_rows,
        "fold_attribution": fold_rows,
        "decision_attribution": decision_rows,
        "manifest": {
            "command": command_line(),
            "git_sha": git_sha(),
            "fold_artifacts_root": str(args.fold_artifacts_root),
            "variant": args.variant,
            "variant_artifact_dir": str(root),
            "variant_artifact_dir_exists": root.exists(),
            "mega_cap_symbols": sorted(mega_caps),
        },
    }
    out_dir = Path(args.out_dir)
    out_dir.mkdir(parents=True, exist_ok=True)
    write_json(out_dir / "ranker_attribution_report.json", report)
    write_csv(out_dir / "ranker_attribution_symbols.csv", symbol_rows)
    write_csv(out_dir / "ranker_attribution_folds.csv", fold_rows)
    write_csv(out_dir / "ranker_attribution_decisions.csv", decision_rows)
    write_markdown(out_dir / "ranker_attribution_report.md", report)
    manifest = dict(report["manifest"])
    manifest["report_sha256"] = file_sha256(out_dir / "ranker_attribution_report.json")
    write_manifest(out_dir / "ranker_attribution_manifest.json", manifest)
    print(
        f"ranker attribution written: symbols={summary['unique_selected_symbols']} "
        f"top5_selection_share={summary['top5_selection_share']:.4f} "
        f"mega_cap_target_weight_share={summary['mega_cap_target_weight_share']:.4f}"
    )


def load_fold_csvs(root: Path, filename: str) -> pd.DataFrame:
    frames = []
    for fold_dir in sorted(path for path in root.iterdir() if path.is_dir()):
        path = fold_dir / filename
        if not path.exists():
            continue
        frame = pd.read_csv(path)
        frame["fold"] = fold_dir.name
        frames.append(frame)
    if not frames:
        return pd.DataFrame()
    return pd.concat(frames, ignore_index=True)


def load_fold_summaries(root: Path) -> pd.DataFrame:
    rows = []
    for fold_dir in sorted(path for path in root.iterdir() if path.is_dir()):
        path = fold_dir / "summary.json"
        if not path.exists():
            continue
        payload = json.loads(path.read_text(encoding="utf-8"))
        summary = payload.get("summary", {})
        row = {"fold": fold_dir.name}
        for key in [
            "total_return",
            "benchmark_return",
            "excess_return_vs_benchmark",
            "sharpe",
            "benchmark_sharpe",
            "max_drawdown",
            "benchmark_max_drawdown",
            "turnover",
            "num_rebalances",
            "num_orders",
            "num_alpha_symbols_traded",
            "mean_selected_forward_excess",
        ]:
            row[key] = finite_float(summary.get(key, 0.0))
        rows.append(row)
    return pd.DataFrame(rows)


def symbol_attribution(selections: pd.DataFrame, orders: pd.DataFrame, mega_caps: set[str]) -> list[dict[str, Any]]:
    frame = selections.copy()
    frame["symbol"] = frame["symbol"].astype(str).str.upper()
    frame["target_weight"] = pd.to_numeric(frame.get("target_weight", 0.0), errors="coerce").fillna(0.0)
    frame["label_excess"] = pd.to_numeric(frame.get("label_excess", 0.0), errors="coerce").fillna(0.0)
    frame["rank"] = pd.to_numeric(frame.get("rank", 0), errors="coerce").fillna(0.0)
    order_summary = {}
    if not orders.empty:
        orders = orders.copy()
        orders["symbol"] = orders["symbol"].astype(str).str.upper()
        orders["notional"] = pd.to_numeric(orders.get("notional", 0.0), errors="coerce").fillna(0.0).abs()
        orders["cost"] = pd.to_numeric(orders.get("cost", 0.0), errors="coerce").fillna(0.0)
        grouped_orders = orders[orders["symbol"] != "VOO"].groupby("symbol", sort=False)
        order_summary = {
            symbol: {
                "order_count": int(len(group)),
                "order_notional": float(group["notional"].sum()),
                "order_cost": float(group["cost"].sum()),
            }
            for symbol, group in grouped_orders
        }
    rows = []
    total_weight = float(frame["target_weight"].sum())
    total_rows = len(frame)
    for symbol, group in frame.groupby("symbol", sort=False):
        order = order_summary.get(symbol, {})
        target_sum = float(group["target_weight"].sum())
        rows.append(
            {
                "symbol": symbol,
                "selection_count": int(len(group)),
                "selection_share": len(group) / total_rows if total_rows else 0.0,
                "target_weight_sum": target_sum,
                "target_weight_share": target_sum / total_weight if total_weight else 0.0,
                "mean_target_weight": float(group["target_weight"].mean()),
                "mean_rank": float(group["rank"].mean()),
                "mean_label_excess": float(group["label_excess"].mean()),
                "positive_label_excess_rate": float((group["label_excess"] > 0).mean()),
                "fold_count": int(group["fold"].nunique()),
                "mega_cap_cohort": symbol in mega_caps,
                "order_count": int(order.get("order_count", 0)),
                "order_notional": float(order.get("order_notional", 0.0)),
                "order_cost": float(order.get("order_cost", 0.0)),
            }
        )
    return sorted(rows, key=lambda row: (row["target_weight_sum"], row["selection_count"]), reverse=True)


def fold_attribution(selections: pd.DataFrame, decisions: pd.DataFrame, summaries: pd.DataFrame, mega_caps: set[str]) -> list[dict[str, Any]]:
    rows = []
    summary_by_fold = {str(row["fold"]): row for _, row in summaries.iterrows()} if not summaries.empty else {}
    for fold, group in selections.groupby("fold", sort=True):
        group = group.copy()
        group["symbol"] = group["symbol"].astype(str).str.upper()
        group["target_weight"] = pd.to_numeric(group.get("target_weight", 0.0), errors="coerce").fillna(0.0)
        group["label_excess"] = pd.to_numeric(group.get("label_excess", 0.0), errors="coerce").fillna(0.0)
        total_weight = float(group["target_weight"].sum())
        mega_weight = float(group.loc[group["symbol"].isin(mega_caps), "target_weight"].sum())
        summary = summary_by_fold.get(str(fold), {})
        decisions_fold = decisions[decisions["fold"].astype(str) == str(fold)] if not decisions.empty else pd.DataFrame()
        rows.append(
            {
                "fold": str(fold),
                "selection_rows": int(len(group)),
                "selection_dates": int(group["decision_time"].nunique()) if "decision_time" in group else 0,
                "unique_selected_symbols": int(group["symbol"].nunique()),
                "mega_cap_selection_rows": int(group["symbol"].isin(mega_caps).sum()),
                "mega_cap_target_weight_share": mega_weight / total_weight if total_weight else 0.0,
                "mean_label_excess": float(group["label_excess"].mean()),
                "positive_label_excess_rate": float((group["label_excess"] > 0).mean()),
                "decision_rows": int(len(decisions_fold)),
                "mean_active_weight": finite_float(decisions_fold.get("active_weight", pd.Series(dtype=float)).mean()) if not decisions_fold.empty else 0.0,
                "total_return": finite_float(summary.get("total_return", 0.0)),
                "benchmark_return": finite_float(summary.get("benchmark_return", 0.0)),
                "excess_return": finite_float(summary.get("excess_return_vs_benchmark", 0.0)),
                "num_alpha_symbols_traded": int(finite_float(summary.get("num_alpha_symbols_traded", 0.0))),
            }
        )
    return rows


def decision_attribution(decisions: pd.DataFrame, mega_caps: set[str]) -> list[dict[str, Any]]:
    if decisions.empty:
        return []
    rows = []
    for _, row in decisions.iterrows():
        active = [symbol for symbol in str(row.get("active_symbols", "")).split(";") if symbol]
        active_upper = [symbol.upper() for symbol in active]
        mega = [symbol for symbol in active_upper if symbol in mega_caps]
        rows.append(
            {
                "fold": str(row.get("fold", "")),
                "decision_time": str(row.get("decision_time", "")),
                "execution_time": str(row.get("execution_time", "")),
                "active_symbols": ";".join(active_upper),
                "active_symbol_count": len(active_upper),
                "mega_cap_symbol_count": len(mega),
                "mega_cap_symbols": ";".join(mega),
                "benchmark_weight": finite_float(row.get("benchmark_weight", 0.0)),
                "active_weight": finite_float(row.get("active_weight", 0.0)),
                "turnover_distance": finite_float(row.get("turnover_distance", 0.0)),
                "execution_cost": finite_float(row.get("execution_cost", 0.0)),
            }
        )
    return rows


def aggregate_summary(
    selections: pd.DataFrame,
    decisions: pd.DataFrame,
    orders: pd.DataFrame,
    summaries: pd.DataFrame,
    symbol_rows: list[dict[str, Any]],
    fold_rows: list[dict[str, Any]],
    mega_caps: set[str],
) -> dict[str, Any]:
    target_sum = sum(row["target_weight_sum"] for row in symbol_rows)
    selection_total = sum(row["selection_count"] for row in symbol_rows)
    sorted_by_selection = sorted(symbol_rows, key=lambda row: row["selection_count"], reverse=True)
    sorted_by_weight = sorted(symbol_rows, key=lambda row: row["target_weight_sum"], reverse=True)
    mega_weight = sum(row["target_weight_sum"] for row in symbol_rows if row["mega_cap_cohort"])
    mega_selection = sum(row["selection_count"] for row in symbol_rows if row["mega_cap_cohort"])
    hhi = sum((row["target_weight_sum"] / target_sum) ** 2 for row in symbol_rows) if target_sum else 0.0
    compounded_strategy = compound(summaries["total_return"]) if not summaries.empty else 0.0
    compounded_benchmark = compound(summaries["benchmark_return"]) if not summaries.empty else 0.0
    return {
        "status": "diagnostic_only",
        "fold_count": int(summaries["fold"].nunique()) if not summaries.empty else int(selections["fold"].nunique()),
        "selection_rows": int(len(selections)),
        "decision_rows": int(len(decisions)),
        "order_rows": int(len(orders)),
        "unique_selected_symbols": int(selections["symbol"].astype(str).str.upper().nunique()),
        "top_symbol_by_selection": sorted_by_selection[0]["symbol"] if sorted_by_selection else "",
        "top_symbol_selection_share": sorted_by_selection[0]["selection_share"] if sorted_by_selection else 0.0,
        "top_symbol_by_weight": sorted_by_weight[0]["symbol"] if sorted_by_weight else "",
        "top_symbol_target_weight_share": sorted_by_weight[0]["target_weight_share"] if sorted_by_weight else 0.0,
        "top5_selection_share": sum(row["selection_count"] for row in sorted_by_selection[:5]) / selection_total if selection_total else 0.0,
        "top5_target_weight_share": sum(row["target_weight_sum"] for row in sorted_by_weight[:5]) / target_sum if target_sum else 0.0,
        "target_weight_hhi": hhi,
        "mega_cap_selection_share": mega_selection / selection_total if selection_total else 0.0,
        "mega_cap_target_weight_share": mega_weight / target_sum if target_sum else 0.0,
        "mega_cap_symbols": sorted(mega_caps),
        "mean_selected_forward_excess": float(pd.to_numeric(selections.get("label_excess", 0.0), errors="coerce").fillna(0.0).mean()),
        "positive_selected_forward_excess_rate": float((pd.to_numeric(selections.get("label_excess", 0.0), errors="coerce").fillna(0.0) > 0).mean()),
        "compounded_strategy_return": compounded_strategy,
        "compounded_benchmark_return": compounded_benchmark,
        "compounded_excess_return": (1 + compounded_strategy) - (1 + compounded_benchmark),
        "promotability_note": "diagnostic only; official promotion remains cmd/alpha-research DSR/PBO reports",
    }


def compound(values: pd.Series) -> float:
    growth = 1.0
    for value in pd.to_numeric(values, errors="coerce").fillna(0.0):
        growth *= 1.0 + float(value)
    return growth - 1.0


def parse_symbol_set(value: str) -> set[str]:
    return {part.strip().upper() for part in value.split(",") if part.strip()}


def finite_float(value: Any) -> float:
    try:
        out = float(value)
    except Exception:
        return 0.0
    if pd.isna(out):
        return 0.0
    return out


def write_markdown(path: Path, report: dict[str, Any]) -> None:
    s = report["summary"]
    lines = [
        "# Ranker Attribution Audit\n\n",
        "- Status: `diagnostic_only`\n",
        "- Promotion evidence remains the official `cmd/alpha-research` report; this file only audits concentration and selected-name behavior.\n",
        f"- Folds: `{s['fold_count']}`\n",
        f"- Selection rows: `{s['selection_rows']}`\n",
        f"- Decision rows: `{s['decision_rows']}`\n",
        f"- Unique selected symbols: `{s['unique_selected_symbols']}`\n",
        f"- Top symbol by selection: `{s['top_symbol_by_selection']}` ({s['top_symbol_selection_share']:.2%})\n",
        f"- Top symbol by target weight: `{s['top_symbol_by_weight']}` ({s['top_symbol_target_weight_share']:.2%})\n",
        f"- Top-5 selection share: `{s['top5_selection_share']:.2%}`\n",
        f"- Top-5 target-weight share: `{s['top5_target_weight_share']:.2%}`\n",
        f"- Target-weight HHI: `{s['target_weight_hhi']:.4f}`\n",
        f"- Mega-cap selection share: `{s['mega_cap_selection_share']:.2%}`\n",
        f"- Mega-cap target-weight share: `{s['mega_cap_target_weight_share']:.2%}`\n",
        f"- Mean selected forward excess: `{s['mean_selected_forward_excess']:.4f}`\n",
        f"- Positive selected forward-excess rate: `{s['positive_selected_forward_excess_rate']:.2%}`\n",
        f"- Compounded strategy return from fold summaries: `{s['compounded_strategy_return']*100:.2f}%`\n",
        f"- Compounded benchmark return from fold summaries: `{s['compounded_benchmark_return']*100:.2f}%`\n\n",
        "## Top Symbols By Target Weight\n\n",
        "| Symbol | Selections | Target Weight Share | Mean Rank | Mean Forward Excess | Positive Rate | Mega Cap |\n",
        "|---|---:|---:|---:|---:|---:|---|\n",
    ]
    for row in report["symbol_attribution"][:20]:
        lines.append(
            f"| {row['symbol']} | {row['selection_count']} | {row['target_weight_share']:.2%} | "
            f"{row['mean_rank']:.2f} | {row['mean_label_excess']:.4f} | "
            f"{row['positive_label_excess_rate']:.2%} | {row['mega_cap_cohort']} |\n"
        )
    lines.extend(
        [
            "\n## Fold Summary\n\n",
            "| Fold | Selections | Unique Symbols | Mega-Cap Weight Share | Mean Forward Excess | Excess Return |\n",
            "|---|---:|---:|---:|---:|---:|\n",
        ]
    )
    for row in report["fold_attribution"]:
        lines.append(
            f"| {row['fold']} | {row['selection_rows']} | {row['unique_selected_symbols']} | "
            f"{row['mega_cap_target_weight_share']:.2%} | {row['mean_label_excess']:.4f} | "
            f"{row['excess_return']*100:.2f}% |\n"
        )
    path.write_text("".join(lines), encoding="utf-8")


if __name__ == "__main__":
    main()
