/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package contextvalues

import (
	"context"
	"testing"
)

// Test_Contextvalues tests Set and Get.
func Test_Contextvalues(t *testing.T) {
	type args struct {
		ctx          context.Context
		setKey       Key
		setValue     string
		wantSetErr   bool
		getKey       Key
		wantGetValue string
		wantGetErr   bool
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "nil ctx",
			args: args{
				ctx:        nil,
				wantSetErr: true,
				wantGetErr: true,
			},
		},
		{
			name: "tenant",
			args: args{
				ctx:          context.Background(),
				setKey:       KeyTenantID,
				setValue:     "123",
				wantSetErr:   false,
				getKey:       KeyTenantID,
				wantGetValue: "123",
				wantGetErr:   false,
			},
		},
		{
			name: "request-id",
			args: args{
				ctx:          context.Background(),
				setKey:       KeyRequestID,
				setValue:     "xxx",
				wantSetErr:   false,
				getKey:       KeyRequestID,
				wantGetValue: "xxx",
				wantGetErr:   false,
			},
		},
		{
			name: "wrong key",
			args: args{
				ctx:        context.Background(),
				setKey:     KeyTenantID,
				setValue:   "abcdef",
				wantSetErr: false,
				getKey:     KeyRequestID,
				wantGetErr: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, err := Set(tt.args.ctx, tt.args.setKey, tt.args.setValue)
			if (err != nil) != tt.args.wantSetErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.args.wantSetErr)
			}

			gotValue, err := Get(ctx, tt.args.getKey)
			if (err != nil) != tt.args.wantGetErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.args.wantGetErr)
			}

			if gotValue != tt.args.wantGetValue {
				t.Errorf("Get() gotValue = %v, want %v", gotValue, tt.args.wantGetValue)
			}
		})
	}
}
