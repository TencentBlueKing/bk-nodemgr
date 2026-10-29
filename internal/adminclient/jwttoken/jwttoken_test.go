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
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPickJWTKey(t *testing.T) {
	pick, err := pickJWTKey("file", "basic")
	require.NoError(t, err)

	conf := config.AdminClientJWT{
		FileBasic: config.JWTServerConfig{SymmetricKey: "file-basic-key"},
	}
	assert.Equal(t, "file-basic-key", pick(conf))
}

func TestPickJWTKeyUnknownService(t *testing.T) {
	_, err := pickJWTKey("unknown", "basic")
	require.ErrorContains(t, err, "unknown service 'unknown'")
	require.ErrorContains(t, err, "application, backend, file")
}

func TestPickJWTKeyUnknownServer(t *testing.T) {
	_, err := pickJWTKey("file", "unknown")
	require.ErrorContains(t, err, "unknown server 'unknown' for service 'file'")
	require.ErrorContains(t, err, "admin, basic, download, export, info")
}
