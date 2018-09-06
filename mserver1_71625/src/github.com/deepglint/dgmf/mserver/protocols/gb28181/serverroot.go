package gb28181

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"os/exec"
	"strings"
	"time"

	"github.com/deepglint/dgmf/mserver/core"
	"github.com/deepglint/dgmf/mserver/protocols/gb28181/utils"
	"github.com/deepglint/dgmf/mserver/protocols/manscdp/manscdpbase"
	"github.com/deepglint/dgmf/mserver/protocols/manscdp/manscdpparser"
	"github.com/deepglint/dgmf/mserver/protocols/sdp"
	"github.com/deepglint/dgmf/mserver/protocols/sip/base"
	"github.com/deepglint/dgmf/mserver/protocols/sip/parser"
)

//ipcamera server root
func (this *GB28181Server) requestRoot(udpConn *net.UDPConn, serverAddr *net.UDPAddr, data []byte) {
	simMessage, err := parser.ParseMessage(data)
	if err != nil {
		fmt.Printf("Some error %v", err)
	} else {
		if value, ok := (simMessage).(*(base.Request)); ok {
			switch value.Method {
			case "MESSAGE":
				this.messageHandle(udpConn, serverAddr, value)
			case "INVITE":
				this.inviteHandle(udpConn, serverAddr, value)
			case "ACK":
				this.ackHandle(udpConn, serverAddr, value)
			case "BYE":
				this.byeHandle(udpConn, serverAddr, value)
			}
		}
	}
}

//Message handle func
func (this *GB28181Server) messageHandle(udpConn *net.UDPConn, serverAddr *net.UDPAddr, request *base.Request) {
	manscdp, err := manscdpparser.ParseManscdp(([]byte)(request.Body))
	if err != nil {
		log.Printf("Some error %v", err)
	} else {

		switch manscdp.(type) {
		// go to Query mode
		case *manscdpbase.Query:
			if value, ok := (manscdp).(*(manscdpbase.Query)); ok {
				switch strings.ToLower(value.CmdType) {
				case "catalog":
					udpConn.WriteToUDP([]byte(this.createSipResponse(request, 200, "OK", "").String()), serverAddr)
					this.sendCatalogRequest(value.SN)
				case "deviceinfo":
					udpConn.WriteToUDP([]byte(this.createSipResponse(request, 200, "OK", "").String()), serverAddr)
					this.sendDeviceInfoRequest(value.SN)
				case "devicestatus":
					udpConn.WriteToUDP([]byte(this.createSipResponse(request, 200, "OK", "").String()), serverAddr)
					this.sendDeviceStatusRequest(value.SN)
				}
			}
		//go to Control mode
		case *manscdpbase.Control:
			if value, ok := (manscdp).(*(manscdpbase.Control)); ok {
				udpConn.WriteToUDP([]byte(this.createSipResponse(request, 200, "OK", "").String()), serverAddr)
				if value.TeleBoot != "" {
					secondRegisterRequest, err := this.sendFirstRegisterRequest(0, 3)
					if err == nil {
						this.sendSecondRegisterRequest(secondRegisterRequest)
					}
					cmd := exec.Command("reboot")
					err = cmd.Run()
					if err != nil {
						log.Println("restart err:", err)
					}
				} else {

				}

			}

		}
	}
}

//Send DeviceStatus Request
func (this *GB28181Server) sendDeviceStatusRequest(sn uint64) {
	data := make([]byte, 2048)
	LocalAddr, err := net.ResolveUDPAddr("udp", this.localhost)
	RemoteEP := net.UDPAddr{
		IP:   net.ParseIP(this.param.ServerHost),
		Port: (int)(this.param.ServerPort),
	}
	conn, err := net.DialUDP("udp", LocalAddr, &RemoteEP)
	if err != nil {
		log.Printf("Some error %v", err)
		return
	}
	defer conn.Close()
	//set time out
	conn.SetReadDeadline(time.Now().Add(TIME_OUT * time.Second))
	// fmt.Fprintf(conn, "Hi UDP Server, How are you doing?")

	//create DeviceStatusRequest
	deviceStatusRequest := this.createDeviceStatusRequest(sn)
	//write DeviceStatusRequest
	conn.Write([]byte(deviceStatusRequest.String()))
	_, err = bufio.NewReader(conn).Read(data)
	if err == nil {
		//fmt.Printf("%s\n", p)
		simMessage, err := parser.ParseMessage(data)
		if err != nil {
			log.Printf("Some error %v", err)
		} else {

			if value, ok := (simMessage).(*(base.Response)); ok {
				if value.StatusCode != 200 {
					log.Printf("sent deviceStatus fail!")
				} else {
					fmt.Println("sent deviceStatus success!!")

				}
			}

		}
	} else {
		log.Printf("sent deviceStatus fail: %v", err)
	}
}

//Send DeviceInfo Request
func (this *GB28181Server) sendDeviceInfoRequest(sn uint64) {
	data := make([]byte, 2048)
	LocalAddr, err := net.ResolveUDPAddr("udp", this.localhost)
	RemoteEP := net.UDPAddr{
		IP:   net.ParseIP(this.param.ServerHost),
		Port: (int)(this.param.ServerPort),
	}
	conn, err := net.DialUDP("udp", LocalAddr, &RemoteEP)
	if err != nil {
		log.Printf("Some error %v", err)
		return
	}
	defer conn.Close()
	//set time out
	conn.SetReadDeadline(time.Now().Add(TIME_OUT * time.Second))
	// fmt.Fprintf(conn, "Hi UDP Server, How are you doing?")

	//create DeviceInfoRequest
	deviceInfoRequest := this.createDeviceInfoRequest(sn)
	//write DeviceInfoRequest
	conn.Write([]byte(deviceInfoRequest.String()))
	_, err = bufio.NewReader(conn).Read(data)
	if err == nil {
		//fmt.Printf("%s\n", p)
		simMessage, err := parser.ParseMessage(data)
		if err != nil {
			log.Printf("Some error %v", err)
		} else {

			if value, ok := (simMessage).(*(base.Response)); ok {
				if value.StatusCode != 200 {
					log.Printf("sent deviceInfo fail!")
				} else {
					fmt.Println("sent deviceInfo success!!")

				}
			}

		}
	} else {
		log.Printf("sent deviceInfo fail: %v", err)
	}
}

//Send Catalog Request
func (this *GB28181Server) sendCatalogRequest(sn uint64) {

	data := make([]byte, 2048)
	LocalAddr, err := net.ResolveUDPAddr("udp", this.localhost)
	RemoteEP := net.UDPAddr{
		IP:   net.ParseIP(this.param.ServerHost),
		Port: (int)(this.param.ServerPort),
	}
	conn, err := net.DialUDP("udp", LocalAddr, &RemoteEP)
	if err != nil {
		log.Printf("Some error %v", err)
		return
	}
	defer conn.Close()
	//set time out
	conn.SetReadDeadline(time.Now().Add(TIME_OUT * time.Second))
	// fmt.Fprintf(conn, "Hi UDP Server, How are you doing?")

	//create catalogRequest
	catalogRequest := this.createCatalogRequest(sn)
	//write catalogRequest
	conn.Write([]byte(catalogRequest.String()))
	_, err = bufio.NewReader(conn).Read(data)
	if err == nil {
		//fmt.Printf("%s\n", p)
		simMessage, err := parser.ParseMessage(data)
		if err != nil {
			log.Printf("Some error %v", err)
		} else {

			if value, ok := (simMessage).(*(base.Response)); ok {
				if value.StatusCode != 200 {
					log.Printf("sent catalog fail!")
				} else {
					fmt.Println("sent catalog success!!")

				}
			}

		}
	} else {
		log.Printf("sent catalog fail: %v", err)
	}

}

//send Register First Request
func (this *GB28181Server) sendFirstRegisterRequest(expiresTime uint32, seqNum uint32) (*base.Request, error) {

	data := make([]byte, 2048)

	// LocalAddr, err := net.ResolveUDPAddr("udp", this.localhost)
	// RemoteEP := net.UDPAddr{IP: net.ParseIP(this.param.ServerHost), Port: (int)(this.param.ServerPort)}

	// conn, err := net.DialUDP("udp", LocalAddr, &RemoteEP)
	// if err != nil {
	// 	return nil, err
	// }

	// defer func() {
	// 	conn.Close()
	// }()

	//firstRegistRequest
	firstRegistRequest := this.createFirstRegisterRequest(expiresTime, seqNum)

	//set read time out
	this.registerConn.SetReadDeadline(time.Now().Add(TIME_OUT * time.Second))

	// set write time out
	this.registerConn.SetWriteDeadline(time.Now().Add(TIME_OUT * time.Second))

	//write  firstRegistRequest
	_, err := this.registerConn.Write([]byte(firstRegistRequest.String()))
	if err != nil {
		return nil, err
	}

	//get response
	_, err = this.registerConn.Read(data)
	if err != nil {
		return nil, err
	}

	simMessage, err := parser.ParseMessage(data)
	if err != nil {
		return nil, err
	}

	if value, ok := (simMessage).(*(base.Response)); ok {
		if value.StatusCode != 401 {
			return nil, errors.New(fmt.Sprint("register response error code ", value.StatusCode))
		}
		authHeaders := simMessage.Headers("www-authenticate")
		if len(authHeaders) == 0 {
			return nil, errors.New(fmt.Sprint("can't get  www-authenticate header"))
		}
		if value, ok := (authHeaders[0]).(*(base.GenericHeader)); ok {
			nonce := utils.Between(value.Contents, "nonce=\"", "\"")
			if nonce == "" {
				return nil, errors.New(fmt.Sprint("register error: can't get nonce"))
			}
			//createSecondRegisterRequest
			secondRegistRequest := this.createSecondRegisterRequest(nonce, expiresTime, seqNum+1)
			return secondRegistRequest, nil

		} else {
			return nil, errors.New("www-authenticate header format err")
		}

	} else {
		return nil, errors.New("register response format sipmessage error")
	}

}

//send Register Second Request
func (this *GB28181Server) sendSecondRegisterRequest(secondRegistRequest *base.Request) error {

	data := make([]byte, 2048)

	// LocalAddr, err := net.ResolveUDPAddr("udp", this.localhost)
	// RemoteEP := net.UDPAddr{IP: net.ParseIP(this.param.ServerHost), Port: (int)(this.param.ServerPort)}

	// conn, err := net.DialUDP("udp", LocalAddr, &RemoteEP)
	// if err != nil {
	// 	return err
	// }

	// defer conn.Close()

	//set read time out
	this.registerConn.SetReadDeadline(time.Now().Add(TIME_OUT * time.Second))

	// set write time out
	this.registerConn.SetWriteDeadline(time.Now().Add(TIME_OUT * time.Second))

	//write createSecondRegisterRequest
	_, err := this.registerConn.Write([]byte(secondRegistRequest.String()))
	if err != nil {
		return err
	}

	//get response
	_, err = this.registerConn.Read(data)
	if err != nil {
		return err
	}

	simMessage, err := parser.ParseMessage(data)
	if err != nil {
		return err
	}

	if value, ok := (simMessage).(*(base.Response)); ok {
		if value.StatusCode != 200 {
			return errors.New(fmt.Sprintf("response statusCode", value.StatusCode))
		} else {
			//calibrate time
			dateHeaders := value.Headers("Date")
			if len(dateHeaders) > 0 {
				splits := strings.Split(dateHeaders[0].String(), "date:")
				if len(splits) > 1 {
					cmdStr := "date -s " + strings.TrimSpace(splits[1])
					list := strings.Split(cmdStr, " ")
					cmd := exec.Command(list[0], list[1:]...)
					err := cmd.Run()
					if err != nil {
						log.Println(cmdStr)
						log.Println("time calibration err:", err)
					}
				}
			}
			return nil
		}
	} else {
		return errors.New("parse sipMessage err")
	}
}

//send KeepAlive Request
func (this *GB28181Server) sendKeepAliveRequest() error {

	data := make([]byte, 2048)
	// LocalAddr, err := net.ResolveUDPAddr("udp", this.localhost)
	// RemoteEP := net.UDPAddr{
	// 	IP:   net.ParseIP(this.param.ServerHost),
	// 	Port: (int)(this.param.ServerPort),
	// }
	// conn, err := net.DialUDP("udp", LocalAddr, &RemoteEP)
	// if err != nil {
	// 	// log.Printf("Create Register Connect err", err)
	// 	return err
	// }
	// defer conn.Close()

	//createKeepLiveRequest

	keepAliveRequest := this.createKeepLiveRequest()

	//set read time out
	this.registerConn.SetReadDeadline(time.Now().Add(TIME_OUT * time.Second))

	// set write time out
	this.registerConn.SetWriteDeadline(time.Now().Add(TIME_OUT * time.Second))

	//write createKeepLiveRequest
	_, err := this.registerConn.Write([]byte(keepAliveRequest.String()))

	if err != nil {
		return err
	}

	// _, err = bufio.NewReader(conn).Read(data)
	//get response
	_, err = this.registerConn.Read(data)
	if err != nil {
		return err
	}

	//fmt.Printf("%s\n", p)
	simMessage, err := parser.ParseMessage(data)
	if err != nil {
		// log.Printf("keep connect error:", err)
		return err
	}
	if value, ok := (simMessage).(*(base.Response)); ok {
		if value.StatusCode != 200 {
			return errors.New(fmt.Sprintf("keep connect response code err:", value.StatusCode))
		} else {
			return nil
		}
	} else {
		return errors.New("parse sipMessage err ")
	}
}

//Invite handle func
func (this *GB28181Server) inviteHandle(udpConn *net.UDPConn, serverAddr *net.UDPAddr, request *base.Request) {

	var ssrc string
	var rtpRemoteIP string
	var rtpRemotePort uint16
	var callID string

	sdpMessage, err := sdp.ParseSdp(request.Body)

	//parse sdp fail
	if err != nil {
		log.Printf("解析sdp失败 %v", err)
		return
	}
	println(sdpMessage.String())
	udpConn.WriteToUDP([]byte(this.createSipResponse(request, 100, "Trying", "").String()), serverAddr)

	//get call-id
	if len(request.Headers("Call-Id")) == 0 {
		return
	}
	callIDHeader := request.Headers("Call-Id")[0]
	if value, ok := (callIDHeader).(*base.CallId); ok {
		callID = string(*value)
	}

	if len(sdpMessage.Media) == 0 {
		return
	}
	for _, media := range sdpMessage.Media {
		if strings.EqualFold(strings.ToLower(media.Type), "video") {
			//get ssrc
			ssrc = media.SSRC
			//get rtpRemotePort
			rtpRemotePort = uint16(media.Port)

		}
	}

	//get rtpRemoteIP
	rtpRemoteIP = sdpMessage.Origin.Address

	var sendDeviceID string = ""

	if len(request.Headers("subject")) == 0 {
		sendDeviceID = sdpMessage.Origin.Username
	} else {
		subject := request.Headers("subject")[0]
		if value, ok := (subject).(*(base.Subject)); ok {
			sendDeviceID = value.SendDeviceID
		}
	}

	// lock GBServer
	this.gbCameraServiceServerLock.Lock()
	defer this.gbCameraServiceServerLock.Unlock()

	//if callID not exist add one
	if _, ok := this.playSessionInfoMap[callID]; !ok {
		pool := core.GetESPool()
		for _, stream := range pool.Live.Maps {
			for _, output := range stream.Outputs {
				if strings.EqualFold(output.Protocol, "gb28181") {
					if subChannel, ok := output.Param.(core.GB28181Channel); ok {
						if strings.EqualFold(subChannel.ID, sendDeviceID) {
							playSessionInfo := NewPlaySeesionInfo(
								callID,
								stream.StreamId,
								this.localhost,
								15060,
								rtpRemoteIP,
								int(rtpRemotePort),
								ssrc,
								&subChannel,
							)
							//if play thread max, close one
							if len(this.playSessionInfoMap) >= MAX_PLAY {
								for _, psi := range this.playSessionInfoMap {
									psi.Stop()
									delete(this.playSessionInfoMap, psi.CallID)
									log.Println("play thread max, closemap:", this.playSessionInfoMap)
									break
								}
							}
							//add playSessionInfo to map
							this.playSessionInfoMap[callID] = playSessionInfo
							udpConn.WriteToUDP([]byte(this.createInviteConfirmResponse(request, ssrc).String()), serverAddr)
							return
						}
					}

				}
			}
		}

	} else {
		//sent 200 ok
		udpConn.WriteToUDP([]byte(this.createInviteConfirmResponse(request, ssrc).String()), serverAddr)
	}

}

//ACK handle func
func (this *GB28181Server) ackHandle(udpConn *net.UDPConn, serverAddr *net.UDPAddr, request *base.Request) {

	if len(request.Headers("Call-Id")) == 0 {
		return
	}
	this.gbCameraServiceServerLock.Lock()
	defer this.gbCameraServiceServerLock.Unlock()

	callIDHeader := request.Headers("Call-Id")[0]
	if value, ok := (callIDHeader).(*base.CallId); ok {
		callID := string(*value)
		if playSessionInfo, ok := this.playSessionInfoMap[callID]; ok {
			playSessionInfo.Start()
			log.Println("startmap:", this.playSessionInfoMap)
		}
	}

}

//BYE handle func
func (this *GB28181Server) byeHandle(udpConn *net.UDPConn, serverAddr *net.UDPAddr, request *base.Request) {

	if len(request.Headers("Call-Id")) == 0 {
		return
	}
	this.gbCameraServiceServerLock.Lock()
	defer this.gbCameraServiceServerLock.Unlock()

	callIDHeader := request.Headers("Call-Id")[0]
	if value, ok := (callIDHeader).(*base.CallId); ok {
		callID := string(*value)
		if playSessionInfo, ok := this.playSessionInfoMap[callID]; ok {
			playSessionInfo.Stop()
			delete(this.playSessionInfoMap, callID)
			udpConn.WriteToUDP([]byte(this.createSipResponse(request, 200, "OK", "").String()), serverAddr)
			log.Println("closemap:", this.playSessionInfoMap)
		}

	}

}
