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
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
)

func (c *cli) listRoles(ctx contextx.IContext, req *ListRolesReq) (*ListRolesResp, error) {
	resp := new(BaseBroker[*ListRolesResp])
	result := c.client.Get().
		SubResourcef("/rbac/model/systems/%s/roles/", req.SystemID).
		WithContext(ctx).
		WithHeaders(c.getHeader(ctx)).
		WithParam("page", strconv.Itoa(req.Page)).
		WithParam("page_size", strconv.Itoa(req.PageSize)).
		EnableLogBody().
		EnableLogResponse().
		Do()
	if err := result.Into(resp); err != nil {
		return nil, fmt.Errorf("list roles page %d: %w", req.Page, err)
	}
	if result.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list_roles: unexpected HTTP %d (request-id: %s)",
			result.StatusCode, result.Header.Get("X-Bkapi-Request-Id"))
	}
	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("list roles failed: %w", err)
	}

	if resp.Data == nil {
		return nil, fmt.Errorf("list roles page %d: incomplete response (request-id: %s)", req.Page, resp.RequestID)
	}

	return resp.Data, nil
}

func (c *cli) batchCreateRole(ctx contextx.IContext, req *BatchCreateRoleReq) (*BatchCreateRoleResp, error) {
	resp := new(BaseBroker[BatchCreateRoleResp])
	result := c.client.Post().
		SubResourcef("/rbac/model/systems/%s/roles/", req.SystemID).
		WithContext(ctx).
		WithHeaders(c.getHeader(ctx)).
		Body(req.Roles).
		EnableLogBody().
		EnableLogResponse().
		Do()
	if err := result.Into(resp); err != nil {
		return nil, fmt.Errorf("create role: %w", err)
	}
	if result.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("create_role: unexpected HTTP %d (request-id: %s)",
			result.StatusCode, result.Header.Get("X-Bkapi-Request-Id"))
	}
	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("create role failed: %w", err)
	}

	return &resp.Data, nil
}

func (c *cli) updateRole(ctx contextx.IContext, req *UpdateRoleReq) (*UpdateRoleResp, error) {
	resp := new(BaseBroker[UpdateRoleResp])
	result := c.client.Put().
		SubResourcef("/rbac/model/systems/%s/roles/%s/", req.SystemID, req.RoleID).
		WithContext(ctx).
		WithHeaders(c.getHeader(ctx)).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do()
	if err := result.Into(resp); err != nil {
		return nil, fmt.Errorf("update role: %w", err)
	}
	if result.StatusCode != http.StatusNoContent {
		return nil, fmt.Errorf("update_role: unexpected HTTP %d (request-id: %s)",
			result.StatusCode, result.Header.Get("X-Bkapi-Request-Id"))
	}
	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("update role failed: %w", err)
	}

	return &resp.Data, nil
}

func (c *cli) batchCreateRoleAction(
	ctx contextx.IContext, req *BatchCreateRoleActionReq,
) (*BatchCreateRoleActionResp, error) {

	resp := new(BaseBroker[BatchCreateRoleActionResp])
	result := c.client.Post().
		SubResourcef("/rbac/model/systems/%s/roles/%s/actions/", req.SystemID, req.RoleID).
		WithContext(ctx).
		WithHeaders(c.getHeader(ctx)).
		Body(req.Actions).
		EnableLogBody().
		EnableLogResponse().
		Do()
	if err := result.Into(resp); err != nil {
		return nil, fmt.Errorf("add role actions: %w; check remote state before rerunning (request-id: %s)",
			err, result.Header.Get("X-Bkapi-Request-Id"))
	}
	if result.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("add_role_actions: unexpected HTTP %d; check remote state before rerunning (request-id: %s)",
			result.StatusCode, result.Header.Get("X-Bkapi-Request-Id"))
	}
	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("add role actions failed: %w; check remote state before rerunning", err)
	}

	return &resp.Data, nil
}

func (c *cli) batchDeleteRoleAction(
	ctx contextx.IContext, req *BatchDeleteRoleActionReq,
) (*BatchDeleteRoleActionResp, error) {

	resp := new(BaseBroker[BatchDeleteRoleActionResp])
	result := c.client.Delete().
		SubResourcef("/rbac/model/systems/%s/roles/%s/actions/", req.SystemID, req.RoleID).
		WithContext(ctx).
		WithHeaders(c.getHeader(ctx)).
		WithParam("ids", strings.Join(req.IDs, ",")).
		Do()
	if err := result.Into(resp); err != nil {
		return nil, fmt.Errorf("delete role actions: %w; check remote state before rerunning (request-id: %s)",
			err, result.Header.Get("X-Bkapi-Request-Id"))
	}
	if result.StatusCode != http.StatusNoContent {
		return nil, fmt.Errorf("delete_role_actions: unexpected HTTP %d; check remote state before rerunning (request-id: %s)",
			result.StatusCode, result.Header.Get("X-Bkapi-Request-Id"))
	}
	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("delete role actions failed: %w; check remote state before rerunning", err)
	}

	return &resp.Data, nil
}

func (c *cli) deleteRole(ctx contextx.IContext, req *DeleteRoleReq) (*DeleteRoleResp, error) {
	resp := new(BaseBroker[DeleteRoleResp])
	result := c.client.Delete().
		SubResourcef("/rbac/model/systems/%s/roles/%s/", req.SystemID, req.RoleID).
		WithContext(ctx).
		WithHeaders(c.getHeader(ctx)).
		Do()
	if err := result.Into(resp); err != nil {
		return nil, fmt.Errorf("delete role: %w", err)
	}
	if result.StatusCode != http.StatusNoContent {
		return nil, fmt.Errorf("delete_role: unexpected HTTP %d (request-id: %s); roles with existing authorizations cannot be deleted",
			result.StatusCode, result.Header.Get("X-Bkapi-Request-Id"))
	}
	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("delete role failed: %w", err)
	}

	return &resp.Data, nil
}
