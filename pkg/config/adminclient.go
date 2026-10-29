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

// AdminClientBKApp the minimal config for the adminclient bkapp command: only the kms
// section is needed, since the bk app code/secret lives in the kms global envelope.
type AdminClientBKApp struct {
	KMS KMS `yaml:"kms" usage:"kms secret config"`

	// AppCode/AppSecret hold the decrypted bk app credentials, populated by LoadFromFile.
	AppCode   string `yaml:"-"`
	AppSecret string `yaml:"-"`
}

// LoadFromFile loads the kms section from a service config file and decrypts the bk app
// code/secret from the kms global envelope into AppCode/AppSecret. Unlike the database
// commands, kms is mandatory here: plaintext deployments read app credentials from values
// directly and never call this command, so a disabled kms section is an error.
func (conf *AdminClientBKApp) LoadFromFile(path string) error {
	if err := loadYAMLFile(path, conf); err != nil {
		return err
	}

	// load kms secret.
	if err := conf.KMS.ParseAdminClientBKApp(conf); err != nil {
		return err
	}

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
