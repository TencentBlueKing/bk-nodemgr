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
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	apigwclient "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/client"
	apigwheader "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/header"
)

var errSystemNotFound = errors.New("IAM system not found")

type cli struct {
	client restclient.IClient
	config *Config
}

func newClient(c *restclient.Capability, conf *Config) (*cli, error) {
	if err := conf.Validate(); err != nil {
		return nil, err
	}
	client, err := apigwclient.NewClient(c, conf.BaseURL+"/api/v1/open", apigwclient.VirtualUserConfig{AppConfig: conf.AppConfig},
		restclient.WithCustomHeaderMasker(apigwheader.BKGWAuthKey, func(string) string { return "[REDACTED]" }))
	if err != nil {
		return nil, fmt.Errorf("create APIGW client: %w", err)
	}

	return &cli{client: client, config: conf}, nil
}

func (c *cli) getHeader(ctx contextx.IContext) http.Header {
	header := http.Header{}
	header.Set(apigwheader.BKGWAuthKey, c.config.GetAuthHeader())
	header.Set(apigwheader.BKGWTenantIDKey, ctx.TenantID())

	return header
}

func (c *cli) retrieveSystem(ctx contextx.IContext, req *RetrieveSystemReq) (*RetrieveSystemResp, error) {
	resp := new(BaseBroker[RetrieveSystemResp])
	header := c.getHeader(ctx)
	result := c.client.Get().
		SubResourcef("/rbac/model/systems/%s/", req.SystemID).
		WithContext(ctx).
		WithHeaders(header).
		EnableLogBody().
		EnableLogResponse().
		Do()
	if err := result.Into(resp); err != nil {
		return nil, err
	}
	// Only this status/code pair means absence rather than a failed query.
	if result.StatusCode == http.StatusNotFound && resp.Error != nil && resp.Error.Code == "NOT_FOUND" {
		return nil, errSystemNotFound
	}
	if result.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("retrieve_system: unexpected HTTP %d (request-id: %s)",
			result.StatusCode, result.Header.Get("X-Bkapi-Request-Id"))
	}
	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("retrieve system failed: %w", err)
	}
	if resp.Data.ID != req.SystemID {
		return nil, fmt.Errorf("retrieve system %s: response system ID does not match", req.SystemID)
	}

	return &resp.Data, nil
}

func (c *cli) createSystem(ctx contextx.IContext, req *CreateSystemReq) (*CreateSystemResp, error) {
	resp := new(BaseBroker[CreateSystemResp])
	header := c.getHeader(ctx)
	result := c.client.Post().
		SubResourcef("/rbac/model/systems/").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do()
	if err := result.Into(resp); err != nil {
		return nil, err
	}
	if result.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("create_system: unexpected HTTP %d (request-id: %s)",
			result.StatusCode, result.Header.Get("X-Bkapi-Request-Id"))
	}
	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("create system failed: %w", err)
	}
	if resp.Data.ID != req.ID {
		return nil, fmt.Errorf("create system %s: response system ID does not match; check remote state before rerunning", req.ID)
	}

	return &resp.Data, nil
}

func (c *cli) updateSystem(ctx contextx.IContext, req *UpdateSystemReq) error {
	// A successful update has no data payload (HTTP 204).
	resp := new(BaseBroker[struct{}])
	header := c.getHeader(ctx)
	result := c.client.Put().
		SubResourcef("/rbac/model/systems/%s/", req.SystemID).
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do()
	if err := result.Into(resp); err != nil {
		return err
	}
	if result.StatusCode != http.StatusNoContent {
		return fmt.Errorf("update_system: unexpected HTTP %d (request-id: %s)",
			result.StatusCode, result.Header.Get("X-Bkapi-Request-Id"))
	}
	if err := resp.IsFailed(); err != nil {
		return fmt.Errorf("update system failed: %w", err)
	}

	return nil
}

func (c *cli) listResourceTypes(
	ctx contextx.IContext, req *ListResourceTypesReq,
) (*BaseBroker[*ListResourceTypesResp], error) {

	resp := new(BaseBroker[*ListResourceTypesResp])
	result := c.client.Get().
		SubResourcef("/rbac/model/systems/%s/resource-types/", req.SystemID).
		WithContext(ctx).
		WithHeaders(c.getHeader(ctx)).
		WithParam("page", strconv.Itoa(req.Page)).
		WithParam("page_size", strconv.Itoa(req.PageSize)).
		EnableLogBody().
		EnableLogResponse().
		Do()
	if err := result.Into(resp); err != nil {
		return nil, fmt.Errorf("list resource types page %d: %w", req.Page, err)
	}
	if result.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list_resource_types: unexpected HTTP %d (request-id: %s)",
			result.StatusCode, result.Header.Get("X-Bkapi-Request-Id"))
	}
	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("list resource types failed: %w", err)
	}

	return resp, nil
}

func (c *cli) batchCreateResourceType(
	ctx contextx.IContext, req *BatchCreateResourceTypeReq,
) (*BaseBroker[BatchCreateResourceTypeResp], error) {

	resp := new(BaseBroker[BatchCreateResourceTypeResp])
	result := c.client.Post().
		SubResourcef("/rbac/model/systems/%s/resource-types/", req.SystemID).
		WithContext(ctx).
		WithHeaders(c.getHeader(ctx)).
		Body(req.Resources).
		EnableLogBody().
		EnableLogResponse().
		Do()
	if err := result.Into(resp); err != nil {
		return nil, fmt.Errorf("create resource type: %w", err)
	}
	if result.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("create_resource_type: unexpected HTTP %d (request-id: %s)",
			result.StatusCode, result.Header.Get("X-Bkapi-Request-Id"))
	}
	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("create resource type failed: %w", err)
	}

	return resp, nil
}

func (c *cli) updateResourceType(ctx contextx.IContext, req *UpdateResourceTypeReq) error {
	// A successful update has no data payload (HTTP 204).
	resp := new(BaseBroker[UpdateResourceTypeResp])
	result := c.client.Put().
		SubResourcef("/rbac/model/systems/%s/resource-types/%s/", req.SystemID, req.ResourceTypeID).
		WithContext(ctx).
		WithHeaders(c.getHeader(ctx)).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do()
	if err := result.Into(resp); err != nil {
		return fmt.Errorf("update resource type: %w", err)
	}
	if result.StatusCode != http.StatusNoContent {
		return fmt.Errorf("update_resource_type: unexpected HTTP %d (request-id: %s)",
			result.StatusCode, result.Header.Get("X-Bkapi-Request-Id"))
	}
	if err := resp.IsFailed(); err != nil {
		return fmt.Errorf("update resource type failed: %w", err)
	}

	return nil
}

func (c *cli) listActions(ctx contextx.IContext, req *ListActionsReq) (*BaseBroker[*ListActionsResp], error) {
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

	return resp, nil
}

func (c *cli) batchCreateAction(ctx contextx.IContext, req *BatchCreateActionReq) (*BaseBroker[BatchCreateActionResp], error) {
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

	return resp, nil
}

func (c *cli) updateAction(ctx contextx.IContext, req *UpdateActionReq) error {
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
		return fmt.Errorf("update action: %w", err)
	}
	if result.StatusCode != http.StatusNoContent {
		return fmt.Errorf("update_action: unexpected HTTP %d (request-id: %s)",
			result.StatusCode, result.Header.Get("X-Bkapi-Request-Id"))
	}
	if err := resp.IsFailed(); err != nil {
		return fmt.Errorf("update action failed: %w", err)
	}

	return nil
}
