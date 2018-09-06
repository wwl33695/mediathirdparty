package rtsp

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

/*
7.1 Status-Line

   The first line of a Response message is the Status-Line, consisting
   of the protocol version followed by a numeric status code, and the
   textual phrase associated with the status code, with each element
   separated by SP characters. No CR or LF is allowed except in the
   final CRLF sequence.

   Status-Line =   RTSP-Version SP Status-Code SP Reason-Phrase CRLF

7.1.1 Status Code and Reason Phrase

     * 1xx: Informational - Request received, continuing process
     * 2xx: Success - The action was successfully received, understood,
       and accepted
     * 3xx: Redirection - Further action must be taken in order to
       complete the request
     * 4xx: Client Error - The request contains bad syntax or cannot be
       fulfilled
     * 5xx: Server Error - The server failed to fulfill an apparently
       valid request

   Code           reason
   100            Continue                         all
   200            OK                               all
   201            Created                          RECORD
   250            Low on Storage Space             RECORD
   300            Multiple Choices                 all
   301            Moved Permanently                all
   302            Moved Temporarily                all
   303            See Other                        all
   305            Use Proxy                        all
   400            Bad Request                      all
   401            Unauthorized                     all
   402            Payment Required                 all
   403            Forbidden                        all
   404            Not Found                        all
   405            Method Not Allowed               all
   406            Not Acceptable                   all
   407            Proxy Authentication Required    all
   408            Request Timeout                  all
   410            Gone                             all
   411            Length Required                  all
   412            Precondition Failed              DESCRIBE, SETUP
   413            Request Entity Too Large         all
   414            Request-URI Too Long             all
   415            Unsupported Media Type           all
   451            Invalid parameter                SETUP
   452            Illegal Conference Identifier    SETUP
   453            Not Enough Bandwidth             SETUP
   454            Session Not Found                all
   455            Method Not Valid In This State   all
   456            Header Field Not Valid           all
   457            Invalid Range                    PLAY
   458            Parameter Is Read-Only           SET_PARAMETER
   459            Aggregate Operation Not Allowed  all
   460            Only Aggregate Operation Allowed all
   461            Unsupported Transport            all
   462            Destination Unreachable          all
   500            Internal Server Error            all
   501            Not Implemented                  all
   502            Bad Gateway                      all
   503            Service Unavailable              all
   504            Gateway Timeout                  all
   505            RTSP Version Not Supported       all
   551            Option not support               all

   https://tools.ietf.org/html/rfc2326
*/
type RTSPStatusLine struct {
	RTSPVersion  string
	StatusCode   int
	ReasonPhrase string
}

func getReasonPhrase(statusCode int) string {
	switch statusCode {
	case 100:
		return "Continue"
	case 200:
		return "OK"
	case 201:
		return "Created"
	case 250:
		return "Low on Storage Space"
	case 300:
		return "Multiple Choices"
	case 301:
		return "Moved Permanently"
	case 302:
		return "Moved Temporarily"
	case 303:
		return "See Other"
	case 305:
		return "Use Proxy"
	case 400:
		return "Bad Request"
	case 401:
		return "Unauthorized"
	case 402:
		return "Payment Required"
	case 403:
		return "Forbidden"
	case 404:
		return "Not Found"
	case 405:
		return "Method Not Allowed"
	case 406:
		return "Not Acceptable"
	case 407:
		return "Proxy Authentication Required"
	case 408:
		return "Request Timeout"
	case 410:
		return "Gone"
	case 411:
		return "Length Required"
	case 412:
		return "Precondition Failed"
	case 413:
		return "Request Entity Too Large"
	case 414:
		return "Request-URI Too Long"
	case 415:
		return "Unsupported Media Type"
	case 451:
		return "Invalid parameter"
	case 452:
		return "Illegal Conference Identifier"
	case 453:
		return "Not Enough Bandwidth"
	case 454:
		return "Session Not Found"
	case 455:
		return "Method Not Valid In This State"
	case 456:
		return "Header Field Not Valid"
	case 457:
		return "Invalid Range"
	case 458:
		return "Parameter Is Read-Only"
	case 459:
		return "Aggregate Operation Not Allowed"
	case 460:
		return "Only Aggregate Operation Allowed"
	case 461:
		return "Unsupported Transport"
	case 462:
		return "Destination Unreachable"
	case 500:
		return "Internal Server Error "
	case 501:
		return "Not Implemented"
	case 502:
		return "Bad Gateway"
	case 503:
		return "Service Unavailable"
	case 504:
		return "Gateway Timeout"
	case 505:
		return "RTSP Version Not Supported"
	case 551:
		return "Option not support"
	default:
		return ""
	}
}

func (this *RTSPStatusLine) Marshal() string {
	return fmt.Sprintf("%s %d %s"+CRLF, this.RTSPVersion, this.StatusCode, this.ReasonPhrase)
}

func (this *RTSPStatusLine) Unmarshal(str string) error {
	fmt.Sscanf(str, "%s %d %s"+CRLF, &this.RTSPVersion, &this.StatusCode, &this.ReasonPhrase)
	if len(this.RTSPVersion) == 0 || this.StatusCode == 0 || len(this.ReasonPhrase) == 0 {
		return errors.New("RTSP status line unmarshal error")
	}
	return nil
}

/*
7.1.2 Response Header Fields

   The response-header fields allow the request recipient to pass
   additional information about the response which cannot be placed in
   the Status-Line. These header fields give information about the
   server and about further access to the resource identified by the
   Request-URI.

   response-header  =     Location             ; Section 12.25
                    |     Proxy-Authenticate   ; Section 12.26
                    |     Public               ; Section 12.28
                    |     Retry-After          ; Section 12.31
                    |     Server               ; Section 12.36
                    |     Vary                 ; Section 12.42
                    |     WWW-Authenticate     ; Section 12.44

   Response-header field names can be extended reliably only in
   combination with a change in the protocol version. However, new or
   experimental header fields MAY be given the semantics of response-
   header fields if all parties in the communication recognize them to
   be response-header fields. Unrecognized header fields are treated as
   entity-header fields.

   https://tools.ietf.org/html/rfc2326
*/
type RTSPResponseHeader struct {
	Location              string `rtspfield:"Location:"`
	ProxyAuthenticate     string `rtspfield:"Proxy-Authenticate:"`
	Public                string `rtspfield:"Public:"`
	RetryAfter            string `rtspfield:"Retry-After:"`
	Server                string `rtspfield:"Server:"`
	Vary                  string `rtspfield:"Vary:"`
	WWWAuthenticateDigest string `rtspfield:"WWW-Authenticate: Digest"`
	WWWAuthenticateBasic  string `rtspfield:"WWW-Authenticate: Basic"`
}

func (this *RTSPResponseHeader) Marshal() string {
	return marshalField(this)
}

func (this *RTSPResponseHeader) Unmarshal(str string) {
	unmarshalField(this, str)
}

/*
7 Response

   [H6] applies except that HTTP-Version is replaced by RTSP-Version.
   Also, RTSP defines additional status codes and does not define some
   HTTP codes. The valid response codes and the methods they can be used
   with are defined in Table 1.

   After receiving and interpreting a request message, the recipient
   responds with an RTSP response message.

     Response    =     Status-Line         ; Section 7.1
                 *(    general-header      ; Section 5
                 |     response-header     ; Section 7.1.2
                 |     entity-header )     ; Section 8.1
                       CRLF
                       [ message-body ]    ; Section 4.3
    ref: https://tools.ietf.org/html/rfc2326
*/
type RTSPResponse struct {
	StatusLine     RTSPStatusLine
	CSeq           int
	GeneralHeader  RTSPGeneralHeader
	ResponseHeader RTSPResponseHeader
	EntityHeader   RTSPEntityHeader
	Session        string
	Transport      string
	MessageBody    string
}

func (this *RTSPResponse) Marshal() string {
	var str string
	str += this.StatusLine.Marshal()
	str += fmt.Sprintf("CSeq: %d"+CRLF, this.CSeq)
	str += this.GeneralHeader.Marshal()
	str += this.ResponseHeader.Marshal()
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
		// str += CRLF
	}
	return str
}

func (this *RTSPResponse) Unmarshal(str string) error {
	lines := strings.Split(str, CRLF)
	if len(lines) == 0 || len(lines) == 1 {
		return errors.New("RTSP response unmarshal error")
	}

	//status line
	err := this.StatusLine.Unmarshal(lines[0])
	if err != nil {
		return err
	}

	//CSeq
	this.CSeq, err = strconv.Atoi(strings.TrimRight(strings.TrimLeft(regexp.MustCompile("[CSeq|Cseq]: .*"+CRLF).FindString(str), "[CSeq|Cseq]: "), CRLF))

	//Fields
	this.GeneralHeader.Unmarshal(str)
	this.ResponseHeader.Unmarshal(str)
	this.EntityHeader.Unmarshal(str)

	//Session
	this.Session = strings.TrimRight(strings.TrimLeft(regexp.MustCompile("Session: .*"+CRLF).FindString(str), "Session: "), CRLF)

	//Transport
	this.Transport = strings.TrimRight(strings.TrimLeft(regexp.MustCompile("Transport: .*"+CRLF).FindString(str), "Transport: "), CRLF)

	//Message body
	msgs := strings.Split(str, CRLF+CRLF)
	if len(msgs) < 2 {
		return errors.New("RTSP response unmarshal error")
	}

	this.MessageBody = strings.TrimRight(msgs[1], "\x00")
	return nil
}
