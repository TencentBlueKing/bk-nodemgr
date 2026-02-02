/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package iamv3 provides the IAM v3 client implementation.
package iamv3

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
)

// NoOpHandler implements IHandler with no-op behavior.
type NoOpHandler struct {
	// No fields needed for no-op implementation
}

// Verify that NoOpHandler implements IHandler interface.
var _ IHandler = (*NoOpHandler)(nil)

// NewNoOpHandler creates a new no-op handler.
// It logs a warning message indicating that IAM v3 is disabled.
func NewNoOpHandler() IHandler {
	logger.G.Sys().Warn("IAM v3 is disabled, all permission checks will be skipped")
	return &NoOpHandler{}
}

// Note: When adding business methods to IHandler interface in the future,
// NoOpHandler must implement them with the following pattern:
// - Log at debug level: "IAM v3 no-op handler: skipping method_name"
// - Return success (nil error or true result)
// This ensures the no-op handler doesn't block business operations.
