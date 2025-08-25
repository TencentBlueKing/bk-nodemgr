/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package local ...
package local

import (
	"os"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/joho/godotenv"
)

func testDirPath(t *testing.T) string {
	err := godotenv.Load(".env")
	if err != nil {
		t.Fatal(err)
	}

	return os.Getenv("LOCAL_DIR_PATH")
}

// TestNewLocalDir ...
func TestNewLocalDir(t *testing.T) {
	type args struct {
		fullPath string
		logger   logger.ILogger
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				fullPath: testDirPath(t),
				logger:   logger.LoggerDefault{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewLocalDir(tt.args.fullPath, tt.args.logger)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewLocalDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			var fn func(dir *LocalDir)
			fn = func(dir *LocalDir) {
				t.Logf("dir: %s", dir.name)
				t.Logf("fullPath: %s", dir.fullPath)
			}

			fn(got)
		})
	}
}
