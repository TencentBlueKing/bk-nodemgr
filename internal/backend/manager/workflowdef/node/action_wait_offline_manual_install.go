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

package node

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	nodeUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/node/utils"
	nodeStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameWaitOfflineManualInstall defines the action name.
	ActionNameWaitOfflineManualInstall = "wait_offline_manual_install"

	waitOfflineManualInstallInterval = 1 * time.Second
	waitOfflineManualInstallTimeout  = 3 * time.Hour // nolint:mnd
)

// installerDataFileContent is the structure of installer.data.json submitted by the user.
// SYNC: must stay in sync with dataFileContent in tools/internal/installer/node/datareporter/step.go.
type installerDataFileContent struct {
	AgentID    string `json:"agent_id"`
	Token      string `json:"token"`
	OperInstID string `json:"oper_inst_id"`
}

// NewActionWaitOfflineManualInstall get a new action.
func NewActionWaitOfflineManualInstall(capability *Capability) action.Definition {
	return &actionWaitOfflineManualInstall{
		storageNodeDeployment: capability.StorageNode,
		storageHost:           capability.StorageTopo,
		storageActionInstance: capability.StorageWorkflow,
	}
}

// ActParamWaitOfflineManualInstall is the parameter for actionWaitOfflineManualInstall.
type ActParamWaitOfflineManualInstall struct {
	nodeUtils.NodeActionStandardParam `json:",inline"`
}

type actionWaitOfflineManualInstall struct {
	storageNodeDeployment nodeStg.IDaoNodeDeployment
	storageHost           topoStg.IStorageHost
	storageActionInstance workflow.IStorageActionInstance
}

// Name returns the name of the action.
func (act *actionWaitOfflineManualInstall) Name() string {
	return ActionNameWaitOfflineManualInstall
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionWaitOfflineManualInstall) DisplayNameZh() string {
	return "等待离线手动安装"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionWaitOfflineManualInstall) DisplayNameEn() string {
	return "Wait for Offline Manual Install"
}

// Version returns the version of the action.
func (act *actionWaitOfflineManualInstall) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionWaitOfflineManualInstall) Description() string {
	return "Block workflow and wait for user to manually execute the offline install script and submit the result data"
}

// Timeout returns the timeout of the action.
func (act *actionWaitOfflineManualInstall) Timeout() time.Duration {
	return waitOfflineManualInstallTimeout
}

// Tags returns the tags of the action.
func (act *actionWaitOfflineManualInstall) Tags() []action.Tag {
	return []action.Tag{action.TagNeedOfflineManualInstall}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionWaitOfflineManualInstall) MaxRetryCount() uint {
	return 0
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionWaitOfflineManualInstall) DelayFn(_ int) func() {
	return func() {}
}

// Do this func define what the action will do.
func (act *actionWaitOfflineManualInstall) Do(ctx *action.InstanceContext) error {
	param := new(ActParamWaitOfflineManualInstall)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	std := nodeUtils.NewNodeActionStandarder(act.storageNodeDeployment, act.storageHost)
	if err = std.Initialize(ctx, param.NodeActionStandardParam); err != nil {
		return err
	}
	defer func() {
		if storeErr := std.Save(); storeErr != nil {
			err = errors.Join(storeErr, err)
		}
	}()

	ticker := time.NewTicker(waitOfflineManualInstallInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Ctx.Done():
			return nil

		case <-ticker.C:
			resultData, fetchErr := act.tryFetchInstallResult(std)
			if fetchErr == errOfflineInstallResultNotReceived {
				continue
			}
			if fetchErr != nil {
				return fetchErr
			}

			if err = act.applyResult(std, resultData); err != nil {
				return err
			}
		}

		break
	}

	return nil
}

var errOfflineInstallResultNotReceived = errors.New("offline install result not received")

func (act *actionWaitOfflineManualInstall) tryFetchInstallResult(
	std *nodeUtils.NodeActionStandarder,
) (*installerDataFileContent, error) {

	privateData, err := act.storageActionInstance.GetActionInstancePrivateData(
		std.Context(),
		std.InstanceData().OperationInstanceID,
		ActionNameWaitOfflineManualInstall)
	if err != nil {
		logger.G.Sys().Ctx(std.Context()).WithErr(err).Error("failed to get action private data")

		return nil, err
	}

	rawValue, exists := privateData[types.PDKeyOfflineInstallResult]
	if !exists {
		logger.G.Sys().Ctx(std.Context()).
			With("oper-inst-id", std.InstanceData().OperationInstanceID, "key", types.PDKeyOfflineInstallResult).
			Debug("no offline install result received")

		return nil, errOfflineInstallResultNotReceived
	}

	rawStr, ok := rawValue.(string)
	if !ok {
		logger.G.Sys().Ctx(std.Context()).
			With("oper-inst-id", std.InstanceData().OperationInstanceID, "key", types.PDKeyOfflineInstallResult, "raw-value", rawValue).
			Error("unexpected type for fetched value")

		return nil, fmt.Errorf("got unexpected value from action private data. key(%s)", types.PDKeyOfflineInstallResult)
	}

	data := new(installerDataFileContent)
	if err = json.Unmarshal([]byte(rawStr), data); err != nil {
		return nil, fmt.Errorf("failed to parse offline install result JSON: %w", err)
	}

	return data, nil
}

func (act *actionWaitOfflineManualInstall) applyResult(
	std *nodeUtils.NodeActionStandarder,
	data *installerDataFileContent,
) error {

	if data.AgentID == "" {
		return fmt.Errorf("offline install result is missing agent_id field")
	}

	if data.OperInstID == "" {
		return fmt.Errorf("offline install result is missing oper_inst_id field")
	}

	if data.OperInstID != std.InstanceData().OperationInstanceID {
		return fmt.Errorf(
			"offline install result oper_inst_id(%s) does not match expected(%s)",
			data.OperInstID, std.InstanceData().OperationInstanceID,
		)
	}

	std.DeployInfo().Host.Dynamic.AgentID = data.AgentID
	std.ResetInstanceDataContext()
	std.InstanceData().Log().
		Zh("接收到离线安装结果, agent-id(%s)", data.AgentID).
		En("received offline install result, agent-id(%s)", data.AgentID).
		Info()

	return nil
}
