/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package precheck ...
package precheck

import (
	"context"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer"
)

func TestNewStep(t *testing.T) {
	type args struct {
		checkListPath string
		ctx           context.Context
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				checkListPath: "./check_list.json",
				ctx:           context.Background(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = main.SetPreCheckFilePath(tt.args.checkListPath)

			if err := NewStep(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("NewStep() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
