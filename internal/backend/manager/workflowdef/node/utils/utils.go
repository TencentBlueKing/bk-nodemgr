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
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/discover"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/installer"
)

const (
	// DefaultEndpointSelectionCount is the default count when selecting endpoints.
	DefaultEndpointSelectionCount = 3
)

// BuildServerURLs builds a comma-separated string of multiple addresses in the format: http://addr1,http://addr2,http://addr3.
// The installer already supports this format and will use utils.SplitServerAddrs() to split and process it.
// If endpoints is empty, returns an empty string.
func BuildServerURLs(endpoints ...discover.Endpoint) string {
	if len(endpoints) == 0 {
		return ""
	}

	addrs := make([]string, 0, len(endpoints))
	for _, ep := range endpoints {
		addr := ep.GetIPV4Address()
		if addr != "" {
			addrs = append(addrs, "http://"+addr)
		}
	}

	return strings.Join(addrs, installer.ServerAddrSeparator)
}
