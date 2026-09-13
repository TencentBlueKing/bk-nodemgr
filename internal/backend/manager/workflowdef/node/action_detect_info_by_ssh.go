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
	"strings"
	"time"

	nodeUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/node/utils"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/credit"
	nodeStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/creditvault"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/sshx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/file"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameDetectInfoBySSH defines the action name.
	ActionNameDetectInfoBySSH = "detect_info_by_ssh"
)

// NewActionDetectInfoBySSH get a new action.
func NewActionDetectInfoBySSH(capability *Capability) action.Definition {
	return &actionDetectInfoBySSH{
		storageHostCredit:     capability.StorageHostCredit,
		storageNodeDeployment: capability.StorageNode,
		storageHost:           capability.StorageTopo,
		fileHandler:           capability.FileHandler,
		passwordVault:         capability.HostPasswordVault,
	}
}

// ActParamDetectInfoBySSH ...
type ActParamDetectInfoBySSH struct {
	nodeUtils.NodeActionStandardParam `json:",inline"`
}

type actionDetectInfoBySSH struct {
	storageHostCredit     credit.IStorageHostCredit
	storageNodeDeployment nodeStg.IDaoNodeDeployment
	storageHost           topoStg.IStorageHost
	fileHandler           file.IPkgReleaseHandler
	passwordVault         creditvault.IHostPasswordVault
}

// Name returns the name of the action.
func (act *actionDetectInfoBySSH) Name() string {
	return ActionNameDetectInfoBySSH
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionDetectInfoBySSH) DisplayNameZh() string {
	return "通过 SSH 探测主机信息"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionDetectInfoBySSH) DisplayNameEn() string {
	return "Detect Host Info via SSH"
}

// Version returns the version of the action.
func (act *actionDetectInfoBySSH) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionDetectInfoBySSH) Description() string {
	return "Use ssh to connect to the target machine, detect the target machine's OS type and CPU architecture, " +
		"then use these info detect pkg version."
}

// Timeout returns the timeout of the action.
func (act *actionDetectInfoBySSH) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionDetectInfoBySSH) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionDetectInfoBySSH) MaxRetryCount() uint {
	return 5 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionDetectInfoBySSH) DelayFn(_ int) func() {
	return func() {
		time.Sleep(5 * time.Second) // nolint: mnd
	}
}

// Do this func define what the action will do.
// To ensure readability, this action uses fmt.Sprintf to concatenate characters.
// nolint: perfsprint,funlen,gocognit
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (act *actionDetectInfoBySSH) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActParamDetectInfoBySSH)
	err = conv.MapToStruct(ctx.Data.Content, param)
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

	// get ssh credit.
	credit := nodeUtils.NewCreditHandler(act.storageHostCredit, act.passwordVault)
	cMethod, cKey, err := credit.GetSSHCredit(std)
	if err != nil {
		return fmt.Errorf("failed to get ssh credit: %w", err)
	}

	std.InstanceData().Log().
		Zh("凭证获取成功, 认证方式(%s), 凭证长度(%d)", cMethod, len(cKey)).
		En("credit loaded, auth-method(%s), credential-length(%d)", cMethod, len(cKey)).
		Info()

	// generate the ssh client.
	client, err := sshx.NewClient(std.Context(), &sshx.Config{
		Network:    sshx.NetworkTCP,
		IP:         std.DeployInfo().Host.Dynamic.LoginIP,
		Port:       int(std.DeployInfo().Host.Dynamic.LoginPort),
		User:       std.DeployInfo().Host.Dynamic.LoginUser,
		AuthMethod: cMethod,
		Password: func() string {
			if cMethod == sshx.AuthMethodPassword {
				return cKey
			}

			return ""
		}(),
		PrivateKey: func() []byte {
			if cMethod == sshx.AuthMethodPrivateKey {
				return []byte(cKey)
			}

			return nil
		}(),
	}, sshx.DefaultTimeout)
	if err != nil {
		return fmt.Errorf("failed to generate new ssh client: %w", err)
	}

	osType, cpuArch, err := act.detectInfo(std.InstanceData(), client)
	if err != nil {
		return err
	}

	std.DeployInfo().Host.Dynamic.NodeOsType = osType
	std.DeployInfo().Host.Dynamic.NodeCPUArch = cpuArch

	releaseType, err := types.ConvertNodeRoleToReleaseType(std.DeployInfo().Host.Dynamic.NodeRole)
	if err != nil {
		std.InstanceData().Log().
			Zh("转换节点角色到发布类型失败, 错误: %v", err).
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
					Zh("用户选择目标版本, 使用版本 version(%s)", std.DeployInfo().Host.Dynamic.NodeVersion).
					En("user select, using target version. version(%s)", std.DeployInfo().Host.Dynamic.NodeVersion).
					Info()

				break
			}
		}
	} else if std.DeployInfo().Host.Dynamic.NodeVersion == "" {
		// we'll automatically use the system information to select the default version,
		// when NodeVersion is empty.
		std.DeployInfo().Host.Dynamic.NodeVersion, err = autoSelectVersion(std.Context(), CheckAndSelectVersionParam{
			fileHandler: act.fileHandler,
			ReleaseType: releaseType,
			Generation:  std.DeployInfo().Host.Dynamic.NodeGeneration,
			OSType:      std.DeployInfo().Host.Dynamic.NodeOsType,
			CPUArch:     std.DeployInfo().Host.Dynamic.NodeCPUArch,
		})
		if err != nil {
			return err
		}
		std.InstanceData().Log().
			Zh("自动选择目标版本, 使用系统默认版本. version(%s)", std.DeployInfo().Host.Dynamic.NodeVersion).
			En("auto select, using system default version. version(%s)", std.DeployInfo().Host.Dynamic.NodeVersion).
			Info()
	}

	err = checkVersionAvailability(
		std.Context(), CheckAndSelectVersionParam{
			fileHandler: act.fileHandler,
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

// nolint: nonamedreturns,perfsprint
func (act *actionDetectInfoBySSH) detectInfo(data *action.InstanceData, client *sshx.Client) (
	osType criteria.OSType, cpuArch criteria.CPUArch, err error) {

	// 1. detect target system
	osTypeStr, _, err := client.RunCommand("uname -s")
	if err != nil {
		err = fmt.Errorf("failed to run uname -s: %w", err)

		return "", "", err
	}
	osTypeStr = strings.TrimFunc(strings.ToLower(osTypeStr), func(r rune) bool {
		return r == '\n'
	})
	osType, err = platfmt.NormalizeOS(osTypeStr)
	if err != nil {
		return "", "", fmt.Errorf("failed to detect info: %w", err)
	}

	switch osType {
	case criteria.OSLinux, criteria.OSDarwin:
	default:
		err = fmt.Errorf("unsupported os type, os-type(%s)", osType)

		return "", "", err
	}
	data.Log().
		Zh("主机操作系统类型(%s)", osType).
		En("host-os-type(%s)", osType).
		Info()

	// 2. detect target cpu arch
	cpuArchStr, _, err := client.RunCommand("uname -m")
	if err != nil {
		return "", "", fmt.Errorf("failed to run uname -m: %w", err)
	}
	cpuArchStr = strings.TrimFunc(strings.ToLower(cpuArchStr), func(r rune) bool {
		return r == '\n'
	})
	cpuArch, err = platfmt.NormalizeArch(cpuArchStr)
	if err != nil {
		return "", "", fmt.Errorf("failed to detect info: %w", err)
	}

	data.Log().
		Zh("主机CPU架构(%s)", cpuArch).
		En("host-cpu-arch(%s)", cpuArch).
		Info()

	// 3. detect target dir
	connectedDir, _, err := client.RunCommand("pwd")
	if err != nil {
		err = fmt.Errorf("failed to run pwd: %w", err)

		return "", "", err
	}
	connectedDir = strings.TrimFunc(connectedDir, func(r rune) bool {
		return r == '\n'
	})
	data.Log().
		Zh("连接目录(%s)", connectedDir).
		En("connected-dir(%s)", connectedDir).
		Info()

	return osType, cpuArch, nil
}
