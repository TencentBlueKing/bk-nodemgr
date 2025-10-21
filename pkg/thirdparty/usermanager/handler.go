/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package usermanager

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	apigwclient "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/client"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IHandler handler interface.
type IHandler interface {
	ListALLTenants(nCtx contextx.IContext) ([]*types.Tenant, error)
}

var _ IHandler = &HandlerMultiTenant{}

// Config handler config of user manager.
type Config struct {
	APIGWUserConfig apigwclient.UserConfig
}

// Validate validate config.
func (conf *Config) Validate() error {
	return nil
}

// HandlerMultiTenant handler of user manager.
type HandlerMultiTenant struct {
	cli *cli
}

// OptionFnMultiTenant ...
type OptionFnMultiTenant func(*HandlerMultiTenant)

// NewHandlerMultiTenant initialize a new user manager Handler.
func NewHandlerMultiTenant(c *restclient.Capability, conf *Config, opts ...OptionFnMultiTenant) (*HandlerMultiTenant, error) {
	cli, err := newClient(c, conf)
	if err != nil {
		return nil, err
	}

	handler := &HandlerMultiTenant{
		cli: cli,
	}

	for _, opt := range opts {
		opt(handler)
	}

	return handler, nil
}

// ListALLTenants implement IHandler.
func (h HandlerMultiTenant) ListALLTenants(nCtx contextx.IContext) ([]*types.Tenant, error) {
	req := &listTenantReq{}
	resp, err := h.cli.listTenant(nCtx, req)
	if err != nil {
		return nil, err
	}

	tenants := make([]*types.Tenant, len(resp))
	for idx, item := range resp {
		enabled, err := item.Status.Bool()
		if err != nil {
			return nil, fmt.Errorf("failed to convert status to bool: %w", err)
		}

		tenants[idx] = &types.Tenant{
			ID:      item.ID,
			Name:    item.Name,
			Enabled: enabled,
		}
	}

	return tenants, nil
}

// HandlerSingle handler of user manager.
type HandlerSingle struct {
	cli *cli
}

var _ IHandler = &HandlerSingle{}

// OptionFnSingle ...
type OptionFnSingle func(*HandlerSingle)

const (
	// singleModeTenantID default tenant id.
	singleModeTenantID = "default"

	// singleModeTenantName default tenant name.
	singleModeTenantName = "default"
)

// NewHandlerSingle initialize a new user manager Handler.
func NewHandlerSingle(c *restclient.Capability, conf *Config, opts ...OptionFnSingle) (*HandlerSingle, error) {
	cli, err := newClient(c, conf)
	if err != nil {
		return nil, err
	}

	handler := &HandlerSingle{
		cli: cli,
	}

	for _, opt := range opts {
		opt(handler)
	}

	return handler, nil
}

// ListALLTenants implement IHandler.
func (h HandlerSingle) ListALLTenants(_ contextx.IContext) ([]*types.Tenant, error) {
	tenants := []*types.Tenant{
		{
			ID:      singleModeTenantID,
			Name:    singleModeTenantName,
			Enabled: true,
		},
	}

	return tenants, nil
}
