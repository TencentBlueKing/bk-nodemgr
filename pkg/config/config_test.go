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

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplicationService_LoadFromFileReadsHelmRenderedKeys(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "application_conf.yaml")
	configContent := []byte(`runMode: debug
tenantMode: single
bkPaaS:
  analysisScript: "<script>window.bkAnalytics=true</script>"
access:
  virtualUser: "bk-nodemgr-app"
`)
	require.NoError(t, os.WriteFile(configPath, configContent, 0o600))

	svc := NewApplicationService()
	require.NoError(t, svc.LoadFromFile(configPath))

	assert.Equal(t, RunModeDebug, svc.RunMode)
	assert.Equal(t, "<script>window.bkAnalytics=true</script>", svc.BKPaas.AnalysisScript)
	assert.Equal(t, "bk-nodemgr-app", svc.Access.VirtualUser)
}

func TestApplicationService_LoadFromFileKeepsLegacyModeKeyCompatible(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "application_conf.yaml")
	configContent := []byte(`mode: debug
tenantMode: single
`)
	require.NoError(t, os.WriteFile(configPath, configContent, 0o600))

	svc := NewApplicationService()
	require.NoError(t, svc.LoadFromFile(configPath))

	assert.Equal(t, RunModeDebug, svc.RunMode)
}

func TestApplicationService_LoadFromFilePrefersRunModeOverLegacyMode(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "application_conf.yaml")
	configContent := []byte(`runMode: release
mode: debug
tenantMode: single
`)
	require.NoError(t, os.WriteFile(configPath, configContent, 0o600))

	svc := NewApplicationService()
	require.NoError(t, svc.LoadFromFile(configPath))

	assert.Equal(t, RunModeRelease, svc.RunMode)
}

func TestBKLogin_ValidateRequiresBackendEndpoints(t *testing.T) {
	conf := BKLogin{
		LoginURL: "https://login.example.com",
		AuthType: LoginAuthTypeBKToken,
	}

	err := conf.Validate()

	require.ErrorContains(t, err, "endpoints of bkLogin is empty")
}

func TestTracing_ValidateTraceSampleRate(t *testing.T) {
	tests := []struct {
		name            string
		traceSampleRate float64
		wantErr         bool
	}{
		{
			name:            "zero sample rate",
			traceSampleRate: 0,
		},
		{
			name:            "full sample rate",
			traceSampleRate: 1,
		},
		{
			name:            "negative sample rate",
			traceSampleRate: -0.1,
			wantErr:         true,
		},
		{
			name:            "sample rate greater than one",
			traceSampleRate: 1.1,
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := Tracing{
				ExporterType: "stdout",
				GlobalService: TraceService{
					TraceSampleRate: tt.traceSampleRate,
				},
			}

			err := conf.Validate()
			if tt.wantErr {
				assert.ErrorContains(t, err, "failed to validate global service trace config")
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestTracing_ValidateOTLPProtocol(t *testing.T) {
	tests := []struct {
		name         string
		otlpProtocol string
		wantErr      bool
	}{
		{
			name: "empty protocol",
		},
		{
			name:         "grpc protocol",
			otlpProtocol: "grpc",
		},
		{
			name:         "http protocol",
			otlpProtocol: "http",
		},
		{
			name:         "invalid protocol",
			otlpProtocol: "invalid",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := Tracing{
				ExporterType:  "otlp",
				OTLPProtocol:  tt.otlpProtocol,
				GlobalService: TraceService{},
			}

			err := conf.Validate()
			if tt.wantErr {
				assert.ErrorContains(t, err, "otlp protocol is invalid")
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestRedisType_Validate(t *testing.T) {
	tests := []struct {
		name    string
		t       RedisType
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid standalone type",
			t:       RedisTypeStandalone,
			wantErr: false,
		},
		{
			name:    "valid sentinel type",
			t:       RedisTypeSentinel,
			wantErr: false,
		},
		{
			name:    "valid cluster type",
			t:       RedisTypeCluster,
			wantErr: false,
		},
		{
			name:    "empty type defaults to standalone",
			t:       "",
			wantErr: false,
		},
		{
			name:    "invalid type",
			t:       "invalid",
			wantErr: true,
			errMsg:  "invalid redis type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.t.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRedis_GetType(t *testing.T) {
	tests := []struct {
		name     string
		config   Redis
		expected RedisType
	}{
		{
			name:     "empty type returns standalone",
			config:   Redis{Type: ""},
			expected: RedisTypeStandalone,
		},
		{
			name:     "explicit standalone",
			config:   Redis{Type: RedisTypeStandalone},
			expected: RedisTypeStandalone,
		},
		{
			name:     "explicit sentinel",
			config:   Redis{Type: RedisTypeSentinel},
			expected: RedisTypeSentinel,
		},
		{
			name:     "explicit cluster",
			config:   Redis{Type: RedisTypeCluster},
			expected: RedisTypeCluster,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.config.GetType())
		})
	}
}

func TestRedis_Validate_StandaloneMode(t *testing.T) {
	tests := []struct {
		name    string
		config  Redis
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid standalone config",
			config: Redis{
				Type:     RedisTypeStandalone,
				Addrs:    []string{"127.0.0.1:6379"},
				Password: "password",
				DB:       0,
			},
			wantErr: false,
		},
		{
			name: "valid standalone config with empty type",
			config: Redis{
				Addrs:    []string{"127.0.0.1:6379"},
				Password: "password",
				DB:       0,
			},
			wantErr: false,
		},
		{
			name: "missing addrs in standalone mode",
			config: Redis{
				Type:     RedisTypeStandalone,
				Addrs:    []string{},
				Password: "password",
				DB:       0,
			},
			wantErr: true,
			errMsg:  "addrs is required",
		},
		{
			name: "missing password in standalone mode",
			config: Redis{
				Type:     RedisTypeStandalone,
				Addrs:    []string{"127.0.0.1:6379"},
				Password: "",
				DB:       0,
			},
			wantErr: true,
			errMsg:  "password of redis is empty",
		},
		{
			name: "negative db in standalone mode",
			config: Redis{
				Type:     RedisTypeStandalone,
				Addrs:    []string{"127.0.0.1:6379"},
				Password: "password",
				DB:       -1,
			},
			wantErr: true,
			errMsg:  "db of redis must be greater than or equal to 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRedis_Validate_ClusterMode(t *testing.T) {
	tests := []struct {
		name    string
		config  Redis
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid cluster config",
			config: Redis{
				Type:     RedisTypeCluster,
				Addrs:    []string{"127.0.0.1:6379", "127.0.0.2:6379"},
				Password: "password",
			},
			wantErr: false,
		},
		{
			name: "missing addrs in cluster mode",
			config: Redis{
				Type:     RedisTypeCluster,
				Addrs:    []string{},
				Password: "password",
			},
			wantErr: true,
			errMsg:  "addrs is required in cluster mode",
		},
		{
			name: "missing password in cluster mode",
			config: Redis{
				Type:     RedisTypeCluster,
				Addrs:    []string{"127.0.0.1:6379"},
				Password: "",
			},
			wantErr: true,
			errMsg:  "password of redis is empty",
		},
		{
			name: "cluster mode ignores db parameter",
			config: Redis{
				Type:     RedisTypeCluster,
				Addrs:    []string{"127.0.0.1:6379"},
				Password: "password",
				DB:       5, // Should be ignored in cluster mode
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRedis_Validate_SentinelMode(t *testing.T) {
	tests := []struct {
		name    string
		config  Redis
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid sentinel config",
			config: Redis{
				Type:       RedisTypeSentinel,
				Addrs:      []string{"sentinel1:26379", "sentinel2:26379", "sentinel3:26379"},
				Password:   "password",
				MasterName: "mymaster",
				DB:         0,
			},
			wantErr: false,
		},
		{
			name: "missing addrs in sentinel mode",
			config: Redis{
				Type:       RedisTypeSentinel,
				Addrs:      []string{},
				Password:   "password",
				MasterName: "mymaster",
			},
			wantErr: true,
			errMsg:  "addrs is required in sentinel mode",
		},
		{
			name: "missing master name in sentinel mode",
			config: Redis{
				Type:       RedisTypeSentinel,
				Addrs:      []string{"sentinel1:26379"},
				Password:   "password",
				MasterName: "",
			},
			wantErr: true,
			errMsg:  "masterName is required in sentinel mode",
		},
		{
			name: "missing password in sentinel mode",
			config: Redis{
				Type:       RedisTypeSentinel,
				Addrs:      []string{"sentinel1:26379"},
				Password:   "",
				MasterName: "mymaster",
			},
			wantErr: true,
			errMsg:  "password of redis is empty",
		},
		{
			name: "sentinel with db parameter",
			config: Redis{
				Type:       RedisTypeSentinel,
				Addrs:      []string{"sentinel1:26379"},
				Password:   "password",
				MasterName: "mymaster",
				DB:         1,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRedis_Validate_InvalidType(t *testing.T) {
	config := Redis{
		Type:     "invalid",
		Addrs:    []string{"127.0.0.1:6379"},
		Password: "password",
	}

	err := config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid redis type")
}

func TestDownloader_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Downloader
		wantErr bool
	}{
		{
			name:    "empty allow hosts",
			config:  Downloader{},
			wantErr: false,
		},
		{
			name:    "valid allow hosts",
			config:  Downloader{AllowHosts: []string{"github.com"}},
			wantErr: false,
		},
		{
			name:    "invalid allow hosts",
			config:  Downloader{AllowHosts: []string{""}},
			wantErr: true,
		},
		{
			name:    "invalid block hosts",
			config:  Downloader{BlockHosts: []string{" "}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
