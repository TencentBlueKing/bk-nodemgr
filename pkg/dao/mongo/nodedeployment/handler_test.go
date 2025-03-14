/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package nodedeployment ...
package nodedeployment

import (
	"context"
	"os"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// testClient ...
func testClient(t *testing.T) IHandler {
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

	return New(mongoClient.Database(os.Getenv("MONGO_DATABASE")), logger.LoggerDefault{})
}

// TestHandler_GetPreSetting ...
func TestHandler_GetPreSetting(t *testing.T) {
	type args struct {
		ctx   context.Context
		Token string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				ctx:   context.Background(),
				Token: "123",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.GetPreSetting(tt.args.ctx, tt.args.Token)
			if err != nil {
				t.Logf("err: %+v", err)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("GetPreSetting() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("got: %+v", got)
		})
	}
}

// TestHandler_Create ...
func TestHandler_Create(t *testing.T) {
	type args struct {
		ctx            context.Context
		nodeDeployment *types.NodeDeployment
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx: context.Background(),
				nodeDeployment: &types.NodeDeployment{
					Token: "123",
					DeploymentInfo: &types.DeploymentInfo{
						OperInstID: "123",
						ActionName: "wait agent install",
					},
					NodePreSetting: &types.NodePreSetting{
						CheckList: map[string]any{
							"port": 8080,
						},
						AgentConf: map[string]any{
							"port": 8081,
						},
						DataProxyConf: map[string]any{
							"port": 8082,
						},
						FileProxyConf: map[string]any{
							"port": 8083,
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			if err := h.Create(tt.args.ctx, tt.args.nodeDeployment); (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestHandler_GetInfo ...
func TestHandler_GetInfo(t *testing.T) {

	type args struct {
		ctx   context.Context
		Token string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx:   context.Background(),
				Token: "123",
			},
			wantErr: false,
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.GetInfo(tt.args.ctx, tt.args.Token)
			if err != nil {
				t.Logf("err: %v", err)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("GetInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("got: %+v", got)
		})
	}
}
