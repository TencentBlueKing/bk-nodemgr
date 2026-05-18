/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package pluginv2

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	pluginV2Utils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/pluginv2/utils"
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
	// ActionNameEnsureAndUpdatePluginConfigDetailsV2 defines the action name.
	ActionNameEnsureAndUpdatePluginConfigDetailsV2 = "ensure_and_update_plugin_config_details_v2"
)

// NewActionEnsureAndUpdatePluginConfigDetailsV2 ...
func NewActionEnsureAndUpdatePluginConfigDetailsV2(capability *Capability) action.Definition {
	return &actionEnsureAndUpdatePluginConfigDetailsV2{
		daoHost:             capability.StorageTopo,
		daoNetworkArea:      capability.StorageTopo,
		daoPluginDeployment: capability.StoragePlugin,
		daoPluginRelease:    capability.StorageRelease,
	}
}

// ActParamEnsureAndUpdatePluginConfigDetailsV2 ...
type ActParamEnsureAndUpdatePluginConfigDetailsV2 struct {
	pluginV2Utils.PluginActionStandardParam `json:",inline"`
}

// actionEnsureAndUpdatePluginConfigDetailsV2 ...
type actionEnsureAndUpdatePluginConfigDetailsV2 struct {
	daoHost             topoStg.IStorageHost
	daoNetworkArea      topoStg.IStorageNetworkArea
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	daoPluginRelease    release.IPlugin
}

// Name returns the name of the action.
func (act *actionEnsureAndUpdatePluginConfigDetailsV2) Name() string {
	return ActionNameEnsureAndUpdatePluginConfigDetailsV2
}

// Version returns the version of the action.
func (act *actionEnsureAndUpdatePluginConfigDetailsV2) Version() string {
	return "1.0.0"
}

// Description returns the description of the action.
func (act *actionEnsureAndUpdatePluginConfigDetailsV2) Description() string {
	return "ensure and update plugin config details v2"
}

// Timeout returns the timeout of the action.
func (act *actionEnsureAndUpdatePluginConfigDetailsV2) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionEnsureAndUpdatePluginConfigDetailsV2) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionEnsureAndUpdatePluginConfigDetailsV2) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionEnsureAndUpdatePluginConfigDetailsV2) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *actionEnsureAndUpdatePluginConfigDetailsV2) Do(ctx *action.InstanceContext) error {
	param := new(ActParamEnsureAndUpdatePluginConfigDetailsV2)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	std := pluginV2Utils.NewPluginActionStandarder(act.daoPluginDeployment)
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
	if err = pluginV2Utils.EnsureHostLoginUser(std, hostInfo); err != nil {
		return err
	}

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
		renderContext = act.generateGoTemplateSystemConfigContext(std, hostInfo)
	case types.TemplateRendererTypeJinja2:
		renderContext, err = act.generateJinja2SystemConfigContext(std, hostInfo)
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

func (act *actionEnsureAndUpdatePluginConfigDetailsV2) fillConfigDetails(std *pluginV2Utils.PluginActionStandarder,
	pluginRelease *types.ReleasePlugin, pluginConf *types.PluginDeploymentPluginConf) error {

	templateMap := make(map[string]types.PluginPkgConfigTemplate, len(pluginRelease.ConfigTemplates))
	var mainTemplate *types.PluginConfigDetail

	for _, tpl := range pluginRelease.ConfigTemplates {
		templateMap[tpl.Name] = tpl

		if tpl.IsMainConfig && mainTemplate == nil {
			mainTemplate = &types.PluginConfigDetail{
				Name:         tpl.Name,
				Content:      tpl.SourceContent,
				IsMainConfig: tpl.IsMainConfig,
				FilePath:     splitPathAndCombineByOS(tpl.FilePath, pluginRelease.Platform.OS),
			}
		}
	}

	for idx, detail := range pluginConf.ConfigFilesDetail {
		tpl, ok := templateMap[detail.Name]
		if !ok {
			std.InstanceData().Log().Zh("未匹配到模板, config-file-name(%s), plugin-name(%s)", detail.Name, pluginRelease.Name).
				En("no matched template, config-file-name(%s), plugin-name(%s)", detail.Name, pluginRelease.Name).
				Error()

			return fmt.Errorf("no matched template for config detail, name(%s)", detail.Name)
		}

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
	RegionID      string   `json:"RegionID"`
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
			RegionID:      hostInfo.Static.RegionID,
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

func (act *actionEnsureAndUpdatePluginConfigDetailsV2) generateGoTemplateSystemConfigContext(
	std *pluginV2Utils.PluginActionStandarder, hostInfo *types.Host) map[string]any {

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

	keyPluginIPC    = "pluginipc"
	keyDataIPC      = "dataipc"
	keyGSEAgentHome = "gse_agent_home"
	keyListenIP     = "listen_ip"
	keyListenPort   = "listen_port"
	keyGroupID      = "group_id"
)

func (act *actionEnsureAndUpdatePluginConfigDetailsV2) generateJinja2SystemConfigContext(
	std *pluginV2Utils.PluginActionStandarder, hostInfo *types.Host) (map[string]any, error) {

	var (
		innerIP   string
		outerIP   string
		innerIPv6 string
		outerIPv6 string
	)

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
		return nil, fmt.Errorf("failed to get network area, networkarea-id(%d): %w", hostInfo.Static.NetworkAreaID, err)
	}

	constants := std.DeployInfo().BaseRuntime.PluginCommonConstants
	constants[keyGlobal] = std.DeployInfo().BaseRuntime.GlobalCommonConstants
	nodeManInfo := map[string]any{
		keyHost: map[string]any{
			keyBkHostID: hostInfo.HostID,
			keyOsType:   hostInfo.Static.OSType,
			keyCPUArch:  hostInfo.Static.Arch,
			keyInnerIP:  innerIP,
			keyOuterIP:  outerIP,
			keyLoginIP:  hostInfo.Dynamic.LoginIP,
		},
		keyIsMultiTenant: tenant.GetMode() == tenant.ModeMultiple,
		keyConstants:     constants,
	}

	cmdbInstance := map[string]any{
		keyHost: map[string]any{
			keyBkBizID:           hostInfo.Static.BizID,
			keyBkHostID:          hostInfo.HostID,
			keyOsType:            hostInfo.Static.OSType,
			keyBkOSType:          hostInfo.Static.OSTypeCCID,
			keyBkAgentID:         hostInfo.Static.SyncedAgentID,
			keyBkCloudID:         hostInfo.Static.NetworkAreaID,
			keyBkCloudName:       networkArea.Name,
			keyBkHostName:        hostInfo.Static.HostName,
			keyBkAddressing:      hostInfo.Static.Addressing,
			keyBkHostInnerIP:     innerIP,
			keyBkHostOuterIP:     outerIP,
			keyBkHostInnerIPv6:   innerIPv6,
			keyBkHostOuterIPv6:   outerIPv6,
			keyBkCPUArchitecture: hostInfo.Static.Arch,
			keyBkCPU:             hostInfo.Static.CPUNum,
			keyBkMem:             hostInfo.Static.MemCap,
			keyUserName:          hostInfo.Static.Operator,
			keyAccount:           hostInfo.Dynamic.LoginUser,
			keyAuthType:          strings.ToUpper(string(hostInfo.Dynamic.LoginMode)),
			keyHostNodeType:      strings.ToUpper(string(hostInfo.Dynamic.NodeRole)),
		},
	}

	// notice: in v2 plugin, subconfig_path still need provide, and the path is fixed
	if err = pluginV2Utils.CheckDirPathSafe(std.DeployInfo().BaseRuntime.SubConfigDir, std.DeployInfo().Process.Platform.OS); err != nil {
		return nil, fmt.Errorf("check subconfig dir safe failed, dir(%s): %w", std.DeployInfo().BaseRuntime.SubConfigDir, err)
	}

	pluginPath := map[string]any{
		keyLogPath:       std.DeployInfo().BaseRuntime.LogDir,
		keyDataPath:      std.DeployInfo().BaseRuntime.DataDir,
		keyPidPath:       std.DeployInfo().BaseRuntime.RunDir,
		keySetupPath:     std.DeployInfo().BaseRuntime.PluginHomeDir,
		keyEndpoint:      std.DeployInfo().BaseRuntime.DataIPC,
		keyHostID:        std.DeployInfo().BaseRuntime.HostIDPath,
		keySubConfigPath: std.DeployInfo().BaseRuntime.SubConfigDir,
	}

	controlInfo := map[string]any{
		keyPluginIPC:    std.DeployInfo().BaseRuntime.PluginIPC,
		keyDataIPC:      std.DeployInfo().BaseRuntime.DataIPC,
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

func (act *actionEnsureAndUpdatePluginConfigDetailsV2) validateCustomContextBlacklist(customContext map[string]any) error {
	blacklistPrefixKeys := getBlacklistKeys()
	for customKey := range customContext {
		if _, exists := blacklistPrefixKeys[customKey]; exists {
			return fmt.Errorf("the key is reserved, cannot be used, custom-context-key(%s)", customKey)
		}
	}

	return nil
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionEnsureAndUpdatePluginConfigDetailsV2) DisplayNameZh() string {
	return "确保并更新 V2 插件配置详情"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionEnsureAndUpdatePluginConfigDetailsV2) DisplayNameEn() string {
	return "Ensure and Update V2 Plugin Config Details"
}
