/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package operinstdata ...
package relayhandler

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/proto/relay"
	"github.com/redis/go-redis/v9"
)

// testClient ...
func testClient(t *testing.T) IServerMessager {
	ctx := contextx.New(context.Background())
	proxyMessanger := NewServerMessager(ServerMessagerConfig{
		SlotID:        0,
		Token:         "",
		AppCode:       "",
		AppSecret:     "",
		GSEBaseURL:    "",
		SkipTLSVerify: true,
		RedisClient: redis.NewClient(&redis.Options{
			Addr:     "",
			Password: "",
			DB:       0,
		}), // Use a mock or real redis client in actual tests.
	})

	if err := proxyMessanger.Start(ctx); err != nil {
		t.Fatal(err)
	}

	return proxyMessanger
}

func TestServerPushMessage(t *testing.T) {
	type args struct {
		ctx contextx.IContext
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		agentID string
	}{
		{
			name: "echo message",
			args: args{
				ctx: contextx.New(context.Background()),
			},
			wantErr: false,
			agentID: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testClient(t)
			echoEvent := &relay.CheckPkgStateReq{

				ActionName: "action-1",

				OperInstID: "oper_inst_id-1",

				// FileList describes the file list.
				FileList: []relay.FileInfo{
					{
						FileName: "file-1",
						FileMD5:  "file_md5-1",
					},
					{
						FileName: "file-2",
						FileMD5:  "file_md5-2",
					},
				},
			}
			data, err := json.Marshal(echoEvent)
			if err != nil {
				t.Fatalf("failed to marshal echo event: %v", err)
			}

			errCh := s.PushToClient(tt.args.ctx, relay.ServerPushEventTypeCheckPkgState, data, tt.agentID)

			select {
			case err := <-errCh:
				if (err != nil) != tt.wantErr {
					t.Errorf("ServerPushMessage() error = %v, wantErr %v", err, tt.wantErr)
				}
			case <-time.After(3 * time.Second):
				if tt.wantErr {
					t.Error("Expected error but got none")
				}
			}
		})
	}
}
