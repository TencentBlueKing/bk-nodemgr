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

import "fmt"

// PluginType defines the plugin type.
type PluginType string

const (
	// PluginTypeOfficial defines the official plugin type.
	PluginTypeOfficial PluginType = "official"

	// PluginTypeExternal defines the external plugin type.
	PluginTypeExternal PluginType = "external"
)

// Validate validates the plugin type.
func (pt PluginType) Validate() error {
	switch pt {
	case PluginTypeOfficial, PluginTypeExternal:
		return nil
	default:
		return fmt.Errorf("plugin type is invalid, plugin-type(%s)", pt)
	}
}
