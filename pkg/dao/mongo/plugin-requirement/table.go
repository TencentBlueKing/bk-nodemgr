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

package pluginrequirement

import (
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
)

// FieldKeyID defines the record ID field key.
const FieldKeyID = "data.id"

// TableName returns the requirement collection for a tenant.
func TableName(tenantID string) string {
	return fmt.Sprintf("plugin_requirement_%s", tenantID)
}

var _ base.IData = &Data{}

// Data stores requirement file metadata without file contents or deployment targets.
type Data struct {
	ID        string    `json:"id" bson:"id"`
	Name      string    `json:"name" bson:"name"`
	SavedName string    `json:"saved_name" bson:"saved_name"`
	MD5       string    `json:"md5" bson:"md5"`
	Size      int64     `json:"size" bson:"size"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
	Operator  string    `json:"operator" bson:"operator"`
}

// UniqueFields returns the fields identifying a requirement record.
func (d *Data) UniqueFields() []string {
	return []string{FieldKeyID}
}

// UniqueKey returns the stable requirement record ID.
func (d *Data) UniqueKey() string {
	return d.ID
}

// Table represents the complete requirement document envelope.
type Table base.TableBroker[*Data]
