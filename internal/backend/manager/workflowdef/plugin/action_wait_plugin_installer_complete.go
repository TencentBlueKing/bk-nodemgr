/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package plugin

import (
	"errors"
	"fmt"
	"time"

	pluginUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/plugin/utils"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/installer"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameWaitPluginInstallerComplete defines the action name.
	ActionNameWaitPluginInstallerComplete = "wait_plugin_installer_complete"

	waitInstallerCompleteInterval = 1 * time.Second
)

// NewActionWaitPluginInstallerComplete get a new action.
func NewActionWaitPluginInstallerComplete(capability *Capability) action.Definition {
	return &actionWaitPluginInstallerComplete{
		daoPluginDeployment:   capability.StoragePlugin,
		storageActionInstance: capability.StorageWorkflow,
	}
}

// ActionWaitPluginInstallerComplete defines the action param.
type ActionWaitPluginInstallerComplete struct {
	pluginUtils.PluginActionStandardParam `json:",inline"`
}

type actionWaitPluginInstallerComplete struct {
	daoPluginDeployment   pluginStg.IDaoPluginDeployment
	storageActionInstance workflow.IStorageActionInstance
}

// Name returns the name of the action.
func (act *actionWaitPluginInstallerComplete) Name() string {
	return ActionNameWaitPluginInstallerComplete
}

// Version returns the version of the action.
func (act *actionWaitPluginInstallerComplete) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionWaitPluginInstallerComplete) Description() string {
	return "Wait for the installer command to complete"
}

// Timeout returns the timeout of the action.
func (act *actionWaitPluginInstallerComplete) Timeout() time.Duration {
	return 30 * time.Minute // nolint:mnd
}

// Tags returns the tags of the action.
func (act *actionWaitPluginInstallerComplete) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount this func define how many times this action will retry.
func (act *actionWaitPluginInstallerComplete) MaxRetryCount() uint {
	return 0
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionWaitPluginInstallerComplete) DelayFn() func() {
	return func() {}
}

// Do this func define what the action will do.
func (act *actionWaitPluginInstallerComplete) Do(ctx *action.InstanceContext) error {
	param := new(ActionParamTrusteeshipProcess)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	// initialize standard data.
	std := pluginUtils.NewPluginActionStandarder(act.daoPluginDeployment)
	if err = std.Initialize(ctx, param.PluginActionStandardParam); err != nil {
		return err
	}
	defer func() {
		if storeErr := std.Save(); storeErr != nil {
			err = errors.Join(storeErr, err)
		}
	}()

	instanceID := std.InstanceData().OperationInstanceID

	// wait for installer result
	rawInstallerResult, err := act.waitInstallerField(std, types.PDKeyInstallerReportStatus)
	if err != nil {
		return fmt.Errorf("failed to wait for installer result: %w", err)
	}

	installerResult := installer.ProcessState(rawInstallerResult)

	// check and update action state
	switch installerResult {
	case installer.ProcessStateSuccess:
		std.InstanceData().LogI("received installer result is success")

		return nil

	case installer.ProcessStateFailed, installer.ProcessStateTimeout:
		std.InstanceData().LogI(fmt.Sprintf("received installer result is not success. installer-result(%s)", installerResult))

		return fmt.Errorf("installer failed. oper-inst-id(%s), action-name(%s), installer-result(%s)",
			instanceID, ActionNameWaitPluginInstallerComplete, installerResult)

	default:
		logger.G.Sys().With("oper-inst-id", instanceID, "state", installerResult).Warn("installer state is not supported")

		return fmt.Errorf("unexpected installer state. state(%s)", installerResult)
	}
}

func (act *actionWaitPluginInstallerComplete) waitInstallerField(std *pluginUtils.PluginActionStandarder, key string) (string, error) {
	ticker := time.NewTicker(waitInstallerCompleteInterval)
	defer ticker.Stop()

	for {
		select {
		case <-std.Context().Done():

			return "", std.Context().Err()
		case <-ticker.C:
			value, err := act.tryFetchValue(std, key)
			if err != nil {
				return "", err
			}

			if value != "" {
				return value, nil
			}
		}
	}
}

func (act *actionWaitPluginInstallerComplete) tryFetchValue(std *pluginUtils.PluginActionStandarder, key string) (string, error) {
	privateData, err := act.storageActionInstance.GetActionInstancePrivateData(
		std.Context(),
		std.InstanceData().OperationInstanceID,
		ActionNameWaitPluginInstallerComplete)
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to get action private data")

		return "", err
	}

	rawValue, exists := privateData[key]
	if !exists {
		logger.G.Sys().With("oper-inst-id", std.InstanceData().OperationInstanceID, "key", key).Debug("no receive data")

		return "", nil
	}

	value, err := conv.ToString(rawValue)
	if err != nil {
		logger.G.Sys().With("oper-inst-id", std.InstanceData().OperationInstanceID, "key", key, "raw-value", rawValue).
			Error("unexpected type for fetched value")

		return "", fmt.Errorf("failed to get value for key. key(%s)", key)
	}

	return value, nil
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionWaitPluginInstallerComplete) DisplayNameZh() string { return act.Name() }

// DisplayNameEn returns the English display name of the action.
func (act *actionWaitPluginInstallerComplete) DisplayNameEn() string { return act.Name() }
