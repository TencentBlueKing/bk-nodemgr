/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package accesspoint

import (
	"fmt"
	"strconv"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
)

// TableName accesspoint table name.
func TableName(tenantID string) string {
	return fmt.Sprintf("%s_%s", tableNamePrefix, tenantID)
}

const tableNamePrefix = "accesspoint"

var _ base.IData = &AccessPoint{}

// AccessPoint represents access point table.
// AccessPointID should be the unique key.
// AccessPointID is generated from a global counter and remains unique across tenant collections.
type AccessPoint struct {
	TenantID        string `json:"tenant_id" bson:"tenant_id"`
	AccessPointID   int64  `json:"accesspoint_id" bson:"accesspoint_id"`
	AccessPointName string `json:"accesspoint_name" bson:"accesspoint_name"`

	NetworkAreaID int64      `json:"networkarea_id" bson:"networkarea_id"`
	Endpoints     *Endpoints `json:"endpoints" bson:"endpoints"`
}

// Endpoints represents endpoints of access point.
type Endpoints struct {
	Cluster []string `json:"cluster" bson:"cluster"`
	File    []string `json:"file" bson:"file"`
	Data    []string `json:"data" bson:"data"`
}

// UniqueFields unique fields of the table.
func (a *AccessPoint) UniqueFields() []string {
	return []string{FieldKeyAccessPointID}
}

// UniqueKey unique key of the table.
func (a *AccessPoint) UniqueKey() string {
	return strconv.FormatInt(a.AccessPointID, 10)
}

// TableAccessPoint represent the complete db structures of a access point.
type TableAccessPoint base.TableBroker[*AccessPoint]
