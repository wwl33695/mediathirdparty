package udp

import (
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
	"github.com/golang/glog"
)

// UDP server for receiving h264 element stream from remote udp client
type UDPReceiver struct {
	core.Receiver
	uri      string
	port     string
	connect  *net.UDPConn
	streamId string
}

// Open a server connection for remote udp client
func (this *UDPReceiver) Open(uri string, streamId string, rtms chan core.RTMessage) {
	glog.V(3).Infof("[UDP_RECEIVER] [STREAM_ID]=%s Open udp receiver started\n", this.AbsStreamId)
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

	this.AbsRunning = true
	select {
	case this.AbsRTMessages <- core.RTMessage{Status: 200}:
	default:
	}
	err = this.receiveFrames()
	this.clear(err)
	return
}

func (this *UDPReceiver) clear(err error) {

	if err != nil {
		select {
		case this.AbsRTMessages <- core.RTMessage{
			Status: 400,
			Error:  err,
		}:
		default:
		}
	}

	if this.connect != nil {
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

	glog.V(3).Infof("[UDP_RECEIVER] [STREAM_ID]=%s Clear udp receiver finished\n", this.AbsStreamId)
}

// Close server connection with remote udp client
func (this *UDPReceiver) Close() {
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
	glog.V(3).Infof("[UDP_RECEIVER] [STREAM_ID]=%s Close udp receiver finished\n", this.AbsStreamId)
}

// Parae udp url and create a udp server
func (this *UDPReceiver) openConnect(uri string) error {

	// Parse udp uri scheme
	urlCtx, err := url.Parse(uri)
	if err != nil {
		glog.Warningf("[UDP_RECEIVER] [STREAM_ID]=%s Open udp server failed, error: %s\n", this.AbsStreamId, err.Error())
		return err
	}
	if !strings.EqualFold(urlCtx.Scheme, "udp") {
		err = errors.New(urlCtx.Scheme + " not support")
		glog.Warningf("[UDP_RECEIVER] [STREAM_ID]=%s Open udp server failed, error: %s\n", this.AbsStreamId, err.Error())
		return err
	}
	if !strings.Contains(urlCtx.Host, "127.0.0.1") && !strings.Contains(urlCtx.Host, "localhost") {
		err = errors.New(urlCtx.Host + " not support")
		glog.Warningf("[UDP_RECEIVER] [STREAM_ID]=%s Open udp server failed, error: %s\n", this.AbsStreamId, err.Error())
		return err
	}
	parts := strings.Split(urlCtx.Host, ":")
	if len(parts) != 2 {
		err = errors.New("URI you provided is invalid")
		glog.Warningf("[UDP_RECEIVER] [STREAM_ID]=%s Open udp server failed, error: %s\n", this.AbsStreamId, err.Error())
		return err
	}
	this.uri = uri
	this.port = ":" + parts[1]

	this.AbsStoped = make(chan bool)

	// Create udp connection between this server and remote udp client
	addr, err := net.ResolveUDPAddr("udp", this.port)
	if err != nil {
		glog.Warningf("[UDP_RECEIVER] [STREAM_ID]=%s Open udp server failed, error: %s\n", this.AbsStreamId, err.Error())
		return err
	}
	this.connect, err = net.ListenUDP("udp", addr)
	if err != nil {
		glog.Warningf("[UDP_RECEIVER] [STREAM_ID]=%s Open udp server failed, error: %s\n", this.AbsStreamId, err.Error())
		return err
	}

	glog.V(3).Infof("[UDP_RECEIVER] [STREAM_ID]=%s Open udp server successed\n", this.AbsStreamId)
	return nil
}

// Receive h264 element stream packet with udp
func (this *UDPReceiver) receiveFrames() error {

	data := make([]byte, 1440)
	var fps uint32 = 30
	var index uint64 = 0
	var ts uint64 = 0
	var t0 time.Time
	var buffer bytes.Buffer
	init := false

	for this.AbsRunning {
		n, _, err := this.connect.ReadFromUDP(data)
		if err != nil {
			return err
		}
		if data[0] == 0x00 && data[1] == 0x00 && data[2] == 0x00 && data[3] == 0x01 {
			if init == false {
				init = true
				buffer.Write(data[:n])
			} else {
				index++
				iFrame := false
				if len(buffer.Bytes()) >= 4 && (buffer.Bytes()[4]&0x1F == 7 || buffer.Bytes()[4]&0x1F == 8 || buffer.Bytes()[4]&0x1F == 5) {
					iFrame = true
				}
				if index%100 == 0 {
					f, _ := strconv.ParseFloat(fmt.Sprintf("%0.2f", 1.0/(float64(time.Now().Sub(t0).Nanoseconds())/1000000000.00)), 64)
					if f*100 > 0 {
						fps = uint32(f * 100)
					}
					t0 = time.Now()
				}

				buf := make([]byte, buffer.Len())
				copy(buf, buffer.Bytes())
				ts += uint64(90000 / fps)

				frame := &core.H264ESFrame{
					Data:      buf,
					Timestamp: uint32(ts),
					IFrame:    iFrame,
					Index:     index,
					// Fps:       fps,
				}
				this.AbsFPS = fps
				if iFrame == true {
					sp := h264.GetLiveSPS(buffer.Bytes())
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
					buffer.Reset()
					init = true
					buffer.Write(data[:n])
				}
			}
		} else {
			if !init {
				continue
			}
			buffer.Write(data[:n])
		}
	}

	return nil
}
