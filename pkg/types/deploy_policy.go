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
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
)

// DeployPolicy defines the deploy policy.
type DeployPolicy struct {
	TenantID       string
	DeployPolicyID int64

	DeployPolicyMeta *DeployPolicyMeta

	Scopes []*Scope
	Specs  []*DeploySpec
}

// DeployPolicyMeta defines the deploy policy meta.
type DeployPolicyMeta struct {
	CreateAt time.Time
}

// DeploySpecType defines the deploy spec type.
type DeploySpecType string

const (
	// DeploySpecTypeSpecifyAgent defines specify agent.
	DeploySpecTypeSpecifyAgent DeploySpecType = "specify_agent"

	// DeploySpecTypeSpecifyPlugin defines specify plugin.
	DeploySpecTypeSpecifyPlugin DeploySpecType = "specify_plugin"
)

// Validate validates the deploy spec type.
func (deploySpecType DeploySpecType) Validate() error {
	switch deploySpecType {
	case DeploySpecTypeSpecifyAgent, DeploySpecTypeSpecifyPlugin:
		return nil
	default:
		return fmt.Errorf("invalid deploy spec type(%s)", deploySpecType)
	}
}

func conflictDeploySpecTypes(specType DeploySpecType) []DeploySpecType {
	switch specType {
	// TODO: 补充 proxy 的冲突
	case DeploySpecTypeSpecifyAgent:
		return []DeploySpecType{}
	default:
		return nil
	}
}

// IsConflict returns whether the deploy spec type is conflict with other.
func (deploySpecType DeploySpecType) IsConflict(other DeploySpecType) bool {
	if deploySpecType == other {
		return true
	}

	conflictTypes := conflictDeploySpecTypes(deploySpecType)
	for _, conflictType := range conflictTypes {
		if conflictType == other {
			return true
		}
	}

	return false
}

// DeploySpec defines the deploy spec of the deploy policy.
type DeploySpec struct {
	Type  DeploySpecType
	Param map[string]any
}

// UniqueID returns the unique id of the deploy spec.
func (spec DeploySpec) UniqueID() (string, error) {
	switch spec.Type {
	case DeploySpecTypeSpecifyPlugin:
		param, err := spec.GetSpecifyPluginParam()
		if err != nil {
			return "", fmt.Errorf("failed to get specify plugin param: %w", err)
		}

		// TODO: 继续补充此处代码
		return param.PluginName, nil
	}

	return "", nil
}

// SpecifyPluginParam defines the specify plugin version param.
type SpecifyPluginParam struct {
	PluginName          string
	PluginVersion       string
	CustomConfigContext map[string]any
}

// GetSpecifyPluginParam returns the specify plugin param.
func (spec DeploySpec) GetSpecifyPluginParam() (*SpecifyPluginParam, error) {
	param := &SpecifyPluginParam{}
	if err := conv.MapToStruct(spec.Param, param); err != nil {
		return nil, fmt.Errorf("failed to convert param to specify plugin param: %w", err)
	}

	return param, nil
}

// SpecifyAgentParam defines the specify plugin version param.
type SpecifyAgentParam struct {
	NodeVersion string
}

// GetSpecifyAgentParam returns the specify agent param.
func (spec DeploySpec) GetSpecifyAgentParam() (*SpecifyAgentParam, error) {
	param := &SpecifyAgentParam{}
	if err := conv.MapToStruct(spec.Param, param); err != nil {
		return nil, fmt.Errorf("failed to convert param to specify agent param: %w", err)
	}

	return param, nil
}

// SpecifyPluginSubConfigParam defines the specify plugin sub config param.
type SpecifyPluginSubConfigParam struct {
	PluginName          string
	ConfigFilesDetail   []*PluginConfigDetail
	CustomConfigContext map[string]any
}

// GetSpecifyPluginSubConfigParam returns the specify plugin param.
func (spec DeploySpec) GetSpecifyPluginSubConfigParam() (*SpecifyPluginSubConfigParam, error) {
	param := &SpecifyPluginSubConfigParam{}
	if err := conv.MapToStruct(spec.Param, param); err != nil {
		return nil, fmt.Errorf("failed to convert param to specify plugin param: %w", err)
	}

	return param, nil
}
