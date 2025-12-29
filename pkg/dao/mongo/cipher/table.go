/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package cipher provides the dao layer for cipher table.
package cipher

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
)

// TableName accesspoint table name.
func TableName(tenantID string) string {
	return fmt.Sprintf("cipher_%s", tenantID)
}

var _ base.IData = &Cipher{}

// Cipher represents an access point for cipher storage.
type Cipher struct {
	Name        string `json:"name" bson:"name"`
	KeyType     string `json:"key_type" bson:"key_type"`
	Description string `json:"description" bson:"description"`
	PrivateKey  []byte `json:"private_key" bson:"private_key"`
	PublicKey   []byte `json:"public_key" bson:"public_key"`
}

// UniqueFields unique fields of the table.
func (a *Cipher) UniqueFields() []string {
	return []string{FieldKeyName}
}

// UniqueKey unique key of the table.
func (a *Cipher) UniqueKey() string {
	return a.Name
}

// TableAccessPoint represent the complete db structures of a access point.
type TableAccessPoint base.TableBroker[*Cipher]
