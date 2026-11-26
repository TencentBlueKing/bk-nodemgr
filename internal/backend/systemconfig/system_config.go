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

// Package systemconfig holds the read-only business configuration view for Backend.
package systemconfig

import (
	"slices"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// ISystemConfig exposes a snapshot without exposing startup configuration objects.
type ISystemConfig interface {
	// Get returns an independent copy of the loaded business configuration view.
	Get() types.SystemConfig
}

type systemConfig struct {
	value types.SystemConfig
}

// New captures the explicitly selected fields injected by the service assembly.
func New(value types.SystemConfig) ISystemConfig {
	value.GSEDeployConfigs = cloneDeployConfigs(value.GSEDeployConfigs)
	return &systemConfig{value: value}
}

func (s *systemConfig) Get() types.SystemConfig {
	value := s.value
	value.GSEDeployConfigs = cloneDeployConfigs(value.GSEDeployConfigs)

	return value
}

func cloneDeployConfigs(configs []types.SystemConfigGSEDeploy) []types.SystemConfigGSEDeploy {
	cloned := slices.Clone(configs)
	for i := range cloned {
		cloned[i].FirewallFeatures = slices.Clone(cloned[i].FirewallFeatures)
	}

	return cloned
}
