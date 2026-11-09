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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/support-files/bkiamv4/migrate/iamv4"
)

type actionReplacement struct {
	iamv4.Action
	ExpectedResourceTypeID string   `json:"expected_resource_type_id"`
	Cascade                []string `json:"cascade,omitempty"`
}

func validateActionReplacement(data map[string]json.RawMessage) error {
	actionData := make(map[string]json.RawMessage, len(data))
	for field, raw := range data {
		if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return fmt.Errorf("data.%s must not be null", field)
		}
		if field != "expected_resource_type_id" && field != "cascade" {
			actionData[field] = raw
		}
	}
	if err := validateAction(actionData); err != nil {
		return err
	}
	for _, field := range []string{fieldName, "resource_type_id", "expected_resource_type_id"} {
		if _, supplied := data[field]; !supplied {
			return fmt.Errorf("data.%s is required for replace_action", field)
		}
	}
	encoded, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("encode replacement fields: %w", err)
	}
	var replacement actionReplacement
	if err := json.Unmarshal(encoded, &replacement); err != nil {
		return fmt.Errorf("decode replacement fields: %w", err)
	}
	if replacement.ExpectedResourceTypeID != "" &&
		!regexp.MustCompile(`^[a-z][a-z0-9_-]{0,31}$`).MatchString(replacement.ExpectedResourceTypeID) {
		return fmt.Errorf("invalid expected resource type ID %q", replacement.ExpectedResourceTypeID)
	}
	if replacement.ExpectedResourceTypeID == replacement.ResourceTypeID {
		return fmt.Errorf("replace_action requires different expected and target bindings; use upsert_action otherwise")
	}
	seen := make(map[string]bool, len(replacement.Cascade))
	for _, scope := range replacement.Cascade {
		switch scope {
		case "role", "rbac_authorization", "abac_authorization":
			if seen[scope] {
				return fmt.Errorf("duplicate cascade scope %q", scope)
			}
			seen[scope] = true
		default:
			return fmt.Errorf("unsupported cascade scope %q", scope)
		}
	}

	return nil
}

func (item migration) executeActionReplacement(
	ctx contextx.IContext, handler iamv4.IHandler, op operation,
	states map[string]map[string]iamv4.Action, resourceTypes map[string]map[string]iamv4.ResourceType,
	roleStates map[string]map[string]iamv4.Role, dryRun bool, out io.Writer,
) error {

	actions, err := loadActions(ctx, handler, item.SystemID, states)
	if err != nil {
		return err
	}
	replacement := op.replacement
	current, exists := actions[replacement.ID]
	if exists && current.ResourceTypeID != replacement.ResourceTypeID &&
		current.ResourceTypeID != replacement.ExpectedResourceTypeID {
		return fmt.Errorf("action %s binding is %q, expected %q or target %q; refusing replacement",
			replacement.ID, current.ResourceTypeID, replacement.ExpectedResourceTypeID, replacement.ResourceTypeID)
	}
	// Reuse normal upsert for completed replacements and recovery after a confirmed deletion.
	actionOp := operation{Operation: operationUpsertAction, action: replacement.Action, Data: map[string]json.RawMessage{
		fieldName: op.Data[fieldName], "resource_type_id": op.Data["resource_type_id"],
	}}
	if !exists || current.ResourceTypeID == replacement.ResourceTypeID {
		return item.executeAction(ctx, handler, actionOp, states, resourceTypes, dryRun, out)
	}
	if err := validateActionCreation(ctx, handler, item.SystemID, replacement.Action, resourceTypes); err != nil {
		return err
	}
	for _, other := range actions {
		if other.ID != replacement.ID && other.Name == replacement.Name {
			return fmt.Errorf("action name %q is already used by %s", replacement.Name, other.ID)
		}
	}
	var roles map[string]iamv4.Role
	if slices.Contains(replacement.Cascade, "role") {
		roles, err = loadRoles(ctx, handler, item.SystemID, roleStates)
		if err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(out, "%s: replace_action %s/%s binding=%q -> %q cascade=%v (dry-run=%t)\n",
		item.filename, item.SystemID, replacement.ID, current.ResourceTypeID,
		replacement.ResourceTypeID, replacement.Cascade, dryRun); err != nil {
		return fmt.Errorf("write action replacement plan: %w", err)
	}
	if !dryRun {
		if err := handler.DeleteAction(ctx, item.SystemID, replacement.ID, replacement.Cascade); err != nil {
			return fmt.Errorf("replace action %s: deletion failed or outcome is unknown; no create attempted, check remote state: %w",
				replacement.ID, err)
		}
	}
	delete(actions, replacement.ID)
	removeActionFromRoles(roles, replacement.ID)
	if !dryRun {
		if err := handler.CreateAction(ctx, item.SystemID, replacement.Action); err != nil {
			return fmt.Errorf("replace action %s: deletion succeeded but creation failed or outcome is unknown; no rollback, check remote state before rerunning: %w",
				replacement.ID, err)
		}
	}
	actions[replacement.ID] = replacement.Action

	return nil
}

func removeActionFromRoles(roles map[string]iamv4.Role, actionID string) {
	for id, role := range roles {
		members := make([]iamv4.RoleAction, 0, len(role.Actions))
		for _, member := range role.Actions {
			if member.ID != actionID {
				members = append(members, member)
			}
		}
		role.Actions = members
		roles[id] = role
	}
}
