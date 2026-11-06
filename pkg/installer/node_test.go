/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and
 * limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package installer

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func newTestNodeInstallParams(sudoInstall bool) *NodeInstallParams {
	return &NodeInstallParams{
		NodeCommonParams: NodeCommonParams{
			DeployEnv:     "test",
			Generation:    2,
			NodeRole:      "agent",
			BaseWorkDir:   "/tmp/nodemgr",
			BaseDeployDir: "/usr/local/gse",
		},
		InstallerPath:   "/tmp/nodemgr/installer",
		DownloadSvrAddr: "http://127.0.0.1:8080",
		CallbackSvrAddr: "http://127.0.0.1:8081",
		NodeVersion:     "1.0.0",
		DeployToken:     "token",
		OperInstID:      "oper-inst-id",
		SudoInstall:     sudoInstall,
	}
}

func TestNodeInstallParamsToUnixScriptSudoInstall(t *testing.T) {
	_, script, err := newTestNodeInstallParams(false).ToUnixScript()
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(script, "/tmp/nodemgr/installer "+NodeCmdFullInstall))
	require.NotContains(t, script, "sudo")

	_, script, err = newTestNodeInstallParams(true).ToUnixScript()
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(script, "sudo -n /tmp/nodemgr/installer "+NodeCmdFullInstall))
}

func TestNodeInstallParamsToUnixScriptManualSudoInstall(t *testing.T) {
	_, script, err := newTestNodeInstallParams(false).ToUnixScriptManual()
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(script, "/tmp/nodemgr/installer "+NodeCmdFullInstall))

	_, script, err = newTestNodeInstallParams(true).ToUnixScriptManual()
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(script, "sudo -n /tmp/nodemgr/installer "+NodeCmdFullInstall))
}

func TestNodeOfflineInstallParamsToUnixScriptOfflineSudoInstall(t *testing.T) {
	newParams := func(sudoInstall bool) *NodeOfflineInstallParams {
		return &NodeOfflineInstallParams{
			NodeCommonParams: NodeCommonParams{
				DeployEnv:     "test",
				Generation:    2,
				NodeRole:      "proxy",
				BaseWorkDir:   "/tmp/nodemgr",
				BaseDeployDir: "/usr/local/gse",
			},
			InstallerFileName: "installer",
			NodeVersion:       "1.0.0",
			DeployToken:       "token",
			OperInstID:        "oper-inst-id",
			SudoInstall:       sudoInstall,
		}
	}

	installerCall := `"${SCRIPT_DIR}/installer" ` + NodeCmdFullInstall + " \\"

	_, script, err := newParams(false).ToUnixScriptOffline()
	require.NoError(t, err)
	require.Contains(t, script, "\n"+installerCall)
	require.NotContains(t, script, "sudo")

	_, script, err = newParams(true).ToUnixScriptOffline()
	require.NoError(t, err)
	require.Contains(t, script, "\nsudo -n "+installerCall)
}
