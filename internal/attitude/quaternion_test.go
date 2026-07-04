package attitude

import (
	"math"
	"testing"
)

func TestQuaternionToEulerDeg(t *testing.T) {
	const s = math.Sqrt2 / 2 // sin(45°)=cos(45°) for a 90° rotation
	cases := []struct {
		name                    string
		x, y, z, w              float64
		roll, pitch, yaw        float64
	}{
		{"identity", 0, 0, 0, 1, 0, 0, 0},
		{"90 yaw about z", 0, 0, s, s, 0, 0, 90},
		{"90 roll about x", s, 0, 0, s, 90, 0, 0},
		// 30° pitch about y (not 90°: that is gimbal lock, where roll/yaw are ambiguous).
		{"30 pitch about y", 0, math.Sin(math.Pi / 12), 0, math.Cos(math.Pi / 12), 0, 30, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r, p, y := QuaternionToEulerDeg(c.x, c.y, c.z, c.w)
			if math.Abs(r-c.roll) > 1e-6 || math.Abs(p-c.pitch) > 1e-6 || math.Abs(y-c.yaw) > 1e-6 {
				t.Fatalf("got (%.4f,%.4f,%.4f) want (%.1f,%.1f,%.1f)", r, p, y, c.roll, c.pitch, c.yaw)
			}
		})
	}
}

func TestPitchClampAtPole(t *testing.T) {
	// Gimbal-lock quaternion (90° pitch) must not NaN via Asin overflow.
	if _, p, _ := QuaternionToEulerDeg(0, math.Sqrt2/2, 0, math.Sqrt2/2); math.IsNaN(p) {
		t.Fatal("pitch is NaN at the pole")
	}
}
