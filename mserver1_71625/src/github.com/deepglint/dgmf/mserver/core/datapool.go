package core

type H264ESFrame struct {
	Data      []byte
	Timestamp uint32
	IFrame    bool
	Index     uint64
}

type Output struct {
	Protocol string
	Enable   bool
	Param    interface{}
}

// type VodSession struct {
// 	SessionId   string
// 	RemoteAddr  string
// 	Network     string
// 	Protocol    string
// 	Index       int
// 	VodInputCtx VodInputLayer
// 	Frame       chan *H264ESFrame
// 	Connect     net.Conn
// }

// type VodStream struct {
// 	StreamId  string
// 	URI       string
// 	Sessions  map[string]*VodSession
// 	Outputs   map[string]*Output
// 	MuxFormat string
// 	Filename  string
// 	Fps       uint32
// 	Width     int
// 	Height    int
// 	SPS       string
// 	PPS       string
// 	FMTP      string
// }

type ESPool struct {
	Live  *LiveStreams
	Proxy *ProxyStreams
}

var esPool *ESPool

func GetESPool() *ESPool {
	if esPool == nil {
		esPool = &ESPool{
			Live:  NewLiveStreams(),
			Proxy: NewProxyStreams(),
		}
	}
	return esPool
}

// func (this *ESPool) RemoveOutput(streamId string, streamType string, protocol string) error {
// 	// if isLive {
// 	if _, ok := this.LiveStreams[streamId]; ok == false {
// 		return errors.New(streamId + " not found")
// 	}
// 	delete(this.LiveStreams[streamId].Outputs, strings.ToLower(protocol))
// 	// } else {
// 	// 	if _, ok := this.VodStreams[streamId]; ok == false {
// 	// 		return errors.New(streamId + " not found")
// 	// 	}
// 	// 	delete(this.VodStreams[streamId].Outputs, strings.ToLower(protocol))
// 	// }
// 	return nil
// }

// func (this *ESPool) RemoveStream(streamId string, streamType string) error {
// 	var err error
// 	// if isLive {
// 	if _, ok := this.LiveStreams[streamId]; ok == true {
// 		for _, session := range this.LiveStreams[streamId].Sessions {
// 			if _, ook := <-session.Frame; ook == true {
// 				close(session.Frame)
// 			}
// 		}
// 		if this.LiveStreams[streamId].InputCtx != nil {
// 			this.LiveStreams[streamId].InputCtx.Close()
// 		}
// 		delete(this.LiveStreams, streamId)
// 	} else {
// 		return errors.New(streamId + " not found")
// 	}
// 	// } else {
// 	// 	if _, ok := this.VodStreams[streamId]; ok == true {
// 	// 		for _, session := range this.VodStreams[streamId].Sessions {
// 	// 			session.VodInputCtx.Close()
// 	// 			// close(session.Frame)
// 	// 		}
// 	// 		delete(this.VodStreams, streamId)
// 	// 	} else {
// 	// 		return errors.New(streamId + " not found")
// 	// 	}
// 	// }
// 	return err
// }
