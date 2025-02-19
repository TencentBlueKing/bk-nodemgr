/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package config provides configuration.
package config

import (
	"fmt"
	"os"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/envx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"gopkg.in/yaml.v2"
)

const (
	// backend service config default values.
	defaultBackendRunMode        = RunModeRelease
	defaultBackendTenantMode     = tenant.ModeSingle
	defaultBackendHTTPBindIP     = "127.0.0.1"
	defaultBackendHTTPPort       = 8000
	defaultBackendAdminBindIP    = "127.0.0.1"
	defaultBackendAdminPort      = 8001
	defaultBackendCallbackBindIP = "127.0.0.1"
	defaultBackendCallbackPort   = 8002
	defaultBackendLogDir         = "/bk-nodeman/log/"
	defaultBackendLogMaxNum      = 10
	defaultBackendLogMaxSizeMB   = 200
	defaultBackendLogLevel       = "INFO"

	// application service config default values.
	defaultApplicationRunMode       = RunModeRelease
	defaultApplicationTenantMode    = tenant.ModeSingle
	defaultApplicationAPIGwUser     = "admin"
	defaultApplicationHTTPBindIP    = "127.0.0.1"
	defaultApplicationHTTPPort      = 5000
	defaultApplicationAdminBindIP   = "127.0.0.1"
	defaultApplicationAdminPort     = 5001
	defaultApplicationHTTPStaticDir = "/bk-nodeman/static/"
	defaultApplicationLogDir        = "/bk-nodeman/log/"
	defaultApplicationLogMaxNum     = 10
	defaultApplicationLogMaxSizeMB  = 200
	defaultApplicationLogLevel      = "INFO"
	defaultEncryptKey               = "1234567890123456"
)

// Etcd the config of etcd.
type Etcd struct {
	Endpoints string `yaml:"endpoints" usage:"endpoints of etcd"`
	Cert      string `yaml:"cert" usage:"cert file of etcd"`
	Key       string `yaml:"key" usage:"key file for etcd"`
	Ca        string `yaml:"ca" usage:"ca file for etcd"`
}

// Redis the config of redis.
type Redis struct {
	Host     string `yaml:"host" usage:"host of redis"`
	Port     int    `yaml:"port" usage:"port of redis"`
	Password string `yaml:"password" usage:"password of redis"`
	DB       int    `yaml:"db" usage:"db of redis"`
}

// Validate configures the config.
func (conf Redis) Validate() error {
	return nil
}

// MongoDB the config of mongodb.
type MongoDB struct {
	Hosts         []string `yaml:"hosts" usage:"hosts list of mongodb"`
	Username      string   `yaml:"username" usage:"user of mongodb"`
	Password      string   `yaml:"password" usage:"password of mongodb"`
	Database      string   `yaml:"database" usage:"database of mongodb"`
	AuthSource    string   `yaml:"authSource" usage:"auth source of mongodb"`
	AuthMechanism string   `yaml:"authMechanism" usage:"auth mechanism of mongodb"`
}

// Log the config of log.
type Log struct {
	Dir          string `yaml:"dir" usage:"log dir of backend server"`
	MaxSizeMB    int    `yaml:"maxSizeMB" usage:"max size in MBytes of single log file"`
	MaxNum       int    `yaml:"maxNum" usage:"max number of log files"`
	Level        string `yaml:"level" usage:"log level of backend server. DEBUG, INFO, WARN, ERROR"`
	ToStdErr     bool   `yaml:"toStderr" usage:"log to stderr instead of files"`
	AlsoToStdErr bool   `yaml:"alsoToStderr" usage:"log to stderr in addition to files"`
}

// HTTPServer the config of http service.
type HTTPServer struct {
	BindIP    string `yaml:"bindIP"`
	Port      int    `yaml:"port"`
	StaticDir string `yaml:"staticDir"`
}

// AdminServer the config of admin service.
type AdminServer struct {
	BindIP    string `yaml:"bindIP"`
	Port      int    `yaml:"port"`
	StaticDir string `yaml:"staticDir"`
}

// CallbackServer the config of callback service.
type CallbackServer struct {
	BindIP    string `yaml:"bindIP"`
	Port      int    `yaml:"port"`
	StaticDir string `yaml:"staticDir"`
}

// APIGateway the config of api-gateway.
type APIGateway struct {
	// Endpoints is a seed list of host:port addresses of api gateway nodes.
	Endpoints []string `yaml:"endpoints"`
	// AppCode is the BlueKing app code of nodeman to request api gateway.
	AppCode string `yaml:"appCode"`
	// AppSecret is the BlueKing app secret of nodeman to request api gateway.
	AppSecret string `yaml:"appSecret"`
	// User is the BlueKing user of nodeman to request api gateway.
	User string `yaml:"user"`
	// AuthMode is the BlueKing api authentication mode.
	AuthMode string `yaml:"authMode"`
	// BkTicket is the BlueKing access ticket of nodeman to request api gateway.
	BkTicket string `yaml:"bkTicket"`
	// BkToken is the BlueKing user token of nodeman to request api gateway.
	BkToken string `yaml:"bkToken"`
	// AccessToken is the BlueKing access token of nodeman to request api gateway.
	AccessToken string `yaml:"accessToken"`
	// TLS defines the tls config of api-gateway.
	TLS TLSConfig `yaml:"tls" usage:"tls config of api-gateway"`
}

// CMDB the config of cmdb.
type CMDB struct {
	TenantID   string
	APIGateway `yaml:",inline" usage:"api-gateway config of cmdb"`
}

// TLSConfig defines tls related options.
type TLSConfig struct {
	// Server should be accessed without verifying the TLS certificate.
	// For testing only.
	InsecureSkipVerify bool `yaml:"insecureSkipVerify"`
	// Server requires TLS client certificate authentication
	CertFile string `yaml:"certFile"`
	// Server requires TLS client certificate authentication
	KeyFile string `yaml:"keyFile"`
	// Trusted root certificates for server
	CAFile string `yaml:"caFile"`
	// the password to decrypt the certificate
	Password string `yaml:"password"`
}

// Workflow the config of workflow.
type Workflow struct {
	WorkerNum int `yaml:"workerNum" usage:"worker num of workflow"`
}

// Validate validates the config.
func (conf Workflow) Validate() error {
	if conf.WorkerNum <= 0 {
		return fmt.Errorf("worker num must be greater than 0, worker-num(%d)", conf.WorkerNum)
	}

	return nil
}

// RunMode the run mode of service.
type RunMode string

const (
	// RunModeRelease release mode.
	RunModeRelease RunMode = "release"

	// RunModeDebug debug mode.
	RunModeDebug RunMode = "debug"
)

// BackendService the config of backend service.
type BackendService struct {
	RunMode        RunMode        `yaml:"runMode" usage:"run mode of service"`
	TenantMode     tenant.Mode    `yaml:"tenantMode" usage:"tenant mode of service"`
	CMDB           CMDB           `yaml:"cmdb" usage:"cmdb config of backend service"`
	Workflow       Workflow       `yaml:"workflow" usage:"workflow config of backend service"`
	HTTPServer     HTTPServer     `yaml:"httpServer" usage:"http server config of backend service"`
	AdminServer    AdminServer    `yaml:"adminServer" usage:"admin server config of backend service"`
	CallbackServer CallbackServer `yaml:"callbackServer" usage:"callback server config of backend service"`
	Redis          Redis          `yaml:"redis" usage:"redis config of backend service"`
	MongoDB        MongoDB        `yaml:"mongodb" usage:"mongodb config of backend service"`
	Log            Log            `yaml:"log" usage:"log config of backend service"`
	EncryptKey     string         `yaml:"encryptKey" usage:"encrypt key of backend service"`
}

// NewBackendService generates a new BackendService with default values.
func NewBackendService() *BackendService {
	return &BackendService{
		RunMode:    defaultBackendRunMode,
		TenantMode: defaultBackendTenantMode,
		HTTPServer: HTTPServer{
			BindIP: defaultBackendHTTPBindIP,
			Port:   defaultBackendHTTPPort,
		},
		AdminServer: AdminServer{
			BindIP: defaultBackendAdminBindIP,
			Port:   defaultBackendAdminPort,
		},
		CallbackServer: CallbackServer{
			BindIP: defaultBackendCallbackBindIP,
			Port:   defaultBackendCallbackPort,
		},
		Log: Log{
			Dir:       defaultBackendLogDir,
			MaxSizeMB: defaultBackendLogMaxSizeMB,
			MaxNum:    defaultBackendLogMaxNum,
			Level:     defaultBackendLogLevel,
		},
		EncryptKey: defaultEncryptKey,
	}
}

// LoadFromFile loads config from file.
func (b *BackendService) LoadFromFile(path string) error {
	configContent, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err = yaml.Unmarshal(configContent, b); err != nil {
		return err
	}

	return nil
}

// Validate validates the config.
func (b *BackendService) Validate() error {
	if err := b.Workflow.Validate(); err != nil {
		return err
	}

	return nil
}

// ApplicationService the config of application service.
type ApplicationService struct {
	RunMode     RunMode     `yaml:"mode" usage:"run mode of service"`
	TenantMode  tenant.Mode `yaml:"tenantMode" usage:"tenant mode of service"`
	APIGateway  APIGateway  `yaml:"apiGateway" usage:"auth config of application service"`
	HTTPServer  HTTPServer  `yaml:"httpServer" usage:"http server config of application service"`
	AdminServer AdminServer `yaml:"adminServer" usage:"admin server config of application service"`
	Log         Log         `yaml:"log" usage:"log config of application service"`
}

// NewApplicationService generatea a new ApplicationService with default values.
func NewApplicationService() *ApplicationService {
	return &ApplicationService{
		RunMode:    defaultApplicationRunMode,
		TenantMode: defaultApplicationTenantMode,
		APIGateway: APIGateway{
			User: defaultApplicationAPIGwUser,
		},
		HTTPServer: HTTPServer{
			BindIP:    defaultApplicationHTTPBindIP,
			Port:      defaultApplicationHTTPPort,
			StaticDir: defaultApplicationHTTPStaticDir,
		},
		AdminServer: AdminServer{
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
	if err := envx.MustLoadString("BKPAAS_APP_ID", &svc.APIGateway.AppCode); err != nil {
		return err
	}
	if err := envx.MustLoadString("BKPAAS_APP_SECRET", &svc.APIGateway.AppSecret); err != nil {
		return err
	}
	_ = envx.LoadString("NODEMAN_APIGW_USER", &svc.APIGateway.User)
	_ = envx.LoadString("NODEMAN_APIGW_AUTH_MODE", &svc.APIGateway.AuthMode)
	_ = envx.LoadString("NODEMAN_APIGW_BK_TICKET", &svc.APIGateway.BkTicket)
	_ = envx.LoadString("NODEMAN_APIGW_BK_TOKEN", &svc.APIGateway.BkToken)
	_ = envx.LoadString("NODEMAN_APIGW_ACCESS_TOKEN", &svc.APIGateway.AccessToken)
	if _, err := envx.LoadBool("NODEMAN_APIGW_TLS_SKIP_VERIFY", &svc.APIGateway.TLS.InsecureSkipVerify); err != nil {
		return err
	}
	_ = envx.LoadString("NODEMAN_APIGW_TLS_CERT", &svc.APIGateway.TLS.CertFile)
	_ = envx.LoadString("NODEMAN_APIGW_TLS_KEY", &svc.APIGateway.TLS.KeyFile)
	_ = envx.LoadString("NODEMAN_APIGW_TLS_CA", &svc.APIGateway.TLS.CAFile)
	_ = envx.LoadString("NODEMAN_APIGW_TLS_PASSWORD", &svc.APIGateway.TLS.Password)

	// http_server.
	_ = envx.LoadString("NODEMAN_HTTPSVR_BIND_IP", &svc.HTTPServer.BindIP)
	if _, err := envx.LoadInt("NODEMAN_HTTPSVR_PORT", &svc.HTTPServer.Port); err != nil {
		return err
	}

	// log.
	_ = envx.LoadString("NODEMAN_LOG_DIR", &svc.Log.Dir)
	_ = envx.LoadString("NODEMAN_LOG_LEVEL", &svc.Log.Level)
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
	configContent, err := os.ReadFile(path)
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

// EnvMustGet read env, panic if not set.
func EnvMustGet(key string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	panic(fmt.Sprintf("required environment variable %s unset", key))
}
