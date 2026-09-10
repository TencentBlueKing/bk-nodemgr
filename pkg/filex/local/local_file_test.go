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

// Package local ...
package local

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/joho/godotenv"
)

func testFilePath(t *testing.T) string {
	err := godotenv.Load(".env")
	if err != nil {
		t.Fatal(err)
	}

	return os.Getenv("LOCAL_FILE_PATH")
}

// TestNewLocalFile ...
func TestNewLocalFile(t *testing.T) {
	type args struct {
		fullPath string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				fullPath: testFilePath(t),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewLocalFile(tt.args.fullPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewLocalFile() error = %v, wantErr %v. path(%s)", err, tt.wantErr, tt.args.fullPath)
				return
			}

			t.Logf("NewLocalFile() = %v", got)

			info := got.Info()
			t.Logf("FileInfo = %v", info)

			file, err := os.Create(filepath.Join(filepath.Dir(tt.args.fullPath), "test.txt"))
			if err != nil {
				t.Fatalf("Create() error = %v", err)
			}
			defer file.Close()

			reader, err := got.Content(contextx.Background())
			if err != nil {
				t.Fatalf("Content() error = %v", err)
			}
			defer reader.Close()

			_, err = io.Copy(file, reader)
		})
	}
}
