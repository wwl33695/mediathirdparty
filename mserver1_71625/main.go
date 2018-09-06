package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strings"

	"github.com/deepglint/dgmf/mserver/core"
	"github.com/deepglint/dgmf/mserver/input"
	"github.com/deepglint/dgmf/mserver/service"
	"github.com/deepglint/dgmf/mserver/service/api"
	"github.com/deepglint/dgmf/mserver/utils"
)

func main() {
//	done := make(chan bool)

	configPath := flag.String("config", "config.json", "Config file path")
	version := flag.Bool("version", false, "Print MServer version")
	profile := flag.Bool("profile", false, "Open cpu profile mode")

	flag.Parse()
	if *version {
		fmt.Println(utils.MANUFACTURER)
		os.Exit(0)
	}

	if *profile {
		log.Println("[MAIN] Run: go tool pprof http://127.0.0.1:6060/debug/pprof/profile")
		go func() {
			http.ListenAndServe("localhost:6060", nil)
		}()
	}

	core.ConfigPath = *configPath

	config := core.GetConfig()
	if err := config.Load(); err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			log.Println(err.Error(), ", MServer create a new one.")
			config.RTSPServer = core.ServerConfig{
				Enable: true,
				Port:   554,
				Param:  nil,
			}
			config.HTTPAPIServer = core.ServerConfig{
				Enable: true,
				Port:   8080,
				Param:  nil,
			}
			config.RTMPServer = core.ServerConfig{
				Enable: true,
				Port:   1935,
				Param:  nil,
			}
			config.GB28181Server = core.ServerConfig{
				Enable: true,
				Port:   5060,
				Param: core.GB28181ServerConfig{
					Interval:       3,
					DeviceID:       "34020000001180000001",
					DeviceAreaID:   "3402000000",
					ServerID:       "34010000002000000001",
					ServerAreaID:   "3401000000",
					ServerHost:     "192.168.5.154",
					ServerPort:     5060,
					ServerPassword: "12345678",
				},
			}
			config.Save()
		} else {
			fmt.Println(err)
			os.Exit(0)
		}
	}

	pool := core.GetESPool()
	for _, inputCfg := range config.LiveInputs {
		inputCtx, err := input.GetLiveInputCtx(inputCfg.Uri)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
		pool.Live.AddStream(inputCfg.StreamId, inputCfg.Uri, inputCtx)
		for _, output := range inputCfg.Outputs {
			pool.Live.AddOutput(inputCfg.StreamId, output.Protocol, output.Enable, output.Param)
		}
	}

	// for _, input := range config.VodInputs {
	// 	service.AddInput(input.StreamId, false, input.Uri)
	// 	for _, output := range input.Outputs {
	// 		service.AddOutput(input.StreamId, false, output.Protocol, output.Enable, output.Param)
	// 	}
	// }

	service.CheckLiveStatus()

	servers := service.GetMediaServerPool()

	if config.RTSPServer.Enable {
		servers.StartServer("rtsp", config.RTSPServer.Port, config.RTMPServer.Param)
	}

	if config.RTMPServer.Enable {
		servers.StartServer("rtmp", config.RTMPServer.Port, config.RTMPServer.Param)
	}

	if config.GB28181Server.Enable {
		servers.StartServer("gb28181", config.GB28181Server.Port, config.GB28181Server.Param)
	}

	if config.DMIServer.Enable {
		servers.StartServer("dmi", config.DMIServer.Port, config.DMIServer.Param)
	}

	if config.HTTPAPIServer.Enable {
		httpServer := api.HttpAPIServer{}
		httpServer.Start(config.HTTPAPIServer.Port)
		log.Printf("[MAIN] HTTPAPIServer start, port: %d\n", config.HTTPAPIServer.Port)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	s := <-c
	fmt.Println("Got signal:", s)
	
	for _, inputCfg := range config.LiveInputs {
		pool.Live.RemoveStream(inputCfg.StreamId)
	}
}
