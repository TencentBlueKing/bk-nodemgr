/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package packageevent

import (
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
)

const tableNamePrefix = "packageevent"

// TableName package event table name.
func TableName(tenantID string) string {
	return fmt.Sprintf("%s_%s", tableNamePrefix, tenantID)
}

var _ base.IData = &PackageEvent{}

// PackageEvent represents package event table.
type PackageEvent struct {
	TenantID    string    `json:"tenant_id" bson:"tenant_id"`
	Name        string    `json:"name" bson:"name"`
	EventID     int64     `json:"event_id" bson:"event_id"`
	EventType   string    `json:"event_type" bson:"event_type"`
	Generation  int64     `json:"generation" bson:"generation"`
	ReleaseType string    `json:"release_type" bson:"release_type"`
	Version     string    `json:"version" bson:"version"`
	CPUArch     string    `json:"cpu_arch" bson:"cpu_arch"`
	OSType      string    `json:"os_type" bson:"os_type"`
	OperateTime time.Time `json:"operate_time" bson:"operate_time"`
	Operator    string    `json:"operator" bson:"operator"`
}

// UniqueFields unique fields of the table.
func (event *PackageEvent) UniqueFields() []string {
	return []string{
		FieldKeyEventID,
	}
}

// UniqueKey unique key of the table.
func (event *PackageEvent) UniqueKey() string {
	return fmt.Sprintf("%d", event.EventID)
}

// TablePackageEvent represents the complete db structures.
type TablePackageEvent base.TableBroker[*PackageEvent]
