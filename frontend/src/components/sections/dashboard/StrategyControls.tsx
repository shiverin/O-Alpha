type RiskProfile = "conservative" | "moderate" | "aggressive";

interface StrategyControlsProps {
  riskProfile: RiskProfile;
  universeSize: number;
}

export default function StrategyControls({
  riskProfile,
  universeSize,
}: StrategyControlsProps) {
  return (
    <div className="md:col-span-12 xl:col-span-4 group relative flex flex-col h-auto xl:h-[460px] bg-surface-container-low border border-outline-variant/30 rounded-[32px] p-5 sm:p-8 overflow-hidden hover:bg-surface-container transition-all duration-700 hover:shadow-[0_20px_40px_rgba(0,0,0,0.2)]">
      <div className="mb-6 xl:mb-8 border-b border-outline-variant/20 pb-4 flex items-center justify-between">
        <h3 className="text-[10px] font-medium tracking-[0.2em] text-on-surface uppercase">
          Strategy Profile
        </h3>
        <span className="text-[9px] text-primary-container uppercase tracking-wider">
          Paper
        </span>
      </div>

      <div className="flex flex-col justify-center flex-grow gap-4 py-4 xl:py-0">
        <div className="rounded-2xl border border-outline-variant/15 bg-void-black/15 px-4 py-4">
          <p className="text-[10px] font-mono uppercase tracking-[0.18em] text-on-surface-variant/50">
            Risk Profile
          </p>
          <div className="mt-2 flex items-center justify-between gap-3">
            <span className="text-lg font-light capitalize tracking-wide text-on-surface">
              {riskProfile}
            </span>
          </div>
        </div>

        <div className="rounded-2xl border border-outline-variant/15 bg-void-black/15 px-4 py-4">
          <p className="text-[10px] font-mono uppercase tracking-[0.18em] text-on-surface-variant/50">
            Asset Universe
          </p>
          <div className="mt-2 flex items-baseline justify-between gap-3">
            <span className="font-mono text-lg tracking-wide text-on-surface">
              {universeSize || 0}
            </span>
            <span className="text-[10px] font-mono uppercase tracking-[0.18em] text-on-surface-variant/60">
              Symbols
            </span>
          </div>
        </div>

        <div className="h-px bg-outline-variant/20" />
      </div>
    </div>
  );
}
