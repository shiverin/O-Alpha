"use client";

import { ServerTradeLog, MockLogItem } from "@/types/dashboard";
import { mockExecutionLogs } from "@/lib/mockAppData";

interface ExecutionLogProps {
  currentUserID: number;
  serverTrades: ServerTradeLog[] | undefined;
}

export default function ExecutionLog({ currentUserID, serverTrades }: ExecutionLogProps) {
  // 1. Determine if we should render the guest sandbox telemetry loop
  const isGuestMode = currentUserID === 999;
  
  // 2. Safely evaluate if our SWR network pipeline is still fetching
  const isLoading = !isGuestMode && serverTrades === undefined;
  
  // 3. Detect if a real user has landed with a completely fresh account history
  const isEmpty = !isGuestMode && serverTrades && serverTrades.length === 0;

  return (
    <div className="md:col-span-12 group relative flex flex-col h-[380px] bg-surface-container-low border border-outline-variant/30 rounded-[32px] p-5 sm:p-8 overflow-hidden hover:bg-surface-container transition-all duration-700 hover:shadow-[0_20px_40px_rgba(0,0,0,0.2)]">
      <h3 className="text-[10px] font-medium tracking-[0.2em] text-on-surface uppercase mb-5 flex items-center gap-2">
        <span className="material-symbols-outlined text-[16px] text-on-surface-variant animate-pulse">terminal</span>
        Live Execution Log
      </h3>
      
      <div className="bg-void-black/40 rounded-xl p-4 flex-grow overflow-y-auto terminal-scroll font-mono text-[11px] leading-relaxed text-on-surface-variant/80 border border-outline-variant/20">
        
        {/* Table Headers (Only show if we aren't loading or empty) */}
        {!isLoading && !isEmpty && (
          <div className="flex justify-between border-b border-outline-variant/20 pb-2 mb-2 text-on-surface-variant/40 font-medium tracking-wider">
            <span className="w-12 sm:w-16">TIME</span>
            <span className="w-16 sm:w-20">ASSET</span>
            <span className="w-12 sm:w-16">SIDE</span>
            <span className="w-20 sm:w-24 text-right">PRICE</span>
          </div>
        )}

        <div className="space-y-1">
          {/* STATE A: CURRENT USER IS THE GUEST SANDBOX ENGINE */}
          {isGuestMode && 
            mockExecutionLogs.map((log: MockLogItem, index: number) => (
              <div key={index} className={`flex justify-between py-1 px-0.5 rounded transition-colors duration-200 hover:bg-white/[0.02] ${log.primary ? "text-primary-fixed-dim" : log.highlight ? "text-secondary-fixed" : ""}`}>
                <span className="w-12 sm:w-16 opacity-60">{log.time}</span>
                <span className="w-16 sm:w-20 font-medium tracking-wide">{log.asset}</span>
                <span className="w-12 sm:w-16">{log.side}</span>
                <span className="w-20 sm:w-24 text-right tracking-tight">{log.price}</span>
              </div>
            ))
          }

          {/* STATE B: NETWORK WIRE IS LOADING IN FLIGHT */}
          {isLoading && (
            <div className="h-full w-full flex items-center justify-center text-xs text-primary-container/40 uppercase tracking-widest animate-pulse py-20">
              ⚡ Connecting core telemetry streaming nodes...
            </div>
          )}

          {/* STATE C: REAL USER WITH 0 HISTORICAL TRADES */}
          {isEmpty && (
            <div className="flex flex-col gap-2 py-12 px-4 text-left text-on-surface-variant/50">
              <p className="text-primary-fixed-dim text-xs font-semibold select-none">[SYSTEM] SECURE CORE TERMINAL LINK OK.</p>
              <p className="text-[11px] font-light leading-relaxed">
                &gt; No transactional execution histories discovered for User #{currentUserID}.<br />
                &gt; Awaiting real-time market engine alerts or live execution triggers...
              </p>
              <div className="w-2 h-3 bg-primary-container/80 animate-pulse mt-2" />
            </div>
          )}

          {/* STATE D: PRODUCTION LIVE ACCOUNT TRACKING REDUCTION MAP */}
          {!isGuestMode && serverTrades && serverTrades.map((log: ServerTradeLog, index: number) => {
            const displayTime = log.timestamp 
              ? new Date(log.timestamp).toLocaleTimeString(undefined, { hour: "2-digit", minute: "2-digit", second: "2-digit" }) 
              : "Live";
              
            const isBuy = log.action.startsWith("BUY");
            const textClass = isBuy ? "text-primary-fixed-dim" : "text-error";
            const sideLabel = log.action.split("_")[0];

            return (
              <div key={index} className="flex justify-between py-1 px-0.5 rounded transition-colors duration-200 hover:bg-white/[0.02]">
                <span className="w-12 sm:w-16 opacity-60 text-on-surface-variant/60">{displayTime}</span>
                <span className="w-16 sm:w-20 font-medium tracking-wide text-on-surface">{log.symbol}</span>
                <span className={`w-12 sm:w-16 font-medium ${textClass}`}>{sideLabel}</span>
                <span className="w-20 sm:w-24 text-right tracking-tight text-on-surface-variant">${log.price.toFixed(2)}</span>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}