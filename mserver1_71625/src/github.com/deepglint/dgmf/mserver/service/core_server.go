package service

import (
	"time"

	"github.com/deepglint/dgmf/mserver/core"
)

var checkRunning bool = false

func CheckLiveStatus() {
	if !checkRunning {
		go func() {
			pool := core.GetESPool()
			timer := time.NewTicker(5 * time.Second)
			for {
				select {
				case <-timer.C:
					for _, stream := range pool.Live.Maps {
						t0 := stream.Index
						time.Sleep(100 * time.Millisecond)
						t1 := stream.Index
						if t0 == t1 {
							stream.InputStatus = false
						} else {
							stream.InputStatus = true
						}
					}

					for _, stream := range pool.Proxy.Maps {
						t0 := stream.Receiver.Index()
						time.Sleep(100 * time.Millisecond)
						t1 := stream.Receiver.Index()
						if t0 == t1 {
							stream.InputStatus = false
						} else {
							stream.InputStatus = true
						}
					}
				}
			}
		}()
	}
}

// func AddInput(streamId string, streamType string, inputUri string) error {
// 	var err error
// 	pool := core.GetESPool()

// 	// if isLive {
// 	var inputCtx core.LiveInputLayer
// 	urlCtx, err := url.Parse(inputUri)
// 	if err != nil {
// 		return err
// 	}

// 	if strings.EqualFold(urlCtx.Scheme, "udp") {
// 		inputCtx = &input.UDPLiveInput{}
// 	} else if strings.EqualFold(urlCtx.Scheme, "rtsp") {
// 		inputCtx = &input.RTSPLiveInput{}
// 	} else if strings.EqualFold(urlCtx.Scheme, "file") && len(strings.Split(inputUri, ".")) == 2 && strings.EqualFold(strings.Split(inputUri, ".")[1], "h264") {
// 		inputCtx = &h264.FileH264LiveInputLayer{}
// 	} else {
// 		return errors.New("Unsuportted protocol: " + urlCtx.Scheme)
// 	}

// 	if len(inputUri) < len(urlCtx.Scheme)+3 {
// 		return errors.New("Invalid URI")
// 	}
// 	if !strings.EqualFold(inputUri[len(urlCtx.Scheme):len(urlCtx.Scheme)+3], "://") {
// 		return errors.New("Invalid URI")
// 	}

// 	err = pool.AddStream(streamId, streamType, inputUri)
// 	if err != nil {
// 		return err
// 	}
// 	inputCtx.Open(inputUri, pool.LiveStreams[streamId])
// 	pool.LiveStreams[streamId].InputCtx = inputCtx

// 	// } else {
// 	// 	urlCtx, err := url.Parse(inputUri)
// 	// 	if err != nil {
// 	// 		return err
// 	// 	}

// 	// 	if !strings.EqualFold(urlCtx.Scheme, "file") {
// 	// 		return errors.New("Unsuportted protocol: " + urlCtx.Scheme)
// 	// 	}
// 	// 	if len(inputUri) < 7 {
// 	// 		return errors.New("Invalid URI")
// 	// 	}
// 	// 	if !strings.EqualFold(inputUri[4:7], "://") {
// 	// 		return errors.New("Invalid URI")
// 	// 	}

// 	// 	parts := strings.Split(inputUri, ".")
// 	// 	if len(parts) != 2 {
// 	// 		return errors.New("Unsuportted file format")
// 	// 	}
// 	// 	if !strings.EqualFold(parts[1], "h264") {
// 	// 		return errors.New("Unsuportted file format: " + parts[1])
// 	// 	}

// 	// 	file, err := os.Open(inputUri[7:])
// 	// 	if err != nil {
// 	// 		return err
// 	// 	}

// 	// 	err = pool.AddInput(streamId, isLive, inputUri)
// 	// 	if err != nil {
// 	// 		return err
// 	// 	}

// 	// 	reader := bufio.NewReader(file)
// 	// 	IDR0, err := h264.ReadNextH264Nalu(reader)
// 	// 	if err != nil {
// 	// 		return err
// 	// 	}
// 	// 	stream := pool.VodStreams[streamId]
// 	// 	h264.GetVodSPS(IDR0, stream)
// 	// 	if len(stream.SPS) == 0 || len(stream.PPS) == 0 ||
// 	// 		stream.Width <= 0 || stream.Height <= 0 {
// 	// 		pool.RemoveInput(streamId, isLive)
// 	// 		return errors.New("H264 file can not get sps or pps")
// 	// 	}
// 	// 	file.Close()
// 	// }

// 	return err
// }

// func RemoveInput(streamId string, streamType string) error {
// 	var err error
// 	pool := core.GetESPool()
// 	err = pool.RemoveStream(streamId, streamType)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// func AddOutput(streamId string, streamType string, protocol string, enable bool, param interface{}) error {
// 	var err error
// 	if !strings.EqualFold(protocol, "rtsp") &&
// 		!strings.EqualFold(protocol, "rtmp") &&
// 		!strings.EqualFold(protocol, "gb28181") {
// 		return errors.New("Unsupported protocol: " + protocol)
// 	}

// 	pool := core.GetESPool()

// 	err = pool.AddOutput(streamId, streamType, protocol, enable, param)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func SetOutput(streamId string, streamType string, protocol string, enable bool, param interface{}) error {
// 	if !strings.EqualFold(protocol, "rtsp") &&
// 		!strings.EqualFold(protocol, "rtmp") &&
// 		!strings.EqualFold(protocol, "gb28181") {
// 		return errors.New("Unsupported protocol: " + protocol)
// 	}
// 	return AddOutput(streamId, streamType, protocol, enable, param)
// }

// func RemoveOutput(streamId string, streamType string, protocol string) error {
// 	var err error
// 	if !strings.EqualFold(protocol, "rtsp") &&
// 		!strings.EqualFold(protocol, "rtmp") &&
// 		!strings.EqualFold(protocol, "gb28181") {
// 		return errors.New("Unsupported protocol: " + protocol)
// 	}

// 	pool := core.GetESPool()
// 	// if isLive {
// 	for _, session := range pool.LiveStreams[streamId].Sessions {
// 		if strings.EqualFold(session.Protocol, protocol) {
// 			pool.RemoveSession(streamId, streamType, session.SessionId)
// 		}
// 	}
// 	// } else {
// 	// 	for _, session := range pool.VodStreams[streamId].Sessions {
// 	// 		if strings.EqualFold(session.Protocol, protocol) {
// 	// 			pool.RemoveSession(streamId, isLive, session.SessionId)
// 	// 		}
// 	// 	}
// 	// }
// 	err = pool.RemoveOutput(streamId, streamType, protocol)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }
