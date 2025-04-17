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

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operengine"
	"time"
)

const (
	// OperDefNameInstallNodeBySSH the name of the operation definition.
	OperDefNameInstallNodeBySSH = "install_node_by_ssh"
)

// OperInstallNodeBySSH the params of OperInstallNodeBySSH.
type OperInstallNodeBySSH struct {
	Token string `json:"token"`
}

// OperDef the operdef of OperInstallNodeBySSH.
func (oper *OperInstallNodeBySSH) OperDef() operengine.OperDefSnapshot {
	return operengine.OperDefSnapshot{
		OperDefName: OperDefNameInstallNodeBySSH,
		ActionNames: []string{
			ActionNameUpsertHost,
			ActionNameRenderNodeDeployment,
			ActionNameInstallNodeBySSH,
			ActionNameWaitComplete,
			ActionNameWaitGseReady,
			ActionNameSyncNodeInfo,
			ActionNameBindAgentHostRel,
			ActionNamePushHostIdentifier,
			ActionNameUpdateHost,
		},
	}
}

// Param the param of OperInstallNodeBySSH.
func (oper *OperInstallNodeBySSH) Param() operengine.OperInstParam {
	return operengine.OperInstParam{
		Timeout:     time.Minute * 10, //nolint: mnd
		InitContent: conv.StructToMapIgnoreError(oper),
	}
}
