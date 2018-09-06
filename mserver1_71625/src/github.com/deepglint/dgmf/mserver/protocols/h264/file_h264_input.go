package h264

import (
	"bufio"
	"bytes"
	// "encoding/base64"
	"errors"
	// "fmt"
	"os"
	"strings"
	"time"

	"github.com/deepglint/dgmf/mserver/core"
)

type FileH264LiveInputLayer struct {
	File    *os.File
	Reader  *bufio.Reader
	running bool
	Uri     string
}

func (this *FileH264LiveInputLayer) Open(uri string, stream *core.LiveStream) {
	var err error
	if len(strings.Split(uri, "file://")) != 2 {
		return
	}
	this.Uri = strings.Split(uri, "file://")[1]
	this.running = true
	this.File, err = os.Open(this.Uri)
	if err != nil {
		return
	}
	this.Reader = bufio.NewReader(this.File)
	var ts uint32 = 0
	stream.Fps = 25
	go func(this *FileH264LiveInputLayer) {
		timer := time.NewTicker(40 * time.Millisecond)
		for this.running {
			select {
			case <-timer.C:
				data, err := ReadNextH264Nalu(this.Reader)
				if err != nil || len(data) <= 4 {
					this.File.Close()
					this.File, _ = os.Open(this.Uri)
					this.Reader = bufio.NewReader(this.File)
					continue
				}

				stream.Index++
				iFrame := false
				if data[4] == 0x67 {
					iFrame = true
				}

				frame := &core.H264ESFrame{
					Data:      data,
					Timestamp: ts,
					IFrame:    iFrame,
				}

				sp := GetLiveSPS(data)
				if sp.Height != 0 {
					stream.Height = sp.Height
				}
				if sp.Width != 0 {
					stream.Width = sp.Width
				}
				if len(sp.SPS) != 0 {
					stream.SPS = sp.SPS
				}
				if len(sp.PPS) != 0 {
					stream.PPS = sp.PPS
				}

				ts += 90000 / stream.Fps
				for _, session := range stream.Sessions {
					select {
					case session.Frame <- frame:
					default:
					}
				}
			}
		}
		timer.Stop()
		this.File.Close()

	}(this)
	return
}

func (this *FileH264LiveInputLayer) Close() {
	this.running = false
}

func (this *FileH264LiveInputLayer) Running() bool {
	return this.running
}

// type FileH264VodInputLayer struct {
// 	File    *os.File
// 	Reader  *bufio.Reader
// 	running bool
// 	stream  *core.VodStream
// 	session *core.VodSession
// }

// func (this *FileH264VodInputLayer) Open(filename string, stream *core.VodStream, session *core.VodSession) error {
// 	var err error
// 	this.File, err = os.Open(filename)
// 	if err != nil {
// 		return err
// 	}
// 	this.Reader = bufio.NewReader(this.File)
// 	this.running = true
// 	this.stream = stream
// 	this.session = session
// 	return nil
// }

// func (this *FileH264VodInputLayer) Start() {
// 	this.stream.Fps = 25
// 	var ts uint32 = 0
// 	timer := time.NewTicker(40 * time.Millisecond)
// 	for this.running {
// 		select {
// 		case <-timer.C:
// 			data, err := ReadNextH264Nalu(this.Reader)
// 			if err != nil || len(data) <= 4 {
// 				break
// 			}
// 			this.session.Index++
// 			iFrame := false
// 			if data[4] == 0x67 {
// 				iFrame = true
// 			}

// 			frame := &core.H264ESFrame{
// 				Data:      data,
// 				Timestamp: ts,
// 				IFrame:    iFrame,
// 			}

// 			ts += 90000 / this.stream.Fps
// 			select {
// 			case this.session.Frame <- frame:
// 			default:
// 			}
// 		}
// 	}
// 	timer.Stop()
// 	this.File.Close()
// 	close(this.session.Frame)
// }

// func (this *FileH264VodInputLayer) Close() error {
// 	this.running = false
// 	return nil
// }

func ReadNextH264Nalu(reader *bufio.Reader) ([]byte, error) {
	var code []byte
	var err error
	var buf bytes.Buffer

	code, err = reader.Peek(5)
	if err != nil {
		return buf.Bytes(), err
	}

	if (code[0] == 0x00 && code[1] == 0x00 && code[2] == 0x00 && code[3] == 0x01 && code[4]&0x1F == 0x07) ||
		(code[0] == 0x00 && code[1] == 0x00 && code[2] == 0x00 && code[3] == 0x01 && code[4]&0x1F == 0x01) {
		for i := 0; i < 5; i++ {
			b, _ := reader.ReadByte()
			buf.WriteByte(b)
		}
	} else {
		return buf.Bytes(), errors.New("H264 nalu format error")
	}

	for {
		code, err = reader.Peek(5)
		if err != nil {
			return buf.Bytes(), err
		}

		if (code[0] == 0x00 && code[1] == 0x00 && code[2] == 0x00 && code[3] == 0x01 && code[4]&0x1F == 0x07) ||
			(code[0] == 0x00 && code[1] == 0x00 && code[2] == 0x00 && code[3] == 0x01 && code[4]&0x1F == 0x01) {
			break
		} else {
			b, _ := reader.ReadByte()
			buf.WriteByte(b)
		}
	}

	return buf.Bytes(), nil
}
