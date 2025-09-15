/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package release

import (
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
)

// TableName release table name.
func TableName(releaseType string) string {
	return fmt.Sprintf("release_%s", releaseType)
}

var _ base.IData = &Release{}

// Release presents a release.
type Release struct {
	Name         string         `json:"name" bson:"name"`
	Generation   int64          `json:"generation" bson:"generation"`
	Type         string         `json:"type" bson:"type"`
	Version      string         `json:"version" bson:"version"`
	CPUArch      string         `json:"cpu_arch" bson:"cpu_arch"`
	OSType       string         `json:"os_type" bson:"os_type"`
	Labels       []string       `json:"labels" bson:"labels"`
	Enabled      bool           `json:"enabled" bson:"enabled"`
	AsDefault    bool           `json:"as_default" bson:"as_default"`
	ChangeLogEN  string         `json:"change_log_en" bson:"change_log_en"`
	ChangeLogZH  string         `json:"change_log_zh" bson:"change_log_zh"`
	FileName     string         `json:"filename" bson:"filename"`
	MD5          string         `json:"md5" bson:"md5"`
	UpdatedAt    time.Time      `json:"updated_at" bson:"updated_at"`
	Operator     string         `json:"operator" bson:"operator"`
	AdditionInfo map[string]any `json:"addition_info" bson:"addition_info"`
}

// UniqueFields unique fields of the table.
func (r *Release) UniqueFields() []string {
	return []string{
		FieldKeyName,
		FieldKeyGeneration,
		FieldKeyType,
		FieldKeyCPUArch,
		FieldKeyOSType,
		FieldKeyVersion,
	}
}

// UniqueKey unique key of the table.
func (r *Release) UniqueKey() string {
	return fmt.Sprintf("%s_%d_%s_%s_%s_%s", r.Name, r.Generation, r.Type, r.CPUArch, r.OSType, r.Version)
}

// TableRelease represents the complete db structures.
type TableRelease base.TableBroker[*Release]
