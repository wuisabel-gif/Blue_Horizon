package calibration

import (
	"fmt"
	"os"
	"strconv"

	"blue-horizon/internal/config"
	"blue-horizon/internal/input"
)

type Calibration struct {
	RollOffsetDeg  float64
	PitchOffsetDeg float64
}

func Estimate(samples []input.Sample) (Calibration, error) {
	if len(samples) == 0 {
		return Calibration{}, fmt.Errorf("cannot calibrate empty sample set")
	}

	var rollSum, pitchSum float64
	for _, sample := range samples {
		rollSum += sample.IMURollDeg
		pitchSum += sample.IMUPitchDeg
	}

	n := float64(len(samples))
	return Calibration{
		RollOffsetDeg:  rollSum / n,
		PitchOffsetDeg: pitchSum / n,
	}, nil
}

func Load(path string) (Calibration, error) {
	if path == "" {
		return Calibration{}, nil
	}

	values, err := readCalibrationYAML(path)
	if err != nil {
		return Calibration{}, err
	}

	var cal Calibration
	if raw, ok := values["roll_offset_deg"]; ok {
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return cal, fmt.Errorf("invalid roll_offset_deg: %w", err)
		}
		cal.RollOffsetDeg = v
	}
	if raw, ok := values["pitch_offset_deg"]; ok {
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return cal, fmt.Errorf("invalid pitch_offset_deg: %w", err)
		}
		cal.PitchOffsetDeg = v
	}
	return cal, nil
}

func Save(path string, cal Calibration) error {
	body := fmt.Sprintf("roll_offset_deg: %.6g\npitch_offset_deg: %.6g\n", cal.RollOffsetDeg, cal.PitchOffsetDeg)
	return os.WriteFile(path, []byte(body), 0644)
}

func readCalibrationYAML(path string) (map[string]string, error) {
	return config.ReadScalarYAML(path)
}
