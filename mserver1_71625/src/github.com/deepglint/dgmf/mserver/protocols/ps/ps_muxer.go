package ps

type StreamType int

const (
	PS_PES_PAYLOAD_MAX_LEN = 65535
)

const (
	DG_PS_STREAM_VIDEO = iota
	DG_PS_STREAM_AUDIO
)

const (
	DG_H264_HEADER_END   = 255
	DG_H264_HEADER_I     = 1
	DG_H264_HEADER_PPS   = 2
	DG_H264_HEADER_SPS   = 3
	DG_H264_HEADER_OTHER = 4
)

/*
 * @brief Callback function of ps_mux, it will return ps mux data
 *
 * @param data The head pointer of program stream mux data
 * @param length The length of data after program stream mux
 */
type PSMuxOut func(out []byte, length uint32)

type PSMuxer struct {
	psMuxout PSMuxOut
}

func NewPSMuxer(psMuxOut PSMuxOut) PSMuxer {
	psM := PSMuxer{}
	psM.psMuxout = psMuxOut
	return psM
}

/*
 * @brief Do video program stream mux from element stream
 *
 * @param buffer The element stream buffer
 * @param length The length of element stream buffer
 * @param timestamp The timestamp of this stream
 * @param stream_type The data type after program stream demux
 * @param is_iframe If the frame in this element stream package is I frame,
 *        this param should be true, else it should be false
 * @param mux_data_handler The callback function for output PS
 */
func (psM *PSMuxer) Mux(esData []byte, length uint32, timestamp uint64, streamType StreamType, isIFrame bool) {
	if streamType == DG_PS_STREAM_VIDEO {
		var pmuxbuf []byte
		if length > 0xffff {
			pmuxbuf = make([]byte, 65535)
		} else {
			pmuxbuf = make([]byte, length+200)
		}
		psM.muxVideoESStream(esData, length, timestamp, isIFrame, pmuxbuf)
	}
}

/*
* @brief Program stream header initialization function
 *
 * @param psheader Program stream header struct
*/
func (psM *PSMuxer) PSHeaderInit(psHeader *PSHeader) {
	psHeader.packStartCode[0] = 0x00
	psHeader.packStartCode[1] = 0x00
	psHeader.packStartCode[2] = 0x01
	psHeader.packStartCode[3] = 0xBA
	psHeader.fixBit = 0x01
	psHeader.markerBit = 0x01
	psHeader.markerBit1 = 0x01
	psHeader.markerBit2 = 0x01
	psHeader.markerBit3 = 0x01
	psHeader.markerBit4 = 0x01
	psHeader.markerBit5 = 0x01
	psHeader.reserved = 0x1F
	psHeader.packStuffingLength = 0x00
	psHeader.systemClockReferenceExtension1 = 0
	psHeader.systemClockReferenceExtension2 = 0
}

/*
 * @brief Set system clock reference base
 *
 * @param psheader Program stream header struct
 * @param scr SCR of program stream
 */
func (psM *PSMuxer) SetSystemClockReferenceBase(psHeader *PSHeader, scr uint64) {
	psHeader.systemClockReferenceBase1 = byte((scr >> 30) & 0x07)
	psHeader.systemClockReferenceBase21 = byte((scr >> 28) & 0x03)
	psHeader.systemClockReferenceBase22 = byte((scr >> 20) & 0xFF)
	psHeader.systemClockReferenceBase23 = byte((scr >> 15) & 0x1F)
	psHeader.systemClockReferenceBase31 = byte((scr >> 13) & 0x03)
	psHeader.systemClockReferenceBase32 = byte((scr >> 5) & 0xFF)
	psHeader.systemClockReferenceBase33 = byte(scr & 0x1F)
}

/*
 * @brief Set program mux rate
 *
 * @param psheader Program stream header struct
 * @param mux_rate Program mux rate
 */
func (psM *PSMuxer) SetProgramMuxRate(psHeader *PSHeader, muxRate uint32) {
	psHeader.programMuxRate1 = byte((muxRate >> 14) & 0xFF)
	psHeader.programMuxRate2 = byte((muxRate >> 6) & 0xFF)
	psHeader.programMuxRate3 = byte(muxRate & 0x3F)
}

/*
 * @brief packetized elementary stream header initialization function
 *
 * @param pesheader Packetized elementary stream header struct
 */
func (psM *PSMuxer) PESHeaderInit(pesHeader *PESHeader) {

	pesHeader.packetStartCodePrefix[0] = 0x00
	pesHeader.packetStartCodePrefix[1] = 0x00
	pesHeader.packetStartCodePrefix[2] = 0x01

	pesHeader.dataAlignmentIndicator = 0x01

	pesHeader.PESPacketLength[0] = 0x00
	pesHeader.PESPacketLength[1] = 0x00

	pesHeader.streamID = 0xE0
	pesHeader.fixBit = 0x02
}

/*
 * @brief partial system header initialization function
 *
 * @param psysheader partial system header struct
 */
func (psM *PSMuxer) PartialSystemHeaderInit(parsysHeader *PartialSystemHeader) {

	parsysHeader.packetStartCodePrefix[0] = 0x00
	parsysHeader.packetStartCodePrefix[1] = 0x00
	parsysHeader.packetStartCodePrefix[2] = 0x01

	parsysHeader.streamID = 0xBB

	parsysHeader.headerLength[0] = 0x0C >> 8
	parsysHeader.headerLength[1] = 0x0C

	parsysHeader.markerBit1 = 1
	parsysHeader.markerBit2 = 1
	parsysHeader.rateBound1 = 0
	parsysHeader.rateBound2 = 0
	parsysHeader.rateBound3 = 0

	parsysHeader.CSPSFlag = 0
	parsysHeader.fixedFlag = 0
	parsysHeader.audioBound = 0

	parsysHeader.videoBound = 0
	parsysHeader.markerBit3 = 1
	parsysHeader.systemVideoLockFlag = 0
	parsysHeader.systemAudioFockFlag = 0

	parsysHeader.packetRateRestrictionFlag = 0
	parsysHeader.reservedByte = 0x7f

}

/*
 * @brief ps map initialization function
 *
 * @param psmap ps map struct
 */
func (psM *PSMuxer) PSMapHeaderInit(psMapHeader *PSMapHeader) {

	psMapHeader.packetStartCodePrefix[0] = 0x00
	psMapHeader.packetStartCodePrefix[1] = 0x00
	psMapHeader.packetStartCodePrefix[2] = 0x01

	psMapHeader.mapStreamID = 0xBC
	psMapHeader.programStreamMapLength[0] = 0x00
	psMapHeader.programStreamMapLength[1] = 0x12

	psMapHeader.programStreamMapVersion = 0x00
	psMapHeader.currentNextIndicator = 0x01
	psMapHeader.reserved1 = 0x03
	psMapHeader.programStreamMapVersion = 0x00

	psMapHeader.reserved2 = 0x7F
	psMapHeader.markerBit = 0x01

	psMapHeader.programStreamInfoLength[0] = 0x00
	psMapHeader.programStreamInfoLength[1] = 0x00
}

/*
 * @brief pts pack initialization function
 *
 * @param pts_pack pts pack struct
 */
func (psM *PSMuxer) PTSPackInit(ptsPack *PTSPack) {

	ptsPack.fixBit = 0x02
	ptsPack.markerBit = 0x01
	ptsPack.markerBit1 = 0x01
	ptsPack.markerBit2 = 0x01

}

/*
 * @brief set pts value to pts pack struct
 *
 * @param pts_pack pts pack struct
 * @param pts pts value
 */
func (psM *PSMuxer) SetPTS(ptsPack *PTSPack, pts uint64) {
	ptsPack.PTS1 = byte((pts >> 30) & 0x07)
	ptsPack.PTS21 = byte((pts >> 22) & 0xFF)
	ptsPack.PTS22 = byte((pts >> 15) & 0x7F)
	ptsPack.PTS31 = byte((pts >> 7) & 0xFF)
	ptsPack.PTS32 = byte(pts & 0x7F)
}

/*
 * @brief pts_dts pack initialization function
 *
 * @param ptsDtsPack pts_dts pack struct
 */
func (psM *PSMuxer) PTSDTSPackInit(ptsDtsPack *PTSDTSPack) {

	ptsDtsPack.DTSFixBit = 0x02
	ptsDtsPack.DTSMarkerBit = 0x01
	ptsDtsPack.DTSMarkerBit1 = 0x01
	ptsDtsPack.DTSMarkerBit2 = 0x01

	ptsDtsPack.PTSFixBit = 0x03
	ptsDtsPack.PTSMarkerBit = 0x01
	ptsDtsPack.PTSMarkerBit1 = 0x01
	ptsDtsPack.PTSMarkerBit2 = 0x01

}

/*
 * @brief set pts value to pts_dts pack struct
 *
 * @param pts_dts_pack pts_dts pack struct
 * @param pts pts value
 */
func (psM *PSMuxer) SetPTS2(ptsDtsPack *PTSDTSPack, pts uint64) {
	ptsDtsPack.PTSPTS1 = byte((pts >> 30) & 0x07)
	ptsDtsPack.PTSPTS21 = byte((pts >> 22) & 0xFF)
	ptsDtsPack.PTSPTS22 = byte((pts >> 15) & 0x7F)
	ptsDtsPack.PTSPTS31 = byte((pts >> 7) & 0xFF)
	ptsDtsPack.PTSPTS32 = byte(pts & 0x7F)
}

/*
 * @brief set dts value to pts_dts pack struct
 *
 * @param pts_dts_pack pts_dts pack struct
 * @param dts dts value
 */
func (psM *PSMuxer) SetDTS(ptsDtsPack *PTSDTSPack, dts uint64) {
	ptsDtsPack.DTSPTS1 = byte((dts >> 30) & 0x07)
	ptsDtsPack.DTSPTS21 = byte((dts >> 22) & 0xFF)
	ptsDtsPack.DTSPTS22 = byte((dts >> 15) & 0x7F)
	ptsDtsPack.DTSPTS31 = byte((dts >> 7) & 0xFF)
	ptsDtsPack.DTSPTS32 = byte(dts & 0x7F)
}

func (psM *PSMuxer) muxVideoESStream(esData []byte, length uint32, timestamp uint64, isIFrame bool, pbuf []byte) {

	var noffset uint32 = 0
	psHeader := new(PSHeader)
	pesHeader := new(PESHeader)
	systemHeader := new(PartialSystemHeader)
	psMapHeader := new(PSMapHeader)

	//PSHeader
	psM.PSHeaderInit(psHeader)
	psM.SetSystemClockReferenceBase(psHeader, timestamp)
	psHeaderData := psHeader.GetHeaderBytes()
	copy(pbuf, psHeaderData)
	noffset = uint32(len(psHeaderData))

	if isIFrame {

		//systemHeader
		psM.PartialSystemHeaderInit(systemHeader)
		systemHeaderData := systemHeader.GetHeaderBytes()
		copy(pbuf[noffset:], systemHeaderData)
		noffset += uint32(len(systemHeaderData))

		//systemHeader Video
		pvediomsg := new(PartialSystemStreamMessage)
		pvediomsg.MarkerBit3 = 3
		pvediomsg.NstreamID = 0xE0
		pvediomsgData := pvediomsg.GetPartialSystemStreamMessageBytes()
		copy(pbuf[noffset:], pvediomsgData)
		noffset += uint32(len(pvediomsgData))

		//systemHeader Audio
		paudiomsg := new(PartialSystemStreamMessage)
		paudiomsg.MarkerBit3 = 3
		paudiomsg.NstreamID = 0xC0
		paudiomsgData := paudiomsg.GetPartialSystemStreamMessageBytes()
		copy(pbuf[noffset:], paudiomsgData)
		noffset += uint32(len(paudiomsgData))

		//PSMapHeader
		psM.PSMapHeaderInit(psMapHeader)
		psMapHeaderData := psMapHeader.GetHeaderBytes()
		copy(pbuf[noffset:], psMapHeaderData)
		noffset += uint32(len(psMapHeaderData))

		pbuf[noffset] = 0x00
		pbuf[noffset+1] = 0x08
		noffset += 2

		// psMap video
		pmapVideo := new(PSMapStream)
		pmapVideo.streamType = 0x1B
		pmapVideo.elementaryStreamID = 0xE0
		pmapVideoData := pmapVideo.GetPSMapStreamBytes()
		copy(pbuf[noffset:], pmapVideoData)
		noffset += uint32(len(pmapVideoData))

		//psmAudio
		pmapAudio := new(PSMapStream)
		pmapAudio.elementaryStreamID = 0xC0
		pmapAudio.streamType = 0x90
		pmapAudioData := pmapAudio.GetPSMapStreamBytes()
		copy(pbuf[noffset:], pmapAudioData)
		noffset += uint32(len(pmapAudioData))

		pbuf[noffset] = 0x45
		pbuf[noffset+1] = 0xBD
		pbuf[noffset+2] = 0xDC
		pbuf[noffset+3] = 0xF4
		noffset += 4

		//pesHeader
		var pesHeaderPoint uint32 = noffset
		psM.PESHeaderInit(pesHeader)
		pesHeader.PTSDTSFlags = 3
		pesHeader.PESHeaderDataLength = 10
		pesHeaderData := pesHeader.GetHeaderBytes()
		// copy(pbuf[noffset:], pesHeaderData)
		noffset += uint32(len(pesHeaderData))

		pack := new(PTSDTSPack)
		psM.PTSDTSPackInit(pack)
		psM.SetPTS2(pack, timestamp)
		psM.SetDTS(pack, timestamp)
		packData := pack.GetPTSDTSPackBytes()
		copy(pbuf[noffset:], packData)
		noffset += uint32(len(packData))

		//var bFindIFrame bool = false
		var bSendPsHeader bool = false
		var ptemp uint32 = 0
		var pnote uint32 = 0
		var tempStreamLen uint32 = length

		for {
			pnote = ptemp
			headerType := psM.findPSHeaderType(esData, ptemp+5, tempStreamLen-5, &ptemp)
			if headerType != DG_H264_HEADER_END {
				if !bSendPsHeader {
					bSendPsHeader = true

					//h264data
					copy(pbuf[noffset:], esData[pnote:ptemp])
					noffset += (ptemp - pnote)

					//pesHeader
					pesDataLen := uint32(len(pesHeaderData)) + uint32(len(packData)) + (ptemp - pnote) - 6
					pesHeader.PESPacketLength[0] = byte(pesDataLen >> 8)
					pesHeader.PESPacketLength[1] = byte(pesDataLen & 0xFF)
					copy(pbuf[pesHeaderPoint:], pesHeader.GetHeaderBytes())

					//callback
					psM.psMuxout(pbuf, noffset)

				} else {

					//h264data
					copy(pbuf[pesHeaderPoint+noffset:], esData[pnote:ptemp])

					//pesheader
					noffset += (ptemp - pnote)
					pesDataLen := noffset - 6
					pesHeader.PESPacketLength[0] = byte(pesDataLen >> 8)
					pesHeader.PESPacketLength[1] = byte(pesDataLen & 0xFF)
					copy(pbuf[pesHeaderPoint:], pesHeader.GetHeaderBytes())

					//callback
					psM.psMuxout(pbuf[pesHeaderPoint:], noffset)
				}
				tempStreamLen -= uint32(ptemp - pnote)
				noffset = uint32(len(pesHeaderData)) + uint32(len(packData))
			} else {
				// bFindIFrame = true
				if tempStreamLen+noffset <= PS_PES_PAYLOAD_MAX_LEN {

					//h264data
					copy(pbuf[pesHeaderPoint+noffset:], esData[pnote:pnote+tempStreamLen])
					noffset += tempStreamLen
					//pesheader
					pesDataLen := noffset - 6
					pesHeader.PESPacketLength[0] = byte(pesDataLen >> 8)
					pesHeader.PESPacketLength[1] = byte(pesDataLen & 0xFF)
					copy(pbuf[pesHeaderPoint:], pesHeader.GetHeaderBytes())

					//callback
					psM.psMuxout(pbuf[pesHeaderPoint:], noffset)
					break
				} else {

					//h264data

					copy(pbuf[pesHeaderPoint+noffset:], esData[pnote:pnote+PS_PES_PAYLOAD_MAX_LEN-200])
					noffset += PS_PES_PAYLOAD_MAX_LEN - 200

					//pesheader
					pesDataLen := noffset - 6
					pesHeader.PESPacketLength[0] = byte(pesDataLen >> 8)
					pesHeader.PESPacketLength[1] = byte(pesDataLen & 0xFF)
					copy(pbuf[pesHeaderPoint:], pesHeader.GetHeaderBytes())

					//callback
					psM.psMuxout(pbuf[pesHeaderPoint:], noffset)

					ptemp += (PS_PES_PAYLOAD_MAX_LEN - 200)
					noffset = uint32(len(pesHeaderData)) + uint32(len(packData))
					tempStreamLen -= (PS_PES_PAYLOAD_MAX_LEN - 200)

				}
			}
		}
	} else {

		// pesHeader
		var pesHeaderPoint uint32 = noffset
		psM.PESHeaderInit(pesHeader)
		pesHeader.PTSDTSFlags = 3
		pesHeader.PESHeaderDataLength = 10
		pesHeaderData := pesHeader.GetHeaderBytes()
		noffset = uint32(len(pesHeaderData))

		pack := new(PTSDTSPack)
		psM.PTSDTSPackInit(pack)
		psM.SetPTS2(pack, timestamp)
		psM.SetDTS(pack, timestamp)
		packData := pack.GetPTSDTSPackBytes()
		copy(pbuf[pesHeaderPoint+noffset:], packData)
		noffset += uint32(len(packData))

		var pnote uint32 = 0
		var tempStreamLen uint32 = length

		for {
			if tempStreamLen+noffset <= PS_PES_PAYLOAD_MAX_LEN {
				//h264data
				copy(pbuf[pesHeaderPoint+noffset:], esData[pnote:pnote+tempStreamLen])
				noffset += tempStreamLen

				//pesheader
				pesDataLen := noffset - 6
				pesHeader.PESPacketLength[0] = byte(pesDataLen >> 8)
				pesHeader.PESPacketLength[1] = byte(pesDataLen & 0xFF)
				// fmt.Println(pesHeaderPoint)
				copy(pbuf[pesHeaderPoint:], pesHeader.GetHeaderBytes())

				//callback
				if tempStreamLen == length {
					psM.psMuxout(pbuf, pesHeaderPoint+noffset)
				} else {
					psM.psMuxout(pbuf[pesHeaderPoint:], noffset)
				}
				break
			} else {
				//h264data
				copy(pbuf[pesHeaderPoint+noffset:], esData[pnote:pnote+PS_PES_PAYLOAD_MAX_LEN-200])
				noffset += PS_PES_PAYLOAD_MAX_LEN - 200

				//pesheader
				pesDataLen := noffset - 6
				pesHeader.PESPacketLength[0] = byte(pesDataLen >> 8)
				pesHeader.PESPacketLength[1] = byte(pesDataLen & 0xFF)
				copy(pbuf[pesHeaderPoint:], pesHeader.GetHeaderBytes())

				//callback
				if tempStreamLen == length {
					psM.psMuxout(pbuf, noffset+uint32(len(psHeaderData)))
				} else {
					psM.psMuxout(pbuf[pesHeaderPoint:], noffset)
				}

				pnote += (PS_PES_PAYLOAD_MAX_LEN - 200)
				noffset = uint32(len(pesHeaderData)) + uint32(len(packData))
				tempStreamLen -= (PS_PES_PAYLOAD_MAX_LEN - 200)
			}
		}
	}
}

func (psM *PSMuxer) findPSHeaderType(esData []byte, pBeign uint32, len uint32, pHeader *uint32) int {
	headerType := DG_H264_HEADER_END
	pTemp := pBeign
	for pTemp+3 < pBeign+len {
		if esData[pTemp] == 0x00 && esData[pTemp+1] == 0x00 && esData[pTemp+2] == 0x00 && esData[pTemp+3] == 0x01 {

			if esData[pTemp+4] == 0x67 {
				headerType = DG_H264_HEADER_SPS
			} else if esData[pTemp+4] == 0x68 {
				headerType = DG_H264_HEADER_PPS
			} else if esData[pTemp+4] == 0x65 {
				headerType = DG_H264_HEADER_I
			} else {
				headerType = DG_H264_HEADER_OTHER
			}
			*pHeader = pTemp
			// fmt.Println(pTemp)
			// fmt.Println(headerType)
			break
		}
		pTemp++
	}
	return headerType
}
