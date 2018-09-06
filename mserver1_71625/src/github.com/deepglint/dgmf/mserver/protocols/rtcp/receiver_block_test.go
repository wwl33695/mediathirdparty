package rtcp

import (
	"testing"
)

func TestReceiverReportMarshal(test *testing.T) {
	report := &ReceiverReport{
		Version:     2,
		Padding:     false,
		ReportCount: 1,
		PacketType:  201,
		SSRC:        0x1234,
		Blocks: []ReportBlock{
			ReportBlock{
				SSRC:               0x1234,
				FractionLost:       0x5,
				CumulativeLost:     0x12,
				EHSNR:              0xAA,
				InterarrivalJitter: 0xBB,
				LSR:                0xCC,
				DLSR:               0xDD,
			},
		},
	}

	buf, err := report.Marshal()
	if err != nil {
		test.Error()
	}

	if buf[0] != 0x81 {
		test.Error()
	}
	if buf[1] != 0xc9 {
		test.Error()
	}
	if buf[2] != 0x00 {
		test.Error()
	}
	if buf[3] != 0x07 {
		test.Error()
	}
	if buf[4] != 0x00 {
		test.Error()
	}
	if buf[5] != 0x00 {
		test.Error()
	}
	if buf[6] != 0x12 {
		test.Error()
	}
	if buf[7] != 0x34 {
		test.Error()
	}
	if buf[8] != 0x00 {
		test.Error()
	}
	if buf[9] != 0x00 {
		test.Error()
	}
	if buf[10] != 0x12 {
		test.Error()
	}
	if buf[11] != 0x34 {
		test.Error()
	}
	if buf[12] != 0x05 {
		test.Error()
	}
	if buf[13] != 0x00 {
		test.Error()
	}
	if buf[14] != 0x00 {
		test.Error()
	}
	if buf[15] != 0x12 {
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
	if buf[19] != 0xaa {
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
	if buf[23] != 0xbb {
		test.Error()
	}
	if buf[24] != 0x00 {
		test.Error()
	}
	if buf[25] != 0x00 {
		test.Error()
	}
	if buf[26] != 0x00 {
		test.Error()
	}
	if buf[27] != 0xcc {
		test.Error()
	}
	if buf[28] != 0x00 {
		test.Error()
	}
	if buf[29] != 0x00 {
		test.Error()
	}
	if buf[30] != 0x00 {
		test.Error()
	}
	if buf[31] != 0xdd {
		test.Error()
	}
}

func TestReceiverReportUnmarshal(test *testing.T) {
	buf := []byte{0x81, 0xc9, 0x00, 0x07, 0x00, 0x00, 0x12, 0x34, 0x00, 0x00, 0x12, 0x34, 0x05, 0x00, 0x00, 0x12, 0x00,
		0x00, 0x00, 0xaa, 0x00, 0x00, 0x00, 0xbb, 0x00, 0x00, 0x00, 0xcc, 0x00, 0x00, 0x00, 0xdd}
	report := &ReceiverReport{}
	err := report.Unmarshal(buf)
	if err != nil {
		test.Error()
	}
	if report.Version != 2 {
		test.Error()
	}
	if report.Padding != false {
		test.Error()
	}
	if report.ReportCount != 1 {
		test.Error()
	}
	if report.PacketType != 201 {
		test.Error()
	}
	if report.Length != 0x07 {
		test.Error()
	}
	if report.SSRC != 0x1234 {
		test.Error()
	}
	if len(report.Blocks) != 1 {
		test.Error()
	}
	if report.Blocks[0].SSRC != 0x1234 {
		test.Error()
	}
	if report.Blocks[0].FractionLost != 0x5 {
		test.Error()
	}
	if report.Blocks[0].CumulativeLost != 0x12 {
		test.Error()
	}
	if report.Blocks[0].EHSNR != 0xAA {
		test.Error()
	}
	if report.Blocks[0].InterarrivalJitter != 0xBB {
		test.Error()
	}
	if report.Blocks[0].LSR != 0xCC {
		test.Error()
	}
	if report.Blocks[0].DLSR != 0xDD {
		test.Error()
	}
}
