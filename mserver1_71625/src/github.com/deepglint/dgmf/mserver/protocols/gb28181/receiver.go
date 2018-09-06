	package gb28181

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"strings"
//	"os/signal"
	"errors"
	"strconv"
	"time"
	"github.com/deepglint/dgmf/mserver/core"
	"github.com/deepglint/dgmf/mserver/protocols/h264"
//	"github.com/deepglint/dgmf/mserver/protocols/sdp"
)

const (
	UNKNOWN = iota
	LOGIN
	REGISTER
	QUERYDEVICE
	LIVEPLAY
	BREAKPLAY
	LOGOUT
)

type GB28181Receiver struct {
	core.Receiver
	uac			   *UASInfo
	conn           net.Conn
	mediaconn 	   *net.UDPConn

	fps 		   uint32
	index 		   uint64
	ts 			   uint64
	t0 			   time.Time
	buffer 		   bytes.Buffer
	init 		   bool
	databuffer	   [2000 * 1024]byte
	offset		   int
	channRecv	   chan string
	status 		   int
}

func ParseURI(uri string, uac *UASInfo ) {
//	ttt := "gb28181://192.168.1.176:5065:15010000004000000001:123456@192.168.6.105:5060:11000000002000000001/34020000001310000001";

	pos := strings.Index(uri, "//")
	temp := uri[pos+2:]
	pos = strings.Index(temp, ":")

//	println(temp[:pos])
	uac.ClientIP = temp[:pos]

	pos++
	pos1 := strings.Index(temp[pos:], ":")
//	println(temp[pos:pos+pos1])
	uac.ClientPort = temp[pos:pos+pos1]

	pos = pos + pos1 + 1
	pos1 = strings.Index(temp[pos:], ":")
//	println(temp[pos:pos+pos1])
	uac.UserName = temp[pos:pos+pos1]

	pos = pos + pos1 + 1
	pos1 = strings.Index(temp[pos:], "@")
//	println(temp[pos:pos+pos1])
	uac.Password = temp[pos:pos+pos1]

	pos = pos + pos1 + 1
	pos1 = strings.Index(temp[pos:], ":")
//	println(temp[pos:pos+pos1])
	uac.ServerIP = temp[pos:pos+pos1]

	pos = pos + pos1 + 1
	pos1 = strings.Index(temp[pos:], ":")
//	println(temp[pos:pos+pos1])
	uac.ServerPort = temp[pos:pos+pos1]

	pos = pos + pos1 + 1
	pos1 = strings.Index(temp[pos:], "/")
//	println(temp[pos:pos+pos1])
	uac.ServerID = temp[pos:pos+pos1]

	pos = pos + pos1 + 1
//	println(temp[pos:])
	uac.ChannelID = temp[pos:]
}

func (this *GB28181Receiver) Open(uri string, streamId string, rtms chan core.RTMessage) {
	this.fps = 30
	this.index = 0
	this.ts = 0
	this.init = false
	this.AbsRunning = false
	this.AbsFrames = make(chan *core.H264ESFrame)
	this.AbsStoped = make(chan bool)
	this.AbsRTMessages = rtms
	this.AbsStreamId = streamId
	this.AbsIndex = 0
	this.offset = 0
	this.channRecv = make(chan string)

	this.uac = &UASInfo{}
	ParseURI(uri, this.uac)

/*		ServerID:   "11000000002000000001",
		ServerIP:   "192.168.6.105",
		ServerPort: "5060",
		UserName:   "15010000004000000001",
		Password:   "123456",
		ClientIP:   "192.168.1.176",
		ClientPort: "5065",
		ChannelID:	"34020000001310000001"}
*/

	conn, err := net.Dial("udp", this.uac.ServerIP+":"+ this.uac.ServerPort)
	if err != nil {
		println("build socket failed")
		return
	}
//	defer conn.Close()

	this.conn = conn

	this.AbsRunning = true
	go this.RecvProc(conn)
	this.SipEventProc(conn, this.uac)

}

func (this *GB28181Receiver) Recv() string{
	select {
		case recvData := <- this.channRecv:
			return recvData
		default:
			return ""
	}

	return ""
}

func (this *GB28181Receiver) RecvProc(conn net.Conn) {

	buf := make([]byte, 2048)
	for this.AbsRunning {
		length, err := conn.Read(buf)
		if err == nil {
//			println(string(buf[:length]))
		} else {
			println("recvproc read error, break the loop", length)
			break
		}

		this.channRecv <- string(buf[:length])
	}

	close(this.channRecv)
	println("exit recvproc routine")
}

func (this *GB28181Receiver) Close() {
	request := this.uac.BuildBYERequest(this.uac.ChannelID)
	println(request)
	for i := 0; i < 5; i++ {
		this.conn.Write([]byte(request))		
	}

	this.AbsRunning = false

	if this.mediaconn != nil {
		this.mediaconn.Close()
	}
	if this.conn != nil {
		this.conn.Close()
	}
}

func (this *GB28181Receiver) SipEventProc(conn net.Conn, uac *UASInfo) {
this.status = REGISTER
	var times int = 0

	for this.AbsRunning {
		if this.status == REGISTER {
			iRet := this.Register(conn, uac)
			if iRet == 0 {
				this.status = LIVEPLAY
			}
		} else if this.status == LIVEPLAY {
			iRet := this.LivePlay(conn, uac)
			if iRet == 0 {
				this.status = UNKNOWN				
			}
		} else if this.status == UNKNOWN {
//			println("state=UNKNOWN--------------")

			if times % 60 == 0 {
				this.Heartbeat(conn, uac)
			}

			if( times >= 10 * 60 ) {
				times = 0
				this.Register(conn, uac)
			}
			times++

			time.Sleep(time.Second * 1)
		}
	}

	println("exit sipeventproc routine")
}

func (this *GB28181Receiver) GetRTPPort(uac *UASInfo) int {
	for port := 20000; port < 30000 ; port += 2 {
		uac.LocalMediaPort = strconv.Itoa(port)
		udp_addr, err := net.ResolveUDPAddr("udp", uac.ClientIP + ":" + uac.LocalMediaPort)
		if err != nil {
			println("MediaStreamProc ResolveUDPAddr failed")
			return -1
		}
		this.mediaconn, err = net.ListenUDP("udp", udp_addr)
		if err != nil {
			println("MediaStreamProc ListenUDP failed")
			continue
		}

		println(uac.LocalMediaPort)
		return 0
	}

//	defer conn.Close()
		return -1
}

func (this *GB28181Receiver) Heartbeat(conn net.Conn, uac *UASInfo) {
//	buf := make([]byte, 2048)
	var times int = 0

	request := uac.BuildHeartbeat()
	println(request)
	conn.Write([]byte(request))

	for {
		time.Sleep(time.Second * 1)

		if times >= 5 {
			break
		}
		times++

		recvData := this.Recv()
		if recvData != "" {
			println(recvData)
			break
		}
	}
/*	length, err := conn.Read(buf)
	if err == nil {
		println(string(buf[:length]))
	} else {
		println("UNKNOWN read error")
	}
*/
}

func (this *GB28181Receiver) Register(conn net.Conn, uac *UASInfo) int {
	var recvData string
	var times int = 0

	request := uac.BuildRegisterRequest()
	println(request)
	conn.Write([]byte(request))

	for {
		time.Sleep(time.Second * 1)
		if times >= 5 {
			return -1
		}
		times++

		recvData = this.Recv()
		if recvData != "" {
			println(recvData)
		}

		if recvData == "" || !MatchResponse("REGISTER1", recvData){
			continue
		} else {
			break
		} 
	}

	errCode, _ := ParseResponseHead(recvData)
	if errCode != "401" {
		return -1
	}

	realm, nonce, _, _ := ParseRegister1(recvData)
	request = uac.BuildRegisterMD5Auth(realm, nonce)
	println(request)
	conn.Write([]byte(request))
	times = 0

	for {
		time.Sleep(time.Second * 1)
		if times >= 5 {
			return -1
		}
		times++

		recvData = this.Recv()
		if recvData != "" {
			println(recvData)
		}
		if recvData == "" || !MatchResponse("REGISTER2", recvData){
			continue
		}

		errCode, err := ParseResponseHead(recvData)
		if err == nil && errCode == "200" {
			return 0
		}			
	}

	return -1
}

func (this *GB28181Receiver) LivePlay(conn net.Conn, uac *UASInfo) int {
	var recvData string
	var times int = 0

	iRet := this.GetRTPPort(this.uac)
	if iRet < 0 {
		println("getrtpport failed")
		return -1
	}

	request := uac.BuildInviteRequest(uac.LocalMediaPort, uac.ChannelID)
	println(request)
	conn.Write([]byte(request))

	for {
		time.Sleep(time.Second * 1)
		if times >= 5 {
			return -1
		}
		times++

		recvData = this.Recv()
		if recvData != "" {
			println(recvData)
		}

		if recvData == ""  || !MatchResponse("INVITE", recvData){
			continue
		}

		if !MatchResponse("INVITE", recvData) {
			continue
		}

		errCode, _ := ParseResponseHead(recvData)
		if errCode == "100" {
			continue
		} else if errCode == "200" {
			remoteMediaIP, remoteMediaPort, remoteSSRC, totag, _ := ParseResponseInvite(recvData)
			uac.RemoteMediaPort = remoteMediaPort
			uac.RemoteMediaIP = remoteMediaIP
			uac.RemoteSSRC = remoteSSRC
			uac.PlayToTag = totag
			break
		}
	}

	request = uac.BuildACKRequest(uac.ChannelID)
	println(request)
	for i:= 0;i<5; i++ {
		conn.Write([]byte(request))		
	}		

	go this.MediaStreamProc(uac)
//	go this.IdleProc(conn, uac)

	times = 0
	for {
		time.Sleep(time.Second * 2)
		if times >= 5 {
			return -1
		}
		times++
	}

	return -1
}

func (this *GB28181Receiver) MediaStreamProc(uac *UASInfo) {
	remotertpport, _ := strconv.Atoi(uac.RemoteMediaPort)
	remotertpport++
	remotertcpport := strconv.Itoa(remotertpport)
	connRTCP, errRTCP := net.Dial("udp", uac.RemoteMediaIP + ":" + remotertcpport)
	if errRTCP != nil {
		println("MediaStreamProc connRTCP failed")
		return
	}
	defer connRTCP.Close()

	remotessrc, _ := strconv.Atoi(uac.RemoteSSRC)
	localssrc, _ := strconv.Atoi(uac.LocalSSRC)

//	println("-----------------", remotessrc, uac.RemoteSSRC, localssrc, uac.LocalSSRC)

	if errRTCP != nil {
		println("build RTCP socket failed")
		return
	}

//	file, err := os.OpenFile("111.264", os.O_WRONLY|os.O_CREATE, 0666)
//	file, err := os.OpenFile("111.ps", os.O_WRONLY|os.O_CREATE, 0666)
/*	if err != nil {
		println("open file failed.", err.Error())
		return
	}
	defer file.Close()
*/
	select {
		case this.AbsRTMessages <- core.RTMessage{Status: 200}:
		default:
	}

	var maxseqnum uint16
	var packetsnum uint
	buf := make([]byte, 2048)
	for this.AbsRunning {
		length, _, err := this.mediaconn.ReadFromUDP(buf)
		if err == nil {
			this.status = UNKNOWN

			var seqnum uint16
			bufTemp := bytes.NewBuffer(buf[2:])
			binary.Read(bufTemp, binary.BigEndian, &(seqnum))
			if maxseqnum < seqnum {
				maxseqnum = seqnum
			}

			packetsnum++
//			println("MediaStreamProc", seqnum, buf[0], buf[1], buf[2], buf[3], length)

			if packetsnum > 5000 {
				bufRTCP := GetRR(uint32(localssrc), uint32(remotessrc), maxseqnum)
				connRTCP.Write([]byte(bufRTCP))
				packetsnum = 0
			}

			this.SetData(buf[12:length], nil)
//			this.SetData(buf[12:length], file)
//			file.Write(buf[12:length])
		} else {
			println("MediaStreamProc read error, break the loop")
			break
		}

		//		time.Sleep(5e6)
	}

	println("exit mediastreamproc routine")
}

func ParsePacket(buffer []byte) (pesbuf []byte, peslen, startpos int) {
	var nFirstPesPos int = -1
	var i int

	for i = 0; i < len(buffer)-4; i++ {
		if (buffer[i]) == (0) &&
			(buffer[i+1]) == (0) &&
			(buffer[i+2]) == (1) &&
			(buffer[i+3]) == (0xe0) {
			nFirstPesPos = i
			i++
			break
		}
	}
	if nFirstPesPos < 0 {
//		println("nFirstPesPos < 0 ")
		return nil, 0, -1
	}

	var nPesEndPos int = -1
	for ; i < len(buffer)-4; i++ {
		if buffer[i] == 0 && buffer[i+1] == 0 && buffer[i+2] == 1 &&
			(buffer[i+3] == 0xba || buffer[i+3] >= 0xc0 ) {
			nPesEndPos = i
			break
		}
	}
	if nPesEndPos < 0 {
//		println("nPesEndPos < 0 ")
		return nil, 0, -1
	}

	return buffer[nFirstPesPos:nPesEndPos], nPesEndPos - nFirstPesPos, nFirstPesPos
}

func (this *GB28181Receiver) SetData(data []byte, file *os.File) {
	if this.offset > len(this.databuffer) {
		println("databuffer overflow ")
		return
	}

	copy(this.databuffer[this.offset:], data)
	this.offset += len(data)
//	if this.offset < 100*1024 {
//		return
//	}

	retbuf, peslen, startpos := ParsePacket(this.databuffer[0:this.offset])
	for peslen > 0 {
		pesheaderlen := 9 + uint8(retbuf[8])
		buf := retbuf[pesheaderlen:]
		this.receiveFrames(buf[0:], len(buf))
//		file.Write(buf[0:])

		copy(this.databuffer[0:], this.databuffer[startpos+peslen:])
		this.offset = this.offset - startpos - peslen
		retbuf, peslen, startpos = ParsePacket(this.databuffer[0:this.offset])
	}
}

// Receive h264 element stream packet with udp
func (this *GB28181Receiver) receiveFrames(data []byte, n int) error {
	if n <= 4 {
		println("receiveFrames datalen <= 4")
		return errors.New("receiveFrames datalen <= 4")
	}

	if data[0] == 0x00 && data[1] == 0x00 && data[2] == 0x00 && data[3] == 0x01 {
		if this.init == false {
			println("receiveFrames nalu got")
			this.init = true
			this.buffer.Write(data[:n])
		} else {
			this.index++
			iFrame := false
			if len(this.buffer.Bytes()) >= 4 && (this.buffer.Bytes()[4]&0x1F == 7 || this.buffer.Bytes()[4]&0x1F == 8 || this.buffer.Bytes()[4]&0x1F == 5) {
				iFrame = true
			}
			if this.index%100 == 0 {
				f, _ := strconv.ParseFloat(fmt.Sprintf("%0.2f", 1.0/(float64(time.Now().Sub(this.t0).Nanoseconds())/1000000000.00)), 64)
				if f*100 > 0 {
					this.fps = uint32(f * 100)
				}
				this.t0 = time.Now()
			}

			buf := make([]byte, this.buffer.Len())
			copy(buf, this.buffer.Bytes())
			this.ts += uint64(90000 / this.fps)

			frame := &core.H264ESFrame{
				Data:      buf,
				Timestamp: uint32(this.ts),
				IFrame:    iFrame,
				Index:     this.index,
			}
			this.AbsFPS = this.fps
			if iFrame == true {
				sp := h264.GetLiveSPS(this.buffer.Bytes())
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
				this.buffer.Reset()
				this.init = true
				this.buffer.Write(data[:n])
			}
		}
	} else {
		if !this.init {
			return errors.New("rawdata hasn't been inited")
		}
		this.buffer.Write(data[:n])
	}

	return nil
}
