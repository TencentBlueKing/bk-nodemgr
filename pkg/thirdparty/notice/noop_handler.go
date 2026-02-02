/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package notice

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// noopHandler is a no-operation implementation of the notice handler.
// It is used when the notice feature is disabled in configuration,
// providing safe interface implementation to avoid nil pointer panics.
type noopHandler struct{}

// NewNoopHandler creates a new noopHandler instance.
// It logs an info message indicating that the notice feature is disabled.
func NewNoopHandler() IHandler {
	logger.G.Sys().Info("Notice feature is disabled, using no-op handler")
	return &noopHandler{}
}

// GetCurrentAnnouncements returns an empty list when notice feature is disabled.
// It logs a debug message and returns successfully to avoid blocking business operations.
func (h *noopHandler) GetCurrentAnnouncements(_ contextx.IContext) ([]*types.Announcement, error) {
	logger.G.Sys().Debug("Notice no-op handler: skipping GetCurrentAnnouncements")
	return []*types.Announcement{}, nil
}

// Verify that noopHandler implements IHandler interface.
var _ IHandler = (*noopHandler)(nil)
