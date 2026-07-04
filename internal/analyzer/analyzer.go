package analyzer

import (
	"blue-horizon/internal/attitude"
	"blue-horizon/internal/calibration"
	"blue-horizon/internal/config"
	"blue-horizon/internal/input"
)

type Result struct {
	SamplesProcessed int
	DurationSec      float64
	MaxRollDeg       float64
	MaxPitchDeg      float64
	MaxMismatchDeg   float64
	WorstEvent       *Event
	Events           []Event
}

func Analyze(samples []input.Sample, cfg config.Config, cal calibration.Calibration) Result {
	result := Result{SamplesProcessed: len(samples)}
	if len(samples) == 0 {
		return result
	}

	result.DurationSec = samples[len(samples)-1].TimeSec - samples[0].TimeSec

	var previous *input.Sample
	for i := range samples {
		sample := samples[i]
		correctedRoll := sample.IMURollDeg - cal.RollOffsetDeg
		correctedPitch := sample.IMUPitchDeg - cal.PitchOffsetDeg
		rollAbs := attitude.AbsDeg(correctedRoll)
		pitchAbs := attitude.AbsDeg(correctedPitch)
		rollMismatch := attitude.AbsDeg(attitude.DeltaDegrees(correctedRoll, sample.GTSAMRollDeg))
		pitchMismatch := attitude.AbsDeg(attitude.DeltaDegrees(correctedPitch, sample.GTSAMPitchDeg))

		result.MaxRollDeg = maxFloat(result.MaxRollDeg, rollAbs)
		result.MaxPitchDeg = maxFloat(result.MaxPitchDeg, pitchAbs)
		result.MaxMismatchDeg = maxFloat(result.MaxMismatchDeg, maxFloat(rollMismatch, pitchMismatch))

		events := EvaluateSample(sample, previous, cfg, cal)
		for _, event := range events {
			result.Events = append(result.Events, event)
			if worseThan(event, result.WorstEvent) {
				eventCopy := event
				result.WorstEvent = &eventCopy
			}
		}
		previous = &samples[i]
	}

	return result
}

// worseThan reports whether candidate should replace the current worst event.
// Values are only comparable within the same unit; degree-valued events outrank
// rate events so a large deg/s number cannot bury a dangerous attitude.
func worseThan(candidate Event, incumbent *Event) bool {
	if incumbent == nil {
		return true
	}
	if Unit(candidate.Rule) != Unit(incumbent.Rule) {
		return Unit(candidate.Rule) == "deg"
	}
	return attitude.AbsDeg(candidate.ValueDeg) > attitude.AbsDeg(incumbent.ValueDeg)
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
