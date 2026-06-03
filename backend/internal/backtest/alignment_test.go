package backtest

import (
	"testing"
	"time"

	"github.com/oalpha/pkg/models"
)

func TestAlignBarsInnerJoinDropsIncompleteTimestamps(t *testing.T) {
	t0 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	t1 := t0.Add(time.Hour)
	t2 := t1.Add(time.Hour)
	panel, err := AlignBars(map[string][]models.Bar{
		"A": {{Time: t0, Symbol: "A", Close: 1}, {Time: t1, Symbol: "A", Close: 2}, {Time: t2, Symbol: "A", Close: 3}},
		"B": {{Time: t1, Symbol: "B", Close: 4}, {Time: t2, Symbol: "B", Close: 5}},
	}, AlignmentConfig{Mode: AlignInnerJoin})
	if err != nil {
		t.Fatalf("align: %v", err)
	}
	if len(panel.Times) != 2 || !panel.Times[0].Equal(t1) || !panel.Times[1].Equal(t2) {
		t.Fatalf("unexpected aligned times: %v", panel.Times)
	}
}

func TestAlignBarsForwardFillMarksStaleBars(t *testing.T) {
	t0 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	t1 := t0.Add(time.Hour)
	panel, err := AlignBars(map[string][]models.Bar{
		"A": {{Time: t0, Symbol: "A", Close: 1}, {Time: t1, Symbol: "A", Close: 2}},
		"B": {{Time: t0, Symbol: "B", Close: 3}},
	}, AlignmentConfig{Mode: AlignForwardFill, MaxStaleBars: 1})
	if err != nil {
		t.Fatalf("align: %v", err)
	}
	if len(panel.Bars["B"]) != 2 || !panel.Bars["B"][1].Time.Equal(t1) || panel.Bars["B"][1].Close != 3 {
		t.Fatalf("expected B to be forward-filled to t1, got %+v", panel.Bars["B"])
	}
}
