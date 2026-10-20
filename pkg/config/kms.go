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
	"fmt"
	"os"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/config/kms"
)

const (
	kmsPrivateKeyGlobalEnvName            = "BK_NODEMGR_KMS_PRIVATE_KEY_GLOBAL"
	kmsPrivateKeyModuleApplicationEnvName = "BK_NODEMGR_KMS_PRIVATE_KEY_MODULE_APPLICATION"
	kmsPrivateKeyModuleBackendEnvName     = "BK_NODEMGR_KMS_PRIVATE_KEY_MODULE_BACKEND"
	kmsPrivateKeyModuleFileEnvName        = "BK_NODEMGR_KMS_PRIVATE_KEY_MODULE_FILE"
)

// KMS defines the kms secret config.
type KMS struct {
	Enabled          bool   `yaml:"enabled" usage:"enable kms secret"`
	GlobalSecretFile string `yaml:"globalSecretFile" usage:"global secret file"`
	ModuleSecretFile string `yaml:"moduleSecretFile" usage:"module secret file"`

	param kms.EnvelopeLoadParam
}

func (k *KMS) load(moduleEnvName string) error {
	if err := k.loadGlobal(); err != nil {
		return err
	}

	if k.ModuleSecretFile == "" {
		return fmt.Errorf("kms moduleSecretFile is empty")
	}

	envelopeModule, err := os.ReadFile(k.ModuleSecretFile)
	if err != nil {
		return err
	}
	k.param.EnvelopeModule = string(envelopeModule)

	// get private key from env.
	k.param.PrivateKeyModule = os.Getenv(moduleEnvName)

	return nil
}

// loadGlobal loads only the global envelope file and the global private key.
func (k *KMS) loadGlobal() error {
	if k.GlobalSecretFile == "" {
		return fmt.Errorf("kms globalSecretFile is empty")
	}

	envelopeGlobal, err := os.ReadFile(k.GlobalSecretFile)
	if err != nil {
		return err
	}
	k.param.EnvelopeGlobal = string(envelopeGlobal)

	// get private key from env.
	k.param.PrivateKeyGlobal = os.Getenv(kmsPrivateKeyGlobalEnvName)

	return nil
}

// ParseAdminClientDatabase parses the adminclient database config.
// Database credentials (etcd/redis/mongodb) all come from the global envelope,
// so only the global private key is required.
func (k *KMS) ParseAdminClientDatabase(conf *AdminClientDatabase) error {
	if !k.Enabled {
		return nil
	}

	if err := k.loadGlobal(); err != nil {
		return err
	}

	dataGlobal, err := kms.LoadGlobalEnvelope(k.param)
	if err != nil {
		return err
	}

	// etcd.
	// - decoding etcd tls.
	etcdTLS, err := parseTLSConfigFromBase64(
		dataGlobal.Etcd.Default.TLSCABase64,
		dataGlobal.Etcd.Default.TLSCertBase64,
		dataGlobal.Etcd.Default.TLSKeyBase64,
		dataGlobal.Etcd.Default.TLSPassword,
	)
	if err != nil {
		return fmt.Errorf("failed to parse etcd tls config: %w", err)
	}
	conf.Etcd.Password = dataGlobal.Etcd.Default.Password
	conf.Etcd.TLS.CA = etcdTLS.CA
	conf.Etcd.TLS.Cert = etcdTLS.Cert
	conf.Etcd.TLS.Key = etcdTLS.Key
	conf.Etcd.TLS.Password = etcdTLS.Password

	// redis.
	conf.Redis.Password = dataGlobal.Redis.Default.Password

	// mongodb.
	conf.MongoDB.Username = dataGlobal.MongoDB.Default.Username
	conf.MongoDB.Password = dataGlobal.MongoDB.Default.Password

	return nil
}

// ParseAdminClientBKApp parses the adminclient bkapp config.
// The bk app credentials live only in the global envelope, so only the global private
// key is required. Unlike the other Parse functions, kms disabled is an error here:
// plaintext deployments read app credentials from values and never call the bkapp command.
func (k *KMS) ParseAdminClientBKApp(conf *AdminClientBKApp) error {
	if !k.Enabled {
		return fmt.Errorf("kms is not enabled, bk app code/secret are only available via kms")
	}

	if err := k.loadGlobal(); err != nil {
		return err
	}

	dataGlobal, err := kms.LoadGlobalEnvelope(k.param)
	if err != nil {
		return err
	}

	conf.AppCode = dataGlobal.BKAppIDSecret.Default.AppCode
	conf.AppSecret = dataGlobal.BKAppIDSecret.Default.AppSecret

	return nil
}

// ParseApplicationServiceConfig parses the application service config.
func (k *KMS) ParseApplicationServiceConfig(service *ApplicationService) error {
	if !k.Enabled {
		return nil
	}

	if err := k.load(kmsPrivateKeyModuleApplicationEnvName); err != nil {
		return err
	}

	dataGlobal, _, err := kms.LoadApplicationEnvelopes(k.param)
	if err != nil {
		return err
	}

	// etcd.
	// - decoding etcd tls.
	etcdTLS, err := parseTLSConfigFromBase64(
		dataGlobal.Etcd.Default.TLSCABase64,
		dataGlobal.Etcd.Default.TLSCertBase64,
		dataGlobal.Etcd.Default.TLSKeyBase64,
		dataGlobal.Etcd.Default.TLSPassword,
	)
	if err != nil {
		return fmt.Errorf("failed to parse etcd tls config: %w", err)
	}
	service.Etcd.Password = dataGlobal.Etcd.Default.Password
	service.Etcd.TLS.CA = etcdTLS.CA
	service.Etcd.TLS.Cert = etcdTLS.Cert
	service.Etcd.TLS.Key = etcdTLS.Key
	service.Etcd.TLS.Password = etcdTLS.Password

	// mongodb.
	service.MongoDB.Username = dataGlobal.MongoDB.Default.Username
	service.MongoDB.Password = dataGlobal.MongoDB.Default.Password

	// services.
	// - info
	service.InfoServer.JWTServerConfig.SymmetricKey = dataGlobal.JwtServerConfig.ApplicationInfo.SymmetricKey
	// - admin
	service.AdminServer.JWTServerConfig.SymmetricKey = dataGlobal.JwtServerConfig.ApplicationAdmin.SymmetricKey
	// - basic
	service.BasicServer.JWTServerConfig.SymmetricKey = dataGlobal.JwtServerConfig.ApplicationBasic.SymmetricKey
	// - file client
	service.File.JWTClientConfig.SymmetricKey = dataGlobal.JwtServerConfig.FileBasic.SymmetricKey

	// bk-app.
	// - backend
	service.Backend.AppCode = dataGlobal.BKAppIDSecret.Default.AppCode
	service.Backend.AppSecret = dataGlobal.BKAppIDSecret.Default.AppSecret
	// - user manager
	service.UserManager.AppCode = dataGlobal.BKAppIDSecret.Default.AppCode
	service.UserManager.AppSecret = dataGlobal.BKAppIDSecret.Default.AppSecret
	// - bk saas
	service.BKSaas.BKLogin.AppCode = dataGlobal.BKAppIDSecret.Default.AppCode
	service.BKSaas.BKLogin.AppSecret = dataGlobal.BKAppIDSecret.Default.AppSecret
	// - iam v3
	service.IAMV3.AppCode = dataGlobal.BKAppIDSecret.Default.AppCode
	service.IAMV3.AppSecret = dataGlobal.BKAppIDSecret.Default.AppSecret
	// - notice
	service.Notice.AppCode = dataGlobal.BKAppIDSecret.Default.AppCode
	service.Notice.AppSecret = dataGlobal.BKAppIDSecret.Default.AppSecret

	return nil
}

// ParseBackendServiceConfig parses the backend service config.
func (k *KMS) ParseBackendServiceConfig(service *BackendService) error {
	if !k.Enabled {
		return nil
	}

	if err := k.load(kmsPrivateKeyModuleBackendEnvName); err != nil {
		return err
	}

	dataGlobal, dataModule, err := kms.LoadBackendEnvelopes(k.param)
	if err != nil {
		return err
	}

	// etcd.
	// - decoding etcd tls.
	etcdTLS, err := parseTLSConfigFromBase64(
		dataGlobal.Etcd.Default.TLSCABase64,
		dataGlobal.Etcd.Default.TLSCertBase64,
		dataGlobal.Etcd.Default.TLSKeyBase64,
		dataGlobal.Etcd.Default.TLSPassword,
	)
	if err != nil {
		return fmt.Errorf("failed to parse etcd tls config: %w", err)
	}
	service.Etcd.Password = dataGlobal.Etcd.Default.Password
	service.Etcd.TLS.CA = etcdTLS.CA
	service.Etcd.TLS.Cert = etcdTLS.Cert
	service.Etcd.TLS.Key = etcdTLS.Key
	service.Etcd.TLS.Password = etcdTLS.Password

	// mongodb.
	service.MongoDB.Username = dataGlobal.MongoDB.Default.Username
	service.MongoDB.Password = dataGlobal.MongoDB.Default.Password

	// redis.
	service.Redis.Password = dataGlobal.Redis.Default.Password

	// services.
	// - info
	service.InfoServer.JWTServerConfig.SymmetricKey = dataGlobal.JwtServerConfig.BackendInfo.SymmetricKey
	// - admin
	service.AdminServer.JWTServerConfig.SymmetricKey = dataGlobal.JwtServerConfig.BackendAdmin.SymmetricKey
	// - basic
	service.BasicServer.JWTServerConfig.SymmetricKey = dataGlobal.JwtServerConfig.BackendBasic.SymmetricKey
	service.BasicServer.JWTServerConfig.PublicKeyPem = dataGlobal.JwtServerConfig.BackendBasic.PublicKeyPEMBase64
	// - callback
	service.CallbackServer.JWTServerConfig.SymmetricKey = dataGlobal.JwtServerConfig.BackendCallback.SymmetricKey
	// - proxy
	service.ProxyServer.JWTServerConfig.SymmetricKey = dataGlobal.JwtServerConfig.BackendProxy.SymmetricKey
	// - file client
	service.File.JWTClientConfig.SymmetricKey = dataGlobal.JwtServerConfig.FileBasic.SymmetricKey

	// bk-app.
	// - cmdb
	service.CMDB.AppCode = dataGlobal.BKAppIDSecret.Default.AppCode
	service.CMDB.AppSecret = dataGlobal.BKAppIDSecret.Default.AppSecret
	// - gse
	service.GSE.AppCode = dataGlobal.BKAppIDSecret.Default.AppCode
	service.GSE.AppSecret = dataGlobal.BKAppIDSecret.Default.AppSecret
	// - user manager
	service.UserManager.AppCode = dataGlobal.BKAppIDSecret.Default.AppCode
	service.UserManager.AppSecret = dataGlobal.BKAppIDSecret.Default.AppSecret
	// - monitor
	service.Monitor.AppCode = dataGlobal.BKAppIDSecret.Default.AppCode
	service.Monitor.AppSecret = dataGlobal.BKAppIDSecret.Default.AppSecret
	// - iamv3
	service.IAMV3.AppCode = dataGlobal.BKAppIDSecret.Default.AppCode
	service.IAMV3.AppSecret = dataGlobal.BKAppIDSecret.Default.AppSecret
	// - iamv4
	service.IAMV4.AppCode = dataGlobal.BKAppIDSecret.Default.AppCode
	service.IAMV4.AppSecret = dataGlobal.BKAppIDSecret.Default.AppSecret
	// - iegtjj
	service.CreditVault.HostCreditVault.IEGTJJ.AppCode = dataGlobal.BKAppIDSecret.Default.AppCode
	service.CreditVault.HostCreditVault.IEGTJJ.AppSecret = dataGlobal.BKAppIDSecret.Default.AppSecret

	// encryption.
	service.EncryptKey = dataModule.Encryption.EncryptKey

	// gse.
	service.GSE.PluginSlotID = dataModule.GSE.PluginSlotID
	service.GSE.PluginSlotToken = dataModule.GSE.PluginSlotToken

	// credit vault.
	service.CreditVault.HostCreditVault.IEGTJJ.Key = dataModule.CreditVault.IEGTJJ.Key
	service.CreditVault.HostCreditVault.IEGTJJ.SecretKey = dataModule.CreditVault.IEGTJJ.SecretKey

	return nil
}

// ParseFileServiceConfig parses the file service config.
func (k *KMS) ParseFileServiceConfig(service *FileService) error {
	if !k.Enabled {
		return nil
	}

	if err := k.load(kmsPrivateKeyModuleFileEnvName); err != nil {
		return err
	}

	dataGlobal, dataModule, err := kms.LoadFileEnvelopes(k.param)
	if err != nil {
		return err
	}

	// etcd.
	// - decoding etcd tls.
	etcdTLS, err := parseTLSConfigFromBase64(
		dataGlobal.Etcd.Default.TLSCABase64,
		dataGlobal.Etcd.Default.TLSCertBase64,
		dataGlobal.Etcd.Default.TLSKeyBase64,
		dataGlobal.Etcd.Default.TLSPassword,
	)
	if err != nil {
		return fmt.Errorf("failed to parse etcd tls config: %w", err)
	}
	service.Etcd.Password = dataGlobal.Etcd.Default.Password
	service.Etcd.TLS.CA = etcdTLS.CA
	service.Etcd.TLS.Cert = etcdTLS.Cert
	service.Etcd.TLS.Key = etcdTLS.Key
	service.Etcd.TLS.Password = etcdTLS.Password

	// mongodb.
	service.MongoDB.Username = dataGlobal.MongoDB.Default.Username
	service.MongoDB.Password = dataGlobal.MongoDB.Default.Password

	// services.
	// - info
	service.InfoServer.JWTServerConfig.SymmetricKey = dataGlobal.JwtServerConfig.FileInfo.SymmetricKey
	// - admin
	service.AdminServer.JWTServerConfig.SymmetricKey = dataGlobal.JwtServerConfig.FileAdmin.SymmetricKey
	// - basic
	service.BasicServer.JWTServerConfig.SymmetricKey = dataGlobal.JwtServerConfig.FileBasic.SymmetricKey
	// - download
	service.DownloadServer.JWTServerConfig.SymmetricKey = dataGlobal.JwtServerConfig.FileDownload.SymmetricKey
	// - export
	service.ExportServer.JWTServerConfig.SymmetricKey = dataGlobal.JwtServerConfig.FileExport.SymmetricKey

	// bk-app.
	// - gse
	service.GSE.AppCode = dataGlobal.BKAppIDSecret.Default.AppCode
	service.GSE.AppSecret = dataGlobal.BKAppIDSecret.Default.AppSecret

	// bkrepo.
	service.Repo.AccessKey = dataModule.BKRepo.Default.AccessKey
	service.Repo.SecretKey = dataModule.BKRepo.Default.SecretKey

	return nil
}

func parseTLSConfigFromBase64(caBase64, certBase64, keyBase64, password string) (*TLSConfig, error) {
	// decoding base64.
	tlsCA, err := base64.StdEncoding.DecodeString(caBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode tls ca: %w", err)
	}
	tlsCert, err := base64.StdEncoding.DecodeString(certBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode tls cert: %w", err)
	}
	tlsKey, err := base64.StdEncoding.DecodeString(keyBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode tls key: %w", err)
	}

	conf := &TLSConfig{
		CA:       string(tlsCA),
		Cert:     string(tlsCert),
		Key:      string(tlsKey),
		Password: password,
	}

	if err = conf.Validate(); err != nil {
		return nil, fmt.Errorf("invalid tls config: %w", err)
	}

	return conf, nil
}
