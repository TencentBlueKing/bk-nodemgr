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
	"fmt"

	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
)

// PermissionDeniedError indicates the user has no permission for the requested actions.
type PermissionDeniedError struct {
	ApplyURL   string
	SystemID   string
	SystemName string
	Actions    []ActionInfo
}

// Error implements the error interface.
func (permErr PermissionDeniedError) Error() string {
	if permErr.ApplyURL != "" {
		return fmt.Sprintf("permission denied, apply at: %s", permErr.ApplyURL)
	}

	return "permission denied"
}

// ActionInfo represents action metadata used for permission-apply responses.
type ActionInfo struct {
	ID                   string
	Name                 string
	RelatedResourceTypes []RelatedResourceType
}

// RelatedResourceType represents an IAM related resource type.
type RelatedResourceType struct {
	SystemID string
	Type     string
	TypeName string
}

// PermissionData converts PermissionDeniedError into a restserver.Permission payload.
// This satisfies the restserver.PermissionProvider interface.
func (permErr PermissionDeniedError) PermissionData() resterrf.Permission {
	actions := make([]resterrf.Action, 0, len(permErr.Actions))
	for _, act := range permErr.Actions {
		rts := make([]resterrf.RelatedResourceType, 0, len(act.RelatedResourceTypes))
		for _, rt := range act.RelatedResourceTypes {
			rts = append(rts, resterrf.RelatedResourceType{
				SystemID: rt.SystemID,
				Type:     rt.Type,
				TypeName: rt.TypeName,
			})
		}
		actions = append(actions, resterrf.Action{
			ID:                   act.ID,
			Name:                 act.Name,
			RelatedResourceTypes: rts,
		})
	}

	return resterrf.Permission{
		System:     permErr.SystemID,
		SystemName: permErr.SystemName,
		ApplyURL:   permErr.ApplyURL,
		Actions:    actions,
	}
}
