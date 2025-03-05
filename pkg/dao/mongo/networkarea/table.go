/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package networkarea

import (
	"strconv"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
)

// TableName networkarea table name.
func TableName() string {
	return "networkarea"
}

// NetworkArea represent network area table.
// NetworkAreaID should be the unique key.
// NetworkAreaID is global unique among all tenants.
type NetworkArea struct {
	TenantID        string `json:"tenant_id" bson:"tenant_id"`
	NetworkAreaID   int64  `json:"networkarea_id" bson:"networkarea_id"`
	NetworkAreaName string `json:"networkarea_name" bson:"networkarea_name"`
	CloudVendor     string `json:"cloud_vendor" bson:"cloud_vendor"`
}

// UniqueKey unique key of the table.
func (area *NetworkArea) UniqueKey() string {
	return strconv.FormatInt(area.NetworkAreaID, 10)
}

// TableNetworkArea represent the complete db structures of a networkarea.
type TableNetworkArea base.TableBroker[*NetworkArea]
