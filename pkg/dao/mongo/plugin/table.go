/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package plugin

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
)

// TableName plugin table name.
func TableName(tenantID string, pluginType string) string {
	return fmt.Sprintf("plugin_%s_%s", pluginType, tenantID)
}

var _ base.IData = &Plugin{}

// Plugin represents the table of plugin deployment.
// Token should be the unique key.
type Plugin struct {
	TenantID string        `json:"tenant_id" bson:"tenant_id"`
	PluginID string        `json:"plugin_id" bson:"plugin_id"`
	Static   pluginStatic  `json:"static" bson:"static"`
	Dynamic  pluginDynamic `json:"dynamic" bson:"dynamic"`
}

// pluginStatic represents the static data of plugin.
type pluginStatic struct {
	Info          ProcessInfo          `json:"info" bson:"info"`
	Identity      processIdentity      `json:"identity" bson:"identity"`
	Controller    processController    `json:"controller" bson:"controller"`
	Resource      processResource      `json:"resource" bson:"resource"`
	MonitorPolicy processMonitorPolicy `json:"monitor_policy" bson:"monitor_policy"`
}

// ProcessInfo represents the info of process.
type ProcessInfo struct {
	Pid         int    `json:"pid" bson:"pid"`
	Version     string `json:"version" bson:"version"`
	AgentID     string `json:"agent_id" bson:"agent_id"`
	Trusteeship bool   `json:"trusteeship" bson:"trusteeship"`
	Status      string `json:"status" bson:"status"`
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
	CPU float64 `json:"cpu" bson:"cpu"`
	Mem float64 `json:"mem" bson:"mem"`
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
	Platform   Platform `json:"platform" bson:"platform"`
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

// UniqueFields unique fields of the table.
func (deploy *Plugin) UniqueFields() []string {
	return []string{FieldKeyPluginID}
}

// UniqueKey unique key of the table.
func (deploy *Plugin) UniqueKey() string {
	return deploy.PluginID
}

// Table represent the complete db structures of plugin deployment.
type Table base.TableBroker[*Plugin]
