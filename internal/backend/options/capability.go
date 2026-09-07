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

// Package options provides the various capabilities the service supports.
package options

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth/provider"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/periodictask"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/bizeventdataidconf"
	cipherStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/cipher"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/configpolicy"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/credit"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/deploypolicy"
	globalsettingsStorage "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/globalsettings"
	nodeStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node"
	pkgStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/pkg"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/release"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/tenant"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/creditvault"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/discover"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/filecache"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/relayhandler"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/crypter"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/locker"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/cmdb"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/file"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/iamv3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/iamv4"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/monitor"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/usermanager"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tracing"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

// Capability encapsulates the various capabilities the service supports.
type Capability struct {
	// Manager workflow management.
	Manager manager.IManager

	// RedisClient redis client (supports both standalone and cluster mode).
	RedisClient redis.UniversalClient

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

	// StoragePackage package storage.
	StoragePackage pkgStg.IStorage

	// StorageRelease release storage.
	StorageRelease release.IStorage

	// StorageGlobalSettings global settings storage.
	StorageGlobalSettings globalsettingsStorage.IStorage

	// StorageBizEventDataIDConf business event data-id config storage.
	StorageBizEventDataIDConf bizeventdataidconf.IStorage

	// StorageCredit credit storage.
	StorageCredit credit.IStorage

	// StorageConfigPolicy config policy storage.
	StorageConfigPolicy configpolicy.IStorage

	// StorageDeployPolicy deploy policy storage.
	StorageDeployPolicy deploypolicy.IStorage

	// StorageTenant tenant storage.
	StorageTenant tenant.IStorage

	// StorageCipher cipher storage.
	StorageCipher cipherStg.IStorage

	// CmdbHandler cmdb handler.
	CmdbHandler cmdb.IHandler

	// GSEHandler gse handler.
	GSEHandler gse.IHandler

	// FileHandler file handler.
	FileHandler file.IHandler

	// UserManagerHandler user manager handler.
	UserManagerHandler usermanager.IHandler

	// IAMV3Handler the IAM v3 handler.
	IAMV3Handler iamv3.IHandler
	// IAMV4Handler the IAM v4 handler.
	IAMV4Handler iamv4.IHandler

	// MonitorHandler the monitor handler.
	MonitorHandler monitor.IHandler

	// Authorizer is the IAM authorization handler for permission checks.
	Authorizer auth.IAuthorizer

	// AuthProviderHandler is the unified handler for auth resource provider handler.
	AuthProviderHandler provider.IHandler

	// LockerFactory locker factory
	LockerFactory locker.MutexFactory

	// Crypter ...
	Crypter crypter.Crypter

	// InstallerFileGroup tool file group.
	InstallerFileGroup fileiface.FileGroup

	// FileCache is the local artifact file cache for SSH install flows.
	FileCache filecache.IFileCache

	// Discover provides discover handler.
	DiscoverProvider discover.IProvider

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

	if err := capability.StoragePackage.Start(ctx); err != nil {
		return err
	}

	if err := capability.StorageRelease.Start(ctx); err != nil {
		return err
	}

	if err := capability.StorageGlobalSettings.Start(ctx); err != nil {
		return err
	}

	if err := capability.StorageBizEventDataIDConf.Start(ctx); err != nil {
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

	if err := capability.StorageCipher.Start(ctx); err != nil {
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

	if capability.FileCache != nil {
		if err := capability.FileCache.Close(); err != nil {
			return err
		}
	}

	if err := tracing.G().ShutdownAll(contextx.Background()); err != nil {
		return err
	}

	return nil
}
