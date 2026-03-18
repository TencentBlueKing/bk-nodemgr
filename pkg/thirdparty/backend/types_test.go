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
	"errors"
	"testing"

	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
)

type fakeBackendResp struct {
	code      int32
	message   string
	requestID string
}

func (resp fakeBackendResp) GetCode() int32 {
	return resp.code
}

func (resp fakeBackendResp) GetMessage() string {
	return resp.message
}

func (resp fakeBackendResp) GetRequestId() string {
	return resp.requestID
}

func TestBuildBackendResponseError_KeepNonPermissionCode(t *testing.T) {
	err := buildBackendResponseError("list business", fakeBackendResp{
		code:      int32(resterrf.ThirdpartyRequestFailed),
		message:   "thirdparty failed",
		requestID: "rid-non-perm",
	}, "backend error detail")

	if err == nil {
		t.Fatal("expected non-nil error")
	}

	var permErr resterrf.PermissionError
	if errors.As(err, &permErr) {
		t.Fatalf("expected no permission error, got %T", permErr)
	}

	code, unwrapErrs := resterrf.ErrUnwrap(resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err))
	if code != resterrf.ThirdpartyRequestFailed {
		t.Fatalf("expected code %d, got %d", resterrf.ThirdpartyRequestFailed, code)
	}
	if len(unwrapErrs) != 1 {
		t.Fatalf("expected 1 unwrap error, got %d", len(unwrapErrs))
	}
}

func TestBuildBackendResponseError_ExposePermissionError(t *testing.T) {
	err := buildBackendResponseError("install node agent", fakeBackendResp{
		code:      int32(resterrf.PermissionDenied),
		message:   "permission denied",
		requestID: "rid-perm",
	}, "permission detail")

	if err == nil {
		t.Fatal("expected non-nil error")
	}

	var permErr resterrf.PermissionError
	if !errors.As(err, &permErr) {
		t.Fatal("expected wrapped permission error")
	}

	code, unwrapErrs := resterrf.ErrUnwrap(resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err))
	if code != resterrf.PermissionDenied {
		t.Fatalf("expected upgraded code %d, got %d", resterrf.PermissionDenied, code)
	}
	if len(unwrapErrs) != 1 {
		t.Fatalf("expected 1 unwrap error, got %d", len(unwrapErrs))
	}
}
