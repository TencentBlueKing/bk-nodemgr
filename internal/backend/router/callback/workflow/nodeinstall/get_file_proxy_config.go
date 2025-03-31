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

// GetFileProxyConfig get file proxy config.
func (h *handler) GetFileProxyConfig(gCtx *gin.Context) {
	req := new(protoBackend.GetFileProxyConfReq)
	if err := gCtx.BindJSON(req); err != nil {
		h.logger.Errorf("get gse file proxy config failed, err: %s", err)
		gCtx.JSON(http.StatusBadRequest, err)

		return
	}

	if err := req.Validate(); err != nil {
		h.logger.Errorf("get gse file proxy config failed, err: %s", err)
		gCtx.JSON(http.StatusBadRequest, err)

		return
	}

	nodeConf, err := h.nodeDeploymentStorage.GetNodeConf(gCtx, req.GetToken())
	if err != nil {
		h.logger.Errorf("get gse file proxy setting failed, err: %s", err)
		gCtx.JSON(http.StatusInternalServerError, err)

		return
	}

	conf, err := RenderConfig(DefaultTemplateGseFileProxy, nodeConf)
	if err != nil {
		h.logger.Errorf("render gse file proxy config failed, err: %s", err)
		gCtx.JSON(http.StatusInternalServerError, err)

		return
	}

	gCtx.String(http.StatusOK, conf)

	return
}

// DefaultTemplateGseFileProxy is the default template of gse file proxy.
const DefaultTemplateGseFileProxy = `{
    "run_mode": "proxy",
    "cloud_id": __BK_GSE_CLOUD_ID__,
    "zone_id": "__BK_GSE_ZONE_ID__",
    "city_id": "__BK_GSE_CITY_ID__",
    "agent": {
        "bind_ip": "__BK_GSE_FILE_AGENT_BIND_IP__",
        "bind_port": __BK_GSE_FILE_AGENT_BIND_PORT__,
        "bind_port_v1": __BK_GSE_FILE_AGENT_BIND_PORT_V1__,
        "advertise_ipv4": "__BK_GSE_FILE_AGENT_ADVERTISE_IPV4__",
        "advertise_ipv6": "__BK_GSE_FILE_AGENT_ADVERTISE_IPV6__",
        "thread_num": __BK_GSE_FILE_AGENT_THREAD_NUM__,
        "tls_ca_file": "__BK_GSE_FILE_AGENT_TLS_CA_FILE__",
        "tls_passwd_file": "__BK_GSE_FILE_AGENT_TLS_PASSWORD_FILE__",
        "tls_cert_file": "__BK_GSE_FILE_AGENT_TLS_CERT_FILE__",
        "tls_key_file": "__BK_GSE_FILE_AGENT_TLS_KEY_FILE__"
    },
    "bittorrent": {
        "bind_ip": "__BK_GSE_FILE_BITTORRENT_BIND_IP__",
        "bind_port": __BK_GSE_FILE_BITTORRENT_BIND_PORT__,
        "tracker_bind_port": __BK_GSE_FILE_BITTORRENT_TRACKER_BIND_PORT__,
        "speed_limit_mb_per_sec": __BK_GSE_FILE_BITTORRENT_SPEED_LIMIT_MB_PER_SEC__
    },
    "topology": {
        "bind_ip": "__BK_GSE_FILE_TOPOLOGY_BIND_IP__",
        "bind_port": __BK_GSE_FILE_TOPOLOGY_BIND_PORT__,
        "thrift_bind_port": __BK_GSE_FILE_TOPOLOGY_THRIFT_BIND_PORT__,
        "advertise_ip": "__BK_GSE_FILE_TOPOLOGY_ADVERTISE_IP__",
        "thread_num": __BK_GSE_FILE_TOPOLOGY_THREAD_NUM__,
        "tls_ca_file": "__BK_GSE_FILE_TOPOLOGY_TLS_CA_FILE__",
        "tls_passwd_file": "__BK_GSE_FILE_TOPOLOGY_TLS_PASSWORD_FILE__",
        "tls_svr_cert_file": "__BK_GSE_FILE_TOPOLOGY_TLS_SVR_CERT_FILE__",
        "tls_svr_key_file": "__BK_GSE_FILE_TOPOLOGY_TLS_SVR_KEY_FILE__",
        "tls_cli_cert_file": "__BK_GSE_FILE_TOPOLOGY_TLS_CLI_CERT_FILE__",
        "tls_cli_key_file": "__BK_GSE_FILE_TOPOLOGY_TLS_CLI_KEY_FILE__",
        "links": [
            {
                "target_ip": "__BK_GSE_FILE_PROXY_UPSTREAM_IP__",
                "target_port": __BK_GSE_FILE_PROXY_UPSTREAM_PORT__,
                "report_ip": "__BK_GSE_FILE_PROXY_REPORT_IP__",
                "report_port": __BK_GSE_FILE_PROXY_REPORT_PORT__
            }
        ]
    },
    "cache": {
        "dirs": "__BK_GSE_FILE_CACHE_DIRS__",
        "expired_time_sec": __BK_GSE_FILE_CACHE_EXPIRED_TIME_SEC__
    },
    "metric": {
        "exporter_bind_ip": "__BK_GSE_FILE_METRIC_EXPORTER_BIND_IP__",
        "exporter_bind_port": __BK_GSE_FILE_METRIC_EXPORTER_BIND_PORT__,
        "exporter_thread_num": __BK_GSE_FILE_METRIC_EXPORTER_THREAD_NUM__
    },
    "logger": {
        "path": "__BK_GSE_LOG_PATH__",
        "level": "__BK_GSE_LOG_LEVEL__",
        "filesize_mb": __BK_GSE_LOG_FILESIZE_MB__,
        "filenum": __BK_GSE_LOG_FILENUM__,
        "rotate": __BK_GSE_LOG_ROTATE__,
        "flush_interval_ms": __BK_GSE_LOG_FLUSH_INTERVAL_MS__
    }
}`
