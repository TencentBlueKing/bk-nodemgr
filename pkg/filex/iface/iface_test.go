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

package iface

import "testing"

func Test_ConvertAbsPathToAbsDirs(t *testing.T) {
	type args struct {
		path string
		want []string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "normal dir",
			args: args{
				path: "/data/test/dir",
				want: []string{"data", "test", "dir"},
			},
		},
		{
			name: "relative dir",
			args: args{
				path: "data/test/dir",
				want: []string{"data", "test", "dir"},
			},
		},
		{
			name: "windows dir",
			args: args{
				path: "c:\\data\\test\\dir",
				want: []string{"c:", "data", "test", "dir"},
			},
		},
		{
			name: "duplicated slash",
			args: args{
				path: "///data//test///dir//",
				want: []string{"data", "test", "dir"},
			},
		},
		{
			name: "multi-mix separator",
			args: args{
				path: "/\\data//\\\\/test/dir",
				want: []string{"data", "test", "dir"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := ConvertAbsPathToAbsDirs(test.args.path)
			if len(got) != len(test.args.want) {
				t.Errorf("ConvertAbsPathToAbsDirs() = %v, want %v", got, test.args.want)
			}
			for i := range got {
				if got[i] != test.args.want[i] {
					t.Errorf("ConvertAbsPathToAbsDirs() = %v, want %v", got, test.args.want)
				}
			}
		})
	}
}
