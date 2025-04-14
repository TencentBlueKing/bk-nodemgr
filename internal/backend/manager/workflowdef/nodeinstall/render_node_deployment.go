/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package nodeinstall

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/nodedeployment"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/deployconstant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/system"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operengine"
)

// NewActionRenderNodeDeployment new action to render node deployment.
func NewActionRenderNodeDeployment(
	storage nodedeployment.IDaoNodeDeployment,
	iDaoHost topo.IDaoHost,
	iDomainGseProxy topo.IDomainGse,
	logger logger.Logger) operengine.ActionDef {

	return &RenderNodeDeployment{
		iDaoNodeDeployment: storage,
		iDaoHost:           iDaoHost,
		IDomainGse:         iDomainGseProxy,
		logger:             logger,
	}
}

// RenderNodeDeploymentParam this is the param for render deployment.
type RenderNodeDeploymentParam struct {
	Token string `json:"token"`
}

// RenderNodeDeployment this is the action to render node deployment.
type RenderNodeDeployment struct {
	iDaoNodeDeployment nodedeployment.IDaoNodeDeployment
	iDaoHost           topo.IDaoHost
	topo.IDomainGse

	logger logger.Logger
}

// Name returns the name of the action.
func (action *RenderNodeDeployment) Name() string {
	return ActionNameRenderNodeDeployment
}

// Version returns the version of the action.
func (action *RenderNodeDeployment) Version() string {
	return "1.0.0"
}

// Description returns the description of the action.
func (action *RenderNodeDeployment) Description() string {
	return "render node deployment"
}

// Timeout returns the timeout of the action.
func (action *RenderNodeDeployment) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (action *RenderNodeDeployment) Tags() []operengine.ActionTag {
	return []operengine.ActionTag{}
}

// MaxRetryCount returns the max retry count of the action.
func (action *RenderNodeDeployment) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (action *RenderNodeDeployment) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (action *RenderNodeDeployment) Do(ctx *operengine.ActionInstContext) error {
	param := new(RenderNodeDeploymentParam)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	info, err := action.iDaoNodeDeployment.GetInfo(ctx.Ctx, param.Token)
	if err != nil {
		return err
	}

	tenantCtx, err := tenant.SetID(ctx.Ctx, info.TenantID)
	if err != nil {
		return err
	}

	info.OperInstID = ctx.Data.OperInstID
	info.BlockingActionName = ActionNameWaitComplete

	if err := action.iDaoNodeDeployment.UpdateInfo(tenantCtx, param.Token, info); err != nil {
		return fmt.Errorf("set node conf failed, err: %w", err)
	}

	// nodeConf comes from db, which means that this node will not overwrite the original configuration in db.
	nodeConf, err := action.iDaoNodeDeployment.GetNodeConf(tenantCtx, param.Token)
	if err != nil {
		return fmt.Errorf("get node conf failed, err: %w", err)
	}

	gp := gopool.NewPool()
	gp.Go(func() error {
		if err := action.renderPreSetting(tenantCtx, nodeConf, &info.Host); err != nil {
			return fmt.Errorf("render pre setting failed, err: %w", err)
		}

		action.logger.Infof("succfessfully render pre setting, token: %s", param.Token)

		return nil
	})

	gp.Go(func() error {
		if err := action.renderCustomSetting(nodeConf, info); err != nil {
			return fmt.Errorf("render custom setting failed, err: %w", err)
		}

		action.logger.Infof("succfessfully render custom setting, token: %s", param.Token)

		return nil
	})

	if err := gp.Wait(); err != nil {
		return fmt.Errorf("failed to render node install config, err: %w", err)
	}

	if err := action.iDaoNodeDeployment.SetNodeConf(tenantCtx, param.Token, nodeConf); err != nil {
		return fmt.Errorf("set node conf failed, err: %w", err)
	}

	return nil
}

func (action *RenderNodeDeployment) renderPreSetting(ctx context.Context, nodeConf *types.NodeConf,
	host *types.Host) error {

	if err := action.renderDefaultSetting(nodeConf.PreSetting, host.Dynamic.NodeRole); err != nil {
		return fmt.Errorf("render default setting failed, err: %w", err)
	}

	if err := action.renderLogicSetting(ctx, nodeConf.PreSetting, host); err != nil {
		return fmt.Errorf("render logic setting failed, err: %w", err)
	}

	return nil
}

func (action *RenderNodeDeployment) renderDefaultSetting(preSetting map[string]any, nodeRole types.NodeRole,
) error {

	switch nodeRole {
	case types.NodeRoleAgent:
		for k, v := range GseAgentSettingDefault() {
			if preSetting[k] == nil {
				preSetting[k] = v
			}
		}
	case types.NodeRoleProxy:
		for k, v := range GseProxySettingDefault() {
			if preSetting[k] == nil {
				preSetting[k] = v
			}
		}
	default:
		return fmt.Errorf("unsupported node role: %s", nodeRole)
	}

	return nil
}

// GseAgentSettingDefault return default gse agent setting
// nolint: mnd
func GseAgentSettingDefault() map[string]any {
	return map[string]any{
		"__BK_GSE_HOME_DIR__":                           "/usr/local/gse/agent",
		"__BK_GSE_RUN_MODE__":                           "agent",
		"__BK_GSE_CLOUD_ID__":                           0,
		"__BK_GSE_ZONE_ID__":                            "default",
		"__BK_GSE_CITY_ID__":                            "default",
		"__BK_GSE_ENABLE_STATIC_ACCESS__":               false,
		"__BK_GSE_ENABLE_FAKE_SEED__":                   false,
		"__BK_GSE_ACCESS_CLUSTER_ENDPOINTS__":           "127.0.0.1:28668",
		"__BK_GSE_ACCESS_DATA_ENDPOINTS__":              "127.0.0.1:28625",
		"__BK_GSE_ACCESS_FILE_ENDPOINTS__":              "127.0.0.1:28925",
		"__BK_GSE_AGENT_BASE_TLS_CA_FILE__":             "",
		"__BK_GSE_AGENT_BASE_TLS_CERT_FILE__":           "",
		"__BK_GSE_AGENT_BASE_TLS_KEY_FILE__":            "",
		"__BK_GSE_AGENT_BASE_TLS_PASSWORD_FILE__":       "",
		"__BK_GSE_AGENT_BASE_PROCESSOR_NUM__":           4,
		"__BK_GSE_AGENT_BASE_PROCESSOR_QUEUE_SIZE__":    4096,
		"__BK_GSE_AGENT_BASE_ALARM_EVENT_DATA_ID__":     1000,
		"__BK_GSE_AGENT_BASE_PLUGIN_IPC__":              "${BK_GSE_HOME_DIR}/data/ipc.state.message",
		"__BK_GSE_PROXY_TLS_CA_FILE__":                  "",
		"__BK_GSE_PROXY_TLS_CERT_FILE__":                "",
		"__BK_GSE_PROXY_TLS_KEY_FILE__":                 "",
		"__BK_GSE_PROXY_TLS_PASSWORD_FILE__":            "",
		"__BK_GSE_PROXY_BIND_IP__":                      "::",
		"__BK_GSE_PROXY_BIND_PORT__":                    28668,
		"__BK_GSE_PROXY_THREAD_NUM__":                   4,
		"__BK_GSE_TASK_PROC_EVENT_DATA_ID__":            1100008,
		"__BK_GSE_TASK_CONCURRENCE_COUNT__":             100,
		"__BK_GSE_TASK_SCRIPT_FILE_EXPIRE_TIME_HOUR__":  72,
		"__BK_GSE_DATA_IPC__":                           "${BK_GSE_HOME_DIR}/data/ipc.state.report",
		"__BK_GSE_DATA_ENABLE_COMPRESSION__":            false,
		"__BK_GSE_FILE_MAX_TRANSFER_SPEED_MB_PER_SEC__": 100,
		"__BK_GSE_FILE_MAX_TRANSFER_CONCURRENT_NUM__":   10,
		"__BK_GSE_FILE_BT_LISTEN_INTERFACE__":           "",
		"__BK_GSE_FILE_BT_OUTGOING_INTERFACE__":         "",
		"__BK_GSE_FILE_BT_ENABLE_OUTGOING_INTERFACE__":  true,
		"__BK_GSE_LOG_PATH__":                           "${BK_GSE_HOME_DIR}/logs",
		"__BK_GSE_LOG_LEVEL__":                          "INFO",
		"__BK_GSE_LOG_FILESIZE_MB__":                    200,
		"__BK_GSE_LOG_FILENUM__":                        10,
		"__BK_GSE_LOG_ROTATE__":                         0,
		"__BK_GSE_LOG_FLUSH_INTERVAL_MS__":              100,
		"__BK_GSE_EXTRA_CONFIG_DIRECTORY__":             "",
	}
}

// GseProxySettingDefault return default gse proxy setting
// nolint: mnd
func GseProxySettingDefault() map[string]any {
	return map[string]any{
		"__BK_GSE_HOME_DIR__":                               "/usr/local/gse/proxy",
		"__BK_GSE_CLOUD_ID__":                               0,
		"__BK_GSE_ZONE_ID__":                                "default",
		"__BK_GSE_CITY_ID__":                                "default",
		"__BK_GSE_DATA_AGENT_BIND_IP__":                     "::",
		"__BK_GSE_DATA_AGENT_BIND_PORT__":                   28625,
		"__BK_GSE_DATA_AGENT_THREAD_NUM__":                  24,
		"__BK_GSE_DATA_MAX_MESSAGE_SIZE__":                  10485760,
		"__BK_GSE_DATA_AGENT_TLS_CA_FILE__":                 "",
		"__BK_GSE_DATA_AGENT_TLS_CERT_FILE__":               "",
		"__BK_GSE_DATA_AGENT_TLS_KEY_FILE__":                "",
		"__BK_GSE_DATA_AGENT_TLS_PASSWORD_FILE__":           "",
		"__BK_GSE_DATA_PROXY_TLS_CA_FILE__":                 "",
		"__BK_GSE_DATA_PROXY_TLS_CERT_FILE__":               "",
		"__BK_GSE_DATA_PROXY_TLS_KEY_FILE__":                "",
		"__BK_GSE_DATA_PROXY_TLS_PASSWORD_FILE__":           "",
		"__BK_GSE_DATA_PROXY_ENDPOINTS__":                   "127.0.0.1:28625",
		"__BK_GSE_DATA_METRIC_EXPORTER_BIND_IP__":           "::",
		"__BK_GSE_DATA_METRIC_EXPORTER_BIND_PORT__":         29402,
		"__BK_GSE_DATA_METRIC_EXPORTER_THREAD_NUM__":        8,
		"__BK_GSE_FILE_AGENT_BIND_IP__":                     "::",
		"__BK_GSE_FILE_AGENT_BIND_PORT__":                   28925,
		"__BK_GSE_FILE_AGENT_BIND_PORT_V1__":                58925,
		"__BK_GSE_FILE_AGENT_ADVERTISE_IPV4__":              "127.0.0.1",
		"__BK_GSE_FILE_AGENT_ADVERTISE_IPV6__":              "::1",
		"__BK_GSE_FILE_AGENT_THREAD_NUM__":                  24,
		"__BK_GSE_FILE_AGENT_TLS_CA_FILE__":                 "",
		"__BK_GSE_FILE_AGENT_TLS_CERT_FILE__":               "",
		"__BK_GSE_FILE_AGENT_TLS_KEY_FILE__":                "",
		"__BK_GSE_FILE_AGENT_TLS_PASSWORD_FILE__":           "",
		"__BK_GSE_FILE_BITTORRENT_BIND_IP__":                "::",
		"__BK_GSE_FILE_BITTORRENT_BIND_PORT__":              10020,
		"__BK_GSE_FILE_BITTORRENT_TRACKER_BIND_PORT__":      10030,
		"__BK_GSE_FILE_BITTORRENT_SPEED_LIMIT_MB_PER_SEC__": 10000,
		"__BK_GSE_FILE_TOPOLOGY_BIND_IP__":                  "::",
		"__BK_GSE_FILE_TOPOLOGY_BIND_PORT__":                28930,
		"__BK_GSE_FILE_TOPOLOGY_THRIFT_BIND_PORT__":         58930,
		"__BK_GSE_FILE_TOPOLOGY_ADVERTISE_IP__":             "127.0.0.1",
		"__BK_GSE_FILE_TOPOLOGY_THREAD_NUM__":               4,
		"__BK_GSE_FILE_TOPOLOGY_TLS_CA_FILE__":              "",
		"__BK_GSE_FILE_TOPOLOGY_TLS_PASSWORD_FILE__":        "",
		"__BK_GSE_FILE_TOPOLOGY_TLS_SVR_CERT_FILE__":        "",
		"__BK_GSE_FILE_TOPOLOGY_TLS_SVR_KEY_FILE__":         "",
		"__BK_GSE_FILE_TOPOLOGY_TLS_CLI_CERT_FILE__":        "",
		"__BK_GSE_FILE_TOPOLOGY_TLS_CLI_KEY_FILE__":         "",
		"__BK_GSE_FILE_PROXY_UPSTREAM_IP__":                 "127.0.0.1",
		"__BK_GSE_FILE_PROXY_UPSTREAM_PORT__":               28930,
		"__BK_GSE_FILE_PROXY_REPORT_IP__":                   "127.0.0.1",
		"__BK_GSE_FILE_PROXY_REPORT_PORT__":                 28930,
		"__BK_GSE_FILE_CACHE_DIRS__":                        "./file_cache",
		"__BK_GSE_FILE_CACHE_EXPIRED_TIME_SEC__":            7200,
		"__BK_GSE_FILE_METRIC_EXPORTER_BIND_IP__":           "::",
		"__BK_GSE_FILE_METRIC_EXPORTER_BIND_PORT__":         29404,
		"__BK_GSE_FILE_METRIC_EXPORTER_THREAD_NUM__":        8,
		"__BK_GSE_LOG_PATH__":                               "${BK_GSE_HOME_DIR}/logs",
		"__BK_GSE_LOG_LEVEL__":                              "INFO",
		"__BK_GSE_LOG_FILESIZE_MB__":                        200,
		"__BK_GSE_LOG_FILENUM__":                            10,
		"__BK_GSE_LOG_ROTATE__":                             0,
		"__BK_GSE_LOG_FLUSH_INTERVAL_MS__":                  100,
	}
}

const (
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
)

// renderLogicSetting load logic setting to the config presetting.
// nolint: nonamedreturns,funlen
func (action *RenderNodeDeployment) renderLogicSetting(ctx context.Context, preSetting map[string]any,
	host *types.Host) (err error) {

	osType, err := platform.NormalizeOS(host.Static.OSType)
	if err != nil {
		return err
	}

	// this is a special case, when the deployment is reverted, the host id is not in the host table.
	if err := action.checkHostExist(ctx, host.HostID); err != nil {
		return err
	}

	deploymentConf, err := deployconstant.GetDeployConf(host.Dynamic.NodeGeneration, osType)
	if err != nil {
		return fmt.Errorf("get deploy conf failed, err: %w", err)
	}

	advertiseIPV4 := strings.Split(host.Static.InnerIP, ",")[0]
	advertiseIPV6 := strings.Split(host.Static.InnerIPV6, ",")[0]
	advertiseIP := func() string {
		if advertiseIPV4 != "" {
			return advertiseIPV4
		}

		return advertiseIPV6
	}()

	homeDir := joinPath(osType, deploymentConf.GseHomeDir, string(host.Dynamic.NodeRole))
	certDir := joinPath(osType, homeDir, "cert")
	preSetting[GseTemplateKeyHomeDir] = homeDir

	// base setting
	preSetting[GseTemplateKeyAgentBaseTLSCAFile] = joinPath(osType, certDir, "gseca.crt")
	preSetting[GseTemplateKeyAgentBaseTLSCertFile] = joinPath(osType, certDir, "gse_agent.crt")
	preSetting[GseTemplateKeyAgentBaseTLSKeyFile] = joinPath(osType, certDir, "gse_agent.key")

	// proxy setting
	preSetting[GseTemplateKeyProxyTLSCaFile] = joinPath(osType, certDir, "gseca.crt")
	preSetting[GseTemplateKeyProxyTLSCertFile] = joinPath(osType, certDir, "gse_server.crt")
	preSetting[GseTemplateKeyProxyTLSKeyFile] = joinPath(osType, certDir, "gse_server.key")

	if system.GetEdition() == system.EditionEE {
		preSetting[GseTemplateKeyAgentBaseTLSPasswordFile] = joinPath(osType, certDir, "cert_encrypt.key")
		preSetting[GseTemplateKeyProxyTLSPasswordFile] = joinPath(osType, certDir, "cert_encrypt.key")
	}

	preSetting[GseTemplateKeyExtraConfigDirectory] = joinPath(osType, deploymentConf.GseEnvironDir, "user_conf")
	preSetting[GseTemplateKeyLogPath] = deploymentConf.GseLogDir
	preSetting[GseTemplateKeyAgentBasePluginIPC] = deploymentConf.GsePluginIPC
	preSetting[GseTemplateKeyDataIPC] = deploymentConf.GseDataDir

	switch host.Dynamic.NodeRole {
	case types.NodeRoleAgent:
		{
			clusters, files, datas, err := action.GetV4AgentAccessEndpoints(ctx, host.Dynamic.NetworkUnitID)
			if err != nil {
				return fmt.Errorf("get agent access endpoints failed, err: %w", err)
			}

			preSetting[GseTemplateKeyAccessClusterEndpoints] = strings.Join(clusters, ",")
			preSetting[GseTemplateKeyAccessDataEndpoints] = strings.Join(datas, ",")
			preSetting[GseTemplateKeyAccessFileEndpoints] = strings.Join(files, ",")
		}
	case types.NodeRoleProxy:
		{
			preSetting[GseTemplateKeyFileAgentAdvertiseIPV4] = advertiseIPV4
			preSetting[GseTemplateKeyFileAgentAdvertiseIPV6] = advertiseIPV6
			preSetting[GseTemplateKeyFileTopologyAdvertiseIP] = advertiseIP

			clusters, files, datas, err := action.GetProxyUpstreamAccessEndpoints(ctx, host.Dynamic.NetworkUnitID)
			if err != nil {
				return fmt.Errorf("get proxy upstream endpoints failed, err: %w", err)
			}

			preSetting[GseTemplateKeyAccessClusterEndpoints] = strings.Join(clusters, ",")
			preSetting[GseTemplateKeyAccessFileEndpoints] = advertiseIP
			preSetting[GseTemplateKeyAccessDataEndpoints] = advertiseIP

			// TODO: save files upstreams for file-proxy links
			_ = files

			preSetting[GseDataProxyEndpoints] = strings.Join(datas, ",")
		}
	default:
		return fmt.Errorf("unsupported node role: %s", host.Dynamic.NodeRole)
	}

	return nil
}

// joinPath joins path elements with the specified separator.
func joinPath(osType string, parts ...string) string {
	separator := "/"
	if osType == criteria.OSWindows {
		separator = "\\"
	}

	if len(parts) == 0 {
		return ""
	}

	result := parts[0]
	for _, part := range parts[1:] {
		if part == "" {
			continue
		}

		if result != "" && !strings.HasSuffix(result, separator) {
			result += separator
		}
		result += part
	}

	return result
}

const (
	// GseKeyRunMode the config template key of gse run mode.
	GseKeyRunMode = "run_mode"
	// GseKeyCloudID the config template key of gse cloud id.
	GseKeyCloudID = "cloud_id"

	// GseKeyZoneID the config template key of gse zone id.
	GseKeyZoneID = "zone_id"

	// GseKeyCityID the config template key of gse city id.
	GseKeyCityID = "city_id"
)

// forbiddenKeys this defines the forbidden keys in custom setting.
func forbiddenKeys() []string {
	return []string{
		GseKeyRunMode,
		"cloud_id",
		"zone_id",
		"city_id",
	}
}

// renderCustomSetting load custom setting to the config presetting.
func (action *RenderNodeDeployment) renderCustomSetting(conf *types.NodeConf, info *types.DeploymentInfo) error {
	if conf.CustomSetting == nil {
		return errors.New("lack custom setting")
	}

	for _, key := range forbiddenKeys() {
		if _, ok := conf.CustomSetting[key]; ok {
			return fmt.Errorf("this key is forbidden, key(%s)", key)
		}
	}

	conf.CustomSetting[GseKeyRunMode] = info.Dynamic.NodeRole
	conf.CustomSetting[GseKeyCloudID] = info.Static.NetworkAreaID
	conf.CustomSetting[GseKeyZoneID] = info.Static.RegionID
	conf.CustomSetting[GseKeyCityID] = info.Static.CityID

	// TODO: Rendering strategy logic

	return nil
}
func (action *RenderNodeDeployment) checkHostExist(ctx context.Context, hostID int64) error {
	host, err := action.iDaoHost.GetHostByID(ctx, hostID)
	if err != nil {
		return fmt.Errorf("get host info failed, err: %w", err)
	}

	if host == nil {
		return fmt.Errorf("host not found, host-id(%d)", hostID)
	}

	return nil
}
