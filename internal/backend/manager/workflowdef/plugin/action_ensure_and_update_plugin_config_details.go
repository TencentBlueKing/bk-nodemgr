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
	"errors"
	"fmt"
	"strings"
	"time"

	pluginUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/plugin/utils"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/release"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/deployconstant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
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

	pluginDeployConstant deployconstant.PluginDeployConf
	nodeDeployConstant   deployconstant.NodeDeployConf
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
func (act *actionEnsureAndUpdatePluginConfigDetails) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *actionEnsureAndUpdatePluginConfigDetails) Do(ctx *action.InstanceContext) error {
	param := new(ActParamEnsureAndUpdatePluginConfigDetails)
	err := conv.MapToStruct(ctx.Data.Content, param)
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

	act.nodeDeployConstant, err = deployconstant.GetNodeDeployConf(std.DeployInfo().Process.Generation, std.DeployInfo().Process.Platform.OS)
	if err != nil {
		return fmt.Errorf("failed to get node deploy constant: %w", err)
	}

	act.pluginDeployConstant, err = deployconstant.GetPluginDeployConf(std.DeployInfo().Process.Generation, std.DeployInfo().Process.Platform.OS)
	if err != nil {
		return fmt.Errorf("failed to get plugin deploy constant: %w", err)
	}

	hostInfo, err := act.daoHost.GetHostByID(std.Context(), std.DeployInfo().Process.HostID)
	if err != nil {
		return fmt.Errorf("failed to get host by id, host-id(%d): %w", std.DeployInfo().Process.HostID, err)
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
	fillConfigDetails(pluginRelease, pluginConf)

	if err := act.daoPluginDeployment.UpdatePluginDeploymentPluginConf(std.Context(), std.Token(), pluginConf); err != nil {
		return fmt.Errorf("failed to update plugin deployment plugin conf: %w", err)
	}

	return nil
}

func fillConfigDetails(pluginRelease *types.ReleasePlugin, pluginConf *types.PluginDeploymentPluginConf) {
	templateMap := make(map[string]types.PluginPkgConfigTemplate, len(pluginRelease.ConfigTemplates))
	var mainTemplate *types.PluginConfigDetail

	for _, tpl := range pluginRelease.ConfigTemplates {
		templateMap[tpl.Name] = tpl

		if tpl.IsMainConfig && mainTemplate == nil {
			mainTemplate = &types.PluginConfigDetail{
				Name:         tpl.Name,
				Content:      tpl.SourceContent,
				IsMainConfig: tpl.IsMainConfig,
			}
		}
	}

	for idx, detail := range pluginConf.ConfigFilesDetail {
		if tpl, ok := templateMap[detail.Name]; ok {
			pluginConf.ConfigFilesDetail[idx].Content = tpl.SourceContent
			pluginConf.ConfigFilesDetail[idx].IsMainConfig = tpl.IsMainConfig
		}
	}

	if len(pluginConf.ConfigFilesDetail) == 0 && mainTemplate != nil {
		pluginConf.ConfigFilesDetail = []*types.PluginConfigDetail{mainTemplate}
	}
}

// ContextPluginInfo plugin info for render context.
type ContextPluginInfo struct {
	Name          string
	LogPath       string
	DataPath      string
	PidPath       string
	SetupPath     string
	HostIDPath    string
	PluginIPC     string
	DataIPC       string
	AgentDir      string
	GroupID       string
	SubConfigPath string
	IsMultiTenant bool
}

// ContextPreDefinitionConstants pre-definition constants for render context.
type ContextPreDefinitionConstants struct {
	Global map[string]any
	Unique map[string]any
}

// ContextNodeStaticInfo node static info for render context.
type ContextNodeStaticInfo struct {
	BizID         int64
	NetworkAreaID int64
	RegionID      string
	CityID        string
	HostName      string
	DeptName      string
	InnerIPList   []string
	InnerIPV6List []string
	OuterIPList   []string
	OuterIPV6List []string
	Operator      string
	Mac           string
	OSTypeCCID    string
	OSType        string
	Arch          string
	Addressing    string
	CPUNum        float64
	MemCap        float64
}

// ContextNodeDynamicInfo node dynamic info for render context.
type ContextNodeDynamicInfo struct {
	NodeRole                 string
	NodeStatus               string
	NodeVersion              string
	NodeGeneration           int64
	NodeCPUArch              string
	NodeOsType               string
	AgentID                  string
	NetworkUnitID            int64
	LoginUser                string
	ExportIP                 string
	ExportIPV6               string
	AdvertiseIP              string
	AdvertiseIPV6            string
	ProxyAccessDisabled      bool
	ProxyTags                []string
	ProxyClusterPort         int64
	ProxyDataPort            int64
	ProxyFilePort            int64
	RelayDownloadPort        int64
	RelayCallbackPort        int64
	ProxyInstallOriginUnitID int64
}

// ContextNodeInfo node info for render context.
type ContextNodeInfo struct {
	HostID   int64
	TenantID string
	Static   ContextNodeStaticInfo
	Dynamic  ContextNodeDynamicInfo
}

func convertHostTypeToContextNodeInfo(hostInfo *types.Host) ContextNodeInfo {
	proxyTags := make([]string, 0, len(hostInfo.Dynamic.ProxyTags))
	for _, tag := range hostInfo.Dynamic.ProxyTags {
		proxyTags = append(proxyTags, string(tag))
	}

	return ContextNodeInfo{
		HostID:   hostInfo.HostID,
		TenantID: hostInfo.TenantID,
		Static: ContextNodeStaticInfo{
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
		Dynamic: ContextNodeDynamicInfo{
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
		LogPath:       act.pluginDeployConstant.LogDir,
		DataPath:      act.pluginDeployConstant.GenerateDefaultDataDir(std.DeployInfo().Process.PluginGroup, std.DeployInfo().Process.PluginName),
		PidPath:       act.pluginDeployConstant.GenerateDefaultRunDir(std.DeployInfo().Process.PluginGroup, std.DeployInfo().Process.PluginName),
		SetupPath:     act.pluginDeployConstant.GenerateDefaultSetupPath(std.DeployInfo().Process.PluginGroup, std.DeployInfo().Process.PluginName),
		HostIDPath:    act.nodeDeployConstant.HostIDPath,
		PluginIPC:     act.nodeDeployConstant.GenerateDefaultPluginIPCPath(hostInfo.Dynamic.NodeRole),
		DataIPC:       act.nodeDeployConstant.GenerateDefaultDataIPCPath(hostInfo.Dynamic.NodeRole),
		AgentDir:      act.nodeDeployConstant.DeployDir,
		GroupID:       std.DeployInfo().Process.PluginGroup,
		SubConfigPath: act.pluginDeployConstant.GenerateDefaultSubConfigDir(std.DeployInfo().Process.PluginGroup, std.DeployInfo().Process.PluginName),
		IsMultiTenant: tenant.GetMode() == tenant.ModeMultiple,
	}

	preDefinitionConstants := ContextPreDefinitionConstants{
		Unique: act.pluginDeployConstant.GetCommonConstants(std.DeployInfo().Process.PluginPkgName),
		Global: act.pluginDeployConstant.GetCommonConstants(keyGlobal),
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

func (act *actionEnsureAndUpdatePluginConfigDetails) generateJinja2SystemConfigContext(
	std *pluginUtils.PluginActionStandarder, hostInfo *types.Host) (map[string]any, error) {

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

	constants := act.pluginDeployConstant.GetCommonConstants(std.DeployInfo().Process.PluginName)
	constants[keyGlobal] = act.pluginDeployConstant.GetCommonConstants(keyGlobal)
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

	pluginPath := map[string]any{
		keyLogPath:       act.pluginDeployConstant.LogDir,
		keyDataPath:      act.pluginDeployConstant.GenerateDefaultDataDir(std.DeployInfo().Process.PluginGroup, std.DeployInfo().Process.PluginName),
		keyPidPath:       act.pluginDeployConstant.GenerateDefaultRunDir(std.DeployInfo().Process.PluginGroup, std.DeployInfo().Process.PluginName),
		keySetupPath:     act.pluginDeployConstant.GenerateDefaultSetupPath(std.DeployInfo().Process.PluginGroup, std.DeployInfo().Process.PluginName),
		keyEndpoint:      act.nodeDeployConstant.GenerateDefaultDataIPCPath(hostInfo.Dynamic.NodeRole),
		keyHostID:        act.nodeDeployConstant.HostIDPath,
		keySubConfigPath: act.pluginDeployConstant.GenerateDefaultSubConfigDir(std.DeployInfo().Process.PluginGroup, std.DeployInfo().Process.PluginName),
	}

	controlInfo := map[string]any{
		keyPluginIPC:    act.nodeDeployConstant.GenerateDefaultPluginIPCPath(hostInfo.Dynamic.NodeRole),
		keyDataIPC:      act.nodeDeployConstant.GenerateDefaultDataIPCPath(hostInfo.Dynamic.NodeRole),
		keyGSEAgentHome: act.nodeDeployConstant.DeployDir,
		keyGroupID:      std.DeployInfo().Process.PluginGroup,
		keyLogPath:      act.pluginDeployConstant.LogDir,
		keyDataPath:     act.pluginDeployConstant.GenerateDefaultDataDir(std.DeployInfo().Process.PluginGroup, std.DeployInfo().Process.PluginName),
		keyPidPath:      act.pluginDeployConstant.GenerateDefaultRunDir(std.DeployInfo().Process.PluginGroup, std.DeployInfo().Process.PluginName),
		keySetupPath:    act.pluginDeployConstant.GenerateDefaultSetupPath(std.DeployInfo().Process.PluginGroup, std.DeployInfo().Process.PluginName),
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

func (act *actionEnsureAndUpdatePluginConfigDetails) validateCustomContextBlacklist(customContext map[string]any) error {
	blacklistPrefixKeys := getBlacklistKeys()
	for customKey := range customContext {
		if _, exists := blacklistPrefixKeys[customKey]; exists {
			return fmt.Errorf("the key is reserved, cannot be used, custom-context-key(%s)", customKey)
		}
	}

	return nil
}
