package analyzer

import (
	"fmt"

	"blue-horizon/internal/attitude"
	"blue-horizon/internal/calibration"
	"blue-horizon/internal/config"
	"blue-horizon/internal/input"
)

type Severity string

const (
	Warning Severity = "WARNING"
	Danger  Severity = "DANGER"
)

type Event struct {
	TimeSec  float64
	Severity Severity
	Rule     string
	Message  string
	ValueDeg float64
}

func EvaluateSample(sample input.Sample, previous *input.Sample, cfg config.Config, cal calibration.Calibration) []Event {
	correctedRoll := sample.IMURollDeg - cal.RollOffsetDeg
	correctedPitch := sample.IMUPitchDeg - cal.PitchOffsetDeg

	var events []Event
	events = append(events, thresholdEvent(sample.TimeSec, "roll", "roll dangerous", "roll too high", correctedRoll, cfg.RollWarningDeg, cfg.RollDangerDeg)...)
	events = append(events, thresholdEvent(sample.TimeSec, "pitch", "pitch dangerous", "pitch too high", correctedPitch, cfg.PitchWarningDeg, cfg.PitchDangerDeg)...)

	rollMismatch := attitude.AbsDeg(correctedRoll - sample.GTSAMRollDeg)
	pitchMismatch := attitude.AbsDeg(correctedPitch - sample.GTSAMPitchDeg)
	events = append(events, mismatchEvent(sample.TimeSec, "estimator_roll_mismatch", "IMU/GTSAM roll mismatch", rollMismatch, cfg)...)
	events = append(events, mismatchEvent(sample.TimeSec, "estimator_pitch_mismatch", "IMU/GTSAM pitch mismatch", pitchMismatch, cfg)...)

	if previous != nil {
		dt := sample.TimeSec - previous.TimeSec
		rate := attitude.AbsDeg(attitude.YawRateDegS(previous.IMUYawDeg, sample.IMUYawDeg, dt))
		if dt > 0 && rate > cfg.YawRateWarningDegS {
			events = append(events, Event{
				TimeSec:  sample.TimeSec,
				Severity: Warning,
				Rule:     "yaw_rate",
				Message:  "abnormal yaw rate",
				ValueDeg: rate,
			})
		}
	}

	return events
}

func thresholdEvent(t float64, rule, dangerMessage, warningMessage string, value, warning, danger float64) []Event {
	abs := attitude.AbsDeg(value)
	if abs > danger {
		return []Event{{TimeSec: t, Severity: Danger, Rule: rule, Message: dangerMessage, ValueDeg: value}}
	}
	if abs > warning {
		return []Event{{TimeSec: t, Severity: Warning, Rule: rule, Message: warningMessage, ValueDeg: value}}
	}
	return nil
}

func mismatchEvent(t float64, rule, message string, value float64, cfg config.Config) []Event {
	if value > cfg.EstimatorMismatchDangerDeg {
		return []Event{{TimeSec: t, Severity: Danger, Rule: rule, Message: message, ValueDeg: value}}
	}
	if value > cfg.EstimatorMismatchWarningDeg {
		return []Event{{TimeSec: t, Severity: Warning, Rule: rule, Message: message, ValueDeg: value}}
	}
	return nil
}

func FormatEvent(event Event) string {
	return fmt.Sprintf("[%s] t=%.2f %s: %.1f deg", event.Severity, event.TimeSec, event.Message, event.ValueDeg)
}
