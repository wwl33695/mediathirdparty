package rtp

import (
	"encoding/binary"
	"errors"
)

/*
5.1 RTP Fixed Header Fields

   The RTP header has the following format:

    0                   1                   2                   3
    0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |V=2|P|X|  CC   |M|     PT      |       sequence number         |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |                           timestamp                           |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |           synchronization source (SSRC) identifier            |
   +=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+
   |            contributing source (CSRC) identifiers             |
   |                             ....                              |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

   version (V): 2 bits
   padding (P): 1 bit
   extension (X): 1 bit
   CSRC count (CC): 4 bits
   marker (M): 1 bit
   payload type (PT): 7 bits
   sequence number: 16 bits
   timestamp: 32 bits
   SSRC: 32 bits
   CSRC list: 0 to 15 items, 32 bits each

  Ref: https://tools.ietf.org/html/rfc3550#section-5.1
*/
type RTPHeader struct {
	Version        uint8
	Padding        bool
	Extend         bool
	CSRCCount      uint8
	Marker         bool
	PayloadType    uint8
	SequenceNumber uint16
	Timestamp      uint32
	SSRC           uint32
	CSRC           []uint32
	size           int
}

func (this *RTPHeader) Marshal() ([]byte, error) {
	var buf []byte
	if this.Version != 2 {
		return buf, errors.New("Version is invalid")
	}

	if this.CSRCCount > 15 {
		return buf, errors.New("CSRC count is invalid")
	}

	if int(this.CSRCCount) != len(this.CSRC) {
		return buf, errors.New("CSRC slice length can not be matched with CSRC count")
	}

	if this.PayloadType > 127 {
		return buf, errors.New("Payload type is invalid")
	}

	buf = make([]byte, 12+4*this.CSRCCount)
	buf[0] = this.Version<<6 | this.CSRCCount
	if this.Padding {
		buf[0] |= 1 << 5
	}
	if this.Extend {
		buf[0] |= 1 << 4
	}

	buf[1] = this.PayloadType
	if this.Marker {
		buf[1] |= 1 << 7
	}

	binary.BigEndian.PutUint16(buf[2:4], this.SequenceNumber)
	binary.BigEndian.PutUint32(buf[4:8], this.Timestamp)
	binary.BigEndian.PutUint32(buf[8:12], this.SSRC)

	for i := 0; i < int(this.CSRCCount); i++ {
		binary.BigEndian.PutUint32(buf[12+4*i:12+4*(i+1)], this.CSRC[i])
	}
	this.size = 12 + 4*int(this.CSRCCount)
	return buf, nil
}

func (this *RTPHeader) Unmarshal(buf []byte) error {
	if len(buf) < 12 {
		return errors.New("Data size can not be less than 12 byte")
	}

	this.Version = buf[0] >> 6
	if buf[0]>>5&0x01 == 1 {
		this.Padding = true
	} else {
		this.Padding = false
	}
	if buf[0]>>4&0x01 == 1 {
		this.Extend = true
	} else {
		this.Extend = false
	}
	this.CSRCCount = buf[0] & 0x0F
	if len(buf) < int(12+this.CSRCCount*4) {
		return errors.New("Data size can not be less than 12+CSRCCount*4 byte")
	}
	if buf[1]>>7 == 1 {
		this.Marker = true
	} else {
		this.Marker = false
	}
	this.PayloadType = buf[1] & 0x7F
	this.SequenceNumber = binary.BigEndian.Uint16(buf[2:4])
	this.Timestamp = binary.BigEndian.Uint32(buf[4:8])
	this.SSRC = binary.BigEndian.Uint32(buf[8:12])
	this.CSRC = []uint32{}
	if this.CSRCCount > 0 {
		this.CSRC = make([]uint32, this.CSRCCount)
		for i := 0; i < int(this.CSRCCount); i++ {
			this.CSRC[i] = binary.BigEndian.Uint32(buf[12+4*i : 12+4*(i+1)])
		}
	}

	this.size = 12 + 4*int(this.CSRCCount)
	return nil
}

type RTPPacket struct {
	Header       RTPHeader
	Payload      []byte
	PaddingCount uint8
	PaddingData  []byte
}

func (this *RTPPacket) Marshal() ([]byte, error) {
	var buf []byte
	header_buf, err := this.Header.Marshal()
	if err != nil {
		return buf, err
	}
	buf = append(buf, header_buf...)
	buf = append(buf, this.Payload...)
	if int(this.PaddingCount) != len(this.PaddingData)+1 {
		return buf, errors.New("Padding count error in RTP packet marshal")
	}
	if this.PaddingCount == 1 {
		buf = append(buf, 0x01)
	}
	if this.PaddingCount > 1 {
		buf = append(buf, this.PaddingData...)
		buf = append(buf, byte(this.PaddingCount))
	}
	return buf, nil
}

func (this *RTPPacket) Unmarshal(buf []byte) error {
	err := this.Header.Unmarshal(buf)
	if err != nil {
		return err
	}
	this.PaddingData = []byte{}
	this.Payload = []byte{}
	this.PaddingCount = 0
	if this.Header.Padding {
		this.PaddingCount = uint8(buf[len(buf)-1])
		if int(this.PaddingCount) > 1 {
			this.PaddingData = buf[len(buf)-int(this.PaddingCount) : len(buf)-1]
		}
		this.Payload = buf[this.Header.size : len(buf)-int(this.PaddingCount)]
	} else {
		this.Payload = buf[this.Header.size:]
	}
	return nil
}
