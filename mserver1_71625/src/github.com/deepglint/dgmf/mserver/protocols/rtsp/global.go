package rtsp

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

const RTSP_VERSION = "RTSP/1.0"
const CRLF = "\r\n"
const REQ_RSP_SIZE = 4096

const DESCRIBE = "DESCRIBE"
const ANNOUNCE = "ANNOUNCE"
const GET_PARAMETER = "GET_PARAMETER"
const OPTIONS = "OPTIONS"
const PAUSE = "PAUSE"
const PLAY = "PLAY"
const RECORD = "RECORD"
const REDIRECT = "REDIRECT"
const SETUP = "SETUP"
const SET_PARAMETER = "SET_PARAMETER"
const TEARDOWN = "TEARDOWN"

/*
5 General Header Fields

   See [H4.5], except that Pragma, Transfer-Encoding and Upgrade headers
   are not defined:

      general-header     =     Cache-Control     ; Section 12.8
                         |     Connection        ; Section 12.10
                         |     Date              ; Section 12.18
                         |     Via               ; Section 12.43

   ref: https://tools.ietf.org/html/rfc2326#section-5
*/
type RTSPGeneralHeader struct {
	CacheControl string `rtspfield:"Cache-Control:"`
	Connection   string `rtspfield:"Connection:"`
	Date         string `rtspfield:"Date:"`
	Via          string `rtspfield:"Via:"`
}

func (this *RTSPGeneralHeader) Marshal() string {
	return marshalField(this)
}

func (this *RTSPGeneralHeader) Unmarshal(str string) {
	unmarshalField(this, str)
}

/*
8.1 Entity Header Fields

   Entity-header fields define optional metainformation about the
   entity-body or, if no body is present, about the resource identified
   by the request.

     entity-header       =    Allow               ; Section 12.4
                         |    Content-Base        ; Section 12.11
                         |    Content-Encoding    ; Section 12.12
                         |    Content-Language    ; Section 12.13
                         |    Content-Length      ; Section 12.14
                         |    Content-Location    ; Section 12.15
                         |    Content-Type        ; Section 12.16
                         |    Expires             ; Section 12.19
                         |    Last-Modified       ; Section 12.24
                         |    extension-header
     extension-header    =    message-header

   The extension-header mechanism allows additional entity-header fields
   to be defined without changing the protocol, but these fields cannot
   be assumed to be recognizable by the recipient. Unrecognized header
   fields SHOULD be ignored by the recipient and forwarded by proxies.

   ref: https://www.ietf.org/rfc/rfc2326.txt
*/
type RTSPEntityHeader struct {
	Allow           string `rtspfield:"Allow:"`
	ContentBase     string `rtspfield:"Content-Base:"`
	ContentEncoding string `rtspfield:"Content-Encoding:"`
	ContentLanguage string `rtspfield:"Content-Language:"`
	ContentLength   string `rtspfield:"Content-Length:"`
	ContentLocation string `rtspfield:"Content-Location:"`
	ContentType     string `rtspfield:"Content-Type:"`
	Expires         string `rtspfield:"Expires:"`
	LastModified    string `rtspfield:"Last-Modified:"`
}

func (this *RTSPEntityHeader) Marshal() string {
	return marshalField(this)
}

func (this *RTSPEntityHeader) Unmarshal(str string) {
	unmarshalField(this, str)
}

func marshalField(obj interface{}) string {
	fields := reflect.ValueOf(obj).Elem()
	str := ""
	for i := 0; i < fields.NumField(); i++ {
		fieldName := fields.Type().Field(i).Tag.Get("rtspfield")
		fieldValue := fields.Field(i).String()
		if len(fieldValue) != 0 {
			str += fmt.Sprintf(fieldName+" %s"+CRLF, fieldValue)
		}
	}
	return str
}

func unmarshalField(obj interface{}, str string) {
	fields := reflect.ValueOf(obj).Elem()
	for i := 0; i < fields.NumField(); i++ {
		fieldName := fields.Type().Field(i).Tag.Get("rtspfield")
		fieldValue := strings.Replace(strings.Replace(regexp.MustCompile(fieldName+" .*"+CRLF).FindString(str), fieldName+" ", "", -1), CRLF, "", -1)
		fields.Field(i).SetString(fieldValue)
	}
}
