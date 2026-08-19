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

package compatibility

import (
	"encoding/json"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
)

// ParseStoredPolicy parses a stored compatibility policy value and falls back to the default policy on invalid data.
func ParseStoredPolicy(raw string, nCtx contextx.IContext) Policy {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return DefaultPolicy()
	}
	if trimmed == "null" {
		logParseFallback(nCtx, "null payload", nil)
		return DefaultPolicy()
	}

	policy := Policy{}
	if hasMissingOrNullFields(trimmed) {
		logParseFallback(nCtx, "missing or null required fields", nil)
		return DefaultPolicy()
	}
	if err := json.Unmarshal([]byte(trimmed), &policy); err != nil {
		logParseFallback(nCtx, "malformed JSON", err)
		return DefaultPolicy()
	}

	policy = policy.Normalize()
	if err := policy.Validate(); err != nil {
		logParseFallback(nCtx, "invalid stored schema", err)
		return DefaultPolicy()
	}

	return policy
}

func hasMissingOrNullFields(raw string) bool {
	var payload map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return false
	}

	editedPlugins, ok := payload["enabled_plugins"]
	if !ok || string(editedPlugins) == "null" {
		return true
	}
	disabledBiz, ok := payload["disabled_biz"]
	if !ok || string(disabledBiz) == "null" {
		return true
	}

	return false
}

func logParseFallback(nCtx contextx.IContext, reason string, err error) {
	if nCtx != nil {
		logger.G.Biz(nCtx).
			With("reason", reason).
			WithErr(err).
			Warn("compatibility policy parse fallback to default")
		return
	}

	logger.G.Sys().
		With("reason", reason).
		WithErr(err).
		Warn("compatibility policy parse fallback to default")
}
