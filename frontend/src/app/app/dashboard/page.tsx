"use client";

import { useMemo, useState, useEffect } from "react";
import useSWR from "swr";
import { AppShell } from "@/components/app/AppShell";
import OnboardingOverlay from "@/components/app/OnboardingOverlay";
import { settingsApi } from "@/lib/api";
import { useAuth } from "@/context/AuthContext";
import { ServerPortfolioSummary, ServerTradeLog, SnapshotPoint } from "@/types/dashboard";

// Import custom dashboard widgets
import BalanceCard from "@/components/sections/dashboard/BalanceCard";
import StrategyControls from "@/components/sections/dashboard/StrategyControls";
import ExecutionLog from "@/components/sections/dashboard/ExecutionLog";
import PortfolioAllocation from "@/components/sections/dashboard/PortfolioAllocation";

const fetcher = <T,>(url: string): Promise<T> => 
  fetch(url).then((res) => {
    if (!res.ok) throw new Error(`Network response error: ${res.status}`);
    return res.json();
  });

export default function DashboardPage() {
  const [isAgentActive, setIsAgentActive] = useState<boolean>(true);
  const [showOnboarding, setShowOnboarding] = useState<boolean>(false);

  const [riskTolerance, setRiskTolerance] = useState<number>(80);
  const [volatilityCap, setVolatilityCap] = useState<number>(30);
  const [leverageMultiplier, setLeverageMultiplier] = useState<number>(50);

  const { user, loading } = useAuth();
  const currentUserID = user?.id || 999;

  // 📡 DYNAMIC SERVER TELEMETRY FETCHERS
  const { data: serverSummary } = useSWR<ServerPortfolioSummary>(
    currentUserID !== 999 ? `http://localhost:8080/api/v1/user/portfolio/summary?user_id=${currentUserID}` : null,
    fetcher
  );

  const { data: serverTrades } = useSWR<ServerTradeLog[]>(
    currentUserID !== 999 ? `http://localhost:8080/api/v1/user/portfolio/trades?user_id=${currentUserID}&limit=8` : null,
    fetcher
  );

    // 📡 Add this query under your existing SWR hooks to retrieve up to 30 historical timeline frames
  const { data: snapshotHistory } = useSWR<SnapshotPoint[]>(
    currentUserID !== 999 ? `http://localhost:8080/api/v1/user/portfolio/history?user_id=${currentUserID}&limit=30` : null,
    fetcher
  );
  // Synchronize layout posture configurations seamlessly
  useEffect(() => {
    if (loading) return;

    if (currentUserID === 999) {
      const demoBlueprint = localStorage.getItem("oa_demo_risk_posture");
      if (!demoBlueprint) {
        setShowOnboarding(true);
      } else {
        configureDashboardFromBlueprint(demoBlueprint);
      }
      return;
    }

    if (user) {
      if (!user.is_onboarded) {
        setShowOnboarding(true);
      } else {
        setShowOnboarding(false);

        const fetchRegistrationPosture = async () => {
          try {
            const response = await settingsApi.check(currentUserID);
            if (response.found && response.settings) {
              configureDashboardFromBlueprint(response.settings.risk_profile);
            }
          } catch (err) {
            console.error("Configuration payload initialization error:", err);
          }
        };

        fetchRegistrationPosture();
      }
    }
  }, [currentUserID, loading, user]);

  const configureDashboardFromBlueprint = (blueprint: string) => {
    if (blueprint === "conservative") {
      setRiskTolerance(25);
      setVolatilityCap(15);
      setLeverageMultiplier(0);
    } else if (blueprint === "moderate") {
      setRiskTolerance(60);
      setVolatilityCap(30);
      setLeverageMultiplier(25);
    } else if (blueprint === "aggressive") {
      setRiskTolerance(95);
      setVolatilityCap(45);
      setLeverageMultiplier(75);
    }
  };

  const handleOnboardingComplete = (finalProfile: string) => {
    configureDashboardFromBlueprint(finalProfile);
    if (user) user.is_onboarded = true;
    setShowOnboarding(false);
  };

  const calculatedLeverageText = useMemo(() => {
    return `${(1.0 + (leverageMultiplier / 100) * 4).toFixed(1)}x`;
  }, [leverageMultiplier]);

  const displayPnL = useMemo(() => {
    if (currentUserID === 999 || !serverSummary || serverSummary.change_dollar_24h === undefined) {
      return "+$12,450.89";
    }
    const prefix = serverSummary.change_dollar_24h >= 0 ? "+" : "";
    return `${prefix}$${serverSummary.change_dollar_24h.toLocaleString(undefined, { minimumFractionDigits: 2 })}`;
  }, [currentUserID, serverSummary]);

  if (loading) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center font-mono text-xs tracking-widest text-on-surface-variant/40 uppercase">
        Waking alpha up...
      </div>
    );
  }

  return (
    <AppShell title="Overview">
      {showOnboarding && <OnboardingOverlay userID={currentUserID} onComplete={handleOnboardingComplete} />}

      <div className="w-full bg-transparent flex flex-col gap-6 md:gap-10 animate-in fade-in duration-700 ease-[cubic-bezier(0.16,1,0.3,1)]">
        <div className="flex flex-col sm:flex-row justify-between items-start sm:items-end gap-4 pb-2">
          <div>
            <p className="text-xs sm:text-sm font-light text-on-surface-variant/70 mt-1">
              Real-time dashboard
            </p>
          </div>

          <button
            onClick={() => setIsAgentActive(!isAgentActive)}
            className="w-full sm:w-auto px-6 py-2.5 rounded-full text-xs font-medium tracking-wider uppercase shadow-md transition-all duration-500 active:scale-95 bg-primary-container text-black shadow-primary-container/20 hover:bg-primary-container/90"
          >
            {isAgentActive ? "Terminate Agent" : "Launch Agent"}
          </button>
        </div>

        {/* Bento Widgets Layer Grid Matrix */}
        <div className="grid grid-cols-1 md:grid-cols-12 gap-6 md:gap-8 items-start">
          <BalanceCard isAgentActive={isAgentActive} displayPnL={displayPnL} historyData={snapshotHistory} />
          
          <StrategyControls 
            riskTolerance={riskTolerance}
            setRiskTolerance={setRiskTolerance}
            volatilityCap={volatilityCap}
            setVolatilityCap={setVolatilityCap}
            leverageMultiplier={leverageMultiplier}
            setLeverageMultiplier={setLeverageMultiplier}
            calculatedLeverageText={calculatedLeverageText}
          />

          <ExecutionLog currentUserID={currentUserID} serverTrades={serverTrades} />

          <PortfolioAllocation currentUserID={currentUserID} serverSummary={serverSummary} />
        </div>
      </div>
    </AppShell>
  );
}