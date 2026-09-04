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
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/locker"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
)

const (
	// defaultExpiry is the TTL of a distributed lock. It also bounds how long a
	// lock stays held after its owner crashes before Redis releases it.
	defaultExpiry = 8 * time.Second

	// renewIntervalRatio controls how often the watchdog renews the lock TTL.
	// The lock is renewed every expiry/renewIntervalRatio so that it never
	// expires while the owner is still alive.
	renewIntervalRatio = 3
)

// Handler this is a interface.
type Handler interface {
	locker.MutexFactory
}

// New creates a new redsync handler.
// Supports both standalone and cluster mode via redis.UniversalClient.
func New(redisClient redis.UniversalClient) Handler {
	return &handler{
		rs:     redsync.New(goredis.NewPool(redisClient)),
		expiry: defaultExpiry,
	}
}

// handler ...
type handler struct {
	rs     *redsync.Redsync
	expiry time.Duration
}

// NewMutex ...
func (l *handler) NewMutex(name string) locker.Mutex {
	return &mutex{
		mutex:  l.rs.NewMutex(name, redsync.WithExpiry(l.expiry)),
		expiry: l.expiry,
	}
}

// mutex wraps a redsync mutex with an auto-renewal watchdog so that a lock held
// longer than its TTL does not silently expire while its owner is still working.
type mutex struct {
	mutex  *redsync.Mutex
	expiry time.Duration

	// cancel stops the renew goroutine. done is closed once the renew goroutine
	// has fully exited, guaranteeing that no Extend races with Unlock.
	cancel context.CancelFunc
	done   chan struct{}
}

// TryLock locks the given key and starts auto-renewal on success.
func (mtx *mutex) TryLock() error {
	ctx := context.Background()
	if err := mtx.mutex.LockContext(ctx); err != nil {
		return err
	}

	mtx.startWatchdog()

	return nil
}

// Unlock stops auto-renewal and unlocks the given key.
func (mtx *mutex) Unlock() error {
	mtx.stopWatchdog()

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
func (mtx *mutex) Name() string {
	return mtx.mutex.Name()
}

// startWatchdog starts a goroutine that renews the lock until cancelled.
func (mtx *mutex) startWatchdog() {
	if mtx.cancel != nil {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	mtx.cancel = cancel
	mtx.done = done

	go mtx.renew(ctx, done)
}

// stopWatchdog cancels the renew goroutine and waits for it to exit.
func (mtx *mutex) stopWatchdog() {
	if mtx.cancel == nil {
		return
	}

	cancel := mtx.cancel
	done := mtx.done
	mtx.cancel = nil
	mtx.done = nil

	cancel()
	<-done
}

// renew periodically extends the lock TTL until ctx is cancelled or the lock is lost.
func (mtx *mutex) renew(ctx context.Context, done chan struct{}) {
	defer close(done)

	interval := mtx.expiry / renewIntervalRatio
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			if _, err := mtx.mutex.ExtendContext(ctx); err != nil {
				// The lock is already lost or expired, nothing more to renew.
				return
			}
		}
	}
}
