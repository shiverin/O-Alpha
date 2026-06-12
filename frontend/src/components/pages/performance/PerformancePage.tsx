"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { LandingShell } from "../../layout/LandingShell";
import StrategySelector, {
  type StrategyConfig,
} from "@/components/StrategySelector";
import { runBacktest, type EquityPoint, type BacktestRequest } from "@/lib/api";
import { DEFAULT_EQUITY_CURVE } from "@/lib/mockData";
import { useRegimeSimulation } from "@/hooks/useRegimeSimulation";
import RegimeDetectionCard from "@/components/sections/performance/RegimeDetectionCard";
import RiskArchitectureCard from "@/components/sections/performance/RiskArchitectureCard";
import TelemetryMetrics from "@/components/sections/performance/TelemetryMetrics";

export function PerformancePage() {
  const [data, setData] = useState<EquityPoint[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [backtestMetrics, setBacktestMetrics] = useState<{
    sharpeRatio: number | null;
    maxDrawdown: number | null;
  }>({
    sharpeRatio: null,
    maxDrawdown: null,
  });

  const [symbol, setSymbol] = useState("AAPL");
  const [strategyConfig, setStrategyConfig] = useState<StrategyConfig | null>(
    null,
  );

  const { regimeLevels, bullStatus, volatileStatus, bearStatus } =
    useRegimeSimulation();

  const currentReturnPct = useMemo(() => {
    if (data.length < 2) return 0;
    const initial = data[0].equity;
    const latest = data[data.length - 1].equity;
    return initial === 0 ? 0 : ((latest - initial) / initial) * 100;
  }, [data]);

  const runBacktestHandler = useCallback(async () => {
    if (!strategyConfig) return;
    setLoading(true);
    setError(null);

    try {
      const endDate = new Date().toISOString();
      const startDate = new Date(
        Date.now() - 365 * 24 * 60 * 60 * 1000,
      ).toISOString();

      const payload: BacktestRequest = {
        symbol,
        start: startDate,
        end: endDate,
        strategy_type: strategyConfig.strategy,
        timeframe: "1Day",
        q_noise: strategyConfig.qNoise,
        r_noise: strategyConfig.rNoise,
        z_threshold: strategyConfig.zThresh,
        fast_period: strategyConfig.fastPeriod,
        slow_period: strategyConfig.slowPeriod,
      };

      const result = await runBacktest(payload);
      setData(result.equity_curve);
      setBacktestMetrics({
        sharpeRatio: result.sharpe ?? null,
        maxDrawdown: result.max_drawdown ?? null,
      });
    } catch (err) {
      setData(DEFAULT_EQUITY_CURVE);
      setBacktestMetrics({
        sharpeRatio: null,
        maxDrawdown: null,
      });
      setError(
        err instanceof Error
          ? `${err.message} Showing fallback performance data.`
          : "Failed to run backtest. Showing fallback performance data.",
      );
    } finally {
      setLoading(false);
    }
  }, [strategyConfig, symbol]);

  useEffect(() => {
    if (strategyConfig) runBacktestHandler();
  }, [strategyConfig, runBacktestHandler]);

  return (
    <LandingShell activePath="/performance" className="bg-performance-grid">
      <main className="pt-32 px-margin-mobile md:px-margin-desktop max-w-[1440px] mx-auto flex flex-col gap-16 md:gap-24">
        <section className="flex flex-col items-start max-w-4xl">
          <h1 className="font-headline-xl text-headline-xl text-on-surface mb-6">
            Institutional-Grade <br className="hidden md:block" />
            <span className="text-secondary-container gold-glow">
              Performance.
            </span>
          </h1>
          <p className="font-body-md text-body-md text-on-surface-variant max-w-2xl">
            Continuous market scanning. Convex optimization. Regime-aware
            execution. O(Alpha) translates your risk appetite into systematic,
            absolute returns without emotional bias.
          </p>

          <div className="mt-8 w-full flex flex-col gap-4">
            <div className="flex flex-col gap-1 w-48">
              <label className="font-data-sm text-data-sm text-on-surface-variant">
                Symbol
              </label>
              <input
                type="text"
                value={symbol}
                onChange={(e) => setSymbol(e.target.value.toUpperCase())}
                className="rounded border border-outline-variant/40 bg-surface-container-low px-3 py-2 text-on-surface"
              />
            </div>

            <StrategySelector onConfigChange={setStrategyConfig} />

            <button
              onClick={runBacktestHandler}
              disabled={loading || !strategyConfig}
              className="rounded-full bg-primary-container px-6 py-3 font-data-md text-background w-48 mt-4 shadow-[0_8px_20px_-12px_rgba(0,213,255,0.6)] hover:shadow-[0_10px_24px_-12px_rgba(0,213,255,0.7)] hover:bg-primary-fixed active:translate-y-[1px] transition-all duration-200 disabled:opacity-50 disabled:shadow-none disabled:hover:bg-primary-container"
            >
              {loading ? "Running..." : "Run Backtest"}
            </button>
            {error && (
              <p className="mt-3 text-error font-data-sm text-data-sm">
                {error}
              </p>
            )}
          </div>
        </section>

        <TelemetryMetrics
          data={data}
          loading={loading}
          currentReturnPct={currentReturnPct}
          sharpeRatio={backtestMetrics.sharpeRatio}
          maxDrawdown={backtestMetrics.maxDrawdown}
        />

        <section className="grid grid-cols-1 md:grid-cols-2 gap-gutter mb-4">
          <RegimeDetectionCard
            levels={regimeLevels}
            statuses={{ bullStatus, volatileStatus, bearStatus }}
          />
          <RiskArchitectureCard />
        </section>
      </main>
    </LandingShell>
  );
}
