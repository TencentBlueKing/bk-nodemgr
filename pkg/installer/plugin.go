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
	// pluginCmdFullInstall defines the installer cmd.
	pluginCmdFullInstall = "plugin full-install"

	// pluginCmdFullUpgrade defines the installer cmd.
	pluginCmdFullUpgrade = "plugin full-upgrade"

	// pluginCmdFullUninstall defines the installer cmd.
	pluginCmdFullUninstall = "plugin full-uninstall"

	// pluginCmdFullDebug defines the debug installer cmd.
	pluginCmdFullDebug = "plugin full-debug"
)

// pluginFlagName defines the plugin flag name.
type pluginFlagName string

// plugin flag name
const (
	// pluginFlagDeployEnv plugin flag name defines the deploy env.
	pluginFlagDeployEnv pluginFlagName = "deploy_env"

	// pluginFlagBaseDeployDir plugin flag name defines the base deploy dir.
	pluginFlagBaseDeployDir pluginFlagName = "base_deploy_dir"

	// pluginFlagBaseWorkDir plugin flag name defines the base work dir.
	pluginFlagBaseWorkDir pluginFlagName = "base_work_dir"

	// pluginFlagLogDir plugin flag name defines the log dir.
	pluginFlagLogDir pluginFlagName = "log_dir"

	// pluginFlagStatus plugin flag name defines the status.
	pluginFlagStatus pluginFlagName = "status"

	// pluginFlagPluginGroup plugin flag name defines the plugin group.
	pluginFlagPluginGroup pluginFlagName = "plugin_group"

	// pluginFlagPluginName plugin flag name defines the plugin name.
	pluginFlagPluginName pluginFlagName = "plugin_name"

	// pluginFlagPkgFile plugin flag name defines the pkg file.
	pluginFlagPkgFile pluginFlagName = "pkg_file"

	// pluginFlagDownloadSvrAdd plugin flag name defines the download server address.
	pluginFlagDownloadSvrAdd pluginFlagName = "dlsvr_addr"

	// pluginFlagCallbackSvrAdd plugin flag name defines the callback server address.
	pluginFlagCallbackSvrAdd pluginFlagName = "cbsvr_addr"

	// pluginFlagDeployToken plugin flag name defines the deploy token.
	pluginFlagDeployToken pluginFlagName = "deploy_token"

	// pluginFlagOperInstID plugin flag name defines the operation instance id.
	pluginFlagOperInstID pluginFlagName = "oper_inst_id"

	// pluginFlagPluginVersion plugin flag name defines the plugin version.
	pluginFlagPluginVersion pluginFlagName = "plugin_version"

	// pluginFlagPluginPkgName plugin flag name defines the plugin package name.
	pluginFlagPluginPkgName pluginFlagName = "plugin_pkg_name"

	// pluginFlagRunCmd plugin flag name defines the debug run command.
	pluginFlagRunCmd pluginFlagName = "run_cmd"

	// pluginFlagPidDir plugin flag name defines the debug pid dir.
	pluginFlagPidDir pluginFlagName = "pid_dir"

	// pluginFlagDebugAction plugin flag name defines the debug action.
	pluginFlagDebugAction pluginFlagName = "debug_action"

	// pluginFlagSkipCallback plugin flag name disables callback reporting.
	// SYNC: tools/cmd/installer/plugin/flag.SkipCallback.
	pluginFlagSkipCallback pluginFlagName = "skip_callback"

	// pluginFlagLogToStd plugin flag name defines whether logs are also written to stdout.
	pluginFlagLogToStd pluginFlagName = "log_to_std"
)

// PluginCommonParams defines the common params of installer.
type PluginCommonParams struct {
	InstallWorkDir    string
	InstallerFileName string

	BaseDeployDir string
	BaseWorkDir   string

	DeployEnv string
}

// Validate validates the common params.
func (params *PluginCommonParams) Validate() error {
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

// PluginInstallParams defines the install params.
type PluginInstallParams struct {
	PluginCommonParams

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
func (params *PluginInstallParams) Validate() error {
	if err := params.PluginCommonParams.Validate(); err != nil {
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

func (params *PluginInstallParams) buildArgs() []string {
	args := []string{
		fmt.Sprintf("--%s %s", pluginFlagBaseDeployDir, params.BaseDeployDir),
		fmt.Sprintf("--%s %s", pluginFlagBaseWorkDir, params.BaseWorkDir),
		fmt.Sprintf("--%s %s", pluginFlagPluginGroup, params.PluginGroup),
		fmt.Sprintf("--%s %s", pluginFlagPluginName, params.PluginName),
		fmt.Sprintf("--%s %s", pluginFlagPluginVersion, params.PluginVersion),
		fmt.Sprintf("--%s %s", pluginFlagPluginPkgName, params.PluginPkgName),
		fmt.Sprintf("--%s %s", pluginFlagDeployEnv, params.DeployEnv),
		fmt.Sprintf("--%s %s", pluginFlagDeployToken, params.DeployToken),
		fmt.Sprintf("--%s %s", pluginFlagOperInstID, params.OperInstID),
	}

	// Add download/callback server addresses only when not skipping
	if !params.SkipDownload && params.DownloadSvrAddr != "" {
		args = append(args, fmt.Sprintf("--%s %s", pluginFlagDownloadSvrAdd, params.DownloadSvrAddr))
	}
	if !params.SkipCallback && params.CallbackSvrAddr != "" {
		args = append(args, fmt.Sprintf("--%s %s", pluginFlagCallbackSvrAdd, params.CallbackSvrAddr))
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
func (params *PluginInstallParams) ToUnixScript() (string, string, error) {
	if err := params.Validate(); err != nil {
		return "", "", err
	}

	args := params.buildArgs()

	// build original cmd.
	installerFilePath := filepath.Join(params.InstallWorkDir, params.InstallerFileName)
	cmdStr := fmt.Sprintf("%s %s %s", installerFilePath, pluginCmdFullInstall, strings.Join(args, " "))

	// wrap cmd with stdout.
	scriptName := fmt.Sprintf("plugin_install_%s_%s.sh", params.PluginGroup, params.PluginName)

	stdoutPath := filepath.Join(params.InstallWorkDir, fmt.Sprintf("%s.stdout", scriptName))
	scriptContent := fmt.Sprintf("%s >%s 2>&1 &", cmdStr, stdoutPath)

	return scriptName, scriptContent, nil
}

// ToWindowsScript converts the install params to a windows script.
func (params *PluginInstallParams) ToWindowsScript() (string, string, error) {
	if err := params.Validate(); err != nil {
		return "", "", err
	}

	args := params.buildArgs()

	// build original cmd.
	installerFilePath := winpath.Join(params.InstallWorkDir, params.InstallerFileName)
	cmdStr := fmt.Sprintf("%s %s %s", installerFilePath, pluginCmdFullInstall, strings.Join(args, " "))

	// wrap cmd with stdout.
	scriptName := fmt.Sprintf("plugin_install_%s_%s.bat", params.PluginGroup, params.PluginName)
	scriptContent := fmt.Sprintf("%s", cmdStr)

	return scriptName, scriptContent, nil
}

// PluginUpgradeParams defines the upgrade params.
type PluginUpgradeParams struct {
	PluginCommonParams

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

// Validate validates the upgrade params.
func (params *PluginUpgradeParams) Validate() error {
	if err := params.PluginCommonParams.Validate(); err != nil {
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

func (params *PluginUpgradeParams) buildArgs() []string {
	args := []string{
		fmt.Sprintf("--%s %s", pluginFlagBaseDeployDir, params.BaseDeployDir),
		fmt.Sprintf("--%s %s", pluginFlagBaseWorkDir, params.BaseWorkDir),
		fmt.Sprintf("--%s %s", pluginFlagPluginGroup, params.PluginGroup),
		fmt.Sprintf("--%s %s", pluginFlagPluginName, params.PluginName),
		fmt.Sprintf("--%s %s", pluginFlagPluginVersion, params.PluginVersion),
		fmt.Sprintf("--%s %s", pluginFlagPluginPkgName, params.PluginPkgName),
		fmt.Sprintf("--%s %s", pluginFlagDeployEnv, params.DeployEnv),
		fmt.Sprintf("--%s %s", pluginFlagDeployToken, params.DeployToken),
		fmt.Sprintf("--%s %s", pluginFlagOperInstID, params.OperInstID),
	}
	// Add download/callback server addresses only when not skipping
	if !params.SkipDownload && params.DownloadSvrAddr != "" {
		args = append(args, fmt.Sprintf("--%s %s", pluginFlagDownloadSvrAdd, params.DownloadSvrAddr))
	}
	if !params.SkipCallback && params.CallbackSvrAddr != "" {
		args = append(args, fmt.Sprintf("--%s %s", pluginFlagCallbackSvrAdd, params.CallbackSvrAddr))
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

// ToUnixScript converts the upgrade params to a unix script.
func (params *PluginUpgradeParams) ToUnixScript() (string, string, error) {
	if err := params.Validate(); err != nil {
		return "", "", err
	}

	args := params.buildArgs()

	// build original cmd.
	installerFilePath := filepath.Join(params.InstallWorkDir, params.InstallerFileName)
	cmdStr := fmt.Sprintf("%s %s %s", installerFilePath, pluginCmdFullUpgrade, strings.Join(args, " "))

	// wrap cmd with stdout.
	scriptName := fmt.Sprintf("plugin_upgrade_%s_%s.sh", params.PluginGroup, params.PluginName)

	stdoutPath := filepath.Join(params.InstallWorkDir, fmt.Sprintf("%s.stdout", scriptName))
	scriptContent := fmt.Sprintf("%s >%s 2>&1 &", cmdStr, stdoutPath)

	return scriptName, scriptContent, nil
}

// ToWindowsScript converts the upgrade params to a windows script.
func (params *PluginUpgradeParams) ToWindowsScript() (string, string, error) {
	if err := params.Validate(); err != nil {
		return "", "", err
	}

	args := params.buildArgs()

	// build original cmd.
	installerFilePath := winpath.Join(params.InstallWorkDir, params.InstallerFileName)
	cmdStr := fmt.Sprintf("%s %s %s", installerFilePath, pluginCmdFullUpgrade, strings.Join(args, " "))

	// wrap cmd with stdout.
	scriptName := fmt.Sprintf("plugin_upgrade_%s_%s.bat", params.PluginGroup, params.PluginName)
	scriptContent := fmt.Sprintf("%s", cmdStr)

	return scriptName, scriptContent, nil
}

// PluginUninstallParams defines the uninstall params.
type PluginUninstallParams struct {
	PluginCommonParams

	PluginGroup string
	PluginName  string

	CallbackSvrAddr string

	DeployToken string
	OperInstID  string
}

// Validate validates the uninstall params.
func (params *PluginUninstallParams) Validate() error {
	if err := params.PluginCommonParams.Validate(); err != nil {
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

func (params *PluginUninstallParams) buildArgs() []string {
	args := []string{
		fmt.Sprintf("--%s %s", pluginFlagBaseDeployDir, params.BaseDeployDir),
		fmt.Sprintf("--%s %s", pluginFlagBaseWorkDir, params.BaseWorkDir),
		fmt.Sprintf("--%s %s", pluginFlagPluginGroup, params.PluginGroup),
		fmt.Sprintf("--%s %s", pluginFlagPluginName, params.PluginName),
		fmt.Sprintf("--%s %s", pluginFlagDeployEnv, params.DeployEnv),
		fmt.Sprintf("--%s %s", pluginFlagCallbackSvrAdd, params.CallbackSvrAddr),
		fmt.Sprintf("--%s %s", pluginFlagDeployToken, params.DeployToken),
		fmt.Sprintf("--%s %s", pluginFlagOperInstID, params.OperInstID),
	}

	return args
}

// ToUnixScript converts the uninstall params to a unix script.
func (params *PluginUninstallParams) ToUnixScript() (string, string, error) {
	if err := params.Validate(); err != nil {
		return "", "", err
	}

	args := params.buildArgs()

	// build original cmd.
	installerFilePath := filepath.Join(params.InstallWorkDir, params.InstallerFileName)
	cmdStr := fmt.Sprintf("%s %s %s", installerFilePath, pluginCmdFullUninstall, strings.Join(args, " "))

	// wrap cmd with stdout.
	scriptName := fmt.Sprintf("plugin_uninstall_%s_%s.sh", params.PluginGroup, params.PluginName)

	stdoutPath := filepath.Join(params.InstallWorkDir, fmt.Sprintf("%s.stdout", scriptName))
	scriptContent := fmt.Sprintf("%s >%s 2>&1 &", cmdStr, stdoutPath)

	return scriptName, scriptContent, nil
}

// ToWindowsScript converts the uninstall params to a windows script.
func (params *PluginUninstallParams) ToWindowsScript() (string, string, error) {
	if err := params.Validate(); err != nil {
		return "", "", err
	}

	args := params.buildArgs()

	// build original cmd.
	installerFilePath := winpath.Join(params.InstallWorkDir, params.InstallerFileName)
	cmdStr := fmt.Sprintf("%s %s %s", installerFilePath, pluginCmdFullUninstall, strings.Join(args, " "))

	// wrap cmd with stdout.
	scriptName := fmt.Sprintf("plugin_uninstall_%s_%s.bat", params.PluginGroup, params.PluginName)
	scriptContent := fmt.Sprintf("%s", cmdStr)

	return scriptName, scriptContent, nil
}

// PluginDebugParams defines the debug params.
type PluginDebugParams struct {
	PluginCommonParams

	PluginGroup string
	PluginName  string

	DeployToken string
	OperInstID  string

	DebugAction string
	RunCmd      string
	PidDir      string
}

// Validate validates the debug params.
func (params *PluginDebugParams) Validate() error {
	if err := params.PluginCommonParams.Validate(); err != nil {
		return err
	}

	if params.PluginGroup == "" {
		return fmt.Errorf("plugin group is empty")
	}

	if params.PluginName == "" {
		return fmt.Errorf("plugin name is empty")
	}

	if params.DeployToken == "" {
		return fmt.Errorf("deploy token is empty")
	}

	if params.OperInstID == "" {
		return fmt.Errorf("operation instance id is empty")
	}
	if params.DebugAction == "" {
		return fmt.Errorf("debug action is empty")
	}

	return nil
}

func (params *PluginDebugParams) buildArgs(quote func(string) string) []string {
	args := []string{
		fmt.Sprintf("--%s %s", pluginFlagBaseDeployDir, quote(params.BaseDeployDir)),
		fmt.Sprintf("--%s %s", pluginFlagBaseWorkDir, quote(params.BaseWorkDir)),
		fmt.Sprintf("--%s %s", pluginFlagPluginGroup, quote(params.PluginGroup)),
		fmt.Sprintf("--%s %s", pluginFlagPluginName, quote(params.PluginName)),
		fmt.Sprintf("--%s %s", pluginFlagDeployEnv, quote(params.DeployEnv)),
		fmt.Sprintf("--%s %s", pluginFlagDeployToken, quote(params.DeployToken)),
		fmt.Sprintf("--%s %s", pluginFlagOperInstID, quote(params.OperInstID)),
	}
	args = append(args, fmt.Sprintf("--%s %s", pluginFlagDebugAction, quote(params.DebugAction)))
	args = append(args, fmt.Sprintf("--%s", pluginFlagSkipCallback))
	args = append(args, fmt.Sprintf("--%s", pluginFlagLogToStd))
	if params.RunCmd != "" {
		args = append(args, fmt.Sprintf("--%s %s", pluginFlagRunCmd, quote(params.RunCmd)))
	}
	if params.PidDir != "" {
		args = append(args, fmt.Sprintf("--%s %s", pluginFlagPidDir, quote(params.PidDir)))
	}

	return args
}

// ToUnixScript converts the debug params to a unix script.
func (params *PluginDebugParams) ToUnixScript() (string, string, error) {
	if err := params.Validate(); err != nil {
		return "", "", err
	}

	args := params.buildArgs(shellSingleQuote)
	installerFilePath := filepath.Join(params.InstallWorkDir, params.InstallerFileName)
	cmdStr := fmt.Sprintf("%s %s %s", shellSingleQuote(installerFilePath), pluginCmdFullDebug, strings.Join(args, " "))
	scriptName := fmt.Sprintf("plugin_debug_%s_%s_%s.sh", params.PluginGroup, params.PluginName, params.DebugAction)

	return scriptName, cmdStr, nil
}

// ToWindowsScript converts the debug params to a windows script.
func (params *PluginDebugParams) ToWindowsScript() (string, string, error) {
	if err := params.Validate(); err != nil {
		return "", "", err
	}

	args := params.buildArgs(powerShellSingleQuote)
	installerFilePath := winpath.Join(params.InstallWorkDir, params.InstallerFileName)
	cmdStr := fmt.Sprintf("%s %s %s", powerShellSingleQuote(installerFilePath), pluginCmdFullDebug, strings.Join(args, " "))
	scriptName := fmt.Sprintf("plugin_debug_%s_%s_%s.bat", params.PluginGroup, params.PluginName, params.DebugAction)

	return scriptName, cmdStr, nil
}

func shellSingleQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func powerShellSingleQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}
