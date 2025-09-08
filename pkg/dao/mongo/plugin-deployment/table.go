/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package plugindeployment

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
)

// TableName plugin deployment table name.
const TableName = "plugin_deployment"

var _ base.IData = &Data{}

// Data represents the table of plugin deployment.
// Token should be the unique key.
type Data struct {
	Token string `json:"token" bson:"token"`
	Info  *Info  `json:"info" bson:"info"`
}

// Info this is the info of this plugin deployment.
type Info struct {
	ActionName       string          `json:"action_name" bson:"action_name"`
	InstallerWorkDir string          `json:"installer_work_dir" bson:"installer_work_dir"`
	Plugin           Plugin          `json:"plugin" bson:"plugin"`
	TransferOptions  TransferOptions `json:"transfer_options" bson:"transfer_options"`
	TargetVersion    []TargetVersion `json:"target_version" bson:"target_version"`
}

// Plugin defines the plugin.
type Plugin struct {
	HostID     int64    `json:"host_id" bson:"host_id"`
	Type       string   `json:"type" bson:"type"`
	Generation int64    `json:"generation" bson:"generation"`
	Platform   Platform `json:"platform" bson:"platform"`
	Version    string   `json:"version" bson:"version"`
}

// Platform defines the platform.
type Platform struct {
	OS   string `json:"os" bson:"os"`
	Arch string `json:"arch" bson:"arch"`
}

// TargetVersion defines the target version.
type TargetVersion struct {
	Platform Platform `json:"platform" bson:"platform"`
	Version  string   `json:"version" bson:"version"`
}

// InstallOptions this is the options for nodemgr tools.
type InstallOptions struct {
}

// UpgradeOptions this is the options for plugin upgrade.
type UpgradeOptions struct {
}

// RestartOptions this is the options for plugin restart.
type RestartOptions struct {
}

// TransferOptions this is the options for plugin transfer.
type TransferOptions struct {
	SelectDownloads      bool `json:"select_downloads" bson:"select_downloads"`
	EnableReleasePackage bool `json:"enable_release_package" bson:"enable_release_package"`
	EnableInstaller      bool `json:"enable_installer" bson:"enable_installer"`
}

// VersionSupports describes this version supports things.
type VersionSupports struct {
}

// UniqueFields unique fields of the table.
func (deploy *Data) UniqueFields() []string {
	return []string{FieldKeyToken}
}

// UniqueKey unique key of the table.
func (deploy *Data) UniqueKey() string {
	return deploy.Token
}

// Table represent the complete db structures of plugin deployment.
type Table base.TableBroker[*Data]
