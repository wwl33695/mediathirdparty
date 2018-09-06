package h264

import (
	// "fmt"
	"testing"
)

func TestU(test *testing.T) {
	var rec int
	var data []byte

	data = []byte{0x02}

	for i := 0; i < 6; i++ {
		if rec = U(data, 1, i); rec != 0 {
			test.Error()
		}
	}

	if rec = U(data, 1, 6); rec != 1 {
		test.Error()
	}

	if rec = U(data, 1, 7); rec != 0 {
		test.Error()
	}

	if rec = U(data, 2, 6); rec != 2 {
		test.Error()
	}

	if rec = U(data, 1, 8); rec != -1 {
		test.Error()
	}

	if rec = U(data, 0, 7); rec != -1 {
		test.Error()
	}

	data = []byte{}
	if rec = U(data, 1, 6); rec != -1 {
		test.Error()
	}
}

func TestUE(test *testing.T) {
	var rec int
	var skip int
	var data []byte

	data = []byte{0x0A, 0x00}
	if rec, skip = UE(data, 0); rec != 19 || skip != 9 {
		test.Error()
	}

	data = []byte{}
	if rec, skip = UE(data, 0); rec != -1 || skip != -1 {
		test.Error()
	}

	data = []byte{0x00, 0x3E, 0x30, 0xA3, 0x00}
	if rec, skip = UE(data, 0); rec != 1989 || skip != 21 {
		test.Error()
	}
	if rec, skip = UE(data, 21); rec != 9 || skip != 7 {
		test.Error()
	}
	if rec, skip = UE(data, 28); rec != 5 || skip != 5 {
		test.Error()
	}

	data = []byte{0x00}
	if rec, skip = UE(data, 28); rec != -1 || skip != -1 {
		test.Error()
	}
}

func TestSE(test *testing.T) {
	var rec int
	var skip int
	var data []byte

	data = []byte{0x0A, 0x80}
	if rec, skip = SE(data, 0); rec != -10 || skip != 9 {
		test.Error()
	}
}

func TestUnmarshal(test *testing.T) {
	var data []byte
	sps := &SPS{}

	data = []byte{0x00, 0x00, 0x00, 0x01, 0x67, 0x42, 0x40, 0x28, 0x95, 0xA0, 0x50, 0x7E, 0x40, 0x00, 0x00, 0x00, 0x01,
		0x68, 0xCE, 0x3C, 0x80}
	sps.Unmarshal(data)
	if sps.LevelIdc != 40 {
		test.Error()
	}
	if sps.PicHeightInMapUnitsMinus1 != 14 {
		test.Error()
	}
	if sps.PicWidthInMbsMinus1 != 19 {
		test.Error()
	}

	data = []byte{0x00, 0x00, 0x00, 0x01, 0x67, 0x4d, 0x00, 0x32, 0x95, 0xa8, 0x0a, 0x00, 0x2d, 0x69, 0xb8, 0x08, 0x08,
		0x08, 0x10}
	sps.Unmarshal(data)
	if sps.PicWidthInMbsMinus1 != 159 {
		test.Error()
	}
	if sps.PicHeightInMapUnitsMinus1 != 89 {
		test.Error()
	}
	if sps.VUIParametersPresentFlag != 1 {
		test.Error()
	}
	if sps.VUIParameters.VideoFormat != 5 {
		test.Error()
	}

	data = []byte{0x00, 0x00, 0x00, 0x01, 0x67, 0x64, 0x00, 0x28, 0x1E, 0x48, 0xB1, 0x7F, 0xFF, 0x80}
	sps.Unmarshal(data)
	if sps.ChromaFormatIdc != 3 {
		test.Error()
	}
	if sps.BitDepthChromaMinus8 != 10 {
		test.Error()
	}
	if len(sps.SeqScalingListPresentFlag) != 12 {
		test.Error()
	}

	data = []byte{0x00, 0x00, 0x00, 0x01, 0x67, 0x64, 0x00, 0x28, 0x16, 0x2C, 0x58, 0xBF, 0xFC}
	sps.Unmarshal(data)
	if len(sps.SeqScalingListPresentFlag) != 8 {
		test.Error()
	}

	data = []byte{0x00, 0x00, 0x00, 0x01, 0x67, 0x64, 0x00, 0x28, 0x16, 0x2C, 0x58, 0xBF, 0xFC, 0xF0, 0xB8}
	sps.Unmarshal(data)
	if sps.Log2MaxPicOrderCntLsbMinus4 != 22 {
		test.Error()
	}

	data = []byte{0x00, 0x00, 0x00, 0x01, 0x67, 0x64, 0x00, 0x28, 0x16, 0x2C, 0x58, 0xBF, 0xFC, 0xE8, 0x15, 0x0A, 0xB0,
		0xA8, 0x54}
	sps.Unmarshal(data)
	if sps.OffsetForNonRefPic != -10 {
		test.Error()
	}
	if sps.OffsetForTopToBottomField != -10 {
		test.Error()
	}
	if len(sps.OffsetForRefFrame) != 2 {
		test.Error()
	}

	if sps.OffsetForRefFrame[0] != -10 {
		test.Error()
	}

	if sps.OffsetForRefFrame[1] != -10 {
		test.Error()
	}

	data = []byte{0x00, 0x00, 0x00, 0x01, 0x67, 0x64, 0x00, 0x28, 0x16, 0x2C, 0x58, 0xBF, 0xFC, 0xE8, 0x15, 0x0A, 0xB0,
		0xA8, 0x54, 0x2A, 0x0A, 0x85, 0x4C, 0x2A, 0x15, 0x0A, 0x85, 0x40}
	sps.Unmarshal(data)
	if sps.FrameMbsOnlyFlag != 0 {
		test.Error()
	}
	if sps.Direct8x8InferenceFlag != 1 {
		test.Error()
	}
	if sps.FrameCropLeftOffset != 20 {
		test.Error()
	}
	if sps.VUIParametersPresentFlag != 0 {
		test.Error()
	}
}

func TestUnmarshalHRD(test *testing.T) {
	var data []byte
	var hrd *HRD

	data = []byte{0x27, 0x71, 0xCF, 0x39, 0xE7, 0x3C, 0xE7, 0xEB, 0x5A, 0xD0}
	hrd, _ = unmarshalHRD(data, 0)
	if hrd.CpbCntMinus1 != 3 {
		test.Error()
	}
	if hrd.BitRateScale != 14 {
		test.Error()
	}
	if hrd.CpbSizeScale != 14 {
		test.Error()
	}
	if len(hrd.BitRateValueMinus1) != 4 {
		test.Error()
	}
	if len(hrd.CpbSizeValueMinus1) != 4 {
		test.Error()
	}
	if len(hrd.CbrFlag) != 4 {
		test.Error()
	}
	if hrd.BitRateValueMinus1[0] != 6 {
		test.Error()
	}
	if hrd.CpbSizeValueMinus1[0] != 6 {
		test.Error()
	}
	if hrd.CbrFlag[0] != 1 {
		test.Error()
	}
	if hrd.TimeOffsetLength != 26 {
		test.Error()
	}
}

func TestUnmarshalVUI(test *testing.T) {
	var data []byte
	var vui *VUI

	data = []byte{0xFF, 0x83, 0xC0, 0x02, 0x1C, 0x75, 0xBD, 0xBD, 0xBD, 0xCE, 0x78, 0x09, 0x9D, 0x10, 0x38, 0x09, 0x9D,
		0x10, 0x3E, 0x4E, 0xE3, 0x9E, 0x73, 0xCE, 0x79, 0xCF, 0xD6, 0xB5, 0xA9, 0x3B, 0x8E, 0x79, 0xCF, 0x39, 0xE7,
		0x3F, 0x5A, 0xD6, 0xBC, 0xE7, 0x39, 0xCE, 0x70}
	vui, _ = unmarshalVUI(data, 0)
	if vui.AspectRatioInfoPresentFlag != 1 {
		test.Error()
	}
	if vui.AspectRatioIdc != 255 {
		test.Error()
	}
	if vui.SarWidth != 1920 {
		test.Error()
	}
	if vui.SarHeight != 1080 {
		test.Error()
	}
	if vui.OverscanInfoPresentFlag != 1 {
		test.Error()
	}
	if vui.OverscanAppropriateFlag != 1 {
		test.Error()
	}
	if vui.VideoSignalTypePresentFlag != 1 {
		test.Error()
	}
	if vui.VideoFormat != 2 {
		test.Error()
	}
	if vui.VideoFullRangeFlag != 1 {
		test.Error()
	}
	if vui.ColourDescriptionPresentFlag != 1 {
		test.Error()
	}
	if vui.ColourPrimaries != 123 {
		test.Error()
	}
	if vui.TransferCharacteristics != 123 {
		test.Error()
	}
	if vui.MatrixCoefficients != 123 {
		test.Error()
	}
	if vui.ChromaLocInfoPresentFlag != 1 {
		test.Error()
	}
	if vui.ChromaSampleLocTypeTopField != 6 {
		test.Error()
	}
	if vui.ChromaSampleLocTypeBottomField != 6 {
		test.Error()
	}
	if vui.TimingInfoPresentFlag != 1 {
		test.Error()
	}
	if vui.NumUnitsInTick != 20161031 {
		test.Error()
	}
	if vui.TimeScale != 20161031 {
		test.Error()
	}
	if vui.FixedFrameRateFlag != 1 {
		test.Error()
	}
	if vui.NalHrdParametersPresentFlag != 1 {
		test.Error()
	}
	if vui.VclHrdParametersPresentFlag != 1 {
		test.Error()
	}
	if vui.LowDelayHrdFlag != 1 {
		test.Error()
	}
	if vui.PicStructPresentFlag != 1 {
		test.Error()
	}
	if vui.BitstreamRestrictionFlag != 1 {
		test.Error()
	}
	if vui.MotionVectorsOverPicBoundariesFlag != 1 {
		test.Error()
	}
	if vui.MaxBytesPerPicDenom != 6 {
		test.Error()
	}
	if vui.MaxBitsPerMbDenom != 6 {
		test.Error()
	}
	if vui.Log2MaxMvLengthHorizontal != 6 {
		test.Error()
	}
	if vui.Log2MaxMvLengthVertical != 6 {
		test.Error()
	}
	if vui.NumReorderFrames != 6 {
		test.Error()
	}
	if vui.MaxDecFrameBuffering != 6 {
		test.Error()
	}
}

func TestFindSPSBytes(test *testing.T) {
	data := FindSPSBytes([]byte{0x00, 0x00, 0x00, 0x01, 0x67, 0x4D, 0x00, 0x32, 0x95, 0xA8, 0x0A, 0x00, 0x2D, 0x69,
		0xB8, 0x08, 0x08, 0x08, 0x10, 0x00, 0x00, 0x00, 0x01, 0x68, 0xEE, 0x3C, 0x80, 0x00, 0x00, 0x00, 0x01,
		0x06, 0xE5, 0x01, 0x58, 0x80, 0x00, 0x00, 0x00, 0x01})

	if len(data) != 15 {
		test.Error()
	}
	if data[0] != 0x67 {
		test.Error()
	}
	if data[1] != 0x4D {
		test.Error()
	}
	if data[2] != 0x00 {
		test.Error()
	}
	if data[3] != 0x32 {
		test.Error()
	}
	if data[4] != 0x95 {
		test.Error()
	}
	if data[5] != 0xA8 {
		test.Error()
	}
	if data[6] != 0x0A {
		test.Error()
	}
	if data[7] != 0x00 {
		test.Error()
	}
	if data[8] != 0x2D {
		test.Error()
	}
	if data[9] != 0x69 {
		test.Error()
	}
	if data[10] != 0xB8 {
		test.Error()
	}
	if data[11] != 0x08 {
		test.Error()
	}
	if data[12] != 0x08 {
		test.Error()
	}
	if data[13] != 0x08 {
		test.Error()
	}
	if data[14] != 0x10 {
		test.Error()
	}

	data = FindSPSBytes([]byte{0x00, 0x00, 0x00, 0x01, 0x67, 0x4D, 0x00, 0x32, 0x95, 0xA8, 0x0A, 0x00, 0x2D, 0x69,
		0xB8, 0x08, 0x08, 0x08, 0x10})

	if len(data) != 15 {
		test.Error()
	}
	if data[0] != 0x67 {
		test.Error()
	}
	if data[1] != 0x4D {
		test.Error()
	}
	if data[2] != 0x00 {
		test.Error()
	}
	if data[3] != 0x32 {
		test.Error()
	}
	if data[4] != 0x95 {
		test.Error()
	}
	if data[5] != 0xA8 {
		test.Error()
	}
	if data[6] != 0x0A {
		test.Error()
	}
	if data[7] != 0x00 {
		test.Error()
	}
	if data[8] != 0x2D {
		test.Error()
	}
	if data[9] != 0x69 {
		test.Error()
	}
	if data[10] != 0xB8 {
		test.Error()
	}
	if data[11] != 0x08 {
		test.Error()
	}
	if data[12] != 0x08 {
		test.Error()
	}
	if data[13] != 0x08 {
		test.Error()
	}
	if data[14] != 0x10 {
		test.Error()
	}

	data = FindSPSBytes([]byte{0x00, 0x00, 0x00, 0x01, 0x68, 0xEE, 0x3C, 0x80, 0x00, 0x00, 0x00, 0x01,
		0x06, 0xE5, 0x01, 0x58, 0x80, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x67, 0x4D, 0x00, 0x32, 0x95, 0xA8, 0x0A, 0x00, 0x2D, 0x69,
		0xB8, 0x08, 0x08, 0x08, 0x10})

	if len(data) != 15 {
		test.Error()
	}
	if data[0] != 0x67 {
		test.Error()
	}
	if data[1] != 0x4D {
		test.Error()
	}
	if data[2] != 0x00 {
		test.Error()
	}
	if data[3] != 0x32 {
		test.Error()
	}
	if data[4] != 0x95 {
		test.Error()
	}
	if data[5] != 0xA8 {
		test.Error()
	}
	if data[6] != 0x0A {
		test.Error()
	}
	if data[7] != 0x00 {
		test.Error()
	}
	if data[8] != 0x2D {
		test.Error()
	}
	if data[9] != 0x69 {
		test.Error()
	}
	if data[10] != 0xB8 {
		test.Error()
	}
	if data[11] != 0x08 {
		test.Error()
	}
	if data[12] != 0x08 {
		test.Error()
	}
	if data[13] != 0x08 {
		test.Error()
	}
	if data[14] != 0x10 {
		test.Error()
	}

	data = FindSPSBytes([]byte{0x00, 0x00, 0x00, 0x01, 0x67, 0x4D, 0x00, 0x32, 0x95, 0xA8, 0x0A, 0x00, 0x2D, 0x69,
		0xB8, 0x08, 0x08, 0x08, 0x10})

	if len(data) != 15 {
		test.Error()
	}
	if data[0] != 0x67 {
		test.Error()
	}
	if data[1] != 0x4D {
		test.Error()
	}
	if data[2] != 0x00 {
		test.Error()
	}
	if data[3] != 0x32 {
		test.Error()
	}
	if data[4] != 0x95 {
		test.Error()
	}
	if data[5] != 0xA8 {
		test.Error()
	}
	if data[6] != 0x0A {
		test.Error()
	}
	if data[7] != 0x00 {
		test.Error()
	}
	if data[8] != 0x2D {
		test.Error()
	}
	if data[9] != 0x69 {
		test.Error()
	}
	if data[10] != 0xB8 {
		test.Error()
	}
	if data[11] != 0x08 {
		test.Error()
	}
	if data[12] != 0x08 {
		test.Error()
	}
	if data[13] != 0x08 {
		test.Error()
	}
	if data[14] != 0x10 {
		test.Error()
	}

	data = FindSPSBytes([]byte{0x00, 0x00, 0x01, 0x67, 0x4D, 0x00, 0x32, 0x95, 0xA8, 0x0A, 0x00, 0x2D, 0x69,
		0xB8, 0x08, 0x08, 0x08, 0x10})

	if len(data) != 0 {
		test.Error()
	}
}

func TestFindPPSBytes(test *testing.T) {
	data := FindPPSBytes([]byte{0x00, 0x00, 0x00, 0x01, 0x67, 0x4D, 0x00, 0x32, 0x95, 0xA8, 0x0A, 0x00, 0x2D, 0x69,
		0xB8, 0x08, 0x08, 0x08, 0x10, 0x00, 0x00, 0x00, 0x01, 0x68, 0xEE, 0x3C, 0x80, 0x00, 0x00, 0x00, 0x01,
		0x06, 0xE5, 0x01, 0x58, 0x80, 0x00, 0x00, 0x00, 0x01})

	if len(data) != 4 {
		test.Error()
	}
	if data[0] != 0x68 {
		test.Error()
	}
	if data[1] != 0xEE {
		test.Error()
	}
	if data[2] != 0x3C {
		test.Error()
	}
	if data[3] != 0x80 {
		test.Error()
	}

	data = FindPPSBytes([]byte{0x00, 0x00, 0x00, 0x01, 0x68, 0xEE, 0x3C, 0x80, 0x00, 0x00, 0x00, 0x01,
		0x06, 0xE5, 0x01, 0x58, 0x80, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x67, 0x4D,
		0x00, 0x32, 0x95, 0xA8, 0x0A, 0x00, 0x2D, 0x69, 0xB8, 0x08, 0x08, 0x08, 0x10, 0x00})

	if len(data) != 4 {
		test.Error()
	}
	if data[0] != 0x68 {
		test.Error()
	}
	if data[1] != 0xEE {
		test.Error()
	}
	if data[2] != 0x3C {
		test.Error()
	}
	if data[3] != 0x80 {
		test.Error()
	}

	data = FindPPSBytes([]byte{0x00, 0x00, 0x00, 0x01, 0x68, 0xEE, 0x3C, 0x80})

	if len(data) != 4 {
		test.Error()
	}
	if data[0] != 0x68 {
		test.Error()
	}
	if data[1] != 0xEE {
		test.Error()
	}
	if data[2] != 0x3C {
		test.Error()
	}
	if data[3] != 0x80 {
		test.Error()
	}
}
