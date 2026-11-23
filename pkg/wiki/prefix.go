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

package wiki

import (
	"fmt"
	"net/url"
	"strings"
)

// PrefixConfig configures a user-accessible Wiki root URL, including hash-route prefixes.
type PrefixConfig struct {
	PrefixURL string
}

// Validate checks the document prefix without network access.
func (conf PrefixConfig) Validate() error {
	prefix, err := url.Parse(conf.PrefixURL)
	if err != nil {
		return fmt.Errorf("invalid wiki prefix url: %w", err)
	}
	if (prefix.Scheme != "https" && prefix.Scheme != "http") || prefix.Host == "" ||
		prefix.User != nil || prefix.RawQuery != "" || prefix.ForceQuery {

		return fmt.Errorf("wiki prefix url must be an absolute http(s) address without credentials or query")
	}

	return nil
}

var _ IHandler = (*prefixHandler)(nil)

type prefixHandler struct {
	prefix string
}

func newPrefixHandler(conf PrefixConfig) *prefixHandler {
	return &prefixHandler{prefix: strings.TrimRight(conf.PrefixURL, "/") + "/"}
}

// URL appends the language and document path without adding a version tag.
func (h *prefixHandler) URL(document Document, language Language) string {
	return h.prefix + string(language) + "/" + string(document)
}
