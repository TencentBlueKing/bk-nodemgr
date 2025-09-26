/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package retrier ...
package retrier

import (
	"context"
	"errors"
	"testing"
	"time"
)

// PollingOptsDefault default retry options.
func TestPolling_Do(t *testing.T) {
	type fields struct {
		opts PollingOpts
	}
	type args struct {
		ctx context.Context
		fn  func(attempt int) error
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test",
			fields: fields{
				opts: PollingOpts{
					Timeout:  35 * time.Second,
					Interval: time.Second,
				},
			},
			args: args{
				ctx: context.Background(),
				fn: func(attempt int) error {
					if attempt == 30 {
						return nil
					}

					return errors.New("test")
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPolling(tt.fields.opts)
			err := p.Do(tt.args.ctx, tt.args.fn)
			if err != nil {
				t.Logf("err: %v", err)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("Do() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
