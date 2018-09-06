package rtp

import (
	// "fmt"
	"strings"
	"testing"
)

func TestRTPHeaderUnmarshal(test *testing.T) {
	var header RTPHeader
	var buf []byte
	var err error

	buf = []byte{0x81, 0x60, 0x10, 0x81, 0xb6, 0x05, 0x6c, 0x44, 0x69, 0x60, 0x73, 0x9a, 0x69, 0x60, 0x73, 0x9b}
	err = header.Unmarshal(buf)
	if err != nil {
		test.Error()
	}
	if header.Version != 2 {
		test.Error()
	}
	if header.Padding != false {
		test.Error()
	}
	if header.Extend != false {
		test.Error()
	}
	if header.CSRCCount != 1 {
		test.Error()
	}
	if header.SequenceNumber != 4225 {
		test.Error()
	}
	if header.Timestamp != 3053808708 {
		test.Error()
	}
	if header.SSRC != 1767928730 {
		test.Error()
	}
	if len(header.CSRC) != 1 {
		test.Error()
	}
	if header.PayloadType != 96 {
		test.Error()
	}
	if header.Marker != false {
		test.Error()
	}
	if header.CSRC[0] != 1767928731 {
		test.Error()
	}

	buf = []byte{0xb1, 0xe0, 0x10, 0x81, 0xb6, 0x05, 0x6c, 0x44, 0x69, 0x60, 0x73, 0x9a, 0x69, 0x60, 0x73, 0x9b}
	err = header.Unmarshal(buf)
	if err != nil {
		test.Error()
	}
	if header.Padding != true {
		test.Error()
	}
	if header.Extend != true {
		test.Error()
	}
	if header.Marker != true {
		test.Error()
	}

	buf = []byte{0x80, 0x60, 0x10, 0x81, 0xb6, 0x05, 0x6c, 0x44, 0x69, 0x60, 0x73}
	err = header.Unmarshal(buf)
	if err == nil || !strings.EqualFold(err.Error(), "Data size can not be less than 12 byte") {
		test.Error()
	}

	buf = []byte{0x81, 0x60, 0x10, 0x81, 0xb6, 0x05, 0x6c, 0x44, 0x69, 0x60, 0x73, 0x9a}
	err = header.Unmarshal(buf)
	if err == nil || !strings.EqualFold(err.Error(), "Data size can not be less than 12+CSRCCount*4 byte") {
		test.Error()
	}

	buf = []byte{0x80, 0x60, 0x10, 0x81, 0xb6, 0x05, 0x6c, 0x44, 0x69, 0x60, 0x73, 0x9a}
	err = header.Unmarshal(buf)
	if err != nil {
		test.Error()
	}
	if len(header.CSRC) != int(header.CSRCCount) || int(header.CSRCCount) != 0 || len(header.CSRC) != 0 {
		test.Error()
	}
}

func TestRTPHeaderMarshal(test *testing.T) {
	var header RTPHeader
	var buf []byte
	var err error

	header = RTPHeader{
		Version:        2,
		Padding:        false,
		Extend:         false,
		CSRCCount:      1,
		Marker:         false,
		PayloadType:    96,
		SequenceNumber: 4225,
		Timestamp:      3053808708,
		SSRC:           1767928730,
		CSRC: []uint32{
			1767928731,
		},
	}

	buf, err = header.Marshal()
	if err != nil {
		test.Error()
	}
	if err != nil || len(buf) != 16 {
		test.Error()
	}
	if buf[0] != 0x81 {
		test.Error()
	}
	if buf[1] != 0x60 {
		test.Error()
	}
	if buf[2] != 0x10 {
		test.Error()
	}
	if buf[3] != 0x81 {
		test.Error()
	}
	if buf[4] != 0xB6 {
		test.Error()
	}
	if buf[5] != 0x05 {
		test.Error()
	}
	if buf[6] != 0x6C {
		test.Error()
	}
	if buf[7] != 0x44 {
		test.Error()
	}
	if buf[8] != 0x69 {
		test.Error()
	}
	if buf[9] != 0x60 {
		test.Error()
	}
	if buf[10] != 0x73 {
		test.Error()
	}
	if buf[11] != 0x9A {
		test.Error()
	}
	if buf[12] != 0x69 {
		test.Error()
	}
	if buf[13] != 0x60 {
		test.Error()
	}
	if buf[14] != 0x73 {
		test.Error()
	}
	if buf[15] != 0x9B {
		test.Error()
	}

	header = RTPHeader{
		Version:        4,
		Padding:        false,
		Extend:         false,
		CSRCCount:      0,
		Marker:         false,
		PayloadType:    96,
		SequenceNumber: 4225,
		Timestamp:      3053808708,
		SSRC:           1767928730,
	}

	buf, err = header.Marshal()
	if err == nil || !strings.EqualFold(err.Error(), "Version is invalid") {
		test.Error()
	}

	header = RTPHeader{
		Version:        2,
		Padding:        false,
		Extend:         false,
		CSRCCount:      16,
		Marker:         false,
		PayloadType:    96,
		SequenceNumber: 4225,
		Timestamp:      3053808708,
		SSRC:           1767928730,
		CSRC: []uint32{
			1767928731, 1767928731, 1767928731, 1767928731, 1767928731, 1767928731,
			1767928731, 1767928731, 1767928731, 1767928731, 1767928731, 1767928731,
			1767928731, 1767928731, 1767928731, 1767928731,
		},
	}

	buf, err = header.Marshal()
	if err == nil || !strings.EqualFold(err.Error(), "CSRC count is invalid") {
		test.Error()
	}

	header = RTPHeader{
		Version:        2,
		Padding:        false,
		Extend:         false,
		CSRCCount:      2,
		Marker:         false,
		PayloadType:    96,
		SequenceNumber: 4225,
		Timestamp:      3053808708,
		SSRC:           1767928730,
		CSRC: []uint32{
			1767928731, 1767928731, 1767928731,
		},
	}

	buf, err = header.Marshal()
	if err == nil || !strings.EqualFold(err.Error(), "CSRC slice length can not be matched with CSRC count") {
		test.Error()
	}

	header = RTPHeader{
		Version:        2,
		Padding:        false,
		Extend:         false,
		CSRCCount:      1,
		Marker:         false,
		PayloadType:    200,
		SequenceNumber: 4225,
		Timestamp:      3053808708,
		SSRC:           1767928730,
		CSRC: []uint32{
			1767928731,
		},
	}

	buf, err = header.Marshal()
	if err == nil || !strings.EqualFold(err.Error(), "Payload type is invalid") {
		test.Error()
	}

	header = RTPHeader{
		Version:        2,
		Padding:        true,
		Extend:         true,
		CSRCCount:      1,
		Marker:         true,
		PayloadType:    96,
		SequenceNumber: 4225,
		Timestamp:      3053808708,
		SSRC:           1767928730,
		CSRC: []uint32{
			1767928731,
		},
	}

	buf, err = header.Marshal()
	if err != nil {
		test.Error()
	}
	if err != nil || len(buf) != 16 {
		test.Error()
	}
	if buf[0] != 0xB1 {
		test.Error()
	}
	if buf[1] != 0xE0 {
		test.Error()
	}
	if buf[2] != 0x10 {
		test.Error()
	}
	if buf[3] != 0x81 {
		test.Error()
	}
	if buf[4] != 0xB6 {
		test.Error()
	}
	if buf[5] != 0x05 {
		test.Error()
	}
	if buf[6] != 0x6C {
		test.Error()
	}
	if buf[7] != 0x44 {
		test.Error()
	}
	if buf[8] != 0x69 {
		test.Error()
	}
	if buf[9] != 0x60 {
		test.Error()
	}
	if buf[10] != 0x73 {
		test.Error()
	}
	if buf[11] != 0x9A {
		test.Error()
	}
	if buf[12] != 0x69 {
		test.Error()
	}
	if buf[13] != 0x60 {
		test.Error()
	}
	if buf[14] != 0x73 {
		test.Error()
	}
	if buf[15] != 0x9B {
		test.Error()
	}
}

func TestRTPPacketMarshal(test *testing.T) {
	var packet RTPPacket
	var err error
	var buf []byte

	packet.Header = RTPHeader{
		Version:        2,
		Padding:        true,
		Extend:         false,
		CSRCCount:      0,
		Marker:         false,
		PayloadType:    96,
		SequenceNumber: 32387,
		Timestamp:      905653278,
		SSRC:           60099170,
	}
	packet.Payload = []byte{0x67, 0x4D, 0x00, 0x32, 0x95}
	packet.PaddingCount = 1
	buf, err = packet.Marshal()
	if err != nil {
		test.Error()
	}
	if len(buf) != 18 {
		test.Error()
	}
	if buf[0] != 0xA0 || buf[1] != 0x60 || buf[2] != 0x7E || buf[3] != 0x83 || buf[4] != 0x35 || buf[5] != 0xFB || buf[6] != 0x2C ||
		buf[7] != 0x1E || buf[8] != 0x03 || buf[9] != 0x95 || buf[10] != 0x0A || buf[11] != 0x62 || buf[12] != 0x67 ||
		buf[13] != 0x4D || buf[14] != 0x00 || buf[15] != 0x32 || buf[16] != 0x95 || buf[17] != 0x01 {
		test.Error()
	}

	packet.Header = RTPHeader{
		Version:        2,
		Padding:        true,
		Extend:         false,
		CSRCCount:      0,
		Marker:         false,
		PayloadType:    96,
		SequenceNumber: 32387,
		Timestamp:      905653278,
		SSRC:           60099170,
	}
	packet.Payload = []byte{0x67, 0x4D, 0x00, 0x32, 0x95}
	packet.PaddingCount = 3
	packet.PaddingData = []byte{0x00, 0x00}
	buf, err = packet.Marshal()
	if err != nil {
		test.Error()
	}
	if len(buf) != 20 {
		test.Error()
	}
	if buf[0] != 0xA0 || buf[1] != 0x60 || buf[2] != 0x7E || buf[3] != 0x83 || buf[4] != 0x35 || buf[5] != 0xFB || buf[6] != 0x2C ||
		buf[7] != 0x1E || buf[8] != 0x03 || buf[9] != 0x95 || buf[10] != 0x0A || buf[11] != 0x62 || buf[12] != 0x67 ||
		buf[13] != 0x4D || buf[14] != 0x00 || buf[15] != 0x32 || buf[16] != 0x95 || buf[17] != 0x00 || buf[18] != 0x00 ||
		buf[19] != 0x03 {
		test.Error()
	}

	packet.Header = RTPHeader{
		Version:        2,
		Padding:        true,
		Extend:         false,
		CSRCCount:      0,
		Marker:         false,
		PayloadType:    96,
		SequenceNumber: 32387,
		Timestamp:      905653278,
		SSRC:           60099170,
	}
	packet.Payload = []byte{0x67, 0x4D, 0x00, 0x32, 0x95}
	packet.PaddingCount = 3
	packet.PaddingData = []byte{0x00}
	buf, err = packet.Marshal()
	if err == nil {
		test.Error()
	}
	if !strings.EqualFold(err.Error(), "Padding count error in RTP packet marshal") {
		test.Error()
	}
}

func TestRTPPacketUnmarshal(test *testing.T) {
	var packet RTPPacket
	var err error
	var buf []byte

	buf = []byte{0xA0, 0x60, 0x7E, 0x85, 0x35, 0xFB, 0x2C, 0x1E, 0x03, 0x95, 0x0A, 0x62, 0x06, 0xE5, 0x01, 0xBC, 0x80, 0x00, 0x00, 0x03}
	err = packet.Unmarshal(buf)
	if err != nil {
		test.Error()
	}
	if packet.PaddingCount != 3 {
		test.Error()
	}
	if len(packet.PaddingData) != 2 {
		test.Error()
	}
	if packet.PaddingData[0] != 0x00 || packet.PaddingData[1] != 0x00 {
		test.Error()
	}
	if len(packet.Payload) != 5 {
		test.Error()
	}
	if packet.Payload[0] != 0x06 || packet.Payload[1] != 0xE5 || packet.Payload[2] != 0x01 || packet.Payload[3] != 0xBC ||
		packet.Payload[4] != 0x80 {
		test.Error()
	}

	buf = []byte{0xA0, 0x60, 0x7E, 0x83, 0x35, 0xFB, 0x2C, 0x1E, 0x03, 0x95, 0x0A, 0x62, 0x67, 0x4D, 0x00, 0x32, 0x95, 0x01}
	err = packet.Unmarshal(buf)
	if err != nil {
		test.Error()
	}
	if packet.PaddingCount != 1 {
		test.Error()
	}
	if len(packet.PaddingData) != 0 {
		test.Error()
	}
	if len(packet.Payload) != 5 {
		test.Error()
	}
	if packet.Payload[0] != 0x67 || packet.Payload[1] != 0x4D || packet.Payload[2] != 0x00 || packet.Payload[3] != 0x32 ||
		packet.Payload[4] != 0x95 {
		test.Error()
	}

	buf = []byte{0x80, 0x60, 0x7E, 0x84, 0x35, 0xFB, 0x2C, 0x1E, 0x03, 0x95, 0x0A, 0x62, 0x68, 0xEE, 0x3C, 0x80}
	err = packet.Unmarshal(buf)
	if err != nil {
		test.Error()
	}
	if packet.PaddingCount != 0 {
		test.Error()
	}
	if len(packet.PaddingData) != 0 {
		test.Error()
	}
	if len(packet.Payload) != 4 {
		test.Error()
	}
	if packet.Payload[0] != 0x68 || packet.Payload[1] != 0xEE || packet.Payload[2] != 0x3C || packet.Payload[3] != 0x80 {
		test.Error()
	}
}
