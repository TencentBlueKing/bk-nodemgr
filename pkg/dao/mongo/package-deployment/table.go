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
	"time"

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
	Release                []release              `json:"release" bson:"release"`
	Upload                 uploadInfo             `json:"upload" bson:"upload"`
	ImportPluginPkgOptions importPluginPkgOptions `json:"import_plugin_pkg_options" bson:"import_plugin_pkg_options"`
}

type importPluginPkgOptions struct {
	FileSourceType string `json:"file_source_type" bson:"file_source_type"`
	FileSource     string `json:"file_source" bson:"file_source"`
	FileName       string `json:"file_name" bson:"file_name"`
	MD5            string `json:"md5" bson:"md5"`
}

type uploadInfo struct {
	UploadID  string     `json:"upload_id" bson:"upload_id"`
	Name      string     `json:"name" bson:"name"`
	Version   string     `json:"version" bson:"version"`
	Platforms []platform `json:"platforms" bson:"platforms"`
}

type platform struct {
	OS   string `json:"os" bson:"os"`
	Arch string `json:"arch" bson:"arch"`
}

type release struct {
	Name         string         `json:"name" bson:"name"`
	Generation   int64          `json:"generation" bson:"generation"`
	Type         string         `json:"type" bson:"type"`
	Version      string         `json:"version" bson:"version"`
	CPUArch      string         `json:"cpu_arch" bson:"cpu_arch"`
	OSType       string         `json:"os_type" bson:"os_type"`
	Labels       []string       `json:"labels" bson:"labels"`
	FileName     string         `json:"filename" bson:"filename"`
	MD5          string         `json:"md5" bson:"md5"`
	Enabled      bool           `json:"enabled" bson:"enabled"`
	IsHidden     bool           `json:"is_hidden" bson:"is_hidden"`
	AsDefault    bool           `json:"as_default" bson:"as_default"`
	UpdatedAt    time.Time      `json:"updated_at" bson:"updated_at"`
	Operator     string         `json:"operator" bson:"operator"`
	AdditionInfo map[string]any `json:"addition_info" bson:"addition_info"`
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
