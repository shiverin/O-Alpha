"use client";

import { useState, useEffect } from "react";
import { settingsApi } from "@/lib/api";

interface OnboardingOverlayProps {
  userID: number;
  onComplete: () => void;
}

export default function OnboardingOverlay({ userID, onComplete }: OnboardingOverlayProps) {
  const [isVisible, setIsVisible] = useState(false);
  const [step, setStep] = useState(1);
  const [riskProfile, setRiskProfile] = useState("moderate");
  const [isSaving, setIsSaving] = useState(false);

  const [flippedCards, setFlippedCards] = useState<Record<string, boolean>>({
    conservative: false,
    moderate: false,
    aggressive: false,
  });

  useEffect(() => {
    const timer = setTimeout(() => setIsVisible(true), 150);
    return () => clearTimeout(timer);
  }, []);

  const profileDescriptions: Record<"conservative" | "moderate" | "aggressive", string> = {
    conservative:
      "Prioritizes structural capital preservation. Minimizes drawdown lengths using tight trailing stops and allocation focuses strictly on low-beta asset pairings.",
    moderate:
      "Engineered for optimal risk-adjusted discovery. Deploys dynamic position scaling rules and active momentum tracking filters to navigate regime transitions.",
    aggressive:
      "Optimized for high-volatility statistical arbitrage workflows. Executes maximum leverage thresholds alongside complex options configurations to yield convex returns.",
  };

  const toggleCardFlip = (profile: string, e: React.MouseEvent) => {
    e.stopPropagation();
    setFlippedCards((prev) => ({ ...prev, [profile]: !prev[profile] }));
  };

  const handleInitializeCore = async () => {
    setIsSaving(true);

    // ─────────────────────────────────────────────────────────────
    // ✅ BRANCH 1: LOCAL STORAGE BYPASS GATE FOR DEMO SESSIONS (userID: 999)
    // ─────────────────────────────────────────────────────────────
    if (userID === 999) {
      await new Promise((resolve) => setTimeout(resolve, 800)); // Maintain premium snappy loading simulation
      localStorage.setItem("oa_demo_risk_posture", riskProfile);
      setIsSaving(false);
      onComplete();
      return;
    }

    // ─────────────────────────────────────────────────────────────
    // 🌍 BRANCH 2: PRODUCTION DATABASE SYNC PIPELINE (REAL USERS)
    // ─────────────────────────────────────────────────────────────
    let configPayload = {
      user_id: userID,
      risk_profile: riskProfile,
      leverage: 2,
      max_positions: 6,
      stop_loss_pct: 2.5,
      take_profit_pct: 5.0,
      rebalance_freq: "daily",
    };

    if (riskProfile === "conservative") {
      configPayload = {
        ...configPayload,
        leverage: 1,
        max_positions: 3,
        stop_loss_pct: 1.5,
        take_profit_pct: 3.0,
        rebalance_freq: "weekly",
      };
    } else if (riskProfile === "aggressive") {
      configPayload = {
        ...configPayload,
        leverage: 4,
        max_positions: 12,
        stop_loss_pct: 4.0,
        take_profit_pct: 12.0,
        rebalance_freq: "hourly",
      };
    }

    try {
      await settingsApi.save(configPayload);
      setIsSaving(false);
      onComplete();
    } catch {
      setIsSaving(false);
      alert("Failed to synchronize execution parameters with the server.");
    }
  };

  if (!isVisible) return null;

  return (
    <div className="fixed inset-0 z-[9999] bg-background/60 backdrop-blur-2xl flex items-center justify-center p-4 sm:p-6 transition-all duration-1000 ease-out">
      <div className="absolute inset-0 opacity-[0.03] pointer-events-none" style={{ backgroundImage: "radial-gradient(circle at 1px 1px, white 1px, transparent 0)", backgroundSize: "32px 32px" }} />
      <div className="w-full max-w-4xl bg-surface-container-low border border-outline-variant/30 rounded-[32px] p-6 sm:p-10 shadow-[0_30px_70px_rgba(0,0,0,0.6)] relative overflow-hidden transition-all duration-500 scale-100">
        <div className="absolute top-0 inset-x-0 h-[1px] bg-gradient-to-r from-transparent via-primary-fixed-dim/40 to-transparent" />

        {step === 1 && (
          <div className="flex flex-col items-center text-center py-10 max-w-2xl mx-auto">
            <div className="h-16 w-16 rounded-2xl flex items-center justify-center mb-8 bg-white/[0.02] border border-outline-variant/20">
              <span className="material-symbols-outlined text-primary-fixed-dim text-3xl animate-pulse">token</span>
            </div>
            <h1 className="text-3xl sm:text-4xl font-light tracking-tight text-on-surface mb-4">
              Welcome to <span className="text-primary-fixed-dim font-normal">O(Alpha)</span>
            </h1>
            <p className="text-sm sm:text-base font-light text-on-surface-variant/70 leading-relaxed mb-10">
              Deploy your quant-level autonomous agent infrastructure. Initialize corporate engine posture metrics below to open up your telemetry overview terminal.
            </p>
            <button onClick={() => setStep(2)} className="px-8 py-3.5 bg-primary-container text-void-black font-mono font-medium text-xs tracking-wider uppercase rounded-full shadow-lg hover:bg-primary-fixed transition-all duration-300">
              Begin Configuration
            </button>
          </div>
        )}

        {step === 2 && (
          <div className="flex flex-col">
            <div className="mb-8 text-center sm:text-left">
              <span className="text-[10px] font-mono tracking-[0.25em] text-primary-fixed-dim uppercase block mb-1">Settings</span>
              <h2 className="text-2xl font-light tracking-tight text-on-surface">Select Your Risk Profile</h2>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-6 w-full mb-10">
              {(["conservative", "moderate", "aggressive"] as const).map((profile) => {
                const isSelected = riskProfile === profile;
                const isFlipped = flippedCards[profile];

                return (
                  <div key={profile} onClick={() => setRiskProfile(profile)} className="[perspective:1000px] h-48 w-full cursor-pointer select-none">
                    <div className={`relative w-full h-full transition-transform duration-500 [transform-style:preserve-3d] ${isFlipped ? "[transform:rotateY(180deg)]" : ""}`}>
                      <div className={`absolute inset-0 [backface-visibility:hidden] flex flex-col justify-center items-center bg-void-black/20 border rounded-2xl p-6 transition-all duration-300 ${isSelected ? "border-primary-fixed-dim shadow-[0_0_20px_rgba(0,240,255,0.08)] bg-surface-container" : "border-outline-variant/20 hover:border-outline-variant/50"}`}>
                        <button type="button" onClick={(e) => toggleCardFlip(profile, e)} className="absolute right-4 top-4 text-on-surface-variant/30 hover:text-primary-fixed-dim transition-colors h-7 w-7 rounded-full flex items-center justify-center hover:bg-white/5"><span className="material-symbols-outlined text-[18px]">help</span></button>
                        <h4 className={`text-base font-light tracking-widest uppercase ${isSelected ? "text-primary-fixed-dim font-medium" : "text-on-surface-variant/70"}`}>{profile}</h4>
                      </div>
                      <div className="absolute inset-0 [backface-visibility:hidden] [transform:rotateY(180deg)] flex flex-col justify-center bg-surface-container-high border border-outline-variant/40 rounded-2xl p-6 shadow-xl">
                        <button type="button" onClick={(e) => toggleCardFlip(profile, e)} className="absolute right-4 top-4 text-primary-fixed-dim/70 hover:text-primary-fixed-dim h-7 w-7 rounded-full flex items-center justify-center bg-white/5"><span className="material-symbols-outlined text-[16px]">flip_to_front</span></button>
                        <p className="text-xs font-light leading-relaxed text-on-surface-variant/90 pr-4 select-text">{profileDescriptions[profile]}</p>
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>

            <div className="flex flex-col sm:flex-row justify-between items-center border-t border-outline-variant/10 pt-6 gap-4">
              <div className="flex gap-4 w-full sm:w-auto ml-auto">
                <button onClick={() => setStep(1)} disabled={isSaving} className="flex-1 sm:flex-none px-6 py-2.5 border border-outline-variant/30 text-on-surface-variant font-mono text-xs tracking-wider uppercase rounded-full hover:text-on-surface transition-colors">Back</button>
                <button onClick={handleInitializeCore} disabled={isSaving} className="flex-1 sm:flex-none px-8 py-2.5 bg-primary-container text-void-black font-mono font-medium text-xs tracking-wider uppercase rounded-full disabled:opacity-50 flex items-center justify-center gap-2 shadow-md">
                  {isSaving ? "Synchronizing..." : "Confirm Deployment"}
                </button>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}