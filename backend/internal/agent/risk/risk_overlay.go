package risk

import (
	"math"
	"time"
)

type RegimeRiskRole string

const (
	RegimeRiskLowVol  RegimeRiskRole = "low_vol"
	RegimeRiskNormal  RegimeRiskRole = "normal"
	RegimeRiskHighVol RegimeRiskRole = "high_vol"
	RegimeRiskCrisis  RegimeRiskRole = "crisis"
	RegimeRiskUnknown RegimeRiskRole = "unknown"
)

type RiskOverlayPolicy struct {
	StateMultipliers           map[RegimeRiskRole]float64
	MinPosteriorConfidence     float64
	EnterHighVolProb           float64
	ExitHighVolProb            float64
	EnterCrisisProb            float64
	ExitCrisisProb             float64
	MinConsecutiveBarsToSwitch int
	ImmediateCrisisDeRisk      bool
	FailClosedOnUncertain      bool
	MaxRealizedAnnualVol       float64
	VolCapMultiplier           float64
	MaxDrawdownPct             float64
	DrawdownMultiplier         float64
	MultiplierSmoothing        float64
}

type RegimeOverlayInput struct {
	Timestamp         time.Time
	BaseExposure      float64
	PosteriorProbs    []float64
	StateRoles        []RegimeRiskRole
	ModelHealthy      bool
	RealizedAnnualVol float64
	PeakEquity        float64
	CurrentEquity     float64
}

type RegimeOverlayDecision struct {
	Timestamp        time.Time      `json:"timestamp"`
	BaseExposure     float64        `json:"base_exposure"`
	AdjustedExposure float64        `json:"adjusted_exposure"`
	Multiplier       float64        `json:"multiplier"`
	RawMultiplier    float64        `json:"raw_multiplier"`
	EffectiveRole    RegimeRiskRole `json:"effective_role"`
	Confidence       float64        `json:"confidence"`
	ModelHealthy     bool           `json:"model_healthy"`
	Vetoed           bool           `json:"vetoed"`
	Reasons          []string       `json:"reasons"`
}

type RegimeRiskOverlay struct {
	policy      RiskOverlayPolicy
	currentRole RegimeRiskRole
	pendingRole RegimeRiskRole
	pendingBars int
	lastMult    float64
}

func DefaultRiskOverlayPolicy() RiskOverlayPolicy {
	return RiskOverlayPolicy{
		StateMultipliers: map[RegimeRiskRole]float64{
			RegimeRiskLowVol:  1.00,
			RegimeRiskNormal:  1.00,
			RegimeRiskHighVol: 0.60,
			RegimeRiskCrisis:  0.25,
			RegimeRiskUnknown: 1.00,
		},
		MinPosteriorConfidence:     0.35,
		EnterHighVolProb:           0.55,
		ExitHighVolProb:            0.40,
		EnterCrisisProb:            0.45,
		ExitCrisisProb:             0.30,
		MinConsecutiveBarsToSwitch: 2,
		ImmediateCrisisDeRisk:      true,
		FailClosedOnUncertain:      false,
		VolCapMultiplier:           0.50,
		DrawdownMultiplier:         0.50,
		MultiplierSmoothing:        1.0,
	}
}

func NewRegimeRiskOverlay(policy RiskOverlayPolicy) *RegimeRiskOverlay {
	policy = policy.withDefaults()
	return &RegimeRiskOverlay{
		policy:      policy,
		currentRole: RegimeRiskNormal,
		pendingRole: RegimeRiskUnknown,
		lastMult:    1.0,
	}
}

func (o *RegimeRiskOverlay) Reset() {
	o.currentRole = RegimeRiskNormal
	o.pendingRole = RegimeRiskUnknown
	o.pendingBars = 0
	o.lastMult = 1.0
}

func (o *RegimeRiskOverlay) Policy() RiskOverlayPolicy {
	return o.policy
}

func (o *RegimeRiskOverlay) Apply(input RegimeOverlayInput) RegimeOverlayDecision {
	if o == nil {
		policy := DefaultRiskOverlayPolicy()
		o = NewRegimeRiskOverlay(policy)
	}

	decision := RegimeOverlayDecision{
		Timestamp:    input.Timestamp,
		BaseExposure: input.BaseExposure,
		ModelHealthy: input.ModelHealthy,
		Reasons:      make([]string, 0, 4),
	}

	role, confidence, healthy := o.classify(input)
	decision.Confidence = confidence
	if !input.ModelHealthy || !healthy {
		decision.EffectiveRole = RegimeRiskUnknown
		decision.RawMultiplier = o.uncertainMultiplier()
		decision.Multiplier = decision.RawMultiplier
		decision.AdjustedExposure = input.BaseExposure * decision.Multiplier
		decision.Vetoed = input.BaseExposure != 0 && decision.AdjustedExposure == 0
		decision.Reasons = append(decision.Reasons, "model_unhealthy_or_uncertain")
		if o.policy.FailClosedOnUncertain {
			decision.Reasons = append(decision.Reasons, "fail_closed")
		} else {
			decision.Reasons = append(decision.Reasons, "fail_open")
		}
		o.lastMult = decision.Multiplier
		return decision
	}

	role = o.applyHysteresis(role, input)
	decision.EffectiveRole = role
	multiplier := o.multiplierFor(role)
	decision.Reasons = append(decision.Reasons, "role_"+string(role))

	if o.policy.MaxRealizedAnnualVol > 0 && input.RealizedAnnualVol > o.policy.MaxRealizedAnnualVol {
		multiplier = math.Min(multiplier, clampMultiplier(o.policy.VolCapMultiplier, 0.5))
		decision.Reasons = append(decision.Reasons, "realized_vol_cap")
	}

	if o.policy.MaxDrawdownPct > 0 && input.PeakEquity > 0 && input.CurrentEquity > 0 {
		drawdown := (input.PeakEquity - input.CurrentEquity) / input.PeakEquity
		if drawdown > o.policy.MaxDrawdownPct {
			multiplier = math.Min(multiplier, clampMultiplier(o.policy.DrawdownMultiplier, 0.5))
			decision.Reasons = append(decision.Reasons, "drawdown_guard")
		}
	}

	rawMultiplier := clampMultiplier(multiplier, 1.0)
	smoothedMultiplier := o.smoothMultiplier(rawMultiplier, role)
	decision.RawMultiplier = rawMultiplier
	decision.Multiplier = smoothedMultiplier
	decision.AdjustedExposure = input.BaseExposure * smoothedMultiplier
	decision.Vetoed = input.BaseExposure != 0 && decision.AdjustedExposure == 0
	o.lastMult = smoothedMultiplier
	return decision
}

func (o *RegimeRiskOverlay) classify(input RegimeOverlayInput) (RegimeRiskRole, float64, bool) {
	if !validPosterior(input.PosteriorProbs, input.StateRoles) {
		return RegimeRiskUnknown, 0, false
	}

	roleProbs := make(map[RegimeRiskRole]float64)
	bestRole := RegimeRiskUnknown
	var bestProb float64
	for i, prob := range input.PosteriorProbs {
		role := input.StateRoles[i]
		roleProbs[role] += prob
		if roleProbs[role] > bestProb {
			bestProb = roleProbs[role]
			bestRole = role
		}
	}

	if roleProbs[RegimeRiskCrisis] >= o.policy.EnterCrisisProb {
		return RegimeRiskCrisis, roleProbs[RegimeRiskCrisis], true
	}
	if o.currentRole == RegimeRiskCrisis && roleProbs[RegimeRiskCrisis] >= o.policy.ExitCrisisProb {
		return RegimeRiskCrisis, roleProbs[RegimeRiskCrisis], true
	}
	if roleProbs[RegimeRiskHighVol] >= o.policy.EnterHighVolProb {
		return RegimeRiskHighVol, roleProbs[RegimeRiskHighVol], true
	}
	if o.currentRole == RegimeRiskHighVol && roleProbs[RegimeRiskHighVol] >= o.policy.ExitHighVolProb {
		return RegimeRiskHighVol, roleProbs[RegimeRiskHighVol], true
	}

	if bestProb < o.policy.MinPosteriorConfidence {
		return RegimeRiskUnknown, bestProb, true
	}
	return bestRole, bestProb, true
}

func (o *RegimeRiskOverlay) applyHysteresis(candidate RegimeRiskRole, input RegimeOverlayInput) RegimeRiskRole {
	if candidate == RegimeRiskUnknown {
		return o.currentRole
	}
	if candidate == o.currentRole {
		o.pendingRole = RegimeRiskUnknown
		o.pendingBars = 0
		return o.currentRole
	}
	if candidate == RegimeRiskCrisis && o.policy.ImmediateCrisisDeRisk {
		o.currentRole = RegimeRiskCrisis
		o.pendingRole = RegimeRiskUnknown
		o.pendingBars = 0
		return o.currentRole
	}
	required := o.policy.MinConsecutiveBarsToSwitch
	if required <= 1 {
		o.currentRole = candidate
		return o.currentRole
	}
	if candidate != o.pendingRole {
		o.pendingRole = candidate
		o.pendingBars = 1
		return o.currentRole
	}
	o.pendingBars++
	if o.pendingBars >= required {
		o.currentRole = candidate
		o.pendingRole = RegimeRiskUnknown
		o.pendingBars = 0
	}
	return o.currentRole
}

func (o *RegimeRiskOverlay) multiplierFor(role RegimeRiskRole) float64 {
	if o.policy.StateMultipliers == nil {
		return 1.0
	}
	if multiplier, ok := o.policy.StateMultipliers[role]; ok {
		return clampMultiplier(multiplier, 1.0)
	}
	return clampMultiplier(o.policy.StateMultipliers[RegimeRiskUnknown], 1.0)
}

func (o *RegimeRiskOverlay) uncertainMultiplier() float64 {
	if o.policy.FailClosedOnUncertain {
		return 0
	}
	return o.multiplierFor(RegimeRiskUnknown)
}

func (o *RegimeRiskOverlay) smoothMultiplier(target float64, role RegimeRiskRole) float64 {
	alpha := o.policy.MultiplierSmoothing
	if alpha <= 0 || alpha >= 1 || (role == RegimeRiskCrisis && o.policy.ImmediateCrisisDeRisk) {
		return target
	}
	return o.lastMult + alpha*(target-o.lastMult)
}

func (p RiskOverlayPolicy) withDefaults() RiskOverlayPolicy {
	defaults := DefaultRiskOverlayPolicy()
	if p.StateMultipliers == nil {
		p.StateMultipliers = defaults.StateMultipliers
	} else {
		for role, multiplier := range defaults.StateMultipliers {
			if _, ok := p.StateMultipliers[role]; !ok {
				p.StateMultipliers[role] = multiplier
			}
		}
	}
	if p.MinPosteriorConfidence <= 0 {
		p.MinPosteriorConfidence = defaults.MinPosteriorConfidence
	}
	if p.EnterHighVolProb <= 0 {
		p.EnterHighVolProb = defaults.EnterHighVolProb
	}
	if p.ExitHighVolProb <= 0 {
		p.ExitHighVolProb = defaults.ExitHighVolProb
	}
	if p.EnterCrisisProb <= 0 {
		p.EnterCrisisProb = defaults.EnterCrisisProb
	}
	if p.ExitCrisisProb <= 0 {
		p.ExitCrisisProb = defaults.ExitCrisisProb
	}
	if p.MinConsecutiveBarsToSwitch <= 0 {
		p.MinConsecutiveBarsToSwitch = defaults.MinConsecutiveBarsToSwitch
	}
	if p.VolCapMultiplier <= 0 {
		p.VolCapMultiplier = defaults.VolCapMultiplier
	}
	if p.DrawdownMultiplier <= 0 {
		p.DrawdownMultiplier = defaults.DrawdownMultiplier
	}
	if p.MultiplierSmoothing <= 0 {
		p.MultiplierSmoothing = defaults.MultiplierSmoothing
	}
	return p
}

func validPosterior(probs []float64, roles []RegimeRiskRole) bool {
	if len(probs) == 0 || len(probs) != len(roles) {
		return false
	}
	var sum float64
	for _, prob := range probs {
		if math.IsNaN(prob) || math.IsInf(prob, 0) || prob < 0 {
			return false
		}
		sum += prob
	}
	return sum > 0
}

func clampMultiplier(value, fallback float64) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
		return fallback
	}
	if value > 1 {
		return 1
	}
	return value
}
