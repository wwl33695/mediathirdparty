package rtp

import (
	"strings"
	"testing"
)

func TestNALUHeaderUnmarshal(test *testing.T) {
	var header NALUHeader
	var buf byte
	var err error

	buf = 0x7C
	err = header.Unmarshal(buf)
	if err != nil {
		test.Error()
	}
	if header.NRI != 3 || header.Type != 28 {
		test.Error()
	}

	buf = 0xFC
	err = header.Unmarshal(buf)
	if err == nil {
		test.Error()
	}
	if !strings.EqualFold(err.Error(), "NALU header forbidden_zero_bit != 0 in NALU header unmarshal") {
		test.Error()
	}
}

func TestNALUHeaderMarshal(test *testing.T) {
	var header NALUHeader
	var buf byte
	var err error

	header.NRI = 3
	header.Type = 28
	buf, err = header.Marshal()
	if err != nil {
		test.Error()
	}
	if buf != 0x7C {
		test.Error()
	}

	header.NRI = 4
	header.Type = 28
	buf, err = header.Marshal()
	if err == nil {
		test.Error()
	}
	if !strings.EqualFold(err.Error(), "NRI can not be larger than 3 in NALU header marshal") {
		test.Error()
	}

	header.NRI = 3
	header.Type = 40
	buf, err = header.Marshal()
	if err == nil {
		test.Error()
	}
	if !strings.EqualFold(err.Error(), "Type can not be larger than 31 in NALU header marshal") {
		test.Error()
	}
}

func TestFUHeaderMarshal(test *testing.T) {
	var fuHeader FUHeader

	fuHeader = FUHeader{
		S:    true,
		E:    false,
		R:    false,
		Type: 0x01,
	}
	if fuHeader.Marshal() != 0x81 {
		test.Error()
	}

	fuHeader = FUHeader{
		S:    false,
		E:    true,
		R:    false,
		Type: 0x01,
	}
	if fuHeader.Marshal() != 0x41 {
		test.Error()
	}

	fuHeader = FUHeader{
		S:    false,
		E:    false,
		R:    true,
		Type: 0x01,
	}
	if fuHeader.Marshal() != 0x21 {
		test.Error()
	}

	fuHeader = FUHeader{
		S:    false,
		E:    false,
		R:    false,
		Type: 0x01,
	}
	if fuHeader.Marshal() != 0x01 {
		test.Error()
	}

	fuHeader = FUHeader{
		S:    false,
		E:    false,
		R:    false,
		Type: 33,
	}
	if fuHeader.Marshal() != 0x00 {
		test.Error()
	}
}

func TestFUHeaderUnmarshal(test *testing.T) {
	fuHeader := FUHeader{}

	fuHeader.Unmarshal(0x81)
	if !fuHeader.S || fuHeader.E || fuHeader.R || fuHeader.Type != 0x01 {
		test.Error()
	}

	fuHeader.Unmarshal(0x41)
	if fuHeader.S || !fuHeader.E || fuHeader.R || fuHeader.Type != 0x01 {
		test.Error()
	}

	fuHeader.Unmarshal(0x21)
	if fuHeader.S || fuHeader.E || !fuHeader.R || fuHeader.Type != 0x01 {
		test.Error()
	}

	fuHeader.Unmarshal(0x01)
	if fuHeader.S || fuHeader.E || fuHeader.R || fuHeader.Type != 0x01 {
		test.Error()
	}
}

func TestRTPNALUPacketMarshal(test *testing.T) {
	var packet RTPNALUPacket
	var err error
	var data []byte

	packet = RTPNALUPacket{
		RTPNALUHeader: NALUHeader{
			NRI:  3,
			Type: 1,
		},
		Payload: []byte{0x11, 0x12, 0x13, 0x14},
	}
	data, err = packet.Marshal()
	if err != nil {
		test.Error()
	}
	if data[0] != 0x61 {
		test.Error()
	}
	if data[1] != 0x11 {
		test.Error()
	}
	if data[2] != 0x12 {
		test.Error()
	}
	if data[3] != 0x13 {
		test.Error()
	}
	if data[4] != 0x14 {
		test.Error()
	}

	packet = RTPNALUPacket{
		RTPNALUHeader: NALUHeader{
			NRI:  3,
			Type: 0,
		},
		Payload: []byte{0x11, 0x12, 0x13, 0x14},
	}
	data, err = packet.Marshal()
	if err == nil || !strings.EqualFold("Unsupport RTP NALU type", err.Error()) {
		test.Error()
	}

	packet = RTPNALUPacket{
		RTPNALUHeader: NALUHeader{
			NRI:  3,
			Type: 28,
		},
		RTPFUHeader: FUHeader{
			S:    true,
			E:    false,
			R:    false,
			Type: 0x01,
		},
		Payload: []byte{0x11, 0x12, 0x13, 0x14},
	}
	data, err = packet.Marshal()
	if err != nil {
		test.Error()
	}
	if data[0] != 0x7C {
		test.Error()
	}
	if data[1] != 0x81 {
		test.Error()
	}
	if data[2] != 0x11 {
		test.Error()
	}
	if data[3] != 0x12 {
		test.Error()
	}
	if data[4] != 0x13 {
		test.Error()
	}
	if data[5] != 0x14 {
		test.Error()
	}
}

func TestRTPNALUPacketUnmarshal(test *testing.T) {
	var packet RTPNALUPacket
	var err error
	var data []byte

	err = packet.Unmarshal(data)
	if err == nil || !strings.EqualFold(err.Error(), "buf size can not be less than 1") {
		test.Error()
	}

	data = []byte{0x60, 0x11, 0x12, 0x13, 0x14}
	err = packet.Unmarshal(data)
	if err == nil || !strings.EqualFold(err.Error(), "Unsupport RTP NALU type") {
		test.Error()
	}

	data = []byte{0x61, 0x11, 0x12, 0x13, 0x14}
	err = packet.Unmarshal(data)
	if err != nil {
		test.Error()
	}
	if packet.RTPNALUHeader.NRI != 3 {
		test.Error()
	}
	if packet.RTPNALUHeader.Type != 1 {
		test.Error()
	}
	if len(packet.Payload) != 4 {
		test.Error()
	}
	if packet.Payload[0] != 0x11 {
		test.Error()
	}
	if packet.Payload[1] != 0x12 {
		test.Error()
	}
	if packet.Payload[2] != 0x13 {
		test.Error()
	}
	if packet.Payload[3] != 0x14 {
		test.Error()
	}

	data = []byte{0x7C, 0x81, 0x11, 0x12, 0x13, 0x14}
	err = packet.Unmarshal(data)
	if err != nil {
		test.Error()
	}
	if packet.RTPNALUHeader.NRI != 3 {
		test.Error()
	}
	if packet.RTPNALUHeader.Type != 28 {
		test.Error()
	}
	if packet.RTPFUHeader.S != true {
		test.Error()
	}
	if packet.RTPFUHeader.E != false {
		test.Error()
	}
	if packet.RTPFUHeader.R != false {
		test.Error()
	}
	if packet.RTPFUHeader.Type != 0x01 {
		test.Error()
	}
	if packet.Payload[0] != 0x11 {
		test.Error()
	}
	if packet.Payload[1] != 0x12 {
		test.Error()
	}
	if packet.Payload[2] != 0x13 {
		test.Error()
	}
	if packet.Payload[3] != 0x14 {
		test.Error()
	}
}

func TestH264NALUPacketInit(test *testing.T) {
	var packet H264NALUPacket
	var rtp RTPNALUPacket
	var err error

	rtp = RTPNALUPacket{
		RTPNALUHeader: NALUHeader{
			NRI:  3,
			Type: 28,
		},
		RTPFUHeader: FUHeader{
			S:    true,
			E:    false,
			R:    false,
			Type: 0x01,
		},
		Payload: []byte{0x11, 0x12, 0x13, 0x14},
	}
	err = packet.Init(rtp)
	if err != nil {
		test.Error()
	}

	rtp = RTPNALUPacket{
		RTPNALUHeader: NALUHeader{
			NRI:  3,
			Type: 1,
		},
		RTPFUHeader: FUHeader{
			S:    true,
			E:    false,
			R:    false,
			Type: 0x01,
		},
		Payload: []byte{0x11, 0x12, 0x13, 0x14},
	}
	err = packet.Init(rtp)
	if err != nil {
		test.Error()
	}

	rtp = RTPNALUPacket{
		RTPNALUHeader: NALUHeader{
			NRI:  3,
			Type: 0,
		},
		RTPFUHeader: FUHeader{
			S:    true,
			E:    false,
			R:    false,
			Type: 0x01,
		},
		Payload: []byte{0x11, 0x12, 0x13, 0x14},
	}
	err = packet.Init(rtp)
	if err == nil || !strings.EqualFold("Unsupport RTP NALU type", err.Error()) {
		test.Error()
	}
}

func TestH264NALUPacketMarshal(test *testing.T) {
	var packet H264NALUPacket
	var rtp RTPNALUPacket
	var err error
	var data []byte

	rtp = RTPNALUPacket{
		RTPNALUHeader: NALUHeader{
			NRI:  3,
			Type: 28,
		},
		RTPFUHeader: FUHeader{
			S:    true,
			E:    false,
			R:    false,
			Type: 0x01,
		},
		Payload: []byte{0x11, 0x12, 0x13, 0x14},
	}
	err = packet.Init(rtp)
	if err != nil {
		test.Error()
	}

	rtp = RTPNALUPacket{
		RTPNALUHeader: NALUHeader{
			NRI:  3,
			Type: 28,
		},
		RTPFUHeader: FUHeader{
			S:    false,
			E:    false,
			R:    false,
			Type: 0x01,
		},
		Payload: []byte{0x11, 0x12, 0x13, 0x14},
	}
	packet.Add(rtp)
	data, err = packet.Marshal()
	if err != nil {
		test.Error()
	}
	if data[0] != 0x00 {
		test.Error()
	}
	if data[1] != 0x00 {
		test.Error()
	}
	if data[2] != 0x00 {
		test.Error()
	}
	if data[3] != 0x01 {
		test.Error()
	}
	if data[4] != 0x61 {
		test.Error()
	}
	if data[5] != 0x11 {
		test.Error()
	}
	if data[6] != 0x12 {
		test.Error()
	}
	if data[7] != 0x13 {
		test.Error()
	}
	if data[8] != 0x14 {
		test.Error()
	}
	if data[9] != 0x11 {
		test.Error()
	}
	if data[10] != 0x12 {
		test.Error()
	}
	if data[11] != 0x13 {
		test.Error()
	}
	if data[12] != 0x14 {
		test.Error()
	}
}
