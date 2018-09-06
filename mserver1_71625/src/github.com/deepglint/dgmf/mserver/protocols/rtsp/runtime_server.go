package rtsp

import (
	"log"
	"time"

	"github.com/deepglint/dgmf/mserver/core"
	"github.com/deepglint/dgmf/mserver/protocols/rtcp"
	"github.com/deepglint/dgmf/mserver/protocols/rtp"
	"github.com/deepglint/dgmf/mserver/utils"
)

func (this *RTSPServer) runtimeRequrst(sessionCtx *SessionContext) {
	pool := core.GetESPool()

	for {
		buf := make([]byte, REQ_RSP_SIZE)
		_, err := sessionCtx.TCPConnect.Read(buf)
		if err != nil {
			if sessionCtx.StreamType == "live" {
				pool.Live.RemoveSession(sessionCtx.StreamId, sessionCtx.SessionId)
			} else if sessionCtx.StreamType == "proxy" {
				pool.Proxy.RemoveStream(sessionCtx.StreamId)
			}
			break
		}

		receiverReport := rtcp.ReceiverReport{}
		err = receiverReport.Unmarshal(buf)
		if err != nil {
			sr := rtcp.SenderReport{
				Version:      2,
				Padding:      false,
				ReportCount:  0,
				PacketType:   200,
				SSRC:         sessionCtx.SSRC,
				NTPTimestamp: uint64(time.Now().Unix()),
				RTPTimestamp: sessionCtx.Timestamp,
			}
			srd, _ := sr.Marshal()
			inv := RTSPInterleavedFrame{}
			inv.Channel = 0x01
			inv.Length = uint16(len(srd))
			send := []byte{}
			send = append(send, inv.Marshal()...)
			send = append(send, srd...)
			sessionCtx.TCPConnect.Write(send)
		}

		// request := RTSPRequest{}
		// err = request.Unmarshal(strings.Split(string(buf), "\r\n\r\n")[0] + "\r\n\r\n")
		// if err == nil {
		// 	log.Println("[RTSP_SERVER] Request:\n" + request.Marshal())
		// 	this.teardown(request, sessionCtx)
		// 	break
		// }
	}
}

func (this *RTSPServer) teardown(request RTSPRequest, sessionCtx *SessionContext) error {
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

	pool := core.GetESPool()
	if sessionCtx.StreamType == "live" {
		pool.Live.RemoveSession(sessionCtx.StreamId, sessionCtx.SessionId)
	} else if sessionCtx.StreamType == "proxy" {
		pool.Proxy.RemoveStream(sessionCtx.StreamId)
	}

	return nil
}

func (this *RTSPServer) sendFrames(sessionCtx *SessionContext) {
	pool := core.GetESPool()
	var seq uint16 = 0

	if sessionCtx.StreamType == "live" {
		pool.Live.AddSession(sessionCtx.StreamId, sessionCtx.SessionId, "rtsp", sessionCtx.TCPConnect)
		find := false
		frames, err := pool.Live.GetFrames(sessionCtx.StreamId, sessionCtx.SessionId)
		stream, err := pool.Live.GetStream(sessionCtx.StreamId)
		if err == nil && stream != nil && frames != nil {
			for frame := range frames {
				if find == false && frame.IFrame == true {
					find = true
				}
				if find == true {
					sendPacket(frame, sessionCtx, sessionCtx.SSRC, &seq)
				} else {
					if len(stream.IFrame.Data) != 0 {
						stream.IFrame.Timestamp = frame.Timestamp
						sendPacket(&stream.IFrame, sessionCtx, sessionCtx.SSRC, &seq)
					}
				}
			}
		}
	} else if sessionCtx.StreamType == "vod" {

	} else if sessionCtx.StreamType == "proxy" {
		frames, err := pool.Proxy.GetFrames(sessionCtx.StreamId)
		runProxy := true
		if err == nil && frames != nil {
			for runProxy {
				select {
				case frame := <-frames:
					if frame != nil {
						sendPacket(frame, sessionCtx, sessionCtx.SSRC, &seq)
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
		sessionCtx.TCPConnect.Close()
		pool.Proxy.RemoveStream(sessionCtx.StreamId)
	}
}

func sendPacket(frame *core.H264ESFrame, sessionCtx *SessionContext, ssrc uint32, seq *uint16) {
	data := frame.Data[4:]
	if len(data) <= MAX_RTP_SIZE {
		*seq++
		rtpPacket := rtp.RTPPacket{
			Header: rtp.RTPHeader{
				Version:        2,
				Padding:        false,
				Extend:         false,
				CSRCCount:      0,
				Marker:         true,
				PayloadType:    96,
				SequenceNumber: *seq,
				Timestamp:      frame.Timestamp,
				SSRC:           ssrc,
			},
		}
		sessionCtx.Timestamp = frame.Timestamp

		naluHeader := rtp.NALUHeader{}
		naluHeader.Unmarshal(data[0])

		rtpNalu := rtp.RTPNALUPacket{
			RTPNALUHeader: naluHeader,
			Payload:       data[1:],
		}

		rtpPacket.Payload, _ = rtpNalu.Marshal()
		rtpBuf, _ := rtpPacket.Marshal()

		inv := RTSPInterleavedFrame{}
		inv.Channel = 0
		inv.Length = uint16(len(rtpBuf))
		send := []byte{}
		send = append(send, inv.Marshal()...)
		send = append(send, rtpBuf...)
		if sessionCtx.IsUDP {
			sessionCtx.UDPConnect.Write(rtpBuf)
		} else {
			sessionCtx.TCPConnect.Write(send)
		}
	} else {
		all := len(data)
		part := 0

		naluHeader := rtp.NALUHeader{}
		naluHeader.Unmarshal(data[0])

		for {
			if all > MAX_RTP_SIZE {
				*seq++
				rtpPacket := rtp.RTPPacket{
					Header: rtp.RTPHeader{
						Version:        2,
						Padding:        false,
						Extend:         false,
						CSRCCount:      0,
						Marker:         false,
						PayloadType:    96,
						SequenceNumber: *seq,
						Timestamp:      frame.Timestamp,
						SSRC:           ssrc,
					},
				}
				sessionCtx.Timestamp = frame.Timestamp

				var rtpNalu rtp.RTPNALUPacket
				if part == 0 {
					rtpNalu.RTPNALUHeader = rtp.NALUHeader{
						NRI:  naluHeader.NRI,
						Type: 28,
					}

					rtpNalu.RTPFUHeader = rtp.FUHeader{
						S:    true,
						E:    false,
						R:    false,
						Type: naluHeader.Type,
					}
				} else {
					rtpNalu.RTPNALUHeader = rtp.NALUHeader{
						NRI:  naluHeader.NRI,
						Type: 28,
					}

					rtpNalu.RTPFUHeader = rtp.FUHeader{
						S:    false,
						E:    false,
						R:    false,
						Type: naluHeader.Type,
					}
				}

				rtpNalu.Payload = data[1+MAX_RTP_SIZE*part : 1+MAX_RTP_SIZE*(part+1)]

				rtpPacket.Payload, _ = rtpNalu.Marshal()
				rtpBuf, _ := rtpPacket.Marshal()

				inv := RTSPInterleavedFrame{}
				inv.Channel = 0
				inv.Length = uint16(len(rtpBuf))
				send := []byte{}
				send = append(send, inv.Marshal()...)
				send = append(send, rtpBuf...)
				if sessionCtx.IsUDP {
					sessionCtx.UDPConnect.Write(rtpBuf)
				} else {
					sessionCtx.TCPConnect.Write(send)
				}

				part++
				all -= MAX_RTP_SIZE
			} else {
				*seq++
				rtpPacket := rtp.RTPPacket{
					Header: rtp.RTPHeader{
						Version:        2,
						Padding:        false,
						Extend:         false,
						CSRCCount:      0,
						Marker:         true,
						PayloadType:    96,
						SequenceNumber: *seq,
						Timestamp:      frame.Timestamp,
						SSRC:           ssrc,
					},
				}
				sessionCtx.Timestamp = frame.Timestamp
				var rtpNalu rtp.RTPNALUPacket
				rtpNalu.RTPNALUHeader = rtp.NALUHeader{
					NRI:  naluHeader.NRI,
					Type: 28,
				}

				rtpNalu.RTPFUHeader = rtp.FUHeader{
					S:    false,
					E:    true,
					R:    false,
					Type: naluHeader.Type,
				}
				rtpNalu.Payload = data[1+MAX_RTP_SIZE*part:]

				rtpPacket.Payload, _ = rtpNalu.Marshal()
				rtpBuf, _ := rtpPacket.Marshal()

				inv := RTSPInterleavedFrame{}
				inv.Channel = 0
				inv.Length = uint16(len(rtpBuf))
				send := []byte{}
				send = append(send, inv.Marshal()...)
				send = append(send, rtpBuf...)
				if sessionCtx.IsUDP {
					sessionCtx.UDPConnect.Write(rtpBuf)
				} else {
					sessionCtx.TCPConnect.Write(send)
				}
				break
			}
		}
	}
}
