package core

import (
	"log"
	"testing"
)

func TestPool(test *testing.T) {
	pool := GetESPool()
	log.Println("[POOL SIZE]", len(pool.LiveStreams))
	if len(pool.LiveStreams) != 0 {
		test.Error()
	}

	pool.AddInput("LEO001", true, "udp://127.0.0.1:9002")
	log.Println("[POOL SIZE]", len(pool.LiveStreams))
	if len(pool.LiveStreams) != 1 {
		test.Error()
	}

	pool.AddSession("LEO001", true, "Sjn3e1", "127.0.0.1:6311", "tcp", "rtsp")

	go func() {
		for i := 0; i < 10; i++ {
			frame := &H264ESFrame{
				Data:      []byte{0x00, 0x00, 0x00, 0x01},
				IFrame:    true,
				Timestamp: 0x1234,
			}
			pool.LiveStreams["LEO001"].Sessions["Sjn3e1"].Frame <- frame
		}
		pool.RemoveSession("LEO001", true, "Sjn3e1")
	}()

	for frame := range pool.LiveStreams["LEO001"].Sessions["Sjn3e1"].Frame {
		log.Println(frame.Data)
		if frame.Data[0] != 0x00 || frame.Data[1] != 0x00 || frame.Data[2] != 0x00 || frame.Data[3] != 0x01 {
			test.Error()
		}
	}

	pool.RemoveInput("LEO001", true)

	log.Println("[POOL SIZE]", len(pool.LiveStreams))
	if len(pool.LiveStreams) != 0 {
		test.Error()
	}d
}
