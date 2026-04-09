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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Validate checks the validity of the request parameters.
func (x *AuthVerifyReq) Validate() error {
	for i, item := range x.GetItems() {
		if item.GetAction() == "" {
			return fmt.Errorf("items[%d].action is required", i)
		}
	}

	return nil
}

// AutoConvert is a no-op for this request type, as it does not require any conversion.
func (x *AuthVerifyReq) AutoConvert() {}

// ConvertItemsToTypes converts the request items to internal types for backend processing.
func (x *AuthVerifyReq) ConvertItemsToTypes() []*types.AuthVerifyItem {
	return conv.SliceToSlice(x.GetItems(), func(item *AuthVerifyItem) *types.AuthVerifyItem {
		return &types.AuthVerifyItem{
			Action:    item.GetAction(),
			Resources: convertAuthResourcesToTypes(item.GetResources()),
		}
	})
}

// ConvertResultsFromTypes populates the response data from IAM verify results.
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

// Validate checks the validity of the request parameters.
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

// AutoConvert is a no-op for this request type, as it does not require any conversion.
func (x *AuthorizedReq) AutoConvert() {}

// ConvertItemsToTypes converts the request items to internal types for backend processing.
func (x *AuthorizedReq) ConvertItemsToTypes() []*types.AuthorizedItem {
	return conv.SliceToSlice(x.GetItems(), func(item *AuthorizedItem) *types.AuthorizedItem {
		return &types.AuthorizedItem{
			Action:       item.GetAction(),
			ResourceType: types.AuthResourceType(item.GetResourceType()),
		}
	})
}

// ConvertResultsFromTypes populates the response data from IAM verify results.
func (x *AuthorizedResp) ConvertResultsFromTypes(results []*types.AuthorizedResult) {
	x.Data = &AuthorizedResp_Data{
		Results: conv.SliceToSlice(results, func(result *types.AuthorizedResult) *AuthorizedResult {
			return &AuthorizedResult{
				Action:       result.Action,
				ResourceType: string(result.ResourceType),
				IsAny:        result.IsAny,
				Resources:    convertAuthResourceFromTypes(result.Resources),
			}
		}),
	}
}

func convertAuthResourcesToTypes(resources []*AuthResource) []types.AuthResource {
	return conv.SliceToSlice(resources, func(resource *AuthResource) types.AuthResource {
		return types.AuthResource{
			SystemID: resource.GetSystemId(),
			Type:     types.AuthResourceType(resource.GetType()),
			ID:       resource.GetId(),
		}
	})
}

func convertAuthResourceFromTypes(resources []types.AuthResource) []*AuthResource {
	return conv.SliceToSlice(resources, func(resource types.AuthResource) *AuthResource {
		return &AuthResource{
			SystemId: resource.SystemID,
			Type:     string(resource.Type),
			Id:       resource.ID,
		}
	})
}
