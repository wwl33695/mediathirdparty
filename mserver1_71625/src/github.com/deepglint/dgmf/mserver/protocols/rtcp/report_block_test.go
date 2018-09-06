package rtcp

import (
	"testing"
)

func TestReportBlocMarshal(test *testing.T) {
	block := &ReportBlock{
		SSRC:               0x1234,
		FractionLost:       0x5,
		CumulativeLost:     0x12,
		EHSNR:              0xAA,
		InterarrivalJitter: 0xBB,
		LSR:                0xCC,
		DLSR:               0xDD,
	}
	buf := block.Marshal()
	if buf[0] != 0x00 {
		test.Error()
	}
	if buf[1] != 0x00 {
		test.Error()
	}
	if buf[2] != 0x12 {
		test.Error()
	}
	if buf[3] != 0x34 {
		test.Error()
	}
	if buf[4] != 0x05 {
		test.Error()
	}
	if buf[5] != 0x00 {
		test.Error()
	}
	if buf[6] != 0x00 {
		test.Error()
	}
	if buf[7] != 0x12 {
		test.Error()
	}
	if buf[8] != 0x00 {
		test.Error()
	}
	if buf[9] != 0x00 {
		test.Error()
	}
	if buf[10] != 0x00 {
		test.Error()
	}
	if buf[11] != 0xaa {
		test.Error()
	}
	if buf[12] != 0x00 {
		test.Error()
	}
	if buf[13] != 0x00 {
		test.Error()
	}
	if buf[14] != 0x00 {
		test.Error()
	}
	if buf[15] != 0xbb {
		test.Error()
	}
	if buf[16] != 0x00 {
		test.Error()
	}
	if buf[17] != 0x00 {
		test.Error()
	}
	if buf[18] != 0x00 {
		test.Error()
	}
	if buf[19] != 0xcc {
		test.Error()
	}
	if buf[20] != 0x00 {
		test.Error()
	}
	if buf[21] != 0x00 {
		test.Error()
	}
	if buf[22] != 0x00 {
		test.Error()
	}
	if buf[23] != 0xdd {
		test.Error()
	}
}

func TestReportBlockUnmarshal(test *testing.T) {
	buf := []byte{0x00, 0x00, 0x12, 0x34, 0x05, 0x00, 0x00, 0x12, 0x00, 0x00, 0x00, 0xaa, 0x00, 0x00, 0x00, 0xbb, 0x00, 0x00, 0x00, 0xcc, 0x00, 0x00, 0x00, 0xdd}
	block := &ReportBlock{}
	block.Unmarshal(buf)
	if block.SSRC != 0x1234 {
		test.Error()
	}
	if block.FractionLost != 0x5 {
		test.Error()
	}
	if block.CumulativeLost != 0x12 {
		test.Error()
	}
	if block.EHSNR != 0xAA {
		test.Error()
	}
	if block.InterarrivalJitter != 0xBB {
		test.Error()
	}
	if block.LSR != 0xCC {
		test.Error()
	}
	if block.DLSR != 0xDD {
		test.Error()
	}
}
