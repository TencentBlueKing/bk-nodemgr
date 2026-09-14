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

package iamv4

import (
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// ErrDisabled identifies calls made while IAM v4 is disabled.
var ErrDisabled = errors.New("IAM v4 is disabled")

// DisabledHandler rejects every operation without contacting IAM v4.
type DisabledHandler struct{}

var _ IHandler = (*DisabledHandler)(nil)

// NewDisabledHandler returns an IHandler that rejects every IAM v4 operation.
func NewDisabledHandler() IHandler {
	return &DisabledHandler{}
}

// IsAllowed returns false and ErrDisabled.
func (h *DisabledHandler) IsAllowed(_ contextx.IContext, _ types.IAMCheckRequest) (bool, error) {
	return false, ErrDisabled
}

// ResourcesAllowed returns no authorization results and ErrDisabled.
func (h *DisabledHandler) ResourcesAllowed(_ contextx.IContext, _ types.IAMCheckRequest) (map[string]bool, error) {
	return nil, ErrDisabled
}

// ActionsAllowed returns no authorization results and ErrDisabled.
func (h *DisabledHandler) ActionsAllowed(_ contextx.IContext, _ types.IAMMultiActionCheckRequest) (map[string]bool, error) {
	return nil, ErrDisabled
}

// ListAuthorizedResources returns no authorized resources and ErrDisabled.
func (h *DisabledHandler) ListAuthorizedResources(
	_ contextx.IContext, _ types.IAMAuthorizedInstancesRequest,
) ([]AuthorizedResourceResponse, error) {
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

// GrantRole returns ErrDisabled without granting permissions.
func (h *DisabledHandler) GrantRole(_ contextx.IContext, _ types.IAMRoleGrantRequest) error {
	return ErrDisabled
}
