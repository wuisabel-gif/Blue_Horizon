package analyzer

import (
	"strings"
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

func TestMismatchWrapsAt180(t *testing.T) {
	// IMU 179 vs estimator -179.5 is a true disagreement of 1.5 deg, not 358.5.
	samples := []input.Sample{
		{TimeSec: 0, IMURollDeg: 179, GTSAMRollDeg: -179.5, IMUPitchDeg: 0, GTSAMPitchDeg: 0},
	}

	result := Analyze(samples, config.Default(), calibration.Calibration{})

	for _, e := range result.Events {
		if e.Rule == "estimator_roll_mismatch" {
			t.Fatalf("wrap-around produced a false mismatch event: %#v", e)
		}
	}
	if result.MaxMismatchDeg > 2 {
		t.Fatalf("expected max mismatch ~1.5, got %v", result.MaxMismatchDeg)
	}
}

func TestTrueFrameFlipStillDetected(t *testing.T) {
	// A genuine ~180 deg flip must still fire, wrap handling must not hide it.
	samples := []input.Sample{
		{TimeSec: 0, IMURollDeg: 0, GTSAMRollDeg: 179, IMUPitchDeg: 0, GTSAMPitchDeg: 0},
	}

	result := Analyze(samples, config.Default(), calibration.Calibration{})

	if result.MaxMismatchDeg < 150 {
		t.Fatalf("expected near-180 mismatch, got %v", result.MaxMismatchDeg)
	}
}

func TestWorstEventPrefersDegreesOverRate(t *testing.T) {
	// Yaw jump gives a huge deg/s value; the dangerous roll must still win.
	samples := []input.Sample{
		{TimeSec: 0, IMURollDeg: 0, IMUYawDeg: 0},
		{TimeSec: 0.01, IMURollDeg: 26.5, GTSAMRollDeg: 26.5, IMUYawDeg: 31.4},
	}

	result := Analyze(samples, config.Default(), calibration.Calibration{})

	if result.WorstEvent == nil || result.WorstEvent.Rule != "roll" {
		t.Fatalf("expected worst event rule=roll, got %#v", result.WorstEvent)
	}
}

func TestFormatEventYawRateUnit(t *testing.T) {
	line := FormatEvent(Event{Severity: Warning, Rule: "yaw_rate", Message: "abnormal yaw rate", ValueDeg: 3140})
	if want := "3140.0 deg/s"; !strings.Contains(line, want) {
		t.Fatalf("expected %q in %q", want, line)
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
