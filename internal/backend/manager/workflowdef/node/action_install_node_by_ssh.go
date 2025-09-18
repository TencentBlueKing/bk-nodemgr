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
	"path"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/common"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/node/utils"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/credit"
	nodeStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/creditvault"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/deployconstant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/discover"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/tool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/installer"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/sshx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/system"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameInstallNodeBySSH defines the action name.
	ActionNameInstallNodeBySSH = "install_node_by_ssh"
)

// NewActionInstallNodeBySSH get a new action.
func NewActionInstallNodeBySSH(
	installerFileGroup fileiface.FileGroup,
	logger logger.ILogger,
	storageNodeDeployment nodeStg.IDaoNodeDeployment,
	provider discover.Provider,
	storageHostCredit credit.IStorageHostCredit,
	passwordVault creditvault.IHostPasswordVault,
) action.Definition {

	return &actionInstallNodeBySSH{
		installerGroup:        installerFileGroup,
		storageHostCredit:     storageHostCredit,
		logger:                logger,
		storageNodeDeployment: storageNodeDeployment,
		provider:              provider,
		passwordVault:         passwordVault,
	}
}

// ActParamInstallAgentBySSH ...
type ActParamInstallAgentBySSH struct {
	utils.NodeActionStandardParam `json:",inline"`
}

// InstallParams this struct defines the parameters for installing agent.
type InstallParams struct {
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

type actionInstallNodeBySSH struct {
	logger logger.ILogger

	installerGroup fileiface.FileGroup

	storageHostCredit     credit.IStorageHostCredit
	storageNodeDeployment nodeStg.IDaoNodeDeployment
	provider              discover.Provider
	passwordVault         creditvault.IHostPasswordVault
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
// nolint: perfsprint,funlen
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (act *actionInstallNodeBySSH) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActParamInstallAgentBySSH)
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
	std.DeployInfo().BlockingActionName = common.ActionNameWaitInstallerComplete

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

	// ensure the workspace dir.
	if err = act.ensureWorkspace(std, client); err != nil {
		return fmt.Errorf("failed to ensure workspace through ssh: %w", err)
	}

	// ensure the installer tool.
	installerPath, err := act.ensureInstallerTool(std, client)
	if err != nil {
		return fmt.Errorf("failed to ensure installer tool through ssh: %w", err)
	}

	// execute install cmd.
	if err := act.executeInstallCMD(std, client, installerPath); err != nil {
		return fmt.Errorf("failed to execute install cmd: %w", err)
	}

	return nil
}

func (act *actionInstallNodeBySSH) ensureWorkspace(std *utils.NodeActionStandarder, client *sshx.Client) error {
	if result, err := client.RunCommand("mkdir -p " + std.DeployInfo().InstallerWorkDir); err != nil {
		err = fmt.Errorf("failed to run command. command(mkdir -p %s), result(%s): %w", std.DeployInfo().InstallerWorkDir, result, err)

		return err
	}

	return nil
}

func (act *actionInstallNodeBySSH) ensureInstallerTool(std *utils.NodeActionStandarder, client *sshx.Client) (string, error) {
	// select matching tools, and use sftp to transfer it.
	toolName, err := tool.FormatInstallerName(std.DeployInfo().Host.Dynamic.NodeOsType, std.DeployInfo().Host.Dynamic.NodeCPUArch)
	if err != nil {
		return "", fmt.Errorf("failed to format installer tool name: %w", err)
	}

	toolFile, err := act.installerGroup.GetFile(std.Context(), toolName)
	if err != nil {
		return "", fmt.Errorf("failed to get installer tool file from local: %w", err)
	}

	reader, err := toolFile.Content(std.Context())
	if err != nil {
		return "", fmt.Errorf("failed to get installer tool file content: %w", err)
	}

	installerPath := path.Clean(path.Join(std.DeployInfo().InstallerWorkDir, toolName))
	if err = client.TransferFile(reader, installerPath); err != nil {
		return "", fmt.Errorf("failed to transfer installer tool to host: %w", err)
	}
	std.InstanceData().LogI(fmt.Sprintf("transferred file to host, path(%s)", installerPath))

	// make sure tool is executable
	if result, err := client.RunCommand("chmod +x " + installerPath); err != nil {
		return "", fmt.Errorf("failed to run command. command(chmod +x %s), result(%s): %w", installerPath, result, err)
	}

	return installerPath, nil
}

func (act *actionInstallNodeBySSH) executeInstallCMD(std *utils.NodeActionStandarder, client *sshx.Client, installerPath string) error {
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

	installParams := &InstallParams{
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

	installCmd := act.buildCMD(installParams)
	std.InstanceData().LogI(fmt.Sprintf("install node cmd: %s", installCmd))

	// exec install command.
	outStr, err := client.RunCommand(fmt.Sprintf(
		`mkdir -p %s && cd %s && echo "%s" > install.sh && sh install.sh`,
		std.DeployInfo().InstallerWorkDir,
		std.DeployInfo().InstallerWorkDir,
		installCmd),
	)
	if err != nil {
		err = fmt.Errorf("failed to run install node: %w", err)

		return err
	}

	std.InstanceData().LogI(fmt.Sprintf("install node result: %s", outStr))

	return nil
}

// To ensure readability, this action uses fmt.Sprintf to concatenate characters.
// nolint: perfsprint
func (act *actionInstallNodeBySSH) buildCMD(param *InstallParams) string {
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

	installCmd := fmt.Sprintf("%s %s %s", param.InstallerPath, installer.NodeCmdFullInstall, strings.Join(args, " "))

	installLogPath := path.Clean(fmt.Sprintf("%s.stdout", param.InstallerPath))
	installCmd = fmt.Sprintf("%s >%s 2>&1 &", installCmd, installLogPath)

	return installCmd
}
