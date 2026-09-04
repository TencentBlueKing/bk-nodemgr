//go:build integration

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

// Package redsync ...
package redsync

import (
	"context"
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/testsuite/support"
)

// testClient ...
func testClient(t *testing.T) (Handler, string) {
	redisClient, keyPrefix := support.RequireRedisClientWithKeyPrefix(t)
	return New(redisClient), keyPrefix
}

// Test_mutex_Lock ...
func Test_mutex_Lock(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name     string
		args     args
		wantErr  bool
		isLock   bool
		isUnlock bool
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
			},
			isLock:   false,
			isUnlock: false,
			wantErr:  false,
		},
		{
			name: "already_lock",
			args: args{
				ctx: context.Background(),
			},
			wantErr:  true,
			isLock:   true,
			isUnlock: false,
		},
		{
			name: "already_unlock",
			args: args{
				ctx: context.Background(),
			},
			wantErr:  false,
			isLock:   true,
			isUnlock: true,
		},
	}
	h, keyPrefix := testClient(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := h.NewMutex(keyPrefix + ":" + tt.name)

			if tt.isLock {
				if err := m.TryLock(); err != nil {
					t.Errorf("Lock() error = %v", err)
				}
			}

			if tt.isUnlock {
				if err := m.Unlock(); err != nil {
					t.Errorf("Unlock() error = %v, wantUnlock %v", err, tt.isUnlock)
				}
			}

			if err := m.TryLock(); (err != nil) != tt.wantErr {
				t.Errorf("Lock() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test_mutex_Unlock ...
func Test_mutex_Unlock(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name     string
		args     args
		want     bool
		wantErr  bool
		isLock   bool
		isUnlock bool
		isExpire bool
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
			},
			want:     true,
			wantErr:  false,
			isLock:   true,
			isUnlock: false,
			isExpire: false,
		},
		{
			name:     "no_lock",
			args:     args{},
			want:     false,
			wantErr:  true,
			isLock:   false,
			isUnlock: false,
			isExpire: false,
		},
		{
			name: "already_unlock",
			args: args{
				ctx: context.Background(),
			},
			want:     false,
			wantErr:  true,
			isLock:   true,
			isUnlock: true,
			isExpire: false,
		},
		{
			// lock held longer than the default TTL should still be valid,
			// because the watchdog keeps renewing it before it expires.
			name: "renew_beyond_ttl",
			args: args{
				ctx: context.Background(),
			},
			want:     true,
			wantErr:  false,
			isLock:   true,
			isUnlock: false,
			isExpire: true,
		},
	}
	h, keyPrefix := testClient(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := h.NewMutex(keyPrefix + ":" + tt.name)

			if tt.isLock {
				if err := m.TryLock(); err != nil {
					t.Errorf("Lock() error = %v", err)
				}
			}

			if tt.isUnlock {
				if err := m.Unlock(); err != nil {
					t.Errorf("Unlock() error = %v, wantUnlock %v", err, tt.isUnlock)
				}
			}

			if tt.isExpire {
				// default redsync expiry is 8s; sleep beyond it to verify the
				// watchdog keeps the lock alive.
				time.Sleep(time.Second * 10)
			}

			err := m.Unlock()
			if err != nil {
				t.Logf("Unlock() error = %v", err)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("Unlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
