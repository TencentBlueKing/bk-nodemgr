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

// Package types ...
package types

import (
	"testing"
)

// TestPage_Validate ...
func TestPage_Validate(t *testing.T) {
	tests := []struct {
		name    string
		page    Page
		wantErr bool
	}{
		{
			name: "valid page",
			page: Page{
				Offset: 0,
				Limit:  10,
				Sort:   "name",
			},
			wantErr: false,
		},
		{
			name: "offset is negative",
			page: Page{
				Offset: -1,
				Limit:  10,
				Sort:   "name",
			},
			wantErr: true,
		},
		{
			name: "limit is zero",
			page: Page{
				Offset: 0,
				Limit:  0,
				Sort:   "name",
			},
			wantErr: true,
		},
		{
			name: "limit is negative",
			page: Page{
				Offset: 0,
				Limit:  -1,
				Sort:   "name",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Page{
				Offset: tt.page.Offset,
				Limit:  tt.page.Limit,
				Sort:   tt.page.Sort,
			}
			if err := p.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
