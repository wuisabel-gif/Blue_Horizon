package input

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type Sample struct {
	TimeSec       float64
	IMURollDeg    float64
	IMUPitchDeg   float64
	IMUYawDeg     float64
	GTSAMRollDeg  float64
	GTSAMPitchDeg float64
	GTSAMYawDeg   float64
	DepthM        float64
}

func ReadCSV(path string) ([]Sample, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return ReadCSVFrom(f)
}

func ReadCSVFrom(r io.Reader) ([]Sample, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true

	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}

	index := map[string]int{}
	for i, h := range header {
		index[strings.TrimSpace(h)] = i
	}

	required := []string{
		"time",
		"imu_roll_deg",
		"imu_pitch_deg",
		"imu_yaw_deg",
		"gtsam_roll_deg",
		"gtsam_pitch_deg",
		"gtsam_yaw_deg",
		"depth_m",
	}
	for _, name := range required {
		if _, ok := index[name]; !ok {
			return nil, fmt.Errorf("missing required CSV column %q", name)
		}
	}

	var samples []Sample
	line := 1
	for {
		record, err := reader.Read()
		line++
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read line %d: %w", line, err)
		}

		value := func(name string) (float64, error) {
			i := index[name]
			if i >= len(record) {
				return 0, fmt.Errorf("line %d missing value for %q", line, name)
			}
			v, err := strconv.ParseFloat(strings.TrimSpace(record[i]), 64)
			if err != nil {
				return 0, fmt.Errorf("line %d invalid %q: %w", line, name, err)
			}
			return v, nil
		}

		sample, err := parseSample(value)
		if err != nil {
			return nil, err
		}
		samples = append(samples, sample)
	}

	return samples, nil
}

func parseSample(value func(string) (float64, error)) (Sample, error) {
	var s Sample
	var err error

	if s.TimeSec, err = value("time"); err != nil {
		return s, err
	}
	if s.IMURollDeg, err = value("imu_roll_deg"); err != nil {
		return s, err
	}
	if s.IMUPitchDeg, err = value("imu_pitch_deg"); err != nil {
		return s, err
	}
	if s.IMUYawDeg, err = value("imu_yaw_deg"); err != nil {
		return s, err
	}
	if s.GTSAMRollDeg, err = value("gtsam_roll_deg"); err != nil {
		return s, err
	}
	if s.GTSAMPitchDeg, err = value("gtsam_pitch_deg"); err != nil {
		return s, err
	}
	if s.GTSAMYawDeg, err = value("gtsam_yaw_deg"); err != nil {
		return s, err
	}
	if s.DepthM, err = value("depth_m"); err != nil {
		return s, err
	}

	return s, nil
}
