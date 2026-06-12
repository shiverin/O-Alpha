package portfolio

import (
	"math"
	"sort"
	"strings"
	"time"

	"github.com/oalpha/internal/backtest"
	"github.com/oalpha/internal/db"
)

const (
	defaultRuntimeRiskProfile    = "moderate"
	defaultRuntimeMaxGross       = 1.0
	defaultRuntimeMaxPositions   = 5
	defaultRuntimeStopLossPct    = 2.0
	defaultRuntimeTakeProfitPct  = 4.0
	defaultRuntimeRebalanceFreq  = "daily"
	runtimeSettingsMetadataKey   = "settings_applied"
	runtimeCadenceMetadataKey    = "rebalance_due"
	runtimeSuppressedMetadataKey = "rebalance_suppressed_reason"
)

type RuntimeSettings struct {
	RiskProfile           string    `json:"risk_profile"`
	MaxGrossExposure      float64   `json:"max_gross_exposure"`
	MaxActivePositions    int       `json:"max_active_positions"`
	StopLossPct           float64   `json:"stop_loss_pct"`
	TakeProfitPct         float64   `json:"take_profit_pct"`
	RebalanceFreq         string    `json:"rebalance_freq"`
	LoadedAt              time.Time `json:"loaded_at"`
	Source                string    `json:"source"`
	LastRebalanceAt       time.Time `json:"last_rebalance_at,omitempty"`
	NextEligibleRebalance time.Time `json:"next_eligible_rebalance,omitempty"`
}

func DefaultRuntimeSettings(now time.Time) RuntimeSettings {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	return RuntimeSettings{
		RiskProfile:        defaultRuntimeRiskProfile,
		MaxGrossExposure:   defaultRuntimeMaxGross,
		MaxActivePositions: defaultRuntimeMaxPositions,
		StopLossPct:        defaultRuntimeStopLossPct,
		TakeProfitPct:      defaultRuntimeTakeProfitPct,
		RebalanceFreq:      defaultRuntimeRebalanceFreq,
		LoadedAt:           now.UTC(),
		Source:             "default",
	}
}

func RuntimeSettingsFromAgentSettings(settings *db.AgentSettings, now time.Time) RuntimeSettings {
	out := DefaultRuntimeSettings(now)
	if settings == nil {
		return out
	}
	out.Source = "agent_settings"
	if strings.TrimSpace(settings.RiskProfile) != "" {
		out.RiskProfile = strings.ToLower(strings.TrimSpace(settings.RiskProfile))
	}
	if settings.Leverage > 0 {
		out.MaxGrossExposure = float64(settings.Leverage)
	}
	if settings.MaxPositions > 0 {
		out.MaxActivePositions = settings.MaxPositions
	}
	if settings.StopLossPct > 0 {
		out.StopLossPct = settings.StopLossPct
	}
	if settings.TakeProfitPct > 0 {
		out.TakeProfitPct = settings.TakeProfitPct
	}
	if strings.TrimSpace(settings.RebalanceFreq) != "" {
		out.RebalanceFreq = strings.ToLower(strings.TrimSpace(settings.RebalanceFreq))
	}
	return out.withBounds()
}

func (s RuntimeSettings) withBounds() RuntimeSettings {
	switch s.RiskProfile {
	case "conservative", "moderate", "aggressive":
	default:
		s.RiskProfile = defaultRuntimeRiskProfile
	}
	if s.MaxGrossExposure <= 0 {
		s.MaxGrossExposure = defaultRuntimeMaxGross
	}
	if s.MaxGrossExposure > 10 {
		s.MaxGrossExposure = 10
	}
	if s.MaxActivePositions <= 0 {
		s.MaxActivePositions = defaultRuntimeMaxPositions
	}
	if s.MaxActivePositions > 100 {
		s.MaxActivePositions = 100
	}
	if s.StopLossPct <= 0 {
		s.StopLossPct = defaultRuntimeStopLossPct
	}
	if s.TakeProfitPct <= 0 {
		s.TakeProfitPct = defaultRuntimeTakeProfitPct
	}
	switch s.RebalanceFreq {
	case "hourly", "daily", "weekly", "monthly":
	default:
		s.RebalanceFreq = defaultRuntimeRebalanceFreq
	}
	return s
}

func (s RuntimeSettings) RebalanceInterval() time.Duration {
	switch s.RebalanceFreq {
	case "hourly":
		return time.Hour
	case "weekly":
		return 7 * 24 * time.Hour
	case "monthly":
		return 30 * 24 * time.Hour
	default:
		return 24 * time.Hour
	}
}

func (s RuntimeSettings) RebalanceDue(now time.Time, lastRebalance time.Time) bool {
	if lastRebalance.IsZero() {
		return true
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	return !now.Before(lastRebalance.Add(s.RebalanceInterval()))
}

func (s RuntimeSettings) ToRuntimeState() map[string]interface{} {
	state := map[string]interface{}{
		"source":                       s.Source,
		"risk_profile":                 s.RiskProfile,
		"effective_max_gross_exposure": s.MaxGrossExposure,
		"effective_max_positions":      s.MaxActivePositions,
		"effective_stop_loss_pct":      s.StopLossPct,
		"effective_take_profit_pct":    s.TakeProfitPct,
		"effective_rebalance_freq":     s.RebalanceFreq,
		"last_settings_loaded_at":      s.LoadedAt,
	}
	if !s.LastRebalanceAt.IsZero() {
		state["last_rebalance_at"] = s.LastRebalanceAt
	}
	if !s.NextEligibleRebalance.IsZero() {
		state["next_eligible_rebalance_at"] = s.NextEligibleRebalance
	}
	return state
}

func applyRuntimeSettingsToOutput(output backtest.PortfolioOutput, benchmarkSymbol string, settings RuntimeSettings, rebalanceDue bool, lastRebalance time.Time) backtest.PortfolioOutput {
	output.Targets = cloneTargets(output.Targets)
	if output.EngineMetadata == nil {
		output.EngineMetadata = make(map[string]interface{})
	}

	settings.LastRebalanceAt = lastRebalance
	if !lastRebalance.IsZero() {
		settings.NextEligibleRebalance = lastRebalance.Add(settings.RebalanceInterval())
	}
	output.EngineMetadata[runtimeSettingsMetadataKey] = settings.ToRuntimeState()
	output.EngineMetadata[runtimeCadenceMetadataKey] = rebalanceDue

	if !rebalanceDue {
		output.Targets = nil
		output.GrossExposure = 0
		output.NetExposure = 0
		output.CashWeight = 1
		output.EngineMetadata[runtimeSuppressedMetadataKey] = "rebalance cadence not due"
		return output
	}

	capActivePositions(output.Targets, benchmarkSymbol, settings.MaxActivePositions)
	scaleTargetsToMaxGross(output.Targets, settings.MaxGrossExposure)
	recomputeOutputExposure(&output)
	return output
}

func cloneTargets(targets map[string]backtest.TargetPosition) map[string]backtest.TargetPosition {
	if len(targets) == 0 {
		return nil
	}
	out := make(map[string]backtest.TargetPosition, len(targets))
	for symbol, target := range targets {
		out[symbol] = target
	}
	return out
}

func capActivePositions(targets map[string]backtest.TargetPosition, benchmarkSymbol string, maxActive int) {
	if maxActive <= 0 || len(targets) == 0 {
		return
	}
	benchmarkSymbol = strings.ToUpper(strings.TrimSpace(benchmarkSymbol))
	type candidate struct {
		symbol string
		target backtest.TargetPosition
	}
	active := make([]candidate, 0, len(targets))
	for symbol, target := range targets {
		normalized := strings.ToUpper(strings.TrimSpace(symbol))
		if normalized == benchmarkSymbol {
			continue
		}
		if target.Side == backtest.PositionSideShort || target.TargetWeight <= 0 {
			delete(targets, symbol)
			continue
		}
		active = append(active, candidate{symbol: symbol, target: target})
	}
	if len(active) <= maxActive {
		return
	}
	sort.Slice(active, func(i, j int) bool {
		left := active[i].target
		right := active[j].target
		if left.TargetWeight != right.TargetWeight {
			return left.TargetWeight > right.TargetWeight
		}
		if left.AlphaScore != right.AlphaScore {
			return left.AlphaScore > right.AlphaScore
		}
		return active[i].symbol < active[j].symbol
	})
	for _, drop := range active[maxActive:] {
		delete(targets, drop.symbol)
	}
}

func scaleTargetsToMaxGross(targets map[string]backtest.TargetPosition, maxGross float64) {
	if maxGross <= 0 || len(targets) == 0 {
		return
	}
	var gross float64
	for _, target := range targets {
		gross += math.Abs(target.TargetWeight)
	}
	if gross <= maxGross || gross <= 0 {
		return
	}
	scale := maxGross / gross
	for symbol, target := range targets {
		target.TargetWeight *= scale
		targets[symbol] = target
	}
}

func recomputeOutputExposure(output *backtest.PortfolioOutput) {
	if output == nil {
		return
	}
	var gross, net float64
	for _, target := range output.Targets {
		gross += math.Abs(target.TargetWeight)
		net += target.TargetWeight
	}
	output.GrossExposure = gross
	output.NetExposure = net
	output.CashWeight = 1 - gross
	if output.CashWeight < 0 {
		output.CashWeight = 0
	}
}
