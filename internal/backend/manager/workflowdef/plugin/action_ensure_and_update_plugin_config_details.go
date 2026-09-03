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

package plugin

import (
	"errors"
	"fmt"
	"net"
	"path/filepath"
	"strings"
	"time"

	pluginUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/plugin/utils"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/release"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/winpath"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameEnsureAndUpdatePluginConfigDetails defines the action name.
	ActionNameEnsureAndUpdatePluginConfigDetails = "ensure_and_update_plugin_config_details"
)

// NewActionEnsureAndUpdatePluginConfigDetails ...
func NewActionEnsureAndUpdatePluginConfigDetails(capability *Capability) action.Definition {
	return &actionEnsureAndUpdatePluginConfigDetails{
		daoHost:             capability.StorageTopo,
		daoNetworkArea:      capability.StorageTopo,
		daoPluginDeployment: capability.StoragePlugin,
		daoPluginRelease:    capability.StorageRelease,
	}
}

// ActParamEnsureAndUpdatePluginConfigDetails ...
type ActParamEnsureAndUpdatePluginConfigDetails struct {
	pluginUtils.PluginActionStandardParam `json:",inline"`
}

// actionEnsureAndUpdatePluginConfigDetails ...
type actionEnsureAndUpdatePluginConfigDetails struct {
	daoHost             topoStg.IStorageHost
	daoNetworkArea      topoStg.IStorageNetworkArea
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	daoPluginRelease    release.IPlugin
}

// Name returns the name of the action.
func (act *actionEnsureAndUpdatePluginConfigDetails) Name() string {
	return ActionNameEnsureAndUpdatePluginConfigDetails
}

// Version returns the version of the action.
func (act *actionEnsureAndUpdatePluginConfigDetails) Version() string {
	return "1.0.0"
}

// Description returns the description of the action.
func (act *actionEnsureAndUpdatePluginConfigDetails) Description() string {
	return "ensure and update plugin config details"
}

// Timeout returns the timeout of the action.
func (act *actionEnsureAndUpdatePluginConfigDetails) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionEnsureAndUpdatePluginConfigDetails) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionEnsureAndUpdatePluginConfigDetails) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionEnsureAndUpdatePluginConfigDetails) DelayFn(_ int) func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *actionEnsureAndUpdatePluginConfigDetails) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActParamEnsureAndUpdatePluginConfigDetails)
	err = conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	std := pluginUtils.NewPluginActionStandarder(act.daoPluginDeployment)
	if err = std.Initialize(ctx, param.PluginActionStandardParam); err != nil {
		return err
	}
	defer func() {
		if storeErr := std.Save(); storeErr != nil {
			err = errors.Join(storeErr, err)
		}
	}()

	hostInfo, err := act.daoHost.GetHostByID(std.Context(), std.DeployInfo().Process.HostID)
	if err != nil {
		return fmt.Errorf("failed to get host by id, host-id(%d): %w", std.DeployInfo().Process.HostID, err)
	}
	if err = pluginUtils.EnsureHostLoginUser(std, hostInfo); err != nil {
		return err
	}
	configSourceHostInfo := getConfigSourceHostInfo(std.DeployInfo(), hostInfo)

	pluginRelease, err := act.daoPluginRelease.GetReleasePlugin(std.Context(), types.ReleasePluginKey{
		Generation: std.DeployInfo().Process.Generation,
		Platform:   std.DeployInfo().Process.Platform,
		Version:    std.DeployInfo().Process.Info.Version,
		Name:       std.DeployInfo().Process.PluginPkgName,
	})
	if err != nil {
		return fmt.Errorf("failed to get plugin release info: %w", err)
	}
	if !pluginRelease.Enabled {
		return fmt.Errorf("plugin release is not enabled, plugin-name(%s), version(%s)",
			std.DeployInfo().Process.PluginName, std.DeployInfo().Process.Info.Version)
	}

	pluginConf, err := act.daoPluginDeployment.GetPluginDeploymentPluginConf(std.Context(), std.Token())
	if err != nil {
		return fmt.Errorf("failed to get plugin deployment plugin conf: %w", err)
	}

	var renderContext map[string]any
	switch pluginRelease.TemplateRendererType {
	case types.TemplateRendererTypeGoTemplate:
		renderContext = act.generateGoTemplateSystemConfigContext(std, configSourceHostInfo)
	case types.TemplateRendererTypeJinja2:
		renderContext, err = act.generateJinja2SystemConfigContext(std, hostInfo, configSourceHostInfo)
		if err != nil {
			return fmt.Errorf("failed to generate jinja2 system config context: %w", err)
		}
	default:
		return fmt.Errorf("unsupported template renderer: %s", pluginRelease.TemplateRendererType)
	}

	if err := act.validateCustomContextBlacklist(pluginConf.CustomConfigContext); err != nil {
		return fmt.Errorf("custom context validation failed: %w", err)
	}

	pluginConf.TemplateRenderer = pluginRelease.TemplateRendererType
	pluginConf.SystemConfigContext = renderContext
	if err := act.fillConfigDetails(std, pluginRelease, pluginConf); err != nil {
		return fmt.Errorf("failed to fill config details: %w", err)
	}

	if err := act.daoPluginDeployment.UpdatePluginDeploymentPluginConf(std.Context(), std.Token(), pluginConf); err != nil {
		return fmt.Errorf("failed to update plugin deployment plugin conf: %w", err)
	}

	return nil
}

func (act *actionEnsureAndUpdatePluginConfigDetails) fillConfigDetails(std *pluginUtils.PluginActionStandarder, pluginRelease *types.ReleasePlugin,
	pluginConf *types.PluginDeploymentPluginConf) error {

	templateMap := make(map[string]types.PluginPkgConfigTemplate, len(pluginRelease.ConfigTemplates))
	var mainTemplate *types.PluginConfigDetail

	for _, tpl := range pluginRelease.ConfigTemplates {
		templateMap[tpl.Name] = tpl

		if tpl.IsMainConfig && mainTemplate == nil {
			mainTemplate = &types.PluginConfigDetail{
				Name:         tpl.Name,
				TemplateName: tpl.Name,
				Content:      tpl.SourceContent,
				IsMainConfig: tpl.IsMainConfig,
				FilePath:     splitPathAndCombineByOS(tpl.FilePath, pluginRelease.Platform.OS),
			}
		}
	}

	for idx, detail := range pluginConf.ConfigFilesDetail {
		tpl, ok := templateMap[detail.TemplateName]
		if !ok {
			std.InstanceData().Log().Zh("未匹配到模板, template-name(%s), plugin-name(%s)", detail.TemplateName, pluginRelease.Name).
				En("no matched template, template-name(%s), plugin-name(%s)", detail.TemplateName, pluginRelease.Name).
				Error()

			return fmt.Errorf("no matched template for config detail, template-name(%s)", detail.TemplateName)
		}

		name := detail.Name
		if name == "" {
			name = tpl.Name
		}

		pluginConf.ConfigFilesDetail[idx].Name = name
		pluginConf.ConfigFilesDetail[idx].TemplateName = tpl.Name
		pluginConf.ConfigFilesDetail[idx].Content = tpl.SourceContent
		pluginConf.ConfigFilesDetail[idx].IsMainConfig = tpl.IsMainConfig
		pluginConf.ConfigFilesDetail[idx].FilePath = splitPathAndCombineByOS(tpl.FilePath, pluginRelease.Platform.OS)
	}

	if len(pluginConf.ConfigFilesDetail) == 0 && mainTemplate != nil {
		pluginConf.ConfigFilesDetail = []*types.PluginConfigDetail{mainTemplate}
	}

	return nil
}

func splitPathAndCombineByOS(fullPath string, osType criteria.OSType) string {
	if len(fullPath) == 0 {
		return fullPath
	}

	var paths []string
	switch {
	case strings.Contains(fullPath, string(filepath.Separator)):
		paths = strings.Split(fullPath, string(filepath.Separator))
	case strings.Contains(fullPath, string(winpath.DirSeparator)):
		paths = strings.Split(fullPath, string(winpath.DirSeparator))
	default:
		return fullPath
	}

	sep := winpath.DirSeparator
	if osType != criteria.OSWindows {
		// notice: nodemgr is running on linux operating system, so we use filepath for non-windows OS type
		sep = filepath.Separator
	}

	return strings.Join(paths, string(sep))
}

func getConfigSourceHostInfo(info *types.PluginDeploymentInfo, runtimeHost *types.Host) *types.Host {
	if info.ConfigSource.Host.HostID <= 0 {
		return runtimeHost
	}

	return &info.ConfigSource.Host
}

// ContextPluginInfo plugin info for render context.
// Use PascalCase to cure the struct field names to be compatible with Go template rendering.
type ContextPluginInfo struct {
	Name          string `json:"Name"`
	LogPath       string `json:"LogPath"`
	DataPath      string `json:"DataPath"`
	PidPath       string `json:"PidPath"`
	SetupPath     string `json:"SetupPath"`
	ConfigPath    string `json:"ConfigPath"`
	HostIDPath    string `json:"HostIDPath"`
	PluginIPC     string `json:"PluginIPC"`
	DataIPC       string `json:"DataIPC"`
	AgentDir      string `json:"AgentDir"`
	GroupID       string `json:"GroupID"`
	IsMultiTenant bool   `json:"IsMultiTenant"`
}

// ContextPreDefinitionConstants pre-definition constants for render context.
// Use PascalCase to cure the struct field names to be compatible with Go template rendering.
type ContextPreDefinitionConstants struct {
	Global map[string]any `json:"Global"`
	Unique map[string]any `json:"Unique"`
}

// ContextNodeInfoStatic node info static for render context.
// Use PascalCase to cure the struct field names to be compatible with Go template rendering.
type ContextNodeInfoStatic struct {
	BizID         int64    `json:"BizID"`
	NetworkAreaID int64    `json:"NetworkAreaID"`
	ZoneID        int64    `json:"ZoneID"`
	CityID        string   `json:"CityID"`
	HostName      string   `json:"HostName"`
	DeptName      string   `json:"DeptName"`
	InnerIPList   []string `json:"InnerIPList"`
	InnerIPV6List []string `json:"InnerIPV6List"`
	OuterIPList   []string `json:"OuterIPList"`
	OuterIPV6List []string `json:"OuterIPV6List"`
	Operator      string   `json:"Operator"`
	Mac           string   `json:"Mac"`
	OSTypeCCID    string   `json:"OSTypeCCID"`
	OSType        string   `json:"OSType"`
	Arch          string   `json:"Arch"`
	Addressing    string   `json:"Addressing"`
	CPUNum        float64  `json:"CPUNum"`
	MemCap        float64  `json:"MemCap"`
}

// ContextNodeInfoDynamic node info dynamic for render context.
// Use PascalCase to cure the struct field names to be compatible with Go template rendering.
type ContextNodeInfoDynamic struct {
	NodeRole                 string   `json:"NodeRole"`
	NodeStatus               string   `json:"NodeStatus"`
	NodeVersion              string   `json:"NodeVersion"`
	NodeGeneration           int64    `json:"NodeGeneration"`
	NodeCPUArch              string   `json:"NodeCPUArch"`
	NodeOsType               string   `json:"NodeOsType"`
	AgentID                  string   `json:"AgentID"`
	NetworkUnitID            int64    `json:"NetworkUnitID"`
	LoginUser                string   `json:"LoginUser"`
	ExportIP                 string   `json:"ExportIP"`
	ExportIPV6               string   `json:"ExportIPV6"`
	AdvertiseIP              string   `json:"AdvertiseIP"`
	AdvertiseIPV6            string   `json:"AdvertiseIPV6"`
	ProxyAccessDisabled      bool     `json:"ProxyAccessDisabled"`
	ProxyTags                []string `json:"ProxyTags"`
	ProxyClusterPort         int64    `json:"ProxyClusterPort"`
	ProxyDataPort            int64    `json:"ProxyDataPort"`
	ProxyFilePort            int64    `json:"ProxyFilePort"`
	RelayDownloadPort        int64    `json:"RelayDownloadPort"`
	RelayCallbackPort        int64    `json:"RelayCallbackPort"`
	ProxyInstallOriginUnitID int64    `json:"ProxyInstallOriginUnitID"`
}

// ContextNodeInfo node info for render context.
// Use PascalCase to cure the struct field names to be compatible with Go template rendering.
type ContextNodeInfo struct {
	HostID   int64                  `json:"HostID"`
	TenantID string                 `json:"TenantID"`
	Static   ContextNodeInfoStatic  `json:"Static"`
	Dynamic  ContextNodeInfoDynamic `json:"Dynamic"`
}

func convertHostTypeToContextNodeInfo(hostInfo *types.Host) ContextNodeInfo {
	proxyTags := make([]string, 0, len(hostInfo.Dynamic.ProxyTags))
	for _, tag := range hostInfo.Dynamic.ProxyTags {
		proxyTags = append(proxyTags, string(tag))
	}

	return ContextNodeInfo{
		HostID:   hostInfo.HostID,
		TenantID: hostInfo.TenantID,
		Static: ContextNodeInfoStatic{
			BizID:         hostInfo.Static.BizID,
			NetworkAreaID: hostInfo.Static.NetworkAreaID,
			ZoneID:        hostInfo.Static.ZoneID,
			CityID:        hostInfo.Static.CityID,
			HostName:      hostInfo.Static.HostName,
			DeptName:      hostInfo.Static.DeptName,
			InnerIPList:   hostInfo.Static.InnerIPList,
			InnerIPV6List: hostInfo.Static.InnerIPV6List,
			OuterIPList:   hostInfo.Static.OuterIPList,
			OuterIPV6List: hostInfo.Static.OuterIPV6List,
			Operator:      hostInfo.Static.Operator,
			Mac:           hostInfo.Static.Mac,
			OSTypeCCID:    hostInfo.Static.OSTypeCCID,
			OSType:        hostInfo.Static.OSType,
			Arch:          hostInfo.Static.Arch,
			Addressing:    string(hostInfo.Static.Addressing),
			CPUNum:        hostInfo.Static.CPUNum,
			MemCap:        hostInfo.Static.MemCap,
		},
		Dynamic: ContextNodeInfoDynamic{
			NodeRole:                 string(hostInfo.Dynamic.NodeRole),
			NodeStatus:               string(hostInfo.Dynamic.NodeStatus),
			NodeVersion:              hostInfo.Dynamic.NodeVersion,
			NodeGeneration:           int64(hostInfo.Dynamic.NodeGeneration),
			NodeCPUArch:              hostInfo.Dynamic.NodeCPUArch.String(),
			NodeOsType:               hostInfo.Dynamic.NodeOsType.String(),
			AgentID:                  hostInfo.Static.SyncedAgentID,
			NetworkUnitID:            hostInfo.Dynamic.NetworkUnitID,
			LoginUser:                hostInfo.Dynamic.LoginUser,
			ExportIP:                 hostInfo.Dynamic.ExportIP,
			ExportIPV6:               hostInfo.Dynamic.ExportIPV6,
			AdvertiseIP:              hostInfo.Dynamic.AdvertiseIP,
			AdvertiseIPV6:            hostInfo.Dynamic.AdvertiseIPV6,
			ProxyAccessDisabled:      hostInfo.Dynamic.ProxyAccessDisabled,
			ProxyTags:                proxyTags,
			ProxyClusterPort:         hostInfo.Dynamic.ProxyClusterPort,
			ProxyDataPort:            hostInfo.Dynamic.ProxyDataPort,
			ProxyFilePort:            hostInfo.Dynamic.ProxyFilePort,
			RelayDownloadPort:        hostInfo.Dynamic.RelayDownloadPort,
			RelayCallbackPort:        hostInfo.Dynamic.RelayCallbackPort,
			ProxyInstallOriginUnitID: hostInfo.Dynamic.ProxyInstallOriginUnitID,
		},
	}
}

const (
	keyPluginInfo             = "PluginInfo"
	keyNodeInfo               = "NodeInfo"
	keyPreDefinitionConstants = "PreDefinitionConstants"
	keyCustomContext          = "CustomContext"
)

func (act *actionEnsureAndUpdatePluginConfigDetails) generateGoTemplateSystemConfigContext(
	std *pluginUtils.PluginActionStandarder, hostInfo *types.Host) map[string]any {

	pluginInfo := ContextPluginInfo{
		Name:          std.DeployInfo().Process.PluginName,
		LogPath:       std.DeployInfo().BaseRuntime.LogDir,
		DataPath:      std.DeployInfo().BaseRuntime.DataDir,
		PidPath:       std.DeployInfo().BaseRuntime.RunDir,
		SetupPath:     std.DeployInfo().BaseRuntime.PluginHomeDir,
		ConfigPath:    std.DeployInfo().BaseRuntime.ConfigDir,
		HostIDPath:    std.DeployInfo().BaseRuntime.HostIDPath,
		PluginIPC:     std.DeployInfo().BaseRuntime.PluginIPC,
		DataIPC:       std.DeployInfo().BaseRuntime.DataIPC,
		AgentDir:      std.DeployInfo().BaseRuntime.GSEHomeDir,
		GroupID:       std.DeployInfo().Process.PluginGroup,
		IsMultiTenant: tenant.GetMode() == tenant.ModeMultiple,
	}

	preDefinitionConstants := ContextPreDefinitionConstants{
		Unique: std.DeployInfo().BaseRuntime.PluginCommonConstants,
		Global: std.DeployInfo().BaseRuntime.GlobalCommonConstants,
	}

	nodeInfo := convertHostTypeToContextNodeInfo(hostInfo)

	renderContext := map[string]any{
		keyPluginInfo:             conv.StructToMapIgnoreError(pluginInfo),
		keyNodeInfo:               conv.StructToMapIgnoreError(nodeInfo),
		keyPreDefinitionConstants: conv.StructToMapIgnoreError(preDefinitionConstants),
	}

	return renderContext
}

// Jinja2 render context keys.
const (
	keyPluginPath   = "plugin_path"
	keyNodeMan      = "nodeman"
	keyCmdbInstance = "cmdb_instance"
	keyTarget       = "target"
	keyControlInfo  = "control_info"

	keyService = "service"
	keyScope   = "scope"
	keyProcess = "process"

	keyLogPath       = "log_path"
	keyDataPath      = "data_path"
	keyPidPath       = "pid_path"
	keySetupPath     = "setup_path"
	keyEndpoint      = "endpoint"
	keyHostID        = "host_id"
	keySubConfigPath = "subconfig_path"

	keyHost = "host"

	keyIsMultiTenant = "is_multi_tenant"
	keyConstants     = "constants"
	keyBkHostID      = "bk_host_id"
	keyOsType        = "os_type"
	keyCPUArch       = "cpu_arch"
	keyInnerIP       = "inner_ip"
	keyOuterIP       = "outer_ip"
	keyLoginIP       = "login_ip"
	keyGlobal        = "global"

	keyBkBizID           = "bk_biz_id"
	keyBkHostName        = "bk_host_name"
	keyBkAddressing      = "bk_addressing"
	keyBkCloudID         = "bk_cloud_id"
	keyBkCloudName       = "bk_cloud_name"
	keyBkHostInnerIP     = "bk_host_innerip"
	keyBkHostOuterIP     = "bk_host_outerip"
	keyBkHostInnerIPv6   = "bk_host_innerip_v6"
	keyBkHostOuterIPv6   = "bk_host_outerip_v6"
	keyBkOSType          = "bk_os_type"
	keyBkAgentID         = "bk_agent_id"
	keyBkCPUArchitecture = "bk_cpu_architecture"
	keyBkCPU             = "bk_cpu"
	keyBkMem             = "bk_mem"
	keyUserName          = "username"
	keyAccount           = "account"
	keyAuthType          = "auth_type"
	keyHostNodeType      = "host_node_type"

	keyID                = "id"
	keyName              = "name"
	keyLabels            = "labels"
	keyBkModuleID        = "bk_module_id"
	keyBkObjID           = "bk_obj_id"
	keyBkInstID          = "bk_inst_id"
	keyServiceTemplateID = "service_template_id"
	keyServiceCategoryID = "service_category_id"

	keyAutoStart         = "auto_start"
	keyBkFuncName        = "bk_func_name"
	keyBkProcessID       = "bk_process_id"
	keyBkProcessName     = "bk_process_name"
	keyBkStartParamRegex = "bk_start_param_regex"
	keyBkSupplierAccount = "bk_supplier_account"
	keyCreateTime        = "create_time"
	keyLastTime          = "last_time"
	keyDescription       = "description"
	keyFaceStopCmd       = "face_stop_cmd"
	keyPidFile           = "pid_file"
	keyPriority          = "priority"
	keyProcNum           = "proc_num"
	keyReloadCmd         = "reload_cmd"
	keyRestartCmd        = "restart_cmd"
	keyStartCmd          = "start_cmd"
	keyStopCmd           = "stop_cmd"
	keyTimeout           = "timeout"
	keyUser              = "user"
	keyWorkPath          = "work_path"
	keyBkCreatedAt       = "bk_created_at"
	keyBkCreatedBy       = "bk_created_by"
	keyBkUpdatedAt       = "bk_updated_at"
	keyBkUpdatedBy       = "bk_updated_by"
	keyBindInfo          = "bind_info"
	keyEnable            = "enable"
	keyIP                = "ip"
	keyPort              = "port"
	keyProtocol          = "protocol"
	keyTemplateRowID     = "template_row_id"

	keyPluginIPC    = "pluginipc"
	keyDataIPC      = "dataipc"
	keyGSEAgentHome = "gse_agent_home"
	keyListenIP     = "listen_ip"
	keyListenPort   = "listen_port"
	keyGroupID      = "group_id"
)

func (act *actionEnsureAndUpdatePluginConfigDetails) generateJinja2SystemConfigContext(
	std *pluginUtils.PluginActionStandarder, runtimeHostInfo *types.Host, configSourceHostInfo *types.Host) (map[string]any, error) {

	runtimeHostRenderData, err := act.generateHostRenderData(std, runtimeHostInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to generate runtime host render data: %w", err)
	}
	configSourceHostRenderData, err := act.generateHostRenderData(std, configSourceHostInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to generate config source host render data: %w", err)
	}

	constants := std.DeployInfo().BaseRuntime.PluginCommonConstants
	constants[keyGlobal] = std.DeployInfo().BaseRuntime.GlobalCommonConstants
	nodeManInfo := map[string]any{
		keyHost: map[string]any{
			keyBkHostID: runtimeHostInfo.HostID,
			keyOsType:   runtimeHostInfo.Static.OSType,
			keyCPUArch:  runtimeHostInfo.Static.Arch,
			keyInnerIP:  runtimeHostRenderData.innerIP,
			keyOuterIP:  runtimeHostRenderData.outerIP,
			keyLoginIP:  runtimeHostInfo.Dynamic.LoginIP,
		},
		keyIsMultiTenant: tenant.GetMode() == tenant.ModeMultiple,
		keyConstants:     constants,
	}

	cmdbInstance := map[string]any{
		keyHost: generateConfigSourceHostRenderContext(configSourceHostInfo, configSourceHostRenderData),
		keyScope: generateMatchedTopoRelationRenderContext(
			std.DeployInfo().ConfigSource.MatchedTopoRelations,
		),
		keyProcess: generateConfigSourceServiceProcessRenderContext(
			std.DeployInfo().ConfigSource.ServiceInstance.Processes,
		),
	}
	if std.DeployInfo().ConfigSource.ServiceInstance.ID > 0 {
		cmdbInstance[keyService] = generateConfigSourceServiceRenderContext(
			std.DeployInfo().ConfigSource.ServiceInstance,
		)
	}

	// notice: in v2 plugin, subconfig_path still need provide, and the path is fixed
	if err = pluginUtils.CheckDirPathSafe(std.DeployInfo().BaseRuntime.SubConfigDir, std.DeployInfo().Process.Platform.OS); err != nil {
		return nil, fmt.Errorf("check subconfig dir safe failed, dir(%s): %w", std.DeployInfo().BaseRuntime.SubConfigDir, err)
	}

	dataEndpoint := renderIPCEndpoint(
		std.DeployInfo().Process.Platform.OS,
		std.DeployInfo().BaseRuntime.DataIPC,
		runtimeHostInfo.Static.InnerIPList,
		runtimeHostInfo.Static.InnerIPV6List,
	)
	pluginEndpoint := renderIPCEndpoint(
		std.DeployInfo().Process.Platform.OS,
		std.DeployInfo().BaseRuntime.PluginIPC,
		runtimeHostInfo.Static.InnerIPList,
		runtimeHostInfo.Static.InnerIPV6List,
	)

	pluginPath := map[string]any{
		keyLogPath:       std.DeployInfo().BaseRuntime.LogDir,
		keyDataPath:      std.DeployInfo().BaseRuntime.DataDir,
		keyPidPath:       std.DeployInfo().BaseRuntime.RunDir,
		keySetupPath:     std.DeployInfo().BaseRuntime.PluginHomeDir,
		keyEndpoint:      dataEndpoint,
		keyHostID:        std.DeployInfo().BaseRuntime.HostIDPath,
		keySubConfigPath: std.DeployInfo().BaseRuntime.SubConfigDir,
	}

	controlInfo := map[string]any{
		keyPluginIPC:    pluginEndpoint,
		keyDataIPC:      dataEndpoint,
		keyGSEAgentHome: std.DeployInfo().BaseRuntime.GSEHomeDir,
		keyGroupID:      std.DeployInfo().Process.PluginGroup,
		keyLogPath:      std.DeployInfo().BaseRuntime.LogDir,
		keyDataPath:     std.DeployInfo().BaseRuntime.DataDir,
		keyPidPath:      std.DeployInfo().BaseRuntime.RunDir,
		keySetupPath:    std.DeployInfo().BaseRuntime.PluginHomeDir,
		// TODO: implement a plugin to obtain listen ip and port
		keyListenIP:   "",
		keyListenPort: 0,
	}

	return map[string]any{
		keyPluginPath:   pluginPath,
		keyNodeMan:      nodeManInfo,
		keyCmdbInstance: cmdbInstance,
		keyTarget:       cmdbInstance,
		keyControlInfo:  controlInfo,
	}, nil
}

type hostRenderData struct {
	innerIP     string
	outerIP     string
	innerIPv6   string
	outerIPv6   string
	networkArea string
}

func (act *actionEnsureAndUpdatePluginConfigDetails) generateHostRenderData(
	std *pluginUtils.PluginActionStandarder, hostInfo *types.Host) (hostRenderData, error) {

	var (
		innerIP   string
		outerIP   string
		innerIPv6 string
		outerIPv6 string
	)

	if hostInfo == nil {
		return hostRenderData{}, fmt.Errorf("host info is nil")
	}
	if hostInfo.Static == nil {
		return hostRenderData{}, fmt.Errorf("host static info is nil, host-id(%d)", hostInfo.HostID)
	}
	if hostInfo.Dynamic == nil {
		return hostRenderData{}, fmt.Errorf("host dynamic info is nil, host-id(%d)", hostInfo.HostID)
	}

	if len(hostInfo.Static.InnerIPList) > 0 {
		innerIP = hostInfo.Static.InnerIPList[0]
	}
	if len(hostInfo.Static.OuterIPList) > 0 {
		outerIP = hostInfo.Static.OuterIPList[0]
	}
	if len(hostInfo.Static.InnerIPV6List) > 0 {
		innerIPv6 = hostInfo.Static.InnerIPV6List[0]
	}
	if len(hostInfo.Static.OuterIPV6List) > 0 {
		outerIPv6 = hostInfo.Static.OuterIPV6List[0]
	}

	networkArea, err := act.daoNetworkArea.GetNetworkArea(std.Context(), hostInfo.Static.NetworkAreaID)
	if err != nil {
		return hostRenderData{}, fmt.Errorf("failed to get network area, networkarea-id(%d): %w", hostInfo.Static.NetworkAreaID, err)
	}

	return hostRenderData{
		innerIP:     innerIP,
		outerIP:     outerIP,
		innerIPv6:   innerIPv6,
		outerIPv6:   outerIPv6,
		networkArea: networkArea.Name,
	}, nil
}

func generateConfigSourceHostRenderContext(hostInfo *types.Host, data hostRenderData) map[string]any {
	return map[string]any{
		keyBkBizID:           hostInfo.Static.BizID,
		keyBkHostID:          hostInfo.HostID,
		keyOsType:            hostInfo.Static.OSType,
		keyBkOSType:          hostInfo.Static.OSTypeCCID,
		keyBkAgentID:         hostInfo.Static.SyncedAgentID,
		keyBkCloudID:         hostInfo.Static.NetworkAreaID,
		keyBkCloudName:       data.networkArea,
		keyBkHostName:        hostInfo.Static.HostName,
		keyBkAddressing:      hostInfo.Static.Addressing,
		keyBkHostInnerIP:     data.innerIP,
		keyBkHostOuterIP:     data.outerIP,
		keyBkHostInnerIPv6:   data.innerIPv6,
		keyBkHostOuterIPv6:   data.outerIPv6,
		keyBkCPUArchitecture: hostInfo.Static.Arch,
		keyBkCPU:             hostInfo.Static.CPUNum,
		keyBkMem:             hostInfo.Static.MemCap,
		keyUserName:          hostInfo.Static.Operator,
		keyAccount:           hostInfo.Dynamic.LoginUser,
		keyAuthType:          strings.ToUpper(string(hostInfo.Dynamic.LoginMode)),
		keyHostNodeType:      strings.ToUpper(string(hostInfo.Dynamic.NodeRole)),
	}
}

func generateConfigSourceServiceRenderContext(serviceInstance types.ServiceInstance) map[string]any {
	return map[string]any{
		keyID:                serviceInstance.ID,
		keyName:              serviceInstance.Name,
		keyLabels:            serviceInstance.Labels,
		keyProcess:           generateConfigSourceServiceProcessRenderContext(serviceInstance.Processes),
		keyBkBizID:           serviceInstance.BizID,
		keyBkHostID:          serviceInstance.HostID,
		keyBkModuleID:        serviceInstance.ModuleID,
		keyServiceTemplateID: serviceInstance.ServiceTemplateID,
		keyServiceCategoryID: serviceInstance.ServiceCategoryID,
	}
}

func generateConfigSourceServiceProcessRenderContext(
	processes map[string]types.ServiceInstanceProcess,
) map[string]any {
	renderProcesses := make(map[string]any, len(processes))
	for name, process := range processes {
		renderProcesses[name] = map[string]any{
			keyAutoStart:         process.AutoStart,
			keyBkBizID:           process.BizID,
			keyBkFuncName:        process.FuncName,
			keyBkProcessID:       process.ProcessID,
			keyBkProcessName:     process.ProcessName,
			keyBkStartParamRegex: process.StartParamRegex,
			keyBkSupplierAccount: process.SupplierAccount,
			keyCreateTime:        process.CreateTime,
			keyLastTime:          process.LastTime,
			keyDescription:       process.Description,
			keyFaceStopCmd:       process.FaceStopCmd,
			keyPidFile:           process.PidFile,
			keyPriority:          process.Priority,
			keyProcNum:           process.ProcNum,
			keyReloadCmd:         process.ReloadCmd,
			keyRestartCmd:        process.RestartCmd,
			keyStartCmd:          process.StartCmd,
			keyStopCmd:           process.StopCmd,
			keyTimeout:           process.Timeout,
			keyUser:              process.User,
			keyWorkPath:          process.WorkPath,
			keyBkCreatedAt:       process.CreateAt,
			keyBkCreatedBy:       process.CreateBy,
			keyBkUpdatedAt:       process.UpdateAt,
			keyBkUpdatedBy:       process.UpdateBy,
			keyBindInfo:          generateServiceInstanceProcessBindInfoRenderContext(process.BindInfo),
		}
	}

	return renderProcesses
}

func generateServiceInstanceProcessBindInfoRenderContext(
	bindInfo []types.ServiceInstanceProcessBindInfo,
) []map[string]any {
	return conv.SliceToSlice(bindInfo, func(info types.ServiceInstanceProcessBindInfo) map[string]any {
		return map[string]any{
			keyEnable:        info.Enable,
			keyIP:            info.IP,
			keyPort:          info.Port,
			keyProtocol:      info.Protocol,
			keyTemplateRowID: info.TemplateRowID,
		}
	})
}

func generateMatchedTopoRelationRenderContext(relations []types.TargetMatchedTopoRelation) []map[string]any {
	return conv.SliceToSlice(relations, func(relation types.TargetMatchedTopoRelation) map[string]any {
		return map[string]any{
			keyBkObjID:  relation.TopoObjID,
			keyBkInstID: relation.TopoInstID,
		}
	})
}

func renderIPCEndpoint(osType criteria.OSType, ipc string, innerIPs, innerIPv6s []string) string {
	if osType != criteria.OSWindows {
		return ipc
	}

	host := "127.0.0.1"
	if len(innerIPs) == 0 && len(innerIPv6s) > 0 {
		host = "::1"
	}

	return net.JoinHostPort(host, ipc)
}

func getBlacklistKeys() map[string]struct{} {
	return map[string]struct{}{
		keyPluginInfo:             {},
		keyNodeInfo:               {},
		keyPreDefinitionConstants: {},
		keyCustomContext:          {},

		keyPluginPath:   {},
		keyNodeMan:      {},
		keyCmdbInstance: {},
		keyTarget:       {},
		keyControlInfo:  {},
	}
}

func (act *actionEnsureAndUpdatePluginConfigDetails) validateCustomContextBlacklist(customContext map[string]any) error {
	blacklistPrefixKeys := getBlacklistKeys()
	for customKey := range customContext {
		if _, exists := blacklistPrefixKeys[customKey]; exists {
			return fmt.Errorf("the key is reserved, cannot be used, custom-context-key(%s)", customKey)
		}
	}

	return nil
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionEnsureAndUpdatePluginConfigDetails) DisplayNameZh() string {
	return "确保并更新插件配置详情"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionEnsureAndUpdatePluginConfigDetails) DisplayNameEn() string {
	return "Ensure and Update Plugin Config Details"
}
