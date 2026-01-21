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

// RandomSelector is a random selector that randomly selects an endpoint on each call, without guaranteeing load balancing.
type RandomSelector struct{}

// Select randomly selects an endpoint from the endpoints list.
// Returns ErrEndpointNotFound error if the endpoints list is empty.
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

// NewRandomSelector creates and returns a new RandomSelector instance.
func NewRandomSelector() *RandomSelector {
	return &RandomSelector{}
}

// RoundRobinSelector is a round-robin selector that selects endpoints in order, ensuring load balancing.
// It is thread-safe and supports concurrent calls. It selects endpoints in order after sorting them,
// and restarts from the beginning when reaching the end of the list.
type RoundRobinSelector struct {
	mutex  sync.Mutex // mutex protecting the cursor field
	cursor int        // current round-robin position
}

// Select selects an endpoint from the endpoints list using round-robin.
// It sorts a copy of endpoints (based on IPV4 address, then IPV6 address) without modifying the original slice.
// Returns ErrEndpointNotFound error if the endpoints list is empty.
func (r *RoundRobinSelector) Select(endpoints []Endpoint) (Endpoint, error) {
	length := len(endpoints)

	if length == 0 {
		return Endpoint{}, ErrEndpointNotFound()
	}

	if length == 1 {
		return endpoints[0], nil
	}

	// Create a copy for sorting to avoid modifying the original slice
	sortedEndpoints := make([]Endpoint, length)
	copy(sortedEndpoints, endpoints)

	sort.Slice(sortedEndpoints, func(i, j int) bool {
		iAddr := sortedEndpoints[i].GetIPV4Address()
		jAddr := sortedEndpoints[j].GetIPV4Address()

		if iAddr == jAddr {
			return sortedEndpoints[i].GetIPV6Address() < sortedEndpoints[j].GetIPV6Address()
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

	return sortedEndpoints[idx], nil
}

// NewRoundRobinSelector creates and returns a new RoundRobinSelector instance.
func NewRoundRobinSelector() *RoundRobinSelector {
	return &RoundRobinSelector{
		cursor: 0,
	}
}

// SelectEndpoints selects a specified number of endpoints from all endpoints using the provided selector.
// The selector must be explicitly provided and cannot be nil.
func SelectEndpoints(endpoints []Endpoint, count int, selector Selector) ([]Endpoint, error) {
	if len(endpoints) == 0 {
		return nil, ErrEndpointNotFound()
	}

	if selector == nil {
		return nil, ErrInvalidSelector()
	}

	if count <= 0 || count >= len(endpoints) {
		return endpoints, nil
	}

	// Create and sort a copy to avoid modifying the original slice and avoid repeated sorting in the loop
	sortedEndpoints := make([]Endpoint, len(endpoints))
	copy(sortedEndpoints, endpoints)
	sort.Slice(sortedEndpoints, func(i, j int) bool {
		iAddr := sortedEndpoints[i].GetIPV4Address()
		jAddr := sortedEndpoints[j].GetIPV4Address()
		if iAddr == jAddr {
			return sortedEndpoints[i].GetIPV6Address() < sortedEndpoints[j].GetIPV6Address()
		}

		return iAddr < jAddr
	})
	selected := make([]Endpoint, 0, count)
	selectedMap := make(map[string]bool) // Used for deduplication, key is the endpoint address.

	for len(selected) < count {
		ep, err := selector.Select(sortedEndpoints)
		if err != nil {
			return nil, err
		}

		// Check if already selected (deduplication by address).
		addr := ep.GetIPV4Address()
		if !selectedMap[addr] {
			selectedMap[addr] = true
			selected = append(selected, ep)
		}

		// If all endpoints have been selected but the count is still insufficient, return what has been selected.
		if len(selectedMap) >= len(sortedEndpoints) {
			break
		}
	}

	return selected, nil
}
