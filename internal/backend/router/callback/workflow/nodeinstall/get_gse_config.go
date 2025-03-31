/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package nodeinstall ...
package nodeinstall

import (
	"net/http"

	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/callback"
	"github.com/gin-gonic/gin"
)

// GetAgentConfig ...
func (h *handler) GetAgentConfig(gCtx *gin.Context) {
	req := new(protoBackend.GetAgentConfReq)
	if err := gCtx.BindJSON(req); err != nil {
		h.logger.Errorf("get gse agent config failed, err: %v", err)
		gCtx.JSON(http.StatusBadRequest, err)

		return
	}

	if err := req.Validate(); err != nil {
		h.logger.Errorf("get gse agent config failed, err: %v", err)
		gCtx.JSON(http.StatusBadRequest, err)

		return
	}

	nodeConf, err := h.nodeDeploymentStorage.GetNodeConf(gCtx, req.GetToken())
	if err != nil {
		h.logger.Errorf("get gse agent setting failed, err: %v", err)
		gCtx.JSON(http.StatusInternalServerError, err)

		return
	}

	conf, err := RenderConfig(DefaultTemplateGseAgent, nodeConf)
	if err != nil {
		h.logger.Errorf("render gse agent config failed, err: %v", err)
		gCtx.JSON(http.StatusInternalServerError, err)

		return
	}

	gCtx.String(http.StatusOK, conf)

	return
}

// DefaultTemplateGseAgent default template for gse agent.
const DefaultTemplateGseAgent = `{
    "run_mode": "__BK_GSE_RUN_MODE__",
    "cloud_id": __BK_GSE_CLOUD_ID__,
    "zone_id": "__BK_GSE_ZONE_ID__",
    "city_id": "__BK_GSE_CITY_ID__",
    "access": {
        "enable_static_access": __BK_GSE_ENABLE_STATIC_ACCESS__,
        "enable_fake_seed": __BK_GSE_ENABLE_FAKE_SEED__,
        "cluster_endpoints": "__BK_GSE_ACCESS_CLUSTER_ENDPOINTS__",
        "data_endpoints": "__BK_GSE_ACCESS_DATA_ENDPOINTS__",
        "file_endpoints": "__BK_GSE_ACCESS_FILE_ENDPOINTS__"
    },
    "base": {
        "tls_ca_file": "__BK_GSE_AGENT_BASE_TLS_CA_FILE__",
        "tls_cert_file": "__BK_GSE_AGENT_BASE_TLS_CERT_FILE__",
        "tls_key_file": "__BK_GSE_AGENT_BASE_TLS_KEY_FILE__",
        "tls_passwd_file": "__BK_GSE_AGENT_BASE_TLS_PASSWORD_FILE__",
        "processor_num": __BK_GSE_AGENT_BASE_PROCESSOR_NUM__,
        "processor_queue_size": __BK_GSE_AGENT_BASE_PROCESSOR_QUEUE_SIZE__,
        "alarm_event_data_id": __BK_GSE_AGENT_BASE_ALARM_EVENT_DATA_ID__,
        "plugin_ipc": "__BK_GSE_AGENT_BASE_PLUGIN_IPC__"
    },
    "proxy": {
        "tls_ca_file": "__BK_GSE_PROXY_TLS_CA_FILE__",
        "tls_cert_file": "__BK_GSE_PROXY_TLS_CERT_FILE__",
        "tls_key_file": "__BK_GSE_PROXY_TLS_KEY_FILE__",
        "tls_passwd_file": "__BK_GSE_PROXY_TLS_PASSWORD_FILE__",
        "bind_ip": "__BK_GSE_PROXY_BIND_IP__",
        "bind_port": __BK_GSE_PROXY_BIND_PORT__,
        "thread_num": __BK_GSE_PROXY_THREAD_NUM__
    },
    "task": {
        "proc_event_data_id": __BK_GSE_TASK_PROC_EVENT_DATA_ID__,
        "concurrence_count": __BK_GSE_TASK_CONCURRENCE_COUNT__,
        "script_file_expire_time_hour": __BK_GSE_TASK_SCRIPT_FILE_EXPIRE_TIME_HOUR__
    },
    "data": {
        "ipc": "__BK_GSE_DATA_IPC__",
        "enable_compression": __BK_GSE_DATA_ENABLE_COMPRESSION__
    },
    "file": {
        "max_transfer_speed_mb_per_sec": __BK_GSE_FILE_MAX_TRANSFER_SPEED_MB_PER_SEC__,
        "max_transfer_concurrent_num": __BK_GSE_FILE_MAX_TRANSFER_CONCURRENT_NUM__,
        "bt_listen_interface": "__BK_GSE_FILE_BT_LISTEN_INTERFACE__",
        "bt_outgoing_interface": "__BK_GSE_FILE_BT_OUTGOING_INTERFACE__",
        "bt_enable_outgoing_interface": __BK_GSE_FILE_BT_ENABLE_OUTGOING_INTERFACE__
    },
    "logger": {
        "path": "__BK_GSE_LOG_PATH__",
        "level": "__BK_GSE_LOG_LEVEL__",
        "filesize_mb": __BK_GSE_LOG_FILESIZE_MB__,
        "filenum": __BK_GSE_LOG_FILENUM__,
        "rotate": __BK_GSE_LOG_ROTATE__,
        "flush_interval_ms": __BK_GSE_LOG_FLUSH_INTERVAL_MS__
    },
    "extra_config_directory": "__BK_GSE_EXTRA_CONFIG_DIRECTORY__"
}`
