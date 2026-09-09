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

func validateResourceType(data map[string]json.RawMessage) error {
	var resource iamv4.ResourceType
	idPattern := regexp.MustCompile(`^[a-z][a-z0-9_-]{0,31}$`)
	if err := json.Unmarshal(data[fieldID], &resource.ID); err != nil {
		return fmt.Errorf("data.id is required: %w", err)
	}
	if !idPattern.MatchString(resource.ID) {
		return fmt.Errorf("invalid resource type ID %q", resource.ID)
	}
	for field, raw := range data {
		if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return fmt.Errorf("data.%s must not be null; omit it to preserve the remote value", field)
		}
		switch field {
		case fieldID:
			// Already validated above.
		case fieldName:
			if err := json.Unmarshal(raw, &resource.Name); err != nil {
				return fmt.Errorf("data.name must be a string: %w", err)
			}
			if strings.TrimSpace(resource.Name) == "" {
				return fmt.Errorf("data.name must not be empty")
			}
		case "ancestors":
			if err := json.Unmarshal(raw, &resource.Ancestors); err != nil {
				return fmt.Errorf("data.ancestors must be a string array: %w", err)
			}
		default:
			return fmt.Errorf("unknown resource type field %q", field)
		}
	}
	for index, ancestor := range resource.Ancestors {
		if !idPattern.MatchString(ancestor) || ancestor == resource.ID || slices.Contains(resource.Ancestors[:index], ancestor) {
			return fmt.Errorf("invalid or repeated ancestor %q for resource type %s", ancestor, resource.ID)
		}
	}

	return nil
}

//nolint:gocognit // Keep topology checks and the corresponding upsert plan in one flow.
func (item migration) executeResourceType(
	ctx contextx.IContext, handler iamv4.ResourceTypeHandler, op operation,
	states map[string]map[string]iamv4.ResourceType, dryRun bool, out io.Writer,
) error {

	resources, err := loadResourceTypes(ctx, handler, item.SystemID, states)
	if err != nil {
		return err
	}
	resource := op.resource
	current, exists := resources[resource.ID]
	if exists {
		if _, supplied := op.Data["ancestors"]; supplied && !slices.Equal(resource.Ancestors, current.Ancestors) {
			return fmt.Errorf("resource type %s ancestors differ from remote state; automatic topology changes are not supported", resource.ID)
		}
		resource.Ancestors = current.Ancestors
		if _, supplied := op.Data[fieldName]; !supplied {
			resource.Name = current.Name
		}
	}
	if !exists && resource.Name == "" {
		return fmt.Errorf("data.name is required to create resource type %s", resource.ID)
	}
	if err := validateResourceTypeRelations(resource, resources); err != nil {
		return err
	}
	action := "create_resource_type"
	if exists {
		action = "update_resource_type"
		if resource.Name == current.Name {
			action = "skip_resource_type"
		}
	}
	if _, err := fmt.Fprintf(out, "%s: %s %s/%s (dry-run=%t)\n", item.filename, action, item.SystemID, resource.ID, dryRun); err != nil {
		return fmt.Errorf("write resource type plan: %w", err)
	}
	if !dryRun && action != "skip_resource_type" {
		if err := applyResourceType(ctx, handler, item.SystemID, resource, exists); err != nil {
			return err
		}
	}
	// Successful writes and dry-run plans both affect later operations in this run.
	resources[resource.ID] = resource

	return nil
}

func loadResourceTypes(
	ctx contextx.IContext, handler iamv4.ResourceTypeHandler, systemID string,
	states map[string]map[string]iamv4.ResourceType,
) (map[string]iamv4.ResourceType, error) {

	if resources, loaded := states[systemID]; loaded {
		return resources, nil
	}
	listed, err := handler.ListResourceTypes(ctx, systemID)
	if err != nil {
		return nil, fmt.Errorf("list resource types for %s: %w", systemID, err)
	}
	resources := make(map[string]iamv4.ResourceType, len(listed))
	for _, resource := range listed {
		resources[resource.ID] = resource
	}
	states[systemID] = resources

	return resources, nil
}

func validateResourceTypeRelations(resource iamv4.ResourceType, resources map[string]iamv4.ResourceType) error {
	for index, ancestorID := range resource.Ancestors {
		ancestor, exists := resources[ancestorID]
		if !exists {
			return fmt.Errorf("ancestor %s must exist before resource type %s", ancestorID, resource.ID)
		}
		if !slices.Equal(ancestor.Ancestors, resource.Ancestors[:index]) {
			return fmt.Errorf("resource type %s ancestor chain does not match ancestor %s", resource.ID, ancestorID)
		}
	}
	for _, other := range resources {
		if other.ID != resource.ID && other.Name == resource.Name {
			return fmt.Errorf("resource type name %q is already used by %s", resource.Name, other.ID)
		}
	}

	return nil
}

func applyResourceType(
	ctx contextx.IContext, handler iamv4.ResourceTypeHandler, systemID string, resource iamv4.ResourceType, exists bool,
) error {

	if !exists {
		if err := handler.CreateResourceType(ctx, systemID, resource); err != nil {
			return fmt.Errorf("create resource type %s: %w", resource.ID, err)
		}

		return nil
	}
	// This migration only updates names; topology changes require a separate decision.
	if err := handler.UpdateResourceType(ctx, systemID, resource.ID, iamv4.ResourceTypeFields{Name: &resource.Name}); err != nil {
		return fmt.Errorf("update resource type %s: %w", resource.ID, err)
	}

	return nil
}
