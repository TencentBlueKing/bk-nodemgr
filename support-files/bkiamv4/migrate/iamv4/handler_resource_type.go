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
	"slices"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/pageexecutor"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IHandlerResourceType exposes resource type operations without HTTP details.
type IHandlerResourceType interface {
	ListResourceTypes(ctx contextx.IContext, systemID string) ([]ResourceType, error)
	CreateResourceType(ctx contextx.IContext, systemID string, resource ResourceType) error
	UpdateResourceType(ctx contextx.IContext, systemID, resourceTypeID string, fields ResourceTypeFields) error
}

// ListResourceTypes retrieves the complete resource type model or returns an error.
func (h *Handler) ListResourceTypes(ctx contextx.IContext, systemID string) ([]ResourceType, error) {
	const pageSize = 100             // IAM's maximum model query page size.
	const timeout = 30 * time.Second // Bound the complete model query, not each page.
	seen := make(map[string]struct{})
	total := 0
	executor := pageexecutor.NewPageExecutor[ResourceType](pageSize, timeout)
	fn := func(ctx contextx.IContext, p types.Page) ([]ResourceType, error) {
		if p.Offset > 0 && len(seen) == total {
			// The last full page already completed the model.
			return nil, nil
		}
		req := &ListResourceTypesReq{SystemID: systemID, Page: p.Offset/pageSize + 1, PageSize: pageSize}
		resp, err := h.cli.listResourceTypes(ctx, req)
		if err != nil {
			return nil, err
		}
		if err := validateResourceTypePage(req, resp, total, seen); err != nil {
			return nil, err
		}
		total = *resp.Count

		return resp.Results, nil
	}
	resources, err := executor.Execute(ctx, types.UnlimitedPage(), fn)
	if err != nil {
		return nil, fmt.Errorf("list resource types: %w", err)
	}
	if len(resources.Items) != total {
		return nil, fmt.Errorf("list resource types: incomplete pagination, got %d of %d", len(resources.Items), total)
	}

	return resources.Items, nil
}

func validateResourceTypePage(
	req *ListResourceTypesReq, resp *ListResourceTypesResp, total int, seen map[string]struct{},
) error {

	if resp.Count == nil || *resp.Count < 0 || resp.Results == nil {
		return fmt.Errorf("list resource types page %d: incomplete response", req.Page)
	}
	if req.Page == 1 {
		total = *resp.Count
	}
	if *resp.Count != total || len(resp.Results) > req.PageSize || len(resp.Results) > total-len(seen) {
		return fmt.Errorf("list resource types page %d: inconsistent count", req.Page)
	}
	for _, resource := range resp.Results {
		if resource.ID == "" || resource.Name == "" || resource.Ancestors == nil || slices.Contains(resource.Ancestors, "") {
			return fmt.Errorf("list resource types page %d: malformed resource type", req.Page)
		}
		if _, exists := seen[resource.ID]; exists {
			return fmt.Errorf("list resource types page %d: duplicate resource type ID", req.Page)
		}
		seen[resource.ID] = struct{}{}
	}

	return nil
}

// CreateResourceType registers a resource type with its ancestor chain.
func (h *Handler) CreateResourceType(ctx contextx.IContext, systemID string, resource ResourceType) error {
	if resource.ID == "" || resource.Name == "" {
		return fmt.Errorf("id and name are required to create a resource type")
	}
	// Root resource types must send an empty array rather than JSON null.
	if resource.Ancestors == nil {
		resource.Ancestors = make([]string, 0)
	}

	resp, err := h.cli.batchCreateResourceType(ctx, &BatchCreateResourceTypeReq{
		SystemID:  systemID,
		Resources: []ResourceType{resource},
	})
	if err != nil {
		return err
	}
	if len(*resp) != 1 || (*resp)[0] != resource.ID {
		return fmt.Errorf("create resource type: response ID does not match; check remote state before rerunning")
	}

	return nil
}

// UpdateResourceType sends supplied fields; the caller owns topology drift checks.
func (h *Handler) UpdateResourceType(
	ctx contextx.IContext, systemID, resourceTypeID string, fields ResourceTypeFields,
) error {

	_, err := h.cli.updateResourceType(ctx, &UpdateResourceTypeReq{
		SystemID:           systemID,
		ResourceTypeID:     resourceTypeID,
		ResourceTypeFields: fields,
	})

	return err
}
