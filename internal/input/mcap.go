package input

import (
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/foxglove/mcap/go/mcap"
)

// ReadMCAP reads attitude from a rosbag2 .mcap file, decoding orientation from
// the IMU topic and (optionally) an estimator topic, and merges them into the
// same []Sample the CSV path produces. estTopic may be empty, in which case the
// estimator fields mirror the IMU so mismatch checks stay silent.
func ReadMCAP(path, imuTopic, estTopic string) ([]Sample, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader, err := mcap.NewReader(f)
	if err != nil {
		return nil, err
	}
	it, err := reader.Messages()
	if err != nil {
		return nil, err
	}

	var imu, est []attSample
	for {
		schema, channel, message, err := it.Next(nil)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if channel.Topic != imuTopic && channel.Topic != estTopic {
			continue
		}
		if schema == nil {
			return nil, fmt.Errorf("channel %q has no schema", channel.Topic)
		}
		s, err := decodeAttitude(schema.Name, message.Data)
		if err != nil {
			return nil, fmt.Errorf("topic %q: %w", channel.Topic, err)
		}
		if s.t == 0 { // header stamp unset: fall back to bag log time
			s.t = float64(message.LogTime) * 1e-9
		}
		if channel.Topic == imuTopic {
			imu = append(imu, s)
		} else {
			est = append(est, s)
		}
	}

	if len(imu) == 0 {
		return nil, fmt.Errorf("no messages on IMU topic %q", imuTopic)
	}
	return mergeSamples(imu, est), nil
}

// mergeSamples walks the IMU timeline and pairs each reading with the nearest
// estimator reading in time. Timestamps are zero-based off the first IMU sample
// to match the CSV convention (t starts at 0).
func mergeSamples(imu, est []attSample) []Sample {
	sort.Slice(imu, func(i, j int) bool { return imu[i].t < imu[j].t })
	sort.Slice(est, func(i, j int) bool { return est[i].t < est[j].t })

	t0 := imu[0].t
	out := make([]Sample, 0, len(imu))
	j := 0
	for _, a := range imu {
		gr, gp, gy := a.roll, a.pitch, a.yaw // default: mirror IMU when no estimator
		if len(est) > 0 {
			// a.t is nondecreasing, so the nearest est index only moves forward.
			for j+1 < len(est) && absf(est[j+1].t-a.t) <= absf(est[j].t-a.t) {
				j++
			}
			gr, gp, gy = est[j].roll, est[j].pitch, est[j].yaw
		}
		out = append(out, Sample{
			TimeSec:       a.t - t0,
			IMURollDeg:    a.roll,
			IMUPitchDeg:   a.pitch,
			IMUYawDeg:     a.yaw,
			GTSAMRollDeg:  gr,
			GTSAMPitchDeg: gp,
			GTSAMYawDeg:   gy,
		})
	}
	return out
}

func absf(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
