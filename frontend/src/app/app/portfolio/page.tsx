"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import useSWR from "swr";
import { AppShell } from "@/components/app/AppShell";
import { Icon } from "@/components/ui/Icon";
import { useAuth } from "@/context/AuthContext";
import { api, streamPortfolioLive } from "@/lib/api";
import {
  applyLivePriceToHistory,
  applyLivePriceToPositions,
  applyLivePriceToSummary,
} from "@/lib/portfolioLiveState";
import {
  portfolioSummary as mockSummary,
  portfolioMetrics as mockMetrics,
  assetPositions as mockPositions,
} from "@/lib/mockAppData";

interface ServerPortfolioSummary {
  total_asset_value: number;
  change_percent_24h: number;
  change_dollar_24h: number;
  estimated_annual_yield: number;
  target_progress_percent: number;
  timestamp: string;
}

interface ServerPositionMetrics {
  symbol: string;
  qty: number;
  avg_entry_price: number;
  current_price: number;
  unrealized_pnl: number;
  exposure: number;
}

interface MockPositionMetrics {
  borderClass?: string;
  initials: string;
  name: string;
  symbol: string;
  category: string;
  allocation: number;
  currentPrice: number;
  isPositive: boolean;
  unrealizedPnL: number;
  exposure: number;
}

const PORTFOLIO_FLATLINE_Y = 70;
const DISPLAY_CENT_EPSILON = 0.005;
const PORTFOLIO_FLAT_SPARKLINE = {
  path: `M 0 ${PORTFOLIO_FLATLINE_Y} L 400 ${PORTFOLIO_FLATLINE_Y}`,
  lastPoint: { x: 400, y: PORTFOLIO_FLATLINE_Y },
};

const fetcher = <T,>(path: string): Promise<T> => api.get<T>(path);
const REALTIME_REFRESH_MS = 15000;
type PriceFlashDirection = "up" | "down";
const priceFlashStyles: Record<
  PriceFlashDirection,
  { backgroundColor: string; boxShadow: string }
> = {
  up: {
    backgroundColor: "rgba(16, 185, 129, 0.16)",
    boxShadow: "inset 4px 0 0 rgba(52, 211, 153, 0.95)",
  },
  down: {
    backgroundColor: "rgba(239, 68, 68, 0.16)",
    boxShadow: "inset 4px 0 0 rgba(239, 68, 68, 0.95)",
  },
};

export default function PortfolioPage() {
  const { user } = useAuth();
  const currentUserID = user?.id || 999;
  const livePositionsRef = useRef<ServerPositionMetrics[] | undefined>();
  const flashTimeoutsRef = useRef<
    Partial<Record<string, ReturnType<typeof setTimeout>>>
  >({});
  const [priceFlashes, setPriceFlashes] = useState<
    Partial<Record<string, PriceFlashDirection>>
  >({});

  const { data: serverSummary, mutate: mutateSummary } =
    useSWR<ServerPortfolioSummary>(
      currentUserID !== 999 ? "/api/v1/user/portfolio/summary" : null,
      fetcher,
      { refreshInterval: REALTIME_REFRESH_MS },
    );

  const { data: serverPositions, mutate: mutatePositions } = useSWR<
    ServerPositionMetrics[]
  >(
    currentUserID !== 999 ? "/api/v1/user/portfolio/positions" : null,
    fetcher,
    { refreshInterval: REALTIME_REFRESH_MS },
  );

  const { data: serverHistory, mutate: mutateHistory } = useSWR<
    ServerPortfolioSummary[]
  >(
    currentUserID !== 999 ? "/api/v1/user/portfolio/history?limit=30" : null,
    fetcher,
    { refreshInterval: REALTIME_REFRESH_MS },
  );

  useEffect(() => {
    if (serverPositions) {
      livePositionsRef.current = serverPositions;
    }
  }, [serverPositions]);

  useEffect(() => {
    const timeouts = flashTimeoutsRef.current;
    return () => {
      Object.values(timeouts).forEach((timeout) => {
        if (timeout) {
          clearTimeout(timeout);
        }
      });
    };
  }, []);

  const flashPositionRow = useCallback(
    (symbol: string, direction: PriceFlashDirection) => {
      const key = symbol.toUpperCase();
      const existingTimeout = flashTimeoutsRef.current[key];
      if (existingTimeout) {
        clearTimeout(existingTimeout);
      }
      setPriceFlashes((current) => ({ ...current, [key]: direction }));
      flashTimeoutsRef.current[key] = setTimeout(() => {
        setPriceFlashes((current) => {
          const next = { ...current };
          delete next[key];
          return next;
        });
        delete flashTimeoutsRef.current[key];
      }, 650);
    },
    [],
  );

  useEffect(() => {
    if (currentUserID === 999) return;

    const controller = new AbortController();
    streamPortfolioLive((event) => {
      if (event.type === "snapshot") {
        if (event.summary) {
          void mutateSummary(event.summary, false);
        }
        if (event.positions) {
          livePositionsRef.current = event.positions;
          void mutatePositions(event.positions, false);
        }
        if (event.history) {
          void mutateHistory(event.history, false);
        }
        return;
      }
      if (event.type === "price") {
        const symbol = event.symbol.toUpperCase();
        const previousPosition = livePositionsRef.current?.find(
          (position) => position.symbol.toUpperCase() === symbol,
        );
        const priceDelta = previousPosition
          ? event.price - previousPosition.current_price
          : 0;
        if (Math.abs(priceDelta) > DISPLAY_CENT_EPSILON) {
          flashPositionRow(symbol, priceDelta > 0 ? "up" : "down");
        }

        const { positions, deltaExposure } = applyLivePriceToPositions(
          livePositionsRef.current,
          event,
        );
        livePositionsRef.current = positions;
        void mutatePositions(positions, false);
        void mutateSummary(
          (summary) =>
            applyLivePriceToSummary(summary, deltaExposure, event.timestamp),
          false,
        );
        void mutateHistory(
          (history) =>
            applyLivePriceToHistory(history, deltaExposure, event.timestamp),
          false,
        );
      }
    }, controller.signal).catch((err) => {
      if (err instanceof DOMException && err.name === "AbortError") return;
      console.error("Portfolio live stream failed:", err);
    });

    return () => controller.abort();
  }, [
    currentUserID,
    flashPositionRow,
    mutateHistory,
    mutatePositions,
    mutateSummary,
  ]);

  const totalAssetValue =
    currentUserID === 999 || !serverSummary
      ? mockSummary.totalAssetValue
      : serverSummary.total_asset_value;
  const changePercent24h =
    currentUserID === 999 || !serverSummary
      ? mockSummary.changePercent24h
      : serverSummary.change_percent_24h;
  const changeDollar24h =
    currentUserID === 999 || !serverSummary
      ? mockSummary.changeDollar24h
      : serverSummary.change_dollar_24h;

  const sparkline = useMemo(() => {
    if (!serverHistory || serverHistory.length < 2) {
      return PORTFOLIO_FLAT_SPARKLINE;
    }

    const orderedHistory = [...serverHistory]
      .filter((snapshot) => Number.isFinite(snapshot.total_asset_value))
      .sort(
        (a, b) =>
          new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime(),
      );
    if (orderedHistory.length < 2) {
      return PORTFOLIO_FLAT_SPARKLINE;
    }

    const values = orderedHistory.map((snapshot) => snapshot.total_asset_value);
    const minVal = Math.min(...values);
    const maxVal = Math.max(...values);
    const valRange = maxVal - minVal;

    if (valRange <= DISPLAY_CENT_EPSILON) {
      return PORTFOLIO_FLAT_SPARKLINE;
    }

    const points = orderedHistory.map((snapshot, index) => {
      const x = (index / (orderedHistory.length - 1)) * 400;
      const y = 85 - ((snapshot.total_asset_value - minVal) / valRange) * 70;
      return { x, y };
    });

    const path = points.reduce(
      (acc, point, index) =>
        index === 0
          ? `M ${point.x} ${point.y}`
          : `${acc} L ${point.x} ${point.y}`,
      "",
    );

    return {
      path,
      lastPoint: points[points.length - 1],
    };
  }, [serverHistory]);

  const progressPercent =
    currentUserID === 999 || !serverSummary
      ? 0
      : serverSummary.target_progress_percent;
  const progressBarWidth = Math.min(Math.max(progressPercent, 0), 100);
  const isPositiveChange = changeDollar24h > DISPLAY_CENT_EPSILON;
  const isNegativeChange = changeDollar24h < -DISPLAY_CENT_EPSILON;

  const totalPositionsExposure = useMemo(() => {
    if (!serverPositions) return 0;
    return serverPositions.reduce(
      (acc, pos) => acc + pos.qty * pos.current_price,
      0,
    );
  }, [serverPositions]);

  const cashBalance = useMemo(() => {
    const diff = totalAssetValue - totalPositionsExposure;
    return diff > 0 ? diff : 0;
  }, [totalAssetValue, totalPositionsExposure]);

  const compositionSegmentsList = useMemo(() => {
    if (currentUserID === 999) {
      return [
        {
          label: "Equities",
          percentage: 60,
          color: "#00dbe9",
          glowClass:
            "bg-primary-fixed-dim shadow-[0_0_8px_rgba(0,219,233,0.5)]",
          dashOffset: 251,
          rotation: 0,
        },
        {
          label: "Crypto Assets",
          percentage: 30,
          color: "#e9c400",
          glowClass:
            "bg-secondary-fixed-dim shadow-[0_0_8px_rgba(233,196,0,0.3)]",
          dashOffset: 439,
          rotation: 216,
        },
        {
          label: "Cash & Equiv",
          percentage: 10,
          color: "#849495",
          glowClass: "bg-outline",
          dashOffset: 565,
          rotation: 324,
        },
      ];
    }

    const colors = ["#00dbe9", "#ffd34d", "#6fe6ff", "#b9f1ff"];
    const glowClasses = [
      "bg-primary-fixed-dim shadow-[0_0_8px_rgba(0,219,233,0.5)]",
      "bg-secondary-fixed-dim shadow-[0_0_8px_rgba(233,196,0,0.3)]",
      "bg-primary-container shadow-[0_0_8px_rgba(0,213,255,0.4)]",
      "bg-secondary-fixed shadow-[0_0_8px_rgba(255,211,77,0.4)]",
    ];

    const rankedPositions = (serverPositions ?? [])
      .map((pos) => ({
        label: pos.symbol,
        exposure: pos.qty * pos.current_price,
      }))
      .filter((pos) => pos.exposure > 0)
      .sort((a, b) => b.exposure - a.exposure);
    const visiblePositions = rankedPositions.slice(0, 4);
    const otherExposure = rankedPositions
      .slice(4)
      .reduce((sum, position) => sum + position.exposure, 0);

    const segments = visiblePositions.map((position, idx) => ({
      label: position.label,
      exposure: position.exposure,
      color: colors[idx % colors.length],
      glowClass: glowClasses[idx % glowClasses.length],
    }));

    if (otherExposure > 0) {
      segments.push({
        label: "Other",
        exposure: otherExposure,
        color: "#94a3b8",
        glowClass: "bg-slate-400",
      });
    }

    if (cashBalance > 0 || segments.length === 0) {
      segments.push({
        label: "Cash & Equiv",
        exposure: segments.length === 0 ? totalAssetValue : cashBalance,
        color: "#849495",
        glowClass: "bg-outline",
      });
    }

    const normalizedSegments = normalizeCompositionPercentages(segments).sort(
      (a, b) => b.percentage - a.percentage,
    );
    let currentRotation = 0;
    return normalizedSegments.map((seg) => {
      const dashOffset = 628 - (628 * seg.percentage) / 100;
      const rotation = currentRotation;
      currentRotation += seg.percentage * 3.6;
      return { ...seg, dashOffset, rotation };
    });
  }, [currentUserID, serverPositions, totalAssetValue, cashBalance]);

  const topPerformerMetrics = useMemo(() => {
    if (currentUserID === 999) {
      return { symbol: "NVDA", contribution: 12.4 };
    }
    if (!serverPositions || serverPositions.length === 0) {
      return { symbol: "None (All Cash)", contribution: 0.0 };
    }
    const sorted = [...serverPositions].sort(
      (a, b) => b.unrealized_pnl - a.unrealized_pnl,
    );
    const top = sorted[0];
    const contributionPct =
      totalAssetValue > 0 ? (top.unrealized_pnl / totalAssetValue) * 100 : 0;
    return {
      symbol: top.symbol,
      contribution: parseFloat(contributionPct.toFixed(2)),
    };
  }, [currentUserID, serverPositions, totalAssetValue]);

  const ringCenterLabel = compositionSegmentsList[0]?.label || "Cash & Equiv";
  const ringCenterPercentage = compositionSegmentsList[0]?.percentage || 100;

  return (
    <AppShell title="Portfolio">
      <div className="w-full max-w-full min-w-0 bg-transparent flex flex-col gap-6 md:gap-10 animate-in fade-in duration-700 ease-[cubic-bezier(0.16,1,0.3,1)] overflow-hidden">
        <section className="w-full min-w-0 max-w-full">
          <div className="group relative bg-surface-container-low border border-outline-variant/30 rounded-[32px] p-5 sm:p-8 xl:p-10 overflow-hidden transition-all duration-700 ease-[cubic-bezier(0.16,1,0.3,1)] hover:shadow-[0_20px_40px_rgba(0,0,0,0.3)]">
            <div className="absolute top-0 inset-x-0 h-[1px] bg-gradient-to-r from-transparent via-primary-container/30 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-700" />
            <div
              className="absolute inset-0 opacity-[0.12] pointer-events-none transition-transform duration-1000 group-hover:scale-105"
              style={{
                backgroundImage:
                  "radial-gradient(circle at 2px 2px, rgba(255,255,255,0.15) 1px, transparent 0)",
                backgroundSize: "32px 32px",
              }}
            />

            <div className="w-full min-w-0 flex flex-col xl:flex-row justify-between items-start xl:items-end gap-6 xl:gap-8 relative z-10">
              <div className="w-full xl:w-auto min-w-0">
                <span className="text-[10px] font-medium tracking-[0.2em] text-on-surface-variant/70 uppercase mb-2 block">
                  Total Asset Value
                </span>
                <h2 className="text-[clamp(1.5rem,6vw,3.75rem)] font-light tracking-tight text-on-surface whitespace-nowrap">
                  $
                  {
                    totalAssetValue
                      .toLocaleString(undefined, {
                        minimumFractionDigits: 2,
                        maximumFractionDigits: 2,
                      })
                      .split(".")[0]
                  }
                  <span className="text-on-surface-variant/30">
                    .{totalAssetValue.toFixed(2).split(".")[1]}
                  </span>
                </h2>

                <div className="flex flex-wrap items-center gap-2.5 mt-3 sm:mt-4">
                  <span
                    className={`font-mono text-[11px] px-2.5 py-0.5 rounded-full flex items-center gap-1 border ${
                      isPositiveChange
                        ? "bg-primary-fixed-dim/10 text-primary-fixed-dim border-primary-fixed-dim/20 shadow-[0_0_12px_rgba(0,240,255,0.15)]"
                        : isNegativeChange
                          ? "bg-error/10 text-error border-error/20"
                          : "bg-white/[0.02] text-on-surface-variant/70 border-outline-variant/20"
                    }`}
                  >
                    <span className="material-symbols-outlined text-[12px]">
                      {isPositiveChange
                        ? "arrow_upward"
                        : isNegativeChange
                          ? "arrow_downward"
                          : "remove"}
                    </span>
                    {changePercent24h > 0 ? "+" : ""}
                    {changePercent24h}% (24h)
                  </span>
                  <span className="text-on-surface-variant/50 font-mono text-[11px] whitespace-nowrap">
                    {changeDollar24h > 0 ? "+" : ""}$
                    {changeDollar24h.toLocaleString()}
                  </span>
                </div>
              </div>

              <div className="w-full xl:w-1/2 max-w-md xl:max-w-none h-24 sm:h-28 relative border-b border-outline-variant/10 mt-2 xl:mt-0 min-w-0">
                <svg
                  className="w-full h-full"
                  preserveAspectRatio="none"
                  viewBox="0 0 400 100"
                >
                  <defs>
                    <linearGradient id="cyan-fade" x1="0" x2="0" y1="0" y2="1">
                      <stop
                        offset="0%"
                        stopColor="rgba(0, 219, 233, 0.15)"
                      ></stop>
                      <stop
                        offset="100%"
                        stopColor="rgba(0, 219, 233, 0)"
                      ></stop>
                    </linearGradient>
                  </defs>
                  <path
                    d={`${sparkline.path} L 400 100 L 0 100 Z`}
                    fill="url(#cyan-fade)"
                  ></path>
                  <path
                    d={sparkline.path}
                    fill="none"
                    stroke="#00dbe9"
                    strokeWidth="1.5"
                    style={{
                      filter: "drop-shadow(0 0 6px rgba(0,219,233,0.4))",
                    }}
                  ></path>
                  <circle
                    cx={sparkline.lastPoint.x}
                    cy={sparkline.lastPoint.y}
                    fill="#00dbe9"
                    r="3.5"
                    style={{ filter: "drop-shadow(0 0 6px #00dbe9)" }}
                  ></circle>
                </svg>
                <div className="absolute bottom-0 inset-x-0 flex justify-between text-[9px] font-medium uppercase tracking-widest text-on-surface-variant/30 pt-1.5 border-t border-outline-variant/10">
                  <span>24h Ago</span>
                  <span>Now</span>
                </div>
              </div>
            </div>
          </div>
        </section>

        <section className="w-full min-w-0 max-w-full grid grid-cols-1 xl:grid-cols-12 gap-6 md:gap-8 items-start">
          <div className="w-full min-w-0 md:col-span-12 xl:col-span-5 bg-surface-container-low border border-outline-variant/30 rounded-[32px] p-5 sm:p-8 flex flex-col justify-between h-auto min-h-[380px] sm:min-h-[420px] xl:min-h-[440px]">
            <h3 className="text-xs font-light tracking-wide text-on-surface border-b border-outline-variant/20 pb-4 mb-2">
              Composition
            </h3>

            <div className="relative flex-grow flex items-center justify-center min-h-[200px] sm:min-h-[240px] my-4 w-full">
              <svg
                className="w-full max-w-[240px] h-auto transform -rotate-90 scale-90 sm:scale-100 transition-transform duration-500"
                viewBox="0 0 240 240"
              >
                <circle
                  cx="120"
                  cy="120"
                  fill="none"
                  r="100"
                  stroke="#222222"
                  strokeWidth="16"
                ></circle>
                {compositionSegmentsList.map((segment, idx) => (
                  <circle
                    key={idx}
                    cx="120"
                    cy="120"
                    fill="none"
                    r="100"
                    stroke={segment.color}
                    strokeDasharray="628"
                    strokeDashoffset={segment.dashOffset}
                    strokeWidth="16"
                    strokeLinecap="round"
                    transform={`rotate(${segment.rotation} 120 120)`}
                    style={{
                      filter: `drop-shadow(0 0 6px ${segment.color}40)`,
                    }}
                  />
                ))}
              </svg>
              <div className="absolute inset-0 flex flex-col items-center justify-center pointer-events-none text-center">
                <span className="text-[9px] font-medium tracking-[0.2em] text-on-surface-variant/50 uppercase">
                  {ringCenterLabel}
                </span>
                <span className="text-2xl sm:text-3xl font-light text-on-surface mt-0.5">
                  {ringCenterPercentage}%
                </span>
              </div>
            </div>

            <div className="flex flex-col gap-3 font-mono text-[11px] sm:text-xs border-t border-outline-variant/10 pt-4 mt-2 w-full">
              {compositionSegmentsList.map((segment, idx) => (
                <div
                  key={idx}
                  className="flex justify-between items-center group cursor-default"
                >
                  <div className="flex items-center gap-3 min-w-0">
                    <div
                      className={`w-2 h-2 sm:w-2.5 sm:h-2.5 rounded-full shrink-0 transition-transform duration-300 group-hover:scale-110 ${segment.glowClass}`}
                    />
                    <span className="text-on-surface-variant/80 group-hover:text-on-surface transition-colors truncate">
                      {segment.label}
                    </span>
                  </div>
                  <span className="text-on-surface/40 group-hover:text-on-surface transition-colors shrink-0">
                    ({segment.percentage}%)
                  </span>
                </div>
              ))}
            </div>
          </div>

          <div className="w-full min-w-0 max-w-full md:col-span-12 xl:col-span-7 grid grid-cols-1 sm:grid-cols-2 gap-6 md:gap-8 items-start">
            <div className="w-full min-w-0 bg-surface-container-low border border-outline-variant/30 rounded-[32px] p-5 sm:p-6 min-h-[180px] sm:min-h-[200px] flex flex-col justify-between hover:bg-surface-container transition-colors duration-300">
              <div className="flex justify-between items-start mb-4">
                <div className="text-primary-fixed-dim">
                  <Icon name="trending_up" />
                </div>
                <span className="bg-white/[0.02] border border-outline-variant/20 px-2.5 py-0.5 rounded-full font-mono text-[9px] tracking-wider text-on-surface-variant/60 uppercase shrink-0">
                  Alpha Gen
                </span>
              </div>
              <div>
                <p className="font-mono text-[10px] tracking-wider text-on-surface-variant/50 uppercase mb-1">
                  Top Performer
                </p>
                <p className="text-xl sm:text-2xl font-light text-on-surface truncate">
                  {topPerformerMetrics.symbol}
                </p>
                <p className="font-mono text-xs text-primary-fixed-dim mt-2 truncate">
                  {topPerformerMetrics.contribution >= 0 ? "+" : ""}
                  {topPerformerMetrics.contribution}%{" "}
                  <span className="text-on-surface-variant/40">
                    contribution
                  </span>
                </p>
              </div>
            </div>

            <div className="w-full min-w-0 bg-surface-container-low border border-outline-variant/30 rounded-[32px] p-5 sm:p-6 min-h-[180px] sm:min-h-[200px] flex flex-col justify-between hover:bg-surface-container transition-colors duration-300 relative overflow-hidden">
              <div className="absolute -right-4 -top-4 w-24 h-24 bg-secondary-fixed-dim opacity-[0.03] rounded-full blur-2xl pointer-events-none" />
              <div className="flex justify-between items-start mb-4">
                <div className="text-secondary-fixed-dim">
                  <Icon name="shield" />
                </div>
                <span className="bg-white/[0.02] border border-outline-variant/20 px-2.5 py-0.5 rounded-full font-mono text-[9px] tracking-wider text-on-surface-variant/60 uppercase shrink-0">
                  Metrics
                </span>
              </div>
              <div>
                <p className="font-mono text-[10px] tracking-wider text-on-surface-variant/50 uppercase mb-1">
                  Risk Profile
                </p>
                <p className="text-xl sm:text-2xl font-light text-on-surface truncate">
                  {currentUserID === 999
                    ? mockMetrics.riskProfile.label
                    : "Moderate"}
                </p>
                <p className="font-mono text-xs text-on-surface-variant/70 mt-2 truncate">
                  Sharpe Ratio:{" "}
                  <span className="text-on-surface font-medium">
                    {currentUserID === 999
                      ? mockMetrics.riskProfile.sharpeRatio
                      : "--"}
                  </span>
                </p>
              </div>
            </div>

            <div className="w-full min-w-0 sm:col-span-2 bg-surface-container-low border border-outline-variant/30 rounded-[32px] p-5 sm:p-6 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 hover:bg-surface-container transition-colors duration-300">
              <div className="min-w-0">
                <p className="font-mono text-[10px] tracking-wider text-on-surface-variant/50 uppercase mb-1">
                  Annual Yield
                </p>
                <p className="text-2xl sm:text-3xl font-light text-on-surface truncate">
                  Pending
                </p>
                <p className="font-mono text-[10px] text-on-surface-variant/40 mt-1">
                  Requires computed trailing return history
                </p>
              </div>
              <div className="w-full sm:w-auto text-left sm:text-right shrink-0">
                <p className="font-mono text-[10px] text-primary-fixed-dim tracking-wider uppercase mb-1.5 sm:mb-2">
                  Target Met
                </p>
                <div className="w-full sm:w-36 h-[3px] bg-surface-container-highest rounded-full overflow-hidden">
                  <div
                    className="h-full bg-primary-fixed-dim shadow-[0_0_10px_rgba(0,219,233,0.8)] rounded-full"
                    style={{ width: `${progressBarWidth}%` }}
                  />
                </div>
              </div>
            </div>
          </div>
        </section>

        <section className="w-full min-w-0 max-w-full bg-surface-container-low border border-outline-variant/30 rounded-[32px] pb-4">
          <div className="p-5 sm:p-6 border-b border-outline-variant/20 flex justify-between items-center bg-white/[0.01]">
            <h3 className="text-sm sm:text-md font-light tracking-wide text-on-surface">
              Positions
            </h3>
            <button className="font-mono text-[11px] text-primary-fixed-dim hover:text-primary-fixed flex items-center gap-1.5 transition-colors duration-300 shrink-0">
              <span className="material-symbols-outlined text-[16px]">
                download
              </span>
              Export
            </button>
          </div>

          <div className="overflow-x-auto w-full max-w-full">
            <table className="w-full text-left border-collapse min-w-[700px] table-fixed">
              <thead>
                <tr className="font-mono text-[10px] tracking-wider uppercase text-on-surface-variant/40 border-b border-outline-variant/20 bg-void-black/20">
                  <th className="p-4 sm:p-6 font-medium w-[30%]">Asset</th>
                  <th className="p-4 sm:p-6 font-medium w-[15%]">Allocation</th>
                  <th className="p-4 sm:p-6 font-medium w-[18%]">
                    Current Price
                  </th>
                  <th className="p-4 sm:p-6 font-medium text-right w-[20%]">
                    Unrealized P&L
                  </th>
                  <th className="p-4 sm:p-6 font-medium text-right w-[17%]">
                    Exposure
                  </th>
                </tr>
              </thead>
              <tbody className="text-xs sm:text-sm text-on-surface/90 divide-y divide-outline-variant/10">
                {currentUserID === 999 ? (
                  mockPositions.map(
                    (position: MockPositionMetrics, idx: number) => (
                      <tr
                        key={idx}
                        className="hover:bg-surface-container-high/30 transition-colors duration-200 group"
                      >
                        <td className="p-4 sm:p-6 min-w-0">
                          <div className="flex items-center gap-3.5 min-w-0">
                            <div
                              className={`w-8 h-8 rounded-full bg-void-black border border-outline-variant/20 flex items-center justify-center font-mono text-[10px] text-on-surface/60 shrink-0 ${position.borderClass || ""}`}
                            >
                              {position.initials}
                            </div>
                            <div className="min-w-0 truncate">
                              <p className="font-medium group-hover:text-primary-fixed-dim transition-colors duration-300 truncate">
                                {position.name}
                              </p>
                              <p className="font-mono text-[10px] text-on-surface-variant/40 mt-0.5 truncate">
                                {position.symbol} • {position.category}
                              </p>
                            </div>
                          </div>
                        </td>
                        <td className="p-4 sm:p-6 font-mono font-light text-on-surface-variant truncate">
                          {position.allocation}%
                        </td>
                        <td className="p-4 sm:p-6 font-mono font-light text-on-surface-variant truncate">
                          $
                          {position.currentPrice.toLocaleString(undefined, {
                            minimumFractionDigits: 2,
                          })}
                        </td>
                        <td className="p-4 sm:p-6 text-right shrink-0">
                          <span
                            className={`font-mono font-medium truncate ${position.isPositive ? "text-primary-fixed-dim drop-shadow-[0_0_6px_rgba(0,219,233,0.25)]" : "text-error"}`}
                          >
                            {position.isPositive ? "+" : ""}$
                            {position.unrealizedPnL.toLocaleString(undefined, {
                              minimumFractionDigits: 2,
                            })}
                          </span>
                        </td>
                        <td className="p-4 sm:p-6 text-right font-mono font-light tracking-tight text-on-surface-variant truncate">
                          $
                          {position.exposure.toLocaleString(undefined, {
                            minimumFractionDigits: 2,
                          })}
                        </td>
                      </tr>
                    ),
                  )
                ) : !serverPositions || serverPositions.length === 0 ? (
                  <tr className="hover:bg-transparent">
                    <td
                      colSpan={5}
                      className="p-8 text-center font-mono text-xs tracking-wider text-on-surface-variant/40 uppercase"
                    >
                      No active asset positions deployed. Liquid Balance held
                      entirely in Cash.
                    </td>
                  </tr>
                ) : (
                  serverPositions.map(
                    (position: ServerPositionMetrics, idx: number) => {
                      const isPositive = position.unrealized_pnl >= 0;
                      const flashDirection =
                        priceFlashes[position.symbol.toUpperCase()];
                      const dynamicAllocation =
                        totalAssetValue > 0
                          ? (
                              (position.exposure / totalAssetValue) *
                              100
                            ).toFixed(1)
                          : "0.0";
                      return (
                        <tr
                          key={idx}
                          className={`group transition-[background-color,box-shadow] duration-500 ${
                            flashDirection
                              ? ""
                              : "hover:bg-surface-container-high/30 shadow-none"
                          }`}
                          style={
                            flashDirection
                              ? priceFlashStyles[flashDirection]
                              : undefined
                          }
                        >
                          <td className="p-4 sm:p-6 min-w-0">
                            <div className="flex items-center gap-3.5 min-w-0">
                              <div className="w-8 h-8 rounded-full bg-void-black border border-outline-variant/20 flex items-center justify-center font-mono text-[10px] text-on-surface/60 shrink-0">
                                {position.symbol.slice(0, 2).toUpperCase()}
                              </div>
                              <div className="min-w-0 truncate">
                                <p className="font-medium group-hover:text-primary-fixed-dim transition-colors duration-300 truncate">
                                  {position.symbol} Asset Node
                                </p>
                                <p className="font-mono text-[10px] text-on-surface-variant/40 mt-0.5 truncate">
                                  {position.symbol} • Core Ledger Partition
                                </p>
                              </div>
                            </div>
                          </td>
                          <td className="p-4 sm:p-6 font-mono font-light text-on-surface-variant truncate">
                            {dynamicAllocation}%
                          </td>
                          <td className="p-4 sm:p-6 font-mono font-light text-on-surface-variant truncate">
                            $
                            {position.current_price.toLocaleString(undefined, {
                              minimumFractionDigits: 2,
                            })}
                          </td>
                          <td className="p-4 sm:p-6 text-right shrink-0">
                            <span
                              className={`font-mono font-medium truncate ${isPositive ? "text-primary-fixed-dim drop-shadow-[0_0_6px_rgba(0,219,233,0.25)]" : "text-error"}`}
                            >
                              {isPositive ? "+" : ""}$
                              {position.unrealized_pnl.toLocaleString(
                                undefined,
                                { minimumFractionDigits: 2 },
                              )}
                            </span>
                          </td>
                          <td className="p-4 sm:p-6 text-right font-mono font-light tracking-tight text-on-surface-variant truncate">
                            $
                            {position.exposure.toLocaleString(undefined, {
                              minimumFractionDigits: 2,
                            })}
                          </td>
                        </tr>
                      );
                    },
                  )
                )}
              </tbody>
            </table>
          </div>

          <div className="p-5 text-center border-t border-outline-variant/10 bg-white/[0.005]">
            <button className="font-mono text-[10px] tracking-wider uppercase text-on-surface-variant/70 hover:text-on-surface border border-outline-variant/20 rounded-full px-5 py-2 hover:bg-void-black transition-all duration-300">
              View All Positions
            </button>
          </div>
        </section>
      </div>
    </AppShell>
  );
}

function normalizeCompositionPercentages<
  T extends {
    exposure: number;
    label: string;
    color: string;
    glowClass: string;
  },
>(segments: T[]) {
  const positiveSegments = segments.filter((segment) => segment.exposure > 0);
  const displayedExposure = positiveSegments.reduce(
    (sum, segment) => sum + segment.exposure,
    0,
  );
  if (positiveSegments.length === 0 || displayedExposure <= 0) {
    return [
      {
        label: "Cash & Equiv",
        exposure: 1,
        percentage: 100,
        color: "#849495",
        glowClass: "bg-outline",
      },
    ];
  }

  const exact = positiveSegments.map((segment) => ({
    ...segment,
    exactPercentage: (segment.exposure / displayedExposure) * 100,
  }));
  const rounded = exact.map((segment) => ({
    ...segment,
    percentage: Math.floor(segment.exactPercentage),
  }));
  let remaining =
    100 - rounded.reduce((sum, segment) => sum + segment.percentage, 0);

  rounded
    .sort(
      (a, b) =>
        b.exactPercentage -
        Math.floor(b.exactPercentage) -
        (a.exactPercentage - Math.floor(a.exactPercentage)),
    )
    .forEach((segment) => {
      if (remaining <= 0) return;
      segment.percentage += 1;
      remaining -= 1;
    });

  return rounded.filter((segment) => segment.percentage > 0);
}
