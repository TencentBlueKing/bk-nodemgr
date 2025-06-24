/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package trigger ...
package trigger

import (
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
)

// TableName trigger table name.
const TableName = "trigger"

// Trigger represents a trigger.
// TriggerID should be the unique key.
type Trigger struct {
	TriggerID        string           `json:"trigger_id" bson:"trigger_id"`
	Category         string           `json:"category" bson:"category"`
	MetadataOnce     MetadataOnce     `json:"metadata_once" bson:"metadata_once"`
	MetadataPeriodic MetadataPeriodic `json:"metadata_periodic" bson:"metadata_periodic"`
	MetadataOrdered  MetadataOrdered  `json:"metadata_ordered" bson:"metadata_ordered"`
	State            string           `json:"state" bson:"state"`
	CreatedAt        time.Time        `json:"created_at" bson:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at" bson:"updated_at"`
	LastTriggeredAt  time.Time        `json:"last_triggered_at" bson:"last_triggered_at"`
}

// UniqueKey unique key of the table.
func (t *Trigger) UniqueKey() string {
	return t.TriggerID
}

// TableTrigger represents the complete db structures.
type TableTrigger base.TableBroker[*Trigger]

// MetadataOnce will store the metadata of a trigger.
type MetadataOnce struct {
}

// MetadataPeriodic will store the metadata of a trigger.
type MetadataPeriodic struct {
	Interval           string `json:"interval" bson:"interval"`
	AllowedConcurrency bool   `json:"allowed_concurrency" bson:"allowed_concurrency"`
}

// MetadataOrdered will store the metadata of a trigger.
type MetadataOrdered struct {
	MaxConcurrencyNum int `json:"max_concurrency_num" bson:"max_concurrency_num"`
}
