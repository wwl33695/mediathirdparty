package rtsp

import (
	"strings"
	"testing"
)

func TestRTSPInterleavedFrameMarshal(test *testing.T) {
	inf := RTSPInterleavedFrame{Channel: 0, Length: 28}
	buf := inf.Marshal()
	if len(buf) != 4 || buf[0] != 0x24 || buf[1] != 0x00 || buf[2] != 0x00 || buf[3] != 0x1c {
		test.Error("RTSP interleaved frame marshal error")
	}
}

func TestRTSPInterleavedFrameUnmarshal(test *testing.T) {
	var inf RTSPInterleavedFrame
	var buf []byte
	var err error
	buf = []byte{0x24, 0x00, 0x00, 0x1c}
	err = inf.Unmarshal(buf)
	if err != nil || inf.Channel != 0 || inf.Length != 28 {
		test.Error("RTSP interleaved frame unmarshal error")
	}

	buf = []byte{0x24, 0x00, 0x00}
	err = inf.Unmarshal(buf)
	if err == nil || !strings.EqualFold(err.Error(), "buf length is smaller than 4") {
		test.Error("RTSP interleaved frame check size error")
	}

	buf = []byte{0x23, 0x00, 0x00, 0x1c}
	err = inf.Unmarshal(buf)
	if err == nil || !strings.EqualFold(err.Error(), "RTSP interleaved frame not found") {
		test.Error("RTSP interleaved frame check magic error")
	}
}
