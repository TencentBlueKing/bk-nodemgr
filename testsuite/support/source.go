//go:build integration

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

package support

import "testing"

type source string

const (
	sourceAuto   source = SourceAuto
	sourceEnv    source = SourceEnv
	sourceDocker source = SourceDocker
)

func resolveSource(t testing.TB) (source, map[string]string) {
	t.Helper()

	values, hasEnv := readEnvFile(t)
	configured := configValue(values, EnvSourceKey)
	if configured == "" {
		configured = SourceAuto
	}

	switch source(configured) {
	case sourceAuto:
		if hasEnv {
			return sourceEnv, values
		}
		return sourceDocker, nil
	case sourceEnv:
		if !hasEnv {
			t.Fatalf("%s=env requires %s", EnvSourceKey, integrationEnvPath)
		}
		return sourceEnv, values
	case sourceDocker:
		return sourceDocker, values
	default:
		t.Fatalf("unsupported %s %q, want auto, env, or docker", EnvSourceKey, configured)
		return "", nil
	}
}
