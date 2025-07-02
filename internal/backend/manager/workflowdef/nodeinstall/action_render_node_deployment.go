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

	nodedeployment "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node-deployment"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/deployconstant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/system"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameRenderNodeDeployment defines the action name.
	ActionNameRenderNodeDeployment = "render_node_deployment"
)

// NewActionRenderNodeDeployment get a new action.
func NewActionRenderNodeDeployment(
	storageNodeDeployment nodedeployment.IStorageNodeDeployment,
	storageHost topo.IStorageHost,
	storageDomainGse topo.IStorageDomainGse,
	logger logger.Logger) action.Definition {

	return &actionRenderNodeDeployment{
		storageNodeDeployment: storageNodeDeployment,
		storageHost:           storageHost,
		storageDomainGse:      storageDomainGse,
		logger:                logger,
	}
}

// ActParamRenderNodeDeployment this is the param for render deployment.
type ActParamRenderNodeDeployment struct {
	Token string `json:"token"`
}

type actionRenderNodeDeployment struct {
	storageNodeDeployment nodedeployment.IStorageNodeDeployment
	storageHost           topo.IStorageHost
	storageDomainGse      topo.IStorageDomainGse

	logger logger.Logger
}

// Name returns the name of the action.
func (act *actionRenderNodeDeployment) Name() string {
	return ActionNameRenderNodeDeployment
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
func (act *actionRenderNodeDeployment) Do(ctx *action.InstanceContext) error {
	param := new(ActParamRenderNodeDeployment)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	info, err := act.storageNodeDeployment.GetInfo(ctx.Ctx, param.Token)
	if err != nil {
		return err
	}

	tenantCtx, err := tenant.SetID(ctx.Ctx, info.TenantID)
	if err != nil {
		return err
	}

	info.OperInstID = ctx.Data.OperationInstanceID
	info.BlockingActionName = ActionNameWaitInstallComplete

	if err := act.storageNodeDeployment.UpdateInfo(tenantCtx, param.Token, info); err != nil {
		return fmt.Errorf("set node conf failed, err: %w", err)
	}

	// nodeConf comes from db, which means that this node will not overwrite the original configuration in db.
	nodeConf, err := act.storageNodeDeployment.GetNodeConf(tenantCtx, param.Token)
	if err != nil {
		return fmt.Errorf("get node conf failed, err: %w", err)
	}

	gp := gopool.NewPool()
	gp.Go(func() error {
		if err := act.renderPreSetting(tenantCtx, nodeConf, &info.Host); err != nil {
			return fmt.Errorf("render pre setting failed, err: %w", err)
		}

		act.logger.Infof("rendered pre setting, token: %s", param.Token)

		return nil
	})

	gp.Go(func() error {
		if err := act.renderCustomSetting(nodeConf, info); err != nil {
			return fmt.Errorf("render custom setting failed, err: %w", err)
		}

		act.logger.Infof("rendered custom setting, token: %s", param.Token)

		return nil
	})

	if err := gp.Wait(); err != nil {
		return fmt.Errorf("failed to render node install config, err: %w", err)
	}

	if err := act.storageNodeDeployment.SetNodeConf(tenantCtx, param.Token, nodeConf); err != nil {
		return fmt.Errorf("set node conf failed, err: %w", err)
	}

	act.renderNodeDeploymentInfo(tenantCtx, info, nodeConf)

	if err := act.storageNodeDeployment.UpdateInfo(tenantCtx, param.Token, info); err != nil {
		return fmt.Errorf("set node deployment info failed, err: %w", err)
	}

	return nil
}

func (act *actionRenderNodeDeployment) renderPreSetting(ctx context.Context, nodeConf *types.NodeConf,
	host *types.Host) error {

	if err := act.renderDefaultSetting(nodeConf.PreSetting, host.Dynamic.NodeRole); err != nil {
		return fmt.Errorf("render default setting failed, err: %w", err)
	}

	if err := act.renderLogicSetting(ctx, nodeConf, host); err != nil {
		return fmt.Errorf("render logic setting failed, err: %w", err)
	}

	return nil
}

func (act *actionRenderNodeDeployment) renderDefaultSetting(preSetting map[string]any, nodeRole types.NodeRole,
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
// nolint: mnd,funlen,fnsize
// NOCC: golint/fnsize(func design is not suitable for splitting).
func GseProxySettingDefault() map[string]any {
	return map[string]any{
		"__BK_GSE_HOME_DIR__":                               "/usr/local/gse/proxy",
		"__BK_GSE_RUN_MODE__":                               "proxy",
		"__BK_GSE_ENABLE_STATIC_ACCESS__":                   false,
		"__BK_GSE_ENABLE_FAKE_SEED__":                       false,
		"__BK_GSE_ACCESS_CLUSTER_ENDPOINTS__":               "127.0.0.1:28668",
		"__BK_GSE_ACCESS_DATA_ENDPOINTS__":                  "127.0.0.1:28625",
		"__BK_GSE_ACCESS_FILE_ENDPOINTS__":                  "127.0.0.1:28925",
		"__BK_GSE_AGENT_BASE_TLS_CA_FILE__":                 "",
		"__BK_GSE_AGENT_BASE_TLS_CERT_FILE__":               "",
		"__BK_GSE_AGENT_BASE_TLS_KEY_FILE__":                "",
		"__BK_GSE_AGENT_BASE_TLS_PASSWORD_FILE__":           "",
		"__BK_GSE_AGENT_BASE_PROCESSOR_NUM__":               4,
		"__BK_GSE_AGENT_BASE_PROCESSOR_QUEUE_SIZE__":        4096,
		"__BK_GSE_AGENT_BASE_ALARM_EVENT_DATA_ID__":         1000,
		"__BK_GSE_AGENT_BASE_PLUGIN_IPC__":                  "${BK_GSE_HOME_DIR}/data/ipc.state.message",
		"__BK_GSE_PROXY_TLS_CA_FILE__":                      "",
		"__BK_GSE_PROXY_TLS_CERT_FILE__":                    "",
		"__BK_GSE_PROXY_TLS_KEY_FILE__":                     "",
		"__BK_GSE_PROXY_TLS_PASSWORD_FILE__":                "",
		"__BK_GSE_PROXY_BIND_IP__":                          "::",
		"__BK_GSE_PROXY_BIND_PORT__":                        28668,
		"__BK_GSE_PROXY_THREAD_NUM__":                       4,
		"__BK_GSE_TASK_PROC_EVENT_DATA_ID__":                1100008,
		"__BK_GSE_TASK_CONCURRENCE_COUNT__":                 100,
		"__BK_GSE_TASK_SCRIPT_FILE_EXPIRE_TIME_HOUR__":      72,
		"__BK_GSE_DATA_IPC__":                               "${BK_GSE_HOME_DIR}/data/ipc.state.report",
		"__BK_GSE_DATA_ENABLE_COMPRESSION__":                false,
		"__BK_GSE_FILE_MAX_TRANSFER_SPEED_MB_PER_SEC__":     100,
		"__BK_GSE_FILE_MAX_TRANSFER_CONCURRENT_NUM__":       10,
		"__BK_GSE_FILE_BT_LISTEN_INTERFACE__":               "",
		"__BK_GSE_FILE_BT_OUTGOING_INTERFACE__":             "",
		"__BK_GSE_FILE_BT_ENABLE_OUTGOING_INTERFACE__":      true,
		"__BK_GSE_EXTRA_CONFIG_DIRECTORY__":                 "",
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

	// GseTemplateKeyProxyBindPort the config template key of gse proxy bind port.
	GseTemplateKeyProxyBindPort = "__BK_GSE_PROXY_BIND_PORT__"
)

const (
	// GseCustomKeyFileTopologyLinks the config template key of gse file topology links.
	GseCustomKeyFileTopologyLinks = "file.topology.links"
)

// renderLogicSetting load logic setting to the config presetting and custom setting .
// nolint: nonamedreturns,funlen,fnsize
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (act *actionRenderNodeDeployment) renderLogicSetting(ctx context.Context, nodeConf *types.NodeConf,
	host *types.Host) (err error) {

	osType, err := platform.NormalizeOS(host.Static.OSType)
	if err != nil {
		return err
	}

	// this is a special case, when the deployment is reverted, the host id is not in the host table.
	if err := act.checkHostExist(ctx, host.HostID); err != nil {
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

	nodeConf.PreSetting[GseTemplateKeyRunMode] = host.Dynamic.NodeRole
	nodeConf.PreSetting[GseTemplateKeyCloudID] = host.Static.NetworkAreaID
	nodeConf.PreSetting[GseTemplateKeyZoneID] = host.Static.RegionID
	nodeConf.PreSetting[GseTemplateKeyCityID] = host.Static.CityID

	homeDir := joinPath(osType, deploymentConf.GseHomeDir, string(host.Dynamic.NodeRole))
	certDir := joinPath(osType, homeDir, "cert")
	nodeConf.PreSetting[GseTemplateKeyHomeDir] = homeDir

	gseCaFilePath := joinPath(osType, certDir, "gseca.crt")
	gsePasswordFilePath := joinPath(osType, certDir, "cert_encrypt.key")
	gseAgentCertFilePath := joinPath(osType, certDir, "gse_agent.crt")
	gseAgentKeyFilePath := joinPath(osType, certDir, "gse_agent.key")
	gseServerCertFilePath := joinPath(osType, certDir, "gse_server.crt")
	gseServerKeyFilePath := joinPath(osType, certDir, "gse_server.key")
	gseAPIClientCertFilePath := joinPath(osType, certDir, "gse_api_client.crt")
	gseAPIClientKeyFilePath := joinPath(osType, certDir, "gse_api_client.key")

	// base setting
	nodeConf.PreSetting[GseTemplateKeyAgentBaseTLSCAFile] = gseCaFilePath
	nodeConf.PreSetting[GseTemplateKeyAgentBaseTLSCertFile] = gseAgentCertFilePath
	nodeConf.PreSetting[GseTemplateKeyAgentBaseTLSKeyFile] = gseAgentKeyFilePath

	if system.GetEdition() == system.EditionEE {
		nodeConf.PreSetting[GseTemplateKeyAgentBaseTLSPasswordFile] = gsePasswordFilePath
	}

	if host.Dynamic.NodeRole == types.NodeRoleProxy {
		// proxy setting
		nodeConf.PreSetting[GseTemplateKeyProxyTLSCaFile] = gseCaFilePath
		nodeConf.PreSetting[GseTemplateKeyProxyTLSCertFile] = gseServerCertFilePath
		nodeConf.PreSetting[GseTemplateKeyProxyTLSKeyFile] = gseServerKeyFilePath

		if system.GetEdition() == system.EditionEE {
			nodeConf.PreSetting[GseTemplateKeyProxyTLSPasswordFile] = gsePasswordFilePath
		}

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

	nodeConf.PreSetting[GseTemplateKeyExtraConfigDirectory] = joinPath(osType, deploymentConf.GseEnvironDir, "user_conf")
	nodeConf.PreSetting[GseTemplateKeyLogPath] = deploymentConf.GseLogDir
	nodeConf.PreSetting[GseTemplateKeyAgentBasePluginIPC] = deploymentConf.GsePluginIPC
	nodeConf.PreSetting[GseTemplateKeyDataIPC] = deploymentConf.GseDataIPC
	nodeConf.PreSetting[GseTemplateKeyEnableStaticAccess], err = act.storageDomainGse.NeedStaticAccess(
		ctx, host.Dynamic.NetworkUnitID)

	if err != nil {
		return fmt.Errorf("check static access failed, err: %w", err)
	}

	// render access endpoints
	switch host.Dynamic.NodeRole {
	case types.NodeRoleAgent:
		{
			clusters, files, datas, err := act.storageDomainGse.GetV4AgentAccessEndpoints(ctx, host.Dynamic.NetworkUnitID)
			if err != nil {
				return fmt.Errorf("get agent access endpoints failed, err: %w", err)
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
			clusters, files, datas, err := act.storageDomainGse.GetProxyUpstreamAccessEndpoints(ctx, host.Dynamic.NetworkUnitID)
			if err != nil {
				return fmt.Errorf("get proxy upstream endpoints failed, err: %w", err)
			}

			nodeConf.PreSetting[GseTemplateKeyAccessClusterEndpoints] = strings.Join(clusters, ",")
			nodeConf.PreSetting[GseTemplateKeyAccessFileEndpoints] =
				fmt.Sprintf("%s:%v", advertiseIP, nodeConf.PreSetting[GseTemplateKeyFileAgentBindPort])
			nodeConf.PreSetting[GseTemplateKeyAccessDataEndpoints] =
				fmt.Sprintf("%s:%v", advertiseIP, nodeConf.PreSetting[GseTemplateKeyDataAgentBindPort])

			nodeConf.PreSetting[GseDataProxyEndpoints] = strings.Join(datas, ",")

			nodeConf.CustomSetting[GseCustomKeyFileTopologyLinks] = act.renderFileLinks(nodeConf, host, files)
		}
	default:
		return fmt.Errorf("unsupported node role: %s", host.Dynamic.NodeRole)
	}

	return nil
}

// joinPath joins path elements with the specified separator.
func joinPath(osType criteria.OSType, parts ...string) string {
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
func (act *actionRenderNodeDeployment) renderCustomSetting(conf *types.NodeConf, _ *types.DeploymentInfo) error {
	if conf.CustomSetting == nil {
		return errors.New("lack custom setting")
	}

	for _, key := range forbiddenKeys() {
		if _, ok := conf.CustomSetting[key]; ok {
			return fmt.Errorf("this key is forbidden, key(%s)", key)
		}
	}

	// TODO: Rendering strategy logic

	return nil
}
func (act *actionRenderNodeDeployment) checkHostExist(ctx context.Context, hostID int64) error {
	host, err := act.storageHost.GetHostByID(ctx, hostID)
	if err != nil {
		return fmt.Errorf("get host info failed, err: %w", err)
	}

	if host == nil {
		return fmt.Errorf("host not found, host-id(%d)", hostID)
	}

	return nil
}

// FileLink file link.
type FileLink struct {
	TargetIP   string `json:"target_ip,omitempty"`
	TargetPort int64  `json:"target_port,omitempty"`
	ReportIP   string `json:"report_ip,omitempty"`
	ReportPort int64  `json:"report_port,omitempty"`
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
				nodeConf.PreSetting[GseTemplateKeyFileAgentBindPort], defaultFileLinkTargetPort),
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

	info.Dynamic.ProxyClusterPort = conv.ToInt64Default(
		conf.PreSetting[GseTemplateKeyProxyBindPort], defaultKeyProxyBindPort)
	info.Dynamic.ProxyDataPort = conv.ToInt64Default(
		conf.PreSetting[GseTemplateKeyDataAgentBindPort], defaultKeyProxyDataPort)
	info.Dynamic.ProxyFilePort = conv.ToInt64Default(
		conf.PreSetting[GseTemplateKeyFileAgentBindPort], defaultKeyProxyFilePort)
}
