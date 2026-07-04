package input

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/foxglove/mcap/go/mcap"
)

// writeMessage appends one CDR message on the given channel to the writer.
func writeCDRMessage(t *testing.T, w *mcap.Writer, chanID uint16, logTime uint64, data []byte) {
	t.Helper()
	if err := w.WriteMessage(&mcap.Message{ChannelID: chanID, LogTime: logTime, Data: data}); err != nil {
		t.Fatal(err)
	}
}

func TestReadMCAPEndToEnd(t *testing.T) {
	path := filepath.Join(t.TempDir(), "flight.mcap")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}

	w, err := mcap.NewWriter(f, &mcap.WriterOptions{Chunked: true, ChunkSize: 1 << 16})
	if err != nil {
		t.Fatal(err)
	}
	if err := w.WriteHeader(&mcap.Header{Profile: "ros2"}); err != nil {
		t.Fatal(err)
	}
	if err := w.WriteSchema(&mcap.Schema{ID: 1, Name: "sensor_msgs/msg/Imu", Encoding: "ros2msg"}); err != nil {
		t.Fatal(err)
	}
	if err := w.WriteSchema(&mcap.Schema{ID: 2, Name: "nav_msgs/msg/Odometry", Encoding: "ros2msg"}); err != nil {
		t.Fatal(err)
	}
	if err := w.WriteChannel(&mcap.Channel{ID: 1, SchemaID: 1, Topic: "/imu/data", MessageEncoding: "cdr"}); err != nil {
		t.Fatal(err)
	}
	if err := w.WriteChannel(&mcap.Channel{ID: 2, SchemaID: 2, Topic: "/odometry/filtered", MessageEncoding: "cdr"}); err != nil {
		t.Fatal(err)
	}

	// Realistic epoch stamps (seconds since 1970). IMU: level at t0, 90° yaw at
	// t0+1. Odometry stays level -> a yaw mismatch appears in the merged samples.
	const t0 = 1_700_000_000
	imu0 := newCDRBuilder()
	imu0.u32(t0)
	imu0.u32(0)
	imu0.str("imu")
	imu0.f64(0)
	imu0.f64(0)
	imu0.f64(0)
	imu0.f64(1) // identity
	writeCDRMessage(t, w, 1, t0*1_000_000_000, imu0.b.Bytes())

	imu1 := newCDRBuilder()
	imu1.u32(t0 + 1)
	imu1.u32(0)
	imu1.str("imu")
	imu1.f64(0)
	imu1.f64(0)
	imu1.f64(0.70710678)
	imu1.f64(0.70710678) // 90° yaw
	writeCDRMessage(t, w, 1, (t0+1)*1_000_000_000, imu1.b.Bytes())

	odom := newCDRBuilder()
	odom.u32(t0)
	odom.u32(500_000_000)
	odom.str("odom")
	odom.str("base_link")
	odom.f64(0)
	odom.f64(0)
	odom.f64(0)
	odom.f64(0)
	odom.f64(0)
	odom.f64(0)
	odom.f64(1) // identity (level)
	writeCDRMessage(t, w, 2, 1_500_000_000, odom.b.Bytes())

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	samples, err := ReadMCAP(path, "/imu/data", "/odometry/filtered")
	if err != nil {
		t.Fatal(err)
	}
	if len(samples) != 2 {
		t.Fatalf("want 2 IMU samples got %d", len(samples))
	}
	if samples[0].TimeSec != 0 || samples[1].TimeSec != 1 {
		t.Fatalf("timeline: %v", []float64{samples[0].TimeSec, samples[1].TimeSec})
	}
	// Second IMU sample yaws to 90° while the estimator stays level.
	if samples[1].IMUYawDeg < 89 || samples[1].IMUYawDeg > 91 {
		t.Fatalf("imu yaw = %v want ~90", samples[1].IMUYawDeg)
	}
	if samples[1].GTSAMYawDeg < -1 || samples[1].GTSAMYawDeg > 1 {
		t.Fatalf("estimator yaw = %v want ~0", samples[1].GTSAMYawDeg)
	}
}
