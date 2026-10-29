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

package cache

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"
)

func TestMemoryCache_GetSetCopiesBytes(t *testing.T) {
	ctx := context.Background()
	c := NewMemoryCache(time.Minute)
	value := []byte("abc")

	if err := c.Set(ctx, "key", value); err != nil {
		t.Fatalf("Set() unexpected error: %v", err)
	}
	value[0] = 'x'

	got, err := c.Get(ctx, "key")
	if err != nil {
		t.Fatalf("Get() unexpected error: %v", err)
	}
	if string(got) != "abc" {
		t.Fatalf("Get() = %q, want stored copy", got)
	}

	got[0] = 'y'
	reloaded, err := c.Get(ctx, "key")
	if err != nil {
		t.Fatalf("Get() after returned slice mutation unexpected error: %v", err)
	}
	if string(reloaded) != "abc" {
		t.Fatalf("Get() after returned slice mutation = %q, want stored copy", reloaded)
	}
}

func TestMemoryCache_GetReturnsMissingKeyError(t *testing.T) {
	c := NewMemoryCache(time.Minute)

	_, err := c.Get(context.Background(), "missing")
	if err == nil || !strings.Contains(err.Error(), "key not found. key(missing)") {
		t.Fatalf("Get() error = %v, want missing key error", err)
	}
}

func TestMemoryCache_GetOrLoadSingleflight(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx := context.Background()
		c := NewMemoryCache(time.Minute)
		release := make(chan struct{})
		value := []byte("value")
		var calls atomic.Int32
		load := func(context.Context) ([]byte, error) {
			calls.Add(1)
			<-release
			return value, nil
		}
		type result struct {
			data []byte
			err  error
		}
		results := make(chan result, 2)
		for range 2 {
			go func() {
				data, err := c.GetOrLoad(ctx, "key", load)
				results <- result{data: data, err: err}
			}()
		}

		synctest.Wait()
		if got := calls.Load(); got != 1 {
			t.Errorf("loader calls = %d, want 1 for concurrent misses", got)
		}
		close(release)
		first, second := <-results, <-results
		for _, got := range []result{first, second} {
			if got.err != nil {
				t.Fatalf("GetOrLoad() unexpected error: %v", got.err)
			}
			if string(got.data) != "value" {
				t.Fatalf("GetOrLoad() = %q, want value", got.data)
			}
		}
		first.data[0] = 'x'
		if string(second.data) != "value" || string(value) != "value" {
			t.Fatal("GetOrLoad() shared returned bytes with another caller or loader")
		}
		value[0] = 'y'

		for range 2 {
			got, err := c.GetOrLoad(ctx, "key", load)
			if err != nil {
				t.Fatalf("GetOrLoad() cache hit unexpected error: %v", err)
			}
			if string(got) != "value" {
				t.Fatalf("GetOrLoad() cache hit = %q, want stored copy", got)
			}
			got[0] = 'z'
		}
		if got := calls.Load(); got != 1 {
			t.Fatalf("loader calls = %d, want 1 after cache hits", got)
		}
	})
}

func TestMemoryCache_GetOrLoadContextCancellation(t *testing.T) {
	c := NewMemoryCache(time.Minute)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := c.GetOrLoad(ctx, "key", func(context.Context) ([]byte, error) {
		t.Error("loader should not run for a cancelled context")
		return nil, nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("GetOrLoad() error = %v, want context cancellation", err)
	}
}

func TestMemoryCache_GetOrLoadDoesNotCacheError(t *testing.T) {
	ctx := context.Background()
	c := NewMemoryCache(time.Minute)
	loadErr := errors.New("temporary failure")
	var calls atomic.Int32
	load := func(context.Context) ([]byte, error) {
		if calls.Add(1) == 1 {
			return nil, loadErr
		}
		return []byte("value"), nil
	}

	if _, err := c.GetOrLoad(ctx, "key", load); !errors.Is(err, loadErr) {
		t.Fatalf("GetOrLoad() error = %v, want loader error", err)
	}
	for range 2 {
		got, err := c.GetOrLoad(ctx, "key", load)
		if err != nil {
			t.Fatalf("GetOrLoad() after loader error unexpected error: %v", err)
		}
		if string(got) != "value" {
			t.Fatalf("GetOrLoad() = %q, want value", got)
		}
	}
	if got := calls.Load(); got != 2 {
		t.Fatalf("loader calls = %d, want 2 after retry and cache hit", got)
	}
}

func TestMemoryCache_GetOrLoadUsesDefaultExpiration(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx := context.Background()
		c := NewMemoryCache(time.Minute)
		var calls atomic.Int32
		load := func(context.Context) ([]byte, error) {
			if calls.Add(1) == 1 {
				return []byte("first"), nil
			}
			return []byte("second"), nil
		}
		for range 2 {
			got, err := c.GetOrLoad(ctx, "key", load)
			if err != nil {
				t.Fatalf("GetOrLoad() before expiration unexpected error: %v", err)
			}
			if string(got) != "first" {
				t.Fatalf("GetOrLoad() before expiration = %q, want first", got)
			}
		}

		time.Sleep(time.Minute + time.Nanosecond)
		got, err := c.GetOrLoad(ctx, "key", load)
		if err != nil {
			t.Fatalf("GetOrLoad() after expiration unexpected error: %v", err)
		}
		if string(got) != "second" || calls.Load() != 2 {
			t.Fatalf("GetOrLoad() after expiration = %q, loader calls = %d, want second and 2", got, calls.Load())
		}
	})
}

func TestMemoryCache_GetOrLoadWaiterCancellation(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx := context.Background()
		c := NewMemoryCache(time.Minute)
		release := make(chan struct{})
		load := func(ctx context.Context) ([]byte, error) {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-release:
				return []byte("value"), nil
			}
		}
		type result struct {
			data []byte
			err  error
		}
		leader := make(chan result, 1)
		go func() {
			data, err := c.GetOrLoad(ctx, "key", load)
			leader <- result{data: data, err: err}
		}()
		synctest.Wait()

		waiterCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		waiter := make(chan error, 1)
		go func() {
			_, err := c.GetOrLoad(waiterCtx, "key", load)
			waiter <- err
		}()
		synctest.Wait()
		cancel()
		if err := <-waiter; !errors.Is(err, context.Canceled) {
			t.Errorf("GetOrLoad() waiter error = %v, want context cancellation", err)
		}
		close(release)
		got := <-leader
		if got.err != nil {
			t.Fatalf("GetOrLoad() leader unexpected error: %v", got.err)
		}
		if string(got.data) != "value" {
			t.Fatalf("GetOrLoad() leader = %q, want value", got.data)
		}
		cached, err := c.Get(ctx, "key")
		if err != nil {
			t.Fatalf("Get() after waiter cancellation unexpected error: %v", err)
		}
		if string(cached) != "value" {
			t.Fatalf("Get() after waiter cancellation = %q, want value", cached)
		}
	})
}

func TestMemoryCache_GetOrLoadDifferentKeys(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx := context.Background()
		c := NewMemoryCache(time.Minute)
		release := make(chan struct{})
		first := make(chan error, 1)
		go func() {
			got, err := c.GetOrLoad(ctx, "first", func(context.Context) ([]byte, error) {
				<-release
				return []byte("first value"), nil
			})
			if err != nil {
				first <- err
				return
			}
			if string(got) != "first value" {
				t.Errorf("GetOrLoad(first) = %q, want first value", got)
			}
			first <- nil
		}()
		synctest.Wait()

		got, err := c.GetOrLoad(ctx, "second", func(context.Context) ([]byte, error) {
			return []byte("second value"), nil
		})
		close(release)
		if err != nil {
			t.Fatalf("GetOrLoad(second) unexpected error: %v", err)
		}
		if string(got) != "second value" {
			t.Fatalf("GetOrLoad(second) = %q, want second value", got)
		}
		if err := <-first; err != nil {
			t.Fatalf("GetOrLoad(first) unexpected error: %v", err)
		}
	})
}

func TestMemoryCache_SetUsesDefaultExpiration(t *testing.T) {
	ctx := context.Background()
	c := NewMemoryCache(20 * time.Millisecond)

	if err := c.Set(ctx, "key", []byte("value")); err != nil {
		t.Fatalf("Set() unexpected error: %v", err)
	}
	if _, err := c.Get(ctx, "key"); err != nil {
		t.Fatalf("Get() before default expiration unexpected error: %v", err)
	}

	waitForMissing(t, c, ctx, "key", 200*time.Millisecond)
}

func TestMemoryCache_SetWithExpiration(t *testing.T) {
	ctx := context.Background()
	c := NewMemoryCache(time.Hour)

	if err := c.SetWithExpiration(ctx, "expiring", []byte("value"), 20*time.Millisecond); err != nil {
		t.Fatalf("SetWithExpiration() unexpected error: %v", err)
	}
	waitForMissing(t, c, ctx, "expiring", 200*time.Millisecond)

	if err := c.SetWithExpiration(ctx, "forever", []byte("value"), 0); err != nil {
		t.Fatalf("SetWithExpiration() zero expiration unexpected error: %v", err)
	}
	time.Sleep(30 * time.Millisecond)
	if _, err := c.Get(ctx, "forever"); err != nil {
		t.Fatalf("Get() zero-expiration key unexpected error: %v", err)
	}
}

func TestMemoryCache_SetNX(t *testing.T) {
	ctx := context.Background()
	c := NewMemoryCache(time.Minute)

	ok, err := c.SetNX(ctx, "key", []byte("first"))
	if err != nil {
		t.Fatalf("SetNX() unexpected error: %v", err)
	}
	if !ok {
		t.Fatalf("SetNX() = false, want true for missing key")
	}

	ok, err = c.SetNX(ctx, "key", []byte("second"))
	if err != nil {
		t.Fatalf("SetNX() existing key unexpected error: %v", err)
	}
	if ok {
		t.Fatalf("SetNX() = true, want false for existing key")
	}

	got, err := c.Get(ctx, "key")
	if err != nil {
		t.Fatalf("Get() unexpected error: %v", err)
	}
	if string(got) != "first" {
		t.Fatalf("Get() = %q, want first value", got)
	}
}

func TestMemoryCache_SetNXWithExpirationAfterExpiry(t *testing.T) {
	ctx := context.Background()
	c := NewMemoryCache(time.Minute)

	ok, err := c.SetNXWithExpiration(ctx, "key", []byte("first"), 20*time.Millisecond)
	if err != nil {
		t.Fatalf("SetNXWithExpiration() unexpected error: %v", err)
	}
	if !ok {
		t.Fatalf("SetNXWithExpiration() = false, want true for missing key")
	}
	waitForMissing(t, c, ctx, "key", 200*time.Millisecond)

	ok, err = c.SetNXWithExpiration(ctx, "key", []byte("second"), 0)
	if err != nil {
		t.Fatalf("SetNXWithExpiration() after expiry unexpected error: %v", err)
	}
	if !ok {
		t.Fatalf("SetNXWithExpiration() after expiry = false, want true")
	}

	time.Sleep(30 * time.Millisecond)
	got, err := c.Get(ctx, "key")
	if err != nil {
		t.Fatalf("Get() zero-expiration SetNX key unexpected error: %v", err)
	}
	if string(got) != "second" {
		t.Fatalf("Get() = %q, want second value", got)
	}
}

func TestMemoryCache_ExistsDelete(t *testing.T) {
	ctx := context.Background()
	c := NewMemoryCache(time.Minute)

	exists, err := c.Exists(ctx, "key")
	if err != nil {
		t.Fatalf("Exists() unexpected error: %v", err)
	}
	if exists {
		t.Fatalf("Exists() = true, want false before Set")
	}

	if err := c.Set(ctx, "key", []byte("value")); err != nil {
		t.Fatalf("Set() unexpected error: %v", err)
	}
	exists, err = c.Exists(ctx, "key")
	if err != nil {
		t.Fatalf("Exists() after Set unexpected error: %v", err)
	}
	if !exists {
		t.Fatalf("Exists() = false, want true after Set")
	}

	deleted, err := c.Delete(ctx, "key")
	if err != nil {
		t.Fatalf("Delete() unexpected error: %v", err)
	}
	if !deleted {
		t.Fatalf("Delete() = false, want true for existing key")
	}
	deleted, err = c.Delete(ctx, "key")
	if err != nil {
		t.Fatalf("Delete() missing key unexpected error: %v", err)
	}
	if deleted {
		t.Fatalf("Delete() = true, want false for missing key")
	}
}

func TestMemoryCache_NilContext(t *testing.T) {
	c := NewMemoryCache(time.Minute)
	var nilCtx context.Context

	if _, err := c.Get(nilCtx, "key"); err == nil || err.Error() != "context is nil" {
		t.Fatalf("Get(nil) error = %v, want context is nil", err)
	}
	if err := c.Set(nilCtx, "key", []byte("value")); err == nil || err.Error() != "context is nil" {
		t.Fatalf("Set(nil) error = %v, want context is nil", err)
	}
	if _, err := c.SetNX(nilCtx, "key", []byte("value")); err == nil || err.Error() != "context is nil" {
		t.Fatalf("SetNX(nil) error = %v, want context is nil", err)
	}
	if _, err := c.SetNXWithExpiration(nilCtx, "key", []byte("value"), time.Minute); err == nil || err.Error() != "context is nil" {
		t.Fatalf("SetNXWithExpiration(nil) error = %v, want context is nil", err)
	}
	if err := c.SetWithExpiration(nilCtx, "key", []byte("value"), time.Minute); err == nil || err.Error() != "context is nil" {
		t.Fatalf("SetWithExpiration(nil) error = %v, want context is nil", err)
	}
	if _, err := c.Exists(nilCtx, "key"); err == nil || err.Error() != "context is nil" {
		t.Fatalf("Exists(nil) error = %v, want context is nil", err)
	}
	if _, err := c.Delete(nilCtx, "key"); err == nil || err.Error() != "context is nil" {
		t.Fatalf("Delete(nil) error = %v, want context is nil", err)
	}
}

func waitForMissing(t *testing.T, c ICache, ctx context.Context, key string, timeout time.Duration) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := c.Get(ctx, key); err != nil && strings.Contains(err.Error(), "key not found") {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}

	t.Fatalf("key %q still existed after %s", key, timeout)
}
