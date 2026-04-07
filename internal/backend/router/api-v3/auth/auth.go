/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package auth provides proactive permission verification endpoints.
package auth

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/options"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
)

type handler struct {
	rg         *gin.RouterGroup
	authorizer auth.IAuthorizer
}

func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{
		rg:         rg.Group("/auth"),
		authorizer: capability.Authorizer,
	}
}

// Load registers the auth verification routes.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	h.rg.POST("/verify", restserver.Handler(h.Verify))
}

// Verify checks whether the current user has permissions for the requested resources.
func (h *handler) Verify(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.AuthVerifyReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to verify auth, failed to decode request body")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if err := req.Validate(); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to verify auth, invalid request")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	results, err := h.verifyItems(rCtx, req.GetItems())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to verify auth")

		var permErr auth.PermissionDeniedError
		if errors.As(err, &permErr) {
			return nil, resterrf.ErrWrap(resterrf.PermissionDenied, err)
		}

		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	resp := new(protoBackend.AuthVerifyResp)
	resp.ConvertResultsFromVerify(results)

	return resp.GetData(), nil
}

func (h *handler) verifyItems(
	rCtx restserver.IContext, items []*protoBackend.AuthVerifyItem,
) ([]*protoBackend.AuthVerifyResult, error) {

	results := make([]*protoBackend.AuthVerifyResult, 0, len(items))

	for _, item := range items {
		result, err := h.verifyItem(rCtx, item)
		if err != nil {
			return nil, err
		}

		results = append(results, result)
	}

	return results, nil
}

func (h *handler) verifyItem(
	rCtx restserver.IContext, item *protoBackend.AuthVerifyItem,
) (*protoBackend.AuthVerifyResult, error) {

	action := auth.Action(item.GetAction())
	resources := protoBackend.ConvertAuthResourcesToInternal(item.GetResources())

	checkErr := h.authorizer.Check(rCtx, action, resources)
	if checkErr == nil {
		return protoBackend.NewAuthVerifyResult(item.GetAction(), true), nil
	}

	var permErr auth.PermissionDeniedError
	if errors.As(checkErr, &permErr) {
		return nil, checkErr
	}

	return nil, fmt.Errorf("check action %s permission: %w", item.GetAction(), checkErr)
}
