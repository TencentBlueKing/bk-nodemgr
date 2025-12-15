/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package types

// PluginGroupDefault default plugin group name.
const PluginGroupDefault = "default"

// Plugin define the all info of plugin.
type Plugin struct {
	TenantID string

	Name    string
	PkgName string
	Group   string
	Memo    string
}

// PluginInstallParam describe the plugin install param.
type PluginInstallParam struct {
	HostID     int64
	PluginName string
	Version    string
}

// PluginListParam describe the plugin list param.
type PluginListParam struct {
}

// PluginApplySubConfigParam defines the plugin apply sub config param.
type PluginApplySubConfigParam struct {
	HostID              int64
	PluginName          string
	Version             string
	ConfigName          []string
	CustomConfigContext map[string]any
}
