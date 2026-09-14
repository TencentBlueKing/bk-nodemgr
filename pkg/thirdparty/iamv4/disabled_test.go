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

package iamv4

import (
	"context"
	"errors"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func TestDisabledHandler(t *testing.T) {
	h := NewDisabledHandler()
	ctx := contextx.New(context.Background())
	req := types.IAMCheckRequest{
		SystemID: "bk_nodemgr", Username: "admin", ActionID: "host_view",
		Resources: []types.IAMResource{{SystemID: "bk_cmdb", Type: "host", ID: "host-1"}},
	}

	t.Run("IsAllowed", func(t *testing.T) {
		allowed, err := h.IsAllowed(ctx, req)
		if allowed || !errors.Is(err, ErrDisabled) {
			t.Fatalf("IsAllowed() = %v, %v; want false, ErrDisabled", allowed, err)
		}
	})
	t.Run("ResourcesAllowed", func(t *testing.T) {
		results, err := h.ResourcesAllowed(ctx, req)
		if results != nil || !errors.Is(err, ErrDisabled) {
			t.Fatalf("ResourcesAllowed() = %v, %v; want nil, ErrDisabled", results, err)
		}
	})
	t.Run("ActionsAllowed", func(t *testing.T) {
		results, err := h.ActionsAllowed(ctx, types.IAMMultiActionCheckRequest{
			SystemID: "bk_nodemgr", Username: "admin", ActionIDs: []string{"host_view", "host_edit"},
		})
		if results != nil || !errors.Is(err, ErrDisabled) {
			t.Fatalf("ActionsAllowed() = %v, %v; want nil, ErrDisabled", results, err)
		}
	})
	t.Run("ListAuthorizedResources", func(t *testing.T) {
		results, err := h.ListAuthorizedResources(ctx, types.IAMAuthorizedInstancesRequest{
			SystemID: "bk_nodemgr", Username: "admin", ActionID: "host_view", ResourceType: "host",
		})
		if results != nil || !errors.Is(err, ErrDisabled) {
			t.Fatalf("ListAuthorizedResources() = %v, %v; want nil, ErrDisabled", results, err)
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
		if !errors.Is(err, ErrDisabled) || errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("IsBasicAuthAllowed() = %v; want ErrDisabled, not ErrInvalidCredentials", err)
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
}
