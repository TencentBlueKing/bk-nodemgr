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

package process

import (
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
)

// TableName process table name.
func TableName(tenantID string) string {
	return fmt.Sprintf("process_%s", tenantID)
}

var _ base.IData = &Process{}

// Process represents the table of process deployment.
// Token should be the unique key.
type Process struct {
	TenantID      string `json:"tenant_id" bson:"tenant_id"`
	HostID        int64  `json:"host_id" bson:"host_id"`
	BizID         int64  `json:"biz_id" bson:"biz_id"`
	Name          string `json:"name" bson:"name"`
	Group         string `json:"group" bson:"group"`
	PluginPkgName string `json:"plugin_pkg_name" bson:"plugin_pkg_name"`

	Generation int64    `json:"generation" bson:"generation"`
	Platform   platform `json:"platform" bson:"platform"`

	BindIP   string `json:"bind_ip" bson:"bind_ip"`
	BindPort int    `json:"bind_port" bson:"bind_port"`

	Info          processInfo          `json:"info" bson:"info"`
	Identity      processIdentity      `json:"identity" bson:"identity"`
	Controller    processController    `json:"controller" bson:"controller"`
	Resource      processResource      `json:"resource" bson:"resource"`
	MonitorPolicy processMonitorPolicy `json:"monitor_policy" bson:"monitor_policy"`
}

// processInfo represents the info of process.
type processInfo struct {
	Pid         int       `json:"pid" bson:"pid"`
	Version     string    `json:"version" bson:"version"`
	AgentID     string    `json:"agent_id" bson:"agent_id"`
	Trusteeship bool      `json:"trusteeship" bson:"trusteeship"`
	Status      string    `json:"status" bson:"status"`
	LastSyncAt  time.Time `json:"last_sync_at" bson:"last_sync_at"`
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
	DebugCmd   string `json:"debug_cmd" bson:"debug_cmd"`
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

// platform defines the platform.
type platform struct {
	OS   string `json:"os" bson:"os"`
	Arch string `json:"arch" bson:"arch"`
}

// UniqueFields unique fields of the table.
func (proc *Process) UniqueFields() []string {
	return []string{
		FieldKeyHostID,
		FieldKeyPluginName,
	}
}

// UniqueKey unique key of the table.
func (proc *Process) UniqueKey() string {
	return fmt.Sprintf("%d:%s", proc.HostID, proc.Name)
}

// Table represent the complete db structures of process deployment.
type Table base.TableBroker[*Process]
