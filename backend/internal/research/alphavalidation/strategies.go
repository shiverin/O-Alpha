package alphavalidation

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	agentportfolio "github.com/oalpha/internal/agent/portfolio"
	"github.com/oalpha/internal/alpha/cointegration"
	"github.com/oalpha/internal/alpha/momentum"
	"github.com/oalpha/internal/alpha/ranker"
	"github.com/oalpha/internal/backtest"
	"github.com/oalpha/pkg/models"
)

func BenchmarkFactories(symbols []string) []StrategyFactory {
	normalized := normalizeSymbols(symbols)
	if len(normalized) == 0 {
		return nil
	}
	factories := []StrategyFactory{
		{
			Name:      "buy_hold",
			Family:    "benchmark",
			Benchmark: "",
			New: func() backtest.PortfolioStrategy {
				return &buyHoldStrategy{symbol: normalized[0]}
			},
		},
		{
			Name:      "equal_weight",
			Family:    "benchmark",
			Benchmark: "",
			New: func() backtest.PortfolioStrategy {
				return &equalWeightStrategy{symbols: normalized}
			},
		},
		{
			Name:      "flat_cash",
			Family:    "benchmark",
			Benchmark: "",
			New: func() backtest.PortfolioStrategy {
				return &flatStrategy{symbols: normalized}
			},
		},
	}
	return factories
}

func CandidateFactories(symbols []string, strategyNames []string) []StrategyFactory {
	normalized := normalizeSymbols(symbols)
	selected := strategySet(strategyNames)
	factories := make([]StrategyFactory, 0)
	if len(normalized) >= 1 {
		symbol := normalized[0]
		if selected["ma"] || selected["all"] {
			factories = append(factories, maFactory(symbol, 20, 50, "ma_crossover_20_50"))
		}
		if selected["kalman"] || selected["all"] {
			factories = append(factories, kalmanFactory(symbol, 0.01, 0.5, 2.0, "kalman_z2"))
		}
	}
	if len(normalized) >= 2 {
		if selected["xsec"] || selected["xsec_momentum"] || selected["all"] {
			factories = append(factories, xsecFactory(normalized, 0.15, "xsec_momentum_top15"))
		}
		if selected["composite"] || selected["composite_momentum"] || selected["composite_momentum_sleeve"] || selected["all"] {
			factories = append(factories, compositeMomentumFactory(normalized, compositeMomentumCheckpointConfig(normalized[0]), "composite_momentum_checkpoint"))
		}
		if selected["benchmark_rotation"] || selected["composite_defensive"] || selected["defensive_composite"] || selected["all"] {
			factories = append(factories, compositeMomentumFamilyFactory(normalized, defensiveCompositeCheckpointConfig(normalized[0]), "benchmark_rotation_defensive", "composite_momentum_defensive"))
		}
		if selected["benchmark_tsmom"] || selected["tsmom_sleeve"] || selected["benchmark_funded_tsmom"] || selected["all"] {
			factories = append(factories, compositeMomentumFamilyFactory(normalized, benchmarkTSMOMCheckpointConfig(normalized[0]), "benchmark_tsmom_checkpoint", "benchmark_tsmom"))
		}
		if selected["benchmark_tsmom_blend"] || selected["tsmom_blend"] || selected["all"] {
			factories = append(factories, compositeMomentumFamilyFactory(normalized, benchmarkTSMOMBlendCheckpointConfig(normalized[0]), "benchmark_tsmom_blend", "benchmark_tsmom_blend"))
		}
		if selected["benchmark_lowvol"] || selected["lowvol_sleeve"] || selected["low_volatility"] || selected["all"] {
			factories = append(factories, compositeMomentumFamilyFactory(normalized, benchmarkLowVolCheckpointConfig(normalized[0]), "benchmark_lowvol_checkpoint", "benchmark_lowvol"))
		}
		if selected["benchmark_reversal"] || selected["reversal_sleeve"] || selected["mean_reversion"] || selected["all"] {
			factories = append(factories, compositeMomentumFamilyFactory(normalized, benchmarkReversalCheckpointConfig(normalized[0]), "benchmark_reversal_checkpoint", "benchmark_reversal"))
		}
		if selected["benchmark_ranked_sleeve"] || selected["ranked_sleeve"] || selected["risk_budgeted_sleeve"] || selected["all"] {
			factories = append(factories, compositeMomentumFamilyFactory(normalized, benchmarkRankedSleeveCheckpointConfig(normalized[0]), "benchmark_ranked_sleeve_checkpoint", "benchmark_ranked_sleeve"))
		}
		if selected["benchmark_ranker_proxy"] || selected["ranker_proxy"] || selected["all"] {
			factories = append(factories, compositeMomentumFamilyFactory(normalized, benchmarkRankerProxyCheckpointConfig(normalized[0]), "benchmark_ranker_proxy_checkpoint", "benchmark_ranker_proxy"))
		}
		if selected["benchmark_ranker_proxy_h63"] || selected["ranker_proxy_h63"] || selected["ranker_h63"] || selected["all"] {
			factories = append(factories, compositeMomentumFamilyFactory(normalized, benchmarkRankerProxyH63CheckpointConfig(normalized[0]), "benchmark_ranker_proxy_h63_checkpoint", "benchmark_ranker_proxy_h63"))
		}
		if selected["benchmark_ranker_proxy_h63_riskcap"] || selected["ranker_proxy_h63_riskcap"] || selected["ranker_h63_riskcap"] || selected["all"] {
			factories = append(factories, compositeMomentumFamilyFactory(normalized, benchmarkRankerProxyH63RiskCapCheckpointConfig(normalized[0]), "benchmark_ranker_proxy_h63_riskcap_checkpoint", "benchmark_ranker_proxy_h63_riskcap"))
		}
		if selected["benchmark_ranker_proxy_h63_trendguard"] || selected["ranker_proxy_h63_trendguard"] || selected["ranker_h63_trendguard"] || selected["all"] {
			factories = append(factories, compositeMomentumFamilyFactory(normalized, benchmarkRankerProxyH63TrendGuardCheckpointConfig(normalized[0]), "benchmark_ranker_proxy_h63_trendguard_checkpoint", "benchmark_ranker_proxy_h63_trendguard"))
		}
		if selected["benchmark_ranker_proxy_h63_liquidity"] || selected["ranker_proxy_h63_liquidity"] || selected["ranker_h63_liquidity"] || selected["all"] {
			factories = append(factories, compositeMomentumFamilyFactory(normalized, benchmarkRankerProxyH63LiquidityCheckpointConfig(normalized[0]), "benchmark_ranker_proxy_h63_liquidity_checkpoint", "benchmark_ranker_proxy_h63_liquidity"))
		}
		if selected["benchmark_ranker_proxy_h84"] || selected["ranker_proxy_h84"] || selected["ranker_h84"] || selected["all"] {
			factories = append(factories, compositeMomentumFamilyFactory(normalized, benchmarkRankerProxyH84CheckpointConfig(normalized[0]), "benchmark_ranker_proxy_h84_checkpoint", "benchmark_ranker_proxy_h84"))
		}
		if selected["benchmark_ranker_proxy_blend"] || selected["ranker_proxy_blend"] || selected["ranker_blend"] || selected["all"] {
			factories = append(factories, compositeMomentumFamilyFactory(normalized, benchmarkRankerProxyBlendCheckpointConfig(normalized[0]), "benchmark_ranker_proxy_blend_checkpoint", "benchmark_ranker_proxy_blend"))
		}
		if selected["benchmark_lgbm_ranker"] || selected["daily_lgbm_ranker"] || selected["lgbm_ranker"] {
			factories = append(factories, dailyRankerFamilyFactory(normalized, "stocks_h63_s15_top3_reb63_z10", dailyRankerSleeveConfig(normalized[0], 0.15, 3, 0.05, 63, 1.0, 0, 0, 1, 0, 1), "benchmark_lgbm_ranker_h63_s15"))
		}
		if selected["benchmark_lgbm_ranker_h63"] || selected["daily_lgbm_ranker_h63"] || selected["lgbm_ranker_h63"] {
			factories = append(factories, dailyRankerFamilyFactoryWithFamily(normalized, "stocks_h63_s15_top3_reb63_z10", dailyRankerSleeveConfig(normalized[0], 0.15, 3, 0.05, 63, 1.0, 0, 0, 1, 0, 1), "benchmark_lgbm_ranker_h63_s15_checkpoint", "benchmark_lgbm_ranker_h63"))
		}
		if selected["benchmark_lgbm_ranker_h63_exmegacap"] || selected["lgbm_ranker_h63_exmegacap"] || selected["ranker_h63_exmegacap"] {
			factories = append(factories, dailyRankerFamilyFactoryWithFamily(normalized, "stocks_h63_s15_top3_reb63_z10", dailyRankerSleeveConfigExcluding(normalized[0], 0.15, 3, 0.05, 63, 1.0, 0, 0, 1, 0, 1, megaCapExclusions()), "benchmark_lgbm_ranker_h63_s15_exmegacap", "benchmark_lgbm_ranker_h63_exmegacap"))
		}
		if selected["benchmark_lgbm_ranker_h63_equal"] || selected["lgbm_ranker_h63_equal"] || selected["ranker_h63_vs_equal"] {
			factories = append(factories, dailyRankerFamilyFactoryWithBenchmark(normalized, "stocks_h63_s15_top3_reb63_z10", dailyRankerSleeveConfig(normalized[0], 0.15, 3, 0.05, 63, 1.0, 0, 0, 1, 0, 1), "benchmark_lgbm_ranker_h63_s15_equal_benchmark", "benchmark_lgbm_ranker_h63_equal", "equal_weight"))
		}
		if selected["sector_ranked_sleeve"] || selected["sector_sleeve"] || selected["all"] {
			factories = append(factories, compositeMomentumFamilyFactory(normalized, sectorRankedSleeveCheckpointConfig(normalized[0]), "sector_ranked_sleeve_checkpoint", "sector_ranked_sleeve"))
		}
		if selected["pair"] || selected["kalman_cointegration"] || selected["all"] {
			factories = append(factories, pairFactory(normalized[0], normalized[1], 2.0, "kalman_cointegration_z2"))
		}
	}
	factories = append(factories, agentCatalogFactories(normalized, selected)...)
	return factories
}

func VariantFactories(factory StrategyFactory, symbols []string) []StrategyFactory {
	normalized := normalizeSymbols(symbols)
	switch factory.Family {
	case "agent_catalog_low", "agent_catalog_medium", "agent_catalog_high":
		return agentCatalogVariantFactories(factory.Family, normalized)
	case "ma_crossover":
		if len(normalized) == 0 {
			return nil
		}
		symbol := normalized[0]
		return []StrategyFactory{
			maFactory(symbol, 10, 30, "ma_crossover_10_30"),
			maFactory(symbol, 20, 50, "ma_crossover_20_50"),
			maFactory(symbol, 50, 100, "ma_crossover_50_100"),
		}
	case "kalman":
		if len(normalized) == 0 {
			return nil
		}
		symbol := normalized[0]
		return []StrategyFactory{
			kalmanFactory(symbol, 0.01, 0.5, 1.5, "kalman_z1_5"),
			kalmanFactory(symbol, 0.01, 0.5, 2.0, "kalman_z2"),
			kalmanFactory(symbol, 0.01, 0.5, 2.5, "kalman_z2_5"),
		}
	case "xsec_momentum":
		if len(normalized) < 2 {
			return nil
		}
		return []StrategyFactory{
			xsecFactory(normalized, 0.10, "xsec_momentum_top10"),
			xsecFactory(normalized, 0.15, "xsec_momentum_top15"),
			xsecFactory(normalized, 0.25, "xsec_momentum_top25"),
		}
	case "composite_momentum":
		if len(normalized) < 2 {
			return nil
		}
		benchmark := normalized[0]
		return []StrategyFactory{
			compositeMomentumFactory(normalized, compositeMomentumCheckpointConfig(benchmark), "composite_momentum_checkpoint"),
			compositeMomentumFactory(normalized, compositeMomentumVariantConfig(benchmark, 0.20, 0.05, 0.25, 0.03), "composite_momentum_sleeve20_broad5"),
			compositeMomentumFactory(normalized, compositeMomentumVariantConfig(benchmark, 0.24, 0.08, 0.22, 0.05), "composite_momentum_strict_etf"),
			compositeMomentumFactory(normalized, compositeMomentumVariantConfig(benchmark, 0.18, 0.05, 0.25, 0.08), "composite_momentum_broader_core"),
		}
	case "composite_momentum_defensive":
		if len(normalized) < 2 {
			return nil
		}
		benchmark := normalized[0]
		return []StrategyFactory{
			compositeMomentumFamilyFactory(normalized, defensiveCompositeCheckpointConfig(benchmark), "benchmark_rotation_defensive", "composite_momentum_defensive"),
			compositeMomentumFamilyFactory(normalized, defensiveCompositeVariantConfig(benchmark, 126, 0.25, 0.12, 0.03), "benchmark_rotation_trend126", "composite_momentum_defensive"),
			compositeMomentumFamilyFactory(normalized, defensiveCompositeVariantConfig(benchmark, 200, 0.50, 0.15, 0.02), "benchmark_rotation_half_defensive", "composite_momentum_defensive"),
			compositeMomentumFamilyFactory(normalized, defensiveCompositeVariantConfig(benchmark, 126, 0.00, 0.18, 0.05), "benchmark_rotation_cash_defensive", "composite_momentum_defensive"),
		}
	case "benchmark_tsmom":
		if len(normalized) < 2 {
			return nil
		}
		benchmark := normalized[0]
		return []StrategyFactory{
			compositeMomentumFamilyFactory(normalized, benchmarkTSMOMCheckpointConfig(benchmark), "benchmark_tsmom_checkpoint", "benchmark_tsmom"),
			compositeMomentumFamilyFactory(normalized, benchmarkTSMOMVariantConfig(benchmark, 42, 126, 252, 0.12, 0.08), "benchmark_tsmom_reb42", "benchmark_tsmom"),
			compositeMomentumFamilyFactory(normalized, benchmarkTSMOMVariantConfig(benchmark, 63, 126, 189, 0.18, 0.05), "benchmark_tsmom_medium", "benchmark_tsmom"),
			compositeMomentumFamilyFactory(normalized, benchmarkTSMOMVariantConfig(benchmark, 63, 189, 252, 0.10, 0.10), "benchmark_tsmom_slow", "benchmark_tsmom"),
		}
	case "benchmark_tsmom_blend":
		if len(normalized) < 2 {
			return nil
		}
		benchmark := normalized[0]
		return []StrategyFactory{
			compositeMomentumFamilyFactory(normalized, benchmarkTSMOMBlendCheckpointConfig(benchmark), "benchmark_tsmom_blend", "benchmark_tsmom_blend"),
			compositeMomentumFamilyFactory(normalized, benchmarkTSMOMBlendConfig(benchmark, 42, 0.06, 0.04), "benchmark_tsmom_blend_reb42", "benchmark_tsmom_blend"),
			compositeMomentumFamilyFactory(normalized, benchmarkTSMOMBlendConfig(benchmark, 63, 0.05, 0.05), "benchmark_tsmom_blend_slow_broad", "benchmark_tsmom_blend"),
			compositeMomentumFamilyFactory(normalized, benchmarkTSMOMBlendConfig(benchmark, 63, 0.09, 0.03), "benchmark_tsmom_blend_etf_tilt", "benchmark_tsmom_blend"),
		}
	case "benchmark_lowvol":
		if len(normalized) < 2 {
			return nil
		}
		benchmark := normalized[0]
		return []StrategyFactory{
			compositeMomentumFamilyFactory(normalized, benchmarkLowVolCheckpointConfig(benchmark), "benchmark_lowvol_checkpoint", "benchmark_lowvol"),
			compositeMomentumFamilyFactory(normalized, benchmarkLowVolConfig(benchmark, 42, 0.15, 5, "low_vol"), "benchmark_lowvol_reb42", "benchmark_lowvol"),
			compositeMomentumFamilyFactory(normalized, benchmarkLowVolConfig(benchmark, 21, 0.25, 8, "low_vol"), "benchmark_lowvol_wider", "benchmark_lowvol"),
			compositeMomentumFamilyFactory(normalized, benchmarkLowVolConfig(benchmark, 42, 0.20, 5, "vol_adjusted_momentum"), "benchmark_lowvol_voladj", "benchmark_lowvol"),
		}
	case "benchmark_reversal":
		if len(normalized) < 2 {
			return nil
		}
		benchmark := normalized[0]
		return []StrategyFactory{
			compositeMomentumFamilyFactory(normalized, benchmarkReversalCheckpointConfig(benchmark), "benchmark_reversal_checkpoint", "benchmark_reversal"),
			compositeMomentumFamilyFactory(normalized, benchmarkReversalConfig(benchmark, 5, 5, 0.12, 8), "benchmark_reversal_fast", "benchmark_reversal"),
			compositeMomentumFamilyFactory(normalized, benchmarkReversalConfig(benchmark, 10, 10, 0.18, 10), "benchmark_reversal_medium", "benchmark_reversal"),
			compositeMomentumFamilyFactory(normalized, benchmarkReversalConfig(benchmark, 21, 21, 0.15, 10), "benchmark_reversal_slow", "benchmark_reversal"),
		}
	case "benchmark_ranked_sleeve":
		if len(normalized) < 2 {
			return nil
		}
		benchmark := normalized[0]
		return []StrategyFactory{
			compositeMomentumFamilyFactory(normalized, benchmarkRankedSleeveCheckpointConfig(benchmark), "benchmark_ranked_sleeve_checkpoint", "benchmark_ranked_sleeve"),
			compositeMomentumFamilyFactory(normalized, benchmarkRankedSleeveConfig(benchmark, 42, 126, 0.20, 3, 0.08, 0.02), "benchmark_ranked_sleeve_conservative", "benchmark_ranked_sleeve"),
			compositeMomentumFamilyFactory(normalized, benchmarkRankedSleeveConfig(benchmark, 21, 189, 0.30, 5, 0.08, 0.04), "benchmark_ranked_sleeve_medium", "benchmark_ranked_sleeve"),
			compositeMomentumFamilyFactory(normalized, benchmarkRankedSleeveConfig(benchmark, 63, 252, 0.25, 5, 0.10, 0.02), "benchmark_ranked_sleeve_slow", "benchmark_ranked_sleeve"),
		}
	case "benchmark_ranker_proxy":
		if len(normalized) < 2 {
			return nil
		}
		benchmark := normalized[0]
		return []StrategyFactory{
			compositeMomentumFamilyFactory(normalized, benchmarkRankerProxyCheckpointConfig(benchmark), "benchmark_ranker_proxy_checkpoint", "benchmark_ranker_proxy"),
			compositeMomentumFamilyFactory(normalized, benchmarkRankerProxyConfig(benchmark, 42, 189, 0.15, 3, 0.05, 0.02), "benchmark_ranker_proxy_medium", "benchmark_ranker_proxy"),
			compositeMomentumFamilyFactory(normalized, benchmarkRankerProxyConfig(benchmark, 42, 126, 0.10, 3, 0.04, 0.02), "benchmark_ranker_proxy_fast", "benchmark_ranker_proxy"),
			compositeMomentumFamilyFactory(normalized, benchmarkRankerProxyConfig(benchmark, 63, 252, 0.10, 3, 0.04, 0.00), "benchmark_ranker_proxy_slow", "benchmark_ranker_proxy"),
		}
	case "benchmark_ranker_proxy_h63":
		if len(normalized) < 2 {
			return nil
		}
		benchmark := normalized[0]
		return []StrategyFactory{
			compositeMomentumFamilyFactory(normalized, benchmarkRankerProxyH63CheckpointConfig(benchmark), "benchmark_ranker_proxy_h63_checkpoint", "benchmark_ranker_proxy_h63"),
			compositeMomentumFamilyFactory(normalized, benchmarkRankerProxyConfig(benchmark, 63, 63, 0.10, 3, 0.04, 0.00), "benchmark_ranker_proxy_h63_sleeve10", "benchmark_ranker_proxy_h63"),
			compositeMomentumFamilyFactory(normalized, benchmarkRankerProxyConfig(benchmark, 63, 63, 0.15, 3, 0.05, 0.03), "benchmark_ranker_proxy_h63_strict", "benchmark_ranker_proxy_h63"),
			compositeMomentumFamilyFactory(normalized, benchmarkRankerProxyConfig(benchmark, 63, 84, 0.15, 3, 0.05, 0.00), "benchmark_ranker_proxy_h84", "benchmark_ranker_proxy_h63"),
			compositeMomentumFamilyFactory(normalized, benchmarkRankerProxyConfig(benchmark, 63, 126, 0.15, 3, 0.05, 0.00), "benchmark_ranker_proxy_h126", "benchmark_ranker_proxy_h63"),
		}
	case "benchmark_ranker_proxy_h63_riskcap":
		if len(normalized) < 2 {
			return nil
		}
		benchmark := normalized[0]
		return []StrategyFactory{
			compositeMomentumFamilyFactory(normalized, benchmarkRankerProxyH63RiskCapCheckpointConfig(benchmark), "benchmark_ranker_proxy_h63_riskcap_checkpoint", "benchmark_ranker_proxy_h63_riskcap"),
			compositeMomentumFamilyFactory(normalized, benchmarkRankerProxyConfigWithMaxVol(benchmark, 63, 63, 0.15, 3, 0.05, 0.00, 0.30), "benchmark_ranker_proxy_h63_riskcap_vol30", "benchmark_ranker_proxy_h63_riskcap"),
			compositeMomentumFamilyFactory(normalized, benchmarkRankerProxyConfigWithMaxVol(benchmark, 63, 63, 0.15, 3, 0.05, 0.00, 0.40), "benchmark_ranker_proxy_h63_riskcap_vol40", "benchmark_ranker_proxy_h63_riskcap"),
			compositeMomentumFamilyFactory(normalized, benchmarkRankerProxyConfigWithMaxVol(benchmark, 63, 63, 0.10, 3, 0.04, 0.00, 0.35), "benchmark_ranker_proxy_h63_riskcap_sleeve10", "benchmark_ranker_proxy_h63_riskcap"),
			compositeMomentumFamilyFactory(normalized, benchmarkRankerProxyConfigWithMaxVol(benchmark, 63, 63, 0.15, 3, 0.05, 0.02, 0.35), "benchmark_ranker_proxy_h63_riskcap_strict", "benchmark_ranker_proxy_h63_riskcap"),
		}
	case "benchmark_ranker_proxy_h63_trendguard":
		if len(normalized) < 2 {
			return nil
		}
		benchmark := normalized[0]
		return []StrategyFactory{
			compositeMomentumFamilyFactory(normalized, benchmarkRankerProxyH63TrendGuardCheckpointConfig(benchmark), "benchmark_ranker_proxy_h63_trendguard_checkpoint", "benchmark_ranker_proxy_h63_trendguard"),
			compositeMomentumFamilyFactory(normalized, benchmarkRankerProxyH63TrendGuardConfig(benchmark, 63, 0.90, 0.10, -0.02), "benchmark_ranker_proxy_h63_trendguard_fast", "benchmark_ranker_proxy_h63_trendguard"),
			compositeMomentumFamilyFactory(normalized, benchmarkRankerProxyH63TrendGuardConfig(benchmark, 126, 1.00, 0.00, -0.02), "benchmark_ranker_proxy_h63_trendguard_voo_only", "benchmark_ranker_proxy_h63_trendguard"),
			compositeMomentumFamilyFactory(normalized, benchmarkRankerProxyH63TrendGuardConfig(benchmark, 189, 0.85, 0.15, 0.00), "benchmark_ranker_proxy_h63_trendguard_slow_defensive", "benchmark_ranker_proxy_h63_trendguard"),
			compositeMomentumFamilyFactory(normalized, benchmarkRankerProxyH63TrendGuardConfig(benchmark, 126, 0.95, 0.05, 0.00), "benchmark_ranker_proxy_h63_trendguard_light", "benchmark_ranker_proxy_h63_trendguard"),
		}
	case "benchmark_ranker_proxy_h63_liquidity":
		if len(normalized) < 2 {
			return nil
		}
		benchmark := normalized[0]
		return []StrategyFactory{
			compositeMomentumFamilyFactory(normalized, benchmarkRankerProxyH63LiquidityCheckpointConfig(benchmark), "benchmark_ranker_proxy_h63_liquidity_checkpoint", "benchmark_ranker_proxy_h63_liquidity"),
			compositeMomentumFamilyFactory(normalized, benchmarkRankerProxyH63LiquidityConfig(benchmark, 500_000_000, 63, 0.15, 0.05), "benchmark_ranker_proxy_h63_liquidity_500m", "benchmark_ranker_proxy_h63_liquidity"),
			compositeMomentumFamilyFactory(normalized, benchmarkRankerProxyH63LiquidityConfig(benchmark, 1_500_000_000, 63, 0.15, 0.05), "benchmark_ranker_proxy_h63_liquidity_1500m", "benchmark_ranker_proxy_h63_liquidity"),
			compositeMomentumFamilyFactory(normalized, benchmarkRankerProxyH63LiquidityConfig(benchmark, 2_000_000_000, 63, 0.15, 0.05), "benchmark_ranker_proxy_h63_liquidity_2000m", "benchmark_ranker_proxy_h63_liquidity"),
			compositeMomentumFamilyFactory(normalized, benchmarkRankerProxyH63LiquidityConfig(benchmark, 1_000_000_000, 126, 0.10, 0.04), "benchmark_ranker_proxy_h63_liquidity_sleeve10", "benchmark_ranker_proxy_h63_liquidity"),
		}
	case "benchmark_ranker_proxy_h84":
		if len(normalized) < 2 {
			return nil
		}
		benchmark := normalized[0]
		return []StrategyFactory{
			compositeMomentumFamilyFactory(normalized, benchmarkRankerProxyH84CheckpointConfig(benchmark), "benchmark_ranker_proxy_h84_checkpoint", "benchmark_ranker_proxy_h84"),
			compositeMomentumFamilyFactory(normalized, benchmarkRankerProxyConfig(benchmark, 63, 63, 0.15, 3, 0.05, 0.00), "benchmark_ranker_proxy_h63", "benchmark_ranker_proxy_h84"),
			compositeMomentumFamilyFactory(normalized, benchmarkRankerProxyConfig(benchmark, 63, 84, 0.10, 3, 0.04, 0.00), "benchmark_ranker_proxy_h84_sleeve10", "benchmark_ranker_proxy_h84"),
			compositeMomentumFamilyFactory(normalized, benchmarkRankerProxyConfig(benchmark, 63, 84, 0.15, 3, 0.05, 0.03), "benchmark_ranker_proxy_h84_strict", "benchmark_ranker_proxy_h84"),
			compositeMomentumFamilyFactory(normalized, benchmarkRankerProxyConfig(benchmark, 63, 126, 0.15, 3, 0.05, 0.00), "benchmark_ranker_proxy_h126", "benchmark_ranker_proxy_h84"),
		}
	case "benchmark_ranker_proxy_blend":
		if len(normalized) < 2 {
			return nil
		}
		benchmark := normalized[0]
		return []StrategyFactory{
			compositeMomentumFamilyFactory(normalized, benchmarkRankerProxyBlendCheckpointConfig(benchmark), "benchmark_ranker_proxy_blend_checkpoint", "benchmark_ranker_proxy_blend"),
			compositeMomentumFamilyFactory(normalized, benchmarkRankerProxyBlendConfig(benchmark, 63, []int{63, 84}, []float64{0.075, 0.075}, 0.05, 0.00, 0.45), "benchmark_ranker_proxy_blend_h63_h84", "benchmark_ranker_proxy_blend"),
			compositeMomentumFamilyFactory(normalized, benchmarkRankerProxyBlendConfig(benchmark, 63, []int{63, 126}, []float64{0.075, 0.075}, 0.05, 0.00, 0.45), "benchmark_ranker_proxy_blend_h63_h126", "benchmark_ranker_proxy_blend"),
			compositeMomentumFamilyFactory(normalized, benchmarkRankerProxyBlendConfig(benchmark, 63, []int{84, 126}, []float64{0.075, 0.075}, 0.05, 0.00, 0.45), "benchmark_ranker_proxy_blend_h84_h126", "benchmark_ranker_proxy_blend"),
			compositeMomentumFamilyFactory(normalized, benchmarkRankerProxyBlendConfig(benchmark, 63, []int{63, 84, 126}, []float64{0.04, 0.04, 0.04}, 0.04, 0.02, 0.40), "benchmark_ranker_proxy_blend_conservative", "benchmark_ranker_proxy_blend"),
		}
	case "benchmark_lgbm_ranker":
		if len(normalized) < 2 {
			return nil
		}
		benchmark := normalized[0]
		return []StrategyFactory{
			dailyRankerFamilyFactory(normalized, "stocks_h63_s15_top3_reb63_z10", dailyRankerSleeveConfig(benchmark, 0.15, 3, 0.05, 63, 1.0, 0, 0, 1, 0, 1), "benchmark_lgbm_ranker_h63_s15"),
			dailyRankerFamilyFactory(normalized, "stocks_h63_s10_top3_reb63_z10", dailyRankerSleeveConfig(benchmark, 0.10, 3, 0.04, 63, 1.0, 0, 0, 1, 0, 1), "benchmark_lgbm_ranker_h63_s10"),
			dailyRankerFamilyFactory(normalized, "stocks_h126_s15_top3_reb63_z10", dailyRankerSleeveConfig(benchmark, 0.15, 3, 0.05, 63, 1.0, 0, 0, 1, 0, 1), "benchmark_lgbm_ranker_h126_s15"),
			dailyRankerFamilyFactory(normalized, "stocks_h126_s10_top3_reb63_z10", dailyRankerSleeveConfig(benchmark, 0.10, 3, 0.04, 63, 1.0, 0, 0, 1, 0, 1), "benchmark_lgbm_ranker_h126_s10"),
		}
	case "benchmark_lgbm_ranker_h63":
		if len(normalized) < 2 {
			return nil
		}
		benchmark := normalized[0]
		return []StrategyFactory{
			dailyRankerFamilyFactoryWithFamily(normalized, "stocks_h63_s15_top3_reb63_z10", dailyRankerSleeveConfig(benchmark, 0.15, 3, 0.05, 63, 1.0, 0, 0, 1, 0, 1), "benchmark_lgbm_ranker_h63_s15_checkpoint", "benchmark_lgbm_ranker_h63"),
			dailyRankerFamilyFactoryWithFamily(normalized, "stocks_h63_s10_top3_reb63_z10", dailyRankerSleeveConfig(benchmark, 0.10, 3, 0.04, 63, 1.0, 0, 0, 1, 0, 1), "benchmark_lgbm_ranker_h63_s10", "benchmark_lgbm_ranker_h63"),
			dailyRankerFamilyFactoryWithFamily(normalized, "stocks_h63_s15_top3_reb63_z10", dailyRankerSleeveConfig(benchmark, 0.15, 3, 0.05, 63, 1.25, 0, 0, 1, 0, 1), "benchmark_lgbm_ranker_h63_s15_z125", "benchmark_lgbm_ranker_h63"),
			dailyRankerFamilyFactoryWithFamily(normalized, "stocks_h63_s10_top3_reb63_z10", dailyRankerSleeveConfig(benchmark, 0.10, 3, 0.04, 63, 1.25, 0, 0, 1, 0, 1), "benchmark_lgbm_ranker_h63_s10_z125", "benchmark_lgbm_ranker_h63"),
		}
	case "benchmark_lgbm_ranker_h63_exmegacap":
		if len(normalized) < 2 {
			return nil
		}
		benchmark := normalized[0]
		exclusions := megaCapExclusions()
		return []StrategyFactory{
			dailyRankerFamilyFactoryWithFamily(normalized, "stocks_h63_s15_top3_reb63_z10", dailyRankerSleeveConfigExcluding(benchmark, 0.15, 3, 0.05, 63, 1.0, 0, 0, 1, 0, 1, exclusions), "benchmark_lgbm_ranker_h63_s15_exmegacap", "benchmark_lgbm_ranker_h63_exmegacap"),
			dailyRankerFamilyFactoryWithFamily(normalized, "stocks_h63_s10_top3_reb63_z10", dailyRankerSleeveConfigExcluding(benchmark, 0.10, 3, 0.04, 63, 1.0, 0, 0, 1, 0, 1, exclusions), "benchmark_lgbm_ranker_h63_s10_exmegacap", "benchmark_lgbm_ranker_h63_exmegacap"),
			dailyRankerFamilyFactoryWithFamily(normalized, "stocks_h63_s15_top3_reb63_z10", dailyRankerSleeveConfigExcluding(benchmark, 0.15, 3, 0.05, 63, 1.25, 0, 0, 1, 0, 1, exclusions), "benchmark_lgbm_ranker_h63_s15_z125_exmegacap", "benchmark_lgbm_ranker_h63_exmegacap"),
			dailyRankerFamilyFactoryWithFamily(normalized, "stocks_h63_s10_top3_reb63_z10", dailyRankerSleeveConfigExcluding(benchmark, 0.10, 3, 0.04, 63, 1.25, 0, 0, 1, 0, 1, exclusions), "benchmark_lgbm_ranker_h63_s10_z125_exmegacap", "benchmark_lgbm_ranker_h63_exmegacap"),
		}
	case "benchmark_lgbm_ranker_h63_equal":
		if len(normalized) < 2 {
			return nil
		}
		benchmark := normalized[0]
		return []StrategyFactory{
			dailyRankerFamilyFactoryWithBenchmark(normalized, "stocks_h63_s15_top3_reb63_z10", dailyRankerSleeveConfig(benchmark, 0.15, 3, 0.05, 63, 1.0, 0, 0, 1, 0, 1), "benchmark_lgbm_ranker_h63_s15_equal_benchmark", "benchmark_lgbm_ranker_h63_equal", "equal_weight"),
			dailyRankerFamilyFactoryWithBenchmark(normalized, "stocks_h63_s10_top3_reb63_z10", dailyRankerSleeveConfig(benchmark, 0.10, 3, 0.04, 63, 1.0, 0, 0, 1, 0, 1), "benchmark_lgbm_ranker_h63_s10_equal_benchmark", "benchmark_lgbm_ranker_h63_equal", "equal_weight"),
			dailyRankerFamilyFactoryWithBenchmark(normalized, "stocks_h63_s15_top3_reb63_z10", dailyRankerSleeveConfig(benchmark, 0.15, 3, 0.05, 63, 1.25, 0, 0, 1, 0, 1), "benchmark_lgbm_ranker_h63_s15_z125_equal_benchmark", "benchmark_lgbm_ranker_h63_equal", "equal_weight"),
			dailyRankerFamilyFactoryWithBenchmark(normalized, "stocks_h63_s10_top3_reb63_z10", dailyRankerSleeveConfig(benchmark, 0.10, 3, 0.04, 63, 1.25, 0, 0, 1, 0, 1), "benchmark_lgbm_ranker_h63_s10_z125_equal_benchmark", "benchmark_lgbm_ranker_h63_equal", "equal_weight"),
		}
	case "sector_ranked_sleeve":
		if len(normalized) < 2 {
			return nil
		}
		benchmark := normalized[0]
		return []StrategyFactory{
			compositeMomentumFamilyFactory(normalized, sectorRankedSleeveCheckpointConfig(benchmark), "sector_ranked_sleeve_checkpoint", "sector_ranked_sleeve"),
			compositeMomentumFamilyFactory(normalized, sectorRankedSleeveConfig(benchmark, 42, 126, 0.20, 3, 0.08, 0.00), "sector_ranked_sleeve_conservative", "sector_ranked_sleeve"),
			compositeMomentumFamilyFactory(normalized, sectorRankedSleeveConfig(benchmark, 21, 189, 0.30, 3, 0.10, 0.02), "sector_ranked_sleeve_medium", "sector_ranked_sleeve"),
			compositeMomentumFamilyFactory(normalized, sectorRankedSleeveConfig(benchmark, 63, 252, 0.25, 4, 0.08, -0.01), "sector_ranked_sleeve_slow", "sector_ranked_sleeve"),
		}
	case "kalman_cointegration":
		if len(normalized) < 2 {
			return nil
		}
		return []StrategyFactory{
			pairFactory(normalized[0], normalized[1], 1.5, "kalman_cointegration_z1_5"),
			pairFactory(normalized[0], normalized[1], 2.0, "kalman_cointegration_z2"),
			pairFactory(normalized[0], normalized[1], 2.5, "kalman_cointegration_z2_5"),
		}
	default:
		return []StrategyFactory{factory}
	}
}

func agentCatalogFactories(symbols []string, selected map[string]bool) []StrategyFactory {
	if len(symbols) == 0 {
		return nil
	}
	cfg := agentportfolio.DefaultStrategyCatalogConfig()
	specs := agentportfolio.AvailableStrategySpecs(symbols, cfg)
	out := make([]StrategyFactory, 0, len(specs))
	for _, spec := range specs {
		if !selected["agent_catalog"] && !selected[spec.Key] {
			continue
		}
		out = append(out, agentCatalogFactory(symbols, spec, cfg))
	}
	return out
}

func agentCatalogVariantFactories(family string, symbols []string) []StrategyFactory {
	if len(symbols) == 0 {
		return nil
	}
	cfg := agentportfolio.DefaultStrategyCatalogConfig()
	specs := agentportfolio.AvailableStrategySpecs(symbols, cfg)
	out := make([]StrategyFactory, 0, len(specs))
	for _, spec := range specs {
		if agentCatalogFamily(spec.RiskProfile) != family {
			continue
		}
		out = append(out, agentCatalogFactory(symbols, spec, cfg))
	}
	return out
}

func agentCatalogFactory(symbols []string, spec agentportfolio.StrategySpec, cfg agentportfolio.StrategyCatalogConfig) StrategyFactory {
	normalized := normalizeSymbols(symbols)
	return StrategyFactory{
		Name:      spec.Key,
		Family:    agentCatalogFamily(spec.RiskProfile),
		Benchmark: "buy_hold",
		New: func() backtest.PortfolioStrategy {
			strategy, _, err := agentportfolio.NewStrategyFromCatalog(spec.Key, normalized, cfg)
			if err != nil {
				return nil
			}
			return strategy
		},
	}
}

func agentCatalogFamily(profile agentportfolio.StrategyRiskProfile) string {
	switch profile {
	case agentportfolio.StrategyRiskLow:
		return "agent_catalog_low"
	case agentportfolio.StrategyRiskMedium:
		return "agent_catalog_medium"
	case agentportfolio.StrategyRiskHigh:
		return "agent_catalog_high"
	default:
		return "agent_catalog_unknown"
	}
}

func compositeMomentumFactory(symbols []string, cfg momentum.CompositeMomentumConfig, name string) StrategyFactory {
	return compositeMomentumFamilyFactory(symbols, cfg, name, "composite_momentum")
}

func compositeMomentumFamilyFactory(symbols []string, cfg momentum.CompositeMomentumConfig, name string, family string) StrategyFactory {
	normalized := normalizeSymbols(symbols)
	return StrategyFactory{
		Name:      name,
		Family:    family,
		Benchmark: "buy_hold",
		New: func() backtest.PortfolioStrategy {
			return momentum.NewCompositeMomentumStrategy(normalized, cfg)
		},
	}
}

func dailyRankerFamilyFactory(symbols []string, variant string, cfg ranker.DailyRankerSleeveConfig, name string) StrategyFactory {
	return dailyRankerFamilyFactoryWithFamily(symbols, variant, cfg, name, "benchmark_lgbm_ranker")
}

func dailyRankerFamilyFactoryWithFamily(symbols []string, variant string, cfg ranker.DailyRankerSleeveConfig, name string, family string) StrategyFactory {
	return dailyRankerFamilyFactoryWithBenchmark(symbols, variant, cfg, name, family, "buy_hold")
}

func dailyRankerFamilyFactoryWithBenchmark(symbols []string, variant string, cfg ranker.DailyRankerSleeveConfig, name string, family string, benchmarkName string) StrategyFactory {
	normalized := normalizeSymbols(symbols)
	root := dailyRankerArtifactRoot()
	cfg.ModelArtifactRoot = root
	cfg.ModelVariant = variant
	cfg.ModelPathsByYear = ranker.DailyRankerModelPaths(root, variant, dailyRankerYears()...)
	return StrategyFactory{
		Name:      name,
		Family:    family,
		Benchmark: benchmarkName,
		New: func() backtest.PortfolioStrategy {
			return ranker.NewDailyRankerSleeveStrategy(normalized, cfg)
		},
	}
}

func dailyRankerArtifactRoot() string {
	if root := strings.TrimSpace(os.Getenv("OALPHA_DAILY_RANKER_ARTIFACT_ROOT")); root != "" {
		return root
	}
	return filepath.Clean("../reports/batches/2026-06-03_yahoo100_daily_ranker_walkforward_slow_horizons_2018_2026/fold_artifacts")
}

func dailyRankerYears() []int {
	return []int{2018, 2019, 2020, 2021, 2022, 2023, 2024, 2025, 2026}
}

func dailyRankerSleeveConfig(
	benchmark string,
	sleeve float64,
	topK int,
	maxName float64,
	rebalanceEvery int,
	minScoreZ float64,
	maxCandidateVol float64,
	maxBenchmarkVol float64,
	highVolScale float64,
	maxBenchmarkDrawdown float64,
	drawdownScale float64,
) ranker.DailyRankerSleeveConfig {
	return ranker.DailyRankerSleeveConfig{
		BenchmarkSymbol:         benchmark,
		CandidateUniverse:       "stocks",
		PointInTimeUniversePath: strings.TrimSpace(os.Getenv("OALPHA_DAILY_RANKER_PIT_UNIVERSE")),
		RebalanceEveryBars:      rebalanceEvery,
		SleeveFraction:          sleeve,
		TopK:                    topK,
		MaxNameWeight:           maxName,
		TurnoverBand:            0.05,
		MinScoreZ:               minScoreZ,
		MaxCandidateVol:         maxCandidateVol,
		MaxBenchmarkVol:         maxBenchmarkVol,
		HighVolScale:            highVolScale,
		MaxBenchmarkDrawdown:    maxBenchmarkDrawdown,
		DrawdownScale:           drawdownScale,
	}
}

func dailyRankerSleeveConfigExcluding(
	benchmark string,
	sleeve float64,
	topK int,
	maxName float64,
	rebalanceEvery int,
	minScoreZ float64,
	maxCandidateVol float64,
	maxBenchmarkVol float64,
	highVolScale float64,
	maxBenchmarkDrawdown float64,
	drawdownScale float64,
	excluded []string,
) ranker.DailyRankerSleeveConfig {
	cfg := dailyRankerSleeveConfig(
		benchmark,
		sleeve,
		topK,
		maxName,
		rebalanceEvery,
		minScoreZ,
		maxCandidateVol,
		maxBenchmarkVol,
		highVolScale,
		maxBenchmarkDrawdown,
		drawdownScale,
	)
	cfg.ExcludedSymbols = append([]string(nil), excluded...)
	return cfg
}

func megaCapExclusions() []string {
	return []string{"AAPL", "AMZN", "AVGO", "GOOG", "GOOGL", "LLY", "META", "MSFT", "NVDA", "TSLA"}
}

func compositeMomentumCheckpointConfig(benchmark string) momentum.CompositeMomentumConfig {
	cfg := momentum.DefaultCompositeMomentumConfig()
	cfg.BenchmarkSymbol = benchmark
	return cfg
}

func defensiveCompositeCheckpointConfig(benchmark string) momentum.CompositeMomentumConfig {
	return defensiveCompositeVariantConfig(benchmark, 200, 0.25, 0.15, 0.03)
}

func benchmarkTSMOMCheckpointConfig(benchmark string) momentum.CompositeMomentumConfig {
	return benchmarkTSMOMVariantConfig(benchmark, 63, 126, 252, 0.15, 0.08)
}

func benchmarkTSMOMBlendCheckpointConfig(benchmark string) momentum.CompositeMomentumConfig {
	return benchmarkTSMOMBlendConfig(benchmark, 63, 0.08, 0.05)
}

func benchmarkLowVolCheckpointConfig(benchmark string) momentum.CompositeMomentumConfig {
	return benchmarkLowVolConfig(benchmark, 21, 0.20, 5, "low_vol")
}

func benchmarkReversalCheckpointConfig(benchmark string) momentum.CompositeMomentumConfig {
	return benchmarkReversalConfig(benchmark, 5, 10, 0.15, 10)
}

func benchmarkRankedSleeveCheckpointConfig(benchmark string) momentum.CompositeMomentumConfig {
	return benchmarkRankedSleeveConfig(benchmark, 21, 189, 0.30, 5, 0.08, 0.03)
}

func benchmarkRankerProxyCheckpointConfig(benchmark string) momentum.CompositeMomentumConfig {
	return benchmarkRankerProxyConfig(benchmark, 42, 252, 0.10, 3, 0.04, 0.00)
}

func benchmarkRankerProxyH63CheckpointConfig(benchmark string) momentum.CompositeMomentumConfig {
	return benchmarkRankerProxyConfig(benchmark, 63, 63, 0.15, 3, 0.05, 0.00)
}

func benchmarkRankerProxyH63RiskCapCheckpointConfig(benchmark string) momentum.CompositeMomentumConfig {
	return benchmarkRankerProxyConfigWithMaxVol(benchmark, 63, 63, 0.15, 3, 0.05, 0.00, 0.35)
}

func benchmarkRankerProxyH63TrendGuardCheckpointConfig(benchmark string) momentum.CompositeMomentumConfig {
	return benchmarkRankerProxyH63TrendGuardConfig(benchmark, 126, 0.90, 0.10, -0.02)
}

func benchmarkRankerProxyH63LiquidityCheckpointConfig(benchmark string) momentum.CompositeMomentumConfig {
	return benchmarkRankerProxyH63LiquidityConfig(benchmark, 1_000_000_000, 63, 0.15, 0.05)
}

func benchmarkRankerProxyH84CheckpointConfig(benchmark string) momentum.CompositeMomentumConfig {
	return benchmarkRankerProxyConfig(benchmark, 63, 84, 0.15, 3, 0.05, 0.00)
}

func benchmarkRankerProxyBlendCheckpointConfig(benchmark string) momentum.CompositeMomentumConfig {
	return benchmarkRankerProxyBlendConfig(benchmark, 63, []int{63, 84, 126}, []float64{0.05, 0.05, 0.05}, 0.05, 0.00, 0.45)
}

func sectorRankedSleeveCheckpointConfig(benchmark string) momentum.CompositeMomentumConfig {
	return sectorRankedSleeveConfig(benchmark, 21, 189, 0.30, 3, 0.10, 0.01)
}

func benchmarkRankedSleeveConfig(
	benchmark string,
	rebalanceEvery int,
	lookback int,
	sleeve float64,
	topK int,
	maxName float64,
	minRelativeMomentum float64,
) momentum.CompositeMomentumConfig {
	cfg := momentum.DefaultCompositeMomentumConfig()
	cfg.BenchmarkSymbol = benchmark
	cfg.RebalanceEveryBars = rebalanceEvery
	cfg.GlobalMaxNameWeight = maxName
	cfg.TurnoverBand = 0.05
	cfg.Legs = []momentum.CompositeMomentumLegConfig{
		{
			Name:                "risk_budgeted_stocks",
			CandidateUniverse:   "stocks",
			RankMode:            "vol_adjusted_momentum",
			WeightMode:          "risk_adjusted_edge",
			LookbackBars:        lookback,
			SleeveFraction:      sleeve,
			TopK:                topK,
			MaxNameWeight:       maxName,
			MinRelativeMomentum: minRelativeMomentum,
			MaxVol20:            0.45,
			EdgeExponent:        2,
			VolFloor:            0.10,
		},
	}
	return cfg
}

func benchmarkRankerProxyConfig(
	benchmark string,
	rebalanceEvery int,
	lookback int,
	sleeve float64,
	topK int,
	maxName float64,
	minRelativeMomentum float64,
) momentum.CompositeMomentumConfig {
	return benchmarkRankerProxyConfigWithMaxVol(benchmark, rebalanceEvery, lookback, sleeve, topK, maxName, minRelativeMomentum, 0.45)
}

func benchmarkRankerProxyConfigWithMaxVol(
	benchmark string,
	rebalanceEvery int,
	lookback int,
	sleeve float64,
	topK int,
	maxName float64,
	minRelativeMomentum float64,
	maxVol20 float64,
) momentum.CompositeMomentumConfig {
	cfg := momentum.DefaultCompositeMomentumConfig()
	cfg.BenchmarkSymbol = benchmark
	cfg.RebalanceEveryBars = rebalanceEvery
	cfg.GlobalMaxNameWeight = maxName
	cfg.TurnoverBand = 0.05
	cfg.Legs = []momentum.CompositeMomentumLegConfig{
		{
			Name:                "ranker_proxy_stocks",
			CandidateUniverse:   "stocks",
			RankMode:            "vol_adjusted_momentum",
			WeightMode:          "risk_adjusted_edge",
			LookbackBars:        lookback,
			SleeveFraction:      sleeve,
			TopK:                topK,
			MaxNameWeight:       maxName,
			MinRelativeMomentum: minRelativeMomentum,
			MaxVol20:            maxVol20,
			EdgeExponent:        2,
			VolFloor:            0.10,
		},
	}
	return cfg
}

func benchmarkRankerProxyH63TrendGuardConfig(
	benchmark string,
	trendLookback int,
	riskOffBenchmarkWeight float64,
	riskOffSleeve float64,
	riskOffMinRelativeMomentum float64,
) momentum.CompositeMomentumConfig {
	cfg := benchmarkRankerProxyConfig(benchmark, 63, 63, 0.15, 3, 0.05, 0.00)
	cfg.BenchmarkTrendLookbackBars = trendLookback
	cfg.MinBenchmarkTrend = 0
	cfg.RiskOffBenchmarkWeight = riskOffBenchmarkWeight
	cfg.RiskOffLegs = []momentum.CompositeMomentumLegConfig{
		{
			Name:                "riskoff_defensive_etfs",
			CandidateSymbols:    []string{"XLU", "XLP", "XLV"},
			RankMode:            "vol_adjusted_momentum",
			WeightMode:          "risk_adjusted_edge",
			LookbackBars:        63,
			SleeveFraction:      riskOffSleeve,
			TopK:                2,
			MaxNameWeight:       riskOffSleeve / 2,
			MinRelativeMomentum: riskOffMinRelativeMomentum,
			MaxVol20:            0.25,
			EdgeExponent:        1.5,
			VolFloor:            0.08,
		},
	}
	return cfg
}

func benchmarkRankerProxyH63LiquidityConfig(
	benchmark string,
	minMedianDollarVolume float64,
	liquidityLookback int,
	sleeve float64,
	maxName float64,
) momentum.CompositeMomentumConfig {
	cfg := benchmarkRankerProxyConfig(benchmark, 63, 63, sleeve, 3, maxName, 0.00)
	if len(cfg.Legs) > 0 {
		cfg.Legs[0].Name = "ranker_proxy_liquid_stocks"
		cfg.Legs[0].DollarVolumeLookbackBars = liquidityLookback
		cfg.Legs[0].MinMedianDollarVolume = minMedianDollarVolume
	}
	return cfg
}

func benchmarkRankerProxyBlendConfig(
	benchmark string,
	rebalanceEvery int,
	lookbacks []int,
	sleeves []float64,
	maxName float64,
	minRelativeMomentum float64,
	maxVol20 float64,
) momentum.CompositeMomentumConfig {
	cfg := momentum.DefaultCompositeMomentumConfig()
	cfg.BenchmarkSymbol = benchmark
	cfg.RebalanceEveryBars = rebalanceEvery
	cfg.GlobalMaxNameWeight = maxName
	cfg.TurnoverBand = 0.05
	cfg.Legs = make([]momentum.CompositeMomentumLegConfig, 0, len(lookbacks))
	for i, lookback := range lookbacks {
		sleeve := 0.0
		if i < len(sleeves) {
			sleeve = sleeves[i]
		}
		if sleeve <= 0 {
			continue
		}
		cfg.Legs = append(cfg.Legs, momentum.CompositeMomentumLegConfig{
			Name:                fmt.Sprintf("ranker_proxy_stocks_h%d", lookback),
			CandidateUniverse:   "stocks",
			RankMode:            "vol_adjusted_momentum",
			WeightMode:          "risk_adjusted_edge",
			LookbackBars:        lookback,
			SleeveFraction:      sleeve,
			TopK:                3,
			MaxNameWeight:       maxName,
			MinRelativeMomentum: minRelativeMomentum,
			MaxVol20:            maxVol20,
			EdgeExponent:        2,
			VolFloor:            0.10,
		})
	}
	return cfg
}

func sectorRankedSleeveConfig(
	benchmark string,
	rebalanceEvery int,
	lookback int,
	sleeve float64,
	topK int,
	maxName float64,
	minRelativeMomentum float64,
) momentum.CompositeMomentumConfig {
	cfg := momentum.DefaultCompositeMomentumConfig()
	cfg.BenchmarkSymbol = benchmark
	cfg.RebalanceEveryBars = rebalanceEvery
	cfg.GlobalMaxNameWeight = maxName
	cfg.TurnoverBand = 0.05
	cfg.Legs = []momentum.CompositeMomentumLegConfig{
		{
			Name:                "risk_budgeted_sector_etfs",
			CandidateSymbols:    []string{"DIA", "IWM", "QQQ", "SMH", "VTI", "XLB", "XLE", "XLF", "XLI", "XLK", "XLP", "XLU", "XLV", "XLY"},
			RankMode:            "vol_adjusted_momentum",
			WeightMode:          "risk_adjusted_edge",
			LookbackBars:        lookback,
			SleeveFraction:      sleeve,
			TopK:                topK,
			MaxNameWeight:       maxName,
			MinRelativeMomentum: minRelativeMomentum,
			MaxVol20:            0.35,
			EdgeExponent:        2,
			VolFloor:            0.08,
		},
	}
	return cfg
}

func benchmarkReversalConfig(
	benchmark string,
	rebalanceEvery int,
	lookback int,
	sleeve float64,
	topK int,
) momentum.CompositeMomentumConfig {
	cfg := momentum.DefaultCompositeMomentumConfig()
	cfg.BenchmarkSymbol = benchmark
	cfg.RebalanceEveryBars = rebalanceEvery
	cfg.GlobalMaxNameWeight = 0.05
	cfg.Legs = []momentum.CompositeMomentumLegConfig{
		{
			Name:                "stock_reversal",
			CandidateUniverse:   "stocks",
			RankMode:            "mean_reversion",
			LookbackBars:        lookback,
			SleeveFraction:      sleeve,
			TopK:                topK,
			MaxNameWeight:       sleeve / float64(topK),
			MinRelativeMomentum: -1,
			MaxVol20:            0.60,
		},
	}
	return cfg
}

func benchmarkLowVolConfig(
	benchmark string,
	rebalanceEvery int,
	sleeve float64,
	topK int,
	rankMode string,
) momentum.CompositeMomentumConfig {
	cfg := momentum.DefaultCompositeMomentumConfig()
	cfg.BenchmarkSymbol = benchmark
	cfg.RebalanceEveryBars = rebalanceEvery
	cfg.GlobalMaxNameWeight = 0.10
	cfg.Legs = []momentum.CompositeMomentumLegConfig{
		{
			Name:                "stock_lowvol",
			CandidateUniverse:   "stocks",
			RankMode:            rankMode,
			LookbackBars:        63,
			SleeveFraction:      sleeve,
			TopK:                topK,
			MaxNameWeight:       sleeve / float64(topK),
			MinRelativeMomentum: -1,
			MaxVol20:            0.35,
		},
	}
	return cfg
}

func benchmarkTSMOMBlendConfig(
	benchmark string,
	rebalanceEvery int,
	etfSleevePerLeg float64,
	broadSleevePerLeg float64,
) momentum.CompositeMomentumConfig {
	cfg := momentum.DefaultCompositeMomentumConfig()
	cfg.BenchmarkSymbol = benchmark
	cfg.RebalanceEveryBars = rebalanceEvery
	cfg.GlobalMaxNameWeight = 0.20
	cfg.Legs = []momentum.CompositeMomentumLegConfig{
		{
			Name:                "etf_126",
			CandidateUniverse:   "etfs",
			LookbackBars:        126,
			SleeveFraction:      etfSleevePerLeg,
			TopK:                2,
			MaxNameWeight:       etfSleevePerLeg / 2,
			MinRelativeMomentum: 0.04,
			MaxVol20:            0.30,
		},
		{
			Name:                "etf_189",
			CandidateUniverse:   "etfs",
			LookbackBars:        189,
			SleeveFraction:      etfSleevePerLeg,
			TopK:                2,
			MaxNameWeight:       etfSleevePerLeg / 2,
			MinRelativeMomentum: 0.04,
			MaxVol20:            0.30,
		},
		{
			Name:                "broad_189",
			CandidateUniverse:   "all",
			LookbackBars:        189,
			SleeveFraction:      broadSleevePerLeg,
			TopK:                5,
			MaxNameWeight:       broadSleevePerLeg / 5,
			MinRelativeMomentum: 0.08,
			MaxVol20:            0.45,
		},
		{
			Name:                "broad_252",
			CandidateUniverse:   "all",
			LookbackBars:        252,
			SleeveFraction:      broadSleevePerLeg,
			TopK:                5,
			MaxNameWeight:       broadSleevePerLeg / 5,
			MinRelativeMomentum: 0.08,
			MaxVol20:            0.45,
		},
	}
	return cfg
}

func benchmarkTSMOMVariantConfig(
	benchmark string,
	rebalanceEvery int,
	etfLookback int,
	broadLookback int,
	etfSleeve float64,
	broadSleeve float64,
) momentum.CompositeMomentumConfig {
	cfg := momentum.DefaultCompositeMomentumConfig()
	cfg.BenchmarkSymbol = benchmark
	cfg.RebalanceEveryBars = rebalanceEvery
	cfg.GlobalMaxNameWeight = 0.25
	cfg.Legs = []momentum.CompositeMomentumLegConfig{
		{
			Name:                "etf_tsmom",
			CandidateUniverse:   "etfs",
			LookbackBars:        etfLookback,
			SleeveFraction:      etfSleeve,
			TopK:                2,
			MaxNameWeight:       etfSleeve / 2,
			MinRelativeMomentum: 0.04,
			MaxVol20:            0.30,
		},
		{
			Name:                "broad_tsmom",
			CandidateUniverse:   "all",
			LookbackBars:        broadLookback,
			SleeveFraction:      broadSleeve,
			TopK:                5,
			MaxNameWeight:       broadSleeve / 5,
			MinRelativeMomentum: 0.08,
			MaxVol20:            0.45,
		},
	}
	return cfg
}

func defensiveCompositeVariantConfig(
	benchmark string,
	trendLookback int,
	riskOffBenchmarkWeight float64,
	etfSleeve float64,
	broadSleeve float64,
) momentum.CompositeMomentumConfig {
	cfg := compositeMomentumVariantConfig(benchmark, etfSleeve, 0.06, 0.24, broadSleeve)
	cfg.GlobalMaxNameWeight = 0.50
	cfg.BenchmarkTrendLookbackBars = trendLookback
	cfg.MinBenchmarkTrend = 0
	cfg.RiskOffBenchmarkWeight = riskOffBenchmarkWeight
	cfg.RiskOffLegs = []momentum.CompositeMomentumLegConfig{
		{
			Name:                "riskoff_defensive_63",
			CandidateSymbols:    []string{"XLU", "XLP", "XLV"},
			LookbackBars:        63,
			SleeveFraction:      math.Max(0, 1-riskOffBenchmarkWeight),
			TopK:                2,
			MaxNameWeight:       0.50,
			MinRelativeMomentum: -0.02,
			MaxVol20:            0.25,
		},
	}
	return cfg
}

func compositeMomentumVariantConfig(
	benchmark string,
	etfSleeve float64,
	etfMinRelativeMomentum float64,
	etfMaxVol20 float64,
	broadSleeve float64,
) momentum.CompositeMomentumConfig {
	cfg := momentum.DefaultCompositeMomentumConfig()
	cfg.BenchmarkSymbol = benchmark
	cfg.Legs = []momentum.CompositeMomentumLegConfig{
		{
			Name:                "etf_21",
			CandidateUniverse:   "etfs",
			LookbackBars:        21,
			SleeveFraction:      etfSleeve,
			TopK:                1,
			MaxNameWeight:       etfSleeve,
			MinRelativeMomentum: etfMinRelativeMomentum,
			MaxVol20:            etfMaxVol20,
		},
		{
			Name:                "all_126",
			CandidateUniverse:   "all",
			LookbackBars:        126,
			SleeveFraction:      broadSleeve,
			TopK:                5,
			MaxNameWeight:       broadSleeve / 5,
			MinRelativeMomentum: 0.10,
		},
	}
	return cfg
}

func maFactory(symbol string, fast, slow int, name string) StrategyFactory {
	return StrategyFactory{
		Name:      name,
		Family:    "ma_crossover",
		Benchmark: "buy_hold",
		New: func() backtest.PortfolioStrategy {
			return &singleSymbolPortfolioStrategy{
				name:   name,
				symbol: symbol,
				newStrategy: func() backtest.Strategy {
					return backtest.NewMACrossoverStrategy(fast, slow)
				},
			}
		},
	}
}

func kalmanFactory(symbol string, q, r, z float64, name string) StrategyFactory {
	return StrategyFactory{
		Name:      name,
		Family:    "kalman",
		Benchmark: "buy_hold",
		New: func() backtest.PortfolioStrategy {
			return &singleSymbolPortfolioStrategy{
				name:   name,
				symbol: symbol,
				newStrategy: func() backtest.Strategy {
					return backtest.NewKalmanStrategy(q, r, 20, z)
				},
			}
		},
	}
}

func xsecFactory(symbols []string, topFraction float64, name string) StrategyFactory {
	normalized := normalizeSymbols(symbols)
	return StrategyFactory{
		Name:      name,
		Family:    "xsec_momentum",
		Benchmark: "equal_weight",
		New: func() backtest.PortfolioStrategy {
			cfg := momentum.DefaultCrossSectionalMomentumConfig()
			cfg.TopFraction = topFraction
			if len(normalized) < cfg.MinPositions {
				cfg.MinPositions = maxInt(1, len(normalized)/2)
			}
			cfg.MaxPositions = maxInt(cfg.MinPositions, minInt(cfg.MaxPositions, len(normalized)))
			return momentum.NewCrossSectionalMomentumStrategy(normalized, cfg, nil)
		},
	}
}

func pairFactory(symbolY, symbolX string, entryZ float64, name string) StrategyFactory {
	return StrategyFactory{
		Name:        name,
		Family:      "kalman_cointegration",
		Benchmark:   "flat_cash",
		AllowShorts: true,
		New: func() backtest.PortfolioStrategy {
			cfg := cointegration.DefaultKalmanPairConfig(symbolY, symbolX)
			cfg.EntryZ = entryZ
			return cointegration.NewKalmanPairStrategy(cfg, nil, nil)
		},
	}
}

type singleSymbolPortfolioStrategy struct {
	name        string
	symbol      string
	newStrategy func() backtest.Strategy
	strategy    backtest.Strategy
}

func (s *singleSymbolPortfolioStrategy) GeneratePortfolioSignals(ctx context.Context, panel backtest.AlignedBars) ([]backtest.PortfolioOutput, error) {
	outputs := make([]backtest.PortfolioOutput, len(panel.Times))
	for i := range panel.Times {
		output, err := s.EvaluatePortfolioLatest(ctx, panelPrefix(panel, i+1))
		if err != nil {
			return nil, err
		}
		outputs[i] = output
	}
	return outputs, nil
}

func (s *singleSymbolPortfolioStrategy) EvaluatePortfolioLatest(ctx context.Context, panel backtest.AlignedBars) (backtest.PortfolioOutput, error) {
	bars := panel.Bars[s.symbol]
	t := time.Time{}
	if len(panel.Times) > 0 {
		t = panel.Times[len(panel.Times)-1]
	}
	if len(bars) == 0 {
		return holdPortfolioOutput(t, "missing_symbol_bars"), nil
	}
	if s.strategy == nil {
		s.strategy = s.newStrategy()
	}
	output, err := s.strategy.EvaluateLatest(ctx, bars)
	if err != nil {
		return holdPortfolioOutput(t, "warmup_or_strategy_error"), nil
	}
	return strategyOutputToPortfolio(s.symbol, t, s.name, output), nil
}

func (s *singleSymbolPortfolioStrategy) Universe() []string {
	return []string{s.symbol}
}

func (s *singleSymbolPortfolioStrategy) Name() string {
	return s.name
}

type buyHoldStrategy struct {
	symbol string
	seen   bool
}

func (s *buyHoldStrategy) GeneratePortfolioSignals(ctx context.Context, panel backtest.AlignedBars) ([]backtest.PortfolioOutput, error) {
	outputs := make([]backtest.PortfolioOutput, len(panel.Times))
	for i := range panel.Times {
		output, err := s.EvaluatePortfolioLatest(ctx, panelPrefix(panel, i+1))
		if err != nil {
			return nil, err
		}
		outputs[i] = output
	}
	return outputs, nil
}

func (s *buyHoldStrategy) EvaluatePortfolioLatest(ctx context.Context, panel backtest.AlignedBars) (backtest.PortfolioOutput, error) {
	_ = ctx
	t := panel.Times[len(panel.Times)-1]
	if s.seen {
		return holdPortfolioOutput(t, "hold_buy_and_hold"), nil
	}
	s.seen = true
	return backtest.PortfolioOutput{
		Time: t,
		Targets: map[string]backtest.TargetPosition{
			s.symbol: {
				Symbol:       s.symbol,
				TargetWeight: 1,
				AlphaScore:   1,
				Confidence:   1,
				Side:         backtest.PositionSideLong,
				Engine:       "buy_hold",
			},
		},
		GrossExposure: 1,
		NetExposure:   1,
		CashWeight:    0,
	}, nil
}

func (s *buyHoldStrategy) Universe() []string { return []string{s.symbol} }
func (s *buyHoldStrategy) Name() string       { return "buy_hold" }

type equalWeightStrategy struct {
	symbols []string
	seen    bool
}

func (s *equalWeightStrategy) GeneratePortfolioSignals(ctx context.Context, panel backtest.AlignedBars) ([]backtest.PortfolioOutput, error) {
	outputs := make([]backtest.PortfolioOutput, len(panel.Times))
	for i := range panel.Times {
		output, err := s.EvaluatePortfolioLatest(ctx, panelPrefix(panel, i+1))
		if err != nil {
			return nil, err
		}
		outputs[i] = output
	}
	return outputs, nil
}

func (s *equalWeightStrategy) EvaluatePortfolioLatest(ctx context.Context, panel backtest.AlignedBars) (backtest.PortfolioOutput, error) {
	_ = ctx
	t := panel.Times[len(panel.Times)-1]
	if s.seen {
		return holdPortfolioOutput(t, "hold_equal_weight"), nil
	}
	s.seen = true
	targets := make(map[string]backtest.TargetPosition, len(s.symbols))
	weight := 1 / float64(len(s.symbols))
	for _, symbol := range s.symbols {
		targets[symbol] = backtest.TargetPosition{
			Symbol:       symbol,
			TargetWeight: weight,
			AlphaScore:   1,
			Confidence:   1,
			Side:         backtest.PositionSideLong,
			Engine:       "equal_weight",
		}
	}
	return backtest.PortfolioOutput{
		Time:          t,
		Targets:       targets,
		GrossExposure: 1,
		NetExposure:   1,
		CashWeight:    0,
	}, nil
}

func (s *equalWeightStrategy) Universe() []string { return append([]string(nil), s.symbols...) }
func (s *equalWeightStrategy) Name() string       { return "equal_weight" }

type flatStrategy struct {
	symbols []string
}

func (s *flatStrategy) GeneratePortfolioSignals(ctx context.Context, panel backtest.AlignedBars) ([]backtest.PortfolioOutput, error) {
	outputs := make([]backtest.PortfolioOutput, len(panel.Times))
	for i, t := range panel.Times {
		outputs[i] = holdPortfolioOutput(t, "flat_cash")
	}
	return outputs, nil
}

func (s *flatStrategy) EvaluatePortfolioLatest(ctx context.Context, panel backtest.AlignedBars) (backtest.PortfolioOutput, error) {
	_ = ctx
	return holdPortfolioOutput(panel.Times[len(panel.Times)-1], "flat_cash"), nil
}

func (s *flatStrategy) Universe() []string { return append([]string(nil), s.symbols...) }
func (s *flatStrategy) Name() string       { return "flat_cash" }

func strategyOutputToPortfolio(symbol string, t time.Time, engine string, output backtest.StrategyOutput) backtest.PortfolioOutput {
	weight := output.TargetWeight
	if weight == 0 {
		weight = output.PositionSizePct
	}
	switch output.Signal {
	case models.SignalBuy:
		if weight <= 0 {
			weight = 0.10
		}
	case models.SignalSell:
		weight = 0
	default:
		if weight == 0 {
			return holdPortfolioOutput(t, "base_hold")
		}
	}
	if weight > 1 {
		weight = 1
	}
	if weight < -1 {
		weight = -1
	}
	side := backtest.PositionSideLong
	if weight < 0 {
		side = backtest.PositionSideShort
	}
	return backtest.PortfolioOutput{
		Time: t,
		Targets: map[string]backtest.TargetPosition{
			symbol: {
				Symbol:       symbol,
				TargetWeight: weight,
				AlphaScore:   output.AlphaScore,
				Confidence:   output.Confidence,
				Side:         side,
				Engine:       engine,
				RegimeLabel:  output.RegimeLabel,
				Metadata:     output.Metadata,
			},
		},
		GrossExposure: math.Abs(weight),
		NetExposure:   weight,
		CashWeight:    math.Max(0, 1-math.Abs(weight)),
	}
}

func holdPortfolioOutput(t time.Time, reason string) backtest.PortfolioOutput {
	return backtest.PortfolioOutput{
		Time:       t,
		Targets:    map[string]backtest.TargetPosition{},
		CashWeight: 1,
		EngineMetadata: map[string]interface{}{
			"action": "hold_targets",
			"reason": reason,
		},
	}
}

func panelPrefix(panel backtest.AlignedBars, length int) backtest.AlignedBars {
	if length > len(panel.Times) {
		length = len(panel.Times)
	}
	out := backtest.AlignedBars{
		Times:      append([]time.Time(nil), panel.Times[:length]...),
		Symbols:    append([]string(nil), panel.Symbols...),
		Bars:       make(map[string][]models.Bar, len(panel.Bars)),
		Timeframe:  panel.Timeframe,
		Feed:       panel.Feed,
		Adjustment: panel.Adjustment,
		Metadata:   panel.Metadata,
	}
	for symbol, bars := range panel.Bars {
		if length > len(bars) {
			out.Bars[symbol] = append([]models.Bar(nil), bars...)
			continue
		}
		out.Bars[symbol] = append([]models.Bar(nil), bars[:length]...)
	}
	return out
}

func normalizeSymbols(symbols []string) []string {
	out := make([]string, 0, len(symbols))
	seen := make(map[string]bool, len(symbols))
	for _, symbol := range symbols {
		normalized := strings.ToUpper(strings.TrimSpace(symbol))
		if normalized == "" || seen[normalized] {
			continue
		}
		seen[normalized] = true
		out = append(out, normalized)
	}
	return out
}

func strategySet(strategyNames []string) map[string]bool {
	out := make(map[string]bool)
	if len(strategyNames) == 0 {
		out["all"] = true
		return out
	}
	for _, name := range strategyNames {
		normalized := strings.ToLower(strings.TrimSpace(name))
		if normalized != "" {
			out[normalized] = true
		}
	}
	return out
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func validateFactory(factory StrategyFactory) error {
	if factory.Name == "" {
		return fmt.Errorf("strategy factory name is required")
	}
	if factory.New == nil {
		return fmt.Errorf("strategy factory %s has no constructor", factory.Name)
	}
	return nil
}
