/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package networkunit

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
)

// TableName networkunit table name.
func TableName() string {
	return "networkunit"
}

var _ base.IData = &NetworkUnit{}

// NetworkUnit represents network unit table.
// NetworkUnitID should be the unique key.
type NetworkUnit struct {
	TenantID        string `json:"tenant_id" bson:"tenant_id"`
	NetworkUnitID   int64  `json:"networkunit_id" bson:"networkunit_id"`
	NetworkUnitName string `json:"networkunit_name" bson:"networkunit_name"`

	NetworkAreaID      int64                         `json:"networkarea_id" bson:"networkarea_id"`
	AccessPoints       []int64                       `json:"accesspoints" bson:"accesspoints"`
	Links              *Links                        `json:"links" bson:"links"`
	IsDirect           bool                          `json:"is_direct" bson:"is_direct"`
	DirectEndpoints    *Endpoints                    `json:"direct_endpoints" bson:"direct_endpoints"`
	Generation         int64                         `json:"generation" bson:"generation"`
	CustomDeployConfig map[string]CustomDeployConfig `json:"custom_deploy_config" bson:"custom_deploy_config"`
}

// UniqueFields unique fields of the table.
func (networkunit *NetworkUnit) UniqueFields() []string {
	return []string{
		FieldKeyNetworkUnitID,
		FieldKeyNetworkAreaID,
	}
}

// UniqueKey unique key of the table.
func (networkunit *NetworkUnit) UniqueKey() string {
	return fmt.Sprintf("%d_%d", networkunit.NetworkAreaID, networkunit.NetworkUnitID)
}

// Link represents link target.
type Link struct {
	NetworkAreaID int64 `json:"networkarea_id" bson:"networkarea_id"`
	NetworkUnitID int64 `json:"networkunit_id" bson:"networkunit_id"`
	AccessPointID int64 `json:"accesspoint_id" bson:"accesspoint_id"`
}

// Links represents links.
type Links struct {
	Cluster *Link `json:"cluster" bson:"cluster"`
	File    *Link `json:"file" bson:"file"`
	Data    *Link `json:"data" bson:"data"`
}

// Endpoints represents endpoints of direct.
type Endpoints struct {
	Cluster []string `json:"cluster" bson:"cluster"`
	File    []string `json:"file" bson:"file"`
	Data    []string `json:"data" bson:"data"`
}

// InstallerRuntime defines the installer runtime.
type InstallerRuntime struct {
	BaseWorkDir string `json:"base_work_dir" bson:"base_work_dir"`
}

// NodeRuntime defines the node runtime.
type NodeRuntime struct {
	BaseDeployDir string `json:"base_deploy_dir" bson:"base_deploy_dir"`
	DataIPC       string `json:"data_ipc" bson:"data_ipc"`
	PluginIPC     string `json:"plugin_ipc" bson:"plugin_ipc"`
	LogDir        string `json:"log_dir" bson:"log_dir"`
}

// PluginRuntime defines the plugin runtime.
type PluginRuntime struct {
	BaseDeployDir string `json:"base_deploy_dir" bson:"base_deploy_dir"`
	LogDir        string `json:"log_dir" bson:"log_dir"`
}

// CustomDeployConfig represents custom deploy config of a network unit.
type CustomDeployConfig struct {
	InstallerRuntime InstallerRuntime `json:"installer_runtime" bson:"installer_runtime"`
	NodeRuntime      NodeRuntime      `json:"node_runtime" bson:"node_runtime"`
	PluginRuntime    PluginRuntime    `json:"plugin_runtime" bson:"plugin_runtime"`
}

// TableNetworkUnit represent the complete db structures of a networkunit.
type TableNetworkUnit base.TableBroker[*NetworkUnit]
