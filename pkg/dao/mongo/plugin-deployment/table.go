/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package plugindeployment

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
)

// TableName plugin deployment table name.
const TableName = "plugin_deployment"

var _ base.IData = &Data{}

// Data represents the table of plugin deployment.
// Token should be the unique key.
type Data struct {
	Token      string      `json:"token" bson:"token"`
	Info       *Info       `json:"info" bson:"info"`
	PluginConf *PluginConf `json:"plugin_config" bson:"plugin_config"`
}

// Info this is the info of this plugin deployment.
type Info struct {
	ActionName       string           `json:"action_name" bson:"action_name"`
	Process          process          `json:"process" bson:"process"`
	InstallerRuntime installerRuntime `json:"installer_runtime" bson:"installer_runtime"`
	BaseRuntime      baseRuntime      `json:"base_runtime" bson:"base_runtime"`
	TransferOptions  transferOptions  `json:"transfer_options" bson:"transfer_options"`
	InstallOptions   installOptions   `json:"install_options" bson:"install_options"`
}

type process struct {
	TenantID string `json:"tenant_id" bson:"tenant_id"`
	HostID   int64  `json:"host_id" bson:"host_id"`
	BizID    int64  `json:"biz_id" bson:"biz_id"`

	Name    string `json:"name" bson:"name"`
	Group   string `json:"group" bson:"group"`
	PkgName string `json:"pkg_name" bson:"pkg_name"`

	Platform   platform `json:"platform" bson:"platform"`
	Generation int64    `json:"generation" bson:"generation"`

	Info          processInfo          `json:"info" bson:"info"`
	Identity      processIdentity      `json:"identity" bson:"identity"`
	Controller    processController    `json:"controller" bson:"controller"`
	Resource      processResource      `json:"resource" bson:"resource"`
	MonitorPolicy processMonitorPolicy `json:"monitor_policy" bson:"monitor_policy"`
}

type platform struct {
	OS   string `json:"os" bson:"os"`
	Arch string `json:"arch" bson:"arch"`
}

type processInfo struct {
	Pid       int    `json:"pid" bson:"pid"`
	Version   string `json:"version" bson:"version"`
	AgentID   string `json:"agent_id" bson:"agent_id"`
	AutoStart bool   `json:"auto_start" bson:"auto_start"`
	Status    string `json:"status" bson:"status"`
}
type processIdentity struct {
	Name       string `json:"name" bson:"name"`
	SetupPath  string `json:"setup_path" bson:"setup_path"`
	PidPath    string `json:"pid_path" bson:"pid_path"`
	ConfigPath string `json:"config_path" bson:"config_path"`
	LogPath    string `json:"log_path" bson:"log_path"`
	User       string `json:"user" bson:"user"`
}
type processController struct {
	StartCmd   string `json:"start_cmd" bson:"start_cmd"`
	StopCmd    string `json:"stop_cmd" bson:"stop_cmd"`
	RestartCmd string `json:"restart_cmd" bson:"restart_cmd"`
	ReloadCmd  string `json:"reload_cmd" bson:"reload_cmd"`
	KillCmd    string `json:"kill_cmd" bson:"kill_cmd"`
	VersionCmd string `json:"version_cmd" bson:"version_cmd"`
	HealthCmd  string `json:"health_cmd" bson:"health_cmd"`
}
type processResource struct {
	CPULimitPercent float64 `json:"cpu_limit_percent" bson:"cpu_limit_percent"`
	MemLimitPercent float64 `json:"mem_limit_percent" bson:"mem_limit_percent"`
}
type processMonitorPolicy struct {
	RestartType    string `json:"restart_type" bson:"restart_type"`
	StartCheckSecs int64  `json:"start_check_secs" bson:"start_check_secs"`
	StopCheckSecs  int64  `json:"stop_check_secs" bson:"stop_check_secs"`
	OpTimeoutSecs  int64  `json:"op_timeout_secs" bson:"op_timeout_secs"`
}

// installerRuntime this is the installer runtime for plugin deployment.
type installerRuntime struct {
	BaseWorkDir string `json:"base_work_dir" bson:"base_work_dir"`
	WorkDir     string `json:"work_dir" bson:"work_dir"`
}

// baseRuntime this is the base runtime for plugin deployment.
type baseRuntime struct {
	BaseDeployDir         string         `json:"base_deploy_dir" bson:"base_deploy_dir"`
	DeployDir             string         `json:"deploy_dir" bson:"deploy_dir"`
	GSEHomeDir            string         `json:"gse_home_dir" bson:"gse_home_dir"`
	PluginHomeDir         string         `json:"plugin_home_dir" bson:"plugin_home_dir"`
	DataIPC               string         `json:"data_ipc" bson:"data_ipc"`
	PluginIPC             string         `json:"plugin_ipc" bson:"plugin_ipc"`
	HostIDPath            string         `json:"host_id_path" bson:"host_id_path"`
	LogDir                string         `json:"log_dir" bson:"log_dir"`
	DataDir               string         `json:"data_dir" bson:"data_dir"`
	RunDir                string         `json:"run_dir" bson:"run_dir"`
	ConfigDir             string         `json:"config_dir" bson:"config_dir"`
	SubConfigDir          string         `json:"sub_config_dir" bson:"sub_config_dir"`
	PluginCommonConstants map[string]any `json:"plugin_common_constants" bson:"plugin_common_constants"`
	GlobalCommonConstants map[string]any `json:"global_common_constants" bson:"global_common_constants"`
}

// installOptions this is the options for nodemgr tools.
type installOptions struct {
	Version                 string `json:"version" bson:"version"`
	IsOffline               bool   `json:"is_offline" bson:"is_offline"`
	EnableCompatibilityMode bool   `json:"enable_compatibility_mode" bson:"enable_compatibility_mode"`
}

// transferOptions this is the options for plugin transfer.
type transferOptions struct {
	SelectDownloads      bool `json:"select_downloads" bson:"select_downloads"`
	EnableReleasePackage bool `json:"enable_release_package" bson:"enable_release_package"`
	EnableInstaller      bool `json:"enable_installer" bson:"enable_installer"`
}

// PluginConf defines the plugin config.
type PluginConf struct {
	TemplateRenderer    string         `json:"template_renderer" bson:"template_renderer"`
	ConfigFilesDetail   []configDetail `json:"config_files_detail" bson:"config_files_detail"`
	SystemConfigContext map[string]any `json:"system_config_context" bson:"system_config_context"`
	CustomConfigContext map[string]any `json:"custom_config_context" bson:"custom_config_context"`
}

// configDetail defines the plugin config detail.
type configDetail struct {
	Name         string `json:"name" bson:"name"`
	Content      string `json:"content" bson:"content"`
	IsMainConfig bool   `json:"is_main_config" bson:"is_main_config"`
	FilePath     string `json:"file_path" bson:"file_path"`
}

// UniqueFields unique fields of the table.
func (deploy *Data) UniqueFields() []string {
	return []string{FieldKeyToken}
}

// UniqueKey unique key of the table.
func (deploy *Data) UniqueKey() string {
	return deploy.Token
}

// Table represent the complete db structures of plugin deployment.
type Table base.TableBroker[*Data]
