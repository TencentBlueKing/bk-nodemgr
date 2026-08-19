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

package errf

import (
	"errors"
	"fmt"
	"testing"
)

type fakePermissionError struct{}

func (fakePermissionError) Error() string {
	return "permission denied"
}

func (fakePermissionError) PermissionData() Permission {
	return Permission{
		System:     "bk_nodeman",
		SystemName: "NodeMan",
		ApplyURL:   "https://iam.example.com/apply",
	}
}

func TestErrUnwrap_ShortCircuitWhenCodeIsPermissionDenied(t *testing.T) {
	originalErr := errors.New("already denied")

	code, unwrapErrs := ErrUnwrap(ErrWrap(PermissionDenied, originalErr))
	if code != PermissionDenied {
		t.Fatalf("expected code %d, got %d", PermissionDenied, code)
	}
	if len(unwrapErrs) != 1 {
		t.Fatalf("expected 1 unwrap error, got %d", len(unwrapErrs))
	}
}

func TestErrUnwrap_UpgradeToPermissionDeniedWhenPermissionErrorExists(t *testing.T) {
	permErr := fmt.Errorf("wrapped permission: %w", fakePermissionError{})

	code, unwrapErrs := ErrUnwrap(ErrWrap(ThirdpartyRequestFailed, permErr))
	if code != PermissionDenied {
		t.Fatalf("expected code %d, got %d", PermissionDenied, code)
	}
	if len(unwrapErrs) != 1 {
		t.Fatalf("expected 1 unwrap error, got %d", len(unwrapErrs))
	}
}

func TestErrUnwrap_KeepOriginalCodeWithoutPermissionError(t *testing.T) {
	code, unwrapErrs := ErrUnwrap(ErrWrap(ThirdpartyRequestFailed, errors.New("plain thirdparty error")))
	if code != ThirdpartyRequestFailed {
		t.Fatalf("expected code %d, got %d", ThirdpartyRequestFailed, code)
	}
	if len(unwrapErrs) != 1 {
		t.Fatalf("expected 1 unwrap error, got %d", len(unwrapErrs))
	}
}

func TestErrUnwrap_OKCodeDoesNotProbePermissionError(t *testing.T) {
	code, unwrapErrs := ErrUnwrap(nil)
	if code != OK {
		t.Fatalf("expected code %d, got %d", OK, code)
	}
	if unwrapErrs != nil {
		t.Fatalf("expected nil unwrap errors, got %v", unwrapErrs)
	}
}
