/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package networkunit

import "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"

// TableName networkunit table name.
func TableName() string {
	return "networkunit"
}

// NetworkUnit represents network unit table.
// NetworkUnitID should be the unique key.
type NetworkUnit struct {
	TenantID        string `json:"tenant_id" bson:"tenant_id"`
	NetworkUnitID   int64  `json:"networkunit_id" bson:"networkunit_id"`
	NetworkUnitName string `json:"networkunit_name" bson:"networkunit_name"`

	NetworkAreaID int64   `json:"networkarea_id" bson:"networkarea_id"`
	AccessPoints  []int64 `json:"accesspoints" bson:"accesspoints"`
	Links         *Links  `json:"links" bson:"links"`
}

// Link represents link target.
type Link struct {
	NetworkAreaID int64 `json:"networkarea_id" bson:"networkarea_id"`
	NetworkUnitID int64 `json:"networkunit_id" bson:"networkunit_id"`
	AccessPointID int64 `json:"accesspoint_id" bson:"accesspoint_id"`
}

// Links represents links.
type Links struct {
	Cluster *Link `json:"cluster" bson:"cluster"`
	File    *Link `json:"file" bson:"file"`
	Data    *Link `json:"data" bson:"data"`
}

// TableNetworkUnit represent the complete db structures of a networkunit.
type TableNetworkUnit base.TableBroker[*NetworkUnit]
