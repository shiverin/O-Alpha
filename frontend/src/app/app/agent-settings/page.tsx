"use client";

import { AppShell } from "@/components/app/AppShell";
import { useState } from "react";

export default function AgentSettingsPage() {
  const [isAdvanced, setIsAdvanced] = useState(false);
  const [isSaving, setIsSaving] = useState(false);

  // Core Posture State
  const [riskProfile, setRiskProfile] = useState("moderate");

  // Tracking flip state configurations independently
  const [flippedCards, setFlippedCards] = useState<{ [key: string]: boolean }>({
    conservative: false,
    moderate: false,
    aggressive: false,
  });

  // Under-the-hood Granular Hyperparameters
  const [leverage, setLeverage] = useState(2);
  const [maxPositions, setMaxPositions] = useState(6);
  const [stopLoss, setStopLoss] = useState(2.5);
  const [takeProfit, setTakeProfit] = useState(5.0);
  const [rebalanceFreq, setRebalanceFreq] = useState("daily");

  const handleProfileSelection = (profile: "conservative" | "moderate" | "aggressive") => {
    setRiskProfile(profile);
    if (profile === "conservative") {
      setLeverage(1);
      setMaxPositions(3);
      setStopLoss(1.5);
      setTakeProfit(3.0);
      setRebalanceFreq("weekly");
    } else if (profile === "moderate") {
      setLeverage(2);
      setMaxPositions(6);
      setStopLoss(2.5);
      setTakeProfit(5.0);
      setRebalanceFreq("daily");
    } else if (profile === "aggressive") {
      setLeverage(4);
      setMaxPositions(12);
      setStopLoss(4.0);
      setTakeProfit(12.0);
      setRebalanceFreq("hourly");
    }
  };

  const toggleCardFlip = (profile: string, e: React.MouseEvent) => {
    e.stopPropagation(); // Restricts parent selection firing during flip review
    setFlippedCards((prev) => ({ ...prev, [profile]: !prev[profile] }));
  };

  // const calculatedLeverageText = useMemo(() => {
  //   return `${(1.0 + (leverage / 5) * 4).toFixed(1)}x`;
  // }, [leverage]);

  const handleSave = async () => {
    setIsSaving(true);
    await new Promise(resolve => setTimeout(resolve, 800));
    setIsSaving(false);
    alert("Neural core settings synchronized successfully.");
  };

  const profileDescriptions = {
    conservative: "Prioritizes structural capital preservation. Minimizes drawdown lengths using tight trailing stops and allocation focuses strictly on low-beta asset pairings.",
    moderate: "Engineered for optimal risk-adjusted discovery. Deploys dynamic position scaling rules and active momentum tracking filters to navigate regime transitions.",
    aggressive: "Optimized for high-volatility statistical arbitrage workflows. Executes maximum leverage thresholds alongside complex options configurations to yield convex returns.",
  };

  return (
    <AppShell title="Agent Settings">
      <div className="w-full bg-transparent flex flex-col gap-6 md:gap-10 animate-in fade-in duration-700 ease-[cubic-bezier(0.16,1,0.3,1)]">
        
        {/* =========================================
            HEADER CONTROL GATEWAY
        ========================================= */}
        <div className="flex flex-col sm:flex-row justify-between items-start sm:items-end gap-4 pb-2 border-b border-outline-variant/10">
          <div>
            <h1 className="text-2xl sm:text-3xl font-light tracking-tight text-on-surface">Agent Configuration</h1>
            <p className="text-xs sm:text-sm font-light text-on-surface-variant/70 mt-1">
              Adjust core risk frameworks and execution boundaries for the O(Alpha) trading core.
            </p>
          </div>
          
          <button
            type="button"
            onClick={() => setIsAdvanced(!isAdvanced)}
            className={`w-full sm:w-auto px-5 py-2 rounded-full text-xs font-mono font-medium tracking-wider uppercase border transition-all duration-300 active:scale-95 flex items-center justify-center gap-2 ${
              isAdvanced 
                ? "bg-white/[0.04] border-primary-container/40 text-primary-fixed-dim" 
                : "bg-transparent border-outline-variant/30 text-on-surface-variant hover:text-on-surface hover:border-outline-variant/60"
            }`}
          >
            <span className="material-symbols-outlined text-[14px]">tune</span>
            {isAdvanced ? "Simple Mode" : "Advanced Tuning"}
          </button>
        </div>

        {/* =========================================
            SIMPLE WORKSPACE: RECONFIGURED BENTO MATRIX
        ========================================= */}
        <div className="flex flex-col gap-2">
          <span className="text-[10px] font-mono tracking-[0.2em] text-on-surface-variant/40 uppercase block mb-1">
            System Posture Blueprint
          </span>
          
          {/* OPTIMIZED: Changed from md:grid-cols-3 to xl:grid-cols-3 so intermediate viewports keep cards full-width */}
          <div className="grid grid-cols-1 xl:grid-cols-3 gap-6 w-full">
            {(["conservative", "moderate", "aggressive"] as const).map((profile) => {
              const isSelected = riskProfile === profile;
              const isFlipped = flippedCards[profile];

              return (
                <div 
                  key={profile}
                  onClick={() => handleProfileSelection(profile)}
                  className="[perspective:1000px] h-44 w-full cursor-pointer select-none"
                >
                  {/* 3D Core Container Engine */}
                  <div className={`relative w-full h-full transition-transform duration-500 [transform-style:preserve-3d] ${
                    isFlipped ? "[transform:rotateY(180deg)]" : ""
                  }`}>
                    
                    {/* ───────────────────────────────────────
                        FRONT FACE: ULTRACLEAN RADIAL LAYOUT
                    ─────────────────────────────────────── */}
                    <div className={`absolute inset-0 [backface-visibility:hidden] flex flex-col justify-center items-center bg-surface-container-low border rounded-[24px] p-6 transition-all duration-300 ${
                      isSelected 
                        ? "border-primary-fixed-dim shadow-[0_0_20px_rgba(0,240,255,0.06)] bg-surface-container" 
                        : "border-outline-variant/30 hover:border-outline-variant/60"
                    }`}>
                      {/* Fixed Top-Right Question Mark Node */}
                      <button
                        type="button"
                        onClick={(e) => toggleCardFlip(profile, e)}
                        className="absolute right-6 top-6 text-on-surface-variant/30 hover:text-primary-fixed-dim transition-colors h-7 w-7 rounded-full flex items-center justify-center hover:bg-white/5 border border-transparent"
                      >
                        <span className="material-symbols-outlined text-[18px]">help</span>
                      </button>
                      
                      {/* Isolated Single Word Brand Typography */}
                      <h4 className={`text-xl font-light tracking-widest uppercase transition-all duration-300 ${
                        isSelected ? "text-primary-fixed-dim font-medium" : "text-on-surface-variant/70"
                      }`}>
                        {profile}
                      </h4>
                    </div>

                    {/* ───────────────────────────────────────
                        BACK FACE: MINIMAL BALANCED SUMMARY
                    ─────────────────────────────────────── */}
                    <div className="absolute inset-0 [backface-visibility:hidden] [transform:rotateY(180deg)] flex flex-col justify-center bg-surface-container border border-outline-variant/40 rounded-[24px] p-6 shadow-xl">
                      {/* Reversion Icon mapped to EXACT identical spatial alignment anchor */}
                      <button
                        type="button"
                        onClick={(e) => toggleCardFlip(profile, e)}
                        className="absolute right-6 top-6 text-primary-fixed-dim/70 hover:text-primary-fixed-dim h-7 w-7 rounded-full flex items-center justify-center bg-white/5 border border-outline-variant/20"
                      >
                        <span className="material-symbols-outlined text-[16px]">flip_to_front</span>
                      </button>

                      {/* Clean description cluster with right side gutter safe-padding */}
                      <p className="text-xs font-light leading-relaxed text-on-surface-variant/80 pr-6 select-text">
                        {profileDescriptions[profile]}
                      </p>
                    </div>

                  </div>
                </div>
              );
            })}
          </div>
        </div>

        {/* =========================================
            ADVANCED EXPERT VARIABLE PARAMETERS
        ========================================= */}
        {isAdvanced && (
          /* OPTIMIZED: Adjusted tuning panels down to xl:grid-cols-2 as well for unified visual symmetry */
          <div className="grid grid-cols-1 xl:grid-cols-2 gap-6 md:gap-8 animate-in fade-in slide-in-from-top-4 duration-500 ease-[cubic-bezier(0.16,1,0.3,1)] border-t border-outline-variant/10 pt-6">
            
            {/* INPUT PANEL MODULE LEFT */}
            <div className="group relative bg-surface-container-low border border-outline-variant/30 rounded-[32px] p-6 sm:p-8 flex flex-col gap-6">
              
              <div className="flex flex-col gap-2">
                <div className="flex justify-between text-[11px] font-mono tracking-wider text-on-surface-variant">
                  <span className="uppercase tracking-widest">Maximum Leverage</span>
                  <span className="text-primary-container font-semibold">{leverage}x</span>
                </div>
                <input
                  type="range" min="1" max="5" value={leverage}
                  onChange={(e) => setLeverage(parseInt(e.target.value))}
                  className="w-full h-[2px] appearance-none bg-outline-variant/30 rounded-full outline-none accent-primary-container cursor-pointer"
                />
              </div>

              <div className="flex flex-col gap-2">
                <div className="flex justify-between text-[11px] font-mono tracking-wider text-on-surface-variant">
                  <span className="uppercase tracking-widest">Max Concurrent Boundaries</span>
                  <span className="text-primary-container font-semibold">{maxPositions} Units</span>
                </div>
                <input
                  type="range" min="1" max="20" value={maxPositions}
                  onChange={(e) => setMaxPositions(parseInt(e.target.value))}
                  className="w-full h-[2px] appearance-none bg-outline-variant/30 rounded-full outline-none accent-primary-container cursor-pointer"
                />
              </div>

              <div className="flex flex-col gap-3 border-t border-outline-variant/10 pt-4 mt-2">
                <span className="text-[10px] font-mono tracking-[0.2em] text-on-surface-variant/50 uppercase">
                  Execution Frequency
                </span>
                <div className="grid grid-cols-3 gap-2 bg-void-black/20 p-1 rounded-xl border border-outline-variant/10">
                  {(["hourly", "daily", "weekly"] as const).map((freq) => {
                    const active = rebalanceFreq === freq;
                    return (
                      <button
                        key={freq}
                        type="button"
                        onClick={() => setRebalanceFreq(freq)}
                        className={`py-1.5 rounded-lg font-mono text-[10px] tracking-wide uppercase transition-all duration-200 ${
                          active
                            ? "bg-surface-container text-on-surface border border-outline-variant/30 font-medium"
                            : "text-on-surface-variant/40 hover:text-on-surface"
                        }`}
                      >
                        {freq}
                      </button>
                    );
                  })}
                </div>
              </div>
            </div>

            {/* INPUT PANEL MODULE RIGHT */}
            <div className="group relative bg-surface-container-low border border-outline-variant/30 rounded-[32px] p-6 sm:p-8 flex flex-col gap-6">
              
              <div className="flex flex-col gap-2">
                <div className="flex justify-between text-[11px] font-mono tracking-wider text-on-surface-variant">
                  <span className="uppercase tracking-widest">Stop Loss Boundary</span>
                  <span className="text-error font-medium">-{stopLoss}%</span>
                </div>
                <input
                  type="range" min="0.5" max="10" step="0.5" value={stopLoss}
                  onChange={(e) => setStopLoss(parseFloat(e.target.value))}
                  className="w-full h-[2px] appearance-none bg-outline-variant/30 rounded-full outline-none accent-error cursor-pointer"
                />
              </div>

              <div className="flex flex-col gap-2">
                <div className="flex justify-between text-[11px] font-mono tracking-wider text-on-surface-variant">
                  <span className="uppercase tracking-widest">Take Profit Target</span>
                  <span className="text-primary-fixed-dim font-medium">+{takeProfit}%</span>
                </div>
                <input
                  type="range" min="1" max="20" step="0.5" value={takeProfit}
                  onChange={(e) => setTakeProfit(parseFloat(e.target.value))}
                  className="w-full h-[2px] appearance-none bg-outline-variant/30 rounded-full outline-none accent-primary-fixed-dim cursor-pointer"
                />
              </div>
              
              <div className="bg-white/[0.01] border border-outline-variant/10 rounded-xl p-3.5 text-[11px] font-light text-on-surface-variant/50 leading-relaxed mt-1">
                System triggers are updated dynamically across active execution loops. Hard targets decouple from standard client frames to prevent slippage anomalies.
              </div>
            </div>

          </div>
        )}

        {/* =========================================
            GLOBAL DISPATCH SYNC ACTION
          ========================================= */}
        <div className="pt-4 border-t border-outline-variant/20 flex justify-end">
          <button
            type="button"
            onClick={handleSave}
            disabled={isSaving}
            className={`w-full sm:w-auto px-8 py-3 rounded-full text-xs font-mono font-medium tracking-wider uppercase text-background transition-all duration-300 active:scale-95 shadow-md ${
              isSaving 
                ? "bg-primary-container/40 cursor-not-allowed text-void-black/40" 
                : "bg-primary-container text-void-black shadow-primary-container/10 hover:bg-primary-container/90"
            }`}
          >
            {isSaving ? "Synchronizing Matrix..." : "Save Terminal Configuration"}
          </button>
        </div>

      </div>
    </AppShell>
  );
}