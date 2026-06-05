package alphavalidation

import "testing"

func TestBuildWalkForwardWindows(t *testing.T) {
	windows, err := BuildWalkForwardWindows(20, 6, 4, 2)
	if err != nil {
		t.Fatalf("build windows: %v", err)
	}
	if len(windows) != 6 {
		t.Fatalf("expected 6 windows, got %d", len(windows))
	}
	if windows[0].TrainStart != 0 || windows[0].TrainEnd != 6 || windows[0].TestStart != 6 || windows[0].TestEnd != 10 {
		t.Fatalf("unexpected first window: %+v", windows[0])
	}
}
