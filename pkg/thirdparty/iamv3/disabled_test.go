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

package iamv3

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func TestDisabledHandler(t *testing.T) {
	h := NewDisabledHandler()
	ctx := contextx.New(context.Background())
	resources := []types.IAMResource{{SystemID: "bk_cmdb", Type: "host", ID: "host-1"}}
	req := types.IAMCheckRequest{
		SystemID: "bk_nodemgr", Username: "admin", ActionID: "host_view", Resources: resources,
	}
	multiReq := types.IAMMultiActionCheckRequest{
		SystemID: "bk_nodemgr", Username: "admin", ActionIDs: []string{"host_view", "host_edit"},
	}
	scopeReq := types.IAMAuthorizedInstancesRequest{
		SystemID: "bk_nodemgr", Username: "admin", ActionID: "host_view", ResourceType: "host",
	}

	t.Run("IsAllowed", func(t *testing.T) {
		allowed, err := h.IsAllowed(ctx, req)
		if allowed || !errors.Is(err, ErrDisabled) {
			t.Fatalf("IsAllowed() = %v, %v; want false, ErrDisabled", allowed, err)
		}
	})
	t.Run("IsAllowedWithCache", func(t *testing.T) {
		allowed, err := h.IsAllowedWithCache(ctx, req, time.Minute)
		if allowed || !errors.Is(err, ErrDisabled) {
			t.Fatalf("IsAllowedWithCache() = %v, %v; want false, ErrDisabled", allowed, err)
		}
	})
	t.Run("BatchIsAllowed", func(t *testing.T) {
		results, err := h.BatchIsAllowed(ctx, req, [][]types.IAMResource{resources})
		if results != nil || !errors.Is(err, ErrDisabled) {
			t.Fatalf("BatchIsAllowed() = %v, %v; want nil, ErrDisabled", results, err)
		}
	})
	t.Run("ResourceMultiActionsAllowed", func(t *testing.T) {
		results, err := h.ResourceMultiActionsAllowed(ctx, multiReq)
		if results != nil || !errors.Is(err, ErrDisabled) {
			t.Fatalf("ResourceMultiActionsAllowed() = %v, %v; want nil, ErrDisabled", results, err)
		}
	})
	t.Run("BatchResourceMultiActionsAllowed", func(t *testing.T) {
		results, err := h.BatchResourceMultiActionsAllowed(ctx, multiReq, [][]types.IAMResource{resources})
		if results != nil || !errors.Is(err, ErrDisabled) {
			t.Fatalf("BatchResourceMultiActionsAllowed() = %v, %v; want nil, ErrDisabled", results, err)
		}
	})
	t.Run("GetToken", func(t *testing.T) {
		token, err := h.GetToken(ctx)
		if token != "" || !errors.Is(err, ErrDisabled) {
			t.Fatalf("GetToken() = %q, %v; want empty token, ErrDisabled", token, err)
		}
	})
	t.Run("IsBasicAuthAllowed", func(t *testing.T) {
		err := h.IsBasicAuthAllowed(ctx, "anyone", "anything")
		if !errors.Is(err, ErrDisabled) {
			t.Fatalf("IsBasicAuthAllowed() = %v; want ErrDisabled", err)
		}
	})
	t.Run("GetApplyURL", func(t *testing.T) {
		url, err := h.GetApplyURL(ctx, types.IAMApplyRequest{
			SystemID: "bk_nodemgr", Actions: []types.IAMApplyAction{{ID: "host_view"}},
		})
		if url != "" || !errors.Is(err, ErrDisabled) {
			t.Fatalf("GetApplyURL() = %q, %v; want empty URL, ErrDisabled", url, err)
		}
	})
	t.Run("GetPolicyExpression", func(t *testing.T) {
		expr, err := h.GetPolicyExpression(ctx, scopeReq)
		if expr != nil || !errors.Is(err, ErrDisabled) {
			t.Fatalf("GetPolicyExpression() = %v, %v; want nil, ErrDisabled", expr, err)
		}
	})
	t.Run("ListAuthorizedInstances", func(t *testing.T) {
		isAny, results, err := h.ListAuthorizedInstances(ctx, scopeReq)
		if isAny || results != nil || !errors.Is(err, ErrDisabled) {
			t.Fatalf("ListAuthorizedInstances() = %v, %v, %v; want false, nil, ErrDisabled", isAny, results, err)
		}
	})
}
