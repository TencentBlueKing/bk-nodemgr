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

	proto "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/callback"
	"github.com/gin-gonic/gin"
)

// GetDataProxyConfig get gse data proxy config.
func (h *handler) GetDataProxyConfig(gCtx *gin.Context) {
	req := new(proto.GetDataProxyConfReq)
	if err := gCtx.BindJSON(req); err != nil {
		h.logger.Errorf("get gse data proxy config failed, err: %s", err)
		gCtx.JSON(http.StatusBadRequest, err)

		return
	}

	if err := req.Validate(); err != nil {
		h.logger.Errorf("get gse data proxy config failed, err: %s", err)
		gCtx.JSON(http.StatusBadRequest, err)

		return
	}

	preSetting, customSetting, err := h.nodeDeploymentStorage.GetGseDataProxySetting(gCtx, req.GetToken())
	if err != nil {
		h.logger.Errorf("get gse data proxy setting failed, err: %s", err)
		gCtx.JSON(http.StatusInternalServerError, err)

		return
	}

	conf, err := RenderConfig(DefaultTemplateGseDataProxy, preSetting, customSetting)
	if err != nil {
		h.logger.Errorf("render gse data proxy config failed, err: %s", err)
		gCtx.JSON(http.StatusInternalServerError, err.Error())

		return
	}

	gCtx.String(http.StatusOK, conf)

	return
}

// DefaultTemplateGseDataProxy  gse data agent.
const DefaultTemplateGseDataProxy = `{
    "run_mode": "proxy",
    "cloud_id": __BK_GSE_CLOUD_ID__,
    "zone_id": "__BK_GSE_ZONE_ID__",
    "city_id": "__BK_GSE_CITY_ID__",
    "agent": {
        "tls_ca_file": "__BK_GSE_DATA_AGENT_TLS_CA_FILE__",
        "tls_cert_file": "__BK_GSE_DATA_AGENT_TLS_CERT_FILE__",
        "tls_key_file": "__BK_GSE_DATA_AGENT_TLS_KEY_FILE__",
        "tls_passwd_file": "__BK_GSE_DATA_AGENT_TLS_PASSWORD_FILE__",
        "tls_proxy_cert_file": "__BK_GSE_DATA_PROXY_TLS_CERT_FILE__",
        "tls_proxy_key_file": "__BK_GSE_DATA_PROXY_TLS_KEY_FILE__",
        "bind_ip": "__BK_GSE_DATA_AGENT_BIND_IP__",
        "bind_port": __BK_GSE_DATA_AGENT_BIND_PORT__,
        "thread_num": __BK_GSE_DATA_AGENT_THREAD_NUM__,
        "proxy_endpoints": "__BK_GSE_DATA_PROXY_ENDPOINTS__"
    },
    "metric":{
        "exporter_bind_ip": "__BK_GSE_DATA_METRIC_EXPORTER_BIND_IP__",
        "exporter_bind_port": __BK_GSE_DATA_METRIC_EXPORTER_BIND_PORT__,
        "exporter_thread_num": __BK_GSE_DATA_METRIC_EXPORTER_THREAD_NUM__
    },
    "logger":{
        "path": "__BK_GSE_LOG_PATH__",
        "level": "__BK_GSE_LOG_LEVEL__",
        "filesize_mb": __BK_GSE_LOG_FILESIZE_MB__,
        "filenum": __BK_GSE_LOG_FILENUM__,
        "rotate": __BK_GSE_LOG_ROTATE__,
        "flush_interval_ms": __BK_GSE_LOG_FLUSH_INTERVAL_MS__
    }
}`
