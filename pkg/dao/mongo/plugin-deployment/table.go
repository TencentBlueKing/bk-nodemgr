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

package plugindeployment

import (
	"time"

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
	ConfigSource     *target          `json:"config_source,omitempty" bson:"config_source,omitempty"`
	InstallerRuntime installerRuntime `json:"installer_runtime" bson:"installer_runtime"`
	BaseRuntime      baseRuntime      `json:"base_runtime" bson:"base_runtime"`
	TransferOptions  transferOptions  `json:"transfer_options" bson:"transfer_options"`
	InstallOptions   installOptions   `json:"install_options" bson:"install_options"`
}

type target struct {
	Host                 targetHost                  `json:"host" bson:"host"`
	ServiceInstance      serviceInstance             `json:"service_instance" bson:"service_instance"`
	MatchedTopoRelations []targetMatchedTopoRelation `json:"matched_topo_relations" bson:"matched_topo_relations"`
}

type targetMatchedTopoRelation struct {
	TopoObjID  string `json:"topo_obj_id" bson:"topo_obj_id"`
	TopoInstID int64  `json:"topo_inst_id" bson:"topo_inst_id"`
}

type serviceInstance struct {
	ID                int64                             `json:"id" bson:"id"`
	Name              string                            `json:"name" bson:"name"`
	Labels            map[string]string                 `json:"labels" bson:"labels"`
	Processes         map[string]serviceInstanceProcess `json:"processes" bson:"processes"`
	BizID             int64                             `json:"biz_id" bson:"biz_id"`
	HostID            int64                             `json:"host_id" bson:"host_id"`
	ModuleID          int64                             `json:"module_id" bson:"module_id"`
	ServiceTemplateID int64                             `json:"service_template_id" bson:"service_template_id"`
	ServiceCategoryID int64                             `json:"service_category_id" bson:"service_category_id"`
}

type serviceInstanceProcess struct {
	AutoStart       bool                             `json:"auto_start" bson:"auto_start"`
	BizID           int64                            `json:"bk_biz_id" bson:"bk_biz_id"`
	FuncName        string                           `json:"bk_func_name" bson:"bk_func_name"`
	ProcessID       int64                            `json:"bk_process_id" bson:"bk_process_id"`
	ProcessName     string                           `json:"bk_process_name" bson:"bk_process_name"`
	StartParamRegex string                           `json:"bk_start_param_regex" bson:"bk_start_param_regex"`
	SupplierAccount string                           `json:"bk_supplier_account" bson:"bk_supplier_account"`
	CreateTime      time.Time                        `json:"create_time" bson:"create_time"`
	LastTime        time.Time                        `json:"last_time" bson:"last_time"`
	Description     string                           `json:"description" bson:"description"`
	FaceStopCmd     string                           `json:"face_stop_cmd" bson:"face_stop_cmd"`
	PidFile         string                           `json:"pid_file" bson:"pid_file"`
	Priority        int64                            `json:"priority" bson:"priority"`
	ProcNum         int64                            `json:"proc_num" bson:"proc_num"`
	ReloadCmd       string                           `json:"reload_cmd" bson:"reload_cmd"`
	RestartCmd      string                           `json:"restart_cmd" bson:"restart_cmd"`
	StartCmd        string                           `json:"start_cmd" bson:"start_cmd"`
	StopCmd         string                           `json:"stop_cmd" bson:"stop_cmd"`
	Timeout         int64                            `json:"timeout" bson:"timeout"`
	User            string                           `json:"user" bson:"user"`
	WorkPath        string                           `json:"work_path" bson:"work_path"`
	CreateAt        string                           `json:"bk_created_at" bson:"bk_created_at"`
	CreateBy        string                           `json:"bk_created_by" bson:"bk_created_by"`
	UpdateAt        string                           `json:"bk_updated_at" bson:"bk_updated_at"`
	UpdateBy        string                           `json:"bk_updated_by" bson:"bk_updated_by"`
	BindInfo        []serviceInstanceProcessBindInfo `json:"bind_info" bson:"bind_info"`
}

type serviceInstanceProcessBindInfo struct {
	Enable        bool   `json:"enable" bson:"enable"`
	IP            string `json:"ip" bson:"ip"`
	Port          string `json:"port" bson:"port"`
	Protocol      string `json:"protocol" bson:"protocol"`
	TemplateRowID int64  `json:"template_row_id" bson:"template_row_id"`
}

type targetHost struct {
	HostID   int64              `json:"host_id" bson:"host_id"`
	TenantID string             `json:"tenant_id" bson:"tenant_id"`
	Static   *targetHostStatic  `json:"static" bson:"static"`
	Dynamic  *targetHostDynamic `json:"dynamic" bson:"dynamic"`
}

type targetHostTopo struct {
	SetID    int64 `json:"set_id" bson:"set_id"`
	ModuleID int64 `json:"module_id" bson:"module_id"`
}

type targetHostStatic struct {
	BizID         int64            `json:"biz_id" bson:"biz_id"`
	Topo          []targetHostTopo `json:"topo" bson:"topo"`
	NetworkAreaID int64            `json:"networkarea_id" bson:"networkarea_id"`
	ZoneID        int64            `json:"zone_id" bson:"zone_id"`
	CityID        string           `json:"city_id" bson:"city_id"`

	HostName      string   `json:"host_name" bson:"host_name"`
	DeptName      string   `json:"dept_name" bson:"dept_name"`
	InnerIPList   []string `json:"inner_ip_list" bson:"inner_ip_list"`
	InnerIPV6List []string `json:"inner_ipv6_list" bson:"inner_ipv6_list"`
	OuterIPList   []string `json:"outer_ip_list" bson:"outer_ip_list"`
	OuterIPV6List []string `json:"outer_ipv6_list" bson:"outer_ipv6_list"`
	Operator      string   `json:"operator" bson:"operator"`
	Mac           string   `json:"mac" bson:"mac"`
	OSTypeCCID    string   `json:"os_type_ccid" bson:"os_type_ccid"`
	OSType        string   `json:"os_type" bson:"os_type"`
	Arch          string   `json:"arch" bson:"arch"`
	Addressing    string   `json:"addressing" bson:"addressing"`
	CPUNum        float64  `json:"cpu_num" bson:"cpu_num"`
	MemCap        float64  `json:"mem_cap" bson:"mem_cap"`

	SyncedAgentID string `json:"synced_agent_id" bson:"synced_agent_id"`
}

type targetHostDynamic struct {
	NodeRole                 string   `json:"node_role" bson:"node_role"`
	NodeStatus               string   `json:"node_status" bson:"node_status"`
	NodeVersion              string   `json:"node_version" bson:"node_version"`
	NodeGeneration           int64    `json:"node_generation" bson:"node_generation"`
	NodeCPUArch              string   `json:"node_cpu_arch" bson:"node_cpu_arch"`
	NodeOsType               string   `json:"node_os_type" bson:"node_os_type"`
	AgentID                  string   `json:"agent_id" bson:"agent_id"`
	NetworkUnitID            int64    `json:"networkunit_id" bson:"networkunit_id"`
	ProxyAccessDisabled      bool     `json:"proxy_access_disabled" bson:"proxy_access_disabled"`
	ProxyTags                []string `json:"proxy_tags" bson:"proxy_tags"`
	ProxyInstallOriginUnitID int64    `json:"proxy_install_origin_unit_id" bson:"proxy_install_origin_unit_id"`
	ProxyClusterPort         int64    `json:"proxy_cluster_port" bson:"proxy_cluster_port"`
	ProxyDataPort            int64    `json:"proxy_data_port" bson:"proxy_data_port"`
	ProxyFilePort            int64    `json:"proxy_file_port" bson:"proxy_file_port"`
	LoginIP                  string   `json:"login_ip" bson:"login_ip"`
	LoginUser                string   `json:"login_user" bson:"login_user"`
	LoginMode                string   `json:"login_mode" bson:"login_mode"`
	ExportIP                 string   `json:"export_ip" bson:"export_ip"`
	ExportIPV6               string   `json:"export_ipv6" bson:"export_ipv6"`
	AdvertiseIP              string   `json:"advertise_ip" bson:"advertise_ip"`
	AdvertiseIPV6            string   `json:"advertise_ipv6" bson:"advertise_ipv6"`
	RelayDownloadPort        int64    `json:"relay_download_port" bson:"relay_download_port"`
	RelayCallbackPort        int64    `json:"relay_callback_port" bson:"relay_callback_port"`
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

	BindIP   string `json:"bind_ip" bson:"bind_ip"`
	BindPort int    `json:"bind_port" bson:"bind_port"`

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
	Pid        int       `json:"pid" bson:"pid"`
	Version    string    `json:"version" bson:"version"`
	AgentID    string    `json:"agent_id" bson:"agent_id"`
	AutoStart  bool      `json:"auto_start" bson:"auto_start"`
	Status     string    `json:"status" bson:"status"`
	LastSyncAt time.Time `json:"last_sync_at" bson:"last_sync_at"`
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

type processSpec struct {
	Resource      processResource      `json:"resource" bson:"resource"`
	MonitorPolicy processMonitorPolicy `json:"monitor_policy" bson:"monitor_policy"`
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
	Version                 string       `json:"version" bson:"version"`
	IsOffline               bool         `json:"is_offline" bson:"is_offline"`
	EnableCompatibilityMode bool         `json:"enable_compatibility_mode" bson:"enable_compatibility_mode"`
	CustomSpec              *processSpec `json:"custom_spec" bson:"custom_spec"`
}

// transferOptions this is the options for plugin transfer.
type transferOptions struct {
	DisableReleasePackage bool `json:"disable_release_package" bson:"disable_release_package"`
	DisableInstaller      bool `json:"disable_installer" bson:"disable_installer"`
}

// PluginConf defines the plugin config.
type PluginConf struct {
	Set                  string         `json:"set" bson:"set"`
	TemplateRenderer     string         `json:"template_renderer" bson:"template_renderer"`
	ConfigFilesDetail    []configDetail `json:"config_files_detail" bson:"config_files_detail"`
	SystemConfigContext  map[string]any `json:"system_config_context" bson:"system_config_context"`
	CustomConfigContext  map[string]any `json:"custom_config_context" bson:"custom_config_context"`
	RemoveConfigFileName []string       `json:"remove_config_file_name" bson:"remove_config_file_name"`
	RemoveAllConfigs     bool           `json:"remove_all_configs" bson:"remove_all_configs"`
}

// configDetail defines the plugin config detail.
type configDetail struct {
	Name         string `json:"name" bson:"name"`
	TemplateName string `json:"template_name" bson:"template_name"`
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
