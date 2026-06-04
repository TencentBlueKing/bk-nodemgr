/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package templates

import (
	"os"
	"strings"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/renderer/gotemplate"
)

// TestRelayConfigTemplate_Render tests the relay config template rendering.
func TestRelayConfigTemplate_Render(t *testing.T) {
	// Read the template file
	templateContent, err := os.ReadFile("bk-nodemgr-relay.conf.template")
	if err != nil {
		t.Fatalf("Failed to read template file: %v", err)
	}

	renderer := gotemplate.New()

	tests := []struct {
		name     string
		data     map[string]any
		validate func(t *testing.T, result string)
	}{
		{
			name: "Use reduced relay server configuration",
			data: map[string]any{
				"PluginInfo": map[string]any{
					"Name":      "bk-nodemgr-relay",
					"PidPath":   "/var/run",
					"LogPath":   "/var/log",
					"PluginIPC": "/tmp/ipc.sock",
				},
				"NodeInfo": map[string]any{
					"Static": map[string]any{
						"InnerIPList": []string{"10.0.0.1", "10.0.0.2"},
					},
					"Dynamic": map[string]any{
						"NodeOsType":        "linux",
						"AdvertiseIP":       "192.168.1.100",
						"AdvertiseIPV6":     "2001:db8::1",
						"RelayCallbackPort": 29000,
						"RelayDownloadPort": 29001,
					},
				},
				"CustomContext": map[string]any{
					"InfoServer": map[string]any{
						"BindIP":       "0.0.0.0",
						"Port":         28000,
						"AuthIdentity": "test-auth",
					},
					"AdminServer": map[string]any{
						"BindIP":       "127.0.0.1",
						"Port":         28001,
						"AuthIdentity": "admin-auth",
					},
					"CallbackServer": map[string]any{
						"BindIP":       "192.168.1.111",
						"Port":         28002,
						"AuthIdentity": "callback-auth",
					},
					"DownloadServer": map[string]any{
						"BindIP":       "192.168.1.112",
						"Port":         28003,
						"AuthIdentity": "download-auth",
					},
					"WorkspaceFileGroupFullPath": "/data/bknm",
					"Log": map[string]any{
						"LogDir":          "/var/log/relay",
						"LogMaxSizeMB":    200,
						"LogMaxNum":       10,
						"LogLevel":        "DEBUG",
						"LogToStderr":     true,
						"LogAlsoToStderr": true,
					},
				},
			},
			validate: func(t *testing.T, result string) {
				// Verify all servers use the first static inner IP as bindIP.
				if strings.Count(result, "bindIP: 10.0.0.1") != 4 {
					t.Error("all bindIP values should use the first NodeInfo.Static.InnerIPList item")
				}
				if !strings.Contains(result, "port: 29000") {
					t.Error("CallbackServer port should use NodeInfo.Dynamic.RelayCallbackPort")
				}
				if !strings.Contains(result, "port: 29001") {
					t.Error("DownloadServer port should use NodeInfo.Dynamic.RelayDownloadPort")
				}
				if strings.Contains(result, "192.168.1.111") || strings.Contains(result, "192.168.1.112") {
					t.Error("CallbackServer and DownloadServer bindIP should ignore CustomContext values")
				}
				if strings.Contains(result, "port: 28002") || strings.Contains(result, "port: 28003") {
					t.Error("CallbackServer and DownloadServer port should ignore CustomContext values")
				}
			},
		},
		{
			name: "CustomContext is empty",
			data: map[string]any{
				"PluginInfo": map[string]any{
					"Name":      "bk-nodemgr-relay",
					"PidPath":   "/var/run",
					"LogPath":   "/var/log",
					"PluginIPC": "/tmp/ipc.sock",
				},
				"NodeInfo": map[string]any{
					"Static": map[string]any{
						"InnerIPList": []string{"10.0.0.1", "10.0.0.2"},
					},
					"Dynamic": map[string]any{
						"NodeOsType":        "linux",
						"AdvertiseIP":       "192.168.1.100",
						"AdvertiseIPV6":     "2001:db8::1",
						"RelayCallbackPort": 29000,
						"RelayDownloadPort": 29001,
					},
				},
				"CustomContext": map[string]any{
					"CallbackServer": map[string]any{
						// BindIP and Port are both empty
					},
					"DownloadServer": map[string]any{
						// BindIP and Port are both empty
					},
				},
			},
			validate: func(t *testing.T, result string) {
				// Verify all servers use the first static inner IP as bindIP.
				if strings.Count(result, "bindIP: 10.0.0.1") != 4 {
					t.Error("all bindIP values should use the first NodeInfo.Static.InnerIPList item")
				}
				if !strings.Contains(result, "port: 29000") {
					t.Error("CallbackServer port should fallback to NodeInfo.Dynamic.RelayCallbackPort")
				}
				if !strings.Contains(result, "port: 29001") {
					t.Error("DownloadServer port should fallback to NodeInfo.Dynamic.RelayDownloadPort")
				}
			},
		},
		{
			name: "Use the default value when both are empty",
			data: map[string]any{
				"PluginInfo": map[string]any{
					"Name":      "bk-nodemgr-relay",
					"PidPath":   "/var/run",
					"LogPath":   "/var/log",
					"PluginIPC": "/tmp/ipc.sock",
				},
				"NodeInfo": map[string]any{
					"Dynamic": map[string]any{
						"NodeOsType": "linux",
						// AdvertiseIP and ports are both empty
					},
				},
				"CustomContext": map[string]any{
					"CallbackServer": map[string]any{},
					"DownloadServer": map[string]any{},
				},
			},
			validate: func(t *testing.T, result string) {
				// Verify default values are used
				if !strings.Contains(result, "bindIP: 127.0.0.1") {
					t.Error("CallbackServer bindIP should use default 127.0.0.1")
				}
				if !strings.Contains(result, "port: 28302") {
					t.Error("CallbackServer port should use default 28302")
				}
				if !strings.Contains(result, "bindIP: 127.0.0.1") {
					t.Error("DownloadServer bindIP should use default 127.0.0.1")
				}
				if !strings.Contains(result, "port: 28303") {
					t.Error("DownloadServer port should use default 28303")
				}
			},
		},
		{
			name: "Windows systems use messageLocalSocketPort",
			data: map[string]any{
				"PluginInfo": map[string]any{
					"Name":      "bk-nodemgr-relay",
					"PidPath":   "C:\\ProgramData",
					"LogPath":   "C:\\Logs",
					"PluginIPC": "26000",
				},
				"NodeInfo": map[string]any{
					"Dynamic": map[string]any{
						"NodeOsType": "windows",
					},
				},
				"CustomContext": map[string]any{},
			},
			validate: func(t *testing.T, result string) {
				// Verify Windows uses messageLocalSocketPort
				if !strings.Contains(result, "messageLocalSocketPort:") {
					t.Error("Windows should use messageLocalSocketPort")
				}
				if strings.Contains(result, "messageDomainSocketPath:") {
					t.Error("Windows should not use messageDomainSocketPath")
				}
			},
		},
		{
			name: "Linux systems use messageDomainSocketPath",
			data: map[string]any{
				"PluginInfo": map[string]any{
					"Name":      "bk-nodemgr-relay",
					"PidPath":   "/var/run",
					"LogPath":   "/var/log",
					"PluginIPC": "/tmp/ipc.sock",
				},
				"NodeInfo": map[string]any{
					"Dynamic": map[string]any{
						"NodeOsType": "linux",
					},
				},
				"CustomContext": map[string]any{},
			},
			validate: func(t *testing.T, result string) {
				// Verify Linux uses messageDomainSocketPath
				if !strings.Contains(result, "messageDomainSocketPath:") {
					t.Error("Linux should use messageDomainSocketPath")
				}
				if strings.Contains(result, "messageLocalSocketPort:") {
					t.Error("Linux should not use messageLocalSocketPort")
				}
			},
		},
		{
			name: "CustomContext Tracing config is ignored",
			data: map[string]any{
				"PluginInfo": map[string]any{
					"Name":      "bk-nodemgr-relay",
					"PidPath":   "/var/run",
					"LogPath":   "/var/log",
					"PluginIPC": "/tmp/ipc.sock",
				},
				"NodeInfo": map[string]any{
					"Dynamic": map[string]any{
						"NodeOsType": "linux",
					},
				},
				"CustomContext": map[string]any{
					"Tracing": map[string]any{
						"ExporterType": "otlp",
						"OtlpEndpoint": "http://localhost:4317",
						"OtlpInsecure": true,
						"OtlpHeaders": map[string]string{
							"Authorization": "Bearer token",
						},
					},
				},
			},
			validate: func(t *testing.T, result string) {
				// Verify Tracing configuration is not rendered after reducing exposed fields.
				if strings.Contains(result, "tracing:") {
					t.Error("Tracing section should not be present")
				}
				if strings.Contains(result, "exporterType: otlp") {
					t.Error("Tracing exporterType should be ignored")
				}
				if strings.Contains(result, "otlpEndpoint: http://localhost:4317") {
					t.Error("Tracing otlpEndpoint should be ignored")
				}
			},
		},
		{
			name: "Tracing config is not included",
			data: map[string]any{
				"PluginInfo": map[string]any{
					"Name":      "bk-nodemgr-relay",
					"PidPath":   "/var/run",
					"LogPath":   "/var/log",
					"PluginIPC": "/tmp/ipc.sock",
				},
				"NodeInfo": map[string]any{
					"Dynamic": map[string]any{
						"NodeOsType": "linux",
					},
				},
				"CustomContext": map[string]any{
					// No Tracing configuration
				},
			},
			validate: func(t *testing.T, result string) {
				// Verify Tracing configuration is not present
				if strings.Contains(result, "tracing:") {
					t.Error("Tracing section should not be present when not configured")
				}
			},
		},
		{
			name: "Both AdvertiseIP and AdvertiseIPV6 are present",
			data: map[string]any{
				"PluginInfo": map[string]any{
					"Name":      "bk-nodemgr-relay",
					"PidPath":   "/var/run",
					"LogPath":   "/var/log",
					"PluginIPC": "/tmp/ipc.sock",
				},
				"NodeInfo": map[string]any{
					"Dynamic": map[string]any{
						"NodeOsType":    "linux",
						"AdvertiseIP":   "192.168.1.100",
						"AdvertiseIPV6": "2001:db8::1",
					},
				},
				"CustomContext": map[string]any{},
			},
			validate: func(t *testing.T, result string) {
				// Verify both advertiseIPV4 and advertiseIPV6 are present
				if !strings.Contains(result, "advertiseIPV4: 192.168.1.100") {
					t.Error("advertiseIPV4 should be present")
				}
				if !strings.Contains(result, "advertiseIPV6: 2001:db8::1") {
					t.Error("advertiseIPV6 should be present")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := renderer.Render(string(templateContent), tt.data)
			if err != nil {
				t.Fatalf("Render() error = %v", err)
			}

			t.Logf("Rendered result:\n%s", result)

			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}
