"use client";

import { useMemo, useState, useEffect } from "react";
import { AppShell } from "@/components/app/AppShell";
import { EquityCurveChart } from "@/components/EquityCurveChart";
import {
  agentStatusApi,
  runBacktestStream,
  settingsApi,
  strategyCatalogApi,
  type BacktestResult,
  type EquityPoint,
  type StrategyCatalogResponse,
  type StrategySpec,
} from "@/lib/api";
import { useAuth } from "@/context/AuthContext";

type RiskProfile = "conservative" | "moderate" | "aggressive";
type RiskBucket = "low" | "medium" | "high";

const BACKTEST_INITIAL_CASH = 100000;

const riskBuckets: Record<RiskProfile, RiskBucket> = {
  conservative: "low",
  moderate: "medium",
  aggressive: "high",
};

const fallbackUniverse = [
  "VOO",
  "AAPL",
  "MSFT",
  "NVDA",
  "AMZN",
  "META",
  "GOOGL",
  "QQQ",
  "SPY",
  "IWM",
  "XLK",
  "XLF",
];

const fallbackCatalog: StrategyCatalogResponse = {
  default_universe: fallbackUniverse,
  recommended: {
    conservative: "ranker_proxy_h63_low",
    moderate: "lgbm_ranker_h63_medium",
    aggressive: "composite_momentum_high",
  },
  strategies: [
    {
      key: "ranker_proxy_h63_low",
      display_name: "Deterministic h63 proxy low risk",
      family: "benchmark_ranker_proxy_h63",
      risk_profile: "low",
      deployment_status: "conservative_variant",
      promoted_checkpoint: false,
      requires_model_artifacts: false,
      paper_only: true,
      benchmark_symbol: "VOO",
      description: "VOO core with an 8% deterministic h63 active sleeve.",
    },
    {
      key: "lgbm_ranker_h63_medium",
      display_name: "LGBM h63 active sleeve medium risk",
      family: "benchmark_lgbm_ranker_h63",
      risk_profile: "medium",
      deployment_status: "promoted_research_checkpoint",
      promoted_checkpoint: true,
      requires_model_artifacts: true,
      paper_only: true,
      benchmark_symbol: "VOO",
      description:
        "VOO core with a 15% learned-ranker active sleeve, top 3 stocks, 63-bar rebalance.",
    },
    {
      key: "composite_momentum_high",
      display_name: "Composite momentum high risk",
      family: "composite_momentum",
      risk_profile: "high",
      deployment_status: "rejected_diagnostic",
      promoted_checkpoint: false,
      requires_model_artifacts: false,
      paper_only: true,
      benchmark_symbol: "VOO",
      description:
        "Higher-active-weight composite momentum sleeve across ETFs and stocks.",
    },
  ],
};

export default function AgentSettingsPage() {
  const [isAdvanced, setIsAdvanced] = useState(false);
  const [isSaving, setIsSaving] = useState(false);

  const [riskProfile, setRiskProfile] = useState<RiskProfile>("moderate");
  const [savedRiskProfile, setSavedRiskProfile] =
    useState<RiskProfile>("moderate");
  const [catalog, setCatalog] =
    useState<StrategyCatalogResponse>(fallbackCatalog);
  const [selectedStrategyKey, setSelectedStrategyKey] = useState("");
  const [isCatalogLoading, setIsCatalogLoading] = useState(false);
  const [isPortfolioAgentRunning, setIsPortfolioAgentRunning] = useState(false);
  const [settingsError, setSettingsError] = useState<string | null>(null);
  const [backtestResult, setBacktestResult] = useState<BacktestResult | null>(
    null,
  );
  const [streamingEquityCurve, setStreamingEquityCurve] = useState<
    EquityPoint[]
  >([]);
  const [backtestProgress, setBacktestProgress] = useState(0);
  const [backtestStatus, setBacktestStatus] = useState<string | null>(null);
  const [backtestError, setBacktestError] = useState<string | null>(null);
  const [isBacktesting, setIsBacktesting] = useState(false);
  const [acceptedStrategyKey, setAcceptedStrategyKey] = useState("");

  const [flippedCards, setFlippedCards] = useState<{ [key: string]: boolean }>({
    conservative: false,
    moderate: false,
    aggressive: false,
  });

  const [leverage, setLeverage] = useState(2);
  const [maxPositions, setMaxPositions] = useState(6);
  const [stopLoss, setStopLoss] = useState(2.5);
  const [takeProfit, setTakeProfit] = useState(5.0);
  const [rebalanceFreq, setRebalanceFreq] = useState("daily");

  const { user } = useAuth();
  const currentUserID = user?.id || 999;
  const riskProfileChanged = riskProfile !== savedRiskProfile;

  const strategiesForRisk = useMemo(() => {
    const bucket = riskBuckets[riskProfile];
    return catalog.strategies.filter(
      (strategy) => strategy.risk_profile === bucket,
    );
  }, [catalog.strategies, riskProfile]);

  const selectedStrategy = useMemo(() => {
    const bucket = riskBuckets[riskProfile];
    const selected = catalog.strategies.find(
      (strategy) => strategy.key === selectedStrategyKey,
    );
    if (selected?.risk_profile === bucket) {
      return selected;
    }
    return strategiesForRisk[0];
  }, [catalog.strategies, riskProfile, selectedStrategyKey, strategiesForRisk]);

  const metrics = useMemo(() => {
    if (!backtestResult) return null;
    const nested = backtestResult.metrics;
    return {
      totalReturn:
        nested?.total_return ??
        backtestResult.total_return ??
        equityReturn(backtestResult),
      sharpe: nested?.sharpe ?? backtestResult.sharpe ?? 0,
      maxDrawdown: nested?.max_drawdown ?? backtestResult.max_drawdown ?? 0,
      numTrades: nested?.num_trades ?? backtestResult.num_trades ?? 0,
    };
  }, [backtestResult]);

  const chartData = backtestResult?.equity_curve ?? streamingEquityCurve;
  const riskBacktestAccepted =
    !riskProfileChanged ||
    (Boolean(backtestResult) && acceptedStrategyKey === selectedStrategyKey);

  useEffect(() => {
    const loadCurrentConfigurationState = async () => {
      if (currentUserID === 999) {
        const demoProfile = normalizeRiskProfile(
          localStorage.getItem("oa_demo_risk_posture") || "moderate",
        );
        setRiskProfile(demoProfile);
        setSavedRiskProfile(demoProfile);
        setLeverage(
          Number(localStorage.getItem("oa_demo_leverage")) ||
            (demoProfile === "conservative"
              ? 1
              : demoProfile === "aggressive"
                ? 4
                : 2),
        );
        setMaxPositions(
          Number(localStorage.getItem("oa_demo_max_positions")) ||
            (demoProfile === "conservative"
              ? 3
              : demoProfile === "aggressive"
                ? 12
                : 6),
        );
        setStopLoss(
          Number(localStorage.getItem("oa_demo_stop_loss_pct")) ||
            (demoProfile === "conservative"
              ? 1.5
              : demoProfile === "aggressive"
                ? 4.0
                : 2.5),
        );
        setTakeProfit(
          Number(localStorage.getItem("oa_demo_take_profit_pct")) ||
            (demoProfile === "conservative"
              ? 3.0
              : demoProfile === "aggressive"
                ? 12.0
                : 5.0),
        );
        setRebalanceFreq(
          localStorage.getItem("oa_demo_rebalance_freq") ||
            (demoProfile === "conservative"
              ? "weekly"
              : demoProfile === "aggressive"
                ? "hourly"
                : "daily"),
        );
        return;
      }

      try {
        const response = await settingsApi.check();
        if (response.found && response.settings) {
          const profile = normalizeRiskProfile(response.settings.risk_profile);
          setRiskProfile(profile);
          setSavedRiskProfile(profile);
          setLeverage(response.settings.leverage);
          setMaxPositions(response.settings.max_positions);
          setStopLoss(response.settings.stop_loss_pct);
          setTakeProfit(response.settings.take_profit_pct);
          setRebalanceFreq(response.settings.rebalance_freq);
        }
      } catch (err) {
        console.error("Failed to read parameters from cloud database:", err);
      }
    };

    loadCurrentConfigurationState();
  }, [currentUserID]);

  useEffect(() => {
    const loadCatalog = async () => {
      setIsCatalogLoading(true);
      try {
        const response = await strategyCatalogApi.list();
        setCatalog(response);
        setSelectedStrategyKey(
          recommendedStrategyKeyForRisk(response, riskProfile),
        );
      } catch {
        setCatalog(fallbackCatalog);
        setSelectedStrategyKey(
          recommendedStrategyKeyForRisk(fallbackCatalog, riskProfile),
        );
      } finally {
        setIsCatalogLoading(false);
      }
    };

    loadCatalog();
  }, [riskProfile]);

  useEffect(() => {
    if (currentUserID === 999) return;

    let cancelled = false;
    const refreshRunningState = async () => {
      try {
        const response = await agentStatusApi.list();
        if (cancelled) return;
        setIsPortfolioAgentRunning(
          response.agents?.some(
            (agent) => agent.strategy_type === "PORTFOLIO_CATALOG",
          ) ?? false,
        );
      } catch (err) {
        console.error("Failed to read active agent state:", err);
      }
    };

    refreshRunningState();
    const interval = window.setInterval(refreshRunningState, 15000);
    return () => {
      cancelled = true;
      window.clearInterval(interval);
    };
  }, [currentUserID]);

  const handleProfileSelection = (profile: RiskProfile) => {
    setSettingsError(null);
    if (profile !== savedRiskProfile && isPortfolioAgentRunning) {
      setSettingsError(
        "Stop the running portfolio agent before changing risk profile.",
      );
      return;
    }
    setRiskProfile(profile);
    setBacktestResult(null);
    setStreamingEquityCurve([]);
    setBacktestProgress(0);
    setBacktestStatus(null);
    setBacktestError(null);
    setAcceptedStrategyKey("");
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
    e.stopPropagation();
    setFlippedCards((prev) => ({ ...prev, [profile]: !prev[profile] }));
  };

  const handleStrategySelect = (strategy: StrategySpec) => {
    setSelectedStrategyKey(strategy.key);
    setBacktestResult(null);
    setStreamingEquityCurve([]);
    setBacktestProgress(0);
    setBacktestStatus(null);
    setBacktestError(null);
    setAcceptedStrategyKey("");
  };

  const handleRunBacktest = async () => {
    if (!selectedStrategy) {
      setBacktestError(
        "No catalog strategy is available for this risk profile.",
      );
      return;
    }
    setIsBacktesting(true);
    setBacktestError(null);
    setBacktestResult(null);
    setAcceptedStrategyKey("");
    setStreamingEquityCurve([]);
    setBacktestProgress(0);
    setBacktestStatus("Preparing historical data...");

    const end = new Date();
    const start = new Date(end);
    start.setFullYear(end.getFullYear() - 5);
    setStreamingEquityCurve([
      { time: start.toISOString(), equity: BACKTEST_INITIAL_CASH },
    ]);

    let receivedProgress = false;
    let streamedPointCount = 1;

    try {
      const result = await runBacktestStream(
        {
          symbol:
            selectedStrategy.benchmark_symbol || catalog.default_universe[0],
          symbols: catalog.default_universe,
          start: start.toISOString(),
          end: end.toISOString(),
          strategy_type: "PORTFOLIO_CATALOG",
          timeframe: "1Day",
          initial_cash: BACKTEST_INITIAL_CASH,
          parameters: {
            strategy_key: selectedStrategy.key,
            max_gross_exposure: leverage,
            max_net_exposure: leverage,
            max_symbol_weight: 1,
          },
        },
        (event) => {
          if (event.type === "started") {
            setBacktestStatus("Loading aligned daily bars...");
            setBacktestProgress(0);
            return;
          }
          if (event.type === "progress") {
            setBacktestStatus("Running simulation...");
            if (!receivedProgress) {
              receivedProgress = true;
              streamedPointCount = 1;
              setStreamingEquityCurve([event.progress.point]);
            } else {
              streamedPointCount += 1;
              setStreamingEquityCurve((points) => [
                ...points,
                event.progress.point,
              ]);
            }
            setBacktestProgress(event.progress.percent);
          }
        },
      );
      setBacktestResult(result);
      if (streamedPointCount < result.equity_curve.length) {
        setBacktestStatus("Rendering equity curve...");
        await revealEquityCurve(result.equity_curve, streamedPointCount, {
          setCurve: setStreamingEquityCurve,
          setProgress: setBacktestProgress,
        });
      } else {
        setStreamingEquityCurve(result.equity_curve);
      }
      setBacktestStatus("Backtest complete.");
      setBacktestProgress(1);
    } catch (err) {
      setBacktestError(
        err instanceof Error
          ? err.message
          : "Backtest failed for the selected strategy.",
      );
    } finally {
      setIsBacktesting(false);
    }
  };

  const handleAcceptBacktest = () => {
    if (!backtestResult || !selectedStrategy) return;
    setAcceptedStrategyKey(selectedStrategy.key);
    setSettingsError(null);
  };

  const handleSave = async () => {
    setSettingsError(null);
    if (riskProfileChanged && isPortfolioAgentRunning) {
      setSettingsError(
        "Stop the running portfolio agent before changing risk profile.",
      );
      return;
    }
    if (riskProfileChanged && !riskBacktestAccepted) {
      setSettingsError(
        "Run and accept a matching strategy backtest before saving this risk profile.",
      );
      return;
    }
    setIsSaving(true);

    if (currentUserID === 999) {
      await new Promise((resolve) => setTimeout(resolve, 600));
      localStorage.setItem("oa_demo_risk_posture", riskProfile);
      localStorage.setItem("oa_demo_leverage", leverage.toString());
      localStorage.setItem("oa_demo_max_positions", maxPositions.toString());
      localStorage.setItem("oa_demo_stop_loss_pct", stopLoss.toString());
      localStorage.setItem("oa_demo_take_profit_pct", takeProfit.toString());
      localStorage.setItem("oa_demo_rebalance_freq", rebalanceFreq);
      if (selectedStrategyKey) {
        localStorage.setItem(
          "oa_demo_onboarding_strategy",
          selectedStrategyKey,
        );
      }
      setSavedRiskProfile(riskProfile);
      setAcceptedStrategyKey("");
      setIsSaving(false);
      alert("Demo frame settings synchronized successfully.");
      return;
    }

    const configPayload = {
      risk_profile: riskProfile,
      leverage: leverage,
      max_positions: maxPositions,
      stop_loss_pct: stopLoss,
      take_profit_pct: takeProfit,
      rebalance_freq: rebalanceFreq,
      strategy_key: riskProfileChanged ? selectedStrategyKey : undefined,
      backtest_accepted: riskProfileChanged ? riskBacktestAccepted : undefined,
    };

    try {
      await settingsApi.save(configPayload);
      setSavedRiskProfile(riskProfile);
      setAcceptedStrategyKey("");
      alert("Settings applied successfully.");
    } catch (err) {
      setSettingsError(
        err instanceof Error
          ? err.message
          : "Failed to communicate with the server.",
      );
    } finally {
      setIsSaving(false);
    }
  };

  const profileDescriptions = {
    conservative:
      "Prioritizes capital preservation with lower exposure, fewer active positions, and slower rebalance cadence.",
    moderate:
      "Balances active sleeve discovery with moderate exposure, daily cadence, and standard exit controls.",
    aggressive:
      "Allows wider paper-risk limits for comparison runs while keeping catalog strategy recipes unchanged.",
  };

  return (
    <AppShell title="Agent Settings">
      <div className="w-full bg-transparent flex flex-col gap-6 md:gap-10 animate-in fade-in duration-700 ease-[cubic-bezier(0.16,1,0.3,1)]">
        <div className="flex flex-col sm:flex-row justify-between items-start sm:items-end gap-4 pb-2 border-b border-outline-variant/10">
          <div>
            <h1 className="text-2xl sm:text-3xl font-light tracking-tight text-on-surface">
              Configuration
            </h1>
            <p className="text-xs sm:text-sm font-light text-on-surface-variant/70 mt-1">
              Adjust settings for your O(Alpha).
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

        <div className="flex flex-col gap-2">
          <span className="text-[10px] font-mono tracking-[0.2em] text-on-surface-variant/40 uppercase block mb-1">
            Risk Profiles
          </span>

          <div className="grid grid-cols-1 xl:grid-cols-3 gap-6 w-full">
            {(["conservative", "moderate", "aggressive"] as const).map(
              (profile) => {
                const isSelected = riskProfile === profile;
                const isFlipped = flippedCards[profile];

                return (
                  <div
                    key={profile}
                    onClick={() => handleProfileSelection(profile)}
                    className={`[perspective:1000px] h-44 w-full select-none ${
                      isPortfolioAgentRunning && profile !== savedRiskProfile
                        ? "cursor-not-allowed opacity-50"
                        : "cursor-pointer"
                    }`}
                  >
                    <div
                      className={`relative w-full h-full transition-transform duration-500 [transform-style:preserve-3d] ${
                        isFlipped ? "[transform:rotateY(180deg)]" : ""
                      }`}
                    >
                      <div
                        className={`absolute inset-0 [backface-visibility:hidden] flex flex-col justify-center items-center bg-surface-container-low border rounded-[24px] p-6 transition-all duration-300 ${
                          isSelected
                            ? "border-primary-fixed-dim shadow-[0_0_20px_rgba(0,240,255,0.06)] bg-surface-container"
                            : "border-outline-variant/30 hover:border-outline-variant/60"
                        }`}
                      >
                        <button
                          type="button"
                          onClick={(e) => toggleCardFlip(profile, e)}
                          className="absolute right-6 top-6 text-on-surface-variant/30 hover:text-primary-fixed-dim transition-colors h-7 w-7 rounded-full flex items-center justify-center hover:bg-white/5 border border-transparent"
                        >
                          <span className="material-symbols-outlined text-[18px]">
                            help
                          </span>
                        </button>

                        <h4
                          className={`text-xl font-light tracking-widest uppercase transition-all duration-300 ${
                            isSelected
                              ? "text-primary-fixed-dim font-medium"
                              : "text-on-surface-variant/70"
                          }`}
                        >
                          {profile}
                        </h4>
                      </div>

                      <div className="absolute inset-0 [backface-visibility:hidden] [transform:rotateY(180deg)] flex flex-col justify-center bg-surface-container border border-outline-variant/40 rounded-[24px] p-6 shadow-xl">
                        <button
                          type="button"
                          onClick={(e) => toggleCardFlip(profile, e)}
                          className="absolute right-6 top-6 text-primary-fixed-dim/70 hover:text-primary-fixed-dim h-7 w-7 rounded-full flex items-center justify-center bg-white/5 border border-outline-variant/20"
                        >
                          <span className="material-symbols-outlined text-[16px]">
                            flip_to_front
                          </span>
                        </button>

                        <p className="text-xs font-light leading-relaxed text-on-surface-variant/80 pr-6 select-text">
                          {profileDescriptions[profile]}
                        </p>
                      </div>
                    </div>
                  </div>
                );
              },
            )}
          </div>
        </div>

        {(settingsError || isPortfolioAgentRunning) && (
          <div
            className={`rounded-2xl border px-4 py-3 text-xs font-mono tracking-wide ${
              settingsError
                ? "border-error/30 bg-error/5 text-error"
                : "border-primary-fixed-dim/25 bg-primary-fixed-dim/5 text-primary-fixed-dim"
            }`}
          >
            {settingsError ||
              "A portfolio agent is running. Stop it from the dashboard before changing risk profile."}
          </div>
        )}

        {riskProfileChanged && (
          <div className="grid grid-cols-1 xl:grid-cols-[minmax(0,0.85fr)_minmax(0,1.15fr)] gap-6 md:gap-8 border-t border-outline-variant/10 pt-6 animate-in fade-in slide-in-from-top-4 duration-500 ease-[cubic-bezier(0.16,1,0.3,1)]">
            <div className="group relative bg-surface-container-low border border-outline-variant/30 rounded-[32px] p-6 sm:p-8 flex flex-col gap-4">
              <div>
                <span className="text-[10px] font-mono tracking-[0.2em] text-primary-fixed-dim uppercase">
                  Risk change validation
                </span>
                <h2 className="mt-2 text-xl font-light tracking-tight text-on-surface">
                  Backtest the new profile
                </h2>
                <p className="mt-2 text-xs font-light leading-relaxed text-on-surface-variant/70">
                  Choose a strategy from the matching catalog bucket, run the
                  five-year backtest, then accept it before saving.
                </p>
              </div>

              <div className="flex flex-col gap-3">
                {isCatalogLoading && (
                  <div className="rounded-2xl border border-outline-variant/20 bg-void-black/20 p-4 text-xs text-on-surface-variant/60">
                    Loading strategies...
                  </div>
                )}
                {strategiesForRisk.map((strategy) => {
                  const selected = selectedStrategyKey === strategy.key;
                  return (
                    <button
                      key={strategy.key}
                      type="button"
                      onClick={() => handleStrategySelect(strategy)}
                      disabled={isBacktesting || isSaving}
                      className={`text-left rounded-2xl border p-4 transition-all duration-200 ${
                        selected
                          ? "border-primary-fixed-dim bg-surface-container shadow-[0_0_20px_rgba(0,240,255,0.06)]"
                          : "border-outline-variant/20 bg-void-black/20 hover:border-outline-variant/50"
                      }`}
                    >
                      <div className="flex items-start justify-between gap-3">
                        <h3 className="text-sm font-medium text-on-surface">
                          {strategy.display_name}
                        </h3>
                        <span className="text-[9px] font-mono uppercase tracking-wider text-primary-fixed-dim">
                          {strategy.risk_profile}
                        </span>
                      </div>
                      <p className="mt-2 text-[11px] font-light leading-relaxed text-on-surface-variant/70">
                        {strategy.description}
                      </p>
                    </button>
                  );
                })}
                {!isCatalogLoading && strategiesForRisk.length === 0 && (
                  <div className="rounded-2xl border border-outline-variant/20 bg-void-black/20 p-4 text-xs text-on-surface-variant/60">
                    No catalog strategy is available for this risk profile.
                  </div>
                )}
              </div>
            </div>

            <div className="group relative bg-surface-container-low border border-outline-variant/30 rounded-[32px] p-6 sm:p-8 flex flex-col gap-4 min-w-0">
              <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
                <MetricTile
                  label="Return"
                  value={
                    metrics
                      ? `${(metrics.totalReturn * 100).toFixed(1)}%`
                      : "--"
                  }
                />
                <MetricTile
                  label="Sharpe"
                  value={metrics ? metrics.sharpe.toFixed(2) : "--"}
                />
                <MetricTile
                  label="Max DD"
                  value={
                    metrics
                      ? `${(metrics.maxDrawdown * 100).toFixed(1)}%`
                      : "--"
                  }
                />
                <MetricTile
                  label="Trades"
                  value={metrics ? String(metrics.numTrades) : "--"}
                />
              </div>

              <div className="min-h-[320px]">
                {chartData.length > 0 ? (
                  <div className="flex flex-col gap-2">
                    <EquityCurveChart data={chartData} />
                    {isBacktesting && (
                      <div className="space-y-2">
                        <div className="h-1.5 overflow-hidden rounded-full bg-void-black/40">
                          <div
                            className="h-full rounded-full bg-primary-fixed-dim transition-all duration-150"
                            style={{
                              width: `${Math.max(2, Math.round(backtestProgress * 100))}%`,
                            }}
                          />
                        </div>
                        {backtestStatus && (
                          <p className="text-[10px] font-mono tracking-[0.18em] uppercase text-on-surface-variant/45">
                            {backtestStatus}
                          </p>
                        )}
                      </div>
                    )}
                  </div>
                ) : (
                  <div className="h-80 w-full rounded-lg border border-outline-variant/20 bg-void-black/20 flex items-center justify-center text-center px-6">
                    <p className="text-xs font-mono tracking-[0.18em] uppercase text-on-surface-variant/45">
                      Run a validation backtest to unlock this risk change.
                    </p>
                  </div>
                )}
              </div>

              {backtestError && (
                <p className="rounded-xl border border-error/30 bg-error/5 px-4 py-3 text-xs text-error">
                  {backtestError}
                </p>
              )}

              {riskBacktestAccepted && (
                <p className="rounded-xl border border-primary-fixed-dim/25 bg-primary-fixed-dim/5 px-4 py-3 text-xs font-mono uppercase tracking-wide text-primary-fixed-dim">
                  Backtest accepted. This profile can be saved.
                </p>
              )}

              <div className="flex flex-col sm:flex-row justify-end gap-3 border-t border-outline-variant/10 pt-5">
                <button
                  type="button"
                  onClick={handleRunBacktest}
                  disabled={
                    isBacktesting ||
                    isSaving ||
                    !selectedStrategy ||
                    isPortfolioAgentRunning
                  }
                  className="px-6 py-3 border border-outline-variant/30 text-on-surface font-mono text-xs tracking-wider uppercase rounded-full hover:border-primary-fixed-dim/60 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                >
                  {isBacktesting ? "Running..." : "Run Backtest"}
                </button>
                <button
                  type="button"
                  onClick={handleAcceptBacktest}
                  disabled={
                    !backtestResult ||
                    isBacktesting ||
                    isSaving ||
                    acceptedStrategyKey === selectedStrategyKey
                  }
                  className="px-8 py-3 bg-primary-container text-void-black font-mono font-medium text-xs tracking-wider uppercase rounded-full disabled:opacity-50 disabled:cursor-not-allowed shadow-md"
                >
                  {acceptedStrategyKey === selectedStrategyKey
                    ? "Accepted"
                    : "Accept Backtest"}
                </button>
              </div>
            </div>
          </div>
        )}

        {isAdvanced && (
          <div className="grid grid-cols-1 xl:grid-cols-2 gap-6 md:gap-8 animate-in fade-in slide-in-from-top-4 duration-500 ease-[cubic-bezier(0.16,1,0.3,1)] border-t border-outline-variant/10 pt-6">
            <div className="group relative bg-surface-container-low border border-outline-variant/30 rounded-[32px] p-6 sm:p-8 flex flex-col gap-6">
              <div className="flex flex-col gap-2">
                <div className="flex justify-between text-[11px] font-mono tracking-wider text-on-surface-variant">
                  <span className="uppercase tracking-widest">
                    Max Gross Exposure
                  </span>
                  <span className="text-primary-container font-semibold">
                    {leverage}x
                  </span>
                </div>
                <input
                  type="range"
                  min="1"
                  max="5"
                  value={leverage}
                  onChange={(e) => setLeverage(parseInt(e.target.value))}
                  className="w-full h-[2px] appearance-none bg-outline-variant/30 rounded-full outline-none accent-primary-container cursor-pointer"
                />
              </div>

              <div className="flex flex-col gap-2">
                <div className="flex justify-between text-[11px] font-mono tracking-wider text-on-surface-variant">
                  <span className="uppercase tracking-widest">
                    Max Active Positions
                  </span>
                  <span className="text-primary-container font-semibold">
                    {maxPositions} Positions
                  </span>
                </div>
                <input
                  type="range"
                  min="1"
                  max="20"
                  value={maxPositions}
                  onChange={(e) => setMaxPositions(parseInt(e.target.value))}
                  className="w-full h-[2px] appearance-none bg-outline-variant/30 rounded-full outline-none accent-primary-container cursor-pointer"
                />
              </div>

              <div className="flex flex-col gap-3 border-t border-outline-variant/10 pt-4 mt-2">
                <span className="text-[10px] font-mono tracking-[0.2em] text-on-surface-variant/50 uppercase">
                  Rebalance Cadence
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

            <div className="group relative bg-surface-container-low border border-outline-variant/30 rounded-[32px] p-6 sm:p-8 flex flex-col gap-6">
              <div className="flex flex-col gap-2">
                <div className="flex justify-between text-[11px] font-mono tracking-wider text-on-surface-variant">
                  <span className="uppercase tracking-widest">
                    Stop-Loss Exit
                  </span>
                  <span className="text-error font-medium">-{stopLoss}%</span>
                </div>
                <input
                  type="range"
                  min="0.5"
                  max="10"
                  step="0.5"
                  value={stopLoss}
                  onChange={(e) => setStopLoss(parseFloat(e.target.value))}
                  className="w-full h-[2px] appearance-none bg-outline-variant/30 rounded-full outline-none accent-error cursor-pointer"
                />
              </div>

              <div className="flex flex-col gap-2">
                <div className="flex justify-between text-[11px] font-mono tracking-wider text-on-surface-variant">
                  <span className="uppercase tracking-widest">
                    Take-Profit Exit
                  </span>
                  <span className="text-primary-fixed-dim font-medium">
                    +{takeProfit}%
                  </span>
                </div>
                <input
                  type="range"
                  min="1"
                  max="20"
                  step="0.5"
                  value={takeProfit}
                  onChange={(e) => setTakeProfit(parseFloat(e.target.value))}
                  className="w-full h-[2px] appearance-none bg-outline-variant/30 rounded-full outline-none accent-primary-fixed-dim cursor-pointer"
                />
              </div>

              <div className="bg-white/[0.01] border border-outline-variant/10 rounded-xl p-3.5 text-[11px] font-light text-on-surface-variant/50 leading-relaxed mt-1">
                Saved controls are applied to catalog paper agents on their next
                evaluation tick. Strategy recipes stay unchanged.
              </div>
            </div>
          </div>
        )}

        <div className="pt-4 border-t border-outline-variant/20 flex justify-end">
          <button
            type="button"
            onClick={handleSave}
            disabled={
              isSaving ||
              isBacktesting ||
              (riskProfileChanged &&
                (isPortfolioAgentRunning || !riskBacktestAccepted))
            }
            className={`w-full sm:w-auto px-8 py-3 rounded-full text-xs font-mono font-medium tracking-wider uppercase text-background transition-all duration-300 active:scale-95 shadow-md ${
              isSaving ||
              isBacktesting ||
              (riskProfileChanged &&
                (isPortfolioAgentRunning || !riskBacktestAccepted))
                ? "bg-primary-container/40 cursor-not-allowed text-void-black/40"
                : "bg-primary-container text-void-black shadow-primary-container/10 hover:bg-primary-container/90"
            }`}
          >
            {isSaving
              ? "Synchronizing Matrix..."
              : "Save Terminal Configuration"}
          </button>
        </div>
      </div>
    </AppShell>
  );
}

function MetricTile({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-2xl border border-outline-variant/20 bg-void-black/20 p-3">
      <span className="block text-[9px] font-mono tracking-[0.18em] uppercase text-on-surface-variant/45">
        {label}
      </span>
      <span className="mt-1 block text-lg font-light text-on-surface">
        {value}
      </span>
    </div>
  );
}

function normalizeRiskProfile(value: string): RiskProfile {
  if (value === "conservative" || value === "aggressive") {
    return value;
  }
  return "moderate";
}

function firstStrategyForRisk(
  strategies: StrategySpec[],
  profile: RiskProfile,
) {
  return strategies.find(
    (strategy) => strategy.risk_profile === riskBuckets[profile],
  );
}

function recommendedStrategyKeyForRisk(
  catalog: StrategyCatalogResponse,
  profile: RiskProfile,
) {
  const recommendedKey = catalog.recommended[profile];
  const recommended = catalog.strategies.find(
    (strategy) => strategy.key === recommendedKey,
  );
  if (recommended?.risk_profile === riskBuckets[profile]) {
    return recommended.key;
  }
  return firstStrategyForRisk(catalog.strategies, profile)?.key || "";
}

function equityReturn(result: BacktestResult) {
  if (result.equity_curve.length < 2) return 0;
  const first = result.equity_curve[0].equity;
  const last = result.equity_curve[result.equity_curve.length - 1].equity;
  return first <= 0 ? 0 : (last - first) / first;
}

async function revealEquityCurve(
  curve: EquityPoint[],
  alreadyVisible: number,
  setters: {
    setCurve: (points: EquityPoint[]) => void;
    setProgress: (progress: number) => void;
  },
) {
  const total = curve.length;
  if (total === 0) return;
  const start = Math.max(1, Math.min(alreadyVisible, total));
  const batchSize = Math.max(4, Math.ceil(total / 40));

  for (let end = start; end < total; end = Math.min(total, end + batchSize)) {
    const nextEnd = Math.min(total, end + batchSize);
    setters.setCurve(curve.slice(0, nextEnd));
    setters.setProgress(nextEnd / total);
    await new Promise((resolve) => window.setTimeout(resolve, 24));
  }
}
