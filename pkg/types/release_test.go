/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package types

import "testing"

func TestPluginPkgAvailablePortRangeValidate(t *testing.T) {
	tests := []struct {
		name      string
		portRange PluginPkgAvailablePortRange
		wantErr   bool
	}{
		{name: "single port", portRange: "8080"},
		{name: "single port followed by range", portRange: "8080,10000-65535"},
		{
			name:      "multiple ports and ranges",
			portRange: "4000-5000,10000,65500,65502-65503,65530",
		},
		{name: "empty range", portRange: "", wantErr: true},
		{name: "non-decimal separator", portRange: "1000~two thousand", wantErr: true},
		{name: "multiple range separators", portRange: "1000-2000-3000", wantErr: true},
		{name: "descending range", portRange: "2000-1000", wantErr: true},
		{name: "empty list item", portRange: "1000,,2000", wantErr: true},
		{name: "port above maximum", portRange: "65536", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.portRange.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("PluginPkgAvailablePortRange.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
