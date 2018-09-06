package rtsp

import (
	"fmt"
	"net"

	"github.com/deepglint/dgmf/mserver/core"
	"github.com/deepglint/dgmf/mserver/protocols/dmi"
	uuid "github.com/satori/go.uuid"
)

const MAX_RTP_SIZE = 1426

type RTSPServer struct {
	Listener net.Listener
	Port     int
	Param    interface{}
}

type SessionContext struct {
	SessionId     string
	StreamId      string
	IsUDP         bool
	StreamType    string
	UDPClientPort string
	SSRC          uint32
	Timestamp     uint32
	UDPConnect    *net.UDPConn
	TCPConnect    net.Conn
	RTMS          chan core.RTMessage
}

func (this *RTSPServer) Start(port int, param interface{}) error {
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

func (this *RTSPServer) listenAndServe(port int) {
	for {
		connect, err := this.Listener.Accept()
		if err != nil {
			break
		}
		sessionCtx := &SessionContext{
			SessionId:  uuid.NewV4().String(),
			StreamId:   "",
			TCPConnect: connect,
		}
		go this.parseRequest(sessionCtx)
	}
}

func (this *RTSPServer) parseRequest(sessionCtx *SessionContext) {
	err := this.initRequest(sessionCtx)
	if err == nil {

		pool := core.GetESPool()
		stream, _ := pool.Live.GetStream(sessionCtx.StreamId)
		if stream != nil && stream.Reserved != nil {
			receiver1 := stream.Reserved.(*dmi.DMIReceiver)
			receiver1.SendForceIRequest()
		}

		go this.sendFrames(sessionCtx)
		this.runtimeRequrst(sessionCtx)
	}
}

func (this *RTSPServer) Stop() {
	pool := core.GetESPool()
	pool.Live.RemoveSessionByOutput("rtsp")
	pool.Proxy.RemoveStreamByOutput("rtsp")
	this.Listener.Close()
}

func (this *RTSPServer) GetPort() int {
	return this.Port
}

func (this *RTSPServer) GetParam() interface{} {
	return this.Param
}
