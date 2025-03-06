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
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operengine"
)

// OperDefName operation definition name.
type OperDefName string

const (
	// OperDefNameSyncHost sync specific biz's host from cmdb.
	OperDefNameSyncHost = "oper_def_sync_host"

	// OperDefNameSyncBiz sync all biz from cmdb.
	OperDefNameSyncBiz = "oper_def_sync_biz"

	// OperDefNameSyncBizAndHost sync all biz and their host from cmdb.
	OperDefNameSyncBizAndHost = "oper_def_sync_biz_and_host"

	// OperDefNameSyncNetworkArea sync networkarea from cmdb.
	OperDefNameSyncNetworkArea = "oper_def_sync_networkarea"
)

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
			OperDefNameSyncHost:        newOperSyncHostFromCMDB,
			OperDefNameSyncBiz:         newOperSyncBizFromCMDB,
			OperDefNameSyncBizAndHost:  newOperSyncBizAndHostFromCMDB,
			OperDefNameSyncNetworkArea: newOperSyncNetworkAreaFromCMDB,
		}
	})
	return instance.registry
}

// newOperSyncBizFromCMDB new an operation to sync all biz's host from cmdb.
func newOperSyncBizFromCMDB(triggerID string) *operengine.Operation {
	defSnapshot := operengine.OperDefSnapshot{
		OperDefName: OperDefNameSyncBiz,
		ActionNames: []string{SyncBizFromCMDB},
	}

	operation := operengine.NewOperation(triggerID, defSnapshot)

	return operation
}

// newOperSyncHostFromCMDB new an operation to sync all biz's host from cmdb.
func newOperSyncHostFromCMDB(triggerID string) *operengine.Operation {
	defSnapshot := operengine.OperDefSnapshot{
		OperDefName: OperDefNameSyncHost,
		ActionNames: []string{SyncHostFromCMDB},
	}

	operation := operengine.NewOperation(triggerID, defSnapshot)

	return operation
}

// newOperSyncBizAndHostFromCMDB new an operation to sync all bizs and their host from cmdb.
func newOperSyncBizAndHostFromCMDB(triggerID string) *operengine.Operation {
	defSnapshot := operengine.OperDefSnapshot{
		OperDefName: OperDefNameSyncHost,
		ActionNames: []string{SyncBizFromCMDB, GenAllBizHostSyncOper},
	}

	operation := operengine.NewOperation(triggerID, defSnapshot)

	return operation
}

// newOperSyncNetworkAreaFromCMDB new an operation to sync all networkareas from cmdb.
func newOperSyncNetworkAreaFromCMDB(triggerID string) *operengine.Operation {
	defSnapshot := operengine.OperDefSnapshot{
		OperDefName: OperDefNameSyncNetworkArea,
		ActionNames: []string{SyncNetworkAreaFromCMDB},
	}

	operation := operengine.NewOperation(triggerID, defSnapshot)

	return operation
}
