package gb28181

import (
	"fmt"
	"strconv"
	"time"

	"github.com/deepglint/dgmf/mserver/core"
	"github.com/deepglint/dgmf/mserver/protocols/gb28181/auth"
	"github.com/deepglint/dgmf/mserver/protocols/manscdp/manscdpbase"
	"github.com/deepglint/dgmf/mserver/protocols/sdp"
	"github.com/deepglint/dgmf/mserver/protocols/sip/base"
	"github.com/deepglint/dgmf/mserver/utils/uuid"

	"strings"
)

var noParams = base.NewParams()

//createFirstRegisterRequest
func (p *GB28181Server) createFirstRegisterRequest(expiresTime uint32, seqNum uint32) *base.Request {

	// var sipUri = base.SipUri{
	// 	User:      base.String{p.param.ServerID},
	// 	Password:  base.NoString{},
	// 	Host:      p.param.ServerAreaID,
	// 	UriParams: base.NewParams(),
	// 	Headers:   base.NewParams(),
	// }

	var sipUri = base.SipUri{
		User:      base.String{p.param.ServerID},
		Password:  base.NoString{},
		Host:      p.param.ServerAreaID,
		UriParams: base.NewParams(),
		Headers:   base.NewParams(),
	}

	request := base.NewRequest("REGISTER", sipUri.Copy(), "SIP/2.0", nil, "")

	//viaHeader
	// viaHeader := base.ViaHeader{&base.ViaHop{"SIP", "2.0", "UDP", p.gbCemera.IP,
	// 	&p.gbCemera.SignalPort, base.NewParams().Add("branch", base.String{time.Now().String()}).Add("rport", base.NoString{})}}
	// request.AddHeader(&viaHeader)

	viaHeader := base.ViaHeader{&base.ViaHop{"SIP", "2.0", "UDP", p.localhost,
		&(p.port),
		base.NewParams().Add("branch", base.String{uuid.NewV4().String()}).Add("rport", base.NoString{})}}
	request.AddHeader(&viaHeader)

	//fromHeader
	// fromHeader := base.FromHeader{DisplayName: base.NoString{},
	// 	Address: &base.SipUri{
	// 		User:      base.String{p.gbCemera.ID},
	// 		Password:  base.NoString{},
	// 		Host:      p.gbCemera.GBServer.AreaID,
	// 		UriParams: noParams,
	// 		Headers:   noParams},
	// 	Params: base.NewParams().Add("tag", base.String{"1963520561"})}
	// request.AddHeader(&fromHeader)

	fromHeader := base.FromHeader{DisplayName: base.NoString{},
		Address: &base.SipUri{
			User:      base.String{p.param.DeviceID},
			Password:  base.NoString{},
			Host:      p.param.ServerAreaID,
			UriParams: noParams,
			Headers:   noParams},
		Params: base.NewParams().Add("tag", base.String{uuid.NewV4().String()})}
	request.AddHeader(&fromHeader)

	//toHeader
	// toHeader := base.ToHeader{DisplayName: base.NoString{},
	// 	Address: &base.SipUri{User: base.String{p.gbCemera.ID}, Password: base.NoString{}, Host: p.gbCemera.GBServer.AreaID, UriParams: noParams, Headers: noParams},
	// 	Params:  base.NewParams()}
	// request.AddHeader(&toHeader)

	toHeader := base.ToHeader{DisplayName: base.NoString{},
		Address: &base.SipUri{
			User:      base.String{p.param.DeviceID},
			Password:  base.NoString{},
			Host:      p.param.ServerAreaID,
			UriParams: noParams,
			Headers:   noParams},
		Params: base.NewParams()}
	request.AddHeader(&toHeader)
	//call-ID
	var callID base.CallId = base.CallId(uuid.NewV4().String())
	request.AddHeader(&callID)

	//CSeq
	cseq := base.CSeq{
		SeqNo:      seqNum,
		MethodName: "REGISTER",
	}
	request.AddHeader(&cseq)

	//contactHeader
	// contactHeader := base.ContactHeader{
	// 	DisplayName: base.NoString{},
	// 	Address: &base.SipUri{
	// 		User:      base.String{p.gbCemera.ID},
	// 		Password:  base.NoString{},
	// 		Host:      p.gbCemera.IP,
	// 		Port:      &p.gbCemera.SignalPort,
	// 		UriParams: noParams,
	// 		Headers:   noParams},
	// 	Params: noParams,
	// }
	// request.AddHeader(&contactHeader)

	contactHeader := base.ContactHeader{
		DisplayName: base.NoString{},
		Address: &base.SipUri{
			User:      base.String{p.param.DeviceID},
			Password:  base.NoString{},
			Host:      p.localhost,
			Port:      &p.port,
			UriParams: noParams,
			Headers:   noParams},
		Params: noParams,
	}
	request.AddHeader(&contactHeader)

	//MaxForwards
	var maxForwards base.MaxForwards = 70
	request.AddHeader(&maxForwards)

	// UserAgent
	var userAgent base.UserAgent = "IP Camera"
	request.AddHeader(&userAgent)

	//Expires
	var expires base.Expires = base.Expires(expiresTime)
	request.AddHeader(&expires)

	//ContentLength
	var contentLength base.ContentLength = 0
	request.AddHeader(contentLength)
	fmt.Println(request.String())

	return request
}

//createSecondRegisterRequest
func (p *GB28181Server) createSecondRegisterRequest(nonce string, expiresTime uint32, seqNum uint32) *base.Request {

	// var sipUri = base.SipUri{
	// 	User:      base.String{p.gbCemera.GBServer.IP},
	// 	Password:  base.NoString{},
	// 	Host:      p.gbCemera.GBServer.AreaID,
	// 	UriParams: base.NewParams(),
	// 	Headers:   base.NewParams(),
	// }

	var sipUri = base.SipUri{
		User:      base.String{p.param.ServerID},
		Password:  base.NoString{},
		Host:      p.param.ServerAreaID,
		UriParams: base.NewParams(),
		Headers:   base.NewParams(),
	}
	request := base.NewRequest("REGISTER", sipUri.Copy(), "SIP/2.0", nil, "")
	//viaHeader
	// viaHeader := base.ViaHeader{&base.ViaHop{"SIP", "2.0", "UDP", p.gbCemera.IP,
	// 	&p.gbCemera.SignalPort, base.NewParams().Add("branch", base.String{time.Now().String()}).Add("rport", base.NoString{})}}
	// request.AddHeader(&viaHeader)

	viaHeader := base.ViaHeader{&base.ViaHop{"SIP", "2.0", "UDP", p.localhost,
		&p.port, base.NewParams().Add("branch", base.String{uuid.NewV4().String()}).Add("rport", base.NoString{})}}
	request.AddHeader(&viaHeader)

	//fromHeader
	// fromHeader := base.FromHeader{DisplayName: base.NoString{},
	// 	Address: &base.SipUri{
	// 		User:      base.String{p.gbCemera.ID},
	// 		Password:  base.NoString{},
	// 		Host:      p.gbCemera.GBServer.AreaID,
	// 		UriParams: noParams,
	// 		Headers:   noParams},
	// 	Params: base.NewParams().Add("tag", base.String{"1963520561"})}
	// request.AddHeader(&fromHeader)

	fromHeader := base.FromHeader{DisplayName: base.NoString{},
		Address: &base.SipUri{
			User:      base.String{p.param.DeviceID},
			Password:  base.NoString{},
			Host:      p.param.ServerAreaID,
			UriParams: noParams,
			Headers:   noParams},
		Params: base.NewParams().Add("tag", base.String{uuid.NewV4().String()})}
	request.AddHeader(&fromHeader)

	//toHeader
	// toHeader := base.ToHeader{DisplayName: base.NoString{},
	// 	Address: &base.SipUri{
	// 		User:      base.String{p.gbCemera.ID},
	// 		Password:  base.NoString{},
	// 		Host:      p.gbCemera.GBServer.AreaID,
	// 		UriParams: noParams,
	// 		Headers:   noParams},
	// 	Params: base.NewParams()}
	// request.AddHeader(&toHeader)
	toHeader := base.ToHeader{DisplayName: base.NoString{},
		Address: &base.SipUri{
			User:      base.String{p.param.DeviceID},
			Password:  base.NoString{},
			Host:      p.param.ServerAreaID,
			UriParams: noParams,
			Headers:   noParams},
		Params: base.NewParams()}
	request.AddHeader(&toHeader)
	//call-ID
	var callID base.CallId = base.CallId(uuid.NewV4().String())
	request.AddHeader(&callID)

	//CSeq
	cseq := base.CSeq{
		SeqNo:      seqNum,
		MethodName: "REGISTER",
	}
	request.AddHeader(&cseq)

	//contactHeader
	// contactHeader := base.ContactHeader{
	// 	DisplayName: base.NoString{},
	// 	Address: &base.SipUri{
	// 		User:      base.String{p.gbCemera.ID},
	// 		Password:  base.NoString{},
	// 		Host:      p.gbCemera.IP,
	// 		Port:      &p.gbCemera.SignalPort,
	// 		UriParams: noParams,
	// 		Headers:   noParams},
	// 	Params: noParams,
	// }
	// request.AddHeader(&contactHeader)

	contactHeader := base.ContactHeader{
		DisplayName: base.NoString{},
		Address: &base.SipUri{
			User:      base.String{p.param.DeviceID},
			Password:  base.NoString{},
			Host:      p.localhost,
			Port:      &p.port,
			UriParams: noParams,
			Headers:   noParams},
		Params: noParams,
	}
	request.AddHeader(&contactHeader)

	//MaxForwards
	var maxForwards base.MaxForwards = 70
	request.AddHeader(&maxForwards)

	// UserAgent
	var userAgent base.UserAgent = "IP Camera"
	request.AddHeader(&userAgent)

	//Expires
	var expires base.Expires = base.Expires(expiresTime)
	request.AddHeader(&expires)

	//ContentLength
	var contentLength base.ContentLength = 0
	request.AddHeader(&contentLength)

	//AuthHeader
	// uri := "sip:" + p.gbCemera.GBServer.ID + "@" + p.gbCemera.GBServer.AreaID
	// authHeader := base.AuthorizationHeader{
	// 	AuthenticationScheme: base.String{"Digest"},
	// 	Params: base.NewParams().
	// 		Add("username", base.String{"\"" + p.gbCemera.ID + "\""}).
	// 		Add("realm", base.String{"\"" + p.gbCemera.GBServer.AreaID + "\""}).
	// 		Add("nonce", base.String{"\"" + nonce + "\""}).
	// 		Add("uri", base.String{"\"" + uri + "\""}).
	// 		Add("response", base.String{"\"" + auth.CreateDigestAuth(p.gbCemera.ID, p.gbCemera.GBServer.Password, p.gbCemera.GBServer.AreaID, "REGISTER", uri, nonce) + "\""}).
	// 		Add("algorithm", base.String{"MD5"}),
	// }
	// request.AddHeader(&authHeader)

	uri := "sip:" + p.param.ServerID + "@" + p.param.ServerAreaID
	authHeader := base.AuthorizationHeader{
		AuthenticationScheme: base.String{"Digest"},
		Params: base.NewParams().
			Add("username", base.String{"\"" + p.param.DeviceID + "\""}).
			Add("realm", base.String{"\"" + p.param.ServerAreaID + "\""}).
			Add("nonce", base.String{"\"" + nonce + "\""}).
			Add("uri", base.String{"\"" + uri + "\""}).
			Add("response", base.String{"\"" + auth.CreateDigestAuth(p.param.DeviceID, p.param.ServerPassword, p.param.ServerAreaID, "REGISTER", uri, nonce) + "\""}).
			Add("algorithm", base.String{"MD5"}),
	}
	request.AddHeader(&authHeader)

	fmt.Println(request.String())
	return request
}

//生成保持连接请求
func (p *GB28181Server) createKeepLiveRequest() *base.Request {

	// var sipUri = base.SipUri{
	// 	User:      base.String{p.gbCemera.GBServer.IP},
	// 	Password:  base.NoString{},
	// 	Host:      p.gbCemera.GBServer.AreaID,
	// 	UriParams: noParams,
	// 	Headers:   noParams,
	// }

	var sipUri = base.SipUri{
		User:      base.String{p.param.ServerID},
		Password:  base.NoString{},
		Host:      p.param.ServerAreaID,
		UriParams: noParams,
		Headers:   noParams,
	}

	request := base.NewRequest("MESSAGE", sipUri.Copy(), "SIP/2.0", nil, "")
	//viaHeader
	// viaHeader := base.ViaHeader{&base.ViaHop{"SIP", "2.0", "UDP", p.gbCemera.IP,
	// 	&p.gbCemera.SignalPort, base.NewParams().Add("branch", base.String{time.Now().String()}).Add("rport", base.NoString{})}}
	// request.AddHeader(&viaHeader)

	//viaHeader
	viaHeader := base.ViaHeader{&base.ViaHop{"SIP", "2.0", "UDP", p.localhost,
		&p.port, base.NewParams().Add("branch", base.String{uuid.NewV4().String()}).Add("rport", base.NoString{})}}
	request.AddHeader(&viaHeader)

	//fromHeader
	// fromHeader := base.FromHeader{
	// 	DisplayName: base.NoString{},
	// 	Address: &base.SipUri{
	// 		User:      base.String{p.gbCemera.ID},
	// 		Password:  base.NoString{},
	// 		Host:      p.gbCemera.GBServer.AreaID,
	// 		UriParams: noParams,
	// 		Headers:   noParams},
	// 	Params: base.NewParams().Add("tag", base.String{time.Now().String()})}
	// request.AddHeader(&fromHeader)

	fromHeader := base.FromHeader{
		DisplayName: base.NoString{},
		Address: &base.SipUri{
			User:      base.String{p.param.DeviceID},
			Password:  base.NoString{},
			Host:      p.param.ServerAreaID,
			UriParams: noParams,
			Headers:   noParams},
		Params: base.NewParams().Add("tag", base.String{uuid.NewV4().String()})}
	request.AddHeader(&fromHeader)

	//toHeader
	// toHeader := base.ToHeader{
	// 	DisplayName: base.NoString{},
	// 	Address: &base.SipUri{
	// 		User:      base.String{p.gbCemera.ID},
	// 		Password:  base.NoString{},
	// 		Host:      p.gbCemera.GBServer.AreaID,
	// 		UriParams: noParams,
	// 		Headers:   noParams},
	// 	Params: base.NewParams()}
	// request.AddHeader(&toHeader)
	toHeader := base.ToHeader{
		DisplayName: base.NoString{},
		Address: &base.SipUri{
			User:      base.String{p.param.DeviceID},
			Password:  base.NoString{},
			Host:      p.param.ServerAreaID,
			UriParams: noParams,
			Headers:   noParams},
		Params: base.NewParams()}
	request.AddHeader(&toHeader)
	//call-ID
	// var callID base.CallId = (base.CallId)("KeepAlive:" + p.gbCemera.IP + time.Now().String())
	// request.AddHeader(&callID)

	var callID base.CallId = (base.CallId)(uuid.NewV4().String())
	request.AddHeader(&callID)

	//CSeq
	cseq := base.CSeq{
		SeqNo:      20,
		MethodName: "MESSAGE",
	}
	request.AddHeader(&cseq)

	//MaxForwards
	var maxForwards base.MaxForwards = 70
	request.AddHeader(&maxForwards)

	// UserAgent
	var userAgent base.UserAgent = "IP Camera"
	request.AddHeader(&userAgent)

	//Content-Type
	contentType := base.GenericHeader{
		HeaderName: "Content-Type",
		Contents:   "Application/MANSCDP+xml",
	}
	request.AddHeader(&contentType)

	p.keepAliveSN = (p.keepAliveSN + 1) % 65535

	var keepalivemanscdp = manscdpbase.Notify{
		CmdType:  "Keepalive",
		SN:       uint64(p.keepAliveSN),
		DeviceID: p.param.DeviceID,
		Status:   "OK",
	}

	xmlOutPut, outPutErr := keepalivemanscdp.ToXml()
	if outPutErr == nil {

		request.SetBody(xmlOutPut)
		//ContentLength
		var contentLength base.ContentLength = base.ContentLength(len(xmlOutPut))
		request.AddHeader(&contentLength)
	}

	// fmt.Println(request.String())

	return request
}

//createSipResponse  create  sip respnese according to request
func (p *GB28181Server) createSipResponse(request *base.Request, statusCode uint16, reason string, body string) *base.Response {

	respone := base.NewResponse("SIP/2.0", statusCode, reason, nil, "")

	// if len(request.Headers("Via")) > 0 {
	// 	viaHeader := request.Headers("Via")[0]
	// 	if value, ok := (viaHeader).(*(base.ViaHeader)); ok {
	// 		if len(*value) > 0 {
	// 			viahop := (*value)[0]
	// 			viahop.Params.Add("rport", base.String{strconv.Itoa(int(p.gbCemera.SignalPort))})
	// 		}
	// 	}
	// 	respone.AddHeader(viaHeader)
	// }

	if len(request.Headers("Via")) > 0 {
		viaHeader := request.Headers("Via")[0]
		if value, ok := (viaHeader).(*(base.ViaHeader)); ok {
			if len(*value) > 0 {
				viahop := (*value)[0]
				viahop.Params.Add("rport", base.String{strconv.Itoa(int(p.port))})
			}
		}
		respone.AddHeader(viaHeader)
	}

	if len(request.Headers("From")) > 0 {
		respone.AddHeader(request.Headers("From")[0])
	}

	if len(request.Headers("To")) > 0 {
		respone.AddHeader(request.Headers("To")[0])
	}
	if len(request.Headers("Call-ID")) > 0 {
		respone.AddHeader(request.Headers("Call-ID")[0])
	}
	if len(request.Headers("CSeq")) > 0 {
		respone.AddHeader(request.Headers("CSeq")[0])
	}

	// UserAgent
	var userAgent base.UserAgent = "IP Camera"
	respone.AddHeader(&userAgent)

	//ContentLength
	var contentLength base.ContentLength = base.ContentLength(len(body))
	respone.AddHeader(&contentLength)
	respone.SetBody(body)
	fmt.Println(respone.String())
	return respone
}

//生成目录查询结果反馈请求
func (p *GB28181Server) createCatalogRequest(sn uint64) *base.Request {

	// var sipUri = base.SipUri{
	// 	User:      base.String{p.gbCemera.GBServer.ID},
	// 	Password:  base.NoString{},
	// 	Host:      p.gbCemera.GBServer.AreaID,
	// 	UriParams: noParams,
	// 	Headers:   noParams,
	// }

	// request := base.NewRequest("MESSAGE", sipUri.Copy(), "SIP/2.0", nil, "")

	var sipUri = base.SipUri{
		User:      base.String{p.param.ServerID},
		Password:  base.NoString{},
		Host:      p.param.ServerAreaID,
		UriParams: noParams,
		Headers:   noParams,
	}

	request := base.NewRequest("MESSAGE", sipUri.Copy(), "SIP/2.0", nil, "")
	//viaHeader
	// viaHeader := base.ViaHeader{&base.ViaHop{"SIP", "2.0", "UDP", p.gbCemera.IP,
	// 	&p.gbCemera.SignalPort, base.NewParams().Add("branch", base.String{time.Now().String()}).Add("rport", base.NoString{})}}
	// request.AddHeader(&viaHeader)

	viaHeader := base.ViaHeader{&base.ViaHop{"SIP", "2.0", "UDP", p.localhost,
		&p.port, base.NewParams().Add("branch", base.String{uuid.NewV4().String()}).Add("rport", base.NoString{})}}
	request.AddHeader(&viaHeader)

	//fromHeader
	// fromHeader := base.FromHeader{
	// 	DisplayName: base.NoString{},
	// 	Address: &base.SipUri{
	// 		User:      base.String{p.gbCemera.ID},
	// 		Password:  base.NoString{},
	// 		Host:      p.gbCemera.GBServer.AreaID,
	// 		UriParams: noParams,
	// 		Headers:   noParams},
	// 	Params: base.NewParams().Add("tag", base.String{time.Now().String()})}
	// request.AddHeader(&fromHeader)

	fromHeader := base.FromHeader{
		DisplayName: base.NoString{},
		Address: &base.SipUri{
			User:      base.String{p.param.DeviceID},
			Password:  base.NoString{},
			Host:      p.param.ServerAreaID,
			UriParams: noParams,
			Headers:   noParams},
		Params: base.NewParams().Add("tag", base.String{uuid.NewV4().String()})}
	request.AddHeader(&fromHeader)

	//toHeader
	// toHeader := base.ToHeader{DisplayName: base.NoString{},
	// 	Address: &base.SipUri{
	// 		User:      base.String{p.gbCemera.ID},
	// 		Password:  base.NoString{},
	// 		Host:      p.gbCemera.GBServer.AreaID,
	// 		UriParams: noParams,
	// 		Headers:   noParams},
	// 	Params: base.NewParams()}
	// request.AddHeader(&toHeader)

	toHeader := base.ToHeader{DisplayName: base.NoString{},
		Address: &base.SipUri{
			User:      base.String{p.param.DeviceID},
			Password:  base.NoString{},
			Host:      p.param.ServerAreaID,
			UriParams: noParams,
			Headers:   noParams},
		Params: base.NewParams()}
	request.AddHeader(&toHeader)
	//call-ID
	// var callID base.CallId = (base.CallId)("CatalogRequest:" + p.gbCemera.IP + time.Now().String())
	// request.AddHeader(&callID)
	var callID base.CallId = (base.CallId)(uuid.NewV4().String())
	request.AddHeader(&callID)

	//CSeq
	cseq := base.CSeq{
		SeqNo:      20,
		MethodName: "MESSAGE",
	}
	request.AddHeader(&cseq)

	//MaxForwards
	var maxForwards base.MaxForwards = 70
	request.AddHeader(&maxForwards)

	// UserAgent
	var userAgent base.UserAgent = "IP Camera"
	request.AddHeader(&userAgent)

	//Content-Type
	contentType := base.GenericHeader{
		HeaderName: "Content-Type",
		Contents:   "Application/MANSCDP+xml",
	}
	request.AddHeader(&contentType)

	catalogResponse := manscdpbase.CatalogResponse{}
	// catalogResponse.DeviceID = p.gbCemera.ID
	catalogResponse.DeviceID = p.param.DeviceID
	catalogResponse.CmdType = "Catalog"
	catalogResponse.SN = sn

	var subChannels []core.GB28181Channel
	pool := core.GetESPool()
	for _, stream := range pool.Live.Maps {
		for _, output := range stream.Outputs {
			if strings.EqualFold(output.Protocol, "gb28181") {
				if subChannel, ok := output.Param.(core.GB28181Channel); ok {
					subChannels = append(subChannels, subChannel)
				}

			}
		}
	}

	if len(subChannels) > 0 {
		for _, value := range subChannels {
			item := manscdpbase.DeviceItem{}
			item.DeviceID = value.ID
			item.Name = value.Name
			item.Manufacturer = value.Manufacturer
			item.Model = value.Model
			item.Owner = value.Owner
			item.CivilCode = value.CivilCode
			item.Address = value.Address
			item.Parental = value.Parental
			item.SafetyWay = value.SafetyWay
			item.RegisterWay = value.RegisterWay
			item.Secrecy = value.Secrecy
			item.Status = value.Status
			catalogResponse.DeviceList.Item = append(catalogResponse.DeviceList.Item, item)
			catalogResponse.DeviceList.Num += 1
			catalogResponse.SumNum += 1

		}
	}

	xmlOutPut, outPutErr := catalogResponse.ToXml()
	if outPutErr == nil {

		request.SetBody(xmlOutPut)
		//ContentLength
		var contentLength base.ContentLength = base.ContentLength(len(xmlOutPut))
		request.AddHeader(&contentLength)
	}

	fmt.Println(request.String())

	return request

}

//生成设备信息查询结果反馈请求
func (p *GB28181Server) createDeviceInfoRequest(sn uint64) *base.Request {

	// var sipUri = base.SipUri{
	// 	User:      base.String{p.gbCemera.GBServer.ID},
	// 	Password:  base.NoString{},
	// 	Host:      p.gbCemera.GBServer.AreaID,
	// 	UriParams: noParams,
	// 	Headers:   noParams,
	// }

	// request := base.NewRequest("MESSAGE", sipUri.Copy(), "SIP/2.0", nil, "")

	var sipUri = base.SipUri{
		User:      base.String{p.param.ServerID},
		Password:  base.NoString{},
		Host:      p.param.ServerAreaID,
		UriParams: noParams,
		Headers:   noParams,
	}

	request := base.NewRequest("MESSAGE", sipUri.Copy(), "SIP/2.0", nil, "")
	//viaHeader
	// viaHeader := base.ViaHeader{&base.ViaHop{"SIP", "2.0", "UDP", p.gbCemera.IP,
	// 	&p.gbCemera.SignalPort, base.NewParams().Add("branch", base.String{time.Now().String()}).Add("rport", base.NoString{})}}
	// request.AddHeader(&viaHeader)

	viaHeader := base.ViaHeader{&base.ViaHop{"SIP", "2.0", "UDP", p.localhost,
		&p.port, base.NewParams().Add("branch", base.String{uuid.NewV4().String()}).Add("rport", base.NoString{})}}
	request.AddHeader(&viaHeader)

	//fromHeader
	// fromHeader := base.FromHeader{
	// 	DisplayName: base.NoString{},
	// 	Address: &base.SipUri{
	// 		User:      base.String{p.gbCemera.ID},
	// 		Password:  base.NoString{},
	// 		Host:      p.gbCemera.GBServer.AreaID,
	// 		UriParams: noParams,
	// 		Headers:   noParams},
	// 	Params: base.NewParams().Add("tag", base.String{time.Now().String()})}
	// request.AddHeader(&fromHeader)

	fromHeader := base.FromHeader{
		DisplayName: base.NoString{},
		Address: &base.SipUri{
			User:      base.String{p.param.DeviceID},
			Password:  base.NoString{},
			Host:      p.param.ServerAreaID,
			UriParams: noParams,
			Headers:   noParams},
		Params: base.NewParams().Add("tag", base.String{uuid.NewV4().String()})}
	request.AddHeader(&fromHeader)

	//toHeader
	// toHeader := base.ToHeader{DisplayName: base.NoString{},
	// 	Address: &base.SipUri{
	// 		User:      base.String{p.gbCemera.ID},
	// 		Password:  base.NoString{},
	// 		Host:      p.gbCemera.GBServer.AreaID,
	// 		UriParams: noParams,
	// 		Headers:   noParams},
	// 	Params: base.NewParams()}
	// request.AddHeader(&toHeader)

	toHeader := base.ToHeader{DisplayName: base.NoString{},
		Address: &base.SipUri{
			User:      base.String{p.param.DeviceID},
			Password:  base.NoString{},
			Host:      p.param.ServerAreaID,
			UriParams: noParams,
			Headers:   noParams},
		Params: base.NewParams()}
	request.AddHeader(&toHeader)
	//call-ID
	// var callID base.CallId = (base.CallId)("CatalogRequest:" + p.gbCemera.IP + time.Now().String())
	// request.AddHeader(&callID)
	var callID base.CallId = (base.CallId)(uuid.NewV4().String())
	request.AddHeader(&callID)

	//CSeq
	cseq := base.CSeq{
		SeqNo:      20,
		MethodName: "MESSAGE",
	}
	request.AddHeader(&cseq)

	//MaxForwards
	var maxForwards base.MaxForwards = 70
	request.AddHeader(&maxForwards)

	// UserAgent
	var userAgent base.UserAgent = "IP Camera"
	request.AddHeader(&userAgent)

	//Content-Type
	contentType := base.GenericHeader{
		HeaderName: "Content-Type",
		Contents:   "Application/MANSCDP+xml",
	}
	request.AddHeader(&contentType)

	deviceInfoResponse := manscdpbase.DeviceInfoResponse{}

	deviceInfoResponse.CmdType = "DeviceInfo"
	deviceInfoResponse.SN = sn
	deviceInfoResponse.DeviceID = p.param.DeviceID
	deviceInfoResponse.Result = "OK"
	deviceInfoResponse.DeviceType = "IP Camera"
	deviceInfoResponse.Manufacturer = "Deepglint"

	xmlOutPut, outPutErr := deviceInfoResponse.ToXml()
	if outPutErr == nil {

		request.SetBody(xmlOutPut)
		//ContentLength
		var contentLength base.ContentLength = base.ContentLength(len(xmlOutPut))
		request.AddHeader(&contentLength)
	}

	fmt.Println(request.String())

	return request

}

//生成设备状态查询结果反馈请求
func (p *GB28181Server) createDeviceStatusRequest(sn uint64) *base.Request {

	// var sipUri = base.SipUri{
	// 	User:      base.String{p.gbCemera.GBServer.ID},
	// 	Password:  base.NoString{},
	// 	Host:      p.gbCemera.GBServer.AreaID,
	// 	UriParams: noParams,
	// 	Headers:   noParams,
	// }

	// request := base.NewRequest("MESSAGE", sipUri.Copy(), "SIP/2.0", nil, "")

	var sipUri = base.SipUri{
		User:      base.String{p.param.ServerID},
		Password:  base.NoString{},
		Host:      p.param.ServerAreaID,
		UriParams: noParams,
		Headers:   noParams,
	}

	request := base.NewRequest("MESSAGE", sipUri.Copy(), "SIP/2.0", nil, "")
	//viaHeader
	// viaHeader := base.ViaHeader{&base.ViaHop{"SIP", "2.0", "UDP", p.gbCemera.IP,
	// 	&p.gbCemera.SignalPort, base.NewParams().Add("branch", base.String{time.Now().String()}).Add("rport", base.NoString{})}}
	// request.AddHeader(&viaHeader)

	viaHeader := base.ViaHeader{&base.ViaHop{"SIP", "2.0", "UDP", p.localhost,
		&p.port, base.NewParams().Add("branch", base.String{uuid.NewV4().String()}).Add("rport", base.NoString{})}}
	request.AddHeader(&viaHeader)

	//fromHeader
	// fromHeader := base.FromHeader{
	// 	DisplayName: base.NoString{},
	// 	Address: &base.SipUri{
	// 		User:      base.String{p.gbCemera.ID},
	// 		Password:  base.NoString{},
	// 		Host:      p.gbCemera.GBServer.AreaID,
	// 		UriParams: noParams,
	// 		Headers:   noParams},
	// 	Params: base.NewParams().Add("tag", base.String{time.Now().String()})}
	// request.AddHeader(&fromHeader)

	fromHeader := base.FromHeader{
		DisplayName: base.NoString{},
		Address: &base.SipUri{
			User:      base.String{p.param.DeviceID},
			Password:  base.NoString{},
			Host:      p.param.ServerAreaID,
			UriParams: noParams,
			Headers:   noParams},
		Params: base.NewParams().Add("tag", base.String{uuid.NewV4().String()})}
	request.AddHeader(&fromHeader)

	//toHeader
	// toHeader := base.ToHeader{DisplayName: base.NoString{},
	// 	Address: &base.SipUri{
	// 		User:      base.String{p.gbCemera.ID},
	// 		Password:  base.NoString{},
	// 		Host:      p.gbCemera.GBServer.AreaID,
	// 		UriParams: noParams,
	// 		Headers:   noParams},
	// 	Params: base.NewParams()}
	// request.AddHeader(&toHeader)

	toHeader := base.ToHeader{DisplayName: base.NoString{},
		Address: &base.SipUri{
			User:      base.String{p.param.DeviceID},
			Password:  base.NoString{},
			Host:      p.param.ServerAreaID,
			UriParams: noParams,
			Headers:   noParams},
		Params: base.NewParams()}
	request.AddHeader(&toHeader)
	//call-ID
	// var callID base.CallId = (base.CallId)("CatalogRequest:" + p.gbCemera.IP + time.Now().String())
	// request.AddHeader(&callID)
	var callID base.CallId = (base.CallId)(uuid.NewV4().String())
	request.AddHeader(&callID)

	//CSeq
	cseq := base.CSeq{
		SeqNo:      20,
		MethodName: "MESSAGE",
	}
	request.AddHeader(&cseq)

	//MaxForwards
	var maxForwards base.MaxForwards = 70
	request.AddHeader(&maxForwards)

	// UserAgent
	var userAgent base.UserAgent = "IP Camera"
	request.AddHeader(&userAgent)

	//Content-Type
	contentType := base.GenericHeader{
		HeaderName: "Content-Type",
		Contents:   "Application/MANSCDP+xml",
	}
	request.AddHeader(&contentType)

	deviceStatusResponse := manscdpbase.DeviceStatusResponse{}

	deviceStatusResponse.CmdType = "DeviceStatus"
	deviceStatusResponse.SN = sn
	deviceStatusResponse.DeviceID = p.param.DeviceID
	deviceStatusResponse.Result = "OK"
	deviceStatusResponse.Record = "OFF"
	deviceStatusResponse.DeviceTime = time.Unix(time.Now().Unix(), 0).Format("2006-01-02T15:04:05")

	status := true
	pool := core.GetESPool()
	for _, stream := range pool.Live.Maps {
		if !stream.InputStatus {
			status = false
			break
		}
	}

	if status {
		deviceStatusResponse.Online = "ONLINE"
		deviceStatusResponse.Encode = "ON"
		deviceStatusResponse.Status = "OK"
	} else {
		deviceStatusResponse.Online = "OFFLINE"
		deviceStatusResponse.Encode = "OFF"
		deviceStatusResponse.Status = "ERROR"
	}

	xmlOutPut, outPutErr := deviceStatusResponse.ToXml()
	if outPutErr == nil {

		request.SetBody(xmlOutPut)
		//ContentLength
		var contentLength base.ContentLength = base.ContentLength(len(xmlOutPut))
		request.AddHeader(&contentLength)
	}

	fmt.Println(request.String())

	return request

}

//生成Invite的确认建立请求
func (p *GB28181Server) createInviteConfirmResponse(request *base.Request, ssrc string) *base.Response {

	v := 0
	// o := sdp.Origin{
	// 	Username:       p.gbCemera.ID,
	// 	SessionId:      3959,
	// 	SessionVersion: 3959,
	// 	Network:        "IN",
	// 	Type:           "IP4",
	// 	Address:        p.gbCemera.IP,
	// }
	o := sdp.Origin{
		Username:       p.param.DeviceID,
		SessionId:      3959,
		SessionVersion: 3959,
		Network:        "IN",
		Type:           "IP4",
		Address:        p.localhost,
	}
	s := "play"
	// c := sdp.Connection{
	// 	Network: "IN",
	// 	Type:    "IP4",
	// 	Address: p.gbCemera.IP,
	// }
	c := sdp.Connection{
		Network: "IN",
		Type:    "IP4",
		Address: p.localhost,
	}
	var t sdp.Timing

	var a []*sdp.Attribute
	a = append(a, &sdp.Attribute{Name: "sendonly"})
	a = append(a, &sdp.Attribute{Name: "rtpmap", Value: "96 PS/90000"})
	// a = append(a, &sdp.Attribute{Name: "username", Value: p.gbCemera.ID})
	// a = append(a, &sdp.Attribute{Name: "password", Value: p.gbCemera.GBServer.Password})
	a = append(a, &sdp.Attribute{Name: "username", Value: p.param.DeviceID})
	a = append(a, &sdp.Attribute{Name: "password", Value: p.param.ServerPassword})

	media := sdp.Media{Type: "video", Port: 15060, Proto: "RTP/AVP"}
	media.Attributes = a
	// media.SSRC = "0000018467"
	var m []*sdp.Media
	m = append(m, &media)

	description := sdp.Description{
		Version:    v,
		Origin:     &o,
		Session:    s,
		Connection: &c,
		Timing:     &t,
		Media:      m,
		SSRC:       ssrc,
	}

	response := p.createSipResponse(request, 200, "OK", description.String())
	conttypeHeader := base.GenericHeader{
		HeaderName: "Content-Type",
		Contents:   "application/sdp",
	}

	// contactHeader := base.ContactHeader{
	// 	DisplayName: base.NoString{},
	// 	Address: &base.SipUri{
	// 		User:      base.String{p.gbCemera.ID},
	// 		Password:  base.NoString{},
	// 		Host:      p.gbCemera.IP,
	// 		Port:      &p.gbCemera.SignalPort,
	// 		UriParams: noParams,
	// 		Headers:   noParams},
	// 	Params: noParams,
	// }

	contactHeader := base.ContactHeader{
		DisplayName: base.NoString{},
		Address: &base.SipUri{
			User:      base.String{p.param.DeviceID},
			Password:  base.NoString{},
			Host:      p.localhost,
			Port:      &p.port,
			UriParams: noParams,
			Headers:   noParams},
		Params: noParams,
	}
	response.AddHeader(&conttypeHeader)
	response.AddHeader(&contactHeader)
	return response

}
