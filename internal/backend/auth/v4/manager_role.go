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

package v4

import (
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/iamv4"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Management grants expire after the IAM V4 maximum lifetime, without automatic renewal.
const managerGrantLifetime = 365 * 24 * time.Hour

const (
	roleNetworkAreaManager = "networkarea_manager"
	roleNetworkUnitManager = "networkunit_manager"
)

type managerRoleGranter struct {
	handler iamv4.IHandlerRole
}

// NewManagerRoleGranter creates the IAM V4 management role grant adapter.
func NewManagerRoleGranter(handler iamv4.IHandlerRole) auth.IManagerRoleGranter {
	return &managerRoleGranter{handler: handler}
}

func (granter *managerRoleGranter) GrantManagerRole(ctx contextx.IContext, username string, resource types.AuthResource) error {
	if resource.SystemID != types.SystemIDNodeMgr {
		return fmt.Errorf("unsupported management grant resource system %q", resource.SystemID)
	}
	var roleID string
	switch resource.Type {
	case types.AuthResourceTypeNetworkArea:
		roleID = roleNetworkAreaManager
	case types.AuthResourceTypeNetworkUnit:
		roleID = roleNetworkUnitManager
	default:
		return fmt.Errorf("unsupported management grant resource type %q", resource.Type)
	}

	if err := granter.handler.GrantRole(ctx, types.IAMRoleGrantRequest{
		Username: username, RoleID: roleID, Resource: resource,
		ExpiresAt: time.Now().Add(managerGrantLifetime),
	}); err != nil {
		return fmt.Errorf("failed to grant resource management permissions: %w", err)
	}

	return nil
}
