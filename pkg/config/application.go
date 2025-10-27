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
	defaultApplicationFrontPasswordVaultSwitch = false
	defaultApplicationFrontPasswordVaultName   = "password_vault"

	// info server config default values.
	defaultApplicationInfoBindIP   = "127.0.0.1"
	defaultApplicationInfoPort     = 28000
	defaultApplicationInfoIdentity = AuthIdentityNone

	// admin server config default values.
	defaultApplicationAdminBindIP   = "127.0.0.1"
	defaultApplicationAdminPort     = 28001
	defaultApplicationAdminIdentity = AuthIdentityRestServer

	// basic server config default values.
	defaultApplicationBasicBindIP    = "127.0.0.1"
	defaultApplicationBasicPort      = 28002
	defaultApplicationBasicStaticDir = "/bk-nodemgr/static/"
	defaultApplicationBasicIdentity  = AuthIdentityBKLogin

	// log config default values.
	defaultApplicationLogDir       = "/bk-nodemgr/log/"
	defaultApplicationLogMaxNum    = 10
	defaultApplicationLogMaxSizeMB = 200
	defaultApplicationLogLevel     = "INFO"

	// advertise config default values.
	defaultApplicationAdvertiseIPv4 = "127.0.0.1"
	defaultApplicationAdvertiseIPv6 = "::1"
)

// Backend the config of backend gateway config.
type Backend struct {
	APIGatewayClient `yaml:",inline" usage:"api-gateway config of backend"`
}

// ApplicationService the config of application service.
type ApplicationService struct {
	RunMode     RunMode     `yaml:"mode" usage:"run mode of service"`
	TenantMode  tenant.Mode `yaml:"tenantMode" usage:"tenant mode of service"`
	BKSaas      BKSaas      `yaml:"bkSaaS" usage:"bk SaaS config of application service"`
	BKPaas      BKPaaS      `yaml:"bkPaaS" usage:"bk paas config of application service"`
	Front       Front       `yaml:"front" usage:"front config of application service"`
	Backend     Backend     `yaml:"backend" usage:"backend gateway config"`
	File        File        `yaml:"file" usage:"file config of backend service"`
	Etcd        Etcd        `yaml:"etcd" usage:"etcd config of application service"`
	MongoDB     MongoDB     `yaml:"mongodb" usage:"mongodb config of application service"`
	InfoServer  HTTPServer  `yaml:"infoServer" usage:"info server config of application service"`
	AdminServer HTTPServer  `yaml:"adminServer" usage:"admin server config of application service"`
	BasicServer HTTPServer  `yaml:"basicServer" usage:"basic server config of application service"`
	Log         Log         `yaml:"log" usage:"log config of application service"`
}

// NewApplicationService generatea a new ApplicationService with default values.
func NewApplicationService() *ApplicationService {
	return &ApplicationService{
		RunMode:    defaultApplicationRunMode,
		TenantMode: defaultApplicationTenantMode,
		Front: Front{
			PasswordVaultSwitch: defaultApplicationFrontPasswordVaultSwitch,
			PasswordVaultName:   defaultApplicationFrontPasswordVaultName,
		},
		InfoServer: HTTPServer{
			BindIP:        defaultApplicationInfoBindIP,
			Port:          defaultApplicationInfoPort,
			AuthIdentity:  defaultApplicationInfoIdentity,
			AdvertiseIPV4: defaultApplicationAdvertiseIPv4,
			AdvertiseIPV6: defaultApplicationAdvertiseIPv6,
		},
		AdminServer: HTTPServer{
			BindIP:          defaultApplicationAdminBindIP,
			Port:            defaultApplicationAdminPort,
			AuthIdentity:    defaultApplicationAdminIdentity,
			AdvertiseIPV4:   defaultApplicationAdvertiseIPv4,
			AdvertiseIPV6:   defaultApplicationAdvertiseIPv6,
			JWTServerConfig: JWTServerConfig{CryptoType: JWTCryptoTypeSymmetric},
		},
		BasicServer: HTTPServer{
			BindIP:       defaultApplicationBasicBindIP,
			Port:         defaultApplicationBasicPort,
			StaticDir:    defaultApplicationBasicStaticDir,
			AuthIdentity: defaultApplicationBasicIdentity,
		},
		Log: Log{
			Dir:       defaultApplicationLogDir,
			MaxSizeMB: defaultApplicationLogMaxSizeMB,
			MaxNum:    defaultApplicationLogMaxNum,
			Level:     defaultApplicationLogLevel,
		},
	}
}

// Load loads config from file or environment variables.
func (svc *ApplicationService) Load(filePath string) error {
	// default options.
	svc.RunMode = RunModeRelease

	if filePath == "" {
		return svc.LoadFromEnv()
	}

	return svc.LoadFromFile(filePath)
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
	_ = envx.LoadString("NODEMAN_ETCD_CERT", &svc.Etcd.Cert)
	_ = envx.LoadString("NODEMAN_ETCD_KEY", &svc.Etcd.Key)
	_ = envx.LoadString("NODEMAN_ETCD_CA", &svc.Etcd.Ca)

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

	return nil
}

// EnvGet read env, supports default value.
func EnvGet(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}
