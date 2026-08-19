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

package upload

import (
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
)

// TableName upload table name.
func TableName(category string) string {
	return fmt.Sprintf("upload_%s", category)
}

var _ base.IData = &Upload{}

// Upload presents a upload table.
type Upload struct {
	UploadID  string    `json:"upload_id" bson:"upload_id"`
	Category  string    `json:"category" bson:"category"`
	SavedName string    `json:"saved_name" bson:"saved_name"`
	Operator  string    `json:"operator" bson:"operator"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}

// UniqueFields unique fields of the table.
func (u *Upload) UniqueFields() []string {
	return []string{FieldKeyUploadID}
}

// UniqueKey unique key of the table.
func (u *Upload) UniqueKey() string {
	return u.UploadID
}

// TableUpload represents the complete db structures of a upload.
type TableUpload base.TableBroker[*Upload]
