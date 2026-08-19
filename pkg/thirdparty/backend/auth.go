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

package backend

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IHandlerAuth defines the backend auth verification capability exposed to callers.
type IHandlerAuth interface {
	// VerifyAuth proxies proactive auth verification requests to the backend service.
	VerifyAuth(nCtx contextx.IContext, req []*types.AuthVerifyItem) ([]*types.AuthVerifyResult, error)

	// Authorized checks whether the user has permissions for the specified actions and resources.
	Authorized(nCtx contextx.IContext, req []*types.AuthorizedItem) ([]*types.AuthorizedResult, error)
}

// VerifyAuth proxies proactive auth verification requests to the backend service.
func (h *Handler) VerifyAuth(nCtx contextx.IContext, req []*types.AuthVerifyItem) ([]*types.AuthVerifyResult, error) {
	backendReq := new(protoBackend.AuthVerifyReq)
	backendReq.ConvertParamFromTypes(req)

	resp, err := h.cli.verifyAuth(nCtx, backendReq)
	if err != nil {
		return nil, err
	}

	if resp.GetData() == nil {
		return nil, nil
	}

	return resp.ConvertResultToTypes(), nil
}

// Authorized checks whether the user has permissions for the specified actions and resources.
func (h *Handler) Authorized(nCtx contextx.IContext, req []*types.AuthorizedItem) ([]*types.AuthorizedResult, error) {
	backendReq := new(protoBackend.AuthorizedReq)
	backendReq.ConvertParamFromTypes(req)

	resp, err := h.cli.authorized(nCtx, backendReq)
	if err != nil {
		return nil, err
	}

	if resp.GetData() == nil {
		return nil, nil
	}

	return resp.ConvertResultToTypes(), nil
}
