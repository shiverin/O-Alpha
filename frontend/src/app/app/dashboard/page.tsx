"use client";

import { useMemo, useRef, useState, useEffect } from "react";
import useSWR from "swr";
import { AppShell } from "@/components/app/AppShell";
import OnboardingOverlay from "@/components/app/OnboardingOverlay";
import {
  agentStatusApi,
  api,
  portfolioAgentApi,
  settingsApi,
  streamPortfolioLive,
  strategyCatalogApi,
  type StrategyCatalogResponse,
} from "@/lib/api";
import { useAuth } from "@/context/AuthContext";
import {
  ServerPositionMetrics,
  ServerPortfolioSummary,
  ServerTradeLog,
  SnapshotPoint,
} from "@/types/dashboard";

import BalanceCard from "@/components/sections/dashboard/BalanceCard";
import StrategyControls from "@/components/sections/dashboard/StrategyControls";
import ExecutionLog from "@/components/sections/dashboard/ExecutionLog";
import PortfolioAllocation from "@/components/sections/dashboard/PortfolioAllocation";
import {
  applyLivePriceToHistory,
  applyLivePriceToPositions,
  applyLivePriceToSummary,
} from "@/lib/portfolioLiveState";

const fetcher = <T,>(path: string): Promise<T> => api.get<T>(path);
type RiskProfile = "conservative" | "moderate" | "aggressive";
const REALTIME_REFRESH_MS = 15000;

const riskBuckets = {
  conservative: "low",
  moderate: "medium",
  aggressive: "high",
} as const;

export default function DashboardPage() {
  const [isDemoAgentActive, setIsDemoAgentActive] = useState<boolean>(false);
  const [agentActionPending, setAgentActionPending] = useState<boolean>(false);
  const [agentError, setAgentError] = useState<string | null>(null);
  const [showOnboarding, setShowOnboarding] = useState<boolean>(false);

  const [riskProfile, setRiskProfile] = useState<RiskProfile>("moderate");
  const [selectedStrategyKey, setSelectedStrategyKey] = useState<string>("");
  const [initialCash, setInitialCash] = useState<number>(100000);

  const { user, loading, markOnboarded } = useAuth();
  const currentUserID = user?.id || 999;
  const livePositionsRef = useRef<ServerPositionMetrics[] | undefined>();

  const { data: serverSummary, mutate: mutateSummary } =
    useSWR<ServerPortfolioSummary>(
      currentUserID !== 999 ? "/api/v1/user/portfolio/summary" : null,
      fetcher,
      { refreshInterval: REALTIME_REFRESH_MS },
    );

  const { data: serverTrades } = useSWR<ServerTradeLog[]>(
    currentUserID !== 999 ? "/api/v1/user/portfolio/trades?limit=8" : null,
    fetcher,
    { refreshInterval: REALTIME_REFRESH_MS },
  );

  const { data: snapshotHistory, mutate: mutateHistory } = useSWR<
    SnapshotPoint[]
  >(
    currentUserID !== 999 ? "/api/v1/user/portfolio/history?limit=30" : null,
    fetcher,
    { refreshInterval: REALTIME_REFRESH_MS },
  );

  const { data: serverPositions, mutate: mutatePositions } = useSWR<
    ServerPositionMetrics[]
  >(
    currentUserID !== 999 ? "/api/v1/user/portfolio/positions" : null,
    fetcher,
    { refreshInterval: REALTIME_REFRESH_MS },
  );

  const { data: strategyCatalog } = useSWR(
    currentUserID !== 999 ? "/api/v1/strategies/catalog" : null,
    () => strategyCatalogApi.list(),
  );

  const { data: agentList, mutate: refreshAgents } = useSWR(
    currentUserID !== 999 ? "/api/v1/agent/list" : null,
    () => agentStatusApi.list(),
    { refreshInterval: 15000 },
  );

  const activePortfolioAgent = useMemo(() => {
    return agentList?.agents?.find(
      (agent) => agent.strategy_type === "PORTFOLIO_CATALOG",
    );
  }, [agentList]);

  const isAgentActive =
    currentUserID === 999 ? isDemoAgentActive : Boolean(activePortfolioAgent);
  const liveStreamEnabled =
    currentUserID !== 999 && Boolean(activePortfolioAgent);

  const regimeLabel = useMemo(() => {
    const label = activePortfolioAgent?.runtime_state?.regime_label;
    if (!label || typeof label !== "string") {
      return "Syncing";
    }
    return label;
  }, [activePortfolioAgent]);

  useEffect(() => {
    if (serverPositions) {
      livePositionsRef.current = serverPositions;
    }
  }, [serverPositions]);

  useEffect(() => {
    if (!liveStreamEnabled) return;

    const controller = new AbortController();
    streamPortfolioLive((event) => {
      if (event.type === "snapshot") {
        if (event.summary) {
          void mutateSummary(event.summary, false);
        }
        if (event.positions) {
          livePositionsRef.current = event.positions;
          void mutatePositions(event.positions, false);
        }
        if (event.history) {
          void mutateHistory(event.history, false);
        }
        return;
      }
      if (event.type === "price") {
        const { positions, deltaExposure } = applyLivePriceToPositions(
          livePositionsRef.current,
          event,
        );
        livePositionsRef.current = positions;
        void mutatePositions(positions, false);
        void mutateSummary(
          (summary) =>
            applyLivePriceToSummary(summary, deltaExposure, event.timestamp),
          false,
        );
        void mutateHistory(
          (history) =>
            applyLivePriceToHistory(history, deltaExposure, event.timestamp),
          false,
        );
      }
    }, controller.signal).catch((err) => {
      if (err instanceof DOMException && err.name === "AbortError") return;
      console.error("Portfolio live stream failed:", err);
    });

    return () => controller.abort();
  }, [liveStreamEnabled, mutateHistory, mutatePositions, mutateSummary]);

  useEffect(() => {
    if (!strategyCatalog) return;
    const nextStrategyKey = dashboardStrategyKeyForRisk(
      strategyCatalog,
      riskProfile,
      selectedStrategyKey,
    );
    if (selectedStrategyKey !== nextStrategyKey) {
      setSelectedStrategyKey(nextStrategyKey);
    }
  }, [riskProfile, selectedStrategyKey, strategyCatalog]);

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
    const normalized =
      blueprint === "conservative" || blueprint === "aggressive"
        ? blueprint
        : "moderate";
    setRiskProfile(normalized);
    if (normalized === "conservative") {
      setInitialCash(50000);
    } else if (normalized === "moderate") {
      setInitialCash(100000);
    } else if (normalized === "aggressive") {
      setInitialCash(150000);
    }
  };

  const handleOnboardingComplete = (
    finalProfile: string,
    finalStrategyKey: string,
  ) => {
    configureDashboardFromBlueprint(finalProfile);
    setSelectedStrategyKey(finalStrategyKey);
    markOnboarded();
    setShowOnboarding(false);
  };

  const handleAgentToggle = async () => {
    setAgentError(null);

    if (currentUserID === 999) {
      setIsDemoAgentActive((active) => !active);
      return;
    }

    setAgentActionPending(true);
    try {
      if (isAgentActive) {
        await portfolioAgentApi.stop();
      } else {
        const launchStrategyKey = strategyCatalog
          ? dashboardStrategyKeyForRisk(
              strategyCatalog,
              riskProfile,
              selectedStrategyKey,
            )
          : "";
        if (launchStrategyKey && launchStrategyKey !== selectedStrategyKey) {
          setSelectedStrategyKey(launchStrategyKey);
        }
        await portfolioAgentApi.start({
          strategy_key: launchStrategyKey || "auto",
          symbols: strategyCatalog?.default_universe,
          risk_profile: riskProfile,
          timeframe: "1Day",
          initial_cash:
            serverSummary?.total_asset_value &&
            serverSummary.total_asset_value > 0
              ? serverSummary.total_asset_value
              : initialCash,
        });
      }
      await refreshAgents();
    } catch (err) {
      setAgentError(
        err instanceof Error ? err.message : "Agent control request failed.",
      );
    } finally {
      setAgentActionPending(false);
    }
  };

  const displayPnL = useMemo(() => {
    if (
      currentUserID === 999 ||
      !serverSummary ||
      serverSummary.change_dollar_24h === undefined
    ) {
      return "+$12,450.89";
    }
    const prefix = serverSummary.change_dollar_24h > 0 ? "+" : "";
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
            regimeLabel={regimeLabel}
          />

          <StrategyControls
            riskProfile={riskProfile}
            universeSize={strategyCatalog?.default_universe.length ?? 0}
          />

          <ExecutionLog
            currentUserID={currentUserID}
            serverTrades={serverTrades}
          />

          <PortfolioAllocation
            currentUserID={currentUserID}
            serverSummary={serverSummary}
            serverPositions={serverPositions}
          />
        </div>
      </div>
    </AppShell>
  );
}

function dashboardRecommendedStrategy(
  catalog: StrategyCatalogResponse,
  riskProfile: RiskProfile,
) {
  const riskBucket = riskBuckets[riskProfile];
  const recommendedKey = catalog.recommended[riskProfile];
  const recommended = catalog.strategies.find(
    (strategy) => strategy.key === recommendedKey,
  );
  if (recommended?.risk_profile === riskBucket) {
    return recommended.key;
  }
  return (
    catalog.strategies.find((strategy) => strategy.risk_profile === riskBucket)
      ?.key || ""
  );
}

function dashboardStrategyKeyForRisk(
  catalog: StrategyCatalogResponse,
  riskProfile: RiskProfile,
  currentStrategyKey: string,
) {
  const riskBucket = riskBuckets[riskProfile];
  const current = catalog.strategies.find(
    (strategy) => strategy.key === currentStrategyKey,
  );
  if (current?.risk_profile === riskBucket) {
    return current.key;
  }
  return dashboardRecommendedStrategy(catalog, riskProfile);
}
