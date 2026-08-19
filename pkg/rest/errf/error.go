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

// Package errf ...
package errf

import (
	"errors"
	"sync"
)

type errMapping struct {
	once sync.Once

	codeErrMap map[Code]error
	errCodeMap map[error]Code
}

// nolint: gochecknoglobals
var errMaps errMapping

// nolint: varnamelen
func (m *errMapping) ensureInit() {
	m.once.Do(func() {
		errorMaps := map[Code]error{
			// retain status code.
			Unknown:          errors.New("unknown error"),
			PermissionDenied: errors.New("permission denied"),
			MaxErrCode:       errors.New("max err code"),

			// custom status code.
			InvalidParameter:        errors.New("invalid parameter"),
			TooManyRequest:          errors.New("too many request"),
			RecordNotFound:          errors.New("record not found"),
			DecodeRequestFailed:     errors.New("decode request failed"),
			UnHealthy:               errors.New("unhealthy"),
			Aborted:                 errors.New("aborted"),
			Unauthorized:            errors.New("unauthorized"),
			PartialFailed:           errors.New("partial failed"),
			DBExecCmdFailed:         errors.New("db exec cmd failed"),
			InvalidCache:            errors.New("invalid cache"),
			InvalidFileResource:     errors.New("invalid file resource"),
			ThirdpartyRequestFailed: errors.New("thirdparty request failed"),
			BackendOperateFailed:    errors.New("backend operate failed"),
			ResourceScanTooLarge:    errors.New("resource scan too large"),
			InvalidKeyword:          errors.New("invalid keyword"),
		}

		m.codeErrMap = make(map[Code]error, len(errorMaps))
		m.errCodeMap = make(map[error]Code, len(errorMaps))

		for code, err := range errorMaps {
			m.codeErrMap[code] = err
			m.errCodeMap[err] = code
		}
	})
}

// CodeErrMap ...
func CodeErrMap(code Code) error {
	errMaps.ensureInit()

	return errMaps.codeErrMap[code]
}

// ErrCodeMap ...
func ErrCodeMap(err error) Code {
	errMaps.ensureInit()
	code, ok := errMaps.errCodeMap[err]
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

	unwrapper, ok := err.(interface {
		Unwrap() []error
	})
	if !ok {
		return ErrCodeMap(err), nil
	}

	errs := unwrapper.Unwrap()
	baseErr, unwrappedErr := errs[0], errs[1:]

	code := ErrCodeMap(baseErr)
	if code != OK && code != PermissionDenied {
		for _, unwrapped := range unwrappedErr {
			var permErr PermissionError
			if errors.As(unwrapped, &permErr) {
				code = PermissionDenied
				break
			}
		}
	}

	return code, unwrappedErr
}
