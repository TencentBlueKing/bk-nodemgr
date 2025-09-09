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
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/node/utils"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/credit"
	nodedeployment "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node-deployment"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/release"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/creditvault"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/deployconstant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/sshx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameDetectInfoBySSH defines the action name.
	ActionNameDetectInfoBySSH = "detect_info_by_ssh"
)

// NewActionDetectInfoBySSH get a new action.
func NewActionDetectInfoBySSH(
	logger logger.ILogger,
	storageNodeDeployment nodedeployment.IStorageNodeDeployment,
	storageRelease release.IStorage,
	storageHostCredit credit.IStorageHostCredit,
	passwordVault creditvault.IHostPasswordVault,
) action.Definition {

	return &actionDetectInfoBySSH{
		logger:                logger,
		storageHostCredit:     storageHostCredit,
		storageNodeDeployment: storageNodeDeployment,
		storageRelease:        storageRelease,
		passwordVault:         passwordVault,
	}
}

// ActParamDetectInfoBySSH ...
type ActParamDetectInfoBySSH struct {
	utils.NodeActionStandardParam `json:",inline"`
}

type actionDetectInfoBySSH struct {
	logger                logger.ILogger
	storageHostCredit     credit.IStorageHostCredit
	storageNodeDeployment nodedeployment.IStorageNodeDeployment
	storageRelease        release.IStorage
	passwordVault         creditvault.IHostPasswordVault
}

// Name returns the name of the action.
func (act *actionDetectInfoBySSH) Name() string {
	return ActionNameDetectInfoBySSH
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
func (act *actionDetectInfoBySSH) DelayFn() func() {
	return func() {
		time.Sleep(5 * time.Second) // nolint: mnd
	}
}

// Do this func define what the action will do.
// To ensure readability, this action uses fmt.Sprintf to concatenate characters.
// nolint: perfsprint,funlen
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (act *actionDetectInfoBySSH) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActParamDetectInfoBySSH)
	err = conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		err = fmt.Errorf("failed to convert param: %w", err)

		return err
	}

	// initialize standard data.
	std := utils.NewNodeActionStandarder(act.storageNodeDeployment)
	if err = std.Initialize(ctx, param.NodeActionStandardParam); err != nil {
		return err
	}
	defer func() {
		if storeErr := std.Save(); storeErr != nil {
			err = errors.Join(storeErr, err)
		}
	}()

	// get ssh credit.
	credit := utils.NewCreditHandler(act.storageHostCredit, act.passwordVault)
	cMethod, cKey, err := credit.GetSSHCredit(std)
	if err != nil {
		return fmt.Errorf("failed to get ssh credit: %w", err)
	}

	// generate the ssh client.
	client, err := sshx.NewClient(std.Context(), &sshx.Config{
		Network:    sshx.NetworkTCP,
		IP:         std.DeployInfo().Host.Dynamic.LoginIP,
		Port:       int(std.DeployInfo().Host.Dynamic.LoginPort),
		User:       std.DeployInfo().Host.Dynamic.LoginUser,
		Logger:     act.logger,
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

	osType, cpuArch, connectedDir, err := act.detectInfo(ctx, client)
	if err != nil {
		return err
	}

	deployConstant, err := deployconstant.GetNodeDeployConf(std.DeployInfo().Host.Dynamic.NodeGeneration, osType)
	if err != nil {
		return fmt.Errorf("failed to get deploy constant, err: %w", err)
	}

	// installer workdir priority: user specified in info > deploy constant default > connected dir.
	if std.DeployInfo().InstallerWorkDir == "" {
		std.DeployInfo().InstallerWorkDir = deployConstant.WorkDir
	}

	// this is a fallback strategy, if system has no specified workdir, use connected dir.
	if std.DeployInfo().InstallerWorkDir == "" {
		std.DeployInfo().InstallerWorkDir = connectedDir
	}

	std.DeployInfo().Host.Dynamic.NodeOsType = osType
	std.DeployInfo().Host.Dynamic.NodeCPUArch = cpuArch

	releaseType, err := types.ConvertNodeRoleToReleaseType(std.DeployInfo().Host.Dynamic.NodeRole)
	if err != nil {
		ctx.Data.LogE(fmt.Sprintf("failed to convert node role to release type. err: %v", err))
		return err
	}

	if len(std.DeployInfo().TargetVersion) > 0 {
		for _, v := range std.DeployInfo().TargetVersion {
			if std.DeployInfo().Host.Dynamic.NodeOsType == v.OsType && std.DeployInfo().Host.Dynamic.NodeCPUArch == v.CPUArch {
				// you can guarantee that there are no duplicates in the TargetVersion.
				std.DeployInfo().Host.Dynamic.NodeVersion = v.Version
				ctx.Data.LogI(fmt.Sprintf("user select, using target version. version(%s)", std.DeployInfo().Host.Dynamic.NodeVersion))
				break
			}
		}
	} else {
		// we'll automatically use the system information to select the default version,
		// when NodeVersion is empty.
		if std.DeployInfo().Host.Dynamic.NodeVersion == "" {
			std.DeployInfo().Host.Dynamic.NodeVersion, err = autoSelectVersion(ctx.Ctx, CheckAndSelectVersionParam{
				daoRelease:  act.storageRelease,
				ReleaseType: releaseType,
				Generation:  std.DeployInfo().Host.Dynamic.NodeGeneration,
				OSType:      std.DeployInfo().Host.Dynamic.NodeOsType,
				CPUArch:     std.DeployInfo().Host.Dynamic.NodeCPUArch,
			})
			if err != nil {
				return err
			}
			ctx.Data.LogI(fmt.Sprintf("auto select, using system default version. version(%s)", std.DeployInfo().Host.Dynamic.NodeVersion))
		}
	}

	err = checkVersionAvailability(
		ctx.Ctx, CheckAndSelectVersionParam{
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

// nolint: nonamedreturns,perfsprint
func (act *actionDetectInfoBySSH) detectInfo(ctx *action.InstanceContext, client *sshx.Client) (
	osType criteria.OSType, cpuArch criteria.CPUArch, connectedDir string, err error) {

	// 1. detect target system
	osTypeStr, err := client.RunCommand("uname -s")
	if err != nil {
		err = fmt.Errorf("failed to run uname -a, err: %w", err)

		return "", "", "", err
	}
	osTypeStr = strings.TrimFunc(strings.ToLower(osTypeStr), func(r rune) bool {
		return r == '\n'
	})
	osType, err = platform.NormalizeOS(osTypeStr)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to detect info, err: %w", err)
	}

	switch osType {
	case criteria.OSLinux, criteria.OSDarwin:
	default:
		err = fmt.Errorf("unsupported os type, os-type(%s)", osType)

		return "", "", "", err
	}
	ctx.Data.LogI(fmt.Sprintf("host-os-type(%s)", osType))

	// 2. detect target cpu arch
	cpuArchStr, err := client.RunCommand("uname -m")
	if err != nil {
		return "", "", "", fmt.Errorf("failed to run uname -m, err: %w", err)
	}
	cpuArchStr = strings.TrimFunc(strings.ToLower(cpuArchStr), func(r rune) bool {
		return r == '\n'
	})
	cpuArch, err = platform.NormalizeArch(cpuArchStr)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to detect info, err: %w", err)
	}

	ctx.Data.LogI(fmt.Sprintf("host-cpu-arch(%s)", cpuArch))

	// 3. detect target dir
	connectedDir, err = client.RunCommand("pwd")
	if err != nil {
		err = fmt.Errorf("failed to run pwd, err: %w", err)

		return "", "", "", err
	}
	connectedDir = strings.TrimFunc(connectedDir, func(r rune) bool {
		return r == '\n'
	})
	ctx.Data.LogI(fmt.Sprintf("connected-dir(%s)", connectedDir))

	return osType, cpuArch, connectedDir, nil
}
