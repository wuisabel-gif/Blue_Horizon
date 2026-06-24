package attitude

import "math"

func AbsDeg(v float64) float64 {
	return math.Abs(v)
}

func DeltaDegrees(a, b float64) float64 {
	delta := a - b
	for delta > 180 {
		delta -= 360
	}
	for delta < -180 {
		delta += 360
	}
	return delta
}

func YawRateDegS(previousYaw, currentYaw, dt float64) float64 {
	if dt <= 0 {
		return 0
	}
	return DeltaDegrees(currentYaw, previousYaw) / dt
}
