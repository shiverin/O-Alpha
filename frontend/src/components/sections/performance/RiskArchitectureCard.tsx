"use client";

export default function RiskArchitectureCard() {
  return (
    <div className="group relative flex flex-col h-full min-h-[340px] bg-surface-container-low border border-outline-variant/30 rounded-[32px] p-8 overflow-hidden hover:bg-surface-container transition-all duration-700 ease-[cubic-bezier(0.16,1,0.3,1)] hover:-translate-y-1 hover:shadow-[0_20px_40px_rgba(0,0,0,0.1)]">
      <div className="absolute top-0 inset-x-0 h-[1px] bg-gradient-to-r from-transparent via-secondary-fixed/30 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-700" />

      <div className="flex items-center gap-2 mb-8">
        <span className="material-symbols-outlined text-sm text-on-surface-variant">shield</span>
        <span className="text-[10px] uppercase tracking-[0.2em] text-on-surface-variant font-medium">
          Risk Architecture
        </span>
      </div>

      <div className="flex flex-col flex-grow justify-center gap-6">
        <div className="group/item flex flex-col gap-2">
          <h3 className="text-lg font-light tracking-wide text-on-surface flex items-center gap-2">
            Kalman Filtering
          </h3>
          <p className="text-sm font-light leading-relaxed text-on-surface-variant group-hover/item:text-on-surface transition-colors duration-300">
            Extracts true signal from market noise, enabling precise entry points even in low-liquidity environments.
          </p>
        </div>

        <div className="h-px w-full bg-gradient-to-r from-transparent via-outline-variant/20 to-transparent" />

        <div className="group/item flex flex-col gap-2">
          <h3 className="text-lg font-light tracking-wide text-on-surface flex items-center gap-2">
            Convex Optimization
          </h3>
          <p className="text-sm font-light leading-relaxed text-on-surface-variant group-hover/item:text-on-surface transition-colors duration-300">
            Portfolio weights are solved continuously to maximize expected return subject to strict variance constraints.
          </p>
        </div>
      </div>
    </div>
  );
}
