/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package basestorage ...
package basestorage

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// testClient ...
func testClient(t *testing.T) *Storage {
	err := godotenv.Load(".env")
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	mongoClient, err := mongo.Connect(
		ctx,
		&options.ClientOptions{
			Hosts: []string{
				os.Getenv("MONGO_ADDRESS"),
			},
			Auth: &options.Credential{
				Username:      os.Getenv("MONGO_USER"),
				Password:      os.Getenv("MONGO_PASSWORD"),
				AuthSource:    os.Getenv("MONGO_AUTH_SOURCE"),
				AuthMechanism: os.Getenv("MONGO_AUTH_MECHANISM"),
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	s := &Storage{
		Name:     "test",
		Database: mongoClient.Database(os.Getenv("MONGO_DATABASE")),
		Logger:   logger.LoggerDefault{},
	}

	return s
}

// TestStorage_Start ...
func TestStorage_Start(t *testing.T) {
	tests := []struct {
		name    string
		isInit  bool
		wantErr bool
	}{
		{
			name:    "init",
			isInit:  true,
			wantErr: false,
		},
		{
			name:    "without init",
			isInit:  false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testClient(t)

			if tt.isInit {
				InitStorage(s)
			}

			if err := s.Start(context.Background()); (err != nil) != tt.wantErr {
				t.Errorf("Start() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestStorage_CheckHealthz ...
func TestStorage_CheckHealthz(t *testing.T) {
	tests := []struct {
		name         string
		healthz      func() error
		isStarted    bool
		isTerminated bool
		wantErr      bool
	}{
		{
			name:         "health",
			healthz:      func() error { return nil },
			isStarted:    true,
			isTerminated: false,
			wantErr:      false,
		},
		{
			name:         "is not running",
			healthz:      func() error { return nil },
			isStarted:    false,
			isTerminated: false,
			wantErr:      true,
		},
		{
			name:         "is terminated",
			healthz:      func() error { return nil },
			isStarted:    true,
			isTerminated: true,
			wantErr:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testClient(t)
			InitStorage(s, WithCheckFunc(tt.healthz))

			if tt.isStarted {
				s.Start(context.Background())
			}

			if tt.isTerminated {
				s.Cancel()
			}

			if err := s.CheckHealthz(); (err != nil) != tt.wantErr {
				t.Errorf("CheckHealthz() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestStorage_Terminate ...
func TestStorage_Terminate(t *testing.T) {
	tests := []struct {
		name         string
		isStarted    bool
		isTerminated bool
		wantErr      bool
	}{
		{
			name:         "not started",
			isStarted:    false,
			isTerminated: false,
			wantErr:      true,
		},
		{
			name:         "has been terminated",
			isStarted:    true,
			isTerminated: true,
			wantErr:      true,
		},
		{
			name:         "normal",
			isStarted:    true,
			isTerminated: false,
			wantErr:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testClient(t)
			InitStorage(s)

			if tt.isStarted {
				if err := s.Start(context.Background()); err != nil {
					t.Errorf("Start() error = %v", err)
				}
			}

			if tt.isTerminated {
				if err := s.Terminate(); err != nil {
					t.Errorf("TerminateOperInst() error = %v", err)
				}

				time.Sleep(time.Second * 1)
			}

			if err := s.Terminate(); (err != nil) != tt.wantErr {
				t.Errorf("TerminateOperInst() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
