package rtcp

import (
	"testing"
)

func TestSenderReportMarshal(test *testing.T) {
	report := &SenderReport{
		Version:      2,
		Padding:      false,
		ReportCount:  1,
		PacketType:   200,
		SSRC:         0x1234,
		NTPTimestamp: 0x987654321,
		RTPTimestamp: 0x1234,
		PacketCount:  0xAA,
		OctetCount:   0xBB,
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
	if buf[1] != 0xc8 {
		test.Error()
	}
	if buf[2] != 0x00 {
		test.Error()
	}
	if buf[3] != 0x0C {
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
	if buf[10] != 0x00 {
		test.Error()
	}
	if buf[11] != 0x09 {
		test.Error()
	}
	if buf[12] != 0x87 {
		test.Error()
	}
	if buf[13] != 0x65 {
		test.Error()
	}
	if buf[14] != 0x43 {
		test.Error()
	}
	if buf[15] != 0x21 {
		test.Error()
	}
	if buf[16] != 0x00 {
		test.Error()
	}
	if buf[17] != 0x00 {
		test.Error()
	}
	if buf[18] != 0x12 {
		test.Error()
	}
	if buf[19] != 0x34 {
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
	if buf[23] != 0xaa {
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
	if buf[27] != 0xbb {
		test.Error()
	}
	if buf[28] != 0x00 {
		test.Error()
	}
	if buf[29] != 0x00 {
		test.Error()
	}
	if buf[30] != 0x12 {
		test.Error()
	}
	if buf[31] != 0x34 {
		test.Error()
	}
	if buf[32] != 0x05 {
		test.Error()
	}
	if buf[33] != 0x00 {
		test.Error()
	}
	if buf[34] != 0x00 {
		test.Error()
	}
	if buf[35] != 0x12 {
		test.Error()
	}
	if buf[36] != 0x00 {
		test.Error()
	}
	if buf[37] != 0x00 {
		test.Error()
	}
	if buf[38] != 0x00 {
		test.Error()
	}
	if buf[39] != 0xaa {
		test.Error()
	}
	if buf[40] != 0x00 {
		test.Error()
	}
	if buf[41] != 0x00 {
		test.Error()
	}
	if buf[42] != 0x00 {
		test.Error()
	}
	if buf[43] != 0xbb {
		test.Error()
	}
	if buf[44] != 0x00 {
		test.Error()
	}
	if buf[45] != 0x00 {
		test.Error()
	}
	if buf[46] != 0x00 {
		test.Error()
	}
	if buf[47] != 0xcc {
		test.Error()
	}
	if buf[48] != 0x00 {
		test.Error()
	}
	if buf[49] != 0x00 {
		test.Error()
	}
	if buf[50] != 0x00 {
		test.Error()
	}
	if buf[51] != 0xdd {
		test.Error()
	}
}

func TestSenderReportUnmarshal(test *testing.T) {
	buf := []byte{0x81, 0xc8, 0x00, 0x0C, 0x00, 0x00, 0x12, 0x34, 0x00, 0x00, 0x00, 0x09, 0x87, 0x65, 0x43, 0x21,
		0x00, 0x00, 0x12, 0x34, 0x00, 0x00, 0x00, 0xaa, 0x00, 0x00, 0x00, 0xbb,
		0x00, 0x00, 0x12, 0x34, 0x05, 0x00, 0x00, 0x12, 0x00, 0x00, 0x00, 0xaa, 0x00, 0x00, 0x00, 0xbb, 0x00, 0x00, 0x00, 0xcc, 0x00, 0x00, 0x00, 0xdd}
	report := &SenderReport{}
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
	if report.PacketType != 200 {
		test.Error()
	}
	if report.Length != 0x0C {
		test.Error()
	}
	if report.SSRC != 0x1234 {
		test.Error()
	}
	if report.NTPTimestamp != 0x987654321 {
		test.Error()
	}
	if report.RTPTimestamp != 0x1234 {
		test.Error()
	}
	if report.PacketCount != 0xAA {
		test.Error()
	}
	if report.OctetCount != 0xBB {
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
