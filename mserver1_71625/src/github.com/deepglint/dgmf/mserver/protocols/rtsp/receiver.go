package rtsp

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/deepglint/dgmf/mserver/core"
	"github.com/deepglint/dgmf/mserver/protocols/h264"
	"github.com/deepglint/dgmf/mserver/protocols/rtcp"
	"github.com/deepglint/dgmf/mserver/protocols/rtp"
	"github.com/deepglint/dgmf/mserver/protocols/sdp"
	"github.com/deepglint/dgmf/mserver/utils"
	"github.com/golang/glog"
)

// RTSP client for receiving h264 element stream from remote rtsp server
type RTSPReceiver struct {
	core.Receiver
	authType       string
	authCtx        WWWAuthenticate
	authStr        string
	sdpDescription *sdp.Description
	ssrc           string
	session        string
	optionNames    string
	host           string
	baseURI        string
	absURI         string
	Username       string
	Password       string
	uri            string
	cseq           int
	buf            bytes.Buffer
	connect        net.Conn
	streamId       string
}

func (this *RTSPReceiver) SDP() *sdp.Description {
	return this.sdpDescription
}

func (this *RTSPReceiver) SSRC() string {
	return this.ssrc
}

func (this *RTSPReceiver) Session() string {
	return this.session
}

func (this *RTSPReceiver) OptionNames() string {
	return this.optionNames
}

func (this *RTSPReceiver) setTimeout() {
	if this.connect != nil {
		this.connect.SetDeadline(time.Now().Add(this.AbsTimeout))
	}
}

// Open a client connection for remote rtsp server
// rtsp sequence: options -> describe -> setup -> play -> get_parameter
func (this *RTSPReceiver) Open(uri string, streamId string, rtms chan core.RTMessage) {
	glog.V(3).Infof("[RTSP_RECEIVER] [STREAM_ID]=%s Open rtsp receiver started\n", this.AbsStreamId)
	var err error

	this.AbsRunning = false
	this.AbsFrames = make(chan *core.H264ESFrame)
	this.AbsStoped = make(chan bool)
	this.AbsRTMessages = rtms
	this.AbsStreamId = streamId
	this.AbsIndex = 0

	err = this.openConnect(uri)
	if err != nil {
		this.clear(err)
		return
	}

	err = this.options()
	if err != nil {
		this.clear(err)
		return
	}

	err = this.describe()
	if err != nil {
		this.clear(err)
		return
	}

	err = this.setup()
	if err != nil {
		this.clear(err)
		return
	}

	err = this.play()
	if err != nil {
		this.clear(err)
		return
	}

	if strings.Contains(this.optionNames, GET_PARAMETER) {
		err = this.getParameter()
		if err != nil {
			this.clear(err)
			return
		}
	}

	this.AbsRunning = true
	select {
	case this.AbsRTMessages <- core.RTMessage{Status: 200}:
	default:
	}
	err = this.receiveFrames()
	this.clear(err)
	return
}

func (this *RTSPReceiver) clear(err error) {

	if err != nil {
		select {
		case this.AbsRTMessages <- core.RTMessage{
			Status: 400,
			Error:  err,
		}:
		}
	}

	if this.AbsRunning == true {
		this.teardown()
	}

	if this.connect != nil {
		this.setTimeout()
		this.connect.Close()
	}

	select {
	case this.AbsStoped <- true:
	default:
	}

	this.AbsRunning = false
	select {
	case this.AbsRTMessages <- core.RTMessage{Status: 201}:
	default:
	}

	if _, ok := <-this.AbsStoped; ok == true {
		close(this.AbsStoped)
	}
	if _, ok := <-this.AbsFrames; ok == true {
		close(this.AbsFrames)
	}

	glog.V(3).Infof("[RTSP_RECEIVER] [STREAM_ID]=%s Clear rtsp receiver finished\n", this.AbsStreamId)
}

// Close client connection with remote rtsp server
// rtsp sequence: teardown
func (this *RTSPReceiver) Close() {
	if this.AbsRunning == true {
		this.AbsRunning = false
		for i := 0; i < 10; i++ {
			select {
			case <-this.AbsStoped:
				break
			default:
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
	glog.V(3).Infof("[RTSP_RECEIVER] [STREAM_ID]=%s Close rtsp receiver finished\n", this.AbsStreamId)
}

// Parae rtsp url and create a rtsp tcp connection
// timeout unit is millisecond
func (this *RTSPReceiver) openConnect(uri string) error {

	// Parse rtsp uri scheme
	urlCtx, err := url.Parse(uri)
	if err != nil {
		glog.Warningf("[RTSP_RECEIVER] [STREAM_ID]=%s Open rtsp connect failed, error: %s\n", this.AbsStreamId, err.Error())
		return err
	}
	if !strings.EqualFold(urlCtx.Scheme, "rtsp") {
		err = errors.New(urlCtx.Scheme + " not support")
		glog.Warningf("[RTSP_RECEIVER] [STREAM_ID]=%s Open rtsp connect failed, error: %s\n", this.AbsStreamId, err.Error())
		return err
	}
	this.uri = uri

	// Parse basic auth username and password
	if urlCtx.User != nil {
		this.Username = urlCtx.User.Username()
		this.Password, _ = urlCtx.User.Password()
	}

	// Port 554 is the default port of protocol rtsp
	if strings.Contains(urlCtx.Host, ":") {
		this.host = urlCtx.Host
	} else {
		this.host = urlCtx.Host + ":554"
	}

	this.cseq = 2

	// Create rtsp connection between this client and remote rtsp server with timeout
	this.connect, err = net.DialTimeout("tcp", this.host, this.AbsTimeout*time.Millisecond)
	if err != nil {
		glog.Warningf("[RTSP_RECEIVER] [STREAM_ID]=%s Open rtsp connect failed, error: %s\n", this.AbsStreamId, err.Error())
		return err
	}

	glog.V(3).Infof("[RTSP_RECEIVER] [STREAM_ID]=%s Open rtsp connect successed\n", this.AbsStreamId)
	return nil
}

// RTSP auth method for digest or basic auth
func (this *RTSPReceiver) auth(request RTSPRequest, response RTSPResponse, method string) (RTSPResponse, error) {
	var err error
	this.authCtx.Unmarshal(response.ResponseHeader.WWWAuthenticateDigest)
	authRequest := Authorization{}
	if len(response.ResponseHeader.WWWAuthenticateDigest) > 0 {
		this.authType = "Digest"
		if len(this.authCtx.Realm) == 0 || len(this.authCtx.Nonce) == 0 {
			err = errors.New("Invalid Digest auth request")
			glog.Warningf("[RTSP_RECEIVER] [STREAM_ID]=%s Rtsp auth failed, error: %s\n", this.AbsStreamId, err.Error())
			return RTSPResponse{}, err
		}
		authRequest.DigestAuth(this.authCtx, this.Username, this.Password, this.uri, method)
	} else if len(response.ResponseHeader.WWWAuthenticateBasic) > 0 {
		this.authType = "Basic"
		authRequest.BasicAuth(this.Username, this.Password)
	} else {
		err = errors.New("WWWAuthenticate error in DESCRIBE 401 auth response")
		glog.Warningf("[RTSP_RECEIVER] [STREAM_ID]=%s Rtsp auth failed, error: %s\n", this.AbsStreamId, err.Error())
		return RTSPResponse{}, err
	}
	this.authStr = authRequest.Marshal()
	if len(this.authStr) == 0 {
		err = errors.New("Auth request maarshal error")
		glog.Warningf("[RTSP_RECEIVER] [STREAM_ID]=%s Rtsp auth failed, error: %s\n", this.AbsStreamId, err.Error())
		return RTSPResponse{}, err
	}

	// Resend request with auth
	request.CSeq = this.cseq
	request.RequestHeader.Authorization = this.authStr
	this.cseq++
	glog.V(3).Infof("[RTSP_RECEIVER] [STREAM_ID]=%s %s Request:\n%s", this.AbsStreamId, method, request.Marshal())
	response = RTSPResponse{}
	this.setTimeout()
	_, err = this.connect.Write([]byte(request.Marshal()))
	if err != nil {
		glog.Warningf("[RTSP_RECEIVER] [STREAM_ID]=%s Rtsp auth failed, error: %s\n", this.AbsStreamId, err.Error())
		return RTSPResponse{}, err
	}
	buf := make([]byte, REQ_RSP_SIZE)
	_, err = this.connect.Read(buf)
	if err != nil {
		glog.Warningf("[RTSP_RECEIVER] [STREAM_ID]=%s Rtsp auth failed, error: %s\n", this.AbsStreamId, err.Error())
		return RTSPResponse{}, err
	}
	response.Unmarshal(string(buf))
	glog.V(3).Infof("[RTSP_RECEIVER] [STREAM_ID]=%s %s Response:\n%s", this.AbsStreamId, method, response.Marshal())

	glog.V(3).Infof("[RTSP_RECEIVER] [STREAM_ID]=%s Rtsp auth successed\n", this.AbsStreamId)
	return response, nil
}

/*
 RTSP OPTIONS method for getting rtsp methods, examples:

 1. request:

 OPTIONS rtsp://admin:deepglint@192.168.4.115:554/cam/realmonitor?channel=1&subtype=0 RTSP/1.0
 CSeq: 3
 User-Agent: MServer/0.9.4 (Deep Glint Inc. 2017.1.21)

 2. response:

 RTSP/1.0 200 OK
 CSeq: 3
 Public: OPTIONS, DESCRIBE, SETUP, PLAY, PAUSE, TEARDOWN, SET_PARAMETER, GET_PARAMETER, ANNOUNCE
 Server: Rtsp Server/3.0
*/
func (this *RTSPReceiver) options() error {
	var err error

	// Send OPTIONS request
	requestLine := RTSPRequestLine{
		Method:      OPTIONS,
		RequestURI:  this.uri,
		RTSPVersion: RTSP_VERSION,
	}
	request := RTSPRequest{
		RequestLine: requestLine,
		CSeq:        this.cseq,
		RequestHeader: RTSPRequestHeader{
			UserAgent: utils.MANUFACTURER,
		},
	}
	this.cseq++
	this.setTimeout()
	_, err = this.connect.Write([]byte(request.Marshal()))
	if err != nil {
		glog.Warningf("[RTSP_RECEIVER] [STREAM_ID]=%s Rtsp OPTIONS failed, error: %s\n", this.AbsStreamId, err.Error())
		return nil
	}
	glog.V(3).Infof("[RTSP_RECEIVER] [STREAM_ID]=%s OPTIONS Request:\n%s", this.AbsStreamId, request.Marshal())

	// Receive OPTIONS response
	buf := make([]byte, REQ_RSP_SIZE)
	this.setTimeout()
	_, err = this.connect.Read(buf)
	if err != nil {
		glog.Warningf("[RTSP_RECEIVER] [STREAM_ID]=%s Rtsp OPTIONS failed, error: %s\n", this.AbsStreamId, err.Error())
		return err
	}
	response := RTSPResponse{}
	err = response.Unmarshal(string(buf))
	if err != nil {
		glog.Warningf("[RTSP_RECEIVER] [STREAM_ID]=%s Rtsp OPTIONS failed, error: %s\n", this.AbsStreamId, err.Error())
		return err
	}
	glog.V(3).Infof("[RTSP_RECEIVER] [STREAM_ID]=%s OPTIONS Response:\n%s", this.AbsStreamId, response.Marshal())

	// Need auth
	if response.StatusLine.StatusCode == 401 {
		response, err = this.auth(request, response, OPTIONS)
		if err != nil {
			glog.Warningf("[RTSP_RECEIVER] [STREAM_ID]=%s Rtsp OPTIONS failed, error: %s\n", this.AbsStreamId, err.Error())
			return err
		}
	}

	// Check response status and result
	if response.StatusLine.StatusCode != 200 {
		err = errors.New(fmt.Sprintf("Received an invalid status code: %d", response.StatusLine.StatusCode))
		glog.Warningf("[RTSP_RECEIVER] [STREAM_ID]=%s Rtsp OPTIONS failed, error: %s\n", this.AbsStreamId, err.Error())
		return err
	}
	if len(response.ResponseHeader.Public) == 0 {
		err = errors.New("Received an empty options")
		glog.Warningf("[RTSP_RECEIVER] [STREAM_ID]=%s Rtsp OPTIONS failed, error: %s\n", this.AbsStreamId, err.Error())
		return err
	}

	// Parse option names
	this.optionNames = strings.ToUpper(response.ResponseHeader.Public)

	glog.V(3).Infof("[RTSP_RECEIVER] [STREAM_ID]=%s Rtsp OPTIONS successed, OptionNames=%s\n", this.AbsStreamId, this.optionNames)
	return nil
}

/*
 RTSP OPTIONS method for getting sdp, examples:

 1. request:

 DESCRIBE rtsp://admin:deepglint@192.168.4.115:554/cam/realmonitor?channel=1&subtype=0 RTSP/1.0
 CSeq: 4
 Accept: application/sdp
 Authorization: Basic YWRtaW46ZGVlcGdsaW50
 User-Agent: MServer/0.9.4 (Deep Glint Inc. 2017.1.21)

 2. response:

 RTSP/1.0 200 OK
 CSeq: 4
 Cache-Control: must-revalidate
 Content-Base: rtsp://admin:deepglint@192.168.4.115:554/cam/realmonitor?channel=1&subtype=0/
 Content-Length: 532
 Content-Type: application/sdp

 v=0
 o=RTSP Session 0 0 IN IP4 0.0.0.0
 s=Media Server
 c=IN IP4 0.0.0.0
 t=0 0
 a=control:*
 a=packetization-supported:DH
 a=rtppayload-supported:DH
 a=range:npt=now-
 m=video 0 RTP/AVP 96
 a=control:trackID=0
 a=framerate:25.000000
 a=rtpmap:96 H264/90000
 a=fmtp:96 packetization-mode=1;profile-level-id=640032;sprop-parameter-sets=J2QAMq2EBUViuKxUcQgKisVxWKjiECSFITk8nyfk/k/J8nm5s00IEkKQnJ5Pk/J/J+T5PNzZphcqAeAIn5YQAAA+gAAMNQBA,KP4Jiw==
 a=recvonly
 m=audio 0 RTP/AVP 8
 a=control:trackID=1
 a=rtpmap:8 PCMA/8000
 a=recvonly
*/
func (this *RTSPReceiver) describe() error {
	var err error

	// Check DESCRIBE is supported or not
	if !strings.Contains(this.optionNames, DESCRIBE) {
		err = errors.New("Option " + DESCRIBE + " not support")
		glog.Warningf("[RTSP_RECEIVER] [STREAM_ID]=%s Rtsp DESCRIBE failed, error: %s\n", this.AbsStreamId, err.Error())
		return err
	}

	// Send DESCRIBE request
	requestLine := RTSPRequestLine{
		Method:      DESCRIBE,
		RequestURI:  this.uri,
		RTSPVersion: RTSP_VERSION,
	}
	request := RTSPRequest{
		RequestLine: requestLine,
		CSeq:        this.cseq,
		RequestHeader: RTSPRequestHeader{
			Accept:    "application/sdp",
			UserAgent: utils.MANUFACTURER,
		},
	}
	if len(this.authStr) > 0 {
		request.RequestHeader.Authorization = this.authStr
	}
	this.cseq++
	this.setTimeout()
	_, err = this.connect.Write([]byte(request.Marshal()))
	if err != nil {
		glog.Warningf("[RTSP_RECEIVER] [STREAM_ID]=%s Rtsp DESCRIBE failed, error: %s\n", this.AbsStreamId, err.Error())
		return err
	}
	glog.V(3).Infof("[RTSP_RECEIVER] [STREAM_ID]=%s DESCRIBE Request:\n%s", this.AbsStreamId, request.Marshal())

	// Receive DESCRIBE response
	response := RTSPResponse{}
	buf := make([]byte, REQ_RSP_SIZE)
	this.setTimeout()
	_, err = this.connect.Read(buf)
	if err != nil {
		glog.Warningf("[RTSP_RECEIVER] [STREAM_ID]=%s Rtsp DESCRIBE failed, error: %s\n", this.AbsStreamId, err.Error())
		return err
	}
	err = response.Unmarshal(string(buf))
	if err != nil {
		glog.Warningf("[RTSP_RECEIVER] [STREAM_ID]=%s Rtsp DESCRIBE failed, error: %s\n", this.AbsStreamId, err.Error())
		return err
	}
	glog.V(3).Infof("[RTSP_RECEIVER] [STREAM_ID]=%s DESCRIBE response:\n%s", this.AbsStreamId, response.Marshal())

	// Need auth
	if response.StatusLine.StatusCode == 401 {
		response, err = this.auth(request, response, DESCRIBE)
		if err != nil {
			glog.Warningf("[RTSP_RECEIVER] [STREAM_ID]=%s Rtsp DESCRIBE failed, error: %s\n", this.AbsStreamId, err.Error())
			return err
		}
	}

	// Check response status and result
	if response.StatusLine.StatusCode != 200 {
		err = errors.New(fmt.Sprintf("Received an invalid status code: %d", response.StatusLine.StatusCode))
		glog.Warningf("[RTSP_RECEIVER] [STREAM_ID]=%s Rtsp DESCRIBE failed, error: %s\n", this.AbsStreamId, err.Error())
		return err
	}
	if len(response.MessageBody) == 0 {
		err = errors.New(fmt.Sprintf("Received an empty message body"))
		glog.Warningf("[RTSP_RECEIVER] [STREAM_ID]=%s Rtsp DESCRIBE failed, error: %s\n", this.AbsStreamId, err.Error())
		return err
	}

	// Parse sdp
	this.sdpDescription, err = sdp.ParseSdp(response.MessageBody)

	if err != nil {
		glog.Warningf("[RTSP_RECEIVER] [STREAM_ID]=%s Rtsp DESCRIBE failed, error: %s\n", this.AbsStreamId, err.Error())
		return err
	}
	for _, media := range this.sdpDescription.Media {
		if strings.EqualFold(media.Type, "video") {
			for _, attr := range media.Attributes {
				if strings.EqualFold(attr.Name, "control") {
					this.baseURI = this.uri + "/"
					if strings.Contains(attr.Value, "rtsp://") {
						this.absURI = attr.Value
					} else {
						this.absURI = this.uri + "/" + attr.Value
					}
					break
				}
			}
			break
		}
	}

	if len(this.baseURI) == 0 {
		err = errors.New("Can not find BaseURI in sdp")
		glog.Warningf("[RTSP_RECEIVER] [STREAM_ID]=%s Rtsp DESCRIBE failed, error: %s\n", this.AbsStreamId, err.Error())
		return err
	}
	if len(this.absURI) == 0 {
		err = errors.New("Can not find AbsURI in sdp")
		glog.Warningf("[RTSP_RECEIVER] [STREAM_ID]=%s Rtsp DESCRIBE failed, error: %s\n", this.AbsStreamId, err.Error())
		return err
	}

	glog.V(3).Infof("[RTSP_RECEIVER] [STREAM_ID]=%s Rtsp DESCRIBE successed, BaseURI=%s, AbsURI=%s\n", this.AbsStreamId, this.baseURI, this.absURI)
	return nil
}

/*
 RTSP SETUP method for setting up rtp transport, examples:

 1. request:

 SETUP rtsp://admin:deepglint@192.168.4.115:554/cam/realmonitor?channel=1&subtype=0/trackID=0 RTSP/1.0
 CSeq: 5
 Authorization: Basic YWRtaW46ZGVlcGdsaW50
 User-Agent: MServer/0.9.4 (Deep Glint Inc. 2017.1.21)
 Transport: RTP/AVP/TCP;unicast;interleaved=0-1

 2. response:

 RTSP/1.0 200 OK
 CSeq: 5
 Session: 118178849315;timeout=60
 Transport: RTP/AVP/TCP;unicast;interleaved=0-1;ssrc=06A3AE0B
*/
func (this *RTSPReceiver) setup() error {
	var err error

	// Check SETUP is supported or not
	if !strings.Contains(this.optionNames, SETUP) {
		err = errors.New("Option " + SETUP + " not support")
		glog.Warningf("[RTSP_RECEIVER] [STREAM_ID]=%s Rtsp SETUP failed, error: %s\n", this.AbsStreamId, err.Error())
		return err
	}

	// Send SETUP request
	transport := RTSPTransport{
		CastType:       "unicast",
		LowerTransport: "TCP",
		Interleaved:    "0-1",
	}
	request := RTSPRequest{
		RequestLine: RTSPRequestLine{
			Method:      SETUP,
			RequestURI:  this.absURI,
			RTSPVersion: RTSP_VERSION,
		},
		CSeq: this.cseq,
		RequestHeader: RTSPRequestHeader{
			UserAgent: utils.MANUFACTURER,
		},
		Transport: transport.Marshal(),
	}
	if len(this.authStr) > 0 {
		request.RequestHeader.Authorization = this.authStr
	}
	this.cseq++
	this.setTimeout()
	_, err = this.connect.Write([]byte(request.Marshal()))
	if err != nil {
		glog.Warningf("[RTSP_RECEIVER] [STREAM_ID]=%s Rtsp SETUP failed, error: %s\n", this.AbsStreamId, err.Error())
		return err
	}
	glog.V(3).Infof("[RTSP_RECEIVER] [STREAM_ID]=%s SETUP request:\n%s", this.AbsStreamId, request.Marshal())

	// Receive SETUP response
	response := RTSPResponse{}
	buf := make([]byte, REQ_RSP_SIZE)
	this.setTimeout()
	_, err = this.connect.Read(buf)
	if err != nil {
		glog.Warningf("[RTSP_RECEIVER] [STREAM_ID]=%s Rtsp SETUP failed, error: %s\n", this.AbsStreamId, err.Error())
		return err
	}
	err = response.Unmarshal(string(buf))
	if err != nil {
		glog.Warningf("[RTSP_RECEIVER] [STREAM_ID]=%s Rtsp SETUP failed, error: %s\n", this.AbsStreamId, err.Error())
		return err
	}
	glog.V(3).Infof("[RTSP_RECEIVER] [STREAM_ID]=%s SETUP response:\n%s", this.AbsStreamId, response.Marshal())

	// Need auth
	if response.StatusLine.StatusCode == 401 {
		response, err = this.auth(request, response, SETUP)
		if err != nil {
			glog.Warningf("[RTSP_RECEIVER] [STREAM_ID]=%s Rtsp SETUP failed, error: %s\n", this.AbsStreamId, err.Error())
			return err
		}
	}

	// Check response status and result
	if response.StatusLine.StatusCode != 200 {
		err = errors.New(fmt.Sprintf("Received an invalid status code: %d", response.StatusLine.StatusCode))
		glog.Warningf("[RTSP_RECEIVER] [STREAM_ID]=%s Rtsp SETUP failed, error: %s\n", this.AbsStreamId, err.Error())
		return err
	}

	// Parse rtp SSRC and session
	transport.Unmarshal(response.Transport)
	if len(transport.SSRC) != 0 {
		this.ssrc = transport.SSRC
	}
	if len(response.Session) != 0 {
		this.session = response.Session
	}

	glog.V(3).Infof("[RTSP_RECEIVER] [STREAM_ID]=%s Rtsp SETUP successed, SSRC=%s, Session=%s\n", this.AbsStreamId, this.ssrc, this.session)
	return nil
}

/*
 RTSP PLAY method for setting up rtp transport, examples:

 1. request:

 PLAY rtsp://admin:deepglint@192.168.4.115:554/cam/realmonitor?channel=1&subtype=0/ RTSP/1.0
 CSeq: 6
 Authorization: Basic YWRtaW46ZGVlcGdsaW50
 Range: npt=0.000-
 User-Agent: MServer/0.9.4 (Deep Glint Inc. 2017.1.21)
 Session: 118178849315;timeout=60

 2. response:

 RTSP/1.0 200 OK
 CSeq: 6
 Session: 118178849315
*/
func (this *RTSPReceiver) play() error {
	var err error

	// Check PLAY is supported or not
	if !strings.Contains(this.optionNames, PLAY) {
		err = errors.New("Option " + PLAY + " not support")
		glog.Warningf("[RTSP_RECEIVER] [STREAM_ID]=%s Rtsp PLAY failed, error: %s\n", this.AbsStreamId, err.Error())
		return err
	}

	// Send PLAY request
	request := RTSPRequest{
		RequestLine: RTSPRequestLine{
			Method:      PLAY,
			RequestURI:  this.baseURI,
			RTSPVersion: RTSP_VERSION,
		},
		CSeq: this.cseq,
		RequestHeader: RTSPRequestHeader{
			Range:     "npt=0.000-",
			UserAgent: utils.MANUFACTURER,
		},
		Session: this.session,
	}
	if len(this.authStr) > 0 {
		request.RequestHeader.Authorization = this.authStr
	}
	this.cseq++
	this.setTimeout()
	_, err = this.connect.Write([]byte(request.Marshal()))
	if err != nil {
		glog.Warningf("[RTSP_RECEIVER] [STREAM_ID]=%s Rtsp PLAY failed, error: %s\n", this.AbsStreamId, err.Error())
		return err
	}
	glog.V(3).Infof("[RTSP_RECEIVER] [STREAM_ID]=%s PLAY Request:\n%s", this.AbsStreamId, request.Marshal())

	// Receive PLAY response
	response := RTSPResponse{}
	buf := make([]byte, REQ_RSP_SIZE)
	this.setTimeout()
	_, err = this.connect.Read(buf)
	if err != nil {
		glog.Warningf("[RTSP_RECEIVER] [STREAM_ID]=%s Rtsp PLAY failed, error: %s\n", this.AbsStreamId, err.Error())
		return err
	}
	err = response.Unmarshal(string(buf))
	if err != nil {
		glog.Warningf("[RTSP_RECEIVER] [STREAM_ID]=%s Rtsp PLAY failed, error: %s\n", this.AbsStreamId, err.Error())
		return err
	}
	glog.V(3).Infof("[RTSP_RECEIVER] [STREAM_ID]=%s PLAY response:\n%s", this.AbsStreamId, response.Marshal())

	// Need auth
	if response.StatusLine.StatusCode == 401 {
		response, err = this.auth(request, response, PLAY)
		if err != nil {
			glog.Warningf("[RTSP_RECEIVER] [STREAM_ID]=%s Rtsp PLAY failed, error: %s\n", this.AbsStreamId, err.Error())
			return err
		}
	}

	// Check response status and result
	if response.StatusLine.StatusCode != 200 {
		err = errors.New(fmt.Sprintf("Received an invalid status code: %d", response.StatusLine.StatusCode))
		glog.Warningf("[RTSP_RECEIVER] [STREAM_ID]=%s Rtsp PLAY failed, error: %s\n", this.AbsStreamId, err.Error())
		return err
	}

	glog.V(3).Infof("[RTSP_RECEIVER] [STREAM_ID]=%s Rtsp PLAY successed", this.AbsStreamId)
	return nil
}

/*
 RTSP GET_PARAMETER method for setting up rtp transport, examples:

 1. request:

 GET_PARAMETER rtsp://admin:deepglint@192.168.4.115:554/cam/realmonitor?channel=1&subtype=0/ RTSP/1.0
 CSeq: 7
 Authorization: Basic YWRtaW46ZGVlcGdsaW50
 User-Agent: MServer/0.9.4 (Deep Glint Inc. 2017.1.21)
 Session: 119488796037;timeout=60
*/
func (this *RTSPReceiver) getParameter() error {
	var err error

	// Check GET_PARAMETER is supported or not
	if !strings.Contains(this.optionNames, GET_PARAMETER) {
		err = errors.New("Option " + GET_PARAMETER + " not support")
		glog.Warningf("[RTSP_RECEIVER] [STREAM_ID]=%s Rtsp GET_PARAMETER failed, error: %s\n", this.AbsStreamId, err.Error())
		return err
	}

	// Send GET_PARAMETER request
	request := RTSPRequest{
		RequestLine: RTSPRequestLine{
			Method:      GET_PARAMETER,
			RequestURI:  this.baseURI,
			RTSPVersion: RTSP_VERSION,
		},
		CSeq: this.cseq,
		RequestHeader: RTSPRequestHeader{
			UserAgent: utils.MANUFACTURER,
		},
		Session: this.session,
	}
	if len(this.authStr) > 0 {
		request.RequestHeader.Authorization = this.authStr
	}
	this.cseq++
	this.setTimeout()
	_, err = this.connect.Write([]byte(request.Marshal()))
	if err != nil {
		glog.Warningf("[RTSP_RECEIVER] [STREAM_ID]=%s Rtsp GET_PARAMETER failed, error: %s\n", this.AbsStreamId, err.Error())
		return err
	}
	glog.V(3).Infof("[RTSP_RECEIVER] [STREAM_ID]=%s GET_PARAMETER Request:\n%s", this.AbsStreamId, request.Marshal())

	glog.V(3).Infof("[RTSP_RECEIVER] [STREAM_ID]=%s Rtsp GET_PARAMETER successed", this.AbsStreamId)
	return nil
}

/*
 RTSP GET_PARAMETER method for setting up rtp transport, examples:

 1. request:

 TEARDOWN rtsp://127.0.0.1:8554/test/ RTSP/1.0
 CSeq: 6
 User-Agent: LibVLC/2.2.4 (LIVE555 Streaming Media v2016.02.22)
*/
func (this *RTSPReceiver) teardown() error {
	var err error

	// Check TEARDOWN is supported or not
	if !strings.Contains(this.optionNames, TEARDOWN) {
		err = errors.New("Option " + TEARDOWN + " not support")
		glog.Warningf("[RTSP_RECEIVER] [STREAM_ID]=%s Rtsp TEARDOWN failed, error: %s\n", this.AbsStreamId, err.Error())
		return err
	}

	// Send TEARDOWN request
	request := RTSPRequest{
		RequestLine: RTSPRequestLine{
			Method:      TEARDOWN,
			RequestURI:  this.baseURI,
			RTSPVersion: RTSP_VERSION,
		},
		CSeq: this.cseq,
		RequestHeader: RTSPRequestHeader{
			UserAgent: utils.MANUFACTURER,
		},
		Session: this.session,
	}
	this.cseq++
	this.setTimeout()
	_, err = this.connect.Write([]byte(request.Marshal()))
	if err != nil {
		glog.Warningf("[RTSP_RECEIVER] [STREAM_ID]=%s Rtsp TEARDOWN failed, error: %s\n", this.AbsStreamId, err.Error())
		return err
	}
	glog.V(3).Infof("[RTSP_RECEIVER] [STREAM_ID]=%s TEARDOWN Request:\n%s", this.AbsStreamId, request.Marshal())

	glog.V(3).Infof("[RTSP_RECEIVER] [STREAM_ID]=%s Rtsp TEARDOWN successed", this.AbsStreamId)

	return nil
}

// Receive rtp h264 packet with rtsp tcp
// Not support rtsp udp rtp data now
func (this *RTSPReceiver) receiveFrames() error {
	var err error
	var inv RTSPInterleavedFrame
	reader := bufio.NewReader(this.connect)
	var h264Nalu rtp.H264NALUPacket
	var t0 time.Time
	var fps uint32 = 30
	var index uint64 = 0

	for this.AbsRunning {
		// Read rtsp interleaved frame
		if inv.Length == 0 {
			this.setTimeout()
			buf, err := reader.Peek(4)
			if err != nil {
				fmt.Println("*******100", err.Error())
				if strings.Contains(err.Error(), "timeout") ||
					strings.Contains(err.Error(), "EOS") ||
					strings.Contains(err.Error(), "EOF") {
					return err
				} else {
					continue
				}
			}
			err = inv.Unmarshal(buf)
			if err != nil {
				this.setTimeout()
				_, err = reader.ReadByte()
				if err != nil {
					fmt.Println("*******101", err.Error())
					if strings.Contains(err.Error(), "timeout") ||
						strings.Contains(err.Error(), "EOS") ||
						strings.Contains(err.Error(), "EOF") {
						return err
					} else {
						continue
					}
				}
				continue
			}
			this.setTimeout()
			_, err = reader.Read(buf)
			if err != nil {
				fmt.Println("*******102", err.Error())
				if strings.Contains(err.Error(), "timeout") ||
					strings.Contains(err.Error(), "EOS") ||
					strings.Contains(err.Error(), "EOF") {
					return err
				} else {
					continue
				}
			}
		}

		if inv.Channel == 0x01 {

			// Read RTCP packet
			buf := make([]byte, inv.Length)

			// Receive sender report
			this.setTimeout()
			_, err = reader.Read(buf)
			if err != nil {
				fmt.Println("*******103", err.Error())
				if strings.Contains(err.Error(), "timeout") ||
					strings.Contains(err.Error(), "EOS") ||
					strings.Contains(err.Error(), "EOF") {
					return err
				} else {
					continue
				}
			}
			senderReport := rtcp.SenderReport{}
			senderReport.Unmarshal(buf)
			receiverReport := rtcp.ReceiverReport{
				Version:     2,
				Padding:     false,
				ReportCount: 1,
				PacketType:  201,
				SSRC:        0x1234,
				Blocks: []rtcp.ReportBlock{
					rtcp.ReportBlock{
						SSRC:           senderReport.SSRC,
						FractionLost:   0xFF,
						CumulativeLost: 0xFFFFFF,
					},
				},
			}

			// Send receiver report
			var newinv RTSPInterleavedFrame
			newinv.Channel = 0x01
			newinv.Length = 52
			send := newinv.Marshal()
			tmp, err := receiverReport.Marshal()
			if err != nil {
				fmt.Println("*******104", err.Error())
				if strings.Contains(err.Error(), "timeout") ||
					strings.Contains(err.Error(), "EOS") ||
					strings.Contains(err.Error(), "EOF") {
					return err
				} else {
					continue
				}
			}
			send = append(send, tmp...)
			send = append(send, buf[senderReport.Length*4+1:]...)
			this.connect.Write(send)

		} else if inv.Channel == 0x00 {

			// Read RTP packet
			all := int(inv.Length)

			// Read until get full packet
			for {
				buf := make([]byte, all)
				this.setTimeout()
				size, err := reader.Read(buf)
				if err != nil {
					fmt.Println("*******105", err.Error())
					if strings.Contains(err.Error(), "timeout") ||
						strings.Contains(err.Error(), "EOS") ||
						strings.Contains(err.Error(), "EOF") {
						return err
					} else {
						break
					}
				}
				this.buf.Write(buf[0:size])
				all -= size

				if all <= 0 {
					payload := make([]byte, inv.Length)
					_, err = this.buf.Read(payload)
					if err != nil {
						fmt.Println("*******106", err.Error())
						if strings.Contains(err.Error(), "timeout") ||
							strings.Contains(err.Error(), "EOS") ||
							strings.Contains(err.Error(), "EOF") {
							return err
						} else {
							break
						}
					}

					var packet rtp.RTPPacket
					packet.Unmarshal(payload)

					var rtpNalu rtp.RTPNALUPacket
					err = rtpNalu.Unmarshal(packet.Payload)
					if err != nil {
						fmt.Println("*******107", err.Error())
						if strings.Contains(err.Error(), "timeout") ||
							strings.Contains(err.Error(), "EOS") ||
							strings.Contains(err.Error(), "EOF") {
							return err
						} else {
							break
						}
					}

					// Get rtp h264 payload NAL unit
					if int(rtpNalu.RTPNALUHeader.Type) >= 1 && int(rtpNalu.RTPNALUHeader.Type) <= 23 {
						h264Nalu.Init(rtpNalu)
						data, err := h264Nalu.Marshal()
						if err != nil {
							fmt.Println("*******108", err.Error())
							if strings.Contains(err.Error(), "timeout") ||
								strings.Contains(err.Error(), "EOS") ||
								strings.Contains(err.Error(), "EOF") {
								return err
							} else {
								break
							}
						}
						index++
						iFrame := false
						if data[4]&0x1F == 0x07 || data[4]&0x1F == 0x08 || data[4]&0x1F == 0x05 {
							iFrame = true
						}
						frame := &core.H264ESFrame{
							Data:      data,
							Timestamp: packet.Header.Timestamp,
							IFrame:    iFrame,
							Index:     index,
							// Fps:       fps,
						}
						this.AbsFPS = fps
						this.packNalu(frame)
					}

					// Get rtp h264 payload FU-A
					if int(rtpNalu.RTPNALUHeader.Type) == 28 && rtpNalu.RTPFUHeader.S && !rtpNalu.RTPFUHeader.E {
						h264Nalu.Init(rtpNalu)
					}
					if int(rtpNalu.RTPNALUHeader.Type) == 28 && !rtpNalu.RTPFUHeader.S && !rtpNalu.RTPFUHeader.E {
						h264Nalu.Add(rtpNalu)
					}
					if int(rtpNalu.RTPNALUHeader.Type) == 28 && !rtpNalu.RTPFUHeader.S && rtpNalu.RTPFUHeader.E {
						h264Nalu.Add(rtpNalu)
						data, err := h264Nalu.Marshal()
						if err != nil {
							fmt.Println("*******109", err.Error())
							if strings.Contains(err.Error(), "timeout") ||
								strings.Contains(err.Error(), "EOS") ||
								strings.Contains(err.Error(), "EOF") {
								return err
							} else {
								break
							}
						}
						index++
						iFrame := false
						if len(data) > 4 && (data[4]&0x1F == 0x07 || data[4]&0x1F == 0x08 || data[4]&0x1F == 0x05) {
							iFrame = true
						}
						if index%500 == 0 {
							f, _ := strconv.ParseFloat(fmt.Sprintf("%0.2f", 1.0/(float64(time.Now().Sub(t0).Nanoseconds())/1000000000.00)), 64)
							if f*500 > 0 {
								fps = uint32(f * 500)
							}
							t0 = time.Now()
						}
						frame := &core.H264ESFrame{
							Data:      data,
							Timestamp: packet.Header.Timestamp,
							IFrame:    iFrame,
							Index:     index,
							// Fps:       fps,
						}
						this.AbsFPS = fps
						err = this.packNalu(frame)
						if err != nil {
							return err
						}
					}
					break
				}
			}
		}
		inv.Length = 0
	}

	fmt.Println("*******110 nil")
	return nil
}

// Merge SPS, PPS and I frame to IDR frame
func (this *RTSPReceiver) packNalu(frame *core.H264ESFrame) error {
	if frame.Data[4]&0x1F == 0x07 {
		sp := h264.GetLiveSPS(frame.Data)
		if sp.Width != 0 {
			this.AbsWidth = sp.Width
		}
		if sp.Height != 0 {
			this.AbsHeight = sp.Height
		}
		if len(sp.SPS) != 0 {
			this.AbsSPS = sp.SPS
		}
		if len(sp.PPS) != 0 {
			// frame.PPS = sp.PPS
			this.AbsPPS = sp.PPS
		}
	}

	if frame.Data[4]&0x1F == 0x08 {
		sp := h264.GetLivePPS(frame.Data)
		if sp.Width != 0 {
			this.AbsWidth = sp.Width
		}
		if sp.Height != 0 {
			this.AbsHeight = sp.Height
		}
		if len(sp.SPS) != 0 {
			this.AbsSPS = sp.SPS
		}
		if len(sp.PPS) != 0 {
			this.AbsPPS = sp.PPS
		}
	}

	select {
	case _, ok := <-this.AbsStoped:
		if ok == false {
			return errors.New("Receiver has been stopped")
		}
	default:
		this.AbsIndex++
		this.AbsFrames <- frame
	}
	return nil
}
