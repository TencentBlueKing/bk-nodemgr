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
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/locker"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
)

// Handler this is a interface.
type Handler interface {
	locker.MutexFactory
}

// New creates a new redsync handler.
// Supports both standalone and cluster mode via redis.UniversalClient.
func New(redisClient redis.UniversalClient) Handler {
	return &handler{
		rs: redsync.New(goredis.NewPool(redisClient)),
	}
}

// handler ...
type handler struct {
	rs *redsync.Redsync
}

// NewMutex ...
func (l handler) NewMutex(name string) locker.Mutex {
	return mutex{
		mutex: l.rs.NewMutex(name),
	}
}

// mutex ...
type mutex struct {
	mutex *redsync.Mutex
}

// TryLock locks the given key.
func (mtx mutex) TryLock() error {
	ctx := context.Background()
	if err := mtx.mutex.LockContext(ctx); err != nil {
		return err
	}

	return nil
}

// Unlock unlocks the given key.
func (mtx mutex) Unlock() error {
	ctx := context.Background()
	result, err := mtx.mutex.UnlockContext(ctx)
	if err != nil {
		return err
	}

	if !result {
		return fmt.Errorf("unlock failed, lock-name(%s)", mtx.mutex.Name())
	}

	return nil
}

// Name ...
func (mtx mutex) Name() string {
	return mtx.mutex.Name()
}
