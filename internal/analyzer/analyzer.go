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
		rollMismatch := attitude.AbsDeg(correctedRoll - sample.GTSAMRollDeg)
		pitchMismatch := attitude.AbsDeg(correctedPitch - sample.GTSAMPitchDeg)

		result.MaxRollDeg = maxFloat(result.MaxRollDeg, rollAbs)
		result.MaxPitchDeg = maxFloat(result.MaxPitchDeg, pitchAbs)
		result.MaxMismatchDeg = maxFloat(result.MaxMismatchDeg, maxFloat(rollMismatch, pitchMismatch))

		events := EvaluateSample(sample, previous, cfg, cal)
		for _, event := range events {
			result.Events = append(result.Events, event)
			if result.WorstEvent == nil || attitude.AbsDeg(event.ValueDeg) > attitude.AbsDeg(result.WorstEvent.ValueDeg) {
				eventCopy := event
				result.WorstEvent = &eventCopy
			}
		}
		previous = &samples[i]
	}

	return result
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
