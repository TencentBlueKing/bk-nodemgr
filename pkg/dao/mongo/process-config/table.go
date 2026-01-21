/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package processconfig provides storage for process configuration.
package processconfig

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
)

// TableName scheduled workflow table name.
func TableName(tenantID string) string {
	return fmt.Sprintf("process_config_%s", tenantID)
}

var _ base.IData = &ProcessConfig{}

// ProcessConfig represents the table of process configuration.
type ProcessConfig struct {
	Name         string `json:"name" bson:"name"`
	ProcessName  string `json:"process_name" bson:"process_name"`
	HostID       int64  `json:"host_id" bson:"host_id"`
	IsMainConfig bool   `json:"is_main_config" bson:"is_main_config"`
	Content      string `json:"content" bson:"content"`
	MD5          string `json:"md5" bson:"md5"`
	FilePath     string `json:"file_path" bson:"file_path"`
}

// UniqueFields unique fields of the table.
func (config *ProcessConfig) UniqueFields() []string {
	return []string{FieldKeyName, FieldKeyProcessName, FieldKeyHostID}
}

// UniqueKey unique key of the table.
func (config *ProcessConfig) UniqueKey() string {
	return fmt.Sprintf("%s_%s_%d", config.Name, config.ProcessName, config.HostID)
}

// TableScheduledWorkflow represent the complete db structures of scheduled workflow.
type TableScheduledWorkflow base.TableBroker[*ProcessConfig]
