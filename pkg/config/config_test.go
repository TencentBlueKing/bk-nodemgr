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
	"encoding/base64"
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
	assert.Equal(t, int64(-1), svc.Front.UnassignedNetworkAreaID)
}

func TestApplicationService_LoadFromFileReadsUnassignedNetworkAreaID(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "application_conf.yaml")
	configContent := []byte(`front:
  unassignedNetworkAreaID: 100
`)
	require.NoError(t, os.WriteFile(configPath, configContent, 0o600))

	svc := NewApplicationService()
	require.NoError(t, svc.LoadFromFile(configPath))

	assert.Equal(t, int64(100), svc.Front.UnassignedNetworkAreaID)
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
			name:    "empty type",
			t:       "",
			wantErr: true,
			errMsg:  "invalid redis type",
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
			name: "empty type is rejected",
			config: Redis{
				Addrs:    []string{"127.0.0.1:6379"},
				Password: "password",
				DB:       0,
			},
			wantErr: true,
			errMsg:  "invalid redis type",
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

func TestTLSConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  TLSConfig
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty config",
			config:  TLSConfig{},
			wantErr: false,
		},
		{
			name: "paired inline cert and key",
			config: TLSConfig{
				Cert: "cert-content",
				Key:  "key-content",
			},
			wantErr: false,
		},
		{
			name: "paired cert and key files",
			config: TLSConfig{
				CertFile: "/tmp/cert.pem",
				KeyFile:  "/tmp/key.pem",
			},
			wantErr: false,
		},
		{
			name: "cert without key",
			config: TLSConfig{
				Cert: "cert-content",
			},
			wantErr: true,
			errMsg:  "cert and key must be both provided or both empty",
		},
		{
			name: "key file without cert",
			config: TLSConfig{
				KeyFile: "/tmp/key.pem",
			},
			wantErr: true,
			errMsg:  "cert and key must be both provided or both empty",
		},
		{
			name: "password without key",
			config: TLSConfig{
				Password: "secret",
			},
			wantErr: true,
			errMsg:  "password provided but no key specified",
		},
		{
			name: "password with key",
			config: TLSConfig{
				Cert:     "cert-content",
				Key:      "key-content",
				Password: "secret",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.ErrorContains(t, err, tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTLSConfig_ToSSLConfigPrefersInlineOverFile(t *testing.T) {
	conf := TLSConfig{
		CA:       "inline-ca",
		Cert:     "inline-cert",
		Key:      "inline-key",
		CAFile:   "/nonexistent/ca.pem",
		CertFile: "/nonexistent/cert.pem",
		KeyFile:  "/nonexistent/key.pem",
	}

	sslConf, err := conf.ToSSLConfig()

	require.NoError(t, err)
	assert.Equal(t, []byte("inline-ca"), sslConf.CA)
	assert.Equal(t, []byte("inline-cert"), sslConf.Cert)
	assert.Equal(t, []byte("inline-key"), sslConf.Key)
}

func TestTLSConfig_ToSSLConfigReadsFromFiles(t *testing.T) {
	dir := t.TempDir()
	caFile := filepath.Join(dir, "ca.pem")
	certFile := filepath.Join(dir, "cert.pem")
	keyFile := filepath.Join(dir, "key.pem")
	require.NoError(t, os.WriteFile(caFile, []byte("file-ca"), 0o600))
	require.NoError(t, os.WriteFile(certFile, []byte("file-cert"), 0o600))
	require.NoError(t, os.WriteFile(keyFile, []byte("file-key"), 0o600))

	conf := TLSConfig{
		CAFile:   caFile,
		CertFile: certFile,
		KeyFile:  keyFile,
		Password: "secret",
	}

	sslConf, err := conf.ToSSLConfig()

	require.NoError(t, err)
	assert.Equal(t, []byte("file-ca"), sslConf.CA)
	assert.Equal(t, []byte("file-cert"), sslConf.Cert)
	assert.Equal(t, []byte("file-key"), sslConf.Key)
	assert.Equal(t, "secret", sslConf.Password)
}

func TestTLSConfig_ToSSLConfigFileErrors(t *testing.T) {
	dir := t.TempDir()
	caFile := filepath.Join(dir, "ca.pem")
	keyFile := filepath.Join(dir, "key.pem")
	require.NoError(t, os.WriteFile(caFile, []byte("file-ca"), 0o600))
	require.NoError(t, os.WriteFile(keyFile, []byte("file-key"), 0o600))

	t.Run("missing cert file", func(t *testing.T) {
		conf := TLSConfig{
			CAFile:   caFile,
			CertFile: filepath.Join(dir, "missing.pem"),
			KeyFile:  keyFile,
		}

		_, err := conf.ToSSLConfig()
		assert.ErrorContains(t, err, "failed to read file")
	})

	t.Run("empty cert file", func(t *testing.T) {
		emptyFile := filepath.Join(dir, "empty.pem")
		require.NoError(t, os.WriteFile(emptyFile, nil, 0o600))
		conf := TLSConfig{
			CAFile:   caFile,
			CertFile: emptyFile,
			KeyFile:  keyFile,
		}

		_, err := conf.ToSSLConfig()
		assert.ErrorContains(t, err, "is empty")
	})
}

func TestTLSConfig_ToSSLConfigIgnoresResidualSourcesWhenDisabled(t *testing.T) {
	// CA is missing, so TLS is not enabled; residual cert/key files must not
	// be read, keeping backward compatibility with stale configs.
	conf := TLSConfig{
		InsecureSkipVerify: true,
		CertFile:           "/nonexistent/cert.pem",
		KeyFile:            "/nonexistent/key.pem",
	}

	sslConf, err := conf.ToSSLConfig()

	require.NoError(t, err)
	assert.False(t, sslConf.IsEnabled())
	assert.True(t, sslConf.InsecureSkipVerify)
	assert.Empty(t, sslConf.CA)
	assert.Empty(t, sslConf.Cert)
	assert.Empty(t, sslConf.Key)
}

func TestParseTLSConfigFromBase64(t *testing.T) {
	t.Run("valid base64", func(t *testing.T) {
		conf, err := parseTLSConfigFromBase64(
			base64.StdEncoding.EncodeToString([]byte("ca-content")),
			base64.StdEncoding.EncodeToString([]byte("cert-content")),
			base64.StdEncoding.EncodeToString([]byte("key-content")),
			"secret",
		)

		require.NoError(t, err)
		assert.Equal(t, "ca-content", conf.CA)
		assert.Equal(t, "cert-content", conf.Cert)
		assert.Equal(t, "key-content", conf.Key)
		assert.Equal(t, "secret", conf.Password)
	})

	t.Run("invalid base64", func(t *testing.T) {
		_, err := parseTLSConfigFromBase64(
			"not-valid-base64!!!",
			base64.StdEncoding.EncodeToString([]byte("cert-content")),
			base64.StdEncoding.EncodeToString([]byte("key-content")),
			"",
		)

		assert.ErrorContains(t, err, "failed to decode tls ca")
	})
}
