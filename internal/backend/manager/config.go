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

// Package manager ...
package manager

import (
	"errors"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/bizeventdataidconf"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/configpolicy"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/credit"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/deploypolicy"
	nodeStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node"
	pkgStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/pkg"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/release"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/tenant"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/creditvault"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/deployconstant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/discover"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/filecache"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/relayhandler"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/cache"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/locker"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/cmdb"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/file"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/monitor"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/usermanager"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tracing"
)

// Config defines the config of Manager.
type Config struct {
	// thridparty handler.
	CmdbHandler        cmdb.IHandler
	GSEHandler         gse.IHandler
	FileHandler        file.IHandler
	UserManagerHandler usermanager.IHandler
	MonitorHandler     monitor.IHandler

	// discover provider.
	Provider discover.IProvider

	// file group.
	InstallerFileGroup fileiface.FileGroup

	// FileCache is the local artifact file cache.
	FileCache filecache.IFileCache

	// workflow locker.
	LockerFactory locker.MutexFactory

	// storages.
	StorageTopo               topoStg.IStorage
	StorageRelease            release.IStorage
	StorageNode               nodeStg.IStorage
	StorageWorkflow           workflow.IStorage
	StoragePackage            pkgStg.IStorage
	StoragePlugin             pluginStg.IStorage
	StorageHostCredit         credit.IStorageHostCredit
	StorageConfigPolicy       configpolicy.IStorage
	StorageTenant             tenant.IStorage
	StorageDeployPolicy       deploypolicy.IStorage
	StorageBizEventDataIDConf bizeventdataidconf.IStorage

	// node event data-id default config.
	NodeEventDataIDConf deployconstant.NodeEventDataIDConf

	// network unit config.
	NetworkUnitConfig config.NetworkUnit

	// credit vault.
	HostPasswordVault creditvault.IHostPasswordVault

	// relay handler.
	ProxyMessager relayhandler.IServerMessager

	// distributed cache.
	Cache cache.ICache

	//  distributed tracing
	TraceService tracing.IService

	WorkflowConfig
}

// WorkflowConfig defines the workflow config.
type WorkflowConfig struct {
	WorkNodeNum             int
	GracefulShutdownTimeout time.Duration
	Redis                   config.Redis
}

// Validate configures the config.
func (conf *WorkflowConfig) Validate() error {
	if conf.WorkNodeNum <= 0 {
		return errors.New("work node num is invalid")
	}

	if conf.GracefulShutdownTimeout <= 0 {
		return errors.New("graceful shutdown timeout is invalid")
	}

	return conf.Redis.Validate()
}

// Validate configures the config.
func (conf *Config) Validate() error {
	if conf.CmdbHandler == nil {
		return errors.New("cmdb handler is nil")
	}
	if conf.StorageTopo == nil {
		return errors.New("topo storage is nil")
	}
	if conf.StorageRelease == nil {
		return errors.New("release storage is nil")
	}
	if conf.StorageNode == nil {
		return errors.New("node storage is nil")
	}
	if conf.StorageWorkflow == nil {
		return errors.New("workflow storage is nil")
	}
	if conf.StoragePackage == nil {
		return errors.New("package storage is nil")
	}
	if conf.StoragePlugin == nil {
		return errors.New("plugin storage is nil")
	}
	if conf.StorageHostCredit == nil {
		return errors.New("host credit storage is nil")
	}
	if conf.StorageConfigPolicy == nil {
		return errors.New("config policy storage is nil")
	}
	if conf.StorageTenant == nil {
		return errors.New("tenant storage is nil")
	}
	if conf.StorageDeployPolicy == nil {
		return errors.New("deploy policy storage is nil")
	}
	if conf.StorageBizEventDataIDConf == nil {
		return errors.New("biz event data-id conf storage is nil")
	}

	return nil
}
