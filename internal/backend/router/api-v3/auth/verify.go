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
	"errors"
	"sort"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Verify checks whether the current user has permissions for requested actions.
func (h *handler) Verify(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.AuthVerifyReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to verify auth, failed to decode request body")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	results, err := h.verifyItems(rCtx, req.ConvertItemsToActionResources())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to verify auth")

		var permErr auth.PermissionDeniedError
		if errors.As(err, &permErr) {
			return nil, resterrf.ErrWrap(resterrf.PermissionDenied, err)
		}

		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	resp := new(protoBackend.AuthVerifyResp)
	resp.ConvertResultsFromTypes(results)

	return resp.GetData(), nil
}

func (h *handler) verifyItems(
	rCtx restserver.IContext, actionResources map[auth.Action][]types.AuthResource,
) ([]*types.AuthVerifyResult, error) {

	err := h.authorizer.CheckMany(rCtx, actionResources)

	if err != nil {
		return nil, err
	}

	return buildVerifyResults(actionResources), nil
}

func buildVerifyResults(actionResources map[auth.Action][]types.AuthResource) []*types.AuthVerifyResult {
	if len(actionResources) == 0 {
		return nil
	}

	results := make([]*types.AuthVerifyResult, 0, len(actionResources))
	actions := make([]string, 0, len(actionResources))
	for action := range actionResources {
		actions = append(actions, string(action))
	}
	sort.Strings(actions)

	for _, action := range actions {
		results = append(results, &types.AuthVerifyResult{
			Action:     action,
			Authorized: true,
		})
	}

	return results
}
