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
	Token      string `json:"token" bson:"token"`
	Info       *Info  `json:"info" bson:"info"`
	MainConfig []byte `json:"main_config" bson:"main_config"`
}

// Info this is the info of this plugin deployment.
type Info struct {
	ActionName       string          `json:"action_name" bson:"action_name"`
	InstallerWorkDir string          `json:"installer_work_dir" bson:"installer_work_dir"`
	Plugin           plugin          `json:"plugin" bson:"plugin"`
	TransferOptions  transferOptions `json:"transfer_options" bson:"transfer_options"`
	InstallOptions   installOptions  `json:"install_options" bson:"install_options"`
	TargetVersion    []targetVersion `json:"target_version" bson:"target_version"`
}

// plugin represents the table of plugin deployment.
// Token should be the unique key.
type plugin struct {
	TenantID string        `json:"tenant_id" bson:"tenant_id"`
	PluginID string        `json:"plugin_id" bson:"plugin_id"`
	Static   pluginStatic  `json:"static" bson:"static"`
	Dynamic  pluginDynamic `json:"dynamic" bson:"dynamic"`
}

// pluginStatic represents the static data of plugin.
type pluginStatic struct {
	Info ProcessInfo `json:"info" bson:"info"`
	Spec processSpec `json:"spec" bson:"spec"`
}

// ProcessInfo represents the info of process.
type ProcessInfo struct {
	Pid         int    `json:"pid" bson:"pid"`
	Version     string `json:"version" bson:"version"`
	AgentID     string `json:"agent_id" bson:"agent_id"`
	Trusteeship bool   `json:"trusteeship" bson:"trusteeship"`
	Status      string `json:"status" bson:"status"`
}

// processSpec represents the spec of process.
type processSpec struct {
	AgentID       string               `json:"agent_id" bson:"agent_id"`
	Identity      processIdentity      `json:"identity" bson:"identity"`
	Controller    processController    `json:"controller" bson:"controller"`
	Resource      processResource      `json:"resource" bson:"resource"`
	MonitorPolicy processMonitorPolicy `json:"monitor_policy" bson:"monitor_policy"`
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
	KillCmd    string `json:"version_cmd" bson:"version_cmd"`
	VersionCmd string `json:"kill_cmd" bson:"kill_cmd"`
	HealthCmd  string `json:"health_cmd" bson:"health_cmd"`
}

type processResource struct {
	CPULimitPercent float64 `json:"cpu_limit_percent" bson:"cpu_limit_percent"`
	MemLimitPercent float64 `json:"mem_limit_percent" bson:"mem_limit_percent"`
}

type processMonitorPolicy struct {
	AutoType       string `json:"auto_type" bson:"auto_type"`
	StartCheckSecs int64  `json:"start_check_secs" bson:"start_check_secs"`
	StopCheckSecs  int64  `json:"stop_check_secs" bson:"stop_check_secs"`
	OpTimeoutSecs  int64  `json:"op_timeout_secs" bson:"op_timeout_secs"`
}

// pluginDynamic represents the dynamic data of plugin.
type pluginDynamic struct {
	Name       string   `json:"name" bson:"name"`
	Type       string   `json:"type" bson:"type"`
	Generation int64    `json:"generation" bson:"generation"`
	Platform   Platform `json:"Platform" bson:"Platform"`
	Version    string   `json:"version" bson:"version"`
	HostID     int64    `json:"host_id" bson:"host_id"`
	AgentID    string   `json:"agent_id" bson:"agent_id"`
	Status     string   `json:"status" bson:"status"`
}

// Platform defines the Platform.
type Platform struct {
	OS   string `json:"os" bson:"os"`
	Arch string `json:"arch" bson:"arch"`
}

// targetVersion defines the target version.
type targetVersion struct {
	Platform Platform `json:"Platform" bson:"Platform"`
	Version  string   `json:"version" bson:"version"`
}

// installOptions this is the options for nodemgr tools.
type installOptions struct {
}

// transferOptions this is the options for plugin transfer.
type transferOptions struct {
	SelectDownloads      bool `json:"select_downloads" bson:"select_downloads"`
	EnableReleasePackage bool `json:"enable_release_package" bson:"enable_release_package"`
	EnableInstaller      bool `json:"enable_installer" bson:"enable_installer"`
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
