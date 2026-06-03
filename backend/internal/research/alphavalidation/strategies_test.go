package alphavalidation

import "testing"

func TestBenchmarkFactoriesIncludesBuyHoldForMultiSymbolPanel(t *testing.T) {
	factories := BenchmarkFactories([]string{"VOO", "AAPL", "MSFT"})
	if !hasFactory(factories, "buy_hold") {
		t.Fatalf("multi-symbol benchmark factories should include buy_hold")
	}
}

func TestCandidateFactoriesRegistersCompositeMomentum(t *testing.T) {
	factories := CandidateFactories([]string{"VOO", "AAPL", "MSFT", "XLU", "SMH"}, []string{"composite_momentum"})
	if !hasFactory(factories, "composite_momentum_checkpoint") {
		t.Fatalf("candidate factories missing composite momentum checkpoint: %#v", factories)
	}
}

func TestCandidateFactoriesRegistersAgentCatalogBuckets(t *testing.T) {
	factories := CandidateFactories(catalogSymbolsForValidation(), []string{"agent_catalog"})
	if len(factories) != 9 {
		t.Fatalf("agent catalog factories=%d, want 9", len(factories))
	}
	for _, name := range []string{"lgbm_ranker_h63_low", "lgbm_ranker_h63_medium", "lgbm_ranker_h63_high"} {
		if !hasFactory(factories, name) {
			t.Fatalf("agent catalog missing %s: %#v", name, factories)
		}
	}
	buckets := map[string]int{}
	for _, factory := range factories {
		buckets[factory.Family]++
		if factory.Benchmark != "buy_hold" {
			t.Fatalf("factory %s benchmark=%s, want buy_hold", factory.Name, factory.Benchmark)
		}
	}
	for _, family := range []string{"agent_catalog_low", "agent_catalog_medium", "agent_catalog_high"} {
		if buckets[family] != 3 {
			t.Fatalf("family %s count=%d, want 3", family, buckets[family])
		}
	}
}

func TestVariantFactoriesProvidesAgentCatalogRiskBucketPBOSet(t *testing.T) {
	factories := CandidateFactories(catalogSymbolsForValidation(), []string{"lgbm_ranker_h63_medium"})
	if len(factories) != 1 {
		t.Fatalf("exact catalog factories=%d, want 1", len(factories))
	}
	variants := VariantFactories(factories[0], catalogSymbolsForValidation())
	if len(variants) != 3 {
		t.Fatalf("agent catalog medium variants=%d, want 3", len(variants))
	}
	for _, variant := range variants {
		if variant.Family != "agent_catalog_medium" {
			t.Fatalf("variant %s family=%s, want agent_catalog_medium", variant.Name, variant.Family)
		}
	}
}

func TestCandidateFactoriesRegistersDefensiveBenchmarkRotation(t *testing.T) {
	factories := CandidateFactories([]string{"VOO", "AAPL", "MSFT", "XLU", "XLP", "XLV"}, []string{"benchmark_rotation"})
	if !hasFactory(factories, "benchmark_rotation_defensive") {
		t.Fatalf("candidate factories missing defensive benchmark rotation: %#v", factories)
	}
}

func TestCandidateFactoriesRegistersBenchmarkTSMOM(t *testing.T) {
	factories := CandidateFactories([]string{"VOO", "AAPL", "MSFT", "XLU", "XLP", "XLV"}, []string{"benchmark_tsmom"})
	if !hasFactory(factories, "benchmark_tsmom_checkpoint") {
		t.Fatalf("candidate factories missing benchmark TSMOM: %#v", factories)
	}
}

func TestCandidateFactoriesRegistersBenchmarkTSMOMBlend(t *testing.T) {
	factories := CandidateFactories([]string{"VOO", "AAPL", "MSFT", "XLU", "XLP", "XLV"}, []string{"benchmark_tsmom_blend"})
	if !hasFactory(factories, "benchmark_tsmom_blend") {
		t.Fatalf("candidate factories missing benchmark TSMOM blend: %#v", factories)
	}
}

func TestCandidateFactoriesRegistersBenchmarkLowVol(t *testing.T) {
	factories := CandidateFactories([]string{"VOO", "AAPL", "MSFT", "JNJ", "PG", "XLV"}, []string{"benchmark_lowvol"})
	if !hasFactory(factories, "benchmark_lowvol_checkpoint") {
		t.Fatalf("candidate factories missing benchmark low-vol: %#v", factories)
	}
}

func TestCandidateFactoriesRegistersBenchmarkReversal(t *testing.T) {
	factories := CandidateFactories([]string{"VOO", "AAPL", "MSFT", "JNJ", "PG", "XLV"}, []string{"benchmark_reversal"})
	if !hasFactory(factories, "benchmark_reversal_checkpoint") {
		t.Fatalf("candidate factories missing benchmark reversal: %#v", factories)
	}
}

func TestCandidateFactoriesRegistersBenchmarkRankedSleeve(t *testing.T) {
	factories := CandidateFactories([]string{"VOO", "AAPL", "MSFT", "JNJ", "PG", "XLV"}, []string{"benchmark_ranked_sleeve"})
	if !hasFactory(factories, "benchmark_ranked_sleeve_checkpoint") {
		t.Fatalf("candidate factories missing benchmark ranked sleeve: %#v", factories)
	}
}

func TestCandidateFactoriesRegistersSectorRankedSleeve(t *testing.T) {
	factories := CandidateFactories([]string{"VOO", "QQQ", "XLK", "XLF", "XLV", "XLE"}, []string{"sector_ranked_sleeve"})
	if !hasFactory(factories, "sector_ranked_sleeve_checkpoint") {
		t.Fatalf("candidate factories missing sector ranked sleeve: %#v", factories)
	}
}

func TestVariantFactoriesProvidesCompositeMomentumPBOSet(t *testing.T) {
	factories := CandidateFactories([]string{"VOO", "AAPL", "MSFT", "XLU", "SMH"}, []string{"composite_momentum"})
	if len(factories) != 1 {
		t.Fatalf("composite factories=%d, want 1", len(factories))
	}
	variants := VariantFactories(factories[0], []string{"VOO", "AAPL", "MSFT", "XLU", "SMH"})
	if len(variants) < 3 {
		t.Fatalf("composite variants=%d, want at least 3 for PBO", len(variants))
	}
	for _, variant := range variants {
		if variant.Benchmark != "buy_hold" {
			t.Fatalf("variant %s benchmark=%s, want buy_hold", variant.Name, variant.Benchmark)
		}
	}
}

func TestVariantFactoriesProvidesBenchmarkTSMOMPBOSet(t *testing.T) {
	factories := CandidateFactories([]string{"VOO", "AAPL", "MSFT", "XLU", "XLP", "XLV"}, []string{"benchmark_tsmom"})
	if len(factories) != 1 {
		t.Fatalf("benchmark TSMOM factories=%d, want 1", len(factories))
	}
	variants := VariantFactories(factories[0], []string{"VOO", "AAPL", "MSFT", "XLU", "XLP", "XLV"})
	if len(variants) < 3 {
		t.Fatalf("benchmark TSMOM variants=%d, want at least 3 for PBO", len(variants))
	}
	for _, variant := range variants {
		if variant.Family != "benchmark_tsmom" {
			t.Fatalf("variant %s family=%s, want benchmark_tsmom", variant.Name, variant.Family)
		}
		if variant.Benchmark != "buy_hold" {
			t.Fatalf("variant %s benchmark=%s, want buy_hold", variant.Name, variant.Benchmark)
		}
	}
}

func TestVariantFactoriesProvidesBenchmarkTSMOMBlendPBOSet(t *testing.T) {
	factories := CandidateFactories([]string{"VOO", "AAPL", "MSFT", "XLU", "XLP", "XLV"}, []string{"benchmark_tsmom_blend"})
	if len(factories) != 1 {
		t.Fatalf("benchmark TSMOM blend factories=%d, want 1", len(factories))
	}
	variants := VariantFactories(factories[0], []string{"VOO", "AAPL", "MSFT", "XLU", "XLP", "XLV"})
	if len(variants) < 3 {
		t.Fatalf("benchmark TSMOM blend variants=%d, want at least 3 for PBO", len(variants))
	}
	for _, variant := range variants {
		if variant.Family != "benchmark_tsmom_blend" {
			t.Fatalf("variant %s family=%s, want benchmark_tsmom_blend", variant.Name, variant.Family)
		}
	}
}

func TestVariantFactoriesProvidesBenchmarkLowVolPBOSet(t *testing.T) {
	factories := CandidateFactories([]string{"VOO", "AAPL", "MSFT", "JNJ", "PG", "XLV"}, []string{"benchmark_lowvol"})
	if len(factories) != 1 {
		t.Fatalf("benchmark low-vol factories=%d, want 1", len(factories))
	}
	variants := VariantFactories(factories[0], []string{"VOO", "AAPL", "MSFT", "JNJ", "PG", "XLV"})
	if len(variants) < 3 {
		t.Fatalf("benchmark low-vol variants=%d, want at least 3 for PBO", len(variants))
	}
	for _, variant := range variants {
		if variant.Family != "benchmark_lowvol" {
			t.Fatalf("variant %s family=%s, want benchmark_lowvol", variant.Name, variant.Family)
		}
	}
}

func TestVariantFactoriesProvidesBenchmarkReversalPBOSet(t *testing.T) {
	factories := CandidateFactories([]string{"VOO", "AAPL", "MSFT", "JNJ", "PG", "XLV"}, []string{"benchmark_reversal"})
	if len(factories) != 1 {
		t.Fatalf("benchmark reversal factories=%d, want 1", len(factories))
	}
	variants := VariantFactories(factories[0], []string{"VOO", "AAPL", "MSFT", "JNJ", "PG", "XLV"})
	if len(variants) < 3 {
		t.Fatalf("benchmark reversal variants=%d, want at least 3 for PBO", len(variants))
	}
	for _, variant := range variants {
		if variant.Family != "benchmark_reversal" {
			t.Fatalf("variant %s family=%s, want benchmark_reversal", variant.Name, variant.Family)
		}
	}
}

func TestVariantFactoriesProvidesBenchmarkRankedSleevePBOSet(t *testing.T) {
	factories := CandidateFactories([]string{"VOO", "AAPL", "MSFT", "JNJ", "PG", "XLV"}, []string{"benchmark_ranked_sleeve"})
	if len(factories) != 1 {
		t.Fatalf("benchmark ranked sleeve factories=%d, want 1", len(factories))
	}
	variants := VariantFactories(factories[0], []string{"VOO", "AAPL", "MSFT", "JNJ", "PG", "XLV"})
	if len(variants) < 3 {
		t.Fatalf("benchmark ranked sleeve variants=%d, want at least 3 for PBO", len(variants))
	}
	for _, variant := range variants {
		if variant.Family != "benchmark_ranked_sleeve" {
			t.Fatalf("variant %s family=%s, want benchmark_ranked_sleeve", variant.Name, variant.Family)
		}
		if variant.Benchmark != "buy_hold" {
			t.Fatalf("variant %s benchmark=%s, want buy_hold", variant.Name, variant.Benchmark)
		}
	}
}

func TestVariantFactoriesProvidesBenchmarkRankerProxyPBOSet(t *testing.T) {
	factories := CandidateFactories([]string{"VOO", "AAPL", "MSFT", "JNJ", "PG", "XLV"}, []string{"benchmark_ranker_proxy"})
	if len(factories) != 1 {
		t.Fatalf("benchmark ranker proxy factories=%d, want 1", len(factories))
	}
	variants := VariantFactories(factories[0], []string{"VOO", "AAPL", "MSFT", "JNJ", "PG", "XLV"})
	if len(variants) < 3 {
		t.Fatalf("benchmark ranker proxy variants=%d, want at least 3 for PBO", len(variants))
	}
	for _, variant := range variants {
		if variant.Family != "benchmark_ranker_proxy" {
			t.Fatalf("variant %s family=%s, want benchmark_ranker_proxy", variant.Name, variant.Family)
		}
		if variant.Benchmark != "buy_hold" {
			t.Fatalf("variant %s benchmark=%s, want buy_hold", variant.Name, variant.Benchmark)
		}
	}
}

func TestCandidateFactoriesRegistersBenchmarkRankerProxyH63(t *testing.T) {
	factories := CandidateFactories([]string{"VOO", "AAPL", "MSFT", "JNJ", "PG", "XLV"}, []string{"benchmark_ranker_proxy_h63"})
	if !hasFactory(factories, "benchmark_ranker_proxy_h63_checkpoint") {
		t.Fatalf("candidate factories missing benchmark ranker proxy h63: %#v", factories)
	}
}

func TestVariantFactoriesProvidesBenchmarkRankerProxyH63PBOSet(t *testing.T) {
	factories := CandidateFactories([]string{"VOO", "AAPL", "MSFT", "JNJ", "PG", "XLV"}, []string{"benchmark_ranker_proxy_h63"})
	if len(factories) != 1 {
		t.Fatalf("benchmark ranker proxy h63 factories=%d, want 1", len(factories))
	}
	variants := VariantFactories(factories[0], []string{"VOO", "AAPL", "MSFT", "JNJ", "PG", "XLV"})
	if len(variants) < 3 {
		t.Fatalf("benchmark ranker proxy h63 variants=%d, want at least 3 for PBO", len(variants))
	}
	for _, variant := range variants {
		if variant.Family != "benchmark_ranker_proxy_h63" {
			t.Fatalf("variant %s family=%s, want benchmark_ranker_proxy_h63", variant.Name, variant.Family)
		}
		if variant.Benchmark != "buy_hold" {
			t.Fatalf("variant %s benchmark=%s, want buy_hold", variant.Name, variant.Benchmark)
		}
	}
}

func TestCandidateFactoriesRegistersBenchmarkLGBMRanker(t *testing.T) {
	factories := CandidateFactories([]string{"VOO", "AAPL", "MSFT", "JNJ", "PG", "XLV"}, []string{"benchmark_lgbm_ranker"})
	if !hasFactory(factories, "benchmark_lgbm_ranker_h63_s15") {
		t.Fatalf("candidate factories missing benchmark LGBM ranker: %#v", factories)
	}
}

func TestVariantFactoriesProvidesBenchmarkLGBMRankerPBOSet(t *testing.T) {
	factories := CandidateFactories([]string{"VOO", "AAPL", "MSFT", "JNJ", "PG", "XLV"}, []string{"benchmark_lgbm_ranker"})
	if len(factories) != 1 {
		t.Fatalf("benchmark LGBM ranker factories=%d, want 1", len(factories))
	}
	variants := VariantFactories(factories[0], []string{"VOO", "AAPL", "MSFT", "JNJ", "PG", "XLV"})
	if len(variants) < 3 {
		t.Fatalf("benchmark LGBM ranker variants=%d, want at least 3 for PBO", len(variants))
	}
	for _, variant := range variants {
		if variant.Family != "benchmark_lgbm_ranker" {
			t.Fatalf("variant %s family=%s, want benchmark_lgbm_ranker", variant.Name, variant.Family)
		}
		if variant.Benchmark != "buy_hold" {
			t.Fatalf("variant %s benchmark=%s, want buy_hold", variant.Name, variant.Benchmark)
		}
	}
}

func TestCandidateFactoriesRegistersBenchmarkLGBMRankerH63(t *testing.T) {
	factories := CandidateFactories([]string{"VOO", "AAPL", "MSFT", "JNJ", "PG", "XLV"}, []string{"benchmark_lgbm_ranker_h63"})
	if !hasFactory(factories, "benchmark_lgbm_ranker_h63_s15_checkpoint") {
		t.Fatalf("candidate factories missing benchmark LGBM ranker h63: %#v", factories)
	}
}

func TestDailyRankerSleeveConfigUsesPITUniverseEnv(t *testing.T) {
	t.Setenv("OALPHA_DAILY_RANKER_PIT_UNIVERSE", "/tmp/pit_universe.json")
	cfg := dailyRankerSleeveConfig("VOO", 0.15, 3, 0.05, 63, 1.0, 0, 0, 1, 0, 1)
	if cfg.PointInTimeUniversePath != "/tmp/pit_universe.json" {
		t.Fatalf("PointInTimeUniversePath=%q, want env path", cfg.PointInTimeUniversePath)
	}
}

func TestDailyRankerArtifactRootUsesEnv(t *testing.T) {
	t.Setenv("OALPHA_DAILY_RANKER_ARTIFACT_ROOT", "/tmp/ranker_artifacts")
	if got := dailyRankerArtifactRoot(); got != "/tmp/ranker_artifacts" {
		t.Fatalf("dailyRankerArtifactRoot=%q, want env path", got)
	}
}

func TestVariantFactoriesProvidesBenchmarkLGBMRankerH63PBOSet(t *testing.T) {
	factories := CandidateFactories([]string{"VOO", "AAPL", "MSFT", "JNJ", "PG", "XLV"}, []string{"benchmark_lgbm_ranker_h63"})
	if len(factories) != 1 {
		t.Fatalf("benchmark LGBM ranker h63 factories=%d, want 1", len(factories))
	}
	variants := VariantFactories(factories[0], []string{"VOO", "AAPL", "MSFT", "JNJ", "PG", "XLV"})
	if len(variants) < 3 {
		t.Fatalf("benchmark LGBM ranker h63 variants=%d, want at least 3 for PBO", len(variants))
	}
	for _, variant := range variants {
		if variant.Family != "benchmark_lgbm_ranker_h63" {
			t.Fatalf("variant %s family=%s, want benchmark_lgbm_ranker_h63", variant.Name, variant.Family)
		}
		if variant.Benchmark != "buy_hold" {
			t.Fatalf("variant %s benchmark=%s, want buy_hold", variant.Name, variant.Benchmark)
		}
	}
}

func TestCandidateFactoriesRegistersBenchmarkLGBMRankerH63ExMegaCap(t *testing.T) {
	factories := CandidateFactories([]string{"VOO", "AAPL", "MSFT", "JNJ", "PG", "XLV"}, []string{"benchmark_lgbm_ranker_h63_exmegacap"})
	if !hasFactory(factories, "benchmark_lgbm_ranker_h63_s15_exmegacap") {
		t.Fatalf("candidate factories missing ex-megacap LGBM ranker h63: %#v", factories)
	}
}

func TestVariantFactoriesProvidesBenchmarkLGBMRankerH63ExMegaCapPBOSet(t *testing.T) {
	factories := CandidateFactories([]string{"VOO", "AAPL", "MSFT", "JNJ", "PG", "XLV"}, []string{"benchmark_lgbm_ranker_h63_exmegacap"})
	if len(factories) != 1 {
		t.Fatalf("ex-megacap LGBM ranker h63 factories=%d, want 1", len(factories))
	}
	variants := VariantFactories(factories[0], []string{"VOO", "AAPL", "MSFT", "JNJ", "PG", "XLV"})
	if len(variants) < 3 {
		t.Fatalf("ex-megacap LGBM ranker h63 variants=%d, want at least 3 for PBO", len(variants))
	}
	for _, variant := range variants {
		if variant.Family != "benchmark_lgbm_ranker_h63_exmegacap" {
			t.Fatalf("variant %s family=%s, want benchmark_lgbm_ranker_h63_exmegacap", variant.Name, variant.Family)
		}
		if variant.Benchmark != "buy_hold" {
			t.Fatalf("variant %s benchmark=%s, want buy_hold", variant.Name, variant.Benchmark)
		}
	}
}

func TestCandidateFactoriesRegistersBenchmarkLGBMRankerH63EqualBenchmark(t *testing.T) {
	factories := CandidateFactories([]string{"VOO", "AAPL", "MSFT", "JNJ", "PG", "XLV"}, []string{"benchmark_lgbm_ranker_h63_equal"})
	if !hasFactory(factories, "benchmark_lgbm_ranker_h63_s15_equal_benchmark") {
		t.Fatalf("candidate factories missing equal-weight benchmark LGBM ranker h63: %#v", factories)
	}
	if factories[0].Benchmark != "equal_weight" {
		t.Fatalf("benchmark=%s, want equal_weight", factories[0].Benchmark)
	}
}

func TestVariantFactoriesProvidesBenchmarkLGBMRankerH63EqualBenchmarkPBOSet(t *testing.T) {
	factories := CandidateFactories([]string{"VOO", "AAPL", "MSFT", "JNJ", "PG", "XLV"}, []string{"benchmark_lgbm_ranker_h63_equal"})
	if len(factories) != 1 {
		t.Fatalf("equal benchmark LGBM ranker h63 factories=%d, want 1", len(factories))
	}
	variants := VariantFactories(factories[0], []string{"VOO", "AAPL", "MSFT", "JNJ", "PG", "XLV"})
	if len(variants) < 3 {
		t.Fatalf("equal benchmark LGBM ranker h63 variants=%d, want at least 3 for PBO", len(variants))
	}
	for _, variant := range variants {
		if variant.Family != "benchmark_lgbm_ranker_h63_equal" {
			t.Fatalf("variant %s family=%s, want benchmark_lgbm_ranker_h63_equal", variant.Name, variant.Family)
		}
		if variant.Benchmark != "equal_weight" {
			t.Fatalf("variant %s benchmark=%s, want equal_weight", variant.Name, variant.Benchmark)
		}
	}
}

func TestCandidateFactoriesRegistersBenchmarkRankerProxyH63RiskCap(t *testing.T) {
	factories := CandidateFactories([]string{"VOO", "AAPL", "MSFT", "JNJ", "PG", "XLV"}, []string{"benchmark_ranker_proxy_h63_riskcap"})
	if !hasFactory(factories, "benchmark_ranker_proxy_h63_riskcap_checkpoint") {
		t.Fatalf("candidate factories missing benchmark ranker proxy h63 riskcap: %#v", factories)
	}
}

func TestVariantFactoriesProvidesBenchmarkRankerProxyH63RiskCapPBOSet(t *testing.T) {
	factories := CandidateFactories([]string{"VOO", "AAPL", "MSFT", "JNJ", "PG", "XLV"}, []string{"benchmark_ranker_proxy_h63_riskcap"})
	if len(factories) != 1 {
		t.Fatalf("benchmark ranker proxy h63 riskcap factories=%d, want 1", len(factories))
	}
	variants := VariantFactories(factories[0], []string{"VOO", "AAPL", "MSFT", "JNJ", "PG", "XLV"})
	if len(variants) < 3 {
		t.Fatalf("benchmark ranker proxy h63 riskcap variants=%d, want at least 3 for PBO", len(variants))
	}
	for _, variant := range variants {
		if variant.Family != "benchmark_ranker_proxy_h63_riskcap" {
			t.Fatalf("variant %s family=%s, want benchmark_ranker_proxy_h63_riskcap", variant.Name, variant.Family)
		}
		if variant.Benchmark != "buy_hold" {
			t.Fatalf("variant %s benchmark=%s, want buy_hold", variant.Name, variant.Benchmark)
		}
	}
}

func TestCandidateFactoriesRegistersBenchmarkRankerProxyH63TrendGuard(t *testing.T) {
	factories := CandidateFactories([]string{"VOO", "AAPL", "MSFT", "JNJ", "PG", "XLU", "XLP", "XLV"}, []string{"benchmark_ranker_proxy_h63_trendguard"})
	if !hasFactory(factories, "benchmark_ranker_proxy_h63_trendguard_checkpoint") {
		t.Fatalf("candidate factories missing benchmark ranker proxy h63 trendguard: %#v", factories)
	}
}

func TestVariantFactoriesProvidesBenchmarkRankerProxyH63TrendGuardPBOSet(t *testing.T) {
	factories := CandidateFactories([]string{"VOO", "AAPL", "MSFT", "JNJ", "PG", "XLU", "XLP", "XLV"}, []string{"benchmark_ranker_proxy_h63_trendguard"})
	if len(factories) != 1 {
		t.Fatalf("benchmark ranker proxy h63 trendguard factories=%d, want 1", len(factories))
	}
	variants := VariantFactories(factories[0], []string{"VOO", "AAPL", "MSFT", "JNJ", "PG", "XLU", "XLP", "XLV"})
	if len(variants) < 3 {
		t.Fatalf("benchmark ranker proxy h63 trendguard variants=%d, want at least 3 for PBO", len(variants))
	}
	for _, variant := range variants {
		if variant.Family != "benchmark_ranker_proxy_h63_trendguard" {
			t.Fatalf("variant %s family=%s, want benchmark_ranker_proxy_h63_trendguard", variant.Name, variant.Family)
		}
		if variant.Benchmark != "buy_hold" {
			t.Fatalf("variant %s benchmark=%s, want buy_hold", variant.Name, variant.Benchmark)
		}
	}
}

func TestCandidateFactoriesRegistersBenchmarkRankerProxyH63Liquidity(t *testing.T) {
	factories := CandidateFactories([]string{"VOO", "AAPL", "MSFT", "JNJ", "PG", "XLU", "XLP", "XLV"}, []string{"benchmark_ranker_proxy_h63_liquidity"})
	if !hasFactory(factories, "benchmark_ranker_proxy_h63_liquidity_checkpoint") {
		t.Fatalf("candidate factories missing benchmark ranker proxy h63 liquidity: %#v", factories)
	}
}

func TestVariantFactoriesProvidesBenchmarkRankerProxyH63LiquidityPBOSet(t *testing.T) {
	factories := CandidateFactories([]string{"VOO", "AAPL", "MSFT", "JNJ", "PG", "XLU", "XLP", "XLV"}, []string{"benchmark_ranker_proxy_h63_liquidity"})
	if len(factories) != 1 {
		t.Fatalf("benchmark ranker proxy h63 liquidity factories=%d, want 1", len(factories))
	}
	variants := VariantFactories(factories[0], []string{"VOO", "AAPL", "MSFT", "JNJ", "PG", "XLU", "XLP", "XLV"})
	if len(variants) < 3 {
		t.Fatalf("benchmark ranker proxy h63 liquidity variants=%d, want at least 3 for PBO", len(variants))
	}
	for _, variant := range variants {
		if variant.Family != "benchmark_ranker_proxy_h63_liquidity" {
			t.Fatalf("variant %s family=%s, want benchmark_ranker_proxy_h63_liquidity", variant.Name, variant.Family)
		}
		if variant.Benchmark != "buy_hold" {
			t.Fatalf("variant %s benchmark=%s, want buy_hold", variant.Name, variant.Benchmark)
		}
	}
}

func TestCandidateFactoriesRegistersBenchmarkRankerProxyH84(t *testing.T) {
	factories := CandidateFactories([]string{"VOO", "AAPL", "MSFT", "JNJ", "PG", "XLV"}, []string{"benchmark_ranker_proxy_h84"})
	if !hasFactory(factories, "benchmark_ranker_proxy_h84_checkpoint") {
		t.Fatalf("candidate factories missing benchmark ranker proxy h84: %#v", factories)
	}
}

func TestVariantFactoriesProvidesBenchmarkRankerProxyH84PBOSet(t *testing.T) {
	factories := CandidateFactories([]string{"VOO", "AAPL", "MSFT", "JNJ", "PG", "XLV"}, []string{"benchmark_ranker_proxy_h84"})
	if len(factories) != 1 {
		t.Fatalf("benchmark ranker proxy h84 factories=%d, want 1", len(factories))
	}
	variants := VariantFactories(factories[0], []string{"VOO", "AAPL", "MSFT", "JNJ", "PG", "XLV"})
	if len(variants) < 3 {
		t.Fatalf("benchmark ranker proxy h84 variants=%d, want at least 3 for PBO", len(variants))
	}
	for _, variant := range variants {
		if variant.Family != "benchmark_ranker_proxy_h84" {
			t.Fatalf("variant %s family=%s, want benchmark_ranker_proxy_h84", variant.Name, variant.Family)
		}
		if variant.Benchmark != "buy_hold" {
			t.Fatalf("variant %s benchmark=%s, want buy_hold", variant.Name, variant.Benchmark)
		}
	}
}

func TestCandidateFactoriesRegistersBenchmarkRankerProxyBlend(t *testing.T) {
	factories := CandidateFactories([]string{"VOO", "AAPL", "MSFT", "JNJ", "PG", "XLV"}, []string{"benchmark_ranker_proxy_blend"})
	if !hasFactory(factories, "benchmark_ranker_proxy_blend_checkpoint") {
		t.Fatalf("candidate factories missing benchmark ranker proxy blend: %#v", factories)
	}
}

func TestVariantFactoriesProvidesBenchmarkRankerProxyBlendPBOSet(t *testing.T) {
	factories := CandidateFactories([]string{"VOO", "AAPL", "MSFT", "JNJ", "PG", "XLV"}, []string{"benchmark_ranker_proxy_blend"})
	if len(factories) != 1 {
		t.Fatalf("benchmark ranker proxy blend factories=%d, want 1", len(factories))
	}
	variants := VariantFactories(factories[0], []string{"VOO", "AAPL", "MSFT", "JNJ", "PG", "XLV"})
	if len(variants) < 3 {
		t.Fatalf("benchmark ranker proxy blend variants=%d, want at least 3 for PBO", len(variants))
	}
	for _, variant := range variants {
		if variant.Family != "benchmark_ranker_proxy_blend" {
			t.Fatalf("variant %s family=%s, want benchmark_ranker_proxy_blend", variant.Name, variant.Family)
		}
		if variant.Benchmark != "buy_hold" {
			t.Fatalf("variant %s benchmark=%s, want buy_hold", variant.Name, variant.Benchmark)
		}
	}
}

func TestVariantFactoriesProvidesSectorRankedSleevePBOSet(t *testing.T) {
	factories := CandidateFactories([]string{"VOO", "QQQ", "XLK", "XLF", "XLV", "XLE"}, []string{"sector_ranked_sleeve"})
	if len(factories) != 1 {
		t.Fatalf("sector ranked sleeve factories=%d, want 1", len(factories))
	}
	variants := VariantFactories(factories[0], []string{"VOO", "QQQ", "XLK", "XLF", "XLV", "XLE"})
	if len(variants) < 3 {
		t.Fatalf("sector ranked sleeve variants=%d, want at least 3 for PBO", len(variants))
	}
	for _, variant := range variants {
		if variant.Family != "sector_ranked_sleeve" {
			t.Fatalf("variant %s family=%s, want sector_ranked_sleeve", variant.Name, variant.Family)
		}
		if variant.Benchmark != "buy_hold" {
			t.Fatalf("variant %s benchmark=%s, want buy_hold", variant.Name, variant.Benchmark)
		}
	}
}

func TestVariantFactoriesProvidesDefensiveRotationPBOSet(t *testing.T) {
	factories := CandidateFactories([]string{"VOO", "AAPL", "MSFT", "XLU", "XLP", "XLV"}, []string{"benchmark_rotation"})
	if len(factories) != 1 {
		t.Fatalf("benchmark rotation factories=%d, want 1", len(factories))
	}
	variants := VariantFactories(factories[0], []string{"VOO", "AAPL", "MSFT", "XLU", "XLP", "XLV"})
	if len(variants) < 3 {
		t.Fatalf("defensive rotation variants=%d, want at least 3 for PBO", len(variants))
	}
	for _, variant := range variants {
		if variant.Family != "composite_momentum_defensive" {
			t.Fatalf("variant %s family=%s, want composite_momentum_defensive", variant.Name, variant.Family)
		}
		if variant.Benchmark != "buy_hold" {
			t.Fatalf("variant %s benchmark=%s, want buy_hold", variant.Name, variant.Benchmark)
		}
	}
}

func catalogSymbolsForValidation() []string {
	return []string{
		"VOO", "AAPL", "MSFT", "NVDA", "AMAT", "INTC", "LRCX", "JNJ", "PG", "COST",
		"QQQ", "SPY", "IWM", "VTI", "XLU", "XLP", "XLV", "XLE", "XLF", "XLK",
	}
}

func hasFactory(factories []StrategyFactory, name string) bool {
	for _, factory := range factories {
		if factory.Name == name {
			return true
		}
	}
	return false
}
