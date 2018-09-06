package rtsp

import (
	"strings"
	"testing"
)

func TestGetReasonPhrase(test *testing.T) {
	for i := 100; i <= 551; i++ {
		getReasonPhrase(i)
	}
}

func TestRTSPStatusLineMarshal(test *testing.T) {
	statusLine := RTSPStatusLine{
		RTSPVersion:  RTSP_VERSION,
		StatusCode:   200,
		ReasonPhrase: getReasonPhrase(200),
	}
	if !strings.EqualFold(statusLine.Marshal(), "RTSP/1.0 200 OK\r\n") {
		test.Error()
	}
}

func TestRTSPStatusLineUnmarshal(test *testing.T) {
	var str string
	var statusLine RTSPStatusLine
	var err error

	str = "200 OK\r\n"
	err = statusLine.Unmarshal(str)
	if err == nil || !strings.EqualFold("RTSP status line unmarshal error", err.Error()) {
		test.Error()
	}

	str = "RTSP/1.1 200 OK\r\n"
	err = statusLine.Unmarshal(str)
	if err != nil {
		test.Error()
	}
}

func TestRTSPResponseHeaderMarshal(test *testing.T) {
	header := RTSPResponseHeader{
		Server: MANUFACTURER,
	}
	if !strings.EqualFold("Server: "+MANUFACTURER+"\r\n", header.Marshal()) {
		test.Error()
	}
}

func TestRTSPResponseHeaderUnmarshal(test *testing.T) {
	header := RTSPResponseHeader{}
	header.Unmarshal("Server: " + MANUFACTURER + "\r\n")
	if !strings.EqualFold(header.Server, MANUFACTURER) {
		test.Error()
	}
}

func TestRTSPResponseMarshal(test *testing.T) {
	response := RTSPResponse{
		StatusLine: RTSPStatusLine{
			RTSPVersion:  RTSP_VERSION,
			StatusCode:   200,
			ReasonPhrase: getReasonPhrase(200),
		},
		CSeq: 4,
		GeneralHeader: RTSPGeneralHeader{
			Date: "Thu, Nov 27 2014 11:59:41 GMT",
		},
		ResponseHeader: RTSPResponseHeader{
			Server: MANUFACTURER,
		},
		Session:     "66bb9f0bf5ac93a909ac8e88877ae727",
		Transport:   "RTP/AVP/TCP;unicast;interleaved=0-1",
		MessageBody: "v=0\r\no=- 0 0 IN IP4 127.0.0.1\r\ns=No Name\r\nc=IN IP4 127.0.0.1\r\nt=0 0\r\na=tool:libavformat 55.8.102\r\nm=video 9002 RTP/AVP 96\r\nb=AS:400\r\na=rtpmap:96 H264/90000\r\na=framerate:25\r\n",
	}

	if !strings.EqualFold(response.Marshal(), "RTSP/1.0 200 OK\r\nCSeq: 4\r\nDate: Thu, Nov 27 2014 11:59:41 GMT\r\nServer: "+MANUFACTURER+"\r\nSession: 66bb9f0bf5ac93a909ac8e88877ae727\r\nTransport: RTP/AVP/TCP;unicast;interleaved=0-1\r\n\r\nv=0\r\no=- 0 0 IN IP4 127.0.0.1\r\ns=No Name\r\nc=IN IP4 127.0.0.1\r\nt=0 0\r\na=tool:libavformat 55.8.102\r\nm=video 9002 RTP/AVP 96\r\nb=AS:400\r\na=rtpmap:96 H264/90000\r\na=framerate:25\r\n") {
		test.Error()
	}
}

func TestRTSPResponseUnmarshal(test *testing.T) {
	var response RTSPResponse
	var err error

	err = response.Unmarshal("")
	if err == nil || !strings.EqualFold("RTSP response unmarshal error", err.Error()) {
		test.Error()
	}

	err = response.Unmarshal("RTSP/1.0 200 OK\r\nCSeq: 4\r\nDate: Thu, Nov 27 2014 11:59:41 GMT\r\nServer: " + MANUFACTURER + "\r\nSession: 66bb9f0bf5ac93a909ac8e88877ae727\r\nTransport: RTP/AVP/TCP;unicast;interleaved=0-1\r\n\r\nv=0\r\no=- 0 0 IN IP4 127.0.0.1\r\ns=No Name\r\nc=IN IP4 127.0.0.1\r\nt=0 0\r\na=tool:libavformat 55.8.102\r\nm=video 9002 RTP/AVP 96\r\nb=AS:400\r\na=rtpmap:96 H264/90000\r\na=framerate:25\r\n")
	if err != nil {
		test.Error()
	}
}
