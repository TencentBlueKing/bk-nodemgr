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

package pluginv2

import (
	"fmt"
	"net/url"
)

const (
	// reportLogPath is the API path for reporting logs.
	reportLogPath = "/api/v3/callback/workflow/plugin/report_log"
)

// reportLogURLs builds report log URLs from multiple server addresses.
func reportLogURLs(callbackSvrAddrs []string) ([]string, error) {
	if len(callbackSvrAddrs) == 0 {
		return nil, fmt.Errorf("callback server addresses are empty")
	}

	urls := make([]string, 0, len(callbackSvrAddrs))
	for _, addr := range callbackSvrAddrs {
		if addr == "" {
			continue
		}
		fullURL, err := url.JoinPath(addr, reportLogPath)
		if err != nil {
			return nil, fmt.Errorf("failed to join URL path for address %s: %w", addr, err)
		}
		urls = append(urls, fullURL)
	}

	if len(urls) == 0 {
		return nil, fmt.Errorf("no valid callback server addresses provided")
	}

	return urls, nil
}
