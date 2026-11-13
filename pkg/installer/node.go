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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
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

	// nodeFlagHTTPProxy node flag name defines the http proxy override for all http requests.
	// SYNC: tools/cmd/installer/node/flag.HTTPProxy.
	nodeFlagHTTPProxy nodeFlagName = "http_proxy"

	// nodeFlagFirewallFeatures node flag name defines the firewall features of a full install.
	// SYNC: tools/cmd/installer/node/flag.FirewallFeatures.
	nodeFlagFirewallFeatures nodeFlagName = "firewall_features"

	// nodeFlagFirewallRemoteIPs node flag name defines the firewall remote IP CIDRs of a full install.
	// SYNC: tools/cmd/installer/node/flag.FirewallRemoteIPs.
	nodeFlagFirewallRemoteIPs nodeFlagName = "firewall_remote_ips"

	// nodeFlagWorkDir node flag name defines the work dir overriding the base/env derived one.
	// SYNC: tools/cmd/installer/node/flag.WorkDir.
	nodeFlagWorkDir nodeFlagName = "work_dir"
)

// NodeCommonParams defines the common installer-facing params shared by node commands.
type NodeCommonParams struct {
	DeployEnv     string
	Generation    int
	NodeRole      string
	BaseWorkDir   string
	BaseDeployDir string

	// InstallerWorkDir is the full installer workdir carrying the per-user isolation segment;
	// when set, it is passed to the installer via --work_dir so the host-side workspace
	// matches the orchestrator instead of the legacy base/env derived one.
	InstallerWorkDir string

	// HTTPProxyOverride whether the installer overrides the target host's env proxy settings.
	HTTPProxyOverride bool

	// HTTPProxyAddress is the proxy URL applied when HTTPProxyOverride is true.
	// Empty address means explicitly skipping the proxy.
	HTTPProxyAddress string

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
	args := []string{
		fmt.Sprintf("--%s %s", nodeFlagDeployEnv, params.DeployEnv),
		fmt.Sprintf("--%s %d", nodeFlagGeneration, params.Generation),
		fmt.Sprintf("--%s %s", nodeFlagNodeRole, params.NodeRole),
		fmt.Sprintf("--%s %s", nodeFlagBaseWorkDir, params.BaseWorkDir),
		fmt.Sprintf("--%s %s", nodeFlagBaseDeployDir, params.BaseDeployDir),
	}
	if params.HTTPProxyOverride {
		args = append(args, fmt.Sprintf(`--%s="%s"`, nodeFlagHTTPProxy, params.HTTPProxyAddress))
	}
	if params.InstallerWorkDir != "" {
		args = append(args, fmt.Sprintf("--%s %s", nodeFlagWorkDir, params.InstallerWorkDir))
	}

	return args
}

// unixWorkDir returns the unix workdir of the params, preferring the explicit
// InstallerWorkDir over the legacy base/env derived one.
func (params *NodeCommonParams) unixWorkDir() string {
	if params.InstallerWorkDir != "" {
		return params.InstallerWorkDir
	}

	return fmt.Sprintf("%s/%s", params.BaseWorkDir, params.DeployEnv)
}

// windowsWorkDir returns the windows workdir of the params, preferring the explicit
// InstallerWorkDir over the legacy base/env derived one.
func (params *NodeCommonParams) windowsWorkDir() string {
	if params.InstallerWorkDir != "" {
		return params.InstallerWorkDir
	}

	return winpath.Join(params.BaseWorkDir, params.DeployEnv)
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

	// FirewallFeatures are the enabled firewall features (gse, nodemgr), separated by comma.
	FirewallFeatures string

	// FirewallRemoteIPs are the firewall remote IP CIDRs, separated by comma.
	FirewallRemoteIPs string

	// SudoInstall escalates the installer invocation through `sudo -n` when rendering
	// unix scripts, for installs driven by a non-root login user with passwordless sudo.
	// It only affects unix rendering (ToUnixScript*); windows scripts ignore it.
	SudoInstall bool
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
	args = appendNodeFirewallArgs(args, params.FirewallFeatures, params.FirewallRemoteIPs)

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
	args = appendNodeFirewallArgs(args, params.FirewallFeatures, params.FirewallRemoteIPs)

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
	args = appendNodeFirewallArgs(args, params.FirewallFeatures, params.FirewallRemoteIPs)

	return append(args, params.AdditionArgs...)
}

// NodeUpgradeParams defines the params for a full node upgrade command.
type NodeUpgradeParams struct {
	NodeCommonParams

	InstallerFileName string

	DownloadSvrAddr string
	CallbackSvrAddr string
	NodeVersion     string
	DeployToken     string
	OperInstID      string

	Force        bool
	Restart      bool
	SkipDownload bool

	// FirewallFeatures are the enabled firewall features (gse, nodemgr), separated by comma.
	FirewallFeatures string

	// FirewallRemoteIPs are the firewall remote IP CIDRs, separated by comma.
	FirewallRemoteIPs string
}

// Validate validates the full node upgrade params required by the installer CLI.
func (params *NodeUpgradeParams) Validate() error {
	if err := params.NodeCommonParams.Validate(); err != nil {
		return err
	}

	if err := validateNodeInstallerFile(params.InstallerWorkDir, params.InstallerFileName); err != nil {
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
	args = appendNodeFirewallArgs(args, params.FirewallFeatures, params.FirewallRemoteIPs)

	return append(args, params.AdditionArgs...)
}

// NodeReconfigParams defines the params for a full node reconfig command.
type NodeReconfigParams struct {
	NodeCommonParams

	InstallerFileName string

	CallbackSvrAddr string
	DeployToken     string
	OperInstID      string

	Force   bool
	Restart bool

	// FirewallFeatures are the enabled firewall features (gse, nodemgr), separated by comma.
	FirewallFeatures string

	// FirewallRemoteIPs are the firewall remote IP CIDRs, separated by comma.
	FirewallRemoteIPs string
}

// Validate validates the full node reconfig params required by the installer CLI.
func (params *NodeReconfigParams) Validate() error {
	if err := params.NodeCommonParams.Validate(); err != nil {
		return err
	}

	if err := validateNodeInstallerFile(params.InstallerWorkDir, params.InstallerFileName); err != nil {
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
	args = appendNodeFirewallArgs(args, params.FirewallFeatures, params.FirewallRemoteIPs)

	return append(args, params.AdditionArgs...)
}

// NodeUninstallParams defines the params for a full node uninstall command.
type NodeUninstallParams struct {
	NodeCommonParams

	InstallerFileName string

	CallbackSvrAddr string
	DeployToken     string
	OperInstID      string
	SkipCallback    bool

	// FirewallFeatures are the enabled firewall features (gse, nodemgr), separated by comma.
	FirewallFeatures string

	// FirewallRemoteIPs are the firewall remote IP CIDRs, separated by comma.
	FirewallRemoteIPs string
}

// Validate validates the full node uninstall params required by the installer CLI.
func (params *NodeUninstallParams) Validate() error {
	if err := params.NodeCommonParams.Validate(); err != nil {
		return err
	}

	if err := validateNodeInstallerFile(params.InstallerWorkDir, params.InstallerFileName); err != nil {
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
	args = appendNodeFirewallArgs(args, params.FirewallFeatures, params.FirewallRemoteIPs)

	return append(args, params.AdditionArgs...)
}

// NodeStepRestartParams defines the params for a node restart step command.
type NodeStepRestartParams struct {
	NodeCommonParams

	InstallerFileName string

	Force bool
}

// Validate validates the node restart step params required by the installer CLI.
func (params *NodeStepRestartParams) Validate() error {
	if err := params.NodeCommonParams.Validate(); err != nil {
		return err
	}

	return validateNodeInstallerFile(params.InstallerWorkDir, params.InstallerFileName)
}

// NodeStepCleanTmpParams defines the params for a clean temporary installer files step command.
type NodeStepCleanTmpParams struct {
	NodeCommonParams

	InstallerFileName string
}

// Validate validates the clean temporary installer files step params required by the installer CLI.
func (params *NodeStepCleanTmpParams) Validate() error {
	if err := params.NodeCommonParams.Validate(); err != nil {
		return err
	}

	return validateNodeInstallerFile(params.InstallerWorkDir, params.InstallerFileName)
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
	installerPath := filepath.Clean(filepath.Join(params.InstallerWorkDir, params.InstallerFileName))
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
	installerPath := winpath.Clean(winpath.Join(params.InstallerWorkDir, params.InstallerFileName))
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
	installerPath := filepath.Clean(filepath.Join(params.InstallerWorkDir, params.InstallerFileName))
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
	installerPath := winpath.Clean(winpath.Join(params.InstallerWorkDir, params.InstallerFileName))
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

	// FirewallFeatures are the enabled firewall features (gse, nodemgr), separated by comma.
	FirewallFeatures string

	// FirewallRemoteIPs are the firewall remote IP CIDRs, separated by comma.
	FirewallRemoteIPs string

	// SudoInstall escalates the installer invocation through `sudo -n` in the generated
	// offline install.sh, for hosts operated by a non-root user with passwordless sudo.
	SudoInstall bool
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
	args = appendNodeFirewallArgs(args, params.FirewallFeatures, params.FirewallRemoteIPs)

	return append(args, params.AdditionArgs...)
}

// ToUnixScriptOffline converts the offline install params to a unix install.sh script.
func (params *NodeOfflineInstallParams) ToUnixScriptOffline() (string, string, error) {
	if err := params.Validate(); err != nil {
		return "", "", err
	}

	dataDir := fmt.Sprintf("%s/data", params.unixWorkDir())
	lines := []string{
		"#!/bin/bash",
		`set -e`,
		`SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"`,
		fmt.Sprintf(`DATA_DIR="%s"`, dataDir),
		`mkdir -p "${DATA_DIR}"`,
		fmt.Sprintf(`cp -rf "${SCRIPT_DIR}/%s/." "${DATA_DIR}/"`, OfflinePkgRelPathData),
		fmt.Sprintf(`chmod +x "${SCRIPT_DIR}/%s"`, params.InstallerFileName),
		unixInstallerBin(params.SudoInstall, fmt.Sprintf(`"${SCRIPT_DIR}/%s" `, params.InstallerFileName)) + NodeCmdFullInstall + ` \`,
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

// unixInstallerBin returns the installer binary invocation for unix scripts,
// escalated through `sudo -n` when sudoInstall is enabled.
func unixInstallerBin(sudoInstall bool, installerPath string) string {
	if !sudoInstall {
		return installerPath
	}

	return "sudo -n " + installerPath
}

// ToUnixScript converts the install params to a unix script.
func (params *NodeInstallParams) ToUnixScript() (string, string, error) {
	if err := params.Validate(); err != nil {
		return "", "", err
	}

	args := params.buildArgs()

	scriptName := "install.sh"
	stdoutPath := filepath.Clean(fmt.Sprintf("%s.stdout", params.InstallerPath))
	installerBin := unixInstallerBin(params.SudoInstall, params.InstallerPath)
	scriptContent := fmt.Sprintf("%s %s %s >%s 2>&1 &", installerBin, NodeCmdFullInstall, strings.Join(args, " "), stdoutPath)

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
	installerBin := unixInstallerBin(params.SudoInstall, params.InstallerPath)
	scriptContent := fmt.Sprintf("%s %s %s >%s 2>&1 &", installerBin, NodeCmdFullInstall, strings.Join(args, " "), stdoutPath)

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
	scriptContent := fmt.Sprintf("cd %s && %s %s %s >%s 2>&1", params.windowsWorkDir(), params.InstallerPath, NodeCmdFullInstall, strings.Join(args, " "), stdoutPath)

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
	workDir := winpath.ToSlash(winpath.Clean(params.windowsWorkDir()))
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
	installerBin := unixInstallerBin(params.SudoInstall, params.InstallerPath)
	scriptContent := fmt.Sprintf("%s %s %s", installerBin, NodeCmdFullInstall, strings.Join(args, " "))

	return scriptName, scriptContent, nil
}

// ToWindowsScriptManual converts the install params to a windows manual install script.
func (params *NodeInstallParams) ToWindowsScriptManual() (string, string, error) {
	if err := params.Validate(); err != nil {
		return "", "", err
	}

	args := params.buildServerFirstArgs()

	scriptName := "install.bat"
	scriptContent := fmt.Sprintf("cd %s && %s %s %s", params.windowsWorkDir(), params.InstallerPath, NodeCmdFullInstall, strings.Join(args, " "))

	return scriptName, scriptContent, nil
}

// ToUnixScript converts the upgrade params to a unix script.
func (params *NodeUpgradeParams) ToUnixScript() (string, string, error) {
	if err := params.Validate(); err != nil {
		return "", "", err
	}

	args := params.buildArgs()
	installerPath := filepath.Clean(filepath.Join(params.InstallerWorkDir, params.InstallerFileName))
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
	installerPath := winpath.Clean(winpath.Join(params.InstallerWorkDir, params.InstallerFileName))
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
	installerPath := filepath.Clean(filepath.Join(params.InstallerWorkDir, params.InstallerFileName))
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
	installerPath := winpath.Clean(winpath.Join(params.InstallerWorkDir, params.InstallerFileName))
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
	installerPath := filepath.Clean(filepath.Join(params.InstallerWorkDir, params.InstallerFileName))
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
	installerPath := winpath.Clean(winpath.Join(params.InstallerWorkDir, params.InstallerFileName))
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

func appendNodeFirewallArgs(args []string, features string, remoteIPs string) []string {
	if features != "" {
		args = append(args, fmt.Sprintf(`--%s="%s"`, nodeFlagFirewallFeatures, features))
	}

	if remoteIPs != "" {
		args = append(args, fmt.Sprintf(`--%s="%s"`, nodeFlagFirewallRemoteIPs, remoteIPs))
	}

	return args
}

// NodeFirewallFlagValues renders a firewall snapshot into the --firewall_features and --firewall_remote_ips
// flag values of a full install command. A nil snapshot renders empty values, which the installer skips.
func NodeFirewallFlagValues(firewall *types.DeploymentFirewall) (features string, remoteIPs string) {
	if firewall == nil {
		return "", ""
	}

	featureList := make([]string, 0, len(firewall.Features))
	for _, feature := range firewall.Features {
		featureList = append(featureList, string(feature))
	}

	return strings.Join(featureList, ","), strings.Join(firewall.RemoteIPs, ",")
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
