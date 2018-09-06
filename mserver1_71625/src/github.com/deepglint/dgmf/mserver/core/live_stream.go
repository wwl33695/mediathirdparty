package core

import (
	"errors"
	"net"
	"net/url"
	"strings"
	"sync"
)

type LiveStream struct {
	StreamId    string
	RequestURI  string
	InputStatus bool
	IFrame      H264ESFrame
	Sessions    map[string]*LiveSession
	Outputs     map[string]*Output
	InputCtx    LiveInputLayer
	Index       uint64
	Fps         uint32
	Width       uint32
	Height      uint32
	SPS         string
	PPS         string
	FMTP        string
	Reserved	interface{}
}

type LiveSession struct {
	SessionId  string
	RemoteAddr string
	Network    string
	Protocol   string
	Frame      chan *H264ESFrame
	Connect    net.Conn
}

type LiveStreams struct {
	sync.RWMutex
	Maps map[string]*LiveStream
}

func NewLiveStreams() *LiveStreams {
	liveStreams := &LiveStreams{
		Maps: make(map[string]*LiveStream),
	}
	return liveStreams
}

func (this *LiveStreams) GetStream(streamId string) (*LiveStream, error) {
	this.RLock()
	defer this.RUnlock()

	if stream, ok := this.Maps[streamId]; ok {
		return stream, nil
	} else {
		return nil, errors.New("Can not find stream id: " + streamId)
	}
}

func (this *LiveStreams) GetSession(streamId string, sessionId string) (*LiveSession, error) {
	var err error

	this.RLock()
	defer this.RUnlock()

	if stream, ok := this.Maps[streamId]; ok {
		if session, ook := stream.Sessions[sessionId]; ook {
			return session, nil
		} else {
			err = errors.New("Can not find session id: " + sessionId)
			return nil, err
		}
	} else {
		err = errors.New("Can not find stream id: " + streamId)
		return nil, err
	}
}

func (this *LiveStreams) GetFrames(streamId string, sessionId string) (<-chan *H264ESFrame, error) {
	var err error

	this.RLock()
	defer this.RUnlock()

	if stream, ok := this.Maps[streamId]; ok {
		if session, ook := stream.Sessions[sessionId]; ook {
			frames := session.Frame
			return frames, nil
		} else {
			err = errors.New("Can not find session id: " + sessionId)
			return nil, err
		}
	} else {
		err = errors.New("Can not find stream id: " + streamId)
		return nil, err
	}
}

func (this *LiveStreams) GetFMTP(streamId string) (string, error) {
	var fmtp string
	var err error

	this.RLock()
	defer this.RUnlock()

	if stream, ok := this.Maps[streamId]; ok && stream != nil {
		fmtp = stream.FMTP
		return fmtp, nil
	} else {
		err = errors.New("Can not find stream id: " + streamId)
		return fmtp, err
	}
}

func (this *LiveStreams) GetSPS(streamId string) (string, error) {
	var sps string
	var err error

	this.RLock()
	defer this.RUnlock()

	if stream, ok := this.Maps[streamId]; ok && stream != nil {
		sps = stream.SPS
		return sps, nil
	} else {
		err = errors.New("Can not find stream id: " + streamId)
		return sps, err
	}
}

func (this *LiveStreams) GetPPS(streamId string) (string, error) {
	var pps string
	var err error

	this.RLock()
	defer this.RUnlock()

	if stream, ok := this.Maps[streamId]; ok && stream != nil {
		pps = stream.PPS
		return pps, nil
	} else {
		err = errors.New("Can not find stream id: " + streamId)
		return pps, err
	}
}

func (this *LiveStreams) ExistOutput(streamId string, protocol string) bool {
	this.RLock()
	defer this.RUnlock()

	if stream, ok := this.Maps[streamId]; ok && stream != nil {
		if output, ook := stream.Outputs[strings.ToLower(protocol)]; ook && output != nil && output.Enable {
			return true
		} else {
			return false
		}
	} else {
		return false
	}
}

func (this *LiveStreams) ExistInput(streamId string) bool {
	this.RLock()
	defer this.RUnlock()

	if stream, ok := this.Maps[streamId]; ok && stream != nil {
		return true
	} else {
		return false
	}
}

func (this *LiveStreams) RemoveSessionByOutput(protocol string) {
	this.Lock()
	defer this.Unlock()

	for _, stream := range this.Maps {
		for _, session := range stream.Sessions {
			if strings.EqualFold(session.Protocol, protocol) {
				this.RemoveSession(stream.StreamId, session.SessionId)
			}
		}
	}

	return
}

func (this *LiveStreams) AddStream(streamId string, uri string, inputCtx LiveInputLayer) error {
	var err error
	this.Lock()
	defer this.Unlock()

	if len(streamId) == 0 {
		err = errors.New("StreamId can not be empty")
		return err
	}

	if len(uri) == 0 {
		err = errors.New("URI can not be empty")
		return err
	}

	if inputCtx == nil {
		err = errors.New("InputCtx can not be empty")
		return err
	}

	urlCtx, err := url.Parse(uri)
	if err != nil {
		return err
	}
	if len(uri) < len(urlCtx.Scheme)+3 {
		err = errors.New("URI is invalid")
		return err
	}
	if !strings.EqualFold(uri[len(urlCtx.Scheme):len(urlCtx.Scheme)+3], "://") {
		err = errors.New("URI is invalid")
		return err
	}

	if _, ok := this.Maps[streamId]; ok {
		err = errors.New("LiveStream stream id: " + streamId + " already in use")
		return err
	}

	this.Maps[streamId] = &LiveStream{
		StreamId:    streamId,
		InputStatus: false,
		Sessions:    make(map[string]*LiveSession),
		Outputs:     make(map[string]*Output),
		RequestURI:  uri,
		InputCtx:    inputCtx,
	}

	inputCtx.Open(uri, this.Maps[streamId])

	return nil
}

func (this *LiveStreams) AddStream2(streamId string, uri string) error {
	var err error
	this.Lock()
	defer this.Unlock()

	if len(streamId) == 0 {
		err = errors.New("StreamId can not be empty")
		return err
	}

	if len(uri) == 0 {
		err = errors.New("URI can not be empty")
		return err
	}

	urlCtx, err := url.Parse(uri)
	if err != nil {
		return err
	}
	if len(uri) < len(urlCtx.Scheme)+3 {
		err = errors.New("URI is invalid")
		return err
	}
	if !strings.EqualFold(uri[len(urlCtx.Scheme):len(urlCtx.Scheme)+3], "://") {
		err = errors.New("URI is invalid")
		return err
	}

	if _, ok := this.Maps[streamId]; ok {
		err = errors.New("LiveStream stream id: " + streamId + " already in use")
		return err
	}

	this.Maps[streamId] = &LiveStream{
		StreamId:    streamId,
		InputStatus: false,
		Sessions:    make(map[string]*LiveSession),
		Outputs:     make(map[string]*Output),
		RequestURI:  uri,
	}

	return nil
}

func (this *LiveStreams) RemoveStream(streamId string) error {
	var err error
	this.Lock()
	defer this.Unlock()

	if len(streamId) == 0 {
		err = errors.New("StreamId can not be empty")
		return err
	}

	if stream, ok := this.Maps[streamId]; ok && stream != nil {
		if stream.InputCtx != nil {
			stream.InputCtx.Close()
		}
		delete(this.Maps, streamId)
	} else {
		err = errors.New("StreamId: " + streamId + " not found")
		return err
	}

	return nil
}

func (this *LiveStreams) AddOutput(streamId string, protocol string, enable bool, param interface{}) error {
	var err error

	this.Lock()
	defer this.Unlock()

	if stream, ok := this.Maps[streamId]; !ok || stream == nil {
		err = errors.New("Can not find live stream: " + streamId)
		return err
	} else {

		if( !strings.EqualFold(protocol, "rtsp") && 
					!strings.EqualFold(protocol, "rtmp") && 
					!strings.EqualFold(protocol, "gb28181") && 
					!strings.EqualFold(protocol, "dmi") ){
			err = errors.New("LiveStream output protocol must be rtsp, rtmp or gb28181")
			return err
		}

		if _, ok := stream.Outputs[strings.ToLower(protocol)]; ok {
			err = errors.New("LiveStream output protocol: " + strings.ToLower(protocol) + " already in use")
			return err
		}

		stream.Outputs[strings.ToLower(protocol)] = &Output{
			Protocol: protocol,
			Enable:   enable,
			Param:    param,
		}
	}

	return nil
}

func (this *LiveStreams) RemoveOutput(streamId string, protocol string) error {
	var err error

	this.Lock()
	defer this.Unlock()

	if stream, ok := this.Maps[streamId]; !ok || stream == nil {
		err = errors.New("Can not find live stream: " + streamId)
		return err
	} else {
		if output, ook := stream.Outputs[strings.ToLower(protocol)]; !ook || output == nil {
			err = errors.New("Can not find output protocol: " + protocol)
			return err
		} else {
			delete(stream.Outputs, protocol)
		}
	}

	return nil
}

func (this *LiveStreams) AddSession(streamId string, sessionId string, protocol string, connect net.Conn) error {
	var err error

	this.Lock()
	defer this.Unlock()

	if stream, ok := this.Maps[streamId]; !ok || stream == nil {
		err = errors.New("Can not find live stream: " + streamId)
		return err
	} else {
		if _, ok := stream.Sessions[sessionId]; ok {
			err = errors.New("LiveStream session: " + sessionId + " already in use")
			return err
		}

		stream.Sessions[sessionId] = &LiveSession{
			SessionId:  sessionId,
			RemoteAddr: connect.RemoteAddr().String(),
			Network:    connect.RemoteAddr().Network(),
			Frame:      make(chan *H264ESFrame, 5),
			Protocol:   strings.ToLower(protocol),
			Connect:    connect,
		}
	}

	return nil
}

func (this *LiveStreams) RemoveSession(streamId string, sessionId string) error {
	var err error

	this.Lock()
	defer this.Unlock()

	if stream, ok := this.Maps[streamId]; !ok || stream == nil {
		err = errors.New("Can not find live stream: " + streamId)
		return err
	} else {
		if session, ok := stream.Sessions[sessionId]; ok && session != nil {
			close(session.Frame)
			if session.Connect != nil {
				session.Connect.Close()
			}
			delete(stream.Sessions, sessionId)
		}
	}

	return nil
}
