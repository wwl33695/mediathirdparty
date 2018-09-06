package gb28181

import (
	"crypto/md5"
	"fmt"
	"math/rand"
	"strconv"
)

type UASInfo struct {
	ServerID        string
	ServerIP        string
	ServerPort      string
	UserName        string
	Password        string
	ClientIP        string
	ClientPort      string
	BranchID        string
	CallID          string
	Tag             string
	RemoteMediaPort string
	RemoteMediaIP string
	LocalMediaPort  string
	LocalSSRC       string
	RemoteSSRC      string
	PlayFromTag     string
	PlayToTag       string
	PlayCallID      string
	ChannelID 		string
}

type SipHeader struct {
	Request string
	From    string
	Via     string
	To      string
	Contact string
}

func (this *SipHeader) ToString() string {
	return this.Via + this.To + this.From + this.Contact
}

func (this *UASInfo) BuildSipHeaders(target, fromtag, totag string) (headers SipHeader) {

	headers.Via = fmt.Sprintf("Via: SIP/2.0/UDP %s:%s;branch=%s;rport\r\n", this.ClientIP, this.ClientPort, this.BranchID)
	if totag == "" {
		headers.To = fmt.Sprintf("To: <sip:%s@%s>\r\n", target, this.ServerID[:10])
	} else {
		headers.To = fmt.Sprintf("To: <sip:%s@%s>;tag=%s\r\n", target, this.ServerID[:10], totag)
	}
	if fromtag == "" {
		headers.From = fmt.Sprintf("From: <sip:%s@%s>;tag=%s\r\n", this.UserName, this.ServerID[:10], this.Tag)
	} else {
		headers.From = fmt.Sprintf("From: <sip:%s@%s>;tag=%s\r\n", this.UserName, this.ServerID[:10], fromtag)
	}
	headers.Contact = fmt.Sprintf("Contact: <sip:%s@%s:%s>\r\n", this.UserName, this.ClientIP, this.ClientPort)

	return
}

func (this *UASInfo) BuildRegisterRequest() (request string) {

	this.BranchID = "z9hG4bK" + strconv.FormatUint(uint64(rand.Uint32()), 16)
	this.Tag = strconv.FormatUint(uint64(rand.Uint32()), 10)

	headers := this.BuildSipHeaders(this.UserName, "", "")
	this.CallID = fmt.Sprintf("%s@%s", this.GetResponse(headers.From, headers.Via, headers.To), this.ServerIP)
	CallIDField := fmt.Sprintf("Call-ID: %s\r\n", this.CallID)
	requestField := fmt.Sprintf("REGISTER sip:%s SIP/2.0\r\n", this.ServerID[:10])

	request = requestField + headers.ToString() +
		CallIDField +
		"Max-Forwards: 70\r\n" +
		"CSeq: 1 REGISTER\r\n" +
		"Expires: 3600\r\n" +
		"Allow: REGISTER, INVITE, MESSAGE, ACK, BYE, CANCEL, INFO, SUBSCRIBE, NOTIFY\r\n" +
		"User-Agent: NPSIPSDK\r\n" +
		"Content-Length: 0\r\n" +
		"\r\n"

	return request
}

func (this *UASInfo) BuildRegisterMD5Auth(realm, nonce string) (request string) {

	response := this.GetResponse(realm, nonce, "REGISTER")

	this.BranchID = "z9hG4bK" + strconv.FormatUint(uint64(rand.Uint32()), 16)

	authentication := fmt.Sprintf("Authorization: Digest username=\"%s\",realm=\"%s\",nonce=\"%s\",uri=\"sip:%s\",response=\"%s\",algorithm=MD5\r\n",
		this.UserName, realm, nonce, this.ServerID[:10], response)

	CallIDField := fmt.Sprintf("Call-ID: %s\r\n", this.CallID)
	requestField := fmt.Sprintf("REGISTER sip:%s SIP/2.0\r\n", this.ServerID[:10])

	headers := this.BuildSipHeaders(this.UserName, "", "")
	request = requestField + headers.ToString() +
		CallIDField +
		"Max-Forwards: 70\r\n" +
		"CSeq: 2 REGISTER\r\n" +
		"Expires: 3600\r\n" +
		"Allow: REGISTER, INVITE, MESSAGE, ACK, BYE, CANCEL, INFO, SUBSCRIBE, NOTIFY\r\n" +
		"User-Agent: NPSIPSDK\r\n" +
		authentication +
		"Content-Length: 0\r\n" +
		"\r\n"

	return
}

func (this *UASInfo) GetResponse(realm, nonce, method string) string {
	//	realm := "1100000000"
	//	nonce := "13133944849:f251436a279f25d0b879d2813df6b5b2"
	//	method := "REGISTER"
	//	println(realm, nonce, method)

	hUsernameRealmPassword := fmt.Sprintf("%x", md5.Sum([]byte(this.UserName+":"+realm+":"+this.Password)))
	//	fmt.Printf("hUsernameRealmPassword=%s \n", hUsernameRealmPassword)

	hMethodUri := fmt.Sprintf("%x", md5.Sum([]byte(method+":sip:"+this.ServerID[:10])))
	//	fmt.Printf("hMethodUri=%s \n", hMethodUri)

	//username:realm:password:nonce:method:uri
	response := fmt.Sprintf("%x", md5.Sum([]byte(hUsernameRealmPassword+":"+nonce+":"+hMethodUri)))
	//	fmt.Printf("%s \n", response)

	return response
}

func (this *UASInfo) BuildHeartbeat() (request string) {

	this.BranchID = "z9hG4bK" + strconv.FormatUint(uint64(rand.Uint32()), 16)
	this.Tag = strconv.FormatUint(uint64(rand.Uint32()), 10)

	headers := this.BuildSipHeaders(this.UserName, "", "")
	this.CallID = fmt.Sprintf("%s@%s", this.GetResponse(headers.From, headers.Via, headers.To), this.ServerIP)
	CallIDField := fmt.Sprintf("Call-ID: %s\r\n", this.CallID)
	requestField := fmt.Sprintf("MESSAGE sip:%s@%s SIP/2.0\r\n", this.ServerID, this.ServerID[:10])

	request = requestField + headers.ToString() +
		CallIDField +
		"Max-Forwards: 70\r\n" +
		"CSeq: 2 MESSAGE\r\n" +
		"Expires: 3600\r\n" +
		"Allow: REGISTER, INVITE, MESSAGE, ACK, BYE, CANCEL, INFO, SUBSCRIBE, NOTIFY\r\n" +
		"User-Agent: NPSIPSDK\r\n" +
		"Content-Type: Application/MANSCDP+xml\r\n" +
		"Content-Length: 150\r\n" +
		"\r\n" +
		"<?xml version=\"1.0\"?>\r\n" +
		"<Notify>\r\n" +
		"<CmdType>Keepalive</CmdType>\r\n" +
		"<SN>8</SN>\r\n" +
		fmt.Sprintf("<DeviceID>%s</DeviceID>\r\n", this.ServerID) +
		"<Status>OK</Status>\r\n" +
		"</Notify>\r\n"

	return request
}

func (this *UASInfo) BuildQueryDeviceRequest() (request string) {
	this.BranchID = "z9hG4bK" + strconv.FormatUint(uint64(rand.Uint32()), 16)
	this.Tag = strconv.FormatUint(uint64(rand.Uint32()), 10)

	headers := this.BuildSipHeaders(this.UserName, "", "")
	this.CallID = fmt.Sprintf("%s@%s", this.GetResponse(headers.From, headers.Via, headers.To), this.ServerIP)
	CallIDField := fmt.Sprintf("Call-ID: %s\r\n", this.CallID)
	requestField := fmt.Sprintf("MESSAGE sip:%s@%s SIP/2.0\r\n", this.ServerID, this.ServerID[:10])

	request = requestField + headers.ToString() +
		CallIDField +
		"Max-Forwards: 70\r\n" +
		"CSeq: 2 MESSAGE\r\n" +
		"Expires: 3600\r\n" +
		"Allow: REGISTER, INVITE, MESSAGE, ACK, BYE, CANCEL, INFO, SUBSCRIBE, NOTIFY\r\n" +
		"User-Agent: NPSIPSDK\r\n" +
		"Content-Type: Application/MANSCDP+xml\r\n" +
		"Content-Length: 125\r\n" +
		"\r\n" +
		"<?xml version=\"1.0\"?>\r\n" +
		"<Query>\r\n" +
		"<CmdType>Catalog</CmdType>\r\n" +
		"<SN>9</SN>\r\n" +
		fmt.Sprintf("<DeviceID>%s</DeviceID>\r\n", this.ServerID) +
		"</Query>\r\n"

	return request
}

func (this *UASInfo) BuildInviteRequest(rtpport string, channelid string) (request string) {

	this.BranchID = "z9hG4bK" + strconv.FormatUint(uint64(rand.Uint32()), 16)
	this.Tag = strconv.FormatUint(uint64(rand.Uint32()), 10)
	this.PlayFromTag = this.Tag

	headers := this.BuildSipHeaders(channelid, "", "")
	this.CallID = fmt.Sprintf("%s@%s", this.GetResponse(headers.From, headers.Via, headers.To), this.ServerIP)
	this.PlayCallID = this.CallID
	CallIDField := fmt.Sprintf("Call-ID: %s\r\n", this.CallID)
	requestField := fmt.Sprintf("INVITE sip:%s@%s SIP/2.0\r\n", channelid, this.ServerID[:10])
	this.LocalSSRC = strconv.FormatUint(uint64(rand.Uint32()), 10)
	//	subjectField := fmt.Sprintf("Subject: %s:%s,%s:%s\r\n", channelid, mediaStreamID[:2], this.UserName, mediaStreamID[:2])
	subjectField := fmt.Sprintf("Subject: %s:%d,%s:%d\r\n", channelid, 1, this.UserName, 1)

	sdpPayload := this.BuildSDPRequest(rtpport, this.LocalSSRC)
	request = requestField + headers.ToString() +
		CallIDField +
		subjectField +
		"Max-Forwards: 70\r\n" +
		"CSeq: 1 INVITE\r\n" +
		"Allow: REGISTER, INVITE, MESSAGE, ACK, BYE, CANCEL, INFO, SUBSCRIBE, NOTIFY\r\n" +
		"User-Agent: NPSIPSDK\r\n" +
		"Content-Type: application/sdp\r\n" +
		fmt.Sprintf("Content-Length: %d\r\n", len(sdpPayload)) +
		"\r\n" +
		sdpPayload

	return request
}

func (this *UASInfo) BuildSDPRequest(rtpport string, ssrc string) (request string) {

	request = "v=0\r\n" +
		fmt.Sprintf("o=%s 0 0 IN IP4 %s\r\n", this.UserName, this.ClientIP) +
		"s=Play\r\n" +
		fmt.Sprintf("c=IN IP4 %s\r\n", this.ClientIP) +
		"t=0 0\r\n" +
		fmt.Sprintf("m=video %s RTP/AVP 96 98 97 \r\n", rtpport) +
		"a=recvonly\r\n" +
		"a=rtpmap:96 PS/90000\r\n" +
		"a=rtpmap:98 H264/90000\r\n" +
		"a=rtpmap:97 MPEG4/90000\r\n" +
		fmt.Sprintf("y=%s\r\n", ssrc) +
		"f=\r\n"

	return
}

func (this *UASInfo) BuildACKRequest(channelid string) (request string) {

	this.BranchID = "z9hG4bK" + strconv.FormatUint(uint64(rand.Uint32()), 16)

	CallIDField := fmt.Sprintf("Call-ID: %s\r\n", this.PlayCallID)
	requestField := fmt.Sprintf("ACK sip:%s@%s SIP/2.0\r\n", channelid, this.ServerID[:10])

	headers := this.BuildSipHeaders(channelid, this.PlayFromTag, this.PlayToTag)
	request = requestField + headers.ToString() +
		CallIDField +
		"Max-Forwards: 70\r\n" +
		"CSeq: 1 ACK\r\n" +
		"Allow: REGISTER, INVITE, MESSAGE, ACK, BYE, CANCEL, INFO, SUBSCRIBE, NOTIFY\r\n" +
		"User-Agent: NPSIPSDK\r\n" +
		"Content-Length: 0\r\n" +
		"\r\n"

	return
}

func (this *UASInfo) BuildBYERequest(channelid string) (request string) {

	this.BranchID = "z9hG4bK" + strconv.FormatUint(uint64(rand.Uint32()), 16)

	CallIDField := fmt.Sprintf("Call-ID: %s\r\n", this.PlayCallID)
	requestField := fmt.Sprintf("BYE sip:%s@%s SIP/2.0\r\n", channelid, this.ServerID[:10])

	headers := this.BuildSipHeaders(channelid, this.PlayFromTag, this.PlayToTag)
	request = requestField + headers.ToString() +
		CallIDField +
		"Max-Forwards: 70\r\n" +
		"CSeq: 2 BYE\r\n" +
		"Allow: REGISTER, INVITE, MESSAGE, ACK, BYE, CANCEL, INFO, SUBSCRIBE, NOTIFY\r\n" +
		"User-Agent: NPSIPSDK\r\n" +
		"Reason: SIP;description=\"User Hung Up\"\r\n" +
		"Content-Length: 0\r\n" +
		"\r\n"

	return
}
