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
)

// DeployPolicy defines the deploy policy.
type DeployPolicy struct {
	DeployPolicyID int64

	Meta DeployPolicyMeta

	Scopes []*Scope
	Specs  []*DeploySpec

	Operator string
	Enabled  bool

	LifeCycle DeployPolicyLifeCycle
}

// DeployPolicyMeta defines the deploy policy meta.
type DeployPolicyMeta struct {
	Name        string
	Description string
}

// DeployPolicyLifeCycle defines the deploy policy life cycle.
type DeployPolicyLifeCycle struct {
	CreatedAt  time.Time
	UpdatedAt  time.Time
	ExecutedAt time.Time
}

// DeploySpecType defines the deploy spec type.
type DeploySpecType string

const (
	// DeploySpecTypeSpecifyAgent defines specify agent.
	DeploySpecTypeSpecifyAgent DeploySpecType = "specify_agent"

	// DeploySpecTypeSpecifyProxy defines specify proxy.
	DeploySpecTypeSpecifyProxy DeploySpecType = "specify_proxy"

	// DeploySpecTypeSpecifyPlugin defines specify plugin.
	DeploySpecTypeSpecifyPlugin DeploySpecType = "specify_plugin"

	// DeploySpecTypeSpecifyPluginPkg defines specify plugin pkg.
	DeploySpecTypeSpecifyPluginPkg DeploySpecType = "specify_plugin_pkg"

	// DeploySpecTypeSpecifyPluginSubConfig defines specify plugin sub config.
	DeploySpecTypeSpecifyPluginSubConfig DeploySpecType = "specify_plugin_sub_config"
)

// Validate validates the deploy spec type.
func (deploySpecType DeploySpecType) Validate() error {
	switch deploySpecType {
	case DeploySpecTypeSpecifyAgent, DeploySpecTypeSpecifyProxy, DeploySpecTypeSpecifyPlugin,
		DeploySpecTypeSpecifyPluginPkg, DeploySpecTypeSpecifyPluginSubConfig:
		return nil
	default:
		return fmt.Errorf("invalid deploy spec type(%s)", deploySpecType)
	}
}

func conflictDeploySpecTypes(specType DeploySpecType) []DeploySpecType {
	switch specType {
	case DeploySpecTypeSpecifyAgent:
		return []DeploySpecType{DeploySpecTypeSpecifyProxy}
	case DeploySpecTypeSpecifyPlugin:
		return []DeploySpecType{DeploySpecTypeSpecifyPluginPkg}
	case DeploySpecTypeSpecifyProxy:
		return []DeploySpecType{}
	case DeploySpecTypeSpecifyPluginPkg:
		return []DeploySpecType{}
	case DeploySpecTypeSpecifyPluginSubConfig:
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
	specType                    DeploySpecType
	paramSpecifyAgent           *SpecifyAgentParam
	paramSpecifyProxy           *SpecifyProxyParam
	paramSpecifyPlugin          *SpecifyPluginParam
	paramSpecifyPluginPkg       *SpecifyPluginPkgParam
	paramSpecifyPluginSubConfig *SpecifyPluginSubConfigParam
}

// Type returns the deploy spec type.
func (spec *DeploySpec) Type() DeploySpecType {
	return spec.specType
}

// NewDeploySpecWithSpecifyAgent creates a new DeploySpec with SpecifyAgent type.
func NewDeploySpecWithSpecifyAgent(param *SpecifyAgentParam) (*DeploySpec, error) {
	if param == nil {
		return nil, fmt.Errorf("param cannot be nil for DeploySpecTypeSpecifyAgent")
	}

	return &DeploySpec{
		specType:          DeploySpecTypeSpecifyAgent,
		paramSpecifyAgent: param,
	}, nil
}

// NewDeploySpecWithSpecifyPlugin creates a new DeploySpec with SpecifyPlugin type.
func NewDeploySpecWithSpecifyPlugin(param *SpecifyPluginParam) (*DeploySpec, error) {
	if param == nil {
		return nil, fmt.Errorf("param cannot be nil for DeploySpecTypeSpecifyPlugin")
	}

	return &DeploySpec{
		specType:           DeploySpecTypeSpecifyPlugin,
		paramSpecifyPlugin: param,
	}, nil
}

// NewDeploySpecWithSpecifyPluginPkg creates a new DeploySpec with SpecifyPluginPkg type.
func NewDeploySpecWithSpecifyPluginPkg(param *SpecifyPluginPkgParam) (*DeploySpec, error) {
	if param == nil {
		return nil, fmt.Errorf("param cannot be nil for DeploySpecTypeSpecifyPluginPkg")
	}

	return &DeploySpec{
		specType:              DeploySpecTypeSpecifyPluginPkg,
		paramSpecifyPluginPkg: param,
	}, nil
}

// NewDeploySpecWithSpecifyPluginSubConfig creates a new DeploySpec with SpecifyPluginSubConfig type.
func NewDeploySpecWithSpecifyPluginSubConfig(param *SpecifyPluginSubConfigParam) (*DeploySpec, error) {
	if param == nil {
		return nil, fmt.Errorf("param cannot be nil for DeploySpecTypeSpecifyPluginSubConfig")
	}

	return &DeploySpec{
		specType:                    DeploySpecTypeSpecifyPluginSubConfig,
		paramSpecifyPluginSubConfig: param,
	}, nil
}

// NewDeploySpecWithSpecifyProxy creates a new DeploySpec with SpecifyProxy type.
func NewDeploySpecWithSpecifyProxy(param *SpecifyProxyParam) (*DeploySpec, error) {
	if param == nil {
		return nil, fmt.Errorf("param cannot be nil for DeploySpecTypeSpecifyProxy")
	}

	return &DeploySpec{
		specType:          DeploySpecTypeSpecifyProxy,
		paramSpecifyProxy: param,
	}, nil
}

// UniqueID returns the unique id of the deploy spec.
func (spec *DeploySpec) UniqueID() (string, error) {
	switch spec.specType {
	case DeploySpecTypeSpecifyPlugin:
		if spec.paramSpecifyPlugin == nil {
			return "", fmt.Errorf("param_specify_plugin is nil")
		}

		return spec.paramSpecifyPlugin.PluginName, nil
	case DeploySpecTypeSpecifyPluginPkg:
		if spec.paramSpecifyPluginPkg == nil {
			return "", fmt.Errorf("param_specify_plugin_pkg is nil")
		}

		return spec.paramSpecifyPluginPkg.PluginPkgName, nil
	default:
		return "", fmt.Errorf("unsupported deploy spec type(%s)", spec.specType)
	}
}

// SpecifyPluginParam defines the specify plugin version param.
type SpecifyPluginParam struct {
	PluginName          string
	Version             string
	CustomConfigContext map[string]any
}

// GetSpecifyPluginParam returns the specify plugin param.
func (spec *DeploySpec) GetSpecifyPluginParam() (*SpecifyPluginParam, error) {
	if spec.paramSpecifyPlugin == nil {
		return nil, fmt.Errorf("param_specify_plugin is nil")
	}

	if err := spec.paramSpecifyPlugin.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate specify plugin param: %w", err)
	}

	return spec.paramSpecifyPlugin, nil
}

// Validate validates the specify plugin param.
func (param *SpecifyPluginParam) Validate() error {
	if param.PluginName == "" {
		return fmt.Errorf("plugin_name is required")
	}

	if param.Version == "" {
		return fmt.Errorf("version is required")
	}

	return nil
}

// SpecifyAgentParam defines the specify plugin version param.
type SpecifyAgentParam struct {
	NodeVersion string
}

// GetSpecifyAgentParam returns the specify agent param.
func (spec *DeploySpec) GetSpecifyAgentParam() (*SpecifyAgentParam, error) {
	if spec.paramSpecifyAgent == nil {
		return nil, fmt.Errorf("param_specify_agent is nil")
	}

	if err := spec.paramSpecifyAgent.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate specify agent param: %w", err)
	}

	return spec.paramSpecifyAgent, nil
}

// Validate validates the specify agent param.
func (param *SpecifyAgentParam) Validate() error {
	if param.NodeVersion == "" {
		return fmt.Errorf("node_version is required")
	}

	return nil
}

// SpecifyProxyParam defines the specify proxy param.
type SpecifyProxyParam struct {
	NodeVersion string
}

// GetSpecifyProxyParam returns the specify proxy param.
func (spec *DeploySpec) GetSpecifyProxyParam() (*SpecifyProxyParam, error) {
	if spec.paramSpecifyProxy == nil {
		return nil, fmt.Errorf("param_specify_proxy is nil")
	}

	if err := spec.paramSpecifyProxy.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate specify proxy param: %w", err)
	}

	return spec.paramSpecifyProxy, nil
}

// Validate validates the specify proxy param.
func (param *SpecifyProxyParam) Validate() error {
	if param.NodeVersion == "" {
		return fmt.Errorf("node_version is required")
	}

	return nil
}

// SpecifyPluginSubConfigParam defines the specify plugin sub config param.
type SpecifyPluginSubConfigParam struct {
	PluginName          string
	ConfigFilesDetail   []*PluginConfigDetail
	CustomConfigContext map[string]any
}

// GetSpecifyPluginSubConfigParam returns the specify plugin param.
func (spec *DeploySpec) GetSpecifyPluginSubConfigParam() (*SpecifyPluginSubConfigParam, error) {
	if spec.paramSpecifyPluginSubConfig == nil {
		return nil, fmt.Errorf("param_specify_plugin_sub_config is nil")
	}

	if err := spec.paramSpecifyPluginSubConfig.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate specify plugin sub config param: %w", err)
	}

	return spec.paramSpecifyPluginSubConfig, nil
}

// Validate validates the specify plugin sub config param.
func (param *SpecifyPluginSubConfigParam) Validate() error {
	if param.PluginName == "" {
		return fmt.Errorf("plugin_name is required")
	}

	return nil
}

// SpecifyPluginPkgParam defines the specify plugin pkg param.
type SpecifyPluginPkgParam struct {
	PluginPkgName       string
	Version             string
	CustomConfigContext map[string]any
}

// GetSpecifyPluginPkgParam returns the specify plugin pkg param.
func (spec *DeploySpec) GetSpecifyPluginPkgParam() (*SpecifyPluginPkgParam, error) {
	if spec.paramSpecifyPluginPkg == nil {
		return nil, fmt.Errorf("param_specify_plugin_pkg is nil")
	}

	if err := spec.paramSpecifyPluginPkg.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate specify plugin pkg param: %w", err)
	}

	return spec.paramSpecifyPluginPkg, nil
}

// Validate validates the specify plugin pkg param.
func (param *SpecifyPluginPkgParam) Validate() error {
	if param.PluginPkgName == "" {
		return fmt.Errorf("plugin_pkg_name is required")
	}

	if param.Version == "" {
		return fmt.Errorf("version is required")
	}

	return nil
}

// DeployPolicyFields represents the fields of DeployPolicy.
type DeployPolicyFields struct {
	Meta    bool
	Scopes  bool
	Specs   bool
	Enabled bool
}

// NewAllDeployPolicyFields returns all fields of DeployPolicy.
func NewAllDeployPolicyFields() DeployPolicyFields {
	return DeployPolicyFields{
		Meta:    true,
		Scopes:  true,
		Specs:   true,
		Enabled: true,
	}
}
