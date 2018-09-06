package dmi

import (
	"time"
	"net"
	"fmt"
	"bufio"
//	"strconv"

	"github.com/deepglint/dgmf/mserver/protocols/dmi/dmic"
	"github.com/deepglint/dgmf/mserver/protocols/dmi/dmid"
	"github.com/deepglint/dgmf/mserver/core"
)

type DMIServer struct {
	Listener *net.TCPListener
	Port     int
	Param    interface{}
}

type Instance struct
{
	conn *net.TCPConn
	h264filepath string
	SessionID uint32

	dmidbuffer *dmid.DMIDBuffer
	fcallback dmid.FrameCallback
	rcallback dmid.ResponseCallback

	chanNet chan *dmid.NetData
	dmiinput DMILiveInput
	
	streamid string
	Running	bool
}

func (this *DMIServer) Start(port int, param interface{}) error {
	this.Port = port
	this.Param = param

    tcpAddr, errResolve := net.ResolveTCPAddr("tcp", fmt.Sprintf("0.0.0.0:%d", port))
    if errResolve != nil {
    	println("net.ResolveTCPAddr error ", errResolve.Error())
    	return errResolve
    }
    
    tcpListener, errListen := net.ListenTCP("tcp", tcpAddr)
    if errListen != nil {
    	println("net.ListenTCP error ", errListen.Error())
    	return errListen
    }
    this.Listener = tcpListener

    go this.listenAndServe()

	return nil
}

func (this *DMIServer) Stop() {
	pool := core.GetESPool()
	pool.Live.RemoveSessionByOutput("dmi")
	this.Listener.Close()
}

func (this *DMIServer) GetPort() int {
	return this.Port
}

func (this *DMIServer) GetParam() interface{} {
	return this.Param
}

func (this *DMIServer) listenAndServe() {
	var cursessionid uint32 = 0

	for {
       tcpConn, err := this.Listener.AcceptTCP()
       if err != nil {
	    	println("tcpListener.AcceptTCP error ", err.Error())
	         continue
       }

       tcpConn.SetNoDelay(true)
       fmt.Println("A client connected : " + tcpConn.RemoteAddr().String())

       inst := &Instance{}
       inst.conn = tcpConn
       inst.Running = true
       inst.SessionID = cursessionid
       cursessionid++

       go tcpPipe(inst)
	}
}

func tcpPipe(inst *Instance) {
	state := dmic.REGISTER1

    ipStr := inst.conn.RemoteAddr().String()
    defer func() {
       fmt.Println("tcpPipe quit:" + ipStr)
       inst.Running = false
       inst.conn.Close()
    }()

    reader := bufio.NewReader(inst.conn)

    buf := make([]byte, 2048)
	proto := &dmic.DmicProto{
		UserID: "11111",
		UserName: "admin",
		Password: "admin",
	}

	proto.SessionID = fmt.Sprintf("%d", inst.SessionID)

	var userid string
	var streamid string

	var response string
	var statuscode int
	var isPull bool = true

	pool := core.GetESPool()

    for inst.Running {
		if state == dmic.REGISTER1 {
			length, err := reader.Read(buf)
			if err != nil {
				println("reader.Read error ", err.Error())
				return
			}
			fmt.Println(string(buf[0:length]))

       		statuscode = 401
       		if !dmic.MatchRequest("REGISTER1", string(buf[0:length])) {
       			println("request doesn't match register1")
       			statuscode = 400
       		}

			userid = dmic.ParseRegister1Request(string(buf[0:length]))
			response = dmic.GetResponse(userid, "", "REGISTER1", statuscode, nil)
			println(response)
			inst.conn.Write([]byte(response))
			if statuscode == 401 {
				state = dmic.REGISTER2
			}
       } else if state == dmic.REGISTER2 {
	       length, err := reader.Read(buf)
	       if err != nil {
		    	println("reader.Read error ", err.Error())
		         return
	       }
	       fmt.Println(string(buf[0:length]))

       		statuscode = 200
       		if !dmic.MatchRequest("REGISTER2", string(buf[0:length])) {
       			println("request doesn't match register2")
       			statuscode = 400
       		}

			userID, username, password := dmic.ParseRegister2Request(string(buf[0:length]))
			if userID != userid {
       			println("userid mismatched")
       			statuscode = 400
			}

			if username != proto.UserName || password != proto.Password {
       			println("username or password is invalid")
       			statuscode = 400			
			}
		   response = dmic.GetResponse(userid, "", "REGISTER2", statuscode, nil)
		   println(response)
		   inst.conn.Write([]byte(response))
		   if statuscode == 200 {
			   state = dmic.REQUESTSTREAM       					
		   }
       } else if state == dmic.REQUESTSTREAM {
	       length, err := reader.Read(buf)
	       if err != nil {
		    	println("reader.Read error ", err.Error())
		         return
	       }
	       fmt.Println(string(buf[0:length]))

       		statuscode = 200
       		if dmic.MatchRequest("PULL", string(buf[0:length])) {

       			streamid = dmic.ParseRequestStreamID(string(buf[0:length]))
				if !pool.Live.ExistOutput(streamid, "dmi") {
       				println("output streamid doesn't exist")

				   response = dmic.GetResponse(userid, proto.SessionID, "PULL", 404, nil )
				   inst.conn.Write([]byte(response))
       				return				
				}

       			var body dmic.MediaInfo
       			body.StreamType = "live"
				var channel dmic.ChannelInfo
				channel.ChannelID = 0
				channel.PayloadType = 0
				body.Channels = append(body.Channels, channel)

			   response = dmic.GetResponse(userid, proto.SessionID, "PULL", 200, &body )
			   println(response)
			   inst.conn.Write([]byte(response))
			   isPull = true
			   state = dmic.PLAY       					
       		} else if dmic.MatchRequest("PUSH", string(buf[0:length])) {
				streamid = dmic.ParseRequestStreamID(string(buf[0:length]))
				inst.streamid = streamid
				println(streamid)
				if pool.Live.ExistInput(streamid) {
					println("input streamid already exist")

					response = dmic.GetResponse(userid, proto.SessionID, "PUSH", 400, nil )
					inst.conn.Write([]byte(response))
					return
				}

			   response = dmic.GetResponse(userid, proto.SessionID, "PUSH", 200, nil)
			   println(response)
			   inst.conn.Write([]byte(response))
			   isPull = false
			   state = dmic.PLAY
			}
       } else if state == dmic.PLAY {
	       length, err := reader.Read(buf)
	       if err != nil {
		    	println("reader.Read error ", err.Error())
		         return
	       }
	       fmt.Println(string(buf[0:length]))

       		if dmic.MatchRequest("PLAY", string(buf[0:length])) {
			   response = dmic.GetResponse(userid, proto.SessionID, "PLAY", 200, nil)
			   println(response)
			   inst.conn.Write([]byte(response))

			   if isPull {
				   state = dmic.STREAMINGPULL

					strstreamid := dmic.ParseRequestStreamID(string(buf[0:length]))
					println(strstreamid)

					stream, errStream := pool.Live.GetStream(strstreamid)
					if errStream != nil {
						println("pull stream doesn't exist", errStream.Error())
						return
					}
					if stream.Reserved != nil {
						receiver1 := stream.Reserved.(*DMIReceiver)
						receiver1.SendForceIRequest()
					}

				   go PullStreamProc(inst, streamid)
			   } else {
				   state = dmic.STREAMINGPUSH       					
					
					strstreamid := dmic.ParseRequestStreamID(string(buf[0:length]))
					uri := dmic.ParseRequestURI(string(buf[0:length]))
					println(strstreamid, uri)

					pool.Live.AddStream2(strstreamid, uri)
					pool.Live.AddOutput(strstreamid, "dmi", true, nil)
					pool.Live.AddOutput(strstreamid, "rtsp", true, nil)
					pool.Live.AddOutput(strstreamid, "rtmp", true, nil)

					stream, errStream := pool.Live.GetStream(strstreamid)
					if errStream != nil {
						println("push stream doesn't exist", errStream.Error())
						return
					}

					inst.dmiinput.receiver = &DMIReceiver{}
					inst.dmiinput.receiver.inst = inst
					stream.Reserved = inst.dmiinput.receiver

//					stream = stream
					inst.dmiinput.Open(uri, stream)				
			   }
			}
       } else if state == dmic.STREAMINGPULL {
   			println("state == dmic.STREAMINGPULL")
			
			length, err := reader.Read(buf)
			if err != nil {
				println("reader.Read error ", err.Error())
				return
			}
			fmt.Println(string(buf[0:length]))

       		if dmic.MatchRequest("BYE", string(buf[0:length])) {
			   response = dmic.GetResponse(userid, proto.SessionID, "BYE", 200, nil )
			   println(response)
			   inst.conn.Write([]byte(response))
			    inst.Running = false
			} else if dmic.MatchRequest("PLAY", string(buf[0:length])) {
			   response = dmic.GetResponse(userid, proto.SessionID, "PLAY", 200, nil )
			   println(response)
			   inst.conn.Write([]byte(response))			
			}
       } else if state == dmic.STREAMINGPUSH {

       }

       time.Sleep(time.Millisecond * 10)
    }

    if !isPull {
		pool.Live.RemoveStream(streamid)
    }

	inst.Running = false
	fmt.Println("tcpPipe quit:")
}

func PullStreamProc(inst *Instance, streamid string) {
	var proto dmid.DMIDProto
	proto.SessionID = inst.SessionID

	proto.Version = 1
	proto.ChannelCount = 1
	proto.PacketID = 0
	proto.TimeStamp = 1

	var channelinfo dmid.ChannelInfo
//	channelinfo.ChannelType = 0
	channelinfo.FrameTime = 0

	pool := core.GetESPool()
	println("streamid = ", streamid, "sessionid =", proto.SessionID)
	pool.Live.AddSession(streamid, fmt.Sprintf("%v",proto.SessionID), "dmi", inst.conn)
	stream, err := pool.Live.GetStream(streamid)
	if err != nil {
		println("new stream error", err, stream)
		return;		
	}
	frames, err := pool.Live.GetFrames(streamid, fmt.Sprintf("%v",proto.SessionID) )
	if err != nil {
		println("new stream error", err, frames)
		return;		
	}

	for frame := range frames {
		if !inst.Running {
			break
		}
		
//		starttime := time.Now().UnixNano() / int64(time.Millisecond)
		sendPacket(frame, inst, &channelinfo, &proto)
//		println("cost ", (time.Now().UnixNano() / int64(time.Millisecond))- starttime, len(frame.Data))
	}

    inst.Running = false
   fmt.Println("MediaStreamProc quit:")
}

func sendPacket(frameinfo *core.H264ESFrame, inst *Instance, channelinfo *dmid.ChannelInfo, proto *dmid.DMIDProto) {
	var sendbuf [2048]byte
	var buflength int
	frame := frameinfo.Data
	if frame == nil {
		println("frame data is nil")
		return
	}

//		println(len(frame))

	var packetLength int = 1400
	var offset int = 0

	channelinfo.PayloadType = 0
	channelinfo.SliceCount = uint16(len(frame) / packetLength)
	if len(frame) % packetLength != 0 {
		channelinfo.SliceCount++
	}

	channelinfo.SliceID = 0
	channelinfo.FrameTime = time.Now().UnixNano() / int64(time.Millisecond)
	for offset < len(frame) {

		dmidheaderbuf := proto.GetDMIDHeader()

		if offset + packetLength < len(frame) {	
			channelinfo.PayloadLength = uint16(packetLength)
		} else {
			channelinfo.PayloadLength = uint16(len(frame) - offset)
		}
		channelheaderbuf := channelinfo.GetChannelHeader()

		buflength = 0
		copy(sendbuf[buflength:], dmidheaderbuf)
		buflength += len(dmidheaderbuf)
		copy(sendbuf[buflength:], channelheaderbuf)
		buflength += len(channelheaderbuf)

		channelinfo.PayloadType = 0
		if offset + packetLength < len(frame) {
			copy(sendbuf[buflength:], frame[offset:offset + packetLength])
			buflength += packetLength
		} else {
			copy(sendbuf[buflength:], frame[offset:])
			buflength += len(frame) - offset
		}

//			fmt.Printf("%c %c %c %c \n", sendbuf[0], sendbuf[1], sendbuf[2], sendbuf[3])
		length, errWrite := inst.conn.Write(sendbuf[0:buflength])
		if errWrite != nil {
			println("write error ", errWrite.Error(), length)
			return
		}

		offset += packetLength
		channelinfo.SliceID++
	
		proto.PacketID++
	}

	channelinfo.FrameID++
}
