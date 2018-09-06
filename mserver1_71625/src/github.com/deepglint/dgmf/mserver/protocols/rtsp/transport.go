package rtsp

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

/*
   Transport           =    "Transport" ":"
                            1\#transport-spec
   transport-spec      =    transport-protocol/profile[/lower-transport]
                            *parameter
   transport-protocol  =    "RTP"
   profile             =    "AVP"
   lower-transport     =    "TCP" | "UDP"
   parameter           =    ( "unicast" | "multicast" )
                       |    ";" "destination" [ "=" address ]
                       |    ";" "interleaved" "=" channel [ "-" channel ]
                       |    ";" "append"
                       |    ";" "ttl" "=" ttl
                       |    ";" "layers" "=" 1*DIGIT
                       |    ";" "port" "=" port [ "-" port ]
                       |    ";" "client_port" "=" port [ "-" port ]
                       |    ";" "server_port" "=" port [ "-" port ]
                       |    ";" "ssrc" "=" ssrc
                       |    ";" "mode" = <"> 1\#mode <">
   ttl                 =    1*3(DIGIT)
   port                =    1*5(DIGIT)
   ssrc                =    8*8(HEX)
   channel             =    1*3(DIGIT)
   address             =    host
   mode                =    <"> *Method <"> | Method

   ref: https://tools.ietf.org/html/rfc2326
*/
type RTSPTransport struct {
	LowerTransport string
	CastType       string
	Destination    string `rtsptransport:"destination"`
	Interleaved    string `rtsptransport:"interleaved"`
	IsAppend       bool
	TTl            string `rtsptransport:"ttl"`
	Layers         string `rtsptransport:"layers"`
	Port           string `rtsptransport:"port"`
	ClientPort     string `rtsptransport:"client_port"`
	ServerPort     string `rtsptransport:"server_port"`
	SSRC           string `rtsptransport:"ssrc"`
	Mode           string
}

func (this *RTSPTransport) Marshal() string {
	var str string
	str += "RTP/AVP"
	if len(this.LowerTransport) != 0 {
		str += "/" + this.LowerTransport
	}
	if len(this.CastType) != 0 {
		str += ";" + this.CastType
	}
	fields := reflect.ValueOf(this).Elem()
	for i := 0; i < fields.NumField(); i++ {
		fieldName := fields.Type().Field(i).Tag.Get("rtsptransport")
		fieldValue := fields.Field(i).String()
		if len(fieldName) != 0 && len(fieldValue) != 0 {
			str += fmt.Sprintf(";%s=%s", fieldName, fieldValue)
		}
	}
	if this.IsAppend {
		str += ";append"
	}
	if len(this.Mode) != 0 {
		str += fmt.Sprintf(";mode=\"%s\"", this.Mode)
	}
	return str
}

func (this *RTSPTransport) Unmarshal(str string) {
	strs := strings.Split(str, ";")
	if len(strs) >= 2 {
		if strings.Contains(strs[0], "RTP/AVP/") {
			this.LowerTransport = strings.Replace(strs[0], "RTP/AVP/", "", -1)
		}
		this.CastType = strs[1]

		fields := reflect.ValueOf(this).Elem()
		for i := 0; i < fields.NumField(); i++ {
			fieldName := fields.Type().Field(i).Tag.Get("rtsptransport")
			if len(fieldName) == 0 {
				continue
			}
			fieldValue := strings.TrimLeft(regexp.MustCompile(";"+fieldName+"=[0-9A-Za-z_-]*").FindString(str), ";"+fieldName+"=")
			fields.Field(i).SetString(fieldValue)
		}

		if strings.Contains(str, ";append") {
			this.IsAppend = true
		} else {
			this.IsAppend = false
		}

		this.Mode = strings.TrimRight(strings.TrimLeft(regexp.MustCompile(";mode=\"[0-9A-Za-z_-]*\"").FindString(str), ";mode=\""), "\"")

	}
}
