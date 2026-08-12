/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package tmp

import (
	"io"
	"os"
	"path/filepath"
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
			if err := file.CleanUp(); err != nil {
				t.Errorf("clean tmp file failed: %s", err)
			}
		})
	}

	if err := Clean(); err != nil {
		t.Errorf("clean tmp dir failed: %s", err)
	}
}

func TestGetTmpDirReturnsStableRoot(t *testing.T) {
	if err := Clean(); err != nil {
		t.Fatalf("clean tmp dir failed: %s", err)
	}
	t.Cleanup(func() {
		_ = Clean()
	})

	firstDir, err := GetTmpDir()
	if err != nil {
		t.Fatalf("get first tmp dir failed: %s", err)
	}

	secondDir, err := GetTmpDir()
	if err != nil {
		t.Fatalf("get second tmp dir failed: %s", err)
	}

	if firstDir != secondDir {
		t.Fatalf("tmp dir changed, first(%s), second(%s)", firstDir, secondDir)
	}

	if _, err := os.Stat(firstDir); err != nil {
		t.Fatalf("stat tmp dir failed: %s", err)
	}

	if err := Clean(); err != nil {
		t.Fatalf("clean tmp dir failed: %s", err)
	}

	if _, err := os.Stat(firstDir); !os.IsNotExist(err) {
		t.Fatalf("tmp dir still exists after clean, dir(%s), err(%v)", firstDir, err)
	}

	thirdDir, err := GetTmpDir()
	if err != nil {
		t.Fatalf("get third tmp dir failed: %s", err)
	}

	if thirdDir == firstDir {
		t.Fatalf("tmp dir was not reset after clean, dir(%s)", thirdDir)
	}
}

func TestNewTempFileWithSpecialNameUsesIsolatedDir(t *testing.T) {
	if err := Clean(); err != nil {
		t.Fatalf("clean tmp dir failed: %s", err)
	}
	t.Cleanup(func() {
		_ = Clean()
	})

	firstFile, err := NewTempFileWithSpecialName(io.NopCloser(strings.NewReader("first")), "fixed.conf")
	if err != nil {
		t.Fatalf("create first temp file failed: %s", err)
	}

	secondFile, err := NewTempFileWithSpecialName(io.NopCloser(strings.NewReader("second")), "fixed.conf")
	if err != nil {
		t.Fatalf("create second temp file failed: %s", err)
	}

	if firstFile.Path() == secondFile.Path() {
		t.Fatalf("temp file paths should differ, path(%s)", firstFile.Path())
	}

	firstDir := filepath.Dir(firstFile.Path())
	secondDir := filepath.Dir(secondFile.Path())
	if firstDir == secondDir {
		t.Fatalf("temp file dirs should differ, dir(%s)", firstDir)
	}

	if err := firstFile.CleanUp(); err != nil {
		t.Fatalf("clean first temp file failed: %s", err)
	}

	if _, err := os.Stat(firstDir); !os.IsNotExist(err) {
		t.Fatalf("first temp dir still exists, dir(%s), err(%v)", firstDir, err)
	}

	if err := secondFile.CleanUp(); err != nil {
		t.Fatalf("clean second temp file failed: %s", err)
	}

	if _, err := os.Stat(secondDir); !os.IsNotExist(err) {
		t.Fatalf("second temp dir still exists, dir(%s), err(%v)", secondDir, err)
	}
}
