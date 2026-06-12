import {
  ServerPortfolioSummary,
  ServerPositionMetrics,
} from "@/types/dashboard";
import { allocationSegments } from "@/lib/mockAppData";

interface PortfolioAllocationProps {
  currentUserID: number;
  serverSummary: ServerPortfolioSummary | undefined;
  serverPositions?: ServerPositionMetrics[];
}

export default function PortfolioAllocation({
  currentUserID,
  serverSummary,
  serverPositions,
}: PortfolioAllocationProps) {
  const totalAssetValue = serverSummary?.total_asset_value ?? 0;
  const positionExposure =
    serverPositions?.reduce((sum, position) => sum + position.exposure, 0) ?? 0;
  const cashExposure = Math.max(totalAssetValue - positionExposure, 0);
  const realSegments =
    currentUserID === 999
      ? allocationSegments
      : buildAllocationSegments(
          serverPositions ?? [],
          cashExposure,
          totalAssetValue,
        );
  const circumference = 251.2;
  let rotation = 0;

  return (
    <div className="md:col-span-12 group relative flex flex-col h-auto md:h-[380px] bg-surface-container-low border border-outline-variant/30 rounded-[32px] p-5 sm:p-8 overflow-hidden hover:bg-surface-container transition-all duration-700 hover:shadow-[0_20px_40px_rgba(0,0,0,0.2)]">
      <h3 className="text-[10px] font-medium tracking-[0.2em] text-on-surface uppercase mb-6">
        Portfolio Allocation
      </h3>
      <div className="flex-grow flex flex-col sm:flex-row items-center justify-center gap-6 sm:gap-8 h-full py-4 sm:py-0">
        <div className="relative w-36 h-36 sm:w-40 sm:h-40 shrink-0 flex items-center justify-center">
          <svg
            className="absolute inset-0 transform -rotate-90"
            viewBox="0 0 100 100"
          >
            <circle
              cx="50"
              cy="50"
              r="40"
              fill="transparent"
              stroke="#2a2a2a"
              strokeWidth="12"
            />
            {realSegments.map((segment) => {
              const dashOffset =
                circumference - (circumference * segment.percentage) / 100;
              const currentRotation = rotation;
              rotation += segment.percentage * 3.6;
              return (
                <circle
                  key={segment.label}
                  cx="50"
                  cy="50"
                  r="40"
                  fill="transparent"
                  stroke={segment.color}
                  strokeWidth="12"
                  strokeDasharray={circumference}
                  strokeDashoffset={dashOffset}
                  strokeLinecap="round"
                  style={{
                    filter: `drop-shadow(0 0 6px ${segment.color}55)`,
                  }}
                  transform={`rotate(${currentRotation} 50 50)`}
                />
              );
            })}
          </svg>
          <div className="absolute flex flex-col items-center justify-center text-center">
            <span className="text-[9px] font-medium tracking-widest text-on-surface-variant/50">
              TOTAL AUM
            </span>
            <span className="text-xl font-light tracking-tight text-on-surface">
              $
              {currentUserID === 999 || !serverSummary
                ? "2.4M"
                : (serverSummary.total_asset_value / 1000000).toFixed(1) + "M"}
            </span>
          </div>
        </div>

        <div className="flex flex-col gap-3 flex-grow justify-center w-full max-w-[200px]">
          {realSegments.map((segment, index) => (
            <div
              key={`${segment.label}-${index}`}
              className="flex items-center justify-between text-xs font-light"
            >
              <div className="flex items-center gap-3">
                <div className={`w-2 h-2 rounded-full ${segment.glowClass}`} />
                <span className="text-on-surface-variant/80">
                  {segment.label}
                </span>
              </div>
              <span className="font-mono text-on-surface/50 text-[11px]">
                ({segment.percentage}%)
              </span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

function buildAllocationSegments(
  positions: ServerPositionMetrics[],
  cashExposure: number,
  totalAssetValue: number,
) {
  const colors = ["#00f0ff", "#ffd700", "#7dd3fc", "#c4b5fd", "#34d399"];
  const glowClasses = [
    "bg-primary-fixed-dim shadow-[0_0_8px_rgba(0,240,255,0.5)]",
    "bg-secondary-fixed-dim shadow-[0_0_8px_rgba(255,215,0,0.3)]",
    "bg-sky-300 shadow-[0_0_8px_rgba(125,211,252,0.35)]",
    "bg-violet-300 shadow-[0_0_8px_rgba(196,181,253,0.3)]",
    "bg-emerald-300 shadow-[0_0_8px_rgba(52,211,153,0.3)]",
  ];

  if (totalAssetValue <= 0) {
    return [
      {
        label: "Cash",
        percentage: 100,
        color: "#849495",
        glowClass: "bg-outline",
      },
    ];
  }

  const rankedPositions = positions
    .filter((position) => position.exposure > 0)
    .sort((a, b) => b.exposure - a.exposure);
  const visiblePositions = rankedPositions.slice(0, 4);
  const otherExposure = rankedPositions
    .slice(4)
    .reduce((sum, position) => sum + position.exposure, 0);

  const rawSegments = visiblePositions.map((position, index) => ({
    label: position.symbol,
    exposure: position.exposure,
    color: colors[index % colors.length],
    glowClass: glowClasses[index % glowClasses.length],
  }));

  if (otherExposure > 0) {
    rawSegments.push({
      label: "Other",
      exposure: otherExposure,
      color: "#94a3b8",
      glowClass: "bg-slate-400",
    });
  }

  if (cashExposure > 0 || rawSegments.length === 0) {
    rawSegments.push({
      label: "Cash",
      exposure: rawSegments.length === 0 ? totalAssetValue : cashExposure,
      color: "#849495",
      glowClass: "bg-outline",
    });
  }

  return normalizeAllocationPercentages(rawSegments, totalAssetValue).map(
    (segment) => ({
      label: segment.label,
      percentage: segment.percentage,
      color: segment.color,
      glowClass: segment.glowClass,
    }),
  );
}

function normalizeAllocationPercentages<
  T extends {
    exposure: number;
    label: string;
    color: string;
    glowClass: string;
  },
>(segments: T[], totalAssetValue: number) {
  const positiveSegments = segments.filter((segment) => segment.exposure > 0);
  if (positiveSegments.length === 0 || totalAssetValue <= 0) {
    return [
      {
        label: "Cash",
        percentage: 100,
        color: "#849495",
        glowClass: "bg-outline",
      },
    ];
  }
  const displayedExposure = positiveSegments.reduce(
    (sum, segment) => sum + segment.exposure,
    0,
  );
  if (displayedExposure <= 0) {
    return [
      {
        label: "Cash",
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

  return rounded
    .filter((segment) => segment.percentage > 0)
    .sort((a, b) => {
      if (a.label === "Cash") return 1;
      if (b.label === "Cash") return -1;
      if (a.label === "Other") return 1;
      if (b.label === "Other") return -1;
      return b.exposure - a.exposure;
    });
}
