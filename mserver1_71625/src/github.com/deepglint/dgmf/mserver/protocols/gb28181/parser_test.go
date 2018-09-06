package gb28181

import (
	"testing"
)

func Test_ParseRegister1(t *testing.T) {
	response := "SIP/2.0 401 Unauthorized\r\n" +
				"To: <sip:34020000001320000001@192.168.10.177>;tag=82283543_53173353_1041d6a9-d495-4dcf-9359-26d7797910d5\r\n" +
				"Via: SIP/2.0/UDP 192.168.10.177:5061;rport=5061;branch=z9hG4bK3675466093;received=192.168.10.177\r\n" +
				"CSeq: 1 REGISTER\r\n" +
				"Call-ID: 423467571\r\n" +
				"From: <sip:34020000001320000001@192.168.10.177>;tag=4210464813\r\n" +
				"WWW-Authenticate: Digest realm=\"3402000000\",nonce=\"8304903d8454ef91\"\r\n" +
				"Content-Length: 0\r\n\r\n"

	realm, nonce, _, _ := ParseRegister1(response)
	if realm == "" || nonce == "" {
		t.Errorf("ParseRegister1 failed")
	}
}

func Test_ParseRegister2(t *testing.T) {
	response := "SIP/2.0 200 OK\r\n" +
				"To: <sip:34020000001320000001@192.168.10.177>;tag=40411091_53173353_4aeafef2-8b2d-44a5-aac8-821490e15653\r\n" +
				"Via: SIP/2.0/UDP 192.168.10.177:5061;rport=5061;branch=z9hG4bK51426462;received=192.168.10.177\r\n" +
				"CSeq: 2 REGISTER\r\n" +
				"Call-ID: 423467571\r\n" +
				"From: <sip:34020000001320000001@192.168.10.177>;tag=4210464813\r\n" +
				"Contact: <sip:34020000001320000001@192.168.10.177:5061;line=ac42de468548252>\r\n" +
				"Expires: 3600\r\n" +
				"Date: 2013-06-26T11:21:50.434\r\n" +
				"Content-Length: 0\r\n\r\n"

	errorCode, _ := ParseRegister2(response)
	if errorCode == "" {
		t.Errorf("ParseRegister2 failed")
	}
}

func Test_ParseResponseHead(t *testing.T) {
	response := "SIP/2.0 401 Unauthorized\r\n" +
				"To: <sip:34020000001320000001@192.168.10.177>;tag=82283543_53173353_1041d6a9-d495-4dcf-9359-26d7797910d5\r\n" +
				"Via: SIP/2.0/UDP 192.168.10.177:5061;rport=5061;branch=z9hG4bK3675466093;received=192.168.10.177\r\n" +
				"CSeq: 1 REGISTER\r\n" +
				"Call-ID: 423467571\r\n" +
				"From: <sip:34020000001320000001@192.168.10.177>;tag=4210464813\r\n" +
				"WWW-Authenticate: Digest realm=\"3402000000\",nonce=\"8304903d8454ef91\"\r\n" +
				"Content-Length: 0\r\n\r\n"

	errorCode, _ := ParseResponseHead(response)
	if errorCode == "" {
		t.Errorf("ParseResponseHead failed")
	}
}

func Test_MatchResponse(t *testing.T) {
	response := "SIP/2.0 401 Unauthorized\r\n" +
				"To: <sip:34020000001320000001@192.168.10.177>;tag=82283543_53173353_1041d6a9-d495-4dcf-9359-26d7797910d5\r\n" +
				"Via: SIP/2.0/UDP 192.168.10.177:5061;rport=5061;branch=z9hG4bK3675466093;received=192.168.10.177\r\n" +
				"CSeq: 1 REGISTER\r\n" +
				"Call-ID: 423467571\r\n" +
				"From: <sip:34020000001320000001@192.168.10.177>;tag=4210464813\r\n" +
				"WWW-Authenticate: Digest realm=\"3402000000\",nonce=\"8304903d8454ef91\"\r\n" +
				"Content-Length: 0\r\n\r\n"

	match := MatchResponse("REGISTER1", response)
	if !match {
		t.Errorf("MatchResponse REGISTER1 failed")
	}
}

func Test_ParseResponseInvite(t *testing.T) {
	response := "SIP/2.0 200 OK\r\n" +
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

	mediaip, mediaport, ssrc, totag, _ := ParseResponseInvite(response)
	if mediaip == "" || mediaport == "" || ssrc == "" || totag == "" {
		t.Errorf("ParseResponseInvite failed")
	}
}

