/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package networkunit

import (
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestHandler_RejectsInvalidParametersBeforeDatabaseAccess(t *testing.T) {
	h := New(nil)
	nCtx := contextx.New(t.Context(), contextx.WithTenantID("tenant-a"))

	tests := []struct {
		name string
		run  func() error
	}{
		{
			name: "get rejects negative id",
			run: func() error {
				_, err := h.Get(nCtx, -1)
				return err
			},
		},
		{
			name: "create rejects nil networkunit",
			run: func() error {
				_, err := h.Create(nCtx, nil)
				return err
			},
		},
		{
			name: "create rejects another tenant",
			run: func() error {
				_, err := h.Create(nCtx, &types.NetworkUnit{TenantID: "tenant-b"})
				return err
			},
		},
		{
			name: "update rejects empty networkunits",
			run: func() error {
				return h.UpdateMany(nCtx, types.NetworkUnitUpdateFields{Name: true})
			},
		},
		{
			name: "update rejects empty field mask",
			run: func() error {
				return h.UpdateMany(nCtx, types.NetworkUnitUpdateFields{}, &types.NetworkUnit{TenantID: "tenant-a"})
			},
		},
		{
			name: "update rejects another tenant",
			run: func() error {
				return h.UpdateMany(nCtx, types.NetworkUnitUpdateFields{Name: true}, &types.NetworkUnit{
					TenantID: "tenant-b",
					Name:     "updated",
				})
			},
		},
		{
			name: "delete rejects empty ids",
			run: func() error {
				return h.DeleteMany(nCtx)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Error(t, tt.run())
		})
	}
}

func TestHandler_RejectsNilContextBeforeDatabaseAccess(t *testing.T) {
	h := New(nil)

	tests := []struct {
		name string
		run  func() error
	}{
		{
			name: "count rejects nil context",
			run: func() error {
				_, err := h.Count(nil)
				return err
			},
		},
		{
			name: "list rejects nil context",
			run: func() error {
				_, _, err := h.List(nil, types.Page{})
				return err
			},
		},
		{
			name: "get rejects nil context",
			run: func() error {
				_, err := h.Get(nil, 1)
				return err
			},
		},
		{
			name: "create rejects nil context",
			run: func() error {
				_, err := h.Create(nil, &types.NetworkUnit{})
				return err
			},
		},
		{
			name: "update rejects nil context",
			run: func() error {
				return h.UpdateMany(nil, types.NetworkUnitUpdateFields{Name: true}, &types.NetworkUnit{})
			},
		},
		{
			name: "delete rejects nil context",
			run: func() error {
				return h.DeleteMany(nil, 1)
			},
		},
		{
			name: "distribution rejects nil context",
			run: func() error {
				_, err := h.GetNetworkUnitDistributionByNetworkAreaID(nil)
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Error(t, tt.run())
		})
	}
}
