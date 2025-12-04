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
	"errors"
	"fmt"
	"time"

	nodeUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/node/utils"
	nodeStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/installer"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameWaitInstallerComplete defines the action name.
	ActionNameWaitInstallerComplete = "wait_node_installer_complete"

	waitReportInterval = 1 * time.Second
)

// NewActionWaitInstallerComplete get a new action.
func NewActionWaitInstallerComplete(capability *Capability) action.Definition {
	return &actionWaitInstallerComplete{
		storageNodeDeployment: capability.StorageNode,
		storageActionInstance: capability.StorageWorkflow,
	}
}

// ActionWaitInstallerComplete defines the action param.
type ActionWaitInstallerComplete struct {
	nodeUtils.NodeActionStandardParam `json:",inline"`
	EnsureAgentID                     bool `json:"ensure_agent_id"`
}

type actionWaitInstallerComplete struct {
	storageNodeDeployment nodeStg.IDaoNodeDeployment
	storageActionInstance workflow.IStorageActionInstance
}

// Name returns the name of the action.
func (act *actionWaitInstallerComplete) Name() string {
	return ActionNameWaitInstallerComplete
}

// Version returns the version of the action.
func (act *actionWaitInstallerComplete) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionWaitInstallerComplete) Description() string {
	return "Wait for the installer command to complete"
}

// Timeout returns the timeout of the action.
func (act *actionWaitInstallerComplete) Timeout() time.Duration {
	return 30 * time.Minute // nolint:mnd
}

// Tags returns the tags of the action.
func (act *actionWaitInstallerComplete) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount this func define how many times this action will retry.
func (act *actionWaitInstallerComplete) MaxRetryCount() uint {
	return 0
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionWaitInstallerComplete) DelayFn() func() {
	return func() {}
}

// Do this func define what the action will do.
// nolint: gocognit
func (act *actionWaitInstallerComplete) Do(ctx *action.InstanceContext) error {
	param := new(ActionWaitInstallerComplete)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	// initialize standard data.
	std := nodeUtils.NewNodeActionStandarder(act.storageNodeDeployment)
	if err = std.Initialize(ctx, param.NodeActionStandardParam); err != nil {
		return err
	}
	defer func() {
		if storeErr := std.Save(); storeErr != nil {
			err = errors.Join(storeErr, err)
		}
	}()

	instanceID := std.InstanceData().OperationInstanceID

	ticker := time.NewTicker(waitReportInterval)
	defer ticker.Stop()

	var rawInstallerResult string

	for {
		select {
		case <-ctx.Ctx.Done():
			return nil

		case <-ticker.C:
			// if agent-id set, fetch agent id
			if param.EnsureAgentID {
				agentID, err := act.tryFetchValue(std, instanceID, installer.WaitInstallerCompleteReportAgentIDKey)
				if err != nil {
					return err
				}

				if agentID == "" {
					continue
				}
				std.DeployInfo().Host.Dynamic.AgentID = agentID
			}

			// fetch installer result
			rawInstallerResult, err = act.tryFetchValue(std, instanceID, installer.WaitInstallerCompleteReportStatusKey)
			if err != nil {
				return err
			}

			if rawInstallerResult == "" {
				continue
			}
		}

		break
	}

	// check and update action state
	installerResult := installer.ProcessState(rawInstallerResult)
	switch installerResult {
	case installer.ProcessStateSuccess:
		// reset action context
		std.ResetInstanceDataContext()
		std.InstanceData().LogI("received installer result is success")

		return nil

	case installer.ProcessStateFailed, installer.ProcessStateTimeout:
		std.InstanceData().LogI(fmt.Sprintf("received installer result is not success. installer-result(%s)", installerResult))

		return fmt.Errorf("installer failed. oper-inst-id(%s), action-name(%s), installer-result(%s)",
			instanceID, ActionNameWaitInstallerComplete, installerResult)

	default:
		logger.G.Sys().With("oper-inst-id", instanceID, "state", installerResult).Warn("installer state is not supported")

		return fmt.Errorf("unexpected installer state. state(%s)", installerResult)
	}
}

func (act *actionWaitInstallerComplete) tryFetchValue(std *nodeUtils.NodeActionStandarder, instanceID, key string) (string, error) {
	privateData, err := act.storageActionInstance.GetActionInstancePrivateData(
		std.Context(),
		instanceID,
		ActionNameWaitInstallerComplete)
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to get action private data")

		return "", err
	}

	rawValue, exists := privateData[key]
	if !exists {
		logger.G.Sys().With("oper-inst-id", instanceID, "key", key).Debug("no receive data, sleep 1 second")

		return "", nil
	}

	value, err := conv.ToString(rawValue)
	if err != nil {
		logger.G.Sys().With("oper-inst-id", instanceID, "key", key, "rawValue", rawValue).Error("unexpected type for fetched value")

		return "", fmt.Errorf("failed to get value for key. key(%s)", key)
	}

	return value, nil
}
