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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
)

var errSystemNotFound = errors.New("IAM system not found")

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

func (c *cli) updateSystem(ctx contextx.IContext, req *UpdateSystemReq) (*UpdateSystemResp, error) {
	// A successful update has no data payload (HTTP 204).
	resp := new(BaseBroker[UpdateSystemResp])
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
		return nil, err
	}
	if result.StatusCode != http.StatusNoContent {
		return nil, fmt.Errorf("update_system: unexpected HTTP %d (request-id: %s)",
			result.StatusCode, result.Header.Get("X-Bkapi-Request-Id"))
	}
	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("update system failed: %w", err)
	}

	return &resp.Data, nil
}
