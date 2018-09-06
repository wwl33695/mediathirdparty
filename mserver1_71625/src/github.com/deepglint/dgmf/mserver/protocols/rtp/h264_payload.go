package rtp

import (
	"errors"
)

/*
5.3.  NAL Unit Header Usage

   The structure and semantics of the NAL unit header were introduced in
   Section 1.3.  For convenience, the format of the NAL unit header is
   reprinted below:

      +---------------+
      |0|1|2|3|4|5|6|7|
      +-+-+-+-+-+-+-+-+
      |F|NRI|  Type   |
      +---------------+

	F:   1 bit
         forbidden_zero_bit.  A value of 0 indicates that the NAL unit
         type octet and payload should not contain bit errors or other
         syntax violations.  A value of 1 indicates that the NAL unit
         type octet and payload may contain bit errors or other syntax
         violations.

         MANEs SHOULD set the F bit to indicate detected bit errors in
         the NAL unit.  The H.264 specification requires that the F bit
         be equal to 0.  When the F bit is set, the decoder is advised
         that bit errors or any other syntax violations may be present
         in the payload or in the NAL unit type octet.  The simplest
         decoder reaction to a NAL unit in which the F bit is equal to 1
         is to discard such a NAL unit and to conceal the lost data in
         the discarded NAL unit.

   NRI:  2 bits
         nal_ref_idc.  The semantics of value 00 and a non-zero value
         remain unchanged from the H.264 specification.  In other words,
         a value of 00 indicates that the content of the NAL unit is not
         used to reconstruct reference pictures for inter picture
         prediction.  Such NAL units can be discarded without risking
         the integrity of the reference pictures.  Values greater than
         00 indicate that the decoding of the NAL unit is required to
         maintain the integrity of the reference pictures.

	Type:5 bits
         Table 3.  Summary of allowed NAL unit types for each packetization
                mode (yes = allowed, no = disallowed, ig = ignore)

	      Payload Packet    Single NAL    Non-Interleaved    Interleaved
	      Type    Type      Unit Mode           Mode             Mode
	      -------------------------------------------------------------
	      0      reserved      ig               ig               ig
	      1-23   NAL unit     yes              yes               no
	      24     STAP-A        no              yes               no
	      25     STAP-B        no               no              yes
	      26     MTAP16        no               no              yes
	      27     MTAP24        no               no              yes
	      28     FU-A          no              yes              yes
	      29     FU-B          no               no              yes
	      30-31  reserved      ig               ig               ig

	      ref: https://tools.ietf.org/html/rfc6184#section-5.3
*/
type NALUHeader struct {
	NRI  uint8
	Type uint8
}

func (this *NALUHeader) Marshal() (byte, error) {
	var buf byte
	if this.NRI > 3 {
		return buf, errors.New("NRI can not be larger than 3 in NALU header marshal")
	}
	if this.Type > 31 {
		return buf, errors.New("Type can not be larger than 31 in NALU header marshal")
	}
	buf = 0x00 | this.NRI<<5 | this.Type

	return buf, nil
}

func (this *NALUHeader) Unmarshal(buf byte) error {
	F := buf >> 7
	if F != 0 {
		return errors.New("NALU header forbidden_zero_bit != 0 in NALU header unmarshal")
	}

	this.NRI = uint8(buf >> 5)
	this.Type = uint8(0x1F & buf)
	return nil
}

/*
   The FU header has the following format:

      +---------------+
      |0|1|2|3|4|5|6|7|
      +-+-+-+-+-+-+-+-+
      |S|E|R|  Type   |
      +---------------+

   S:     1 bit
          When set to one, the Start bit indicates the start of a
          fragmented NAL unit.  When the following FU payload is not the
          start of a fragmented NAL unit payload, the Start bit is set
          to zero.

   E:     1 bit
          When set to one, the End bit indicates the end of a fragmented
          NAL unit, i.e., the last byte of the payload is also the last
          byte of the fragmented NAL unit.  When the following FU
          payload is not the last fragment of a fragmented NAL unit, the
          End bit is set to zero.

   R:     1 bit
          The Reserved bit MUST be equal to 0 and MUST be ignored by the
          receiver.

   Type:  5 bits
          The NAL unit payload type as defined in Table 7-1 of [1].

   ref: https://tools.ietf.org/html/rfc6184#section-5.8
*/
type FUHeader struct {
	S    bool
	E    bool
	R    bool
	Type uint8
}

func (this *FUHeader) Marshal() byte {
	if this.Type > 31 {
		this.Type = 0
	}
	var buf byte
	buf = 0x00
	if this.S {
		buf |= 1 << 7
	}
	if this.E {
		buf |= 1 << 6
	}
	if this.R {
		buf |= 1 << 5
	}
	buf |= byte(this.Type)
	return buf
}

func (this *FUHeader) Unmarshal(buf byte) {
	if int(buf>>7) == 1 {
		this.S = true
	} else {
		this.S = false
	}
	if int((buf&0x7F)>>6) == 1 {
		this.E = true
	} else {
		this.E = false
	}
	if int((buf&0x3F)>>5) == 1 {
		this.R = true
	} else {
		this.R = false
	}
	this.Type = buf & 0x1F
}

type RTPNALUPacket struct {
	RTPNALUHeader NALUHeader
	RTPFUHeader   FUHeader
	Payload       []byte
}

func (this *RTPNALUPacket) Marshal() ([]byte, error) {
	var buf []byte
	if int(this.RTPNALUHeader.Type) < 1 || (int(this.RTPNALUHeader.Type) > 23 && int(this.RTPNALUHeader.Type) != 28) {
		return buf, errors.New("Unsupport RTP NALU type")
	}
	if int(this.RTPNALUHeader.Type) >= 1 && int(this.RTPNALUHeader.Type) <= 23 {
		naluHeader, err := this.RTPNALUHeader.Marshal()
		if err != nil {
			return buf, err
		}
		buf = append(buf, naluHeader)
	}
	if int(this.RTPNALUHeader.Type) == 28 {
		naluHeader, err := this.RTPNALUHeader.Marshal()
		if err != nil {
			return buf, err
		}
		buf = append(buf, naluHeader)
		buf = append(buf, this.RTPFUHeader.Marshal())
	}
	buf = append(buf, this.Payload...)
	return buf, nil
}

func (this *RTPNALUPacket) Unmarshal(buf []byte) error {
	if len(buf) < 1 {
		return errors.New("buf size can not be less than 1")
	}
	this.RTPNALUHeader.Unmarshal(buf[0])
	if int(this.RTPNALUHeader.Type) < 1 || (int(this.RTPNALUHeader.Type) > 23 && int(this.RTPNALUHeader.Type) != 28) {
		return errors.New("Unsupport RTP NALU type")
	}
	if int(this.RTPNALUHeader.Type) >= 1 && int(this.RTPNALUHeader.Type) <= 23 {
		this.Payload = buf[1:]
	}
	if int(this.RTPNALUHeader.Type) == 28 {
		if len(buf) < 2 {
			return errors.New("buf size can not be less than 2")
		}
		this.RTPFUHeader.Unmarshal(buf[1])
		this.Payload = buf[2:]
	}
	return nil
}

type H264NALUPacket struct {
	H264NALUHeader NALUHeader
	Payload        []byte
}

func (this *H264NALUPacket) Init(rtpNALU RTPNALUPacket) error {
	if int(rtpNALU.RTPNALUHeader.Type) < 1 || (int(rtpNALU.RTPNALUHeader.Type) > 23 && int(rtpNALU.RTPNALUHeader.Type) != 28) {
		return errors.New("Unsupport RTP NALU type")
	}
	if int(rtpNALU.RTPNALUHeader.Type) >= 1 && int(rtpNALU.RTPNALUHeader.Type) <= 23 {
		this.H264NALUHeader = NALUHeader{
			NRI:  rtpNALU.RTPNALUHeader.NRI,
			Type: rtpNALU.RTPNALUHeader.Type,
		}
	}
	if int(rtpNALU.RTPNALUHeader.Type) == 28 {
		this.H264NALUHeader = NALUHeader{
			NRI:  rtpNALU.RTPNALUHeader.NRI,
			Type: rtpNALU.RTPFUHeader.Type,
		}
	}
	this.Payload = rtpNALU.Payload

	return nil
}

func (this *H264NALUPacket) Add(rtpNALU RTPNALUPacket) {
	this.Payload = append(this.Payload, rtpNALU.Payload...)
}

func (this *H264NALUPacket) Marshal() ([]byte, error) {
	var buf []byte
	buf = append(buf, 0x00)
	buf = append(buf, 0x00)
	buf = append(buf, 0x00)
	buf = append(buf, 0x01)

	naluHeader, err := this.H264NALUHeader.Marshal()
	if err != nil {
		return buf, err
	}
	buf = append(buf, naluHeader)
	buf = append(buf, this.Payload...)
	return buf, nil
}
