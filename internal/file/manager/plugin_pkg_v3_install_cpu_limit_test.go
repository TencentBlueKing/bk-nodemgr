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
 */

package manager

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPluginV3ProjectValidate(t *testing.T) {
	for _, limit := range []float64{0, 0.5, 1} {
		require.NoError(t, (&PluginV3Project{InstallCPULimit: limit}).Validate())
	}
	for _, limit := range []float64{-0.1, 1.1} {
		require.Error(t, (&PluginV3Project{InstallCPULimit: limit}).Validate())
	}
}

func TestReadOriginPluginV3ProjectInstallCPULimit(t *testing.T) {
	project, err := readOriginPluginV3Project(
		[]string{"demo", "project.yaml"},
		strings.NewReader("name: demo\nversion: 1.0.0\ninstallCpuLimit: 0.5\n"),
		nil,
	)
	require.NoError(t, err)
	require.Equal(t, 0.5, project.InstallCPULimit)

	_, err = readOriginPluginV3Project(
		[]string{"demo", "project.yaml"},
		strings.NewReader("name: demo\nversion: 1.0.0\ninstallCpuLimit: invalid\n"),
		nil,
	)
	require.Error(t, err)
}
