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
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/credit"
	nodeStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/release"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	workflowStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/creditvault"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/sshx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/wmix"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameDetectInfoByWindowsAuto defines the action name.
	ActionNameDetectInfoByWindowsAuto = "detect_info_by_windows_auto"

	windowsAutoDetectMethodSSH windowsAutoDetectMethod = "ssh"
	windowsAutoDetectMethodWMI windowsAutoDetectMethod = "wmi"
)

// NewActionDetectInfoByWindowsAuto get a new action.
func NewActionDetectInfoByWindowsAuto(capability *Capability) action.Definition {
	return &actionDetectInfoByWindowsAuto{
		storageHostCredit:     capability.StorageHostCredit,
		storageNodeDeployment: capability.StorageNode,
		storageHost:           capability.StorageTopo,
		storageRelease:        capability.StorageRelease,
		storageActionInstance: capability.StorageWorkflow,
		passwordVault:         capability.HostPasswordVault,
	}
}

// ActParamDetectInfoByWindowsAuto defines the action parameter.
type ActParamDetectInfoByWindowsAuto struct {
	nodeUtils.NodeActionStandardParam `json:",inline"`
}

type actionDetectInfoByWindowsAuto struct {
	storageHostCredit     credit.IStorageHostCredit
	storageNodeDeployment nodeStg.IDaoNodeDeployment
	storageHost           topoStg.IStorageHost
	storageRelease        release.IStorage
	storageActionInstance workflowStg.IStorage
	passwordVault         creditvault.IHostPasswordVault
}

type windowsAutoDetectMethod string

type windowsAutoDetectResult struct {
	osType  criteria.OSType
	cpuArch criteria.CPUArch
	method  windowsAutoDetectMethod
	profile string
	sshErr  error
}

// Name returns the name of the action.
func (act *actionDetectInfoByWindowsAuto) Name() string {
	return ActionNameDetectInfoByWindowsAuto
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionDetectInfoByWindowsAuto) DisplayNameZh() string {
	return "自动探测 Windows 主机信息"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionDetectInfoByWindowsAuto) DisplayNameEn() string {
	return "Detect Windows Host Info Automatically"
}

// Version returns the version of the action.
func (act *actionDetectInfoByWindowsAuto) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionDetectInfoByWindowsAuto) Description() string {
	return "Use Windows SSH first to detect host info, and fallback to WMI when SSH detection fails."
}

// Timeout returns the timeout of the action.
func (act *actionDetectInfoByWindowsAuto) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionDetectInfoByWindowsAuto) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionDetectInfoByWindowsAuto) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionDetectInfoByWindowsAuto) DelayFn() func() {
	return func() {
		time.Sleep(5 * time.Second) // nolint: mnd
	}
}

// Do this func define what the action will do.
// nolint: perfsprint,funlen,gocognit
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (act *actionDetectInfoByWindowsAuto) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActParamDetectInfoByWindowsAuto)
	err = conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return fmt.Errorf("failed to convert param: %w", err)
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

	result, err := detectWindowsAutoInfo(
		func() (windowsSSHDetectResult, error) {
			return act.detectByWindowsSSH(std)
		},
		func() (windowsWMIDetectResult, error) {
			return act.detectByWMI(std)
		},
	)
	if err != nil {
		return err
	}

	if result.method == windowsAutoDetectMethodWMI {
		std.InstanceData().Log().
			Zh("Windows SSH 探测失败，切换到 WMI 兜底: %s", result.sshErr.Error()).
			En("windows ssh detect failed, fallback to wmi: %s", result.sshErr.Error()).
			Warn()
		logWindowsWMIInfo(std.InstanceData(), windowsWMIDetectResult{osType: result.osType, cpuArch: result.cpuArch})
	}

	if err = act.applyWindowsAutoDetectResult(std, result); err != nil {
		return err
	}

	if err = act.storageActionInstance.UpsertActionInstancePrivateData(
		std.Context(),
		std.InstanceData().OperationInstanceID,
		ActionNameDetectInfoByWindowsAuto,
		buildWindowsAutoDetectPrivateData(result),
	); err != nil {
		return fmt.Errorf("failed to save windows auto detect private data: %w", err)
	}

	return nil
}

func (act *actionDetectInfoByWindowsAuto) detectByWindowsSSH(std *nodeUtils.NodeActionStandarder) (windowsSSHDetectResult, error) {
	credit := nodeUtils.NewCreditHandler(act.storageHostCredit, act.passwordVault)
	cMethod, cKey, err := credit.GetSSHCredit(std)
	if err != nil {
		return windowsSSHDetectResult{}, fmt.Errorf("failed to get ssh credit: %w", err)
	}

	std.InstanceData().Log().
		Zh("凭证获取成功, 认证方式(%s), 凭证长度(%d)", cMethod, len(cKey)).
		En("credit loaded, auth-method(%s), credential-length(%d)", cMethod, len(cKey)).
		Info()

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
		Ciphers: sshx.WindowsCompatibleCiphers(),
		MACs:    sshx.WindowsCompatibleMACs(),
	}, sshx.DefaultTimeout)
	if err != nil {
		return windowsSSHDetectResult{}, fmt.Errorf("failed to generate new ssh client: %w", err)
	}
	defer func() {
		_ = client.Close()
	}()

	return detectWindowsSSHInfo(std.InstanceData(), client)
}

func (act *actionDetectInfoByWindowsAuto) detectByWMI(std *nodeUtils.NodeActionStandarder) (windowsWMIDetectResult, error) {
	credit := nodeUtils.NewCreditHandler(act.storageHostCredit, act.passwordVault)
	cMethod, cKey, err := credit.GetWMICredit(std)
	if err != nil {
		return windowsWMIDetectResult{}, fmt.Errorf("failed to get wmi credit: %w", err)
	}

	std.InstanceData().Log().
		Zh("凭证获取成功, 认证方式(%s), 凭证长度(%d)", cMethod, len(cKey)).
		En("credit loaded, auth-method(%s), credential-length(%d)", cMethod, len(cKey)).
		Info()

	client, err := wmix.NewClient(&wmix.Config{
		IP:         std.DeployInfo().Host.Dynamic.LoginIP,
		User:       std.DeployInfo().Host.Dynamic.LoginUser,
		AuthMethod: cMethod,
		Password: func() string {
			if cMethod == wmix.AuthMethodPassword {
				return cKey
			}

			return ""
		}(),
		Timeout: wmix.DefaultTimeout,
	})
	if err != nil {
		return windowsWMIDetectResult{}, fmt.Errorf("failed to generate new wmi client: %w", err)
	}

	return detectWindowsWMIInfo(std.Context(), client)
}

func (act *actionDetectInfoByWindowsAuto) applyWindowsAutoDetectResult(
	std *nodeUtils.NodeActionStandarder, result windowsAutoDetectResult,
) error {
	std.DeployInfo().Host.Dynamic.NodeOsType = result.osType
	std.DeployInfo().Host.Dynamic.NodeCPUArch = result.cpuArch

	releaseType, err := types.ConvertNodeRoleToReleaseType(std.DeployInfo().Host.Dynamic.NodeRole)
	if err != nil {
		return err
	}
	if len(std.DeployInfo().TargetVersion) > 0 {
		for _, v := range std.DeployInfo().TargetVersion {
			if std.DeployInfo().Host.Dynamic.NodeOsType == v.OsType && std.DeployInfo().Host.Dynamic.NodeCPUArch == v.CPUArch {
				// you can guarantee that there are no duplicates in the TargetVersion.
				std.DeployInfo().Host.Dynamic.NodeVersion = v.Version
				std.InstanceData().Log().
					Zh("用户选择, 使用目标版本 version(%s)", std.DeployInfo().Host.Dynamic.NodeVersion).
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

	if err = checkVersionAvailability(
		std.Context(), CheckAndSelectVersionParam{
			daoRelease:  act.storageRelease,
			ReleaseType: releaseType,
			Generation:  std.DeployInfo().Host.Dynamic.NodeGeneration,
			OSType:      std.DeployInfo().Host.Dynamic.NodeOsType,
			CPUArch:     std.DeployInfo().Host.Dynamic.NodeCPUArch,
			Version:     std.DeployInfo().Host.Dynamic.NodeVersion,
		}); err != nil {
		return err
	}

	return nil
}

func detectWindowsAutoInfo(
	detectBySSH func() (windowsSSHDetectResult, error),
	detectByWMI func() (windowsWMIDetectResult, error),
) (windowsAutoDetectResult, error) {
	sshResult, sshErr := detectBySSH()
	if sshErr == nil {
		return windowsAutoDetectResult{
			osType:  sshResult.osType,
			cpuArch: sshResult.cpuArch,
			method:  windowsAutoDetectMethodSSH,
			profile: sshResult.profile,
		}, nil
	}

	wmiResult, wmiErr := detectByWMI()
	if wmiErr != nil {
		return windowsAutoDetectResult{}, fmt.Errorf(
			"failed to detect windows info by auto, ssh detect error(%v), wmi detect error: %w", sshErr, wmiErr,
		)
	}

	return windowsAutoDetectResult{
		osType:  wmiResult.osType,
		cpuArch: wmiResult.cpuArch,
		method:  windowsAutoDetectMethodWMI,
		sshErr:  sshErr,
	}, nil
}

func buildWindowsAutoDetectPrivateData(result windowsAutoDetectResult) map[string]any {
	privateData := map[string]any{
		types.PDKeyWindowsAutoDetectMethod: string(result.method),
	}
	if result.method == windowsAutoDetectMethodSSH {
		privateData[types.PDKeyWindowsSSHProfile] = result.profile
	}
	if result.method == windowsAutoDetectMethodWMI && result.sshErr != nil {
		privateData[types.PDKeyWindowsAutoSSHDetectError] = result.sshErr.Error()
	}

	return privateData
}
