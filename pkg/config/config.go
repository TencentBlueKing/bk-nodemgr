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
	"strings"
)

// Etcd the config of etcd.
type Etcd struct {
	Endpoints []string  `yaml:"endpoints" usage:"endpoints of etcd"`
	Username  string    `yaml:"username" usage:"username of etcd"`
	Password  string    `yaml:"password" usage:"password of etcd"`
	TLS       TLSConfig `yaml:"tls" usage:"tls of etcd"`
}

// Validate configures the config.
func (conf Etcd) Validate() error {
	if len(conf.Endpoints) == 0 {
		return errors.New("endpoints of etcd is empty")
	}

	if err := conf.TLS.Validate(); err != nil {
		return err
	}

	return nil
}

// RedisType defines the Redis deployment type.
type RedisType string

const (
	// RedisTypeStandalone standalone mode (single Redis instance).
	RedisTypeStandalone RedisType = "standalone"
	// RedisTypeSentinel sentinel mode (Redis Sentinel for high availability).
	RedisTypeSentinel RedisType = "sentinel"
	// RedisTypeCluster cluster mode (Redis Cluster for horizontal scaling).
	RedisTypeCluster RedisType = "cluster"
)

// Validate validates the Redis type.
func (redisType RedisType) Validate() error {
	switch redisType {
	case RedisTypeStandalone, RedisTypeSentinel, RedisTypeCluster:
		return nil
	default:
		return fmt.Errorf("invalid redis type: %s, must be one of: standalone, sentinel, cluster", redisType)
	}
}

// Redis the config of redis.
type Redis struct {
	// Type defines the Redis deployment type: standalone, sentinel, cluster.
	// Defaults to "standalone" if not specified.
	Type RedisType `yaml:"type" usage:"redis deployment type: standalone, sentinel, cluster"`
	// Addrs is the list of Redis addresses.
	// - standalone: ["host:port"] (single address)
	// - sentinel: ["sentinel1:26379", "sentinel2:26379"] (sentinel addresses)
	// - cluster: ["node1:6379", "node2:6379"] (cluster node addresses)
	Addrs []string `yaml:"addrs" usage:"redis addresses"`
	// Password is the password for Redis authentication.
	Password string `yaml:"password" usage:"password of redis"`
	// DB is the database index (standalone and sentinel mode only, ignored in cluster mode).
	DB int `yaml:"db" usage:"db of redis (standalone/sentinel mode only)"`
	// MasterName is the name of the master node (sentinel mode only).
	MasterName string `yaml:"masterName" usage:"master name (sentinel mode only)"`
	// TLS defines the TLS configuration.
	TLS TLSConfig `yaml:"tls" usage:"tls of redis"`

	TraceService `yaml:",inline"`
}

// GetType returns the Redis type, defaulting to standalone if not specified.
func (conf Redis) GetType() RedisType {
	if conf.Type == "" {
		return RedisTypeStandalone
	}

	return conf.Type
}

// Validate validates the Redis configuration.
func (conf Redis) Validate() error {
	if err := conf.Type.Validate(); err != nil {
		return err
	}

	if err := conf.validateAddrs(); err != nil {
		return err
	}

	if conf.Password == "" {
		return errors.New("password of redis is empty")
	}

	if err := conf.TLS.Validate(); err != nil {
		return err
	}

	return nil
}

// validateAddrs validates the Redis addresses configuration based on the deployment type.
func (conf Redis) validateAddrs() error {
	switch conf.GetType() {
	case RedisTypeSentinel:
		return conf.validateSentinelAddrs()
	case RedisTypeCluster:
		return conf.validateClusterAddrs()
	default:
		return conf.validateStandaloneAddrs()
	}
}

// validateSentinelAddrs validates sentinel mode addresses.
func (conf Redis) validateSentinelAddrs() error {
	if len(conf.Addrs) == 0 {
		return errors.New("addrs is required in sentinel mode")
	}

	if conf.MasterName == "" {
		return errors.New("masterName is required in sentinel mode")
	}

	return nil
}

// validateClusterAddrs validates cluster mode addresses.
func (conf Redis) validateClusterAddrs() error {
	if len(conf.Addrs) == 0 {
		return errors.New("addrs is required in cluster mode")
	}

	return nil
}

// validateStandaloneAddrs validates standalone mode addresses.
func (conf Redis) validateStandaloneAddrs() error {
	if len(conf.Addrs) == 0 {
		return errors.New("addrs is required (format: [\"host:port\"])")
	}

	if conf.DB < 0 {
		return fmt.Errorf("db of redis must be greater than or equal to 0, db(%d)", conf.DB)
	}

	return nil
}

// MongoDB the config of mongodb.
type MongoDB struct {
	AppName       string    `yaml:"appName" usage:"app name of mongodb"`
	Hosts         []string  `yaml:"hosts" usage:"hosts list of mongodb"`
	Username      string    `yaml:"username" usage:"user of mongodb"`
	Password      string    `yaml:"password" usage:"password of mongodb"`
	Database      string    `yaml:"database" usage:"database of mongodb"`
	AuthSource    string    `yaml:"authSource" usage:"auth source of mongodb"`
	AuthMechanism string    `yaml:"authMechanism" usage:"auth mechanism of mongodb"`
	ReplicaSet    string    `yaml:"replicaSet" usage:"replica set name of mongodb"`
	TLS           TLSConfig `yaml:"tls" usage:"tls of mongodb"`
	TraceService  `yaml:",inline"`
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

	if err := conf.TLS.Validate(); err != nil {
		return err
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
	BindIP                     string          `yaml:"bindIP"`
	BindIPV6                   string          `yaml:"bindIPV6"`
	AdvertiseIPV4              string          `yaml:"advertiseIPV4"`
	AdvertiseIPV6              string          `yaml:"advertiseIPV6"`
	Port                       int             `yaml:"port"`
	GracefulShutdownTimeoutSec int             `yaml:"gracefulShutdownTimeoutSec" usage:"graceful shutdown timeout in seconds"`
	AuthIdentity               AuthIdentity    `yaml:"authIdentity" usage:"identity of auth"`
	JWTServerConfig            JWTServerConfig `yaml:"jwtServerConfig" usage:"JWT configuration for authentication"`
	StaticDir                  string          `yaml:"staticDir"`
	TLSConfig                  TLSConfig       `yaml:"tls"`
	TraceService               `yaml:",inline"`
}

// Validate validates the config.
func (conf HTTPServer) Validate() error {
	if conf.BindIP == "" && conf.BindIPV6 == "" {
		return errors.New("failed to validate http server config: bindIP and bindIPV6 are both empty")
	}

	if conf.Port <= 0 {
		return fmt.Errorf("failed to validate http server config: port(%d) must be greater than 0", conf.Port)
	}

	if conf.GracefulShutdownTimeoutSec <= 0 {
		return fmt.Errorf("failed to validate http server config: gracefulShutdownTimeoutSec(%d) must be greater than 0",
			conf.GracefulShutdownTimeoutSec)
	}

	// Only validate JWT config for authentication methods that require JWT
	switch conf.AuthIdentity {
	case AuthIdentityAPIGW, AuthIdentityRestServer:
		// Validate JWT config based on authentication identity
		if err := conf.JWTServerConfig.Validate(); err != nil {
			return fmt.Errorf("failed to validate JWT config: %w", err)
		}

		if conf.AuthIdentity == AuthIdentityAPIGW && conf.JWTServerConfig.CryptoType != JWTCryptoTypeAsymmetric {
			return fmt.Errorf("crypto type of api-gateway's jwt server crypto type must be asymmetric, but got %s",
				conf.JWTServerConfig.CryptoType)
		}

		if conf.AuthIdentity == AuthIdentityRestServer && conf.JWTServerConfig.CryptoType != JWTCryptoTypeSymmetric {
			return fmt.Errorf("crypto type of rest-server's jwt server crypto type must be symmetric, but got %s",
				conf.JWTServerConfig.CryptoType)
		}

	case AuthIdentityBKLogin, AuthIdentityNone:
		// No JWT validation needed for these authentication methods
		return nil
	default:
		return fmt.Errorf("invalid auth identity: %s", conf.AuthIdentity)
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
	TraceService `yaml:",inline"`

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

// File the config of file.
type File struct {
	TraceService    `yaml:",inline"`
	JWTClientConfig JWTClientConfig `yaml:"jwtClientConfig" usage:"jwt config of api-gateway"`
}

// Validate validates the config.
func (file *File) Validate() error {
	if err := file.JWTClientConfig.Validate(); err != nil {
		return fmt.Errorf("jwt config of file is invalid: %s", err)
	}

	if file.JWTClientConfig.CryptoType != JWTCryptoTypeSymmetric {
		return errors.New("jwt config of file is invalid: only support symmetric algorithm")
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
	APIGatewayClient `yaml:",inline" usage:"api-gateway config of gse"`

	PluginSlotID    int    `yaml:"pluginSlotID"`
	PluginSlotToken string `yaml:"pluginSlotToken"`
}

// Validate validates the config.
func (conf GSE) Validate() error {
	if err := conf.APIGatewayClient.Validate(); err != nil {
		return fmt.Errorf("api-gateway config of gse is invalid: %s", err)
	}

	return nil
}

// UserManager the config of user manager.
type UserManager struct {
	APIGatewayClient `yaml:",inline" usage:"api-gateway config of user manager"`
}

// Validate validates the config.
func (conf UserManager) Validate() error {
	if err := conf.APIGatewayClient.Validate(); err != nil {
		return fmt.Errorf("api-gateway config of user manager is invalid: %s", err)
	}

	return nil
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
	Endpoint     string `yaml:"endpoint" usage:"endpoint of repo"`
	ProjectID    string `yaml:"projectID" usage:"projectID of repo"`
	RepoName     string `yaml:"repoName" usage:"name of repo"`
	AccessKey    string `yaml:"accessKey" usage:"access key of repo"`
	SecretKey    string `yaml:"secretKey" usage:"secret key of repo"`
	TraceService `yaml:",inline"`
}

// Validate validates the config.
func (repo Repo) Validate() error {
	if repo.Endpoint == "" {
		return errors.New("endpoint of repo is empty")
	}

	if repo.ProjectID == "" {
		return errors.New("projectID of repo is empty")
	}

	if repo.RepoName == "" {
		return errors.New("repo name of repo is empty")
	}

	if repo.AccessKey == "" {
		return errors.New("access key of repo is empty")
	}

	if repo.SecretKey == "" {
		return errors.New("secret key of repo is empty")
	}

	return nil
}

// TLSConfig defines tls related options.
type TLSConfig struct {
	InsecureSkipVerify bool   `yaml:"insecureSkipVerify"`
	VerifyClient       bool   `yaml:"verifyClient"`
	CAFile             string `yaml:"caFile"`
	CertFile           string `yaml:"certFile"`
	KeyFile            string `yaml:"keyFile"`
	Password           string `yaml:"password"`
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

	// case3: two-way authentication
	case len(conf.CertFile) > 0 && len(conf.KeyFile) > 0:
		return nil
	default:
		return fmt.Errorf("failed to validate tls config: invalid combination of tls config")
	}
}

// Workflow the config of workflow.
type Workflow struct {
	TraceService                   `yaml:",inline"`
	WorkerNum                      int `yaml:"workerNum" usage:"worker num of workflow"`
	GracefulShutdownTimeoutSeconds int `yaml:"gracefulShutdownTimeoutSeconds" usage:"graceful shutdown timeout in seconds of workflow"`
}

// Validate validates the config.
func (conf Workflow) Validate() error {
	if conf.WorkerNum <= 0 {
		return fmt.Errorf("worker num must be greater than 0, worker-num(%d)", conf.WorkerNum)
	}

	if conf.GracefulShutdownTimeoutSeconds <= 0 {
		return fmt.Errorf("graceful shutdown timeout seconds must be greater than 0, seconds(%d)",
			conf.GracefulShutdownTimeoutSeconds)
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

// Front the config of front.
type Front struct {
	// PasswordVault Options.
	PasswordVaultSwitch bool   `yaml:"passwordVaultSwitch" usage:"switch of password vault"`
	PasswordVaultName   string `yaml:"passwordVaultName" usage:"name of password vault"`

	// BKUserWebURL is the URL of bk user web service.
	BKUserWebURL string `yaml:"bkUserWebURL" usage:"bk user web url"`

	// BKDomain is the domain of bk platform.
	BKDomain string `yaml:"bkDomain" usage:"bk domain"`

	// BKDocsCenterURL is the URL of bk docs center.
	BKDocsCenterURL string `yaml:"bkDocsCenterURL" usage:"bk docs center url"`

	// BKAppNavOpenSourceURL is the URL of bk app nav open source.
	BKAppNavOpenSourceURL string `yaml:"bkAppNavOpenSourceURL" usage:"bk app nav open source url"`

	// WindowsWMIPortDefault is the default port for Windows WMI connection.
	WindowsWMIPortDefault int `yaml:"windowsWMIPortDefault" usage:"default port for Windows WMI connection"`

	// UnixSSHPortDefault is the default port for Unix-like OS (Linux, AIX, Darwin, etc.) SSH connection.
	UnixSSHPortDefault int `yaml:"unixSSHPortDefault" usage:"default port for Unix-like OS SSH connection"`

	// BKIamSaaSHost is the host of bk iam saas.
	BKIamSaaSHost string `yaml:"bkIamSaaSHost" usage:"bk iam saas host of application service"`

	// BKUserSaaSHost is the host of bk user saas.
	BKUserSaaSHost string `yaml:"bkUserSaaSHost" usage:"bk user saas host of application service"`
}

// Validate validates the config.
func (front *Front) Validate() error {
	return nil
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

// LoginAuthType defines the auth type of bklogin.
type LoginAuthType string

const (
	// LoginAuthTypeBKTicket BKTicket mode.
	LoginAuthTypeBKTicket LoginAuthType = "bk_ticket"

	// LoginAuthTypeBKToken BKToken mode.
	LoginAuthTypeBKToken LoginAuthType = "bk_token"
)

// Validate validates the auth type.
func (authType LoginAuthType) Validate() error {
	switch authType {
	case LoginAuthTypeBKTicket, LoginAuthTypeBKToken:
		return nil
	default:
		return fmt.Errorf("invalid login auth type (%s)", authType)
	}
}

// String returns the string of auth type.
func (authType LoginAuthType) String() string {
	return string(authType)
}

// BKLogin the config of bklogin.
type BKLogin struct {
	// LoginURL defines the login url of bklogin.
	LoginURL string `yaml:"loginURL" usage:"login url of bklogin"`

	// AuthType defines the auth type of bklogin, support 'bk_token' and 'bk_ticket'.
	AuthType LoginAuthType `yaml:"authType" usage:"auth type of bklogin, support 'bk_token' and 'bk_ticket'"`

	// TLS defines the tls config of bklogin.
	TLS TLSConfig `yaml:"tls" usage:"tls config of bklogin"`

	TraceService `yaml:",inline"`
}

// Validate validates the config.
func (bklogin *BKLogin) Validate() error {
	if bklogin.LoginURL == "" {
		return errors.New("login url of bkLogin is empty")
	}

	if err := bklogin.TLS.Validate(); err != nil {
		return fmt.Errorf("failed to validate bklogin config: %w", err)
	}

	if err := bklogin.AuthType.Validate(); err != nil {
		return fmt.Errorf("failed to validate bklogin config: %w", err)
	}

	return nil
}

// JWTCryptoType defines the JWT encryption type.
type JWTCryptoType string

const (
	// JWTCryptoTypeSymmetric symmetric encryption.
	JWTCryptoTypeSymmetric JWTCryptoType = "symmetric"
	// JWTCryptoTypeAsymmetric asymmetric encryption.
	JWTCryptoTypeAsymmetric JWTCryptoType = "asymmetric"
)

// Validate validates the JWT crypto type.
func (cryptoType JWTCryptoType) Validate() error {
	switch cryptoType {
	case JWTCryptoTypeSymmetric, JWTCryptoTypeAsymmetric:
		return nil
	default:
		return fmt.Errorf("invalid JWT crypto type: %s", cryptoType)
	}
}

// JWTServerConfig defines the unified JWT configuration supporting both symmetric and asymmetric encryption.
type JWTServerConfig struct {
	// CryptoType specifies which encryption method to use
	CryptoType JWTCryptoType `yaml:"cryptoType" usage:"JWT encryption type: symmetric or asymmetric"`
	// SymmetricKey is the secret key for symmetric encryption.
	SymmetricKey string `yaml:"symmetricKey" usage:"symmetric key for JWT signature algorithms"`
	// PublicKeyPem is the public key in PEM format for asymmetric encryption.
	PublicKeyPem string `yaml:"publicKeyPem" usage:"public key in PEM format for JWT signature algorithms"`
}

// Validate validates the JWT config based on the selected crypto type.
func (conf JWTServerConfig) Validate() error {
	if err := conf.CryptoType.Validate(); err != nil {
		return err
	}

	switch conf.CryptoType {
	case JWTCryptoTypeSymmetric:
		if conf.SymmetricKey == "" {
			return errors.New("symmetric key is required for symmetric encryption")
		}
	case JWTCryptoTypeAsymmetric:
		if conf.PublicKeyPem == "" {
			return errors.New("public key pem is required for asymmetric encryption")
		}
	}

	return nil
}

// JWTClientConfig defines the JWT configuration for client-side authentication.
// This is used when the service needs to act as a client and authenticate with other services.
type JWTClientConfig struct {
	// CryptoType specifies which encryption method to use
	CryptoType JWTCryptoType `yaml:"cryptoType" usage:"JWT encryption type: symmetric (HMAC) or asymmetric (RSA/ECDSA)"`
	// SymmetricKey is the secret key for symmetric encryption (HMAC algorithms)
	SymmetricKey string `yaml:"symmetricKey" usage:"symmetric key for JWT HMAC algorithms (HS256, HS384, HS512)"`
	// PrivateKeyPem is the private key in PEM format for asymmetric encryption (RSA/ECDSA algorithms)
	PrivateKeyPem string `yaml:"privateKeyPem" usage:"private key in PEM format for JWT RSA/ECDSA algorithms (RS256, ES256, etc.)"`
	// TokenExpirationHour defines the expiration time for JWT tokens
	TokenExpirationHour int64 `yaml:"tokenExpirationHour" usage:"token expiration duration (e.g., 1h, 24h)"`
}

// Validate validates the JWT client config based on the selected crypto type.
func (conf JWTClientConfig) Validate() error {
	if err := conf.CryptoType.Validate(); err != nil {
		return fmt.Errorf("invalid JWT crypto type: %w", err)
	}

	switch conf.CryptoType {
	case JWTCryptoTypeSymmetric:
		if conf.SymmetricKey == "" {
			return errors.New("symmetric key is required for symmetric encryption")
		}
	case JWTCryptoTypeAsymmetric:
		if conf.PrivateKeyPem == "" {
			return errors.New("private key pem is required for asymmetric encryption")
		}
	default:
		return fmt.Errorf("invalid JWT crypto type: %s", conf.CryptoType)
	}

	if conf.TokenExpirationHour == 0 {
		return errors.New("token expiration is required for JWT client configuration")
	}

	return nil
}

// PluginName the name of plugin.
type PluginName string

const (
	// PluginNameRelay relay plugin name.
	PluginNameRelay PluginName = "bk-nodemgr-relay"
)

// Validate validates the config.
func (name PluginName) Validate() error {
	switch name {
	case PluginNameRelay:
		return nil
	default:
		return fmt.Errorf("invalid plugin name, plugin-name(%s)", name)
	}
}

// TraceService defines the trace service.
type TraceService struct {
	// TraceServiceName is the trace service name.
	TraceServiceName string `yaml:"traceServiceName" usage:"trace service name"`

	// TraceSampleRate is the trace sample rate.
	TraceSampleRate float64 `yaml:"traceSampleRate" usage:"trace sample rate"`
}

const (
	defaultIAMV3SystemID     = "bk_nodemgr"
	defaultIAMV3CMDBSystemID = "bk_cmdb"
)

// IAMV3 the config of IAM v3 gateway config.
type IAMV3 struct {
	// Enable indicates whether IAM v3 is enabled.
	// When disabled, a no-op handler will be used and all permission checks will be skipped.
	Enable           bool `yaml:"enable" usage:"enable IAM v3 permission management"`
	APIGatewayClient `yaml:",inline" usage:"api-gateway config of IAM v3"`
	// SystemID is the system identifier registered in IAM.
	SystemID string `yaml:"systemID" usage:"system ID registered in IAM v3"`
	// CallbackPath is the callback path for IAM resource provider.
	// This field is required when IAM v3 is enabled.
	CallbackPath string `yaml:"callbackPath" usage:"callback path for IAM resource provider"`
	// CMDBSystemID is the system identifier registered in IAM for CMDB, used for fetching CMDB resources from IAM.
	CMDBSystemID string `yaml:"cmdbSystemID" usage:"system ID registered in IAM v3"`
}

// Downloader defines the shared remote package downloader configuration.
type Downloader struct {
	AllowHosts   []string `yaml:"allowHosts" usage:"download host allow list, exact hostname match, no wildcard or subdomain"`
	BlockHosts   []string `yaml:"blockHosts" usage:"download host block list"`
	MaxBytes     int64    `yaml:"maxBytes" usage:"max bytes of download file, default is 1 GiB"`
	TraceService `yaml:",inline"`
}

// Validate validates the downloader configuration.
func (conf Downloader) Validate() error {
	for _, host := range conf.AllowHosts {
		if strings.TrimSpace(host) == "" {
			return errors.New("allow host must not be empty")
		}
	}

	for _, host := range conf.BlockHosts {
		if strings.TrimSpace(host) == "" {
			return errors.New("block host must not be empty")
		}
	}

	return nil
}
