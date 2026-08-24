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

package usermanager

import (
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/access"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
	apigwclient "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/client"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IHandler handler interface.
type IHandler interface {
	tenant.ITenantIDProvider
	access.IVirtualUserResolver

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

func (h HandlerMultiTenant) ListTenantIDs(nCtx contextx.IContext) ([]string, error) {
	// Tenant listing is a platform-level bk-user call; hold the system tenant to avoid caller-tenant permission denial.
	systemTenantCtx := contextx.From(nCtx, contextx.WithTenantID(tenant.SystemTenantID))
	resp, err := h.cli.listTenant(systemTenantCtx)
	if err != nil {
		return nil, err
	}

	tenantIDs := make([]string, len(resp))
	for idx, item := range resp {
		enabled, err := item.Status.Bool()
		if err != nil {
			return nil, fmt.Errorf("failed to convert status to bool: %w", err)
		}

		if !enabled {
			continue
		}

		tenantIDs[idx] = item.ID
	}

	return tenantIDs, nil
}

// ListALLTenants implement IHandler.
func (h HandlerMultiTenant) ListALLTenants(nCtx contextx.IContext) ([]*types.Tenant, error) {
	// Tenant listing is a platform-level bk-user call; hold the system tenant to avoid caller-tenant permission denial.
	systemTenantCtx := contextx.From(nCtx, contextx.WithTenantID(tenant.SystemTenantID))
	resp, err := h.cli.listTenant(systemTenantCtx)
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

// GetBKUsernameByLoginName gets the tenant-scoped bk_username by login_name.
//
//nolint:varnamelen // h is the conventional handler receiver name.
func (h HandlerMultiTenant) GetBKUsernameByLoginName(nCtx contextx.IContext, loginName string) (string, error) {
	if nCtx == nil {
		return "", errors.New("context is nil")
	}
	if err := nCtx.CheckTenantID(); err != nil {
		return "", err
	}
	if loginName == "" {
		return "", errors.New("login name is empty")
	}

	virtualUsers, err := h.cli.batchLookupVirtualUser(nCtx, loginName)
	if err != nil {
		return "", fmt.Errorf("failed to get bk username by login name: %w", err)
	}

	var (
		bkUsername string
		matched    bool
	)
	for _, item := range virtualUsers {
		if item.LoginName != loginName {
			continue
		}

		if matched {
			return "", fmt.Errorf("multiple virtual users found, login-name(%s)", loginName)
		}

		matched = true
		bkUsername = item.BKUsername
	}

	if !matched {
		return "", fmt.Errorf("virtual user not found, login-name(%s)", loginName)
	}

	if bkUsername == "" {
		return "", fmt.Errorf("virtual user bk username is empty, login-name(%s)", loginName)
	}

	return bkUsername, nil
}

// HandlerSingle handler of user manager.
type HandlerSingle struct {
	cli *cli
}

func (h HandlerSingle) ListTenantIDs(_ contextx.IContext) ([]string, error) {
	return []string{tenant.SingleModeTenantID}, nil
}

var _ IHandler = &HandlerSingle{}

// OptionFnSingle ...
type OptionFnSingle func(*HandlerSingle)

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
			ID:      tenant.SingleModeTenantID,
			Name:    tenant.SingleModeTenantName,
			Enabled: true,
		},
	}

	return tenants, nil
}

// GetBKUsernameByLoginName returns loginName as bk_username in single tenant mode.
func (h HandlerSingle) GetBKUsernameByLoginName(_ contextx.IContext, loginName string) (string, error) {
	if loginName == "" {
		return "", errors.New("login name is empty")
	}

	return loginName, nil
}
