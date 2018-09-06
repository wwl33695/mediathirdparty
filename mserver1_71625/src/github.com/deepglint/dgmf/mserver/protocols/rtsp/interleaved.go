package rtsp

import (
	"encoding/binary"
	"errors"
)

type RTSPInterleavedFrame struct {
	Channel uint8
	Length  uint16
}

func (this *RTSPInterleavedFrame) Marshal() []byte {
	buf := make([]byte, 4)
	buf[0] = 0x24
	buf[1] = byte(this.Channel)
	binary.BigEndian.PutUint16(buf[2:4], this.Length)
	return buf
}

func (this *RTSPInterleavedFrame) Unmarshal(buf []byte) error {
	if len(buf) < 4 {
		return errors.New("buf length is smaller than 4")
	}
	if buf[0] != 0x24 {
		return errors.New("RTSP interleaved frame not found")
	}
	this.Channel = uint8(buf[1])
	this.Length = binary.BigEndian.Uint16(buf[2:4])
	return nil
}
