/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
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
}

// Validate validates the install params.
func (params *PluginInstallParams) Validate() error {

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
		fmt.Sprintf("--%s %s", pluginFlagDownloadSvrAdd, params.DownloadSvrAddr),
		fmt.Sprintf("--%s %s", pluginFlagCallbackSvrAdd, params.CallbackSvrAddr),
		fmt.Sprintf("--%s %s", pluginFlagDeployToken, params.DeployToken),
		fmt.Sprintf("--%s %s", pluginFlagOperInstID, params.OperInstID),
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
