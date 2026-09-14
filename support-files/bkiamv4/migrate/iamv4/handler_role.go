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
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/pageexecutor"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IHandlerRole exposes role and member operations without HTTP details.
type IHandlerRole interface {
	ListRoles(ctx contextx.IContext, systemID string) ([]Role, error)
	CreateRole(ctx contextx.IContext, systemID string, role Role) error
	UpdateRole(ctx contextx.IContext, systemID, roleID string, fields RoleFields) error
	DeleteRole(ctx contextx.IContext, systemID, roleID string) error
	AddRoleActions(ctx contextx.IContext, systemID, roleID string, actions []RoleAction) error
	DeleteRoleActions(ctx contextx.IContext, systemID, roleID string, actionIDs []string) error
}

// ListRoles retrieves the complete role model or returns an error.
func (h *Handler) ListRoles(ctx contextx.IContext, systemID string) ([]Role, error) {
	const pageSize = 100             // IAM's maximum model query page size.
	const timeout = 30 * time.Second // Bound the complete model query, not each page.
	seen := make(map[string]struct{})
	total := 0
	executor := pageexecutor.NewPageExecutor[Role](pageSize, timeout)
	fn := func(ctx contextx.IContext, p types.Page) ([]Role, error) {
		if p.Offset > 0 && len(seen) == total {
			// The last full page already completed the model.
			return nil, nil
		}
		req := &ListRolesReq{SystemID: systemID, Page: p.Offset/pageSize + 1, PageSize: pageSize}
		resp, err := h.cli.listRoles(ctx, req)
		if err != nil {
			return nil, err
		}
		if err := validateRolePage(req, resp, total, seen); err != nil {
			return nil, err
		}
		total = *resp.Count

		items := make([]Role, 0, len(resp.Results))
		for _, role := range resp.Results {
			members := make([]RoleAction, 0, len(role.Actions))
			for _, action := range role.Actions {
				members = append(members, RoleAction{ID: action.ID, ResourceTypeID: *action.ResourceTypeID})
			}
			items = append(items, Role{ID: role.ID, Name: role.Name, Description: role.Description, Actions: members})
		}

		return items, nil
	}
	roles, err := executor.Execute(ctx, types.UnlimitedPage(), fn)
	if err != nil {
		return nil, fmt.Errorf("list roles: %w", err)
	}
	if len(roles.Items) != total {
		return nil, fmt.Errorf("list roles: incomplete pagination, got %d of %d", len(roles.Items), total)
	}

	return roles.Items, nil
}

func validateRolePage(
	req *ListRolesReq, resp *ListRolesResp, total int, seen map[string]struct{},
) error {

	if resp.Count == nil || *resp.Count < 0 || resp.Results == nil {
		return fmt.Errorf("list roles page %d: incomplete response", req.Page)
	}
	if req.Page == 1 {
		total = *resp.Count
	}
	if *resp.Count != total || len(resp.Results) > req.PageSize || len(resp.Results) > total-len(seen) {
		return fmt.Errorf("list roles page %d: inconsistent count", req.Page)
	}
	for _, role := range resp.Results {
		if role.ID == "" || role.Name == "" || role.Actions == nil {
			return fmt.Errorf("list roles page %d: malformed role", req.Page)
		}
		if _, exists := seen[role.ID]; exists {
			return fmt.Errorf("list roles page %d: duplicate role ID", req.Page)
		}
		seen[role.ID] = struct{}{}
		if err := validateRoleActions(role.Actions); err != nil {
			return fmt.Errorf("list roles page %d: %w", req.Page, err)
		}
	}

	return nil
}

func validateRoleActions(actions []RoleActionResp) error {
	seen := make(map[string]struct{}, len(actions))
	for _, action := range actions {
		if action.ID == "" || action.ResourceTypeID == nil {
			return fmt.Errorf("malformed role action")
		}
		if _, exists := seen[action.ID]; exists {
			return fmt.Errorf("duplicate role action ID")
		}
		seen[action.ID] = struct{}{}
	}

	return nil
}

// CreateRole registers one role through the batch API and verifies the created ID.
func (h *Handler) CreateRole(ctx contextx.IContext, systemID string, role Role) error {
	// Roles without member actions must send an empty array rather than JSON null.
	if role.Actions == nil {
		role.Actions = make([]RoleAction, 0)
	}
	resp, err := h.cli.batchCreateRole(ctx, &BatchCreateRoleReq{SystemID: systemID, Roles: []Role{role}})
	if err != nil {
		return err
	}
	if len(*resp) != 1 || (*resp)[0] != role.ID {
		return fmt.Errorf("create role: response ID does not match; check remote state before rerunning")
	}

	return nil
}

// UpdateRole sends supplied fields; the caller owns member drift checks.
func (h *Handler) UpdateRole(ctx contextx.IContext, systemID, roleID string, fields RoleFields) error {
	_, err := h.cli.updateRole(ctx, &UpdateRoleReq{SystemID: systemID, RoleID: roleID, RoleFields: fields})

	return err
}

// DeleteRole deletes a role; IAM rejects roles with existing authorizations.
func (h *Handler) DeleteRole(ctx contextx.IContext, systemID, roleID string) error {
	_, err := h.cli.deleteRole(ctx, &DeleteRoleReq{SystemID: systemID, RoleID: roleID})

	return err
}

// DeleteRoleActions removes members by action ID without deleting the role itself.
func (h *Handler) DeleteRoleActions(ctx contextx.IContext, systemID, roleID string, actionIDs []string) error {
	_, err := h.cli.batchDeleteRoleAction(ctx, &BatchDeleteRoleActionReq{
		SystemID: systemID,
		RoleID:   roleID,
		IDs:      actionIDs,
	})

	return err
}

// AddRoleActions adds member actions and verifies the returned IDs independently of order.
func (h *Handler) AddRoleActions(ctx contextx.IContext, systemID, roleID string, actions []RoleAction) error {
	resp, err := h.cli.batchCreateRoleAction(ctx, &BatchCreateRoleActionReq{
		SystemID: systemID,
		RoleID:   roleID,
		Actions:  actions,
	})
	if err != nil {
		return err
	}
	if len(*resp) != len(actions) {
		return fmt.Errorf("add role actions: response ID count does not match; check remote state before rerunning")
	}
	seen := make(map[string]bool, len(actions))
	for _, action := range actions {
		seen[action.ID] = false
	}
	for _, id := range *resp {
		found, expected := seen[id]
		if !expected {
			return fmt.Errorf("add role actions: unexpected response ID %s; check remote state before rerunning", id)
		}
		if found {
			return fmt.Errorf("add role actions: duplicate response ID %s; check remote state before rerunning", id)
		}
		seen[id] = true
	}
	for _, action := range actions {
		if !seen[action.ID] {
			return fmt.Errorf("add role actions: missing response ID %s; check remote state before rerunning", action.ID)
		}
	}

	return nil
}
