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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/deployconstant"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameResolveOfflineDetectInfo defines the action name.
	ActionNameResolveOfflineDetectInfo = "resolve_offline_detect_info"
)

// NewActionResolveOfflineDetectInfo get a new action.
func NewActionResolveOfflineDetectInfo(capability *Capability) action.Definition {
	return &actionResolveOfflineDetectInfo{
		storageNodeDeployment: capability.StorageNode,
		storageHost:           capability.StorageTopo,
		storageRelease:        capability.StorageRelease,
	}
}

// ActParamResolveOfflineDetectInfo is the parameter for actionResolveOfflineDetectInfo.
type ActParamResolveOfflineDetectInfo struct {
	nodeUtils.NodeActionStandardParam `json:",inline"`
}

type actionResolveOfflineDetectInfo struct {
	storageNodeDeployment nodeStg.IDaoNodeDeployment
	storageHost           topoStg.IStorageHost
	storageRelease        release.IStorage
}

// Name returns the name of the action.
func (act *actionResolveOfflineDetectInfo) Name() string {
	return ActionNameResolveOfflineDetectInfo
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionResolveOfflineDetectInfo) DisplayNameZh() string {
	return "解析离线安装信息"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionResolveOfflineDetectInfo) DisplayNameEn() string {
	return "Resolve Offline Detect Info"
}

// Version returns the version of the action.
func (act *actionResolveOfflineDetectInfo) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionResolveOfflineDetectInfo) Description() string {
	return "Resolve OS/Arch info from user-provided cpu_arch for offline installation, without SSH detection"
}

// Timeout returns the timeout of the action.
func (act *actionResolveOfflineDetectInfo) Timeout() time.Duration {
	return 5 * time.Minute // nolint:mnd
}

// Tags returns the tags of the action.
func (act *actionResolveOfflineDetectInfo) Tags() []action.Tag {
	return nil
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionResolveOfflineDetectInfo) MaxRetryCount() uint {
	return 3 // nolint:mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionResolveOfflineDetectInfo) DelayFn() func() {
	return func() {}
}

// Do this func define what the action will do.
func (act *actionResolveOfflineDetectInfo) Do(ctx *action.InstanceContext) error {
	param := new(ActParamResolveOfflineDetectInfo)
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

	// Read os_type from host static (user-provided at install time).
	rawOsType := std.DeployInfo().Host.Static.OSType
	osType, err := platfmt.NormalizeOS(rawOsType)
	if err != nil {
		return fmt.Errorf("failed to normalize os_type(%s): %w", rawOsType, err)
	}

	// Read cpu_arch from host dynamic (set from NodeProxyInstallHost.cpu_arch at install time).
	rawCPUArch := string(std.DeployInfo().Host.Dynamic.NodeCPUArch)
	cpuArch, err := platfmt.NormalizeArch(rawCPUArch)
	if err != nil {
		return fmt.Errorf("failed to normalize cpu_arch(%s): %w", rawCPUArch, err)
	}

	std.DeployInfo().Host.Dynamic.NodeOsType = osType
	std.DeployInfo().Host.Dynamic.NodeCPUArch = cpuArch

	deployConstant, err := deployconstant.GetNodeDeployConf(std.DeployInfo().Host.Dynamic.NodeGeneration, osType)
	if err != nil {
		return fmt.Errorf("failed to get deploy constant: %w", err)
	}

	if std.DeployInfo().InstallerWorkDir == "" {
		std.DeployInfo().InstallerWorkDir = deployConstant.WorkDir
	}

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
				std.DeployInfo().Host.Dynamic.NodeVersion = v.Version
				std.InstanceData().Log().
					Zh("用户选择, 使用目标版本 version(%s)", std.DeployInfo().Host.Dynamic.NodeVersion).
					En("user select, using target version. version(%s)", std.DeployInfo().Host.Dynamic.NodeVersion).
					Info()

				break
			}
		}
	} else if std.DeployInfo().Host.Dynamic.NodeVersion == "" {
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

	std.InstanceData().Log().
		Zh("离线探测信息解析完成. os-type(%s) cpu-arch(%s) version(%s)",
			std.DeployInfo().Host.Dynamic.NodeOsType,
			std.DeployInfo().Host.Dynamic.NodeCPUArch,
			std.DeployInfo().Host.Dynamic.NodeVersion).
		En("offline detect info resolved. os-type(%s) cpu-arch(%s) version(%s)",
			std.DeployInfo().Host.Dynamic.NodeOsType,
			std.DeployInfo().Host.Dynamic.NodeCPUArch,
			std.DeployInfo().Host.Dynamic.NodeVersion).
		Info()

	return nil
}
