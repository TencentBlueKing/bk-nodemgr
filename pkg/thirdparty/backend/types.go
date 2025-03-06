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

	proto "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// CodeOK defines the success code.
const CodeOK = 0

func convertPage(page types.Page) *proto.Page {
	return &proto.Page{
		Offset: int32(page.Offset),
		Limit:  int32(page.Limit),
	}
}

var (
	errConditionTypeNotSupport = errors.New("condition type not support")
)

// ErrConditionTypeNotSupport err
func ErrConditionTypeNotSupport() error {
	return errConditionTypeNotSupport
}
