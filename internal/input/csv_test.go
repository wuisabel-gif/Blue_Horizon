package input

import (
	"strings"
	"testing"
)

const header = "time,imu_roll_deg,imu_pitch_deg,imu_yaw_deg,gtsam_roll_deg,gtsam_pitch_deg,gtsam_yaw_deg,depth_m"

func TestReadCSVHappyPath(t *testing.T) {
	in := header + "\n0.00,1.2,-0.5,10.0,1.0,-0.3,10.1,0.2\n"
	samples, err := ReadCSVFrom(strings.NewReader(in))
	if err != nil {
		t.Fatal(err)
	}
	if len(samples) != 1 {
		t.Fatalf("want 1 sample got %d", len(samples))
	}
	s := samples[0]
	if s.TimeSec != 0 || s.IMURollDeg != 1.2 || s.GTSAMYawDeg != 10.1 || s.DepthM != 0.2 {
		t.Fatalf("parsed wrong values: %+v", s)
	}
}

func TestReadCSVColumnOrderIndependent(t *testing.T) {
	// Columns shuffled: parser must map by header name, not position.
	in := "depth_m,time,gtsam_yaw_deg,gtsam_pitch_deg,gtsam_roll_deg,imu_yaw_deg,imu_pitch_deg,imu_roll_deg\n" +
		"0.2,0.00,10.1,-0.3,1.0,10.0,-0.5,1.2\n"
	samples, err := ReadCSVFrom(strings.NewReader(in))
	if err != nil {
		t.Fatal(err)
	}
	s := samples[0]
	if s.IMURollDeg != 1.2 || s.DepthM != 0.2 || s.GTSAMYawDeg != 10.1 {
		t.Fatalf("column-order mapping wrong: %+v", s)
	}
}

func TestReadCSVMissingColumn(t *testing.T) {
	in := "time,imu_roll_deg\n0.0,1.2\n"
	_, err := ReadCSVFrom(strings.NewReader(in))
	if err == nil || !strings.Contains(err.Error(), "missing required CSV column") {
		t.Fatalf("want missing-column error got %v", err)
	}
}

func TestReadCSVInvalidValue(t *testing.T) {
	in := header + "\n0.00,not_a_number,-0.5,10.0,1.0,-0.3,10.1,0.2\n"
	_, err := ReadCSVFrom(strings.NewReader(in))
	if err == nil || !strings.Contains(err.Error(), "invalid") {
		t.Fatalf("want invalid-value error got %v", err)
	}
}
