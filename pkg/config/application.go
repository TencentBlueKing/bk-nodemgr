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

// Package config ...
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	defaultApplicationFrontWindowsWMIPortDefault = 135
	defaultApplicationFrontUnixSSHPortDefault    = 22

	defaultApplicationBackendTraceServiceName     = "application-client-backend"
	defaultApplicationUserManagerTraceServiceName = "application-client-usermanager"
	defaultApplicationFileTraceServiceName        = "application-client-file"
	defaultApplicationNoticeTraceServiceName      = "application-client-notice"

	// info server config default values.
	defaultApplicationInfoBindIP                     = "127.0.0.1"
	defaultApplicationInfoBindIPV6                   = "::1"
	defaultApplicationInfoPort                       = 28000
	defaultApplicationInfoGracefulShutdownTimeoutSec = 60
	defaultApplicationInfoIdentity                   = AuthIdentityNone
	defaultApplicationInfoTraceServiceName           = "application-server-info"

	// admin server config default values.
	defaultApplicationAdminBindIP                     = "127.0.0.1"
	defaultApplicationAdminBindIPV6                   = "::1"
	defaultApplicationAdminPort                       = 28001
	defaultApplicationAdminGracefulShutdownTimeoutSec = 60
	defaultApplicationAdminIdentity                   = AuthIdentityRestServer
	defaultApplicationAdminTraceServiceName           = "application-server-admin"

	// basic server config default values.
	defaultApplicationBasicBindIP                     = "127.0.0.1"
	defaultApplicationBasicBindIPV6                   = "::1"
	defaultApplicationBasicPort                       = 28002
	defaultApplicationBasicGracefulShutdownTimeoutSec = 60
	defaultApplicationBasicStaticDir                  = "/bk-nodemgr/static/"
	defaultApplicationBasicIdentity                   = AuthIdentityBKLogin
	defaultApplicationBasicTraceServiceName           = "application-server-basic"

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

	defaultApplicationConfigPolicyOptionFilePath = "/bk-nodemgr/support-files/configpolicy"

	defaultApplicationIAMV3TraceServiceName = "application-client-iam-v3"
)

// ConfigPolicyOption defines config policy option file settings.
// nolint: revive
type ConfigPolicyOption struct {
	FilePath string `yaml:"filePath" usage:"config policy option file path"`
}

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
	RunMode            RunMode            `yaml:"runMode" usage:"run mode of service"`
	TenantMode         tenant.Mode        `yaml:"tenantMode" usage:"tenant mode of service"`
	BKSaas             BKSaas             `yaml:"bkSaaS" usage:"bk SaaS config of application service"`
	BKPaas             BKPaaS             `yaml:"bkPaaS" usage:"bk paas config of application service"`
	Front              Front              `yaml:"front" usage:"front config of application service"`
	Backend            Backend            `yaml:"backend" usage:"backend gateway config"`
	UserManager        UserManager        `yaml:"userManager" usage:"user manager config of application service"`
	Access             Access             `yaml:"access" usage:"access config of application service"`
	Notice             Notice             `yaml:"notice" usage:"notice gateway config"`
	File               File               `yaml:"file" usage:"file config of backend service"`
	Etcd               Etcd               `yaml:"etcd" usage:"etcd config of application service"`
	MongoDB            MongoDB            `yaml:"mongodb" usage:"mongodb config of application service"`
	InfoServer         HTTPServer         `yaml:"infoServer" usage:"info server config of application service"`
	AdminServer        HTTPServer         `yaml:"adminServer" usage:"admin server config of application service"`
	BasicServer        HTTPServer         `yaml:"basicServer" usage:"basic server config of application service"`
	Log                Log                `yaml:"log" usage:"log config of application service"`
	Tracing            Tracing            `yaml:"tracing" usage:"tracing config of file service"`
	Profiling          Profiling          `yaml:"profiling" usage:"profiling config of application service"`
	ConfigPolicyOption ConfigPolicyOption `yaml:"configPolicyOption" usage:"config policy option file settings"`
	IAMV3              IAMV3              `yaml:"iamV3" usage:"IAM v3 gateway config"`
}

// NewApplicationService generatea a new ApplicationService with default values.
func NewApplicationService() *ApplicationService {
	return &ApplicationService{
		RunMode:    defaultApplicationRunMode,
		TenantMode: defaultApplicationTenantMode,
		BKSaas: BKSaas{
			BKLogin: BKLogin{
				APIGatewayClient: APIGatewayClient{
					TraceService: TraceService{
						TraceServiceName: defaultApplicationBKLoginTraceServiceName,
					},
				},
			},
		},
		IAMV3: IAMV3{
			SystemID:     defaultIAMV3SystemID,
			CMDBSystemID: defaultIAMV3CMDBSystemID,
			APIGatewayClient: APIGatewayClient{
				TraceService: TraceService{
					TraceServiceName: defaultApplicationIAMV3TraceServiceName,
				},
			},
		},
		Front: Front{
			PasswordVaultSwitch:   defaultApplicationFrontPasswordVaultSwitch,
			PasswordVaultName:     defaultApplicationFrontPasswordVaultName,
			BKAppNavOpenSourceURL: defaultApplicationFrontBKAppNavOpenSourceURL,
			WindowsWMIPortDefault: defaultApplicationFrontWindowsWMIPortDefault,
			UnixSSHPortDefault:    defaultApplicationFrontUnixSSHPortDefault,
		},
		Backend: Backend{
			APIGatewayClient: APIGatewayClient{
				TraceService: TraceService{
					TraceServiceName: defaultApplicationBackendTraceServiceName,
				},
			},
		},
		UserManager: UserManager{
			APIGatewayClient: APIGatewayClient{
				TraceService: TraceService{
					TraceServiceName: defaultApplicationUserManagerTraceServiceName,
				},
			},
		},
		Access: Access{
			VirtualUser: defaultAccessVirtualUser,
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
			BindIP:                     defaultApplicationInfoBindIP,
			BindIPV6:                   defaultApplicationInfoBindIPV6,
			Port:                       defaultApplicationInfoPort,
			GracefulShutdownTimeoutSec: defaultApplicationInfoGracefulShutdownTimeoutSec,
			AuthIdentity:               defaultApplicationInfoIdentity,
			AdvertiseIPV4:              defaultApplicationAdvertiseIPv4,
			AdvertiseIPV6:              defaultApplicationAdvertiseIPv6,
			TraceService: TraceService{
				TraceServiceName: defaultApplicationInfoTraceServiceName,
			},
		},
		AdminServer: HTTPServer{
			BindIP:                     defaultApplicationAdminBindIP,
			BindIPV6:                   defaultApplicationAdminBindIPV6,
			Port:                       defaultApplicationAdminPort,
			GracefulShutdownTimeoutSec: defaultApplicationAdminGracefulShutdownTimeoutSec,
			AuthIdentity:               defaultApplicationAdminIdentity,
			AdvertiseIPV4:              defaultApplicationAdvertiseIPv4,
			AdvertiseIPV6:              defaultApplicationAdvertiseIPv6,
			JWTServerConfig:            JWTServerConfig{CryptoType: JWTCryptoTypeSymmetric},
			TraceService: TraceService{
				TraceServiceName: defaultApplicationAdminTraceServiceName,
			},
		},
		BasicServer: HTTPServer{
			BindIP:                     defaultApplicationBasicBindIP,
			BindIPV6:                   defaultApplicationBasicBindIPV6,
			Port:                       defaultApplicationBasicPort,
			GracefulShutdownTimeoutSec: defaultApplicationBasicGracefulShutdownTimeoutSec,
			StaticDir:                  defaultApplicationBasicStaticDir,
			AuthIdentity:               defaultApplicationBasicIdentity,
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
		ConfigPolicyOption: ConfigPolicyOption{
			FilePath: defaultApplicationConfigPolicyOptionFilePath,
		},
	}
}

// Load loads config from file or environment variables.
func (svc *ApplicationService) Load(filePath string) error {
	// default options.
	svc.RunMode = RunModeRelease

	if filePath == "" {
		return fmt.Errorf("failed to load config: file path is empty")
	}

	err := svc.LoadFromFile(filePath)
	if err != nil {
		return err
	}

	// Set default values for front config if not set.
	if svc.Front.BKAppNavOpenSourceURL == "" {
		svc.Front.BKAppNavOpenSourceURL = defaultApplicationFrontBKAppNavOpenSourceURL
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

	if err := svc.validateBKLogin(); err != nil {
		return err
	}

	if err := svc.BKPaas.Validate(); err != nil {
		return fmt.Errorf("failed to validate bkpaas config: %w", err)
	}

	if err := svc.Backend.Validate(); err != nil {
		return fmt.Errorf("failed to validate backend config: %w", err)
	}

	if err := svc.UserManager.Validate(); err != nil {
		return fmt.Errorf("failed to validate user manager config: %w", err)
	}

	if err := svc.Access.Validate(); err != nil {
		return fmt.Errorf("failed to validate access config: %w", err)
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

	if err := svc.Profiling.Validate(); err != nil {
		return fmt.Errorf("failed to validate profiling config: %w", err)
	}

	if strings.TrimSpace(svc.ConfigPolicyOption.FilePath) == "" {
		return errors.New("configPolicyOption.filePath is empty")
	}

	return nil
}

func (svc *ApplicationService) validateBKLogin() error {
	if err := svc.BKSaas.Validate(); err != nil {
		return fmt.Errorf("failed to validate bksaas config: %w", err)
	}

	if svc.TenantMode != tenant.ModeMultiple || svc.BKSaas.BKLogin.AuthType != LoginAuthTypeBKToken {
		return nil
	}

	if err := svc.BKSaas.BKLogin.APIGatewayClient.Validate(); err != nil {
		return fmt.Errorf("failed to validate bklogin api-gateway config: %w", err)
	}

	switch svc.BKSaas.BKLogin.AuthMode {
	case "un":
		if svc.BKSaas.BKLogin.User == "" {
			return errors.New("failed to validate bklogin api-gateway config: user is empty in un auth mode")
		}
	case "at":
		if svc.BKSaas.BKLogin.AccessToken == "" {
			return errors.New("failed to validate bklogin api-gateway config: access token is empty in at auth mode")
		}
	default:
		return fmt.Errorf(
			"failed to validate bklogin api-gateway config: unsupported auth mode: %s",
			svc.BKSaas.BKLogin.AuthMode,
		)
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
