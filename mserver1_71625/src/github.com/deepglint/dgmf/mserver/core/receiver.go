package core

import (
	"time"
)

type IReceiver interface {
	Open(uri string, streamId string, rtms chan RTMessage)
	Close()
	Frames() <-chan *H264ESFrame
	Running() bool
	SetTimeout(timeout time.Duration)
	FPS() uint32
	Index() uint64
	PPS() string
	SPS() string
	Width() uint32
	Height() uint32
}

type RTMessage struct {
	Status int
	Error  error
}

type Receiver struct {
	AbsRunning    bool
	AbsFrames     chan *H264ESFrame
	AbsRTMessages chan RTMessage
	AbsStoped     chan bool
	AbsTimeout    time.Duration
	AbsStreamId   string
	AbsFPS        uint32
	AbsSPS        string
	AbsPPS        string
	AbsWidth      uint32
	AbsHeight     uint32
	AbsIndex      uint64
}

func (this *Receiver) Frames() <-chan *H264ESFrame {
	return this.AbsFrames
}

func (this *Receiver) Running() bool {
	return this.AbsRunning
}

func (this *Receiver) SetTimeout(timeout time.Duration) {
	this.AbsTimeout = timeout
}

func (this *Receiver) FPS() uint32 {
	return this.AbsFPS
}

func (this *Receiver) Index() uint64 {
	return this.AbsIndex
}

func (this *Receiver) PPS() string {
	return this.AbsPPS
}

func (this *Receiver) SPS() string {
	return this.AbsSPS
}

func (this *Receiver) Width() uint32 {
	return this.AbsWidth
}

func (this *Receiver) Height() uint32 {
	return this.AbsHeight
}
