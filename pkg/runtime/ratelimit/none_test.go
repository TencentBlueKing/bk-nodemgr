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

package ratelimit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNoneLimiter(t *testing.T) {
	lim, err := NewNone().NewLimiter("quota", NoneRule{})
	require.NoError(t, err)
	for range 100 {
		allowed, err := lim.Allow(t.Context())
		require.NoError(t, err)
		require.True(t, allowed)
		require.NoError(t, lim.Wait(t.Context()))
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	allowed, err := lim.Allow(ctx)
	require.ErrorIs(t, err, context.Canceled)
	require.False(t, allowed)
	require.ErrorIs(t, lim.Wait(ctx), context.Canceled)
}

func TestNoneHandlerRejectsFiniteRules(t *testing.T) {
	handler := NewNone()
	_, err := handler.NewLimiter("quota", TokenBucketRule{Limit: 1, Interval: time.Second, Burst: 1})
	require.Error(t, err)
	_, err = handler.NewLimiter("quota", nil)
	require.Error(t, err)
	_, err = handler.NewLimiter("", NoneRule{})
	require.Error(t, err)
}
