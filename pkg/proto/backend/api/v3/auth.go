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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
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

func (x *AuthVerifyResp) ConvertResultsFromTypes(results []*types.IAMCheckResult) {
	x.Data = &AuthVerifyResp_Data{
		Results: conv.SliceToSlice(results, func(result *types.IAMCheckResult) *AuthVerifyResult {
			return &AuthVerifyResult{
				Action:     result.ActionID,
				Authorized: result.Authorized,
			}
		}),
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

// ConvertItemsToActionResources converts verify request items to action-resource map.
func (x *AuthVerifyReq) ConvertItemsToActionResources() map[auth.Action][]auth.Resource {
	items := x.GetItems()
	if len(items) == 0 {
		return nil
	}

	actionResources := make(map[auth.Action][]auth.Resource, len(items))
	for _, item := range items {
		action := auth.Action(item.GetAction())
		resources := ConvertAuthResourcesToInternal(item.GetResources())
		actionResources[action] = append(actionResources[action], resources...)
	}

	return actionResources
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

// Validate validates the authorized request body.
func (x *AuthorizedReq) Validate() error {
	for i, item := range x.GetItems() {
		if item.GetAction() == "" {
			return fmt.Errorf("items[%d].action is required", i)
		}
		if item.GetResourceType() == "" {
			return fmt.Errorf("items[%d].resource_type is required", i)
		}
	}

	return nil
}

// AutoConvert is a no-op for this request type.
func (x *AuthorizedReq) AutoConvert() {}

// ConvertResultsFromScopes populates the response data from authorized scopes.
func (x *AuthorizedResp) ConvertResultsFromScopes(
	items []*AuthorizedItem,
	scopes []auth.AuthorizedScope,
) {
	if len(items) != len(scopes) {
		return
	}

	results := make([]*AuthorizedResult, len(items))
	for i, item := range items {
		results[i] = &AuthorizedResult{
			Action:       item.GetAction(),
			ResourceType: item.GetResourceType(),
			IsAny:        scopes[i].IsAny,
			Resources:    ConvertInternalToAuthResources(scopes[i].Resources),
		}
	}

	x.Data = &AuthorizedResp_Data{
		Results: results,
	}
}

// ConvertInternalToAuthResources converts internal auth resources to proto auth resources.
func ConvertInternalToAuthResources(resources []auth.Resource) []*AuthResource {
	if len(resources) == 0 {
		return nil
	}

	protoResources := make([]*AuthResource, len(resources))
	for i, r := range resources {
		protoResources[i] = &AuthResource{
			SystemId: r.SystemID,
			Type:     string(r.Type),
			Id:       r.ID,
		}
	}

	return protoResources
}
