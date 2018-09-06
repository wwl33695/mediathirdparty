// Packege pubsub implements publisher-subscribers model used in multi-channel streaming.
package pubsub

import (
	"github.com/deepglint/dgmf/mserver/protocols/rtmp/av"
	"github.com/deepglint/dgmf/mserver/protocols/rtmp/av/pktque"
	"io"
	"sync"
	"time"
)

//        time
// ----------------->
//
// V-A-V-V-A-V-V-A-V-V
// |                 |
// 0        5        10
// head             tail
// oldest          latest
//

// One publisher and multiple subscribers thread-safe packet buffer queue.
type Queue struct {
	buf                      *pktque.Buf
	head, tail               int
	lock                     *sync.RWMutex
	cond                     *sync.Cond
	curgopcount, maxgopcount int
	streams                  []av.CodecData
	videoidx                 int
	closed                   bool
}

func NewQueue() *Queue {
	q := &Queue{}
	q.buf = pktque.NewBuf()
	q.maxgopcount = 2
	q.lock = &sync.RWMutex{}
	q.cond = sync.NewCond(q.lock.RLocker())
	q.videoidx = -1
	return q
}

func (this *Queue) SetMaxGopCount(n int) {
	this.lock.Lock()
	this.maxgopcount = n
	this.lock.Unlock()
	return
}

func (this *Queue) WriteHeader(streams []av.CodecData) error {
	this.lock.Lock()

	this.streams = streams
	for i, stream := range streams {
		if stream.Type().IsVideo() {
			this.videoidx = i
		}
	}
	this.cond.Broadcast()

	this.lock.Unlock()

	return nil
}

func (this *Queue) WriteTrailer() error {
	return nil
}

// After Close() called, all QueueCursor's ReadPacket will return io.EOF.
func (this *Queue) Close() (err error) {
	this.lock.Lock()

	this.closed = true
	this.cond.Broadcast()

	this.lock.Unlock()
	return
}

// Put packet into buffer, old packets will be discared.
func (this *Queue) WritePacket(pkt av.Packet) (err error) {
	this.lock.Lock()

	this.buf.Push(pkt)
	if pkt.Idx == int8(this.videoidx) && pkt.IsKeyFrame {
		this.curgopcount++
	}

	for this.curgopcount >= this.maxgopcount && this.buf.Count > 1 {
		pkt := this.buf.Pop()
		if pkt.Idx == int8(this.videoidx) && pkt.IsKeyFrame {
			this.curgopcount--
		}
		if this.curgopcount < this.maxgopcount {
			break
		}
	}
	//println("shrink", this.curgopcount, this.maxgopcount, this.buf.Head, this.buf.Tail, "count", this.buf.Count, "size", this.buf.Size)

	this.cond.Broadcast()

	this.lock.Unlock()
	return
}

type QueueCursor struct {
	que    *Queue
	pos    pktque.BufPos
	gotpos bool
	init   func(buf *pktque.Buf, videoidx int) pktque.BufPos
}

func (this *Queue) newCursor() *QueueCursor {
	return &QueueCursor{
		que: this,
	}
}

// Create cursor position at latest packet.
func (this *Queue) Latest() *QueueCursor {
	cursor := this.newCursor()
	cursor.init = func(buf *pktque.Buf, videoidx int) pktque.BufPos {
		return buf.Tail
	}
	return cursor
}

// Create cursor position at oldest buffered packet.
func (this *Queue) Oldest() *QueueCursor {
	cursor := this.newCursor()
	cursor.init = func(buf *pktque.Buf, videoidx int) pktque.BufPos {
		return buf.Head
	}
	return cursor
}

// Create cursor position at specific time in buffered packets.
func (this *Queue) DelayedTime(dur time.Duration) *QueueCursor {
	cursor := this.newCursor()
	cursor.init = func(buf *pktque.Buf, videoidx int) pktque.BufPos {
		i := buf.Tail - 1
		if buf.IsValidPos(i) {
			end := buf.Get(i)
			for buf.IsValidPos(i) {
				if end.Time-buf.Get(i).Time > dur {
					break
				}
				i--
			}
		}
		return i
	}
	return cursor
}

// Create cursor position at specific delayed GOP count in buffered packets.
func (this *Queue) DelayedGopCount(n int) *QueueCursor {
	cursor := this.newCursor()
	cursor.init = func(buf *pktque.Buf, videoidx int) pktque.BufPos {
		i := buf.Tail - 1
		if videoidx != -1 {
			for gop := 0; buf.IsValidPos(i) && gop < n; i-- {
				pkt := buf.Get(i)
				if pkt.Idx == int8(this.videoidx) && pkt.IsKeyFrame {
					gop++
				}
			}
		}
		return i
	}
	return cursor
}

func (this *QueueCursor) Streams() (streams []av.CodecData, err error) {
	this.que.cond.L.Lock()
	for this.que.streams == nil && !this.que.closed {
		this.que.cond.Wait()
	}
	if this.que.streams != nil {
		streams = this.que.streams
	} else {
		err = io.EOF
	}
	this.que.cond.L.Unlock()
	return
}

// ReadPacket will not consume packets in Queue, it's just a cursor.
func (this *QueueCursor) ReadPacket() (pkt av.Packet, err error) {
	this.que.cond.L.Lock()
	buf := this.que.buf
	if !this.gotpos {
		this.pos = this.init(buf, this.que.videoidx)
		this.gotpos = true
	}
	for {
		if this.pos.LT(buf.Head) {
			this.pos = buf.Head
		} else if this.pos.GT(buf.Tail) {
			this.pos = buf.Tail
		}
		if buf.IsValidPos(this.pos) {
			pkt = buf.Get(this.pos)
			this.pos++
			break
		}
		if this.que.closed {
			err = io.EOF
			break
		}
		this.que.cond.Wait()
	}
	this.que.cond.L.Unlock()
	return
}
