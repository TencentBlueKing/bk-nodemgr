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

// Package retrier ...
package retrier

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestExpoBackoff ...
// NOCC: golint/fnsize(func design is not suitable for splitting).
func TestExpoBackoff(t *testing.T) {
	backgroundContext := func(t *testing.T) context.Context {
		t.Helper()
		return context.Background()
	}
	type fields struct {
		opts ExpoBackoffOpts
	}
	type args struct {
		ctxFn func(t *testing.T) context.Context
		fn    func(attempt int) error
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantErr      bool
		wantAttempts int
	}{
		{
			name: "success_without_retry",
			fields: fields{
				opts: ExpoBackoffOptsDefault(),
			},
			args: args{
				ctxFn: backgroundContext,
				fn: func(attempt int) error {
					return nil
				},
			},
			wantErr:      false,
			wantAttempts: 1,
		},
		{
			name: "failed but success after retry",
			fields: fields{
				opts: ExpoBackoffOptsDefault(),
			},
			args: args{
				ctxFn: backgroundContext,
				fn: func(attempt int) error {
					if attempt < 1 {
						return errors.New("temporary error")
					}

					return nil
				},
			},
			wantErr:      false,
			wantAttempts: 2,
		},
		{
			name: "all_retries_failed",
			fields: fields{
				opts: ExpoBackoffOptsDefault(),
			},
			args: args{
				ctxFn: backgroundContext,
				fn: func(attempt int) error {
					return errors.New("persistent error")
				},
			},
			wantErr:      true,
			wantAttempts: 3,
		},
		{
			name: "context_cancelled",
			fields: fields{
				opts: ExpoBackoffOptsDefault(),
			},
			args: args{
				ctxFn: func(t *testing.T) context.Context {
					t.Helper()
					ctx, cancel := context.WithCancel(context.Background())
					cancel()
					return ctx
				},
				fn: func(attempt int) error {
					return errors.New("some error")
				},
			},
			wantErr: true,
		},
		{
			name: "custom_max_retries",
			fields: fields{
				opts: func() ExpoBackoffOpts {
					opts := ExpoBackoffOptsDefault()
					opts.MaxRetries = 5
					return opts
				}(),
			},
			args: args{
				ctxFn: backgroundContext,
				fn: func(attempt int) error {
					if attempt < 4 {
						return errors.New("temporary error")
					}
					return nil
				},
			},
			wantErr:      false,
			wantAttempts: 5,
		},
		{
			name: "context_timeout",
			fields: fields{
				opts: ExpoBackoffOptsDefault(),
			},
			args: args{
				ctxFn: func(t *testing.T) context.Context {
					t.Helper()
					ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
					t.Cleanup(cancel)
					return ctx
				},
				fn: func(attempt int) error {
					if attempt < 1 {
						time.Sleep(200 * time.Millisecond)
					}

					return errors.New("timeout error")
				},
			},
			wantErr:      true,
			wantAttempts: 1,
		},
		{
			name: "zero_max_retries",
			fields: fields{
				opts: func() ExpoBackoffOpts {
					opts := ExpoBackoffOptsDefault()
					opts.MaxRetries = 0
					return opts
				}(),
			},
			args: args{
				ctxFn: backgroundContext,
				fn: func(attempt int) error {
					return errors.New("error")
				},
			},
			wantErr:      false,
			wantAttempts: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewExpoBackoff(tt.fields.opts)

			attemptCount := 0
			wrapperFn := func(attempt int) error {
				attemptCount++
				return tt.args.fn(attempt)
			}

			ctx := tt.args.ctxFn(t)
			err := e.Do(ctx, wrapperFn)
			if err != nil {
				t.Logf("Do() error = %v", err)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("Do() error = %v, wantErr %v", err, tt.wantErr)
			}

			if attemptCount != tt.wantAttempts {
				t.Errorf("Do() wantAttempts = %v, got = %v", tt.wantAttempts, attemptCount)
			}
		})
	}
}
