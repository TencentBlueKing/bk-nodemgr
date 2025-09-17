/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package manager

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/node"
)

// registerOperExecDefs init operation execution definitions.
func (mgr *Manager) registerOperExecDefs() error {
	if err := mgr.registerOperExecDefNodeInstall(); err != nil {
		return fmt.Errorf("register oper extra action def node install failed: %w", err)
	}

	return nil
}

// registerOperExecDefNodeInstall registers the operation execution definitions for node installation operations.
func (mgr *Manager) registerOperExecDefNodeInstall() error {
	return mgr.workflowMgr.RegisterOperExtraExecutions(
		node.NewOperationExtraExecution(mgr.conf.Cache, mgr.conf.StorageNode),
	)
}
