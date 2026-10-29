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

// AdminClientJWT the minimal config for the adminclient jwt_token command: only the kms
// section is needed, since the jwt server keys live in the kms global envelope.
type AdminClientJWT struct {
	KMS KMS `yaml:"kms" usage:"kms secret config"`

	// The per-server jwt configs are populated by LoadFromFile from the kms global
	// envelope; only SymmetricKey is set, because the jwt_token command signs tokens
	// with the symmetric key verified by the rest-server auth identity.
	ApplicationInfo  JWTServerConfig `yaml:"-"`
	ApplicationAdmin JWTServerConfig `yaml:"-"`
	ApplicationBasic JWTServerConfig `yaml:"-"`
	BackendInfo      JWTServerConfig `yaml:"-"`
	BackendAdmin     JWTServerConfig `yaml:"-"`
	BackendBasic     JWTServerConfig `yaml:"-"`
	BackendCallback  JWTServerConfig `yaml:"-"`
	BackendProxy     JWTServerConfig `yaml:"-"`
	FileInfo         JWTServerConfig `yaml:"-"`
	FileAdmin        JWTServerConfig `yaml:"-"`
	FileBasic        JWTServerConfig `yaml:"-"`
	FileDownload     JWTServerConfig `yaml:"-"`
	FileExport       JWTServerConfig `yaml:"-"`
}

// LoadFromFile loads the kms section from a service config file and decrypts the jwt
// server keys from the kms global envelope into JWT. Like the bkapp command, kms is
// mandatory here: plaintext deployments read jwt keys from values and never call this
// command, so a disabled kms section is an error.
func (conf *AdminClientJWT) LoadFromFile(path string) error {
	if err := loadYAMLFile(path, conf); err != nil {
		return err
	}

	// load kms secret.
	if err := conf.KMS.ParseAdminClientJWT(conf); err != nil {
		return err
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
