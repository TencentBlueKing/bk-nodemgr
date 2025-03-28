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
package operinstdata

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operengine"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// testClient ...
func testClient(t *testing.T) IStorage {
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

	s, err := NewStorage(mongoClient, "test", logger.LoggerDefault{})
	if err != nil {
		t.Fatal(err)
	}

	if err = s.Start(ctx); err != nil {
		t.Fatal(err)
	}

	return s
}

func Test_storage_UpsertOperInstData(t *testing.T) {
	type args struct {
		ctx  context.Context
		data *operengine.OperInstData
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
				data: &operengine.OperInstData{
					OperInstID:        "",
					OperDefName:       "",
					ActionNames:       nil,
					ActionInstDataMap: nil,
					ParentOperationID: "",
					Timeout:           0,
					InitContent:       map[string]any{},
					Lifecycle: &operengine.Lifecycle{
						State:     operengine.OperInstStateInit,
						CreatedAt: time.Time{},
						StartedAt: time.Time{},
						EndedAt:   time.Time{},
						StoppedAt: time.Time{},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testClient(t)
			if err := s.UpsertOperInstData(tt.args.ctx, tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("UpsertOperInstData() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
