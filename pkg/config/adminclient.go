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

	"gopkg.in/yaml.v2"
)

// AdminClientDatabase the minimal config for adminclient database commands (check/proxy).
// It aggregates only the database sections shared by all service config files, so the same
// loader works against backend/application/file conf files.
type AdminClientDatabase struct {
	KMS     KMS     `yaml:"kms" usage:"kms secret config"`
	Etcd    Etcd    `yaml:"etcd" usage:"etcd config"`
	Redis   Redis   `yaml:"redis" usage:"redis config"`
	MongoDB MongoDB `yaml:"mongodb" usage:"mongodb config"`
}

// LoadFromFile loads config from file.
// When kms is enabled, database credentials are overridden from the kms global envelope.
// Validation is left to callers: each command validates only the database section it uses,
// because the raw config file may contain unreplaced rendering placeholders in unrelated fields.
func (conf *AdminClientDatabase) LoadFromFile(path string) error {
	if err := loadYAMLFile(path, conf); err != nil {
		return err
	}

	// load kms secret.
	if conf.KMS.Enabled {
		if err := conf.KMS.ParseAdminClientDatabase(conf); err != nil {
			return err
		}
	}

	return nil
}

// bk app credential environment variable names: the ops job layer (e.g. the
// apigw-sync job) injects the plaintext apigw-sync credential through these
// variables, the same deployment-layer channel as the KMS private key env.
const (
	bkAppCodeEnvName   = "BK_APP_CODE"
	bkAppSecretEnvName = "BK_APP_SECRET"
)

// AdminClientBKApp the minimal config for the adminclient bkapp command. It mirrors
// the backend config sections holding the named bk app credentials; with kms enabled,
// LoadFromFile overrides all of them with the default credential decrypted from the
// kms global envelope.
type AdminClientBKApp struct {
	KMS KMS `yaml:"kms" usage:"kms secret config"`

	// IAMV3/IAMV4 decode the appCode/appSecret of the backend iamV3/iamV4 sections;
	// the remaining APIGatewayClient fields are ignored.
	IAMV3 APIGatewayClient `yaml:"iamV3"`
	IAMV4 APIGatewayClient `yaml:"iamV4"`

	// Default is the apigw-sync credential, populated by LoadFromFile: decrypted
	// from the kms global envelope when kms is enabled, read from the
	// BK_APP_CODE/BK_APP_SECRET env (injected by the ops job layer) otherwise.
	Default APIGatewayClient `yaml:"-"`
}

// LoadFromFile loads the backend config file and resolves the named bk app
// credentials: kms enabled overrides every credential from the global envelope,
// plaintext keeps the yaml iamV3/iamV4 values and resolves Default from the env.
func (conf *AdminClientBKApp) LoadFromFile(path string) error {
	if err := loadYAMLFile(path, conf); err != nil {
		return err
	}

	if conf.KMS.Enabled {
		return conf.KMS.ParseAdminClientBKApp(conf)
	}

	// load plaintext apigw-sync credential from the env.
	conf.Default.AppCode = os.Getenv(bkAppCodeEnvName)
	conf.Default.AppSecret = os.Getenv(bkAppSecretEnvName)

	return nil
}

// AdminClientApplication mirrors the application service config sections relevant to
// the adminclient jwt_token command. Plaintext deployments read the jwt keys straight
// from the yaml server sections; with kms enabled, LoadFromFile overrides
// JWTServerConfig.SymmetricKey with the keys decrypted from the kms global envelope.
type AdminClientApplication struct {
	KMS KMS `yaml:"kms" usage:"kms secret config"`

	InfoServer  HTTPServer `yaml:"infoServer"`
	AdminServer HTTPServer `yaml:"adminServer"`
	BasicServer HTTPServer `yaml:"basicServer"`
}

// LoadFromFile loads the application config file and resolves the jwt server keys.
func (conf *AdminClientApplication) LoadFromFile(path string) error {
	if err := loadYAMLFile(path, conf); err != nil {
		return err
	}

	if conf.KMS.Enabled {
		return conf.KMS.ParseAdminClientApplicationJWT(conf)
	}

	return nil
}

// AdminClientBackend mirrors the backend service config sections relevant to the
// adminclient jwt_token command. See AdminClientApplication for the kms semantics.
type AdminClientBackend struct {
	KMS KMS `yaml:"kms" usage:"kms secret config"`

	InfoServer  HTTPServer `yaml:"infoServer"`
	AdminServer HTTPServer `yaml:"adminServer"`
	BasicServer HTTPServer `yaml:"basicServer"`
	// CallbackServer/ProxyServer are HTTPServer wrappers in BackendService, so plain
	// HTTPServer fields decode them compatibly.
	CallbackServer HTTPServer `yaml:"callbackServer"`
	ProxyServer    HTTPServer `yaml:"proxyServer"`
}

// LoadFromFile loads the backend config file and resolves the jwt server keys.
func (conf *AdminClientBackend) LoadFromFile(path string) error {
	if err := loadYAMLFile(path, conf); err != nil {
		return err
	}

	if conf.KMS.Enabled {
		return conf.KMS.ParseAdminClientBackendJWT(conf)
	}

	return nil
}

// AdminClientFile mirrors the file service config sections relevant to the adminclient
// jwt_token command. See AdminClientApplication for the kms semantics.
type AdminClientFile struct {
	KMS KMS `yaml:"kms" usage:"kms secret config"`

	InfoServer     HTTPServer `yaml:"infoServer"`
	AdminServer    HTTPServer `yaml:"adminServer"`
	BasicServer    HTTPServer `yaml:"basicServer"`
	DownloadServer HTTPServer `yaml:"downloadServer"`
	// ExportServer is an HTTPServer wrapper in FileService, so a plain HTTPServer
	// field decodes it compatibly.
	ExportServer HTTPServer `yaml:"exportServer"`
}

// LoadFromFile loads the file config file and resolves the jwt server keys.
func (conf *AdminClientFile) LoadFromFile(path string) error {
	if err := loadYAMLFile(path, conf); err != nil {
		return err
	}

	if conf.KMS.Enabled {
		return conf.KMS.ParseAdminClientFileJWT(conf)
	}

	return nil
}

// loadYAMLFile reads a yaml config file into v.
func loadYAMLFile(path string, v any) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	absPath = filepath.Clean(absPath)
	configContent, err := os.ReadFile(absPath)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(configContent, v)
}
