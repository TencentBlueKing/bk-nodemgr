/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package templates

import (
	"os"
	"strings"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/renderer/gotemplate"
)

// TestRelayConfigTemplate_Render tests the relay config template rendering.
// NOCC: golint/fnsize(template rendering fixture and expected output belong together).
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
				if !strings.Contains(result, "tracing:") {
					t.Error("Tracing section should be present")
				}
				if strings.Contains(result, "exporterType: otlp") {
					t.Error("Tracing exporterType should be ignored")
				}
				if strings.Contains(result, "otlpEndpoint: http://localhost:4317") {
					t.Error("Tracing otlpEndpoint should be ignored")
				}
				if !strings.Contains(result, "traceServiceName: relay") {
					t.Error("Tracing globalService traceServiceName should default to relay")
				}
			},
		},
		{
			name: "Workspace path is isolated per deploy env on linux",
			data: map[string]any{
				"PluginInfo": map[string]any{
					"Name":      "bk-nodemgr-relay",
					"PidPath":   "/var/run",
					"LogPath":   "/var/log",
					"PluginIPC": "/tmp/ipc.sock",
					"DeployEnv": "gse2_opbk",
				},
				"NodeInfo": map[string]any{
					"Dynamic": map[string]any{
						"NodeOsType": "linux",
					},
				},
				"CustomContext": map[string]any{},
			},
			validate: func(t *testing.T, result string) {
				if !strings.Contains(result, "fullPath: /tmp/bknm/gse2_opbk/relay") {
					t.Error("Workspace path should carry the deploy env so relays of different envs stay isolated")
				}
			},
		},
		{
			name: "Workspace path is isolated per deploy env on windows",
			data: map[string]any{
				"PluginInfo": map[string]any{
					"Name":      "bk-nodemgr-relay",
					"PidPath":   `c:\var\run`,
					"LogPath":   `c:\var\log`,
					"PluginIPC": "26000",
					"DeployEnv": "gse2_opbk",
				},
				"NodeInfo": map[string]any{
					"Dynamic": map[string]any{
						"NodeOsType": "windows",
					},
				},
				"CustomContext": map[string]any{},
			},
			validate: func(t *testing.T, result string) {
				if !strings.Contains(result, "fullPath: c:/tmp/bknm/gse2_opbk/relay") {
					t.Error("Windows workspace path should carry the deploy env")
				}
			},
		},
		{
			name: "Explicit workspace path overrides the per env default",
			data: map[string]any{
				"PluginInfo": map[string]any{
					"Name":      "bk-nodemgr-relay",
					"PidPath":   "/var/run",
					"LogPath":   "/var/log",
					"PluginIPC": "/tmp/ipc.sock",
					"DeployEnv": "gse2_opbk",
				},
				"NodeInfo": map[string]any{
					"Dynamic": map[string]any{
						"NodeOsType": "linux",
					},
				},
				"CustomContext": map[string]any{
					"WorkspaceFileGroupFullPath": "/data/bknm/relay",
				},
			},
			validate: func(t *testing.T, result string) {
				if !strings.Contains(result, "fullPath: /data/bknm/relay") {
					t.Error("An explicitly configured workspace path must win over the default")
				}
				if strings.Contains(result, "/tmp/bknm/gse2_opbk/relay") {
					t.Error("The per env default must not be emitted when overridden")
				}
			},
		},
		{
			name: "Workspace path falls back to dev when deploy env is absent",
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
				// Never render an empty segment: /tmp/bknm//relay would silently collapse the
				// isolation this default exists to provide.
				if !strings.Contains(result, "fullPath: /tmp/bknm/dev/relay") {
					t.Error("A missing deploy env should fall back to dev, not produce an empty path segment")
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
				if !strings.Contains(result, "tracing:") {
					t.Error("Tracing section should be present when not configured")
				}
				if !strings.Contains(result, "traceSampleRate: 0") {
					t.Error("Tracing globalService traceSampleRate should default to 0")
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
