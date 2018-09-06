package gb28181

import (
	"strings"
)

func ParseRegister1(response string) (realm, nonce, authMethod string, err error) {

	pos := strings.Index(response, " ")
	pos1 := strings.Index(response[pos+1:], " ")
	//	println(pos, pos+pos1+1)
//	errCode = response[pos+1 : pos+pos1+1]
	//	println(errCode)

	pos = strings.Index(response, "Authenticate")
	temp := response[pos:]

	templen := len("realm=\"")
	pos = strings.Index(temp, "realm=\"")
	pos1 = strings.Index(temp[pos+templen:], "\"")
	realm = temp[pos+templen : pos+templen+pos1]
	//	println(realm)

	templen = len("nonce=\"")
	pos = strings.Index(temp, "nonce=\"")
	pos1 = strings.Index(temp[pos+templen:], "\"")
	nonce = temp[pos+templen : pos+templen+pos1]
	//	println(nonce)

	templen = len("algorithm=")
	pos = strings.Index(temp, "algorithm=")
	pos1 = strings.Index(temp[pos+templen:], ",")
	authMethod = temp[pos+templen : pos+templen+pos1]
	//	println(authMethod)

	err = nil

	return
}

func ParseRegister2(response string) (errCode string, err error) {

	pos := strings.Index(response, " ")
	pos1 := strings.Index(response[pos+1:], " ")
	errCode = response[pos+1 : pos+pos1+1]

	err = nil

	return
}

func ParseResponseHead(response string) (errCode string, err error) {

	pos := strings.Index(response, " ")
	pos1 := strings.Index(response[pos+1:], " ")
	errCode = response[pos+1 : pos+pos1+1]

	err = nil

	return
}

func MatchResponse(key, response string) bool {
	pos := strings.Index(response, "Call-ID:")
	pos1 := strings.Index(response[pos:], "\r\n")
//	callid := response[pos+8:pos+pos1]
//	println(callid)

	pos = strings.Index(response, "CSeq:")
	pos1 = strings.Index(response[pos:], "\r\n")
	cseq := response[pos+5:pos+pos1]
//	println(cseq)

	if key == "INVITE" && ( cseq == "1 INVITE" || cseq == " 1 INVITE") {
		return true
	} else if key == "ACK" && ( cseq == "2 ACK" || cseq == " 2 ACK") {
		return true	
	} else if key == "REGISTER1" && ( cseq == "1 REGISTER" || cseq == " 1 REGISTER") {
		return true	
	} else if key == "REGISTER2" && ( cseq == "2 REGISTER" || cseq == " 2 REGISTER") {
		return true	
	}

	return false
}

func ParseResponseInvite(response string) (mediaip, mediaport, ssrc, totag string, err error) {

	/*	str := "SIP/2.0 200 OK\r\n" +
			"Via: SIP/2.0/UDP 192.168.1.71:5069;branch=z9hG4bK-d8754z-51060901bcee1b5b-1---d8754z-;rport=5069\r\n" +
			"Contact: <sip:11000000001320000011@192.168.1.71:5060>\r\n" +
			"To: <sip:11000000001320000011@1100000000>;tag=2108aa2c\r\n" +
			"From: <sip:15010000004000000001@1100000000>;tag=ca37f475\r\n" +
			"Call-ID: YTcwZDYzM2ZmZjE0Y2Y5MzFjYWUxMmVkNGNlNzZjMjg@192.168.1.71\r\n" +
			"CSeq: 1 INVITE\r\n" +
			"Allow: REGISTER, INVITE, MESSAGE, ACK, BYE, CANCEL, INFO, SUBSCRIBE, NOTIFY\r\n" +
			"Content-Type: Application/sdp\r\n" +
			"User-Agent: NetPosa\r\n" +
			"Content-Length: 166\r\n" +
			"\r\n" +
			"v=0\r\n" +
			"o=11000000001320000011 0 0 IN IP4 192.168.4.114\r\n" +
			"s=Play\r\n" +
			"c=IN IP4 192.168.4.114\r\n" +
			"t=0 0\r\n" +
			"m=video 9708 RTP/AVP 96\r\n" +
			"a=sendonly\r\n" +
			"a=rtpmap:96 PS/90000\r\n" +
			"y=2147483647\r\n"

		response = str
	*/
	pos := strings.Index(response, "To:")
	temp := response[pos+1:]
	pos = strings.Index(temp, "tag=")
	pos1 := strings.Index(temp[pos:], "\r\n")
//	println(pos, pos1)
	totag = temp[pos+4 : pos+pos1]
//	println(totag)

	pos = strings.Index(response, "v=0")
	temp = response[pos+1:]
	//	println(pos)

	pos = strings.Index(temp, "c=IN IP4 ")
	pos1 = strings.Index(temp[pos+9:], "\r\n")
	if pos1 < 0 {
		pos1 = strings.Index(temp[pos+9:], "\n")		
	}
//	println( pos, pos1)
	mediaip = temp[pos+9 : pos+9+pos1]
//	println(mediaip)

	pos = strings.Index(temp, "m=video ")
	pos1 = strings.Index(temp[pos+8:], " ")
	mediaport = temp[pos+8 : pos+8+pos1]
//	println(mediaport)

	pos = strings.Index(temp, "y=")
	pos1 = strings.Index(temp[pos+2:], "\r\n")
	if pos1 < 0 {
		pos1 = strings.Index(temp[pos+9:], "\n")		
	}
//	println(pos, pos1)
	ssrc = temp[pos+2 : pos+2+pos1]
//	println(ssrc)

	err = nil

	return
}
