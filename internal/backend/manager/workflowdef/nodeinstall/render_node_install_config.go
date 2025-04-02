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

// NewActionRenderNodeInstallConfig ...
func NewActionRenderNodeInstallConfig(
	storage nodedeployment.IDaoNodeDeployment,
	iDaoHost topo.IDaoHost,
	iDomainGseProxy topo.IDomainGse,
	logger logger.Logger) operengine.ActionDef {

	return &RenderNodeInstallConfig{
		iDaoNodeDeployment: storage,
		iDaoHost:           iDaoHost,
		iDomainGseProxy:    iDomainGseProxy,
		logger:             logger,
	}
}

// RenderNodeInstallConfigParam ...
type RenderNodeInstallConfigParam struct {
	Token string `json:"token"`
}

// RenderNodeInstallConfig ...
type RenderNodeInstallConfig struct {
	iDaoNodeDeployment nodedeployment.IDaoNodeDeployment
	iDaoHost           topo.IDaoHost
	iDomainGseProxy    topo.IDomainGse

	logger logger.Logger
}

// Name returns the name of the action.
func (action *RenderNodeInstallConfig) Name() string {
	return ActionNameRenderNodeInstallConfig
}

// Version returns the version of the action.
func (action *RenderNodeInstallConfig) Version() string {
	return "1.0.0"
}

// Description returns the description of the action.
func (action *RenderNodeInstallConfig) Description() string {
	return "render node install necessary config"
}

// Timeout returns the timeout of the action.
func (action *RenderNodeInstallConfig) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (action *RenderNodeInstallConfig) Tags() []operengine.ActionTag {
	return []operengine.ActionTag{}
}

// MaxRetryCount returns the max retry count of the action.
func (action *RenderNodeInstallConfig) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (action *RenderNodeInstallConfig) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (action *RenderNodeInstallConfig) Do(ctx *operengine.ActionInstContext) error {
	param := new(RenderNodeInstallConfigParam)
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

func (action *RenderNodeInstallConfig) renderPreSetting(ctx context.Context, nodeConf *types.NodeConf,
	host *types.Host) error {

	if err := action.renderDefaultSetting(nodeConf.PreSetting, host.Dynamic.NodeRole); err != nil {
		return fmt.Errorf("render default setting failed, err: %w", err)
	}

	if err := action.renderLogicSetting(ctx, nodeConf.PreSetting, host); err != nil {
		return fmt.Errorf("render logic setting failed, err: %w", err)
	}

	return nil
}

func (action *RenderNodeInstallConfig) renderDefaultSetting(preSetting map[string]any, nodeRole types.NodeRole,
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
	// GseLogPathKey the config template key of gse log path.
	GseLogPathKey = "__BK_GSE_LOG_PATH__"

	// GseHomeDirKey the config template key of gse home dir.
	GseHomeDirKey = "__BK_GSE_HOME_DIR__"

	// GseAgentBasePluginIPCKey the config template key of gse agent base plugin ipc.
	GseAgentBasePluginIPCKey = "__BK_GSE_AGENT_BASE_PLUGIN_IPC__"

	// GseDataIPCKey the config template key of gse agent base ipc.
	GseDataIPCKey = "__BK_GSE_DATA_IPC__"

	// GseProxyTLSCaFile the config template key of gse proxy tls ca file.
	GseProxyTLSCaFile = "__BK_GSE_PROXY_TLS_CA_FILE__"

	// GseProxyTLSCertFile the config template key of gse proxy tls cert file.
	GseProxyTLSCertFile = "__BK_GSE_PROXY_TLS_CERT_FILE__"

	// GseProxyTLSKeyFile the config template key of gse proxy tls key file.
	GseProxyTLSKeyFile = "__BK_GSE_PROXY_TLS_KEY_FILE__"

	// GseProxyTLSPasswordFile the config template key of gse proxy tls password file.
	GseProxyTLSPasswordFile = "__BK_GSE_PROXY_TLS_PASSWORD_FILE__"

	// GseAgentBaseTLSCAFile the config template key of gse agent base tls ca file.
	GseAgentBaseTLSCAFile = "__BK_GSE_AGENT_BASE_TLS_CA_FILE__"

	// GseAgentBaseTLSCertFile the config template key of gse agent base tls cert file.
	GseAgentBaseTLSCertFile = "__BK_GSE_AGENT_BASE_TLS_CERT_FILE__"

	// GseAgentBaseTLSKeyFile the config template key of gse agent base tls key file.
	GseAgentBaseTLSKeyFile = "__BK_GSE_AGENT_BASE_TLS_KEY_FILE__"

	// GseAgentBaseTLSPasswordFile the config template key of gse agent base tls password file.
	GseAgentBaseTLSPasswordFile = "__BK_GSE_AGENT_BASE_TLS_PASSWORD_FILE__"

	// GseExtraConfigDirectory the config template key of gse extra config directory.
	GseExtraConfigDirectory = "__BK_GSE_EXTRA_CONFIG_DIRECTORY__"

	// GseAccessClusterEndpoints the config template key of gse access cluster endpoints.
	GseAccessClusterEndpoints = "__BK_GSE_ACCESS_CLUSTER_ENDPOINTS__"

	// GseAccessDataEndpoints the config template key of gse access data endpoints.
	GseAccessDataEndpoints = "__BK_GSE_ACCESS_DATA_ENDPOINTS__"

	// GseAccessFileEndpoints the config template key of gse access file endpoints.
	GseAccessFileEndpoints = "__BK_GSE_ACCESS_FILE_ENDPOINTS__"

	// GseDataProxyEndpoints the config template key of gse data proxy endpoints.
	GseDataProxyEndpoints = "__BK_GSE_DATA_PROXY_ENDPOINTS__"

	// GseFileAgentAdvertiseIPV4 the config template key of gse file agent advertise ipv4.
	GseFileAgentAdvertiseIPV4 = "__BK_GSE_FILE_AGENT_ADVERTISE_IPV4__"

	// GseFileAgentAdvertiseIPV6 the config template key of gse file agent advertise ipv6.
	GseFileAgentAdvertiseIPV6 = "__BK_GSE_FILE_AGENT_ADVERTISE_IPV6__"

	// GseFileTopologyAdvertiseIP the config template key of gse file topology advertise ip.
	GseFileTopologyAdvertiseIP = "__BK_GSE_FILE_TOPOLOGY_ADVERTISE_IP__"
)

// renderLogicSetting load logic setting to the config presetting.
// nolint: nonamedreturns
func (action *RenderNodeInstallConfig) renderLogicSetting(ctx context.Context, preSetting map[string]any,
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
	certDir := joinPath(osType, deploymentConf.GseHomeDir, "cert")
	preSetting[GseHomeDirKey] = homeDir

	// base setting
	preSetting[GseAgentBaseTLSCAFile] = joinPath(osType, certDir, "gseca.crt")
	preSetting[GseAgentBaseTLSCertFile] = joinPath(osType, certDir, "gse_agent.crt")
	preSetting[GseAgentBaseTLSKeyFile] = joinPath(osType, certDir, "gse_agent.key")

	// proxy setting
	preSetting[GseProxyTLSCaFile] = joinPath(osType, certDir, "gseca.crt")
	preSetting[GseProxyTLSCertFile] = joinPath(osType, certDir, "gse_server.crt")
	preSetting[GseProxyTLSKeyFile] = joinPath(osType, certDir, "gse_server.key")

	if system.GetEdition() == system.EditionEE {
		preSetting[GseAgentBaseTLSPasswordFile] = joinPath(osType, certDir, "cert_encrypt.key")
		preSetting[GseProxyTLSPasswordFile] = joinPath(osType, certDir, "cert_encrypt.key")
	}

	preSetting[GseExtraConfigDirectory] = joinPath(osType, deploymentConf.GseEnvironDir, "user_conf")
	preSetting[GseLogPathKey] = deploymentConf.GseLogDir
	preSetting[GseAgentBasePluginIPCKey] = deploymentConf.GsePluginIPC
	preSetting[GseDataIPCKey] = deploymentConf.GseDataDir

	switch host.Dynamic.NodeRole {
	case types.NodeRoleAgent:
		{
			clusters, files, datas, err := action.iDomainGseProxy.GetV4AgentAccessEndpoints(ctx, host.Dynamic.NetworkUnitID)
			if err != nil {
				return fmt.Errorf("get agent access endpoints failed, err: %w", err)
			}

			preSetting[GseAccessClusterEndpoints] = strings.Join(clusters, ",")
			preSetting[GseAccessDataEndpoints] = strings.Join(datas, ",")
			preSetting[GseAccessFileEndpoints] = strings.Join(files, ",")
		}
	case types.NodeRoleProxy:
		{
			preSetting[GseFileAgentAdvertiseIPV4] = advertiseIPV4
			preSetting[GseFileAgentAdvertiseIPV6] = advertiseIPV6
			preSetting[GseFileTopologyAdvertiseIP] = advertiseIP

			clusters, files, datas, err := action.iDomainGseProxy.GetProxyUpstreamAccessEndpoints(ctx, host.Dynamic.NetworkUnitID)
			if err != nil {
				return fmt.Errorf("get proxy upstream endpoints failed, err: %w", err)
			}

			preSetting[GseAccessClusterEndpoints] = strings.Join(clusters, ",")
			preSetting[GseAccessFileEndpoints] = advertiseIP
			preSetting[GseAccessDataEndpoints] = advertiseIP

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

// necessaryKeys this defines the necessary keys in custom setting.
func necessaryKeys() []string {
	return []string{
		"run_mode",
		"cloud_id",
		"zone_id",
		"city_id",
	}
}

// forbiddenKeys this defines the forbidden keys in custom setting.
func forbiddenKeys() []string {
	return []string{}
}

// renderCustomSetting load custom setting to the config presetting.
func (action *RenderNodeInstallConfig) renderCustomSetting(conf *types.NodeConf, _ *types.DeploymentInfo) error {
	if conf.CustomSetting == nil {
		return errors.New("lack custom setting")
	}

	for _, key := range necessaryKeys() {
		if _, ok := conf.CustomSetting[key]; !ok {
			return fmt.Errorf("lack necessary key, key(%s)", key)
		}
	}

	for _, key := range forbiddenKeys() {
		if _, ok := conf.CustomSetting[key]; ok {
			return fmt.Errorf("this key is forbidden, key(%s)", key)
		}
	}

	// TODO: Rendering strategy logic

	return nil
}
func (action *RenderNodeInstallConfig) checkHostExist(ctx context.Context, hostID int64) error {
	host, err := action.iDaoHost.GetHostByID(ctx, hostID)
	if err != nil {
		return fmt.Errorf("get host info failed, err: %w", err)
	}

	if host == nil {
		return fmt.Errorf("host not found, host-id(%d)", hostID)
	}

	return nil
}
