/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package errf

import "net/http"

var (
	codeHttpStatus = map[Code]int{
		// retain status code.
		OK:               http.StatusOK,
		PermissionDenied: http.StatusForbidden,
		MaxErrCode:       http.StatusInternalServerError,

		// custom status code.
		InvalidParameter:        http.StatusBadRequest,
		TooManyRequest:          http.StatusTooManyRequests,
		RecordNotFound:          http.StatusNotFound,
		DecodeRequestFailed:     http.StatusBadRequest,
		UnHealthy:               http.StatusServiceUnavailable,
		Aborted:                 http.StatusInternalServerError,
		Unauthorized:            http.StatusUnauthorized,
		PartialFailed:           http.StatusInternalServerError,
		DBExecCmdFailed:         http.StatusInternalServerError,
		InvalidCache:            http.StatusInternalServerError,
		InvalidFileResource:     http.StatusInternalServerError,
		ThirdpartyRequestFailed: http.StatusInternalServerError,
		BackendOperateFailed:    http.StatusInternalServerError,
		ResourceScanTooLarge:    http.StatusUnprocessableEntity, // 422
		InvalidKeyword:          http.StatusNotAcceptable,       // 406
	}
)

// HttpStatusCode returns the http status code.
func (code Code) HttpStatusCode() int {
	httpStatus, ok := codeHttpStatus[code]
	if !ok {
		return http.StatusInternalServerError
	}

	return httpStatus
}
