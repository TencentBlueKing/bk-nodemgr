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

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/configpolicy"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/credit"
	nodeStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/release"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/creditvault"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/relayhandler"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/cache"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/discover"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/locker"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/cmdb"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/file"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
)

// Config defines the config of Manager.
type Config struct {
	CmdbHandler cmdb.IHandler
	GSEHandler  gse.IHandler
	Provider    discover.Provider

	InstallerFileGroup fileiface.FileGroup

	LockerFactory locker.MutexFactory

	StorageTopo         topo.IStorage
	StorageRelease      release.IStorage
	StorageNode         nodeStg.IStorage
	StorageWorkflow     workflow.IStorage
	StorageHostCredit   credit.IStorageHostCredit
	StorageConfigPolicy configpolicy.IStorage

	HostPasswordVault creditvault.IHostPasswordVault

	FileHandler   file.IHandler
	ProxyMessager relayhandler.IServerMessager
	Cache         cache.ICache
	WorkflowConfig
}

// WorkflowConfig defines the workflow config.
type WorkflowConfig struct {
	WorkNodeNum int
	Redis       RedisConfig
}

// Validate configures the config.
func (conf *WorkflowConfig) Validate() error {
	if conf.WorkNodeNum <= 0 {
		return errors.New("work node num is invalid")
	}

	return nil
}

// RedisConfig defines the redis config.
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// Validate configures the config.
func (conf *RedisConfig) Validate() error {
	if conf.Addr == "" {
		return errors.New("redis addr is invalid")
	}
	if conf.Password == "" {
		return errors.New("redis password is invalid")
	}
	if conf.DB < 0 {
		return errors.New("redis db is invalid")
	}

	return nil
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
