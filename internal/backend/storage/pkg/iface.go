/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package pkg provides package storage.
package pkg

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	packagedeployment "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/package-deployment"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IStorage defines the package storage interface.
type IStorage interface {
	basestorage.Interface

	IStoragePackageWorkflow
	IStoragePackageDeployment
}

// IStoragePackageWorkflow defines the interface of package workflow storage.
type IStoragePackageWorkflow interface {
	// CreatePackageWorkflow creates a new package workflow.
	CreatePackageWorkflow(nCtx contextx.IContext, workflow *types.PackageWorkflow) error

	// GetPackageWorkflow gets a package workflow by workflow ID.
	GetPackageWorkflow(nCtx contextx.IContext, workflowID string) (*types.PackageWorkflow, error)
}

// IStoragePackageDeployment defines the interface of package deployment storage.
type IStoragePackageDeployment interface {
	// CreatePackageDeployment creates a package deployment record.
	CreatePackageDeployment(nCtx contextx.IContext, deployment *types.PackageDeployment) error

	// ListPackageDeployment lists package deployment records.
	ListPackageDeployment(nCtx contextx.IContext, page types.Page, opts ...packagedeployment.OptFn) ([]*types.PackageDeployment, int64, error)

	// GetPackageDeploymentInfo gets package deployment info by token.
	GetPackageDeploymentInfo(nCtx contextx.IContext, token string) (*types.PackageDeploymentInfo, error)

	// UpdatePackageDeploymentInfo updates package deployment info by token.
	UpdatePackageDeploymentInfo(nCtx contextx.IContext, token string, info *types.PackageDeploymentInfo) error
}
