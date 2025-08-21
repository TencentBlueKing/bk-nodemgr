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
	"errors"
	"fmt"
)

// Etcd the config of etcd.
type Etcd struct {
	Endpoints []string `yaml:"endpoints" usage:"endpoints of etcd"`
	Username  string   `yaml:"username" usage:"username of etcd"`
	Password  string   `yaml:"password" usage:"password of etcd"`
	Cert      string   `yaml:"cert" usage:"cert file of etcd"`
	Key       string   `yaml:"key" usage:"key file for etcd"`
	Ca        string   `yaml:"ca" usage:"ca file for etcd"`
}

// Validate configures the config.
func (conf Etcd) Validate() error {
	if len(conf.Endpoints) == 0 {
		return errors.New("endpoints of etcd is empty")
	}

	if (len(conf.Cert) == 0) != (len(conf.Key) == 0) {
		return errors.New("cert file and key file must be both empty or both not empty")
	}

	return nil
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
	if conf.Host == "" {
		return errors.New("host of redis is empty")
	}

	if conf.Port <= 0 {
		return fmt.Errorf("port of redis must be greater than 0, port(%d)", conf.Port)
	}

	if conf.DB < 0 {
		return fmt.Errorf("db of redis must be greater than or equal to 0, db(%d)", conf.DB)
	}

	if conf.Password == "" {
		return errors.New("password of redis is empty")
	}

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

// Validate configures the config.
func (conf MongoDB) Validate() error {
	if len(conf.Hosts) == 0 {
		return errors.New("hosts of mongodb is empty")
	}

	if conf.Username == "" {
		return errors.New("username of mongodb is empty")
	}

	if conf.Password == "" {
		return errors.New("password of mongodb is empty")
	}

	if conf.Database == "" {
		return errors.New("database of mongodb is empty")
	}

	return nil
}

// LogLevel the log level of service.
type LogLevel string

const (
	// LogLevelDebug debug level.
	LogLevelDebug LogLevel = "DEBUG"

	// LogLevelInfo info level.
	LogLevelInfo LogLevel = "INFO"

	// LogLevelWarn warn level.
	LogLevelWarn LogLevel = "WARN"

	// LogLevelError error level.
	LogLevelError LogLevel = "ERROR"
)

// Validate validates the log level.
func (logLevel LogLevel) Validate() error {
	switch logLevel {
	case LogLevelDebug, LogLevelInfo, LogLevelWarn, LogLevelError:
		return nil
	default:
		return fmt.Errorf("invalid log level, level(%s)", logLevel)
	}
}

// Log the config of log.
type Log struct {
	Dir          string   `yaml:"dir" usage:"log dir of server"`
	MaxSizeMB    int      `yaml:"maxSizeMB" usage:"max size in MBytes of single log file"`
	MaxNum       int      `yaml:"maxNum" usage:"max number of log files"`
	Level        LogLevel `yaml:"level" usage:"log level of server. DEBUG, INFO, WARN, ERROR"`
	ToStdErr     bool     `yaml:"toStderr" usage:"log to stderr instead of files"`
	AlsoToStdErr bool     `yaml:"alsoToStderr" usage:"log to stderr in addition to files"`
}

// Validate validates the config.
func (conf Log) Validate() error {
	if conf.Dir == "" {
		return errors.New("log dir is empty")
	}

	if conf.MaxSizeMB <= 0 {
		return fmt.Errorf("max size of log file must be greater than 0, max-size(%d)", conf.MaxSizeMB)
	}

	if conf.MaxNum <= 0 {
		return fmt.Errorf("max number of log files must be greater than 0, max-num(%d)", conf.MaxNum)
	}

	if err := conf.Level.Validate(); err != nil {
		return err
	}

	return nil
}

// AuthIdentity the identity of auth.
type AuthIdentity string

const (
	// AuthIdentityAPIGW api-gateway identity.
	AuthIdentityAPIGW AuthIdentity = "api-gateway"
	// AuthIdentityBKLogin bk-login identity.
	AuthIdentityBKLogin AuthIdentity = "bk-login"
	// AuthIdentityRestServer rest server identity.
	AuthIdentityRestServer AuthIdentity = "rest-server"
	// AuthIdentityNone none identity.
	AuthIdentityNone AuthIdentity = "none"
)

// Validate validates the auth identity.
func (authIdentity AuthIdentity) Validate() error {
	switch authIdentity {
	case AuthIdentityAPIGW, AuthIdentityBKLogin, AuthIdentityRestServer, AuthIdentityNone:
		return nil
	default:
		return fmt.Errorf("invalid auth identity, identity(%s)", authIdentity)
	}
}

// HTTPServer the config of http service.
type HTTPServer struct {
	BindIP        string       `yaml:"bindIP"`
	AdvertiseIPV4 string       `yaml:"advertiseIPV4"`
	AdvertiseIPV6 string       `yaml:"advertiseIPV6"`
	Port          int          `yaml:"port"`
	AuthIdentity  AuthIdentity `yaml:"authIdentity" usage:"identity of auth"`
	StaticDir     string       `yaml:"staticDir"`
}

// Validate validates the config.
func (conf HTTPServer) Validate() error {
	if conf.BindIP == "" {
		return errors.New("failed to validate http server config: bindIP is empty")
	}

	if conf.Port <= 0 {
		return fmt.Errorf("failed to validate http server config: port(%d) must be greater than 0", conf.Port)
	}

	return nil
}

// CallbackServer the config of callback service.
type CallbackServer struct {
	HTTPServer `yaml:",inline"`
}

// ProxyServer the config of proxy service.
type ProxyServer struct {
	HTTPServer `yaml:",inline"`
}

// APIGatewayClient the config of api-gateway.
type APIGatewayClient struct {
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
	// AccessToken is the BlueKing access token of nodeman to request api gateway.
	AccessToken string `yaml:"accessToken"`
	// TLS defines the tls config of api-gateway.
	TLS TLSConfig `yaml:"tls" usage:"tls config of api-gateway"`
}

// Validate validates the config.
func (conf APIGatewayClient) Validate() error {
	if len(conf.Endpoints) == 0 {
		return errors.New("endpoints of api-gateway is empty")
	}

	if conf.AppCode == "" {
		return errors.New("app code of api-gateway is empty")
	}

	if conf.AppSecret == "" {
		return errors.New("app secret of api-gateway is empty")
	}

	if conf.AuthMode == "" {
		return errors.New("auth mode of api-gateway is empty")
	}

	if err := conf.TLS.Validate(); err != nil {
		return fmt.Errorf("tls config of api-gateway is invalid: %s", err)
	}

	return nil
}

// CMDB the config of cmdb.
type CMDB struct {
	SupplierAccount  string `yaml:"supplierAccount" usage:"cmdb api request parameter"`
	APIGatewayClient `yaml:",inline" usage:"api-gateway config of cmdb"`
}

// Validate validates the config.
func (conf CMDB) Validate() error {
	if conf.SupplierAccount == "" {
		return errors.New("supplier account is empty")
	}

	if err := conf.APIGatewayClient.Validate(); err != nil {
		return fmt.Errorf("api-gateway config of cmdb is invalid: %s", err)
	}

	return nil
}

// GSE the config of gse.
type GSE struct {
	APIGatewayClient `yaml:",inline" usage:"api-gateway config of cmdb"`

	PluginSlotID    int    `yaml:"pluginSlotID"`
	PluginSlotToken string `yaml:"pluginSlotToken"`
}

// CreditVault the config of credit vault.
type CreditVault struct {
	HostCreditVault HostCreditVault `yaml:"hostCreditVault" usage:"host credit vault config of backend service"`
}

// Validate validates the config.
func (vault CreditVault) Validate() error {
	if err := vault.HostCreditVault.Validate(); err != nil {
		return fmt.Errorf("host credit vault config of backend service is invalid: %s", err)
	}

	return nil
}

// HostCreditVault the config of host credit vault.
type HostCreditVault struct {
	Enable bool   `yaml:"enable" usage:"enable credit vault"`
	Type   string `yaml:"type" usage:"type of credit vault"`
	IEGTJJ IEGTJJ `yaml:"iegtjj" usage:"ieg tjj config of backend service"`
}

// Validate validates the config.
func (vault HostCreditVault) Validate() error {
	if !vault.Enable {
		return nil
	}

	switch vault.Type {
	case "iegtjj":
		return vault.IEGTJJ.Validate()
	default:
		return errors.New("invalid type of host credit vault")
	}
}

// IEGTJJ the config of iegtjj.
type IEGTJJ struct {
	APIGatewayClient `yaml:",inline" usage:"api-gateway config of cmdb"`
	Key              string `yaml:"key" usage:"key of iegtjj"`
	SecretKey        string `yaml:"secretKey" usage:"secret key of iegtjj"`
}

// Validate validates the config.
func (iegtjj IEGTJJ) Validate() error {
	if iegtjj.Key == "" {
		return errors.New("key of iegtjj is empty")
	}

	if iegtjj.SecretKey == "" {
		return errors.New("secret key of iegtjj is empty")
	}

	if err := iegtjj.APIGatewayClient.Validate(); err != nil {
		return fmt.Errorf("api-gateway config of iegtjj is invalid: %s", err)
	}

	return nil
}

// Repo the config of repo.
type Repo struct {
	Endpoint  string `yaml:"endpoint" usage:"endpoint of repo"`
	ProjectID string `yaml:"projectID" usage:"projectID of repo"`
	RepoName  string `yaml:"repoName" usage:"name of repo"`
	AccessKey string `yaml:"accessKey" usage:"access key of repo"`
	SecretKey string `yaml:"secretKey" usage:"secret key of repo"`
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

// Validate validates the config.
func (conf TLSConfig) Validate() error {
	// certificate and key must be paired.
	if (len(conf.CertFile) == 0) != (len(conf.KeyFile) == 0) {
		return errors.New("cert file and key file must be both provided or both empty")
	}

	// the password must be present at the same time as the key file.
	if len(conf.Password) > 0 && len(conf.KeyFile) == 0 {
		return errors.New("password provided but no key file specified")
	}

	switch {
	// case1: unused tls config.
	case len(conf.CertFile) == 0 && len(conf.KeyFile) == 0 &&
		len(conf.CAFile) == 0 && len(conf.Password) == 0:
		return nil

	// case2: one-way authentication
	case len(conf.CertFile) == 0 && len(conf.CAFile) > 0:
		return nil

	// case2: two-way authentication
	case len(conf.CertFile) > 0 && len(conf.KeyFile) > 0:
		return nil
	default:
		return fmt.Errorf("failed to validate tls config: invalid combination of tls config")
	}
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

// Validate validates the config.
func (conf RunMode) Validate() error {
	switch conf {
	case RunModeRelease, RunModeDebug:
		return nil
	default:
		return fmt.Errorf("invalid run mode: %s", conf)
	}
}

// System the info of deploy info.
type System struct {
	Env     string `yaml:"env" usage:"env of system"`
	Edition string `yaml:"edition" usage:"edition of system"`
}

// Validate validates the config.
func (conf System) Validate() error {
	if conf.Env == "" {
		return errors.New("env is empty")
	}

	if conf.Edition == "" {
		return errors.New("edition is empty")
	}

	return nil
}

// GSEPlugin the config of gse agent plugin.
type GSEPlugin struct {
	PidFile                 string `yaml:"pidFile" usage:"pid file to save pid"`
	MessageDomainSocketPath string `yaml:"messageDomainSocketPath" usage:"message domain socket path of gse agent plugin"`
	MessageLocalSocketPort  int    `yaml:"messageLocalSocketPort" usage:"message local socket port of gse agent plugin"`
}

// BKPaaS the config of bk PaaS.
type BKPaaS struct {
	AnalysisScript string `yaml:"analysisScript" usage:"analysis script of bk PaaS"`
}

// Validate validates the config.
func (paas *BKPaaS) Validate() error {
	return nil
}

// BKSaas the config of bk SaaS.
type BKSaas struct {
	BKLogin BKLogin `yaml:"bkLogin" usage:"bklogin config of bk SaaS"`
}

// Validate validates the config.
func (saas *BKSaas) Validate() error {
	if err := saas.BKLogin.Validate(); err != nil {
		return fmt.Errorf("failed to validate bkSaaS config: %w", err)
	}

	return nil
}

// BKLogin the config of bklogin.
type BKLogin struct {
	// LoginURL defines the login url of bklogin.
	LoginURL string `yaml:"loginURL" usage:"login url of bklogin"`

	// TLS defines the tls config of bklogin.
	TLS TLSConfig `yaml:"tls" usage:"tls config of bklogin"`
}

// Validate validates the config.
func (bklogin *BKLogin) Validate() error {
	if bklogin.LoginURL == "" {
		return errors.New("login url of bkLogin is empty")
	}

	if err := bklogin.TLS.Validate(); err != nil {
		return fmt.Errorf("failed to validate bklogin config: %w", err)
	}

	return nil
}

// APIGateWayServer defines the api gateway config of backend service.
type APIGateWayServer struct {
	PublickeyPem string `yaml:"publickeyPem" usage:"publickey pem of api gateway, which is saved in base64 format"`
}

// Validate validates the config.
func (conf *APIGateWayServer) Validate() error {
	if conf.PublickeyPem == "" {
		return errors.New("publickey pem of api gateway is empty")
	}

	return nil
}
