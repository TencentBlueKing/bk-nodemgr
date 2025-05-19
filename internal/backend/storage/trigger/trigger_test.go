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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/identifier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/trigger"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// testClient creates a client for test.
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

var globalTrigger *trigger.Trigger

// Test_storage_Create tests storage.Create.
func Test_storage_Create(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "test")

	globalTrigger = &trigger.Trigger{
		TriggerID: identifier.GenTriggerID(),
		Category:  trigger.CategoryOnce,
		Metadata:  &trigger.MetadataOnce{},
		State:     trigger.StateInit,
	}

	type args struct {
		ctx  context.Context
		trig *trigger.Trigger
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "base",
			args: args{
				ctx:  ctx,
				trig: globalTrigger,
			},
			wantErr: false,
		},
		{
			name: "nil ctx",
			args: args{
				ctx: nil,
				trig: &trigger.Trigger{
					TriggerID: "trigger-nil-ctx",
					Category:  trigger.CategoryOnce,
					Metadata:  &trigger.MetadataOnce{},
					State:     trigger.StateInit,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testClient(t)
			if err := s.CreateTrigger(tt.args.ctx, tt.args.trig); (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test_storage_Get tests storage.Get.
func Test_storage_Get(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "test")

	type args struct {
		ctx         context.Context
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
				ctx:         ctx,
				wantTrigger: globalTrigger,
			},
			wantErr: false,
		},
		{
			name: "not found",
			args: args{
				ctx:         ctx,
				wantTrigger: &trigger.Trigger{TriggerID: "not-found"},
			},
			wantErr: true,
		},
		{
			name: "nil ctx",
			args: args{
				ctx:         nil,
				wantTrigger: globalTrigger,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testClient(t)
			trig, err := s.GetTrigger(tt.args.ctx, tt.args.wantTrigger.TriggerID)
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

// Test_storage_UpdateTrigger tests storage.UpdateTrigger.
func Test_storage_UpdateTrigger(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "test")

	type args struct {
		ctx  context.Context
		trig *trigger.Trigger
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "base",
			args: args{
				ctx: ctx,
				trig: &trigger.Trigger{
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
			name: "nil ctx",
			args: args{
				ctx: nil,
				trig: &trigger.Trigger{
					TriggerID: "trigger-nil-ctx",
					Category:  trigger.CategoryOnce,
					Metadata:  &trigger.MetadataOnce{},
					State:     trigger.StateInit,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testClient(t)
			if err := s.UpdateTrigger(tt.args.ctx, tt.args.trig); (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test_storage_ListAliveTrigger tests storage.ListAliveTrigger.
func Test_storage_ListAliveTrigger(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "test")

	type args struct {
		ctx      context.Context
		category trigger.Category
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "base",
			args: args{
				ctx:      ctx,
				category: trigger.CategoryPeriodic,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testClient(t)
			got, err := s.ListAliveTrigger(tt.args.ctx, tt.args.category)
			if (err != nil) != tt.wantErr {
				t.Errorf("AllPeriodicTrigger() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for _, trigger := range got {
				t.Logf("trigger: %#v\n", trigger)
			}
		})
	}
}
