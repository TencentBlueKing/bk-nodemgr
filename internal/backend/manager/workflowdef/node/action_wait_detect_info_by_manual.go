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
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/release"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameWaitDetectInfoByManual defines the action name.
	ActionNameWaitDetectInfoByManual = "wait_detect_info_by_manual"

	waitDetectInfoByManualInterval = 1 * time.Second
)

// NewActionWaitDetectInfoByManual get a new action.
func NewActionWaitDetectInfoByManual(capability *Capability) action.Definition {
	return &actionWaitDetectInfoByManual{
		storageNodeDeployment: capability.StorageNode,
		storageHost:           capability.StorageTopo,
		storageActionInstance: capability.StorageWorkflow,
		storageRelease:        capability.StorageRelease,
	}
}

// ActParamWaitDetectInfoByManual ...
type ActParamWaitDetectInfoByManual struct {
	nodeUtils.NodeActionStandardParam `json:",inline"`
}

type actionWaitDetectInfoByManual struct {
	storageNodeDeployment nodeStg.IDaoNodeDeployment
	storageHost           topoStg.IStorageHost
	storageActionInstance workflow.IStorageActionInstance
	storageRelease        release.IStorage
}

// Name returns the name of the action.
func (act *actionWaitDetectInfoByManual) Name() string {
	return ActionNameWaitDetectInfoByManual
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionWaitDetectInfoByManual) DisplayNameZh() string {
	return "等待手动探测信息"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionWaitDetectInfoByManual) DisplayNameEn() string {
	return "Wait for Manual Detect Info"
}

// Version returns the version of the action.
func (act *actionWaitDetectInfoByManual) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionWaitDetectInfoByManual) Description() string {
	return "Wait for user to manually report system detection information (OS type, CPU architecture, etc.)"
}

// Timeout returns the timeout of the action.
func (act *actionWaitDetectInfoByManual) Timeout() time.Duration {
	return 5 * time.Minute // nolint:mnd
}

// Tags returns the tags of the action.
func (act *actionWaitDetectInfoByManual) Tags() []action.Tag {
	return []action.Tag{action.TagNeedManualExecInstallScript}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionWaitDetectInfoByManual) MaxRetryCount() uint {
	return 0
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionWaitDetectInfoByManual) DelayFn() func() {
	return func() {}
}

// Do this func define what the action will do.
// nolint: gocognit
func (act *actionWaitDetectInfoByManual) Do(ctx *action.InstanceContext) error {
	param := new(ActParamWaitDetectInfoByManual)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	// initialize standard data.
	std := nodeUtils.NewNodeActionStandarder(act.storageNodeDeployment, act.storageHost)
	if err = std.Initialize(ctx, param.NodeActionStandardParam); err != nil {
		return err
	}
	defer func() {
		if storeErr := std.Save(); storeErr != nil {
			err = errors.Join(storeErr, err)
		}
	}()

	ticker := time.NewTicker(waitDetectInfoByManualInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Ctx.Done():
			return nil

		case <-ticker.C:
			detectInfo, err := act.tryFetchDetectInfo(std, types.PDKeyReportDetectInfo)
			if err == errDetectInfoNotReceived {
				continue
			}
			if err != nil {
				return err
			}

			if detectInfo.ErrMsg != "" {
				std.InstanceData().Log().
					Zh("获取探测信息失败: %v", detectInfo.ErrMsg).
					En("failed to get detect info: %v", detectInfo.ErrMsg).
					Error()

				return fmt.Errorf("failed to get detect info: %v", detectInfo.ErrMsg)
			}

			if err = act.applyInfo(std, detectInfo); err != nil {
				return err
			}
		}

		break
	}

	return nil
}

func (act *actionWaitDetectInfoByManual) applyInfo(std *nodeUtils.NodeActionStandarder, detectInfo *types.PDDetectInfo) error {
	osType, err := platform.NormalizeOS(detectInfo.OsType)
	if err != nil {
		return err
	}

	cpuArch, err := platform.NormalizeArch(detectInfo.CPUArch)
	if err != nil {
		return err
	}

	std.DeployInfo().Host.Dynamic.NodeOsType = osType
	std.DeployInfo().Host.Dynamic.NodeCPUArch = cpuArch

	releaseType, err := types.ConvertNodeRoleToReleaseType(std.DeployInfo().Host.Dynamic.NodeRole)
	if err != nil {
		std.InstanceData().Log().
			Zh("节点角色转换为发布类型失败: %v", err).
			En("failed to convert node role to release type. err: %v", err).
			Error()

		return err
	}

	if len(std.DeployInfo().TargetVersion) > 0 {
		for _, v := range std.DeployInfo().TargetVersion {
			if std.DeployInfo().Host.Dynamic.NodeOsType == v.OsType && std.DeployInfo().Host.Dynamic.NodeCPUArch == v.CPUArch {
				// you can guarantee that there are no duplicates in the TargetVersion.
				std.DeployInfo().Host.Dynamic.NodeVersion = v.Version
				std.InstanceData().Log().
					Zh("用户选择, 使用目标版本. version(%s)", std.DeployInfo().Host.Dynamic.NodeVersion).
					En("user select, using target version. version(%s)", std.DeployInfo().Host.Dynamic.NodeVersion).
					Info()

				break
			}
		}
	} else if std.DeployInfo().Host.Dynamic.NodeVersion == "" {
		// we'll automatically use the system information to select the default version,
		// when NodeVersion is empty.
		std.DeployInfo().Host.Dynamic.NodeVersion, err = autoSelectVersion(std.Context(), CheckAndSelectVersionParam{
			daoRelease:  act.storageRelease,
			ReleaseType: releaseType,
			Generation:  std.DeployInfo().Host.Dynamic.NodeGeneration,
			OSType:      std.DeployInfo().Host.Dynamic.NodeOsType,
			CPUArch:     std.DeployInfo().Host.Dynamic.NodeCPUArch,
		})
		if err != nil {
			return err
		}
		std.InstanceData().Log().
			Zh("自动选择, 使用系统默认版本. version(%s)", std.DeployInfo().Host.Dynamic.NodeVersion).
			En("auto select, using system default version. version(%s)", std.DeployInfo().Host.Dynamic.NodeVersion).
			Info()
	}

	err = checkVersionAvailability(
		std.Context(), CheckAndSelectVersionParam{
			daoRelease:  act.storageRelease,
			ReleaseType: releaseType,
			Generation:  std.DeployInfo().Host.Dynamic.NodeGeneration,
			OSType:      std.DeployInfo().Host.Dynamic.NodeOsType,
			CPUArch:     std.DeployInfo().Host.Dynamic.NodeCPUArch,
			Version:     std.DeployInfo().Host.Dynamic.NodeVersion,
		})
	if err != nil {
		return err
	}

	return nil
}

var (
	errDetectInfoNotReceived = errors.New("detect info not received")
)

func (act *actionWaitDetectInfoByManual) tryFetchDetectInfo(std *nodeUtils.NodeActionStandarder, key string) (*types.PDDetectInfo, error) {
	privateData, err := act.storageActionInstance.GetActionInstancePrivateData(
		std.Context(),
		std.InstanceData().OperationInstanceID,
		ActionNameWaitDetectInfoByManual)
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to get action private data")

		return nil, err
	}

	rawValue, exists := privateData[key]
	if !exists {
		logger.G.Sys().With("oper-inst-id", std.InstanceData().OperationInstanceID, "key", key).Debug("no receive data")

		return nil, errDetectInfoNotReceived
	}

	value, ok := rawValue.(map[string]any)
	if !ok {
		logger.G.Sys().With("oper-inst-id", std.InstanceData().OperationInstanceID, "key", key, "raw-value", rawValue).
			Error("unexpected type for fetched value")

		return nil, fmt.Errorf("got unexpected value from action private data. key(%s)", key)
	}

	data := new(types.PDDetectInfo)
	if err = conv.MapToStruct(value, data); err != nil {
		logger.G.Sys().With("oper-inst-id", std.InstanceData().OperationInstanceID, "key", key, "raw-value", rawValue).
			Error("failed to convert value to struct")

		return nil, err
	}

	return data, nil
}
