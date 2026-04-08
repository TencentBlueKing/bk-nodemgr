/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package auth

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

type noopAuthorizer struct{}

// NewNoOpAuthorizer creates an IAuthorizer that always grants permission.
func NewNoOpAuthorizer() IAuthorizer {
	return &noopAuthorizer{}
}

func (a *noopAuthorizer) Check(_ contextx.IContext, _ Action, _ []types.AuthResource) error {
	return nil
}

func (a *noopAuthorizer) CheckMany(_ contextx.IContext, _ map[Action][]types.AuthResource) error {
	return nil
}

func (a *noopAuthorizer) ListAuthorizedInstances(
	_ contextx.IContext, _ Action, _ types.AuthResourceType,
) (AuthorizedScope, error) {

	return AuthorizedScope{IsAny: true, Resources: []types.AuthResource{}}, nil
}
