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
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// NoOpHandler is a full bypass implementation of IHandler used when IAM v3 is
// disabled. Every permission check returns "allowed", GetToken and GetApplyURL
// return empty strings, and no IAM service is ever contacted. Callers that
// instantiate NoOpHandler accept that all authorization decisions are bypassed;
// this is NOT a degraded or partial IAM mode.
type NoOpHandler struct{}

// Verify that NoOpHandler implements IHandler interface.
var _ IHandler = (*NoOpHandler)(nil)

// NewNoOpHandler returns an IHandler that bypasses all IAM authorization. It
// logs a warning at construction time so operators are aware that permission
// enforcement is completely disabled for this process.
func NewNoOpHandler() IHandler {
	logger.G.Sys().Warn("IAM v3 is disabled, all permission checks will be skipped")
	return &NoOpHandler{}
}

// IsAllowed unconditionally returns true (bypass).
func (h *NoOpHandler) IsAllowed(_ contextx.IContext, _ types.IAMCheckRequest) (bool, error) {
	logger.G.Sys().Debug("IAM v3 no-op handler: skipping IsAllowed")
	return true, nil
}

// IsAllowedWithCache unconditionally returns true (bypass); the ttl is ignored.
func (h *NoOpHandler) IsAllowedWithCache(_ contextx.IContext, _ types.IAMCheckRequest, _ time.Duration) (bool, error) {
	logger.G.Sys().Debug("IAM v3 no-op handler: skipping IsAllowedWithCache")
	return true, nil
}

// BatchIsAllowed returns true for every resource set (bypass).
func (h *NoOpHandler) BatchIsAllowed(_ contextx.IContext, _ types.IAMCheckRequest,
	resourcesList [][]types.IAMResource,
) (map[string]bool, error) {

	logger.G.Sys().Debug("IAM v3 no-op handler: skipping BatchIsAllowed")
	results := make(map[string]bool, len(resourcesList))

	for _, resources := range resourcesList {
		key := buildResourceID(toWireResources(resources))
		results[key] = true
	}

	return results, nil
}

// ResourceMultiActionsAllowed returns true for every action (bypass).
func (h *NoOpHandler) ResourceMultiActionsAllowed(
	_ contextx.IContext, request types.IAMMultiActionCheckRequest,
) (map[string]bool, error) {

	logger.G.Sys().Debug("IAM v3 no-op handler: skipping ResourceMultiActionsAllowed")
	results := make(map[string]bool, len(request.ActionIDs))

	for _, actionID := range request.ActionIDs {
		results[actionID] = true
	}

	return results, nil
}

// BatchResourceMultiActionsAllowed returns true for every action on every
// resource set (bypass).
func (h *NoOpHandler) BatchResourceMultiActionsAllowed(
	_ contextx.IContext, request types.IAMMultiActionCheckRequest, resourcesList [][]types.IAMResource,
) (map[string]map[string]bool, error) {

	logger.G.Sys().Debug("IAM v3 no-op handler: skipping BatchResourceMultiActionsAllowed")
	results := make(map[string]map[string]bool, len(resourcesList))
	for _, resources := range resourcesList {
		resourceKey := buildResourceID(toWireResources(resources))
		actionResults := make(map[string]bool, len(request.ActionIDs))
		for _, actionID := range request.ActionIDs {
			actionResults[actionID] = true
		}
		results[resourceKey] = actionResults
	}

	return results, nil
}

// GetToken returns an empty string; no IAM service is contacted in bypass mode.
func (h *NoOpHandler) GetToken(_ contextx.IContext) (string, error) {
	logger.G.Sys().Debug("IAM v3 no-op handler: skipping GetToken")
	return "", nil
}

// IsBasicAuthAllowed unconditionally succeeds (bypass).
func (h *NoOpHandler) IsBasicAuthAllowed(_ contextx.IContext, _, _ string) error {
	logger.G.Sys().Debug("IAM v3 no-op handler: skipping IsBasicAuthAllowed")
	return nil
}

// GetApplyURL returns an empty string; no IAM service is contacted in bypass mode.
func (h *NoOpHandler) GetApplyURL(_ contextx.IContext, _ types.IAMApplyRequest) (string, error) {
	logger.G.Sys().Debug("IAM v3 no-op handler: skipping GetApplyURL")
	return "", nil
}
