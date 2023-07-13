// Copyright © 2022 The sealos Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package udp

import (
	"net"
	"strings"
	"testing"
	"time"

	"github.com/labring/operator-sdk/probe"
)

const maxBufferSize = 1024

func udpServer(t *testing.T, addr string) {
	listener, err := net.ListenPacket("udp", addr)
	if err != nil {
		return
	}
	defer listener.Close()

	doneChan := make(chan error, 1)
	buffer := make([]byte, maxBufferSize)

	go func() {
		for {
			//read
			n, caddr, err := listener.ReadFrom(buffer)
			if err != nil {
				doneChan <- err
				return
			}
			str1 := string(buffer[:n])
			if strings.ToLower(str1) == "toerror" {
				t.Log("nothing no send ...")
				buffer = []byte{}
				n = 0
				break
			} else {
				t.Logf("packet-received: bytes=%d from=%s", n, caddr.String())
			}

			// write
			deadline := time.Now().Add(time.Duration(1) * time.Second)
			err = listener.SetWriteDeadline(deadline)
			if err != nil {
				doneChan <- err
				return
			}

			n, err = listener.WriteTo(buffer[:n], caddr)
			if err != nil {
				doneChan <- err
				return
			}

			t.Logf("packet-written: bytes=%d to=%s\n", n, caddr.String())

		}
	}()
	for {
		if <-doneChan == err {
			return
		}
	}
}

func TestUDPProbe(t *testing.T) {
	type args struct {
		addr     string
		testData string
		timeout  time.Duration
	}
	tests := []struct {
		name    string
		args    args
		want    probe.Result
		want1   string
		wantErr bool
	}{
		// TODO: Add test cases.
		{name: "UDPTestWithdata_Successed", args: struct {
			addr     string
			testData string
			timeout  time.Duration
		}{addr: "127.0.0.1:38888", testData: "ABCD", timeout: 1}, want: probe.Success, want1: "", wantErr: false},

		{name: "UDPTestWithoutData_Failed", args: struct {
			addr     string
			testData string
			timeout  time.Duration
		}{addr: "127.0.0.1:38889", testData: "toerror", timeout: 1}, want: probe.Failure, want1: "io read timout", wantErr: false},
	}
	// start two udp server
	go udpServer(t, "127.0.0.1:38888")
	go udpServer(t, "127.0.0.1:38889")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := DoUDPProbe(tt.args.addr, []byte(tt.args.testData), tt.args.timeout)

			if (err != nil) != tt.wantErr {
				t.Errorf("DoUDPProbe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DoUDPProbe() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("DoUDPProbe() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
