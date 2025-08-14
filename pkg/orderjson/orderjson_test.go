/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package orderjson

import (
	"encoding/json"
	"testing"
)

// Test_Ordered tests the ordered result.
func Test_Ordered(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantErr bool
	}{
		{
			name:    "test-1",
			raw:     `{"a":{"c":123,"d":true},"b":[0,1,2,3,4,5],"c":false}`,
			wantErr: false,
		},
		{
			name:    "test-2",
			raw:     `{"run_mode":"agent","cloud_id":0,"zone_id":"test","city_id":"test","access":{"enable_static_access":false,"enable_fake_seed":false,"cluster_endpoints":"","data_endpoints":"","file_endpoints":""},"base":{"tls_ca_file":"/usr/local/gse/agent/cert/gseca.crt","tls_cert_file":"/usr/local/gse/agent/cert/gse_agent.crt","tls_key_file":"/usr/local/gse/agent/cert/gse_agent.key","tls_passwd_file":"/usr/local/gse/agent/cert/cert_encrypt.key","processor_num":4,"processor_queue_size":4096,"alarm_event_data_id":0,"plugin_ipc":"/usr/local/gse/agent/lib/ipc.state.message"},"proxy":{"tls_ca_file":"/usr/local/gse/agent/cert/gseca.crt","tls_cert_file":"/usr/local/gse/agent/cert/gse_server.crt","tls_key_file":"/usr/local/gse/agent/cert/gse_server.key","tls_passwd_file":"/usr/local/gse/agent/cert/cert_encrypt.key","bind_ip":"::","bind_port":28668,"thread_num":4},"task":{"proc_event_data_id":0,"concurrence_count":100,"script_file_expire_time_hour":72},"data":{"ipc":"/usr/local/gse/agent/data/ipc.state.report","enable_compression":false},"file":{"max_transfer_speed_mb_per_sec":100,"max_transfer_concurrent_num":10,"bt_listen_interface":"","bt_outgoing_interface":"","bt_enable_outgoing_interface":true},"logger":{"path":"/var/log/gse","level":"INFO","filesize_mb":200,"filenum":10,"rotate":0,"flush_interval_ms":1000},"extra_config_directory":""}`,
			wantErr: false,
		},
		{
			name:    "test-3",
			raw:     `[1,2,3,{"hello":true,"world":[{"a":1},2,false,""]}]`,
			wantErr: false,
		},
		{
			name:    "test-4",
			raw:     `{"yes":{"\"\"\"{}}}}{}{}]][[[{}{{}}{":"\"\"[][]{}{}\"\"","\"][]{}{}{}{}{}\"\\\\\\\\\\":123}}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var data OrderedData
			err := json.Unmarshal([]byte(tt.raw), &data)
			if (err != nil) != tt.wantErr {
				t.Errorf("json.Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			got, err := json.Marshal(data)
			if err != nil {
				t.Errorf("json.Marshal() error = %v", err)
				return
			}

			if string(got) != tt.raw {
				t.Errorf("unmarshal and marshal got = %s, want %s", got, tt.raw)
			}
		})
	}
}
