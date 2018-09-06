package rtsp

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/deepglint/dgmf/mserver/core"
	"github.com/deepglint/dgmf/mserver/utils"
)

func (this *RTSPServer) errorResponse(request RTSPRequest, statusCode int, sessionCtx *SessionContext) {
	response := RTSPResponse{
		StatusLine: RTSPStatusLine{
			RTSPVersion:  RTSP_VERSION,
			StatusCode:   statusCode,
			ReasonPhrase: getReasonPhrase(statusCode),
		},
		CSeq: request.CSeq,
		ResponseHeader: RTSPResponseHeader{
			Server: utils.MANUFACTURER,
		},
		Session: sessionCtx.SessionId,
	}

	log.Println("[RTSP_SERVER] Response:\n" + response.Marshal())
	sessionCtx.TCPConnect.Write([]byte(response.Marshal()))
}

func (this *RTSPServer) initRequest(sessionCtx *SessionContext) error {
	var err error

	request := RTSPRequest{}
	buf := make([]byte, REQ_RSP_SIZE)
	size, err := sessionCtx.TCPConnect.Read(buf)
	if err != nil {
		fmt.Println(err)
		this.errorResponse(request, 500, sessionCtx)
		sessionCtx.TCPConnect.Close()
		return err
	}
	if size >= REQ_RSP_SIZE {
		err = errors.New("rtsp request size error")
		this.errorResponse(request, 413, sessionCtx)
		sessionCtx.TCPConnect.Close()
		return err
	}

	err = request.Unmarshal(string(buf))
	if err != nil {
		this.errorResponse(request, 400, sessionCtx)
		sessionCtx.TCPConnect.Close()
		return err
	}
	log.Println("[RTSP_SERVER] Request:\n" + request.Marshal())

	if !strings.EqualFold(request.RequestLine.RTSPVersion, RTSP_VERSION) {
		err = errors.New("rtsp version error")
		this.errorResponse(request, 505, sessionCtx)
		sessionCtx.TCPConnect.Close()
		return err
	}

	switch {
	case strings.EqualFold(request.RequestLine.Method, OPTIONS):
		err = this.options(request, sessionCtx)
	case strings.EqualFold(request.RequestLine.Method, DESCRIBE):
		err = this.describe(request, sessionCtx)
	case strings.EqualFold(request.RequestLine.Method, SETUP):
		err = this.setup(request, sessionCtx)
	case strings.EqualFold(request.RequestLine.Method, PLAY):
		err = this.play(request, sessionCtx)
	case strings.EqualFold(request.RequestLine.Method, TEARDOWN):
		err = this.teardown(request, sessionCtx)
	default:
		err = errors.New("Unsupported rtsp method")
		this.errorResponse(request, 405, sessionCtx)
		sessionCtx.TCPConnect.Close()
		return err
	}

	return err
}

func (this *RTSPServer) options(request RTSPRequest, sessionCtx *SessionContext) error {
	var err error

	response := RTSPResponse{
		StatusLine: RTSPStatusLine{
			RTSPVersion:  RTSP_VERSION,
			StatusCode:   200,
			ReasonPhrase: getReasonPhrase(200),
		},
		CSeq: request.CSeq,
		ResponseHeader: RTSPResponseHeader{
			Server: utils.MANUFACTURER,
			Public: fmt.Sprintf("%s, %s, %s, %s, %s", OPTIONS, DESCRIBE, SETUP, PLAY, TEARDOWN),
		},
		Session: sessionCtx.SessionId,
	}

	log.Println("[RTSP_SERVER] Response:\n" + response.Marshal())
	sessionCtx.TCPConnect.Write([]byte(response.Marshal()))

	err = this.initRequest(sessionCtx)
	return err
}

func (this *RTSPServer) describe(request RTSPRequest, sessionCtx *SessionContext) error {
	var err error

	urlCtx, err := url.Parse(request.RequestLine.RequestURI)
	if err != nil {
		this.errorResponse(request, 404, sessionCtx)
		sessionCtx.TCPConnect.Close()
		return err
	}

	if len(urlCtx.Path) < 2 {
		err = errors.New("rtsp url is invalid")
		this.errorResponse(request, 404, sessionCtx)
		sessionCtx.TCPConnect.Close()
		return err
	}

	parts := strings.Split(urlCtx.Path[1:], "/")
	if len(parts) != 2 {
		err = errors.New("rtsp url is invalid")
		this.errorResponse(request, 404, sessionCtx)
		sessionCtx.TCPConnect.Close()
		return err
	}

	if !strings.EqualFold(parts[0], "live") &&
		!strings.EqualFold(parts[0], "vod") &&
		!strings.EqualFold(parts[0], "proxy") {

		err = errors.New("Unsupported rtsp type")
		this.errorResponse(request, 404, sessionCtx)
		sessionCtx.TCPConnect.Close()
		return err
	}
	sessionCtx.StreamType = strings.ToLower(parts[0])
	streamId := strings.Split(strings.ToLower(urlCtx.Path[1:]), "/")[1]

	pool := core.GetESPool()
	sdp := "v=0\r\n" +
		"o=- 0 0 IN IP4 0.0.0.0\r\n" +
		"s=H.264 Video, streamed by " + utils.MANUFACTURER + "\r\n" +
		"c=IN IP4 0.0.0.0\r\n" +
		"t=0 0\r\n" +
		"m=video 0 RTP/AVP 96\r\n" +
		"b=AS:500\r\n" +
		"a=tool:" + utils.MANUFACTURER + "\r\n" +
		"a=rtpmap:96 H264/90000\r\n"

	if sessionCtx.StreamType == "live" {
		if pool.Live.ExistOutput(streamId, "rtsp") {
			sessionCtx.StreamId = streamId

			if fmtp, err := pool.Live.GetFMTP(sessionCtx.StreamId); err != nil && len(fmtp) > 0 {
				sdp += "a=fmtp:" + fmtp + "\r\n"
			}

			sps, err := pool.Live.GetSPS(sessionCtx.StreamId)
			pps, err := pool.Live.GetPPS(sessionCtx.StreamId)
			if err == nil {
				sdp += "a=fmtp:96 packetization-mode=1;sprop-parameter-sets=" + sps + "," + pps + "\r\n"
			} else {
				sdp += "a=fmtp:96 packetization-mode=1"
			}

		}
	} else if sessionCtx.StreamType == "vod" {

	} else if sessionCtx.StreamType == "proxy" {
		if uris, ok := urlCtx.Query()["uri"]; ok && len(uris) == 1 && len(uris[0]) > 0 {
			uri := uris[0]
			sessionCtx.StreamId = streamId
			sdp += "a=fmtp:96 profile-level-id=1\r\n"
			receiver := &RTSPReceiver{}
			receiver.SetTimeout(20 * time.Second)
			sessionCtx.RTMS = make(chan core.RTMessage)
			go receiver.Open(uri, sessionCtx.StreamId, sessionCtx.RTMS)

			rtm := <-sessionCtx.RTMS
			if rtm.Status != 200 {
				this.errorResponse(request, 404, sessionCtx)
				sessionCtx.TCPConnect.Close()
				return rtm.Error
			}
			if pool.Proxy.ExistStream(sessionCtx.StreamId) {
				sessionCtx.TCPConnect.Close()
				return errors.New("StreamId existed")
			}

			pool.Proxy.AddStream(sessionCtx.StreamId, uri, urlCtx.String(), "rtsp", receiver, sessionCtx.TCPConnect)
		} else {
			err = errors.New("rtsp proxy uri is invalid")
			this.errorResponse(request, 404, sessionCtx)
			sessionCtx.TCPConnect.Close()
			return err
		}
	}

	if len(sessionCtx.StreamId) == 0 {
		err = errors.New("No StreamId found")
		this.errorResponse(request, 404, sessionCtx)
		sessionCtx.TCPConnect.Close()
		return err
	}

	sdp += "a=control:rtsp://" + urlCtx.Host + "/" + sessionCtx.StreamId + "/trackID=1"

	response := RTSPResponse{
		StatusLine: RTSPStatusLine{
			RTSPVersion:  RTSP_VERSION,
			StatusCode:   200,
			ReasonPhrase: getReasonPhrase(200),
		},
		CSeq: request.CSeq,
		ResponseHeader: RTSPResponseHeader{
			Server: utils.MANUFACTURER,
		},
		EntityHeader: RTSPEntityHeader{
			ContentType:   "application/sdp",
			ContentBase:   "rtsp://" + urlCtx.Host + "/" + sessionCtx.StreamId + "/",
			ContentLength: fmt.Sprintf("%d", len(sdp)),
		},
		Session:     sessionCtx.SessionId,
		MessageBody: sdp,
	}

	log.Println("[RTSP_SERVER] Response:\n" + response.Marshal())
	sessionCtx.TCPConnect.Write([]byte(response.Marshal()))

	err = this.initRequest(sessionCtx)

	return err
}

func (this *RTSPServer) setup(request RTSPRequest, sessionCtx *SessionContext) error {
	var err error

	if len(sessionCtx.StreamId) == 0 {
		urlCtx, err := url.Parse(request.RequestLine.RequestURI)
		if err != nil {
			this.errorResponse(request, 404, sessionCtx)
			sessionCtx.TCPConnect.Close()
			return err
		}

		if len(urlCtx.Path) < 2 {
			err = errors.New("rtsp url invalid")
			this.errorResponse(request, 404, sessionCtx)
			sessionCtx.TCPConnect.Close()
			return err
		}

		if len(strings.Split(urlCtx.Path[1:], "/")) != 2 {
			err = errors.New("rtsp url invalid")
			this.errorResponse(request, 404, sessionCtx)
			sessionCtx.TCPConnect.Close()
			return err
		}

		if len(sessionCtx.StreamId) == 0 {
			err = errors.New("No StreamId found")
			this.errorResponse(request, 404, sessionCtx)
			sessionCtx.TCPConnect.Close()
			return err
		}
	}

	transport := RTSPTransport{}
	transport.Unmarshal(request.Transport)

	if strings.Contains(request.Transport, "TCP") {
		if !strings.EqualFold(transport.LowerTransport, "TCP") {
			err = errors.New("No TCP found")
			this.errorResponse(request, 461, sessionCtx)
			sessionCtx.TCPConnect.Close()
			return err
		}
		if !strings.EqualFold(transport.CastType, "unicast") {
			err = errors.New("No unicast found")
			this.errorResponse(request, 461, sessionCtx)
			sessionCtx.TCPConnect.Close()
			return err
		}
		if !strings.EqualFold(transport.Interleaved, "0-1") {
			err = errors.New("No 0-1 found")
			this.errorResponse(request, 461, sessionCtx)
			sessionCtx.TCPConnect.Close()
			return err
		}
		sessionCtx.IsUDP = false
		response := RTSPResponse{
			StatusLine: RTSPStatusLine{
				RTSPVersion:  RTSP_VERSION,
				StatusCode:   200,
				ReasonPhrase: getReasonPhrase(200),
			},
			CSeq: request.CSeq,
			ResponseHeader: RTSPResponseHeader{
				Server: utils.MANUFACTURER,
			},
			Transport: request.Transport + ";mode=\"play\"",
			Session:   sessionCtx.SessionId,
		}
		log.Println("[RTSP_SERVER] Response:\n" + response.Marshal())
		sessionCtx.TCPConnect.Write([]byte(response.Marshal()))

		this.initRequest(sessionCtx)

	} else if strings.Contains(request.Transport, "client_port") {
		if len(transport.ClientPort) == 0 {
			err = errors.New("No client_port found")
			this.errorResponse(request, 461, sessionCtx)
			sessionCtx.TCPConnect.Close()
			return err
		}
		sessionCtx.UDPClientPort = transport.ClientPort
		sessionCtx.IsUDP = true

		udpAddr, err := net.ResolveUDPAddr("udp", strings.Split(sessionCtx.TCPConnect.RemoteAddr().String(), ":")[0]+":"+strings.Split(transport.ClientPort, "-")[0])
		if err != nil {
			this.errorResponse(request, 461, sessionCtx)
			sessionCtx.TCPConnect.Close()
			return err
		}
		sessionCtx.UDPConnect, err = net.DialUDP("udp", nil, udpAddr)
		if err != nil {
			this.errorResponse(request, 461, sessionCtx)
			sessionCtx.TCPConnect.Close()
			return err
		}

		lop, err := strconv.Atoi(strings.Split(sessionCtx.UDPConnect.LocalAddr().String(), ":")[1])

		sessionCtx.SSRC = uint32(rand.Int31())

		response := RTSPResponse{
			StatusLine: RTSPStatusLine{
				RTSPVersion:  RTSP_VERSION,
				StatusCode:   200,
				ReasonPhrase: getReasonPhrase(200),
			},
			CSeq: request.CSeq,
			ResponseHeader: RTSPResponseHeader{
				Server: utils.MANUFACTURER,
			},
			Transport: fmt.Sprintf("RTP/AVP;unicast;client_port=%s;server_port=%d-%d;ssrc=%d", sessionCtx.UDPClientPort, lop, lop+1, sessionCtx.SSRC),
			Session:   sessionCtx.SessionId,
		}
		log.Println("[RTSP_SERVER] Response:\n" + response.Marshal())
		sessionCtx.TCPConnect.Write([]byte(response.Marshal()))

		err = this.initRequest(sessionCtx)

	} else {
		err = errors.New("Unsupported transport method")
		this.errorResponse(request, 461, sessionCtx)
		sessionCtx.TCPConnect.Close()
		return err
	}
	return err
}

func (this *RTSPServer) play(request RTSPRequest, sessionCtx *SessionContext) error {
	response := RTSPResponse{
		StatusLine: RTSPStatusLine{
			RTSPVersion:  RTSP_VERSION,
			StatusCode:   200,
			ReasonPhrase: getReasonPhrase(200),
		},
		CSeq: request.CSeq,
		ResponseHeader: RTSPResponseHeader{
			Server: utils.MANUFACTURER,
		},
		Session: sessionCtx.SessionId,
	}

	log.Println("[RTSP_SERVER] Response:\n" + response.Marshal())
	sessionCtx.TCPConnect.Write([]byte(response.Marshal()))

	return nil
}
