/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package v3

import (
	"fmt"

	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
)

// Validate validates the request body.
func (x *AuthVerifyReq) Validate() error {
	for i, item := range x.GetItems() {
		if item.GetAction() == "" {
			return fmt.Errorf("items[%d].action is required", i)
		}
	}

	return nil
}

// AutoConvert is a no-op for this request type.
func (x *AuthVerifyReq) AutoConvert() {}

// ConvertResultsFromVerify populates the response data from verification results.
func (x *AuthVerifyResp) ConvertResultsFromVerify(results []*AuthVerifyResult) {
	x.Data = &AuthVerifyResp_Data{
		Results: results,
	}
}

// SetPermissionFromErrf populates the response permission from resterrf.Permission.
func (x *AuthVerifyResp) SetPermissionFromErrf(perm *resterrf.Permission) {
	if perm == nil {
		return
	}

	protoActions := make([]*Action, len(perm.Actions))
	for i, action := range perm.Actions {
		relatedTypes := make([]*RelatedResourceType, len(action.RelatedResourceTypes))
		for j, rt := range action.RelatedResourceTypes {
			relatedTypes[j] = &RelatedResourceType{
				SystemId: rt.SystemID,
				Type:     rt.Type,
				TypeName: rt.TypeName,
			}
		}

		protoActions[i] = &Action{
			Id:                   action.ID,
			Name:                 action.Name,
			RelatedResourceTypes: relatedTypes,
		}
	}

	x.Permission = &Permission{
		System:     perm.System,
		SystemName: perm.SystemName,
		ApplyUrl:   perm.ApplyURL,
		Actions:    protoActions,
	}
}

// NewAuthVerifyResult creates a new AuthVerifyResult.
func NewAuthVerifyResult(action string, authorized bool) *AuthVerifyResult {
	return &AuthVerifyResult{
		Action:     action,
		Authorized: authorized,
	}
}
