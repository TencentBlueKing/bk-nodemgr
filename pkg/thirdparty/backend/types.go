/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package backend

import (
	"fmt"

	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// CodeOK defines the success code.
const CodeOK = 0

type backendBaseResp interface {
	GetCode() int32
	GetMessage() string
	GetRequestId() string
}

type backendPermissionError struct {
	message string
}

func (err *backendPermissionError) Error() string {
	if err.message != "" {
		return err.message
	}

	return "permission denied"
}

func (err *backendPermissionError) PermissionData() resterrf.Permission {
	return resterrf.Permission{}
}

func buildBackendResponseError(apiName string, resp backendBaseResp, errInfo any) error {
	baseErr := fmt.Errorf("%s failed. code(%d), message(%s), error(%v), request-id(%s)",
		apiName, resp.GetCode(), resp.GetMessage(), errInfo, resp.GetRequestId())

	if resterrf.Code(resp.GetCode()) != resterrf.PermissionDenied {
		return baseErr
	}

	permErr := &backendPermissionError{message: resp.GetMessage()}

	return fmt.Errorf("%w: %w", baseErr, permErr)
}

func convertPage(page types.Page) *protoBackend.Page {
	return &protoBackend.Page{
		Offset: int32(page.Offset),
		Limit:  int32(page.Limit),
	}
}
