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

package cmdb

import (
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func TestNewHostQueryPage(t *testing.T) {
	tests := []struct {
		name string
		page types.Page
		want Page
	}{
		{
			name: "default sort",
			page: types.Page{Offset: 100, Limit: 500},
			want: Page{Start: 100, Limit: 500, Sort: string(ccFieldBKHostID)},
		},
		{
			name: "preserve explicit sort",
			page: types.Page{Offset: 50, Limit: 100, Sort: "-bk_host_name"},
			want: Page{Start: 50, Limit: 100, Sort: "-bk_host_name"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newHostQueryPage(tt.page)
			if got != tt.want {
				t.Fatalf("newHostQueryPage() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
