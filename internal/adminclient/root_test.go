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

package adminclient

import (
	"testing"

	backendadmin "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/backendadmin"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRootCMDRegistersBackendSubcommand(t *testing.T) {
	cmd := NewRootCMD(func(string) (backendadmin.IHandler, error) {
		return &backendadmin.Handler{}, nil
	})

	require.NotNil(t, cmd)
	assert.Equal(t, "bk-nodemgr-adminclient", cmd.Use)
	assert.NotNil(t, findSubcommand(cmd, "backend"))
}

func findSubcommand(cmd *cobra.Command, use string) *cobra.Command {
	for _, child := range cmd.Commands() {
		if child.Use == use {
			return child
		}
	}

	return nil
}
