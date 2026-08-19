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

package manager

import (
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/node"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func TestManagerGetNodeInstallOperationDefAgentWindowsDirectMethods(t *testing.T) {
	tests := []struct {
		name          string
		installMethod types.NodeInstallMethod
		directInstall bool
		wantOperName  string
	}{
		{
			name:          "direct windows auto uses auto operation",
			installMethod: types.NodeInstallMethodAuto,
			directInstall: true,
			wantOperName:  node.OperDefNameInstallNodeByWindowsAuto,
		},
		{
			name:          "direct windows ssh stays explicit ssh operation",
			installMethod: types.NodeInstallMethodSSH,
			directInstall: true,
			wantOperName:  node.OperDefNameInstallNodeByWindowsSSH,
		},
		{
			name:          "direct windows wmi stays explicit wmi operation",
			installMethod: types.NodeInstallMethodWMI,
			directInstall: true,
			wantOperName:  node.OperDefNameInstallNodeByWMI,
		},
		{
			name:          "relay windows auto stays windows ssh pagent operation",
			installMethod: types.NodeInstallMethodAuto,
			directInstall: false,
			wantOperName:  node.OperDefNameInstallPagntNodeByWindowsSSH,
		},
	}

	mgr := &Manager{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deploy := &types.NodeDeployment{
				Token: "test-token",
				Info: &types.DeploymentInfo{
					Host: types.Host{
						Static:  &types.HostStatic{OSType: string(criteria.OSWindows)},
						Dynamic: &types.HostDynamic{NodeRole: types.NodeRoleAgent},
					},
					InstallOptions: types.DeploymentInstallOptions{
						DirectInstall: tt.directInstall,
						InstallMethod: tt.installMethod,
					},
				},
			}

			oper, err := mgr.getNodeInstallOperationDefAgent(deploy, "tester")
			if err != nil {
				t.Fatalf("getNodeInstallOperationDefAgent() error = %v", err)
			}
			if got := oper.Name(); got != tt.wantOperName {
				t.Fatalf("Name() = %q, want %q", got, tt.wantOperName)
			}
		})
	}
}
