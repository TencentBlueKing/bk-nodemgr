/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package manager ...
package manager

import (
	"errors"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/configpolicy"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/credit"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/deploypolicy"
	nodeStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/release"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/tenant"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/creditvault"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/discover"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/filecache"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/relayhandler"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/cache"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/locker"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/cmdb"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/file"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
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

	// discover provider.
	Provider discover.Provider

	// file group.
	InstallerFileGroup fileiface.FileGroup

	// FileCache is the local artifact file cache.
	FileCache filecache.IFileCache

	// workflow locker.
	LockerFactory locker.MutexFactory

	// storages.
	StorageTopo         topoStg.IStorage
	StorageRelease      release.IStorage
	StorageNode         nodeStg.IStorage
	StorageWorkflow     workflow.IStorage
	StoragePlugin       pluginStg.IStorage
	StorageHostCredit   credit.IStorageHostCredit
	StorageConfigPolicy configpolicy.IStorage
	StorageTenant       tenant.IStorage
	StorageDeployPolicy deploypolicy.IStorage

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

	return nil
}
