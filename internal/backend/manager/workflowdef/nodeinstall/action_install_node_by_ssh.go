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
	"path"
	"strings"
	"time"

	nodedeployment "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node-deployment"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/deployconstant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/tool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/system"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/crypter"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/discover"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameInstallNodeBySSH defines the action name.
	ActionNameInstallNodeBySSH = "install_node_by_ssh"
)

// NewActionInstallNodeBySSH get a new action.
func NewActionInstallNodeBySSH(
	installerFileGroup iface.FileGroup,
	crypter crypter.Crypter,
	logger logger.Logger,
	storageNodeDeployment nodedeployment.IStorageNodeDeployment,
	provider discover.Provider,
) action.Definition {

	return &actionInstallNodeBySSH{
		installerGroup:        installerFileGroup,
		crypter:               crypter,
		logger:                logger,
		storageNodeDeployment: storageNodeDeployment,
		provider:              provider,
	}
}

// ActParamInstallAgentBySSH ...
type ActParamInstallAgentBySSH struct {
	Token string `json:"token"`
}

// InstallParams this struct defines the parameters for installing agent.
type InstallParams struct {
	InstallerPath      string
	NodeRole           types.NodeRole
	CallbackEndpoint   string
	DownloadEndpoint   string
	PkgVersion         string
	PkgGeneration      types.Generation
	GseRoot            string
	Token              string
	OperInstID         string
	InstallerWorkspace string
	AdditionArgs       []string
}

type actionInstallNodeBySSH struct {
	installerGroup        iface.FileGroup
	crypter               crypter.Crypter
	logger                logger.Logger
	storageNodeDeployment nodedeployment.IStorageNodeDeployment
	provider              discover.Provider
}

// Name returns the name of the action.
func (act *actionInstallNodeBySSH) Name() string {
	return ActionNameInstallNodeBySSH
}

// Version returns the version of the action.
func (act *actionInstallNodeBySSH) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionInstallNodeBySSH) Description() string {
	return "Use ssh to connect to the target machine, transfer files through sftp, and execute the installation command"
}

// Timeout returns the timeout of the action.
func (act *actionInstallNodeBySSH) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionInstallNodeBySSH) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionInstallNodeBySSH) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionInstallNodeBySSH) DelayFn() func() {
	return func() {
		time.Sleep(5 * time.Second) // nolint: mnd
	}
}

// Do this func define what the action will do.
// To ensure readability, this action uses fmt.Sprintf to concatenate characters.
// nolint: perfsprint,funlen,fnsize
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (act *actionInstallNodeBySSH) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActParamInstallAgentBySSH)
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

	// ensure the workspace dir.
	if result, err := client.RunCommand("mkdir -p " + info.InstallerWorkspace); err != nil {
		err = fmt.Errorf("failed to mkdir -p %s , result(%s), err: %w", info.InstallerWorkspace, result, err)

		return err
	}

	// select matching tools, and use sftp to transfer it.
	toolName, err := tool.FormatInstallerName(info.Dynamic.NodeOsType, info.Dynamic.NodeCPUArch)
	if err != nil {
		err = fmt.Errorf("failed to format tools name, err: %w", err)

		return err
	}

	toolFile, err := act.installerGroup.GetFile(ctx.Ctx, toolName)
	if err != nil {
		err = fmt.Errorf("failed to get file, err: %w", err)

		return err
	}

	reader, err := toolFile.Content(ctx.Ctx)
	if err != nil {
		err = fmt.Errorf("failed to get file content, err: %w", err)

		return err
	}

	installerPath := path.Clean(path.Join(info.InstallerWorkspace, toolName))
	if err := client.TransferFile(reader, installerPath); err != nil {
		return fmt.Errorf("failed to transfer file, err: %w", err)
	}

	// make sure tool is executable
	if result, err := client.RunCommand("chmod +x " + installerPath); err != nil {
		err = fmt.Errorf("failed to chmod +x, result(%s), err: %w", result, err)

		return err
	}

	randSelector := discover.NewRandomSelector()
	downloadEndpoint, err := act.provider.GetEndpoint(
		discover.ServiceNameFile,
		discover.EndpointNameFileBasic,
		randSelector)
	if err != nil {
		return fmt.Errorf("failed to get file endpoint, err: %w", err)
	}

	callbackEndpoint, err := act.provider.GetEndpoint(
		discover.ServiceNameBackend,
		discover.EndpointNameBackendCallback,
		randSelector)
	if err != nil {
		return fmt.Errorf("failed to get backend callback endpoint, err: %w", err)
	}

	deployConstant, err := deployconstant.GetDeployConf(info.Dynamic.NodeGeneration, info.Dynamic.NodeOsType)
	if err != nil {
		return fmt.Errorf("failed to get deploy constant, err: %w", err)
	}

	installParams := &InstallParams{
		InstallerPath:      installerPath,
		NodeRole:           info.Host.Dynamic.NodeRole,
		CallbackEndpoint:   "http://" + callbackEndpoint.GetIPV4Address(),
		DownloadEndpoint:   "http://" + downloadEndpoint.GetIPV4Address(),
		PkgVersion:         info.Dynamic.NodeVersion,
		PkgGeneration:      info.Dynamic.NodeGeneration,
		GseRoot:            deployConstant.GseHomeDir,
		Token:              param.Token,
		OperInstID:         ctx.Data.OperationInstanceID,
		InstallerWorkspace: info.InstallerWorkspace,
		AdditionArgs: []string{
			"--reinstall",
		},
	}

	if !info.ReRegister && info.Dynamic.AgentID != "" {
		installParams.AdditionArgs = append(installParams.AdditionArgs,
			fmt.Sprintf("--agent_id %s", info.Dynamic.AgentID))
	}

	// exec install command
	installCmd := act.buildCMD(installParams)

	ctx.Data.LogI(fmt.Sprintf("install node cmd: %s", installCmd))

	outStr, err := client.RunCommand(fmt.Sprintf(
		`mkdir -p %s && cd %s && echo "%s" > install.sh && sh install.sh`,
		info.InstallerWorkspace,
		info.InstallerWorkspace,
		installCmd),
	)
	if err != nil {
		err = fmt.Errorf("failed to run install node, err: %w", err)

		return err
	}
	ctx.Data.LogI(fmt.Sprintf("install node result: %s", outStr))

	return nil
}

// To ensure readability, this action uses fmt.Sprintf to concatenate characters.
// nolint: perfsprint
func (act *actionInstallNodeBySSH) buildCMD(param *InstallParams) string {
	args := []string{
		fmt.Sprintf(`--node_role "%s"`, param.NodeRole),
		fmt.Sprintf(`--callback_endpoint "%s"`, param.CallbackEndpoint),
		fmt.Sprintf(`--download_endpoint "%s"`, param.DownloadEndpoint),
		fmt.Sprintf(`--pkg_version "%s"`, param.PkgVersion),
		fmt.Sprintf(`--pkg_generation "%d"`, param.PkgGeneration),
		fmt.Sprintf(`--gse_root "%s"`, param.GseRoot),
		fmt.Sprintf(`--token "%s"`, param.Token),
		fmt.Sprintf(`--oper_inst_id "%s"`, param.OperInstID),
		fmt.Sprintf(`--workspace "%s"`, param.InstallerWorkspace),
		fmt.Sprintf(`--deploy_env "%s"`, system.GetEnv()),
	}
	if len(param.AdditionArgs) > 0 {
		args = append(args, param.AdditionArgs...)
	}

	installCmd := fmt.Sprintf("%s %s", param.InstallerPath, strings.Join(args, " "))

	installLogPath := path.Clean(fmt.Sprintf("%s.stdout", param.InstallerPath))
	installCmd = fmt.Sprintf("%s >%s 2>&1 &", installCmd, installLogPath)

	return installCmd
}
