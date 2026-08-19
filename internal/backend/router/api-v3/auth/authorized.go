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

package auth

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Authorized queries the authorized resource scope for requested action-resource type pairs.
func (h *handler) Authorized(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.AuthorizedReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to query authorized scope, failed to decode request body")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	scopes, err := h.queryAuthorizedScopes(rCtx, req.GetItems())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to query authorized scope")
		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	resp := new(protoBackend.AuthorizedResp)
	resp.ConvertResultsFromScopes(req.GetItems(), scopes)

	return resp.GetData(), nil
}

func (h *handler) queryAuthorizedScopes(rCtx restserver.IContext, items []*protoBackend.AuthorizedItem) ([]auth.AuthorizedScope, error) {
	scopes := make([]auth.AuthorizedScope, len(items))
	for i, item := range items {
		action := auth.Action(item.GetAction())
		resourceType := types.AuthResourceType(item.GetResourceType())

		scope, err := h.authorizer.ListAuthorizedInstances(rCtx, action, resourceType)
		if err != nil {
			return nil, err
		}

		scopes[i] = scope
	}

	return scopes, nil
}
