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
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/envx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"gopkg.in/yaml.v2"
)

const (
	// application service config default values.
	defaultApplicationRunMode       = RunModeRelease
	defaultApplicationTenantMode    = tenant.ModeSingle
	defaultApplicationAPIGwUser     = "admin"
	defaultApplicationHTTPBindIP    = "127.0.0.1"
	defaultApplicationHTTPPort      = 5000
	defaultApplicationAdminBindIP   = "127.0.0.1"
	defaultApplicationAdminPort     = 5001
	defaultApplicationHTTPStaticDir = "/bk-nodemgr/static/"
	defaultApplicationLogDir        = "/bk-nodemgr/log/"
	defaultApplicationLogMaxNum     = 10
	defaultApplicationLogMaxSizeMB  = 200
	defaultApplicationLogLevel      = "INFO"
)

// BackendGateway the config of backend gateway config.
type BackendGateway struct {
	APIGatewayClient `yaml:",inline" usage:"api-gateway config of backend"`
}

// ApplicationService the config of application service.
type ApplicationService struct {
	RunMode     RunMode        `yaml:"mode" usage:"run mode of service"`
	TenantMode  tenant.Mode    `yaml:"tenantMode" usage:"tenant mode of service"`
	BKSaas      BKSaas         `yaml:"bkSaaS" usage:"bk SaaS config of application service"`
	Backend     BackendGateway `yaml:"backend" usage:"backend gateway config"`
	Etcd        Etcd           `yaml:"etcd" usage:"etcd config of application service"`
	MongoDB     MongoDB        `yaml:"mongodb" usage:"mongodb config of application service"`
	HTTPServer  HTTPServer     `yaml:"httpServer" usage:"http server config of application service"`
	AdminServer HTTPServer     `yaml:"adminServer" usage:"admin server config of application service"`
	Log         Log            `yaml:"log" usage:"log config of application service"`
}

// NewApplicationService generatea a new ApplicationService with default values.
func NewApplicationService() *ApplicationService {
	return &ApplicationService{
		RunMode:    defaultApplicationRunMode,
		TenantMode: defaultApplicationTenantMode,
		HTTPServer: HTTPServer{
			BindIP:    defaultApplicationHTTPBindIP,
			Port:      defaultApplicationHTTPPort,
			StaticDir: defaultApplicationHTTPStaticDir,
		},
		AdminServer: HTTPServer{
			BindIP: defaultApplicationAdminBindIP,
			Port:   defaultApplicationAdminPort,
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
func (svc *ApplicationService) LoadFromEnv() error {
	// run mode.
	var runMode string
	if envx.LoadString("NODEMAN_MODE", &runMode) {
		svc.RunMode = RunMode(runMode)
	}

	// api_gateway.
	if err := envx.MustLoadString("BKPAAS_APP_ID", &svc.Backend.AppCode); err != nil {
		return err
	}
	if err := envx.MustLoadString("BKPAAS_APP_SECRET", &svc.Backend.AppSecret); err != nil {
		return err
	}
	_ = envx.LoadString("NODEMAN_BACKEND_USER", &svc.Backend.User)
	_ = envx.LoadString("NODEMAN_BACKEND_AUTH_MODE", &svc.Backend.AuthMode)
	_ = envx.LoadString("NODEMAN_BACKEND_BK_TICKET", &svc.Backend.BkTicket)
	_ = envx.LoadString("NODEMAN_BACKEND_BK_TOKEN", &svc.Backend.BkToken)
	_ = envx.LoadString("NODEMAN_BACKEND_ACCESS_TOKEN", &svc.Backend.AccessToken)
	if _, err := envx.LoadBool("NODEMAN_BACKEND_TLS_SKIP_VERIFY", &svc.Backend.TLS.InsecureSkipVerify); err != nil {
		return err
	}
	_ = envx.LoadString("NODEMAN_BACKEND_TLS_CERT", &svc.Backend.TLS.CertFile)
	_ = envx.LoadString("NODEMAN_BACKEND_TLS_KEY", &svc.Backend.TLS.KeyFile)
	_ = envx.LoadString("NODEMAN_BACKEND_TLS_CA", &svc.Backend.TLS.CAFile)
	_ = envx.LoadString("NODEMAN_BACKEND_TLS_PASSWORD", &svc.Backend.TLS.Password)

	// http_server.
	_ = envx.LoadString("NODEMAN_HTTPSVR_BIND_IP", &svc.HTTPServer.BindIP)
	if _, err := envx.LoadInt("NODEMAN_HTTPSVR_PORT", &svc.HTTPServer.Port); err != nil {
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
	if err := svc.BKSaas.Validate(); err != nil {
		return fmt.Errorf("failed to validate application config: %w", err)
	}

	// TODO: validate the config
	return nil
}

// EnvGet read env, supports default value.
func EnvGet(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

// Front front setting of application service.
type Front struct {
	BKLoginURL string `yaml:"bkLoginURL" usage:"bk login url of front setting"`
}

// Validate validates the config.
func (svc *Front) Validate() error {
	if svc.BKLoginURL == "" {
		return errors.New("failed to validate front config: bkLoginURL can not be empty")
	}

	return nil
}
