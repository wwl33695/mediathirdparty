package dmid

import (
	"fmt"
	"bytes"
	"encoding/binary"
)

func test() {
	fmt.Println("aaa")
}

func (self *DMIDProto) GetDMIDHeader() []byte {

	buffer := bytes.NewBuffer([]byte{})

	//DMID header
	binary.Write(buffer, binary.LittleEndian, uint8('D'))
	binary.Write(buffer, binary.LittleEndian, uint8('M'))
	binary.Write(buffer, binary.LittleEndian, uint8('I'))
	binary.Write(buffer, binary.LittleEndian, uint8('D'))
	binary.Write(buffer, binary.LittleEndian, self.Version)

	//stream header
	binary.Write(buffer, binary.LittleEndian, self.PacketID)
	binary.Write(buffer, binary.LittleEndian, self.TimeStamp)
	binary.Write(buffer, binary.LittleEndian, self.SessionID)
//	binary.Write(buffer, binary.LittleEndian, self.StreamID)
	binary.Write(buffer, binary.LittleEndian, self.ChannelCount)

	return buffer.Bytes()
}

func (self *ChannelInfo) GetChannelHeader() []byte {

	buffer := bytes.NewBuffer([]byte{})

	//channel header
	binary.Write(buffer, binary.LittleEndian, self.ChannelID)
//	binary.Write(buffer, binary.LittleEndian, self.ChannelType)
	binary.Write(buffer, binary.LittleEndian, self.FrameID)
	binary.Write(buffer, binary.LittleEndian, self.FrameTime)
	binary.Write(buffer, binary.LittleEndian, self.SliceCount)
	binary.Write(buffer, binary.LittleEndian, self.SliceID)
	binary.Write(buffer, binary.LittleEndian, self.PayloadType)
	binary.Write(buffer, binary.LittleEndian, self.PayloadLength)

	return buffer.Bytes()
}
