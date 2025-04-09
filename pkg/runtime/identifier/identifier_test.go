/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package identifier

import (
	"context"
	"testing"
)

// Test_GenID test severail generating id funcs.
func Test_GenID(t *testing.T) {
	for i := 0; i < 10000; i++ {
		GenTriggerID()
	}

	for i := 0; i < 10000; i++ {
		GenRequestID()
	}

	for i := 0; i < 10000; i++ {
		GenOperationID()
	}

	for i := 0; i < 10000; i++ {
		GenOperationInstanceID()
	}

	for i := 0; i < 10000; i++ {
		GenActionInstanceID()
	}

	for i := 0; i < 10000; i++ {
		GenServiceID()
	}
}

// Test_HandleRequestID test request-id get and set.
func Test_HandleRequestID(t *testing.T) {
	type args struct {
		ctx          context.Context
		setValue     string
		wantSetErr   bool
		wantGetValue string
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
			},
		},
		{
			name: "tenant",
			args: args{
				ctx:          context.Background(),
				setValue:     "123",
				wantSetErr:   false,
				wantGetValue: "123",
			},
		},
		{
			name: "request-id",
			args: args{
				ctx:          context.Background(),
				setValue:     "xxx",
				wantSetErr:   false,
				wantGetValue: "xxx",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, err := SetRequestID(tt.args.ctx, tt.args.setValue)
			if (err != nil) != tt.args.wantSetErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.args.wantSetErr)
			}

			if gotValue := GetRequestID(ctx); gotValue != tt.args.wantGetValue {
				t.Errorf("Get() gotValue = %v, want %v", gotValue, tt.args.wantGetValue)
			}
		})
	}
}
