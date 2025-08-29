/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package globalsettings provides storage for global settings.
package globalsettings

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
)

// TableName global settings table name.
func TableName() string {
	return "global_settings"
}

var _ base.IData = &GlobalSettings{}

// GlobalSettings represents the table of global settings.
type GlobalSettings struct {
	SettingName string `bson:"setting_name" json:"setting_name"`
	Value       string `bson:"value" json:"value"`
}

// UniqueFields unique fields of the table.
func (event *GlobalSettings) UniqueFields() []string {
	return []string{FieldKeySettingName}
}

// UniqueKey unique key of the table.
func (event *GlobalSettings) UniqueKey() string {
	return event.SettingName
}

// TableGlobalSettings represent the complete db structures of global settings.
type TableGlobalSettings base.TableBroker[*GlobalSettings]
