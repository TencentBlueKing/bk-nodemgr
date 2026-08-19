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

// Package common defines workflow common things.
package common

import "errors"

var (
	errActionSkipped               = errors.New("action skipped")
	errActionAlreadySucceeded      = errors.New("action already succeeded")
	errInvalidParameters           = errors.New("invalid parameters")
	errInvalidState                = errors.New("invalid state")
	errNoActionTodo                = errors.New("no action to do")
	errUnknownTriggerCategory      = errors.New("unknown trigger category")
	errInvalidTriggerMetadata      = errors.New("invalid trigger metadata")
	errInvalidPeriodicOperationNum = errors.New("invalid periodic operation num")
	errTriggerInactive             = errors.New("trigger is inactive")
	errPeriodicTriggerNotReady     = errors.New("periodic trigger not ready")
)

// ErrActionSkipped gets action skipped error.
func ErrActionSkipped() error {
	return errActionSkipped
}

// ErrActionAlreadySucceeded gets action already succeeded error.
func ErrActionAlreadySucceeded() error {
	return errActionAlreadySucceeded
}

// ErrInvalidParameters gets invalid parameters error.
func ErrInvalidParameters() error {
	return errInvalidParameters
}

// ErrInvalidState gets invalid state error.
func ErrInvalidState() error {
	return errInvalidState
}

// ErrNoActionTodo gets no action todo error.
func ErrNoActionTodo() error {
	return errNoActionTodo
}

// ErrUnknownTriggerCategory gets unknown trigger category error.
func ErrUnknownTriggerCategory() error {
	return errUnknownTriggerCategory
}

// ErrInvalidTriggerMetadata gets invalid trigger metadata error.
func ErrInvalidTriggerMetadata() error {
	return errInvalidTriggerMetadata
}

// ErrInvalidPeriodicOperationNum gets invalid periodic operation num error.
func ErrInvalidPeriodicOperationNum() error {
	return errInvalidPeriodicOperationNum
}

// ErrTriggerInactive gets trigger inactive error.
func ErrTriggerInactive() error {
	return errTriggerInactive
}

// ErrPeriodicTriggerNotReady gets periodic trigger not ready error.
func ErrPeriodicTriggerNotReady() error {
	return errPeriodicTriggerNotReady
}
