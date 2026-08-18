/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package node

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/deployconstant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func (act *actionRenderNodeDeployment) resolveNodeEventDataIDConf(
	nCtx contextx.IContext,
	deployInfo *types.DeploymentInfo,
) (deployconstant.NodeEventDataIDConf, error) {

	if act.storageBizEventDataIDConf == nil {
		return deployconstant.NodeEventDataIDConf{}, fmt.Errorf("biz event data-id conf storage is nil")
	}
	if act.monitorHandler == nil {
		return deployconstant.NodeEventDataIDConf{}, fmt.Errorf("monitor handler is nil")
	}

	bkBizID, err := resolveDeploymentBizID(deployInfo)
	if err != nil {
		return deployconstant.NodeEventDataIDConf{}, err
	}

	bizConf, bizConfFound, err := act.storageBizEventDataIDConf.GetBizEventDataIDConf(nCtx, bkBizID)
	if err != nil {
		return deployconstant.NodeEventDataIDConf{}, fmt.Errorf("failed to get biz event data-id conf: %w", err)
	}

	eventDataIDConf := deployconstant.NodeEventDataIDConf{
		AgentBaseAlarmEventDataID: act.resolveAgentBaseAlarmEventDataID(bizConf, bizConfFound),
		TaskProcEventDataID:       act.defaultNodeEventDataIDConf.TaskProcEventDataID,
	}

	if bizConfFound && bizConf.TaskProcEventDataID != nil {
		eventDataIDConf.TaskProcEventDataID = *bizConf.TaskProcEventDataID
		return validateNodeEventDataIDConf(eventDataIDConf)
	}

	taskProcEventDataID, found, err := act.monitorHandler.GetOrCreateAgentEventDataID(nCtx, bkBizID)
	if err != nil {
		return deployconstant.NodeEventDataIDConf{}, fmt.Errorf("failed to get or create agent event data-id: %w", err)
	}
	if !found {
		return validateNodeEventDataIDConf(eventDataIDConf)
	}

	if err := act.storageBizEventDataIDConf.UpdateBizTaskProcEventDataID(nCtx, bkBizID, taskProcEventDataID); err != nil {
		return deployconstant.NodeEventDataIDConf{}, fmt.Errorf("failed to write back task process event data-id: %w", err)
	}

	eventDataIDConf.TaskProcEventDataID = taskProcEventDataID

	return validateNodeEventDataIDConf(eventDataIDConf)
}

func resolveDeploymentBizID(deployInfo *types.DeploymentInfo) (int64, error) {
	if deployInfo == nil || deployInfo.Host.Static == nil {
		return 0, fmt.Errorf("deployment biz id is missing")
	}

	if deployInfo.Host.Static.BizID <= 0 {
		return 0, fmt.Errorf("deployment biz id must be positive, got %d", deployInfo.Host.Static.BizID)
	}

	return deployInfo.Host.Static.BizID, nil
}

func (act *actionRenderNodeDeployment) resolveAgentBaseAlarmEventDataID(
	bizConf types.BizEventDataIDConf,
	bizConfFound bool,
) int64 {

	if bizConfFound && bizConf.AgentBaseAlarmEventDataID != nil {
		return *bizConf.AgentBaseAlarmEventDataID
	}

	return act.defaultNodeEventDataIDConf.AgentBaseAlarmEventDataID
}

func validateNodeEventDataIDConf(
	conf deployconstant.NodeEventDataIDConf,
) (deployconstant.NodeEventDataIDConf, error) {

	if err := conf.Validate(); err != nil {
		return deployconstant.NodeEventDataIDConf{}, fmt.Errorf("invalid node event data-id conf: %w", err)
	}

	return conf, nil
}
