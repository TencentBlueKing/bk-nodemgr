//go:build integration

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

package packageevent

import (
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/testsuite/support"
)

func testClient(t *testing.T) IHandler {
	t.Helper()

	_, database := support.RequireMongoDatabase(t)

	return New(database)
}

func prepareData(t *testing.T, nCtx contextx.IContext, h IHandler) {
	t.Helper()

	err := h.CreateMany(nCtx,
		&types.PackageEvent{
			EventType:   types.PackageEventTypePublish,
			ReleaseType: types.ReleaseTypeAgent,
			Generation:  2,
			OSType:      criteria.OSLinux,
			CPUArch:     criteria.CPUArchAmd64,
			Version:     "1.0.0",
			Operator:    "admin",
			OperateTime: time.Date(2024, 3, 1, 0, 0, 0, 0, time.Local),
		},
		&types.PackageEvent{
			EventType:   types.PackageEventTypePublish,
			ReleaseType: types.ReleaseTypeAgent,
			Generation:  2,
			OSType:      criteria.OSLinux,
			CPUArch:     criteria.CPUArchAmd64,
			Version:     "2.0.0",
			Operator:    "admin",
			OperateTime: time.Date(2025, 1, 1, 0, 0, 0, 0, time.Local),
		},
		&types.PackageEvent{
			EventType:   types.PackageEventTypePublish,
			ReleaseType: types.ReleaseTypeProxy,
			Generation:  2,
			OSType:      criteria.OSLinux,
			CPUArch:     criteria.CPUArchAmd64,
			Version:     "3.0.0",
			Operator:    "admin",
			OperateTime: time.Date(2025, 1, 1, 0, 0, 0, 0, time.Local),
		},
		&types.PackageEvent{
			EventType:   types.PackageEventTypePublish,
			ReleaseType: types.ReleaseTypeProxy,
			Generation:  2,
			OSType:      criteria.OSLinux,
			CPUArch:     criteria.CPUArchAmd64,
			Version:     "4.0.0",
			Operator:    "admin",
			OperateTime: time.Date(2025, 1, 1, 0, 0, 0, 0, time.Local),
		},
	)
	if err != nil {
		t.Fatalf("prepare package events: %v", err)
	}
}

// Test_handler_List list event by page and conditions.
func Test_handler_List(t *testing.T) {
	nCtx := contextx.New(t.Context(), contextx.WithTenantID("test"))
	h := testClient(t)
	prepareData(t, nCtx, h)

	tests := []struct {
		name      string
		nCtx      contextx.IContext
		page      types.Page
		optFn     []OptFn
		wantTotal int64
		wantNum   int64
		wantErr   bool
	}{
		{
			name: "nil nCtx",
			nCtx: nil,
			page: types.Page{
				Offset: 0,
				Limit:  0,
			},
			optFn:     nil,
			wantTotal: -1,
			wantNum:   -1,
			wantErr:   true,
		},
		{
			name: "normal",
			nCtx: nCtx,
			page: types.Page{
				Offset: 0,
				Limit:  0,
			},
			optFn:     nil,
			wantTotal: -1,
			wantNum:   -1,
			wantErr:   false,
		},
		{
			name: "filter by release type",
			nCtx: nCtx,
			page: types.Page{
				Offset: 0,
				Limit:  1,
			},
			optFn:     []OptFn{WithReleaseType(types.ReleaseTypeAgent)},
			wantTotal: -1,
			wantNum:   1,
			wantErr:   false,
		},
		{
			name: "filter by event type",
			nCtx: nCtx,
			page: types.Page{
				Offset: 0,
				Limit:  4,
			},
			optFn:     []OptFn{WithEventType(types.PackageEventTypePublish)},
			wantTotal: -1,
			wantNum:   4,
			wantErr:   false,
		},
		{
			name: "filter with time range",
			nCtx: nCtx,
			page: types.UnlimitedPage(),
			optFn: []OptFn{WithOperateTimeRange(types.TimeRange{
				StartTime: time.Date(2022, 1, 1, 0, 0, 0, 0, time.Local),
				EndTime:   time.Date(2026, 5, 1, 0, 0, 0, 0, time.Local),
			})},
			wantTotal: -1,
			wantNum:   4,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, total, err := h.List(tt.nCtx, tt.page, tt.optFn...)
			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantTotal > 0 && tt.wantTotal != total {
				t.Errorf("List() total = %d, wantTotal %d", total, tt.wantTotal)
				return
			}
			t.Logf("List() total = %d", total)

			if tt.wantNum > 0 && tt.wantNum != int64(len(got)) {
				t.Errorf("List() num = %d, wantNum %d", len(got), tt.wantNum)
				return
			}
			t.Logf("List() num = %d", len(got))

			for _, v := range got {
				t.Logf("List() got = %v", v)
			}
		})
	}
}

// Test_handler_Count tests the count with filter.
func Test_handler_Count(t *testing.T) {
	nCtx := contextx.New(t.Context(), contextx.WithTenantID("test"))
	h := testClient(t)
	prepareData(t, nCtx, h)

	tests := []struct {
		name    string
		nCtx    contextx.IContext
		optFn   []OptFn
		wantErr bool
	}{
		{
			name:    "nil nCtx",
			nCtx:    nil,
			optFn:   nil,
			wantErr: true,
		},
		{
			name:    "normal",
			nCtx:    nCtx,
			optFn:   nil,
			wantErr: false,
		},
		{
			name:    "filter by release type",
			nCtx:    nCtx,
			optFn:   []OptFn{WithReleaseType(types.ReleaseTypeAgent)},
			wantErr: false,
		},
		{
			name:    "filter by event type",
			nCtx:    nCtx,
			optFn:   []OptFn{WithEventType(types.PackageEventTypePublish)},
			wantErr: false,
		},
		{
			name:    "filter by os type",
			nCtx:    nCtx,
			optFn:   []OptFn{WithOSType(criteria.OSLinux)},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := h.Count(tt.nCtx, tt.optFn...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Count() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("Count() total = %d", got)
		})
	}
}

// Test_handler_DistinctEventType tests the distinct with type field.
func Test_handler_DistinctEventType(t *testing.T) {
	nCtx := contextx.New(t.Context(), contextx.WithTenantID("test"))
	h := testClient(t)
	prepareData(t, nCtx, h)

	type args struct {
		nCtx  contextx.IContext
		optFn []OptFn
	}
	tests := []struct {
		name       string
		args       args
		wantResult []types.PackageEventType
		wantErr    bool
	}{
		{
			name: "invalid nCtx",
			args: args{
				nCtx:  nil,
				optFn: []OptFn{},
			},
			wantResult: nil,
			wantErr:    true,
		},
		{
			name: "normal",
			args: args{
				nCtx:  nCtx,
				optFn: []OptFn{},
			},
			wantResult: []types.PackageEventType{types.PackageEventTypePublish},
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, err := h.DistinctEventType(tt.args.nCtx, tt.args.optFn...)
			if (err != nil) != tt.wantErr {
				t.Errorf("DistinctType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			sort.Slice(gotResult, func(i, j int) bool { return gotResult[i] < gotResult[j] })
			sort.Slice(tt.wantResult, func(i, j int) bool { return tt.wantResult[i] < tt.wantResult[j] })

			if !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("DistinctType() gotResult = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

// Test_handler_DistinctReleaseType tests the distinct with type field.
func Test_handler_DistinctReleaseType(t *testing.T) {
	nCtx := contextx.New(t.Context(), contextx.WithTenantID("test"))
	h := testClient(t)
	prepareData(t, nCtx, h)

	type args struct {
		nCtx  contextx.IContext
		optFn []OptFn
	}
	tests := []struct {
		name       string
		args       args
		wantResult []types.ReleaseType
		wantErr    bool
	}{
		{
			name: "invalid nCtx",
			args: args{
				nCtx:  nil,
				optFn: []OptFn{},
			},
			wantResult: nil,
			wantErr:    true,
		},
		{
			name: "normal",
			args: args{
				nCtx:  nCtx,
				optFn: []OptFn{},
			},
			wantResult: []types.ReleaseType{types.ReleaseTypeAgent, types.ReleaseTypeProxy},
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, err := h.DistinctReleaseType(tt.args.nCtx, tt.args.optFn...)
			if (err != nil) != tt.wantErr {
				t.Errorf("DistinctType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			sort.Slice(gotResult, func(i, j int) bool { return gotResult[i] < gotResult[j] })
			sort.Slice(tt.wantResult, func(i, j int) bool { return tt.wantResult[i] < tt.wantResult[j] })

			if !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("DistinctType() gotResult = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

// Test_handler_DistinctOSType tests the distinct with type field.
func Test_handler_DistinctOSType(t *testing.T) {
	nCtx := contextx.New(t.Context(), contextx.WithTenantID("test"))
	h := testClient(t)
	prepareData(t, nCtx, h)

	type args struct {
		nCtx  contextx.IContext
		optFn []OptFn
	}
	tests := []struct {
		name       string
		args       args
		wantResult []criteria.OSType
		wantErr    bool
	}{
		{
			name: "invalid nCtx",
			args: args{
				nCtx:  nil,
				optFn: []OptFn{},
			},
			wantResult: nil,
			wantErr:    true,
		},
		{
			name: "normal",
			args: args{
				nCtx:  nCtx,
				optFn: []OptFn{},
			},
			wantResult: []criteria.OSType{criteria.OSLinux},
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, err := h.DistinctOsType(tt.args.nCtx, tt.args.optFn...)
			if (err != nil) != tt.wantErr {
				t.Errorf("DistinctType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			sort.Slice(gotResult, func(i, j int) bool { return gotResult[i] < gotResult[j] })
			sort.Slice(tt.wantResult, func(i, j int) bool { return tt.wantResult[i] < tt.wantResult[j] })

			if !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("DistinctType() gotResult = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

// Test_handler_DistinctOperator tests the distinct with operator field.
func Test_handler_DistinctOperator(t *testing.T) {
	nCtx := contextx.New(t.Context(), contextx.WithTenantID("test"))
	h := testClient(t)
	prepareData(t, nCtx, h)
	type args struct {
		nCtx  contextx.IContext
		optFn []OptFn
	}
	tests := []struct {
		name       string
		args       args
		wantResult []string
		wantErr    bool
	}{
		{
			name: "invalid nCtx",
			args: args{
				nCtx:  nil,
				optFn: []OptFn{},
			},
			wantResult: nil,
			wantErr:    true,
		},
		{
			name: "normal",
			args: args{
				nCtx:  nCtx,
				optFn: []OptFn{},
			},
			wantResult: []string{"admin"},
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, err := h.DistinctOperator(tt.args.nCtx, tt.args.optFn...)
			if (err != nil) != tt.wantErr {
				t.Errorf("DistinctOperator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			sort.Slice(gotResult, func(i, j int) bool { return gotResult[i] < gotResult[j] })
			sort.Slice(tt.wantResult, func(i, j int) bool { return tt.wantResult[i] < tt.wantResult[j] })

			if !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("DistinctOperator() gotResult = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}
