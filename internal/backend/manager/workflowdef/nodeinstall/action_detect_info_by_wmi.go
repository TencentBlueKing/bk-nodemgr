/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package nodeinstall

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/credit"
	nodedeployment "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node-deployment"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/release"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/creditvault"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/deployconstant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/wmix"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameDetectInfoByWMI defines the action name.
	ActionNameDetectInfoByWMI = "detect_info_by_wmi"
)

// NewActionDetectInfoByWMI get a new action.
func NewActionDetectInfoByWMI(
	logger logger.Logger,
	storageNodeDeployment nodedeployment.IStorageNodeDeployment,
	storageRelease release.IStorage,
	storageHostCredit credit.IStorageHostCredit,
	passwordVault creditvault.IHostPasswordVault,
) action.Definition {

	return &actionDetectInfoByWMI{
		storageHostCredit:     storageHostCredit,
		logger:                logger,
		storageNodeDeployment: storageNodeDeployment,
		storageRelease:        storageRelease,
		passwordVault:         passwordVault,
	}
}

// ActParamDetectInfoByWMI ...
type ActParamDetectInfoByWMI struct {
	Token    string `json:"token"`
	Operator string `json:"operator"`
}

type actionDetectInfoByWMI struct {
	logger                logger.Logger
	storageHostCredit     credit.IStorageHostCredit
	storageNodeDeployment nodedeployment.IStorageNodeDeployment
	storageRelease        release.IStorage
	passwordVault         creditvault.IHostPasswordVault
}

// Name returns the name of the action.
func (act *actionDetectInfoByWMI) Name() string {
	return ActionNameDetectInfoByWMI
}

// Version returns the version of the action.
func (act *actionDetectInfoByWMI) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionDetectInfoByWMI) Description() string {
	return "Use wmi to connect to the target machine, transfer files through sftp, and execute the installation command"
}

// Timeout returns the timeout of the action.
func (act *actionDetectInfoByWMI) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionDetectInfoByWMI) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionDetectInfoByWMI) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionDetectInfoByWMI) DelayFn() func() {
	return func() {
		time.Sleep(5 * time.Second) // nolint: mnd
	}
}

const windowsDefaultInstallerWorkDir = "C:\\tmp"

// Do this func define what the action will do.
// To ensure readability, this action uses fmt.Sprintf to concatenate characters.
// nolint: perfsprint,funlen,fnsize
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (act *actionDetectInfoByWMI) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActParamDetectInfoByWMI)
	err = conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		err = fmt.Errorf("failed to convert param, err: %w", err)

		return err
	}

	info, err := act.storageNodeDeployment.GetInfo(ctx.Ctx, param.Token)
	if err != nil {
		return err
	}

	defer func() {
		if storeErr := act.storageNodeDeployment.UpdateInfo(ctx.Ctx, param.Token, info); storeErr != nil {
			err = errors.Join(storeErr, err)
		}
	}()

	client, err := generateWMIClient(ctx.Ctx, param.Operator, act.logger, act.storageHostCredit, act.passwordVault, info)
	if err != nil {
		return err
	}

	osType, cpuArch, err := act.detectInfo(ctx, client)
	if err != nil {
		return err
	}

	deployConstant, err := deployconstant.GetDeployConf(info.Host.Dynamic.NodeGeneration, osType)
	if err != nil {
		return fmt.Errorf("failed to get deploy constant, err: %w", err)
	}

	// installer workdir priority: user specified in info > deploy constant default > connected dir.
	if info.InstallerWorkDir == "" {
		info.InstallerWorkDir = deployConstant.WorkDir
	}

	if info.InstallerWorkDir == "" {
		info.InstallerWorkDir = windowsDefaultInstallerWorkDir
	}

	info.Host.Dynamic.NodeOsType = osType
	info.Host.Dynamic.NodeCPUArch = cpuArch

	releaseType, err := types.ConvertNodeRoleToReleaseType(info.Host.Dynamic.NodeRole)
	if err != nil {
		return err
	}
	if len(info.TargetVersion) > 0 {
		for _, v := range info.TargetVersion {
			if info.Host.Dynamic.NodeOsType == v.OsType && info.Host.Dynamic.NodeCPUArch == v.CPUArch {
				// you can guarantee that there are no duplicates in the TargetVersion.
				info.Host.Dynamic.NodeVersion = v.Version
				ctx.Data.LogI(fmt.Sprintf("user select, using target version. version(%s)", info.Host.Dynamic.NodeVersion))
				break
			}
		}
	} else {
		// we'll automatically use the system information to select the default version,
		// when NodeVersion is empty.
		if info.Host.Dynamic.NodeVersion == "" {
			info.Host.Dynamic.NodeVersion, err = autoSelectVersion(ctx.Ctx, CheckAndSelectVersionParam{
				daoRelease:  act.storageRelease,
				ReleaseType: releaseType,
				Generation:  info.Host.Dynamic.NodeGeneration,
				OSType:      info.Host.Dynamic.NodeOsType,
				CPUArch:     info.Host.Dynamic.NodeCPUArch,
			})
			if err != nil {
				return err
			}
			ctx.Data.LogI(fmt.Sprintf("auto select, using system default version. version(%s)", info.Host.Dynamic.NodeVersion))
		}
	}

	err = checkVersionAvailability(
		ctx.Ctx, CheckAndSelectVersionParam{
			daoRelease:  act.storageRelease,
			ReleaseType: releaseType,
			Generation:  info.Host.Dynamic.NodeGeneration,
			OSType:      info.Host.Dynamic.NodeOsType,
			CPUArch:     info.Host.Dynamic.NodeCPUArch,
			Version:     info.Host.Dynamic.NodeVersion,
		})
	if err != nil {
		return err
	}

	return nil
}

// inorder to improve readability, use fmt.Sprintf to construct command line, and use named return.
// nolint: nonamedreturns,perfsprint
func (act *actionDetectInfoByWMI) detectInfo(ctx *action.InstanceContext, client *wmix.Client) (
	osType criteria.OSType, cpuArch criteria.CPUArch, err error) {

	// 1. detect target system
	osTypeStr, _, err := client.RunCommand(ctx.Ctx, "ver")
	if err != nil {
		err = fmt.Errorf("failed to run ver, err: %w", err)

		return "", "", err
	}
	osTypeStr = strings.TrimFunc(strings.ToLower(osTypeStr), func(r rune) bool {
		return r == '\n'
	})
	osType, err = platform.NormalizeOS(osTypeStr)
	if err != nil {
		return "", "", fmt.Errorf("failed to detect info, err: %w", err)
	}

	switch osType {
	case criteria.OSWindows:
	default:
		err = fmt.Errorf("unsupported os type, os-type(%s)", osType)

		return "", "", err
	}
	ctx.Data.LogI(fmt.Sprintf("host-os-type(%s)", osType))

	// 2. detect target cpu arch
	cpuArchStr, _, err := client.RunCommand(ctx.Ctx, "echo %PROCESSOR_ARCHITECTURE%")
	if err != nil {
		return "", "", fmt.Errorf("failed to run uname -m, err: %w", err)
	}
	cpuArchStr = strings.TrimFunc(strings.ToLower(cpuArchStr), func(r rune) bool {
		return r == '\n'
	})
	cpuArch, err = platform.NormalizeArch(cpuArchStr)
	if err != nil {
		return "", "", fmt.Errorf("failed to detect info, err: %w", err)
	}

	ctx.Data.LogI(fmt.Sprintf("host-cpu-arch(%s)", cpuArch))

	return osType, cpuArch, nil
}
