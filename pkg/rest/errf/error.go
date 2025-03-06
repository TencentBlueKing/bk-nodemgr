/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package errf ...
package errf

import (
	"errors"
)

var instance = struct {
	codeErrMap map[Code]error
	errCodeMap map[error]Code
}{}

func init() {
	errorMaps := map[Code]error{
		// retain status code.
		Unknown:          errors.New("unauthorized"),
		PermissionDenied: errors.New("permission denied"),
		MaxErrCode:       errors.New("max err code"),

		// custom status code.
		InvalidParameter:        errors.New("invalid parameter"),
		TooManyRequest:          errors.New("too many request"),
		RecordNotFound:          errors.New("record not found"),
		DecodeRequestFailed:     errors.New("decode request failed"),
		UnHealthy:               errors.New("unhealthy"),
		Aborted:                 errors.New("aborted"),
		Unauthorized:            errors.New("unknown error"),
		PartialFailed:           errors.New("partial failed"),
		DBExecCmdFailed:         errors.New("db exec cmd failed"),
		InvalidCache:            errors.New("invalid cache"),
		InvalidFileResource:     errors.New("invalid file resource"),
		ThirdpartyRequestFailed: errors.New("thirdparty request failed"),
	}

	instance.codeErrMap = make(map[Code]error)
	instance.errCodeMap = make(map[error]Code)

	for code, err := range errorMaps {
		instance.codeErrMap[code] = err
		instance.errCodeMap[err] = code
	}
}

// CodeErrMap ...
func CodeErrMap(code Code) error {
	return instance.codeErrMap[code]
}

// ErrCodeMap ...
func ErrCodeMap(err error) Code {
	code, ok := instance.errCodeMap[err]
	if !ok {
		return Unknown
	}

	return code
}

// ErrWrap this function is used to wrap a base error with a code.
// notice: this is single layer wrap, don't use it too much.
func ErrWrap(code Code, err error) error {
	baseErr := CodeErrMap(code)
	if baseErr == nil {
		return err
	}

	return errors.Join(baseErr, err)
}

// ErrUnwrap this function is used to unwrap a error to get the base error.
// notice: this is single layer unwrap, don't use it too much.
func ErrUnwrap(err error) (Code, []error) {
	if err == nil {
		return OK, nil
	}

	u, ok := err.(interface {
		Unwrap() []error
	})
	if !ok {
		return ErrCodeMap(err), nil
	}

	errs := u.Unwrap()
	baseErr, unwrappedErr := errs[0], errs[1:]

	return ErrCodeMap(baseErr), unwrappedErr
}
