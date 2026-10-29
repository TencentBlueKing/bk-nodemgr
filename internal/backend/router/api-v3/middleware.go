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

package apiv3

import (
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/gin-gonic/gin"
	gocache "github.com/patrickmn/go-cache"
)

const systemAccessCacheTTL = 5 * time.Minute

// MiddlewareSystemAccess checks resource-free system access after identity verification.
// Reuse the returned middleware across basic route groups to share the positive-only cache.
func MiddlewareSystemAccess(authorizer auth.IAuthorizer) gin.HandlerFunc {
	allowedUsers := gocache.New(systemAccessCacheTTL, systemAccessCacheTTL)

	return func(gCtx *gin.Context) {
		rCtx, err := restserver.GenRestContext(gCtx)
		if err != nil {
			logger.G.Sys().WithErr(err).Error("failed to build system access context")
			restserver.NewRequest(gCtx).AbortWithJSONError(resterrf.Unauthorized, []error{err})

			return
		}

		if err := checkSystemAccess(rCtx, authorizer, allowedUsers); err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to check system access")
			code, unwrapErrs := resterrf.ErrUnwrap(err)
			if code == resterrf.PermissionDenied {
				rCtx.AbortWithJSONPermDenied(code, unwrapErrs)

				return
			}
			rCtx.AbortWithJSONError(code, unwrapErrs)

			return
		}

		gCtx.Next()
	}
}

func checkSystemAccess(rCtx restserver.IContext, authorizer auth.IAuthorizer, allowedUsers *gocache.Cache) error {
	if err := rCtx.CheckTenantID(); err != nil {
		return resterrf.ErrWrap(resterrf.Unauthorized, err)
	}
	if err := rCtx.CheckBKUsername(); err != nil {
		return resterrf.ErrWrap(resterrf.Unauthorized, err)
	}

	// Quoted components prevent collisions between tenant and username boundaries.
	key := fmt.Sprintf("%q:%q", rCtx.TenantID(), rCtx.BKUsername())
	if _, found := allowedUsers.Get(key); found {
		return nil
	}

	// Empty resources intentionally request an action-level check.
	if err := authorizer.Check(rCtx, auth.ActionSystemAccess, nil); err != nil {
		return resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	// Hits do not refresh the TTL; denied results and IAM failures are never cached.
	allowedUsers.Set(key, true, systemAccessCacheTTL)

	return nil
}
