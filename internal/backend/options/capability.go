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
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/deploypolicy"
	globalsettingsStorage "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/globalsettings"
	nodeStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/release"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/tenant"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/creditvault"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/discover"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/relayhandler"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/crypter"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/locker"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/cmdb"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/file"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/usermanager"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

// Capability encapsulates the various capabilities the service supports.
type Capability struct {
	// Manager workflow management.
	Manager manager.IManager

	// RedisClient redis client.
	RedisClient *redis.Client

	// MongoClient mongo client.
	MongoClient *mongo.Client

	// StorageTopo topo storage.
	StorageTopo topoStg.IStorage

	// StorageNode node storage.
	StorageNode nodeStg.IStorage

	// StoragePlugin plugin deployment storage.
	StoragePlugin pluginStg.IStorage

	// StorageScheduleWorkflow schedule workflow storage.
	StorageWorkflow workflow.IStorage

	// StorageRelease release storage.
	StorageRelease release.IStorage

	// StorageGlobalSettings global settings storage.
	StorageGlobalSettings globalsettingsStorage.IStorage

	// StorageCredit credit storage.
	StorageCredit credit.IStorage

	// StorageConfigPolicy config policy storage.
	StorageConfigPolicy configpolicy.IStorage

	// StorageDeployPolicy deploy policy storage.
	StorageDeployPolicy deploypolicy.IStorage

	// StorageTenant tenant storage.
	StorageTenant tenant.IStorage

	// CmdbHandler cmdb handler.
	CmdbHandler cmdb.IHandler

	// GSEHandler gse handler.
	GSEHandler gse.IHandler

	// FileHandler file handler.
	FileHandler file.IHandler

	// UserManagerHandler user manager handler.
	UserManagerHandler usermanager.IHandler

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
	PeriodicTask periodictask.IPeriodicTask
}

// Start ...
func (capability *Capability) Start(ctx contextx.IContext) error {
	if err := capability.DiscoverProvider.Start(ctx); err != nil {
		return err
	}

	if err := capability.StorageTopo.Start(ctx); err != nil {
		return err
	}

	if err := capability.StorageNode.Start(ctx); err != nil {
		return err
	}

	if err := capability.StoragePlugin.Start(ctx); err != nil {
		return err
	}

	if err := capability.StorageWorkflow.Start(ctx); err != nil {
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

	if err := capability.StorageDeployPolicy.Start(ctx); err != nil {
		return err
	}

	if err := capability.StorageTenant.Start(ctx); err != nil {
		return err
	}

	if err := capability.Manager.Start(ctx); err != nil {
		return err
	}

	if err := capability.PeriodicTask.Start(ctx); err != nil {
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
