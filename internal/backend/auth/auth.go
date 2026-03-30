/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package auth provider the interface to check permission.
package auth

import "github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"

// IAuthorizer defines the contract for permission checking.
type IAuthorizer interface {
	// Check verifies whether the current request context can perform the action on resources.
	Check(ctx contextx.IContext, action Action, resources []Resource) error
	// CheckMany verifies multiple actions over their corresponding resources and aggregates denied actions.
	CheckMany(ctx contextx.IContext, actionResources map[Action][]Resource) error
}
