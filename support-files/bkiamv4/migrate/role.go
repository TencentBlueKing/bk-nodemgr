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

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"slices"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/support-files/bkiamv4/migrate/iamv4"
)

func validateRole(data map[string]json.RawMessage) error {
	var role iamv4.Role
	idPattern := regexp.MustCompile(`^[a-z][a-z0-9_-]{0,31}$`)
	if err := json.Unmarshal(data[fieldID], &role.ID); err != nil {
		return fmt.Errorf("data.id is required: %w", err)
	}
	if !idPattern.MatchString(role.ID) {
		return fmt.Errorf("invalid role ID %q", role.ID)
	}
	for field, raw := range data {
		if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return fmt.Errorf("data.%s must not be null; omit optional fields to preserve the remote value", field)
		}
		switch field {
		case fieldID, "actions":
			// Identity is validated above; the required member set is validated below.
		case fieldName:
			if err := json.Unmarshal(raw, &role.Name); err != nil {
				return fmt.Errorf("data.name must be a string: %w", err)
			}
			if strings.TrimSpace(role.Name) == "" {
				return fmt.Errorf("data.name must not be empty")
			}
		case "description":
			if err := json.Unmarshal(raw, &role.Description); err != nil {
				return fmt.Errorf("data.description must be a string: %w", err)
			}
		default:
			return fmt.Errorf("unknown role field %q", field)
		}
	}

	return validateRoleActions(data["actions"])
}

func validateRoleActions(raw json.RawMessage) error {
	var members []map[string]*string
	if err := json.Unmarshal(raw, &members); err != nil {
		return fmt.Errorf("data.actions is required and must be an object array: %w", err)
	}
	if len(members) == 0 {
		return fmt.Errorf("data.actions must not be empty")
	}
	idPattern := regexp.MustCompile(`^[a-z][a-z0-9_-]{0,31}$`)
	seen := make(map[string]struct{}, len(members))
	for index, member := range members {
		if err := validateRoleActionFields(member); err != nil {
			return fmt.Errorf("data.actions[%d]: %w", index, err)
		}
		id, dimension := member[fieldID], member["resource_type_id"]
		if !idPattern.MatchString(*id) || (*dimension != "" && !idPattern.MatchString(*dimension)) {
			return fmt.Errorf("data.actions[%d]: invalid action ID %q or resource type ID %q", index, *id, *dimension)
		}
		if _, exists := seen[*id]; exists {
			return fmt.Errorf("data.actions[%d]: repeated action ID %q", index, *id)
		}
		seen[*id] = struct{}{}
	}

	return nil
}

func validateRoleActionFields(member map[string]*string) error {
	for field, value := range member {
		if field != fieldID && field != "resource_type_id" {
			return fmt.Errorf("unknown field %q", field)
		}
		if value == nil {
			return fmt.Errorf("%s must not be null", field)
		}
	}
	if member[fieldID] == nil || member["resource_type_id"] == nil {
		return fmt.Errorf("id and resource_type_id are required strings")
	}

	return nil
}

func (item migration) executeRole(
	ctx contextx.IContext, handler iamv4.IHandler, op operation,
	states map[string]map[string]iamv4.Role, actions map[string]map[string]iamv4.Action,
	resourceTypes map[string]map[string]iamv4.ResourceType, dryRun bool, out io.Writer,
) error {

	roles, loaded := states[item.SystemID]
	if !loaded {
		listed, err := handler.ListRoles(ctx, item.SystemID)
		if err != nil {
			return fmt.Errorf("list roles for %s: %w", item.SystemID, err)
		}
		roles = make(map[string]iamv4.Role, len(listed))
		for _, role := range listed {
			roles[role.ID] = role
		}
		states[item.SystemID] = roles
	}
	role, fields, err := prepareRole(op, roles)
	if err != nil {
		return err
	}
	if err := validateRoleRelations(ctx, handler, item.SystemID, role, actions, resourceTypes); err != nil {
		return err
	}
	current, exists := roles[role.ID]
	var additions []iamv4.RoleAction
	if exists {
		additions = missingRoleActions(role.Actions, current.Actions)
	}
	plan := "skip_role"
	metadataChanged := fields.Name != nil || fields.Description != nil
	switch {
	case !exists:
		plan = "create_role"
	case len(additions) > 0 && metadataChanged:
		plan = "add_role_actions+update_role"
	case len(additions) > 0:
		plan = "add_role_actions"
	case metadataChanged:
		plan = "update_role"
	}
	if _, err := fmt.Fprintf(out, "%s: %s %s/%s (dry-run=%t)\n", item.filename, plan, item.SystemID, role.ID, dryRun); err != nil {
		return fmt.Errorf("write role plan: %w", err)
	}
	if !dryRun && plan != "skip_role" {
		if err := applyRole(ctx, handler, item.SystemID, role, fields, exists, additions); err != nil {
			return err
		}
	}
	// Successful writes and dry-run plans both affect later operations in this run.
	roles[role.ID] = role

	return nil
}

func prepareRole(op operation, roles map[string]iamv4.Role) (iamv4.Role, iamv4.RoleFields, error) {
	role := op.role
	fields := iamv4.RoleFields{}
	current, exists := roles[role.ID]
	if exists && !containsRoleActions(role.Actions, current.Actions) {
		return role, fields, fmt.Errorf("role %s removes members or changes their dimensions; automatic removal or rebinding is not supported", role.ID)
	}
	if exists {
		if _, supplied := op.Data[fieldName]; !supplied {
			role.Name = current.Name
		}
		if _, supplied := op.Data["description"]; !supplied {
			role.Description = current.Description
		}
	}
	if exists && role.Name != current.Name {
		fields.Name = &role.Name
	}
	if exists && role.Description != current.Description {
		fields.Description = &role.Description
	}
	if !exists && role.Name == "" {
		return role, fields, fmt.Errorf("data.name is required to create role %s", role.ID)
	}
	for _, other := range roles {
		if other.ID != role.ID && other.Name == role.Name {
			return role, fields, fmt.Errorf("role name %q is already used by %s", role.Name, other.ID)
		}
	}

	return role, fields, nil
}

func missingRoleActions(desired, current []iamv4.RoleAction) []iamv4.RoleAction {
	additions := make([]iamv4.RoleAction, 0)
	for _, member := range desired {
		if !slices.Contains(current, member) {
			additions = append(additions, member)
		}
	}

	return additions
}

func containsRoleActions(desired, current []iamv4.RoleAction) bool {
	members := make(map[iamv4.RoleAction]struct{}, len(desired))
	for _, member := range desired {
		members[member] = struct{}{}
	}
	for _, member := range current {
		if _, exists := members[member]; !exists {
			return false
		}
		delete(members, member)
	}

	return true
}

func loadActions(
	ctx contextx.IContext, handler iamv4.IHandlerAction, systemID string, states map[string]map[string]iamv4.Action,
) (map[string]iamv4.Action, error) {

	if actions, loaded := states[systemID]; loaded {
		return actions, nil
	}
	listed, err := handler.ListActions(ctx, systemID)
	if err != nil {
		return nil, fmt.Errorf("list actions for %s: %w", systemID, err)
	}
	actions := make(map[string]iamv4.Action, len(listed))
	for _, action := range listed {
		actions[action.ID] = action
	}
	states[systemID] = actions

	return actions, nil
}

func validateRoleRelations(
	ctx contextx.IContext, handler iamv4.IHandler, systemID string, role iamv4.Role,
	actionStates map[string]map[string]iamv4.Action, resourceTypes map[string]map[string]iamv4.ResourceType,
) error {

	actions, err := loadActions(ctx, handler, systemID, actionStates)
	if err != nil {
		return err
	}
	for _, member := range role.Actions {
		action, exists := actions[member.ID]
		if !exists {
			return fmt.Errorf("action %s must exist before role %s", member.ID, role.ID)
		}
		if action.ResourceTypeID == "" {
			if member.ResourceTypeID != "" {
				return fmt.Errorf("role %s action %s is resource-free and requires an empty resource_type_id", role.ID, member.ID)
			}

			continue
		}
		if member.ResourceTypeID == "" {
			return fmt.Errorf("role %s action %s requires a nonempty resource_type_id", role.ID, member.ID)
		}
		if err := validateRoleDimension(ctx, handler, systemID, member, action, resourceTypes); err != nil {
			return fmt.Errorf("role %s: %w", role.ID, err)
		}
	}

	return nil
}

func validateRoleDimension(
	ctx contextx.IContext, handler iamv4.IHandlerResourceType, systemID string, member iamv4.RoleAction, action iamv4.Action,
	states map[string]map[string]iamv4.ResourceType,
) error {

	resources, err := loadResourceTypes(ctx, handler, systemID, states)
	if err != nil {
		return err
	}
	resource, exists := resources[action.ResourceTypeID]
	if !exists {
		return fmt.Errorf("resource type %s must exist before the role", action.ResourceTypeID)
	}
	if _, exists := resources[member.ResourceTypeID]; !exists {
		return fmt.Errorf("resource type %s must exist before the role", member.ResourceTypeID)
	}
	if member.ResourceTypeID != action.ResourceTypeID && !slices.Contains(resource.Ancestors, member.ResourceTypeID) {
		return fmt.Errorf("action %s dimension %s must match its binding or an ancestor", member.ID, member.ResourceTypeID)
	}

	return nil
}

func applyRole(
	ctx contextx.IContext, handler iamv4.IHandlerRole, systemID string,
	role iamv4.Role, fields iamv4.RoleFields, exists bool, additions []iamv4.RoleAction,
) error {

	if !exists {
		if err := handler.CreateRole(ctx, systemID, role); err != nil {
			return fmt.Errorf("create role %s: %w", role.ID, err)
		}

		return nil
	}
	if len(additions) > 0 {
		if err := handler.AddRoleActions(ctx, systemID, role.ID, additions); err != nil {
			return fmt.Errorf("add actions to role %s: %w", role.ID, err)
		}
	}
	if fields.Name != nil || fields.Description != nil {
		if err := handler.UpdateRole(ctx, systemID, role.ID, fields); err != nil {
			return fmt.Errorf("update role %s: %w", role.ID, err)
		}
	}

	return nil
}
