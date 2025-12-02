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
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
)

// TableName credit table name.
// nolint: perfsprint
func TableName(tenantID string) string {
	return fmt.Sprintf("credit_%s", tenantID)
}

var _ base.IData = &Credit{}

// Credit represents a credit.
type Credit struct {
	TenantID   string    `json:"tenant_id" bson:"tenant_id"`
	CreditID   string    `json:"credit_id" bson:"credit_id"`
	CreditData []byte    `json:"credit_data" bson:"credit_data"`
	ExpireAt   time.Time `json:"expire_at" bson:"expire_at"`
}

// UniqueFields unique fields of the table.
func (c *Credit) UniqueFields() []string {
	return []string{FieldKeyTenantID, FieldKeyCreditID}
}

// UniqueKey unique key of the table.
func (c *Credit) UniqueKey() string {
	return fmt.Sprintf("%s_%s", c.TenantID, c.CreditID)
}

// TableCredit represents the complete db structures of a credit.
type TableCredit base.TableBroker[*Credit]
