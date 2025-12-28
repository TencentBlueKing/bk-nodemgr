/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package asymmetricencryption provides the dao layer for asymmetric encryption table.
package asymmetricencryption

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
)

// TableName accesspoint table name.
func TableName() string {
	return "asymmetricencryption"
}

var _ base.IData = &AsymmetricEncryption{}

// AsymmetricEncryption represents an access point for asymmetric encryption.
type AsymmetricEncryption struct {
	CipherType  string `json:"cipher_type" bson:"cipher_type"`
	KeyType     string `json:"key_type" bson:"key_type"`
	Description string `json:"description" bson:"description"`
	Content     []byte `json:"content" bson:"content"`
}

// UniqueFields unique fields of the table.
func (a *AsymmetricEncryption) UniqueFields() []string {
	return []string{FieldKeyCipherType, FieldKeyKeyType}
}

// UniqueKey unique key of the table.
func (a *AsymmetricEncryption) UniqueKey() string {
	return fmt.Sprintf("%s_%s", a.CipherType, a.KeyType)
}

// TableAccessPoint represent the complete db structures of a access point.
type TableAccessPoint base.TableBroker[*AsymmetricEncryption]
