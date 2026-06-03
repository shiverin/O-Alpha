package api

import (
	"testing"

	"github.com/oalpha/internal/alpha/cointegration"
	"github.com/oalpha/internal/alpha/momentum"
	"github.com/oalpha/pkg/models"
)

func TestIsPortfolioBacktestStrategy(t *testing.T) {
	if !isPortfolioBacktestStrategy("xsec_momentum") {
		t.Fatalf("xsec_momentum should route as portfolio strategy")
	}
	if !isPortfolioBacktestStrategy("KALMAN_COINTEGRATION") {
		t.Fatalf("KALMAN_COINTEGRATION should route as portfolio strategy")
	}
	if isPortfolioBacktestStrategy("MA_CROSSOVER") {
		t.Fatalf("MA_CROSSOVER should remain single-symbol")
	}
	if isPortfolioBacktestStrategy("ML_META_LABEL") {
		t.Fatalf("ML_META_LABEL should remain single-symbol wrapper")
	}
}

func TestNormalizeRequestSymbolsPreservesPairOrder(t *testing.T) {
	symbols := normalizeRequestSymbols(models.BacktestRequest{
		Symbols: []string{"Y", "X", "Y"},
		Symbol:  "Z",
	})
	want := []string{"Y", "X", "Z"}
	if len(symbols) != len(want) {
		t.Fatalf("symbols=%v, want %v", symbols, want)
	}
	for i := range want {
		if symbols[i] != want[i] {
			t.Fatalf("symbols=%v, want %v", symbols, want)
		}
	}
}

func TestBuildPortfolioStrategyMomentum(t *testing.T) {
	strategy, allowShorts, err := buildPortfolioStrategy(models.BacktestRequest{
		StrategyType: "XSEC_MOMENTUM",
		Parameters: map[string]interface{}{
			"formation_days":    120.0,
			"top_fraction":      0.25,
			"target_vol_annual": 0.10,
		},
	}, []string{"AAPL", "MSFT", "NVDA"})
	if err != nil {
		t.Fatalf("buildPortfolioStrategy: %v", err)
	}
	if allowShorts {
		t.Fatalf("momentum should be long-only")
	}
	if _, ok := strategy.(*momentum.CrossSectionalMomentumStrategy); !ok {
		t.Fatalf("strategy type=%T, want momentum strategy", strategy)
	}
}

func TestBuildPortfolioStrategyKalmanPair(t *testing.T) {
	strategy, allowShorts, err := buildPortfolioStrategy(models.BacktestRequest{
		StrategyType: "KALMAN_COINTEGRATION",
		Parameters: map[string]interface{}{
			"symbol_y": "KRE",
			"symbol_x": "XLF",
			"entry_z":  1.5,
		},
	}, []string{"VOO", "SPY"})
	if err != nil {
		t.Fatalf("buildPortfolioStrategy: %v", err)
	}
	if !allowShorts {
		t.Fatalf("kalman pair should allow shorts in backtest")
	}
	if _, ok := strategy.(*cointegration.KalmanPairStrategy); !ok {
		t.Fatalf("strategy type=%T, want kalman pair strategy", strategy)
	}
	if got := strategy.Universe(); got[0] != "KRE" || got[1] != "XLF" {
		t.Fatalf("universe=%v, want KRE/XLF", got)
	}
}

func TestBuildBaseSingleSymbolStrategyForMLWrapper(t *testing.T) {
	strategy, err := buildBaseSingleSymbolStrategy("KALMAN", models.BacktestRequest{
		QNoise:     0.01,
		RNoise:     0.5,
		ZThreshold: 2,
	})
	if err != nil {
		t.Fatalf("buildBaseSingleSymbolStrategy: %v", err)
	}
	if strategy == nil {
		t.Fatalf("expected strategy")
	}
}

func TestLoadMLPredictorRequiresArtifact(t *testing.T) {
	_, _, _, _, err := loadMLPredictorFromParams(nil)
	if err == nil {
		t.Fatalf("expected missing artifact error")
	}
}
