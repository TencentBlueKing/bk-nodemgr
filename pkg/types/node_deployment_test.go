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

package types

import (
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
)

func TestNodeInstallMethodCheckAvailable(t *testing.T) {
	tests := []struct {
		name          string
		installMethod NodeInstallMethod
		osType        criteria.OSType
		wantErr       bool
	}{
		{
			name:          "windows auto is supported",
			installMethod: NodeInstallMethodAuto,
			osType:        criteria.OSWindows,
		},
		{
			name:          "windows ssh is supported",
			installMethod: NodeInstallMethodSSH,
			osType:        criteria.OSWindows,
		},
		{
			name:          "windows wmi is supported",
			installMethod: NodeInstallMethodWMI,
			osType:        criteria.OSWindows,
		},
		{
			name:          "linux ssh is supported",
			installMethod: NodeInstallMethodSSH,
			osType:        criteria.OSLinux,
		},
		{
			name:          "linux wmi is unsupported",
			installMethod: NodeInstallMethodWMI,
			osType:        criteria.OSLinux,
			wantErr:       true,
		},
		{
			name:          "unknown method is invalid",
			installMethod: NodeInstallMethod("telnet"),
			osType:        criteria.OSWindows,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.installMethod.CheckAvailable(tt.osType)
			if (err != nil) != tt.wantErr {
				t.Fatalf("CheckAvailable() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
