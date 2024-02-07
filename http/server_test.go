package http

import (
	"bytes"
	"chatservice/config"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

var cfg config.Config

func init() {
	cfg = config.Config{
		HttpIp:     "127.0.0.1",
		HttpPort:   "8080",
		TelNetIp:   "127.0.0.1",
		TelNetPort: "8181",
		LogFile:    "fooFile.txt",
	}
	//Init http server
	InitHttpServer(cfg)
}

func TestSubmitMessages(t *testing.T) {
	//Expecting all these to "fail". This just verifies parsing of the json
	//and logic, not the sending of a message into telnet
	tests := []struct {
		name     string
		postBody string
		expected string
	}{
		{
			"send to all",
			`{"message":"hello"}`,
			"Message submitted successfully",
		},
		{
			"send to pm",
			`{"user":"foo", "message":"hello"}`,
			"User does not exist",
		},
		{
			"send to channel",
			`{"channel":"foo", "message":"hello"}`,
			"Channel does not exist",
		},
		{
			"json parse failure",
			`{"foo:"foo"}`,
			"invalid character 'f' after object key\n",
		},
	}
	for _, tt := range tests {
		jsonBody := []byte(tt.postBody)
		bodyReader := bytes.NewReader(jsonBody)

		req := httptest.NewRequest(http.MethodPost, "/submitMessage", bodyReader)
		w := httptest.NewRecorder()
		submitMessage(w, req)
		res := w.Result()
		defer res.Body.Close()
		data, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Errorf("Error: %v", err)
		}
		if string(data) != tt.expected {
			t.Errorf("Expected "+tt.expected+" but got %v", string(data))
		}
	}
}

func TestGetLogs(t *testing.T) {
	expected := "" //Expect nothing and no errors
	req := httptest.NewRequest(http.MethodGet, "/getLogs", nil)
	w := httptest.NewRecorder()
	getLogs(w, req)
	res := w.Result()
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	if string(data) != expected {
		t.Errorf("Expected "+expected+" but got %v", string(data))
	}
}

func TestGetStats(t *testing.T) {
	expected := `{"channels":0,"messages_sent":1,"users":0}`
	req := httptest.NewRequest(http.MethodGet, "/stats", nil)
	w := httptest.NewRecorder()
	getStats(w, req)
	res := w.Result()
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	if string(data) != expected {
		t.Errorf("Expected "+expected+" but got %v", string(data))
	}
}
