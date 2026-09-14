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

package iamv3

import (
	"errors"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/iam-go-sdk/expression"
)

// ErrDisabled identifies calls made while IAM v3 is disabled.
var ErrDisabled = errors.New("IAM v3 is disabled")

// DisabledHandler rejects every operation without contacting IAM v3.
type DisabledHandler struct{}

var _ IHandler = (*DisabledHandler)(nil)

// NewDisabledHandler returns an IHandler that rejects every IAM v3 operation.
func NewDisabledHandler() IHandler {
	return &DisabledHandler{}
}

// IsAllowed returns false and ErrDisabled.
func (h *DisabledHandler) IsAllowed(_ contextx.IContext, _ types.IAMCheckRequest) (bool, error) {
	return false, ErrDisabled
}

// IsAllowedWithCache returns false and ErrDisabled without accessing a cache.
func (h *DisabledHandler) IsAllowedWithCache(_ contextx.IContext, _ types.IAMCheckRequest, _ time.Duration) (bool, error) {
	return false, ErrDisabled
}

// BatchIsAllowed returns no authorization results and ErrDisabled.
func (h *DisabledHandler) BatchIsAllowed(
	_ contextx.IContext, _ types.IAMCheckRequest, _ [][]types.IAMResource,
) (map[string]bool, error) {
	return nil, ErrDisabled
}

// ResourceMultiActionsAllowed returns no authorization results and ErrDisabled.
func (h *DisabledHandler) ResourceMultiActionsAllowed(
	_ contextx.IContext, _ types.IAMMultiActionCheckRequest,
) (map[string]bool, error) {
	return nil, ErrDisabled
}

// BatchResourceMultiActionsAllowed returns no authorization results and ErrDisabled.
func (h *DisabledHandler) BatchResourceMultiActionsAllowed(
	_ contextx.IContext, _ types.IAMMultiActionCheckRequest, _ [][]types.IAMResource,
) (map[string]map[string]bool, error) {
	return nil, ErrDisabled
}

// GetToken returns an empty token and ErrDisabled.
func (h *DisabledHandler) GetToken(_ contextx.IContext) (string, error) {
	return "", ErrDisabled
}

// IsBasicAuthAllowed returns ErrDisabled regardless of the credentials.
func (h *DisabledHandler) IsBasicAuthAllowed(_ contextx.IContext, _, _ string) error {
	return ErrDisabled
}

// GetApplyURL returns an empty URL and ErrDisabled.
func (h *DisabledHandler) GetApplyURL(_ contextx.IContext, _ types.IAMApplyRequest) (string, error) {
	return "", ErrDisabled
}

// GetPolicyExpression returns no policy expression and ErrDisabled.
func (h *DisabledHandler) GetPolicyExpression(
	_ contextx.IContext, _ types.IAMAuthorizedInstancesRequest,
) (*expression.ExprCell, error) {
	return nil, ErrDisabled
}

// ListAuthorizedInstances returns no authorized instances and ErrDisabled.
func (h *DisabledHandler) ListAuthorizedInstances(
	_ contextx.IContext, _ types.IAMAuthorizedInstancesRequest,
) (bool, []types.IAMResource, error) {
	return false, nil, ErrDisabled
}
