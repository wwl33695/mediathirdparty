package rtmp

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/deepglint/dgmf/mserver/core"
	"github.com/deepglint/dgmf/mserver/protocols/rtmp/av"
	"github.com/deepglint/dgmf/mserver/protocols/rtsp"
	"github.com/deepglint/dgmf/mserver/protocols/dmi"
	uuid "github.com/satori/go.uuid"
)

type RTMPServer struct {
	Listener   net.Listener
	Port       int
	Param      interface{}
	Addr       string
	HandleConn func(*Conn)
}

type SessionContext struct {
	SessionId     string
	StreamId      string
	StreamType    string
	TCPConnect    *Conn
	RTMS          chan core.RTMessage
	proxyURI      string
	proxyReceiver *rtsp.RTSPReceiver
}

func (this *RTMPServer) Start(port int, param interface{}) error {
	var err error

	this.Port = port
	this.Param = param

	this.Listener, err = net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", this.Port))
	if err != nil {
		return err
	}

	go this.listenAndServe(port)

	return nil
}

func (this *RTMPServer) listenAndServe(port int) {
	for {
		connect, err := this.Listener.Accept()
		if err != nil {
			break
		}
		sessionCtx := &SessionContext{
			SessionId:  uuid.NewV4().String(),
			StreamId:   "",
			TCPConnect: NewConn(connect),
		}
		go this.parseRequest(sessionCtx)

	}
}

func (this *RTMPServer) parseRequest(sessionCtx *SessionContext) (err error) {
	err = this.initRequest(sessionCtx)
	if err == nil {

		pool := core.GetESPool()
		stream, _ := pool.Live.GetStream(sessionCtx.StreamId)
		if stream != nil && stream.Reserved != nil {
			receiver1 := stream.Reserved.(*dmi.DMIReceiver)
			receiver1.SendForceIRequest()
		}

		this.sendFrames(sessionCtx)
	}
	return
}

func (this *RTMPServer) sendFrames(sessionCtx *SessionContext) {
	pool := core.GetESPool()

	t0 := time.Now()
	if strings.EqualFold(sessionCtx.StreamType, "live") {
		pool.Live.AddSession(sessionCtx.StreamId, sessionCtx.StreamId, "rtmp", sessionCtx.TCPConnect.NetConn())
		frames, err := pool.Live.GetFrames(sessionCtx.StreamId, sessionCtx.StreamId)
		if err == nil && frames != nil {
			for frame := range frames {
				if frame != nil {
					err = this.sendPacket(t0, frame, sessionCtx.TCPConnect, sessionCtx.StreamId, sessionCtx.StreamId)
					if err != nil {
						break
					}
				} else {
					break
				}
			}
		}
		sessionCtx.TCPConnect.WriteTrailer()
		sessionCtx.TCPConnect.Close()
		pool.Live.RemoveSession(sessionCtx.StreamId, sessionCtx.StreamId)
	} else if strings.EqualFold(sessionCtx.StreamType, "vod") {

	} else if strings.EqualFold(sessionCtx.StreamType, "proxy") {
		frames, err := pool.Proxy.GetFrames(sessionCtx.StreamId)
		runProxy := true
		if err == nil && frames != nil {
			for runProxy {
				select {
				case frame := <-frames:
					if frame != nil {
						err = this.sendPacket(t0, frame, sessionCtx.TCPConnect, sessionCtx.StreamId, sessionCtx.StreamId)
						if err != nil {
							runProxy = false
							break
						}
					} else {
						runProxy = false
						break
					}
				case <-sessionCtx.RTMS:
					runProxy = false
					break
				}
			}
		}

		pool.Proxy.RemoveStream(sessionCtx.StreamId)
		// sessionCtx.TCPConnect.WriteTrailer()
		sessionCtx.TCPConnect.Close()
	}
}

func (this *RTMPServer) sendPacket(t0 time.Time, frame *core.H264ESFrame, conn *Conn, streamId string, sessionId string) error {
	var pkt av.Packet
	var data []byte
	var err error

	key := false
	if frame.Data[4]&0x1F == 0x07 {
		size := 0
		begin := 0
		key = true
		for i := 0; i < len(frame.Data); i++ {
			if i+3 < len(frame.Data) && frame.Data[i] == 0x00 && frame.Data[i+1] == 0x00 && frame.Data[i+2] == 0x00 && frame.Data[i+3] == 0x01 {
				if size > 0 {
					sb := make([]byte, 4)
					sb[0] = byte(size >> 24)
					sb[1] = byte((size & 0xFFFFFF) >> 16)
					sb[2] = byte((size & 0xFFFF) >> 8)
					sb[3] = byte(size & 0xFF)
					data = append(data, sb...)
					data = append(data, frame.Data[begin:begin+size]...)
				}
				size = 0
				i += 3
				begin = i + 1
				continue
			} else {
				size++
			}

			if i == len(frame.Data)-1 {
				sb := make([]byte, 4)
				sb[0] = byte(size >> 24)
				sb[1] = byte((size & 0xFFFFFF) >> 16)
				sb[2] = byte((size & 0xFFFF) >> 8)
				sb[3] = byte(size & 0xFF)
				data = append(data, sb...)
				data = append(data, frame.Data[begin:begin+size]...)
			}
		}
	} else {
		size := len(frame.Data) - 4
		sb := make([]byte, 4)
		sb[0] = byte(size >> 24)
		sb[1] = byte((size & 0xFFFFFF) >> 16)
		sb[2] = byte((size & 0xFFFF) >> 8)
		sb[3] = byte(size & 0xFF)
		data = append(data, sb...)
		data = append(data, frame.Data[4:]...)
	}

	pkt.Data = data
	pkt.IsKeyFrame = key
	pkt.Time = time.Now().Sub(t0)
	err = conn.WritePacket(pkt)
	if err != nil {
		return err
	}
	return nil
}

func (this *RTMPServer) Stop() {
	pool := core.GetESPool()
	pool.Live.RemoveSessionByOutput("rtmp")
	this.Listener.Close()
}

func (this *RTMPServer) GetPort() int {
	return this.Port
}

func (this *RTMPServer) GetParam() interface{} {
	return this.Param
}
