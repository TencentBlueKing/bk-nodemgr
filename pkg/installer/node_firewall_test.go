/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package installer

import (
	"strings"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestNodeFirewallFlagValues(t *testing.T) {
	features, remoteIPs := NodeFirewallFlagValues(nil)
	require.Empty(t, features)
	require.Empty(t, remoteIPs)

	features, remoteIPs = NodeFirewallFlagValues(&types.DeploymentFirewall{
		RemoteIPs: []string{"10.0.0.0/24", "fd00::/8"},
		Features:  []types.FirewallFeature{types.FirewallFeatureGSE, types.FirewallFeatureNodeMgr},
	})
	require.Equal(t, "gse,nodemgr", features)
	require.Equal(t, "10.0.0.0/24,fd00::/8", remoteIPs)
}

func TestNodeInstallParamsFirewallArgs(t *testing.T) {
	params := &NodeInstallParams{
		NodeCommonParams: NodeCommonParams{
			DeployEnv:     "test",
			Generation:    2,
			NodeRole:      "agent",
			BaseWorkDir:   "/var/lib",
			BaseDeployDir: "/opt",
		},
		InstallerPath:     "/tmp/installer",
		DownloadSvrAddr:   "http://127.0.0.1",
		CallbackSvrAddr:   "http://127.0.0.1",
		DeployToken:       "token",
		NodeVersion:       "v1",
		OperInstID:        "1",
		FirewallFeatures:  "gse,nodemgr",
		FirewallRemoteIPs: "10.0.0.0/24,fd00::/8",
	}

	for _, args := range [][]string{params.buildArgs(), params.buildDownloadBeforeCallbackArgs(), params.buildServerFirstArgs()} {
		joined := strings.Join(args, " ")
		require.Contains(t, joined, `--firewall_features="gse,nodemgr"`)
		require.Contains(t, joined, `--firewall_remote_ips="10.0.0.0/24,fd00::/8"`)
	}

	params.FirewallFeatures = ""
	params.FirewallRemoteIPs = ""
	for _, args := range [][]string{params.buildArgs(), params.buildDownloadBeforeCallbackArgs(), params.buildServerFirstArgs()} {
		require.NotContains(t, strings.Join(args, " "), "firewall")
	}
}

func TestNodeOfflineInstallParamsFirewallArgs(t *testing.T) {
	params := &NodeOfflineInstallParams{
		NodeCommonParams: NodeCommonParams{
			DeployEnv:     "test",
			Generation:    2,
			NodeRole:      "agent",
			BaseWorkDir:   "/var/lib",
			BaseDeployDir: "/opt",
		},
		InstallerFileName: "installer",
		DeployToken:       "token",
		NodeVersion:       "v1",
		OperInstID:        "1",
		FirewallFeatures:  "gse",
		FirewallRemoteIPs: "10.0.0.0/24",
	}

	joined := strings.Join(params.buildArgs(), " ")
	require.Contains(t, joined, `--firewall_features="gse"`)
	require.Contains(t, joined, `--firewall_remote_ips="10.0.0.0/24"`)
}

func TestNodeLifecycleParamsFirewallArgs(t *testing.T) {
	common := NodeCommonParams{
		DeployEnv:        "test",
		Generation:       2,
		NodeRole:         "agent",
		BaseWorkDir:      "/var/lib",
		BaseDeployDir:    "/opt",
		InstallerWorkDir: "/tmp",
	}

	upgrade := &NodeUpgradeParams{
		NodeCommonParams:  common,
		InstallerFileName: "installer",
		CallbackSvrAddr:   "http://127.0.0.1",
		DeployToken:       "token",
		NodeVersion:       "v1",
		OperInstID:        "1",
		FirewallFeatures:  "gse",
		FirewallRemoteIPs: "10.0.0.0/24",
	}
	reconfig := &NodeReconfigParams{
		NodeCommonParams:  common,
		InstallerFileName: "installer",
		CallbackSvrAddr:   "http://127.0.0.1",
		DeployToken:       "token",
		OperInstID:        "1",
		FirewallFeatures:  "gse",
		FirewallRemoteIPs: "10.0.0.0/24",
	}
	uninstall := &NodeUninstallParams{
		NodeCommonParams:  common,
		InstallerFileName: "installer",
		CallbackSvrAddr:   "http://127.0.0.1",
		DeployToken:       "token",
		OperInstID:        "1",
		FirewallFeatures:  "gse",
		FirewallRemoteIPs: "10.0.0.0/24",
	}

	for _, args := range [][]string{upgrade.buildArgs(), reconfig.buildArgs(), uninstall.buildArgs()} {
		joined := strings.Join(args, " ")
		require.Contains(t, joined, `--firewall_features="gse"`)
		require.Contains(t, joined, `--firewall_remote_ips="10.0.0.0/24"`)
	}
}
