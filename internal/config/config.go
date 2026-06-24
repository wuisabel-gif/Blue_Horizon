package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	RollWarningDeg              float64
	RollDangerDeg               float64
	PitchWarningDeg             float64
	PitchDangerDeg              float64
	EstimatorMismatchWarningDeg float64
	EstimatorMismatchDangerDeg  float64
	YawRateWarningDegS          float64
}

func Default() Config {
	return Config{
		RollWarningDeg:              10,
		RollDangerDeg:               25,
		PitchWarningDeg:             8,
		PitchDangerDeg:              20,
		EstimatorMismatchWarningDeg: 7,
		EstimatorMismatchDangerDeg:  15,
		YawRateWarningDegS:          30,
	}
}

func Load(path string) (Config, error) {
	cfg := Default()
	if path == "" {
		return cfg, nil
	}

	values, err := ReadScalarYAML(path)
	if err != nil {
		return cfg, err
	}

	assignments := map[string]*float64{
		"roll_warning_deg":               &cfg.RollWarningDeg,
		"roll_danger_deg":                &cfg.RollDangerDeg,
		"pitch_warning_deg":              &cfg.PitchWarningDeg,
		"pitch_danger_deg":               &cfg.PitchDangerDeg,
		"estimator_mismatch_warning_deg": &cfg.EstimatorMismatchWarningDeg,
		"estimator_mismatch_danger_deg":  &cfg.EstimatorMismatchDangerDeg,
		"yaw_rate_warning_deg_s":         &cfg.YawRateWarningDegS,
	}

	for key, raw := range values {
		target, ok := assignments[key]
		if !ok {
			return cfg, fmt.Errorf("unknown config key %q", key)
		}
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return cfg, fmt.Errorf("invalid value for %q: %w", key, err)
		}
		*target = v
	}

	return cfg, nil
}

func ReadScalarYAML(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	values := map[string]string{}
	scanner := bufio.NewScanner(f)
	line := 0
	for scanner.Scan() {
		line++
		text := strings.TrimSpace(scanner.Text())
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		parts := strings.SplitN(text, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("%s:%d expected key: value", path, line)
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, `"'`)
		values[key] = value
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return values, nil
}
