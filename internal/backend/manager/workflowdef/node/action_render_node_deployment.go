/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package node

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	nodeUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/node/utils"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/configpolicy"
	nodeStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/release"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/tool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/system"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameRenderNodeDeployment defines the action name.
	ActionNameRenderNodeDeployment = "render_node_deployment"
)

// NewActionRenderNodeDeployment get a new action.
func NewActionRenderNodeDeployment(capability *Capability) action.Definition {
	return &actionRenderNodeDeployment{
		storageNodeDeployment: capability.StorageNode,
		storageHost:           capability.StorageTopo,
		storageDomainGse:      capability.StorageTopo,
		storageRelease:        capability.StorageRelease,
		storageConfigPolicy:   capability.StorageConfigPolicy,
	}
}

// ActParamRenderNodeDeployment this is the param for render deployment.
type ActParamRenderNodeDeployment struct {
	nodeUtils.NodeActionStandardParam `json:",inline"`
}

type actionRenderNodeDeployment struct {
	storageNodeDeployment nodeStg.IDaoNodeDeployment
	storageHost           topoStg.IStorageHost
	storageDomainGse      topoStg.IStorageDomainGse
	storageRelease        release.IStorage
	storageConfigPolicy   configpolicy.IStorage
}

// Name returns the name of the action.
func (act *actionRenderNodeDeployment) Name() string {
	return ActionNameRenderNodeDeployment
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionRenderNodeDeployment) DisplayNameZh() string {
	return "渲染节点部署信息"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionRenderNodeDeployment) DisplayNameEn() string {
	return "Render Node Deployment"
}

// Version returns the version of the action.
func (act *actionRenderNodeDeployment) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionRenderNodeDeployment) Description() string {
	return "render node deployment"
}

// Timeout returns the timeout of the action.
func (act *actionRenderNodeDeployment) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionRenderNodeDeployment) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionRenderNodeDeployment) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionRenderNodeDeployment) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
// nolint: perfsprint,funlen,gocognit
func (act *actionRenderNodeDeployment) Do(ctx *action.InstanceContext) error {
	param := new(ActParamRenderNodeDeployment)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	// initialize standard data.
	std := nodeUtils.NewNodeActionStandarder(act.storageNodeDeployment, act.storageHost)
	if err = std.Initialize(ctx, param.NodeActionStandardParam); err != nil {
		return err
	}
	defer func() {
		if storeErr := std.Save(); storeErr != nil {
			err = errors.Join(storeErr, err)
		}
	}()

	// nodeConf comes from db, which means that this node will not overwrite the original configuration in db.
	nodeConf, err := act.storageNodeDeployment.GetNodeDeploymentNodeConf(std.Context(), std.Token())
	if err != nil {
		return fmt.Errorf("failed to get node deployment node conf: %w", err)
	}

	// get release of this node.
	releaseType, err := types.ConvertNodeRoleToReleaseType(std.DeployInfo().Host.Dynamic.NodeRole)
	if err != nil {
		return fmt.Errorf("failed to convert node role to release type: %w", err)
	}

	if err := act.ensureHostDynamicAdvertiseIPAndExportIP(std); err != nil {
		return fmt.Errorf("failed to ensure host dynamic advertise ip and export ip: %w", err)
	}

	// Validate CPU architecture before rendering deployment.
	if err := std.DeployInfo().Host.Dynamic.NodeCPUArch.Validate(); err != nil {
		return fmt.Errorf("failed to validate cpu arch for render: %w", err)
	}

	switch releaseType {
	case types.ReleaseTypeAgent:
		rlsAgent, err := act.getReleaseAgentForRender(std)
		if err != nil {
			return err
		}

		nodeConf.PreSetting = rlsAgent.ReleaseAdditionInfoAgent.ConfigEnviron
		nodeConf.ConfigTemplate = rlsAgent.ReleaseAdditionInfoAgent.ConfigTemplate
	case types.ReleaseTypeProxy:
		rlsProxy, err := act.getReleaseProxyForRender(std)
		if err != nil {
			return err
		}

		nodeConf.PreSetting = rlsProxy.ReleaseAdditionInfoProxy.ConfigEnviron
		nodeConf.ConfigTemplate = rlsProxy.ReleaseAdditionInfoProxy.ConfigTemplate
	default:
		return fmt.Errorf("unsupported release type: %s", releaseType)
	}

	gp := gopool.NewPool()
	gp.Go(func() error {
		if err := act.renderLogicSetting(std, nodeConf); err != nil {
			return fmt.Errorf("failed to render logic setting: %w", err)
		}

		logger.G.Sys().Ctx(std.Context()).With("token", std.Token()).Info("rendered logic setting")

		return nil
	})

	gp.Go(func() error {
		if err := act.renderCustomSetting(std, nodeConf); err != nil {
			return fmt.Errorf("failed to render custom setting: %w", err)
		}

		logger.G.Sys().Ctx(std.Context()).With("token", std.Token()).Info("rendered custom setting")

		return nil
	})

	if err := gp.Wait(); err != nil {
		return fmt.Errorf("failed to render node install config: %w", err)
	}

	if err := act.storageNodeDeployment.SetNodeDeploymentNodeConf(std.Context(), std.Token(), nodeConf); err != nil {
		return fmt.Errorf("failed to set node conf: %w", err)
	}

	act.renderNodeDeploymentInfo(std.Context(), std.DeployInfo(), nodeConf)

	if err := act.storageNodeDeployment.UpdateNodeDeploymentInfo(std.Context(), std.Token(), std.DeployInfo()); err != nil {
		return fmt.Errorf("failed to set node deployment info: %w", err)
	}

	return nil
}

// getReleaseAgentForRender gets agent release for rendering node config.
// If the exact version is not found and AllowReleaseFallback is enabled, it falls back to the default release.
func (act *actionRenderNodeDeployment) getReleaseAgentForRender(
	std *nodeUtils.NodeActionStandarder) (*types.ReleaseAgent, error) {

	plat := platfmt.Platform{OS: std.DeployInfo().Host.Dynamic.NodeOsType, Arch: std.DeployInfo().Host.Dynamic.NodeCPUArch}
	gen := std.DeployInfo().Host.Dynamic.NodeGeneration
	version := std.DeployInfo().Host.Dynamic.NodeVersion

	// check if the exact version exists.
	cond := &types.ReleaseCondition{
		ExactInclude: &types.ReleaseExactFields{
			Platform:   []platfmt.Platform{plat},
			Generation: []types.Generation{gen},
			Version:    []string{version},
			Enabled:    []bool{true},
		},
	}

	releases, _, err := act.storageRelease.ListReleaseAgent(std.Context(), types.UnlimitedPage(), cond)
	if err != nil {
		return nil, fmt.Errorf("failed to list release agent: %w", err)
	}

	if len(releases) > 0 {
		return releases[0], nil
	}

	// version not found, try fallback to default release if allowed.
	if !std.DeployInfo().ReconfigOptions.AllowReleaseFallback {
		return nil, fmt.Errorf("release agent version(%s) not found for platform(%v)", version, plat)
	}

	return act.fallbackDefaultReleaseAgent(std, plat, gen, version)
}

// fallbackDefaultReleaseAgent queries the default agent release and logs the fallback.
func (act *actionRenderNodeDeployment) fallbackDefaultReleaseAgent(
	std *nodeUtils.NodeActionStandarder,
	plat platfmt.Platform, gen types.Generation, originalVersion string,
) (*types.ReleaseAgent, error) {

	cond := &types.ReleaseCondition{
		ExactInclude: &types.ReleaseExactFields{
			Platform:   []platfmt.Platform{plat},
			Generation: []types.Generation{gen},
			AsDefault:  []bool{true},
			Enabled:    []bool{true},
		},
	}

	defaults, _, err := act.storageRelease.ListReleaseAgent(std.Context(), types.UnlimitedPage(), cond)
	if err != nil {
		return nil, fmt.Errorf("release agent version(%s) not found, "+
			"and failed to list default agent release for platform(%v): %w", originalVersion, plat, err)
	}
	if len(defaults) == 0 {
		return nil, fmt.Errorf("release agent version(%s) not found, "+
			"and no default agent release available for platform(%v)", originalVersion, plat)
	}
	if len(defaults) > 1 {
		return nil, fmt.Errorf("release agent version(%s) not found, "+
			"and multiple default agent releases found for platform(%v)", originalVersion, plat)
	}

	rls := defaults[0]
	act.logReleaseFallback(std, originalVersion, rls.Version)

	return rls, nil
}

// getReleaseProxyForRender gets proxy release for rendering node config.
// If the exact version is not found and AllowReleaseFallback is enabled, it falls back to the default release.
func (act *actionRenderNodeDeployment) getReleaseProxyForRender(
	std *nodeUtils.NodeActionStandarder) (*types.ReleaseProxy, error) {

	plat := platfmt.Platform{OS: std.DeployInfo().Host.Dynamic.NodeOsType, Arch: std.DeployInfo().Host.Dynamic.NodeCPUArch}
	gen := std.DeployInfo().Host.Dynamic.NodeGeneration
	version := std.DeployInfo().Host.Dynamic.NodeVersion

	// check if the exact version exists.
	cond := &types.ReleaseCondition{
		ExactInclude: &types.ReleaseExactFields{
			Platform:   []platfmt.Platform{plat},
			Generation: []types.Generation{gen},
			Version:    []string{version},
			Enabled:    []bool{true},
		},
	}

	releases, _, err := act.storageRelease.ListReleaseProxy(std.Context(), types.UnlimitedPage(), cond)
	if err != nil {
		return nil, fmt.Errorf("failed to list release proxy: %w", err)
	}

	if len(releases) > 0 {
		return releases[0], nil
	}

	// version not found, try fallback to default release if allowed.
	if !std.DeployInfo().ReconfigOptions.AllowReleaseFallback {
		return nil, fmt.Errorf("release proxy version(%s) not found for platform(%v)", version, plat)
	}

	return act.fallbackDefaultReleaseProxy(std, plat, gen, version)
}

// fallbackDefaultReleaseProxy queries the default proxy release and logs the fallback.
func (act *actionRenderNodeDeployment) fallbackDefaultReleaseProxy(
	std *nodeUtils.NodeActionStandarder,
	plat platfmt.Platform, gen types.Generation, originalVersion string,
) (*types.ReleaseProxy, error) {

	cond := &types.ReleaseCondition{
		ExactInclude: &types.ReleaseExactFields{
			Platform:   []platfmt.Platform{plat},
			Generation: []types.Generation{gen},
			AsDefault:  []bool{true},
			Enabled:    []bool{true},
		},
	}

	defaults, _, err := act.storageRelease.ListReleaseProxy(std.Context(), types.UnlimitedPage(), cond)
	if err != nil {
		return nil, fmt.Errorf("release proxy version(%s) not found, "+
			"and failed to list default proxy release for platform(%v): %w", originalVersion, plat, err)
	}
	if len(defaults) == 0 {
		return nil, fmt.Errorf("release proxy version(%s) not found, "+
			"and no default proxy release available for platform(%v)", originalVersion, plat)
	}
	if len(defaults) > 1 {
		return nil, fmt.Errorf("release proxy version(%s) not found, "+
			"and multiple default proxy releases found for platform(%v)", originalVersion, plat)
	}

	rls := defaults[0]
	act.logReleaseFallback(std, originalVersion, rls.Version)

	return rls, nil
}

// logReleaseFallback records fallback behavior in both system log and workflow log.
func (act *actionRenderNodeDeployment) logReleaseFallback(
	std *nodeUtils.NodeActionStandarder,
	originalVersion, fallbackVersion string,
) {

	logger.G.Sys().Ctx(std.Context()).With("token", std.Token()).
		With("original_version", originalVersion).
		With("fallback_version", fallbackVersion).
		Warn("release not found, fallback to default release for reconfig")

	std.InstanceData().Log().
		Zh("原始版本(%s)的 release 不存在, 回退使用默认版本(%s)的配置模板",
			originalVersion, fallbackVersion).
		En("release for version(%s) not found, fallback to default release version(%s) for reconfig",
			originalVersion, fallbackVersion).
		Warn()
}

const (
	// GseTemplateKeyRunMode the config template key of gse run mode.
	GseTemplateKeyRunMode = "__BK_GSE_RUN_MODE__"

	// GseTemplateKeyCloudID the config template key of gse cloud id.
	GseTemplateKeyCloudID = "__BK_GSE_CLOUD_ID__"

	// GseTemplateKeyZoneID the config template key of gse zone id.
	GseTemplateKeyZoneID = "__BK_GSE_ZONE_ID__"

	// GseTemplateKeyCityID the config template key of gse city id.
	GseTemplateKeyCityID = "__BK_GSE_CITY_ID__"

	// GseTemplateKeyLogPath the config template key of gse log path.
	GseTemplateKeyLogPath = "__BK_GSE_LOG_PATH__"

	// GseTemplateKeyHomeDir the config template key of gse home dir.
	GseTemplateKeyHomeDir = "__BK_GSE_HOME_DIR__"

	// GseTemplateKeyAgentBasePluginIPC the config template key of gse agent base plugin ipc.
	GseTemplateKeyAgentBasePluginIPC = "__BK_GSE_AGENT_BASE_PLUGIN_IPC__"

	// GseTemplateKeyDataIPC the config template key of gse agent base ipc.
	GseTemplateKeyDataIPC = "__BK_GSE_DATA_IPC__"

	// GseTemplateKeyProxyTLSCaFile the config template key of gse proxy tls ca file.
	GseTemplateKeyProxyTLSCaFile = "__BK_GSE_PROXY_TLS_CA_FILE__"

	// GseTemplateKeyProxyTLSCertFile the config template key of gse proxy tls cert file.
	GseTemplateKeyProxyTLSCertFile = "__BK_GSE_PROXY_TLS_CERT_FILE__"

	// GseTemplateKeyProxyTLSKeyFile the config template key of gse proxy tls key file.
	GseTemplateKeyProxyTLSKeyFile = "__BK_GSE_PROXY_TLS_KEY_FILE__"

	// GseTemplateKeyProxyTLSPasswordFile the config template key of gse proxy tls password file.
	GseTemplateKeyProxyTLSPasswordFile = "__BK_GSE_PROXY_TLS_PASSWORD_FILE__"

	// GseTemplateKeyAgentBaseTLSCAFile the config template key of gse agent base tls ca file.
	GseTemplateKeyAgentBaseTLSCAFile = "__BK_GSE_AGENT_BASE_TLS_CA_FILE__"

	// GseTemplateKeyAgentBaseTLSCertFile the config template key of gse agent base tls cert file.
	GseTemplateKeyAgentBaseTLSCertFile = "__BK_GSE_AGENT_BASE_TLS_CERT_FILE__"

	// GseTemplateKeyAgentBaseTLSKeyFile the config template key of gse agent base tls key file.
	GseTemplateKeyAgentBaseTLSKeyFile = "__BK_GSE_AGENT_BASE_TLS_KEY_FILE__"

	// GseTemplateKeyAgentBaseTLSPasswordFile the config template key of gse agent base tls password file.
	GseTemplateKeyAgentBaseTLSPasswordFile = "__BK_GSE_AGENT_BASE_TLS_PASSWORD_FILE__"

	// GseTemplateKeyDataAgentTLSCaFile the config template key of gse data agent tls ca file.
	GseTemplateKeyDataAgentTLSCaFile = "__BK_GSE_DATA_AGENT_TLS_CA_FILE__"

	// GseTemplateKeyDataAgentTLSCertFile the config template key of gse data agent tls cert file.
	GseTemplateKeyDataAgentTLSCertFile = "__BK_GSE_DATA_AGENT_TLS_CERT_FILE__"

	// GseTemplateKeyDataAgentTLSKeyFile the config template key of gse data agent tls key file.
	GseTemplateKeyDataAgentTLSKeyFile = "__BK_GSE_DATA_AGENT_TLS_KEY_FILE__"

	// GseTemplateKeyDataAgentTLSPasswordFile the config template key of gse data agent tls password file.
	GseTemplateKeyDataAgentTLSPasswordFile = "__BK_GSE_DATA_AGENT_TLS_PASSWORD_FILE__"

	// GseTemplateKeyDataProxyTLSCaFile the config template key of gse data proxy tls ca file.
	GseTemplateKeyDataProxyTLSCaFile = "__BK_GSE_DATA_PROXY_TLS_CA_FILE__"

	// GseTemplateKeyDataProxyTLSCertFile the config template key of gse data proxy tls cert file.
	GseTemplateKeyDataProxyTLSCertFile = "__BK_GSE_DATA_PROXY_TLS_CERT_FILE__"

	// GseTemplateKeyDataProxyTLSKeyFile the config template key of gse data proxy tls key file.
	GseTemplateKeyDataProxyTLSKeyFile = "__BK_GSE_DATA_PROXY_TLS_KEY_FILE__"

	// GseTemplateKeyDataProxyTLSPasswordFile the config template key of gse data proxy tls password file.
	GseTemplateKeyDataProxyTLSPasswordFile = "__BK_GSE_DATA_PROXY_TLS_PASSWORD_FILE__"

	// GseTemplateKeyFileAgentTLSCaFile the config template key of gse file agent tls ca file.
	GseTemplateKeyFileAgentTLSCaFile = "__BK_GSE_FILE_AGENT_TLS_CA_FILE__"

	// GseTemplateKeyFileAgentTLSCertFile the config template key of gse file agent tls cert file.
	GseTemplateKeyFileAgentTLSCertFile = "__BK_GSE_FILE_AGENT_TLS_CERT_FILE__"

	// GseTemplateKeyFileAgentTLSKeyFile the config template key of gse file agent tls key file.
	GseTemplateKeyFileAgentTLSKeyFile = "__BK_GSE_FILE_AGENT_TLS_KEY_FILE__"

	// GseTemplateKeyFileAgentTLSPasswordFile the config template key of gse file agent tls password file.
	GseTemplateKeyFileAgentTLSPasswordFile = "__BK_GSE_FILE_AGENT_TLS_PASSWORD_FILE__"

	// GseTemplateKeyFileTopologyTLSCaFile the config template key of gse file topology tls ca file.
	GseTemplateKeyFileTopologyTLSCaFile = "__BK_GSE_FILE_TOPOLOGY_TLS_CA_FILE__"

	// GseTemplateKeyFileTopologyTLSPasswordFile the config template key of gse file topology tls password file.
	GseTemplateKeyFileTopologyTLSPasswordFile = "__BK_GSE_FILE_TOPOLOGY_TLS_PASSWORD_FILE__"

	// GseTemplateKeyFileTopologyTLSSvrCertFile the config template key of gse file topology tls svr cert file.
	GseTemplateKeyFileTopologyTLSSvrCertFile = "__BK_GSE_FILE_TOPOLOGY_TLS_SVR_CERT_FILE__"

	// GseTemplateKeyFileTopologyTLSSvrKeyFile the config template key of gse file topology tls svr key file.
	GseTemplateKeyFileTopologyTLSSvrKeyFile = "__BK_GSE_FILE_TOPOLOGY_TLS_SVR_KEY_FILE__"

	// GseTemplateKeyFileTopologyTLSCliCertFile the config template key of gse file topology tls cli cert file.
	GseTemplateKeyFileTopologyTLSCliCertFile = "__BK_GSE_FILE_TOPOLOGY_TLS_CLI_CERT_FILE__"

	// GseTemplateKeyFileTopologyTLSCliKeyFile the config template key of gse file topology tls cli key file.
	GseTemplateKeyFileTopologyTLSCliKeyFile = "__BK_GSE_FILE_TOPOLOGY_TLS_CLI_KEY_FILE__"

	// GseTemplateKeyExtraConfigDirectory the config template key of gse extra config directory.
	GseTemplateKeyExtraConfigDirectory = "__BK_GSE_EXTRA_CONFIG_DIRECTORY__"

	// GseTemplateKeyAccessClusterEndpoints the config template key of gse access cluster endpoints.
	GseTemplateKeyAccessClusterEndpoints = "__BK_GSE_ACCESS_CLUSTER_ENDPOINTS__"

	// GseTemplateKeyAccessDataEndpoints the config template key of gse access data endpoints.
	GseTemplateKeyAccessDataEndpoints = "__BK_GSE_ACCESS_DATA_ENDPOINTS__"

	// GseTemplateKeyAccessFileEndpoints the config template key of gse access file endpoints.
	GseTemplateKeyAccessFileEndpoints = "__BK_GSE_ACCESS_FILE_ENDPOINTS__"

	// GseDataProxyEndpoints the config template key of gse data proxy endpoints.
	GseDataProxyEndpoints = "__BK_GSE_DATA_PROXY_ENDPOINTS__"

	// GseTemplateKeyFileAgentAdvertiseIPV4 the config template key of gse file agent advertise ipv4.
	GseTemplateKeyFileAgentAdvertiseIPV4 = "__BK_GSE_FILE_AGENT_ADVERTISE_IPV4__"

	// GseTemplateKeyFileAgentAdvertiseIPV6 the config template key of gse file agent advertise ipv6.
	GseTemplateKeyFileAgentAdvertiseIPV6 = "__BK_GSE_FILE_AGENT_ADVERTISE_IPV6__"

	// GseTemplateKeyFileTopologyAdvertiseIP the config template key of gse file topology advertise ip.
	GseTemplateKeyFileTopologyAdvertiseIP = "__BK_GSE_FILE_TOPOLOGY_ADVERTISE_IP__"

	// GseTemplateKeyEnableStaticAccess the config template key of gse enable static access.
	GseTemplateKeyEnableStaticAccess = "__BK_GSE_ENABLE_STATIC_ACCESS__"

	// GseTemplateKeyDataAgentBindPort the config template key of gse data agent bind port.
	GseTemplateKeyDataAgentBindPort = "__BK_GSE_DATA_AGENT_BIND_PORT__"

	// GseTemplateKeyFileAgentBindPort the config template key of gse file agent bind port.
	GseTemplateKeyFileAgentBindPort = "__BK_GSE_FILE_AGENT_BIND_PORT__"

	// GseTemplateKeyFileTopologyBindPort the config template key of gse file topology bind port.
	GseTemplateKeyFileTopologyBindPort = "__BK_GSE_FILE_TOPOLOGY_BIND_PORT__"

	// GseTemplateKeyProxyBindPort the config template key of gse proxy bind port.
	GseTemplateKeyProxyBindPort = "__BK_GSE_PROXY_BIND_PORT__"
)

const (
	// GseCustomKeyFileTopologyLinks the config template key of gse file topology links.
	GseCustomKeyFileTopologyLinks = "file.topology.links"

	// GseCustomKeyFileTopologyProxyGroupTag the config template key of gse file topology proxy group tag.
	// this config is used to split the proxy group for file topology, set this config will only allow the machines
	// in the same unit to transfer files to each other.
	GseCustomKeyFileTopologyProxyGroupTag = "file.topology.proxy_group_tag"
)

// renderLogicSetting load logic setting to the config presetting and custom setting .
// nolint: nonamedreturns,funlen,gocognit
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (act *actionRenderNodeDeployment) renderLogicSetting(std *nodeUtils.NodeActionStandarder, nodeConf *types.NodeConf) error {
	// this is a special case, when the deployment is reverted, the host id is not in the host table.
	if err := act.checkHostExist(std.Context(), std.DeployInfo().Host.HostID); err != nil {
		return err
	}

	osType := std.DeployInfo().Host.Dynamic.NodeOsType

	advertiseIPV4 := std.DeployInfo().Host.Dynamic.AdvertiseIP
	advertiseIPV6 := std.DeployInfo().Host.Dynamic.AdvertiseIPV6
	advertiseIP := advertiseIPV4
	if advertiseIP == "" {
		advertiseIP = advertiseIPV6
	}

	nodeConf.PreSetting[GseTemplateKeyRunMode] = std.DeployInfo().Host.Dynamic.NodeRole
	nodeConf.PreSetting[GseTemplateKeyCloudID] = std.DeployInfo().Host.Static.NetworkAreaID
	nodeConf.PreSetting[GseTemplateKeyZoneID] = std.DeployInfo().Host.Static.RegionID
	nodeConf.PreSetting[GseTemplateKeyCityID] = std.DeployInfo().Host.Static.CityID

	homeDir := std.DeployInfo().BaseRuntime.HomeDir
	certDir := tool.JoinPath(osType, homeDir, "cert")
	nodeConf.PreSetting[GseTemplateKeyHomeDir] = homeDir

	gseCaFilePath := tool.JoinPath(osType, certDir, "gseca.crt")
	gsePasswordFilePath := tool.JoinPath(osType, certDir, "cert_encrypt.key")
	gseAgentCertFilePath := tool.JoinPath(osType, certDir, "gse_agent.crt")
	gseAgentKeyFilePath := tool.JoinPath(osType, certDir, "gse_agent.key")
	gseServerCertFilePath := tool.JoinPath(osType, certDir, "gse_server.crt")
	gseServerKeyFilePath := tool.JoinPath(osType, certDir, "gse_server.key")
	gseAPIClientCertFilePath := tool.JoinPath(osType, certDir, "gse_api_client.crt")
	gseAPIClientKeyFilePath := tool.JoinPath(osType, certDir, "gse_api_client.key")

	// base setting
	// always set tls settings no matter it is agent or proxy.
	// cause agent and proxy share the same config template.
	nodeConf.PreSetting[GseTemplateKeyAgentBaseTLSCAFile] = gseCaFilePath
	nodeConf.PreSetting[GseTemplateKeyAgentBaseTLSCertFile] = gseAgentCertFilePath
	nodeConf.PreSetting[GseTemplateKeyAgentBaseTLSKeyFile] = gseAgentKeyFilePath
	nodeConf.PreSetting[GseTemplateKeyProxyTLSCaFile] = gseCaFilePath
	nodeConf.PreSetting[GseTemplateKeyProxyTLSCertFile] = gseServerCertFilePath
	nodeConf.PreSetting[GseTemplateKeyProxyTLSKeyFile] = gseServerKeyFilePath

	if system.GetEdition() == system.EditionEE || system.GetEdition() == system.EditionInner {
		nodeConf.PreSetting[GseTemplateKeyAgentBaseTLSPasswordFile] = gsePasswordFilePath
		nodeConf.PreSetting[GseTemplateKeyProxyTLSPasswordFile] = gsePasswordFilePath
	} else {
		nodeConf.PreSetting[GseTemplateKeyAgentBaseTLSPasswordFile] = ""
		nodeConf.PreSetting[GseTemplateKeyProxyTLSPasswordFile] = ""
	}

	if std.DeployInfo().Host.Dynamic.NodeRole == types.NodeRoleProxy {
		// proxy setting
		nodeConf.PreSetting[GseTemplateKeyDataAgentTLSCaFile] = gseCaFilePath
		nodeConf.PreSetting[GseTemplateKeyDataAgentTLSCertFile] = gseServerCertFilePath
		nodeConf.PreSetting[GseTemplateKeyDataAgentTLSKeyFile] = gseServerKeyFilePath
		nodeConf.PreSetting[GseTemplateKeyDataAgentTLSPasswordFile] = gsePasswordFilePath
		nodeConf.PreSetting[GseTemplateKeyDataProxyTLSCaFile] = gseCaFilePath
		nodeConf.PreSetting[GseTemplateKeyDataProxyTLSCertFile] = gseAgentCertFilePath
		nodeConf.PreSetting[GseTemplateKeyDataProxyTLSKeyFile] = gseAgentKeyFilePath
		nodeConf.PreSetting[GseTemplateKeyDataProxyTLSPasswordFile] = gsePasswordFilePath

		nodeConf.PreSetting[GseTemplateKeyFileAgentTLSCaFile] = gseCaFilePath
		nodeConf.PreSetting[GseTemplateKeyFileAgentTLSCertFile] = gseServerCertFilePath
		nodeConf.PreSetting[GseTemplateKeyFileAgentTLSKeyFile] = gseServerKeyFilePath
		nodeConf.PreSetting[GseTemplateKeyFileAgentTLSPasswordFile] = gsePasswordFilePath
		nodeConf.PreSetting[GseTemplateKeyFileTopologyTLSCaFile] = gseCaFilePath
		nodeConf.PreSetting[GseTemplateKeyFileTopologyTLSPasswordFile] = gsePasswordFilePath
		nodeConf.PreSetting[GseTemplateKeyFileTopologyTLSSvrCertFile] = gseServerCertFilePath
		nodeConf.PreSetting[GseTemplateKeyFileTopologyTLSSvrKeyFile] = gseServerKeyFilePath
		nodeConf.PreSetting[GseTemplateKeyFileTopologyTLSCliCertFile] = gseAPIClientCertFilePath
		nodeConf.PreSetting[GseTemplateKeyFileTopologyTLSCliKeyFile] = gseAPIClientKeyFilePath
	}

	nodeConf.PreSetting[GseTemplateKeyExtraConfigDirectory] = std.DeployInfo().BaseRuntime.ExtraConfigDir
	nodeConf.PreSetting[GseTemplateKeyLogPath] = std.DeployInfo().BaseRuntime.LogDir
	nodeConf.PreSetting[GseTemplateKeyAgentBasePluginIPC] = std.DeployInfo().BaseRuntime.PluginIPC
	nodeConf.PreSetting[GseTemplateKeyDataIPC] = std.DeployInfo().BaseRuntime.DataIPC
	needStaticAccess, err := act.storageDomainGse.NeedStaticAccess(std.Context(), std.DeployInfo().Host.Dynamic.NetworkUnitID)
	if err != nil {
		return fmt.Errorf("failed to get need static access: %w", err)
	}
	nodeConf.PreSetting[GseTemplateKeyEnableStaticAccess] = needStaticAccess

	// render access endpoints
	switch std.DeployInfo().Host.Dynamic.NodeRole {
	case types.NodeRoleAgent:
		{
			clusters, files, datas, err := act.storageDomainGse.GetV4AgentAccessEndpoints(std.Context(), std.DeployInfo().Host.Dynamic.NetworkUnitID)
			if err != nil {
				return fmt.Errorf("failed to get agent access endpoints: %w", err)
			}

			if len(clusters) == 0 {
				return fmt.Errorf("agent cluster access endpoints is empty")
			}
			if len(files) == 0 {
				return fmt.Errorf("agent file access endpoints is empty")
			}
			if len(datas) == 0 {
				return fmt.Errorf("agent data access endpoints is empty")
			}

			nodeConf.PreSetting[GseTemplateKeyAccessClusterEndpoints] = strings.Join(clusters, ",")
			nodeConf.PreSetting[GseTemplateKeyAccessDataEndpoints] = strings.Join(datas, ",")
			nodeConf.PreSetting[GseTemplateKeyAccessFileEndpoints] = strings.Join(files, ",")
		}
	case types.NodeRoleProxy:
		{
			nodeConf.PreSetting[GseTemplateKeyFileAgentAdvertiseIPV4] = advertiseIPV4
			nodeConf.PreSetting[GseTemplateKeyFileAgentAdvertiseIPV6] = advertiseIPV6
			nodeConf.PreSetting[GseTemplateKeyFileTopologyAdvertiseIP] = advertiseIP
			clusters, files, datas, err := act.storageDomainGse.GetProxyUpstreamAccessEndpoints(std.Context(), std.DeployInfo().Host.Dynamic.NetworkUnitID)
			if err != nil {
				return fmt.Errorf("get proxy upstream endpoints failed: %w", err)
			}

			nodeConf.PreSetting[GseTemplateKeyAccessClusterEndpoints] = strings.Join(clusters, ",")
			nodeConf.PreSetting[GseTemplateKeyAccessFileEndpoints] =
				fmt.Sprintf("%s:%v", advertiseIP, nodeConf.PreSetting[GseTemplateKeyFileAgentBindPort])
			nodeConf.PreSetting[GseTemplateKeyAccessDataEndpoints] =
				fmt.Sprintf("%s:%v", advertiseIP, nodeConf.PreSetting[GseTemplateKeyDataAgentBindPort])

			nodeConf.PreSetting[GseDataProxyEndpoints] = strings.Join(datas, ",")

			nodeConf.CustomSetting[GseCustomKeyFileTopologyLinks] = act.renderFileLinks(nodeConf, &std.DeployInfo().Host, files)

			// use the network unit id as the proxy group tag for file topology.
			nodeConf.CustomSetting[GseCustomKeyFileTopologyProxyGroupTag] = strconv.Itoa(int(std.DeployInfo().Host.Dynamic.NetworkUnitID))
		}
	default:
		return fmt.Errorf("unsupported node role: %s", std.DeployInfo().Host.Dynamic.NodeRole)
	}

	return nil
}

const (
	// GseCustomKeyAgentRunMode the config template key of gse run mode.
	GseCustomKeyAgentRunMode = "agent.run_mode"

	// GseCustomKeyAgentCloudID the config template key of gse cloud id.
	GseCustomKeyAgentCloudID = "agent.cloud_id"

	// GseCustomKeyAgentZoneID the config template key of gse zone id.
	GseCustomKeyAgentZoneID = "agent.zone_id"

	// GseCustomKeyAgentCityID the config template key of gse city id.
	GseCustomKeyAgentCityID = "agent.city_id"
)

// forbiddenKeys this defines the forbidden keys in custom setting.
func forbiddenKeys() []string {
	return []string{
		GseCustomKeyAgentRunMode,
		GseCustomKeyAgentCloudID,
		GseCustomKeyAgentZoneID,
		GseCustomKeyAgentCityID,
	}
}

// renderCustomSetting load custom setting to the config presetting.
func (act *actionRenderNodeDeployment) renderCustomSetting(std *nodeUtils.NodeActionStandarder, conf *types.NodeConf) error {
	matchResult, err := act.storageConfigPolicy.MatchConfigPolicyNode(std.Context(),
		std.DeployInfo().Host.Static.BizID,
		std.DeployInfo().Host.Static.NetworkAreaID,
		std.DeployInfo().Host.Dynamic.NetworkUnitID,
		std.DeployInfo().Host.Dynamic.NodeOsType,
		std.DeployInfo().Host.Dynamic.NodeCPUArch,
		std.DeployInfo().Host.Dynamic.NodeRole,
		std.DeployInfo().Host.HostID)
	if err != nil {
		return fmt.Errorf("match config policy failed. "+
			"biz-id(%d), networkarea-id(%d), networkunit-id(%d), os-type(%s), cpu-arch(%s), role(%s), host-id(%d): %w",
			std.DeployInfo().Host.Static.BizID,
			std.DeployInfo().Host.Static.NetworkAreaID,
			std.DeployInfo().Host.Dynamic.NetworkUnitID,
			std.DeployInfo().Host.Dynamic.NodeOsType,
			std.DeployInfo().Host.Dynamic.NodeCPUArch,
			std.DeployInfo().Host.Dynamic.NodeRole,
			std.DeployInfo().Host.HostID,
			err)
	}

	std.InstanceData().Log().
		Zh("匹配配置策略. "+
			"biz-id(%d), networkarea-id(%d), networkunit-id(%d), os-type(%s), cpu-arch(%s), role(%s), host-id(%d)",
			std.DeployInfo().Host.Static.BizID,
			std.DeployInfo().Host.Static.NetworkAreaID,
			std.DeployInfo().Host.Dynamic.NetworkUnitID,
			std.DeployInfo().Host.Dynamic.NodeOsType,
			std.DeployInfo().Host.Dynamic.NodeCPUArch,
			std.DeployInfo().Host.Dynamic.NodeRole,
			std.DeployInfo().Host.HostID).
		En("match config policy. "+
			"biz-id(%d), networkarea-id(%d), networkunit-id(%d), os-type(%s), cpu-arch(%s), role(%s), host-id(%d)",
			std.DeployInfo().Host.Static.BizID,
			std.DeployInfo().Host.Static.NetworkAreaID,
			std.DeployInfo().Host.Dynamic.NetworkUnitID,
			std.DeployInfo().Host.Dynamic.NodeOsType,
			std.DeployInfo().Host.Dynamic.NodeCPUArch,
			std.DeployInfo().Host.Dynamic.NodeRole,
			std.DeployInfo().Host.HostID).
		Info()

	if len(matchResult.MatchedPolicies) > 0 {
		for _, p := range matchResult.MatchedPolicies {
			std.InstanceData().Log().
				Zh("命中原始策略. configpolicy-id(%d), configpolicy-name(%s), priority(%d)", p.PolicyID, p.PolicyName, p.Priority).
				En("matched original policy. configpolicy-id(%d), configpolicy-name(%s), priority(%d)", p.PolicyID, p.PolicyName, p.Priority).
				Info()
		}

		conf.CustomSetting = matchResult.MergedConfig
		std.InstanceData().Log().
			Zh("将按照优先级合并配置策略, 并应用到节点配置").
			En("merge config policies by priority and apply to node config").
			Info()
	} else {
		std.InstanceData().Log().
			Zh("未命中任何策略. 保持默认配置").
			En("no policy matched. keep default config").
			Info()
	}

	if conf.CustomSetting != nil {
		for _, key := range forbiddenKeys() {
			if _, ok := conf.CustomSetting[key]; ok {
				return fmt.Errorf("this key is forbidden, key(%s)", key)
			}
		}
	}

	return nil
}

func (act *actionRenderNodeDeployment) checkHostExist(nCtx contextx.IContext, hostID int64) error {
	host, err := act.storageHost.GetHostByID(nCtx, hostID)
	if err != nil {
		return fmt.Errorf("get host info failed: %w", err)
	}

	if host == nil {
		return fmt.Errorf("host not found, host-id(%d)", hostID)
	}

	return nil
}

// FileLink file link.
type FileLink struct {
	TargetIP   string `json:"target_ip,omitempty" bson:"target_ip,omitempty"`
	TargetPort int64  `json:"target_port,omitempty" bson:"target_port,omitempty"`
	ReportIP   string `json:"report_ip,omitempty" bson:"report_ip,omitempty"`
	ReportPort int64  `json:"report_port,omitempty" bson:"report_port,omitempty"`
}

// NewFileLink new file link.
// nolint: mnd
func NewFileLink() *FileLink {
	return &FileLink{
		TargetIP:   "127.0.0.1",
		TargetPort: defaultFileLinkTargetPort,
		ReportIP:   "127.0.0.1",
		ReportPort: defaultFileLinkTargetPort,
	}
}

const defaultFileLinkTargetPort = 28930

func (act *actionRenderNodeDeployment) renderFileLinks(nodeConf *types.NodeConf, host *types.Host,
	fileUpstreams []string) []FileLink {

	reportIP := func() string {
		outerIps := host.Static.GetOuterIPList()
		if len(outerIps) > 0 {
			return outerIps[0]
		}

		innerIPs := host.Static.GetInnerIPList()
		if len(innerIPs) > 0 {
			return innerIPs[0]
		}

		return ""
	}()

	links := make([]FileLink, 0, len(fileUpstreams))
	for _, upstream := range fileUpstreams {
		strs := strings.Split(upstream, ":")

		link := FileLink{
			TargetIP:   strs[0],
			TargetPort: conv.ToInt64Default(strs[1], defaultFileLinkTargetPort),
			ReportIP:   reportIP,
			ReportPort: conv.ToInt64Default(
				nodeConf.PreSetting[GseTemplateKeyFileTopologyBindPort], defaultFileLinkTargetPort),
		}
		links = append(links, link)
	}

	return links
}

const (
	defaultKeyProxyBindPort = 28668
	defaultKeyProxyDataPort = 28625
	defaultKeyProxyFilePort = 28925
)

func (act *actionRenderNodeDeployment) renderNodeDeploymentInfo(
	_ context.Context,
	info *types.DeploymentInfo,
	conf *types.NodeConf) {

	info.Host.Dynamic.ProxyClusterPort = conv.ToInt64Default(
		conf.PreSetting[GseTemplateKeyProxyBindPort], defaultKeyProxyBindPort)
	info.Host.Dynamic.ProxyDataPort = conv.ToInt64Default(
		conf.PreSetting[GseTemplateKeyDataAgentBindPort], defaultKeyProxyDataPort)
	info.Host.Dynamic.ProxyFilePort = conv.ToInt64Default(
		conf.PreSetting[GseTemplateKeyFileAgentBindPort], defaultKeyProxyFilePort)
}

func (act *actionRenderNodeDeployment) ensureHostDynamicAdvertiseIPAndExportIP(std *nodeUtils.NodeActionStandarder) error {
	advertiseIPV4 := std.DeployInfo().Host.Dynamic.AdvertiseIP
	advertiseIPV6 := std.DeployInfo().Host.Dynamic.AdvertiseIPV6
	if advertiseIPV4 == "" && len(std.DeployInfo().Host.Static.InnerIPList) > 0 {
		advertiseIPV4 = std.DeployInfo().Host.Static.InnerIPList[0]
	}
	if advertiseIPV6 == "" && len(std.DeployInfo().Host.Static.InnerIPV6List) > 0 {
		advertiseIPV6 = std.DeployInfo().Host.Static.InnerIPV6List[0]
	}

	std.DeployInfo().Host.Dynamic.AdvertiseIP = advertiseIPV4
	std.DeployInfo().Host.Dynamic.AdvertiseIPV6 = advertiseIPV6

	if advertiseIPV4 == "" && advertiseIPV6 == "" {
		return fmt.Errorf("advertise ipv4 and ipv6 are empty")
	}

	exportIPV4 := std.DeployInfo().Host.Dynamic.ExportIP
	exportIPV6 := std.DeployInfo().Host.Dynamic.ExportIPV6
	if exportIPV4 == "" && len(std.DeployInfo().Host.Static.InnerIPList) > 0 {
		exportIPV4 = std.DeployInfo().Host.Static.InnerIPList[0]
	}
	if exportIPV6 == "" && len(std.DeployInfo().Host.Static.InnerIPV6List) > 0 {
		exportIPV6 = std.DeployInfo().Host.Static.InnerIPV6List[0]
	}

	std.DeployInfo().Host.Dynamic.ExportIP = exportIPV4
	std.DeployInfo().Host.Dynamic.ExportIPV6 = exportIPV6

	if exportIPV4 == "" && exportIPV6 == "" {
		return fmt.Errorf("export ipv4 and ipv6 are empty")
	}

	return nil
}
