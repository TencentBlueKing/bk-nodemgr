/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package backend provides handlers to operate nodeman backend api.
package backend

import (
	"context"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Handler is interface for nodeman backend handler.
type Handler interface {
	// ListBusiness list business within specified tenant in context.
	// @param ctx context, contains tenant-id.
	// @param page describes the page info when listing.
	// @return business list.
	ListBusiness(ctx context.Context, page types.Page) ([]*types.Business, error)

	// ListHost list host within specified tenant in context.
	// @param ctx context, contains tenant-id.
	// @param page describes the page info when listing.
	// @return host list.
	ListHost(ctx context.Context, page types.Page) ([]*types.Host, error)

	// ListNetworkArea list network area within specified tenant in context.
	// @param ctx context, contains tenant-id.
	// @param page describes the page info when listing.
	// @return the network-area list.
	ListNetworkArea(ctx context.Context, page types.Page) ([]*types.NetworkArea, error)

	// ListNetworkUnit list network unit within specified tenant in context.
	// @param ctx content, contains tenant-id.
	// @param page describes the page info when listing.
	// @return the network-unit list.
	ListNetworkUnit(ctx context.Context, cloudID int64) ([]*types.NetworkUnit, error)
}

type handler struct {
	cli *cli
}

// New initialize a new nodeman backend handler.
func New(c *client.Capability, conf *Config) (Handler, error) {
	cli, err := newClient(c, conf)
	if err != nil {
		return nil, err
	}

	return &handler{cli: cli}, nil
}

// ListBusiness list business within specified tenant in context.
func (h *handler) ListBusiness(ctx context.Context, page types.Page) ([]*types.Business, error) {
	return nil, nil
}

// ListHost list host within specified tenant in context.
func (h *handler) ListHost(ctx context.Context, page types.Page) ([]*types.Host, error) {
	return nil, nil
}

// ListNetworkArea list network area within specified tenant in context.
func (h *handler) ListNetworkArea(ctx context.Context, page types.Page) ([]*types.NetworkArea, error) {
	return nil, nil
}

// ListNetworkUnit list network unit within specified tenant in context.
func (h *handler) ListNetworkUnit(ctx context.Context, cloudID int64) ([]*types.NetworkUnit, error) {
	return nil, nil
}
