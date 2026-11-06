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

package plugin

import (
	"fmt"
	"time"

	pluginUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/plugin/utils"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/retrier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

func saveAndWaitCompositionChild(act action.Definition, storagePlugin pluginStg.IStorage,
	storageAction workflow.IStorageActionInstance, std *pluginUtils.PluginActionStandarder, workflowID string) error {

	// Empty workflow ID means no source operation was requested.
	if workflowID == "" {
		return nil
	}
	refs := []types.SubWorkflowRef{{WorkflowID: workflowID, WorkflowDomain: types.WorkflowDomainPlugin}}
	serialized, err := types.SerializeSubWorkflowRefs(refs)
	if err != nil {
		return fmt.Errorf("failed to serialize composition child references: %w", err)
	}
	std.InstanceData().PrivateData[types.PDKeySubWorkflowRefs] = serialized
	if err := storageAction.UpsertActionInstancePrivateData(std.Context(), std.InstanceData().OperationInstanceID,
		act.Name(), map[string]any{types.PDKeySubWorkflowRefs: serialized}); err != nil {
		return fmt.Errorf("failed to save composition child references: %w", err)
	}
	for _, ref := range refs {
		if err := waitCompositionChild(storagePlugin, std, ref.WorkflowID, act.Timeout()); err != nil {
			return err
		}
	}

	return nil
}

func waitCompositionChild(storagePlugin pluginStg.IStorage, std *pluginUtils.PluginActionStandarder,
	workflowID string, timeout time.Duration) error {

	var status types.PluginWorkflowStatus
	polling := retrier.NewPolling(retrier.PollingOpts{Timeout: timeout, Interval: 10 * time.Second}) // nolint: mnd
	if err := polling.Do(std.Context(), func(_ int) error {
		var err error
		status, err = storagePlugin.GetPluginWorkflowStatus(std.Context(), workflowID)
		if err != nil {
			return fmt.Errorf("failed to get child workflow status: %w", err)
		}
		if status == types.PluginWorkflowStatusRunning {
			return fmt.Errorf("child workflow is still running: %s", workflowID)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("failed to wait for child workflow(%s): %w", workflowID, err)
	}
	if status != types.PluginWorkflowStatusSuccess {
		return fmt.Errorf("child workflow(%s) ended with status(%s)", workflowID, status)
	}

	return nil
}
