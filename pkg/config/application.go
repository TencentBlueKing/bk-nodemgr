/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package config ...
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/envx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
	"gopkg.in/yaml.v2"
)

const (
	// run mode default values.
	defaultApplicationRunMode = RunModeRelease

	// tenant mode default values.
	defaultApplicationTenantMode = tenant.ModeSingle

	// front config default values.
	defaultApplicationFrontPasswordVaultSwitch   = false
	defaultApplicationFrontPasswordVaultName     = "password_vault"
	defaultApplicationFrontBKAppNavOpenSourceURL = "https://github.com/TencentBlueKing/bk-nodemgr"

	defaultApplicationBackendTraceServiceName = "application-client-backend"
	defaultApplicationFileTraceServiceName    = "application-client-file"
	defaultApplicationNoticeTraceServiceName  = "application-client-notice"

	// info server config default values.
	defaultApplicationInfoBindIP           = "127.0.0.1"
	defaultApplicationInfoPort             = 28000
	defaultApplicationInfoIdentity         = AuthIdentityNone
	defaultApplicationInfoTraceServiceName = "application-server-info"

	// admin server config default values.
	defaultApplicationAdminBindIP           = "127.0.0.1"
	defaultApplicationAdminPort             = 28001
	defaultApplicationAdminIdentity         = AuthIdentityRestServer
	defaultApplicationAdminTraceServiceName = "application-server-admin"

	// basic server config default values.
	defaultApplicationBasicBindIP           = "127.0.0.1"
	defaultApplicationBasicPort             = 28002
	defaultApplicationBasicStaticDir        = "/bk-nodemgr/static/"
	defaultApplicationBasicIdentity         = AuthIdentityBKLogin
	defaultApplicationBasicTraceServiceName = "application-server-basic"

	// log config default values.
	defaultApplicationLogDir       = "/bk-nodemgr/log/"
	defaultApplicationLogMaxNum    = 10
	defaultApplicationLogMaxSizeMB = 200
	defaultApplicationLogLevel     = "INFO"

	defaultApplicationMongoDBAppName          = "bk_nodemgr_application"
	defaultApplicationMongoDBTraceServiceName = "bk_nodemgr_mongo"

	// advertise config default values.
	defaultApplicationAdvertiseIPv4 = "127.0.0.1"
	defaultApplicationAdvertiseIPv6 = "::1"

	defaultApplicationTracingExporterType = "stdout"

	defaultApplicationBKLoginTraceServiceName = "application-client-bklogin"
)

// Backend the config of backend gateway config.
type Backend struct {
	APIGatewayClient `yaml:",inline" usage:"api-gateway config of backend"`
}

// Notice the config of notice gateway config.
type Notice struct {
	Enabled          bool `yaml:"enabled" usage:"enable notice feature"`
	APIGatewayClient `yaml:",inline" usage:"api-gateway config of notice"`
}

// Validate validates the notice config.
// When Enabled is false, validation is skipped to allow graceful degradation.
// When Enabled is true, validates all APIGatewayClient fields.
func (n *Notice) Validate() error {
	if !n.Enabled {
		return nil
	}

	return n.APIGatewayClient.Validate()
}

// ApplicationService the config of application service.
type ApplicationService struct {
	RunMode     RunMode     `yaml:"mode" usage:"run mode of service"`
	TenantMode  tenant.Mode `yaml:"tenantMode" usage:"tenant mode of service"`
	BKSaas      BKSaas      `yaml:"bkSaaS" usage:"bk SaaS config of application service"`
	BKPaas      BKPaaS      `yaml:"bkPaaS" usage:"bk paas config of application service"`
	Front       Front       `yaml:"front" usage:"front config of application service"`
	Backend     Backend     `yaml:"backend" usage:"backend gateway config"`
	Notice      Notice      `yaml:"notice" usage:"notice gateway config"`
	File        File        `yaml:"file" usage:"file config of backend service"`
	Etcd        Etcd        `yaml:"etcd" usage:"etcd config of application service"`
	MongoDB     MongoDB     `yaml:"mongodb" usage:"mongodb config of application service"`
	InfoServer  HTTPServer  `yaml:"infoServer" usage:"info server config of application service"`
	AdminServer HTTPServer  `yaml:"adminServer" usage:"admin server config of application service"`
	BasicServer HTTPServer  `yaml:"basicServer" usage:"basic server config of application service"`
	Log         Log         `yaml:"log" usage:"log config of application service"`
	Tracing     Tracing     `yaml:"tracing" usage:"tracing config of file service"`
}

// NewApplicationService generatea a new ApplicationService with default values.
func NewApplicationService() *ApplicationService {
	return &ApplicationService{
		RunMode:    defaultApplicationRunMode,
		TenantMode: defaultApplicationTenantMode,
		BKSaas: BKSaas{
			BKLogin: BKLogin{
				TraceService: TraceService{
					TraceServiceName: defaultApplicationBKLoginTraceServiceName,
				},
			},
		},
		Front: Front{
			PasswordVaultSwitch:   defaultApplicationFrontPasswordVaultSwitch,
			PasswordVaultName:     defaultApplicationFrontPasswordVaultName,
			BKAppNavOpenSourceURL: defaultApplicationFrontBKAppNavOpenSourceURL,
		},
		Backend: Backend{
			APIGatewayClient: APIGatewayClient{
				TraceService: TraceService{
					TraceServiceName: defaultApplicationBackendTraceServiceName,
				},
			},
		},
		Notice: Notice{
			Enabled: false, // Disabled by default, must be explicitly enabled
			APIGatewayClient: APIGatewayClient{
				TraceService: TraceService{
					TraceServiceName: defaultApplicationNoticeTraceServiceName,
				},
			},
		},
		File: File{
			TraceService: TraceService{
				TraceServiceName: defaultApplicationFileTraceServiceName,
			},
		},
		Etcd: Etcd{},
		MongoDB: MongoDB{
			AppName: defaultApplicationMongoDBAppName,
			TraceService: TraceService{
				TraceServiceName: defaultApplicationMongoDBTraceServiceName,
			},
		},
		InfoServer: HTTPServer{
			BindIP:        defaultApplicationInfoBindIP,
			Port:          defaultApplicationInfoPort,
			AuthIdentity:  defaultApplicationInfoIdentity,
			AdvertiseIPV4: defaultApplicationAdvertiseIPv4,
			AdvertiseIPV6: defaultApplicationAdvertiseIPv6,
			TraceService: TraceService{
				TraceServiceName: defaultApplicationInfoTraceServiceName,
			},
		},
		AdminServer: HTTPServer{
			BindIP:          defaultApplicationAdminBindIP,
			Port:            defaultApplicationAdminPort,
			AuthIdentity:    defaultApplicationAdminIdentity,
			AdvertiseIPV4:   defaultApplicationAdvertiseIPv4,
			AdvertiseIPV6:   defaultApplicationAdvertiseIPv6,
			JWTServerConfig: JWTServerConfig{CryptoType: JWTCryptoTypeSymmetric},
			TraceService: TraceService{
				TraceServiceName: defaultApplicationAdminTraceServiceName,
			},
		},
		BasicServer: HTTPServer{
			BindIP:       defaultApplicationBasicBindIP,
			Port:         defaultApplicationBasicPort,
			StaticDir:    defaultApplicationBasicStaticDir,
			AuthIdentity: defaultApplicationBasicIdentity,
			TraceService: TraceService{
				TraceServiceName: defaultApplicationBasicTraceServiceName,
			},
		},
		Log: Log{
			Dir:       defaultApplicationLogDir,
			MaxSizeMB: defaultApplicationLogMaxSizeMB,
			MaxNum:    defaultApplicationLogMaxNum,
			Level:     defaultApplicationLogLevel,
		},
		Tracing: Tracing{
			ExporterType: defaultApplicationTracingExporterType,
		},
	}
}

// Load loads config from file or environment variables.
func (svc *ApplicationService) Load(filePath string) error {
	// default options.
	svc.RunMode = RunModeRelease

	var err error
	if filePath == "" {
		err = svc.LoadFromEnv()
	} else {
		err = svc.LoadFromFile(filePath)
	}
	if err != nil {
		return err
	}

	// Set default values for front config if not set.
	if svc.Front.BKAppNavOpenSourceURL == "" {
		svc.Front.BKAppNavOpenSourceURL = defaultApplicationFrontBKAppNavOpenSourceURL
	}

	return nil
}

// LoadFromEnv loads config from environment variables.
// nolint: gocyclo, cyclop, funlen
func (svc *ApplicationService) LoadFromEnv() error {
	// run mode.
	var runMode string
	if err := envx.MustLoadString("NODEMAN_MODE", &runMode); err != nil {
		return err
	}
	svc.RunMode = RunMode(runMode)

	var tenantMode string
	if err := envx.MustLoadString("NODEMAN_TENANT_MODE", &tenantMode); err != nil {
		return err
	}
	svc.TenantMode = tenant.Mode(tenantMode)

	// bk SaaS.
	if err := envx.MustLoadString("BK_BKSAAS_BKLOGIN_LOGIN_URL", &svc.BKSaas.BKLogin.LoginURL); err != nil {
		return err
	}
	if _, err := envx.LoadBool("BK_BKSAAS_BKLOGIN_TLS_INSECURE_SKIP_VERIFY", &svc.Backend.TLS.InsecureSkipVerify); err != nil {
		return err
	}
	_ = envx.LoadString("BK_BKSAAS_BKLOGIN_TLS_CERT", &svc.Backend.TLS.CertFile)
	_ = envx.LoadString("BK_BKSAAS_BKLOGIN_TLS_KEY", &svc.Backend.TLS.KeyFile)
	_ = envx.LoadString("BK_BKSAAS_BKLOGIN_TLS_CA", &svc.Backend.TLS.CAFile)
	_ = envx.LoadString("BK_BKSAAS_BKLOGIN_TLS_PASSWORD", &svc.Backend.TLS.Password)

	// bk PaaS
	_ = envx.LoadString("BK_PAAS_ANALYSIS_SCRIPT", &svc.BKPaas.AnalysisScript)

	// backend.
	if err := envx.MustLoadString("BKPAAS_APP_ID", &svc.Backend.AppCode); err != nil {
		return err
	}
	if err := envx.MustLoadString("BKPAAS_APP_SECRET", &svc.Backend.AppSecret); err != nil {
		return err
	}
	_ = envx.MustLoadString("NODEMAN_BACKEND_USER", &svc.Backend.User)
	_ = envx.MustLoadString("NODEMAN_BACKEND_AUTH_MODE", &svc.Backend.AuthMode)
	_ = envx.MustLoadString("NODEMAN_BACKEND_ACCESS_TOKEN", &svc.Backend.AccessToken)
	if _, err := envx.LoadBool("NODEMAN_BACKEND_TLS_SKIP_VERIFY", &svc.Backend.TLS.InsecureSkipVerify); err != nil {
		return err
	}

	if _, err := envx.LoadBool("NODEMAN_BACKEND_TLS_INSECURE_SKIP_VERIFY", &svc.Backend.TLS.InsecureSkipVerify); err != nil {
		return err
	}
	_ = envx.LoadString("NODEMAN_BACKEND_TLS_CERT", &svc.Backend.TLS.CertFile)
	_ = envx.LoadString("NODEMAN_BACKEND_TLS_KEY", &svc.Backend.TLS.KeyFile)
	_ = envx.LoadString("NODEMAN_BACKEND_TLS_CA", &svc.Backend.TLS.CAFile)
	_ = envx.LoadString("NODEMAN_BACKEND_TLS_PASSWORD", &svc.Backend.TLS.Password)

	// notice.
	if _, err := envx.LoadBool("NODEMAN_NOTICE_ENABLED", &svc.Notice.Enabled); err != nil {
		return err
	}
	// Only load Notice configuration when enabled.
	if svc.Notice.Enabled {
		if err := envx.MustLoadString("NODEMAN_NOTICE_APP_CODE", &svc.Notice.AppCode); err != nil {
			return err
		}
		if err := envx.MustLoadString("NODEMAN_NOTICE_APP_SECRET", &svc.Notice.AppSecret); err != nil {
			return err
		}
		if err := envx.MustLoadString("NODEMAN_NOTICE_USER", &svc.Notice.User); err != nil {
			return err
		}
		if err := envx.MustLoadString("NODEMAN_NOTICE_AUTH_MODE", &svc.Notice.AuthMode); err != nil {
			return err
		}
		_ = envx.MustLoadString("NODEMAN_NOTICE_ACCESS_TOKEN", &svc.Notice.AccessToken)
		if _, err := envx.LoadBool("NODEMAN_NOTICE_TLS_INSECURE_SKIP_VERIFY", &svc.Notice.TLS.InsecureSkipVerify); err != nil {
			return err
		}
		_ = envx.LoadString("NODEMAN_NOTICE_TLS_CERT", &svc.Notice.TLS.CertFile)
		_ = envx.LoadString("NODEMAN_NOTICE_TLS_KEY", &svc.Notice.TLS.KeyFile)
		_ = envx.LoadString("NODEMAN_NOTICE_TLS_CA", &svc.Notice.TLS.CAFile)
		_ = envx.LoadString("NODEMAN_NOTICE_TLS_PASSWORD", &svc.Notice.TLS.Password)
	}

	// etcd
	var etcdEndpoints string
	if err := envx.MustLoadString("NODEMAN_ETCD_ENDPOINTS", &etcdEndpoints); err != nil {
		return err
	}
	svc.Etcd.Endpoints = strings.Split(etcdEndpoints, ",")
	if err := envx.MustLoadString("NODEMAN_ETCD_USERNAME", &svc.Etcd.Username); err != nil {
		return err
	}
	if err := envx.MustLoadString("NODEMAN_ETCD_PASSWORD", &svc.Etcd.Password); err != nil {
		return err
	}
	_ = envx.LoadString("NODEMAN_ETCD_CERT", &svc.Etcd.TLS.CertFile)
	_ = envx.LoadString("NODEMAN_ETCD_KEY", &svc.Etcd.TLS.KeyFile)
	_ = envx.LoadString("NODEMAN_ETCD_CA", &svc.Etcd.TLS.CAFile)

	// mongodb.
	var mongoDBHosts string
	if err := envx.MustLoadString("NODEMAN_MONGODB_HOSTS", &mongoDBHosts); err != nil {
		return err
	}
	svc.MongoDB.Hosts = strings.Split(mongoDBHosts, ",")

	if err := envx.MustLoadString("NODEMAN_MONGODB_USERNAME", &svc.MongoDB.Username); err != nil {
		return err
	}
	if err := envx.MustLoadString("NODEMAN_MONGODB_PASSWORD", &svc.MongoDB.Password); err != nil {
		return err
	}
	if err := envx.MustLoadString("NODEMAN_MONGODB_DATABASE", &svc.MongoDB.Database); err != nil {
		return err
	}
	if err := envx.MustLoadString("NODEMAN_MONGODB_AUTH_SOURCE", &svc.MongoDB.AuthSource); err != nil {
		return err
	}
	if err := envx.MustLoadString("NODEMAN_MONGODB_AUTH_MECHANISM", &svc.MongoDB.AuthMechanism); err != nil {
		return err
	}

	// http_server.
	_ = envx.LoadString("NODEMAN_HTTPSVR_BIND_IP", &svc.BasicServer.BindIP)
	if _, err := envx.LoadInt("NODEMAN_HTTPSVR_PORT", &svc.BasicServer.Port); err != nil {
		return err
	}

	// log.
	_ = envx.LoadString("NODEMAN_LOG_DIR", &svc.Log.Dir)
	_ = envx.LoadString("NODEMAN_LOG_LEVEL", (*string)(&svc.Log.Level))
	if _, err := envx.LoadInt("NODEMAN_LOG_MAX_NUM", &svc.Log.MaxNum); err != nil {
		return err
	}
	if _, err := envx.LoadInt("NODEMAN_LOG_MAX_SIZE_MB", &svc.Log.MaxSizeMB); err != nil {
		return err
	}

	// front config.
	_ = envx.LoadString("BK_NODEMGR_APPLICATION_USER_WEB_URL", &svc.Front.BKUserWebURL)
	_ = envx.LoadString("BK_NODEMGR_APPLICATION_DOMAIN", &svc.Front.BKDomain)
	_ = envx.LoadString("BK_NODEMGR_APPLICATION_DOCS_CENTER_URL", &svc.Front.BKDocsCenterURL)
	_ = envx.LoadString("BK_NODEMGR_APPLICATION_NAV_OPEN_SOURCE_URL", &svc.Front.BKAppNavOpenSourceURL)

	return nil
}

// LoadFromFile loads config from file.
func (svc *ApplicationService) LoadFromFile(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	absPath = filepath.Clean(absPath)
	configContent, err := os.ReadFile(absPath)
	if err != nil {
		return err
	}

	if err = yaml.Unmarshal(configContent, svc); err != nil {
		return err
	}

	return nil
}

// Validate validates the config.
func (svc *ApplicationService) Validate() error {
	if err := svc.RunMode.Validate(); err != nil {
		return fmt.Errorf("failed to validate run mode config: %w", err)
	}

	if err := svc.TenantMode.Validate(); err != nil {
		return fmt.Errorf("failed to validate tenant mode config: %w", err)
	}

	if err := svc.BKSaas.Validate(); err != nil {
		return fmt.Errorf("failed to validate bksaas config: %w", err)
	}

	if err := svc.BKPaas.Validate(); err != nil {
		return fmt.Errorf("failed to validate bkpaas config: %w", err)
	}

	if err := svc.Backend.Validate(); err != nil {
		return fmt.Errorf("failed to validate backend config: %w", err)
	}

	if err := svc.Notice.Validate(); err != nil {
		return fmt.Errorf("failed to validate notice config: %w", err)
	}

	if err := svc.File.Validate(); err != nil {
		return fmt.Errorf("failed to validate file config: %w", err)
	}

	if err := svc.Etcd.Validate(); err != nil {
		return fmt.Errorf("failed to validate etcd config: %w", err)
	}

	if err := svc.MongoDB.Validate(); err != nil {
		return fmt.Errorf("failed to validate mongodb config: %w", err)
	}

	if err := svc.InfoServer.Validate(); err != nil {
		return fmt.Errorf("failed to validate info service config: %w", err)
	}

	if err := svc.AdminServer.Validate(); err != nil {
		return fmt.Errorf("failed to validate admin service config: %w", err)
	}

	if err := svc.BasicServer.Validate(); err != nil {
		return fmt.Errorf("failed to validate basic service config: %w", err)
	}

	if err := svc.Log.Validate(); err != nil {
		return fmt.Errorf("failed to validate log config: %w", err)
	}

	if err := svc.Front.Validate(); err != nil {
		return fmt.Errorf("failed to validate front config: %w", err)
	}

	if err := svc.Tracing.Validate(); err != nil {
		return fmt.Errorf("failed to validate tracing config: %w", err)
	}

	return nil
}

// EnvGet read env, supports default value.
func EnvGet(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}
