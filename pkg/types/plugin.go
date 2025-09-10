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

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
)

// PluginType represents the type of plugin.
type PluginType string

const (
	// PluginTypeUnknown represents unknown plugin type.
	PluginTypeUnknown PluginType = "unknown"
	// PluginTypeOfficial represents official plugin.
	PluginTypeOfficial PluginType = "official"
	// PluginTypeExternal represents external plugin.
	PluginTypeExternal PluginType = "external"
)

// ConvPluginTypeToReleaseType convert plugin type to release type.
func ConvPluginTypeToReleaseType(pluginType PluginType) (ReleaseType, error) {
	switch pluginType {
	case PluginTypeOfficial:
		return ReleaseTypeOfficialPlugin, nil
	case PluginTypeExternal:
		return ReleaseTypeExternalPlugin, nil
	default:
		return "", fmt.Errorf("unsupport plugin type: %s", pluginType)
	}
}

// Plugin define the all info of plugin.
type Plugin struct {
	HostID     int64
	Type       PluginType
	Generation Generation
	Platform   platform.Platform
	Version    string
}
