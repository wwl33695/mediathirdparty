package service

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/deepglint/dgmf/mserver/core"
	"github.com/deepglint/dgmf/mserver/protocols/gb28181"
	"github.com/deepglint/dgmf/mserver/protocols/rtmp"
	"github.com/deepglint/dgmf/mserver/protocols/rtsp"
	"github.com/deepglint/dgmf/mserver/protocols/dmi"
)

type MediaServer interface {
	Start(port int, param interface{}) error
	Stop()
	GetPort() int
	GetParam() interface{}
}

type MediaServerPool struct {
	servers map[string]MediaServer
}

var mediaServerPool *MediaServerPool

func GetMediaServerPool() *MediaServerPool {
	if mediaServerPool == nil {
		mediaServerPool = &MediaServerPool{
			servers: make(map[string]MediaServer),
		}
	}
	return mediaServerPool
}

func (this *MediaServerPool) StartServer(serverId string, port int, param interface{}) error {
	if _, ok := this.servers[serverId]; ok == true {
		return errors.New(serverId + " service has been started at :" + fmt.Sprintf("%d", this.servers[serverId].GetPort()))
	}

	var server MediaServer
	config := core.GetConfig()

	if strings.EqualFold("rtsp", serverId) {
		server = &rtsp.RTSPServer{}
		config.RTSPServer.Enable = true
		config.RTSPServer.Port = port
		config.RTSPServer.Param = param
		log.Printf("[SERVICE] RTSPServer start, port: %d\n", config.RTSPServer.Port)
	} else if strings.EqualFold("rtmp", serverId) {
		server = &rtmp.RTMPServer{}
		config.RTMPServer.Enable = true
		config.RTMPServer.Port = port
		config.RTMPServer.Param = param
		log.Printf("[SERVICE] RTMPServer start, port: %d\n", config.RTMPServer.Port)
	} else if strings.EqualFold("gb28181", serverId) {
		server = gb28181.NewGB28181Server()
		config.GB28181Server.Enable = true
		config.GB28181Server.Port = port
		config.GB28181Server.Param = param
		log.Printf("[SERVICE] GB28181Server start, port: %d\n", config.GB28181Server.Port)
	} else if strings.EqualFold("dmi", serverId) {
		server = &dmi.DMIServer{}	
		config.DMIServer.Enable = true
		config.DMIServer.Port = port
		config.DMIServer.Param = param
		log.Printf("[SERVICE] DMIServer start, port: %d\n", config.DMIServer.Port)
	} else {
		return errors.New("MServer can not support media server: " + serverId)
	}
	err := server.Start(port, param)
	if err != nil {
		return err
	}
	this.servers[serverId] = server
	return nil
}

func (this *MediaServerPool) StopServer(serverId string) error {
	if _, ok := this.servers[serverId]; ok == false {
		return errors.New(serverId + " not found")
	}

	config := core.GetConfig()

	if strings.EqualFold("rtsp", serverId) {
		config.RTSPServer.Enable = false
		log.Printf("[SERVICE] RTSPServer stop")
	} else if strings.EqualFold("rtmp", serverId) {
		config.RTMPServer.Enable = false
		log.Printf("[SERVICE] RTMPServer stop")
	} else if strings.EqualFold("gb28181", serverId) {
		config.GB28181Server.Enable = false
		log.Printf("[SERVICE] RTMPServer stop")
	} else if strings.EqualFold("dmi", serverId) {
		config.DMIServer.Enable = false
		log.Printf("[SERVICE] DMIServer stop")
	} else {
		return errors.New("MServer can not support media server: " + serverId)
	}
	this.servers[serverId].Stop()
	delete(this.servers, serverId)
	return nil
}

func (this *MediaServerPool) GetServer(serverId string) MediaServer {
	if _, ok := this.servers[serverId]; ok == true {
		return this.servers[serverId]
	} else {
		return nil
	}
}
