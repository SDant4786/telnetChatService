package telnet

import (
	"bytes"
	"chatservice/config"
	"net"
	"sync"
	"testing"
	"time"
)

var cfg config.Config

func init() {
	cfg = config.Config{
		HttpIp:     "127.0.0.1",
		HttpPort:   "8080",
		TelNetIp:   "127.0.0.0",
		TelNetPort: "8181",
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go InitTelnetServer(cfg, nil, &wg)
	wg.Wait()
}

// Since many tests need to be done sequentially, test everything as once
func TestInitTelNetServer(t *testing.T) {
	conn, err := net.Dial("tcp", cfg.TelNetIp+":"+cfg.TelNetPort)
	if err != nil {
		t.Error("could not connect to TCP server: ", err)
	}
	defer conn.Close()

	tests := []struct {
		name    string
		payload [][]byte
		want    []byte
	}{
		{
			"startup",
			[][]byte{
				[]byte(""),
			},
			[]byte("Enter username:"),
		},
		{
			"create user",
			[][]byte{
				[]byte("foouser\n"),
			},
			[]byte("/listchannels"),
		},
		{
			"list user",
			[][]byte{
				[]byte("/listusers\n"),
			},
			[]byte("foouser"),
		},
		{
			"create channel",
			[][]byte{
				[]byte("/create\n"),
				[]byte("foochannel\n"),
			},
			[]byte("Channel: foochannel created"),
		},
		{
			"list channels",
			[][]byte{
				[]byte("/listchannels\n"),
			},
			[]byte("foochannel"),
		},
		{
			"join channel",
			[][]byte{
				[]byte("/join\n"),
				[]byte("foochannel\n"),
			},
			[]byte("Joined channel: foochannel"),
		},
		{
			"list my channels",
			[][]byte{
				[]byte("/listmychannels\n"),
			},
			[]byte("foochannel"),
		},
		{
			"send to channel",
			[][]byte{
				[]byte("/sendchannel\n"),
				[]byte("foochannel\n"),
				[]byte("fooMessage\n"),
			},
			[]byte("fooMessage"),
		},
		{
			"leave channel",
			[][]byte{
				[]byte("/leave\n"),
				[]byte("foochannel\n"),
			},
			[]byte("Left channel: foochannel"),
		},
		{
			"help menu",
			[][]byte{
				[]byte("/help\n"),
			},
			[]byte("/listchannels"),
		},
	}
	for _, tt := range tests {
		//Send in whats needed
		for _, send := range tt.payload {
			conn.Write(send)
			time.Sleep(time.Second / 10)
		}
		//Verify its what is expected
		out := make([]byte, 1024)
		if _, err := conn.Read(out); err == nil {
			if !bytes.Contains(out, tt.want) {
				t.Error(tt.name+" test failed. got: "+string(out)+" want: ", string(tt.want))
			}
		}
	}

	//2 user stuff
	conn2, err := net.Dial("tcp", cfg.TelNetIp+":"+cfg.TelNetPort)
	if err != nil {
		t.Error("could not connect to TCP server: ", err)
	}
	defer conn2.Close()

	//Create user
	conn2.Write([]byte("baruser\n"))
	time.Sleep(time.Second / 10)
	//Create channel
	conn2.Write([]byte("/create\n"))
	time.Sleep(time.Second / 10)
	conn2.Write([]byte("foochannel\n"))
	time.Sleep(time.Second / 10)
	//Join channels
	conn.Write([]byte("/join\n"))
	conn2.Write([]byte("/join\n"))
	time.Sleep(time.Second / 10)
	conn.Write([]byte("foochannel\n"))
	conn2.Write([]byte("foochannel\n"))
	time.Sleep(time.Second / 10)

	tests2 := []struct {
		name       string
		u1Action   [][]byte
		u2Action   [][]byte
		u1Response []byte
		u2Response []byte
	}{
		{
			"receive from all",
			[][]byte{
				[]byte("fooMessage\n"),
			},
			[][]byte{
				[]byte("\n"),
			},
			[]byte(""),
			[]byte("fooMessage"),
		},
		{
			"receive from pm",
			[][]byte{
				[]byte("/pm\n"),
				[]byte("baruser\n"),
				[]byte("fooMessage\n"),
			},
			[][]byte{
				[]byte("\n"),
			},
			[]byte(""),
			[]byte("fooMessage"),
		},
		{
			"receive from channel",
			[][]byte{
				[]byte("/sendchannel\n"),
				[]byte("foochannel\n"),
				[]byte("fooMessage\n"),
			},
			[][]byte{
				[]byte("\n"),
			},
			[]byte(""),
			[]byte("fooMessage"),
		},
		{
			"ignore from all",
			[][]byte{
				[]byte("/ignore\n"),
				[]byte("baruser\n"),
			},
			[][]byte{
				[]byte("fooMessage\n"),
			},
			[]byte(""),
			[]byte(""),
		},
		{
			"ignore from pm",
			[][]byte{
				[]byte("\n"),
			},
			[][]byte{
				[]byte("/pm\n"),
				[]byte("foouser\n"),
				[]byte("fooMessage\n"),
			},
			[]byte(""),
			[]byte(""),
		},
		{
			"ignore from channel",
			[][]byte{
				[]byte("\n"),
			},
			[][]byte{
				[]byte("/sendchannel\n"),
				[]byte("foochannel\n"),
				[]byte("fooMessage\n"),
			},
			[]byte(""),
			[]byte(""),
		},
	}
	for _, tt := range tests2 {
		//First user actions
		for _, send := range tt.u1Action {
			conn.Write(send)
			time.Sleep(time.Second / 10)
		}
		//Second user actions
		for _, send := range tt.u2Action {
			conn2.Write(send)
			time.Sleep(time.Second / 10)
		}
		//First user expectations
		out := make([]byte, 1024)
		if _, err := conn.Read(out); err == nil {
			if !bytes.Contains(out, tt.u1Response) {
				t.Error(tt.name+" test failed. got: "+string(out)+" want: ", string(tt.u1Response))
			}
		}
		//Second user expectations
		out = make([]byte, 1024)
		if _, err := conn2.Read(out); err == nil {
			if !bytes.Contains(out, tt.u2Response) {
				t.Error(tt.name+" test failed. got: "+string(out)+" want: ", string(tt.u2Response))
			}
		}
	}

	//HTTP tests
	HTTPSendChannelMessage("hello", "foochannel", Channels["foochannel"])
	time.Sleep(time.Second / 10)
	out := make([]byte, 1024)
	if _, err := conn.Read(out); err == nil {
		if !bytes.Contains(out, []byte("http")) {
			t.Error("HTTPSendChannelMessage test failed. got: " + string(out) + " want: ")
		}
	}
	HTTPSendUserMessage("hello", Users["foouser"])
	time.Sleep(time.Second / 10)
	out = make([]byte, 1024)
	if _, err := conn.Read(out); err == nil {
		if !bytes.Contains(out, []byte("http")) {
			t.Error("HTTPSendChannelMessage test failed. got: " + string(out) + " want: ")
		}
	}
	HTTPSendAllMessage("hello")
	time.Sleep(time.Second / 10)
	out = make([]byte, 1024)
	if _, err := conn.Read(out); err == nil {
		if !bytes.Contains(out, []byte("http")) {
			t.Error("HTTPSendChannelMessage test failed. got: " + string(out) + " want: ")
		}
	}

	//Close connections
	conn.Write([]byte("/quit\n"))
	conn2.Write([]byte("/quit\n"))
}
