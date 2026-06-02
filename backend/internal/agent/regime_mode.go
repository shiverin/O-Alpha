package agent

type RegimeMode string

const (
	RegimeModeNone    RegimeMode = "none"
	RegimeModeOverlay RegimeMode = "overlay"
)

func (m RegimeMode) String() string {
	if m == "" {
		return string(RegimeModeOverlay)
	}
	return string(m)
}
