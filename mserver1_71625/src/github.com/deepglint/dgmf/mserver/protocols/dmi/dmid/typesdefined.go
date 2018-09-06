package dmid

import (
)

type DMIDProto struct {
	Version uint8 //fixed
	PacketID uint32
	TimeStamp uint32
	SessionID uint32 //fixed
//	StreamID uint32 //fixed
	ChannelCount uint8 //fixed
}

type ChannelInfo struct {
	ChannelID uint8
//	ChannelType uint8
	FrameID uint32
	FrameTime int64
	SliceCount uint16
	SliceID uint16
	PayloadType uint8
	PayloadLength uint16
}

type FrameInfo struct {
	dmidheader DMIDProto
	Channelheader ChannelInfo
	Data [1 * 1024 * 1024]byte
	Length int
	slicecount uint16
}

type NetData struct {
	Data [500 * 1024]byte
	Length int
}

type DMIDBuffer struct {
	Buffer []byte
	Offset int
	frameinfo [256]*FrameInfo

	FCallback FrameCallback
	RCallback ResponseCallback
	FrameMapID uint32

	ChanFrame chan *FrameInfo
}

type FrameCallback func (data []byte, channelid, payloadtype uint8, frametime int64) int

type ResponseCallback func (request, statuscode string) int
