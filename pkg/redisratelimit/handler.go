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

// Package redisratelimit implements shared token buckets backed by Redis.
package redisratelimit

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/ratelimit"
	"github.com/redis/go-redis/v9"
)

var (
	_ ratelimit.IHandler = (*Handler)(nil)
	_ ratelimit.ILimiter = (*limiter)(nil)
)

// Handler creates limiters sharing a Redis namespace. The caller owns the client.
// Local rule registrations are retained for the lifetime of the handler.
type Handler struct {
	client    redis.UniversalClient
	namespace string
	mu        sync.Mutex
	rules     map[string]ratelimit.Rule
}

// New creates a factory without contacting Redis. The injected client must respect
// context deadlines and must not retry ambiguous writes: configure ContextTimeoutEnabled=true,
// MaxRetries=-1 and, for cluster clients, MaxRedirects=-1 before creating the client.
// Cluster redirects are consequently returned as errors rather than retried.
// The caller must keep these options unchanged and own the client's lifecycle.
// Redis state must not be evicted: eviction, deletion or failover with lost writes
// can reset quotas. Cross-process rules must agree even after idle state expires.
func New(client redis.UniversalClient, namespace string) (*Handler, error) {
	if namespace == "" {
		return nil, errors.New("rate limiter namespace is empty")
	}
	if err := validateClient(client); err != nil {
		return nil, err
	}

	return &Handler{
		client: client, namespace: namespace, rules: make(map[string]ratelimit.Rule),
	}, nil
}

func validateClient(client redis.UniversalClient) error {
	switch concrete := client.(type) {
	case *redis.Client:
		if concrete != nil && concrete.Options().MaxRetries == 0 && concrete.Options().ContextTimeoutEnabled {
			return nil
		}
	case *redis.ClusterClient:
		if concrete != nil && concrete.Options().MaxRetries == -1 && concrete.Options().MaxRedirects == 0 &&
			concrete.Options().ContextTimeoutEnabled {

			return nil
		}
	}

	return errors.New("rate limiter requires a Redis client with command retries disabled and context timeouts enabled")
}

// NewLimiter creates a handle without resetting the named Redis quota.
// Cross-process rule conflicts are detected by Allow and Wait.
func (handler *Handler) NewLimiter(name string, rule ratelimit.Rule) (ratelimit.ILimiter, error) {
	if name == "" {
		return nil, errors.New("limiter name is empty")
	}
	if err := rule.Validate(); err != nil {
		return nil, fmt.Errorf("invalid limiter rule: %w", err)
	}

	handler.mu.Lock()
	defer handler.mu.Unlock()
	if existing, exists := handler.rules[name]; exists && existing != rule {
		return nil, fmt.Errorf("limiter %q: %w", name, ratelimit.ErrRuleConflict)
	}
	handler.rules[name] = rule

	return &limiter{
		client: handler.client,
		key:    handler.namespace + ":ratelimit:v1:" + base64.RawURLEncoding.EncodeToString([]byte(name)),
		rule:   rule,
	}, nil
}

type limiter struct {
	client redis.UniversalClient
	key    string
	rule   ratelimit.Rule
}

func (lim *limiter) Allow(ctx context.Context) (bool, error) {
	delay, err := lim.acquire(ctx)
	if err != nil {
		return false, err
	}

	return delay == 0, nil
}

func (lim *limiter) Wait(ctx context.Context) error {
	for {
		delay, err := lim.acquire(ctx)
		if err != nil {
			return err
		}
		if delay == 0 {
			return nil
		}

		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()

			return ctx.Err()
		case <-timer.C:
		}
	}
}

func (lim *limiter) acquire(ctx context.Context) (time.Duration, error) {
	if ctx == nil {
		return 0, errors.New("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	result, err := lim.client.Eval(ctx, acquireScript, []string{lim.key},
		lim.rule.Limit, lim.rule.Interval.Microseconds(), lim.rule.Burst).Int64()
	if err != nil {
		return 0, fmt.Errorf("failed to acquire Redis quota: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	if result == -1 {
		return 0, ratelimit.ErrRuleConflict
	}
	if result < 0 {
		return 0, errors.New("invalid Redis quota state or clock regression")
	}

	return time.Duration(result) * time.Microsecond, nil
}
