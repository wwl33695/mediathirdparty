package gb28181

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/deepglint/dgmf/mserver/core"
)

const (
	TIME_OUT = 5
	MAX_PLAY = 4
)

type GB28181Server struct {
	localhost                  string
	port                       uint16
	param                      core.GB28181ServerConfig
	isRegisterServiceStart     bool
	isListenSignalServiceStart bool
	isLive                     bool
	registerStopChan           chan bool
	listenStopChan             chan bool
	registerConn               *net.UDPConn
	keepAliveSN                uint64
	keepAliveConnectCount      int
	playSessionInfoMap         map[string]*PlaySessionInfo
	gbCameraServiceServerLock  *sync.Mutex
}

func NewGB28181Server() *GB28181Server {
	var lock sync.Mutex
	return &GB28181Server{
		gbCameraServiceServerLock: &lock,
		registerStopChan:          make(chan bool),
		listenStopChan:            make(chan bool),
		playSessionInfoMap:        make(map[string]*PlaySessionInfo),
	}
}

func (this *GB28181Server) Start(port int, param interface{}) error {

	this.gbCameraServiceServerLock.Lock()
	if this.isRegisterServiceStart {
		this.gbCameraServiceServerLock.Unlock()
		return errors.New("this register thread has start please close it first")
	}

	if this.isListenSignalServiceStart {
		this.gbCameraServiceServerLock.Unlock()
		return errors.New("this ListenToServer thread has start please close it first")
	}

	this.isListenSignalServiceStart = true
	this.isRegisterServiceStart = true
	this.gbCameraServiceServerLock.Unlock()

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println(err)
		return err
	}
	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				this.localhost = ipnet.IP.String()
			}
		}
	}

	this.port = uint16(port)
	this.param = param.(core.GB28181ServerConfig)

	LocalAddr, err := net.ResolveUDPAddr("udp", this.localhost)
	RemoteEP := net.UDPAddr{IP: net.ParseIP(this.param.ServerHost), Port: (int)(this.param.ServerPort)}

	this.registerConn, err = net.DialUDP("udp", LocalAddr, &RemoteEP)
	if err != nil {
		return err
	}

	go this.register()
	go this.listenToServer()

	return nil

}

func (this *GB28181Server) Stop() {

	this.gbCameraServiceServerLock.Lock()
	defer this.gbCameraServiceServerLock.Unlock()
	this.closePlaySessionInfo()
	this.registerStopChan <- true
	this.listenStopChan <- true
	for {
		if !this.isRegisterServiceStart && !this.isListenSignalServiceStart {
			this.registerConn.Close()
			return
		}
		time.Sleep(time.Second)
	}

}

func (this *GB28181Server) GetPort() int {
	return int(this.port)
}

func (this *GB28181Server) GetParam() interface{} {
	return this.param
}

//Register start the Register thread
func (this *GB28181Server) register() {

	defer log.Println("register thread close")

	defer func() {
		this.isRegisterServiceStart = false
		this.isLive = false
		secondRegisterRequest, err := this.sendFirstRegisterRequest(0, 3)
		if err == nil {
			err = this.sendSecondRegisterRequest(secondRegisterRequest)
		}
	}()

	for {
		// camera is not live
		if !this.isLive {
			log.Println("start register")
			secondRegisterRequest, err := this.sendFirstRegisterRequest(this.param.ExpiresTime, 1)
			if err != nil {
				log.Println("register err:", err)
			} else {
				err = this.sendSecondRegisterRequest(secondRegisterRequest)
				if err != nil {
					log.Println("register err:", err)
				} else {
					log.Println("register success")
					this.isLive = true
				}
			}
		} else {
			err := this.sendKeepAliveRequest()
			if err != nil {
				log.Println("keep connect err:", err)
				this.isLive = false
				this.gbCameraServiceServerLock.Lock()
				this.closePlaySessionInfo()
				this.gbCameraServiceServerLock.Unlock()
			} else {
				this.isLive = true
			}
		}
		select {
		case <-this.registerStopChan:
			return
		case <-time.After(time.Duration(this.param.Interval) * time.Second):
			continue

		}
	}
}

//ListenToServer start the ListenToServer thread
func (this *GB28181Server) listenToServer() {

	defer func() {
		this.isListenSignalServiceStart = false
		log.Println("listen to server thread close")
	}()

	addr := net.UDPAddr{
		Port: (int)(this.port),
		IP:   net.ParseIP(this.localhost),
	}

	listenToServerConn, err := net.ListenUDP("udp", &addr)
	for err != nil {
		listenToServerConn, err = net.ListenUDP("udp", &addr)
		log.Printf("can't create the listen to server connect %s\n")
		select {
		case <-this.listenStopChan:
			return
		case <-time.After(time.Second):
			continue
		}
	}

	defer listenToServerConn.Close()

	isNeedStop := false
	go func() {
		for {
			if isNeedStop {
				break
			}
			data := make([]byte, 2048)
			_, serveraddr, err := listenToServerConn.ReadFromUDP(data)
			if err != nil {
				log.Printf("listen to server error  %v", err)
				continue
			}

			// fmt.Println("--------------------", string(data))
			//root thread
			go this.requestRoot(listenToServerConn, serveraddr, data)
		}
	}()

	select {
	case <-this.listenStopChan:
		isNeedStop = true
		return
	}

}

//close all playSessionInfo
func (this *GB28181Server) closePlaySessionInfo() {
	for _, playSessionInfo := range this.playSessionInfoMap {
		playSessionInfo.Stop()
	}

}
