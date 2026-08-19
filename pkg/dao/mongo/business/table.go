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

// Package business ...
package business

import (
	"fmt"
	"strconv"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
)

// TableName business table name.
func TableName(tenantID string) string {
	return fmt.Sprintf("business_%s", tenantID)
}

var _ base.IData = &Business{}

// Business represents a business under a tenant.
// BizID should be the unique key.
type Business struct {
	TenantID string `json:"tenant_id" bson:"tenant_id"`
	BizID    int64  `json:"biz_id" bson:"biz_id"`
	BizName  string `json:"biz_name" bson:"biz_name"`
}

// UniqueFields unique fields of the table.
func (biz *Business) UniqueFields() []string {
	return []string{FieldKeyBizID}
}

// UniqueKey unique key of the table.
func (biz *Business) UniqueKey() string {
	return strconv.FormatInt(biz.BizID, 10)
}

// TableBusiness represents the complete db structures of a business.
type TableBusiness base.TableBroker[*Business]
