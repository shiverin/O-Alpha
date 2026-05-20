"use client";

import { FormEvent, useState } from "react";
import { EquityCurveChart } from "@/components/EquityCurveChart";
import { runBacktest, type BacktestResult } from "@/lib/api";

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border border-slate-700 bg-panel px-4 py-3">
      <p className="text-xs uppercase tracking-wide text-muted">{label}</p>
      <p className="mt-1 text-lg font-semibold text-white">{value}</p>
    </div>
  );
}

export default function HomePage() {
  const [symbol, setSymbol] = useState("AAPL");
  const [start, setStart] = useState("2023-01-01");
  const [end, setEnd] = useState("2024-12-31");
  const [fast, setFast] = useState(10);
  const [slow, setSlow] = useState(30);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [result, setResult] = useState<BacktestResult | null>(null);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError(null);
    try {
      const data = await runBacktest({
        symbol: symbol.toUpperCase(),
        start,
        end,
        fast_window: fast,
        slow_window: slow,
      });
      setResult(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Backtest failed");
      setResult(null);
    } finally {
      setLoading(false);
    }
  }

  const pct = (n: number) => `${(n * 100).toFixed(2)}%`;

  return (
    <main className="mx-auto min-h-screen max-w-5xl px-6 py-10">
      <header className="mb-10">
        <p className="text-sm font-medium text-accent">O(Alpha) · Milestone 1</p>
        <h1 className="mt-2 text-3xl font-bold tracking-tight">
          MA Crossover Backtest
        </h1>
        <p className="mt-2 max-w-2xl text-muted">
          Run a time-indexed backtest on OHLCV data stored in TimescaleDB.
          Signals at <em>t</em> execute at <em>t+1</em> open.
        </p>
      </header>

      <form
        onSubmit={onSubmit}
        className="mb-8 grid gap-4 rounded-xl border border-slate-700 bg-panel p-6 sm:grid-cols-2 lg:grid-cols-3"
      >
        <label className="flex flex-col gap-1 text-sm">
          Symbol
          <input
            className="rounded-md border border-slate-600 bg-surface px-3 py-2"
            value={symbol}
            onChange={(e) => setSymbol(e.target.value)}
          />
        </label>
        <label className="flex flex-col gap-1 text-sm">
          Start
          <input
            type="date"
            className="rounded-md border border-slate-600 bg-surface px-3 py-2"
            value={start}
            onChange={(e) => setStart(e.target.value)}
          />
        </label>
        <label className="flex flex-col gap-1 text-sm">
          End
          <input
            type="date"
            className="rounded-md border border-slate-600 bg-surface px-3 py-2"
            value={end}
            onChange={(e) => setEnd(e.target.value)}
          />
        </label>
        <label className="flex flex-col gap-1 text-sm">
          Fast MA
          <input
            type="number"
            min={2}
            className="rounded-md border border-slate-600 bg-surface px-3 py-2"
            value={fast}
            onChange={(e) => setFast(Number(e.target.value))}
          />
        </label>
        <label className="flex flex-col gap-1 text-sm">
          Slow MA
          <input
            type="number"
            min={3}
            className="rounded-md border border-slate-600 bg-surface px-3 py-2"
            value={slow}
            onChange={(e) => setSlow(Number(e.target.value))}
          />
        </label>
        <div className="flex items-end">
          <button
            type="submit"
            disabled={loading}
            className="w-full rounded-md bg-accent px-4 py-2 font-medium text-white hover:bg-blue-500 disabled:opacity-50"
          >
            {loading ? "Running…" : "Run Backtest"}
          </button>
        </div>
      </form>

      {error && (
        <div className="mb-6 rounded-lg border border-red-800 bg-red-950/50 px-4 py-3 text-red-200">
          {error}
        </div>
      )}

      {result && (
        <section className="space-y-6">
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            <Metric label="Sharpe" value={result.sharpe.toFixed(2)} />
            <Metric label="Sortino" value={result.sortino.toFixed(2)} />
            <Metric label="Max Drawdown" value={pct(result.max_drawdown)} />
            <Metric label="Total Return" value={pct(result.total_return)} />
          </div>
          <EquityCurveChart data={result.equity_curve} />
        </section>
      )}
    </main>
  );
}
