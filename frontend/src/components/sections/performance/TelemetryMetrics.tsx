"use client";

import { useState, useMemo } from "react";
import { EquityCurveChart } from "@/components/EquityCurveChart";
import type { EquityPoint } from "@/lib/api";

const metricTabs = ["1W", "1M", "YTD"] as const;
type TimeframeTab = (typeof metricTabs)[number];

interface TelemetryMetricsProps {
  data: EquityPoint[];
  loading: boolean;
  currentReturnPct?: number;
  sharpeRatio?: number | null;
  maxDrawdown?: number | null;
}

export default function TelemetryMetrics({
  data,
  loading,
  sharpeRatio,
  maxDrawdown,
}: TelemetryMetricsProps) {
  const [activeTab, setActiveTab] = useState<TimeframeTab>("YTD");

  const { displayData, displayReturnPct } = useMemo(() => {
    if (!data || data.length === 0) {
      return { displayData: [], displayReturnPct: 0 };
    }

    const latestPoint = data[data.length - 1];
    const latestDate = new Date(latestPoint.time);
    let cutoffDate = new Date(latestDate);

    if (activeTab === "1W") {
      cutoffDate.setDate(latestDate.getDate() - 7);
    } else if (activeTab === "1M") {
      cutoffDate.setDate(latestDate.getDate() - 30);
    } else if (activeTab === "YTD") {
      cutoffDate = new Date(latestDate.getFullYear(), 0, 1);
    }

    const filtered = data.filter((d) => new Date(d.time) >= cutoffDate);
    const finalData = filtered.length > 1 ? filtered : data;

    const firstEquity = finalData[0].equity;
    const lastEquity = finalData[finalData.length - 1].equity;
    const pctReturn =
      firstEquity === 0 ? 0 : ((lastEquity - firstEquity) / firstEquity) * 100;

    return { displayData: finalData, displayReturnPct: pctReturn };
  }, [data, activeTab]);

  const isPositive = displayReturnPct >= 0;
  const normalizedSharpeRatio =
    typeof sharpeRatio === "number" && Number.isFinite(sharpeRatio)
      ? sharpeRatio
      : null;
  const normalizedMaxDrawdown =
    typeof maxDrawdown === "number" && Number.isFinite(maxDrawdown)
      ? maxDrawdown
      : null;
  const maxDrawdownPct =
    normalizedMaxDrawdown === null
      ? null
      : Math.abs(normalizedMaxDrawdown) * 100;
  const drawdownWidth = maxDrawdownPct
    ? `${Math.min(maxDrawdownPct, 100)}%`
    : "0%";

  return (
    <section className="w-full">
      <div className="flex items-center gap-2 mb-8">
        <span className="material-symbols-outlined text-sm text-on-surface-variant">
          monitoring
        </span>
        <span className="text-[10px] uppercase tracking-[0.2em] text-on-surface-variant font-medium">
          Real-Time Metrics
        </span>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 md:gap-8 items-start">
        <div className="lg:col-span-8 group relative flex flex-col h-[500px] lg:h-[540px] bg-surface-container-low border border-outline-variant/30 rounded-[32px] p-6 sm:p-8 overflow-hidden hover:bg-surface-container transition-colors duration-500 ease-out">
          <div className="absolute -top-32 -left-32 w-64 h-64 bg-primary-container/5 rounded-full blur-3xl pointer-events-none" />

          <div className="flex flex-col sm:flex-row justify-between items-start sm:items-end gap-6 mb-8 relative z-10">
            <div>
              <span className="text-[10px] font-medium tracking-[0.2em] text-on-surface-variant uppercase block mb-2">
                Cumulative P&L ({activeTab})
              </span>
              <span
                className={`text-4xl sm:text-5xl font-light tracking-tight transition-colors duration-500 ${isPositive ? "text-primary-container" : "text-error"}`}
              >
                {isPositive ? "+" : ""}
                {displayReturnPct.toFixed(2)}%
              </span>
            </div>

            <div className="flex p-1 bg-surface-container-highest/30 rounded-lg border border-outline-variant/20 backdrop-blur-md">
              {metricTabs.map((tab) => (
                <button
                  key={tab}
                  onClick={() => setActiveTab(tab)}
                  className={`px-4 py-1.5 text-xs font-medium tracking-wide rounded-md transition-all duration-300 ease-out ${
                    tab === activeTab
                      ? "bg-surface-container-low text-on-surface shadow-sm border border-outline-variant/30"
                      : "text-on-surface-variant hover:text-on-surface"
                  }`}
                >
                  {tab}
                </button>
              ))}
            </div>
          </div>

          <div className="w-full flex-grow min-h-0 relative overflow-hidden rounded-xl">
            {loading ? (
              <div className="absolute inset-0 flex flex-col items-center justify-center gap-4">
                <div className="w-8 h-8 border-2 border-outline-variant/30 border-t-primary-container rounded-full animate-spin" />
                <span className="text-xs font-light tracking-widest text-on-surface-variant uppercase">
                  Crunching Telemetry...
                </span>
              </div>
            ) : displayData.length > 0 ? (
              <div
                key={activeTab}
                className="absolute inset-0 animate-in fade-in duration-1000 ease-out"
              >
                <EquityCurveChart data={displayData} />
              </div>
            ) : (
              <div className="absolute inset-0 flex items-center justify-center text-sm font-light tracking-wide text-on-surface-variant bg-surface-container-highest/10 rounded-xl border border-dashed border-outline-variant/20">
                Awaiting Engine Initialization
              </div>
            )}
          </div>
        </div>

        <div className="lg:col-span-4 flex flex-col gap-6 self-start lg:mt-0">
          <div className="group relative flex flex-col justify-center bg-surface-container-low border border-outline-variant/30 rounded-[32px] p-6 lg:p-8 overflow-hidden hover:bg-surface-container transition-all duration-700 ease-[cubic-bezier(0.16,1,0.3,1)] hover:-translate-y-1 hover:shadow-[0_20px_40px_rgba(0,0,0,0.1)]">
            <div className="absolute top-0 inset-x-0 h-[1px] bg-gradient-to-r from-transparent via-primary-container/40 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-700" />

            <span className="text-[10px] font-medium tracking-[0.2em] text-on-surface-variant uppercase mb-4">
              Sharpe Ratio
            </span>
            <div className="flex items-baseline gap-3">
              <span className="text-4xl font-light tracking-tight text-on-surface">
                {normalizedSharpeRatio === null
                  ? "--"
                  : normalizedSharpeRatio.toFixed(2)}
              </span>
            </div>
          </div>

          <div className="group relative flex flex-col justify-center bg-surface-container-low border border-outline-variant/30 rounded-[32px] p-6 lg:p-8 overflow-hidden hover:bg-surface-container transition-all duration-700 ease-[cubic-bezier(0.16,1,0.3,1)] hover:-translate-y-1 hover:shadow-[0_20px_40px_rgba(0,0,0,0.1)]">
            <div className="absolute top-0 inset-x-0 h-[1px] bg-gradient-to-r from-transparent via-secondary-fixed/40 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-700" />

            <span className="text-[10px] font-medium tracking-[0.2em] text-on-surface-variant uppercase mb-4">
              Max Drawdown (Controlled)
            </span>
            <div className="flex items-baseline gap-3 mb-5">
              <span className="text-4xl font-light tracking-tight text-on-surface">
                {maxDrawdownPct === null
                  ? "--"
                  : `-${maxDrawdownPct.toFixed(2)}%`}
              </span>
            </div>

            <div className="w-full bg-outline-variant/20 h-[2px] rounded-full overflow-hidden relative">
              <div
                className="absolute left-0 top-0 h-full bg-secondary-fixed rounded-full"
                style={{ width: drawdownWidth }}
              />
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
