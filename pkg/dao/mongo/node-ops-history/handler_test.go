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

package nodeopshistory

import (
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/testsuite/support"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestHandler_RejectsInvalidContext(t *testing.T) {
	h := new(Handler)
	calls := map[string]func(contextx.IContext) error{
		"Create": func(ctx contextx.IContext) error { return h.Create(ctx, historyFixture()) },
		"GetWithFields": func(ctx contextx.IContext) error {
			_, err := h.GetWithFields(ctx, "op-1", nil)
			return err
		},
		"ListWithFields": func(ctx contextx.IContext) error {
			_, _, err := h.ListWithFields(ctx, types.SingleItemPage(), nil)
			return err
		},
		"Count": func(ctx contextx.IContext) error { _, err := h.Count(ctx); return err },
	}
	for name, call := range calls {
		t.Run(name, func(t *testing.T) {
			require.Error(t, call(nil))
			require.Error(t, call(contextx.New(t.Context())))
		})
	}
}

func TestHandler_RejectsInvalidInput(t *testing.T) {
	h := new(Handler)
	ctx := contextx.New(t.Context(), contextx.WithTenantID("tenant-a"))
	for _, history := range []*types.NodeOpsHistory{
		nil,
		{TenantID: "tenant-a"},
		{TenantID: "tenant-a", Snapshots: []*types.NodeOpsHistoryHostSnapshot{nil}},
		{TenantID: "tenant-b", Snapshots: []*types.NodeOpsHistoryHostSnapshot{{HostID: 1}}},
	} {
		require.Error(t, h.Create(ctx, history))
	}
	_, err := h.GetWithFields(ctx, "", nil)
	require.Error(t, err)
	for _, page := range []types.Page{{Limit: 0}, {Limit: -1}, {Offset: -1, Limit: 1}} {
		_, _, err := h.ListWithFields(ctx, page, nil)
		require.Error(t, err)
	}
}

func TestHandler_CreateAndGetWithFields(t *testing.T) {
	h, ctx := testHistoryHandler(t)
	full := historyFixture()
	require.NoError(t, h.Create(ctx, full))
	tests := []struct {
		name      string
		selection *types.NodeOpsHistoryFieldSelection
		want      *types.NodeOpsHistory
	}{
		{name: "nil selection", want: full},
		{name: "empty selection", selection: &types.NodeOpsHistoryFieldSelection{}, want: full},
		{name: "snapshot detail",
			selection: &types.NodeOpsHistoryFieldSelection{NodeRole: true, Snapshots: true},
			want: &types.NodeOpsHistory{
				Info: types.NodeOpsHistoryInfo{NodeRole: full.Info.NodeRole}, Snapshots: full.Snapshots,
			},
		},
		{name: "business IDs only",
			selection: &types.NodeOpsHistoryFieldSelection{BizIDs: true},
			want: &types.NodeOpsHistory{
				Info: types.NodeOpsHistoryInfo{BizIDs: full.Info.BizIDs}, Snapshots: []*types.NodeOpsHistoryHostSnapshot{},
			},
		},
		{name: "tenant only",
			selection: &types.NodeOpsHistoryFieldSelection{TenantID: true},
			want:      &types.NodeOpsHistory{TenantID: "tenant-a", Snapshots: []*types.NodeOpsHistoryHostSnapshot{}},
		},
		{name: "metadata only",
			selection: &types.NodeOpsHistoryFieldSelection{
				OperateID: true, Operator: true, OperateTime: true, NodeRole: true, BizIDs: true,
			},
			want: &types.NodeOpsHistory{Info: full.Info, Snapshots: []*types.NodeOpsHistoryHostSnapshot{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := h.GetWithFields(ctx, full.Info.OperateID, tt.selection)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}

	_, err := h.GetWithFields(ctx, "missing", nil)
	require.ErrorIs(t, err, mongo.ErrNoDocuments)
	otherCtx := contextx.New(t.Context(), contextx.WithTenantID("tenant-b"))
	_, err = h.GetWithFields(otherCtx, full.Info.OperateID, nil)
	require.ErrorIs(t, err, mongo.ErrNoDocuments)
}

func TestHandler_ListWithFieldsAndCount(t *testing.T) {
	h, ctx := testHistoryHandler(t)
	for _, item := range []struct {
		id     string
		bizIDs []int64
	}{
		{id: "op-1", bizIDs: []int64{100, 200}},
		{id: "op-2", bizIDs: []int64{200}},
		{id: "op-3", bizIDs: []int64{300}},
	} {
		history := historyFixture()
		history.Info.OperateID = item.id
		history.Info.BizIDs = item.bizIDs
		// All operations reference the same hosts; filtering uses archived businesses only.
		require.NoError(t, h.Create(ctx, history))
	}
	tests := []struct {
		name string
		opts []OptFn
		want []string
	}{
		{name: "all", want: []string{"op-1", "op-2", "op-3"}},
		{name: "includes business", opts: []OptFn{WithBizID(200)}, want: []string{"op-1", "op-2"}},
		{name: "includes any business", opts: []OptFn{WithBizID(100, 300)}, want: []string{"op-1", "op-3"}},
		{name: "excludes business", opts: []OptFn{WithoutBizID(100)}, want: []string{"op-2", "op-3"}},
		{name: "includes and excludes", opts: []OptFn{WithBizID(200), WithoutBizID(100)}, want: []string{"op-2"}},
		{name: "no matching business", opts: []OptFn{WithBizID(400)}, want: []string{}},
		{name: "empty business filters", opts: []OptFn{WithBizID(), WithoutBizID()}, want: []string{"op-1", "op-2", "op-3"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows, total, err := h.ListWithFields(ctx, types.UnlimitedPage(), nil, tt.opts...)
			require.NoError(t, err)
			require.Equal(t, int64(len(tt.want)), total)
			ids := make([]string, 0, len(rows))
			for _, row := range rows {
				ids = append(ids, row.Info.OperateID)
				require.Equal(t, historyFixture().Snapshots, row.Snapshots)
			}
			require.Equal(t, tt.want, ids)
			count, err := h.Count(ctx, tt.opts...)
			require.NoError(t, err)
			require.Equal(t, total, count)
		})
	}

	rows, total, err := h.ListWithFields(ctx, types.Page{Offset: 1, Limit: 1},
		&types.NodeOpsHistoryFieldSelection{OperateID: true, BizIDs: true}, WithBizID(200))
	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	require.Equal(t, []*types.NodeOpsHistory{{
		Info:      types.NodeOpsHistoryInfo{OperateID: "op-2", BizIDs: []int64{200}},
		Snapshots: []*types.NodeOpsHistoryHostSnapshot{},
	}}, rows)

	otherCtx := contextx.New(t.Context(), contextx.WithTenantID("tenant-b"))
	count, err := h.Count(otherCtx, WithBizID(200))
	require.NoError(t, err)
	require.Zero(t, count)
}

func testHistoryHandler(t *testing.T) (*Handler, contextx.IContext) {
	t.Helper()
	_, db := support.RequireMongoDatabase(t)
	return New(db), contextx.New(t.Context(), contextx.WithTenantID("tenant-a"))
}

func historyFixture() *types.NodeOpsHistory {
	return &types.NodeOpsHistory{
		TenantID: "tenant-a",
		Info: types.NodeOpsHistoryInfo{OperateID: "op-1", Operator: "alice",
			OperateTime: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC), NodeRole: types.NodeRoleAgent,
			BizIDs: []int64{100, 200}},
		Snapshots: []*types.NodeOpsHistoryHostSnapshot{
			{HostID: 1, OpsConsoleHostID: 11, OpsConsoleHostInnerIP: []string{"10.0.0.1"},
				OpsConsoleHostInnerIPV6: []string{"::1"}, OpsOutBandIP: "10.0.0.2", OpsOutBandPort: 443,
				OpsOutBandProtocol: "https", OpsOutBandType: "redfish"},
			{HostID: 2, OpsOutBandIP: "10.0.0.3", OpsOutBandPort: 22},
		},
	}
}
