/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package header define the blueking common header.
package header

import (
	"net/http"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/identifier"
)

const (
	// BKRIDKey is request id header key.
	BKRIDKey = "X-Bkapi-Request-Id"

	// BKUserKey is operator name header key.
	BKUserKey = "X-Bkapi-User-Name"

	// BKAppCodeKey is blueking application code header key.
	BKAppCodeKey = "X-Bkapi-App-Code"

	// BKLanguageKey the language key word.
	BKLanguageKey = "HTTP_BLUEKING_LANGUAGE"

	// BKGWJWTTokenKey is blueking api gateway jwt header key.
	BKGWJWTTokenKey = "X-Bkapi-JWT"

	// BKTenantIDKey is tenant id header key.
	BKTenantIDKey = "X-Bk-Tenant-Id"

	// BKGWAuthKey is blueking api gateway authorization header key.
	BKGWAuthKey = "X-Bkapi-Authorization"
)

// ContentType http request content type
type ContentType string

// ContentType http request content type
const (
	FormDataContent ContentType = "application/x-www-form-urlencoded"
	JsonContent     ContentType = "application/json"
)

// BKRIDGetter request id value
func BKRIDGetter(req *http.Request, autoGen bool) string {
	id := req.Header.Get(BKRIDKey)
	if id == "" && autoGen {
		id = BKRIDGenerator()
	}
	return id
}

// BKRIDGenerator generate request id
func BKRIDGenerator() string {
	return identifier.GenRequestID()
}
