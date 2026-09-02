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

package packageexport

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
)

// TableName is the fixed package export collection name.
const TableName = "package_export"

var _ base.IData = &Data{}

// Data represents the package export document payload.
type Data struct {
	ExportID   string `json:"export_id" bson:"export_id"`
	WorkflowID string `json:"workflow_id" bson:"workflow_id"`
	TenantID   string `json:"tenant_id" bson:"tenant_id"`

	StorageKey   string `json:"storage_key" bson:"storage_key"`
	DownloadName string `json:"download_name" bson:"download_name"`
	MD5          string `json:"md5" bson:"md5"`
	Size         int64  `json:"size" bson:"size"`

	Operator string `json:"operator" bson:"operator"`
}

// UniqueFields returns the unique fields of the package export document.
func (data *Data) UniqueFields() []string {
	return []string{FieldKeyExportID}
}

// UniqueKey returns the unique key of the package export document.
func (data *Data) UniqueKey() string {
	return data.ExportID
}

// Table represents the complete package export MongoDB document.
type Table base.TableBroker[*Data]
