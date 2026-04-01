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

func (x *AuthVerifyReq) Validate() error {
	for i, item := range x.GetItems() {
		if item.GetAction() == "" {
			return fmt.Errorf("items[%d].action is required", i)
		}
	}

	return nil
}

func (x *AuthVerifyReq) AutoConvert() {}

func (x *AuthVerifyReq) ConvertItemsToTypes() []*types.IAMCheckRequest {
	return conv.SliceToSlice(x.GetItems(), func(item *AuthVerifyItem) *types.IAMCheckRequest {
		return &types.IAMCheckRequest{
			ActionID:  item.GetAction(),
			Resources: convertAuthResourcesToTypes(item.GetResources()),
		}
	})
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

func convertAuthResourcesToTypes(resources []*AuthResource) []types.IAMResource {
	return conv.SliceToSlice(resources, func(resource *AuthResource) types.IAMResource {
		return types.IAMResource{
			SystemID: resource.GetSystemId(),
			Type:     resource.GetType(),
			ID:       resource.GetId(),
		}
	})
}
