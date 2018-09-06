package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestStatusHandler00(test *testing.T) {
	request, _ := http.NewRequest("GET", "http://example.com/status", nil)
	response := httptest.NewRecorder()
	statusHandler(response, request)
	if response.Code != 200 {
		test.Error()
	}

	status := &SystemStatus{}
	err := json.Unmarshal(response.Body.Bytes(), status)
	if err != nil {
		test.Error()
	}
	if status.StreamCount != 0 {
		test.Error()
	}
	if status.SessionCount != 0 {
		test.Error()
	}
	if len(status.LiveStreams) != 0 {
		test.Error()
	}
}

func TestAddInputHandler0(test *testing.T) {
	request, _ := http.NewRequest("GET", "http://example.com/add-input?streamid=test0&islive=1&uri=udp://127.0.0.1:9002", nil)
	response := httptest.NewRecorder()
	addInputHandler(response, request)
	if response.Code != 200 {
		test.Error()
	}

	syserr := &SystemError{}
	err := json.Unmarshal(response.Body.Bytes(), syserr)
	if err != nil {
		test.Error()
	}
	if !strings.EqualFold(syserr.Content, "OK") {
		test.Error()
	}
	if syserr.StatusCode != 200 {
		test.Error()
	}
}

func TestStatusHandler01(test *testing.T) {
	request, _ := http.NewRequest("GET", "http://example.com/status", nil)
	response := httptest.NewRecorder()
	statusHandler(response, request)
	if response.Code != 200 {
		test.Error()
	}

	status := &SystemStatus{}
	err := json.Unmarshal(response.Body.Bytes(), status)
	if err != nil {
		test.Error()
	}
	if status.StreamCount != 1 {
		test.Error()
	}
	if status.SessionCount != 0 {
		test.Error()
	}
	if len(status.LiveStreams) != 1 {
		test.Error()
	}
	if len(status.LiveStreams[0].Protocols) != 0 {
		test.Error()
	}
}

func TestAddOutputHandler0(test *testing.T) {
	request, _ := http.NewRequest("GET", "http://example.com/add-output?streamid=test0&islive=1&protocol=rtsp&enable=1", nil)
	response := httptest.NewRecorder()
	addOutputHandler(response, request)
	if response.Code != 200 {
		test.Error()
	}

	syserr := &SystemError{}
	err := json.Unmarshal(response.Body.Bytes(), syserr)
	if err != nil {
		test.Error()
	}
	if !strings.EqualFold(syserr.Content, "OK") {
		test.Error()
	}
	if syserr.StatusCode != 200 {
		test.Error()
	}
}

func TestStatusHandler02(test *testing.T) {
	request, _ := http.NewRequest("GET", "http://example.com/status", nil)
	response := httptest.NewRecorder()
	statusHandler(response, request)
	if response.Code != 200 {
		test.Error()
	}

	status := &SystemStatus{}
	err := json.Unmarshal(response.Body.Bytes(), status)
	if err != nil {
		test.Error()
	}
	if status.StreamCount != 1 {
		test.Error()
	}
	if status.SessionCount != 0 {
		test.Error()
	}
	if len(status.LiveStreams) != 1 {
		test.Error()
	}
	if len(status.LiveStreams[0].Protocols) != 1 {
		test.Error()
	}
	if enable, ok := status.LiveStreams[0].Protocols["rtsp"]; enable == false || ok == false {
		test.Error()
	}
}

func TestRemoveOutputHandler0(test *testing.T) {
	request, _ := http.NewRequest("GET", "http://example.com/remove-output?streamid=test0&islive=1&protocol=rtsp", nil)
	response := httptest.NewRecorder()
	removeOutputHandler(response, request)
	if response.Code != 200 {
		test.Error()
	}

	syserr := &SystemError{}
	err := json.Unmarshal(response.Body.Bytes(), syserr)
	if err != nil {
		test.Error()
	}
	if !strings.EqualFold(syserr.Content, "OK") {
		test.Error()
	}
	if syserr.StatusCode != 200 {
		test.Error()
	}
}

func TestStatusHandler03(test *testing.T) {
	request, _ := http.NewRequest("GET", "http://example.com/status", nil)
	response := httptest.NewRecorder()
	statusHandler(response, request)
	if response.Code != 200 {
		test.Error()
	}

	status := &SystemStatus{}
	err := json.Unmarshal(response.Body.Bytes(), status)
	if err != nil {
		test.Error()
	}
	if status.StreamCount != 1 {
		test.Error()
	}
	if status.SessionCount != 0 {
		test.Error()
	}
	if len(status.LiveStreams) != 1 {
		test.Error()
	}
	if len(status.LiveStreams[0].Protocols) != 0 {
		test.Error()
	}
}

func TestRemoveInputHandler0(test *testing.T) {
	request, _ := http.NewRequest("GET", "http://example.com/remove-input?streamid=test0&islive=1", nil)
	response := httptest.NewRecorder()
	removeInputHandler(response, request)
	if response.Code != 200 {
		test.Error()
	}

	syserr := &SystemError{}
	err := json.Unmarshal(response.Body.Bytes(), syserr)
	if err != nil {
		test.Error()
	}
	if !strings.EqualFold(syserr.Content, "OK") {
		test.Error()
	}
	if syserr.StatusCode != 200 {
		test.Error()
	}
}

func TestStatusHandler04(test *testing.T) {
	request, _ := http.NewRequest("GET", "http://example.com/status", nil)
	response := httptest.NewRecorder()
	statusHandler(response, request)
	if response.Code != 200 {
		test.Error()
	}

	status := &SystemStatus{}
	err := json.Unmarshal(response.Body.Bytes(), status)
	if err != nil {
		test.Error()
	}
	if status.StreamCount != 0 {
		test.Error()
	}
	if status.SessionCount != 0 {
		test.Error()
	}
	if len(status.LiveStreams) != 0 {
		test.Error()
	}
}

func TestAddInputHandler10(test *testing.T) {
	time.Sleep(time.Second)
	request, _ := http.NewRequest("GET", "http://example.com/add-input?islive=1&uri=udp://127.0.0.1:9002", nil)
	response := httptest.NewRecorder()
	addInputHandler(response, request)
	if response.Code != 200 {
		test.Error()
	}

	syserr := &SystemError{}
	err := json.Unmarshal(response.Body.Bytes(), syserr)
	if err != nil {
		test.Error()
	}
	if !strings.EqualFold(syserr.Content, "streamid not found") {
		test.Error()
	}
	if syserr.StatusCode != 400 {
		test.Error()
	}
}

func TestAddInputHandler11(test *testing.T) {
	time.Sleep(time.Second)
	request, _ := http.NewRequest("GET", "http://example.com/add-input?streamid=test0&islive=1", nil)
	response := httptest.NewRecorder()
	addInputHandler(response, request)
	if response.Code != 200 {
		test.Error()
	}

	syserr := &SystemError{}
	err := json.Unmarshal(response.Body.Bytes(), syserr)
	if err != nil {
		test.Error()
	}
	if !strings.EqualFold(syserr.Content, "uri not found") {
		test.Error()
	}
	if syserr.StatusCode != 400 {
		test.Error()
	}
}

func TestAddInputHandler12(test *testing.T) {
	time.Sleep(time.Second)
	request, _ := http.NewRequest("GET", "http://example.com/add-input?streamid=test0&uri=udp://127.0.0.1:9002", nil)
	response := httptest.NewRecorder()
	addInputHandler(response, request)
	if response.Code != 200 {
		test.Error()
	}

	syserr := &SystemError{}
	err := json.Unmarshal(response.Body.Bytes(), syserr)
	if err != nil {
		test.Error()
	}
	if !strings.EqualFold(syserr.Content, "islive not found") {
		test.Error()
	}
	if syserr.StatusCode != 400 {
		test.Error()
	}
}

func TestAddInputHandler13(test *testing.T) {
	request, _ := http.NewRequest("GET", "http://example.com/add-input?streamid=test0&islive=1&uri=udp://127.0.0.1:9002", nil)
	response := httptest.NewRecorder()
	addInputHandler(response, request)
	if response.Code != 200 {
		test.Error()
	}

	syserr := &SystemError{}
	err := json.Unmarshal(response.Body.Bytes(), syserr)
	if err != nil {
		test.Error()
	}
	if !strings.EqualFold(syserr.Content, "OK") {
		test.Error()
	}
	if syserr.StatusCode != 200 {
		test.Error()
	}
}

func TestAddOutputHandler10(test *testing.T) {
	request, _ := http.NewRequest("GET", "http://example.com/add-output?islive=1&protocol=rtsp&enable=1", nil)
	response := httptest.NewRecorder()
	addOutputHandler(response, request)
	if response.Code != 200 {
		test.Error()
	}

	syserr := &SystemError{}
	err := json.Unmarshal(response.Body.Bytes(), syserr)
	if err != nil {
		test.Error()
	}
	if !strings.EqualFold(syserr.Content, "streamid not found") {
		test.Error()
	}
	if syserr.StatusCode != 400 {
		test.Error()
	}
}

func TestAddOutputHandler11(test *testing.T) {
	request, _ := http.NewRequest("GET", "http://example.com/add-output?streamid=test0&islive=1&enable=1", nil)
	response := httptest.NewRecorder()
	addOutputHandler(response, request)
	if response.Code != 200 {
		test.Error()
	}

	syserr := &SystemError{}
	err := json.Unmarshal(response.Body.Bytes(), syserr)
	if err != nil {
		test.Error()
	}
	if !strings.EqualFold(syserr.Content, "protocol not found") {
		test.Error()
	}
	if syserr.StatusCode != 400 {
		test.Error()
	}
}

func TestAddOutputHandler12(test *testing.T) {
	request, _ := http.NewRequest("GET", "http://example.com/add-output?streamid=test0&islive=1&protocol=rtsp&enable=2", nil)
	response := httptest.NewRecorder()
	addOutputHandler(response, request)
	if response.Code != 200 {
		test.Error()
	}

	syserr := &SystemError{}
	err := json.Unmarshal(response.Body.Bytes(), syserr)
	if err != nil {
		test.Error()
	}
	if !strings.EqualFold(syserr.Content, "enable not found") {
		test.Error()
	}
	if syserr.StatusCode != 400 {
		test.Error()
	}
}

func TestAddOutputHandler13(test *testing.T) {
	request, _ := http.NewRequest("GET", "http://example.com/add-output?streamid=test0&islive=2&protocol=rtsp&enable=1", nil)
	response := httptest.NewRecorder()
	addOutputHandler(response, request)
	if response.Code != 200 {
		test.Error()
	}

	syserr := &SystemError{}
	err := json.Unmarshal(response.Body.Bytes(), syserr)
	if err != nil {
		test.Error()
	}
	if !strings.EqualFold(syserr.Content, "islive not found") {
		test.Error()
	}
	if syserr.StatusCode != 400 {
		test.Error()
	}
}

func TestAddOutputHandler14(test *testing.T) {
	request, _ := http.NewRequest("GET", "http://example.com/add-output?streamid=test0&islive=1&protocol=rtsp&enable=1", nil)
	response := httptest.NewRecorder()
	addOutputHandler(response, request)
	if response.Code != 200 {
		test.Error()
	}

	syserr := &SystemError{}
	err := json.Unmarshal(response.Body.Bytes(), syserr)
	if err != nil {
		test.Error()
	}
	if !strings.EqualFold(syserr.Content, "OK") {
		test.Error()
	}
	if syserr.StatusCode != 200 {
		test.Error()
	}
}

func TestRemoveOutputHandler10(test *testing.T) {
	request, _ := http.NewRequest("GET", "http://example.com/remove-output?islive=1&protocol=rtsp", nil)
	response := httptest.NewRecorder()
	removeOutputHandler(response, request)
	if response.Code != 200 {
		test.Error()
	}

	syserr := &SystemError{}
	err := json.Unmarshal(response.Body.Bytes(), syserr)
	if err != nil {
		test.Error()
	}
	if !strings.EqualFold(syserr.Content, "streamid not found") {
		test.Error()
	}
	if syserr.StatusCode != 400 {
		test.Error()
	}
}

func TestRemoveOutputHandler11(test *testing.T) {
	request, _ := http.NewRequest("GET", "http://example.com/remove-output?streamid=test0&islive=1", nil)
	response := httptest.NewRecorder()
	removeOutputHandler(response, request)
	if response.Code != 200 {
		test.Error()
	}

	syserr := &SystemError{}
	err := json.Unmarshal(response.Body.Bytes(), syserr)
	if err != nil {
		test.Error()
	}
	if !strings.EqualFold(syserr.Content, "protocol not found") {
		test.Error()
	}
	if syserr.StatusCode != 400 {
		test.Error()
	}
}

func TestRemoveOutputHandler12(test *testing.T) {
	request, _ := http.NewRequest("GET", "http://example.com/remove-output?streamid=test0&protocol=rtsp", nil)
	response := httptest.NewRecorder()
	removeOutputHandler(response, request)
	if response.Code != 200 {
		test.Error()
	}

	syserr := &SystemError{}
	err := json.Unmarshal(response.Body.Bytes(), syserr)
	if err != nil {
		test.Error()
	}
	if !strings.EqualFold(syserr.Content, "islive not found") {
		test.Error()
	}
	if syserr.StatusCode != 400 {
		test.Error()
	}
}

func TestRemoveOutputHandler13(test *testing.T) {
	request, _ := http.NewRequest("GET", "http://example.com/remove-output?streamid=test0&islive=1&protocol=rtsp", nil)
	response := httptest.NewRecorder()
	removeOutputHandler(response, request)
	if response.Code != 200 {
		test.Error()
	}

	syserr := &SystemError{}
	err := json.Unmarshal(response.Body.Bytes(), syserr)
	if err != nil {
		test.Error()
	}
	if !strings.EqualFold(syserr.Content, "OK") {
		test.Error()
	}
	if syserr.StatusCode != 200 {
		test.Error()
	}
}

func TestRemoveInputHandler10(test *testing.T) {
	request, _ := http.NewRequest("GET", "http://example.com/remove-input?islive=1", nil)
	response := httptest.NewRecorder()
	removeInputHandler(response, request)
	if response.Code != 200 {
		test.Error()
	}

	syserr := &SystemError{}
	err := json.Unmarshal(response.Body.Bytes(), syserr)
	if err != nil {
		test.Error()
	}
	if !strings.EqualFold(syserr.Content, "streamid not found") {
		test.Error()
	}
	if syserr.StatusCode != 400 {
		test.Error()
	}
}

func TestRemoveInputHandler11(test *testing.T) {
	request, _ := http.NewRequest("GET", "http://example.com/remove-input?streamid=test0", nil)
	response := httptest.NewRecorder()
	removeInputHandler(response, request)
	if response.Code != 200 {
		test.Error()
	}

	syserr := &SystemError{}
	err := json.Unmarshal(response.Body.Bytes(), syserr)
	if err != nil {
		test.Error()
	}
	if !strings.EqualFold(syserr.Content, "islive not found") {
		test.Error()
	}
	if syserr.StatusCode != 400 {
		test.Error()
	}
}

func TestRemoveInputHandler12(test *testing.T) {
	request, _ := http.NewRequest("GET", "http://example.com/remove-input?streamid=test0&islive=1", nil)
	response := httptest.NewRecorder()
	removeInputHandler(response, request)
	if response.Code != 200 {
		test.Error()
	}

	syserr := &SystemError{}
	err := json.Unmarshal(response.Body.Bytes(), syserr)
	if err != nil {
		test.Error()
	}
	if !strings.EqualFold(syserr.Content, "OK") {
		test.Error()
	}
	if syserr.StatusCode != 200 {
		test.Error()
	}
}
