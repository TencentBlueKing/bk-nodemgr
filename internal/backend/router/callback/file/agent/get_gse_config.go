/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package agent ...
package agent

import (
	"bytes"
	"encoding/json"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
)

// GetGSEConfig ...
func (h handler) GetGSEConfig(ctx *rest.Context) (*rest.FileResponse, error) {
	//TODO: Mock 方案

	data := map[string]any{
		"run_mode": "agent",
		"cloud_id": 0,
		"zone_id":  "2",
		"city_id":  "30",
		"access": map[string]any{
			"enable_static_access": true,
			"enable_fake_seed":     false,
			"cluster_endpoints":    "9.134.151.50:28668,9.134.149.131:28668",
			"data_endpoints":       "9.134.151.50:28625,9.134.149.131:28625",
			"file_endpoints":       "9.134.151.50:28925,9.134.149.131:28925",
		},
		"base": map[string]any{
			"tls_ca_file":          "/usr/local/gse2_opbk/agent/cert/gseca.crt",
			"tls_cert_file":        "/usr/local/gse2_opbk/agent/cert/gse_agent.crt",
			"tls_key_file":         "/usr/local/gse2_opbk/agent/cert/gse_agent.key",
			"tls_passwd_file":      "/usr/local/gse2_opbk/agent/cert/cert_encrypt.key",
			"processor_num":        4,
			"processor_queue_size": 4096,
		},
		"proxy": map[string]any{
			"tls_ca_file":     "/usr/local/gse2_opbk/agent/cert/gseca.crt",
			"tls_cert_file":   "/usr/local/gse2_opbk/agent/cert/gse_server.crt",
			"tls_key_file":    "/usr/local/gse2_opbk/agent/cert/gse_server.key",
			"tls_passwd_file": "/usr/local/gse2_opbk/agent/cert/cert_encrypt.key",
			"bind_ip":         "::",
			"bind_port":       28668,
			"thread_num":      4,
		},
		"task": map[string]any{
			"proc_event_data_id":           1100008,
			"concurrence_count":            100,
			"script_file_expire_time_hour": 72,
		},
		"data": map[string]any{
			"ipc":                "/usr/local/gse2_opbk/agent/data/ipc.state.report",
			"enable_compression": true,
		},
		"file": map[string]any{
			"max_transfer_speed_mb_per_sec": 100,
			"max_transfer_concurrent_num":   10,
			"bt_listen_interface":           "",
			"bt_outgoing_interface":         "",
			"bt_enable_outgoing_interface":  true,
		},
		"logger": map[string]any{
			"path":              "/var/log/gse2_opbk",
			"level":             "INFO",
			"filesize_mb":       200,
			"filenum":           10,
			"rotate":            0,
			"flush_interval_ms": 1000,
		},
		"extra_config_directory": "/etc/sysconfig/gse/opbk/user_conf",
	}

	buffer := bytes.NewBuffer(nil)

	err := json.NewEncoder(buffer).Encode(data)
	if err != nil {
		return nil, err
	}

	resp := &rest.FileResponse{
		Data:        buffer,
		Size:        int64(buffer.Len()),
		FilePath:    "",
		FileName:    "gse_config.json",
		ContentType: "",
		Headers:     nil,
	}

	return resp, nil
}
