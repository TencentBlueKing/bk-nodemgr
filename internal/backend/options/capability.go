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
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/periodictask"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/configpolicy"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/credit"
	globalsettingsStorage "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/globalsettings"
	nodedeployment "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node-deployment"
	nodeworkflow "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node-workflow"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/operation"
	operinstdataStorage "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/operinstdata"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/release"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/scheduleworkflow"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/trigger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/creditvault"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/relayhandler"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/crypter"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/discover"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/locker"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/cmdb"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/file"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
)

// Capability encapsulates the various capabilities the service supports.
type Capability struct {
	// Manager workflow management.
	Manager manager.IManager

	// StorageTopo topo storage.
	StorageTopo topo.IStorage

	// StorageTrigger trigger storage.
	StorageTrigger trigger.IStorage

	// StorageOperInst operation_inst storage.
	StorageOperInst operinstdataStorage.IStorage

	// StorageOperation operation storage.
	StorageOperation operation.IStorage

	// StorageNodeDeployment node deployment storage.
	StorageNodeDeployment nodedeployment.IStorage

	// StorageNodeWorkflow node workflow storage.
	StorageNodeWorkflow nodeworkflow.IStorage

	// StorageScheduleWorkflow schedule workflow storage.
	StorageScheduleWorkflow scheduleworkflow.IStorage

	// StorageRelease release storage.
	StorageRelease release.IStorage

	// StorageGlobalSettings global settings storage.
	StorageGlobalSettings globalsettingsStorage.IStorage

	// StorageCredit credit storage.
	StorageCredit credit.IStorage

	// StorageConfigPolicy config policy storage.
	StorageConfigPolicy configpolicy.IStorage

	// CmdbHandler cmdb handler.
	CmdbHandler cmdb.IHandler

	// GSEHandler gse handler.
	GSEHandler gse.IHandler

	// FileHandler file handler.
	FileHandler file.IHandler

	// Logger logger
	Logger logger.ILogger

	// LockerFactory locker factory
	LockerFactory locker.MutexFactory

	// Crypter ...
	Crypter crypter.Crypter

	// InstallerFileGroup tool file group.
	InstallerFileGroup fileiface.FileGroup

	// Discover provides discover handler.
	DiscoverProvider discover.Provider

	// ProxyMessager provides the proxy messager.
	ProxyMessager relayhandler.IServerMessager

	// CreditVault credit vault.
	CreditVault creditvault.ICreditVault

	// PeriodicTask periodic task manager.
	PeriodicTask periodictask.PeriodicTask
}

// Start ...
func (capability *Capability) Start(ctx contextx.IContext) error {
	if err := capability.DiscoverProvider.Start(ctx); err != nil {
		return err
	}

	if err := capability.StorageTopo.Start(ctx); err != nil {
		return err
	}

	if err := capability.StorageTrigger.Start(ctx); err != nil {
		return err
	}

	if err := capability.StorageOperInst.Start(ctx); err != nil {
		return err
	}

	if err := capability.StorageOperation.Start(ctx); err != nil {
		return err
	}

	if err := capability.StorageNodeDeployment.Start(ctx); err != nil {
		return err
	}

	if err := capability.StorageNodeWorkflow.Start(ctx); err != nil {
		return err
	}

	if err := capability.StorageScheduleWorkflow.Start(ctx); err != nil {
		return err
	}

	if err := capability.StorageRelease.Start(ctx); err != nil {
		return err
	}

	if err := capability.StorageGlobalSettings.Start(ctx); err != nil {
		return err
	}

	if err := capability.ProxyMessager.Start(ctx); err != nil {
		return err
	}

	if err := capability.StorageCredit.Start(ctx); err != nil {
		return err
	}

	if err := capability.StorageConfigPolicy.Start(ctx); err != nil {
		return err
	}

	if err := capability.Manager.Start(ctx); err != nil {
		return err
	}

	if err := capability.PeriodicTask.Start(); err != nil {
		return err
	}

	return nil
}

// GracefulShutdown ...
func (capability *Capability) GracefulShutdown() error {
	if err := capability.Manager.GracefulShutdown(); err != nil {
		return err
	}

	return nil
}
