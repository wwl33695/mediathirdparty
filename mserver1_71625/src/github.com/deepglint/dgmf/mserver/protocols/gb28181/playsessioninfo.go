package gb28181

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"sync"

	"github.com/deepglint/dgmf/mserver/core"
	"github.com/deepglint/dgmf/mserver/protocols/ps"
	"github.com/deepglint/dgmf/mserver/protocols/rtp"
	"github.com/deepglint/dgmf/mserver/utils/uuid"
)

const (
	RTP_MAX_LENGTH = 1400
)

//RTPSession struct of RTP Session
type PlaySessionInfo struct {
	CallID              string
	LiveStreamID        string
	LocalIP             string
	LocalPort           int
	RTPServerIP         string
	RTPServerPort       int
	SSRC                string
	Channel             *core.GB28181Channel
	isLive              bool
	rtpConnect          net.Conn
	liveStreamSessionID string
	sequenceNumber      uint16
	psMuxer             ps.PSMuxer
	playSeeionInfoLock  *sync.Mutex
	waitgroup           *sync.WaitGroup
}

func NewPlaySeesionInfo(callID string, liveStreamID string, localIP string, localPort int, rtpRemoteIP string, rtpRemotePort int, ssrc string, channel *core.GB28181Channel) *PlaySessionInfo {

	var lock sync.Mutex
	playSeesionInfo := PlaySessionInfo{
		CallID:             callID,
		LiveStreamID:       liveStreamID,
		LocalIP:            localIP,
		RTPServerIP:        rtpRemoteIP,
		RTPServerPort:      rtpRemotePort,
		SSRC:               ssrc,
		Channel:            channel,
		sequenceNumber:     0,
		isLive:             false,
		playSeeionInfoLock: &lock,
		waitgroup:          &sync.WaitGroup{},
	}

	//Initialize ps muMuxer
	playSeesionInfo.psMuxer = ps.NewPSMuxer(playSeesionInfo.psMuxerCallBack)

	return &playSeesionInfo
}

func (this *PlaySessionInfo) Start() {
	fmt.Println("open the rtp id", this.CallID)

	go func(this *PlaySessionInfo) {
		defer fmt.Println("getFrameThread close")

		//avoid start mutily get frame thread
		this.playSeeionInfoLock.Lock()
		if this.isLive {
			fmt.Println("suchannel ", this.CallID, "has open")
			this.playSeeionInfoLock.Unlock()
			return
		}

		if this.rtpConnect != nil {
			this.rtpConnect.Close()
		}

		var err error
		// LocalAddr, _ := net.ResolveUDPAddr("udp4", localIp)
		RemoteEP := net.UDPAddr{IP: net.ParseIP(this.RTPServerIP), Port: (int)(this.RTPServerPort)}
		this.rtpConnect, err = net.DialUDP("udp", nil, &RemoteEP)
		if err != nil {
			log.Printf("strat PRTSession error %v", err)
			this.playSeeionInfoLock.Unlock()
			return
		}
		//close rtpConnect when exit
		defer func() {
			this.rtpConnect.Close()
			this.rtpConnect = nil
			fmt.Println("rtp close")
		}()

		pool := core.GetESPool()

		if pool.Live.Maps[this.LiveStreamID] == nil {
			fmt.Println("video streamId", this.LiveStreamID, " not exist")
			this.playSeeionInfoLock.Unlock()
			return
		}

		this.liveStreamSessionID = uuid.NewV4().String()
		pool.Live.AddSession(this.LiveStreamID, this.liveStreamSessionID, "gb28181", this.rtpConnect)
		if err != nil {
			fmt.Println(this.LiveStreamID, " add session failed")
			this.playSeeionInfoLock.Unlock()
			return
		}

		//release liveStreamSession when exit
		defer func() {
			// pool.RemoveSession(this.LiveStreamID, true, this.liveStreamSessionID)
			this.playSeeionInfoLock.Lock()
			this.isLive = false
			fmt.Println("delete livestreamSession")
			this.playSeeionInfoLock.Unlock()
		}()

		frame := pool.Live.Maps[this.LiveStreamID].Sessions[this.liveStreamSessionID].Frame

		this.isLive = true
		this.playSeeionInfoLock.Unlock()
		for esData := range frame {
			if esData.Data != nil {
				this.psMuxer.Mux(esData.Data, uint32(len(esData.Data)), uint64(esData.Timestamp), 0, esData.IFrame)
			}
		}
	}(this)
}

//psMuxerCallBack ps callback
func (this *PlaySessionInfo) psMuxerCallBack(out []byte, length uint32) {

	this.playSeeionInfoLock.Lock()
	defer this.playSeeionInfoLock.Unlock()

	ssrcInt, err := strconv.Atoi(this.SSRC)
	if err != nil {
		fmt.Printf("convert ssrc error ssrc=%s", this.SSRC)
		return
	}

	sentDataFunc := func() {
		var offset = 0
		for i := 0; i < int(length)/RTP_MAX_LENGTH; i++ {
			header := rtp.RTPHeader{
				Version:        2,
				Padding:        false,
				Extend:         false,
				CSRCCount:      0,
				Marker:         false,
				PayloadType:    96,
				SequenceNumber: this.sequenceNumber,
				Timestamp:      0,
				SSRC:           uint32(ssrcInt),
			}
			this.sequenceNumber++

			rtpPacket := rtp.RTPPacket{
				Header:  header,
				Payload: out[offset : offset+RTP_MAX_LENGTH],
			}

			data, _ := rtpPacket.Marshal()

			if this.rtpConnect != nil {
				this.rtpConnect.Write(data)
			}
			// fmt.Printf("[SIZE] %d\n", RTP_MAX_LENGTH)
			offset += RTP_MAX_LENGTH

		}
		header := rtp.RTPHeader{
			Version:        2,
			Padding:        false,
			Extend:         false,
			CSRCCount:      0,
			Marker:         true,
			PayloadType:    96,
			SequenceNumber: this.sequenceNumber,
			Timestamp:      0,
			SSRC:           uint32(ssrcInt),
		}
		this.sequenceNumber++
		rtpPacket := rtp.RTPPacket{
			Header:  header,
			Payload: out[offset:length],
		}

		data, _ := rtpPacket.Marshal()
		if this.rtpConnect != nil {
			this.rtpConnect.Write(data)
		}
		// fmt.Printf("[SIZE] %d\n", int(length)-offset)
		this.waitgroup.Done()
	}

	// open multily thread to sent data
	this.waitgroup.Add(1)
	go sentDataFunc()

	//wait all thread close
	this.waitgroup.Wait()

}

func (this *PlaySessionInfo) Stop() {
	fmt.Println("close playSession", this.CallID)
	this.playSeeionInfoLock.Lock()
	defer this.playSeeionInfoLock.Unlock()

	if this.isLive {
		pool := core.GetESPool()
		pool.Live.RemoveSession(this.LiveStreamID, this.liveStreamSessionID)
	}
}
