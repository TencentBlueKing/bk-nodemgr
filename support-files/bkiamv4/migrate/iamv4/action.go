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
	"fmt"
	"net/http"
	"strconv"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
)

func (c *cli) listActions(ctx contextx.IContext, req *ListActionsReq) (*ListActionsResp, error) {
	resp := new(BaseBroker[*ListActionsResp])
	result := c.client.Get().
		SubResourcef("/rbac/model/systems/%s/actions/", req.SystemID).
		WithContext(ctx).
		WithHeaders(c.getHeader(ctx)).
		WithParam("page", strconv.Itoa(req.Page)).
		WithParam("page_size", strconv.Itoa(req.PageSize)).
		EnableLogBody().
		EnableLogResponse().
		Do()
	if err := result.Into(resp); err != nil {
		return nil, fmt.Errorf("list actions page %d: %w", req.Page, err)
	}
	if result.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list_actions: unexpected HTTP %d (request-id: %s)",
			result.StatusCode, result.Header.Get("X-Bkapi-Request-Id"))
	}
	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("list actions failed: %w", err)
	}

	if resp.Data == nil {
		return nil, fmt.Errorf("list actions page %d: incomplete response (request-id: %s)", req.Page, resp.RequestID)
	}

	return resp.Data, nil
}

func (c *cli) batchCreateAction(ctx contextx.IContext, req *BatchCreateActionReq) (*BatchCreateActionResp, error) {
	resp := new(BaseBroker[BatchCreateActionResp])
	result := c.client.Post().
		SubResourcef("/rbac/model/systems/%s/actions/", req.SystemID).
		WithContext(ctx).
		WithHeaders(c.getHeader(ctx)).
		Body(req.Actions).
		EnableLogBody().
		EnableLogResponse().
		Do()
	if err := result.Into(resp); err != nil {
		return nil, fmt.Errorf("create action: %w", err)
	}
	if result.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("create_action: unexpected HTTP %d (request-id: %s)",
			result.StatusCode, result.Header.Get("X-Bkapi-Request-Id"))
	}
	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("create action failed: %w", err)
	}

	return &resp.Data, nil
}

func (c *cli) updateAction(ctx contextx.IContext, req *UpdateActionReq) (*UpdateActionResp, error) {
	resp := new(BaseBroker[UpdateActionResp])
	result := c.client.Put().
		SubResourcef("/rbac/model/systems/%s/actions/%s/", req.SystemID, req.ActionID).
		WithContext(ctx).
		WithHeaders(c.getHeader(ctx)).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do()
	if err := result.Into(resp); err != nil {
		return nil, fmt.Errorf("update action: %w", err)
	}
	if result.StatusCode != http.StatusNoContent {
		return nil, fmt.Errorf("update_action: unexpected HTTP %d (request-id: %s)",
			result.StatusCode, result.Header.Get("X-Bkapi-Request-Id"))
	}
	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("update action failed: %w", err)
	}

	return &resp.Data, nil
}
