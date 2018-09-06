package core

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
)

type GB28181ServerConfig struct {
	Interval       int
	DeviceID       string
	DeviceAreaID   string
	ServerID       string
	ServerAreaID   string
	ServerHost     string
	ServerPort     uint16
	ServerPassword string
	ExpiresTime    uint32
}

type ServerConfig struct {
	Enable bool
	Port   int
	Param  interface{}
}

type InputConfig struct {
	StreamId string
	Uri      string
	Outputs  map[string]*OutputConfig
}

type OutputConfig struct {
	Protocol string
	Enable   bool
	Param    interface{}
}

type Config struct {
	RTSPServer    ServerConfig
	HTTPAPIServer ServerConfig
	RTMPServer    ServerConfig
	GB28181Server ServerConfig
	DMIServer	  ServerConfig
	LiveInputs    map[string]*InputConfig
	// VodInputs     map[string]*InputConfig
}

var config *Config
var ConfigPath string

func GetConfig() *Config {
	if config == nil {
		config = &Config{}
	}
	return config
}

func (this *Config) Load() error {
	buf, err := ioutil.ReadFile(ConfigPath)
	if err != nil {
		return err
	}
	err = json.Unmarshal(buf, this)
	if err != nil {
		return err
	}

	if this.GB28181Server.Param != nil {
		gb, _ := json.Marshal(this.GB28181Server.Param)
		var gbConfig GB28181ServerConfig
		json.Unmarshal(gb, &gbConfig)
		this.GB28181Server.Param = gbConfig
	}

	for _, stream := range this.LiveInputs {
		for _, output := range stream.Outputs {
			if strings.EqualFold(output.Protocol, "gb28181") {
				gb, _ := json.Marshal(output.Param)
				var gbChannel GB28181Channel
				json.Unmarshal(gb, &gbChannel)
				output.Param = gbChannel
			}
		}
	}

	return nil
}

func (this *Config) Update() {
	esPool := GetESPool()

	this.LiveInputs = make(map[string]*InputConfig)
	for _, stream := range esPool.Live.Maps {
		input := &InputConfig{}
		input.Outputs = make(map[string]*OutputConfig)
		input.StreamId = stream.StreamId
		input.Uri = stream.RequestURI
		for protocol, output := range stream.Outputs {
			outputConfig := &OutputConfig{}
			outputConfig.Protocol = output.Protocol
			outputConfig.Enable = output.Enable
			outputConfig.Param = output.Param
			input.Outputs[protocol] = outputConfig
		}
		this.LiveInputs[stream.StreamId] = input
	}

	// this.VodInputs = make(map[string]*InputConfig)
	// for _, stream := range esPool.VodStreams {
	// 	input := &InputConfig{}
	// 	input.Outputs = make(map[string]*OutputConfig)
	// 	input.StreamId = stream.StreamId
	// 	input.Uri = stream.URI
	// 	for protocol, output := range stream.Outputs {
	// 		outputConfig := &OutputConfig{}
	// 		outputConfig.Protocol = output.Protocol
	// 		outputConfig.Enable = output.Enable
	// 		outputConfig.Param = output.Param
	// 		input.Outputs[protocol] = outputConfig
	// 	}
	// 	this.VodInputs[stream.StreamId] = input
	// }
}

func (this *Config) Save() error {
	buf, err := json.MarshalIndent(this, "", "\t")
	if err != nil {
		return err
	}
	file, err := os.Create(ConfigPath)
	if err != nil {
		return err
	}
	_, err = file.Write(buf)
	if err != nil {
		return err
	}
	err = file.Close()
	if err != nil {
		return err
	}
	return nil
}
