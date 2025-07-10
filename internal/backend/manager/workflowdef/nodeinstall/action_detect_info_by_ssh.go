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

	nodedeployment "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node-deployment"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/deployconstant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/crypter"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/discover"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/sshx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameDetectInfoBySSH defines the action name.
	ActionNameDetectInfoBySSH = "detect_info_by_ssh"
)

// NewActionDetectInfoBySSH get a new action.
func NewActionDetectInfoBySSH(
	crypter crypter.Crypter,
	logger logger.Logger,
	storageNodeDeployment nodedeployment.IStorageNodeDeployment,
	provider discover.Provider,
) action.Definition {

	return &actionDetectInfoBySSH{
		crypter:               crypter,
		logger:                logger,
		storageNodeDeployment: storageNodeDeployment,
		provider:              provider,
	}
}

// ActParamDetectInfoBySSH ...
type ActParamDetectInfoBySSH struct {
	Token string `json:"token"`
}

type actionDetectInfoBySSH struct {
	installerGroup        iface.FileGroup
	crypter               crypter.Crypter
	logger                logger.Logger
	storageNodeDeployment nodedeployment.IStorageNodeDeployment
	provider              discover.Provider
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
// nolint: perfsprint,funlen,fnsize
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (act *actionDetectInfoBySSH) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActParamDetectInfoBySSH)
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

	client, err := buildSSHClient(ctx.Ctx, act.logger, act.crypter, info)
	if err != nil {
		return err
	}

	osType, cpuArch, connectedDir, err := act.detectInfo(ctx, client)
	if err != nil {
		return err
	}

	deployConstant, err := deployconstant.GetDeployConf(info.Dynamic.NodeGeneration, osType)
	if err != nil {
		return fmt.Errorf("failed to get deploy constant, err: %w", err)
	}

	// installer workspace priority: user specified in info > deploy constant default > connected dir.
	if info.InstallerWorkspace == "" {
		info.InstallerWorkspace = deployConstant.InstallerWorkspace
	}

	// this is a fallback strategy, if system has no specified workspace, use connected dir.
	if info.InstallerWorkspace == "" {
		info.InstallerWorkspace = connectedDir
	}

	info.Dynamic.NodeOsType = osType
	info.Dynamic.NodeCPUArch = cpuArch

	// we'll automatically use the system information to select the default version,
	// when NodeVersion is empty.
	if info.Dynamic.NodeVersion == "" {
		info.Dynamic.NodeVersion, err = autoSelectVersion(ctx, osType, cpuArch)
		if err != nil {
			return err
		}
	}

	err = checkVersionAvailability(
		ctx,
		info.Dynamic.NodeOsType,
		info.Dynamic.NodeCPUArch,
		info.Dynamic.NodeVersion,
	)
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
