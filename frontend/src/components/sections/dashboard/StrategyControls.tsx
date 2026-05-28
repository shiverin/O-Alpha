interface StrategyControlsProps {
  riskTolerance: number;
  setRiskTolerance: (value: number) => void;
  volatilityCap: number;
  setVolatilityCap: (value: number) => void;
  leverageMultiplier: number;
  setLeverageMultiplier: (value: number) => void;
  calculatedLeverageText: string;
}

export default function StrategyControls({
  riskTolerance,
  setRiskTolerance,
  volatilityCap,
  setVolatilityCap,
  leverageMultiplier,
  setLeverageMultiplier,
  calculatedLeverageText,
}: StrategyControlsProps) {
  return (
    <div className="md:col-span-12 xl:col-span-4 group relative flex flex-col h-auto xl:h-[460px] bg-surface-container-low border border-outline-variant/30 rounded-[32px] p-5 sm:p-8 overflow-hidden hover:bg-surface-container transition-all duration-700 hover:shadow-[0_20px_40px_rgba(0,0,0,0.2)]">
      <div className="absolute top-0 inset-x-0 h-[1px] bg-gradient-to-r from-transparent via-secondary-fixed-dim/30 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-700" />
      
      <div className="mb-6 xl:mb-8 border-b border-outline-variant/20 pb-4 flex items-center justify-between">
        <h3 className="text-[10px] font-medium tracking-[0.2em] text-on-surface uppercase">Strategy Controls</h3>
        <span className="text-[9px] text-primary-container uppercase tracking-wider">Sync Active</span>
      </div>

      <div className="flex flex-col gap-6 flex-grow justify-center py-4 xl:py-0">
        {/* Risk Tolerance */}
        <div className="flex flex-col gap-2">
          <div className="flex justify-between text-[10px] font-medium tracking-wider text-on-surface-variant">
            <span>Risk Tolerance</span>
            <span className="text-primary-container">
              {riskTolerance > 75 ? "High" : riskTolerance > 40 ? "Balanced" : "Conservative"}
            </span>
          </div>
          <input type="range" min="1" max="100" value={riskTolerance} onChange={(e) => setRiskTolerance(parseInt(e.target.value))} className="w-full h-[2px] appearance-none bg-outline-variant/30 rounded-full outline-none accent-primary-container cursor-pointer" />
        </div>

        {/* Volatility Cap */}
        <div className="flex flex-col gap-2">
          <div className="flex justify-between text-[10px] font-medium tracking-wider text-on-surface-variant">
            <span>Volatility Cap</span>
            <span className="text-primary-container">{volatilityCap}%</span>
          </div>
          <input type="range" min="5" max="50" value={volatilityCap} onChange={(e) => setVolatilityCap(parseInt(e.target.value))} className="w-full h-[2px] appearance-none bg-outline-variant/30 rounded-full outline-none accent-primary-container cursor-pointer" />
        </div>

        {/* Leverage Multiplier */}
        <div className="flex flex-col gap-2">
          <div className="flex justify-between text-[10px] font-medium tracking-wider text-on-surface-variant">
            <span>Leverage Multiplier</span>
            <span className="text-secondary-fixed">{calculatedLeverageText}</span>
          </div>
          <input type="range" min="0" max="100" value={leverageMultiplier} onChange={(e) => setLeverageMultiplier(parseInt(e.target.value))} className="w-full h-[2px] appearance-none bg-outline-variant/30 rounded-full outline-none accent-secondary-fixed-dim cursor-pointer" />
        </div>
      </div>
    </div>
  );
}