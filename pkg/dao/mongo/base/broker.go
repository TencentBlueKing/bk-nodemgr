/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package base provides some base table struct.
package base

import "time"

// BasicInfo represents a basic info for every table.
type BasicInfo struct {
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
	IsDeleted bool      `json:"is_deleted" bson:"is_deleted"`
}

// TableBroker is a broker for every table.
type TableBroker[T IData] struct {
	BasicInfo `json:"basic" bson:"basic"`
	Data      T `json:"data" bson:"data"`
}

// IData this ia the table data.
type IData interface {
	UniqueKey() string
	UniqueFields() []string
}

// NewBasicInfo creates a new BasicInfo.
func NewBasicInfo() BasicInfo {
	return BasicInfo{
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IsDeleted: false,
	}
}

// TableChangeEventBroker table change event broker
type TableChangeEventBroker[T IData] struct {
	FullDocument *TableBroker[T] `json:"fullDocument" bson:"fullDocument"`
}
