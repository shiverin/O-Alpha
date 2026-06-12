"use client";

import { useEffect, useMemo, useState } from "react";
import { EquityCurveChart } from "@/components/EquityCurveChart";
import {
  runBacktest,
  settingsApi,
  strategyCatalogApi,
  userApi,
  type BacktestResult,
  type ServerAgentSettings,
  type StrategyCatalogResponse,
  type StrategySpec,
} from "@/lib/api";

interface OnboardingOverlayProps {
  userID: number;
  onComplete: (riskProfile: string) => void;
}

type RiskProfile = "conservative" | "moderate" | "aggressive";
type RiskBucket = "low" | "medium" | "high";

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
      description: "VOO core with a 15% learned-ranker active sleeve.",
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
      description: "Higher-active-weight composite momentum sleeve.",
    },
  ],
};

export default function OnboardingOverlay({
  userID,
  onComplete,
}: OnboardingOverlayProps) {
  const [isVisible, setIsVisible] = useState(false);
  const [step, setStep] = useState<1 | 2 | 3>(1);
  const [riskProfile, setRiskProfile] = useState<RiskProfile>("moderate");
  const [catalog, setCatalog] =
    useState<StrategyCatalogResponse>(fallbackCatalog);
  const [selectedStrategyKey, setSelectedStrategyKey] = useState(
    fallbackCatalog.recommended.moderate,
  );
  const [backtestResult, setBacktestResult] = useState<BacktestResult | null>(
    null,
  );
  const [backtestError, setBacktestError] = useState<string | null>(null);
  const [isCatalogLoading, setIsCatalogLoading] = useState(false);
  const [isBacktesting, setIsBacktesting] = useState(false);
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

  useEffect(() => {
    const loadCatalog = async () => {
      setIsCatalogLoading(true);
      try {
        const response = await strategyCatalogApi.list();
        setCatalog(response);
        setSelectedStrategyKey(
          recommendedStrategyKeyForRisk(response, riskProfile) ||
            fallbackCatalog.recommended.moderate,
        );
      } catch {
        setCatalog(fallbackCatalog);
      } finally {
        setIsCatalogLoading(false);
      }
    };

    loadCatalog();
  }, [riskProfile]);

  const profileDescriptions: Record<RiskProfile, string> = {
    conservative:
      "Prioritizes capital preservation with lower exposure, fewer active positions, and slower rebalance cadence.",
    moderate:
      "Balances active sleeve discovery with moderate exposure, daily cadence, and standard exit controls.",
    aggressive:
      "Allows wider paper-risk limits for comparison runs while keeping catalog strategy recipes unchanged.",
  };

  const strategiesForRisk = useMemo(() => {
    const bucket = riskBuckets[riskProfile];
    return catalog.strategies.filter(
      (strategy) => strategy.risk_profile === bucket,
    );
  }, [catalog.strategies, riskProfile]);

  const selectedStrategy = useMemo(() => {
    return (
      catalog.strategies.find(
        (strategy) => strategy.key === selectedStrategyKey,
      ) ||
      strategiesForRisk[0] ||
      catalog.strategies[0]
    );
  }, [catalog.strategies, selectedStrategyKey, strategiesForRisk]);

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
      finalEquity:
        backtestResult.final_equity ??
        backtestResult.equity_curve[backtestResult.equity_curve.length - 1]
          ?.equity ??
        0,
    };
  }, [backtestResult]);

  const toggleCardFlip = (profile: string, e: React.MouseEvent) => {
    e.stopPropagation();
    setFlippedCards((prev) => ({ ...prev, [profile]: !prev[profile] }));
  };

  const handleProfileSelect = (profile: RiskProfile) => {
    setRiskProfile(profile);
    const nextKey =
      recommendedStrategyKeyForRisk(catalog, profile) || selectedStrategyKey;
    setSelectedStrategyKey(nextKey);
    setBacktestResult(null);
    setBacktestError(null);
    setStep(3);
  };

  const handleStrategySelect = (strategy: StrategySpec) => {
    setSelectedStrategyKey(strategy.key);
    setBacktestResult(null);
    setBacktestError(null);
  };

  const handleRunBacktest = async () => {
    if (!selectedStrategy) return;
    setIsBacktesting(true);
    setBacktestError(null);
    setBacktestResult(null);

    const end = new Date();
    const start = new Date(end);
    start.setFullYear(end.getFullYear() - 5);
    const settings = settingsForRisk(riskProfile);

    try {
      const result = await runBacktest({
        symbol:
          selectedStrategy.benchmark_symbol || catalog.default_universe[0],
        symbols: catalog.default_universe,
        start: start.toISOString(),
        end: end.toISOString(),
        strategy_type: "PORTFOLIO_CATALOG",
        timeframe: "1Day",
        initial_cash: 100000,
        parameters: {
          strategy_key: selectedStrategy.key,
          max_gross_exposure: settings.leverage,
          max_net_exposure: settings.leverage,
          max_symbol_weight: 1,
        },
      });
      setBacktestResult(result);
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

  const handleAcceptBacktest = async () => {
    if (!backtestResult) return;
    setIsSaving(true);
    const configPayload = settingsForRisk(riskProfile);

    if (userID === 999) {
      await new Promise((resolve) => setTimeout(resolve, 400));
      localStorage.setItem("oa_demo_risk_posture", riskProfile);
      localStorage.setItem("oa_demo_onboarding_strategy", selectedStrategyKey);
      setIsSaving(false);
      onComplete(riskProfile);
      return;
    }

    try {
      await settingsApi.save(configPayload);
      await userApi.completeOnboarding({
        risk_profile: riskProfile,
        strategy_key: selectedStrategyKey,
        backtest_accepted: true,
      });
      setIsSaving(false);
      onComplete(riskProfile);
    } catch {
      setIsSaving(false);
      alert("Failed to complete onboarding after backtest acceptance.");
    }
  };

  if (!isVisible) return null;

  return (
    <div className="fixed inset-0 z-[9999] bg-background/60 backdrop-blur-2xl flex items-center justify-center p-4 sm:p-6 transition-all duration-1000 ease-out">
      <div
        className="absolute inset-0 opacity-[0.03] pointer-events-none"
        style={{
          backgroundImage:
            "radial-gradient(circle at 1px 1px, white 1px, transparent 0)",
          backgroundSize: "32px 32px",
        }}
      />
      <div className="w-full max-w-5xl max-h-[92vh] overflow-y-auto bg-surface-container-low border border-outline-variant/30 rounded-[32px] p-6 sm:p-10 shadow-[0_30px_70px_rgba(0,0,0,0.6)] relative transition-all duration-500 scale-100">
        <div className="absolute top-0 inset-x-0 h-[1px] bg-gradient-to-r from-transparent via-primary-fixed-dim/40 to-transparent" />

        {step === 1 && (
          <div className="flex flex-col items-center text-center py-10 max-w-2xl mx-auto">
            <span className="material-symbols-outlined text-primary-fixed-dim text-3xl mb-4">
              token
            </span>
            <h1 className="text-3xl sm:text-4xl font-light tracking-tight text-on-surface mb-4">
              Welcome to{" "}
              <span className="text-primary-fixed-dim font-normal">
                O(Alpha)
              </span>
            </h1>
            <p className="text-sm sm:text-base font-light text-on-surface-variant/70 leading-relaxed mb-10">
              Choose a risk profile, backtest a matching catalog strategy, then
              accept the result to activate your paper agent workspace.
            </p>
            <button
              onClick={() => setStep(2)}
              className="px-8 py-3.5 bg-primary-container text-void-black font-mono font-medium text-xs tracking-wider uppercase rounded-full shadow-lg hover:bg-primary-fixed transition-all duration-300"
            >
              Begin
            </button>
          </div>
        )}

        {step === 2 && (
          <div className="flex flex-col">
            <div className="mb-8 text-center sm:text-left">
              <span className="text-[10px] font-mono tracking-[0.25em] text-primary-fixed-dim uppercase block mb-1">
                Step 1
              </span>
              <h2 className="text-2xl font-light tracking-tight text-on-surface">
                Select Your Risk Profile
              </h2>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-6 w-full mb-10">
              {(["conservative", "moderate", "aggressive"] as const).map(
                (profile) => {
                  const isSelected = riskProfile === profile;
                  const isFlipped = flippedCards[profile];

                  return (
                    <div
                      key={profile}
                      onClick={() => handleProfileSelect(profile)}
                      className="[perspective:1000px] h-48 w-full cursor-pointer select-none"
                    >
                      <div
                        className={`relative w-full h-full transition-transform duration-500 [transform-style:preserve-3d] ${isFlipped ? "[transform:rotateY(180deg)]" : ""}`}
                      >
                        <div
                          className={`absolute inset-0 [backface-visibility:hidden] flex flex-col justify-center items-center bg-void-black/20 border rounded-2xl p-6 transition-all duration-300 ${isSelected ? "border-primary-fixed-dim shadow-[0_0_20px_rgba(0,240,255,0.08)] bg-surface-container" : "border-outline-variant/20 hover:border-outline-variant/50"}`}
                        >
                          <button
                            type="button"
                            onClick={(e) => toggleCardFlip(profile, e)}
                            className="absolute right-4 top-4 text-on-surface-variant/30 hover:text-primary-fixed-dim transition-colors h-7 w-7 rounded-full flex items-center justify-center hover:bg-white/5"
                          >
                            <span className="material-symbols-outlined text-[18px]">
                              help
                            </span>
                          </button>
                          <h4
                            className={`text-base font-light tracking-widest uppercase ${isSelected ? "text-primary-fixed-dim font-medium" : "text-on-surface-variant/70"}`}
                          >
                            {profile}
                          </h4>
                        </div>
                        <div className="absolute inset-0 [backface-visibility:hidden] [transform:rotateY(180deg)] flex flex-col justify-center bg-surface-container-high border border-outline-variant/40 rounded-2xl p-6 shadow-xl">
                          <button
                            type="button"
                            onClick={(e) => toggleCardFlip(profile, e)}
                            className="absolute right-4 top-4 text-primary-fixed-dim/70 hover:text-primary-fixed-dim h-7 w-7 rounded-full flex items-center justify-center bg-white/5"
                          >
                            <span className="material-symbols-outlined text-[16px]">
                              flip_to_front
                            </span>
                          </button>
                          <p className="text-xs font-light leading-relaxed text-on-surface-variant/90 pr-4 select-text">
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
        )}

        {step === 3 && (
          <div className="flex flex-col gap-6">
            <div className="flex flex-col sm:flex-row sm:items-end justify-between gap-4">
              <div>
                <span className="text-[10px] font-mono tracking-[0.25em] text-primary-fixed-dim uppercase block mb-1">
                  Step 2
                </span>
                <h2 className="text-2xl font-light tracking-tight text-on-surface">
                  Backtest Your Chosen Strategy
                </h2>
                <p className="text-xs sm:text-sm font-light text-on-surface-variant/70 mt-2 max-w-2xl">
                  Pick a catalog strategy from your {riskProfile} profile and
                  run a 5-year paper backtest before onboarding is finalized.
                </p>
              </div>
              <button
                onClick={() => setStep(2)}
                disabled={isBacktesting || isSaving}
                className="px-5 py-2.5 border border-outline-variant/30 text-on-surface-variant font-mono text-xs tracking-wider uppercase rounded-full hover:text-on-surface transition-colors"
              >
                Change Profile
              </button>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-[minmax(0,0.8fr)_minmax(0,1.2fr)] gap-6">
              <div className="flex flex-col gap-3">
                <span className="text-[10px] font-mono tracking-[0.2em] text-on-surface-variant/50 uppercase">
                  Catalog Strategies
                </span>
                {isCatalogLoading && (
                  <div className="rounded-2xl border border-outline-variant/20 bg-void-black/20 p-4 text-xs text-on-surface-variant/60">
                    Loading catalog...
                  </div>
                )}
                {(strategiesForRisk.length > 0
                  ? strategiesForRisk
                  : catalog.strategies
                ).map((strategy) => {
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
              </div>

              <div className="flex flex-col gap-4 min-w-0">
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

                <div className="min-h-[340px]">
                  {backtestResult ? (
                    <EquityCurveChart data={backtestResult.equity_curve} />
                  ) : (
                    <div className="h-80 w-full rounded-lg border border-outline-variant/20 bg-void-black/20 flex items-center justify-center text-center px-6">
                      <p className="text-xs font-mono tracking-[0.18em] uppercase text-on-surface-variant/45">
                        {isBacktesting
                          ? "Running 5-year catalog backtest..."
                          : "Run a backtest to unlock onboarding acceptance."}
                      </p>
                    </div>
                  )}
                </div>

                {backtestError && (
                  <p className="rounded-xl border border-error/30 bg-error/5 px-4 py-3 text-xs text-error">
                    {backtestError}
                  </p>
                )}

                <div className="flex flex-col sm:flex-row justify-end gap-3 border-t border-outline-variant/10 pt-5">
                  <button
                    onClick={handleRunBacktest}
                    disabled={isBacktesting || isSaving || !selectedStrategy}
                    className="px-6 py-3 border border-outline-variant/30 text-on-surface font-mono text-xs tracking-wider uppercase rounded-full hover:border-primary-fixed-dim/60 disabled:opacity-50 transition-colors"
                  >
                    {isBacktesting ? "Running..." : "Start Backtest"}
                  </button>
                  <button
                    onClick={handleAcceptBacktest}
                    disabled={!backtestResult || isBacktesting || isSaving}
                    className="px-8 py-3 bg-primary-container text-void-black font-mono font-medium text-xs tracking-wider uppercase rounded-full disabled:opacity-50 shadow-md"
                  >
                    {isSaving ? "Finalizing..." : "Accept Backtest"}
                  </button>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
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
  return (
    firstStrategyForRisk(catalog.strategies, profile)?.key ||
    catalog.strategies[0]?.key
  );
}

function settingsForRisk(profile: RiskProfile): ServerAgentSettings {
  const base = {
    risk_profile: profile,
    leverage: 2,
    max_positions: 6,
    stop_loss_pct: 2.5,
    take_profit_pct: 5.0,
    rebalance_freq: "daily",
  };
  if (profile === "conservative") {
    return {
      ...base,
      leverage: 1,
      max_positions: 3,
      stop_loss_pct: 1.5,
      take_profit_pct: 3.0,
      rebalance_freq: "weekly",
    };
  }
  if (profile === "aggressive") {
    return {
      ...base,
      leverage: 4,
      max_positions: 12,
      stop_loss_pct: 4.0,
      take_profit_pct: 12.0,
      rebalance_freq: "hourly",
    };
  }
  return base;
}

function equityReturn(result: BacktestResult) {
  if (result.equity_curve.length < 2) return 0;
  const first = result.equity_curve[0].equity;
  const last = result.equity_curve[result.equity_curve.length - 1].equity;
  return first <= 0 ? 0 : (last - first) / first;
}
