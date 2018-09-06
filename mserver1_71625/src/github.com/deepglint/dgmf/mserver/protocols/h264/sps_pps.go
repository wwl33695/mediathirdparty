package h264

import (
	"encoding/base64"
	"math"
)

/*
7.3.2.1.1 Sequence parameter set data syntax

seq_parameter_set_data( ) {
	profile_idc																		u(8)
	constraint_set0_flag															u(1)
	constraint_set1_flag															u(1)
	constraint_set2_flag															u(1)
	constraint_set3_flag															u(1)
	constraint_set4_flag															u(1)
	constraint_set5_flag															u(1)
	reserved_zero_2bits																u(2)
	level_idc																		u(8)
	seq_parameter_set_id															ue(v)
	if( profile_idc = = 100 || profile_idc = = 110 ||
		profile_idc = = 122 || profile_idc = = 244 ||
		profile_idc = = 44 || profile_idc = = 83 ||
		profile_idc = = 86 || profile_idc = = 118 ||
		profile_idc = = 128 ) {
			chroma_format_idc														ue(v)
			if( chroma_format_idc = = 3 )
				separate_colour_plane_flag											u(1)
			bit_depth_luma_minus8													ue(v)
			bit_depth_chroma_minus8													ue(v)
			qpprime_y_zero_transform_bypass_flag									u(1)
			seq_scaling_matrix_present_flag											u(1)
			if( seq_scaling_matrix_present_flag )
				for(i=0;i<((chroma_format_idc != 3)?8:12);i++) {
					seq_scaling_list_present_flag[i]								u(1)
					if( seq_scaling_list_present_flag[i] )
						if( i < 6 )
							scaling_list( ScalingList4x4[i], 16,
							UseDefaultScalingMatrix4x4Flag[i])
						else
							scaling_list( ScalingList8x8[i−6], 64,
							UseDefaultScalingMatrix8x8Flag[i−6] )
				}
	}
	log2_max_frame_num_minus4														ue(v)
	pic_order_cnt_type																ue(v)
	if( pic_order_cnt_type = = 0 )
		log2_max_pic_order_cnt_lsb_minus4											ue(v)
	else if( pic_order_cnt_type = = 1 ) {
		delta_pic_order_always_zero_flag											u(1)
		offset_for_non_ref_pic														se(v)
		offset_for_top_to_bottom_field												se(v)
		num_ref_frames_in_pic_order_cnt_cycle										ue(v)
		for( i = 0; i < num_ref_frames_in_pic_order_cnt_cycle; i++ )
			offset_for_ref_frame[i]													se(v)
		max_num_ref_frames															ue(v)
		gaps_in_frame_num_value_allowed_flag										u(1)
		pic_width_in_mbs_minus1														ue(v)
		pic_height_in_map_units_minus1												ue(v)
		frame_mbs_only_flag															u(1)
		if( !frame_mbs_only_flag )
			mb_adaptive_frame_field_flag											u(1)
		direct_8x8_inference_flag													u(1)
		frame_cropping_flag															u(1)
		if( frame_cropping_flag ) {
			frame_crop_left_offset													ue(v)
			frame_crop_right_offset													ue(v)
			frame_crop_top_offset													ue(v)
			frame_crop_bottom_offset												ue(v)
		}
		vui_parameters_present_flag													u(1)
		if( vui_parameters_present_flag )
			vui_parameters( )
	}

https://www.itu.int/rec/T-REC-H.264-201003-S/en
*/
type SPS struct {
	ForbiddenZeroBit                int
	NalRefIdc                       int
	NalUnitType                     int
	ProfileIdc                      int
	ConstraintSet0Flag              int
	ConstraintSet1Flag              int
	ConstraintSet2Flag              int
	ConstraintSet3Flag              int
	ConstraintSet4Flag              int
	ConstraintSet5Flag              int
	ReservedZero2Bits               int
	LevelIdc                        int
	SeqParameterSetId               int
	ChromaFormatIdc                 int
	SeparateColourPlaneFlag         int
	BitDepthLumaMinus8              int
	BitDepthChromaMinus8            int
	QpprimeYZeroTransformBypassFlag int
	SeqScalingMatrixPresentFlag     int
	SeqScalingListPresentFlag       []int
	ScalingList4x4                  [][]int
	UseDefaultScalingMatrix4x4Flag  []int
	ScalingList8x8                  [][]int
	UseDefaultScalingMatrix8x8Flag  []int
	Log2MaxFrameNumMinus4           int
	PicOrderCntType                 int
	Log2MaxPicOrderCntLsbMinus4     int
	DeltaPicOrderAlwaysZeroFlag     int
	OffsetForNonRefPic              int
	OffsetForTopToBottomField       int
	NumRefFramesInPicOrderCntCycle  int
	OffsetForRefFrame               []int
	MaxNumRefFrames                 int
	GapsInFrameNumValueAllowedFlag  int
	PicWidthInMbsMinus1             int
	PicHeightInMapUnitsMinus1       int
	FrameMbsOnlyFlag                int
	MbAdaptiveFrameFieldFlag        int
	Direct8x8InferenceFlag          int
	FrameCroppingFlag               int
	FrameCropLeftOffset             int
	FrameCropRightOffset            int
	FrameCropTopOffset              int
	FrameCropBottomOffset           int
	VUIParametersPresentFlag        int
	VUIParameters                   *VUI
}

/*
E.1.1 VUI parameters syntax

vui_parameters( ) {
	aspect_ratio_info_present_flag													u(1)
	if( aspect_ratio_info_present_flag ) {
		aspect_ratio_idc															u(8)
		if( aspect_ratio_idc = = Extended_SAR ) {
			sar_width																u(16)
			sar_height																u(16)
		}
	}
	overscan_info_present_flag														u(1)
	if( overscan_info_present_flag )
		overscan_appropriate_flag													u(1)
	video_signal_type_present_flag													u(1)
	if( video_signal_type_present_flag ) {
		video_format																u(3)
		video_full_range_flag														u(1)
		colour_description_present_flag												u(1)
		if( colour_description_present_flag ) {
			colour_primaries														u(8)
			transfer_characteristics												u(8)
			matrix_coefficients														u(8)
		}
	}
	chroma_loc_info_present_flag													u(1)
	if( chroma_loc_info_present_flag ) {
		chroma_sample_loc_type_top_field											ue(v)
		chroma_sample_loc_type_bottom_field											ue(v)
	}
	timing_info_present_flag														u(1)
	if( timing_info_present_flag ) {
		num_units_in_tick															u(32)
		time_scale																	u(32)
		fixed_frame_rate_flag														u(1)
	}
	nal_hrd_parameters_present_flag													u(1)
	if( nal_hrd_parameters_present_flag )
		hrd_parameters( )
	vcl_hrd_parameters_present_flag													u(1)
	if( vcl_hrd_parameters_present_flag )
		hrd_parameters( )
	if( nal_hrd_parameters_present_flag | | vcl_hrd_parameters_present_flag )
		low_delay_hrd_flag															u(1)
	pic_struct_present_flag															u(1)
	bitstream_restriction_flag														u(1)
	if( bitstream_restriction_flag ) {
		motion_vectors_over_pic_boundaries_flag										u(1)
		max_bytes_per_pic_denom														ue(v)
		max_bits_per_mb_denom														ue(v)
		log2_max_mv_length_horizontal												ue(v)
		log2_max_mv_length_vertical													ue(v)
		num_reorder_frames															ue(v)
		max_dec_frame_buffering														ue(v)
	}
}

https://www.itu.int/rec/T-REC-H.264-201003-S/en
*/
type VUI struct {
	AspectRatioInfoPresentFlag         int
	AspectRatioIdc                     int
	SarWidth                           int
	SarHeight                          int
	OverscanInfoPresentFlag            int
	OverscanAppropriateFlag            int
	VideoSignalTypePresentFlag         int
	VideoFormat                        int
	VideoFullRangeFlag                 int
	ColourDescriptionPresentFlag       int
	ColourPrimaries                    int
	TransferCharacteristics            int
	MatrixCoefficients                 int
	ChromaLocInfoPresentFlag           int
	ChromaSampleLocTypeTopField        int
	ChromaSampleLocTypeBottomField     int
	TimingInfoPresentFlag              int
	NumUnitsInTick                     int
	TimeScale                          int
	FixedFrameRateFlag                 int
	NalHrdParametersPresentFlag        int
	NalHrdParameters                   *HRD
	VclHrdParametersPresentFlag        int
	VclHrdParameters                   *HRD
	LowDelayHrdFlag                    int
	PicStructPresentFlag               int
	BitstreamRestrictionFlag           int
	MotionVectorsOverPicBoundariesFlag int
	MaxBytesPerPicDenom                int
	MaxBitsPerMbDenom                  int
	Log2MaxMvLengthHorizontal          int
	Log2MaxMvLengthVertical            int
	NumReorderFrames                   int
	MaxDecFrameBuffering               int
}

/*
E.1.2 HRD parameters syntax
hrd_parameters( ) {
	cpb_cnt_minus1																	ue(v)
	bit_rate_scale																	u(4)
	cpb_size_scale																	u(4)
	for( SchedSelIdx = 0; SchedSelIdx <= cpb_cnt_minus1; SchedSelIdx++ ) {
		bit_rate_value_minus1[ SchedSelIdx ]										ue(v)
		cpb_size_value_minus1[ SchedSelIdx ]										ue(v)
		cbr_flag[ SchedSelIdx ]														u(1)
	}
	initial_cpb_removal_delay_length_minus1											u(5)
	cpb_removal_delay_length_minus1													u(5)
	dpb_output_delay_length_minus1													u(5)
	time_offset_length																u(5)
}

https://www.itu.int/rec/T-REC-H.264-201003-S/en
*/
type HRD struct {
	CpbCntMinus1                       int
	BitRateScale                       int
	CpbSizeScale                       int
	BitRateValueMinus1                 []int
	CpbSizeValueMinus1                 []int
	CbrFlag                            []int
	InitialCpbRemovalDelayLengthMinus1 int
	CpbRemovalDelayLengthMinus1        int
	DpbOutputDelayLengthMinus1         int
	TimeOffsetLength                   int
}

func U(data []byte, bitCnt int, startBit int) int {
	if len(data) == 0 {
		return -1
	}
	if bitCnt <= 0 {
		return -1
	}
	if startBit < 0 {
		return -1
	}
	if bitCnt+startBit > len(data)*8 {
		return -1
	}
	ret := 0
	start := startBit
	for i := 0; i < bitCnt; i++ {
		ret <<= 1
		if (data[start/8] & (0x80 >> uint(start%8))) != 0 {
			ret += 1
		}
		start++
	}
	return ret
}

/*
https://en.wikipedia.org/wiki/Exponential-Golomb_coding

*/
func UE(data []byte, startBit int) (int, int) {
	if len(data) == 0 {
		return -1, -1
	}
	var ret int = 0
	var leadingZeroBits int = -1
	for b := 0; b != 1; leadingZeroBits++ {
		b = U(data, 1, startBit)
		if b < 0 {
			return -1, -1
		}
		startBit++
	}

	if leadingZeroBits > 0 {
		ret = int(math.Exp2(float64(leadingZeroBits))) - 1 + U(data, leadingZeroBits, startBit)
	} else {
		ret = 0
	}
	return ret, 2*leadingZeroBits + 1
}

func SE(data []byte, startBit int) (int, int) {
	k, skip := UE(data, startBit)

	if k&0x01 == 1 {
		k = (k + 1) / 2
	} else {
		k = -(k / 2)
	}
	return k, skip
}

/*
scaling_list( scalingList, sizeOfScalingList, useDefaultScalingMatrixFlag ) {
	lastScale = 8
 	nextScale = 8
	for( j = 0; j < sizeOfScalingList; j++ ) {
 		if( nextScale != 0 ) {
 			delta_scale																se(v)
 			nextScale = ( lastScale + delta_scale + 256 ) % 256
 			useDefaultScalingMatrixFlag = ( j = = 0 && nextScale = = 0 )
 		}
 		scalingList[j] = ( nextScale = = 0 ) ? lastScale : nextScale
 		lastScale = scalingList[j]
 	}
}
*/

func scalingList(list []int, sizeOfScalingList int, useDefaultScalingMatrixFlag *int, data []byte, position int) int {
	lastScale := 8
	nextScale := 8
	posTotal := 0
	list = make([]int, sizeOfScalingList)
	for j := 0; j < sizeOfScalingList; j++ {
		if nextScale != 0 {
			deltaScale, pos := SE(data, position)
			position += pos
			posTotal += pos
			nextScale = (lastScale + deltaScale + 256) % 256
			if j == 0 && nextScale == 0 {
				*useDefaultScalingMatrixFlag = 1
			} else {
				*useDefaultScalingMatrixFlag = 0
			}
		}

		if nextScale == 0 {
			list[j] = lastScale
		} else {
			list[j] = nextScale
		}
		lastScale = list[j]
	}
	return posTotal
}

func unmarshalHRD(data []byte, position int) (*HRD, int) {
	var skip int = 0
	hrdParameters := &HRD{}

	hrdParameters.CpbCntMinus1, skip = UE(data, position)
	position += skip

	hrdParameters.BitRateScale = U(data, 4, position)
	position += 4

	hrdParameters.CpbSizeScale = U(data, 4, position)
	position += 4

	hrdParameters.BitRateValueMinus1 = []int{}
	hrdParameters.CpbSizeValueMinus1 = []int{}
	hrdParameters.CbrFlag = []int{}
	for i := 0; i <= hrdParameters.CpbCntMinus1; i++ {
		v0, skip := UE(data, position)
		position += skip
		hrdParameters.BitRateValueMinus1 = append(hrdParameters.BitRateValueMinus1, v0)

		v1, skip := UE(data, position)
		position += skip
		hrdParameters.CpbSizeValueMinus1 = append(hrdParameters.CpbSizeValueMinus1, v1)

		hrdParameters.CbrFlag = append(hrdParameters.CbrFlag, U(data, 1, position))
		position += 1
	}

	hrdParameters.InitialCpbRemovalDelayLengthMinus1 = U(data, 5, position)
	position += 5

	hrdParameters.CpbRemovalDelayLengthMinus1 = U(data, 5, position)
	position += 5

	hrdParameters.DpbOutputDelayLengthMinus1 = U(data, 5, position)
	position += 5

	hrdParameters.TimeOffsetLength = U(data, 5, position)
	position += 5

	return hrdParameters, position
}

func unmarshalVUI(data []byte, position int) (*VUI, int) {
	var skip int = 0
	vuiParameters := &VUI{}

	vuiParameters.AspectRatioInfoPresentFlag = U(data, 1, position)
	position += 1

	if vuiParameters.AspectRatioInfoPresentFlag > 0 {
		vuiParameters.AspectRatioIdc = U(data, 8, position)
		position += 8

		if vuiParameters.AspectRatioIdc == 255 {
			vuiParameters.SarWidth = U(data, 16, position)
			position += 16

			vuiParameters.SarHeight = U(data, 16, position)
			position += 16
		}
	}

	vuiParameters.OverscanInfoPresentFlag = U(data, 1, position)
	position += 1

	if vuiParameters.OverscanInfoPresentFlag > 0 {
		vuiParameters.OverscanAppropriateFlag = U(data, 1, position)
		position += 1
	}

	vuiParameters.VideoSignalTypePresentFlag = U(data, 1, position)
	position += 1

	if vuiParameters.VideoSignalTypePresentFlag > 0 {
		vuiParameters.VideoFormat = U(data, 3, position)
		position += 3

		vuiParameters.VideoFullRangeFlag = U(data, 1, position)
		position += 1

		vuiParameters.ColourDescriptionPresentFlag = U(data, 1, position)
		position += 1

		if vuiParameters.ColourDescriptionPresentFlag > 0 {
			vuiParameters.ColourPrimaries = U(data, 8, position)
			position += 8

			vuiParameters.TransferCharacteristics = U(data, 8, position)
			position += 8

			vuiParameters.MatrixCoefficients = U(data, 8, position)
			position += 8
		}
	}

	vuiParameters.ChromaLocInfoPresentFlag = U(data, 1, position)
	position += 1

	if vuiParameters.ChromaLocInfoPresentFlag > 0 {
		vuiParameters.ChromaSampleLocTypeTopField, skip = UE(data, position)
		position += skip

		vuiParameters.ChromaSampleLocTypeBottomField, skip = UE(data, position)
		position += skip
	}

	vuiParameters.TimingInfoPresentFlag = U(data, 1, position)
	position += 1

	if vuiParameters.TimingInfoPresentFlag > 0 {
		vuiParameters.NumUnitsInTick = U(data, 32, position)
		position += 32

		vuiParameters.TimeScale = U(data, 32, position)
		position += 32

		vuiParameters.FixedFrameRateFlag = U(data, 1, position)
		position += 1
	}

	vuiParameters.NalHrdParametersPresentFlag = U(data, 1, position)
	position += 1

	if vuiParameters.NalHrdParametersPresentFlag > 0 {
		vuiParameters.NalHrdParameters, position = unmarshalHRD(data, position)
	}

	vuiParameters.VclHrdParametersPresentFlag = U(data, 1, position)
	position += 1

	if vuiParameters.VclHrdParametersPresentFlag > 0 {
		vuiParameters.VclHrdParameters, position = unmarshalHRD(data, position)
	}

	if vuiParameters.NalHrdParametersPresentFlag > 0 || vuiParameters.VclHrdParametersPresentFlag > 0 {
		vuiParameters.LowDelayHrdFlag = U(data, 1, position)
		position += 1
	}

	vuiParameters.PicStructPresentFlag = U(data, 1, position)
	position += 1

	vuiParameters.BitstreamRestrictionFlag = U(data, 1, position)
	position += 1

	if vuiParameters.BitstreamRestrictionFlag > 0 {
		vuiParameters.MotionVectorsOverPicBoundariesFlag = U(data, 1, position)
		position += 1

		vuiParameters.MaxBytesPerPicDenom, skip = UE(data, position)
		position += skip

		vuiParameters.MaxBitsPerMbDenom, skip = UE(data, position)
		position += skip

		vuiParameters.Log2MaxMvLengthHorizontal, skip = UE(data, position)
		position += skip

		vuiParameters.Log2MaxMvLengthVertical, skip = UE(data, position)
		position += skip

		vuiParameters.NumReorderFrames, skip = UE(data, position)
		position += skip

		vuiParameters.MaxDecFrameBuffering, skip = UE(data, position)
		position += skip
	}

	return vuiParameters, position
}

func (this *SPS) Unmarshal(data []byte) {
	var skip int = 0
	var position int = 0

	this.ForbiddenZeroBit = U(data[4:], 1, position)
	position += 1

	this.NalRefIdc = U(data[4:], 2, position)
	position += 2

	this.NalUnitType = U(data[4:], 5, position)
	position += 5

	if this.NalUnitType == 0x07 {
		this.ProfileIdc = U(data[4:], 8, position)
		position += 8

		this.ConstraintSet0Flag = U(data[4:], 1, position)
		position += 1

		this.ConstraintSet1Flag = U(data[4:], 1, position)
		position += 1

		this.ConstraintSet2Flag = U(data[4:], 1, position)
		position += 1

		this.ConstraintSet3Flag = U(data[4:], 1, position)
		position += 1

		this.ConstraintSet4Flag = U(data[4:], 1, position)
		position += 1

		this.ConstraintSet5Flag = U(data[4:], 1, position)
		position += 1

		this.ReservedZero2Bits = U(data[4:], 2, position)
		position += 2

		this.LevelIdc = U(data[4:], 8, position)
		position += 8

		this.SeqParameterSetId, skip = UE(data[4:], position)
		position += skip

		if this.ProfileIdc == 100 || this.ProfileIdc == 110 || this.ProfileIdc == 122 ||
			this.ProfileIdc == 244 || this.ProfileIdc == 44 || this.ProfileIdc == 83 ||
			this.ProfileIdc == 86 || this.ProfileIdc == 118 || this.ProfileIdc == 128 ||
			this.ProfileIdc == 138 || this.ProfileIdc == 139 || this.ProfileIdc == 134 {

			this.ChromaFormatIdc, skip = UE(data[4:], position)
			position += skip

			if this.ChromaFormatIdc == 3 {
				this.SeparateColourPlaneFlag = U(data[4:], 1, position)
				position += 1
			}

			this.BitDepthLumaMinus8, skip = UE(data[4:], position)
			position += skip

			this.BitDepthChromaMinus8, skip = UE(data[4:], position)
			position += skip

			this.QpprimeYZeroTransformBypassFlag = U(data[4:], 1, position)
			position += 1

			this.SeqScalingMatrixPresentFlag = U(data[4:], 1, position)
			position += 1

			if this.SeqScalingMatrixPresentFlag > 0 {
				this.ScalingList4x4 = make([][]int, 6)
				this.ScalingList8x8 = make([][]int, 2)
				this.UseDefaultScalingMatrix4x4Flag = make([]int, 6)
				this.UseDefaultScalingMatrix8x8Flag = make([]int, 2)
				this.SeqScalingListPresentFlag = make([]int, 8)

				for i := 0; i < 8; i++ {
					this.SeqScalingListPresentFlag[i] = U(data[4:], 1, position)
					position += 1
					if this.SeqScalingListPresentFlag[i] == 1 {
						if i < 6 {
							position += scalingList(this.ScalingList4x4[i], 16, &this.UseDefaultScalingMatrix4x4Flag[i], data[4:], position)
						} else {
							position += scalingList(this.ScalingList8x8[i-6], 64, &this.UseDefaultScalingMatrix8x8Flag[i-6], data[4:], position)
						}
					}
				}
			}
		}

		this.Log2MaxFrameNumMinus4, skip = UE(data[4:], position)
		position += skip

		this.PicOrderCntType, skip = UE(data[4:], position)
		position += skip

		if this.PicOrderCntType == 0 {
			this.Log2MaxPicOrderCntLsbMinus4, skip = UE(data[4:], position)
			position += skip
		} else if this.PicOrderCntType == 1 {
			this.DeltaPicOrderAlwaysZeroFlag = U(data[4:], 1, position)
			position += 1

			this.OffsetForNonRefPic, skip = SE(data[4:], position)
			position += skip

			this.OffsetForTopToBottomField, skip = SE(data[4:], position)
			position += skip

			this.NumRefFramesInPicOrderCntCycle, skip = UE(data[4:], position)
			position += skip

			for i := 0; i < this.NumRefFramesInPicOrderCntCycle; i++ {
				v, skip := SE(data[4:], position)
				position += skip
				this.OffsetForRefFrame = append(this.OffsetForRefFrame, v)
			}
		}

		this.MaxNumRefFrames, skip = UE(data[4:], position)
		position += skip

		this.GapsInFrameNumValueAllowedFlag = U(data[4:], 1, position)
		position += 1

		this.PicWidthInMbsMinus1, skip = UE(data[4:], position)
		position += skip

		this.PicHeightInMapUnitsMinus1, skip = UE(data[4:], position)
		position += skip

		this.FrameMbsOnlyFlag = U(data[4:], 1, position)
		position += 1

		if this.FrameMbsOnlyFlag <= 0 {
			this.MbAdaptiveFrameFieldFlag = U(data[4:], 1, position)
			position += 1
		}

		this.Direct8x8InferenceFlag = U(data[4:], 1, position)
		position += 1

		this.FrameCroppingFlag = U(data[4:], 1, position)
		position += 1

		if this.FrameCroppingFlag > 0 {
			this.FrameCropLeftOffset, skip = UE(data[4:], position)
			position += skip

			this.FrameCropRightOffset, skip = UE(data[4:], position)
			position += skip

			this.FrameCropTopOffset, skip = UE(data[4:], position)
			position += skip

			this.FrameCropBottomOffset, skip = UE(data[4:], position)
			position += skip
		}

		this.VUIParametersPresentFlag = U(data[4:], 1, position)
		position += 1

		if this.VUIParametersPresentFlag > 0 {
			this.VUIParameters, position = unmarshalVUI(data[4:], position)
		}
	}
}

func FindSPSBytes(IDR []byte) []byte {
	begin := -1
	end := 0
	for i := 0; i < len(IDR); i++ {
		if i+4 < len(IDR) && begin < 0 && IDR[i] == 0x00 && IDR[i+1] == 0x00 && IDR[i+2] == 0x00 && IDR[i+3] == 0x01 && IDR[i+4]&0x1F == 0x07 {
			i += 3
			begin = i + 1
			continue
		}
		if i+4 < len(IDR) && begin > 0 && IDR[i] == 0x00 && IDR[i+1] == 0x00 && IDR[i+2] == 0x00 && IDR[i+3] == 0x01 && IDR[i+4]&0x1F != 0x07 {
			end = i
			break
		}
		if i == len(IDR)-1 {
			end = len(IDR)
		}
	}

	if end > begin && end <= len(IDR) && begin > 0 {
		return IDR[begin:end]
	} else {
		return []byte{}
	}
}

func FindPPSBytes(IDR []byte) []byte {
	begin := -1
	end := 0
	for i := 0; i < len(IDR); i++ {
		if i+4 < len(IDR) && begin < 0 && IDR[i] == 0x00 && IDR[i+1] == 0x00 && IDR[i+2] == 0x00 && IDR[i+3] == 0x01 && IDR[i+4]&0x1F == 0x08 {
			i += 3
			begin = i + 1
			continue
		}
		if i+4 < len(IDR) && begin > 0 && IDR[i] == 0x00 && IDR[i+1] == 0x00 && IDR[i+2] == 0x00 && IDR[i+3] == 0x01 && IDR[i+4]&0x1F != 0x08 {
			end = i
			break
		}
		if i == len(IDR)-1 {
			end = len(IDR)
		}
	}

	if end > begin && end <= len(IDR) && begin > 0 {
		return IDR[begin:end]
	} else {
		return []byte{}
	}
}

type SPSPPS struct {
	SPS    string
	PPS    string
	Width  uint32
	Height uint32
}

func GetLivePPS(data []byte) SPSPPS {
	sp := SPSPPS{}
	n := len(data)
	if n >= 4 && data[0] == 0x00 && data[1] == 0x00 && data[2] == 0x00 && data[3] == 0x01 && data[4]&0x1F == 8 {
		if len(data) > 200 {
			for i := 4; i < n-4; i++ {
				if data[i] == 0x00 && data[i+1] == 0x00 && data[i+2] == 0x01 {
					sp.PPS = base64.StdEncoding.EncodeToString(data[4:i])
					return sp
				}
			}
		}
		sp.PPS = base64.StdEncoding.EncodeToString(data[4:])
		return sp
	}
	return sp
}

func GetLiveSPS(data []byte) SPSPPS {
	sp := SPSPPS{}
	n := len(data)
	if n >= 4 && data[0] == 0x00 && data[1] == 0x00 && data[2] == 0x00 && data[3] == 0x01 && data[4]&0x1F == 7 {
		sps := &SPS{}
		if len(data) > 200 {
			for i := 4; i < n-4; i++ {
				if data[i] == 0x00 && data[i+1] == 0x00 && data[i+2] == 0x00 && data[i+3] == 0x01 {
					sps.Unmarshal(data[:i])
					sp.SPS = base64.StdEncoding.EncodeToString(data[4:i])

					if sps.FrameCroppingFlag == 1 {
						sp.Width = uint32(16*(sps.PicWidthInMbsMinus1+1) - sps.FrameCropLeftOffset*2 - sps.FrameCropRightOffset*2)
						sp.Height = uint32((2-sps.FrameMbsOnlyFlag)*16*(sps.PicHeightInMapUnitsMinus1+1) - sps.FrameCropTopOffset*2 - sps.FrameCropBottomOffset*2)
					} else {
						sp.Width = uint32((sps.PicWidthInMbsMinus1 + 1) * 16)
						sp.Height = uint32((2 - sps.FrameMbsOnlyFlag) * (sps.PicHeightInMapUnitsMinus1 + 1) * 16)
					}
					if data[i+4]&0x1F == 8 {
						p := GetLivePPS(data[i:])
						if len(p.PPS) != 0 {
							sp.PPS = p.PPS
						}
					}
					return sp
				}
			}
		}

		sps.Unmarshal(data)
		sp.SPS = base64.StdEncoding.EncodeToString(data[4:])

		// if sps.FrameCroppingFlag == 1 {
		// 	sp.Width = uint32(16*(sps.PicWidthInMbsMinus1+1) - sps.FrameCropLeftOffset*2 - sps.FrameCropRightOffset*2)
		// 	sp.Height = uint32((2-sps.FrameMbsOnlyFlag)*16*(sps.PicHeightInMapUnitsMinus1+1) - sps.FrameCropTopOffset*2 - sps.FrameCropBottomOffset*2)
		// } else {
		// 	fmt.Println(sps.SeparateColourPlaneFlag)
		// 	sp.Width = uint32((sps.PicWidthInMbsMinus1 + 1) * 16)
		// 	sp.Height = uint32((2 - sps.FrameMbsOnlyFlag) * (sps.PicHeightInMapUnitsMinus1 + 1) * 16)
		// }

		sp.Width = uint32(16 * (sps.PicWidthInMbsMinus1 + 1))
		sp.Height = uint32(16 * (2 - sps.FrameMbsOnlyFlag) * (sps.PicHeightInMapUnitsMinus1 + 1))
		if sps.SeparateColourPlaneFlag == 1 || sps.ChromaFormatIdc == 0 {
			sps.FrameCropBottomOffset *= (2 - sps.FrameMbsOnlyFlag)
			sps.FrameCropTopOffset *= (2 - sps.FrameMbsOnlyFlag)
		} else if sps.SeparateColourPlaneFlag == 0 && sps.ChromaFormatIdc > 0 {
			if sps.ChromaFormatIdc == 1 || sps.ChromaFormatIdc == 2 {
				sps.FrameCropLeftOffset *= 2
				sps.FrameCropRightOffset *= 2
			}
			if sps.ChromaFormatIdc == 1 {
				sps.FrameCropTopOffset *= 2
				sps.FrameCropBottomOffset *= 2
			}
		}
		sp.Width -= uint32(sps.FrameCropLeftOffset + sps.FrameCropRightOffset)
		sp.Height -= uint32(sps.FrameCropTopOffset + sps.FrameCropBottomOffset)

		return sp
	}
	return sp
}

// func GetVodPPS(data []byte, stream *core.VodStream) {
// 	n := len(data)
// 	if n >= 4 && data[0] == 0x00 && data[1] == 0x00 && data[2] == 0x00 && data[3] == 0x01 && data[4]&0x1F == 8 {
// 		for i := 4; i < n-4; i++ {
// 			if data[i] == 0x00 && data[i+1] == 0x00 && data[i+2] == 0x01 {
// 				stream.PPS = base64.StdEncoding.EncodeToString(data[4:i])
// 				break
// 			}
// 		}
// 	}
// }

// func GetVodSPS(data []byte, stream *core.VodStream) {
// 	n := len(data)
// 	if n >= 4 && data[0] == 0x00 && data[1] == 0x00 && data[2] == 0x00 && data[3] == 0x01 && data[4]&0x1F == 7 {
// 		sps := &SPS{}
// 		for i := 4; i < n-4; i++ {
// 			if data[i] == 0x00 && data[i+1] == 0x00 && data[i+2] == 0x00 && data[i+3] == 0x01 {
// 				sps.Unmarshal(data[:i])
// 				stream.SPS = base64.StdEncoding.EncodeToString(data[4:i])
// 				if data[i+4]&0x1F == 8 {
// 					GetVodPPS(data[i:], stream)
// 				}
// 				break
// 			}
// 		}

// 		stream.Width = 16*(sps.PicWidthInMbsMinus1+1) - sps.FrameCropRightOffset*2 - sps.FrameCropLeftOffset*2
// 		stream.Height = (2-sps.FrameMbsOnlyFlag)*16*(sps.PicHeightInMapUnitsMinus1+1) - sps.FrameCropTopOffset*2 - sps.FrameCropBottomOffset*2
// 	}
// }
