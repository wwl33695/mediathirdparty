package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/deepglint/dgmf/mserver/core"
	"github.com/deepglint/dgmf/mserver/input"
	"github.com/deepglint/dgmf/mserver/service"
	"github.com/deepglint/dgmf/mserver/utils"
	"github.com/golang/glog"
)

type HttpAPIServer struct {
}

type SystemError struct {
	StatusCode int
	Content    string
}

type LiveSessionStatus struct {
	SessionId  string
	RemoteAddr string
	Network    string
	Protocol   string
}

// type VodSessionStatus struct {
// 	SessionId  string
// 	RemoteAddr string
// 	Network    string
// 	Protocol   string
// 	Index      int
// }

type OutputStatus struct {
	Protocol string
	Enable   bool
	Param    interface{}
}

type LiveStreamStatus struct {
	StreamId       string
	InputStatus    bool
	RequestURI     string
	FPS            uint32
	Index          uint64
	Width          uint32
	Height         uint32
	SPS            string
	PPS            string
	SessionsStatus []LiveSessionStatus
	Outputs        map[string]OutputStatus
}

type ProxyStreamStatus struct {
	StreamId    string
	ProxyURI    string
	RequestURI  string
	InputStatus bool
	RemoteAddr  string
	Network     string
	Protocol    string
	Index       uint64
	FPS         uint32
	Width       uint32
	Height      uint32
	SPS         string
	PPS         string
}

// type MediaServerStatus struct {
// 	Protocol string
// 	Enable   bool
// 	Port     int
// 	Param    interface{}
// }

type SystemStatus struct {
	StreamCount  int
	SessionCount int
	Lives        []LiveStreamStatus
	Proxies      []ProxyStreamStatus
	// VodStreams   []VodStreamStatus
	// MediaServers []MediaServerStatus
}

func addInputHandler(w http.ResponseWriter, r *http.Request) {
	glog.V(2).Infof("[API_SERVER] Request:\n%s\n\n\n", r.URL.RequestURI())
	w.Header().Set("content-type", "json")
	var systemError SystemError
	var err error
	streamId := r.URL.Query().Get("stream_id")
	uri := r.URL.Query().Get("uri")
	streamType := r.URL.Query().Get("stream_type")
	if len(streamId) == 0 {
		systemError = SystemError{
			StatusCode: 400,
			Content:    "stream_id not found",
		}
		data, _ := json.Marshal(systemError)
		w.Write(data)
		glog.V(2).Infof("[API_SERVER] Response:\n%s\n\n\n", string(data))
		return
	}
	if len(uri) == 0 {
		systemError = SystemError{
			StatusCode: 400,
			Content:    "uri not found",
		}
		data, _ := json.Marshal(systemError)
		w.Write(data)
		glog.V(2).Infof("[API_SERVER] Response:\n%s\n\n\n", string(data))
		return
	}
	if len(streamType) == 0 {
		systemError = SystemError{
			StatusCode: 400,
			Content:    "stream_type not found",
		}
		data, _ := json.Marshal(systemError)
		w.Write(data)
		glog.V(2).Infof("[API_SERVER] Response:\n%s\n\n\n", string(data))
		return
	}
	if !strings.EqualFold(streamType, "live") && !strings.EqualFold(streamType, "vod") {
		systemError = SystemError{
			StatusCode: 400,
			Content:    "stream_type must be live or vod",
		}
		data, _ := json.Marshal(systemError)
		w.Write(data)
		glog.V(2).Infof("[API_SERVER] Response:\n%s\n\n\n", string(data))
		return
	}

	if strings.EqualFold(streamType, "live") {
		pool := core.GetESPool()

		if pool.Live.ExistInput(streamId) {
			systemError = SystemError{
				StatusCode: 400,
				Content:    "stream_id: " + streamId + " exist",
			}
			data, _ := json.Marshal(systemError)
			w.Write(data)
			glog.V(2).Infof("[API_SERVER] Response:\n%s\n\n\n", string(data))
			return
		}

		inputCtx, err := input.GetLiveInputCtx(uri)
		if err != nil {
			systemError = SystemError{
				StatusCode: 400,
				Content:    err.Error(),
			}
			data, _ := json.Marshal(systemError)
			w.Write(data)
			glog.V(2).Infof("[API_SERVER] Response:\n%s\n\n\n", string(data))
			return
		}

		err = pool.Live.AddStream(streamId, uri, inputCtx)
		if err != nil {
			systemError = SystemError{
				StatusCode: 400,
				Content:    err.Error(),
			}
			data, _ := json.Marshal(systemError)
			w.Write(data)
			glog.V(2).Infof("[API_SERVER] Response:\n%s\n\n\n", string(data))
			return
		}
	}

	config := core.GetConfig()
	config.Update()
	config.Save()
	if err == nil {
		systemError = SystemError{
			StatusCode: 200,
			Content:    "OK",
		}
	} else {
		systemError = SystemError{
			StatusCode: 400,
			Content:    err.Error(),
		}
	}
	data, _ := json.Marshal(systemError)
	w.Write(data)
	glog.V(2).Infof("[API_SERVER] Response:\n%s\n\n\n", string(data))
}

func removeInputHandler(w http.ResponseWriter, r *http.Request) {
	glog.V(2).Infof("[API_SERVER] Request:\n%s\n\n\n", r.URL.RequestURI())
	w.Header().Set("content-type", "json")
	var systemError SystemError
	var err error
	streamId := r.URL.Query().Get("stream_id")
	streamType := r.URL.Query().Get("stream_type")
	if len(streamId) == 0 {
		systemError = SystemError{
			StatusCode: 400,
			Content:    "stream_id not found",
		}
		data, _ := json.Marshal(systemError)
		w.Write(data)
		glog.V(2).Infof("[API_SERVER] Response:\n%s\n\n\n", string(data))
		return
	}
	if len(streamType) == 0 {
		systemError = SystemError{
			StatusCode: 400,
			Content:    "stream_type not found",
		}
		data, _ := json.Marshal(systemError)
		w.Write(data)
		glog.V(2).Infof("[API_SERVER] Response:\n%s\n\n\n", string(data))
		return
	}
	if !strings.EqualFold(streamType, "live") && !strings.EqualFold(streamType, "vod") {
		systemError = SystemError{
			StatusCode: 400,
			Content:    "stream_type must be live or vod",
		}
		data, _ := json.Marshal(systemError)
		w.Write(data)
		glog.V(2).Infof("[API_SERVER] Response:\n%s\n\n\n", string(data))
		return
	}

	if strings.EqualFold(streamType, "live") {
		pool := core.GetESPool()

		if !pool.Live.ExistInput(streamId) {
			systemError = SystemError{
				StatusCode: 400,
				Content:    "stream_id: " + streamId + " not exist",
			}
			data, _ := json.Marshal(systemError)
			w.Write(data)
			glog.V(2).Infof("[API_SERVER] Response:\n%s\n\n\n", string(data))
			return
		}

		err = pool.Live.RemoveStream(streamId)
		if err != nil {
			systemError = SystemError{
				StatusCode: 400,
				Content:    err.Error(),
			}
			data, _ := json.Marshal(systemError)
			w.Write(data)
			glog.V(2).Infof("[API_SERVER] Response:\n%s\n\n\n", string(data))
			return
		}
	}

	config := core.GetConfig()
	config.Update()
	config.Save()
	if err == nil {
		systemError = SystemError{
			StatusCode: 200,
			Content:    "OK",
		}
	} else {
		systemError = SystemError{
			StatusCode: 400,
			Content:    err.Error(),
		}
	}
	data, _ := json.Marshal(systemError)
	w.Write(data)
	glog.V(2).Infof("[API_SERVER] Response:\n%s\n\n\n", string(data))
}

func addOutputHandler(w http.ResponseWriter, r *http.Request) {
	glog.V(2).Infof("[API_SERVER] Request:\n%s\n\n\n", r.URL.RequestURI())
	w.Header().Set("content-type", "json")
	var systemError SystemError
	var err error
	streamId := r.URL.Query().Get("stream_id")
	protocol := r.URL.Query().Get("protocol")
	streamType := r.URL.Query().Get("stream_type")

	if len(streamId) == 0 {
		systemError = SystemError{
			StatusCode: 400,
			Content:    "streamid not found",
		}
		data, _ := json.Marshal(systemError)
		w.Write(data)
		glog.V(2).Infof("[API_SERVER] Response:\n%s\n\n\n", string(data))
		return
	}
	if len(protocol) == 0 {
		systemError = SystemError{
			StatusCode: 400,
			Content:    "protocol not found",
		}
		data, _ := json.Marshal(systemError)
		w.Write(data)
		glog.V(2).Infof("[API_SERVER] Response:\n%s\n\n\n", string(data))
		return
	}
	if len(streamType) == 0 {
		systemError = SystemError{
			StatusCode: 400,
			Content:    "stream_type not found",
		}
		data, _ := json.Marshal(systemError)
		w.Write(data)
		glog.V(2).Infof("[API_SERVER] Response:\n%s\n\n\n", string(data))
		return
	}
	if !strings.EqualFold(streamType, "live") && !strings.EqualFold(streamType, "vod") {
		systemError = SystemError{
			StatusCode: 400,
			Content:    "stream_type must be live or vod",
		}
		data, _ := json.Marshal(systemError)
		w.Write(data)
		glog.V(2).Infof("[API_SERVER] Response:\n%s\n\n\n", string(data))
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		systemError = SystemError{
			StatusCode: 400,
			Content:    "body invalid",
		}
		data, _ := json.Marshal(systemError)
		w.Write(data)
		glog.V(2).Infof("[API_SERVER] Response:\n%s\n\n\n", string(data))
		return
	}

	var param interface{}

	if len(body) > 0 && strings.EqualFold(protocol, "gb28181") {
		var gbChannel core.GB28181Channel
		err = json.Unmarshal(body, &gbChannel)
		if err != nil {
			systemError = SystemError{
				StatusCode: 400,
				Content:    "body json unmarshal error",
			}
			data, _ := json.Marshal(systemError)
			w.Write(data)
			glog.V(2).Infof("[API_SERVER] Response:\n%s\n\n\n", string(data))
			return
		}
		param = gbChannel
	}

	if strings.EqualFold(streamType, "live") {
		pool := core.GetESPool()

		if !pool.Live.ExistInput(streamId) {
			systemError = SystemError{
				StatusCode: 400,
				Content:    "stream_id: " + streamId + " not exist",
			}
			data, _ := json.Marshal(systemError)
			w.Write(data)
			glog.V(2).Infof("[API_SERVER] Response:\n%s\n\n\n", string(data))
			return
		}

		if pool.Live.ExistOutput(streamId, protocol) {
			systemError = SystemError{
				StatusCode: 400,
				Content:    "protocol: " + protocol + " exist",
			}
			data, _ := json.Marshal(systemError)
			w.Write(data)
			glog.V(2).Infof("[API_SERVER] Response:\n%s\n\n\n", string(data))
			return
		}

		err = pool.Live.AddOutput(streamId, protocol, true, param)
		if err != nil {
			systemError = SystemError{
				StatusCode: 400,
				Content:    err.Error(),
			}
			data, _ := json.Marshal(systemError)
			w.Write(data)
			glog.V(2).Infof("[API_SERVER] Response:\n%s\n\n\n", string(data))
			return
		}
	}

	config := core.GetConfig()
	config.Update()
	config.Save()
	if err == nil {
		systemError = SystemError{
			StatusCode: 200,
			Content:    "OK",
		}
	} else {
		systemError = SystemError{
			StatusCode: 400,
			Content:    err.Error(),
		}
	}
	data, _ := json.Marshal(systemError)
	w.Write(data)
	glog.V(2).Infof("[API_SERVER] Response:\n%s\n\n\n", string(data))
}

func removeOutputHandler(w http.ResponseWriter, r *http.Request) {
	glog.V(2).Infof("[API_SERVER] Request:\n%s\n\n\n", r.URL.RequestURI())
	w.Header().Set("content-type", "json")
	var systemError SystemError
	var err error
	streamId := r.URL.Query().Get("stream_id")
	protocol := r.URL.Query().Get("protocol")
	streamType := r.URL.Query().Get("stream_type")

	if len(streamId) == 0 {
		systemError = SystemError{
			StatusCode: 400,
			Content:    "streamid not found",
		}
		data, _ := json.Marshal(systemError)
		w.Write(data)
		glog.V(2).Infof("[API_SERVER] Response:\n%s\n\n\n", string(data))
		return
	}
	if len(protocol) == 0 {
		systemError = SystemError{
			StatusCode: 400,
			Content:    "protocol not found",
		}
		data, _ := json.Marshal(systemError)
		w.Write(data)
		glog.V(2).Infof("[API_SERVER] Response:\n%s\n\n\n", string(data))
		return
	}
	if len(streamType) == 0 {
		systemError = SystemError{
			StatusCode: 400,
			Content:    "stream_type not found",
		}
		data, _ := json.Marshal(systemError)
		w.Write(data)
		glog.V(2).Infof("[API_SERVER] Response:\n%s\n\n\n", string(data))
		return
	}
	if !strings.EqualFold(streamType, "live") && !strings.EqualFold(streamType, "vod") {
		systemError = SystemError{
			StatusCode: 400,
			Content:    "stream_type must be live or vod",
		}
		data, _ := json.Marshal(systemError)
		w.Write(data)
		glog.V(2).Infof("[API_SERVER] Response:\n%s\n\n\n", string(data))
		return
	}

	if strings.EqualFold(streamType, "live") {
		pool := core.GetESPool()

		if !pool.Live.ExistInput(streamId) {
			systemError = SystemError{
				StatusCode: 400,
				Content:    "stream_id: " + streamId + " not exist",
			}
			data, _ := json.Marshal(systemError)
			w.Write(data)
			glog.V(2).Infof("[API_SERVER] Response:\n%s\n\n\n", string(data))
			return
		}

		if !pool.Live.ExistOutput(streamId, protocol) {
			systemError = SystemError{
				StatusCode: 400,
				Content:    "protocol: " + protocol + " not exist",
			}
			data, _ := json.Marshal(systemError)
			w.Write(data)
			glog.V(2).Infof("[API_SERVER] Response:\n%s\n\n\n", string(data))
			return
		}

		err = pool.Live.RemoveOutput(streamId, protocol)
		if err != nil {
			systemError = SystemError{
				StatusCode: 400,
				Content:    err.Error(),
			}
			data, _ := json.Marshal(systemError)
			w.Write(data)
			glog.V(2).Infof("[API_SERVER] Response:\n%s\n\n\n", string(data))
			return
		}
	}

	config := core.GetConfig()
	config.Update()
	config.Save()
	if err == nil {
		systemError = SystemError{
			StatusCode: 200,
			Content:    "OK",
		}
	} else {
		systemError = SystemError{
			StatusCode: 400,
			Content:    err.Error(),
		}
	}
	data, _ := json.Marshal(systemError)
	w.Write(data)
	glog.V(2).Infof("[API_SERVER] Response:\n%s\n\n\n", string(data))
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	glog.V(2).Infof("[API_SERVER] Request:\n%s\n\n\n", r.URL.RequestURI())
	w.Header().Set("content-type", "json")
	pool := core.GetESPool()
	var systemStatus SystemStatus

	for _, stream := range pool.Live.Maps {
		streamStatus := LiveStreamStatus{
			StreamId:    stream.StreamId,
			InputStatus: stream.InputStatus,
			RequestURI:  stream.RequestURI,
			FPS:         stream.Fps,
			Index:       stream.Index,
			Width:       stream.Width,
			Height:      stream.Height,
			SPS:         stream.SPS,
			PPS:         stream.PPS,
			Outputs:     make(map[string]OutputStatus),
		}

		for _, output := range stream.Outputs {
			outputStatus := OutputStatus{
				Protocol: output.Protocol,
				Enable:   output.Enable,
				Param:    output.Param,
			}
			streamStatus.Outputs[output.Protocol] = outputStatus
		}

		var sessionsStatus []LiveSessionStatus
		for _, session := range stream.Sessions {
			sessionStatus := LiveSessionStatus{
				SessionId:  session.SessionId,
				RemoteAddr: session.RemoteAddr,
				Network:    session.Network,
				Protocol:   session.Protocol,
			}
			sessionsStatus = append(sessionsStatus, sessionStatus)
			systemStatus.SessionCount++
		}

		streamStatus.SessionsStatus = sessionsStatus
		systemStatus.Lives = append(systemStatus.Lives, streamStatus)
		systemStatus.StreamCount++
	}

	for _, stream := range pool.Proxy.Maps {
		streamStatus := ProxyStreamStatus{
			StreamId:    stream.StreamId,
			InputStatus: stream.InputStatus,
			ProxyURI:    stream.ProxyURI,
			RequestURI:  stream.RequestURI,
			RemoteAddr:  stream.RemoteAddr,
			Network:     stream.Network,
			Protocol:    stream.Protocol,
			Index:       stream.Receiver.Index(),
			FPS:         stream.Receiver.FPS(),
			SPS:         stream.Receiver.SPS(),
			PPS:         stream.Receiver.PPS(),
			Width:       stream.Receiver.Width(),
			Height:      stream.Receiver.Height(),
		}
		systemStatus.Proxies = append(systemStatus.Proxies, streamStatus)
		systemStatus.StreamCount++
		systemStatus.SessionCount++
	}

	// 	// for _, stream := range pool.VodStreams {
	// 	// 	streamStatus := VodStreamStatus{
	// 	// 		StreamId:  stream.StreamId,
	// 	// 		URI:       stream.URI,
	// 	// 		Filename:  stream.Filename,
	// 	// 		MuxFormat: stream.MuxFormat,
	// 	// 		Fps:       stream.Fps,
	// 	// 		Width:     stream.Width,
	// 	// 		Height:    stream.Height,
	// 	// 		SPS:       stream.SPS,
	// 	// 		PPS:       stream.PPS,
	// 	// 		Outputs:   make(map[string]OutputStatus),
	// 	// 	}

	// 	// 	for _, output := range stream.Outputs {
	// 	// 		outputStatus := OutputStatus{
	// 	// 			Protocol: output.Protocol,
	// 	// 			Enable:   output.Enable,
	// 	// 			Param:    output.Param,
	// 	// 		}
	// 	// 		streamStatus.Outputs[output.Protocol] = outputStatus
	// 	// 	}

	// 	// 	var sessionsStatus []VodSessionStatus
	// 	// 	for _, session := range stream.Sessions {
	// 	// 		sessionStatus := VodSessionStatus{
	// 	// 			SessionId:  session.SessionId,
	// 	// 			RemoteAddr: session.RemoteAddr,
	// 	// 			Network:    session.Network,
	// 	// 			Protocol:   session.Protocol,
	// 	// 			Index:      session.Index,
	// 	// 		}
	// 	// 		sessionsStatus = append(sessionsStatus, sessionStatus)
	// 	// 		systemStatus.SessionCount++
	// 	// 	}

	// 	// 	streamStatus.SessionsStatus = sessionsStatus
	// 	// 	systemStatus.VodStreams = append(systemStatus.VodStreams, streamStatus)
	// 	// 	systemStatus.StreamCount++
	// 	// }

	// 	var server service.MediaServer
	// 	var serverStatus MediaServerStatus
	// 	servers := service.GetMediaServerPool()

	// 	server = servers.GetServer("rtsp")
	// 	if server != nil {
	// 		serverStatus = MediaServerStatus{
	// 			Protocol: "rtsp",
	// 			Enable:   true,
	// 			Port:     server.GetPort(),
	// 			Param:    server.GetParam(),
	// 		}
	// 	} else {
	// 		serverStatus = MediaServerStatus{
	// 			Protocol: "rtsp",
	// 			Enable:   false,
	// 			Port:     0,
	// 			Param:    nil,
	// 		}
	// 	}
	// 	systemStatus.MediaServers = append(systemStatus.MediaServers, serverStatus)

	// 	server = servers.GetServer("rtmp")
	// 	if server != nil {
	// 		serverStatus = MediaServerStatus{
	// 			Protocol: "rtmp",
	// 			Enable:   true,
	// 			Port:     server.GetPort(),
	// 			Param:    server.GetParam(),
	// 		}
	// 	} else {
	// 		serverStatus = MediaServerStatus{
	// 			Protocol: "rtmp",
	// 			Enable:   false,
	// 			Port:     0,
	// 			Param:    nil,
	// 		}
	// 	}
	// 	systemStatus.MediaServers = append(systemStatus.MediaServers, serverStatus)

	// 	server = servers.GetServer("gb28181")
	// 	if server != nil {
	// 		serverStatus = MediaServerStatus{
	// 			Protocol: "gb28181",
	// 			Enable:   true,
	// 			Port:     server.GetPort(),
	// 			Param:    server.GetParam(),
	// 		}
	// 	} else {
	// 		serverStatus = MediaServerStatus{
	// 			Protocol: "gb28181",
	// 			Enable:   false,
	// 			Port:     0,
	// 			Param:    nil,
	// 		}
	// 	}
	// 	systemStatus.MediaServers = append(systemStatus.MediaServers, serverStatus)

	data, _ := json.MarshalIndent(systemStatus, "", "\t")
	w.Write(data)
	glog.V(2).Infof("[API_SERVER] Response:\n%s\n\n\n", string(data))
}

func startServerHandler(w http.ResponseWriter, r *http.Request) {
	glog.V(2).Infof("[API_SERVER] Request:\n%s\n\n\n", r.URL.RequestURI())
	w.Header().Set("content-type", "json")
	var systemError SystemError
	var err error
	serverId := r.URL.Query().Get("serverid")
	port, err := strconv.Atoi(r.URL.Query().Get("port"))

	if len(serverId) == 0 {
		systemError = SystemError{
			StatusCode: 400,
			Content:    "serverId not found",
		}
		data, _ := json.Marshal(systemError)
		w.Write(data)
		glog.V(2).Infof("[API_SERVER] Response:\n%s\n\n\n", string(data))
		return
	}
	if err != nil {
		systemError = SystemError{
			StatusCode: 400,
			Content:    "port invalid",
		}
		data, _ := json.Marshal(systemError)
		w.Write(data)
		glog.V(2).Infof("[API_SERVER] Response:\n%s\n\n\n", string(data))
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		systemError = SystemError{
			StatusCode: 400,
			Content:    "body invalid",
		}
		data, _ := json.Marshal(systemError)
		w.Write(data)
		glog.V(2).Infof("[API_SERVER] Response:\n%s\n\n\n", string(data))
		return
	}

	var param interface{}
	if len(body) > 0 && strings.EqualFold(serverId, "gb28181") {
		var gbConfig core.GB28181ServerConfig
		err = json.Unmarshal(body, &gbConfig)
		if err != nil {
			systemError = SystemError{
				StatusCode: 400,
				Content:    "body json unmarshal error",
			}
			data, _ := json.Marshal(systemError)
			w.Write(data)
			glog.V(2).Infof("[API_SERVER] Response:\n%s\n\n\n", string(data))
			return
		}
		//check if invalid
		param = gbConfig
	}

	servers := service.GetMediaServerPool()
	err = servers.StartServer(serverId, port, param)

	config := core.GetConfig()
	config.Update()
	config.Save()
	if err == nil {
		systemError = SystemError{
			StatusCode: 200,
			Content:    "OK",
		}
	} else {
		systemError = SystemError{
			StatusCode: 400,
			Content:    err.Error(),
		}
	}
	data, _ := json.Marshal(systemError)
	w.Write(data)
	glog.V(2).Infof("[API_SERVER] Response:\n%s\n\n\n", string(data))
}

func stopServerHandler(w http.ResponseWriter, r *http.Request) {
	glog.V(2).Infof("[API_SERVER] Request:\n%s\n\n\n", r.URL.RequestURI())
	w.Header().Set("content-type", "json")
	var systemError SystemError
	var err error
	serverId := r.URL.Query().Get("serverid")

	if len(serverId) == 0 {
		systemError = SystemError{
			StatusCode: 400,
			Content:    "serverId not found",
		}
		data, _ := json.Marshal(systemError)
		w.Write(data)
		glog.V(2).Infof("[API_SERVER] Response:\n%s\n\n\n", string(data))
		return
	}

	servers := service.GetMediaServerPool()
	err = servers.StopServer(serverId)

	config := core.GetConfig()
	config.Update()
	config.Save()
	if err == nil {
		systemError = SystemError{
			StatusCode: 200,
			Content:    "OK",
		}
	} else {
		systemError = SystemError{
			StatusCode: 400,
			Content:    err.Error(),
		}
	}
	data, _ := json.Marshal(systemError)
	w.Write(data)
	glog.V(2).Infof("[API_SERVER] Response:\n%s\n\n\n", string(data))
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	glog.V(2).Infof("[API_SERVER] Request:\n%s\n\n\n", r.URL.RequestURI())
	w.Header().Set("content-type", "text")
	w.Write([]byte("OK"))
}

func versionHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(utils.MANUFACTURER))
}

func configHandler(w http.ResponseWriter, r *http.Request) {
	glog.V(2).Infof("[API_SERVER] Request:\n%s\n\n\n", r.URL.RequestURI())
	w.Header().Set("content-type", "json")
	body, err := ioutil.ReadFile(core.ConfigPath)
	if err != nil {
		systemError := SystemError{
			StatusCode: 500,
			Content:    core.ConfigPath + " not found",
		}
		data, _ := json.Marshal(systemError)
		w.Write(data)
		glog.V(2).Infof("[API_SERVER] Response:\n%s\n\n\n", string(data))
		return
	}
	w.Write(body)
}

func (this *HttpAPIServer) Start(port int) {
	go func(this *HttpAPIServer, port int) {
		http.HandleFunc("/ping", pingHandler)
		http.HandleFunc("/version", versionHandler)
		http.HandleFunc("/config", configHandler)
		http.HandleFunc("/status", statusHandler)
		http.HandleFunc("/add-input", addInputHandler)
		http.HandleFunc("/remove-input", removeInputHandler)
		http.HandleFunc("/add-output", addOutputHandler)
		http.HandleFunc("/remove-output", removeOutputHandler)
		http.HandleFunc("/start-server", startServerHandler)
		http.HandleFunc("/stop-server", stopServerHandler)
		err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
		if err != nil {
			glog.Errorf("[API_SERVER] Start API_SERVER failed, error: %s\n", err)
			os.Exit(0)
		}
	}(this, port)
}
