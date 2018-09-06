package codec

import (
	"time"

	"github.com/deepglint/dgmf/mserver/protocols/rtmp/av"
)

type PCMUCodecData struct {
	typ av.CodecType
}

func (this PCMUCodecData) Type() av.CodecType {
	return this.typ
}

func (this PCMUCodecData) SampleRate() int {
	return 8000
}

func (this PCMUCodecData) ChannelLayout() av.ChannelLayout {
	return av.CH_MONO
}

func (this PCMUCodecData) SampleFormat() av.SampleFormat {
	return av.S16
}

func (this PCMUCodecData) PacketDuration(data []byte) (time.Duration, error) {
	return time.Duration(len(data)) * time.Second / time.Duration(8000), nil
}

func NewPCMMulawCodecData() av.AudioCodecData {
	return PCMUCodecData{
		typ: av.PCM_MULAW,
	}
}

func NewPCMAlawCodecData() av.AudioCodecData {
	return PCMUCodecData{
		typ: av.PCM_ALAW,
	}
}

type SpeexCodecData struct {
	CodecData
}

func (this SpeexCodecData) PacketDuration(data []byte) (time.Duration, error) {
	// libavcodec/libspeexdec.c
	// samples = samplerate/50
	// duration = 0.02s
	return time.Millisecond * 20, nil
}

func NewSpeexCodecData(sr int, cl av.ChannelLayout) SpeexCodecData {
	codec := SpeexCodecData{}
	codec.CodecType_ = av.SPEEX
	codec.SampleFormat_ = av.S16
	codec.SampleRate_ = sr
	codec.ChannelLayout_ = cl
	return codec
}

type CodecData struct {
	CodecType_     av.CodecType
	SampleRate_    int
	SampleFormat_  av.SampleFormat
	ChannelLayout_ av.ChannelLayout
}

func (this CodecData) Type() av.CodecType {
	return this.CodecType_
}

func (this CodecData) SampleFormat() av.SampleFormat {
	return this.SampleFormat_
}

func (this CodecData) ChannelLayout() av.ChannelLayout {
	return this.ChannelLayout_
}

func (this CodecData) SampleRate() int {
	return this.SampleRate_
}
