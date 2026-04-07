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

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
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
			instances := make([]*ResourceNode, len(rt.Instances))
			for k, node := range rt.Instances {
				instances[k] = &ResourceNode{
					Type:     node.Type,
					TypeName: node.TypeName,
					Id:       node.ID,
					Name:     node.Name,
				}
			}
			relatedTypes[j] = &RelatedResourceType{
				SystemId:   rt.SystemID,
				SystemName: rt.SystemName,
				Type:       rt.Type,
				TypeName:   rt.TypeName,
				Instances:  instances,
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

// AuthCheckItem describes an internal auth check item converted from verify request.
type AuthCheckItem struct {
	Action    auth.Action
	Resources []auth.Resource
}

// ConvertItemsToCheckItems converts verify request items to internal auth check items.
func (x *AuthVerifyReq) ConvertItemsToCheckItems() []AuthCheckItem {
	items := x.GetItems()
	if len(items) == 0 {
		return nil
	}

	checkItems := make([]AuthCheckItem, 0, len(items))
	for _, item := range items {
		checkItems = append(checkItems, AuthCheckItem{
			Action:    auth.Action(item.GetAction()),
			Resources: ConvertAuthResourcesToInternal(item.GetResources()),
		})
	}

	return checkItems
}

// ConvertItemsToVerifyResults converts verify request items to successful verify results.
func (x *AuthVerifyReq) ConvertItemsToVerifyResults() []*AuthVerifyResult {
	items := x.GetItems()
	if len(items) == 0 {
		return nil
	}

	results := make([]*AuthVerifyResult, 0, len(items))
	for _, item := range items {
		results = append(results, NewAuthVerifyResult(item.GetAction(), true))
	}

	return results
}

// ConvertAuthResourcesToInternal converts proto auth resources to internal auth resources.
func ConvertAuthResourcesToInternal(protoResources []*AuthResource) []auth.Resource {
	if len(protoResources) == 0 {
		return nil
	}

	resources := make([]auth.Resource, len(protoResources))
	for i, pr := range protoResources {
		resources[i] = auth.Resource{
			SystemID: pr.GetSystemId(),
			Type:     auth.ResourceType(pr.GetType()),
			ID:       pr.GetId(),
		}
	}

	return resources
}
