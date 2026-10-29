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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdminClientDatabase_LoadFromFile(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "backend_conf.yaml")
	configContent := []byte(`# unrelated fields may contain unreplaced rendering placeholders.
advertiseIPV4: "__ADVERTISE_IPV4__"
etcd:
  endpoints:
    - etcd-0.etcd-headless:2379
  username: root
  password: etcd-pass
  tls:
    caFile: /bk-nodemgr/cert/etcdca.crt
redis:
  type: sentinel
  addrs:
    - redis-0.redis-headless:26379
  password: redis-pass
  masterName: mymaster
mongodb:
  hosts:
    - mongodb-0.mongodb-headless:27017
  username: bk-nodemgr
  password: mongo-pass
  database: bk-nodemgr
  authSource: bk-nodemgr
`)
	require.NoError(t, os.WriteFile(configPath, configContent, 0o600))

	conf := &AdminClientDatabase{}
	require.NoError(t, conf.LoadFromFile(configPath))

	assert.False(t, conf.KMS.Enabled)
	assert.Equal(t, []string{"etcd-0.etcd-headless:2379"}, conf.Etcd.Endpoints)
	assert.Equal(t, "root", conf.Etcd.Username)
	assert.Equal(t, "etcd-pass", conf.Etcd.Password)
	assert.Equal(t, "/bk-nodemgr/cert/etcdca.crt", conf.Etcd.TLS.CAFile)
	assert.Equal(t, RedisTypeSentinel, conf.Redis.Type)
	assert.Equal(t, "redis-pass", conf.Redis.Password)
	assert.Equal(t, "mymaster", conf.Redis.MasterName)
	assert.Equal(t, []string{"mongodb-0.mongodb-headless:27017"}, conf.MongoDB.Hosts)
	assert.Equal(t, "bk-nodemgr", conf.MongoDB.Username)
	assert.Equal(t, "mongo-pass", conf.MongoDB.Password)
	assert.Equal(t, "bk-nodemgr", conf.MongoDB.Database)
}

func TestAdminClientDatabase_LoadFromFileNotFound(t *testing.T) {
	conf := &AdminClientDatabase{}
	err := conf.LoadFromFile(filepath.Join(t.TempDir(), "not-exist.yaml"))
	require.Error(t, err)
}

func TestAdminClientDatabase_LoadFromFileKMSMissingSecretFile(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "backend_conf.yaml")
	configContent := []byte(`kms:
  enabled: true
`)
	require.NoError(t, os.WriteFile(configPath, configContent, 0o600))

	conf := &AdminClientDatabase{}
	err := conf.LoadFromFile(configPath)
	require.ErrorContains(t, err, "globalSecretFile is empty")
}

func TestAdminClientBKApp_LoadFromFile(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "backend_conf.yaml")
	configContent := []byte(`# unrelated fields may contain unreplaced rendering placeholders.
advertiseIPV4: "__ADVERTISE_IPV4__"
kms:
  enabled: true
  globalSecretFile: /bk-nodemgr/secret/kms-global-envelope
`)
	require.NoError(t, os.WriteFile(configPath, configContent, 0o600))

	conf := &AdminClientBKApp{}
	// LoadFromFile decrypts inline; the missing envelope file surfaces its parsed path,
	// which also proves the kms section yaml fields were loaded.
	err := conf.LoadFromFile(configPath)
	require.ErrorContains(t, err, "/bk-nodemgr/secret/kms-global-envelope")

	assert.True(t, conf.KMS.Enabled)
	assert.Equal(t, "/bk-nodemgr/secret/kms-global-envelope", conf.KMS.GlobalSecretFile)
}

func TestAdminClientBKApp_LoadFromFileKMSDisabled(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "backend_conf.yaml")
	configContent := []byte(`kms:
  enabled: false
`)
	require.NoError(t, os.WriteFile(configPath, configContent, 0o600))

	conf := &AdminClientBKApp{}
	err := conf.LoadFromFile(configPath)
	require.ErrorContains(t, err, "kms is not enabled")
}

func TestAdminClientJWT_LoadFromFile(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "file_conf.yaml")
	configContent := []byte(`# unrelated fields may contain unreplaced rendering placeholders.
advertiseIPV4: "__ADVERTISE_IPV4__"
kms:
  enabled: true
  globalSecretFile: /bk-nodemgr/secret/kms-global-envelope
`)
	require.NoError(t, os.WriteFile(configPath, configContent, 0o600))

	conf := &AdminClientJWT{}
	// LoadFromFile decrypts inline; the missing envelope file surfaces its parsed path,
	// which also proves the kms section yaml fields were loaded.
	err := conf.LoadFromFile(configPath)
	require.ErrorContains(t, err, "/bk-nodemgr/secret/kms-global-envelope")

	assert.True(t, conf.KMS.Enabled)
	assert.Equal(t, "/bk-nodemgr/secret/kms-global-envelope", conf.KMS.GlobalSecretFile)
}

func TestAdminClientJWT_LoadFromFileKMSDisabled(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "file_conf.yaml")
	configContent := []byte(`kms:
  enabled: false
`)
	require.NoError(t, os.WriteFile(configPath, configContent, 0o600))

	conf := &AdminClientJWT{}
	err := conf.LoadFromFile(configPath)
	require.ErrorContains(t, err, "kms is not enabled")
}
