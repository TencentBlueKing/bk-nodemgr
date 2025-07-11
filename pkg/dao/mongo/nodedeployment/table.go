/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package nodedeployment

import (
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
)

// TableName node deployment table name.
const TableName = "node_deployment"

// Data represents the table of node deployment.
// Token should be the unique key.
type Data struct {
	Token    string    `json:"token" bson:"token"`
	Info     *Info     `json:"info" bson:"info"`
	NodeConf *NodeConf `json:"node_conf" bson:"node_conf"`
}

// Info this is the info of this node deployment.
type Info struct {
	ActionName         string          `json:"action_name" bson:"action_name"`
	HostID             int64           `json:"host_id" bson:"host_id"`
	OSType             string          `json:"os_type" bson:"os_type"`
	TenantID           string          `json:"tenant_id" bson:"tenant_id"`
	NodeRole           string          `json:"node_role" bson:"node_role"`
	NodeStatus         string          `json:"node_status" bson:"node_status"`
	NodeVersion        string          `json:"node_version" bson:"node_version"`
	NodeGeneration     int64           `json:"node_generation" bson:"node_generation"`
	NodeCPUArch        string          `json:"node_cpu_arch" bson:"node_cpu_arch"`
	NodeOsType         string          `json:"node_os_type" bson:"node_os_type"`
	AgentID            string          `json:"agent_id" bson:"agent_id"`
	NetworkUnitID      int64           `json:"network_unit_id" bson:"network_unit_id"`
	NetworkAreaID      int64           `json:"network_area_id" bson:"network_area_id"`
	BizID              int64           `json:"biz_id" bson:"biz_id"`
	InnerIP            string          `json:"inner_ip" bson:"inner_ip"`
	Addressing         string          `json:"addressing" bson:"addressing"`
	ProxyTags          []string        `json:"proxy_tags" bson:"proxy_tags"`
	ProxyClusterPort   int64           `json:"proxy_cluster_port" bson:"proxy_cluster_port"`
	ProxyDataPort      int64           `json:"proxy_data_port" bson:"proxy_data_port"`
	ProxyFilePort      int64           `json:"proxy_file_port" bson:"proxy_file_port"`
	InstallerWorkspace string          `json:"installer_workspace" bson:"installer_workspace"`
	LoginInfo          LoginInfo       `json:"login_info" bson:"login_info"`
	InstallOptions     InstallOptions  `json:"install_options" bson:"install_options"`
	UpgradeOptions     UpgradeOptions  `json:"upgrade_options" bson:"upgrade_options"`
	TargetVersion      []TargetVersion `json:"target_version" bson:"target_version"`
}

// TargetVersion this is the target version for node deployment.
type TargetVersion struct {
	OsType  string `json:"os_type" bson:"os_type"`
	CPUArch string `json:"cpu_arch" bson:"cpu_arch"`
	Version string `json:"version" bson:"version"`
}

// LoginInfo this is the login info for node deployment.
type LoginInfo struct {
	IP       string `json:"ip" bson:"ip"`
	Port     int64  `json:"port" bson:"port"`
	User     string `json:"user" bson:"user"`
	Mode     string `json:"mode" bson:"mode"`
	Password []byte `json:"password" bson:"password"`
	KeyFile  []byte `json:"key_file" bson:"key_file"`
}

// InstallOptions this is the options for nodemgr tools.
type InstallOptions struct {
	ReRegister bool `json:"re_register" bson:"re_register"`
}

// UpgradeOptions this is the options for node upgrade.
type UpgradeOptions struct {
	ForceRestart           bool          `json:"force_restart" bson:"force_restart"`
	GracefulRestartTimeout time.Duration `json:"graceful_restart_timeout" bson:"graceful_restart_timeout"`
}

// NodeConf this is the node conf for node deployment.
type NodeConf struct {
	PreSetting    map[string]any `json:"pre_setting" bson:"pre_setting"`
	CustomSetting map[string]any `json:"custom_setting" bson:"custom_setting"`
}

// UniqueKey unique key of the table.
func (deploy *Data) UniqueKey() string {
	return deploy.Token
}

// Table represent the complete db structures of node deployment.
type Table base.TableBroker[*Data]
