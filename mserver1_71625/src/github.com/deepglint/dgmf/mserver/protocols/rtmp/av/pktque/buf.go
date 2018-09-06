package pktque

import (
	"github.com/deepglint/dgmf/mserver/protocols/rtmp/av"
)

type Buf struct {
	Head, Tail BufPos
	pkts       []av.Packet
	Size       int
	Count      int
}

func NewBuf() *Buf {
	return &Buf{
		pkts: make([]av.Packet, 64),
	}
}

func (this *Buf) Pop() av.Packet {
	if this.Count == 0 {
		panic("pktque.Buf: Pop() when count == 0")
	}

	i := int(this.Head) & (len(this.pkts) - 1)
	pkt := this.pkts[i]
	this.pkts[i] = av.Packet{}
	this.Size -= len(pkt.Data)
	this.Head++
	this.Count--

	return pkt
}

func (this *Buf) grow() {
	newpkts := make([]av.Packet, len(this.pkts)*2)
	for i := this.Head; i.LT(this.Tail); i++ {
		newpkts[int(i)&(len(newpkts)-1)] = this.pkts[int(i)&(len(this.pkts)-1)]
	}
	this.pkts = newpkts
}

func (this *Buf) Push(pkt av.Packet) {
	if this.Count == len(this.pkts) {
		this.grow()
	}
	this.pkts[int(this.Tail)&(len(this.pkts)-1)] = pkt
	this.Tail++
	this.Count++
	this.Size += len(pkt.Data)
}

func (this *Buf) Get(pos BufPos) av.Packet {
	return this.pkts[int(pos)&(len(this.pkts)-1)]
}

func (this *Buf) IsValidPos(pos BufPos) bool {
	return pos.GE(this.Head) && pos.LT(this.Tail)
}

type BufPos int

func (this BufPos) LT(pos BufPos) bool {
	return this-pos < 0
}

func (this BufPos) GE(pos BufPos) bool {
	return this-pos >= 0
}

func (this BufPos) GT(pos BufPos) bool {
	return this-pos > 0
}
