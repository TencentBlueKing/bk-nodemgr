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

var _ base.IData = &Data{}

// Data represents the table of node deployment.
// Token should be the unique key.
type Data struct {
	Token    string    `json:"token" bson:"token"`
	Info     *Info     `json:"info" bson:"info"`
	NodeConf *NodeConf `json:"node_conf" bson:"node_conf"`
}

// Info this is the info of this node deployment.
type Info struct {
	ActionName             string          `json:"action_name" bson:"action_name"`
	HostID                 int64           `json:"host_id" bson:"host_id"`
	OSType                 string          `json:"os_type" bson:"os_type"`
	TenantID               string          `json:"tenant_id" bson:"tenant_id"`
	NodeRole               string          `json:"node_role" bson:"node_role"`
	NodeStatus             string          `json:"node_status" bson:"node_status"`
	NodeVersion            string          `json:"node_version" bson:"node_version"`
	NodeGeneration         int64           `json:"node_generation" bson:"node_generation"`
	NodeCPUArch            string          `json:"node_cpu_arch" bson:"node_cpu_arch"`
	NodeOsType             string          `json:"node_os_type" bson:"node_os_type"`
	AgentID                string          `json:"agent_id" bson:"agent_id"`
	NetworkUnitID          int64           `json:"networkunit_id" bson:"networkunit_id"`
	NetworkAreaID          int64           `json:"networkarea_id" bson:"networkarea_id"`
	BizID                  int64           `json:"biz_id" bson:"biz_id"`
	InnerIPList            []string        `json:"inner_ip_list" bson:"inner_ip_list"`
	Addressing             string          `json:"addressing" bson:"addressing"`
	ExportIP               string          `json:"export_ip" bson:"export_ip"`
	ExportIPV6             string          `json:"export_ipv6" bson:"export_ipv6"`
	AdvertiseIP            string          `json:"advertise_ip" bson:"advertise_ip"`
	AdvertiseIPV6          string          `json:"advertise_ipv6" bson:"advertise_ipv6"`
	ProxyTags              []string        `json:"proxy_tags" bson:"proxy_tags"`
	ProxyClusterPort       int64           `json:"proxy_cluster_port" bson:"proxy_cluster_port"`
	ProxyDataPort          int64           `json:"proxy_data_port" bson:"proxy_data_port"`
	ProxyFilePort          int64           `json:"proxy_file_port" bson:"proxy_file_port"`
	RelayDownloadPort      int64           `json:"relay_download_port" bson:"relay_download_port"`
	RelayCallbackPort      int64           `json:"relay_callback_port" bson:"relay_callback_port"`
	LoginIP                string          `json:"login_ip" bson:"login_ip"`
	LoginPort              int64           `json:"login_port" bson:"login_port"`
	LoginUser              string          `json:"login_user" bson:"login_user"`
	LoginMode              string          `json:"login_mode" bson:"login_mode"`
	LoginCreditID          string          `json:"login_credit_id" bson:"login_credit_id"`
	InstallerWorkDir       string          `json:"installer_workdir" bson:"installer_workdir"`
	InstallOptions         InstallOptions  `json:"install_options" bson:"install_options"`
	UpgradeOptions         UpgradeOptions  `json:"upgrade_options" bson:"upgrade_options"`
	RestartOptions         RestartOptions  `json:"restart_options" bson:"restart_options"`
	TransferOptions        TransferOptions `json:"transfer_options" bson:"transfer_options"`
	CurrentVersionSupports VersionSupports `json:"current_version_supports" bson:"current_version_supports"`
	TargetVersion          []TargetVersion `json:"target_version" bson:"target_version"`
	RelayInfo              RelayInfo       `json:"relay_info" bson:"relay_info"`
}

// TargetVersion this is the target version for node deployment.
type TargetVersion struct {
	OsType  string `json:"os_type" bson:"os_type"`
	CPUArch string `json:"cpu_arch" bson:"cpu_arch"`
	Version string `json:"version" bson:"version"`
}

// InstallOptions this is the options for nodemgr tools.
type InstallOptions struct {
	ReRegister    bool `json:"re_register" bson:"re_register"`
	DirectInstall bool `json:"direct_install" bson:"direct_install"`
	IsManual      bool `json:"is_manual" bson:"is_manual"`
}

// UpgradeOptions this is the options for node upgrade.
type UpgradeOptions struct {
	DirectLink bool `json:"direct_link" bson:"direct_link"`
}

// RestartOptions this is the options for node restart.
type RestartOptions struct {
	ForceRestart           bool          `json:"force_restart" bson:"force_restart"`
	GracefulRestartTimeout time.Duration `json:"graceful_restart_timeout" bson:"graceful_restart_timeout"`
}

// TransferOptions this is the options for node transfer.
type TransferOptions struct {
	SelectDownloads      bool `json:"select_downloads" bson:"select_downloads"`
	EnableReleasePackage bool `json:"enable_release_package" bson:"enable_release_package"`
	EnableInstaller      bool `json:"enable_installer" bson:"enable_installer"`
}

// VersionSupports describes this version supports things.
type VersionSupports struct {
	// OperateAgentRestart means this version agent supports operates restart through cluster,
	// which will check if the agent is idle.
	OperateAgentRestart bool `json:"operate_agent_restart" bson:"operate_agent_restart"`
}

// NodeConf this is the node conf for node deployment.
type NodeConf struct {
	ConfigTemplate map[string]string `json:"config_template" bson:"config_template"`
	PreSetting     map[string]any    `json:"pre_setting" bson:"pre_setting"`
	CustomSetting  map[string]any    `json:"custom_setting" bson:"custom_setting"`
}

// RelayInfo this is the relay info for node deployment.
type RelayInfo struct {
	HostID          int64  `json:"host_id" bson:"host_id"`
	AgentID         string `json:"agent_id" bson:"agent_id"`
	InnerIP         string `json:"inner_ip" bson:"inner_ip"`
	InnerIPV6       string `json:"inner_ipv6" bson:"inner_ipv6"`
	PackageDestDir  string `json:"package_dest_dir" bson:"package_dest_dir"`
	DownloadSvcPort int64  `json:"download_svc_port" bson:"download_svc_port"`
	CallbackSvcPort int64  `json:"callback_svc_port" bson:"callback_svc_port"`
}

// UniqueFields unique fields of the table.
func (deploy *Data) UniqueFields() []string {
	return []string{FieldKeyToken}
}

// UniqueKey unique key of the table.
func (deploy *Data) UniqueKey() string {
	return deploy.Token
}

// Table represent the complete db structures of node deployment.
type Table base.TableBroker[*Data]
