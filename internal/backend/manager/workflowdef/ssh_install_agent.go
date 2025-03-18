/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package workflowdef ...
package workflowdef

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/criteria/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/crypter"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/sshx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operengine"
)

// NewActionInstallAgentBySSH ...
func NewActionInstallAgentBySSH(toolsGroup iface.FileGroup, crypter crypter.Crypter, logger logger.Logger,
) *InstallAgentBySSH {

	return &InstallAgentBySSH{
		toolsGroup: toolsGroup,
		crypter:    crypter,
		logger:     logger,
	}
}

// InstallAgentParamBySSH ...
type InstallAgentParamBySSH struct {
	sshHostInfo     `json:",inline"`
	installerParams `json:",inline"`
}

type sshHostInfo struct {
	IP       string `json:"ip"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password []byte `json:"passwd"`
}

type installerParams struct {
	NodeType         gse.NodeType `json:"node_type"`
	CallbackEndpoint string       `json:"callback_endpoint"`
	DownloadEndpoint string       `json:"download_endpoint"`
	PkgVersion       string       `json:"pkg_version"`
	PkgGeneration    int          `json:"pkg_generation"`
	InstallEnv       string       `json:"install_env"`
	Token            string       `json:"token"`
	TmpDir           string       `json:"tmp_dir"`
	AdditionArgs     []string     `json:"addition_args"`
}

// InstallAgentBySSH ...
type InstallAgentBySSH struct {
	toolsGroup iface.FileGroup
	crypter    crypter.Crypter
	logger     logger.Logger
}

// Name returns the name of the action.
func (action *InstallAgentBySSH) Name() string {
	return "install_agent_by_ssh"
}

// Version returns the version of the action.
func (action *InstallAgentBySSH) Version() string {
	return "1.0.0"
}

// Description returns the description of the action.
func (action *InstallAgentBySSH) Description() string {
	return "Use ssh to connect to the target machine, transfer files through sftp, and execute the installation command"
}

// Timeout returns the timeout of the action.
func (action *InstallAgentBySSH) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (action *InstallAgentBySSH) Tags() []operengine.ActionTag {
	return []operengine.ActionTag{}
}

// MaxRetryCount returns the max retry count of the action.
func (action *InstallAgentBySSH) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (action *InstallAgentBySSH) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
// To ensure readability, this action uses fmt.Sprintf to concatenate characters.
// nolint: perfsprint
func (action *InstallAgentBySSH) Do(ctx *operengine.ActionInstContext) error {
	param := new(InstallAgentParamBySSH)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		err = fmt.Errorf("failed to convert param, err: %w", err)

		return err
	}

	passwd, err := action.crypter.Decrypt(param.Password)
	if err != nil {
		err = fmt.Errorf("failed to decrypt password, err: %w", err)

		return err
	}

	client, err := sshx.NewClient(ctx.Ctx, &sshx.Config{
		Network:  sshx.NetworkTCP,
		IP:       param.IP,
		Port:     param.Port,
		User:     param.User,
		Password: string(passwd),
		Logger:   action.logger,
	}, sshx.DefaultTimeout)
	if err != nil {
		err = fmt.Errorf("failed to connect to host, host(%s), err: %w",
			fmt.Sprintf("%s:%d", param.IP, param.Port), err)

		return err
	}

	osType, cpuArch, targetDir, err := action.detectInfo(ctx, client)
	if err != nil {
		return err
	}
	if param.TmpDir == "" {
		param.TmpDir = targetDir
	}

	// 4. select matching installer, and use sftp to transfer it.
	toolName := fmt.Sprintf("installer-%s-%s", osType, cpuArch)
	tool, err := action.toolsGroup.GetFile(toolName)
	if err != nil {
		err = fmt.Errorf("failed to get file, err: %w", err)

		return err
	}

	reader, err := tool.Content()
	if err != nil {
		err = fmt.Errorf("failed to get file content, err: %w", err)

		return err
	}

	installerPath := path.Clean(path.Join(targetDir, toolName))
	if err := client.TransferFile(reader, installerPath); err != nil {
		err = fmt.Errorf("failed to transfer file, err: %w", err)

		return err
	}

	// 6. make sure tool is executable
	if result, err := client.RunCommand("chmod +x " + installerPath); err != nil {
		err = fmt.Errorf("failed to chmod +x, result(%s), err: %w", result, err)

		return err
	}

	// 7. exec install command
	installCmd := action.buildCMD(installerPath, param.installerParams, ctx.Data.OperInstID)

	outStr, err := client.RunCommand(installCmd)
	if err != nil {
		err = fmt.Errorf("failed to run install agent, err: %w", err)

		return err
	}

	ctx.Data.Log(outStr)

	return nil
}

// inorder to improve readability, use fmt.Sprintf to construct command line, and use named return.
// nolint: nonamedreturns,perfsprint
func (action *InstallAgentBySSH) detectInfo(ctx *operengine.ActionInstContext, client *sshx.Client) (
	osType string, cpuArch string, targetDir string, err error) {

	// 1. detect target system
	osType, err = client.RunCommand("uname -s")
	if err != nil {
		err = fmt.Errorf("failed to run uname -a, err: %w", err)

		return "", "", "", err
	}
	osType = strings.TrimFunc(strings.ToLower(osType), func(r rune) bool {
		return r == '\n'
	})
	switch osType {
	case "linux", "darwin":
	default:
		err = fmt.Errorf("unsupported os type, os-type(%s)", osType)

		return "", "", "", err
	}
	ctx.Data.Log(fmt.Sprintf("host os info: %s", osType))

	// 2. detect target cpu arch
	cpuArch, err = client.RunCommand("uname -m")
	if err != nil {
		return "", "", "", fmt.Errorf("failed to run uname -m, err: %w", err)
	}
	cpuArch = strings.TrimFunc(strings.ToLower(cpuArch), func(r rune) bool {
		return r == '\n'
	})
	ctx.Data.Log(fmt.Sprintf("host cpu arch: %s", cpuArch))

	// 3. detect target dir
	targetDir, err = client.RunCommand("pwd")
	if err != nil {
		err = fmt.Errorf("failed to run pwd, err: %w", err)

		return "", "", "", err
	}
	targetDir = strings.TrimFunc(targetDir, func(r rune) bool {
		return r == '\n'
	})
	ctx.Data.Log(fmt.Sprintf("target dir: %s", targetDir))

	return osType, cpuArch, targetDir, nil
}

// To ensure readability, this action uses fmt.Sprintf to concatenate characters.
// nolint: perfsprint
func (action *InstallAgentBySSH) buildCMD(installerPath string, param installerParams, operInstID string) string {
	args := []string{
		fmt.Sprintf("--node_type %s", param.NodeType),
		fmt.Sprintf("--callback_endpoint %s", param.CallbackEndpoint),
		fmt.Sprintf("--download_endpoint %s", param.DownloadEndpoint),
		fmt.Sprintf("--pkg_version %s", param.PkgVersion),
		fmt.Sprintf("--pkg_generation %d", param.PkgGeneration),
		fmt.Sprintf("--install_env %s", param.InstallEnv),
		fmt.Sprintf("--token %s", param.Token),
		fmt.Sprintf("--oper_inst_id %s", operInstID),
		fmt.Sprintf("--tmp_dir %s", param.TmpDir),
	}
	if len(param.AdditionArgs) > 0 {
		args = append(args, param.AdditionArgs...)
	}

	installCmd := fmt.Sprintf("%s %s", installerPath, strings.Join(args, " "))

	installLogPath := path.Clean(path.Join(param.TmpDir, "install.log"))
	installCmd = fmt.Sprintf("%s >%s 2>&1 &", installCmd, installLogPath)

	return installCmd
}
