// Package pktque provides packet Filter interface and structures used by other components.
package pktque

import (
	"github.com/deepglint/dgmf/mserver/protocols/rtmp/av"
	"time"
)

type Filter interface {
	// Change packet time or drop packet
	ModifyPacket(pkt *av.Packet, streams []av.CodecData, videoidx int, audioidx int) (drop bool, err error)
}

// Combine multiple Filters into one, ModifyPacket will be called in order.
type Filters []Filter

func (this Filters) ModifyPacket(pkt *av.Packet, streams []av.CodecData, videoidx int, audioidx int) (drop bool, err error) {
	for _, filter := range this {
		if drop, err = filter.ModifyPacket(pkt, streams, videoidx, audioidx); err != nil {
			return
		}
		if drop {
			return
		}
	}
	return
}

// Wrap origin Demuxer and Filter into a new Demuxer, when read this Demuxer filters will be called.
type FilterDemuxer struct {
	av.Demuxer
	Filter   Filter
	streams  []av.CodecData
	videoidx int
	audioidx int
}

func (this FilterDemuxer) ReadPacket() (pkt av.Packet, err error) {
	if this.streams == nil {
		if this.streams, err = this.Demuxer.Streams(); err != nil {
			return
		}
		for i, stream := range this.streams {
			if stream.Type().IsVideo() {
				this.videoidx = i
			} else if stream.Type().IsAudio() {
				this.audioidx = i
			}
		}
	}

	for {
		if pkt, err = this.Demuxer.ReadPacket(); err != nil {
			return
		}
		var drop bool
		if drop, err = this.Filter.ModifyPacket(&pkt, this.streams, this.videoidx, this.audioidx); err != nil {
			return
		}
		if !drop {
			break
		}
	}

	return
}

// Drop packets until first video key frame arrived.
type WaitKeyFrame struct {
	ok bool
}

func (this *WaitKeyFrame) ModifyPacket(pkt *av.Packet, streams []av.CodecData, videoidx int, audioidx int) (drop bool, err error) {
	if !this.ok && pkt.Idx == int8(videoidx) && pkt.IsKeyFrame {
		this.ok = true
	}
	drop = !this.ok
	return
}

// Fix incorrect packet timestamps.
type FixTime struct {
	zerobase      time.Duration
	incrbase      time.Duration
	lasttime      time.Duration
	StartFromZero bool // make timestamp start from zero
	MakeIncrement bool // force timestamp increment
}

func (this *FixTime) ModifyPacket(pkt *av.Packet, streams []av.CodecData, videoidx int, audioidx int) (drop bool, err error) {
	if this.StartFromZero {
		if this.zerobase == 0 {
			this.zerobase = pkt.Time
		}
		pkt.Time -= this.zerobase
	}

	if this.MakeIncrement {
		pkt.Time -= this.incrbase
		if this.lasttime == 0 {
			this.lasttime = pkt.Time
		}
		if pkt.Time < this.lasttime || pkt.Time > this.lasttime+time.Millisecond*500 {
			this.incrbase += pkt.Time - this.lasttime
			pkt.Time = this.lasttime
		}
		this.lasttime = pkt.Time
	}

	return
}

// Drop incorrect packets to make A/V sync.
type AVSync struct {
	MaxTimeDiff time.Duration
	time        []time.Duration
}

func (this *AVSync) ModifyPacket(pkt *av.Packet, streams []av.CodecData, videoidx int, audioidx int) (drop bool, err error) {
	if this.time == nil {
		this.time = make([]time.Duration, len(streams))
		if this.MaxTimeDiff == 0 {
			this.MaxTimeDiff = time.Millisecond * 500
		}
	}

	start, end, correctable, correcttime := this.check(int(pkt.Idx))
	if pkt.Time >= start && pkt.Time < end {
		this.time[pkt.Idx] = pkt.Time
	} else {
		if correctable {
			pkt.Time = correcttime
			for i := range this.time {
				this.time[i] = correcttime
			}
		} else {
			drop = true
		}
	}
	return
}

func (this *AVSync) check(i int) (start time.Duration, end time.Duration, correctable bool, correcttime time.Duration) {
	minidx := -1
	maxidx := -1
	for j := range this.time {
		if minidx == -1 || this.time[j] < this.time[minidx] {
			minidx = j
		}
		if maxidx == -1 || this.time[j] > this.time[maxidx] {
			maxidx = j
		}
	}
	allthesame := this.time[minidx] == this.time[maxidx]

	if i == maxidx {
		if allthesame {
			correctable = true
		} else {
			correctable = false
		}
	} else {
		correctable = true
	}

	start = this.time[minidx]
	end = start + this.MaxTimeDiff
	correcttime = start + time.Millisecond*40
	return
}

// Make packets reading speed as same as walltime, effect like ffmpeg -re option.
type Walltime struct {
	firsttime time.Time
}

func (this *Walltime) ModifyPacket(pkt *av.Packet, streams []av.CodecData, videoidx int, audioidx int) (drop bool, err error) {
	if pkt.Idx == 0 {
		if this.firsttime.IsZero() {
			this.firsttime = time.Now()
		}
		pkttime := this.firsttime.Add(pkt.Time)
		delta := pkttime.Sub(time.Now())
		if delta > 0 {
			time.Sleep(delta)
		}
	}
	return
}
