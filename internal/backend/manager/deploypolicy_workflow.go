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

package manager

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func (mgr *Manager) recordDeployPolicyWorkflowChild(
	nCtx contextx.IContext, workflowIDs []string, child types.DeployPolicyWorkflowChild,
) error {

	if len(workflowIDs) == 0 {
		return nil
	}

	if err := mgr.conf.StorageDeployPolicy.RecordDeployPolicyWorkflowChild(nCtx, workflowIDs, child); err != nil {
		return fmt.Errorf("failed to record deploy policy workflow child: %w", err)
	}

	return nil
}
