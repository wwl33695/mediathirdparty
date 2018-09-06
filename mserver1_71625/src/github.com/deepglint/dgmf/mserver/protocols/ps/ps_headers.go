package ps

//PSHeader 14 Byte
/*
 * Program stream (PS or MPEG-PS) is a container format for multiplexing digital
 * audio, video and more. The PS format is specified in MPEG-1 Part 1 (ISO/IEC
 * 11172-1) and MPEG-2 Part 1, Systems (ISO/IEC standard 13818-1/ITU-T H.222.0).
 * The MPEG-2 Program Stream is analogous and similar to ISO/IEC 11172 Systems
 * layer and it is forward compatible.
 *
 * Program streams are used on DVD-Video discs and HD DVD video discs, but with
 * some restrictions and extensions.[9][10] The filename extensions are VOB and
 * EVO respectively.
 *
 * PS Header (14 Byte):
 *
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 * |     Byte0     |     Byte1     |     Byte2     |     Byte3     |
 * |0 1 2 3 4 5 6 7 0 1 2 3 4 5 6 7 0 1 2 3 4 5 6 7 0 1 2 3 4 5 6 7|
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 * |         0000 0000 0000 0000 0000 0001         |   1011 1010   |
 * |                   start code                  |PACK identifier|
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 *
 *
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 * |     Byte4     |     Byte5     |     Byte6     |     Byte7     |
 * |0 1 2 3 4 5 6 7 0 1 2 3 4 5 6 7 0 1 2 3 4 5 6 7 0 1 2 3 4 5 6 7|
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 * |0 1|SCR  |1|          SCR 29..15         |1|    SCR 14..00   ...
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 *
 *
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 * |     Byte8     |     Byte9     |     Byte10    |     Byte11    |
 * |0 1 2 3 4 5 6 7 0 1 2 3 4 5 6 7 0 1 2 3 4 5 6 7 0 1 2 3 4 5 6 7|
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 * ...       |1|     SCR_ext     |1|       Program_Mux_Rate      ...
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 *
 *
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 * |      Byte12   |     Byte13    |
 * |0 1 2 3 4 5 6 7 0 1 2 3 4 5 6 7|
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 * ...         |1|1|reserved | PSL |
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 *
 * pack identifier -- 0xBA
 * PSL pack_stuffing_length
 *
 * SCR and SCR_ext together are the System Clock Reference, a counter driven at
 * 27MHz, used as a reference to synchronize streams. The clock is divided by
 * 300 (to match the 90KHz clocks such as PTS/DTS), the quotient is SCR (33
 * bits), the remainder is SCR_ext (9 bits)
 *
 * Program_Mux_Rate -- This is a 22 bit integer specifying the rate at which the
 * program stream target decoder receives the Program Stream during the pack in
 * which it is included. The value of program_mux_rate is measured in units of
 * 50 bytes/second. The value 0 is forbidden.
 *
 * pack_stuffing_length -- A 3 bit integer specifying the number of stuffing
 * bytes which follow this field.
 *
 * stuffing byte -- This is a fixed 8-bit value equal to '1111 1111' that can be
 * inserted by the encoder, for example to meet the requirements of the channel.
 * It is discarded by the decoder.
 *
 * Ref: http://dvd.sourceforge.net/dvdinfo/packhdr.html
 */
type PSHeader struct {
	packStartCode [4]byte //[4]

	systemClockReferenceBase21 byte //:2
	markerBit                  byte //:1
	systemClockReferenceBase1  byte //:3
	fixBit                     byte //:2

	systemClockReferenceBase22 byte //[1]

	systemClockReferenceBase31 byte //:2
	markerBit1                 byte //:1
	systemClockReferenceBase23 byte //:5

	systemClockReferenceBase32 byte //[1]

	systemClockReferenceExtension1 byte //:2
	markerBit2                     byte //:1
	systemClockReferenceBase33     byte //:5

	markerBit3                     byte //:1
	systemClockReferenceExtension2 byte //:7

	programMuxRate1 byte //[1]
	programMuxRate2 byte //[1]

	markerBit5      byte //:1
	markerBit4      byte //:1
	programMuxRate3 byte //:6

	packStuffingLength byte //:3
	reserved           byte //:5
}

func (psHeader *PSHeader) GetHeaderBytes() (desBytes []byte) {

	psHeaderData := [14]byte{}
	/* PS头起始码前缀：0x000001BA */
	psHeaderData[0] = psHeader.packStartCode[0]
	psHeaderData[1] = psHeader.packStartCode[1]
	psHeaderData[2] = psHeader.packStartCode[2]
	psHeaderData[3] = psHeader.packStartCode[3]

	/* 4~9字节为：SCR */

	psHeaderData[4] = byte(psHeader.systemClockReferenceBase21 |
		psHeader.markerBit<<2 |
		psHeader.systemClockReferenceBase1<<3 |
		psHeader.fixBit<<6)

	psHeaderData[5] = psHeader.systemClockReferenceBase22

	psHeaderData[6] = byte(psHeader.systemClockReferenceBase31 |
		psHeader.markerBit1<<2 |
		psHeader.systemClockReferenceBase23<<3)

	psHeaderData[7] = psHeader.systemClockReferenceBase32

	psHeaderData[8] = byte(psHeader.systemClockReferenceExtension1 |
		psHeader.markerBit2<<2 |
		psHeader.systemClockReferenceBase33<<3)

	psHeaderData[9] = byte(psHeader.markerBit3 |
		psHeader.systemClockReferenceExtension2<<1)

	/* 10~12字节为：PS流速率*/

	psHeaderData[10] = psHeader.programMuxRate1
	psHeaderData[11] = psHeader.programMuxRate2

	psHeaderData[12] = byte(psHeader.markerBit5 |
		psHeader.markerBit4<<1 |
		psHeader.programMuxRate3<<2)
	/* 填充字节数：2 */

	psHeaderData[13] = byte(psHeader.packStuffingLength |
		psHeader.reserved<<3)

	desBytes = []byte{}
	for _, value := range psHeaderData {
		desBytes = append(desBytes, value)
	}
	return desBytes
}

/*
 * Packetized Elementary Stream (PES) is a specification in the MPEG-2 Part 1
 * (Systems) (ISO/IEC 13818-1) and ITU-T H.222.0 that defines carrying of
 * elementary streams (usually the output of an audio or video encoder) in
 * packets within MPEG program streams and MPEG transport streams. The
 * elementary stream is packetized by encapsulating sequential data bytes from
 * the elementary stream inside PES packet headers.
 *
 * A typical method of transmitting elementary stream data from a video or audio
 * encoder is to first create PES packets from the elementary stream data and
 * then to encapsulate these PES packets inside Transport Stream (TS) packets or
 * Program Stream (PS) packets. The TS packets can then be multiplexed and
 * transmitted using broadcasting techniques, such as those used in an ATSC and
 * DVB.
 *
 * Transport Streams and Program Streams are each logically constructed from PES
 * packets. PES packets shall be used to convert between Transport Streams and
 * Program Streams. In some cases the PES packets need not be modified when
 * performing such conversions. PES packets may be much larger than the size of
 * a Transport Stream packet.
 *
 * PES Header (9 Byte):
 *
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 * |     Byte0     |     Byte1     |     Byte2     |     Byte3     |
 * |0 1 2 3 4 5 6 7 0 1 2 3 4 5 6 7 0 1 2 3 4 5 6 7 0 1 2 3 4 5 6 7|
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 * |         0000 0000 0000 0000 0000 0001         |               |
 * |                   start code                  |   Stream ID   |
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 *
 *
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 * |     Byte4     |     Byte5     |     Byte6     |     Byte7     |
 * |0 1 2 3 4 5 6 7 0 1 2 3 4 5 6 7 0 1 2 3 4 5 6 7 0 1 2 3 4 5 6 7|
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 * |       PES packet length       |           The extension     ...
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 *
 *
 * +-+-+-+-+-+-+-+-+
 * |      Byte8    |
 * |0 1 2 3 4 5 6 7|
 * +-+-+-+-+-+-+-+-+
 * ...             |
 * +-+-+-+-+-+-+-+-+
 *
 * Stream ID's which pertain to DVD
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 * |    Stream ID   |               Stream type              | ext |
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 * | 1011 1101 0xBD | Private stream 1 (non MPEG audio,      |     |
 * |                | subpictures)                           | Yes |
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 * | 1011 1110 0xBE	| Padding stream                         | No  |
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 * | 1011 1111 0xBF | Private stream 2 (navigation data)     | No  |
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 * | 110x xxxx 0xC0 | MPEG-1 or MPEG-2 audio stream number x | Yes |
 * | - 0xDF         | note: DVD allows only 8 audio streams  |     |
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 * | 1110 xxxx 0xE0 | MPEG-1 or MPEG-2 video stream number x | Yes |
 * | - 0xEF         | note: DVD allows only 1 video stream   |     |
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 *
 * The extension:
 *
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 * |                           Byte 6                              |
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 * | bit 0:     1                                                  |
 * | bit 1:     0                                                  |
 * | bit 2-3:   PES scrambling control                             |
 * | bit 4:     PES priority                                       |
 * | bit 5:     data alignment indicator                           |
 * | bit 6:     copyright                                          |
 * | bit 7:     original or copy                                   |
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 *
 *
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 * |                           Byte 7                              |
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 * | bit 0-1:   PTS DTS flags                                      |
 * | bit 2:     ESCR flag                                          |
 * | bit 3:     ES rate flag                                       |
 * | bit 4:     DSM trick mode flag                                |
 * | bit 5:     additional copy info flag                          |
 * | bit 6:     PES CRC flag                                       |
 * | bit 7:     PES extension flag                                 |
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 *
 *
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 * |                           Byte 8                              |
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 * |                    PES header data length                     |
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 *
 * The bit fields of the basic extension:
 *
 * PES scrambling control -- 00 = not scrambled, others are user defined.
 *
 * PES priority -- provides 2 priority levels, 0 and 1.
 *
 * data alignment indicator -- if set to 1 indicates that the PES packet
 * header is immediately followed by the video start code or audio syncword.
 *
 * copyright -- 1 = packet contains copyrighted material.
 *
 * original or copy -- 1 = original, 0 = copy.
 *
 * Ref: http://dvd.sourceforge.net/dvdinfo/pes-hdr.html
 */

//9 Byte
type PESHeader struct {
	packetStartCodePrefix [3]byte //[3]

	streamID byte //[1]

	PESPacketLength [2]byte //[2]

	originalOrCopy         byte //:1
	copyright              byte //:1
	dataAlignmentIndicator byte //:1
	PESPriority            byte //:1
	PESScramblingControl   byte //:2
	fixBit                 byte //:2

	PESExtensionFlag       byte //:1
	PESCRCFlag             byte //:1
	additionalCopyInfoFlag byte //:1
	DSMTrickModeFlag       byte //:1
	ESRateFlag             byte //:1
	ESCRFlag               byte //:1
	PTSDTSFlags            byte //:2

	PESHeaderDataLength byte //[1]
}

func (pesHeader *PESHeader) GetHeaderBytes() (desBytes []byte) {

	var pesHeaderData [9]byte

	/* Packet start code prefix*/
	pesHeaderData[0] = pesHeader.packetStartCodePrefix[0]
	pesHeaderData[1] = pesHeader.packetStartCodePrefix[1]
	pesHeaderData[2] = pesHeader.packetStartCodePrefix[2]

	/*Stream id：在节目流中，它规定了基本流的号码和类型。0x(C0~DF)指音频，0x(E0~EF)为视频*/
	pesHeaderData[3] = pesHeader.streamID

	/* 4~5字节为：PES包长/为0，即不限制后面ES的长度 */
	pesHeaderData[4] = pesHeader.PESPacketLength[0]
	pesHeaderData[5] = pesHeader.PESPacketLength[1]

	/* 6~7字节为：PES包头识别标识 */

	pesHeaderData[6] = byte(pesHeader.originalOrCopy |
		pesHeader.copyright<<1 |
		pesHeader.dataAlignmentIndicator<<2 |
		pesHeader.PESPriority<<3 |
		pesHeader.PESScramblingControl<<4 |
		pesHeader.fixBit<<6)

	pesHeaderData[7] = byte(pesHeader.PESExtensionFlag |
		pesHeader.PESCRCFlag<<1 |
		pesHeader.additionalCopyInfoFlag<<2 |
		pesHeader.DSMTrickModeFlag<<3 |
		pesHeader.ESRateFlag<<4 |
		pesHeader.ESCRFlag<<5 |
		pesHeader.PTSDTSFlags<<6)

	/* 8字节为：PES包头长 */
	/* 0011填充字段，表示既含有PTS，又含有DTS */
	pesHeaderData[8] = pesHeader.PESHeaderDataLength /* 可选字段和填充字段所占的字节数为10 */

	desBytes = []byte{}
	for _, value := range pesHeaderData {
		desBytes = append(desBytes, value)
	}
	return desBytes

}

/*
 * PTS DTS flags -- Presentation Time Stamp / Decode Time Stamp. 00 = no
 * PTS or DTS data present, 01 is forbidden.
 *
 * if set to 10 the following data is appended to the header data field
 * (5 Byte):
 *
 * Byte 0:
 * bit 0-3:    0010
 * bit 4-6:    PTS 32..30
 * bit 7:      1
 *
 * Byte 1-2:
 * bit 0-14    PTS 29..15
 * bit 15:     1
 *
 * Byte 3-4:
 * bit 0-14    PTS 14..00
 * bit 15:     1
 *
 * Ref: http://dvd.sourceforge.net/dvdinfo/pes-hdr.html
 */

//5 Byte
type PTSPack struct {
	markerBit byte //:1
	PTS1      byte //:3
	fixBit    byte //:4

	PTS21 byte //[1]

	markerBit1 byte //:1
	PTS22      byte //:7

	PTS31 byte //[1]

	markerBit2 byte //:1
	PTS32      byte //:7
}

func (ptsPack *PTSPack) GetPTSPackBytes() (desBytes []byte) {
	ptsPackData := [5]byte{}

	ptsPackData[0] = byte(ptsPack.markerBit |
		ptsPack.PTS1<<1 |
		ptsPack.fixBit<<4)

	ptsPackData[1] = ptsPack.PTS21

	ptsPackData[2] = byte(ptsPack.markerBit1 |
		ptsPack.PTS22<<1)

	ptsPackData[3] = ptsPack.PTS31

	ptsPackData[4] = byte(ptsPack.markerBit2 |
		ptsPack.PTS32<<1)

	desBytes = []byte{}
	for _, value := range ptsPackData {
		desBytes = append(desBytes, value)
	}
	return desBytes
}

/*
 * PTS DTS flags -- if set to 11 the following data is appended to the
 * header data field (10 Byte):
 *
 * Byte 0:
 * bit 0-3:    0011
 * bit 4-6:    PTS 32..30
 * bit 7:      1
 *
 * Byte 1-2:
 * bit 0-14    PTS 29..15
 * bit 15:     1
 *
 * Byte 3-4:
 * bit 0-14    PTS 14..00
 * bit 15:     1
 *
 * Byte 5:
 * bit 0-3:    0001
 * bit 4-6:    DTS 32..30
 * bit 7:      1
 *
 * Byte 6-7:
 * bit 0-14    DTS 29..15
 * bit 15:     1
 *
 * Byte 8-9:
 * bit 0-14    DTS 14..00
 * bit 15:     1
 *
 * Ref: http://dvd.sourceforge.net/dvdinfo/pes-hdr.html
 */

//10 Byte
type PTSDTSPack struct {
	//PTS
	PTSMarkerBit byte //:1
	PTSPTS1      byte //:3
	PTSFixBit    byte //:4

	PTSPTS21 byte //[1]

	PTSMarkerBit1 byte //:1
	PTSPTS22      byte //:7

	PTSPTS31 byte // [1]

	PTSMarkerBit2 byte //:1
	PTSPTS32      byte //:7

	//DTS
	DTSMarkerBit byte // :1
	DTSPTS1      byte // :3
	DTSFixBit    byte //:4

	DTSPTS21 byte //[1]

	DTSMarkerBit1 byte //:1
	DTSPTS22      byte //:7

	DTSPTS31 byte // [1]

	DTSMarkerBit2 byte //:1
	DTSPTS32      byte //:7
}

func (ptsDtsPack *PTSDTSPack) GetPTSDTSPackBytes() (desBytes []byte) {

	ptsDtsPackData := [10]byte{}

	/* 视频PTS */
	ptsDtsPackData[0] = byte(ptsDtsPack.PTSMarkerBit |
		ptsDtsPack.PTSPTS1<<1 |
		ptsDtsPack.PTSFixBit<<4)

	ptsDtsPackData[1] = ptsDtsPack.PTSPTS21

	ptsDtsPackData[2] = byte(ptsDtsPack.PTSMarkerBit1 |
		ptsDtsPack.PTSPTS22<<1)

	ptsDtsPackData[3] = ptsDtsPack.PTSPTS31

	ptsDtsPackData[4] = byte(ptsDtsPack.PTSMarkerBit2 |
		ptsDtsPack.PTSPTS32<<1)

	/* 视频DTS */

	ptsDtsPackData[5] = byte(ptsDtsPack.DTSMarkerBit |
		ptsDtsPack.DTSPTS1<<1 |
		ptsDtsPack.DTSFixBit<<4)

	ptsDtsPackData[6] = ptsDtsPack.DTSPTS21

	ptsDtsPackData[7] = byte(ptsDtsPack.DTSMarkerBit1 |
		ptsDtsPack.DTSPTS22<<1)

	ptsDtsPackData[8] = ptsDtsPack.DTSPTS31

	ptsDtsPackData[9] = byte(ptsDtsPack.DTSMarkerBit2 |
		ptsDtsPack.DTSPTS32<<1)

	desBytes = []byte{}
	for _, value := range ptsDtsPackData {
		desBytes = append(desBytes, value)
	}
	return desBytes
}

/*
 * Partial system header format (12 Byte):
 *
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 * |     Byte0     |     Byte1     |     Byte2     |     Byte3     |
 * |0 1 2 3 4 5 6 7 0 1 2 3 4 5 6 7 0 1 2 3 4 5 6 7 0 1 2 3 4 5 6 7|
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 * |         0000 0000 0000 0000 0000 0001         |   10111011    |
 * |                   start code                  |   Stream ID   |
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 *
 *
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 * |     Byte4     |     Byte5     |     Byte6     |     Byte7     |
 * |0 1 2 3 4 5 6 7 0 1 2 3 4 5 6 7 0 1 2 3 4 5 6 7 0 1 2 3 4 5 6 7|
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 * |          header length        |1| rate_bound1 |  rate_bound2  |
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 *
 *
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 * |     Byte8     |     Byte9     |     Byte10    |     Byte11    |
 * |0 1 2 3 4 5 6 7 0 1 2 3 4 5 6 7 0 1 2 3 4 5 6 7 0 1 2 3 4 5 6 7|
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 * | rate_bound1 |1|  audio bound  |  video bound  |    others     |
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 */

//12 Byte
type PartialSystemHeader struct {
	packetStartCodePrefix [3]byte //[3]

	streamID byte //[1]

	headerLength [2]byte //;[2]

	rateBound1 byte //:7
	markerBit1 byte //:1

	rateBound2 byte //[1]

	markerBit2 byte //:1
	rateBound3 byte //:7

	CSPSFlag   byte //:1
	fixedFlag  byte //:1
	audioBound byte //:6

	videoBound          byte //: 5
	markerBit3          byte //:1
	systemVideoLockFlag byte //:1
	systemAudioFockFlag byte //:1

	reservedByte              byte //:7
	packetRateRestrictionFlag byte //:1
}

func (systemHeader *PartialSystemHeader) GetHeaderBytes() (desBytes []byte) {

	systemHeaderData := [12]byte{}

	/* 系统头起始码前缀：0x000001BB */
	systemHeaderData[0] = systemHeader.packetStartCodePrefix[0]
	systemHeaderData[1] = systemHeader.packetStartCodePrefix[1]
	systemHeaderData[2] = systemHeader.packetStartCodePrefix[2]
	systemHeaderData[3] = systemHeader.streamID

	/* 4~5字节为：该字段后的系统头长度 */

	systemHeaderData[4] = systemHeader.headerLength[0]
	systemHeaderData[5] = systemHeader.headerLength[1]

	/* rate_bound：0x0F7F */
	/* rate_bound为大于或等于在任意节目流包中编码的program_mux_rate字段的最大值的整数值 */

	systemHeaderData[6] = byte(systemHeader.rateBound1 |
		systemHeader.markerBit1<<7)

	systemHeaderData[7] = systemHeader.rateBound2

	systemHeaderData[8] = byte(systemHeader.markerBit2 |
		systemHeader.rateBound3<<1)

	/* audio_bound：0x3F */
	/* audio_bound为大于或等于ISO/IEC 13818-3和ISO/IEC 11172-3音频流最大数的整数值 */
	/* fixed_flag：1 */
	/* fixed_flag为1时表示固定的比特速率操作，为0时表示可变比特速率操作 */
	/* CSPS_flag：0 */
	/* CSPS_flag为1表示节目流满足2.7.9中规定的限制 */
	systemHeaderData[9] = byte(systemHeader.CSPSFlag |
		systemHeader.fixedFlag<<1 |
		systemHeader.audioBound<<2)

	/* system_audio_lock_flag：1 */
	/* system_audio_lock_flag为1表示音频采样速率和系统目标解码器的system_clock_frequency之间存在特定的常量比率关系 */
	/* system_video_lock_flag：1 */
	/* system_video_lock_flag为1表示视频时间基和系统目标解码器的系统时钟频率之间存在特定的常量比率关系 */
	/* video_bound：1 */
	/* 在解码过程同时被激活的节目流中，video_bound被设置为大于或等于视流的最大数的整数值 */

	systemHeaderData[10] = byte(systemHeader.videoBound |
		systemHeader.markerBit3<<5 |
		systemHeader.systemVideoLockFlag<<6 |
		systemHeader.systemAudioFockFlag<<7)

	/* packet_rate_restriction_flag：0 */
	/* 若CSPS标志设置为1，则packet_rate_restriction_flag指示适用于该包速率的那些限制，如2.7.9 中所指定的 */
	/* 若CSPS标志设置为0，则packet_rate_restriction_flag的含义未确定 */
	systemHeaderData[11] = byte(systemHeader.reservedByte |
		systemHeader.packetRateRestrictionFlag<<7)

	desBytes = []byte{}
	for _, value := range systemHeaderData {
		desBytes = append(desBytes, value)
	}
	return desBytes
}

//PartialSystemStreamMessage 3 Byte
type PartialSystemStreamMessage struct {
	NstreamID byte //[1]

	PSTDBufferScaleBound1 byte //:5
	PSTDBufferBoundScale  byte //:1
	MarkerBit3            byte //:2

	PSTDBufferScaleBound2 byte //[1]
}

func (psSystemStreamMessage *PartialSystemStreamMessage) GetPartialSystemStreamMessageBytes() (desBytes []byte) {

	psSystemStreamMessageData := [3]byte{}

	psSystemStreamMessageData[0] = psSystemStreamMessage.NstreamID

	psSystemStreamMessageData[1] = byte(psSystemStreamMessage.PSTDBufferScaleBound1 |
		psSystemStreamMessage.PSTDBufferBoundScale<<5 |
		psSystemStreamMessage.MarkerBit3<<6)

	psSystemStreamMessageData[2] = psSystemStreamMessage.PSTDBufferScaleBound2

	desBytes = []byte{}
	for _, value := range psSystemStreamMessageData {
		desBytes = append(desBytes, value)
	}
	return desBytes

}

//PSMap 10 Byte
type PSMapHeader struct {
	packetStartCodePrefix [3]byte //[3]

	mapStreamID byte //[1]

	programStreamMapLength [2]byte //[2]

	programStreamMapVersion byte //:5
	reserved1               byte //:2
	currentNextIndicator    byte //:1

	markerBit byte //:1
	reserved2 byte //:7

	programStreamInfoLength [2]byte //[2]
}

func (psMapHeader *PSMapHeader) GetHeaderBytes() (desBytes []byte) {
	psMapHeaderData := [10]byte{}

	/* 节目流映射起始码前缀：0x000001BC */
	psMapHeaderData[0] = psMapHeader.packetStartCodePrefix[0]
	psMapHeaderData[1] = psMapHeader.packetStartCodePrefix[1]
	psMapHeaderData[2] = psMapHeader.packetStartCodePrefix[2]
	psMapHeaderData[3] = psMapHeader.mapStreamID

	/* 4~5字节为：该字段后的节目流映射长度 */
	psMapHeaderData[4] = psMapHeader.programStreamMapLength[0]
	psMapHeaderData[5] = psMapHeader.programStreamMapLength[1]

	/* current_next_indicator：1 */
	/* 当current_next_indicator为1时表示发送的节目流映射为当前有效，为0时表示发送的节目流映射尚未有效并且下一个节目流映射表将生效 */
	/* program_stream_map_version：1*/
	/* 当current_next_indicator为1时，program_stream_map_version是当前有效的节目流映射的版本 */
	psMapHeaderData[6] = byte(psMapHeader.programStreamMapVersion |
		psMapHeader.reserved1<<5 |
		psMapHeader.currentNextIndicator<<7)

	psMapHeaderData[7] = byte(psMapHeader.markerBit |
		psMapHeader.reserved2<<1)

	/* program_stream_info_length：0*/
	/* program_stream_info_length表示紧随此字段的描述符的总长 */
	psMapHeaderData[8] = psMapHeader.programStreamInfoLength[0]
	psMapHeaderData[9] = psMapHeader.programStreamInfoLength[1]

	desBytes = []byte{}
	for _, value := range psMapHeaderData {
		desBytes = append(desBytes, value)
	}
	return desBytes
}

//PSMapStream 4 Byte
type PSMapStream struct {
	streamType                 byte    //[1];
	elementaryStreamID         byte    //[1];
	elementaryStreamInfoLength [2]byte //[2];
}

func (psMapStream *PSMapStream) GetPSMapStreamBytes() (desBytes []byte) {
	psMapStreamData := [4]byte{}

	psMapStreamData[0] = psMapStream.streamType

	psMapStreamData[1] = psMapStream.elementaryStreamID

	psMapStreamData[2] = psMapStream.elementaryStreamInfoLength[0]
	psMapStreamData[3] = psMapStream.elementaryStreamInfoLength[1]

	desBytes = []byte{}
	for _, value := range psMapStreamData {
		desBytes = append(desBytes, value)
	}
	return desBytes
}
