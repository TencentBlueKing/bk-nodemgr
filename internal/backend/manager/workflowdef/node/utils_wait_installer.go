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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
)

func saveWaitInstallerPrivateData(
	nCtx contextx.IContext,
	storage workflow.IStorageActionInstance,
	operationInstanceID string,
	pollingSwitch bool,
	ensureAgentID bool,
) error {

	if err := storage.UpsertActionInstancePrivateData(
		nCtx,
		operationInstanceID,
		ActionNameWaitInstallerComplete,
		map[string]any{
			types.PDKeyActionWaitInstallerCompletePollingSwitch: pollingSwitch,
			types.PDKeyActionWaitInstallerCompleteEnsureAgentID: ensureAgentID,
		},
	); err != nil {
		return fmt.Errorf("failed to save wait installer private data: %w", err)
	}

	return nil
}
