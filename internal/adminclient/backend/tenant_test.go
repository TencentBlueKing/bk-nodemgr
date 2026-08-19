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

package backend

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/backendadmin"
)

type fakeInitTenantHandler struct {
	nCtx           contextx.IContext
	targetTenantID string
	calls          int
	err            error
}

func (f *fakeInitTenantHandler) InitTenant(nCtx contextx.IContext, tenantID string) error {
	f.nCtx = nCtx
	f.targetTenantID = tenantID
	f.calls++

	return f.err
}

func TestTenantCMDRegistersInitSubcommand(t *testing.T) {
	cmd := NewTenantCMD(
		func() backendadmin.IInitTenantHandler { return &backendadmin.Handler{} },
		func() (string, string) { return "default", "admin" },
	)

	require.NotNil(t, cmd)
	assert.Equal(t, "tenant", cmd.Use)
	assert.NotNil(t, findSubcommand(cmd, "init"))
}

func TestInitTenantCommandSeparatesAuthTenantAndTargetTenant(t *testing.T) {
	handler := &fakeInitTenantHandler{}
	cmd := NewInitTenantCMD(
		func() backendadmin.IInitTenantHandler { return handler },
		func() (string, string) { return "auth-tenant", "admin" },
	)
	out := new(bytes.Buffer)
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"--target-tenant-id", "target-tenant"})
	cmd.SetContext(contextx.New(context.Background(), contextx.WithTenantID("base"), contextx.WithLoginName("base-login")))

	err := cmd.Execute()

	require.NoError(t, err)
	assert.Equal(t, 1, handler.calls)
	assert.Equal(t, "auth-tenant", handler.nCtx.TenantID())
	assert.Equal(t, "admin", handler.nCtx.LoginName())
	assert.Equal(t, "target-tenant", handler.targetTenantID)
	assert.Contains(t, out.String(), "Successfully initialized tenant")
}

func TestBackendCMDRejectsMissingTargetTenantIDBeforeHandlerCreation(t *testing.T) {
	factoryCalls := 0
	cmd := NewBackendCMD(func(string) (backendadmin.IHandler, error) {
		factoryCalls++
		return &backendadmin.Handler{}, nil
	})
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"tenant", "init"})

	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "target-tenant-id")
	assert.Zero(t, factoryCalls)
}

func TestInitTenantCommandReturnsHandlerErrorWithoutSuccessOutput(t *testing.T) {
	handler := &fakeInitTenantHandler{err: assert.AnError}
	cmd := NewInitTenantCMD(
		func() backendadmin.IInitTenantHandler { return handler },
		func() (string, string) { return "auth-tenant", "admin" },
	)
	out := new(bytes.Buffer)
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"--target-tenant-id", "target-tenant"})

	err := cmd.Execute()

	require.ErrorIs(t, err, assert.AnError)
	assert.Equal(t, 1, handler.calls)
	assert.NotContains(t, out.String(), "Successfully initialized tenant")
}

func TestInitTenantCommandReturnsUnsupportedHandlerError(t *testing.T) {
	cmd := NewInitTenantCMD(
		func() backendadmin.IInitTenantHandler { return nil },
		func() (string, string) { return "auth-tenant", "admin" },
	)
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"--target-tenant-id", "target-tenant"})

	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not support initializing tenants")
}
