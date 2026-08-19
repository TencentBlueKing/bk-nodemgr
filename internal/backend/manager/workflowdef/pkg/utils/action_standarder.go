/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package utils provides common utilities for package workflow actions.
// nolint:revive // package name follows the workflowdef XxxUtils convention (node/utils, plugin/utils).
package utils

import (
	"fmt"

	pkgStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/pkg"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

// NewPackageActionStandarder creates a new PackageActionStandarder.
func NewPackageActionStandarder(storagePackage pkgStg.IStorage) *PackageActionStandarder {
	return &PackageActionStandarder{storagePackage: storagePackage}
}

// PackageActionStandarder defines the standard parameters of package actions.
type PackageActionStandarder struct {
	storagePackage pkgStg.IStorage

	instanceContext *action.InstanceContext
	param           PackageActionStandardParam

	ctx  contextx.IContext
	info *types.PackageDeploymentInfo
}

// Initialize initializes the PackageActionStandarder.
func (std *PackageActionStandarder) Initialize(instanceContext *action.InstanceContext, param PackageActionStandardParam) error {
	std.instanceContext = instanceContext
	std.param = param

	var err error
	std.info, err = std.storagePackage.GetPackageDeploymentInfo(std.instanceContext.Ctx, std.param.Token)
	if err != nil {
		return fmt.Errorf("failed to get package deployment info: %w", err)
	}

	std.ctx = contextx.From(std.instanceContext.Ctx, contextx.WithTenantID(param.TenantID), contextx.WithBKUsername(param.Operator))

	return nil
}

// Context returns the tenant user context.
func (std *PackageActionStandarder) Context() contextx.IContext {
	return std.ctx
}

// Param returns the package action standard parameters.
func (std *PackageActionStandarder) Param() PackageActionStandardParam {
	return std.param
}

// DeployInfo returns the package deployment info.
func (std *PackageActionStandarder) DeployInfo() *types.PackageDeploymentInfo {
	return std.info
}

// Save saves the package deployment info.
func (std *PackageActionStandarder) Save() error {
	if err := std.storagePackage.UpdatePackageDeploymentInfo(std.instanceContext.Ctx, std.param.Token, std.info); err != nil {
		return fmt.Errorf("failed to update package deployment info: %w", err)
	}

	return nil
}

// InstanceData returns the instance data.
func (std *PackageActionStandarder) InstanceData() *action.InstanceData {
	return std.instanceContext.Data
}

// ResetInstanceDataContext resets the instance data content.
func (std *PackageActionStandarder) ResetInstanceDataContext() {
	std.instanceContext.Data.Content = conv.StructToMapIgnoreError(std.param)
}

// PackageActionStandardParam defines the standard parameters of package actions.
type PackageActionStandardParam struct {
	TenantID string `json:"tenant_id"`
	Token    string `json:"token"`
	Operator string `json:"operator"`
}
