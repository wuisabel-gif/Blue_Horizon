package input

import (
	"encoding/binary"
	"fmt"
	"math"

	"blue-horizon/internal/attitude"
)

// attSample is one decoded attitude reading from a ROS 2 message.
type attSample struct {
	t                float64 // seconds (message header stamp)
	roll, pitch, yaw float64 // degrees
}

// cdrReader decodes ROS 2 CDR-serialized messages. It handles only the leading
// fields of the message types Blue Horizon needs (stamp + orientation), reading
// just far enough to reach the quaternion.
//
// ponytail: alignment origin is the byte after the 4-byte encapsulation header
// (the FastCDR / ROS 2 convention). Verify against one real recorded bag before
// trusting decodes on a new stack; that is the one thing this can't self-check.
type cdrReader struct {
	buf []byte
	pos int
	bo  binary.ByteOrder
}

func newCDRReader(data []byte) (*cdrReader, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("cdr: buffer too short (%d bytes)", len(data))
	}
	bo := binary.ByteOrder(binary.BigEndian)
	if data[1]&1 == 1 { // encapsulation header: low bit of byte 1 = little-endian
		bo = binary.LittleEndian
	}
	return &cdrReader{buf: data, pos: 4, bo: bo}, nil
}

func (r *cdrReader) align(n int) {
	if off := (r.pos - 4) % n; off != 0 {
		r.pos += n - off
	}
}

func (r *cdrReader) need(n int) error {
	if r.pos+n > len(r.buf) {
		return fmt.Errorf("cdr: truncated at %d (need %d, have %d)", r.pos, n, len(r.buf))
	}
	return nil
}

func (r *cdrReader) uint32() (uint32, error) {
	r.align(4)
	if err := r.need(4); err != nil {
		return 0, err
	}
	v := r.bo.Uint32(r.buf[r.pos:])
	r.pos += 4
	return v, nil
}

func (r *cdrReader) float64() (float64, error) {
	r.align(8)
	if err := r.need(8); err != nil {
		return 0, err
	}
	v := math.Float64frombits(r.bo.Uint64(r.buf[r.pos:]))
	r.pos += 8
	return v, nil
}

func (r *cdrReader) skipString() error {
	n, err := r.uint32() // CDR string length includes the null terminator
	if err != nil {
		return err
	}
	if err := r.need(int(n)); err != nil {
		return err
	}
	r.pos += int(n)
	return nil
}

func (r *cdrReader) stamp() (float64, error) {
	sec, err := r.uint32()
	if err != nil {
		return 0, err
	}
	nsec, err := r.uint32()
	if err != nil {
		return 0, err
	}
	return float64(int32(sec)) + float64(nsec)*1e-9, nil
}

func (r *cdrReader) quaternion() (attSample, error) {
	x, err := r.float64()
	if err != nil {
		return attSample{}, err
	}
	y, err := r.float64()
	if err != nil {
		return attSample{}, err
	}
	z, err := r.float64()
	if err != nil {
		return attSample{}, err
	}
	w, err := r.float64()
	if err != nil {
		return attSample{}, err
	}
	roll, pitch, yaw := attitude.QuaternionToEulerDeg(x, y, z, w)
	return attSample{roll: roll, pitch: pitch, yaw: yaw}, nil
}

// decodeAttitude dispatches on the ROS 2 schema name and pulls stamp+orientation.
func decodeAttitude(schemaName string, data []byte) (attSample, error) {
	r, err := newCDRReader(data)
	if err != nil {
		return attSample{}, err
	}
	t, err := r.stamp()
	if err != nil {
		return attSample{}, err
	}
	if err := r.skipString(); err != nil { // header.frame_id
		return attSample{}, err
	}

	switch schemaName {
	case "sensor_msgs/msg/Imu":
		// orientation is the next field after the header.
	case "nav_msgs/msg/Odometry":
		if err := r.skipString(); err != nil { // child_frame_id
			return attSample{}, err
		}
		if err := skipFloats(r, 3); err != nil { // pose.pose.position x,y,z
			return attSample{}, err
		}
	case "geometry_msgs/msg/PoseStamped":
		if err := skipFloats(r, 3); err != nil { // pose.position x,y,z
			return attSample{}, err
		}
	default:
		return attSample{}, fmt.Errorf("unsupported message type %q (want sensor_msgs/msg/Imu, nav_msgs/msg/Odometry, or geometry_msgs/msg/PoseStamped)", schemaName)
	}

	s, err := r.quaternion()
	if err != nil {
		return attSample{}, err
	}
	s.t = t
	return s, nil
}

func skipFloats(r *cdrReader, n int) error {
	for i := 0; i < n; i++ {
		if _, err := r.float64(); err != nil {
			return err
		}
	}
	return nil
}
