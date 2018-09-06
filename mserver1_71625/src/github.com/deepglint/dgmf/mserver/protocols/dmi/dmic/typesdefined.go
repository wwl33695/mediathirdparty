package dmic

import (
)

type DmicProto struct {
	UserID string
	ServerIP string
	ServerPort string
	UserName string
	Password string
	SessionID string
}

type ChannelInfo struct {
	ChannelID uint8
	PayloadType uint8
}

type MediaInfo struct {
	StreamType string
	Channels []ChannelInfo 
}

const (
	UNKNOWN = iota
	REGISTER1
	REGISTER2
	REQUESTSTREAM
	PLAY
	STREAMINGPULL
	STREAMINGPUSH
)