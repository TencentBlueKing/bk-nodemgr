/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package utils ...
package utils

import (
	"context"
	"testing"
	"time"
)

// TestCheckDiskFreeSpace ...
func TestCheckDiskFreeSpace(t *testing.T) {
	type args struct {
		dirPath string
	}
	tests := []struct {
		name    string
		args    args
		want    uint64
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				dirPath: ".",
			},
			want:    0,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CountDiskFreeSpace(tt.args.dirPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("CountDiskFreeSpace() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CountDiskFreeSpace() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckTCPPortOpen(t *testing.T) {
	type args struct {
		host    string
		port    int
		timeout time.Duration
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				host:    "127.0.0.1",
				port:    22,
				timeout: time.Second,
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CheckNetTCPOpen(tt.args.host, tt.args.port, tt.args.timeout)
			if err != nil {
				t.Logf("CheckNetTCPOpen() error = %v", err)
				return
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("CheckNetTCPOpen() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("CheckNetTCPOpen() got = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestCheckTCPPortIdle test
func TestCheckTCPPortIdle(t *testing.T) {
	type args struct {
		port uint64
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				port: 8001,
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CheckTCPPortIdle(context.Background(), tt.args.port)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckTCPPortIdle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CheckTCPPortIdle() got = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestCheckTCP6PortIdle test
func TestCheckTCP6PortIdle(t *testing.T) {
	type args struct {
		port uint64
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				port: 3390,
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CheckTCP6PortIdle(context.Background(), tt.args.port)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckTCP6PortIdle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CheckTCP6PortIdle() got = %v, want %v", got, tt.want)
			}
		})
	}
}
