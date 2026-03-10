/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package utils use to provide some common utils for node actions.
package utils

import (
	"math/rand/v2"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/discover"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/installer"
)

const (
	// DefaultEndpointSelectionCount is the default count when selecting endpoints.
	DefaultEndpointSelectionCount = 3
)

func buildServerURL(needV4, needV6 bool, endpoints ...discover.Endpoint) []string {
	if len(endpoints) == 0 {
		return nil
	}

	addrs := make([]string, 0, len(endpoints))
	for _, ep := range endpoints {
		if needV4 && ep.IPV4 != "" {
			addr := ep.GetIPV4Address()
			if addr != "" {
				addrs = append(addrs, "http://"+addr)
			}
		}

		if needV6 && ep.IPV6 != "" {
			addr := ep.GetIPV6Address()
			if addr != "" {
				addrs = append(addrs, "http://"+addr)
			}
		}
	}

	if len(addrs) == 0 {
		return nil
	}

	rand.Shuffle(len(addrs), func(i, j int) {
		addrs[i], addrs[j] = addrs[j], addrs[i]
	})

	return addrs
}

// BuildServerURLs builds a comma-separated string of multiple addresses in the format: http://addr1,http://addr2,http://addr3.
// The installer already supports this format and will use utils.SplitServerAddrs() to split and process it.
// If endpoints is empty, returns an empty string.
func BuildServerURLs(endpoints ...discover.Endpoint) string {
	addrs := buildServerURL(true, true, endpoints...)
	if len(addrs) == 0 {
		return ""
	}

	return strings.Join(addrs, installer.ServerAddrSeparator)
}

// SelectOneServerV4URL selects one ipv4 address from the given endpoints.
func SelectOneServerV4URL(endpoints []discover.Endpoint) string {
	addrs := buildServerURL(true, false, endpoints...)
	if len(addrs) == 0 {
		return ""
	}

	return addrs[0]
}

// SelectOneServerV6URL selects one ipv6 address from the given endpoints.
func SelectOneServerV6URL(endpoints []discover.Endpoint) string {
	addrs := buildServerURL(false, true, endpoints...)
	if len(addrs) == 0 {
		return ""
	}

	return addrs[0]
}
