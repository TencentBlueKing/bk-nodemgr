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

// Package iamv4 encapsulates IAM V4 model APIs for the migration tool.
package iamv4

import (
	"errors"
	"fmt"
	"slices"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
)

// IHandler exposes system registration operations without HTTP details.
type IHandler interface {
	SystemExists(ctx contextx.IContext, systemID string) (bool, error)
	CreateSystem(ctx contextx.IContext, systemID string, fields SystemFields) error
	UpdateSystem(ctx contextx.IContext, systemID string, fields SystemFields) error
}

// ResourceTypeHandler exposes resource type operations without HTTP details.
type ResourceTypeHandler interface {
	ListResourceTypes(ctx contextx.IContext, systemID string) ([]ResourceType, error)
	CreateResourceType(ctx contextx.IContext, systemID string, resource ResourceType) error
	UpdateResourceType(ctx contextx.IContext, systemID, resourceTypeID string, fields ResourceTypeFields) error
}

// Handler adapts model registration operations to the IAM V4 client.
type Handler struct {
	cli *cli
}

var _ IHandler = (*Handler)(nil)
var _ ResourceTypeHandler = (*Handler)(nil)

// New initializes the migration tool's IAM V4 handler.
func New(c *restclient.Capability, conf *Config) (*Handler, error) {
	client, err := newClient(c, conf)
	if err != nil {
		return nil, err
	}

	return &Handler{cli: client}, nil
}

// SystemExists distinguishes an absent system from a failed query.
func (h *Handler) SystemExists(ctx contextx.IContext, systemID string) (bool, error) {
	_, err := h.cli.retrieveSystem(ctx, &RetrieveSystemReq{SystemID: systemID})
	if errors.Is(err, errSystemNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

// CreateSystem registers a system with the supplied fields.
func (h *Handler) CreateSystem(ctx contextx.IContext, systemID string, fields SystemFields) error {
	if fields.Name == nil || fields.Clients == nil {
		return fmt.Errorf("name and clients are required to create system %s", systemID)
	}
	_, err := h.cli.createSystem(ctx, &CreateSystemReq{ID: systemID, SystemFields: fields})

	return err
}

// UpdateSystem sends only supplied mutable fields, never the immutable system ID.
func (h *Handler) UpdateSystem(ctx contextx.IContext, systemID string, fields SystemFields) error {
	return h.cli.updateSystem(ctx, &UpdateSystemReq{SystemID: systemID, SystemFields: fields})
}

// ListResourceTypes retrieves the complete resource type model or returns an error.
//
//nolint:gocognit // Keep pagination completeness and duplicate checks beside aggregation.
func (h *Handler) ListResourceTypes(ctx contextx.IContext, systemID string) ([]ResourceType, error) {
	const pageSize = 100 // IAM's maximum model query page size.
	resources := make([]ResourceType, 0)
	seen := make(map[string]struct{})
	total := 0
	for page := 1; ; page++ {
		resp, err := h.cli.listResourceTypes(ctx, &ListResourceTypesReq{
			SystemID: systemID,
			Page:     page,
			PageSize: pageSize,
		})
		if err != nil {
			return nil, err
		}
		if resp.Data == nil || resp.Data.Count == nil || *resp.Data.Count < 0 || resp.Data.Results == nil {
			return nil, fmt.Errorf("list resource types page %d: incomplete response (request-id: %s)", page, resp.RequestID)
		}
		if page == 1 {
			total = *resp.Data.Count
		}
		if *resp.Data.Count != total || len(resp.Data.Results) > pageSize || len(resp.Data.Results) > total-len(resources) {
			return nil, fmt.Errorf("list resource types page %d: inconsistent count (request-id: %s)", page, resp.RequestID)
		}
		for _, resource := range resp.Data.Results {
			if resource.ID == "" || resource.Name == "" || resource.Ancestors == nil || slices.Contains(resource.Ancestors, "") {
				return nil, fmt.Errorf("list resource types page %d: malformed resource type (request-id: %s)", page, resp.RequestID)
			}
			if _, exists := seen[resource.ID]; exists {
				return nil, fmt.Errorf("list resource types page %d: duplicate resource type ID (request-id: %s)", page, resp.RequestID)
			}
			seen[resource.ID] = struct{}{}
			resources = append(resources, resource)
		}
		if len(resources) == total {
			return resources, nil
		}
		if len(resp.Data.Results) == 0 {
			return nil, fmt.Errorf("list resource types page %d: pagination made no progress (request-id: %s)", page, resp.RequestID)
		}
	}
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
	if len(resp.Data) != 1 || resp.Data[0] != resource.ID {
		return fmt.Errorf("create resource type: response ID does not match; check remote state before rerunning (request-id: %s)",
			resp.RequestID)
	}

	return nil
}

// UpdateResourceType sends supplied fields; the caller owns topology drift checks.
func (h *Handler) UpdateResourceType(
	ctx contextx.IContext, systemID, resourceTypeID string, fields ResourceTypeFields,
) error {

	return h.cli.updateResourceType(ctx, &UpdateResourceTypeReq{
		SystemID:           systemID,
		ResourceTypeID:     resourceTypeID,
		ResourceTypeFields: fields,
	})
}
