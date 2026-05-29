import { ServerPortfolioSummary } from "@/types/dashboard";
import { allocationSegments } from "@/lib/mockAppData";

interface PortfolioAllocationProps {
  currentUserID: number;
  serverSummary: ServerPortfolioSummary | undefined;
}

export default function PortfolioAllocation({
  currentUserID,
  serverSummary,
}: PortfolioAllocationProps) {
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
            <circle
              cx="50"
              cy="50"
              r="40"
              fill="transparent"
              stroke="#00f0ff"
              strokeWidth="12"
              strokeDasharray="251.2"
              strokeDashoffset="150.72"
              strokeLinecap="round"
              style={{ filter: "drop-shadow(0 0 6px rgba(0,240,255,0.3))" }}
            />
            <circle
              cx="50"
              cy="50"
              r="40"
              fill="transparent"
              stroke="#ffd700"
              strokeWidth="12"
              strokeDasharray="251.2"
              strokeDashoffset="175.84"
              strokeLinecap="round"
              style={{ filter: "drop-shadow(0 0 6px rgba(255,215,0,0.2))" }}
              transform="rotate(144 50 50)"
            />
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
          {allocationSegments.map((segment, index) => (
            <div
              key={index}
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
