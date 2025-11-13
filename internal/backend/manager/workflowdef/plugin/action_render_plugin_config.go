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
	"time"

	pluginUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/plugin/utils"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/release"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/deployconstant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/renderer"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameRenderPluginConfig defines the action name.
	ActionNameRenderPluginConfig = "render_plugin_config"

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

	keyPluginIPC    = "pluginipc"
	keyDataIPC      = "dataipc"
	keyGSEAgentHome = "gse_agent_home"
	keyListenIP     = "listen_ip"
	keyListenPort   = "listen_port"
	keyGroupID      = "group_id"
)

// NewActionRenderPluginConfig ...
func NewActionRenderPluginConfig(capability *Capability) action.Definition {
	return &actionRenderPluginConfig{
		daoHost:             capability.StorageTopo,
		daoNetworkArea:      capability.StorageTopo,
		daoPluginDeployment: capability.StoragePlugin,
		daoPluginRelease:    capability.StorageRelease,
	}
}

// ActParamRenderPluginConfig ...
type ActParamRenderPluginConfig struct {
	pluginUtils.PluginActionStandardParam `json:",inline"`
}

// actionRenderPluginConfig ...
type actionRenderPluginConfig struct {
	daoHost             topoStg.IStorageHost
	daoNetworkArea      topoStg.IStorageNetworkArea
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	daoPluginRelease    release.IPlugin
}

// Name returns the name of the action.
func (act *actionRenderPluginConfig) Name() string {
	return ActionNameRenderPluginConfig
}

// Version returns the version of the action.
func (act *actionRenderPluginConfig) Version() string {
	return "1.0.0"
}

// Description returns the description of the action.
func (act *actionRenderPluginConfig) Description() string {
	return "render plugin config"
}

// Timeout returns the timeout of the action.
func (act *actionRenderPluginConfig) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionRenderPluginConfig) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionRenderPluginConfig) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionRenderPluginConfig) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *actionRenderPluginConfig) Do(ctx *action.InstanceContext) error {
	param := new(ActParamRenderPluginConfig)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	// initialize standard data.
	std := pluginUtils.NewPluginActionStandarder(act.daoPluginDeployment)
	if err = std.Initialize(ctx, param.PluginActionStandardParam); err != nil {
		return err
	}
	defer func() {
		if storeErr := std.Save(); storeErr != nil {
			err = errors.Join(storeErr, err)
		}
	}()

	host, err := act.daoHost.GetHostByID(std.Context(), std.DeployInfo().Process.HostID)
	if err != nil {
		return fmt.Errorf("failed to get host by id. host-id(%d): %w", std.DeployInfo().Process.HostID, err)
	}

	pluginRelease, err := act.daoPluginRelease.GetEnabledReleasePlugin(std.Context(), std.DeployInfo().Process.PluginName,
		std.DeployInfo().Process.Generation, std.DeployInfo().Process.Platform, std.DeployInfo().Process.Info.Version)
	if err != nil {
		return fmt.Errorf("failed to get plugin release info: %w", err)
	}

	renderContext, err := act.getRenderContext(std, host)
	if err != nil {
		return fmt.Errorf("failed to get render context: %w", err)
	}

	configFiles, err := act.daoPluginDeployment.GetPluginDeploymentPluginConfConfigFilesDetail(std.Context(), std.Token())
	if err != nil {
		return fmt.Errorf("failed to get plugin config files detail: %w", err)
	}

	configNameMap := make(map[string]types.PluginPkgConfigTemplate)
	for _, template := range pluginRelease.ConfigTemplates {
		configNameMap[template.Name] = template
	}

	for idx := range configFiles {
		template, ok := configNameMap[configFiles[idx].Name]
		if !ok {
			std.InstanceData().LogE(fmt.Sprintf("template(%s) not found in plugin release", configFiles[idx].Name))
			return fmt.Errorf("template(%s) not found in plugin release", configFiles[idx].Name)
		}

		renderer, err := renderer.NewRenderer(template.TemplateRenderer)
		if err != nil {
			return fmt.Errorf("failed to create template renderer: %w", err)
		}

		std.InstanceData().LogI(fmt.Sprintf("using template renderer(%s) to render template(%s)",
			template.TemplateRenderer, configFiles[idx].Name))

		result, err := renderer.Render(template.SourceContent, renderContext)
		if err != nil {
			logger.G.Sys().WithErr(err).Error("failed to render template")
			return fmt.Errorf("failed to render template: %w", err)
		}

		std.InstanceData().LogI(fmt.Sprintf("rendered plugin(%s-%s-%s) config(%s) success",
			std.DeployInfo().Process.PluginName, std.DeployInfo().Process.Platform.String(),
			std.DeployInfo().Process.Info.Version, configFiles[idx].Name))

		configFiles[idx].Content = result
	}

	if err := std.UpdatePluginConfConfigFilesDetail(configFiles...); err != nil {
		return err
	}

	return nil
}

func (act *actionRenderPluginConfig) getRenderContext(std *pluginUtils.PluginActionStandarder, hostInfo *types.Host) (map[string]any, error) {
	pluginPath, err := act.getPluginPath(std.DeployInfo(), hostInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to get plugin path: %w", err)
	}

	nodeManContext, err := act.getNodeContext(std.DeployInfo(), hostInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to get node context: %w", err)
	}

	cmdbInstance, err := act.getCMDBInstance(std.Context(), hostInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to get cmdb instance: %w", err)
	}

	controlInfo, err := act.getControlInfo(std.DeployInfo(), hostInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to get control info: %w", err)
	}

	customContext, err := act.daoPluginDeployment.GetPluginDeploymentPluginConfCustomConfigContext(std.Context(), std.Token())
	if err != nil {
		return nil, fmt.Errorf("failed to get custom config context: %w", err)
	}

	customContext[keyPluginPath] = pluginPath
	customContext[keyNodeMan] = nodeManContext
	customContext[keyCmdbInstance] = cmdbInstance
	customContext[keyTarget] = cmdbInstance
	customContext[keyControlInfo] = controlInfo

	return customContext, nil
}

func (act *actionRenderPluginConfig) getPluginPath(info *types.PluginDeploymentInfo, hostInfo *types.Host) (map[string]any, error) {
	pluginDeployConf, err := deployconstant.GetPluginDeployConf(
		info.Process.Generation, info.Process.Platform.OS)
	if err != nil {
		return nil, err
	}

	nodeDeployConf, err := deployconstant.GetNodeDeployConf(
		info.Process.Generation, info.Process.Platform.OS)
	if err != nil {
		return nil, err
	}

	paths := map[string]any{
		keyLogPath:       pluginDeployConf.LogDir,
		keyDataPath:      pluginDeployConf.GenerateDefaultDataDir(info.Process.PluginGroup, info.Process.PluginName),
		keyPidPath:       pluginDeployConf.GenerateDefaultRunDir(info.Process.PluginGroup, info.Process.PluginName),
		keySetupPath:     pluginDeployConf.GenerateDefaultSetupPath(info.Process.PluginGroup, info.Process.PluginName),
		keyEndpoint:      nodeDeployConf.GenerateDefaultDataIPCPath(hostInfo.Dynamic.NodeRole),
		keyHostID:        pluginDeployConf.HostIDPath,
		keySubConfigPath: pluginDeployConf.GenerateDefaultSubConfigDir(info.Process.PluginGroup, info.Process.PluginName),
	}

	return paths, nil
}

func (act *actionRenderPluginConfig) getNodeContext(info *types.PluginDeploymentInfo, hostInfo *types.Host) (
	map[string]any, error) {

	pluginDeployConf, err := deployconstant.GetPluginDeployConf(
		info.Process.Generation, info.Process.Platform.OS)
	if err != nil {
		return nil, err
	}

	nodeContext := map[string]any{
		keyHost: map[string]any{
			keyBkHostID: hostInfo.HostID,
			keyOsType:   hostInfo.Dynamic.NodeOsType,
			keyCPUArch:  hostInfo.Dynamic.NodeCPUArch,
			keyInnerIP:  hostInfo.Static.InnerIP,
			keyOuterIP:  hostInfo.Static.OuterIP,
			keyLoginIP:  hostInfo.Dynamic.LoginIP,
		},
		keyIsMultiTenant: tenant.GetMode() == tenant.ModeMultiple,
		keyConstants:     map[string]any{},
	}

	if commonConstants, ok := pluginDeployConf.CommonConstants[info.Process.PluginPkgName]; ok {
		nodeContext[keyConstants] = commonConstants
	}

	if _, ok := pluginDeployConf.CommonConstants[keyGlobal]; !ok {
		return nodeContext, nil
	}

	if constantsMap, ok := nodeContext[keyConstants].(map[string]any); ok {
		constantsMap[keyGlobal] = pluginDeployConf.CommonConstants[keyGlobal]
	}

	return nodeContext, nil
}

func (act *actionRenderPluginConfig) getCMDBInstance(nCtx contextx.IContext, hostInfo *types.Host) (
	map[string]any, error) {

	networkArea, err := act.daoNetworkArea.GetNetworkArea(nCtx, hostInfo.Static.NetworkAreaID)
	if err != nil {
		return nil, fmt.Errorf("failed to get network area by id: %w", err)
	}

	return map[string]any{
		keyHost: map[string]any{
			keyBkBizID:           hostInfo.Static.BizID,
			keyBkHostID:          hostInfo.HostID,
			keyBkOSType:          hostInfo.Static.OSTypeCCID,
			keyBkAgentID:         hostInfo.Static.SyncedAgentID,
			keyBkCloudID:         hostInfo.Static.NetworkAreaID,
			keyBkCloudName:       networkArea.Name,
			keyBkHostName:        hostInfo.Static.HostName,
			keyBkAddressing:      hostInfo.Static.Addressing,
			keyBkHostInnerIP:     hostInfo.Static.InnerIP,
			keyBkHostOuterIP:     hostInfo.Static.OuterIP,
			keyBkHostInnerIPv6:   hostInfo.Static.InnerIPV6,
			keyBkHostOuterIPv6:   hostInfo.Static.OuterIPV6,
			keyBkCPUArchitecture: hostInfo.Static.Arch,
			keyBkCPU:             hostInfo.Static.CPUNum,
			keyBkMem:             hostInfo.Static.MemCap,
		},
	}, nil
}

func (act *actionRenderPluginConfig) getControlInfo(info *types.PluginDeploymentInfo, hostInfo *types.Host) (map[string]any, error) {
	pluginDeployConf, err := deployconstant.GetPluginDeployConf(info.Process.Generation, info.Process.Platform.OS)
	if err != nil {
		return nil, err
	}

	nodeDeployConf, err := deployconstant.GetNodeDeployConf(info.Process.Generation, info.Process.Platform.OS)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		keyPluginIPC:    nodeDeployConf.GenerateDefaultPluginIPCPath(hostInfo.Dynamic.NodeRole),
		keyDataIPC:      nodeDeployConf.GenerateDefaultDataIPCPath(hostInfo.Dynamic.NodeRole),
		keyGSEAgentHome: nodeDeployConf.GenerateNodeHomeDir(hostInfo.Dynamic.NodeRole),
		keyGroupID:      info.Process.PluginGroup,
		keyLogPath:      pluginDeployConf.LogDir,
		keyDataPath:     pluginDeployConf.GenerateDefaultDataDir(info.Process.PluginGroup, info.Process.PluginName),
		keyPidPath:      pluginDeployConf.GenerateDefaultRunDir(info.Process.PluginGroup, info.Process.PluginName),
		keySetupPath:    pluginDeployConf.GenerateDefaultSetupPath(info.Process.PluginGroup, info.Process.PluginName),

		// TODO: implement a plugin to obtain listen ip and port
		keyListenIP:   "",
		keyListenPort: 0,
	}, nil
}
