/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package packagedeployment

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
)

// TableName is the name of table.
const TableName = "package_deployment"

var _ base.IData = &Data{}

// Data represents the table of package deployment.
type Data struct {
	Token string `json:"token" bson:"token"`
	Info  *Info  `json:"info" bson:"info"`
}

// Info represents package deployment detail in database.
type Info struct {
	UploadID               string                 `json:"upload_id" bson:"upload_id"`
	ImportPluginPkgOptions importPluginPkgOptions `json:"import_plugin_pkg_options" bson:"import_plugin_pkg_options"`
}

type importPluginPkgOptions struct {
	FileSourceType string     `json:"file_source_type" bson:"file_source_type"`
	FileSource     string     `json:"file_source" bson:"file_source"`
	MD5            string     `json:"md5" bson:"md5"`
	PluginPkgName  string     `json:"plugin_pkg_name" bson:"plugin_pkg_name"`
	PluginName     string     `json:"plugin_name" bson:"plugin_name"`
	Version        string     `json:"version" bson:"version"`
	Platforms      []platform `json:"platforms" bson:"platforms"`
}

type platform struct {
	OS   string `json:"os" bson:"os"`
	Arch string `json:"arch" bson:"arch"`
}

// UniqueFields returns unique fields of the table.
func (deploy *Data) UniqueFields() []string {
	return []string{FieldKeyToken}
}

// UniqueKey returns the unique key of the table.
func (deploy *Data) UniqueKey() string {
	return deploy.Token
}

// Table represents the complete db structures of package deployment.
type Table base.TableBroker[*Data]
