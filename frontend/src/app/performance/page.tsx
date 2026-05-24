"use client";

import { useEffect, useState } from "react";
import { runBacktest } from "@/lib/api";
import { EquityCurveChart } from "@/components/EquityCurveChart";
import { LandingShell } from "@/components/layout/LandingShell";

const metricTabs = ["1W", "1M", "YTD"] as const;

export default function PerformancePage() {
  const [data, setData] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [symbol, setSymbol] = useState("AAPL");
  const [fastWindow, setFastWindow] = useState(10);
  const [slowWindow, setSlowWindow] = useState(30);

  const runBacktestHandler = async () => {
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

      // Convert API response to format expected by EquityCurveChart
      const chartData = result.equity_curve.map((point: any) => ({
        time: point.time,
        equity: point.equity,
      }));

      setData(chartData);
    } catch (err: any) {
      setError(err.message || "Failed to run backtest");
      console.error("Backtest error:", err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    // Run initial backtest on mount
    runBacktestHandler();
  }, []);

  return (
    <LandingShell activePath="/performance" className="bg-performance-grid">
      <main className="pt-32 px-margin-mobile md:px-margin-desktop max-w-[1440px] mx-auto flex flex-col gap-16 md:gap-24">
        {/* Configuration Controls */}
        <section className="flex flex-col items-start max-w-4xl">
          <h1 className="font-headline-xl text-headline-xl text-on-surface mb-6">
            Interactive <br className="hidden md:block" />
            <span className="text-secondary-container">Performance</span>
          </h1>
          <p className="font-body-md text-body-md text-on-surface-variant max-w-2xl">
            Run live backtests with configurable parameters to analyze strategy performance.
          </p>

          <div className="w-full flex flex-wrap gap-4 items-end mb-6">
            <div className="flex flex-col">
              <label className="font-data-sm text-data-sm text-on-surface-variant mb-1">
                Symbol
              </label>
              <input
                type="text"
                value={symbol}
                onChange={(e) => setSymbol(e.target.value.toUpperCase())}
                className="w-full px-3 py-2 border border-input-bg rounded bg-surface"
                placeholder="AAPL"
              />
            </div>
            <div className="flex flex-col">
              <label className="font-data-sm text-data-sm text-on-surface-variant mb-1">
                Fast MA
              </label>
              <input
                type="number"
                value={fastWindow}
                onChange={(e) => setFastWindow(parseInt(e.target.value) || 10)}
                className="w-full px-3 py-2 border border-input-bg rounded bg-surface"
                min="1"
                placeholder="10"
              />
            </div>
            <div className="flex flex-col">
              <label className="font-data-sm text-data-sm text-on-surface-variant mb-1">
                Slow MA
              </label>
              <input
                type="number"
                value={slowWindow}
                onChange={(e) => setSlowWindow(parseInt(e.target.value) || 30)}
                className="w-full px-3 py-2 border border-input-bg rounded bg-surface"
                min="1"
                placeholder="30"
              />
            </div>
            <button
              onClick={runBacktestHandler}
              disabled={loading}
              className="px-6 py-3 bg-primary-container text-primary hover:bg-primary-container/90 rounded-lg font-data-md text-data-md transition-colors"
            >
              {loading ? "Running..." : "Run Backtest"}
            </button>
          </div>
        </section>

        {/* Error Display */}
        {error && (
          <div className="w-full bg-error/10 border border-error/20 rounded-lg p-4 mb-6">
            <span className="font-data-sm text-data-sm text-error">{error}</span>
          </div>
        )}

        {/* Loading State */}
        {loading && !data.length && (
          <div className="w-full flex items-center justify-center py-12">
            <span className="font-data-sm text-data-sm text-on-surface-variant animate-pulse">
              Calculating performance...
            </span>
          </div>
        )}

        {/* Results Section */}
        {data.length > 0 && (
          <>
            <section>
              <h2 className="font-label-caps text-label-caps text-surface-tint uppercase mb-6 flex items-center gap-2 border-b border-outline-variant/30 pb-2 inline-flex">
                <span className="material-symbols-outlined text-sm">monitoring</span>
                Equity Curve
              </h2>
              <div className="w-full">
                <EquityCurveChart data={data} />
              </div>
            </section>

            <section className="grid grid-cols-1 md:grid-cols-2 gap-gutter">
              <div>
                <h2 className="font-label-caps text-label-caps text-surface-tint uppercase mb-6 flex items-center gap-2 border-b border-outline-variant/30 pb-2 inline-flex">
                  <span className="material-symbols-outlined text-sm">monitoring</span>
                  Real-Time Metrics
                </h2>
                <div className="grid grid-cols-1 md:grid-cols-12 gap-gutter">
                  <div className="md:col-span-8 glass-card rounded p-6 flex flex-col justify-between h-[400px]">
                    <div className="flex justify-between items-start mb-6">
                      <div>
                        <span className="font-data-sm text-data-sm text-on-surface-variant block mb-1">
                          CUMULATIVE P&L (YTD)
                        </span>
                        <span className="font-data-lg text-[32px] font-medium text-primary-container">
                          {((data[data.length - 1]?.equity || 100000) / 100000 - 1) * 100 >= 0 ? "+" : ""}{((data[data.length - 1]?.equity || 100000) / 100000 - 1) * 100.toFixed(2)}%
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
                    <div className="w-full flex-grow relative mt-4">
                      <svg
                        className="absolute inset-0 w-full h-full"
                        preserveAspectRatio="none"
                        viewBox="0 0 100 50"
                      >
                        <defs>
                          <linearGradient id="chart-gradient" x1="0%" y1="0%" x2="0%" y2="100%">
                            <stop offset="0%" stopColor="rgba(0, 229, 255, 0.2)"></stop>
                            <stop offset="100%" stopColor="rgba(0, 229, 255, 0)"></stop>
                          </linearGradient>
                        </defs>
                        <line
                          x1="0"
                          y1="25"
                          x2="100"
                          y2="25"
                          stroke="rgba(255,255,255,0.05)"
                          strokeWidth="0.5"
                          strokeDasharray="1 2"
                        ></line>
                        <path
                          className="chart-area chart-area-animate"
                          d="M0,40 C10,38 20,45 30,30 C40,15 50,25 60,10 C70,-5 80,15 90,5 L100,0 L100,50 L0,50 Z"
                        ></path>
                        <path
                          className="chart-line chart-line-animate"
                          pathLength={100}
                          d="M0,40 C10,38 20,45 30,30 C40,15 50,25 60,10 C70,-5 80,15 90,5 L100,0"
                        ></path>
                      </svg>
                    </div>
                  </div>

                  <div className="md:col-span-4 flex flex-col gap-gutter">
                    <div className="glass-card rounded p-6 flex-1 flex flex-col justify-center">
                      <span className="font-data-sm text-data-sm text-on-surface-variant mb-2">
                        SHARPE RATIO
                      </span>
                      <div className="flex items-baseline gap-2">
                        <span className="font-data-lg text-[40px] text-on-surface">
                          {/* Would calculate from actual returns in a real implementation */}
                          2.4
                        </span>
                        <span className="font-data-sm text-data-sm text-surface-tint">
                          Top Decile
                        </span>
                      </div>
                    </div>
                    <div className="glass-card rounded p-6 flex-1 flex flex-col justify-center">
                      <span className="font-data-sm text-data-sm text-on-surface-variant mb-2">
                        MAX DRAWDOWN (CONTROLLED)
                      </span>
                      <div className="flex items-baseline gap-2">
                        <span className="font-data-lg text-[40px] text-error">
                          {/* Would calculate from actual equity curve in a real implementation */}
                          -4.2%
                        </span>
                      </div>
                      <div className="mt-4 w-full bg-surface-container-highest h-1 rounded-full overflow-hidden">
                        <div className="bg-error h-full w-1/4"></div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-gutter">
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
                          <div className="absolute left-0 top-1/2 -translate-y-1/2 w-3/4 h-[2px] bg-secondary-container"></div>
                        </div>
                        <span className="font-data-sm text-data-sm text-secondary-container">
                          SCALING
                        </span>
                      </div>
                      <div className="flex items-center justify-between">
                        <span className="font-data-sm text-data-sm text-on-surface w-24">
                          VOLATILE
                        </span>
                        <div className="flex-grow mx-4 bg-surface-container-highest h-[2px] relative">
                          <div className="absolute left-0 top-1/2 -translate-y-1/2 w-1/2 h-[2px] bg-surface-tint"></div>
                        </div>
                        <span className="font-data-sm text-data-sm text-surface-tint">
                          ACTIVE
                        </span>
                      </div>
                      <div className="flex items-center justify-between opacity-50">
                        <span className="font-data-sm text-data-sm text-on-surface w-24">
                          BEAR
                        </span>
                        <div className="flex-grow mx-4 bg-surface-container-highest h-[2px] relative"></div>
                        <span className="font-data-sm text-data-sm text-on-surface-variant">
                          HEDGED
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
              </div>
            </>
          </>
        )}
      </main>
    </LandingShell>
  );
}