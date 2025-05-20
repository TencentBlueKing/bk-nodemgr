/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package base ...
package base

import "errors"

// ErrUpsertNilData this error indicates that the user inserted an empty block of data when inserting data.
func ErrUpsertNilData() error {
	return errors.New("upsert nil data")
}

// ErrNilContent this error indicates that the user inserted an empty block of data when inserting data.
func ErrNilContent() error {
	return errors.New("ctx is nil")
}

// ErrEmptyUniqueKey this error indicates that the user inserted an empty block of data when inserting data.
func ErrEmptyUniqueKey() error {
	return errors.New("empty unique key")
}

// ErrEmptyTriggerID this error indicates that the user inserted an empty block of data when inserting data.
func ErrEmptyTriggerID() error {
	return errors.New("empty trigger key")
}

// ErrEmptyActionName this error indicates that the user inserted an empty block of data when inserting data.
func ErrEmptyActionName() error {
	return errors.New("empty action name")
}

// ErrEmptyOperaInstID this error indicates that the user inserted an empty block of data when inserting data.
func ErrEmptyOperaInstID() error {
	return errors.New("empty operation inst id name")
}
