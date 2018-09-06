package dmi

import (
	"time"
//	"net"
	"fmt"
//	"bufio"
	"strconv"

	"github.com/deepglint/dgmf/mserver/protocols/dmi/dmic"
	"github.com/deepglint/dgmf/mserver/protocols/dmi/dmid"
	"github.com/deepglint/dgmf/mserver/protocols/h264"
	"github.com/deepglint/dgmf/mserver/core"
)

type DMIReceiver struct {
	core.Receiver

	inst *Instance

	fps 		   uint32
	index 		   uint64
	ts 			   uint64
	t0 			   time.Time
}

func (this *DMIReceiver) Open(uri string, streamId string, rtms chan core.RTMessage) {
	this.fps = 30
	this.index = 0
	this.ts = 0

	this.AbsRunning = false
	this.AbsFrames = make(chan *core.H264ESFrame)
	this.AbsStoped = make(chan bool)
	this.AbsRTMessages = rtms
	this.AbsStreamId = streamId
	this.AbsIndex = 0

	go this.PushStreamProc(this.inst)

	this.AbsRunning = true
}

func (this *DMIReceiver) Close() {

	this.AbsRunning = false

	this.inst.Running = false
}

func (this *DMIReceiver) PushStreamProc(inst *Instance) {
	inst.dmidbuffer = &dmid.DMIDBuffer{}
	inst.dmidbuffer.Buffer = make([]byte, 2 * 1024 * 1024)
	inst.dmidbuffer.FrameMapID = 0
	inst.dmidbuffer.ChanFrame = make(chan *dmid.FrameInfo, 10)

	inst.chanNet = make(chan *dmid.NetData, 10)
	buffer := make([]byte, 500 * 1024)

	select {
		case this.AbsRTMessages <- core.RTMessage{Status: 200}:
		default:
	}

	go this.ParseDataProc(inst)

	for inst.Running {
		length, errRead := inst.conn.Read(buffer)
		if errRead != nil {
			println("read error", errRead.Error())
			break;
		}

		netdata := &dmid.NetData{}
	
		copy(netdata.Data[0:], buffer[0:length])
		netdata.Length = length

//		println("PushStreamProc read=", length)
		inst.chanNet <- netdata		

		time.Sleep(time.Millisecond * 10)
	}

	inst.Running = false
	close(inst.chanNet)
	println("NetDataProc exit")
}

func (this *DMIReceiver) SendForceIRequest() {
	println("SendForceIRequest", this.inst.conn.RemoteAddr().String())
	request := dmic.GetForceIRequest(this.inst.conn.RemoteAddr().String(), this.inst.streamid, fmt.Sprintf("%v", this.inst.SessionID))
	this.inst.conn.Write([]byte(request))
	println(request)
}

func (this *DMIReceiver)ParseDataProc(inst *Instance) {
	go this.FrameDataProc(inst)

	for inst.Running {
		select {
			case recvData := <- inst.chanNet:
//				println("parseproc", recvData.Length)
				inst.dmidbuffer.SetData(recvData.Data[0:recvData.Length])
			default:
		}

//		starttime := time.Now().UnixNano() / int64(time.Millisecond)

		inst.dmidbuffer.GetFrame()

//		println("ParseDataProc cost ", (time.Now().UnixNano() / int64(time.Millisecond))- starttime)

		time.Sleep(time.Millisecond * 10)
	}

	inst.Running = false
	println("ParseDataProc exit")
}

func (this *DMIReceiver) FrameDataProc(inst *Instance) {

	println("FrameDataProc begin")

	for inst.Running {
		select {
			case frame := <- inst.dmidbuffer.ChanFrame:
//				println("FrameDataProc", frame.Length)
				if frame.Channelheader.PayloadType == 0 {
					this.PushH264Frame(frame)					
				} else {
					println("payloadtype is not supported")
				}
			default:

		}


		time.Sleep(time.Millisecond * 10)
	}

	inst.Running = false
	println("FrameDataProc exit")
}

func (this *DMIReceiver) PushH264Frame(frameinfo *dmid.FrameInfo) {

	data := frameinfo.Data
	if data[0] == 0x00 && data[1] == 0x00 && data[2] == 0x00 && data[3] == 0x01 {
		iFrame := false
		this.index++
		if frameinfo.Length >= 4 && (data[4]&0x1F == 7 || data[4]&0x1F == 8 || data[4]&0x1F == 5) {
			iFrame = true
		}
		if this.index%100 == 0 {
			f, _ := strconv.ParseFloat(fmt.Sprintf("%0.2f", 1.0/(float64(time.Now().Sub(this.t0).Nanoseconds())/1000000000.00)), 64)
			if f*100 > 0 {
				this.fps = uint32(f * 100)
			}
			this.t0 = time.Now()
		}

		buf := make([]byte, frameinfo.Length)
//		println(frameinfo.Length)

		copy(buf, frameinfo.Data[0:frameinfo.Length])
		this.ts += uint64(90000 / this.fps)

		frame := &core.H264ESFrame{
			Data:      buf,
			Timestamp: uint32(this.ts),
			IFrame:    iFrame,
			Index:     uint64(this.index),
			// Fps:       fps,
		}

		this.AbsFPS = this.fps
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
				return 
			}
		default:
			this.AbsIndex++
//			frame = frame
			this.AbsFrames <- frame
//			println("pushframe ", len(frame.Data))
		}
	}
}