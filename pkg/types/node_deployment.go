/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package types

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/google/uuid"
)

// NewNodeDeployment this is the info for node deployment.
func NewNodeDeployment(info *DeploymentInfo) *NodeDeployment {
	return &NodeDeployment{
		Token: strings.ReplaceAll(uuid.New().String(), "-", ""),
		Info:  info,
		NodeConf: &NodeConf{
			PreSetting:    make(map[string]any),
			CustomSetting: make(map[string]any),
		},
	}
}

// NodeDeployment this is the info for node deployment.
type NodeDeployment struct {
	Token    string
	Info     *DeploymentInfo
	NodeConf *NodeConf
}

// Validate this is the validate for node deployment.
func (deployment NodeDeployment) Validate() error {
	if err := deployment.Info.Validate(); err != nil {
		return err
	}

	return nil
}

// LoginMode this is the login mode for node deployment.
type LoginMode string

const (
	// LoginModePasswordVault this mode means system will automic query password to login.
	LoginModePasswordVault LoginMode = "password_vault"

	// LoginModePassword this mode means system will use password to login.
	LoginModePassword LoginMode = "password"

	// LoginModeKeyFile this mode means system will use key file to login.
	LoginModeKeyFile LoginMode = "keyfile"
)

// Validate this is the validate for login mode.
func (mode LoginMode) Validate() error {
	switch mode {
	case LoginModePassword, LoginModeKeyFile, LoginModePasswordVault:
		return nil
	default:
		return fmt.Errorf("invalid login mode: %s", mode)
	}
}

// DeploymentInstallerRuntime this is the installer runtime for node deployment.
type DeploymentInstallerRuntime struct {
	BaseWorkDir string
	WorkDir     string
}

// DeploymentBaseRuntime  this is the base runtime for deployment.
type DeploymentBaseRuntime struct {
	BaseDeployDir     string
	DeployDir         string
	HomeDir           string
	DataIPC           string
	PluginIPC         string
	ExtraConfigDir    string
	LogDir            string
	ProxyFileCacheDir string
}

// DeploymentInstallOptions this is the options for nodemgr tools.
type DeploymentInstallOptions struct {
	ReRegister               bool
	RenewGSETask             bool
	RenewGSEProc             bool
	InstallPreOrderedPlugins bool
	DirectInstall            bool
	IsManual                 bool
	IsOffline                bool
	EnableCompatibilityMode  bool
}

// DeploymentUpgradeOptions this is the options for node upgrade.
type DeploymentUpgradeOptions struct {
	DirectLink bool
}

// DeploymentReconfigOptions this is the options for node reconfig.
type DeploymentReconfigOptions struct {
	DirectLink           bool
	AllowReleaseFallback bool
}

// DeploymentUninstallOptions this is the options for node uninstall.
type DeploymentUninstallOptions struct {
	DirectLink bool
}

// DeploymentRestartOptions this is the options for node restart.
type DeploymentRestartOptions struct {
	ForceRestart           bool
	GracefulRestartTimeout time.Duration
}

// DeploymentTransferOptions this is the options for node transfer.
type DeploymentTransferOptions struct {
	// SelectDownloads set false by default, will download all things.
	// set true, then will only download the enabled ones following.
	SelectDownloads bool

	EnableReleasePackage bool
	EnableInstaller      bool
}

// DeploymentVersionSupports describes this version supports things.
type DeploymentVersionSupports struct {
	// OperateAgentRestart means this version agent supports operates restart through cluster,
	// which will check if the agent is idle.
	OperateAgentRestart bool
}

// DeploymentInfo this is the info for node deployment.
type DeploymentInfo struct {
	BlockingActionName string
	Host               Host
	RelayInfo          RelayInfo

	// InstallerRuntime is used to store the installer runtime.
	InstallerRuntime DeploymentInstallerRuntime

	// BaseRuntime is used to store the base runtime.
	BaseRuntime DeploymentBaseRuntime

	// CurrentVersionSupports is used to mark the source agent(before any workflow) supports things.
	CurrentVersionSupports DeploymentVersionSupports

	// InstallOptions is used to control the tools when install node.
	InstallOptions DeploymentInstallOptions

	// UpgradeOptions is used to control the tools when upgrade node.
	UpgradeOptions DeploymentUpgradeOptions

	// RestartOptions is used to control the tools when restart node.
	RestartOptions DeploymentRestartOptions

	// ReconfigOptions is used to control the tools when reconfig node.
	ReconfigOptions DeploymentReconfigOptions

	// UninstallOptions is used to control the tools when uninstall node.
	UninstallOptions DeploymentUninstallOptions

	// TransferOptions is used to control the tools when transfer node.
	TransferOptions DeploymentTransferOptions

	// TargetVersion is used to control the target version for node.
	TargetVersion []TargetVersion
}

// Validate this is the validate for node deployment.
// nolint: gocognit,gocyclo,cyclop
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (info DeploymentInfo) Validate() error {
	if info.BlockingActionName == "" {
		return errors.New("blocking_action_name shouldn't not be empty")
	}

	if info.Host.HostID < 0 {
		return errors.New("host_id should be equal or greater than 0")
	}

	if info.Host.Static.OSType == "" {
		return errors.New("os_type shouldn't not be empty")
	}

	if info.Host.TenantID == "" {
		return errors.New("tenant_id shouldn't not be empty")
	}

	if err := info.Host.Dynamic.NodeRole.Validate(); err != nil {
		return fmt.Errorf("node_role validate failed: %w", err)
	}

	if err := info.Host.Dynamic.NodeStatus.Validate(); err != nil {
		return fmt.Errorf("node_status validate failed: %w", err)
	}

	if info.Host.Dynamic.NodeVersion == "" {
		return errors.New("node_version shouldn't not be empty")
	}
	if err := info.Host.Dynamic.NodeGeneration.Validate(); err != nil {
		return fmt.Errorf("node_generation validate failed: %w", err)
	}
	if info.Host.Dynamic.NetworkUnitID < 0 {
		return errors.New("network_unit_id should be equal or greater than 0")
	}
	if info.Host.Static.BizID < 0 {
		return errors.New("biz_id should be equal or greater than 0")
	}
	if info.Host.Static.NetworkAreaID < 0 {
		return errors.New("networkarea_id should be equal or greater than 0")
	}
	if len(info.Host.Static.InnerIPList) == 0 {
		return errors.New("inner_ip shouldn't not be empty")
	}
	if info.Host.Static.Addressing == "" {
		return errors.New("addressing shouldn't not be empty")
	}

	// nolint: nestif
	if info.Host.Dynamic.NodeRole == NodeRoleProxy {
		if info.Host.Dynamic.ProxyClusterPort > 0 {
			return errors.New("proxy_cluster_port should be 0")
		}
		if info.Host.Dynamic.ProxyDataPort > 0 {
			return errors.New("proxy_data_port should be 0")
		}
		if info.Host.Dynamic.ProxyFilePort > 0 {
			return errors.New("proxy_file_port should be 0")
		}
	} else {
		if info.Host.Dynamic.ProxyClusterPort != 0 {
			return errors.New("node_role is not proxy, proxy_cluster_port should be 0")
		}

		if info.Host.Dynamic.ProxyDataPort != 0 {
			return errors.New("node_role is not proxy, proxy_data_port should be 0")
		}

		if info.Host.Dynamic.ProxyFilePort != 0 {
			return errors.New("node_role is not proxy, proxy_file_port should be 0")
		}
	}

	if err := info.Host.Dynamic.LoginMode.Validate(); err != nil {
		return fmt.Errorf("login_mode validate failed: %w", err)
	}

	return nil
}

const (
	// ConfigKeyAgent agent config key.
	ConfigKeyAgent = "agent"

	// ConfigKeyFile file config key.
	ConfigKeyFile = "file"

	// ConfigKeyData data config key.
	ConfigKeyData = "data"
)

// NodeConf this is the node conf for node deployment.
type NodeConf struct {
	ConfigTemplate map[string]string
	PreSetting     map[string]any
	CustomSetting  map[string]any
}

// TargetVersion this is the target version for node.
type TargetVersion struct {
	OsType  criteria.OSType
	CPUArch criteria.CPUArch
	Version string
}

// Validate validate target version.
func (v TargetVersion) Validate() error {
	if err := v.OsType.Validate(); err != nil {
		return fmt.Errorf("os_type validate failed: %w", err)
	}

	if err := v.CPUArch.Validate(); err != nil {
		return fmt.Errorf("cpu_arch validate failed: %w", err)
	}

	if v.Version == "" {
		return errors.New("version shouldn't not be empty")
	}

	return nil
}
