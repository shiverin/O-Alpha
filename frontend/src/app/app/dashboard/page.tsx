"use client";

import { useMemo, useState, useEffect } from "react";
import useSWR from "swr";
import { AppShell } from "@/components/app/AppShell";
import OnboardingOverlay from "@/components/app/OnboardingOverlay";
import { agentApi, api, settingsApi } from "@/lib/api";
import { useAuth } from "@/context/AuthContext";
import {
  ServerPortfolioSummary,
  ServerTradeLog,
  SnapshotPoint,
} from "@/types/dashboard";

import BalanceCard from "@/components/sections/dashboard/BalanceCard";
import StrategyControls from "@/components/sections/dashboard/StrategyControls";
import ExecutionLog from "@/components/sections/dashboard/ExecutionLog";
import PortfolioAllocation from "@/components/sections/dashboard/PortfolioAllocation";

const fetcher = <T,>(path: string): Promise<T> => api.get<T>(path);

export default function DashboardPage() {
  const [isAgentActive, setIsAgentActive] = useState<boolean>(false);
  const [agentActionPending, setAgentActionPending] = useState<boolean>(false);
  const [agentError, setAgentError] = useState<string | null>(null);
  const [showOnboarding, setShowOnboarding] = useState<boolean>(false);

  const [riskTolerance, setRiskTolerance] = useState<number>(80);
  const [volatilityCap, setVolatilityCap] = useState<number>(30);
  const [leverageMultiplier, setLeverageMultiplier] = useState<number>(50);

  const { user, loading, markOnboarded } = useAuth();
  const currentUserID = user?.id || 999;

  const { data: serverSummary } = useSWR<ServerPortfolioSummary>(
    currentUserID !== 999 ? "/api/v1/user/portfolio/summary" : null,
    fetcher,
  );

  const { data: serverTrades } = useSWR<ServerTradeLog[]>(
    currentUserID !== 999 ? "/api/v1/user/portfolio/trades?limit=8" : null,
    fetcher,
  );

  const { data: snapshotHistory } = useSWR<SnapshotPoint[]>(
    currentUserID !== 999 ? "/api/v1/user/portfolio/history?limit=30" : null,
    fetcher,
  );

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
            const response = await settingsApi.check();
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
    markOnboarded();
    setShowOnboarding(false);
  };

  const handleAgentToggle = async () => {
    setAgentError(null);

    if (currentUserID === 999) {
      setIsAgentActive((active) => !active);
      return;
    }

    setAgentActionPending(true);
    try {
      if (isAgentActive) {
        await agentApi.stop("AAPL");
        setIsAgentActive(false);
      } else {
        await agentApi.start({
          symbol: "AAPL",
          strategy_type: "KALMAN",
          timeframe: "1Hour",
          initial_cash: serverSummary?.total_asset_value ?? 50000,
          q_noise: 0.01,
          r_noise: 0.5,
          z_threshold: 2,
        });
        setIsAgentActive(true);
      }
    } catch (err) {
      setAgentError(
        err instanceof Error ? err.message : "Agent control request failed.",
      );
    } finally {
      setAgentActionPending(false);
    }
  };

  const calculatedLeverageText = useMemo(() => {
    return `${(1.0 + (leverageMultiplier / 100) * 4).toFixed(1)}x`;
  }, [leverageMultiplier]);

  const displayPnL = useMemo(() => {
    if (
      currentUserID === 999 ||
      !serverSummary ||
      serverSummary.change_dollar_24h === undefined
    ) {
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
      {showOnboarding && (
        <OnboardingOverlay
          userID={currentUserID}
          onComplete={handleOnboardingComplete}
        />
      )}

      <div className="w-full bg-transparent flex flex-col gap-6 md:gap-10 animate-in fade-in duration-700 ease-[cubic-bezier(0.16,1,0.3,1)]">
        <div className="flex flex-col sm:flex-row justify-between items-start sm:items-end gap-4 pb-2">
          <div>
            <p className="text-xs sm:text-sm font-light text-on-surface-variant/70 mt-1">
              Real-time dashboard
            </p>
          </div>

          <button
            onClick={handleAgentToggle}
            disabled={agentActionPending}
            className="w-full sm:w-auto px-6 py-2.5 rounded-full text-xs font-medium tracking-wider uppercase shadow-md transition-all duration-500 active:scale-95 bg-primary-container text-black shadow-primary-container/20 hover:bg-primary-container/90 disabled:cursor-not-allowed disabled:opacity-60"
          >
            {agentActionPending
              ? "Synchronizing"
              : isAgentActive
                ? "Terminate Agent"
                : "Launch Agent"}
          </button>
        </div>

        {agentError && (
          <div className="rounded-xl border border-error/30 bg-error/5 px-4 py-3 text-xs font-mono tracking-wide text-error">
            {agentError}
          </div>
        )}

        <div className="grid grid-cols-1 md:grid-cols-12 gap-6 md:gap-8 items-start">
          <BalanceCard
            isAgentActive={isAgentActive}
            displayPnL={displayPnL}
            historyData={snapshotHistory}
          />

          <StrategyControls
            riskTolerance={riskTolerance}
            setRiskTolerance={setRiskTolerance}
            volatilityCap={volatilityCap}
            setVolatilityCap={setVolatilityCap}
            leverageMultiplier={leverageMultiplier}
            setLeverageMultiplier={setLeverageMultiplier}
            calculatedLeverageText={calculatedLeverageText}
          />

          <ExecutionLog
            currentUserID={currentUserID}
            serverTrades={serverTrades}
          />

          <PortfolioAllocation
            currentUserID={currentUserID}
            serverSummary={serverSummary}
          />
        </div>
      </div>
    </AppShell>
  );
}
