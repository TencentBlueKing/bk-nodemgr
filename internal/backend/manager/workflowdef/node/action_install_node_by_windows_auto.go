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
	"errors"
	"fmt"
	"time"

	nodeUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/node/utils"
	nodeStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameInstallNodeByWindowsAuto defines the action name.
	ActionNameInstallNodeByWindowsAuto = "install_node_by_windows_auto"
)

// NewActionInstallNodeByWindowsAuto get a new action.
func NewActionInstallNodeByWindowsAuto(capability *Capability) action.Definition {
	return &actionInstallNodeByWindowsAuto{
		ssh:                   NewActionInstallNodeByWindowsSSH(capability).(*actionInstallNodeByWindowsSSH),
		wmi:                   NewActionInstallNodeByWMI(capability).(*actionInstallNodeByWMI),
		storageNodeDeployment: capability.StorageNode,
		storageHost:           capability.StorageTopo,
		storageActionInstance: capability.StorageWorkflow,
	}
}

// ActParamInstallAgentByWindowsAuto defines the action parameter.
type ActParamInstallAgentByWindowsAuto struct {
	nodeUtils.NodeActionStandardParam `json:",inline"`
}

type actionInstallNodeByWindowsAuto struct {
	ssh *actionInstallNodeByWindowsSSH
	wmi *actionInstallNodeByWMI

	storageNodeDeployment nodeStg.IDaoNodeDeployment
	storageHost           topoStg.IStorageHost
	storageActionInstance workflow.IStorageActionInstance
}

// Name returns the name of the action.
func (act *actionInstallNodeByWindowsAuto) Name() string {
	return ActionNameInstallNodeByWindowsAuto
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionInstallNodeByWindowsAuto) DisplayNameZh() string {
	return "自动安装 Windows 节点"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionInstallNodeByWindowsAuto) DisplayNameEn() string {
	return "Install Windows Node Automatically"
}

// Version returns the version of the action.
func (act *actionInstallNodeByWindowsAuto) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionInstallNodeByWindowsAuto) Description() string {
	return "Use Windows SSH first to install the target machine, and fallback to WMI before installer launch."
}

// Timeout returns the timeout of the action.
func (act *actionInstallNodeByWindowsAuto) Timeout() time.Duration {
	return 5 * time.Minute // nolint: mnd
}

// Tags returns the tags of the action.
func (act *actionInstallNodeByWindowsAuto) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionInstallNodeByWindowsAuto) MaxRetryCount() uint {
	return 0
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionInstallNodeByWindowsAuto) DelayFn(_ int) func() {
	return func() {
		time.Sleep(5 * time.Second) // nolint: mnd
	}
}

// Do this func define what the action will do.
// nolint: perfsprint,funlen
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (act *actionInstallNodeByWindowsAuto) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActParamInstallAgentByWindowsAuto)
	err = conv.MapToStruct(ctx.Data.Content, param)
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

	if err := std.SaveBlockingActionName(ActionNameWaitInstallerComplete); err != nil {
		return fmt.Errorf("failed to save blocking action name: %w", err)
	}

	installCmd, sshErr := act.ssh.prepareWindowsSSHInstall(std)
	if !windowsAutoShouldFallback(installCmd != nil, sshErr) {
		defer installCmd.Close()
		if err = act.saveWaitInstallerPrivateData(std); err != nil {
			return err
		}

		return act.ssh.launchWindowsSSHInstall(std, installCmd)
	}

	std.InstanceData().Log().
		Zh("Windows SSH 安装准备失败，切换到 WMI 兜底: %s", sshErr.Error()).
		En("windows ssh install preparation failed, fallback to wmi: %s", sshErr.Error()).
		Warn()

	if err = act.saveWaitInstallerPrivateData(std); err != nil {
		return err
	}

	if err = act.wmi.installByWMI(std); err != nil {
		return fmt.Errorf("failed to install node by windows auto: %w", errors.Join(
			fmt.Errorf("windows ssh prepare failed: %w", sshErr),
			fmt.Errorf("wmi install failed: %w", err),
		))
	}

	return nil
}

func (act *actionInstallNodeByWindowsAuto) saveWaitInstallerPrivateData(std *nodeUtils.NodeActionStandarder) error {
	return saveWaitInstallerPrivateData(
		std.Context(), act.storageActionInstance, std.InstanceData().OperationInstanceID,
		false, true,
	)
}

func windowsAutoShouldFallback(sshLaunchReady bool, sshErr error) bool {
	return !sshLaunchReady && sshErr != nil
}
