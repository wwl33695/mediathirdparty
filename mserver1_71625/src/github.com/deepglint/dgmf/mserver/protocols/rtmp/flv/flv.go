package flv

import (
	"bufio"
	"fmt"
	"io"

	"github.com/deepglint/dgmf/mserver/protocols/rtmp/av"
	"github.com/deepglint/dgmf/mserver/protocols/rtmp/av/avutil"
	"github.com/deepglint/dgmf/mserver/protocols/rtmp/codec"
	"github.com/deepglint/dgmf/mserver/protocols/rtmp/codec/aacparser"
	"github.com/deepglint/dgmf/mserver/protocols/rtmp/codec/h264parser"
	"github.com/deepglint/dgmf/mserver/utils/bits/pio"
)

var MaxProbePacketCount = 20

func NewMetadataByStreams(streams []av.CodecData) (metadata AMFMap, err error) {
	metadata = AMFMap{}

	for _, _stream := range streams {
		typ := _stream.Type()
		switch {
		case typ.IsVideo():
			stream := _stream.(av.VideoCodecData)
			switch typ {
			case av.H264:
				metadata["videocodecid"] = VIDEO_H264

			default:
				err = fmt.Errorf("flv: metadata: unsupported video codecType=%v", stream.Type())
				return
			}

			metadata["width"] = stream.Width()
			metadata["height"] = stream.Height()
			metadata["displayWidth"] = stream.Width()
			metadata["displayHeight"] = stream.Height()

		case typ.IsAudio():
			stream := _stream.(av.AudioCodecData)
			switch typ {
			case av.AAC:
				metadata["audiocodecid"] = SOUND_AAC

			case av.SPEEX:
				metadata["audiocodecid"] = SOUND_SPEEX

			default:
				err = fmt.Errorf("flv: metadata: unsupported audio codecType=%v", stream.Type())
				return
			}

			metadata["audiosamplerate"] = stream.SampleRate()
		}
	}

	return
}

type Prober struct {
	HasAudio, HasVideo             bool
	GotAudio, GotVideo             bool
	VideoStreamIdx, AudioStreamIdx int
	PushedCount                    int
	Streams                        []av.CodecData
	CachedPkts                     []av.Packet
}

func (this *Prober) CacheTag(_tag Tag, timestamp int32) {
	pkt, _ := this.TagToPacket(_tag, timestamp)
	this.CachedPkts = append(this.CachedPkts, pkt)
}

func (this *Prober) PushTag(tag Tag, timestamp int32) (err error) {
	this.PushedCount++

	if this.PushedCount > MaxProbePacketCount {
		err = fmt.Errorf("flv: max probe packet count reached")
		return
	}

	switch tag.Type {
	case TAG_VIDEO:
		switch tag.AVCPacketType {
		case AVC_SEQHDR:
			if !this.GotVideo {
				var stream h264parser.CodecData
				if stream, err = h264parser.NewCodecDataFromAVCDecoderConfRecord(tag.Data); err != nil {
					err = fmt.Errorf("flv: h264 seqhdr invalid")
					return
				}
				this.VideoStreamIdx = len(this.Streams)
				this.Streams = append(this.Streams, stream)
				this.GotVideo = true
			}

		case AVC_NALU:
			this.CacheTag(tag, timestamp)
		}

	case TAG_AUDIO:
		switch tag.SoundFormat {
		case SOUND_AAC:
			switch tag.AACPacketType {
			case AAC_SEQHDR:
				if !this.GotAudio {
					var stream aacparser.CodecData
					if stream, err = aacparser.NewCodecDataFromMPEG4AudioConfigBytes(tag.Data); err != nil {
						err = fmt.Errorf("flv: aac seqhdr invalid")
						return
					}
					this.AudioStreamIdx = len(this.Streams)
					this.Streams = append(this.Streams, stream)
					this.GotAudio = true
				}

			case AAC_RAW:
				this.CacheTag(tag, timestamp)
			}

		case SOUND_SPEEX:
			if !this.GotAudio {
				stream := codec.NewSpeexCodecData(16000, tag.ChannelLayout())
				this.AudioStreamIdx = len(this.Streams)
				this.Streams = append(this.Streams, stream)
				this.GotAudio = true
				this.CacheTag(tag, timestamp)
			}

		case SOUND_NELLYMOSER:
			if !this.GotAudio {
				stream := codec.CodecData{
					CodecType_:     av.NELLYMOSER,
					SampleRate_:    16000,
					SampleFormat_:  av.S16,
					ChannelLayout_: tag.ChannelLayout(),
				}
				this.AudioStreamIdx = len(this.Streams)
				this.Streams = append(this.Streams, stream)
				this.GotAudio = true
				this.CacheTag(tag, timestamp)
			}

		}
	}

	return
}

func (this *Prober) Probed() (ok bool) {
	if this.HasAudio || this.HasVideo {
		if this.HasAudio == this.GotAudio && this.HasVideo == this.GotVideo {
			return true
		}
	} else {
		if this.PushedCount == MaxProbePacketCount {
			return true
		}
	}
	return
}

func (this *Prober) TagToPacket(tag Tag, timestamp int32) (pkt av.Packet, ok bool) {
	switch tag.Type {
	case TAG_VIDEO:
		pkt.Idx = int8(this.VideoStreamIdx)
		switch tag.AVCPacketType {
		case AVC_NALU:
			ok = true
			pkt.Data = tag.Data
			pkt.CompositionTime = TsToTime(tag.CompositionTime)
			pkt.IsKeyFrame = tag.FrameType == FRAME_KEY
		}

	case TAG_AUDIO:
		pkt.Idx = int8(this.AudioStreamIdx)
		switch tag.SoundFormat {
		case SOUND_AAC:
			switch tag.AACPacketType {
			case AAC_RAW:
				ok = true
				pkt.Data = tag.Data
			}

		case SOUND_SPEEX:
			ok = true
			pkt.Data = tag.Data

		case SOUND_NELLYMOSER:
			ok = true
			pkt.Data = tag.Data
		}
	}

	pkt.Time = TsToTime(timestamp)
	return
}

func (this *Prober) Empty() bool {
	return len(this.CachedPkts) == 0
}

func (this *Prober) PopPacket() av.Packet {
	pkt := this.CachedPkts[0]
	this.CachedPkts = this.CachedPkts[1:]
	return pkt
}

func CodecDataToTag(stream av.CodecData) (_tag Tag, ok bool, err error) {
	switch stream.Type() {
	case av.H264:
		h264 := stream.(h264parser.CodecData)
		tag := Tag{
			Type:          TAG_VIDEO,
			AVCPacketType: AVC_SEQHDR,
			CodecID:       VIDEO_H264,
			Data:          h264.AVCDecoderConfRecordBytes(),
			FrameType:     FRAME_KEY,
		}
		ok = true
		_tag = tag

	case av.NELLYMOSER:
	case av.SPEEX:

	case av.AAC:
		aac := stream.(aacparser.CodecData)
		tag := Tag{
			Type:          TAG_AUDIO,
			SoundFormat:   SOUND_AAC,
			SoundRate:     SOUND_44Khz,
			AACPacketType: AAC_SEQHDR,
			Data:          aac.MPEG4AudioConfigBytes(),
		}
		switch aac.SampleFormat().BytesPerSample() {
		case 1:
			tag.SoundSize = SOUND_8BIT
		default:
			tag.SoundSize = SOUND_16BIT
		}
		switch aac.ChannelLayout().Count() {
		case 1:
			tag.SoundType = SOUND_MONO
		case 2:
			tag.SoundType = SOUND_STEREO
		}
		ok = true
		_tag = tag

	default:
		err = fmt.Errorf("flv: unspported codecType=%v", stream.Type())
		return
	}
	return
}

func PacketToTag(pkt av.Packet, stream av.CodecData) (tag Tag, timestamp int32) {
	switch stream.Type() {
	case av.H264:
		tag = Tag{
			Type:            TAG_VIDEO,
			AVCPacketType:   AVC_NALU,
			CodecID:         VIDEO_H264,
			Data:            pkt.Data,
			CompositionTime: TimeToTs(pkt.CompositionTime),
		}
		if pkt.IsKeyFrame {
			tag.FrameType = FRAME_KEY
		} else {
			tag.FrameType = FRAME_INTER
		}

	case av.AAC:
		tag = Tag{
			Type:          TAG_AUDIO,
			SoundFormat:   SOUND_AAC,
			SoundRate:     SOUND_44Khz,
			AACPacketType: AAC_RAW,
			Data:          pkt.Data,
		}
		astream := stream.(av.AudioCodecData)
		switch astream.SampleFormat().BytesPerSample() {
		case 1:
			tag.SoundSize = SOUND_8BIT
		default:
			tag.SoundSize = SOUND_16BIT
		}
		switch astream.ChannelLayout().Count() {
		case 1:
			tag.SoundType = SOUND_MONO
		case 2:
			tag.SoundType = SOUND_STEREO
		}

	case av.SPEEX:
		tag = Tag{
			Type:        TAG_AUDIO,
			SoundFormat: SOUND_SPEEX,
			Data:        pkt.Data,
		}

	case av.NELLYMOSER:
		tag = Tag{
			Type:        TAG_AUDIO,
			SoundFormat: SOUND_NELLYMOSER,
			Data:        pkt.Data,
		}
	}

	timestamp = TimeToTs(pkt.Time)
	return
}

type Muxer struct {
	bufw    writeFlusher
	b       []byte
	streams []av.CodecData
}

type writeFlusher interface {
	io.Writer
	Flush() error
}

func NewMuxerWriteFlusher(w writeFlusher) *Muxer {
	return &Muxer{
		bufw: w,
		b:    make([]byte, 256),
	}
}

func NewMuxer(w io.Writer) *Muxer {
	return NewMuxerWriteFlusher(bufio.NewWriterSize(w, pio.RecommendBufioSize))
}

var CodecTypes = []av.CodecType{av.H264, av.AAC, av.SPEEX}

func (this *Muxer) WriteHeader(streams []av.CodecData) (err error) {
	var flags uint8
	for _, stream := range streams {
		if stream.Type().IsVideo() {
			flags |= FILE_HAS_VIDEO
		} else if stream.Type().IsAudio() {
			flags |= FILE_HAS_AUDIO
		}
	}

	n := FillFileHeader(this.b, flags)
	if _, err = this.bufw.Write(this.b[:n]); err != nil {
		return
	}

	for _, stream := range streams {
		var tag Tag
		var ok bool
		if tag, ok, err = CodecDataToTag(stream); err != nil {
			return
		}
		if ok {
			if err = WriteTag(this.bufw, tag, 0, this.b); err != nil {
				return
			}
		}
	}

	this.streams = streams
	return
}

func (this *Muxer) WritePacket(pkt av.Packet) (err error) {
	stream := this.streams[pkt.Idx]
	tag, timestamp := PacketToTag(pkt, stream)

	if err = WriteTag(this.bufw, tag, timestamp, this.b); err != nil {
		return
	}
	return
}

func (this *Muxer) WriteTrailer() (err error) {
	if err = this.bufw.Flush(); err != nil {
		return
	}
	return
}

type Demuxer struct {
	prober *Prober
	bufr   *bufio.Reader
	b      []byte
	stage  int
}

func NewDemuxer(r io.Reader) *Demuxer {
	return &Demuxer{
		bufr:   bufio.NewReaderSize(r, pio.RecommendBufioSize),
		prober: &Prober{},
		b:      make([]byte, 256),
	}
}

func (this *Demuxer) prepare() (err error) {
	for this.stage < 2 {
		switch this.stage {
		case 0:
			if _, err = io.ReadFull(this.bufr, this.b[:FileHeaderLength]); err != nil {
				return
			}
			var flags uint8
			var skip int
			if flags, skip, err = ParseFileHeader(this.b); err != nil {
				return
			}
			if _, err = this.bufr.Discard(skip); err != nil {
				return
			}
			if flags&FILE_HAS_AUDIO != 0 {
				this.prober.HasAudio = true
			}
			if flags&FILE_HAS_VIDEO != 0 {
				this.prober.HasVideo = true
			}
			this.stage++

		case 1:
			for !this.prober.Probed() {
				var tag Tag
				var timestamp int32
				if tag, timestamp, err = ReadTag(this.bufr, this.b); err != nil {
					return
				}
				if err = this.prober.PushTag(tag, timestamp); err != nil {
					return
				}
			}
			this.stage++
		}
	}
	return
}

func (this *Demuxer) Streams() (streams []av.CodecData, err error) {
	if err = this.prepare(); err != nil {
		return
	}
	streams = this.prober.Streams
	return
}

func (this *Demuxer) ReadPacket() (pkt av.Packet, err error) {
	if err = this.prepare(); err != nil {
		return
	}

	if !this.prober.Empty() {
		pkt = this.prober.PopPacket()
		return
	}

	for {
		var tag Tag
		var timestamp int32
		if tag, timestamp, err = ReadTag(this.bufr, this.b); err != nil {
			return
		}

		var ok bool
		if pkt, ok = this.prober.TagToPacket(tag, timestamp); ok {
			return
		}
	}

	return
}

func Handler(h *avutil.RegisterHandler) {
	h.Probe = func(b []byte) bool {
		return b[0] == 'F' && b[1] == 'L' && b[2] == 'V'
	}

	h.Ext = ".flv"

	h.ReaderDemuxer = func(r io.Reader) av.Demuxer {
		return NewDemuxer(r)
	}

	h.WriterMuxer = func(w io.Writer) av.Muxer {
		return NewMuxer(w)
	}

	h.CodecTypes = CodecTypes
}
