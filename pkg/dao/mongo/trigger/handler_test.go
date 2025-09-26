/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package trigger ...
package trigger

import (
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/identifier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/trigger"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// testClient creates a client for test.
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

var globalTrigger *trigger.Trigger

// Test_handler_Create tests handler Create.
func Test_handler_Create(t *testing.T) {
	type args struct {
		nCtx context.Context
		trig *trigger.Trigger
	}

	globalTrigger = &trigger.Trigger{
		TriggerID: identifier.GenTriggerID(),
		Category:  trigger.CategoryOnce,
		Metadata:  &trigger.MetadataOnce{},
		State:     trigger.StateInit,
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "base",
			args: args{
				nCtx: context.Background(),
				trig: globalTrigger,
			},
			wantErr: false,
		},
		{
			name: "nil nCtx",
			args: args{
				nCtx: nil,
				trig: &trigger.Trigger{
					TriggerID: "trigger-nil-nCtx",
					Category:  trigger.CategoryOnce,
					Metadata:  &trigger.MetadataOnce{},
					State:     trigger.StateInit,
				},
			},
			wantErr: true,
		},
		{
			name: "nil trigger",
			args: args{
				nCtx: context.Background(),
				trig: nil,
			},
			wantErr: true,
		},
		{
			name: "invalid category",
			args: args{
				nCtx: context.Background(),
				trig: &trigger.Trigger{
					TriggerID: "trigger-invalid-category",
					Category:  "test",
					Metadata:  &trigger.MetadataOnce{},
					State:     trigger.StateInit,
				},
			},
			wantErr: true,
		},
		{
			name: "mismatch metadata",
			args: args{
				nCtx: context.Background(),
				trig: &trigger.Trigger{
					TriggerID: "trigger-invalid-state",
					Category:  trigger.CategoryOnce,
					Metadata:  &trigger.MetadataOrdered{},
					State:     trigger.StateInit,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid state",
			args: args{
				nCtx: context.Background(),
				trig: &trigger.Trigger{
					TriggerID: "trigger-invalid-state",
					Category:  trigger.CategoryOnce,
					Metadata:  &trigger.MetadataOnce{},
					State:     "test",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			if err := h.Create(tt.args.nCtx, tt.args.trig); (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test_handler_Get tests handler Get.
func Test_handler_Get(t *testing.T) {
	type args struct {
		nCtx        context.Context
		wantTrigger *trigger.Trigger
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "base",
			args: args{
				nCtx:        context.Background(),
				wantTrigger: globalTrigger,
			},
			wantErr: false,
		},
		{
			name: "not found",
			args: args{
				nCtx:        context.Background(),
				wantTrigger: &trigger.Trigger{TriggerID: "not-found"},
			},
			wantErr: true,
		},
		{
			name: "nil nCtx",
			args: args{
				nCtx:        nil,
				wantTrigger: globalTrigger,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			trig, err := h.Get(tt.args.nCtx, tt.args.wantTrigger.TriggerID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				if !reflect.DeepEqual(trig, tt.args.wantTrigger) {
					t.Errorf("Get() error, got %v, want %v", trig, tt.args.wantTrigger)
					return
				}
			}
		})
	}
}

// Test_handler_Update tests handler Update.
func Test_handler_Update(t *testing.T) {
	type args struct {
		nCtx    context.Context
		trigger *trigger.Trigger
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "base",
			args: args{
				nCtx: context.Background(),
				trigger: &trigger.Trigger{
					TriggerID: globalTrigger.TriggerID,
					Category:  trigger.CategoryOrdered,
					Metadata: &trigger.MetadataOrdered{
						MaxConcurrencyNum: 10,
					},
					State: trigger.StateRunning,
				},
			},
			wantErr: false,
		},
		{
			name: "nil nCtx",
			args: args{
				nCtx:    nil,
				trigger: globalTrigger,
			},
			wantErr: true,
		},
		{
			name: "not exist trigger",
			args: args{
				nCtx:    context.Background(),
				trigger: &trigger.Trigger{TriggerID: "not-exist"},
			},
			wantErr: true,
		},
		{
			name: "nil trigger",
			args: args{
				nCtx:    context.Background(),
				trigger: nil,
			},
			wantErr: true,
		},
		{
			name: "invalid category",
			args: args{
				nCtx: context.Background(),
				trigger: &trigger.Trigger{
					TriggerID: globalTrigger.TriggerID,
					Category:  "test",
					Metadata:  &trigger.MetadataOnce{},
					State:     trigger.StateInit,
				},
			},
			wantErr: true,
		},
		{
			name: "mismatch metadata",
			args: args{
				nCtx: context.Background(),
				trigger: &trigger.Trigger{
					TriggerID: globalTrigger.TriggerID,
					Category:  trigger.CategoryOnce,
					Metadata:  &trigger.MetadataOrdered{},
					State:     trigger.StateInit,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid state",
			args: args{
				nCtx: context.Background(),
				trigger: &trigger.Trigger{
					TriggerID: globalTrigger.TriggerID,
					Category:  trigger.CategoryOnce,
					Metadata:  &trigger.MetadataOnce{},
					State:     "test",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			err := h.Update(tt.args.nCtx, tt.args.trigger)
			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				got, err := h.Get(tt.args.nCtx, tt.args.trigger.TriggerID)
				if err != nil {
					t.Errorf("Update() error, got %v, want %v", got, tt.args.trigger)
					return
				}

				if !reflect.DeepEqual(got, tt.args.trigger) {
					t.Errorf("Update() error, got %v, want %v", got, tt.args.trigger)
					return
				}
			}
		})
	}
}

// Test_handler_List tests handler List.
func Test_handler_List(t *testing.T) {
	type args struct {
		nCtx  context.Context
		page  types.Page
		optFn []OptFn
	}

	tests := []struct {
		name      string
		args      args
		wantTotal int64
		wantNum   int
		wantErr   bool
	}{
		{
			name: "nil nCtx",
			args: args{
				nCtx:  nil,
				page:  types.Page{Limit: 10},
				optFn: nil,
			},
			wantTotal: -1,
			wantNum:   -1,
			wantErr:   true,
		},
		{
			name: "base",
			args: args{
				nCtx:  context.Background(),
				page:  types.Page{Limit: 1},
				optFn: []OptFn{WithCategory(trigger.CategoryOrdered), WithState(trigger.StateRunning)},
			},
			wantTotal: -1,
			wantNum:   1,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, total, err := h.List(tt.args.nCtx, tt.args.page, tt.args.optFn...)
			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			if tt.wantTotal >= 0 && total != tt.wantTotal {
				t.Errorf("List() total = %v, want %v", total, tt.wantTotal)
			}

			if tt.wantNum >= 0 && len(got) != tt.wantNum {
				t.Errorf("List() num = %v, want %v", len(got), tt.wantNum)
			}
		})
	}
}
