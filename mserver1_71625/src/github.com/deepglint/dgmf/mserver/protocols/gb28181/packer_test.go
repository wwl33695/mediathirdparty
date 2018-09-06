package gb28181

import (
	"testing"
)

func Test_BuildRegisterRequest(t *testing.T) {
	info := &UASInfo{
		ServerID:   "11000000002000000001",
		ServerIP:   "192.168.6.105",
		ServerPort: "5060",
		UserName:   "15010000004000000001",
		Password:   "123456",
		ClientIP:   "192.168.1.176",
		ClientPort: "5065",
		ChannelID:	"34020000001310000001"}

	request := info.BuildRegisterRequest()
	if len(request) <= 0 {
		t.Errorf("uasinfo BuildRegisterRequest failed, request returned length <= 0")
	}
}

func Test_BuildRegisterMD5Auth(t *testing.T) {
	info := &UASInfo{
		ServerID:   "11000000002000000001",
		ServerIP:   "192.168.6.105",
		ServerPort: "5060",
		UserName:   "15010000004000000001",
		Password:   "123456",
		ClientIP:   "192.168.1.176",
		ClientPort: "5065",
		ChannelID:	"34020000001310000001"}

	request := info.BuildRegisterMD5Auth("1100000000", "0010101022")
	if len(request) <= 0 {
		t.Errorf("uasinfo BuildRegisterMD5Auth failed, request returned length <= 0")
	}
}

func Test_BuildHeartbeat(t *testing.T) {
	info := &UASInfo{
		ServerID:   "11000000002000000001",
		ServerIP:   "192.168.6.105",
		ServerPort: "5060",
		UserName:   "15010000004000000001",
		Password:   "123456",
		ClientIP:   "192.168.1.176",
		ClientPort: "5065",
		ChannelID:	"34020000001310000001"}

	request := info.BuildHeartbeat()
	if len(request) <= 0 {
		t.Errorf("uasinfo BuildHeartbeat failed, request returned length <= 0")
	}
}

func Test_BuildQueryDeviceRequest(t *testing.T) {
	info := &UASInfo{
		ServerID:   "11000000002000000001",
		ServerIP:   "192.168.6.105",
		ServerPort: "5060",
		UserName:   "15010000004000000001",
		Password:   "123456",
		ClientIP:   "192.168.1.176",
		ClientPort: "5065",
		ChannelID:	"34020000001310000001"}

	request := info.BuildQueryDeviceRequest()
	if len(request) <= 0 {
		t.Errorf("uasinfo BuildQueryDeviceRequest failed, request returned length <= 0")
	}
}

func Test_BuildInviteRequest(t *testing.T) {
	info := &UASInfo{
		ServerID:   "11000000002000000001",
		ServerIP:   "192.168.6.105",
		ServerPort: "5060",
		UserName:   "15010000004000000001",
		Password:   "123456",
		ClientIP:   "192.168.1.176",
		ClientPort: "5065",
		ChannelID:	"34020000001310000001"}

	request := info.BuildInviteRequest("20010", "0010101022")
	if len(request) <= 0 {
		t.Errorf("uasinfo BuildInviteRequest failed, request returned length <= 0")
	}
}

func Test_BuildACKRequest(t *testing.T) {
	info := &UASInfo{
		ServerID:   "11000000002000000001",
		ServerIP:   "192.168.6.105",
		ServerPort: "5060",
		UserName:   "15010000004000000001",
		Password:   "123456",
		ClientIP:   "192.168.1.176",
		ClientPort: "5065",
		ChannelID:	"34020000001310000001"}

	request := info.BuildACKRequest("0010101022")
	if len(request) <= 0 {
		t.Errorf("uasinfo BuildACKRequest failed, request returned length <= 0")
	}
}

func Test_BuildBYERequest(t *testing.T) {
	info := &UASInfo{
		ServerID:   "11000000002000000001",
		ServerIP:   "192.168.6.105",
		ServerPort: "5060",
		UserName:   "15010000004000000001",
		Password:   "123456",
		ClientIP:   "192.168.1.176",
		ClientPort: "5065",
		ChannelID:	"34020000001310000001"}

	request := info.BuildBYERequest("0010101022")
	if len(request) <= 0 {
		t.Errorf("uasinfo BuildBYERequest failed, request returned length <= 0")
	}
}
