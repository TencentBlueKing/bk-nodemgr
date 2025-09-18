/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package discover ...
package discover

import (
	"math/rand"
	"sort"
	"sync"
)

// RandomSelector this defines the random selector.
type RandomSelector struct{}

// Select select one instance from the endpoints.
func (r *RandomSelector) Select(endpoints []Endpoint) (Endpoint, error) {
	length := len(endpoints)

	if length == 0 {
		return Endpoint{}, ErrEndpointNotFound()
	}

	// this just is a simple random selector.
	// nolint: gosec
	idx := rand.Intn(length)

	return endpoints[idx], nil
}

// NewRandomSelector creates a new random selector.
func NewRandomSelector() *RandomSelector {
	return &RandomSelector{}
}

// RoundRobinSelector this defines the round robin selector.
type RoundRobinSelector struct {
	mutex  sync.Mutex
	cursor int
}

// Select select one instance from the endpoints.
func (r *RoundRobinSelector) Select(endpoints []Endpoint) (Endpoint, error) {
	length := len(endpoints)

	if length == 0 {
		return Endpoint{}, ErrEndpointNotFound()
	}

	if length == 1 {
		return endpoints[0], nil
	}

	sort.Slice(endpoints, func(i, j int) bool {
		iAddr := endpoints[i].GetIPV4Address()
		jAddr := endpoints[j].GetIPV4Address()

		if iAddr == jAddr {
			return endpoints[i].GetIPV6Address() < endpoints[j].GetIPV6Address()
		}

		return iAddr < jAddr
	})

	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.cursor >= length {
		r.cursor = 0
	}

	idx := r.cursor

	r.cursor++

	return endpoints[idx], nil
}

// NewRoundRobinSelector creates a new round robin selector.
func NewRoundRobinSelector() *RoundRobinSelector {
	return &RoundRobinSelector{
		cursor: 0,
	}
}
