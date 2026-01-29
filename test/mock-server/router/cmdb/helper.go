/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package cmdb provides the CMDB mock API helper.
package cmdb

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/cmdb"
	"github.com/TencentBlueKing/bk-nodemgr/test/mock-server/router/common"
	"github.com/gin-gonic/gin"
)

// respondSuccess sends a success response with CMDB format.
func respondSuccess[T any](gCtx *gin.Context, data T) {
	common.RespondJSON(gCtx, cmdb.BaseBroker[T]{
		RespCommon: cmdb.RespCommon{
			Result:  true,
			Code:    CodeOK,
			Message: "",
		},
		Data: data,
	})
}

// respondError sends an error response with CMDB format.
func respondError(gCtx *gin.Context, code int, message string) {
	common.RespondJSON(gCtx, cmdb.BaseBroker[any]{
		RespCommon: cmdb.RespCommon{
			Result:  false,
			Code:    code,
			Message: message,
		},
		Data: nil,
	})
}
