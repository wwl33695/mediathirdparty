package h264

import (
	// "fmt"
	"github.com/deepglint/dgmf/mserver/core"
	"testing"
)

func TestOpen(test *testing.T) {
	var err error
	stream := &core.LiveStream{
		StreamId:    "stream",
		URI:         "file://testdata/test.h264",
		InputStatus: false,
		Sessions:    make(map[string]*core.Session),
		Protocols:   make(map[string]bool),
		InputCtx:    &FileH264InputLayer{},
	}

	stream.Sessions["session"] = &core.Session{
		SessionId:  "session",
		RemoteAddr: "127.0.0.1",
		Network:    "tcp",
		Frame:      make(chan *core.H264ESFrame),
		Protocol:   "rtsp",
	}

	err = stream.InputCtx.Open(stream.URI, stream)
	if err != nil {
		test.Error()
	}
	for i := 0; i < 10; i++ {
		<-stream.Sessions["session"].Frame
	}
	err = stream.InputCtx.Close()
	if err != nil {
		test.Error()
	}
}
