package rtsp

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

/*
6.1 Request Line

  Request-Line = Method SP Request-URI SP RTSP-Version CRLF

   Method         =         "DESCRIBE"              ; Section 10.2
                  |         "ANNOUNCE"              ; Section 10.3
                  |         "GET_PARAMETER"         ; Section 10.8
                  |         "OPTIONS"               ; Section 10.1
                  |         "PAUSE"                 ; Section 10.6
                  |         "PLAY"                  ; Section 10.5
                  |         "RECORD"                ; Section 10.11
                  |         "REDIRECT"              ; Section 10.10
                  |         "SETUP"                 ; Section 10.4
                  |         "SET_PARAMETER"         ; Section 10.9
                  |         "TEARDOWN"              ; Section 10.7
                  |         extension-method

  extension-method = token

  Request-URI = "*" | absolute_URI

  RTSP-Version = "RTSP" "/" 1*DIGIT "." 1*DIGIT

  ref: https://tools.ietf.org/html/rfc2326
*/
type RTSPRequestLine struct {
	Method      string
	RequestURI  string
	RTSPVersion string
}

func (this *RTSPRequestLine) Marshal() string {
	return fmt.Sprintf("%s %s %s"+CRLF, this.Method, this.RequestURI, this.RTSPVersion)
}

func (this *RTSPRequestLine) Unmarshal(str string) error {
	this.Method = ""
	this.RequestURI = ""
	this.RTSPVersion = ""
	fmt.Sscanf(str, "%s %s %s"+CRLF, &this.Method, &this.RequestURI, &this.RTSPVersion)
	if len(this.Method) == 0 || len(this.RequestURI) == 0 || len(this.RTSPVersion) == 0 {
		return errors.New("RTSP request line unmarshal error")
	}
	return nil
}

/*
6.2 Request Header Fields

  request-header  =          Accept                   ; Section 12.1
                  |          Accept-Encoding          ; Section 12.2
                  |          Accept-Language          ; Section 12.3
                  |          Authorization            ; Section 12.5
                  |          From                     ; Section 12.20
                  |          If-Modified-Since        ; Section 12.23
                  |          Range                    ; Section 12.29
                  |          Referer                  ; Section 12.30
                  |          User-Agent               ; Section 12.41

   Note that in contrast to HTTP/1.1 [2], RTSP requests always contain
   the absolute URL (that is, including the scheme, host and port)
   rather than just the absolute path.

   ref: https://tools.ietf.org/html/rfc2326
*/
type RTSPRequestHeader struct {
	Accept          string `rtspfield:"Accept:"`
	AcceptEncoding  string `rtspfield:"Accept-Encoding:"`
	AcceptLanguage  string `rtspfield:"Accept-Language:"`
	Authorization   string `rtspfield:"Authorization:"`
	From            string `rtspfield:"From:"`
	IfModifiedSince string `rtspfield:"If-Modified-Since:"`
	Range           string `rtspfield:"Range:"`
	Referer         string `rtspfield:"Referer:"`
	UserAgent       string `rtspfield:"User-Agent:"`
}

func (this *RTSPRequestHeader) Marshal() string {
	return marshalField(this)
}

func (this *RTSPRequestHeader) Unmarshal(str string) {
	unmarshalField(this, str)
}

/*
   Request      =       Request-Line          ; Section 6.1
                *(      general-header        ; Section 5
                |       request-header        ; Section 6.2
                |       entity-header )       ; Section 8.1
                        CRLF
                        [ message-body ]      ; Section 4.3

   ref: https://tools.ietf.org/html/rfc2326
*/
type RTSPRequest struct {
	RequestLine   RTSPRequestLine
	CSeq          int
	GeneralHeader RTSPGeneralHeader
	RequestHeader RTSPRequestHeader
	EntityHeader  RTSPEntityHeader
	Session       string
	Transport     string
	MessageBody   string
}

func (this *RTSPRequest) Marshal() string {
	var str string
	str += this.RequestLine.Marshal()
	str += fmt.Sprintf("CSeq: %d"+CRLF, this.CSeq)
	str += this.GeneralHeader.Marshal()
	str += this.RequestHeader.Marshal()
	str += this.EntityHeader.Marshal()
	if len(this.Session) != 0 {
		str += fmt.Sprintf("Session: %s"+CRLF, this.Session)
	}
	if len(this.Transport) != 0 {
		str += fmt.Sprintf("Transport: %s"+CRLF, this.Transport)
	}
	str += CRLF
	if len(this.MessageBody) > 0 {
		str += this.MessageBody
	}
	return str
}

func (this *RTSPRequest) Unmarshal(str string) error {
	lines := strings.Split(str, CRLF)
	if len(lines) == 0 || len(lines) == 1 {
		return errors.New("RTSP request unmarshal error")
	}

	//Request line
	err := this.RequestLine.Unmarshal(lines[0])
	if err != nil {
		return err
	}

	//CSeq
	this.CSeq, err = strconv.Atoi(strings.TrimRight(strings.TrimLeft(regexp.MustCompile("CSeq: .*"+CRLF).FindString(str), "CSeq: "), CRLF))
	if err != nil {
		return err
	}

	//Fields
	this.GeneralHeader.Unmarshal(str)
	this.RequestHeader.Unmarshal(str)
	this.EntityHeader.Unmarshal(str)

	//Session
	this.Transport = strings.TrimRight(strings.TrimLeft(regexp.MustCompile("Session: .*"+CRLF).FindString(str), "Session: "), CRLF)

	//Transport
	this.Transport = strings.TrimRight(strings.TrimLeft(regexp.MustCompile("Transport: .*"+CRLF).FindString(str), "Transport: "), CRLF)

	//Message body
	this.MessageBody = lines[len(lines)-1]
	return nil
}
