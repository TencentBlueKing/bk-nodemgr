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
	"context"
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/nodedeployment"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/deployconstant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/tool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/crypter"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/discover"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/sshx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

// NewActionInstallNodeBySSH ...
func NewActionInstallNodeBySSH(
	installerFileGroup iface.FileGroup,
	crypter crypter.Crypter,
	logger logger.Logger,
	iDaoNodeDeployment nodedeployment.IDaoNodeDeployment,
	provider discover.Provider,
) action.Definition {

	return &InstallNodeBySSH{
		installerGroup:     installerFileGroup,
		crypter:            crypter,
		logger:             logger,
		iDaoNodeDeployment: iDaoNodeDeployment,
		provider:           provider,
	}
}

// InstallAgentParamBySSH ...
type InstallAgentParamBySSH struct {
	Token string `json:"token"`
}

// InstallParams this struct defines the parameters for installing agent.
type InstallParams struct {
	InstallerPath    string
	NodeRole         types.NodeRole
	CallbackEndpoint string
	DownloadEndpoint string
	PkgVersion       string
	PkgGeneration    types.NodeGeneration
	GseRoot          string
	Token            string
	TmpDir           string
	AdditionArgs     []string
}

// InstallNodeBySSH ...
type InstallNodeBySSH struct {
	installerGroup     iface.FileGroup
	crypter            crypter.Crypter
	logger             logger.Logger
	iDaoNodeDeployment nodedeployment.IDaoNodeDeployment
	provider           discover.Provider
}

// Name returns the name of the action.
func (act *InstallNodeBySSH) Name() string {
	return ActionNameInstallNodeBySSH
}

// Version returns the version of the action.
func (act *InstallNodeBySSH) Version() string {
	return "1.0.0"
}

// Description returns the description of the action.
func (act *InstallNodeBySSH) Description() string {
	return "Use ssh to connect to the target machine, transfer files through sftp, and execute the installation command"
}

// Timeout returns the timeout of the action.
func (act *InstallNodeBySSH) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *InstallNodeBySSH) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *InstallNodeBySSH) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *InstallNodeBySSH) DelayFn() func() {
	return func() {
		time.Sleep(5 * time.Second) // nolint: mnd
	}
}

// Do this func define what the action will do.
// To ensure readability, this action uses fmt.Sprintf to concatenate characters.
// nolint: perfsprint
func (act *InstallNodeBySSH) Do(ctx *action.InstanceContext) (err error) {
	param := new(InstallAgentParamBySSH)
	err = conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		err = fmt.Errorf("failed to convert param, err: %w", err)

		return err
	}

	info, err := act.iDaoNodeDeployment.GetInfo(ctx.Ctx, param.Token)
	if err != nil {
		return err
	}

	defer func() {
		if storeErr := act.iDaoNodeDeployment.UpdateInfo(ctx.Ctx, param.Token, info); storeErr != nil {
			err = errors.Join(storeErr, err)
		}
	}()

	client, err := act.buildSSH(ctx.Ctx, info)
	if err != nil {
		return err
	}

	osType, cpuArch, targetDir, err := act.detectInfo(ctx, client)
	if err != nil {
		return err
	}
	if info.TmpDir == "" {
		info.TmpDir = targetDir
	}

	// 4. select matching tools, and use sftp to transfer it.
	toolName, err := tool.FormatInstallerName(osType, cpuArch)
	if err != nil {
		err = fmt.Errorf("failed to format tools name, err: %w", err)

		return err
	}

	toolFile, err := act.installerGroup.GetFile(toolName)
	if err != nil {
		err = fmt.Errorf("failed to get file, err: %w", err)

		return err
	}

	reader, err := toolFile.Content()
	if err != nil {
		err = fmt.Errorf("failed to get file content, err: %w", err)

		return err
	}

	installerPath := path.Clean(path.Join(targetDir, toolName))
	if err := client.TransferFile(reader, installerPath); err != nil {
		return fmt.Errorf("failed to transfer file, err: %w", err)
	}

	// 6. make sure tool is executable
	if result, err := client.RunCommand("chmod +x " + installerPath); err != nil {
		err = fmt.Errorf("failed to chmod +x, result(%s), err: %w", result, err)

		return err
	}

	randSelector := discover.NewRandomSelector()
	downloadEndpoint, err := act.provider.GetEndpoint(discover.ServiceNameFile, discover.EndpointNameFileBasic, randSelector)
	if err != nil {
		return fmt.Errorf("failed to get file endpoint, err: %w", err)
	}

	callbackEndpoint, err := act.provider.GetEndpoint(discover.ServiceNameBackend, discover.EndpointNameBackendCallback, randSelector)
	if err != nil {
		return fmt.Errorf("failed to get backend callback endpoint, err: %w", err)
	}

	deployConstant, err := deployconstant.GetDeployConf(info.Dynamic.NodeGeneration, osType)
	if err != nil {
		return fmt.Errorf("failed to get deploy constant, err: %w", err)
	}

	installParams := &InstallParams{
		InstallerPath:    installerPath,
		NodeRole:         info.Host.Dynamic.NodeRole,
		CallbackEndpoint: "http://" + callbackEndpoint.GetIPV4Address(),
		DownloadEndpoint: "http://" + downloadEndpoint.GetIPV4Address(),
		PkgVersion:       info.Dynamic.NodeVersion,
		PkgGeneration:    info.Dynamic.NodeGeneration,
		GseRoot:          deployConstant.GseHomeDir,
		Token:            param.Token,
		TmpDir:           info.TmpDir,
		AdditionArgs: []string{
			"--reinstall",
		},
	}

	if installParams.TmpDir == "" {
		installParams.TmpDir = targetDir
	}

	// 7. exec install command
	installCmd := action.buildCMD(installParams)
	ctx.Data.Log(fmt.Sprintf("install node cmd: %s", installCmd))

	outStr, err := client.RunCommand(installCmd)
	if err != nil {
		err = fmt.Errorf("failed to run install node, err: %w", err)

		return err
	}
	ctx.Data.Log(fmt.Sprintf("install node result: %s", outStr))

	return nil
}

func (act *InstallNodeBySSH) buildSSH(ctx context.Context, info *types.DeploymentInfo) (*sshx.Client, error) {
	sshConf := &sshx.Config{
		Network: sshx.NetworkTCP,
		IP:      info.LoginIP,
		Port:    int(info.LoginPort),
		User:    info.LoginUser,
		Logger:  act.logger,
	}

	switch info.LoginMode {
	case types.LoginModePassword:
		passwd, err := act.crypter.Decrypt(info.LoginPassword)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt password, err: %w", err)
		}

		sshConf.AuthMethod = sshx.AuthMethodPassword
		sshConf.Password = string(passwd)

	case types.LoginModeKeyFile:
		privateKey, err := act.crypter.Decrypt(info.LoginKeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt private key, err: %w", err)
		}

		sshConf.AuthMethod = sshx.AuthMethodPrivateKey
		sshConf.PrivateKey = privateKey
	case types.LoginModeNone:
		sshConf.AuthMethod = sshx.AuthMethodNone
	default:
		return nil, fmt.Errorf("unsupported login mode, mode(%s)", info.LoginMode)
	}

	client, err := sshx.NewClient(ctx, sshConf, sshx.DefaultTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to host, host(%s), err: %w",
			fmt.Sprintf("%s:%d", info.LoginIP, info.LoginPort), err)
	}

	return client, nil
}

// inorder to improve readability, use fmt.Sprintf to construct command line, and use named return.
// nolint: nonamedreturns,perfsprint
func (act *InstallNodeBySSH) detectInfo(ctx *action.InstanceContext, client *sshx.Client) (
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
	osType, err = platform.NormalizeOS(osType)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to detect info, err: %w", err)
	}

	switch osType {
	case criteria.OSLinux, criteria.OSDarwin:
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
	cpuArch, err = platform.NormalizeArch(cpuArch)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to detect info, err: %w", err)
	}

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
func (act *InstallNodeBySSH) buildCMD(param *InstallParams) string {
	args := []string{
		fmt.Sprintf("--node_role %s", param.NodeRole),
		fmt.Sprintf("--callback_endpoint %s", param.CallbackEndpoint),
		fmt.Sprintf("--download_endpoint %s", param.DownloadEndpoint),
		fmt.Sprintf("--pkg_version %s", param.PkgVersion),
		fmt.Sprintf("--pkg_generation %d", param.PkgGeneration),
		fmt.Sprintf("--gse_root %s", param.GseRoot),
		fmt.Sprintf("--token %s", param.Token),
		fmt.Sprintf("--tmp_dir %s", param.TmpDir),
	}
	if len(param.AdditionArgs) > 0 {
		args = append(args, param.AdditionArgs...)
	}

	installCmd := fmt.Sprintf("%s %s", param.InstallerPath, strings.Join(args, " "))

	installLogPath := path.Clean(fmt.Sprintf("%s.stdout", param.InstallerPath))
	installCmd = fmt.Sprintf("%s >%s 2>&1 &", installCmd, installLogPath)

	return installCmd
}
