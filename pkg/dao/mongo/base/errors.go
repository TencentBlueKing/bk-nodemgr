/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package base

import (
	"errors"

	"go.mongodb.org/mongo-driver/mongo"
)

// ErrInvalidContext return the error when context is invalid.
func ErrInvalidContext() error {
	return errInvalidContext
}

// ErrEnsureIndexesFailed return the error when ensure indexes failed.
func ErrEnsureIndexesFailed() error {
	return errEnsureIndexesFailed
}

// ErrInvalidID return the error when specified id is invalid.
// Deprecated: please use every pkg's their own error.
func ErrInvalidID() error {
	return errInvalidID
}

// ErrEmptyParamData return the error when param data is empty.
func ErrEmptyParamData() error {
	return errEmptyParamData
}

// ErrInvalidItemInParamList return the error when param item in param list is invalid.
func ErrInvalidItemInParamList() error {
	return errInvalidItemInParamList
}

// ErrTenantIDNotMatched return the error when tenant id is not matched.
func ErrTenantIDNotMatched() error {
	return errTenantIDNotMatched
}

// ErrRecordNoFound return the error when record no found.
func ErrRecordNoFound() error {
	return mongo.ErrNoDocuments
}

var (
	errInvalidParam           = errors.New("param is invalid")
	errEnsureIndexesFailed    = errors.New("ensure indexes failed")
	errInvalidID              = errors.New("invalid id")
	errEmptyParamData         = errors.New("empty param data")
	errInvalidItemInParamList = errors.New("invalid item in param list")
	errTenantIDNotMatched     = errors.New("tenant id is not matched")
	errInvalidContext         = errors.New("context is nil")
)

// ErrInvalidParam return the error when param is invalid.
func ErrInvalidParam(err error) error {
	return errors.Join(errInvalidParam, err)
}
