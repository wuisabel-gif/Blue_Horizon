package input

import (
	"bytes"
	"encoding/binary"
	"math"
	"testing"

	"blue-horizon/internal/attitude"
)

// cdrBuilder writes ROS 2 CDR (little-endian) with the same alignment origin the
// decoder assumes (byte after the 4-byte encapsulation header), so a round trip
// pins field order, string handling, and endianness plumbing.
type cdrBuilder struct{ b bytes.Buffer }

func newCDRBuilder() *cdrBuilder {
	c := &cdrBuilder{}
	c.b.Write([]byte{0x00, 0x01, 0x00, 0x00}) // CDR_LE encapsulation header
	return c
}

func (c *cdrBuilder) align(n int) {
	for (c.b.Len()-4)%n != 0 {
		c.b.WriteByte(0)
	}
}

func (c *cdrBuilder) u32(v uint32) {
	c.align(4)
	var tmp [4]byte
	binary.LittleEndian.PutUint32(tmp[:], v)
	c.b.Write(tmp[:])
}

func (c *cdrBuilder) f64(v float64) {
	c.align(8)
	var tmp [8]byte
	binary.LittleEndian.PutUint64(tmp[:], math.Float64bits(v))
	c.b.Write(tmp[:])
}

func (c *cdrBuilder) str(s string) {
	c.u32(uint32(len(s) + 1))
	c.b.WriteString(s)
	c.b.WriteByte(0)
}

func TestDecodeImu(t *testing.T) {
	const x, y, z, w = 0.0, 0.0, 0.70710678, 0.70710678 // 90° yaw
	c := newCDRBuilder()
	c.u32(12)             // stamp.sec
	c.u32(500000000)      // stamp.nanosec
	c.str("imu_link")     // header.frame_id
	c.f64(x)
	c.f64(y)
	c.f64(z)
	c.f64(w)

	got, err := decodeAttitude("sensor_msgs/msg/Imu", c.b.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(got.t-12.5) > 1e-9 {
		t.Fatalf("stamp = %v want 12.5", got.t)
	}
	wantR, wantP, wantYaw := attitude.QuaternionToEulerDeg(x, y, z, w)
	if math.Abs(got.roll-wantR) > 1e-6 || math.Abs(got.pitch-wantP) > 1e-6 || math.Abs(got.yaw-wantYaw) > 1e-6 {
		t.Fatalf("euler = (%.4f,%.4f,%.4f) want (%.4f,%.4f,%.4f)", got.roll, got.pitch, got.yaw, wantR, wantP, wantYaw)
	}
}

func TestDecodeOdometrySkipsPositionAndChildFrame(t *testing.T) {
	const x, y, z, w = 0.70710678, 0.0, 0.0, 0.70710678 // 90° roll
	c := newCDRBuilder()
	c.u32(1)
	c.u32(0)
	c.str("odom")     // header.frame_id
	c.str("base_link") // child_frame_id
	c.f64(1.1)         // position.x
	c.f64(2.2)         // position.y
	c.f64(3.3)         // position.z
	c.f64(x)
	c.f64(y)
	c.f64(z)
	c.f64(w)

	got, err := decodeAttitude("nav_msgs/msg/Odometry", c.b.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(got.roll-90) > 1e-4 {
		t.Fatalf("roll = %v want ~90 (position/child_frame not skipped correctly?)", got.roll)
	}
}

func TestDecodeUnsupportedType(t *testing.T) {
	c := newCDRBuilder()
	c.u32(0)
	c.u32(0)
	c.str("f")
	if _, err := decodeAttitude("std_msgs/msg/String", c.b.Bytes()); err == nil {
		t.Fatal("want error for unsupported type")
	}
}

func TestMergeSamplesNearestAndZeroBase(t *testing.T) {
	imu := []attSample{{t: 100.0, roll: 1}, {t: 100.2, roll: 2}}
	est := []attSample{{t: 100.19, roll: 20}, {t: 100.05, roll: 10}} // unsorted on purpose
	out := mergeSamples(imu, est)
	if len(out) != 2 {
		t.Fatalf("want 2 samples got %d", len(out))
	}
	if out[0].TimeSec != 0 || math.Abs(out[1].TimeSec-0.2) > 1e-9 {
		t.Fatalf("timeline not zero-based: %v", []float64{out[0].TimeSec, out[1].TimeSec})
	}
	if out[0].GTSAMRollDeg != 10 || out[1].GTSAMRollDeg != 20 {
		t.Fatalf("nearest-neighbor pairing wrong: %v", []float64{out[0].GTSAMRollDeg, out[1].GTSAMRollDeg})
	}
}

func TestMergeSamplesNoEstimatorMirrors(t *testing.T) {
	imu := []attSample{{t: 5, roll: 3, pitch: 4, yaw: 5}}
	out := mergeSamples(imu, nil)
	if out[0].GTSAMRollDeg != 3 || out[0].GTSAMPitchDeg != 4 || out[0].GTSAMYawDeg != 5 {
		t.Fatalf("no-estimator case should mirror IMU, got %+v", out[0])
	}
}
