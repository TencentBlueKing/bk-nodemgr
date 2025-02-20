/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package workflowdef ...
package workflowdef

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/discovery"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/ssl"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/cmdb"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operengine"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type testHeaderSetter struct{}

// GetAuthHeader ...
func (testHeaderSetter) GetAuthHeader() (string, error) {
	return os.Getenv("BK_APIGW_AUTHHEADER"), nil
}

// testClient ...
func testClient(t *testing.T) operengine.ActionDef {
	err := godotenv.Load(".env")
	if err != nil {
		t.Fatal(err)
	}

	ctx, _ := tenant.SetID(context.Background(), "bk_nodeman")
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

	httpClient, err := client.NewClient(&ssl.TLSConfig{
		InsecureSkipVerify: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	topoStorage, err := topo.NewStorage(mongoClient, "bk_nodeman", logger.LoggerDefault{})
	if err != nil {
		t.Fatal(err)
	}

	topoStorage.Start(ctx)

	clientCap := &client.Capability{
		Client:               httpClient,
		Discover:             discovery.NewDiscovery("apigateway", []string{os.Getenv("BK_APIGW_ENDPOINT")}),
		ToleranceLatencyTime: client.ToleranceLatencyTimeDefault,
		MetricOpts:           client.MetricOption{},
		Logger:               logger.LoggerDefault{},
	}

	cmdbHandler, err := cmdb.New(clientCap, &cmdb.Config{
		TenantID:     os.Getenv("BK_APIGW_TENANT_ID"),
		HeaderSetter: testHeaderSetter{},
	})

	return NewActionSyncHostFromCMDB(cmdbHandler, topoStorage)
}

// Test_syncHostFromCMDB_Do ...
func Test_syncHostFromCMDB_Do(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "bk_nodeman")

	type fields struct {
		cmdbHandler cmdb.Handler
		topoStorage topo.Storage
	}
	type args struct {
		ctx *operengine.ActionInstContext
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "normal",
			fields: fields{},
			args: args{
				ctx: &operengine.ActionInstContext{
					Ctx: ctx,
					Data: &operengine.ActionInstData{
						TriggerID:  "",
						OperInstID: "",
						Name:       "",
						Index:      0,
						Lifecycle: &operengine.ActInstLifeCycle{
							State:     operengine.ActionInstStateRunning,
							StartedAt: time.Time{},
							EndedAt:   time.Time{},
							StoppedAt: time.Time{},
						},
						Messages: nil,
						Content: map[string]any{
							"biz_id":   2,
							"biz_name": "test",
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testClient(t)
			if err := s.Do(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Do() error = %v, wantErr %v", err, tt.wantErr)
			}

			t.Logf("Data %#v", tt.args.ctx.Data)
		})
	}
}
