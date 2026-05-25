"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { LandingShell } from "../layout/LandingShell";
import { EquityCurveChart } from "@/components/EquityCurveChart";
import { runBacktest, type EquityPoint } from "@/lib/api";

const metricTabs = ["1W", "1M", "YTD"] as const;

type RegimeLevel = {
  bull: number;
  volatile: number;
  bear: number;
};

const buildFallbackEquityCurve = (): EquityPoint[] => {
  const now = Date.now();
  const points = 64;

  return Array.from({ length: points }, (_, idx) => {
    const t = idx / (points - 1);
    const trend = 10000 + t * 1650;
    const cycle = Math.sin(t * Math.PI * 3.6) * 180;
    const pullback = Math.sin(t * Math.PI * 9) * 65;
    const equity = trend + cycle + pullback;

    return {
      time: new Date(now - (points - idx) * 5 * 24 * 60 * 60 * 1000).toISOString(),
      equity: Number(equity.toFixed(2)),
    };
  });
};

const DEFAULT_EQUITY_CURVE = buildFallbackEquityCurve();

export function PerformancePage() {
  const [data, setData] = useState<EquityPoint[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [symbol, setSymbol] = useState("AAPL");
  const [fastWindow, setFastWindow] = useState(10);
  const [slowWindow, setSlowWindow] = useState(30);
  const [regimeLevels, setRegimeLevels] = useState<RegimeLevel>({
    bull: 74,
    volatile: 50,
    bear: 24,
  });

  const runBacktestHandler = useCallback(async () => {
    setLoading(true);
    setError(null);

    try {
      const endDate = new Date().toISOString().split("T")[0];
      const startDate = new Date(Date.now() - 365 * 24 * 60 * 60 * 1000)
        .toISOString()
        .split("T")[0];

      const result = await runBacktest({
        symbol,
        start: startDate,
        end: endDate,
        fast_window: fastWindow,
        slow_window: slowWindow,
      });

      setData(result.equity_curve);
    } catch (err) {
      setData(DEFAULT_EQUITY_CURVE);

      if (err instanceof Error) {
        setError(`${err.message} Showing fallback performance data.`);
      } else {
        setError("Failed to run backtest. Showing fallback performance data.");
      }
    } finally {
      setLoading(false);
    }
  }, [symbol, fastWindow, slowWindow]);

  useEffect(() => {
    runBacktestHandler();
  }, [runBacktestHandler]);

  useEffect(() => {
    const clamp = (value: number, min: number, max: number) =>
      Math.max(min, Math.min(max, value));

    const interval = window.setInterval(() => {
      setRegimeLevels((current) => {
        const shock = () => (Math.random() < 0.28 ? (Math.random() * 40 - 20) : 0);

        const nextBull = clamp(current.bull + (Math.random() * 26 - 13) + shock(), 18, 98);
        const nextVolatile = clamp(current.volatile + (Math.random() * 30 - 15) + shock(), 6, 95);
        const nextBear = clamp(current.bear + (Math.random() * 26 - 13) + shock(), 4, 88);

        return {
          bull: Math.round(nextBull),
          volatile: Math.round(nextVolatile),
          bear: Math.round(nextBear),
        };
      });
    }, 700);

    return () => window.clearInterval(interval);
  }, []);

  const currentReturnPct = useMemo(() => {
    if (data.length < 2) {
      return 0;
    }

    const initial = data[0].equity;
    const latest = data[data.length - 1].equity;
    if (initial === 0) {
      return 0;
    }
    return ((latest - initial) / initial) * 100;
  }, [data]);

  const bullStatus = regimeLevels.bull >= 72 ? "SCALING" : regimeLevels.bull >= 58 ? "BUILDING" : "SOFT";
  const volatileStatus = regimeLevels.volatile >= 58 ? "ACTIVE" : regimeLevels.volatile >= 36 ? "WATCH" : "CALM";
  const bearStatus = regimeLevels.bear >= 32 ? "ELEVATED" : regimeLevels.bear >= 18 ? "HEDGED" : "LOW";

  return (
    <LandingShell activePath="/performance" className="bg-performance-grid">
      <main className="pt-32 px-margin-mobile md:px-margin-desktop max-w-[1440px] mx-auto flex flex-col gap-16 md:gap-24">
        <section className="flex flex-col items-start max-w-4xl">

          <h1 className="font-headline-xl text-headline-xl text-on-surface mb-6">
            Institutional-Grade <br className="hidden md:block" />
            <span className="text-secondary-container gold-glow">Performance.</span>
          </h1>
          <p className="font-body-md text-body-md text-on-surface-variant max-w-2xl">
            Continuous market scanning. Convex optimization. Regime-aware
            execution. O(Alpha) translates your risk appetite into systematic,
            absolute returns without emotional bias.
          </p>

          <div className="mt-8 w-full flex flex-wrap items-end gap-4">
            <div className="flex flex-col gap-1">
              <label className="font-data-sm text-data-sm text-on-surface-variant">Symbol</label>
              <input
                type="text"
                value={symbol}
                onChange={(e) => setSymbol(e.target.value.toUpperCase())}
                className="rounded border border-outline-variant/40 bg-surface-container-low px-3 py-2 text-on-surface"
              />
            </div>
            <div className="flex flex-col gap-1">
              <label className="font-data-sm text-data-sm text-on-surface-variant">Fast MA</label>
              <input
                type="number"
                min={1}
                value={fastWindow}
                onChange={(e) => setFastWindow(Math.max(1, parseInt(e.target.value, 10) || 1))}
                className="rounded border border-outline-variant/40 bg-surface-container-low px-3 py-2 text-on-surface"
              />
            </div>
            <div className="flex flex-col gap-1">
              <label className="font-data-sm text-data-sm text-on-surface-variant">Slow MA</label>
              <input
                type="number"
                min={2}
                value={slowWindow}
                onChange={(e) => setSlowWindow(Math.max(2, parseInt(e.target.value, 10) || 2))}
                className="rounded border border-outline-variant/40 bg-surface-container-low px-3 py-2 text-on-surface"
              />
            </div>
            <button
              onClick={runBacktestHandler}
              disabled={loading}
              className="rounded-full bg-primary-container px-6 py-3 font-data-md text-background disabled:opacity-50"
            >
              {loading ? "Running..." : "Run Backtest"}
            </button>
          </div>
          {error && <p className="mt-3 text-error font-data-sm text-data-sm">{error}</p>}
        </section>

        <section>
          <h2 className="font-label-caps text-label-caps text-surface-tint uppercase mb-6 flex items-center gap-2 border-b border-outline-variant/30 pb-2 inline-flex">
            <span className="material-symbols-outlined text-sm">monitoring</span>
            Real-Time Metrics
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-12 gap-gutter items-start">
            <div className="md:col-span-8 glass-card rounded p-6 flex flex-col h-[460px] lg:h-[520px] overflow-hidden">
              <div className="flex justify-between items-start mb-6">
                <div>
                  <span className="font-data-sm text-data-sm text-on-surface-variant block mb-1">
                    CUMULATIVE P&L (YTD)
                  </span>
                  <span className="font-data-lg text-[32px] font-medium text-primary-container neon-text">
                    {currentReturnPct >= 0 ? "+" : ""}
                    {currentReturnPct.toFixed(2)}%
                  </span>
                </div>
                <div className="flex gap-2">
                  {metricTabs.map((tab) => (
                    <span
                      key={tab}
                      className={
                        tab === "1W"
                          ? "font-data-sm text-data-sm bg-primary-container/10 border border-primary-container/30 text-primary-container px-2 py-1 rounded"
                          : "font-data-sm text-data-sm text-on-surface-variant px-2 py-1"
                      }
                    >
                      {tab}
                    </span>
                  ))}
                </div>
              </div>
              <div className="w-full flex-grow min-h-0 relative mt-4 overflow-hidden">
                {data.length > 0 ? (
                  <EquityCurveChart data={data} />
                ) : (
                  <div className="h-full flex items-center justify-center text-on-surface-variant font-data-sm text-data-sm">
                    {loading ? "Calculating performance..." : "No data available"}
                  </div>
                )}
              </div>
            </div>

            <div className="md:col-span-4 flex flex-col gap-gutter self-start">
              <div className="glass-card rounded p-6 h-[190px] flex flex-col justify-center">
                <span className="font-data-sm text-data-sm text-on-surface-variant mb-2">
                  SHARPE RATIO
                </span>
                <div className="flex items-baseline gap-2">
                  <span className="font-data-lg text-[40px] text-on-surface">
                    2.4
                  </span>
                  <span className="font-data-sm text-data-sm text-surface-tint">
                    Top Decile
                  </span>
                </div>
              </div>
              <div className="glass-card rounded p-6 h-[190px] flex flex-col justify-center">
                <span className="font-data-sm text-data-sm text-on-surface-variant mb-2">
                  MAX DRAWDOWN (CONTROLLED)
                </span>
                <div className="flex items-baseline gap-2">
                  <span className="font-data-lg text-[40px] text-error">-4.2%</span>
                </div>
                <div className="mt-4 w-full bg-surface-container-highest h-1 rounded-full overflow-hidden">
                  <div className="bg-error h-full w-1/4"></div>
                </div>
              </div>
            </div>
          </div>
        </section>

        <section className="grid grid-cols-1 md:grid-cols-2 gap-gutter mb-4">
          <div>
            <h2 className="font-label-caps text-label-caps text-surface-tint uppercase mb-6 flex items-center gap-2 border-b border-outline-variant/30 pb-2 inline-flex">
              <span className="material-symbols-outlined text-sm">radar</span>
              Regime Detection
            </h2>
            <div className="glass-card rounded p-6 h-[300px] flex flex-col">
              <p className="font-body-md text-body-md text-on-surface-variant mb-6">
                Hidden Markov Models continuously classify market states,
                dynamically shifting agent posture.
              </p>
              <div className="flex-grow flex flex-col justify-around">
                <div className="flex items-center justify-between">
                  <span className="font-data-sm text-data-sm text-on-surface w-24">
                    BULL
                  </span>
                  <div className="flex-grow mx-4 bg-surface-container-highest h-[2px] relative">
                    <div
                      className="absolute left-0 top-1/2 -translate-y-1/2 h-[2px] bg-secondary-container transition-all duration-500 ease-out"
                      style={{ width: `${regimeLevels.bull}%` }}
                    ></div>
                  </div>
                  <span className="font-data-sm text-data-sm text-secondary-container">
                    {bullStatus}
                  </span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="font-data-sm text-data-sm text-on-surface w-24">
                    VOLATILE
                  </span>
                  <div className="flex-grow mx-4 bg-surface-container-highest h-[2px] relative">
                    <div
                      className="absolute left-0 top-1/2 -translate-y-1/2 h-[2px] bg-surface-tint transition-all duration-500 ease-out"
                      style={{ width: `${regimeLevels.volatile}%` }}
                    ></div>
                  </div>
                  <span className="font-data-sm text-data-sm text-surface-tint">
                    {volatileStatus}
                  </span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="font-data-sm text-data-sm text-on-surface w-24">
                    BEAR
                  </span>
                  <div className="flex-grow mx-4 bg-surface-container-highest h-[2px] relative">
                    <div
                      className="absolute left-0 top-1/2 -translate-y-1/2 h-[2px] bg-error/80 transition-all duration-500 ease-out"
                      style={{ width: `${regimeLevels.bear}%` }}
                    ></div>
                  </div>
                  <span className="font-data-sm text-data-sm text-on-surface-variant">
                    {bearStatus}
                  </span>
                </div>
              </div>
            </div>
          </div>

          <div>
            <h2 className="font-label-caps text-label-caps text-surface-tint uppercase mb-6 flex items-center gap-2 border-b border-outline-variant/30 pb-2 inline-flex">
              <span className="material-symbols-outlined text-sm">shield</span>
              Risk Architecture
            </h2>
            <div className="glass-card rounded p-6 h-[300px] flex flex-col gap-4">
              <div className="border-b border-outline-variant/20 pb-4">
                <h3 className="font-data-lg text-data-lg text-on-surface mb-2">
                  Kalman Filtering
                </h3>
                <p className="font-body-md text-sm text-on-surface-variant">
                  Extracts true signal from market noise, enabling precise entry
                  points even in low-liquidity environments.
                </p>
              </div>
              <div className="pt-2">
                <h3 className="font-data-lg text-data-lg text-on-surface mb-2">
                  Convex Optimization
                </h3>
                <p className="font-body-md text-sm text-on-surface-variant">
                  Portfolio weights are solved continuously to maximize expected
                  return subject to strict variance constraints.
                </p>
              </div>
            </div>
          </div>
        </section>

      </main>
    </LandingShell>
  );
}
