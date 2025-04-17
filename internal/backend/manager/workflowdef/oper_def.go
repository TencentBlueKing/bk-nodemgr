/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package workflowdef ...
package workflowdef

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/syncdata"
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/nodeinstall"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operengine"
)

// OperDefName operation definition name.
type OperDefName string

// OperBuilder used to create an operation.
type OperBuilder func(triggerID string) *operengine.Operation

// operBuilderRegistry ...
type operBuilderRegistry map[OperDefName]OperBuilder

var instance = struct {
	registry operBuilderRegistry
	once     sync.Once
}{
	registry: make(map[OperDefName]OperBuilder),
	once:     sync.Once{},
}

// OperBuilderRegistry ...
func OperBuilderRegistry() map[OperDefName]OperBuilder {
	instance.once.Do(func() {
		instance.registry = map[OperDefName]OperBuilder{
			syncdata.OperDefNameSyncHost:               syncdata.NewOperSyncHostFromCMDB,
			syncdata.OperDefNameSyncBiz:                syncdata.NewOperSyncBizFromCMDB,
			syncdata.OperDefNameSyncBizAndHostFromCMDB: syncdata.NewOperSyncBizAndHostFromCMDB,
			syncdata.OperDefNameSyncNetworkArea:        syncdata.NewOperSyncNetworkAreaFromCMDB,
			nodeinstall.OperDefNameInstallNodeBySSH:    nodeinstall.NewOperationInstallNodeBySSH,
		}
	})

	return instance.registry
}
