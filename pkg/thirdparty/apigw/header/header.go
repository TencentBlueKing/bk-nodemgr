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

// Package header define the blueking common header.
package header

import (
	"net/http"
)

const (
	// BKGWRIDKey is request id header key.
	BKGWRIDKey = "X-Bkapi-Request-Id"

	// BKGWJWTTokenKey is blueking api gateway jwt header key.
	BKGWJWTTokenKey = "X-Bkapi-JWT"

	// BKGWAuthKey is blueking api gateway authorization header key.
	BKGWAuthKey = "X-Bkapi-Authorization"

	// BKGWTenantIDKey is blueking api gateway tenant id header key.
	BKGWTenantIDKey = "X-Bk-Tenant-Id"
)

// BKRIDGetter request id value.
func BKRIDGetter(req *http.Request) string {
	id := req.Header.Get(BKGWRIDKey)

	return id
}
