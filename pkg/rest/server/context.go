/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package server defines rest server context.
package server

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	"github.com/gin-gonic/gin"
)

// IContext rest server context interface, contains the global IContext and rest IRequest.
type IContext interface {
	// the global IContext.
	contextx.IContext

	// the rest IRequest.
	IRequest
}

// restRequestKey was used to store the restRequest in gin.Context.
const restRequestKey = "rest_request"

// initRestRequest initializes a new rest request.
// notice: please don't set Context's fields here, use Middleware to set fields.
func initRestRequest(gCtx *gin.Context) IRequest {
	r := NewRequest(gCtx)
	gCtx.Set(restRequestKey, r)

	return r
}

// loadRestRequest loads rest request from gin.Context.
func loadRestRequest(gCtx *gin.Context) IRequest {
	r, ok := gCtx.Value(restRequestKey).(IRequest)
	if !ok {
		return initRestRequest(gCtx)
	}

	return r
}

// GenRestContext only when user has authenticated, otherwise, return ErrorUnauthorized.
func GenRestContext(c *gin.Context) (*Context, error) {
	rObj, ok := c.Get(restRequestKey)
	if !ok {
		return nil, resterrf.CodeErrMap(resterrf.Unauthorized)
	}

	r, ok := rObj.(IRequest)
	if !ok {
		return nil, resterrf.CodeErrMap(resterrf.Unauthorized)
	}

	return &Context{
		Context: contextx.New(
			r.GContext(),
			contextx.WithTenantID(r.Data().GetTenantID()),
			contextx.WithBKUsername(r.Data().GetBKUsername()),
			contextx.WithMessageID(r.Data().GetRequestID()),
		),
		IRequest: r,
	}, nil
}

// Context rest context.
type Context struct {
	// implements the global IContext
	*contextx.Context

	// implements the rest IRequest.
	IRequest
}
