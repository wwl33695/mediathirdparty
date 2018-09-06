package rtsp

import (
	"strings"
	"testing"
)

func TestRTSPRequestLineMarshal(test *testing.T) {
	line := RTSPRequestLine{
		Method:      "OPTIONS",
		RequestURI:  "www.deepglint.com/",
		RTSPVersion: RTSP_VERSION,
	}

	if !strings.EqualFold(line.Marshal(), "OPTIONS www.deepglint.com/ RTSP/1.0\r\n") {
		test.Error()
	}
}

func TestRTSPRequestLineUnmarshal(test *testing.T) {
	var err error
	var line RTSPRequestLine

	err = line.Unmarshal("OPTIONS www.deepglint.com/ RTSP/1.0\r\n")
	if err != nil {
		test.Error()
	}
	if !strings.EqualFold(line.RTSPVersion, RTSP_VERSION) {
		test.Error()
	}
	if !strings.EqualFold(line.Method, "OPTIONS") {
		test.Error()
	}
	if !strings.EqualFold(line.RequestURI, "www.deepglint.com/") {
		test.Error()
	}

	err = line.Unmarshal("www.deepglint.com/ RTSP/1.0\r\n")
	if err == nil {
		test.Error()
	}
}

func TestRTSPRequestHeaderMarshal(test *testing.T) {
	header := RTSPRequestHeader{
		Accept:        "application/sdp",
		UserAgent:     MANUFACTURER,
		Authorization: "Digest username=\"admin\", realm=\"4419b727ab09\", nonce=\"66bb9f0bf5ac93a909ac8e88877ae727\", uri=\"rtsp://192.168.1.145:554/MPEG-4/ch2/main/av_stream\", response=\"108084646408d21aa255664781c886fc",
	}
	if !strings.EqualFold(header.Marshal(), "Accept: application/sdp\r\nAuthorization: Digest username=\"admin\", realm=\"4419b727ab09\", nonce=\"66bb9f0bf5ac93a909ac8e88877ae727\", uri=\"rtsp://192.168.1.145:554/MPEG-4/ch2/main/av_stream\", response=\"108084646408d21aa255664781c886fc\r\nUser-Agent: "+MANUFACTURER+"\r\n") {
		test.Error()
	}
}

func TestRTSPRequestHeaderUnmarshal(test *testing.T) {
	str := "Accept: application/sdp\r\nAuthorization: Digest username=\"admin\", realm=\"4419b727ab09\", nonce=\"66bb9f0bf5ac93a909ac8e88877ae727\", uri=\"rtsp://192.168.1.145:554/MPEG-4/ch2/main/av_stream\", response=\"108084646408d21aa255664781c886fc\r\nUser-Agent: " + MANUFACTURER + "\r\n"
	header := RTSPRequestHeader{}
	header.Unmarshal(str)
	if !strings.EqualFold(header.Accept, "application/sdp") {
		test.Error()
	}
	if !strings.EqualFold(header.UserAgent, MANUFACTURER) {
		test.Error()
	}
	if !strings.EqualFold(header.Authorization, "Digest username=\"admin\", realm=\"4419b727ab09\", nonce=\"66bb9f0bf5ac93a909ac8e88877ae727\", uri=\"rtsp://192.168.1.145:554/MPEG-4/ch2/main/av_stream\", response=\"108084646408d21aa255664781c886fc") {
		test.Error()
	}
}

func TestRTSPRequestMarshal(test *testing.T) {
	request := RTSPRequest{
		RequestLine: RTSPRequestLine{
			Method:      "OPTIONS",
			RequestURI:  "www.deepglint.com/",
			RTSPVersion: RTSP_VERSION,
		},
		CSeq: 5,
		GeneralHeader: RTSPGeneralHeader{
			Date: "Thu, Nov 27 2014 11:59:41 GMT",
		},
		RequestHeader: RTSPRequestHeader{
			Accept: "application/sdp",
		},
		EntityHeader: RTSPEntityHeader{
			ContentBase: "www.deepglint.com/",
		},
		Session:     "66bb9f0bf5ac93a909ac8e88877ae727",
		Transport:   "RTP/AVP/TCP;unicast;interleaved=0-1",
		MessageBody: "v=0\r\no=- 0 0 IN IP4 127.0.0.1\r\ns=No Name\r\nc=IN IP4 127.0.0.1\r\nt=0 0\r\na=tool:libavformat 55.8.102\r\nm=video 9002 RTP/AVP 96\r\nb=AS:400\r\na=rtpmap:96 H264/90000\r\na=framerate:25\r\n",
	}
	if !strings.EqualFold(request.Marshal(), "OPTIONS www.deepglint.com/ RTSP/1.0\r\nCSeq: 5\r\nDate: Thu, Nov 27 2014 11:59:41 GMT\r\nAccept: application/sdp\r\nContent-Base: www.deepglint.com/\r\nSession: 66bb9f0bf5ac93a909ac8e88877ae727\r\nTransport: RTP/AVP/TCP;unicast;interleaved=0-1\r\n\r\nv=0\r\no=- 0 0 IN IP4 127.0.0.1\r\ns=No Name\r\nc=IN IP4 127.0.0.1\r\nt=0 0\r\na=tool:libavformat 55.8.102\r\nm=video 9002 RTP/AVP 96\r\nb=AS:400\r\na=rtpmap:96 H264/90000\r\na=framerate:25\r\n") {
		test.Error()
	}
}

func TestRTSPRequestUnmarshal(test *testing.T) {
	var request RTSPRequest
	var err error

	err = request.Unmarshal("")
	if err == nil || !strings.EqualFold("RTSP request unmarshal error", err.Error()) {
		test.Error()
	}

	err = request.Unmarshal("OPTIONS www.deepglint.com/ RTSP/1.0\r\nCSeq: 5\r\nDate: Thu, Nov 27 2014 11:59:41 GMT\r\nAccept: application/sdp\r\nContent-Base: www.deepglint.com/\r\nSession: 66bb9f0bf5ac93a909ac8e88877ae727\r\nTransport: RTP/AVP/TCP;unicast;interleaved=0-1\r\n\r\nv=0\r\no=- 0 0 IN IP4 127.0.0.1\r\ns=No Name\r\nc=IN IP4 127.0.0.1\r\nt=0 0\r\na=tool:libavformat 55.8.102\r\nm=video 9002 RTP/AVP 96\r\nb=AS:400\r\na=rtpmap:96 H264/90000\r\na=framerate:25\r\n")
	if err != nil {
		test.Error()
	}
}
