package analyzer

import (
	"testing"

	"blue-horizon/internal/calibration"
	"blue-horizon/internal/config"
	"blue-horizon/internal/input"
)

func TestAnalyzeDetectsRollAndMismatch(t *testing.T) {
	samples := []input.Sample{
		{TimeSec: 0, IMURollDeg: 1, IMUPitchDeg: 0, IMUYawDeg: 10, GTSAMRollDeg: 1, GTSAMPitchDeg: 0},
		{TimeSec: 1, IMURollDeg: 18.5, IMUPitchDeg: 0, IMUYawDeg: 10, GTSAMRollDeg: 3, GTSAMPitchDeg: 0},
	}

	result := Analyze(samples, config.Default(), calibration.Calibration{})

	if len(result.Events) != 2 {
		t.Fatalf("expected 2 events, got %d: %#v", len(result.Events), result.Events)
	}
	if result.MaxRollDeg != 18.5 {
		t.Fatalf("expected max roll 18.5, got %v", result.MaxRollDeg)
	}
	if result.MaxMismatchDeg != 15.5 {
		t.Fatalf("expected max mismatch 15.5, got %v", result.MaxMismatchDeg)
	}
}

func TestCalibrationOffsetIsApplied(t *testing.T) {
	samples := []input.Sample{
		{TimeSec: 0, IMURollDeg: 14, IMUPitchDeg: 0, IMUYawDeg: 10, GTSAMRollDeg: 4, GTSAMPitchDeg: 0},
	}

	result := Analyze(samples, config.Default(), calibration.Calibration{RollOffsetDeg: 10})

	if len(result.Events) != 0 {
		t.Fatalf("expected calibration to suppress events, got %#v", result.Events)
	}
}
