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
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// TestRenderConfig ...
// NOCC: golint/fnsize.
func TestRenderConfig(t *testing.T) {
	type args struct {
		template Template
		nodeConf types.NodeConf
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				template: Template{
					UniqueKey: UniqueKeyFile,
					Content: `{
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
}`,
				},
				nodeConf: types.NodeConf{
					PreSetting: map[string]any{
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
						"__BK_GSE_LOG_PATH__":                               "__BK_GSE_HOME_DIR__/logs",
						"__BK_GSE_LOG_LEVEL__":                              "INFO",
						"__BK_GSE_LOG_FILESIZE_MB__":                        200,
						"__BK_GSE_LOG_FILENUM__":                            10,
						"__BK_GSE_LOG_ROTATE__":                             0,
						"__BK_GSE_LOG_FLUSH_INTERVAL_MS__":                  100,
					},
					CustomSetting: map[string]any{
						"file.topology.links": "127.0.0.1",
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := RenderConfig(tt.args.template, &tt.args.nodeConf)
			if err != nil {
				t.Logf("RenderConfig() error = %v", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("RenderConfig() error = %v, wantErr %v", err, tt.wantErr)
			}

			t.Logf("RenderConfig() config: \n%v\n", config)
		})
	}
}
