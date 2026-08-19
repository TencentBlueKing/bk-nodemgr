/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package wmix

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
)

// testClient ...
func testClient(t *testing.T) *Client {
	err := godotenv.Load(".env")
	if err != nil {
		t.Fatal(err)
	}

	config := &Config{
		Domain:     os.Getenv("WIN-DOMAIN"),
		IP:         os.Getenv("WIN-IP"),
		User:       os.Getenv("WIN-USERNAME"),
		Password:   os.Getenv("WIN-PASSWORD"),
		Timeout:    10 * time.Second,
		AuthMethod: AuthMethodPassword,
	}

	h, err := NewClient(config)
	if err != nil {
		t.Fatal(err)
	}

	return h
}

// TestClient_RunCommand ...
func TestClient_RunCommand(t *testing.T) {
	type args struct {
		ctx     context.Context
		command string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				ctx:     context.Background(),
				command: "whoami",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := testClient(t)
			stdOut, stdErr, err := client.RunCommand(tt.args.ctx, tt.args.command)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("RunCommand() stdOut(%s), stdErr(%s)", stdOut, stdErr)
		})
	}
}

// TestClient_UploadFile ...
func TestClient_UploadFile(t *testing.T) {
	type args struct {
		ctx         context.Context
		srcFilePath string
		dstFilePath string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				ctx:         context.Background(),
				srcFilePath: "./binaries/wmiexec-amd64",
				dstFilePath: "c:/wmiexec-amd64",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := testClient(t)
			stdOut, stdErr, err := client.UploadFile(tt.args.ctx, tt.args.srcFilePath, tt.args.dstFilePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("RunCommand() stdOut(%s), stdErr(%s)", stdOut, stdErr)
		})
	}
}
