/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package backend

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IHandlerAuth defines the backend auth verification capability exposed to callers.
type IHandlerAuth interface {
	VerifyAuth(nCtx contextx.IContext, req []*types.IAMCheckRequest) ([]*types.IAMCheckResult, error)
}

// VerifyAuth proxies proactive auth verification requests to the backend service.
func (h *Handler) VerifyAuth(nCtx contextx.IContext, req []*types.IAMCheckRequest) ([]*types.IAMCheckResult, error) {
	backendReq := new(protoBackend.AuthVerifyReq)
	backendReq.Items = convertIAMCheckRequestsToBackend(req)

	resp, err := h.cli.verifyAuth(nCtx, backendReq)
	if err != nil {
		return nil, err
	}

	if resp.GetData() == nil {
		return nil, nil
	}

	return convertAuthVerifyResultsToTypes(resp.GetData().GetResults()), nil
}

func convertIAMCheckRequestsToBackend(items []*types.IAMCheckRequest) []*protoBackend.AuthVerifyItem {
	if len(items) == 0 {
		return nil
	}

	backendItems := make([]*protoBackend.AuthVerifyItem, len(items))
	for i, item := range items {
		backendItems[i] = &protoBackend.AuthVerifyItem{
			Action:    item.ActionID,
			Resources: convertIAMResourcesToBackend(item.Resources),
		}
	}

	return backendItems
}

func convertIAMResourcesToBackend(resources []types.IAMResource) []*protoBackend.AuthResource {
	if len(resources) == 0 {
		return nil
	}

	backendResources := make([]*protoBackend.AuthResource, len(resources))
	for i, resource := range resources {
		backendResources[i] = &protoBackend.AuthResource{
			SystemId: resource.SystemID,
			Type:     resource.Type,
			Id:       resource.ID,
		}
	}

	return backendResources
}

func convertAuthVerifyResultsToTypes(results []*protoBackend.AuthVerifyResult) []*types.IAMCheckResult {
	if len(results) == 0 {
		return nil
	}

	checkResults := make([]*types.IAMCheckResult, len(results))
	for i, result := range results {
		checkResults[i] = &types.IAMCheckResult{
			ActionID:   result.GetAction(),
			Authorized: result.GetAuthorized(),
		}
	}

	return checkResults
}
