package dmic

import (
	"strings"
	"encoding/base64"
)

func MatchRequest(key, response string) bool {
	pos := strings.Index(response, "CSeq:")
	pos1 := strings.Index(response[pos:], "\r\n")
	cseq := response[pos+5:pos+pos1]
//	println(cseq)

	if key == "PUSH" && ( cseq == "1 PUSH" || cseq == " 1 PUSH") {
		return true
	} else if key == "PULL" && ( cseq == "1 PULL" || cseq == " 1 PULL") {
		return true
	} else if key == "PLAY" && ( cseq == "2 PLAY" || cseq == " 2 PLAY") {
		return true	
	} else if key == "REGISTER1" && ( cseq == "1 REGISTER" || cseq == " 1 REGISTER") {
		return true	
	} else if key == "REGISTER2" && ( cseq == "2 REGISTER" || cseq == " 2 REGISTER") {
		return true	
	}

	return false
}

func ParseRegister1Request(response string) (userid string) {
	pos := strings.Index(response, "UserID: ")
	pos1 := strings.Index(response[pos:], "\r\n")
	userid = response[pos+8:pos+pos1]
//	println(userid)

	return 
}

func ParseRegister2Request(response string) (userid, username, password string) {
	pos := strings.Index(response, "UserID: ")
	pos1 := strings.Index(response[pos:], "\r\n")
	userid = response[pos+8:pos+pos1]
//	println(userid)

	pos = strings.Index(response, "Authorization:Basic ")
	pos1 = strings.Index(response[pos:], "\r\n")
	base64Data := response[pos+20:pos+pos1]
//	println(base64Data)

	decodebytes, _ := base64.StdEncoding.DecodeString(base64Data)
	strTemp := string(decodebytes[0:])
//	println(strTemp)

	pos = strings.Index(strTemp, ":")
	if pos < 0 {
		println("parse error")
		return "", "", ""
	}
	username = strTemp[0:pos]
	password = strTemp[pos+1:]
//	println(username, password)

	return 
}

func ParseRequestStreamID(response string) (streamid string) {
	pos := strings.Index(response, "//")
	pos1 := strings.Index(response[pos+2:], "/")
	pos1 += pos + 2
	pos = strings.Index(response[pos1:], "?")
	if pos <= 0 {
		pos = strings.Index(response[pos1:], " ")
	}
	streamid = response[pos1+1:pos+pos1]
//	println(streamid)

	return 
}

func ParseRequestURI(response string) (uri string) {
	pos := strings.Index(response, " ")
	pos1 := strings.Index(response[pos+1:], " ")
	uri = response[pos+1:pos+pos1+1]
//	println(streamid)

	return 
}
