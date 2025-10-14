/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package scheduledworkflow provides storage for scheduled workflow.
package scheduledworkflow

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func testClient(t *testing.T) IHandler {
	err := godotenv.Load(".env")
	if err != nil {
		t.Fatal(err)
	}

	nCtx := context.Background()
	mongoClient, err := mongo.Connect(
		nCtx,
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

	return New(mongoClient.Database(os.Getenv("MONGO_DATABASE")))
}

var once = sync.Once{}

// prepareData for all tests.
func prepareData(t *testing.T, nCtx contextx.IContext) {
	testDatas := []*types.ScheduledWorkflow{
		{
			WorkflowID:   "1",
			WorkflowName: "schedule_sync_host",
			TriggerID:    "T-00001",
		},
		{
			WorkflowID:   "2",
			WorkflowName: "schedule_sync_biz",
			TriggerID:    "T-00002",
		},
	}

	once.Do(func() {
		h := testClient(t)
		for _, data := range testDatas {
			err := h.Create(nCtx, data)
			if err != nil {
				t.Errorf("prepareData() error = %v", err)
			}
		}
	})
}

// Test_handler_Create tests the Create method of the handler.
func Test_handler_Create(t *testing.T) {
	nCtx := contextx.New(context.Background(), contextx.WithTenantID("test"))

	prepareData(t, nCtx)

	type args struct {
		nCtx              contextx.IContext
		scheduledWorkflow *types.ScheduledWorkflow
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				nCtx: nCtx,
				scheduledWorkflow: &types.ScheduledWorkflow{
					WorkflowID:   "3",
					WorkflowName: "schedule_sync_networkarea",
					TriggerID:    "T-00003",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			err := h.Create(tt.args.nCtx, tt.args.scheduledWorkflow)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr: %v", err, tt.wantErr)
				return
			}
		})
	}
}

// Test_handler_Get tests the Create method of the handler.
func Test_handler_Get(t *testing.T) {
	nCtx := contextx.New(context.Background(), contextx.WithTenantID("test"))

	prepareData(t, nCtx)

	type args struct {
		nCtx       contextx.IContext
		workflowID string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    *types.ScheduledWorkflow
	}{
		{
			name: "normal",
			args: args{
				nCtx:       nCtx,
				workflowID: "1",
			},
			wantErr: false,
			want: &types.ScheduledWorkflow{
				WorkflowID:   "1",
				WorkflowName: "schedule_sync_host",
				TriggerID:    "T-00002",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.Get(tt.args.nCtx, tt.args.workflowID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr: %v", err, tt.wantErr)
				return
			}

			if got.WorkflowName != tt.want.WorkflowName {
				t.Errorf("Get() error, got %v, want: %v", got, tt.want)
			}

			t.Logf("Get() got = %v", got)
		})
	}
}

// Test_handler_List tests the List method of the handler.
func Test_handler_List(t *testing.T) {
	nCtx := contextx.New(context.Background(), contextx.WithTenantID("test"))

	type args struct {
		nCtx contextx.IContext
		page types.Page
		opts []OptFn
	}
	tests := []struct {
		name    string
		args    args
		want    []*types.ScheduledWorkflow
		wantNum int64
		wantErr bool
	}{
		{
			name: "filter by workflow id",
			args: args{
				nCtx: nCtx,
				page: types.Page{
					Offset: 0,
					Limit:  1,
					Sort:   "",
				},
				opts: []OptFn{
					WithWorkflowID("1"),
				},
			},
			want: []*types.ScheduledWorkflow{
				{
					WorkflowID:   "1",
					WorkflowName: "schedule_sync_host",
					TriggerID:    "T-00001",
				},
			},
			wantNum: 1,
			wantErr: false,
		},
		{
			name: "filter by oper type",
			args: args{
				nCtx: nCtx,
				page: types.Page{
					Offset: 0,
					Limit:  1,
				},
				opts: []OptFn{
					WithWorkflowID("2"),
					WithWorkflowName("schedule_sync_biz"),
				},
			},
			want: []*types.ScheduledWorkflow{
				{
					WorkflowID:   "2",
					WorkflowName: "schedule_sync_biz",
					TriggerID:    "T-00002",
				},
			},
			wantNum: 1,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, gotNum, err := h.List(tt.args.nCtx, tt.args.page, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantNum > 0 && tt.wantNum != gotNum {
				t.Errorf("List() num = %d, wantNum %d", gotNum, tt.wantNum)
				return
			}
			t.Logf("List() num = %d", gotNum)
			for _, v := range got {
				t.Logf("List() got = %v", v)
			}
		})
	}
}

// Test_handler_Count tests the Count method of the handler.
func Test_handler_Count(t *testing.T) {
	nCtx := contextx.New(context.Background(), contextx.WithTenantID("test"))

	type args struct {
		nCtx contextx.IContext
		opts []OptFn
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "filter by workflow id",
			args: args{
				nCtx: nCtx,
				opts: []OptFn{
					WithWorkflowID("1"),
				},
			},
			want:    1,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.Count(tt.args.nCtx, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Count() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.want > 0 && got != tt.want {
				t.Errorf("Count() got = %d, want %d", got, tt.want)
				return
			}
			t.Logf("Count() got = %d", got)
		})
	}
}

// Test_handler_UpdateTriggerID tests the UpdateTriggerID method of the handler.
func Test_handler_UpdateTriggerID(t *testing.T) {
	nCtx := contextx.New(context.Background(), contextx.WithTenantID("test"))

	type args struct {
		nCtx       contextx.IContext
		workflowID string
		triggerID  string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "update",
			args: args{
				nCtx:       nCtx,
				workflowID: "1",
				triggerID:  "updated trigger",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			if err := h.UpdateTriggerID(tt.args.nCtx, tt.args.workflowID, tt.args.triggerID); (err != nil) != tt.wantErr {
				t.Errorf("UpdateTriggerID() error = %v, wantErr %v", err, tt.wantErr)
			}

			sw, err := h.Get(tt.args.nCtx, tt.args.workflowID)
			if err != nil {
				t.Errorf("UpdateTriggerID() error = %v, wantErr %v", err, tt.wantErr)
			}

			if sw.TriggerID != tt.args.triggerID {
				t.Errorf("UpdateTriggerID() got = %v, want %v", sw.TriggerID, tt.args.triggerID)
			}
		})
	}
}

// Test_hander_UpdatePrivateData tests the UpdatePrivateData method of the handler.
func Test_hander_UpdatePrivateData(t *testing.T) {
	nCtx := contextx.New(context.Background(), contextx.WithTenantID("test"))

	type args struct {
		nCtx        contextx.IContext
		workflowID  string
		privateData map[string]any
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "update",
			args: args{
				nCtx:       nCtx,
				workflowID: "1",
				privateData: map[string]any{
					"test":  "1",
					"test2": 2,
					"test3": true,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			if err := h.UpdatePrivateData(tt.args.nCtx, tt.args.workflowID, tt.args.privateData); (err != nil) != tt.wantErr {
				t.Errorf("UpdatePrivateData() error = %v, wantErr %v", err, tt.wantErr)
			}

			sw, err := h.Get(tt.args.nCtx, tt.args.workflowID)
			if err != nil {
				t.Errorf("UpdatePrivateData() error = %v, wantErr %v", err, tt.wantErr)
			}

			for k := range tt.args.privateData {
				if _, ok := sw.PrivateData[k]; !ok {
					t.Errorf("UpdatePrivateData() want = %v, got nothing", k)
				}
			}
		})
	}
}
