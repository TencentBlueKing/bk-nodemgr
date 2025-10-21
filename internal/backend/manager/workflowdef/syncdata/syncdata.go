/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package syncdata this package define the operation witch is used to sync data.
package syncdata

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/configpolicy"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/credit"
	nodeStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/release"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/tenant"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	wfStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/discover"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/cache"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/cmdb"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/file"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/usermanager"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
)

// HostIDAgentID defines the host ID and agent ID.
type HostIDAgentID struct {
	HostID  int64  `json:"host_id"`
	AgentID string `json:"agent_id"`
}

// Capability encapsulates the various capabilities the service supports.
type Capability struct {
	// thridparty handler.
	CMDBHandler        cmdb.IHandler
	GSEHandler         gse.IHandler
	FileHandler        file.IHandler
	UserManagerHandler usermanager.IHandler

	// stroage.
	StorageTopo         topoStg.IStorage
	StorageRelease      release.IStorage
	StorageNode         nodeStg.IStorage
	StorageWorkflow     wfStg.IStorage
	StoragePlugin       plugin.IStorage
	StorageHostCredit   credit.IStorageHostCredit
	StorageConfigPolicy configpolicy.IStorage
	StorageTenant       tenant.IStorage

	// discover provider.
	DiscoverProvider discover.Provider

	// cache
	Cache cache.ICache

	// workflow manager.
	WorkflowCtl workflow.IManager
}
