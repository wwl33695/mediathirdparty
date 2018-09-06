package core

import (
	"errors"
	"net"
	"net/url"
	"strings"
	"sync"
)

type ProxyStream struct {
	StreamId    string
	RequestURI  string
	InputStatus bool
	ProxyURI    string
	RemoteAddr  string
	Network     string
	Protocol    string
	Connect     net.Conn
	Receiver    IReceiver
}

type ProxyStreams struct {
	sync.RWMutex
	Maps map[string]*ProxyStream
}

func NewProxyStreams() *ProxyStreams {
	proxyStreams := &ProxyStreams{
		Maps: make(map[string]*ProxyStream),
	}
	return proxyStreams
}

func (this *ProxyStreams) ExistStream(streamId string) bool {
	this.RLock()
	defer this.RUnlock()

	_, ok := this.Maps[streamId]
	return ok
}

func (this *ProxyStreams) GetStream(streamId string) (*ProxyStream, error) {
	this.RLock()
	defer this.RUnlock()

	if stream, ok := this.Maps[streamId]; ok {
		return stream, nil
	} else {
		return nil, errors.New("Can not find stream id: " + streamId)
	}
}
func (this *ProxyStreams) GetFrames(streamId string) (<-chan *H264ESFrame, error) {
	this.RLock()
	defer this.RUnlock()

	if stream, ok := this.Maps[streamId]; ok && stream != nil && stream.Receiver != nil {
		return stream.Receiver.Frames(), nil
	} else {
		return nil, errors.New("Can not find stream id: " + streamId)
	}
}

func (this *ProxyStreams) RemoveStreamByOutput(protocol string) {
	this.Lock()
	defer this.Unlock()

	for _, stream := range this.Maps {
		if strings.EqualFold(stream.Protocol, protocol) {
			this.RemoveStream(stream.StreamId)
		}
	}

	return
}

func (this *ProxyStreams) AddStream(streamId string, proxyURI string, requestURI string, protocol string, receiver IReceiver, connect net.Conn) error {
	var err error
	this.Lock()
	defer this.Unlock()

	if len(streamId) == 0 {
		err = errors.New("StreamId can not be empty")
		return err
	}

	if len(proxyURI) == 0 {
		err = errors.New("URI can not be empty")
		return err
	}

	if receiver == nil {
		err = errors.New("Receiver can not be empty")
		return err
	}

	urlCtx, err := url.Parse(proxyURI)
	if err != nil {
		return err
	}
	if len(proxyURI) < len(urlCtx.Scheme)+3 {
		err = errors.New("URI is invalid")
		return err
	}
	if !strings.EqualFold(proxyURI[len(urlCtx.Scheme):len(urlCtx.Scheme)+3], "://") {
		err = errors.New("URI is invalid")
		return err
	}

	if _, ok := this.Maps[streamId]; ok {
		err = errors.New("ProxyStream stream id: " + streamId + " already in use")
		return err
	}

	this.Maps[streamId] = &ProxyStream{
		StreamId:    streamId,
		InputStatus: false,
		ProxyURI:    proxyURI,
		RequestURI:  requestURI,
		Protocol:    strings.ToLower(protocol),
		Connect:     connect,
		RemoteAddr:  connect.RemoteAddr().String(),
		Network:     connect.RemoteAddr().Network(),
		Receiver:    receiver,
	}

	return nil
}

func (this *ProxyStreams) RemoveStream(streamId string) {
	// this.Lock()
	// defer this.Unlock()

	if stream, ok := this.Maps[streamId]; ok && stream != nil && stream.Receiver != nil {
		delete(this.Maps, streamId)
		stream.Receiver.Close()
	}
}
