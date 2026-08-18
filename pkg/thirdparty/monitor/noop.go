/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package monitor

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
)

// NoOpHandler is used when monitor event data-id integration is disabled.
type NoOpHandler struct{}

var _ IHandler = (*NoOpHandler)(nil)

// NewNoOpHandler returns a monitor handler that never sends outbound requests.
func NewNoOpHandler() IHandler {
	logger.G.Sys().Warn("Monitor is disabled, event data-id creation will be skipped")
	return &NoOpHandler{}
}

// GetOrCreateAgentEventDataID returns no value so callers can apply their defaults.
func (h *NoOpHandler) GetOrCreateAgentEventDataID(_ contextx.IContext, _ int64) (int64, bool, error) {
	logger.G.Sys().Debug("Monitor no-op handler: skipping GetOrCreateAgentEventDataID")
	return 0, false, nil
}
