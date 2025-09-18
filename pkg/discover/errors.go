/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package discover ...
package discover

import "errors"

// ErrServiceNotFound this defines the error of service not found.
func ErrServiceNotFound() error {
	return errors.New("service not found")
}

// ErrEndpointNotFound this defines the error of endpoint not found.
func ErrEndpointNotFound() error {
	return errors.New("endpoint not found")
}

// ErrInvalidInstance this defines the error of invalid instance.
func ErrInvalidInstance() error {
	return errors.New("invalid instance")
}

// ErrNotRegistered this defines the error of not registered.
func ErrNotRegistered() error {
	return errors.New("not registered")
}

// ErrInvalidInstanceID this defines the error of invalid instance ID.
func ErrInvalidInstanceID() error {
	return errors.New("invalid instance ID")
}

// ErrInvalidInstanceName this defines the error of invalid instance name.
func ErrInvalidInstanceName() error {
	return errors.New("invalid instance name")
}

// ErrInvalidServiceName this defines the error of invalid service name.
func ErrInvalidServiceName() error {
	return errors.New("invalid service name")
}

// ErrInvalidServiceIP this defines the error of invalid service IP.
func ErrInvalidServiceIP() error {
	return errors.New("invalid service IP")
}

// ErrInvalidServicePort this defines the error of invalid service port.
func ErrInvalidServicePort() error {
	return errors.New("invalid service port")
}

// ErrInvalidSelector this defines the error of invalid selector.
func ErrInvalidSelector() error {
	return errors.New("invalid selector")
}

// ErrDiscoverNotStarted this defines the error of discover not started.
func ErrDiscoverNotStarted() error {
	return errors.New("discover not started")
}

// ErrDiscoverInternalError this defines the error of discover internal error.
func ErrDiscoverInternalError() error {
	return errors.New("discover internal error")
}

// ErrMetaValueNotFound this defines the error of meta value not found.
func ErrMetaValueNotFound() error {
	return errors.New("meta value not found")
}
