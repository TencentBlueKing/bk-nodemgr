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

package database

import (
	"io"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveRedisAddrPreservesSentinelMasterHostPort(t *testing.T) {
	binDir := t.TempDir()
	redisCLI := filepath.Join(binDir, "redis-cli")
	require.NoError(t, os.WriteFile(redisCLI, []byte(`#!/bin/sh
printf '%s\r\n' "$TEST_SENTINEL_HOST" "$TEST_SENTINEL_PORT"
`), 0o700))
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	tests := []struct {
		name string
		host string
		port string
	}{
		{name: "ipv4", host: "127.0.0.1", port: "6379"},
		{name: "hostname", host: "redis-master", port: "6380"},
		{name: "ipv6", host: "::1", port: "6380"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := &config.Redis{
				Type:       config.RedisTypeSentinel,
				Addrs:      []string{"127.0.0.1:26379"},
				MasterName: "mymaster",
			}
			addr, err := resolveRedisAddr(conf, nil, []string{
				"TEST_SENTINEL_HOST=" + tt.host,
				"TEST_SENTINEL_PORT=" + tt.port,
			}, io.Discard)
			require.NoError(t, err)

			host, port, err := net.SplitHostPort(addr)
			require.NoError(t, err)
			assert.Equal(t, tt.host, host)
			assert.Equal(t, tt.port, port)
		})
	}
}
