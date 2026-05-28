/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package utils

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

// FinishOperationInstance finish operation.
func FinishOperationInstance(std *PluginActionStandarder,
	daoOperationInstance workflow.IStorageOperationInstance,
	daoActionInstance workflow.IStorageActionInstance,
) error {
	operInstanceID := std.InstanceData().OperationInstanceID
	operationInstance, err := daoOperationInstance.GetOperationInstanceFullData(std.Context(), operInstanceID)
	if err != nil {
		std.InstanceData().Log().
			Zh("获取 Operation Instance 失败, operation-instance-id(%s), err(%v)", operInstanceID, err).
			En("failed to get operation instance, operation-instance-id(%s), err(%v)", operInstanceID, err).
			Error()

		return fmt.Errorf("get operation instance, operation-instance-id(%s): %w", operInstanceID, err)
	}

	for actionName, actionInstData := range operationInstance.ActionInstanceDataMap {
		if actionInstData.Index <= std.InstanceData().Index {
			continue
		}

		// mark action as skipped
		actionInstData.Lifecycle.State = action.StateSkipped

		err := daoActionInstance.UpdateOperInstActionStatus(std.Context(), operInstanceID, actionName, action.StateSkipped)
		if err != nil {
			std.InstanceData().Log().
				Zh("更新 Action Instance 失败, operation-instance-id(%s), action-name(%s), err(%v)",
					operInstanceID, actionName, err).
				En("failed to update action instance, operation-instance-id(%s), action-name(%s), err(%v)",
					operInstanceID, actionName, err).
				Error()

			return fmt.Errorf("update action instance, operation-instance-id(%s), action-name(%s): %w",
				operInstanceID, actionName, err)
		}
	}

	std.InstanceData().Log().
		Zh("operation instance(%s) 在 action(%s) 阶段已结束，后续流程将被跳过",
			std.InstanceData().OperationInstanceID, std.InstanceData().Name).
		En("operation instance(%s) finished at action(%s), skipping subsequent processes",
			std.InstanceData().OperationInstanceID, std.InstanceData().Name).
		Info()

	return nil
}
