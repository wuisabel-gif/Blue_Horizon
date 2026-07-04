package attitude

import "math"

// QuaternionToEulerDeg converts a unit quaternion (x,y,z,w) to roll/pitch/yaw
// in degrees using the ZYX convention: roll about x, pitch about y, yaw about z
// (REP-103 body frame). This is how sensor_msgs/Imu and nav_msgs/Odometry carry
// attitude on the wire, so MCAP ingestion funnels through here.
func QuaternionToEulerDeg(x, y, z, w float64) (roll, pitch, yaw float64) {
	sinrCosp := 2 * (w*x + y*z)
	cosrCosp := 1 - 2*(x*x+y*y)
	roll = math.Atan2(sinrCosp, cosrCosp)

	sinp := 2 * (w*y - z*x)
	if sinp > 1 {
		sinp = 1 // clamp: numerical drift past the pole would NaN the Asin
	}
	if sinp < -1 {
		sinp = -1
	}
	pitch = math.Asin(sinp)

	sinyCosp := 2 * (w*z + x*y)
	cosyCosp := 1 - 2*(y*y+z*z)
	yaw = math.Atan2(sinyCosp, cosyCosp)

	const r2d = 180 / math.Pi
	return roll * r2d, pitch * r2d, yaw * r2d
}
