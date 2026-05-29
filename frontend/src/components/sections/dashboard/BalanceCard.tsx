"use client";

import { useMemo } from "react";

interface SnapshotPoint {
  total_asset_value: number;
  timestamp: string;
}

interface BalanceCardProps {
  isAgentActive: boolean;
  displayPnL: string;
  historyData?: SnapshotPoint[];
}

const FLATLINE_Y = 70;
const FLAT_CHART_COORDINATES = {
  pathString: `M 0 ${FLATLINE_Y} L 100 ${FLATLINE_Y}`,
  lastPoint: { x: 100, y: FLATLINE_Y },
};

export default function BalanceCard({
  isAgentActive,
  displayPnL,
  historyData,
}: BalanceCardProps) {
  const chartCoordinates = useMemo(() => {
    if (!historyData || historyData.length < 2) {
      return FLAT_CHART_COORDINATES;
    }

    const values = historyData.map((d) => d.total_asset_value);
    const minVal = Math.min(...values);
    const maxVal = Math.max(...values);
    const valRange = maxVal - minVal;

    if (valRange === 0) {
      return FLAT_CHART_COORDINATES;
    }

    // Leave vertical padding so the sparkline never clips the SVG edges.
    const points = historyData.map((snapshot, index) => {
      const x = (index / (historyData.length - 1)) * 100;
      const y = 85 - ((snapshot.total_asset_value - minVal) / valRange) * 70;
      return { x, y };
    });

    const pathString = points.reduce(
      (acc, p, i) => (i === 0 ? `M ${p.x} ${p.y}` : `${acc} L ${p.x} ${p.y}`),
      "",
    );

    return {
      pathString,
      lastPoint: points[points.length - 1],
    };
  }, [historyData]);

  return (
    <div className="md:col-span-12 xl:col-span-8 group relative flex flex-col h-auto min-h-[380px] sm:h-[460px] bg-surface-container-low border border-outline-variant/30 rounded-[32px] p-5 sm:p-8 overflow-hidden hover:bg-surface-container transition-all duration-700 hover:shadow-[0_20px_40px_rgba(0,0,0,0.2)]">
      <div className="absolute top-0 inset-x-0 h-[1px] bg-gradient-to-r from-transparent via-primary-container/30 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-700" />

      <div className="flex justify-between items-center mb-6 sm:mb-8 border-b border-outline-variant/20 pb-4 relative z-10">
        <div className="flex items-center gap-3">
          <div
            className={`w-2.5 h-2.5 rounded-full shadow-[0_0_10px_rgba(0,240,255,0.4)] ${isAgentActive ? "bg-primary-fixed-dim animate-pulse" : "bg-on-surface-variant/40"}`}
          />
          <span className="text-[10px] font-medium tracking-[0.2em] text-on-surface uppercase">
            Agent Status: {isAgentActive ? "Optimising" : "Idle"}
          </span>
        </div>
        <span className="px-2.5 py-0.5 bg-white/[0.02] border border-outline-variant/30 rounded-full text-[9px] font-medium tracking-widest text-secondary-fixed">
          REGIME: VOLATILE
        </span>
      </div>

      <div className="flex-grow flex flex-col justify-center relative z-10">
        <span className="text-[10px] font-medium tracking-[0.2em] text-on-surface-variant uppercase block mb-1">
          Current P&L (24h)
        </span>
        <h2 className="text-4xl sm:text-5xl xl:text-6xl font-light tracking-tight text-primary-fixed">
          {displayPnL}
        </h2>

        <div className="mt-6 sm:mt-8 h-28 sm:h-36 w-full relative flex items-end border-b border-outline-variant/20">
          <svg
            className="w-full h-full"
            preserveAspectRatio="none"
            viewBox="0 0 100 100"
          >
            <path
              d={chartCoordinates.pathString}
              fill="none"
              stroke="#00f0ff"
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth="1.2"
              style={{ filter: "drop-shadow(0 0 6px rgba(0,240,255,0.5))" }}
            />
          </svg>

          <div
            className="absolute w-2 h-2 rounded-full bg-primary-container shadow-[0_0_12px_#00f0ff] transition-all duration-500 ease-out"
            style={{
              left: `${chartCoordinates.lastPoint.x}%`,
              top: `${chartCoordinates.lastPoint.y}%`,
              transform: "translate(-50%, -50%)",
            }}
          >
            <div className="absolute inset-0 rounded-full bg-primary-fixed-dim animate-ping opacity-75" />
          </div>
        </div>
      </div>
    </div>
  );
}
