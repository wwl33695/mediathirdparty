package rtcp

import (
	"encoding/binary"
	"errors"
)

/*
6.4.2 RR: Receiver Report RTCP Packet

        0                   1                   2                   3
        0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
       +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
header |V=2|P|    RC   |   PT=RR=201   |             length            |
       +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
       |                     SSRC of packet sender                     |
       +=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+
report |                 SSRC_1 (SSRC of first source)                 |
block  +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
  1    | fraction lost |       cumulative number of packets lost       |
       +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
       |           extended highest sequence number received           |
       +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
       |                      interarrival jitter                      |
       +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
       |                         last SR (LSR)                         |
       +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
       |                   delay since last SR (DLSR)                  |
       +=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+
report |                 SSRC_2 (SSRC of second source)                |
block  +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
  2    :                               ...                             :
       +=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+
       |                  profile-specific extensions                  |
       +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
ref: https://www.ietf.org/rfc/rfc3550.txt
*/
type ReceiverReport struct {
	Version     uint8
	Padding     bool
	ReportCount uint8
	PacketType  uint8
	Length      uint16
	SSRC        uint32
	Blocks      []ReportBlock
}

func (this *ReceiverReport) Marshal() ([]byte, error) {
	var buf []byte
	if this.Version != 2 {
		return buf, errors.New("Version is invalid")
	}
	if int(this.ReportCount) != len(this.Blocks) {
		return buf, errors.New("ReportCount is invalid")
	}
	if this.ReportCount > 31 {
		return buf, errors.New("ReportCount is invalid")
	}
	if this.PacketType != 201 {
		return buf, errors.New("PacketType is invalid")
	}

	buf = make([]byte, 8)
	buf[0] = 0x80
	if this.Padding == true {
		buf[0] |= 0x20
	}
	buf[0] |= this.ReportCount
	buf[1] = this.PacketType
	this.Length = uint16((8+len(this.Blocks)*24)/4 - 1)
	binary.BigEndian.PutUint16(buf[2:4], this.Length)
	binary.BigEndian.PutUint32(buf[4:8], this.SSRC)
	for i := 0; i < len(this.Blocks); i++ {
		buf = append(buf, this.Blocks[i].Marshal()...)
	}
	return buf, nil
}

func (this *ReceiverReport) Unmarshal(buf []byte) error {
	if len(buf) < 8 {
		return errors.New("buffer size error it can not be less than 8")
	}
	this.Version = buf[0] >> 6
	if (buf[0]>>5)&0x01 == 1 {
		this.Padding = true
	} else {
		this.Padding = false
	}
	this.ReportCount = buf[0] & 0x1F
	if len(buf) < int(8+this.ReportCount*24) {
		return errors.New("ReportCount is invalid")
	}
	this.PacketType = buf[1]
	this.Length = binary.BigEndian.Uint16(buf[2:4])
	this.SSRC = binary.BigEndian.Uint32(buf[4:8])
	this.Blocks = make([]ReportBlock, this.ReportCount)
	for i := 0; i < int(this.ReportCount); i++ {
		this.Blocks[i].Unmarshal(buf[8+i*24 : 8+(i+1)*24])
	}

	return nil
}
