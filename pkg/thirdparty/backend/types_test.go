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

	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
)

type fakeBackendResp struct {
	code      int32
	message   string
	requestID string
}

type fakeBackendRespWithPermission struct {
	fakeBackendResp
	permission *protoBackend.Permission
}

type fakePermissionError struct {
	permission resterrf.Permission
}

func (err fakePermissionError) Error() string {
	return "permission denied"
}

func (err fakePermissionError) PermissionData() resterrf.Permission {
	return err.permission
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

func (resp fakeBackendRespWithPermission) GetPermission() *protoBackend.Permission {
	return resp.permission
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

func TestBuildBackendResponseError_ReusePermissionDataFromErrorInfo(t *testing.T) {
	err := buildBackendResponseError("install node agent", fakeBackendResp{
		code:      int32(resterrf.PermissionDenied),
		message:   "permission denied",
		requestID: "rid-perm-data",
	}, fakePermissionError{permission: resterrf.Permission{
		System:     "bk_iam",
		SystemName: "IAM",
		ApplyURL:   "https://iam.example/apply",
	}})

	if err == nil {
		t.Fatal("expected non-nil error")
	}

	var permErr resterrf.PermissionError
	if !errors.As(err, &permErr) {
		t.Fatal("expected wrapped permission error")
	}

	permData := permErr.PermissionData()
	if permData.System != "bk_iam" {
		t.Fatalf("expected system bk_iam, got %s", permData.System)
	}
	if permData.SystemName != "IAM" {
		t.Fatalf("expected system name IAM, got %s", permData.SystemName)
	}
	if permData.ApplyURL != "https://iam.example/apply" {
		t.Fatalf("expected apply url https://iam.example/apply, got %s", permData.ApplyURL)
	}
}

func TestBuildBackendResponseError_RecognizePermissionFromErrorInfo(t *testing.T) {
	err := buildBackendResponseError("list business", fakeBackendResp{
		code:      int32(resterrf.ThirdpartyRequestFailed),
		message:   "thirdparty failed",
		requestID: "rid-perm-recognize",
	}, fakePermissionError{permission: resterrf.Permission{
		System: "bk_iam",
	}})

	if err == nil {
		t.Fatal("expected non-nil error")
	}

	var permErr resterrf.PermissionError
	if !errors.As(err, &permErr) {
		t.Fatal("expected wrapped permission error from error info")
	}

	code, unwrapErrs := resterrf.ErrUnwrap(resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err))
	if code != resterrf.PermissionDenied {
		t.Fatalf("expected upgraded code %d, got %d", resterrf.PermissionDenied, code)
	}
	if len(unwrapErrs) != 1 {
		t.Fatalf("expected 1 unwrap error, got %d", len(unwrapErrs))
	}
}

func TestBuildBackendResponseError_RecognizePermissionFromResponse(t *testing.T) {
	err := buildBackendResponseError("select inner ip", fakeBackendRespWithPermission{
		fakeBackendResp: fakeBackendResp{
			code:      int32(resterrf.ThirdpartyRequestFailed),
			message:   "thirdparty failed",
			requestID: "rid-perm-from-resp",
		},
		permission: &protoBackend.Permission{
			System:     "bk_iam",
			SystemName: "IAM",
			ApplyUrl:   "https://iam.example/apply",
			Actions: []*protoBackend.Action{
				{
					Id:   "host_manage",
					Name: "Host Manage",
					RelatedResourceTypes: []*protoBackend.RelatedResourceType{
						{
							SystemId: "bk_cmdb",
							Type:     "biz",
							TypeName: "Business",
						},
					},
				},
			},
		},
	}, nil)

	if err == nil {
		t.Fatal("expected non-nil error")
	}

	var permErr resterrf.PermissionError
	if !errors.As(err, &permErr) {
		t.Fatal("expected wrapped permission error from response")
	}

	permissionData := permErr.PermissionData()
	if permissionData.System != "bk_iam" {
		t.Fatalf("expected system bk_iam, got %s", permissionData.System)
	}
	if permissionData.SystemName != "IAM" {
		t.Fatalf("expected system_name IAM, got %s", permissionData.SystemName)
	}
	if permissionData.ApplyURL != "https://iam.example/apply" {
		t.Fatalf("expected apply_url https://iam.example/apply, got %s", permissionData.ApplyURL)
	}
	if len(permissionData.Actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(permissionData.Actions))
	}
	if len(permissionData.Actions[0].RelatedResourceTypes) != 1 {
		t.Fatalf("expected 1 related_resource_type, got %d", len(permissionData.Actions[0].RelatedResourceTypes))
	}
	if permissionData.Actions[0].RelatedResourceTypes[0].SystemID != "bk_cmdb" {
		t.Fatalf("expected related system_id bk_cmdb, got %s", permissionData.Actions[0].RelatedResourceTypes[0].SystemID)
	}

	code, _ := resterrf.ErrUnwrap(resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err))
	if code != resterrf.PermissionDenied {
		t.Fatalf("expected upgraded code %d, got %d", resterrf.PermissionDenied, code)
	}
}
