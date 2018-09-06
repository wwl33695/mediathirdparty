package dmic

import (
	"encoding/json"
	"fmt"
)

func GetResponse(userid, sessionid, cseq string, statuscode int, channels interface{}) string {
	var head string
	if statuscode == 401 {
		head = "DMIC/1.0 401 Unauthorized\r\n" +
				"WWW-Authenticate： Basic realm=\"WallyWorld\"\r\n"
	} else if statuscode == 200 {
		head = "DMIC/1.0 200 OK \r\n"		
	} else if statuscode == 400 {
		head = "DMIC/1.0 400 Bad Request \r\n"		
	} else if statuscode == 404 {
		head = "DMIC/1.0 404 Not Found \r\n"		
	} else if statuscode == 405 {
		head = "DMIC/1.0 405 Method Not Allowed \r\n"		
	} else if statuscode == 406 {
		head = "DMIC/1.0 406 Not Acceptable \r\n"		
	} else if statuscode == 451 {
		head = "DMIC/1.0 451 Invalid Parameter \r\n"		
	} else {
		println("unknown statuscode")
		return ""
	}

//	var cseq string
	if cseq == "REGISTER1" {
		cseq = "CSeq: 1 REGISTER\r\n"
	} else if cseq == "REGISTER2" {
		cseq = "CSeq: 2 REGISTER\r\n"		
	} else if cseq == "PUSH" {
		cseq = "CSeq: 1 PUSH\r\n"		
	} else if cseq == "PULL" {
		cseq = "CSeq: 1 PULL\r\n"		
	} else if cseq == "PLAY" {
		cseq = "CSeq: 2 PLAY\r\n"		
	} else if cseq == "BYE" {
		cseq = "CSeq: 2 BYE\r\n"		
	} else {
		println("unknown cseq mark")
		return ""		
	}

	var userID string
	if userid != "" {
		userID = "UserID: " + userid + "\r\n"		
	}

	var sessionID string
	if sessionid != "" {
		sessionID ="SessionID: " + sessionid + "\r\n"
	}

	request := head +
			cseq +
			"Server: MServer 1.1.0\r\n"

	if userID != "" {
		request += userID
	}
	if sessionID != "" {
		request += sessionID
	}

	if channels != nil {
		bs, err := json.Marshal(channels)
		if err != nil {
			return ""
		}		
		body := string(bs[0:])
		contentLengthField := fmt.Sprintf("Content-Length: %d\r\n", len(body))
		request += "Content-Type:aplication/json\r\n"
		request += contentLengthField

		request += "\r\n"
		request += body
		return request
	}

	request += "\r\n"
	return request
}
