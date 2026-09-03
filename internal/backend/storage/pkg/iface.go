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

// Package pkg provides package storage.
package pkg

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IStorage defines the package storage interface.
type IStorage interface {
	basestorage.Interface

	IStoragePackageWorkflow
	IStoragePackageDeployment
	IStoragePackageExport
}

// IStoragePackageExport defines the interface of package export storage.
type IStoragePackageExport interface {
	// ListPackageExport lists package export records.
	ListPackageExport(nCtx contextx.IContext, page types.Page, conditions ...*types.PackageExportCondition) ([]*types.PackageExport, int64, error)

	// CreatePackageExport creates a package export record.
	CreatePackageExport(nCtx contextx.IContext, exportData *types.PackageExport) error

	// GetPackageExport gets a package export by export ID.
	GetPackageExport(nCtx contextx.IContext, exportID string) (*types.PackageExport, error)

	// UpdatePackageExport updates package export fields.
	UpdatePackageExport(nCtx contextx.IContext, fields types.PackageExportFields, exportData ...*types.PackageExport) error

	// DeletePackageExport deletes a package export by export ID.
	DeletePackageExport(nCtx contextx.IContext, exportID string) error
}

// IStoragePackageWorkflow defines the interface of package workflow storage.
type IStoragePackageWorkflow interface {
	// CreatePackageWorkflow creates a new package workflow.
	CreatePackageWorkflow(nCtx contextx.IContext, workflow *types.PackageWorkflow) error

	// GetPackageWorkflow gets a package workflow by workflow ID.
	GetPackageWorkflow(nCtx contextx.IContext, workflowID string) (*types.PackageWorkflow, error)

	// GetPackageExportWorkflowByTriggerID gets the unique export workflow by trigger ID.
	GetPackageExportWorkflowByTriggerID(nCtx contextx.IContext, triggerID string) (*types.PackageWorkflow, error)
}

// IStoragePackageDeployment defines the interface of package deployment storage.
type IStoragePackageDeployment interface {
	// CreatePackageDeployment creates a package deployment record.
	CreatePackageDeployment(nCtx contextx.IContext, deployment *types.PackageDeployment) error

	// ListPackageDeployment lists package deployment records.
	ListPackageDeployment(nCtx contextx.IContext, page types.Page, conditions ...*types.PackageDeploymentCondition) (
		[]*types.PackageDeployment, int64, error)

	// GetPackageDeploymentInfo gets package deployment info by token.
	GetPackageDeploymentInfo(nCtx contextx.IContext, token string) (*types.PackageDeploymentInfo, error)

	// UpdatePackageDeploymentInfo updates package deployment info by token.
	UpdatePackageDeploymentInfo(nCtx contextx.IContext, token string, info *types.PackageDeploymentInfo) error
}
