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

// ConvertParamFromTypes populates the request items from IAM check requests.
func (x *AuthVerifyReq) ConvertParamFromTypes(items []*types.AuthVerifyItem) {
	x.Items = make([]*AuthVerifyItem, 0, len(items))
	for _, item := range items {
		x.Items = append(x.Items, &AuthVerifyItem{
			Action:    item.Action,
			Resources: convertAuthResourceFromTypes(item.Resources),
		})
	}
}

// ConvertItemsToActionResources converts verify request items to action-resource map.
func (x *AuthVerifyReq) ConvertItemsToActionResources() map[auth.Action][]types.AuthResource {
	items := x.GetItems()
	if len(items) == 0 {
		return nil
	}

	actionResources := make(map[auth.Action][]types.AuthResource, len(items))
	for _, item := range items {
		action := auth.Action(item.GetAction())
		resources := convertAuthResourceToTypes(item.GetResources())
		actionResources[action] = append(actionResources[action], resources...)
	}

	return actionResources
}

// ConvertResultToTypes converts the response data to auth verify results.
func (x *AuthVerifyResp) ConvertResultToTypes() []*types.AuthVerifyResult {
	data := x.GetData()
	if data == nil {
		return nil
	}

	results := data.GetResults()
	checkResults := make([]*types.AuthVerifyResult, 0, len(results))
	for _, result := range results {
		checkResults = append(checkResults, &types.AuthVerifyResult{
			Action:     result.GetAction(),
			Authorized: result.GetAuthorized(),
		})
	}

	return checkResults
}

// ConvertResultsFromTypes populates the response data from auth verify results.
func (x *AuthVerifyResp) ConvertResultsFromTypes(results []*types.AuthVerifyResult) {
	x.Data = &AuthVerifyResp_Data{
		Results: conv.SliceToSlice(results, func(result *types.AuthVerifyResult) *AuthVerifyResult {
			return &AuthVerifyResult{
				Action:     result.Action,
				Authorized: result.Authorized,
			}
		}),
	}
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

// ConvertParamFromAuthorizedInstances populates the request items from IAM authorized instances requests.
func (x *AuthorizedReq) ConvertParamFromTypes(items []*types.AuthorizedItem) {
	x.Items = conv.SliceToSlice(items, func(item *types.AuthorizedItem) *AuthorizedItem {
		return &AuthorizedItem{
			Action:       item.Action,
			ResourceType: string(item.ResourceType),
		}
	})
}

// ConvertResultToTypes converts the response data to authorized results.
func (x *AuthorizedResp) ConvertResultToTypes() []*types.AuthorizedResult {
	data := x.GetData()
	if data == nil {
		return nil
	}

	results := make([]*types.AuthorizedResult, len(data.GetResults()))
	for i, result := range data.GetResults() {
		results[i] = &types.AuthorizedResult{
			Action:       result.GetAction(),
			ResourceType: types.AuthResourceType(result.GetResourceType()),
			IsAny:        result.GetIsAny(),
			Resources:    convertAuthResourceToTypes(result.GetResources()),
		}
	}

	return results
}

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
			Resources:    convertAuthResourceFromTypes(scopes[i].Resources),
		}
	}

	x.Data = &AuthorizedResp_Data{
		Results: results,
	}
}

// convertAuthResourceToTypes converts proto auth resources to types auth resources.
func convertAuthResourceToTypes(resources []*AuthResource) []types.AuthResource {
	if len(resources) == 0 {
		return nil
	}

	result := make([]types.AuthResource, 0, len(resources))
	for _, resource := range resources {
		result = append(result, types.AuthResource{
			SystemID: resource.GetSystemId(),
			Type:     types.AuthResourceType(resource.GetType()),
			ID:       resource.GetId(),
		})
	}

	return result
}

// convertAuthResourceFromTypes converts types auth resources to proto auth resources.
func convertAuthResourceFromTypes(resources []types.AuthResource) []*AuthResource {
	if len(resources) == 0 {
		return nil
	}

	backendResources := make([]*AuthResource, len(resources))
	for i, resource := range resources {
		backendResources[i] = &AuthResource{
			SystemId: resource.SystemID,
			Type:     string(resource.Type),
			Id:       resource.ID,
		}
	}

	return backendResources
}
