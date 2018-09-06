package gb28181

import (
	"bytes"
	"encoding/binary"
)

func GetRR(localssrc, remotessrc uint32, highestSqNum uint16) []byte {
	buffer := bytes.NewBuffer([]byte{})
	binary.Write(buffer, binary.LittleEndian, uint8(0x81))
	binary.Write(buffer, binary.BigEndian, uint8(0xc9))
	binary.Write(buffer, binary.BigEndian, uint8(0x0))
	binary.Write(buffer, binary.BigEndian, uint8(0x7))
	binary.Write(buffer, binary.BigEndian, localssrc)
	binary.Write(buffer, binary.BigEndian, remotessrc)
	binary.Write(buffer, binary.BigEndian, uint32(0x0))

	binary.Write(buffer, binary.BigEndian, uint16(0x01))
	binary.Write(buffer, binary.BigEndian, highestSqNum)

	binary.Write(buffer, binary.BigEndian, uint32(0x0))
	binary.Write(buffer, binary.BigEndian, uint32(0x0))
	binary.Write(buffer, binary.BigEndian, uint32(0x0))

	binary.Write(buffer, binary.BigEndian, GetSDES(localssrc))

	return buffer.Bytes()
}

func GetSDES(localssrc uint32) []byte {
	buffer := bytes.NewBuffer([]byte{})
	binary.Write(buffer, binary.LittleEndian, uint8(0x81))
	binary.Write(buffer, binary.BigEndian, uint8(0xca))
	binary.Write(buffer, binary.BigEndian, uint8(0x0))
	binary.Write(buffer, binary.BigEndian, uint8(0x7))

	binary.Write(buffer, binary.BigEndian, localssrc)

	binary.Write(buffer, binary.BigEndian, uint8(0x1))

	binary.Write(buffer, binary.BigEndian, uint8(0x13))

	binary.Write(buffer, binary.BigEndian, []byte("deepglint@DG0102021"))

	/*	binary.Write(buffer, binary.BigEndian, uint32(0x65666768))
		binary.Write(buffer, binary.BigEndian, uint32(0x65666768))
		binary.Write(buffer, binary.BigEndian, uint32(0x65666768))
		binary.Write(buffer, binary.BigEndian, uint32(0x65666768))
		binary.Write(buffer, binary.BigEndian, uint32(0x65666768))
	*/

	binary.Write(buffer, binary.BigEndian, uint8(0x0))
	binary.Write(buffer, binary.BigEndian, uint16(0x0))
	return buffer.Bytes()
}
