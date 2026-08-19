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

package tmp

import (
	"io"
	"strings"
	"testing"
)

// Test test NewTempFile.
func Test(t *testing.T) {
	type args struct {
		data io.ReadCloser
		name string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "txt_test",
			args: args{
				data: func() io.ReadCloser {
					str := "hello this is a test"
					return io.NopCloser(strings.NewReader(str))
				}(),
				name: "txt_test",
			},
			wantErr: false,
		},
		{
			name: "nil data",
			args: args{
				data: nil,
				name: "",
			},
			wantErr: true,
		},
		{
			name: "empty name",
			args: args{
				data: func() io.ReadCloser {
					str := "hello this is a test"
					return io.NopCloser(strings.NewReader(str))
				}(),
				name: "",
			},
			wantErr: false,
		},
		{
			name: "close data",
			args: args{
				data: func() io.ReadCloser {
					str := "hello this is a test"
					data := io.NopCloser(strings.NewReader(str))
					data.Close()

					return data
				}(),
				name: "close_data",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := NewTempFile(tt.args.data, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTempFile() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if err != nil {
				t.Logf("err: %v", err)

				return
			}

			t.Logf("tmp file path: %s", file.Path())
		})
	}

	if err := Clean(); err != nil {
		t.Errorf("clean tmp file failed: %s", err)
	}
}
