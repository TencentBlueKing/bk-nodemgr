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
	// NodeCmdFullInstall defines the installer cmd.
	NodeCmdFullInstall = "node full-install"

	// NodeCmdFullUpgrade defines the installer cmd.
	NodeCmdFullUpgrade = "node full-upgrade"

	// NodeCmdFullReconfig defines the installer cmd.
	NodeCmdFullReconfig = "node full-reconfig"

	// NodeCmdFullUninstall defines the installer cmd.
	NodeCmdFullUninstall = "node full-uninstall"

	// NodeCmdStepCleanTmp defines the installer cmd.
	NodeCmdStepCleanTmp = "node step clean-tmp"

	// NodeCmdStepRestart defines the installer cmd.
	NodeCmdStepRestart = "node step restart"
)

// nodeFlagName defines the node flag name.
type nodeFlagName string

// node flag name
const (
	// nodeFlagRenewGSEProc node flag name defines whether to renew the generated GSE .proc file.
	// SYNC: tools/cmd/installer/node/flag.RenewGSEProc.
	nodeFlagRenewGSEProc nodeFlagName = "renew_gse_proc"

	// nodeFlagRenewGSETask node flag name defines whether to renew the generated GSE .task file.
	// SYNC: tools/cmd/installer/node/flag.RenewGSETask.
	nodeFlagRenewGSETask nodeFlagName = "renew_gse_task"

	// nodeFlagDeployEnv node flag name defines the deploy environment.
	// SYNC: tools/cmd/installer/node/flag.DeployEnv.
	nodeFlagDeployEnv nodeFlagName = "deploy_env"

	// nodeFlagGeneration node flag name defines the node generation.
	// SYNC: tools/cmd/installer/node/flag.Generation.
	nodeFlagGeneration nodeFlagName = "generation"

	// nodeFlagNodeRole node flag name defines the node role.
	// SYNC: tools/cmd/installer/node/flag.NodeRole.
	nodeFlagNodeRole nodeFlagName = "node_role"

	// nodeFlagBaseDeployDir node flag name defines the base deploy directory.
	// SYNC: tools/cmd/installer/node/flag.BaseDeployDir.
	nodeFlagBaseDeployDir nodeFlagName = "base_deploy_dir"

	// nodeFlagBaseWorkDir node flag name defines the base work directory.
	// SYNC: tools/cmd/installer/node/flag.BaseWorkDir.
	nodeFlagBaseWorkDir nodeFlagName = "base_work_dir"

	// nodeFlagLogToStd node flag name defines whether logs are also written to stdout.
	// SYNC: tools/cmd/installer/node/flag.LogToStd.
	nodeFlagLogToStd nodeFlagName = "log_to_std"

	// nodeFlagAgentID node flag name defines the agent id.
	// SYNC: tools/cmd/installer/node/flag.AgentID.
	nodeFlagAgentID nodeFlagName = "agent_id"

	// nodeFlagDeployToken node flag name defines the deploy token.
	// SYNC: tools/cmd/installer/node/flag.DeployToken.
	nodeFlagDeployToken nodeFlagName = "deploy_token"

	// nodeFlagDownloadSvrAddr node flag name defines the download server address.
	// SYNC: tools/cmd/installer/node/flag.DownloadSvrAddr.
	nodeFlagDownloadSvrAddr nodeFlagName = "dlsvr_addr"

	// nodeFlagCallbackSvrAddr node flag name defines the callback server address.
	// SYNC: tools/cmd/installer/node/flag.CallbackSvrAddr.
	nodeFlagCallbackSvrAddr nodeFlagName = "cbsvr_addr"

	// nodeFlagNodeVersion node flag name defines the node version.
	// SYNC: tools/cmd/installer/node/flag.NodeVersion.
	nodeFlagNodeVersion nodeFlagName = "node_version"

	// nodeFlagOperInstID node flag name defines the operation instance id.
	// SYNC: tools/cmd/installer/node/flag.OperInstID.
	nodeFlagOperInstID nodeFlagName = "oper_inst_id"

	// nodeFlagForce node flag name defines whether to force a node operation.
	// SYNC: tools/cmd/installer/node/flag.Force.
	nodeFlagForce nodeFlagName = "force"

	// nodeFlagRestart node flag name defines whether to restart the node after a full command.
	// SYNC: tools/cmd/installer/node/flag.Restart.
	nodeFlagRestart nodeFlagName = "restart"

	// nodeFlagSkipDownload node flag name defines whether to skip downloading files.
	// SYNC: tools/cmd/installer/node/flag.SkipDownload.
	nodeFlagSkipDownload nodeFlagName = "skip_download"

	// nodeFlagSkipCallback node flag name defines whether to skip callback reporting.
	// SYNC: tools/cmd/installer/node/flag.SkipCallback.
	nodeFlagSkipCallback nodeFlagName = "skip_callback"
)

// NodeCommonParams defines the common installer-facing params shared by node commands.
type NodeCommonParams struct {
	DeployEnv     string
	Generation    int
	NodeRole      string
	BaseWorkDir   string
	BaseDeployDir string

	AdditionArgs []string
}

// Validate validates the common node params required by installer CLI flags.
func (params *NodeCommonParams) Validate() error {
	if params.DeployEnv == "" {
		return fmt.Errorf("deploy env is empty")
	}

	if params.Generation == 0 {
		return fmt.Errorf("generation is empty")
	}

	if params.NodeRole == "" {
		return fmt.Errorf("node role is empty")
	}

	if params.BaseWorkDir == "" {
		return fmt.Errorf("base work dir is empty")
	}

	if params.BaseDeployDir == "" {
		return fmt.Errorf("base deploy dir is empty")
	}

	return nil
}

func (params *NodeCommonParams) buildArgs() []string {
	return []string{
		fmt.Sprintf("--%s %s", nodeFlagDeployEnv, params.DeployEnv),
		fmt.Sprintf("--%s %d", nodeFlagGeneration, params.Generation),
		fmt.Sprintf("--%s %s", nodeFlagNodeRole, params.NodeRole),
		fmt.Sprintf("--%s %s", nodeFlagBaseWorkDir, params.BaseWorkDir),
		fmt.Sprintf("--%s %s", nodeFlagBaseDeployDir, params.BaseDeployDir),
	}
}

// NodeInstallParams defines the params for a full node install command.
type NodeInstallParams struct {
	NodeCommonParams

	InstallerPath string

	DownloadSvrAddr string
	CallbackSvrAddr string
	NodeVersion     string
	DeployToken     string
	OperInstID      string
	AgentID         string

	LogToStd     bool
	SkipDownload bool
	SkipCallback bool
	RenewGSEProc bool
	RenewGSETask bool
	// DownloadBeforeCallback keeps legacy token-first install arg order for compatibility.
	DownloadBeforeCallback bool
}

// Validate validates the full node install params required by the installer CLI.
func (params *NodeInstallParams) Validate() error {
	if err := params.NodeCommonParams.Validate(); err != nil {
		return err
	}

	if err := validateNodeInstallerPath(params.InstallerPath); err != nil {
		return err
	}

	return validateNodeInstallCommandParams(
		params.DeployToken,
		params.NodeVersion,
		params.OperInstID,
		params.DownloadSvrAddr,
		params.CallbackSvrAddr,
		params.SkipDownload,
		params.SkipCallback,
	)
}

func (params *NodeInstallParams) buildArgs() []string {
	args := params.NodeCommonParams.buildArgs()
	args = append(args,
		fmt.Sprintf("--%s %s", nodeFlagDeployToken, params.DeployToken),
		fmt.Sprintf("--%s %s", nodeFlagNodeVersion, params.NodeVersion),
		fmt.Sprintf("--%s %s", nodeFlagOperInstID, params.OperInstID),
	)
	args = appendNodeCallbackArg(args, params.SkipCallback, params.CallbackSvrAddr)
	args = appendNodeDownloadArg(args, params.SkipDownload, params.DownloadSvrAddr)
	args = appendNodeInstallOptionalArgs(args, params.LogToStd, params.AgentID, params.RenewGSEProc, params.RenewGSETask)

	return append(args, params.AdditionArgs...)
}

func (params *NodeInstallParams) buildDownloadBeforeCallbackArgs() []string {
	args := params.NodeCommonParams.buildArgs()
	args = append(args,
		fmt.Sprintf("--%s %s", nodeFlagDeployToken, params.DeployToken),
		fmt.Sprintf("--%s %s", nodeFlagNodeVersion, params.NodeVersion),
		fmt.Sprintf("--%s %s", nodeFlagOperInstID, params.OperInstID),
	)
	args = appendNodeDownloadArg(args, params.SkipDownload, params.DownloadSvrAddr)
	args = appendNodeCallbackArg(args, params.SkipCallback, params.CallbackSvrAddr)
	args = appendNodeInstallOptionalArgs(args, params.LogToStd, params.AgentID, params.RenewGSEProc, params.RenewGSETask)

	return append(args, params.AdditionArgs...)
}

func (params *NodeInstallParams) buildServerFirstArgs() []string {
	args := params.NodeCommonParams.buildArgs()
	args = appendNodeDownloadArg(args, params.SkipDownload, params.DownloadSvrAddr)
	args = appendNodeCallbackArg(args, params.SkipCallback, params.CallbackSvrAddr)
	args = append(args,
		fmt.Sprintf("--%s %s", nodeFlagDeployToken, params.DeployToken),
		fmt.Sprintf("--%s %s", nodeFlagNodeVersion, params.NodeVersion),
		fmt.Sprintf("--%s %s", nodeFlagOperInstID, params.OperInstID),
	)
	args = appendNodeInstallOptionalArgs(args, params.LogToStd, params.AgentID, params.RenewGSEProc, params.RenewGSETask)

	return append(args, params.AdditionArgs...)
}

// NodeUpgradeParams defines the params for a full node upgrade command.
type NodeUpgradeParams struct {
	NodeCommonParams

	InstallWorkDir    string
	InstallerFileName string

	DownloadSvrAddr string
	CallbackSvrAddr string
	NodeVersion     string
	DeployToken     string
	OperInstID      string

	Force        bool
	Restart      bool
	SkipDownload bool
}

// Validate validates the full node upgrade params required by the installer CLI.
func (params *NodeUpgradeParams) Validate() error {
	if err := params.NodeCommonParams.Validate(); err != nil {
		return err
	}

	if err := validateNodeInstallerFile(params.InstallWorkDir, params.InstallerFileName); err != nil {
		return err
	}

	if params.CallbackSvrAddr == "" {
		return fmt.Errorf("callback server address is empty")
	}

	if params.DeployToken == "" {
		return fmt.Errorf("deploy token is empty")
	}

	if params.NodeVersion == "" {
		return fmt.Errorf("node version is empty")
	}

	if params.OperInstID == "" {
		return fmt.Errorf("operation instance id is empty")
	}

	return nil
}

func (params *NodeUpgradeParams) buildArgs() []string {
	args := params.NodeCommonParams.buildArgs()
	if params.SkipDownload {
		args = append(args,
			fmt.Sprintf("--%s %s", nodeFlagCallbackSvrAddr, params.CallbackSvrAddr),
			fmt.Sprintf("--%s %s", nodeFlagDeployToken, params.DeployToken),
			fmt.Sprintf("--%s %s", nodeFlagNodeVersion, params.NodeVersion),
			fmt.Sprintf("--%s %s", nodeFlagOperInstID, params.OperInstID),
		)
	} else {
		args = append(args,
			fmt.Sprintf("--%s %s", nodeFlagDownloadSvrAddr, params.DownloadSvrAddr),
			fmt.Sprintf("--%s %s", nodeFlagCallbackSvrAddr, params.CallbackSvrAddr),
			fmt.Sprintf("--%s %s", nodeFlagDeployToken, params.DeployToken),
			fmt.Sprintf("--%s %s", nodeFlagNodeVersion, params.NodeVersion),
			fmt.Sprintf("--%s %s", nodeFlagOperInstID, params.OperInstID),
		)
	}
	if params.SkipDownload {
		args = append(args, fmt.Sprintf("--%s", nodeFlagSkipDownload))
	}

	return append(args, params.AdditionArgs...)
}

// NodeReconfigParams defines the params for a full node reconfig command.
type NodeReconfigParams struct {
	NodeCommonParams

	InstallWorkDir    string
	InstallerFileName string

	CallbackSvrAddr string
	DeployToken     string
	OperInstID      string

	Force   bool
	Restart bool
}

// Validate validates the full node reconfig params required by the installer CLI.
func (params *NodeReconfigParams) Validate() error {
	if err := params.NodeCommonParams.Validate(); err != nil {
		return err
	}

	if err := validateNodeInstallerFile(params.InstallWorkDir, params.InstallerFileName); err != nil {
		return err
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

func (params *NodeReconfigParams) buildArgs() []string {
	args := params.NodeCommonParams.buildArgs()
	args = append(args,
		fmt.Sprintf("--%s %s", nodeFlagCallbackSvrAddr, params.CallbackSvrAddr),
		fmt.Sprintf("--%s %s", nodeFlagDeployToken, params.DeployToken),
		fmt.Sprintf("--%s %s", nodeFlagOperInstID, params.OperInstID),
	)

	return append(args, params.AdditionArgs...)
}

// NodeUninstallParams defines the params for a full node uninstall command.
type NodeUninstallParams struct {
	NodeCommonParams

	InstallWorkDir    string
	InstallerFileName string

	CallbackSvrAddr string
	DeployToken     string
	OperInstID      string
	SkipCallback    bool
}

// Validate validates the full node uninstall params required by the installer CLI.
func (params *NodeUninstallParams) Validate() error {
	if err := params.NodeCommonParams.Validate(); err != nil {
		return err
	}

	if err := validateNodeInstallerFile(params.InstallWorkDir, params.InstallerFileName); err != nil {
		return err
	}

	if !params.SkipCallback && params.CallbackSvrAddr == "" {
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

func (params *NodeUninstallParams) buildArgs() []string {
	args := params.NodeCommonParams.buildArgs()
	args = appendNodeCallbackArg(args, params.SkipCallback, params.CallbackSvrAddr)
	args = append(args,
		fmt.Sprintf("--%s %s", nodeFlagDeployToken, params.DeployToken),
		fmt.Sprintf("--%s %s", nodeFlagOperInstID, params.OperInstID),
	)

	return append(args, params.AdditionArgs...)
}

// NodeStepRestartParams defines the params for a node restart step command.
type NodeStepRestartParams struct {
	NodeCommonParams

	InstallWorkDir    string
	InstallerFileName string

	Force bool
}

// Validate validates the node restart step params required by the installer CLI.
func (params *NodeStepRestartParams) Validate() error {
	if err := params.NodeCommonParams.Validate(); err != nil {
		return err
	}

	return validateNodeInstallerFile(params.InstallWorkDir, params.InstallerFileName)
}

// NodeStepCleanTmpParams defines the params for a clean temporary installer files step command.
type NodeStepCleanTmpParams struct {
	NodeCommonParams

	InstallWorkDir    string
	InstallerFileName string
}

// Validate validates the clean temporary installer files step params required by the installer CLI.
func (params *NodeStepCleanTmpParams) Validate() error {
	if err := params.NodeCommonParams.Validate(); err != nil {
		return err
	}

	return validateNodeInstallerFile(params.InstallWorkDir, params.InstallerFileName)
}

func (params *NodeStepRestartParams) buildArgs() []string {
	args := params.NodeCommonParams.buildArgs()
	args = append(args, fmt.Sprintf("--%s", nodeFlagForce))

	return append(args, params.AdditionArgs...)
}

// ToUnixScript converts the restart step params to a unix script.
func (params *NodeStepRestartParams) ToUnixScript() (string, string, error) {
	if err := params.Validate(); err != nil {
		return "", "", err
	}

	args := params.buildArgs()
	installerPath := filepath.Clean(filepath.Join(params.InstallWorkDir, params.InstallerFileName))
	stdoutPath := filepath.Clean(fmt.Sprintf("%s.stdout", installerPath))

	scriptName := "restart.sh"
	scriptContent := fmt.Sprintf("chmod +x %s && %s %s %s >%s 2>&1 &", installerPath, installerPath, NodeCmdStepRestart, strings.Join(args, " "), stdoutPath)

	return scriptName, scriptContent, nil
}

// ToWindowsScript converts the restart step params to a windows script.
func (params *NodeStepRestartParams) ToWindowsScript() (string, string, error) {
	if err := params.Validate(); err != nil {
		return "", "", err
	}

	args := params.buildArgs()
	installerPath := winpath.Clean(winpath.Join(params.InstallWorkDir, params.InstallerFileName))
	stdoutPath := winpath.Clean(fmt.Sprintf("%s.stdout", installerPath))

	scriptName := "restart.bat"
	scriptContent := fmt.Sprintf("%s %s %s >%s 2>&1", installerPath, NodeCmdStepRestart, strings.Join(args, " "), stdoutPath)

	return scriptName, scriptContent, nil
}

func (params *NodeStepCleanTmpParams) buildArgs() []string {
	args := params.NodeCommonParams.buildArgs()

	return append(args, params.AdditionArgs...)
}

// ToUnixScript converts the clean temporary files step params to a unix script.
func (params *NodeStepCleanTmpParams) ToUnixScript() (string, string, error) {
	if err := params.Validate(); err != nil {
		return "", "", err
	}

	args := params.buildArgs()
	installerPath := filepath.Clean(filepath.Join(params.InstallWorkDir, params.InstallerFileName))
	stdoutPath := filepath.Clean(fmt.Sprintf("%s.stdout", installerPath))

	scriptName := "clean.sh"
	scriptContent := fmt.Sprintf("chmod +x %s && %s %s %s >%s 2>&1 &", installerPath, installerPath, NodeCmdStepCleanTmp, strings.Join(args, " "), stdoutPath)

	return scriptName, scriptContent, nil
}

// ToWindowsScript converts the clean temporary files step params to a windows script.
func (params *NodeStepCleanTmpParams) ToWindowsScript() (string, string, error) {
	if err := params.Validate(); err != nil {
		return "", "", err
	}

	args := params.buildArgs()
	installerPath := winpath.Clean(winpath.Join(params.InstallWorkDir, params.InstallerFileName))
	stdoutPath := winpath.Clean(fmt.Sprintf("%s.stdout", installerPath))

	scriptName := "clean.bat"
	scriptContent := fmt.Sprintf("%s %s %s >%s 2>&1", installerPath, NodeCmdStepCleanTmp, strings.Join(args, " "), stdoutPath)

	return scriptName, scriptContent, nil
}

// NodeOfflineInstallParams defines the params for rendering an offline node install command.
type NodeOfflineInstallParams struct {
	NodeCommonParams

	InstallerFileName string

	NodeVersion string
	DeployToken string
	OperInstID  string
	AgentID     string

	LogToStd     bool
	RenewGSEProc bool
	RenewGSETask bool
}

// Validate validates the offline node install params required by the installer CLI.
func (params *NodeOfflineInstallParams) Validate() error {
	if err := params.NodeCommonParams.Validate(); err != nil {
		return err
	}

	if params.InstallerFileName == "" {
		return fmt.Errorf("installer file name is empty")
	}

	return validateNodeInstallCommandParams(
		params.DeployToken,
		params.NodeVersion,
		params.OperInstID,
		"",
		"",
		true,
		true,
	)
}

func (params *NodeOfflineInstallParams) buildArgs() []string {
	args := params.NodeCommonParams.buildArgs()
	args = append(args,
		fmt.Sprintf("--%s %s", nodeFlagDeployToken, params.DeployToken),
		fmt.Sprintf("--%s %s", nodeFlagNodeVersion, params.NodeVersion),
		fmt.Sprintf("--%s %s", nodeFlagOperInstID, params.OperInstID),
		fmt.Sprintf("--%s", nodeFlagSkipCallback),
		fmt.Sprintf("--%s", nodeFlagSkipDownload),
	)
	args = appendNodeInstallOptionalArgs(args, params.LogToStd, params.AgentID, params.RenewGSEProc, params.RenewGSETask)

	return append(args, params.AdditionArgs...)
}

// ToUnixScriptOffline converts the offline install params to a unix install.sh script.
func (params *NodeOfflineInstallParams) ToUnixScriptOffline() (string, string, error) {
	if err := params.Validate(); err != nil {
		return "", "", err
	}

	dataDir := fmt.Sprintf("%s/%s/data", params.BaseWorkDir, params.DeployEnv)
	lines := []string{
		"#!/bin/bash",
		`set -e`,
		`SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"`,
		fmt.Sprintf(`DATA_DIR="%s"`, dataDir),
		`mkdir -p "${DATA_DIR}"`,
		fmt.Sprintf(`cp -rf "${SCRIPT_DIR}/%s/." "${DATA_DIR}/"`, OfflinePkgRelPathData),
		fmt.Sprintf(`chmod +x "${SCRIPT_DIR}/%s"`, params.InstallerFileName),
		fmt.Sprintf(`"${SCRIPT_DIR}/%s" `, params.InstallerFileName) + NodeCmdFullInstall + ` \`,
	}

	args := params.buildArgs()
	for i, arg := range args {
		if i < len(args)-1 {
			lines = append(lines, "  "+arg+` \`)
		} else {
			lines = append(lines, "  "+arg)
		}
	}

	lines = append(lines,
		fmt.Sprintf(`echo "--- %s ---"`, DataFileName),
		fmt.Sprintf(`cat "%s/%s"`, dataDir, DataFileName),
	)

	return OfflinePkgInstallScriptName, strings.Join(lines, "\n") + "\n", nil
}

// ToUnixScript converts the install params to a unix script.
func (params *NodeInstallParams) ToUnixScript() (string, string, error) {
	if err := params.Validate(); err != nil {
		return "", "", err
	}

	args := params.buildArgs()

	scriptName := "install.sh"
	stdoutPath := filepath.Clean(fmt.Sprintf("%s.stdout", params.InstallerPath))
	scriptContent := fmt.Sprintf("%s %s %s >%s 2>&1 &", params.InstallerPath, NodeCmdFullInstall, strings.Join(args, " "), stdoutPath)

	return scriptName, scriptContent, nil
}

// ToUnixScriptDownloadBeforeCallback converts install params to a unix script while preserving legacy
// download-before-callback flag order after deploy token, node version, and operation instance id.
func (params *NodeInstallParams) ToUnixScriptDownloadBeforeCallback() (string, string, error) {
	if err := params.Validate(); err != nil {
		return "", "", err
	}

	args := params.buildDownloadBeforeCallbackArgs()

	scriptName := "install.sh"
	stdoutPath := filepath.Clean(fmt.Sprintf("%s.stdout", params.InstallerPath))
	scriptContent := fmt.Sprintf("%s %s %s >%s 2>&1 &", params.InstallerPath, NodeCmdFullInstall, strings.Join(args, " "), stdoutPath)

	return scriptName, scriptContent, nil
}

// ToWindowsScript converts the install params to a windows script.
func (params *NodeInstallParams) ToWindowsScript() (string, string, error) {
	if err := params.Validate(); err != nil {
		return "", "", err
	}

	args := params.buildServerFirstArgs()
	if params.DownloadBeforeCallback {
		args = params.buildDownloadBeforeCallbackArgs()
	}
	scriptName := "install.bat"
	stdoutPath := winpath.Clean(fmt.Sprintf("%s.stdout", params.InstallerPath))
	scriptContent := fmt.Sprintf("cd %s && %s %s %s >%s 2>&1", winpath.Join(params.BaseWorkDir, params.DeployEnv), params.InstallerPath, NodeCmdFullInstall, strings.Join(args, " "), stdoutPath)

	return scriptName, scriptContent, nil
}

// ToWindowsShellScript converts the install params to a Cygwin shell script for Windows installer execution.
func (params *NodeInstallParams) ToWindowsShellScript() (string, string, error) {
	if err := params.Validate(); err != nil {
		return "", "", err
	}

	shellParams := *params
	shellParams.BaseWorkDir = winpath.ToSlash(winpath.Clean(params.BaseWorkDir))
	shellParams.BaseDeployDir = winpath.ToSlash(winpath.Clean(params.BaseDeployDir))

	args := shellParams.buildServerFirstArgs()
	if params.DownloadBeforeCallback {
		args = shellParams.buildDownloadBeforeCallbackArgs()
	}
	scriptName := "install.sh"
	workDir := winpath.ToSlash(winpath.Clean(winpath.Join(params.BaseWorkDir, params.DeployEnv)))
	installerPath := winpath.ToSlash(winpath.Clean(params.InstallerPath))
	stdoutPath := winpath.ToSlash(winpath.Clean(fmt.Sprintf("%s.stdout", params.InstallerPath)))
	scriptContent := fmt.Sprintf(
		"cd \"%s\" && \"%s\" %s %s >%s 2>&1 &",
		workDir,
		installerPath,
		NodeCmdFullInstall,
		strings.Join(args, " "),
		stdoutPath,
	)

	return scriptName, scriptContent, nil
}

// ToUnixScriptManual converts the install params to a unix manual install script.
func (params *NodeInstallParams) ToUnixScriptManual() (string, string, error) {
	if err := params.Validate(); err != nil {
		return "", "", err
	}

	args := params.buildServerFirstArgs()

	scriptName := "install.sh"
	scriptContent := fmt.Sprintf("%s %s %s", params.InstallerPath, NodeCmdFullInstall, strings.Join(args, " "))

	return scriptName, scriptContent, nil
}

// ToWindowsScriptManual converts the install params to a windows manual install script.
func (params *NodeInstallParams) ToWindowsScriptManual() (string, string, error) {
	if err := params.Validate(); err != nil {
		return "", "", err
	}

	args := params.buildServerFirstArgs()

	scriptName := "install.bat"
	scriptContent := fmt.Sprintf("cd %s && %s %s %s", winpath.Join(params.BaseWorkDir, params.DeployEnv), params.InstallerPath, NodeCmdFullInstall, strings.Join(args, " "))

	return scriptName, scriptContent, nil
}

// ToUnixScript converts the upgrade params to a unix script.
func (params *NodeUpgradeParams) ToUnixScript() (string, string, error) {
	if err := params.Validate(); err != nil {
		return "", "", err
	}

	args := params.buildArgs()
	installerPath := filepath.Clean(filepath.Join(params.InstallWorkDir, params.InstallerFileName))
	stdoutPath := filepath.Clean(fmt.Sprintf("%s.stdout", installerPath))

	scriptName := "upgrade.sh"
	scriptContent := fmt.Sprintf("chmod +x %s && %s %s %s >%s 2>&1 &", installerPath, installerPath, NodeCmdFullUpgrade, strings.Join(args, " "), stdoutPath)

	return scriptName, scriptContent, nil
}

// ToWindowsScript converts the upgrade params to a windows script.
func (params *NodeUpgradeParams) ToWindowsScript() (string, string, error) {
	if err := params.Validate(); err != nil {
		return "", "", err
	}

	args := params.buildArgs()
	installerPath := winpath.Clean(winpath.Join(params.InstallWorkDir, params.InstallerFileName))
	stdoutPath := winpath.Clean(fmt.Sprintf("%s.stdout", installerPath))

	scriptName := "upgrade.bat"
	scriptContent := fmt.Sprintf("%s %s %s >%s 2>&1", installerPath, NodeCmdFullUpgrade, strings.Join(args, " "), stdoutPath)

	return scriptName, scriptContent, nil
}

// ToUnixScript converts the reconfig params to a unix script.
func (params *NodeReconfigParams) ToUnixScript() (string, string, error) {
	if err := params.Validate(); err != nil {
		return "", "", err
	}

	args := params.buildArgs()
	installerPath := filepath.Clean(filepath.Join(params.InstallWorkDir, params.InstallerFileName))
	stdoutPath := filepath.Clean(fmt.Sprintf("%s.stdout", installerPath))

	scriptName := "reconfig.sh"
	scriptContent := fmt.Sprintf("chmod +x %s && %s %s %s >%s 2>&1 &", installerPath, installerPath, NodeCmdFullReconfig, strings.Join(args, " "), stdoutPath)

	return scriptName, scriptContent, nil
}

// ToWindowsScript converts the reconfig params to a windows script.
func (params *NodeReconfigParams) ToWindowsScript() (string, string, error) {
	if err := params.Validate(); err != nil {
		return "", "", err
	}

	args := params.buildArgs()
	installerPath := winpath.Clean(winpath.Join(params.InstallWorkDir, params.InstallerFileName))
	stdoutPath := winpath.Clean(fmt.Sprintf("%s.stdout", installerPath))

	scriptName := "reconfig.bat"
	scriptContent := fmt.Sprintf("%s %s %s >%s 2>&1", installerPath, NodeCmdFullReconfig, strings.Join(args, " "), stdoutPath)

	return scriptName, scriptContent, nil
}

// ToUnixScript converts the uninstall params to a unix script.
func (params *NodeUninstallParams) ToUnixScript() (string, string, error) {
	if err := params.Validate(); err != nil {
		return "", "", err
	}

	args := params.buildArgs()
	installerPath := filepath.Clean(filepath.Join(params.InstallWorkDir, params.InstallerFileName))
	stdoutPath := filepath.Clean(fmt.Sprintf("%s.stdout", installerPath))

	scriptName := "uninstall.sh"
	scriptContent := fmt.Sprintf("chmod +x %s && %s %s %s >%s 2>&1 &", installerPath, installerPath, NodeCmdFullUninstall, strings.Join(args, " "), stdoutPath)

	return scriptName, scriptContent, nil
}

// ToWindowsScript converts the uninstall params to a windows script.
func (params *NodeUninstallParams) ToWindowsScript() (string, string, error) {
	if err := params.Validate(); err != nil {
		return "", "", err
	}

	args := params.buildArgs()
	installerPath := winpath.Clean(winpath.Join(params.InstallWorkDir, params.InstallerFileName))
	stdoutPath := winpath.Clean(fmt.Sprintf("%s.stdout", installerPath))

	scriptName := "uninstall.bat"
	scriptContent := fmt.Sprintf("%s %s %s >%s 2>&1", installerPath, NodeCmdFullUninstall, strings.Join(args, " "), stdoutPath)

	return scriptName, scriptContent, nil
}

func appendNodeDownloadArg(args []string, skipDownload bool, downloadSvrAddr string) []string {
	if skipDownload {
		return append(args, fmt.Sprintf("--%s", nodeFlagSkipDownload))
	}

	return append(args, fmt.Sprintf("--%s %s", nodeFlagDownloadSvrAddr, downloadSvrAddr))
}

func appendNodeCallbackArg(args []string, skipCallback bool, callbackSvrAddr string) []string {
	if skipCallback {
		return append(args, fmt.Sprintf("--%s", nodeFlagSkipCallback))
	}

	return append(args, fmt.Sprintf("--%s %s", nodeFlagCallbackSvrAddr, callbackSvrAddr))
}

func appendNodeInstallOptionalArgs(args []string, logToStd bool, agentID string, renewGSEProc bool, renewGSETask bool) []string {
	if logToStd {
		args = append(args, fmt.Sprintf("--%s", nodeFlagLogToStd))
	}

	if agentID != "" {
		args = append(args, fmt.Sprintf("--%s %s", nodeFlagAgentID, agentID))
	}

	if renewGSEProc {
		args = append(args, fmt.Sprintf("--%s", nodeFlagRenewGSEProc))
	}

	if renewGSETask {
		args = append(args, fmt.Sprintf("--%s", nodeFlagRenewGSETask))
	}

	return args
}

func validateNodeInstallerPath(installerPath string) error {
	if installerPath == "" {
		return fmt.Errorf("installer path is empty")
	}

	return nil
}

func validateNodeInstallerFile(installWorkDir string, installerFileName string) error {
	if installWorkDir == "" {
		return fmt.Errorf("install work dir is empty")
	}

	if installerFileName == "" {
		return fmt.Errorf("installer file name is empty")
	}

	return nil
}

func validateNodeInstallCommandParams(
	deployToken string,
	nodeVersion string,
	operInstID string,
	downloadSvrAddr string,
	callbackSvrAddr string,
	skipDownload bool,
	skipCallback bool,
) error {
	if deployToken == "" {
		return fmt.Errorf("deploy token is empty")
	}

	if nodeVersion == "" {
		return fmt.Errorf("node version is empty")
	}

	if operInstID == "" {
		return fmt.Errorf("operation instance id is empty")
	}

	if !skipDownload && downloadSvrAddr == "" {
		return fmt.Errorf("download server address is empty")
	}

	if !skipCallback && callbackSvrAddr == "" {
		return fmt.Errorf("callback server address is empty")
	}

	return nil
}
