/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package download ...
package filedownloader

import (
	"context"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer"
)

// TestNewStep ...
func TestNewStep(t *testing.T) {
	type args struct {
		ctx              context.Context
		callbackEndpoint string
		downloadEndpoint string
		nodeType         string
		version          string
		tag              string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx:              context.Background(),
				callbackEndpoint: "http://9.134.43.70:8002",
				downloadEndpoint: "http://9.134.43.70:7000",
				nodeType:         "agent",
				version:          "2",
				tag:              "2.1",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = main.SetCallbackEndPoint(tt.args.callbackEndpoint)
			_ = main.SetDownloadEndPoint(tt.args.downloadEndpoint)
			_ = main.SetNodeType(main.NodeType(tt.args.nodeType))
			_ = main.SetNodeGeneration(tt.args.version)
			_ = main.SetNodeVersion(tt.args.tag)

			if err := NewStep(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("NewStep() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
