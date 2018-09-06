package rtmp

import (
	"bufio"
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/deepglint/dgmf/mserver/core"
	"github.com/deepglint/dgmf/mserver/protocols/rtmp/av"
	"github.com/deepglint/dgmf/mserver/protocols/rtmp/codec/h264parser"
	"github.com/deepglint/dgmf/mserver/protocols/rtmp/flv"
	"github.com/deepglint/dgmf/mserver/protocols/rtsp"
	"github.com/deepglint/dgmf/mserver/utils/bits/pio"
)

const (
	stageHandshakeDone = iota + 1
	stageCommandDone
	stageCodecDataDone
)

const (
	prepareReading = iota + 1
	prepareWriting
)

func (this *RTMPServer) initRequest(sessionCtx *SessionContext) error {
	var err error

	var streams []av.CodecData
	var sps []byte
	var pps []byte

	sessionCtx.TCPConnect.isserver = true
	err = sessionCtx.TCPConnect.prepare(stageCommandDone, 0)
	if err != nil {
		sessionCtx.TCPConnect.Close()
		return err
	}

	urlCtx := sessionCtx.TCPConnect.URL
	if len(urlCtx.Path) < 2 {
		err = errors.New("rtmp url is invalid")
		sessionCtx.TCPConnect.Close()
		return err
	}

	parts := strings.Split(urlCtx.Path[1:], "/")
	if len(parts) != 2 {
		err = errors.New("rtmp url is invalid")
		sessionCtx.TCPConnect.Close()
		return err
	}

	if !strings.EqualFold(parts[0], "live") &&
		!strings.EqualFold(parts[0], "vod") &&
		!strings.EqualFold(parts[0], "proxy") {

		err = errors.New("Unsupported rtmp type")
		sessionCtx.TCPConnect.Close()
		return err
	}
	sessionCtx.StreamType = strings.ToLower(parts[0])
	streamId := strings.Split(strings.ToLower(urlCtx.Path[1:]), "/")[1]

	pool := core.GetESPool()

	if sessionCtx.StreamType == "live" {
		if !pool.Live.ExistOutput(streamId, "rtmp") {
			err = errors.New("No StreamId found")
			sessionCtx.TCPConnect.Close()
			return err
		}

		sessionCtx.StreamId = streamId
		spsStr, err := pool.Live.GetSPS(sessionCtx.StreamId)
		ppsStr, err := pool.Live.GetSPS(sessionCtx.StreamId)
		sps, err = base64.StdEncoding.DecodeString(spsStr)
		pps, err = base64.StdEncoding.DecodeString(ppsStr)

		if err != nil || len(sps) == 0 || len(pps) == 0 {
			err = errors.New("No sps or pps found")
			sessionCtx.TCPConnect.Close()
			return err
		}

	} else if sessionCtx.StreamType == "vod" {

	} else if sessionCtx.StreamType == "proxy" {
		if uris, ok := urlCtx.Query()["uri"]; ok && len(uris) == 1 && len(uris[0]) > 0 {
			uri := uris[0]
			sessionCtx.proxyURI = uri
			sessionCtx.StreamId = streamId
			sessionCtx.proxyReceiver = &rtsp.RTSPReceiver{}
			sessionCtx.proxyReceiver.SetTimeout(20 * time.Second)
			sessionCtx.RTMS = make(chan core.RTMessage)
			go sessionCtx.proxyReceiver.Open(sessionCtx.proxyURI, sessionCtx.StreamId, sessionCtx.RTMS)

			rtm := <-sessionCtx.RTMS
			if rtm.Status != 200 {
				sessionCtx.TCPConnect.Close()
				return rtm.Error
			}
			if pool.Proxy.ExistStream(sessionCtx.StreamId) {
				sessionCtx.TCPConnect.Close()
				return errors.New("StreamId existed")
			}

			pool.Proxy.AddStream(sessionCtx.StreamId, sessionCtx.proxyURI, urlCtx.String(), "rtmp", sessionCtx.proxyReceiver, sessionCtx.TCPConnect.NetConn())
			frames, err := pool.Proxy.GetFrames(sessionCtx.StreamId)
			if err != nil {
				sessionCtx.TCPConnect.Close()
				pool.Proxy.RemoveStream(sessionCtx.StreamId)
				return err
			}
			var spsStr, ppsStr string
			for i := 0; i < 300; i++ {
				_, ok := <-frames
				if !ok {
					sessionCtx.TCPConnect.Close()
					pool.Proxy.RemoveStream(sessionCtx.StreamId)
					return errors.New("frame closed")
				}
				if len(sessionCtx.proxyReceiver.SPS()) != 0 {
					spsStr = sessionCtx.proxyReceiver.SPS()
				}
				if len(sessionCtx.proxyReceiver.PPS()) != 0 {
					ppsStr = sessionCtx.proxyReceiver.PPS()
				}
				if len(spsStr) != 0 && len(ppsStr) != 0 {
					break
				}
			}
			if len(spsStr) == 0 || len(ppsStr) == 0 {
				sessionCtx.TCPConnect.Close()
				pool.Proxy.RemoveStream(sessionCtx.StreamId)
				return errors.New("No sps or pps found")
			}
			sps, err = base64.StdEncoding.DecodeString(spsStr)
			pps, err = base64.StdEncoding.DecodeString(ppsStr)

		} else {
			err = errors.New("rtmp proxy uri error")
			sessionCtx.TCPConnect.Close()
			return err
		}
	}

	stream, err := h264parser.NewCodecDataFromSPSAndPPS(sps, pps)
	if err != nil {
		sessionCtx.TCPConnect.Close()
		return err
	}
	streams = append(streams, stream)
	sessionCtx.TCPConnect.WriteHeader(streams)

	return err
}

type Conn struct {
	URL             *url.URL
	OnPlayOrPublish func(string, flv.AMFMap) error

	prober  *flv.Prober
	streams []av.CodecData

	txbytes uint64
	rxbytes uint64

	bufr *bufio.Reader
	bufw *bufio.Writer
	ackn uint32

	writebuf []byte
	readbuf  []byte

	netconn   net.Conn
	txrxcount *txrxcount

	writeMaxChunkSize int
	readMaxChunkSize  int
	readAckSize       uint32
	readcsmap         map[uint32]*chunkStream

	isserver            bool
	publishing, playing bool
	reading, writing    bool
	stage               int

	avmsgsid uint32

	gotcommand     bool
	commandname    string
	commandtransid float64
	commandobj     flv.AMFMap
	commandparams  []interface{}

	gotmsg      bool
	timestamp   uint32
	msgdata     []byte
	msgtypeid   uint8
	datamsgvals []interface{}
	avtag       flv.Tag

	eventtype uint16
}

type txrxcount struct {
	io.ReadWriter
	txbytes uint64
	rxbytes uint64
}

func (this *txrxcount) Read(p []byte) (int, error) {
	n, err := this.ReadWriter.Read(p)
	this.rxbytes += uint64(n)
	return n, err
}

func (this *txrxcount) Write(p []byte) (int, error) {
	n, err := this.ReadWriter.Write(p)
	this.txbytes += uint64(n)
	return n, err
}

func NewConn(netconn net.Conn) *Conn {
	conn := &Conn{}
	conn.prober = &flv.Prober{}
	conn.netconn = netconn
	conn.readcsmap = make(map[uint32]*chunkStream)
	conn.readMaxChunkSize = 128
	conn.writeMaxChunkSize = 128
	conn.bufr = bufio.NewReaderSize(netconn, pio.RecommendBufioSize)
	conn.bufw = bufio.NewWriterSize(netconn, pio.RecommendBufioSize)
	conn.txrxcount = &txrxcount{ReadWriter: netconn}
	conn.writebuf = make([]byte, 4096)
	conn.readbuf = make([]byte, 4096)
	return conn
}

type chunkStream struct {
	timenow     uint32
	timedelta   uint32
	hastimeext  bool
	msgsid      uint32
	msgtypeid   uint8
	msgdatalen  uint32
	msgdataleft uint32
	msghdrtype  uint8
	msgdata     []byte
}

func (this *chunkStream) Start() {
	this.msgdataleft = this.msgdatalen
	this.msgdata = make([]byte, this.msgdatalen)
}

const (
	msgtypeidUserControl      = 4
	msgtypeidAck              = 3
	msgtypeidWindowAckSize    = 5
	msgtypeidSetPeerBandwidth = 6
	msgtypeidSetChunkSize     = 1
	msgtypeidCommandMsgAMF0   = 20
	msgtypeidCommandMsgAMF3   = 17
	msgtypeidDataMsgAMF0      = 18
	msgtypeidDataMsgAMF3      = 15
	msgtypeidVideoMsg         = 9
	msgtypeidAudioMsg         = 8
)

const (
	eventtypeStreamBegin      = 0
	eventtypeSetBufferLength  = 3
	eventtypeStreamIsRecorded = 4
)

func (this *Conn) NetConn() net.Conn {
	return this.netconn
}

func (this *Conn) TxBytes() uint64 {
	return this.txrxcount.txbytes
}

func (this *Conn) RxBytes() uint64 {
	return this.txrxcount.rxbytes
}

func (this *Conn) Close() (err error) {
	return this.netconn.Close()
}

func (this *Conn) pollCommand() (err error) {
	for {
		if err = this.pollMsg(); err != nil {
			return
		}
		if this.gotcommand {
			return
		}
	}
}

func (this *Conn) pollAVTag() (tag flv.Tag, err error) {
	for {
		if err = this.pollMsg(); err != nil {
			return
		}
		switch this.msgtypeid {
		case msgtypeidVideoMsg, msgtypeidAudioMsg:
			tag = this.avtag
			return
		}
	}
}

func (this *Conn) pollMsg() (err error) {
	this.gotmsg = false
	this.gotcommand = false
	this.datamsgvals = nil
	this.avtag = flv.Tag{}
	for {
		if err = this.readChunk(); err != nil {
			return
		}
		if this.gotmsg {
			return
		}
	}
}

func SplitPath(u *url.URL) (app, stream string) {
	pathsegs := strings.SplitN(u.RequestURI(), "/", 3)
	if len(pathsegs) > 1 {
		app = pathsegs[1]
	}
	if len(pathsegs) > 2 {
		stream = pathsegs[2]
	}
	return
}

func getTcUrl(u *url.URL) string {
	app, _ := SplitPath(u)
	nu := *u
	nu.Path = "/" + app
	return nu.String()
}

func createURL(tcurl, app, play string) (u *url.URL) {
	tcurl = strings.Split(tcurl, "?")[0]
	app = strings.Split(app, "?")[0]
	play = strings.Replace(play, "&", "%26", -1)
	path := "/" + app + "/" + play

	u, _ = url.ParseRequestURI(path)
	if tcurl != "" {
		tu, _ := url.Parse(tcurl)
		if tu != nil {
			u.Host = tu.Host
			u.Scheme = tu.Scheme
		}
	}

	return
}

var CodecTypes = flv.CodecTypes

func (this *Conn) writeBasicConf() (err error) {
	// > SetChunkSize
	if err = this.writeSetChunkSize(1024 * 1024 * 128); err != nil {
		return
	}
	// > WindowAckSize
	if err = this.writeWindowAckSize(5000000); err != nil {
		return
	}
	// > SetPeerBandwidth
	if err = this.writeSetPeerBandwidth(5000000, 2); err != nil {
		return
	}
	return
}

func (this *Conn) readConnect() (err error) {
	var connectpath string

	// < connect("app")
	if err = this.pollCommand(); err != nil {
		return
	}
	if this.commandname != "connect" {
		err = fmt.Errorf("rtmp: first command is not connect")
		return
	}
	if this.commandobj == nil {
		err = fmt.Errorf("rtmp: connect command params invalid")
		return
	}

	var ok bool
	var _app, _tcurl interface{}
	if _app, ok = this.commandobj["app"]; !ok {
		err = fmt.Errorf("rtmp: `connect` params missing `app`")
		return
	}
	connectpath, _ = _app.(string)

	var tcurl string
	if _tcurl, ok = this.commandobj["tcUrl"]; !ok {
		_tcurl, ok = this.commandobj["tcurl"]
	}
	if ok {
		tcurl, _ = _tcurl.(string)
	}
	connectparams := this.commandobj

	if err = this.writeBasicConf(); err != nil {
		return
	}

	// > _result("NetConnection.Connect.Success")
	if err = this.writeCommandMsg(3, 0, "_result", this.commandtransid,
		flv.AMFMap{
			"fmtVer":       "FMS/3,0,1,123",
			"capabilities": 31,
		},
		flv.AMFMap{
			"level":          "status",
			"code":           "NetConnection.Connect.Success",
			"description":    "Connection succeeded.",
			"objectEncoding": 3,
		},
	); err != nil {
		return
	}

	if err = this.flushWrite(); err != nil {
		return
	}

	for {
		if err = this.pollMsg(); err != nil {
			return
		}
		if this.gotcommand {
			switch this.commandname {

			// < createStream
			case "createStream":
				this.avmsgsid = uint32(1)
				// > _result(streamid)
				if err = this.writeCommandMsg(3, 0, "_result", this.commandtransid, nil, this.avmsgsid); err != nil {
					return
				}
				if err = this.flushWrite(); err != nil {
					return
				}

			// < publish("path")
			case "publish":

				if len(this.commandparams) < 1 {
					err = fmt.Errorf("rtmp: publish params invalid")
					return
				}
				publishpath, _ := this.commandparams[0].(string)

				var cberr error
				if this.OnPlayOrPublish != nil {
					cberr = this.OnPlayOrPublish(this.commandname, connectparams)
				}

				// > onStatus()
				if err = this.writeCommandMsg(5, this.avmsgsid,
					"onStatus", this.commandtransid, nil,
					flv.AMFMap{
						"level":       "status",
						"code":        "NetStream.Publish.Start",
						"description": "Start publishing",
					},
				); err != nil {
					return
				}
				if err = this.flushWrite(); err != nil {
					return
				}

				if cberr != nil {
					err = fmt.Errorf("rtmp: OnPlayOrPublish check failed")
					return
				}

				this.URL = createURL(tcurl, connectpath, publishpath)
				this.publishing = true
				this.reading = true
				this.stage++
				return

			// < play("path")
			case "play":
				if len(this.commandparams) < 1 {
					err = fmt.Errorf("rtmp: command play params invalid")
					return
				}
				playpath, _ := this.commandparams[0].(string)

				// > streamBegin(streamid)
				if err = this.writeStreamBegin(this.avmsgsid); err != nil {
					return
				}

				// > onStatus()
				if err = this.writeCommandMsg(5, this.avmsgsid,
					"onStatus", this.commandtransid, nil,
					flv.AMFMap{
						"level":       "status",
						"code":        "NetStream.Play.Start",
						"description": "Start live",
					},
				); err != nil {
					return
				}

				// > |RtmpSampleAccess()
				if err = this.writeDataMsg(5, this.avmsgsid,
					"|RtmpSampleAccess", true, true,
				); err != nil {
					return
				}

				if err = this.flushWrite(); err != nil {
					return
				}

				this.URL = createURL(tcurl, connectpath, playpath)
				this.playing = true
				this.writing = true
				this.stage++
				return
			}

		}
	}

	return
}

func (this *Conn) checkConnectResult() (ok bool, errmsg string) {
	if len(this.commandparams) < 1 {
		errmsg = "params length < 1"
		return
	}

	obj, _ := this.commandparams[0].(flv.AMFMap)
	if obj == nil {
		errmsg = "params[0] not object"
		return
	}

	_code, _ := obj["code"]
	if _code == nil {
		errmsg = "code invalid"
		return
	}

	code, _ := _code.(string)
	if code != "NetConnection.Connect.Success" {
		errmsg = "code != NetConnection.Connect.Success"
		return
	}

	ok = true
	return
}

func (this *Conn) checkCreateStreamResult() (ok bool, avmsgsid uint32) {
	if len(this.commandparams) < 1 {
		return
	}

	ok = true
	_avmsgsid, _ := this.commandparams[0].(float64)
	avmsgsid = uint32(_avmsgsid)
	return
}

func (this *Conn) probe() (err error) {
	for !this.prober.Probed() {
		var tag flv.Tag
		if tag, err = this.pollAVTag(); err != nil {
			return
		}
		if err = this.prober.PushTag(tag, int32(this.timestamp)); err != nil {
			return
		}
	}

	this.streams = this.prober.Streams
	this.stage++
	return
}

func (this *Conn) writeConnect(path string) (err error) {
	if err = this.writeBasicConf(); err != nil {
		return
	}

	if err = this.writeCommandMsg(3, 0, "connect", 1,
		flv.AMFMap{
			"app":           path,
			"flashVer":      "MAC 22,0,0,192",
			"tcUrl":         getTcUrl(this.URL),
			"fpad":          false,
			"capabilities":  15,
			"audioCodecs":   4071,
			"videoCodecs":   252,
			"videoFunction": 1,
		},
	); err != nil {
		return
	}

	if err = this.flushWrite(); err != nil {
		return
	}

	for {
		if err = this.pollMsg(); err != nil {
			return
		}
		if this.gotcommand {
			// < _result("NetConnection.Connect.Success")
			if this.commandname == "_result" {
				var ok bool
				var errmsg string
				if ok, errmsg = this.checkConnectResult(); !ok {
					err = fmt.Errorf("rtmp: command connect failed: %s", errmsg)
					return
				}

				break
			}
		} else {
			if this.msgtypeid == msgtypeidWindowAckSize {
				if len(this.msgdata) == 4 {
					this.readAckSize = pio.U32BE(this.msgdata)
				}
				if err = this.writeWindowAckSize(0xffffffff); err != nil {
					return
				}
			}
		}
	}

	return
}

func (this *Conn) connectPublish() (err error) {
	connectpath, publishpath := SplitPath(this.URL)

	if err = this.writeConnect(connectpath); err != nil {
		return
	}

	transid := 2

	if err = this.writeCommandMsg(3, 0, "createStream", transid, nil); err != nil {
		return
	}
	transid++

	if err = this.flushWrite(); err != nil {
		return
	}

	for {
		if err = this.pollMsg(); err != nil {
			return
		}
		if this.gotcommand {
			// < _result(avmsgsid) of createStream
			if this.commandname == "_result" {
				var ok bool
				if ok, this.avmsgsid = this.checkCreateStreamResult(); !ok {
					err = fmt.Errorf("rtmp: createStream command failed")
					return
				}
				break
			}
		}
	}

	// > publish('app')

	if err = this.writeCommandMsg(8, this.avmsgsid, "publish", transid, nil, publishpath); err != nil {
		return
	}
	transid++

	if err = this.flushWrite(); err != nil {
		return
	}

	this.writing = true
	this.publishing = true
	this.stage++
	return
}

func (this *Conn) connectPlay() (err error) {
	connectpath, playpath := SplitPath(this.URL)
	if err = this.writeConnect(connectpath); err != nil {
		return
	}

	if err = this.writeCommandMsg(3, 0, "createStream", 2, nil); err != nil {
		return
	}

	// > SetBufferLength 0,100ms
	if err = this.writeSetBufferLength(0, 100); err != nil {
		return
	}

	if err = this.flushWrite(); err != nil {
		return
	}

	for {
		if err = this.pollMsg(); err != nil {
			return
		}
		if this.gotcommand {
			// < _result(avmsgsid) of createStream
			if this.commandname == "_result" {
				var ok bool
				if ok, this.avmsgsid = this.checkCreateStreamResult(); !ok {
					err = fmt.Errorf("rtmp: createStream command failed")
					return
				}
				break
			}
		}
	}

	if err = this.writeCommandMsg(8, this.avmsgsid, "play", 0, nil, playpath); err != nil {
		return
	}
	if err = this.flushWrite(); err != nil {
		return
	}

	this.reading = true
	this.playing = true
	this.stage++
	return
}

func (this *Conn) ReadPacket() (pkt av.Packet, err error) {
	if err = this.prepare(stageCodecDataDone, prepareReading); err != nil {
		return
	}

	if !this.prober.Empty() {
		pkt = this.prober.PopPacket()
		return
	}

	for {
		var tag flv.Tag
		if tag, err = this.pollAVTag(); err != nil {
			return
		}

		var ok bool
		if pkt, ok = this.prober.TagToPacket(tag, int32(this.timestamp)); ok {
			return
		}
	}

	return
}

func (this *Conn) Prepare() (err error) {
	return this.prepare(stageCommandDone, 0)
}

func (this *Conn) prepare(stage int, flags int) (err error) {
	for this.stage < stage {
		switch this.stage {
		case 0:
			if this.isserver {
				if err = this.handshakeServer(); err != nil {
					return
				}
			} else {
				if err = this.handshakeClient(); err != nil {
					return
				}
			}

		case stageHandshakeDone:
			if this.isserver {
				if err = this.readConnect(); err != nil {
					return
				}
			} else {
				if flags == prepareReading {
					if err = this.connectPlay(); err != nil {
						return
					}
				} else {
					if err = this.connectPublish(); err != nil {
						return
					}
				}
			}

		case stageCommandDone:
			if flags == prepareReading {
				if err = this.probe(); err != nil {
					return
				}
			} else {
				err = fmt.Errorf("rtmp: call WriteHeader() before WritePacket()")
				return
			}
		}
	}
	return
}

func (this *Conn) Streams() (streams []av.CodecData, err error) {
	if err = this.prepare(stageCodecDataDone, prepareReading); err != nil {
		return
	}
	streams = this.streams
	return
}

func (this *Conn) WritePacket(pkt av.Packet) (err error) {
	if err = this.prepare(stageCodecDataDone, prepareWriting); err != nil {
		return
	}

	stream := this.streams[pkt.Idx]
	tag, timestamp := flv.PacketToTag(pkt, stream)

	if err = this.writeAVTag(tag, int32(timestamp)); err != nil {
		return
	}

	return
}

func (this *Conn) WriteTrailer() (err error) {
	if err = this.flushWrite(); err != nil {
		return
	}
	return
}

func (this *Conn) WriteHeader(streams []av.CodecData) (err error) {
	if err = this.prepare(stageCommandDone, prepareWriting); err != nil {
		return
	}

	var metadata flv.AMFMap
	if metadata, err = flv.NewMetadataByStreams(streams); err != nil {
		return
	}

	// > onMetaData()
	if err = this.writeDataMsg(5, this.avmsgsid, "onMetaData", metadata); err != nil {
		return
	}

	// > Videodata(decoder config)
	// > Audiodata(decoder config)
	for _, stream := range streams {
		var ok bool
		var tag flv.Tag
		if tag, ok, err = flv.CodecDataToTag(stream); err != nil {
			return
		}
		if ok {
			if err = this.writeAVTag(tag, 0); err != nil {
				return
			}
		}
	}

	this.streams = streams
	this.stage++
	return
}

func (this *Conn) tmpwbuf(n int) []byte {
	if len(this.writebuf) < n {
		this.writebuf = make([]byte, n)
	}
	return this.writebuf
}

func (this *Conn) writeSetChunkSize(size int) (err error) {
	this.writeMaxChunkSize = size
	b := this.tmpwbuf(chunkHeaderLength + 4)
	n := this.fillChunkHeader(b, 2, 0, msgtypeidSetChunkSize, 0, 4)
	pio.PutU32BE(b[n:], uint32(size))
	n += 4
	_, err = this.bufw.Write(b[:n])
	return
}

func (this *Conn) writeAck(seqnum uint32) (err error) {
	b := this.tmpwbuf(chunkHeaderLength + 4)
	n := this.fillChunkHeader(b, 2, 0, msgtypeidAck, 0, 4)
	pio.PutU32BE(b[n:], seqnum)
	n += 4
	_, err = this.bufw.Write(b[:n])
	return
}

func (this *Conn) writeWindowAckSize(size uint32) (err error) {
	b := this.tmpwbuf(chunkHeaderLength + 4)
	n := this.fillChunkHeader(b, 2, 0, msgtypeidWindowAckSize, 0, 4)
	pio.PutU32BE(b[n:], size)
	n += 4
	_, err = this.bufw.Write(b[:n])
	return
}

func (this *Conn) writeSetPeerBandwidth(acksize uint32, limittype uint8) (err error) {
	b := this.tmpwbuf(chunkHeaderLength + 5)
	n := this.fillChunkHeader(b, 2, 0, msgtypeidSetPeerBandwidth, 0, 5)
	pio.PutU32BE(b[n:], acksize)
	n += 4
	b[n] = limittype
	n++
	_, err = this.bufw.Write(b[:n])
	return
}

func (this *Conn) writeCommandMsg(csid, msgsid uint32, args ...interface{}) (err error) {
	return this.writeAMF0Msg(msgtypeidCommandMsgAMF0, csid, msgsid, args...)
}

func (this *Conn) writeDataMsg(csid, msgsid uint32, args ...interface{}) (err error) {
	return this.writeAMF0Msg(msgtypeidDataMsgAMF0, csid, msgsid, args...)
}

func (this *Conn) writeAMF0Msg(msgtypeid uint8, csid, msgsid uint32, args ...interface{}) (err error) {
	size := 0
	for _, arg := range args {
		size += flv.LenAMF0Val(arg)
	}

	b := this.tmpwbuf(chunkHeaderLength + size)
	n := this.fillChunkHeader(b, csid, 0, msgtypeid, msgsid, size)
	for _, arg := range args {
		n += flv.FillAMF0Val(b[n:], arg)
	}

	_, err = this.bufw.Write(b[:n])
	return
}

func (this *Conn) writeAVTag(tag flv.Tag, ts int32) (err error) {
	var msgtypeid uint8
	var csid uint32
	var data []byte

	switch tag.Type {
	case flv.TAG_AUDIO:
		msgtypeid = msgtypeidAudioMsg
		csid = 6
		data = tag.Data

	case flv.TAG_VIDEO:
		msgtypeid = msgtypeidVideoMsg
		csid = 7
		data = tag.Data
	}

	b := this.tmpwbuf(chunkHeaderLength + flv.MaxTagSubHeaderLength)
	hdrlen := tag.FillHeader(b[chunkHeaderLength:])
	this.fillChunkHeader(b, csid, ts, msgtypeid, this.avmsgsid, hdrlen+len(data))
	n := hdrlen + chunkHeaderLength

	if n+len(data) > this.writeMaxChunkSize {
		if err = this.writeSetChunkSize(n + len(data)); err != nil {
			return
		}
	}

	if _, err = this.bufw.Write(b[:n]); err != nil {
		return
	}
	_, err = this.bufw.Write(data)
	return
}

func (this *Conn) writeStreamBegin(msgsid uint32) (err error) {
	b := this.tmpwbuf(chunkHeaderLength + 6)
	n := this.fillChunkHeader(b, 2, 0, msgtypeidUserControl, 0, 6)
	pio.PutU16BE(b[n:], eventtypeStreamBegin)
	n += 2
	pio.PutU32BE(b[n:], msgsid)
	n += 4
	_, err = this.bufw.Write(b[:n])
	return
}

func (this *Conn) writeSetBufferLength(msgsid uint32, timestamp uint32) (err error) {
	b := this.tmpwbuf(chunkHeaderLength + 10)
	n := this.fillChunkHeader(b, 2, 0, msgtypeidUserControl, 0, 10)
	pio.PutU16BE(b[n:], eventtypeSetBufferLength)
	n += 2
	pio.PutU32BE(b[n:], msgsid)
	n += 4
	pio.PutU32BE(b[n:], timestamp)
	n += 4
	_, err = this.bufw.Write(b[:n])
	return
}

const chunkHeaderLength = 12

func (this *Conn) fillChunkHeader(b []byte, csid uint32, timestamp int32, msgtypeid uint8, msgsid uint32, msgdatalen int) (n int) {
	//  0                   1                   2                   3
	//  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |                   timestamp                   |message length |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |     message length (cont)     |message type id| msg stream id |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |           message stream id (cont)            |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	//
	//       Figure 9 Chunk Message Header – Type 0

	b[n] = byte(csid) & 0x3f
	n++
	pio.PutU24BE(b[n:], uint32(timestamp))
	n += 3
	pio.PutU24BE(b[n:], uint32(msgdatalen))
	n += 3
	b[n] = msgtypeid
	n++
	pio.PutU32LE(b[n:], msgsid)
	n += 4

	return
}

func (this *Conn) flushWrite() (err error) {
	if err = this.bufw.Flush(); err != nil {
		return
	}
	return
}

func (this *Conn) readChunk() (err error) {
	b := this.readbuf
	n := 0
	if _, err = io.ReadFull(this.bufr, b[:1]); err != nil {
		return
	}
	header := b[0]
	n += 1

	var msghdrtype uint8
	var csid uint32

	msghdrtype = header >> 6

	csid = uint32(header) & 0x3f
	switch csid {
	default: // Chunk basic header 1
	case 0: // Chunk basic header 2
		if _, err = io.ReadFull(this.bufr, b[:1]); err != nil {
			return
		}
		n += 1
		csid = uint32(b[0]) + 64
	case 1: // Chunk basic header 3
		if _, err = io.ReadFull(this.bufr, b[:2]); err != nil {
			return
		}
		n += 2
		csid = uint32(pio.U16BE(b)) + 64
	}

	cs := this.readcsmap[csid]
	if cs == nil {
		cs = &chunkStream{}
		this.readcsmap[csid] = cs
	}

	var timestamp uint32

	switch msghdrtype {
	case 0:
		//  0                   1                   2                   3
		//  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
		// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
		// |                   timestamp                   |message length |
		// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
		// |     message length (cont)     |message type id| msg stream id |
		// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
		// |           message stream id (cont)            |
		// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
		//
		//       Figure 9 Chunk Message Header – Type 0
		if cs.msgdataleft != 0 {
			err = fmt.Errorf("rtmp: chunk msgdataleft=%d invalid", cs.msgdataleft)
			return
		}
		h := b[:11]
		if _, err = io.ReadFull(this.bufr, h); err != nil {
			return
		}
		n += len(h)
		timestamp = pio.U24BE(h[0:3])
		cs.msghdrtype = msghdrtype
		cs.msgdatalen = pio.U24BE(h[3:6])
		cs.msgtypeid = h[6]
		cs.msgsid = pio.U32LE(h[7:11])
		if timestamp == 0xffffff {
			if _, err = io.ReadFull(this.bufr, b[:4]); err != nil {
				return
			}
			n += 4
			timestamp = pio.U32BE(b)
			cs.hastimeext = true
		} else {
			cs.hastimeext = false
		}
		cs.timenow = timestamp
		cs.Start()

	case 1:
		//  0                   1                   2                   3
		//  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
		// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
		// |                timestamp delta                |message length |
		// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
		// |     message length (cont)     |message type id|
		// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
		//
		//       Figure 10 Chunk Message Header – Type 1
		if cs.msgdataleft != 0 {
			err = fmt.Errorf("rtmp: chunk msgdataleft=%d invalid", cs.msgdataleft)
			return
		}
		h := b[:7]
		if _, err = io.ReadFull(this.bufr, h); err != nil {
			return
		}
		n += len(h)
		timestamp = pio.U24BE(h[0:3])
		cs.msghdrtype = msghdrtype
		cs.msgdatalen = pio.U24BE(h[3:6])
		cs.msgtypeid = h[6]
		if timestamp == 0xffffff {
			if _, err = io.ReadFull(this.bufr, b[:4]); err != nil {
				return
			}
			n += 4
			timestamp = pio.U32BE(b)
			cs.hastimeext = true
		} else {
			cs.hastimeext = false
		}
		cs.timedelta = timestamp
		cs.timenow += timestamp
		cs.Start()

	case 2:
		//  0                   1                   2
		//  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3
		// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
		// |                timestamp delta                |
		// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
		//
		//       Figure 11 Chunk Message Header – Type 2
		if cs.msgdataleft != 0 {
			err = fmt.Errorf("rtmp: chunk msgdataleft=%d invalid", cs.msgdataleft)
			return
		}
		h := b[:3]
		if _, err = io.ReadFull(this.bufr, h); err != nil {
			return
		}
		n += len(h)
		cs.msghdrtype = msghdrtype
		timestamp = pio.U24BE(h[0:3])
		if timestamp == 0xffffff {
			if _, err = io.ReadFull(this.bufr, b[:4]); err != nil {
				return
			}
			n += 4
			timestamp = pio.U32BE(b)
			cs.hastimeext = true
		} else {
			cs.hastimeext = false
		}
		cs.timedelta = timestamp
		cs.timenow += timestamp
		cs.Start()

	case 3:
		if cs.msgdataleft == 0 {
			switch cs.msghdrtype {
			case 0:
				if cs.hastimeext {
					if _, err = io.ReadFull(this.bufr, b[:4]); err != nil {
						return
					}
					n += 4
					timestamp = pio.U32BE(b)
					cs.timenow = timestamp
				}
			case 1, 2:
				if cs.hastimeext {
					if _, err = io.ReadFull(this.bufr, b[:4]); err != nil {
						return
					}
					n += 4
					timestamp = pio.U32BE(b)
				} else {
					timestamp = cs.timedelta
				}
				cs.timenow += timestamp
			}
			cs.Start()
		}

	default:
		err = fmt.Errorf("rtmp: invalid chunk msg header type=%d", msghdrtype)
		return
	}

	size := int(cs.msgdataleft)
	if size > this.readMaxChunkSize {
		size = this.readMaxChunkSize
	}
	off := cs.msgdatalen - cs.msgdataleft
	buf := cs.msgdata[off : int(off)+size]
	if _, err = io.ReadFull(this.bufr, buf); err != nil {
		return
	}
	n += len(buf)
	cs.msgdataleft -= uint32(size)

	if cs.msgdataleft == 0 {

		if err = this.handleMsg(cs.timenow, cs.msgsid, cs.msgtypeid, cs.msgdata); err != nil {
			return
		}
	}

	this.ackn += uint32(n)
	if this.readAckSize != 0 && this.ackn > this.readAckSize {
		if err = this.writeAck(this.ackn); err != nil {
			return
		}
		this.ackn = 0
	}

	return
}

func (this *Conn) handleCommandMsgAMF0(b []byte) (n int, err error) {
	var name, transid, obj interface{}
	var size int

	if name, size, err = flv.ParseAMF0Val(b[n:]); err != nil {
		return
	}
	n += size
	if transid, size, err = flv.ParseAMF0Val(b[n:]); err != nil {
		return
	}
	n += size
	if obj, size, err = flv.ParseAMF0Val(b[n:]); err != nil {
		return
	}
	n += size

	var ok bool
	if this.commandname, ok = name.(string); !ok {
		err = fmt.Errorf("rtmp: CommandMsgAMF0 command is not string")
		return
	}
	this.commandtransid, _ = transid.(float64)
	this.commandobj, _ = obj.(flv.AMFMap)
	this.commandparams = []interface{}{}

	for n < len(b) {
		if obj, size, err = flv.ParseAMF0Val(b[n:]); err != nil {
			return
		}
		n += size
		this.commandparams = append(this.commandparams, obj)
	}
	if n < len(b) {
		err = fmt.Errorf("rtmp: CommandMsgAMF0 left bytes=%d", len(b)-n)
		return
	}

	this.gotcommand = true
	return
}

func (this *Conn) handleMsg(timestamp uint32, msgsid uint32, msgtypeid uint8, msgdata []byte) (err error) {
	this.msgdata = msgdata
	this.msgtypeid = msgtypeid
	this.timestamp = timestamp

	switch msgtypeid {
	case msgtypeidCommandMsgAMF0:
		if _, err = this.handleCommandMsgAMF0(msgdata); err != nil {
			return
		}

	case msgtypeidCommandMsgAMF3:
		if len(msgdata) < 1 {
			err = fmt.Errorf("rtmp: short packet of CommandMsgAMF3")
			return
		}
		// skip first byte
		if _, err = this.handleCommandMsgAMF0(msgdata[1:]); err != nil {
			return
		}

	case msgtypeidUserControl:
		if len(msgdata) < 2 {
			err = fmt.Errorf("rtmp: short packet of UserControl")
			return
		}
		this.eventtype = pio.U16BE(msgdata)

	case msgtypeidDataMsgAMF0:
		b := msgdata
		n := 0
		for n < len(b) {
			var obj interface{}
			var size int
			if obj, size, err = flv.ParseAMF0Val(b[n:]); err != nil {
				return
			}
			n += size
			this.datamsgvals = append(this.datamsgvals, obj)
		}
		if n < len(b) {
			err = fmt.Errorf("rtmp: DataMsgAMF0 left bytes=%d", len(b)-n)
			return
		}

	case msgtypeidVideoMsg:
		if len(msgdata) == 0 {
			return
		}
		tag := flv.Tag{Type: flv.TAG_VIDEO}
		var n int
		if n, err = (&tag).ParseHeader(msgdata); err != nil {
			return
		}
		if !(tag.FrameType == flv.FRAME_INTER || tag.FrameType == flv.FRAME_KEY) {
			return
		}
		tag.Data = msgdata[n:]
		this.avtag = tag

	case msgtypeidAudioMsg:
		if len(msgdata) == 0 {
			return
		}
		tag := flv.Tag{Type: flv.TAG_AUDIO}
		var n int
		if n, err = (&tag).ParseHeader(msgdata); err != nil {
			return
		}
		tag.Data = msgdata[n:]
		this.avtag = tag

	case msgtypeidSetChunkSize:
		if len(msgdata) < 4 {
			err = fmt.Errorf("rtmp: short packet of SetChunkSize")
			return
		}
		this.readMaxChunkSize = int(pio.U32BE(msgdata))
		return
	}

	this.gotmsg = true
	return
}

var (
	hsClientFullKey = []byte{
		'G', 'e', 'n', 'u', 'i', 'n', 'e', ' ', 'A', 'd', 'o', 'b', 'e', ' ',
		'F', 'l', 'a', 's', 'h', ' ', 'P', 'l', 'a', 'y', 'e', 'r', ' ',
		'0', '0', '1',
		0xF0, 0xEE, 0xC2, 0x4A, 0x80, 0x68, 0xBE, 0xE8, 0x2E, 0x00, 0xD0, 0xD1,
		0x02, 0x9E, 0x7E, 0x57, 0x6E, 0xEC, 0x5D, 0x2D, 0x29, 0x80, 0x6F, 0xAB,
		0x93, 0xB8, 0xE6, 0x36, 0xCF, 0xEB, 0x31, 0xAE,
	}
	hsServerFullKey = []byte{
		'G', 'e', 'n', 'u', 'i', 'n', 'e', ' ', 'A', 'd', 'o', 'b', 'e', ' ',
		'F', 'l', 'a', 's', 'h', ' ', 'M', 'e', 'd', 'i', 'a', ' ',
		'S', 'e', 'r', 'v', 'e', 'r', ' ',
		'0', '0', '1',
		0xF0, 0xEE, 0xC2, 0x4A, 0x80, 0x68, 0xBE, 0xE8, 0x2E, 0x00, 0xD0, 0xD1,
		0x02, 0x9E, 0x7E, 0x57, 0x6E, 0xEC, 0x5D, 0x2D, 0x29, 0x80, 0x6F, 0xAB,
		0x93, 0xB8, 0xE6, 0x36, 0xCF, 0xEB, 0x31, 0xAE,
	}
	hsClientPartialKey = hsClientFullKey[:30]
	hsServerPartialKey = hsServerFullKey[:36]
)

func hsMakeDigest(key []byte, src []byte, gap int) (dst []byte) {
	h := hmac.New(sha256.New, key)
	if gap <= 0 {
		h.Write(src)
	} else {
		h.Write(src[:gap])
		h.Write(src[gap+32:])
	}
	return h.Sum(nil)
}

func hsCalcDigestPos(p []byte, base int) (pos int) {
	for i := 0; i < 4; i++ {
		pos += int(p[base+i])
	}
	pos = (pos % 728) + base + 4
	return
}

func hsFindDigest(p []byte, key []byte, base int) int {
	gap := hsCalcDigestPos(p, base)
	digest := hsMakeDigest(key, p, gap)
	if bytes.Compare(p[gap:gap+32], digest) != 0 {
		return -1
	}
	return gap
}

func hsParse1(p []byte, peerkey []byte, key []byte) (ok bool, digest []byte) {
	var pos int
	if pos = hsFindDigest(p, peerkey, 772); pos == -1 {
		if pos = hsFindDigest(p, peerkey, 8); pos == -1 {
			return
		}
	}
	ok = true
	digest = hsMakeDigest(key, p[pos:pos+32], -1)
	return
}

func hsCreate01(p []byte, time uint32, ver uint32, key []byte) {
	p[0] = 3
	p1 := p[1:]
	rand.Read(p1[8:])
	pio.PutU32BE(p1[0:4], time)
	pio.PutU32BE(p1[4:8], ver)
	gap := hsCalcDigestPos(p1, 8)
	digest := hsMakeDigest(key, p1, gap)
	copy(p1[gap:], digest)
}

func hsCreate2(p []byte, key []byte) {
	rand.Read(p)
	gap := len(p) - 32
	digest := hsMakeDigest(key, p, gap)
	copy(p[gap:], digest)
}

func (this *Conn) handshakeClient() (err error) {
	var random [(1 + 1536*2) * 2]byte

	C0C1C2 := random[:1536*2+1]
	C0 := C0C1C2[:1]
	//C1 := C0C1C2[1:1536+1]
	C0C1 := C0C1C2[:1536+1]
	C2 := C0C1C2[1536+1:]

	S0S1S2 := random[1536*2+1:]
	//S0 := S0S1S2[:1]
	S1 := S0S1S2[1 : 1536+1]
	//S0S1 := S0S1S2[:1536+1]
	//S2 := S0S1S2[1536+1:]

	C0[0] = 3
	//hsCreate01(C0C1, hsClientFullKey)

	// > C0C1
	if _, err = this.bufw.Write(C0C1); err != nil {
		return
	}
	if err = this.bufw.Flush(); err != nil {
		return
	}

	// < S0S1S2
	if _, err = io.ReadFull(this.bufr, S0S1S2); err != nil {
		return
	}

	if ver := pio.U32BE(S1[4:8]); ver != 0 {
		C2 = S1
	} else {
		C2 = S1
	}

	// > C2
	if _, err = this.bufw.Write(C2); err != nil {
		return
	}

	this.stage++
	return
}

func (this *Conn) handshakeServer() (err error) {
	var random [(1 + 1536*2) * 2]byte

	C0C1C2 := random[:1536*2+1]
	C0 := C0C1C2[:1]
	C1 := C0C1C2[1 : 1536+1]
	C0C1 := C0C1C2[:1536+1]
	C2 := C0C1C2[1536+1:]

	S0S1S2 := random[1536*2+1:]
	S0 := S0S1S2[:1]
	S1 := S0S1S2[1 : 1536+1]
	S0S1 := S0S1S2[:1536+1]
	S2 := S0S1S2[1536+1:]

	// < C0C1
	if _, err = io.ReadFull(this.bufr, C0C1); err != nil {
		return
	}
	if C0[0] != 3 {
		err = fmt.Errorf("rtmp: handshake version=%d invalid", C0[0])
		return
	}

	S0[0] = 3

	clitime := pio.U32BE(C1[0:4])
	srvtime := clitime
	srvver := uint32(0x0d0e0a0d)
	cliver := pio.U32BE(C1[4:8])

	if cliver != 0 {
		var ok bool
		var digest []byte
		if ok, digest = hsParse1(C1, hsClientPartialKey, hsServerFullKey); !ok {
			err = fmt.Errorf("rtmp: handshake server: C1 invalid")
			return
		}
		hsCreate01(S0S1, srvtime, srvver, hsServerPartialKey)
		hsCreate2(S2, digest)
	} else {
		copy(S1, C1)
		copy(S2, C2)
	}

	// > S0S1S2
	if _, err = this.bufw.Write(S0S1S2); err != nil {
		return
	}
	if err = this.bufw.Flush(); err != nil {
		return
	}

	// < C2
	if _, err = io.ReadFull(this.bufr, C2); err != nil {
		return
	}

	this.stage++
	return
}

type closeConn struct {
	*Conn
	waitclose chan bool
}

func (this closeConn) Close() error {
	this.waitclose <- true
	return nil
}
