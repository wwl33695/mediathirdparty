package rtcp

import (
	"encoding/binary"
	"errors"
)

/*
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
*/
type ReportBlock struct {
	SSRC               uint32
	FractionLost       uint8
	CumulativeLost     uint32
	EHSNR              uint32
	InterarrivalJitter uint32
	LSR                uint32
	DLSR               uint32
}

func (this *ReportBlock) Marshal() []byte {
	buf := make([]byte, 24)
	binary.BigEndian.PutUint32(buf[0:4], this.SSRC)
	buf[4] = this.FractionLost
	tmp := make([]byte, 4)
	binary.BigEndian.PutUint32(tmp, this.CumulativeLost)
	buf[5] = tmp[1]
	buf[6] = tmp[2]
	buf[7] = tmp[3]
	binary.BigEndian.PutUint32(buf[8:12], this.EHSNR)
	binary.BigEndian.PutUint32(buf[12:16], this.InterarrivalJitter)
	binary.BigEndian.PutUint32(buf[16:20], this.LSR)
	binary.BigEndian.PutUint32(buf[20:24], this.DLSR)

	return buf
}

func (this *ReportBlock) Unmarshal(buf []byte) error {
	if len(buf) != 24 {
		return errors.New("buffer size error, it should be 24")
	}

	this.SSRC = binary.BigEndian.Uint32(buf[0:4])
	this.FractionLost = buf[4]
	tmp := make([]byte, 4)
	tmp[0] = 0
	tmp[1] = buf[5]
	tmp[2] = buf[6]
	tmp[3] = buf[7]
	this.CumulativeLost = binary.BigEndian.Uint32(tmp)
	this.EHSNR = binary.BigEndian.Uint32(buf[8:12])
	this.InterarrivalJitter = binary.BigEndian.Uint32(buf[12:16])
	this.LSR = binary.BigEndian.Uint32(buf[16:20])
	this.DLSR = binary.BigEndian.Uint32(buf[20:24])
	return nil
}
