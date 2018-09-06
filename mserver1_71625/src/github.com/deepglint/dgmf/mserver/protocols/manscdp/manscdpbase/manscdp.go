package manscdpbase

import "encoding/xml"

//manscdp 接口
type Manscdp interface {
	ToXml() (string, error)
}

//保持连接命令
type Notify struct {
	CmdType  string
	SN       uint64
	DeviceID string
	Status   string
}

//获取Notify的xml字符串
func (notify *Notify) ToXml() (string, error) {
	xmlOutPut, outPutErr := xml.MarshalIndent(*notify, "", "    ")

	if outPutErr == nil {
		//加入XML头
		headerBytes := []byte(xml.Header)
		//拼接XML头和实际XML内容
		xmlOutPutData := (string)(append(headerBytes, xmlOutPut...))
		return xmlOutPutData, nil
	}
	return "", outPutErr
}

//控制命令
type Control struct {
	CmdType  string
	SN       uint64
	DeviceID string
	PTZCmd   string
	TeleBoot string
	Info     Info
}

type Info struct {
	ControlPriority int
}

//获取Control的xml字符串
func (control *Control) ToXml() (string, error) {
	xmlOutPut, outPutErr := xml.MarshalIndent(*control, "", "    ")

	if outPutErr == nil {
		//加入XML头
		headerBytes := []byte(xml.Header)
		//拼接XML头和实际XML内容
		xmlOutPutData := (string)(append(headerBytes, xmlOutPut...))
		return xmlOutPutData, nil
	}
	return "", outPutErr
}

//查询命令
type Query struct {
	CmdType  string
	SN       uint64
	DeviceID string
}

//获取Query的xml字符串
func (query *Query) ToXml() (string, error) {
	xmlOutPut, outPutErr := xml.MarshalIndent(*query, "", "    ")

	if outPutErr == nil {
		//加入XML头
		headerBytes := []byte(xml.Header)
		//拼接XML头和实际XML内容
		xmlOutPutData := (string)(append(headerBytes, xmlOutPut...))
		return xmlOutPutData, nil
	}
	return "", outPutErr
}

//目录响应根节点转化
func marshalCatalogResponse(p CatalogResponse) ([]byte, error) {
	tmp := struct {
		CatalogResponse
		XMLName struct{} `xml:"Response"`
	}{CatalogResponse: p}

	return xml.MarshalIndent(tmp, "", "   ")
}

//目录结果响应
type CatalogResponse struct {
	CmdType    string
	SN         uint64
	DeviceID   string
	SumNum     uint
	DeviceList DeviceList
}

//设备列表
type DeviceList struct {
	Num  int `xml:",attr"`
	Item []DeviceItem
}

//子设备
type DeviceItem struct {
	DeviceID     string
	Name         string
	Manufacturer string
	Model        string
	Owner        string
	CivilCode    string
	Address      string
	Parental     int
	SafetyWay    int
	RegisterWay  int
	Secrecy      int
	Status       string
}

//获取Catalog的xml字符串
func (catalogResopnse *CatalogResponse) ToXml() (string, error) {
	xmlOutPut, outPutErr := marshalCatalogResponse(*catalogResopnse)
	if outPutErr == nil {
		//加入XML头
		headerBytes := []byte(xml.Header)
		//拼接XML头和实际XML内容
		xmlOutPutData := (string)(append(headerBytes, xmlOutPut...))
		return xmlOutPutData, nil
	}
	return "", outPutErr
}

//设备信息响应
type DeviceInfoResponse struct {
	CmdType      string
	SN           uint64
	DeviceID     string
	Result       string
	DeviceType   string
	Manufacturer string
}

//设备信息响应根节点转化
func marshalDeviceInfoResponse(p DeviceInfoResponse) ([]byte, error) {
	tmp := struct {
		DeviceInfoResponse
		XMLName struct{} `xml:"Response"`
	}{DeviceInfoResponse: p}

	return xml.MarshalIndent(tmp, "", "   ")
}

//获取DeviceInfoResponse 的xml字符串
func (deviceInfoResponse *DeviceInfoResponse) ToXml() (string, error) {
	xmlOutPut, outPutErr := marshalDeviceInfoResponse(*deviceInfoResponse)
	if outPutErr == nil {
		//加入XML头
		headerBytes := []byte(xml.Header)
		//拼接XML头和实际XML内容
		xmlOutPutData := (string)(append(headerBytes, xmlOutPut...))
		return xmlOutPutData, nil
	}
	return "", outPutErr
}

//设备状态响应
type DeviceStatusResponse struct {
	CmdType    string
	SN         uint64
	DeviceID   string
	Result     string
	Online     string
	Status     string
	DeviceTime string
	Encode     string
	Record     string
}

//设备状态响应根节点转化
func marshalDeviceStatusResponse(p DeviceStatusResponse) ([]byte, error) {
	tmp := struct {
		DeviceStatusResponse
		XMLName struct{} `xml:"Response"`
	}{DeviceStatusResponse: p}
	return xml.MarshalIndent(tmp, "", "   ")
}

//获取DeviceStatusResopnse 的xml字符串
func (deviceStatusResponse *DeviceStatusResponse) ToXml() (string, error) {
	xmlOutPut, outPutErr := marshalDeviceStatusResponse(*deviceStatusResponse)
	if outPutErr == nil {
		//加入XML头
		headerBytes := []byte(xml.Header)
		//拼接XML头和实际XML内容
		xmlOutPutData := (string)(append(headerBytes, xmlOutPut...))
		return xmlOutPutData, nil
	}
	return "", outPutErr
}
