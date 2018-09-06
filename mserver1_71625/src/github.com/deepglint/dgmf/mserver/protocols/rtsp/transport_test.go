package rtsp

import (
	"strings"
	"testing"
)

func TestRTSPTransportMarshal(test *testing.T) {
	transport := RTSPTransport{
		LowerTransport: "TCP",
		CastType:       "unicast",
		Interleaved:    "0-1",
	}
	if !strings.EqualFold(transport.Marshal(), "RTP/AVP/TCP;unicast;interleaved=0-1") {
		test.Error("Transport marshal error")
	}
}

func TestRTSPTransportUnmarshal(test *testing.T) {
	transport := RTSPTransport{}
	transport.Unmarshal("RTP/AVP/TCP;unicast;interleaved=0-1")
	if !strings.EqualFold(transport.LowerTransport, "TCP") {
		test.Error("Transport unmarshal error")
	}
	if !strings.EqualFold(transport.CastType, "unicast") {
		test.Error("Transport unmarshal error")
	}
	if !strings.EqualFold(transport.Interleaved, "0-1") {
		test.Error("Transport unmarshal error")
	}
}
