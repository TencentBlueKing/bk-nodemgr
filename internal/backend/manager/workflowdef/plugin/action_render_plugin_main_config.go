/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

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
	"path/filepath"
	"time"

	pluginUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/plugin/utils"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/release"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/deployconstant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/winpath"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/templaterender"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameRenderPluginMainConfig defines the action name.
	ActionNameRenderPluginMainConfig = "render_plugin_main_config"

	keyPluginPath   = "plugin_path"
	keyNodeMan      = "nodeman"
	keyCmdbInstance = "cmdb_instance"
	keyTarget       = "target"

	keyLogPath       = "log_path"
	keyDataPath      = "data_path"
	keyPidPath       = "pid_path"
	keySetupPath     = "setup_path"
	keyEndpoint      = "endpoint"
	keyHostID        = "host_id"
	keySubConfigPath = "subconfig_path"

	keyHost = "host"

	keyConstants = "constants"
	keyBkHostID  = "bk_host_id"
	keyOsType    = "os_type"
	keyCPUArch   = "cpu_arch"
	keyInnerIP   = "inner_ip"
	keyOuterIP   = "outer_ip"
	keyLoginIP   = "login_ip"
	keyGlobal    = "global"

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
)

// NewActionRenderPluginMainConfig ...
func NewActionRenderPluginMainConfig(capability *Capability) action.Definition {
	return &RenderPluginMainConfig{
		daoHost:             capability.StorageTopo,
		daoNetworkArea:      capability.StorageTopo,
		daoPluginDeployment: capability.StoragePlugin,
		daoPluginRelease:    capability.StorageRelease,
	}
}

// ActParamRenderPluginMainConfig ...
type ActParamRenderPluginMainConfig struct {
	pluginUtils.PluginActionStandardParam `json:",inline"`
}

// RenderPluginMainConfig ...
type RenderPluginMainConfig struct {
	daoHost             topoStg.IStorageHost
	daoNetworkArea      topoStg.IStorageNetworkArea
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	daoPluginRelease    release.IPlugin
}

// Name returns the name of the action.
func (act *RenderPluginMainConfig) Name() string {
	return ActionNameRenderPluginMainConfig
}

// Version returns the version of the action.
func (act *RenderPluginMainConfig) Version() string {
	return "1.0.0"
}

// Description returns the description of the action.
func (act *RenderPluginMainConfig) Description() string {
	return "render plugin main config"
}

// Timeout returns the timeout of the action.
func (act *RenderPluginMainConfig) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *RenderPluginMainConfig) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *RenderPluginMainConfig) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *RenderPluginMainConfig) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *RenderPluginMainConfig) Do(ctx *action.InstanceContext) error {
	param := new(ActParamRenderPluginMainConfig)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	std := pluginUtils.NewPluginActionStandarder(act.daoPluginDeployment)
	if err = std.Initialize(ctx, param.PluginActionStandardParam); err != nil {
		return err
	}

	host, err := act.daoHost.GetHostByID(std.Context(), std.DeployInfo().Process.HostID)
	if err != nil {
		return fmt.Errorf("failed to get host by id. host-id(%d): %w", std.DeployInfo().Process.HostID, err)
	}

	pluginRelease, err := act.daoPluginRelease.GetEnabledReleasePlugin(std.Context(), std.DeployInfo().Process.PluginName,
		std.DeployInfo().Process.Generation, std.DeployInfo().Process.Platform, std.DeployInfo().Process.Info.Version)
	if err != nil {
		return fmt.Errorf("failed to get plugin release info: %w", err)
	}

	renderContext, err := act.getRenderContext(std.Context(), std.DeployInfo(), host)
	if err != nil {
		return fmt.Errorf("failed to get render context: %w", err)
	}

	var templateContent string
	for _, tmpl := range pluginRelease.ConfigTemplates {
		if tmpl.IsMainConfig {
			templateContent = tmpl.SourceContent
		}
	}

	result, err := templaterender.New().Render(templateContent, renderContext)
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to render template")
		return fmt.Errorf("failed to render template: %w", err)
	}

	ctx.Data.LogI(fmt.Sprintf("rendered plugin(%s-%s-%s) main config success",
		std.DeployInfo().Process.PluginName, std.DeployInfo().Process.Platform.String(),
		std.DeployInfo().Process.Info.Version))

	if err := std.UpdateMainConfig([]byte(result)); err != nil {
		return err
	}

	return nil
}

func (act *RenderPluginMainConfig) getRenderContext(
	nCtx contextx.IContext, info *types.PluginDeploymentInfo, hostInfo *types.Host) (map[string]any, error) {

	pluginPath, err := act.getPluginPath(info)
	if err != nil {
		return nil, fmt.Errorf("failed to get plugin path: %w", err)
	}

	nodeManContext, err := act.getNodeContext(info, hostInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to get node context: %w", err)
	}

	cmdbInstance, err := act.getCMDBInstance(nCtx, hostInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to get cmdb instance: %w", err)
	}

	return map[string]any{
		keyPluginPath:   pluginPath,
		keyNodeMan:      nodeManContext,
		keyCmdbInstance: cmdbInstance,
		keyTarget:       cmdbInstance,
	}, nil
}

func (act *RenderPluginMainConfig) getPluginPath(info *types.PluginDeploymentInfo) (map[string]any, error) {
	pluginDeployConf, err := deployconstant.GetPluginDeployConf(
		info.Process.Generation, info.Process.Platform.OS)
	if err != nil {
		return nil, err
	}

	paths := map[string]any{
		keyLogPath:   pluginDeployConf.LogDir,
		keyDataPath:  pluginDeployConf.DataDir,
		keyPidPath:   info.Process.Identity.PidPath,
		keySetupPath: info.Process.Identity.SetupPath,
		keyEndpoint:  pluginDeployConf.AgentDataIPCPath,
		keyHostID:    pluginDeployConf.HostIDPath,
	}

	if info.Process.Platform.OS == criteria.OSWindows {
		paths[keySubConfigPath] = winpath.Join(info.Process.PluginName, pluginDeployConf.SubConfigBaseDir)
	} else {
		paths[keySubConfigPath] = filepath.Join(info.Process.PluginName, pluginDeployConf.SubConfigBaseDir)
	}

	return paths, nil
}

func (act *RenderPluginMainConfig) getNodeContext(info *types.PluginDeploymentInfo, hostInfo *types.Host) (
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
		keyConstants: map[string]any{},
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

func (act *RenderPluginMainConfig) getCMDBInstance(nCtx contextx.IContext, hostInfo *types.Host) (
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
		},
	}, nil
}
