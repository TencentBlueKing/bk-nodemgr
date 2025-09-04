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

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/nodeinstall/utils"
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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/wmix"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameInstallNodeByWMI defines the action name.
	ActionNameInstallNodeByWMI = "install_node_by_wmi"
)

// NewActionInstallNodeByWMI get a new action.
func NewActionInstallNodeByWMI(
	installerFileGroup fileiface.FileGroup,
	logger logger.ILogger,
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
	utils.NodeActionStandardParam `json:",inline"`
}

// InstallParamsWin this struct defines the parameters for installing agent.
type InstallParamsWin struct {
	InstallerPath   string
	Generation      types.Generation
	NodeRole        types.NodeRole
	CallbackSvrAddr string
	DownloadSvrAddr string
	NodeVersion     string
	DeployToken     string
	OperInstID      string
	BaseWorkDir     string
	BaseDeployDir   string
	AdditionArgs    []string
}

type actionInstallNodeByWMI struct {
	installerGroup        fileiface.FileGroup
	logger                logger.ILogger
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
// nolint: perfsprint,funlen
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (act *actionInstallNodeByWMI) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActParamInstallAgentByWMI)
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

	// let the callback server known which action to mark and log.
	std.DeployInfo().BlockingActionName = ActionNameWaitInstallerComplete

	// get wmi credit.
	credit := utils.NewCreditHandler(act.storageHostCredit, act.passwordVault)
	cMethod, cKey, err := credit.GetWMICredit(std)
	if err != nil {
		return fmt.Errorf("failed to get wmi credit: %w", err)
	}

	// generate the wmi client.
	client, err := wmix.NewClient(&wmix.Config{
		IP:         std.DeployInfo().Host.Dynamic.LoginIP,
		User:       std.DeployInfo().Host.Dynamic.LoginUser,
		Logger:     act.logger,
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
		return fmt.Errorf("failed to generate new wmi client: %w", err)
	}

	// ensure the workspace dir.
	if err = act.ensureWorkspace(std, client); err != nil {
		return fmt.Errorf("failed to ensure workspace through wmi: %w", err)
	}

	// ensure the installer tool.
	installerPath, err := act.ensureInstallerTool(std, client)
	if err != nil {
		return fmt.Errorf("failed to ensure installer tool through wmi: %w", err)
	}

	// execute install cmd.
	if err := act.executeInstallCMD(std, client, installerPath); err != nil {
		return fmt.Errorf("failed to execute install cmd: %w", err)
	}

	return nil
}

func (act *actionInstallNodeByWMI) ensureWorkspace(std *utils.NodeActionStandarder, client *wmix.Client) error {
	stdout, stderr, err := client.RunCommand(std.Context(), "mkdir "+std.DeployInfo().InstallerWorkDir)
	if err != nil {
		err = fmt.Errorf("failed to run command. command(mkdir %s), stdout(%s), stderr(%s): %w",
			std.DeployInfo().InstallerWorkDir, stdout, stderr, err)

		return err
	}

	std.InstanceData().LogI(fmt.Sprintf("make sure the installer workspace exists, stdout(%s), stderr(%s)",
		strings.Split(strings.TrimSpace(stdout), "\n"), strings.Split(strings.TrimSpace(stderr), "\n")))

	return nil
}

func (act *actionInstallNodeByWMI) ensureInstallerTool(std *utils.NodeActionStandarder, client *wmix.Client) (string, error) {
	// select matching tools, and use sftp to transfer it.
	toolName, err := tool.FormatInstallerName(std.DeployInfo().Host.Dynamic.NodeOsType, std.DeployInfo().Host.Dynamic.NodeCPUArch)
	if err != nil {
		return "", fmt.Errorf("failed to format installer tool name: %w", err)
	}

	toolFile, err := act.installerGroup.GetFile(std.Context(), toolName)
	if err != nil {
		return "", fmt.Errorf("failed to get installer tool file from local: %w", err)
	}

	tmpInstallFilePath := local.GetLocalFileAbsFilePath(toolFile)
	stdout, stderr, err := client.UploadFile(std.Context(), tmpInstallFilePath, std.DeployInfo().InstallerWorkDir)
	if err != nil {
		return "", fmt.Errorf("failed to transfer installer tool to host, stdout(%s), stderr(%s): %w",
			stdout, stderr, err)
	}
	installerPath := winpath.Clean(winpath.Join(std.DeployInfo().InstallerWorkDir, toolName))
	std.InstanceData().LogI(fmt.Sprintf("transferred file to host, path(%s)", installerPath))

	return installerPath, nil
}

func (act *actionInstallNodeByWMI) executeInstallCMD(std *utils.NodeActionStandarder, client *wmix.Client, installerPath string) error {
	randSelector := discover.NewRandomSelector()
	downloadSvrEndpoint, err := act.provider.GetEndpoint(
		discover.ServiceNameFile,
		discover.EndpointNameFileDownload,
		randSelector)
	if err != nil {
		return fmt.Errorf("failed to get file endpoint: %w", err)
	}

	callbackSvrEndpoint, err := act.provider.GetEndpoint(
		discover.ServiceNameBackend,
		discover.EndpointNameBackendCallback,
		randSelector)
	if err != nil {
		return fmt.Errorf("failed to get backend callback endpoint: %w", err)
	}

	deployConstant, err := deployconstant.GetNodeDeployConf(std.DeployInfo().Host.Dynamic.NodeGeneration, std.DeployInfo().Host.Dynamic.NodeOsType)
	if err != nil {
		return fmt.Errorf("failed to get deploy constant: %w", err)
	}

	installParams := &InstallParamsWin{
		NodeVersion:     std.DeployInfo().Host.Dynamic.NodeVersion,
		Generation:      std.DeployInfo().Host.Dynamic.NodeGeneration,
		InstallerPath:   installerPath,
		NodeRole:        std.DeployInfo().Host.Dynamic.NodeRole,
		CallbackSvrAddr: "http://" + callbackSvrEndpoint.GetIPV4Address(),
		DownloadSvrAddr: "http://" + downloadSvrEndpoint.GetIPV4Address(),
		DeployToken:     std.Token(),
		OperInstID:      std.InstanceData().OperationInstanceID,
		BaseWorkDir:     deployConstant.BaseWorkDir,
		BaseDeployDir:   deployConstant.BaseDeployDir,
	}

	if !std.DeployInfo().InstallOptions.ReRegister && std.DeployInfo().Host.Dynamic.AgentID != "" {
		installParams.AdditionArgs = append(installParams.AdditionArgs,
			fmt.Sprintf("--agent_id %s", std.DeployInfo().Host.Dynamic.AgentID))
	}

	installBat := act.buildBat(installParams)
	std.InstanceData().LogI(fmt.Sprintf("install node cmd: %s", installBat))

	// exec install command.
	tmpInstallBat, err := tmp.NewTempFileWithSpecialName(io.NopCloser(strings.NewReader(installBat)), installBatName)
	if err != nil {
		return fmt.Errorf("failed to create temp bat file for wmi execution: %w", err)
	}
	defer tmp.Clean()

	_, _, err = client.UploadFile(std.Context(), tmpInstallBat.Path(), std.DeployInfo().InstallerWorkDir)
	if err != nil {
		return fmt.Errorf("failed to transfer bat file for wmi execution: %w", err)
	}

	installCMD := winpath.Clean(winpath.Join(std.DeployInfo().InstallerWorkDir, installBatName))
	stdout, stderr, err := client.RunSilentCommand(std.Context(), installCMD)
	if err != nil {
		err = fmt.Errorf("failed to run install node, err: %w", err)

		return err
	}

	std.InstanceData().LogI(fmt.Sprintf("install node stdout: %s", strings.Split(strings.TrimSpace(stdout), "\n")))
	std.InstanceData().LogI(fmt.Sprintf("install node stderr: %s", strings.Split(strings.TrimSpace(stderr), "\n")))

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
		fmt.Sprintf("--dlsvr_addr %s", param.DownloadSvrAddr),
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
