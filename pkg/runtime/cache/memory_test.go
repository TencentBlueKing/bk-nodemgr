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
	"strings"
	"testing"
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
