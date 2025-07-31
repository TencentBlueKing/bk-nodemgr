/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package credit

import (
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
)

// ErrInvalidCreditID return the error when creditID is invalid.
func ErrInvalidCreditID() error {
	return base.ErrInvalidParam(errors.New("credit id is invalid"))
}

// ErrInvalidExpireAt return the error when expireAt is invalid.
func ErrInvalidExpireAt() error {
	return base.ErrInvalidParam(errors.New("expire at must be greater than current time"))
}

// ErrEmptyCreditData return the error when credit data is empty.
func ErrEmptyCreditData() error {
	return base.ErrInvalidParam(errors.New("credit data is empty"))
}
