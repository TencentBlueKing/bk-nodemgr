/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package options provides the various capabilities the service supports.
package options

import (
	"context"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager"
	nodedeployment "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node-deployment"
	nodeworkflow "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node-workflow"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/operation"
	operinstdataStorage "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/operinstdata"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/trigger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/crypter"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/discover"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/locker"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/cmdb"
)

// Capability encapsulates the various capabilities the service supports.
type Capability struct {
	// Manager workflow management.
	Manager manager.Manager

	// TopoStorage topo storage.
	TopoStorage topoStg.IStorage

	// TriggerStorage trigger storage.
	TriggerStorage trigger.IStorage

	// OperInstStorage operation_inst storage.
	OperInstStorage operinstdataStorage.IStorage

	// OperStorage operation storage.
	OperStorage operation.Storage

	// NodeDeploymentStorage node deployment storage.
	NodeDeploymentStorage nodedeployment.IStorage

	// NodeWorkflowStorage node workflow storage.
	NodeWorkflowStorage nodeworkflow.IStorage

	// CmdbHandler cmdb handler.
	CmdbHandler cmdb.IHandler

	// GSEHandler gse handler.
	GSEHandler gse.IHandler

	// Logger logger
	Logger logger.Logger

	// LockerFactory locker factory
	LockerFactory locker.MutexFactory

	// Crypter ...
	Crypter crypter.Crypter

	// InstallerFileGroup tool file group.
	InstallerFileGroup iface.FileGroup

	// Discover provides discover handler.
	DiscoverProvider discover.Provider
}

// Start ...
func (c *Capability) Start(ctx context.Context) error {
	if err := c.DiscoverProvider.Start(ctx); err != nil {
		return err
	}

	if err := c.TopoStorage.Start(ctx); err != nil {
		return err
	}

	if err := c.TriggerStorage.Start(ctx); err != nil {
		return err
	}

	if err := c.OperInstStorage.Start(ctx); err != nil {
		return err
	}

	if err := c.OperStorage.Start(ctx); err != nil {
		return err
	}

	if err := c.NodeDeploymentStorage.Start(ctx); err != nil {
		return err
	}

	if err := c.NodeWorkflowStorage.Start(ctx); err != nil {
		return err
	}

	if err := c.Manager.Start(ctx); err != nil {
		return err
	}

	return nil
}

// GracefulShutdown ...
func (c *Capability) GracefulShutdown() error {
	if err := c.Manager.GracefulShutdown(); err != nil {
		return err
	}

	return nil
}
