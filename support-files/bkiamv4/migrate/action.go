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
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/support-files/bkiamv4/migrate/iamv4"
)

func validateAction(data map[string]json.RawMessage) error {
	var action iamv4.Action
	idPattern := regexp.MustCompile(`^[a-z][a-z0-9_-]{0,31}$`)
	if err := json.Unmarshal(data[fieldID], &action.ID); err != nil {
		return fmt.Errorf("data.id is required: %w", err)
	}
	if !idPattern.MatchString(action.ID) {
		return fmt.Errorf("invalid action ID %q", action.ID)
	}
	for field, raw := range data {
		if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return fmt.Errorf("data.%s must not be null; omit it to preserve the remote value", field)
		}
		switch field {
		case fieldID:
			// Already validated above.
		case fieldName:
			if err := json.Unmarshal(raw, &action.Name); err != nil {
				return fmt.Errorf("data.name must be a string: %w", err)
			}
			if strings.TrimSpace(action.Name) == "" {
				return fmt.Errorf("data.name must not be empty")
			}
		case "resource_type_id":
			if err := json.Unmarshal(raw, &action.ResourceTypeID); err != nil {
				return fmt.Errorf("data.resource_type_id must be a string: %w", err)
			}
			if action.ResourceTypeID != "" && !idPattern.MatchString(action.ResourceTypeID) {
				return fmt.Errorf("invalid resource type ID %q", action.ResourceTypeID)
			}
		default:
			return fmt.Errorf("unknown action field %q", field)
		}
	}

	return nil
}

func (item migration) executeAction(
	ctx contextx.IContext, handler iamv4.IHandler, op operation,
	states map[string]map[string]iamv4.Action, resourceTypes map[string]map[string]iamv4.ResourceType,
	dryRun bool, out io.Writer,
) error {

	actions, loaded := states[item.SystemID]
	if !loaded {
		listed, err := handler.ListActions(ctx, item.SystemID)
		if err != nil {
			return fmt.Errorf("list actions for %s: %w", item.SystemID, err)
		}
		actions = make(map[string]iamv4.Action, len(listed))
		for _, action := range listed {
			actions[action.ID] = action
		}
		states[item.SystemID] = actions
	}
	action, err := prepareAction(op, actions)
	if err != nil {
		return err
	}
	current, exists := actions[action.ID]
	if !exists {
		if err := validateActionCreation(ctx, handler, item.SystemID, action, resourceTypes); err != nil {
			return err
		}
	}
	plan := "create_action"
	if exists {
		plan = "update_action"
		if action.Name == current.Name {
			plan = "skip_action"
		}
	}
	if _, err := fmt.Fprintf(out, "%s: %s %s/%s (dry-run=%t)\n", item.filename, plan, item.SystemID, action.ID, dryRun); err != nil {
		return fmt.Errorf("write action plan: %w", err)
	}
	if !dryRun && plan != "skip_action" {
		if err := applyAction(ctx, handler, item.SystemID, action, exists); err != nil {
			return err
		}
	}
	// Successful writes and dry-run plans both affect later operations in this run.
	actions[action.ID] = action

	return nil
}

func prepareAction(op operation, actions map[string]iamv4.Action) (iamv4.Action, error) {
	action := op.action
	current, exists := actions[action.ID]
	if exists {
		if _, supplied := op.Data["resource_type_id"]; supplied && action.ResourceTypeID != current.ResourceTypeID {
			return action, fmt.Errorf("action %s resource binding differs from remote state; automatic binding changes are not supported", action.ID)
		}
		action.ResourceTypeID = current.ResourceTypeID
		if _, supplied := op.Data[fieldName]; !supplied {
			action.Name = current.Name
		}
	}
	for _, other := range actions {
		if other.ID != action.ID && other.Name == action.Name {
			return action, fmt.Errorf("action name %q is already used by %s", action.Name, other.ID)
		}
	}

	return action, nil
}

func validateActionCreation(
	ctx contextx.IContext, handler iamv4.IHandlerResourceType, systemID string, action iamv4.Action,
	states map[string]map[string]iamv4.ResourceType,
) error {

	if action.Name == "" {
		return fmt.Errorf("data.name is required to create action %s", action.ID)
	}
	if action.ResourceTypeID == "" {
		// Omitted and explicitly empty bindings both create a resource-free action.
		return nil
	}
	resources, err := loadResourceTypes(ctx, handler, systemID, states)
	if err != nil {
		return err
	}
	if _, exists := resources[action.ResourceTypeID]; !exists {
		return fmt.Errorf("resource type %s must exist before action %s", action.ResourceTypeID, action.ID)
	}

	return nil
}

func applyAction(ctx contextx.IContext, handler iamv4.IHandlerAction, systemID string, action iamv4.Action, exists bool) error {
	if !exists {
		if err := handler.CreateAction(ctx, systemID, action); err != nil {
			return fmt.Errorf("create action %s: %w", action.ID, err)
		}

		return nil
	}
	if err := handler.UpdateAction(ctx, systemID, action.ID, action.Name); err != nil {
		return fmt.Errorf("update action %s: %w", action.ID, err)
	}

	return nil
}
