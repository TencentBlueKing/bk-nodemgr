/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package installer

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/winpath"
)

const (
	// pluginV2CmdFullInstall defines the installer cmd.
	pluginV2CmdFullInstall = "pluginv2 full-install"

	// pluginV2CmdFullUpgrade defines the installer cmd.
	pluginV2CmdFullUpgrade = "pluginv2 full-upgrade"

	// pluginV2CmdFullUninstall defines the installer cmd.
	pluginV2CmdFullUninstall = "pluginv2 full-uninstall"
)

// pluginV2FlagName defines the plugin flag name.
type pluginV2FlagName string

// plugin flag name
const (
	// pluginV2FlagDeployEnv plugin flag name defines the deploy env.
	pluginV2FlagDeployEnv pluginV2FlagName = "deploy_env"

	// pluginV2FlagBaseDeployDir plugin flag name defines the base deploy dir.
	pluginV2FlagBaseDeployDir pluginV2FlagName = "base_deploy_dir"

	// pluginV2FlagBaseWorkDir plugin flag name defines the base work dir.
	pluginV2FlagBaseWorkDir pluginV2FlagName = "base_work_dir"

	// pluginV2FlagLogDir plugin flag name defines the log dir.
	pluginV2FlagLogDir pluginV2FlagName = "log_dir"

	// pluginV2FlagStatus plugin flag name defines the status.
	pluginV2FlagStatus pluginV2FlagName = "status"

	// pluginV2FlagPluginGroup plugin flag name defines the plugin group.
	pluginV2FlagPluginGroup pluginV2FlagName = "plugin_group"

	// pluginV2FlagPluginName plugin flag name defines the plugin name.
	pluginV2FlagPluginName pluginV2FlagName = "plugin_name"

	// pluginV2FlagPkgFile plugin flag name defines the pkg file.
	pluginV2FlagPkgFile pluginV2FlagName = "pkg_file"

	// pluginV2FlagDownloadSvrAdd plugin flag name defines the download server address.
	pluginV2FlagDownloadSvrAdd pluginV2FlagName = "dlsvr_addr"

	// pluginV2FlagCallbackSvrAdd plugin flag name defines the callback server address.
	pluginV2FlagCallbackSvrAdd pluginV2FlagName = "cbsvr_addr"

	// pluginV2FlagDeployToken plugin flag name defines the deploy token.
	pluginV2FlagDeployToken pluginV2FlagName = "deploy_token"

	// pluginV2FlagOperInstID plugin flag name defines the operation instance id.
	pluginV2FlagOperInstID pluginV2FlagName = "oper_inst_id"

	// pluginV2FlagPluginVersion plugin flag name defines the plugin version.
	pluginV2FlagPluginVersion pluginV2FlagName = "plugin_version"

	// pluginV2FlagPluginPkgName plugin flag name defines the plugin package name.
	pluginV2FlagPluginPkgName pluginV2FlagName = "plugin_pkg_name"
)

// PluginV2CommonParams defines the common params of installer.
type PluginV2CommonParams struct {
	InstallWorkDir    string
	InstallerFileName string

	BaseDeployDir string
	BaseWorkDir   string

	DeployEnv string
}

// Validate validates the common params.
func (params *PluginV2CommonParams) Validate() error {
	if params.BaseDeployDir == "" {
		return fmt.Errorf("base deploy dir is empty")
	}

	if params.BaseWorkDir == "" {
		return fmt.Errorf("base work dir is empty")
	}

	if params.InstallWorkDir == "" {
		return fmt.Errorf("install work dir is empty")
	}

	if params.InstallerFileName == "" {
		return fmt.Errorf("installer file name is empty")
	}

	if params.DeployEnv == "" {
		return fmt.Errorf("deploy env is empty")
	}

	return nil
}

// PluginV2InstallParams defines the install params.
type PluginV2InstallParams struct {
	PluginV2CommonParams

	PluginGroup   string
	PluginName    string
	PluginVersion string
	PluginPkgName string

	DownloadSvrAddr string
	CallbackSvrAddr string

	DeployToken string
	OperInstID  string

	SkipCallback bool
	SkipDownload bool
}

// Validate validates the install params.
func (params *PluginV2InstallParams) Validate() error {
	if err := params.PluginV2CommonParams.Validate(); err != nil {
		return err
	}

	if params.PluginGroup == "" {
		return fmt.Errorf("plugin group is empty")
	}

	if params.PluginName == "" {
		return fmt.Errorf("plugin name is empty")
	}

	if params.PluginVersion == "" {
		return fmt.Errorf("plugin version is empty")
	}

	if params.PluginPkgName == "" {
		return fmt.Errorf("plugin package name is empty")
	}

	// When skip_callback is set, callback server address can be empty
	if !params.SkipCallback && params.CallbackSvrAddr == "" {
		return fmt.Errorf("callback server address is empty")
	}

	// When skip_download is set, download server address can be empty
	if !params.SkipDownload && params.DownloadSvrAddr == "" {
		return fmt.Errorf("download server address is empty")
	}

	if params.DeployToken == "" {
		return fmt.Errorf("deploy token is empty")
	}

	if params.OperInstID == "" {
		return fmt.Errorf("operation instance id is empty")
	}

	return nil
}

func (params *PluginV2InstallParams) buildArgs() []string {
	args := []string{
		fmt.Sprintf("--%s %s", pluginV2FlagBaseDeployDir, params.BaseDeployDir),
		fmt.Sprintf("--%s %s", pluginV2FlagBaseWorkDir, params.BaseWorkDir),
		fmt.Sprintf("--%s %s", pluginV2FlagPluginGroup, params.PluginGroup),
		fmt.Sprintf("--%s %s", pluginV2FlagPluginName, params.PluginName),
		fmt.Sprintf("--%s %s", pluginV2FlagPluginVersion, params.PluginVersion),
		fmt.Sprintf("--%s %s", pluginV2FlagPluginPkgName, params.PluginPkgName),
		fmt.Sprintf("--%s %s", pluginV2FlagDeployEnv, params.DeployEnv),
		fmt.Sprintf("--%s %s", pluginV2FlagDeployToken, params.DeployToken),
		fmt.Sprintf("--%s %s", pluginV2FlagOperInstID, params.OperInstID),
	}

	// Add download/callback server addresses only when not skipping
	if !params.SkipDownload && params.DownloadSvrAddr != "" {
		args = append(args, fmt.Sprintf("--%s %s", pluginV2FlagDownloadSvrAdd, params.DownloadSvrAddr))
	}
	if !params.SkipCallback && params.CallbackSvrAddr != "" {
		args = append(args, fmt.Sprintf("--%s %s", pluginV2FlagCallbackSvrAdd, params.CallbackSvrAddr))
	}

	// Add skip flags when set
	if params.SkipCallback {
		args = append(args, "--skip_callback")
	}
	if params.SkipDownload {
		args = append(args, "--skip_download")
	}

	return args
}

// ToUnixScript converts the install params to a unix script.
func (params *PluginV2InstallParams) ToUnixScript() (string, string, error) {
	if err := params.Validate(); err != nil {
		return "", "", err
	}

	args := params.buildArgs()

	// build original cmd.
	installerFilePath := filepath.Join(params.InstallWorkDir, params.InstallerFileName)
	cmdStr := fmt.Sprintf("%s %s %s", installerFilePath, pluginV2CmdFullInstall, strings.Join(args, " "))

	// wrap cmd with stdout.
	scriptName := fmt.Sprintf("pluginv2_install_%s_%s.sh", params.PluginGroup, params.PluginName)

	stdoutPath := filepath.Join(params.InstallWorkDir, fmt.Sprintf("%s.stdout", scriptName))
	scriptContent := fmt.Sprintf("%s >%s 2>&1 &", cmdStr, stdoutPath)

	return scriptName, scriptContent, nil
}

// ToWindowsScript converts the install params to a windows script.
func (params *PluginV2InstallParams) ToWindowsScript() (string, string, error) {
	if err := params.Validate(); err != nil {
		return "", "", err
	}

	args := params.buildArgs()

	// build original cmd.
	installerFilePath := winpath.Join(params.InstallWorkDir, params.InstallerFileName)
	cmdStr := fmt.Sprintf("%s %s %s", installerFilePath, pluginV2CmdFullInstall, strings.Join(args, " "))

	// wrap cmd with stdout.
	scriptName := fmt.Sprintf("pluginv2_install_%s_%s.bat", params.PluginGroup, params.PluginName)
	scriptContent := fmt.Sprintf("%s", cmdStr)

	return scriptName, scriptContent, nil
}

// PluginV2UpgradeParams defines the upgrade params.
type PluginV2UpgradeParams struct {
	PluginV2CommonParams

	PluginGroup   string
	PluginName    string
	PluginVersion string
	PluginPkgName string

	DownloadSvrAddr string
	CallbackSvrAddr string

	DeployToken string
	OperInstID  string
}

// Validate validates the upgrade params.
func (params *PluginV2UpgradeParams) Validate() error {
	if err := params.PluginV2CommonParams.Validate(); err != nil {
		return err
	}

	if params.PluginGroup == "" {
		return fmt.Errorf("plugin group is empty")
	}

	if params.PluginName == "" {
		return fmt.Errorf("plugin name is empty")
	}

	if params.PluginVersion == "" {
		return fmt.Errorf("plugin version is empty")
	}

	if params.PluginPkgName == "" {
		return fmt.Errorf("plugin package name is empty")
	}

	if params.CallbackSvrAddr == "" {
		return fmt.Errorf("callback server address is empty")
	}

	if params.DownloadSvrAddr == "" {
		return fmt.Errorf("download server address is empty")
	}

	if params.DeployToken == "" {
		return fmt.Errorf("deploy token is empty")
	}

	if params.OperInstID == "" {
		return fmt.Errorf("operation instance id is empty")
	}

	return nil
}

func (params *PluginV2UpgradeParams) buildArgs() []string {
	args := []string{
		fmt.Sprintf("--%s %s", pluginV2FlagBaseDeployDir, params.BaseDeployDir),
		fmt.Sprintf("--%s %s", pluginV2FlagBaseWorkDir, params.BaseWorkDir),
		fmt.Sprintf("--%s %s", pluginV2FlagPluginGroup, params.PluginGroup),
		fmt.Sprintf("--%s %s", pluginV2FlagPluginName, params.PluginName),
		fmt.Sprintf("--%s %s", pluginV2FlagPluginVersion, params.PluginVersion),
		fmt.Sprintf("--%s %s", pluginV2FlagPluginPkgName, params.PluginPkgName),
		fmt.Sprintf("--%s %s", pluginV2FlagDeployEnv, params.DeployEnv),
		fmt.Sprintf("--%s %s", pluginV2FlagDownloadSvrAdd, params.DownloadSvrAddr),
		fmt.Sprintf("--%s %s", pluginV2FlagCallbackSvrAdd, params.CallbackSvrAddr),
		fmt.Sprintf("--%s %s", pluginV2FlagDeployToken, params.DeployToken),
		fmt.Sprintf("--%s %s", pluginV2FlagOperInstID, params.OperInstID),
	}

	return args
}

// ToUnixScript converts the upgrade params to a unix script.
func (params *PluginV2UpgradeParams) ToUnixScript() (string, string, error) {
	if err := params.Validate(); err != nil {
		return "", "", err
	}

	args := params.buildArgs()

	// build original cmd.
	installerFilePath := filepath.Join(params.InstallWorkDir, params.InstallerFileName)
	cmdStr := fmt.Sprintf("%s %s %s", installerFilePath, pluginV2CmdFullUpgrade, strings.Join(args, " "))

	// wrap cmd with stdout.
	scriptName := fmt.Sprintf("pluginv2_upgrade_%s_%s.sh", params.PluginGroup, params.PluginName)

	stdoutPath := filepath.Join(params.InstallWorkDir, fmt.Sprintf("%s.stdout", scriptName))
	scriptContent := fmt.Sprintf("%s >%s 2>&1 &", cmdStr, stdoutPath)

	return scriptName, scriptContent, nil
}

// ToWindowsScript converts the upgrade params to a windows script.
func (params *PluginV2UpgradeParams) ToWindowsScript() (string, string, error) {
	if err := params.Validate(); err != nil {
		return "", "", err
	}

	args := params.buildArgs()

	// build original cmd.
	installerFilePath := winpath.Join(params.InstallWorkDir, params.InstallerFileName)
	cmdStr := fmt.Sprintf("%s %s %s", installerFilePath, pluginV2CmdFullUpgrade, strings.Join(args, " "))

	// wrap cmd with stdout.
	scriptName := fmt.Sprintf("pluginv2_upgrade_%s_%s.bat", params.PluginGroup, params.PluginName)
	scriptContent := fmt.Sprintf("%s", cmdStr)

	return scriptName, scriptContent, nil
}

// PluginV2UninstallParams defines the uninstall params.
type PluginV2UninstallParams struct {
	PluginV2CommonParams

	PluginGroup string
	PluginName  string

	CallbackSvrAddr string

	DeployToken string
	OperInstID  string
}

// Validate validates the uninstall params.
func (params *PluginV2UninstallParams) Validate() error {
	if err := params.PluginV2CommonParams.Validate(); err != nil {
		return err
	}

	if params.PluginGroup == "" {
		return fmt.Errorf("plugin group is empty")
	}

	if params.PluginName == "" {
		return fmt.Errorf("plugin name is empty")
	}

	if params.CallbackSvrAddr == "" {
		return fmt.Errorf("callback server address is empty")
	}

	if params.DeployToken == "" {
		return fmt.Errorf("deploy token is empty")
	}

	if params.OperInstID == "" {
		return fmt.Errorf("operation instance id is empty")
	}

	return nil
}

func (params *PluginV2UninstallParams) buildArgs() []string {
	args := []string{
		fmt.Sprintf("--%s %s", pluginV2FlagBaseDeployDir, params.BaseDeployDir),
		fmt.Sprintf("--%s %s", pluginV2FlagBaseWorkDir, params.BaseWorkDir),
		fmt.Sprintf("--%s %s", pluginV2FlagPluginGroup, params.PluginGroup),
		fmt.Sprintf("--%s %s", pluginV2FlagPluginName, params.PluginName),
		fmt.Sprintf("--%s %s", pluginV2FlagDeployEnv, params.DeployEnv),
		fmt.Sprintf("--%s %s", pluginV2FlagCallbackSvrAdd, params.CallbackSvrAddr),
		fmt.Sprintf("--%s %s", pluginV2FlagDeployToken, params.DeployToken),
		fmt.Sprintf("--%s %s", pluginV2FlagOperInstID, params.OperInstID),
	}

	return args
}

// ToUnixScript converts the uninstall params to a unix script.
func (params *PluginV2UninstallParams) ToUnixScript() (string, string, error) {
	if err := params.Validate(); err != nil {
		return "", "", err
	}

	args := params.buildArgs()

	// build original cmd.
	installerFilePath := filepath.Join(params.InstallWorkDir, params.InstallerFileName)
	cmdStr := fmt.Sprintf("%s %s %s", installerFilePath, pluginV2CmdFullUninstall, strings.Join(args, " "))

	// wrap cmd with stdout.
	scriptName := fmt.Sprintf("pluginv2_uninstall_%s_%s.sh", params.PluginGroup, params.PluginName)

	stdoutPath := filepath.Join(params.InstallWorkDir, fmt.Sprintf("%s.stdout", scriptName))
	scriptContent := fmt.Sprintf("%s >%s 2>&1 &", cmdStr, stdoutPath)

	return scriptName, scriptContent, nil
}

// ToWindowsScript converts the uninstall params to a windows script.
func (params *PluginV2UninstallParams) ToWindowsScript() (string, string, error) {
	if err := params.Validate(); err != nil {
		return "", "", err
	}

	args := params.buildArgs()

	// build original cmd.
	installerFilePath := winpath.Join(params.InstallWorkDir, params.InstallerFileName)
	cmdStr := fmt.Sprintf("%s %s %s", installerFilePath, pluginV2CmdFullUninstall, strings.Join(args, " "))

	// wrap cmd with stdout.
	scriptName := fmt.Sprintf("pluginv2_uninstall_%s_%s.bat", params.PluginGroup, params.PluginName)
	scriptContent := fmt.Sprintf("%s", cmdStr)

	return scriptName, scriptContent, nil
}
