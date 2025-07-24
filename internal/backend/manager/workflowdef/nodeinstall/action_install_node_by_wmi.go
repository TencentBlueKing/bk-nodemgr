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
	"io"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/credit"
	nodedeployment "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node-deployment"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/creditvault"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/deployconstant"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/local"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/tool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/discover"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/system"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tmp"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/winpath"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameInstallNodeByWMI defines the action name.
	ActionNameInstallNodeByWMI = "install_node_by_wmi"
)

// NewActionInstallNodeByWMI get a new action.
func NewActionInstallNodeByWMI(
	installerFileGroup fileiface.FileGroup,
	logger logger.Logger,
	storageNodeDeployment nodedeployment.IStorageNodeDeployment,
	provider discover.Provider,
	storageHostCredit credit.IStorageHostCredit,
	passwordVault creditvault.IHostPasswordVault,
) action.Definition {

	return &actionInstallNodeByWMI{
		installerGroup:        installerFileGroup,
		storageHostCredit:     storageHostCredit,
		logger:                logger,
		storageNodeDeployment: storageNodeDeployment,
		provider:              provider,
		passwordVault:         passwordVault,
	}
}

// ActParamInstallAgentByWMI ...
type ActParamInstallAgentByWMI struct {
	Token    string `json:"token"`
	Operator string `json:"operator"`
}

// InstallParamsWin this struct defines the parameters for installing agent.
type InstallParamsWin struct {
	InstallerPath   string
	Generation      types.Generation
	NodeRole        types.NodeRole
	CallbackSvrAddr string
	FileSvrAddr     string
	NodeVersion     string
	DeployToken     string
	OperInstID      string
	BaseWorkDir     string
	BaseDeployDir   string
	AdditionArgs    []string
}

type actionInstallNodeByWMI struct {
	installerGroup        fileiface.FileGroup
	logger                logger.Logger
	storageHostCredit     credit.IStorageHostCredit
	storageNodeDeployment nodedeployment.IStorageNodeDeployment
	provider              discover.Provider
	passwordVault         creditvault.IHostPasswordVault
}

// Name returns the name of the action.
func (act *actionInstallNodeByWMI) Name() string {
	return ActionNameInstallNodeByWMI
}

// Version returns the version of the action.
func (act *actionInstallNodeByWMI) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionInstallNodeByWMI) Description() string {
	return "Use wmi to connect to the target machine, transfer files through sftp, and execute the installation command"
}

// Timeout returns the timeout of the action.
func (act *actionInstallNodeByWMI) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionInstallNodeByWMI) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionInstallNodeByWMI) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionInstallNodeByWMI) DelayFn() func() {
	return func() {
		time.Sleep(5 * time.Second) // nolint: mnd
	}
}

const installBatName = "install.bat"

// Do this func define what the action will do.
// To ensure readability, this action uses fmt.Sprintf to concatenate characters.
// nolint: perfsprint,funlen,fnsize
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (act *actionInstallNodeByWMI) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActParamInstallAgentByWMI)
	err = conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		err = fmt.Errorf("failed to convert param, err: %w", err)

		return err
	}

	info, err := act.storageNodeDeployment.GetInfo(ctx.Ctx, param.Token)
	if err != nil {
		return err
	}
	// let the callback server known which action to mark and log.
	info.BlockingActionName = ActionNameWaitInstallerComplete

	defer func() {
		if storeErr := act.storageNodeDeployment.UpdateInfo(ctx.Ctx, param.Token, info); storeErr != nil {
			err = errors.Join(storeErr, err)
		}
	}()

	client, err := generateWMIClient(ctx.Ctx, param.Operator, act.logger, act.storageHostCredit, act.passwordVault, info)
	if err != nil {
		return err
	}

	stdOut, stdErr, err := client.RunCommand(ctx.Ctx, "mkdir "+info.InstallerWorkDir)
	if err != nil {
		err = fmt.Errorf("failed to run mkdir %s, err: %w", info.InstallerWorkDir, err)

		return err
	}

	ctx.Data.LogI(fmt.Sprintf("make sure the installer workspace exists, stdOut: %s, stdErr: %s",
		strings.Split(strings.TrimSpace(stdOut), "\n"), strings.Split(strings.TrimSpace(stdErr), "\n")))

	// select matching tools, and use sftp to transfer it.
	toolName, err := tool.FormatInstallerName(info.Host.Dynamic.NodeOsType, info.Host.Dynamic.NodeCPUArch)
	if err != nil {
		err = fmt.Errorf("failed to format tools name, err: %w", err)

		return err
	}

	toolFile, err := act.installerGroup.GetFile(ctx.Ctx, toolName)
	if err != nil {
		err = fmt.Errorf("failed to get file, err: %w", err)

		return err
	}

	if toolFile.FileObject() != fileiface.LocalFile {
		err = fmt.Errorf("installer file is not a local file, file-info(%v)", toolFile.Info())

		return err
	}

	tmpInstallFilePath := local.GetLocalFileAbsFilePath(toolFile)
	stdOut, stdErr, err = client.UploadFile(ctx.Ctx, tmpInstallFilePath, info.InstallerWorkDir)
	if err != nil {
		return fmt.Errorf("failed to transfer file, stdOut: %s, stdErr: %s err: %w",
			stdOut, stdErr, err)
	}

	installerPath := winpath.Clean(winpath.Join(info.InstallerWorkDir, toolName))

	ctx.Data.LogI(fmt.Sprintf("upload file to remote, path(%s)", installerPath))
	act.logger.Infof("upload file to remote, path(%s)", installerPath)

	ctx.Data.LogI(fmt.Sprintf("upload file stdout: %s, stdErr: %s",
		strings.Split(strings.TrimSpace(stdOut), "\n"), strings.Split(strings.TrimSpace(stdErr), "\n")))
	act.logger.Infof("upload file stdout: %s, stdErr: %s",
		strings.Split(strings.TrimSpace(stdOut), "\n"), strings.Split(strings.TrimSpace(stdErr), "\n"))

	randSelector := discover.NewRandomSelector()
	fileSvrEndpoint, err := act.provider.GetEndpoint(
		discover.ServiceNameFile,
		discover.EndpointNameFileBasic,
		randSelector)
	if err != nil {
		return fmt.Errorf("failed to get file endpoint, err: %w", err)
	}

	callbackSvrEndpoint, err := act.provider.GetEndpoint(
		discover.ServiceNameBackend,
		discover.EndpointNameBackendCallback,
		randSelector)
	if err != nil {
		return fmt.Errorf("failed to get backend callback endpoint, err: %w", err)
	}

	deployConstant, err := deployconstant.GetDeployConf(info.Host.Dynamic.NodeGeneration, info.Host.Dynamic.NodeOsType)
	if err != nil {
		return fmt.Errorf("failed to get deploy constant, err: %w", err)
	}

	installParams := &InstallParamsWin{
		NodeVersion:     info.Host.Dynamic.NodeVersion,
		Generation:      info.Host.Dynamic.NodeGeneration,
		InstallerPath:   installerPath,
		NodeRole:        info.Host.Dynamic.NodeRole,
		CallbackSvrAddr: "http://" + callbackSvrEndpoint.GetIPV4Address(),
		FileSvrAddr:     "http://" + fileSvrEndpoint.GetIPV4Address(),
		DeployToken:     param.Token,
		OperInstID:      ctx.Data.OperationInstanceID,
		BaseWorkDir:     deployConstant.BaseWorkDir,
		BaseDeployDir:   deployConstant.BaseDeployDir,
	}

	if !info.InstallOptions.ReRegister && info.Host.Dynamic.AgentID != "" {
		installParams.AdditionArgs = append(installParams.AdditionArgs,
			fmt.Sprintf("--agent_id %s", info.Host.Dynamic.AgentID))
	}

	// exec install command
	installBat := act.buildBat(installParams)
	ctx.Data.LogI(fmt.Sprintf("install-node-cmd(%s)", installBat))

	tmpInstallBat, err := tmp.NewTempFileWithSpecialName(io.NopCloser(strings.NewReader(installBat)), installBatName)
	if err != nil {
		return fmt.Errorf("failed to create temp file, err: %w", err)
	}
	defer tmp.Clean()

	_, _, err = client.UploadFile(ctx.Ctx, tmpInstallBat.Path(), info.InstallerWorkDir)
	if err != nil {
		return fmt.Errorf("failed to transfer file, err: %w", err)
	}

	installCMD := winpath.Clean(winpath.Join(info.InstallerWorkDir, installBatName))
	stdOutStr, stdErrStr, err := client.RunSilentCommand(ctx.Ctx, installCMD)
	if err != nil {
		err = fmt.Errorf("failed to run install node, err: %w", err)

		return err
	}

	ctx.Data.LogI(fmt.Sprintf("install node stdout: %s", strings.Split(strings.TrimSpace(stdOutStr), "\n")))
	ctx.Data.LogI(fmt.Sprintf("install node stderr: %s", strings.Split(strings.TrimSpace(stdErrStr), "\n")))

	return nil
}

// To ensure readability, this action uses fmt.Sprintf to concatenate characters.
// nolint: perfsprint
func (act *actionInstallNodeByWMI) buildBat(param *InstallParamsWin) string {
	args := []string{
		fmt.Sprintf("--deploy_env %s", system.GetEnv()),
		fmt.Sprintf("--generation %d", param.Generation),
		fmt.Sprintf("--node_role %s", param.NodeRole),
		fmt.Sprintf("--base_work_dir %s", param.BaseWorkDir),
		fmt.Sprintf("--base_deploy_dir %s", param.BaseDeployDir),
		fmt.Sprintf("--filesvr_addr %s", param.FileSvrAddr),
		fmt.Sprintf("--cbsvr_addr %s", param.CallbackSvrAddr),
		fmt.Sprintf("--deploy_token %s", param.DeployToken),
		fmt.Sprintf("--node_version %s", param.NodeVersion),
		fmt.Sprintf("--oper_inst_id %s", param.OperInstID),
	}
	if len(param.AdditionArgs) > 0 {
		args = append(args, param.AdditionArgs...)
	}
	installLogPath := winpath.Clean(fmt.Sprintf("%s.stdout", param.InstallerPath))

	installCmd := fmt.Sprintf("cd %s && %s full-install %s >%s 2>&1",
		winpath.Join(param.BaseWorkDir, system.GetEnv()), param.InstallerPath, strings.Join(args, " "), installLogPath)

	return installCmd
}
