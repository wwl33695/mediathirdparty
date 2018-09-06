package pktque

import (
	"time"
)

/*
pop                                   push

     seg                 seg        seg
  |--------|         |---------|   |---|
     20ms                40ms       5ms
----------------- time -------------------->
headtm                               tailtm
*/

type tlSeg struct {
	tm, dur time.Duration
}

type Timeline struct {
	segs   []tlSeg
	headtm time.Duration
}

func (this *Timeline) Push(tm time.Duration, dur time.Duration) {
	if len(this.segs) > 0 {
		tail := this.segs[len(this.segs)-1]
		diff := tm - (tail.tm + tail.dur)
		if diff < 0 {
			tm -= diff
		}
	}
	this.segs = append(this.segs, tlSeg{tm, dur})
}

func (this *Timeline) Pop(dur time.Duration) (tm time.Duration) {
	if len(this.segs) == 0 {
		return this.headtm
	}

	tm = this.segs[0].tm
	for dur > 0 && len(this.segs) > 0 {
		seg := &this.segs[0]
		sub := dur
		if seg.dur < sub {
			sub = seg.dur
		}
		seg.dur -= sub
		dur -= sub
		seg.tm += sub
		this.headtm += sub
		if seg.dur == 0 {
			copy(this.segs[0:], this.segs[1:])
			this.segs = this.segs[:len(this.segs)-1]
		}
	}

	return
}
