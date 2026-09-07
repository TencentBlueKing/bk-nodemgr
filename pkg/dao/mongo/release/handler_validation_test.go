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

package release

import (
	"context"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func TestHandler_RejectsNilContextBeforeDatabaseAccess(t *testing.T) {
	h := New(nil)

	tests := []struct {
		name string
		run  func() error
	}{
		{
			name: "get rejects nil context",
			run: func() error {
				_, err := h.Get(nil, types.ReleaseTypeAgent)
				return err
			},
		},
		{
			name: "exist rejects nil context",
			run: func() error {
				_, err := h.Exist(nil, types.ReleaseTypeAgent)
				return err
			},
		},
		{
			name: "list rejects nil context",
			run: func() error {
				_, _, err := h.List(nil, types.ReleaseTypeAgent, types.Page{})
				return err
			},
		},
		{
			name: "set labels rejects nil context",
			run: func() error {
				return h.SetLabels(nil, types.ReleaseTypeAgent, nil)
			},
		},
		{
			name: "count rejects nil context",
			run: func() error {
				_, err := h.Count(nil, types.ReleaseTypeAgent)
				return err
			},
		},
		{
			name: "upsert rejects nil context",
			run: func() error {
				return h.UpsertMany(nil, types.ReleaseTypeAgent, &types.Release{})
			},
		},
		{
			name: "delete rejects nil context",
			run: func() error {
				return h.Delete(nil, types.ReleaseTypeAgent)
			},
		},
		{
			name: "set enabled rejects nil context",
			run: func() error {
				return h.SetEnabled(nil, types.ReleaseTypeAgent, true)
			},
		},
		{
			name: "set hidden rejects nil context",
			run: func() error {
				return h.SetHidden(nil, types.ReleaseTypeAgent, true)
			},
		},
		{
			name: "set default rejects nil context",
			run: func() error {
				return h.SetAsDefault(nil, types.ReleaseTypeAgent, true)
			},
		},
		{
			name: "cancel default rejects nil context",
			run: func() error {
				return h.CancelPlatformDefault(nil, types.ReleaseTypeAgent)
			},
		},
		{
			name: "distinct os rejects nil context",
			run: func() error {
				_, err := h.DistinctOsType(nil, types.ReleaseTypeAgent)
				return err
			},
		},
		{
			name: "distinct arch rejects nil context",
			run: func() error {
				_, err := h.DistinctCPUArch(nil, types.ReleaseTypeAgent)
				return err
			},
		},
		{
			name: "distinct name rejects nil context",
			run: func() error {
				_, err := h.DistinctName(nil, types.ReleaseTypeAgent)
				return err
			},
		},
		{
			name: "distinct version rejects nil context",
			run: func() error {
				_, err := h.DistinctVersion(nil, types.ReleaseTypeAgent)
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.run(); err != nil {
				return
			}

			t.Fatal("expected error")
		})
	}
}

func TestHandler_RejectsMissingTenantBeforeDatabaseAccess(t *testing.T) {
	h := New(nil)
	nCtx := contextx.New(context.Background())

	tests := []struct {
		name string
		run  func() error
	}{
		{
			name: "get rejects missing tenant",
			run: func() error {
				_, err := h.Get(nCtx, types.ReleaseTypeAgent)
				return err
			},
		},
		{
			name: "exist rejects missing tenant",
			run: func() error {
				_, err := h.Exist(nCtx, types.ReleaseTypeAgent)
				return err
			},
		},
		{
			name: "list rejects missing tenant",
			run: func() error {
				_, _, err := h.List(nCtx, types.ReleaseTypeAgent, types.Page{})
				return err
			},
		},
		{
			name: "set labels rejects missing tenant",
			run: func() error {
				return h.SetLabels(nCtx, types.ReleaseTypeAgent, nil)
			},
		},
		{
			name: "count rejects missing tenant",
			run: func() error {
				_, err := h.Count(nCtx, types.ReleaseTypeAgent)
				return err
			},
		},
		{
			name: "upsert rejects missing tenant",
			run: func() error {
				return h.UpsertMany(nCtx, types.ReleaseTypeAgent, &types.Release{})
			},
		},
		{
			name: "delete rejects missing tenant",
			run: func() error {
				return h.Delete(nCtx, types.ReleaseTypeAgent)
			},
		},
		{
			name: "set enabled rejects missing tenant",
			run: func() error {
				return h.SetEnabled(nCtx, types.ReleaseTypeAgent, true)
			},
		},
		{
			name: "set hidden rejects missing tenant",
			run: func() error {
				return h.SetHidden(nCtx, types.ReleaseTypeAgent, true)
			},
		},
		{
			name: "set default rejects missing tenant",
			run: func() error {
				return h.SetAsDefault(nCtx, types.ReleaseTypeAgent, true)
			},
		},
		{
			name: "cancel default rejects missing tenant",
			run: func() error {
				return h.CancelPlatformDefault(nCtx, types.ReleaseTypeAgent)
			},
		},
		{
			name: "distinct os rejects missing tenant",
			run: func() error {
				_, err := h.DistinctOsType(nCtx, types.ReleaseTypeAgent)
				return err
			},
		},
		{
			name: "distinct arch rejects missing tenant",
			run: func() error {
				_, err := h.DistinctCPUArch(nCtx, types.ReleaseTypeAgent)
				return err
			},
		},
		{
			name: "distinct name rejects missing tenant",
			run: func() error {
				_, err := h.DistinctName(nCtx, types.ReleaseTypeAgent)
				return err
			},
		},
		{
			name: "distinct version rejects missing tenant",
			run: func() error {
				_, err := h.DistinctVersion(nCtx, types.ReleaseTypeAgent)
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.run(); err != nil {
				return
			}

			t.Fatal("expected error")
		})
	}
}
