/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package nodeinstall

import "github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operengine"

const (
	// OperDefNameInstallNodeBySSH the name of the operation definition.
	OperDefNameInstallNodeBySSH = "install_node_by_ssh"
)

// NewOperationInstallNodeBySSH new an operation to install node.
func NewOperationInstallNodeBySSH(triggerID string) *operengine.Operation {
	defSnapshot := operengine.OperDefSnapshot{
		OperDefName: OperDefNameInstallNodeBySSH,
		ActionNames: []string{
			ActionNameUpsertHost,
			ActionNameRenderNodeDeployment,
			ActionNameInstallAgentBySSH,
			ActionNameWaitComplete,
			ActionNameWaitGseReady,
			ActionNameSyncNodeInfo,
			ActionNameBindAgentHostRel,
			ActionNamePushHostIdentifier,
			ActionNameUpdateHost,
		},
	}

	operation := operengine.NewOperation(triggerID, defSnapshot)

	return operation
}
