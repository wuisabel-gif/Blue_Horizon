package attitude

import (
	"math"
	"testing"
)

func TestDeltaDegreesWrap(t *testing.T) {
	cases := []struct {
		name    string
		a, b    float64
		want    float64
	}{
		{"no wrap", 10, 4, 6},
		{"cross +180 boundary", -179, 179, 2}, // shortest path is +2, not -358
		{"cross -180 boundary", 179, -179, -2},
		{"exactly +180 stays", 180, 0, 180},
		{"exactly -180 stays", -180, 0, -180},
		{"large positive", 540, 0, 180},
		{"large negative", -540, 0, -180},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := DeltaDegrees(c.a, c.b); math.Abs(got-c.want) > 1e-9 {
				t.Fatalf("DeltaDegrees(%v,%v)=%v want %v", c.a, c.b, got, c.want)
			}
		})
	}
}

func TestYawRateWrapDoesNotSpike(t *testing.T) {
	// Yaw ticks 179 -> -179 over 0.1s is a 2 deg move, i.e. 20 deg/s.
	// Without wrapping it would falsely read -3580 deg/s.
	if got := YawRateDegS(179, -179, 0.1); math.Abs(got-20) > 1e-9 {
		t.Fatalf("YawRateDegS wrap = %v want 20", got)
	}
}

func TestYawRateNonPositiveDt(t *testing.T) {
	if got := YawRateDegS(10, 40, 0); got != 0 {
		t.Fatalf("dt=0 want 0 got %v", got)
	}
	if got := YawRateDegS(10, 40, -1); got != 0 {
		t.Fatalf("dt<0 want 0 got %v", got)
	}
}
