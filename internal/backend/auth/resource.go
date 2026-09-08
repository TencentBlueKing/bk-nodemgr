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

package auth

import "github.com/TencentBlueKing/bk-nodemgr/pkg/types"

// BuildPackageResources constructs canonical package resources for permission checks.
// For agent/proxy/cert/bintool types, the resource ID is the release type itself.
// For plugin types, the resource ID is the plugin name.
func BuildPackageResources(releaseType ...string) []types.AuthResource {
	resources := make([]types.AuthResource, 0, len(releaseType))
	for _, rt := range releaseType {
		resources = append(resources, types.AuthResource{
			SystemID: types.SystemIDNodeMgr,
			Type:     types.AuthResourceTypePackage,
			ID:       rt,
		})
	}

	return resources
}
