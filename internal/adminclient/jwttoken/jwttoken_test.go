/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package jwttoken

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadFileJWTKey(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "file_conf.yaml")
	configContent := []byte(`kms:
  enabled: false
basicServer:
  jwtServerConfig:
    symmetricKey: file-basic-key
downloadServer:
  jwtServerConfig:
    symmetricKey: file-download-key
`)
	require.NoError(t, os.WriteFile(configPath, configContent, 0o600))

	key, err := loadFileJWTKey(configPath, serverBasic)
	require.NoError(t, err)
	assert.Equal(t, "file-basic-key", key)

	// server names are validated before the loader runs; an unvalidated name yields
	// an empty key and no error.
	key, err = loadFileJWTKey(configPath, "unknown")
	require.NoError(t, err)
	assert.Empty(t, key)
}

func TestValidateServiceServerUnknownService(t *testing.T) {
	err := validateServiceServer("unknown", "basic")
	require.ErrorContains(t, err, "unknown service 'unknown'")
	require.ErrorContains(t, err, "application, backend, file")
}

func TestValidateServiceServerUnknownServer(t *testing.T) {
	err := validateServiceServer("file", "unknown")
	require.ErrorContains(t, err, "unknown server 'unknown' for service 'file'")
	require.ErrorContains(t, err, "admin, basic, download, export, info")
}
