"use client";

import { useMemo, useState } from "react";
import { AppShell } from "@/components/app/AppShell";
import { mockExecutionLogs, allocationSegments } from "@/lib/mockAppData";

export default function DashboardPage() {
  const [isAgentActive, setIsAgentActive] = useState(true);
  
  // Simulation control parameters (Loaded dynamically from user profile context rules)
  const [riskTolerance, setRiskTolerance] = useState(80);
  const [volatilityCap, setVolatilityCap] = useState(30);
  const [leverageMultiplier, setLeverageMultiplier] = useState(50);

  const calculatedLeverageText = useMemo(() => {
    return `${(1.0 + (leverageMultiplier / 100) * 4).toFixed(1)}x`;
  }, [leverageMultiplier]);

  return (
    <AppShell title="System Overview">
      <div className="w-full bg-transparent flex flex-col gap-6 md:gap-10 animate-in fade-in duration-700 ease-[cubic-bezier(0.16,1,0.3,1)]">
        
        <div className="flex flex-col sm:flex-row justify-between items-start sm:items-end gap-4 pb-2">
          <div>
            <p className="text-xs sm:text-sm font-light text-on-surface-variant/70 mt-1">
              Real-time telemetry and execution analytics.
            </p>
          </div>
          
          <button
            onClick={() => setIsAgentActive(!isAgentActive)}
            className="w-full sm:w-auto px-6 py-2.5 rounded-full text-xs font-medium tracking-wider uppercase shadow-md transition-all duration-500 ease-[cubic-bezier(0.16,1,0.3,1)] active:scale-95 bg-primary-container text-black shadow-primary-container/20 hover:bg-primary-container/90"
          >
            {isAgentActive ? "Terminate Agent" : "Launch Agent"}
          </button>
        </div>

        {/* =========================================
            OPTIMIZED RESPONSIVE BENTO GRID
        ========================================= */}
        <div className="grid grid-cols-1 md:grid-cols-12 gap-6 md:gap-8 items-start">
          
          {/* MAIN BALANCES TERMINAL CARD (OPTIMIZED: SPANS FULL 12 COLUMNS ON MD & LG; DROPS TO 8 ONLY ON XL) */}
          <div className="md:col-span-12 xl:col-span-8 group relative flex flex-col h-auto min-h-[380px] sm:h-[460px] bg-surface-container-low border border-outline-variant/30 rounded-[32px] p-5 sm:p-8 overflow-hidden hover:bg-surface-container transition-all duration-700 ease-[cubic-bezier(0.16,1,0.3,1)] hover:shadow-[0_20px_40px_rgba(0,0,0,0.2)]">
            <div className="absolute top-0 inset-x-0 h-[1px] bg-gradient-to-r from-transparent via-primary-container/30 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-700" />
            
            <div className="flex justify-between items-center mb-6 sm:mb-8 border-b border-outline-variant/20 pb-4 relative z-10">
              <div className="flex items-center gap-3">
                <div className={`w-2.5 h-2.5 rounded-full shadow-[0_0_10px_rgba(0,240,255,0.4)] ${isAgentActive ? "bg-primary-fixed-dim animate-pulse" : "bg-on-surface-variant/40"}`} />
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
              {/* Text scales smoothly across custom breakpoints */}
              <h2 className="text-4xl sm:text-5xl xl:text-6xl font-light tracking-tight text-primary-fixed">
                +$12,450.89
              </h2>
              
              {/* VECTOR LINE CHART LAYER */}
              <div className="mt-6 sm:mt-8 h-28 sm:h-36 w-full relative flex items-end border-b border-outline-variant/20">
                <svg className="w-full h-full" preserveAspectRatio="none" viewBox="0 0 100 100">
                  <path 
                    d="M 0 85 Q 15 95 30 75 T 60 55 T 90 20 L 100 15" 
                    fill="none" 
                    stroke="#00f0ff" 
                    strokeLinecap="round" 
                    strokeLinejoin="round" 
                    strokeWidth="1.5" 
                    style={{ filter: "drop-shadow(0 0 8px rgba(0,240,255,0.4))" }}
                  />
                </svg>
                <div className="absolute bottom-[13%] left-[30%] w-1.5 h-1.5 rounded-full bg-primary-container shadow-[0_0_8px_#00f0ff]" />
                <div className="absolute bottom-[43%] left-[60%] w-1.5 h-1.5 rounded-full bg-primary-container shadow-[0_0_8px_#00f0ff]" />
                <div className="absolute bottom-[78%] left-[90%] w-1.5 h-1.5 rounded-full bg-primary-container shadow-[0_0_8px_#00f0ff]" />
              </div>
            </div>
          </div>

          {/* STRATEGY CONTROLS TUNING CARD (SPAN 12 ON MD & LG; DEFERRED TO SIDEBAR SPAN 4 ON XL) */}
          <div className="md:col-span-12 xl:col-span-4 group relative flex flex-col h-auto xl:h-[460px] bg-surface-container-low border border-outline-variant/30 rounded-[32px] p-5 sm:p-8 overflow-hidden hover:bg-surface-container transition-all duration-700 ease-[cubic-bezier(0.16,1,0.3,1)] hover:shadow-[0_20px_40px_rgba(0,0,0,0.2)]">
            <div className="absolute top-0 inset-x-0 h-[1px] bg-gradient-to-r from-transparent via-secondary-fixed-dim/30 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-700" />
            
            <div className="mb-6 xl:mb-8 border-b border-outline-variant/20 pb-4 flex items-center justify-between">
              <h3 className="text-[10px] font-medium tracking-[0.2em] text-on-surface uppercase">
                Strategy Controls
              </h3>
              <span className="text-[9px] text-primary-container uppercase tracking-wider animate-pulse">Sync Active</span>
            </div>
            
            <div className="flex flex-col gap-6 flex-grow justify-center py-4 xl:py-0">
              {/* Risk Tolerance Selection Track */}
              <div className="flex flex-col gap-2">
                <div className="flex justify-between text-[10px] font-medium tracking-wider text-on-surface-variant">
                  <span>Risk Tolerance</span>
                  <span className="text-primary-container">{riskTolerance > 75 ? "High" : riskTolerance > 40 ? "Balanced" : "Conservative"}</span>
                </div>
                <input 
                  type="range" min="1" max="100" value={riskTolerance} 
                  onChange={(e) => setRiskTolerance(parseInt(e.target.value))}
                  className="w-full h-[2px] appearance-none bg-outline-variant/30 rounded-full outline-none accent-primary-container cursor-pointer"
                />
              </div>

              {/* Volatility Cap Selection Track */}
              <div className="flex flex-col gap-2">
                <div className="flex justify-between text-[10px] font-medium tracking-wider text-on-surface-variant">
                  <span>Volatility Cap</span>
                  <span className="text-primary-container">{volatilityCap}%</span>
                </div>
                <input 
                  type="range" min="5" max="50" value={volatilityCap} 
                  onChange={(e) => setVolatilityCap(parseInt(e.target.value))}
                  className="w-full h-[2px] appearance-none bg-outline-variant/30 rounded-full outline-none accent-primary-container cursor-pointer"
                />
              </div>

              {/* Leverage Multiplier Selection Track */}
              <div className="flex flex-col gap-2">
                <div className="flex justify-between text-[10px] font-medium tracking-wider text-on-surface-variant">
                  <span>Leverage Multiplier</span>
                  <span className="text-secondary-fixed">{calculatedLeverageText}</span>
                </div>
                <input 
                  type="range" min="0" max="100" value={leverageMultiplier} 
                  onChange={(e) => setLeverageMultiplier(parseInt(e.target.value))}
                  className="w-full h-[2px] appearance-none bg-outline-variant/30 rounded-full outline-none accent-secondary-fixed-dim cursor-pointer"
                />
              </div>
            </div>
          </div>

          {/* TELEMETRY LIVE EXECUTION LOG TERMINAL (SPAN 6 ON MD & LG COHORTS) */}
          <div className="md:col-span-6 group relative flex flex-col h-[380px] bg-surface-container-low border border-outline-variant/30 rounded-[32px] p-5 sm:p-8 overflow-hidden hover:bg-surface-container transition-all duration-700 ease-[cubic-bezier(0.16,1,0.3,1)] hover:shadow-[0_20px_40px_rgba(0,0,0,0.2)]">
            <h3 className="text-[10px] font-medium tracking-[0.2em] text-on-surface uppercase mb-5 flex items-center gap-2">
              <span className="material-symbols-outlined text-[16px] text-on-surface-variant">terminal</span>
              Live Execution Log
            </h3>
            
            <div className="bg-void-black/40 rounded-xl p-4 flex-grow overflow-y-auto terminal-scroll font-mono text-[11px] leading-relaxed text-on-surface-variant/80 border border-outline-variant/20">
              <div className="flex justify-between border-b border-outline-variant/20 pb-2 mb-2 text-on-surface-variant/40 font-medium tracking-wider">
                <span className="w-12 sm:w-16">TIME</span>
                <span className="w-16 sm:w-20">ASSET</span>
                <span className="w-12 sm:w-16">SIDE</span>
                <span className="w-20 sm:w-24 text-right">PRICE</span>
              </div>
              <div className="space-y-1">
                {mockExecutionLogs.map((log, index) => (
                  <div 
                    key={index} 
                    className={`flex justify-between py-1 px-0.5 rounded transition-colors duration-200 hover:bg-white/[0.02] ${
                      log.primary ? "text-primary-fixed-dim" : log.highlight ? "text-secondary-fixed" : ""
                    }`}
                  >
                    <span className="w-12 sm:w-16 opacity-60">{log.time}</span>
                    <span className="w-16 sm:w-20 font-medium tracking-wide">{log.asset}</span>
                    <span className="w-12 sm:w-16">{log.side}</span>
                    <span className="w-20 sm:w-24 text-right tracking-tight">{log.price}</span>
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* PORTFOLIO ALLOCATION DOCK CHART (SPAN 6 ON MD & LG COHORTS) */}
          <div className="md:col-span-6 group relative flex flex-col h-auto md:h-[380px] bg-surface-container-low border border-outline-variant/30 rounded-[32px] p-5 sm:p-8 overflow-hidden hover:bg-surface-container transition-all duration-700 ease-[cubic-bezier(0.16,1,0.3,1)] hover:shadow-[0_20px_40px_rgba(0,0,0,0.2)]">
            <h3 className="text-[10px] font-medium tracking-[0.2em] text-on-surface uppercase mb-6">
              Portfolio Allocation
            </h3>
            
            <div className="flex-grow flex flex-col sm:flex-row items-center justify-center gap-6 sm:gap-8 h-full py-4 sm:py-0">
              {/* DONUT CHART COMPONENT */}
              <div className="relative w-36 h-36 sm:w-40 sm:h-40 shrink-0 flex items-center justify-center">
                <svg className="absolute inset-0 transform -rotate-90" viewBox="0 0 100 100">
                  <circle cx="50" cy="50" r="40" fill="transparent" stroke="#2a2a2a" strokeWidth="12" />
                  <circle cx="50" cy="50" r="40" fill="transparent" stroke="#00f0ff" strokeWidth="12" 
                    strokeDasharray="251.2" strokeDashoffset="150.72" strokeLinecap="round"
                    style={{ filter: "drop-shadow(0 0 6px rgba(0,240,255,0.3))" }}
                  />
                  <circle cx="50" cy="50" r="40" fill="transparent" stroke="#ffd700" strokeWidth="12" 
                    strokeDasharray="251.2" strokeDashoffset="175.84" strokeLinecap="round"
                    style={{ filter: "drop-shadow(0 0 6px rgba(255,215,0,0.2))" }}
                    transform="rotate(144 50 50)"
                  />
                </svg>
                <div className="absolute flex flex-col items-center justify-center text-center">
                  <span className="text-[9px] font-medium tracking-widest text-on-surface-variant/50">TOTAL AUM</span>
                  <span className="text-xl font-light tracking-tight text-on-surface">$2.4M</span>
                </div>
              </div>

              {/* Metric Legend Items */}
              <div className="flex flex-col gap-3 flex-grow justify-center w-full max-w-[200px]">
                {allocationSegments.map((segment, index) => (
                  <div key={index} className="flex items-center justify-between text-xs font-light">
                    <div className="flex items-center gap-3">
                      <div className={`w-2 h-2 rounded-full ${segment.glowClass}`} />
                      <span className="text-on-surface-variant/80">{segment.label}</span>
                    </div>
                    <span className="font-mono text-on-surface/50 text-[11px]">({segment.percentage}%)</span>
                  </div>
                ))}
              </div>
            </div>
          </div>

        </div>
      </div>
    </AppShell>
  );
}