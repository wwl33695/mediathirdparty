package dmic

import (
	"fmt"
	"encoding/base64"
	"encoding/json"
)

func (this *DmicProto) GetRegister1Request() string {
	request := fmt.Sprintf("REGISTER dmic://%s:%s DMIC/1.0\r\n", this.ServerIP, this.ServerPort)

	request += 	"CSeq: 1 REGISTER\r\n" +
				"Expires: 3600\r\n"

	useridfield := fmt.Sprintf("UserID: %s\r\n", this.UserID)
	request += useridfield

	request += "\r\n"
	return request
}

func (this *DmicProto) GetRegister2Request() string {
	request := fmt.Sprintf("REGISTER dmic://%s:%s DMIC/1.0\r\n", this.ServerIP, this.ServerPort)

	request += 	"CSeq: 2 REGISTER\r\n" +
				"Expires: 3600\r\n"

	useridField := fmt.Sprintf("UserID: %s\r\n", this.UserID)
	request += useridField

	encodeString := base64.StdEncoding.EncodeToString([]byte(this.UserName+":"+this.Password) )

	authorizationField := fmt.Sprintf("Authorization:Basic %s\r\n",  encodeString)
	request += authorizationField

	request += "\r\n"
	return request
}

func (this *DmicProto) GetPushRequest(streamid, streamtype string, channels interface{}) string {
	bs, err := json.Marshal(channels)
	if err != nil {
		return ""
	}

	request := fmt.Sprintf("PUSH dmic://%s:%s/%s?streamtype=%s DMIC/1.0\r\n", 
					this.ServerIP, this.ServerPort, streamid, streamtype)

	request += 	"CSeq: 1 PUSH\r\n" +
				"Content-Type:aplication/json\r\n"

	useridField := fmt.Sprintf("UserID: %s\r\n", this.UserID)
	request += useridField
	
	body := string(bs[0:])
	contentLengthField := fmt.Sprintf("Content-Length: %d\r\n", len(body))
	request += contentLengthField

	request += "\r\n"
	request += body

	return request
}

func (this *DmicProto) GetPullRequest(streamid string) string {
	request := fmt.Sprintf("PULL dmic://%s:%s/%s DMIC/1.0\r\n", this.ServerIP, this.ServerPort, streamid)

	request += 	"CSeq: 1 PULL\r\n" +
				"Accept: application/json\r\n"

	useridField := fmt.Sprintf("UserID: %s\r\n", this.UserID)
	request += useridField


	request += "\r\n"
	return request
}

func (this *DmicProto) GetPlayRequest(streamid string) string {
	request := fmt.Sprintf("PLAY dmic://%s:%s/%s DMIC/1.0\r\n", this.ServerIP, this.ServerPort, streamid)

	request += 	"CSeq: 2 PLAY\r\n"

	sessionidField := fmt.Sprintf("SessionID: %s\r\n", this.SessionID)
	request += sessionidField

	request += "\r\n"
	return request
}

func (this *DmicProto) GetByeRequest(streamid string) string {
	request := fmt.Sprintf("BYE dmic://%s:%s/%s DMIC/1.0\r\n", this.ServerIP, this.ServerPort, streamid)

	request += 	"CSeq: 2 BYE\r\n"

	sessionidField := fmt.Sprintf("SessionID: %s\r\n", this.SessionID)
	request += sessionidField

	request += "\r\n"
	return request
}

func GetForceIRequest(remoteaddr, streamid, sessionid string) string {
	request := fmt.Sprintf("FORCEI dmic://%s/%s DMIC/1.0\r\n", remoteaddr, streamid)

	request += 	"CSeq: 2 FORCEI\r\n"

	sessionidField := fmt.Sprintf("SessionID: %s\r\n", sessionid)
	request += sessionidField

	request += "\r\n"
	return request
}