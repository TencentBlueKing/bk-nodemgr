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

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/nodedeployment"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/operation"
	operinstdataStorage "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/operinstdata"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/trigengine"
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

	// TopoStorage bk nodeman topo storage.
	TopoStorage topo.IStorage

	// TrigEngineStorage bk nodeman trigengine storage.
	TrigEngineStorage trigengine.Storage

	// OperInstStorage bk nodeman operation_inst storage.
	OperInstStorage operinstdataStorage.IStorage

	// OperStorage bk nodeman operation storage.
	OperStorage operation.Storage

	// NodeDeploymentStorage bk nodeman node deployment storage.
	NodeDeploymentStorage nodedeployment.IStorage

	// CmdbHandler cmdb handler.
	CmdbHandler cmdb.Handler

	// Logger logger
	Logger logger.Logger

	// LockerFactory locker factory
	LockerFactory locker.MutexFactory

	// Crypter ...
	Crypter crypter.Crypter

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

	if err := c.TrigEngineStorage.Start(ctx); err != nil {
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
