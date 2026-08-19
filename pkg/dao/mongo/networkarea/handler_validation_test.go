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

package networkarea

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
			name: "upsert rejects empty networkareas",
			run: func() error {
				return h.UpsertMany(nCtx)
			},
		},
		{
			name: "upsert rejects nil networkarea",
			run: func() error {
				return h.UpsertMany(nCtx, nil)
			},
		},
		{
			name: "upsert rejects another tenant",
			run: func() error {
				return h.UpsertMany(nCtx, &types.NetworkArea{TenantID: "tenant-b"})
			},
		},
		{
			name: "update rejects empty networkareas",
			run: func() error {
				return h.UpdateMany(nCtx)
			},
		},
		{
			name: "update rejects nil networkarea",
			run: func() error {
				return h.UpdateMany(nCtx, nil)
			},
		},
		{
			name: "update rejects another tenant",
			run: func() error {
				return h.UpdateMany(nCtx, &types.NetworkArea{TenantID: "tenant-b"})
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
			name: "upsert rejects nil context",
			run: func() error {
				return h.UpsertMany(nil, &types.NetworkArea{})
			},
		},
		{
			name: "update rejects nil context",
			run: func() error {
				return h.UpdateMany(nil, &types.NetworkArea{})
			},
		},
		{
			name: "delete rejects nil context",
			run: func() error {
				return h.DeleteMany(nil, 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Error(t, tt.run())
		})
	}
}
