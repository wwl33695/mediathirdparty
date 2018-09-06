package rtcp

import (
	"encoding/binary"
	"errors"
)

/*
6.4.1 SR: Sender Report RTCP Packet

        0                   1                   2                   3
        0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
       +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
header |V=2|P|    RC   |   PT=SR=200   |             length            |
       +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
       |                         SSRC of sender                        |
       +=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+
sender |              NTP timestamp, most significant word             |
info   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
       |             NTP timestamp, least significant word             |
       +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
       |                         RTP timestamp                         |
       +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
       |                     sender's packet count                     |
       +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
       |                      sender's octet count                     |
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

   version (V): 2 bits
   padding (P): 1 bit
   reception report count (RC): 5 bits
   packet type (PT): 8 bits
   length: 16 bits
   SSRC: 32 bits
   NTP timestamp: 64 bits
   RTP timestamp: 32 bits
   sender's packet count: 32 bits
   sender's octet count: 32 bits
   SSRC_n (source identifier): 32 bits
   fraction lost: 8 bits
   cumulative number of packets lost: 24 bits
   extended highest sequence number received: 32 bits
   interarrival jitter: 32 bits
   last SR timestamp (LSR): 32 bits
   delay since last SR (DLSR): 32 bits

    ref: https://www.ietf.org/rfc/rfc3550.txt
*/
type SenderReport struct {
	Version      uint8
	Padding      bool
	ReportCount  uint8
	PacketType   uint8
	Length       uint16
	SSRC         uint32
	NTPTimestamp uint64
	RTPTimestamp uint32
	PacketCount  uint32
	OctetCount   uint32
	Blocks       []ReportBlock
}

func (this *SenderReport) Marshal() ([]byte, error) {
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
	if this.PacketType != 200 {
		return buf, errors.New("PacketType is invalid")
	}
	buf = make([]byte, 28)
	buf[0] = 0x80
	if this.Padding == true {
		buf[0] |= 0x20
	}
	buf[0] |= this.ReportCount
	buf[1] = this.PacketType
	this.Length = uint16((28+len(this.Blocks)*24)/4 - 1)
	binary.BigEndian.PutUint16(buf[2:4], this.Length)
	binary.BigEndian.PutUint32(buf[4:8], this.SSRC)
	binary.BigEndian.PutUint64(buf[8:16], this.NTPTimestamp)
	binary.BigEndian.PutUint32(buf[16:20], this.RTPTimestamp)
	binary.BigEndian.PutUint32(buf[20:24], this.PacketCount)
	binary.BigEndian.PutUint32(buf[24:28], this.OctetCount)
	for i := 0; i < len(this.Blocks); i++ {
		buf = append(buf, this.Blocks[i].Marshal()...)
	}

	return buf, nil
}

func (this *SenderReport) Unmarshal(buf []byte) error {
	if len(buf) < 28 {
		return errors.New("buffer size error it can not be less than 28")
	}
	this.Version = buf[0] >> 6
	if (buf[0]>>5)&0x01 == 1 {
		this.Padding = true
	} else {
		this.Padding = false
	}
	this.ReportCount = buf[0] & 0x1F
	if len(buf) < int(28+this.ReportCount*24) {
		return errors.New("ReportCount is invalid")
	}
	this.PacketType = buf[1]
	this.Length = binary.BigEndian.Uint16(buf[2:4])
	this.SSRC = binary.BigEndian.Uint32(buf[4:8])
	this.NTPTimestamp = binary.BigEndian.Uint64(buf[8:16])
	this.RTPTimestamp = binary.BigEndian.Uint32(buf[16:20])
	this.PacketCount = binary.BigEndian.Uint32(buf[20:24])
	this.OctetCount = binary.BigEndian.Uint32(buf[24:28])
	this.Blocks = make([]ReportBlock, this.ReportCount)
	for i := 0; i < int(this.ReportCount); i++ {
		this.Blocks[i].Unmarshal(buf[28+i*24 : 28+(i+1)*24])
	}

	return nil
}
