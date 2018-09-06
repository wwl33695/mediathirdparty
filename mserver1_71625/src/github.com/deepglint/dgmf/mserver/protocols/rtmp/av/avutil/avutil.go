package avutil

import (
	"bytes"
	"fmt"
	"github.com/deepglint/dgmf/mserver/protocols/rtmp/av"
	"io"
	"net/url"
	"os"
	"path"
	"strings"
)

type HandlerDemuxer struct {
	av.Demuxer
	r io.ReadCloser
}

func (this *HandlerDemuxer) Close() error {
	return this.r.Close()
}

type HandlerMuxer struct {
	av.Muxer
	w     io.WriteCloser
	stage int
}

func (this *HandlerMuxer) WriteHeader(streams []av.CodecData) (err error) {
	if this.stage == 0 {
		if err = this.Muxer.WriteHeader(streams); err != nil {
			return
		}
		this.stage++
	}
	return
}

func (this *HandlerMuxer) WriteTrailer() (err error) {
	if this.stage == 1 {
		this.stage++
		if err = this.Muxer.WriteTrailer(); err != nil {
			return
		}
	}
	return
}

func (this *HandlerMuxer) Close() (err error) {
	if err = this.WriteTrailer(); err != nil {
		return
	}
	return this.w.Close()
}

type RegisterHandler struct {
	Ext           string
	ReaderDemuxer func(io.Reader) av.Demuxer
	WriterMuxer   func(io.Writer) av.Muxer
	UrlMuxer      func(string) (bool, av.MuxCloser, error)
	UrlDemuxer    func(string) (bool, av.DemuxCloser, error)
	UrlReader     func(string) (bool, io.ReadCloser, error)
	Probe         func([]byte) bool
	AudioEncoder  func(av.CodecType) (av.AudioEncoder, error)
	AudioDecoder  func(av.AudioCodecData) (av.AudioDecoder, error)
	ServerDemuxer func(string) (bool, av.DemuxCloser, error)
	ServerMuxer   func(string) (bool, av.MuxCloser, error)
	CodecTypes    []av.CodecType
}

type Handlers struct {
	handlers []RegisterHandler
}

func (this *Handlers) Add(fn func(*RegisterHandler)) {
	handler := &RegisterHandler{}
	fn(handler)
	this.handlers = append(this.handlers, *handler)
}

func (this *Handlers) openUrl(u *url.URL, uri string) (r io.ReadCloser, err error) {
	if u != nil && u.Scheme != "" {
		for _, handler := range this.handlers {
			if handler.UrlReader != nil {
				var ok bool
				if ok, r, err = handler.UrlReader(uri); ok {
					return
				}
			}
		}
		err = fmt.Errorf("avutil: openUrl %s failed", uri)
	} else {
		r, err = os.Open(uri)
	}
	return
}

func (this *Handlers) createUrl(u *url.URL, uri string) (w io.WriteCloser, err error) {
	w, err = os.Create(uri)
	return
}

func (this *Handlers) NewAudioEncoder(typ av.CodecType) (enc av.AudioEncoder, err error) {
	for _, handler := range this.handlers {
		if handler.AudioEncoder != nil {
			if enc, _ = handler.AudioEncoder(typ); enc != nil {
				return
			}
		}
	}
	err = fmt.Errorf("avutil: encoder", typ, "not found")
	return
}

func (this *Handlers) NewAudioDecoder(codec av.AudioCodecData) (dec av.AudioDecoder, err error) {
	for _, handler := range this.handlers {
		if handler.AudioDecoder != nil {
			if dec, _ = handler.AudioDecoder(codec); dec != nil {
				return
			}
		}
	}
	err = fmt.Errorf("avutil: decoder", codec.Type(), "not found")
	return
}

func (this *Handlers) Open(uri string) (demuxer av.DemuxCloser, err error) {
	listen := false
	if strings.HasPrefix(uri, "listen:") {
		uri = uri[len("listen:"):]
		listen = true
	}

	for _, handler := range this.handlers {
		if listen {
			if handler.ServerDemuxer != nil {
				var ok bool
				if ok, demuxer, err = handler.ServerDemuxer(uri); ok {
					return
				}
			}
		} else {
			if handler.UrlDemuxer != nil {
				var ok bool
				if ok, demuxer, err = handler.UrlDemuxer(uri); ok {
					return
				}
			}
		}
	}

	var r io.ReadCloser
	var ext string
	var u *url.URL
	if u, _ = url.Parse(uri); u != nil && u.Scheme != "" {
		ext = path.Ext(u.Path)
	} else {
		ext = path.Ext(uri)
	}

	if ext != "" {
		for _, handler := range this.handlers {
			if handler.Ext == ext {
				if handler.ReaderDemuxer != nil {
					if r, err = this.openUrl(u, uri); err != nil {
						return
					}
					demuxer = &HandlerDemuxer{
						Demuxer: handler.ReaderDemuxer(r),
						r:       r,
					}
					return
				}
			}
		}
	}

	var probebuf [1024]byte
	if r, err = this.openUrl(u, uri); err != nil {
		return
	}
	if _, err = io.ReadFull(r, probebuf[:]); err != nil {
		return
	}

	for _, handler := range this.handlers {
		if handler.Probe != nil && handler.Probe(probebuf[:]) && handler.ReaderDemuxer != nil {
			var _r io.Reader
			if rs, ok := r.(io.ReadSeeker); ok {
				if _, err = rs.Seek(0, 0); err != nil {
					return
				}
				_r = rs
			} else {
				_r = io.MultiReader(bytes.NewReader(probebuf[:]), r)
			}
			demuxer = &HandlerDemuxer{
				Demuxer: handler.ReaderDemuxer(_r),
				r:       r,
			}
			return
		}
	}

	r.Close()
	err = fmt.Errorf("avutil: open %s failed", uri)
	return
}

func (this *Handlers) Create(uri string) (muxer av.MuxCloser, err error) {
	_, muxer, err = this.FindCreate(uri)
	return
}

func (this *Handlers) FindCreate(uri string) (handler RegisterHandler, muxer av.MuxCloser, err error) {
	listen := false
	if strings.HasPrefix(uri, "listen:") {
		uri = uri[len("listen:"):]
		listen = true
	}

	for _, handler = range this.handlers {
		if listen {
			if handler.ServerMuxer != nil {
				var ok bool
				if ok, muxer, err = handler.ServerMuxer(uri); ok {
					return
				}
			}
		} else {
			if handler.UrlMuxer != nil {
				var ok bool
				if ok, muxer, err = handler.UrlMuxer(uri); ok {
					return
				}
			}
		}
	}

	var ext string
	var u *url.URL
	if u, _ = url.Parse(uri); u != nil && u.Scheme != "" {
		ext = path.Ext(u.Path)
	} else {
		ext = path.Ext(uri)
	}

	if ext != "" {
		for _, handler = range this.handlers {
			if handler.Ext == ext && handler.WriterMuxer != nil {
				var w io.WriteCloser
				if w, err = this.createUrl(u, uri); err != nil {
					return
				}
				muxer = &HandlerMuxer{
					Muxer: handler.WriterMuxer(w),
					w:     w,
				}
				return
			}
		}
	}

	err = fmt.Errorf("avutil: create muxer %s failed", uri)
	return
}

var DefaultHandlers = &Handlers{}

func Open(url string) (demuxer av.DemuxCloser, err error) {
	return DefaultHandlers.Open(url)
}

func Create(url string) (muxer av.MuxCloser, err error) {
	return DefaultHandlers.Create(url)
}

func CopyPackets(dst av.PacketWriter, src av.PacketReader) (err error) {
	for {
		var pkt av.Packet
		if pkt, err = src.ReadPacket(); err != nil {
			if err == io.EOF {
				break
			}
			return
		}
		if err = dst.WritePacket(pkt); err != nil {
			return
		}
	}
	return
}

func CopyFile(dst av.Muxer, src av.Demuxer) (err error) {
	var streams []av.CodecData
	if streams, err = src.Streams(); err != nil {
		return
	}
	if err = dst.WriteHeader(streams); err != nil {
		return
	}
	if err = CopyPackets(dst, src); err != nil {
		return
	}
	if err = dst.WriteTrailer(); err != nil {
		return
	}
	return
}
