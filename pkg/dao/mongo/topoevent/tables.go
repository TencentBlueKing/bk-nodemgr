/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package topoevent

import (
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
)

// TableName topoevent table name.
func TableName(tenantID string) string {
	return fmt.Sprintf("topoevent_%s", tenantID)
}

// TopoEvent represents topo event table.
type TopoEvent struct {
	TenantID        string    `json:"tenant_id" bson:"tenant_id"`
	Type            string    `json:"type" bson:"type"`
	NetworkAreaID   int64     `json:"networkarea_id" bson:"networkarea_id"`
	NetworkAreaName string    `json:"networkarea_name" bson:"networkarea_name"`
	NetworkUnitID   int64     `json:"networkunit_id" bson:"networkunit_id"`
	NetworkUnitName string    `json:"networkunit_name" bson:"networkunit_name"`
	AccessPointID   int64     `json:"accesspoint_id" bson:"accesspoint_id"`
	AccessPointName string    `json:"accesspoint_name" bson:"accesspoint_name"`
	OperateTime     time.Time `json:"operate_time" bson:"operate_time"`
	Operator        string    `json:"operator" bson:"operator"`
}

// TableTopoEvent represents the complete db structures of an event topo.
type TableTopoEvent base.TableBroker[*TopoEvent]
