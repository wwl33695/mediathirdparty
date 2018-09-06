package dmid

import (
	"fmt"
	"bytes"
	"encoding/binary"
//	"time"

	"github.com/deepglint/dgmf/mserver/protocols/dmi/dmic"
)

func parser_test() {
	fmt.Println("aaa")
}

func (this *DMIDBuffer) SetData(data []byte) int {
	if this.Offset + len(data) > len(this.Buffer) {
		println("DMIDBuffer buffer overflow")
		return -1;
	}

	copy(this.Buffer[this.Offset:], data)
	this.Offset += len(data)
	return 0
}

func (this *DMIDBuffer) GetStartPos() int {
	if this.Offset < 10 {
		return -1;
	}

	var datatype int = -1

	var i int
	for i = 0;i < this.Offset - 3;i++ {
		if this.Buffer[i] == 'D' && this.Buffer[i+1] == 'M' && this.Buffer[i+2] == 'I' && this.Buffer[i+3] == 'D' {
			datatype = 1
			break;
		} else if this.Buffer[i] == 'D' && this.Buffer[i+1] == 'M' && this.Buffer[i+2] == 'I' && this.Buffer[i+3] == 'C' {
			datatype = 2
			break;
		}
	}

	if i == this.Offset -3 {
		return -1
	}

	copy(this.Buffer[0:], this.Buffer[i:this.Offset])
	this.Offset -= i	

	return datatype
}

func (this *DMIDBuffer) GetFrame() int {
	for {
		ret := this.GetStartPos()
		if ret < 0 {
			return -1
		}

		if ret == 1 && this.ParseDMID() < 0 {
			return -1
		} else if ret == 2 && this.ParseResponse() < 0 {
			return -1
		}
	}

	return 0;
}

func dumpDMIDProto(proto *DMIDProto) {
	println("Version=", proto.Version)
	println("PacketID=", proto.PacketID)
	println("TimeStamp=", proto.TimeStamp)
	println("SessionID=", proto.SessionID)
//	println("StreamID=", proto.StreamID)
	println("ChannelCount=", proto.ChannelCount)
}

func dumpChannelInfo(info *ChannelInfo) {
	println("ChannelID=", info.ChannelID)
//	println("ChannelType=", info.ChannelType)
	println("FrameID=", info.FrameID)
	println("FrameTime=", info.FrameTime)
	println("SliceCount=", info.SliceCount)
	println("SliceID=", info.SliceID)
	println("PayloadType=", info.PayloadType)
	println("PayloadLength=", info.PayloadLength)
}

func (this *DMIDBuffer) ParseDMID() int {

	var proto DMIDProto 
	var info ChannelInfo

	if this.Buffer[0] == 'D' && this.Buffer[1] == 'M' && this.Buffer[2] == 'I' && this.Buffer[3] == 'D' {
//		starttime := time.Now().UnixNano() / int64(time.Millisecond)

		headerlength := binary.Size(proto) + binary.Size(info) + 4
		if this.Offset < headerlength {
			return -1
		}

		buf := bytes.NewReader(this.Buffer[4:])
		binary.Read(buf, binary.LittleEndian, &proto)
//		dumpDMIDProto(&proto)
		binary.Read(buf, binary.LittleEndian, &info)
//		dumpChannelInfo(&info)

		payloadlen := int(info.PayloadLength)
		if this.Offset < headerlength + payloadlen {
//			println("buffer length is too short=========")
			return -1
		}

		if this.frameinfo[info.ChannelID] == nil {
			this.frameinfo[info.ChannelID] = &FrameInfo{}
		}

		if info.SliceID == 0 {
			this.frameinfo[info.ChannelID].dmidheader = proto
			this.frameinfo[info.ChannelID].Channelheader = info
			copy(this.frameinfo[info.ChannelID].Data[0:], this.Buffer[headerlength:headerlength+payloadlen])
			this.frameinfo[info.ChannelID].Length = payloadlen
			this.frameinfo[info.ChannelID].slicecount = 1
		} else {
			copy(this.frameinfo[info.ChannelID].Data[this.frameinfo[info.ChannelID].Length:], this.Buffer[headerlength:headerlength+payloadlen])
			this.frameinfo[info.ChannelID].Length += payloadlen
			this.frameinfo[info.ChannelID].slicecount++
		}

		if this.frameinfo[info.ChannelID].slicecount == this.frameinfo[info.ChannelID].Channelheader.SliceCount {
//			println(this.frameinfo[info.ChannelID].Channelheader.FrameID, this.frameinfo[info.ChannelID].Length)
			this.ChanFrame <- this.frameinfo[info.ChannelID]
		}

		copy(this.Buffer[0:], this.Buffer[headerlength+payloadlen:this.Offset])
		this.Offset -= headerlength + payloadlen

//		println("ParseDMID cost ", (time.Now().UnixNano() / int64(time.Millisecond))- starttime)

//		println("offset = ", this.Offset)

		return 0
	}

	return -1
}

func (this *DMIDBuffer) ParseResponse() int {

	if this.Buffer[0] == 'D' && this.Buffer[1] == 'M' && this.Buffer[2] == 'I' && this.Buffer[3] == 'C' {

//		starttime := time.Now().UnixNano() / int64(time.Millisecond)

		var i int
		for i = 0;i < this.Offset -3;i++ {
			if this.Buffer[i] == '\r' && this.Buffer[i+1] == '\n' && this.Buffer[i+2] == '\r' && this.Buffer[i+3] == '\n' {
				break				
			}  
		}

		if i == this.Offset - 3 {
			return -1;
		}

		response := string(this.Buffer[0:i+4])
		println(response)

		if this.RCallback != nil {
			errCode := dmic.ParseResponseHead(response)
			cseq := dmic.ParseResponseCSeq(response)
			this.RCallback(cseq, errCode)
		}

		copy(this.Buffer[0:], this.Buffer[i+4:])		
		this.Offset -= i+4

//		println("ParseResponse cost ", (time.Now().UnixNano() / int64(time.Millisecond))- starttime)
		return 0
	}
	
	return -1
}