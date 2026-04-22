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
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
)

// ConfigPolicy defines the config policy.
type ConfigPolicy struct {
	TenantID      string
	Version       int
	ID            int64
	Name          string
	Type          ConfigPolicyType
	BizID         int64
	Remark        string
	Scopes        []ConfigPolicyScope
	TargetHostIDs []int64
	Configs       map[string]any
	Enabled       bool
	Priority      int64
	UpdatedAt     time.Time
	Operator      string
}

// ConfigPolicyMatchedPolicy describes a single matched policy in preview results.
type ConfigPolicyMatchedPolicy struct {
	PolicyID   int64
	PolicyName string
	Priority   int64
}

// ConfigPolicyMatchResult holds the match result.
type ConfigPolicyMatchResult struct {
	HostID          int64
	MatchedPolicies []ConfigPolicyMatchedPolicy
	MergedConfig    map[string]any
}

// ConfigPolicyPreviewResult holds the preview results.
type ConfigPolicyPreviewResult struct {
	ReliableResults   []ConfigPolicyMatchResult
	UnreliableResults []ConfigPolicyMatchResult
}

// ConfigPolicyPreviewHost carries resolved host attributes for preview matching.
type ConfigPolicyPreviewHost struct {
	HostID        int64
	NetworkAreaID int64
	NetworkUnitID int64
	OSType        criteria.OSType
	CPUArch       criteria.CPUArch
}

// ConfigPolicyScope defines the config policy scope.
type ConfigPolicyScope struct {
	NetworkAreaID int64
	NetworkUnitID int64
	NodeOsType    criteria.OSType
	NodeCPUArch   criteria.CPUArch
}

const (
	// ConfigPolicyScopeAnyID indicates a scope dimension matches any network area or unit.
	ConfigPolicyScopeAnyID int64 = -1

	// ConfigPolicyScopeAnyOSType indicates a scope matches any OS type.
	ConfigPolicyScopeAnyOSType criteria.OSType = ""

	// ConfigPolicyScopeAnyCPUArch indicates a scope matches any CPU architecture.
	ConfigPolicyScopeAnyCPUArch criteria.CPUArch = ""
)

// ConfigPolicyTemplate defines the config policy template.
type ConfigPolicyTemplate struct {
	TenantID string
	ID       int64
	Template string
}

// ConfigPolicyOptionSet defines the config policy option set.
type ConfigPolicyOptionSet struct {
	Agent  []ConfigPolicyTemplateBlock
	Proxy  []ConfigPolicyTemplateBlock
	Plugin []ConfigPolicyTemplateBlock
}

// GetOptionsByPolicyType gets the config policy template blocks by policy type.
func (set ConfigPolicyOptionSet) GetOptionsByPolicyType(policyType ConfigPolicyType) ([]ConfigPolicyTemplateBlock, error) {
	switch policyType {
	case ConfigPolicyTypeAgent:
		return set.Agent, nil
	case ConfigPolicyTypeProxy:
		return set.Proxy, nil
	case ConfigPolicyTypePlugin:
		return set.Plugin, nil
	default:
		return nil, fmt.Errorf("invalid config policy type, type(%s)", policyType)
	}
}

// ConfigPolicyTemplateBlock defines the config policy template block.
type ConfigPolicyTemplateBlock struct {
	ID      string                     `json:"id"`
	TitleEN string                     `json:"title_en"`
	TitleZH string                     `json:"title_zh"`
	Items   []ConfigPolicyTemplateItem `json:"items"`
}

// ConfigPolicyTemplateItem defines the config policy template item.
type ConfigPolicyTemplateItem struct {
	Enabled           bool                     `json:"enabled"`
	ID                string                   `json:"id"`
	NameEN            string                   `json:"name_en"`
	NameZH            string                   `json:"name_zh"`
	RemarkEN          string                   `json:"remark_en"`
	RemarkZH          string                   `json:"remark_zh"`
	Key               string                   `json:"key"`
	Type              ConfigPolicyTemplateType `json:"type"`
	ValueString       string                   `json:"value_string"`
	ValueInt          int64                    `json:"value_int"`
	ValueBool         bool                     `json:"value_bool"`
	ValueStringSelect []string                 `json:"value_string_select"`
	ValueIntSelect    []int64                  `json:"value_int_select"`

	/**
	 * ValueGroup is only for definition, will not return to frontend.
	**/

	// ValueGroupFixed defines the fixed value to set.
	ValueGroupFixed map[string]any `json:"value_group_fixed"`

	// ValueGroupAssigned defines the value assgined from Value(Type) to set.
	// if the value type is different from Type, it will be ignored.
	ValueGroupAssigned map[string]any `json:"value_group_assigned"`
}

// ConfigPolicyTemplateType defines the config policy config template type.
type ConfigPolicyTemplateType int

const (
	// ConfigPolicyTemplateTypeString defines the config policy template type string.
	// regards value_string as the value.
	ConfigPolicyTemplateTypeString ConfigPolicyTemplateType = 0

	// ConfigPolicyTemplateTypeInt defines the config policy template type int.
	// regards value_int as the value.
	ConfigPolicyTemplateTypeInt ConfigPolicyTemplateType = 1

	// ConfigPolicyTemplateTypeBool defines the config policy template type bool.
	// regards value_bool as the value.
	ConfigPolicyTemplateTypeBool ConfigPolicyTemplateType = 2

	// ConfigPolicyTemplateTypeStringSelect defines the config policy template type string select.
	// regards value_string_select as the options, and value_string as the value.
	ConfigPolicyTemplateTypeStringSelect ConfigPolicyTemplateType = 3

	// ConfigPolicyTemplateTypeIntSelect defines the config policy template type int select.
	// regards value_int_select as the options, and value_int as the value.
	ConfigPolicyTemplateTypeIntSelect ConfigPolicyTemplateType = 4
)

// ConfigPolicyPriorityDisabled is the priority value for disabled policies.
const ConfigPolicyPriorityDisabled int64 = -1

// ConfigPolicyType defines the policy type.
type ConfigPolicyType string

const (
	// ConfigPolicyTypeAgent defines the agent policy type.
	ConfigPolicyTypeAgent ConfigPolicyType = "config_policy_agent"

	// ConfigPolicyTypeProxy defines the proxy policy type.
	ConfigPolicyTypeProxy ConfigPolicyType = "config_policy_proxy"

	// ConfigPolicyTypePlugin defines the plugin policy type.
	ConfigPolicyTypePlugin ConfigPolicyType = "config_policy_plugin"
)

// Validate validates the policy type.
func (policyType ConfigPolicyType) Validate() error {
	switch policyType {
	case ConfigPolicyTypeAgent, ConfigPolicyTypeProxy, ConfigPolicyTypePlugin:
		return nil
	default:
		return fmt.Errorf("invalid config policy type, type(%s)", policyType)
	}
}

// ConfigPolicyTypeListToStringList converts a config policy evnt type list to a string list.
func ConfigPolicyTypeListToStringList(configPolicyTypeList []ConfigPolicyType) []string {
	data := make([]string, len(configPolicyTypeList))
	for idx, configPolicyType := range configPolicyTypeList {
		data[idx] = string(configPolicyType)
	}

	return data
}

// StringListToConfigPolicyTypeList converts a string list to a config policy type list.
func StringListToConfigPolicyTypeList(stringList []string) ([]ConfigPolicyType, error) {
	data := make([]ConfigPolicyType, len(stringList))
	for idx, configPolicyType := range stringList {
		if err := ConfigPolicyType(configPolicyType).Validate(); err != nil {
			return nil, err
		}
		data[idx] = ConfigPolicyType(configPolicyType)
	}

	return data, nil
}

// ConvertNodeRoleToConfigPolicyType converts node role to config policy type.
func ConvertNodeRoleToConfigPolicyType(nodeRole NodeRole) (ConfigPolicyType, error) {
	switch nodeRole {
	case NodeRoleAgent:
		return ConfigPolicyTypeAgent, nil
	case NodeRoleProxy:
		return ConfigPolicyTypeProxy, nil
	default:
		return "", fmt.Errorf("invalid node role. role(%s)", nodeRole)
	}
}

const (
	// ConfigOptionDefinitionFileNameAgent defines the config option definition file name for agent.
	ConfigOptionDefinitionFileNameAgent = "agent_option.json"

	// ConfigOptionDefinitionFileNameProxy defines the config option definition file name for proxy.
	ConfigOptionDefinitionFileNameProxy = "proxy_option.json"

	// ConfigOptionDefinitionFileNamePlugin defines the config option definition file name for plugin.
	ConfigOptionDefinitionFileNamePlugin = "plugin_option.json"
)

// FlatMapToNestedMap converts a flat map with dot-notation keys into a nested map.
// For example, {"agent.access.enable_fake_seed": true} becomes {"agent":{"access":{"enable_fake_seed":true}}}.
func FlatMapToNestedMap(flat map[string]any) map[string]any {
	nested := make(map[string]any)
	for key, value := range flat {
		parts := strings.Split(key, ".")
		current := nested
		for i, part := range parts {
			if i == len(parts)-1 {
				current[part] = value
			} else {
				next, ok := current[part].(map[string]any)
				if !ok {
					next = make(map[string]any)
					current[part] = next
				}
				current = next
			}
		}
	}
	return nested
}

// NestedMapToFlatMap converts a nested map into a flat map with dot-notation keys.
// For example, {"agent":{"access":{"enable_fake_seed":true}}} becomes {"agent.access.enable_fake_seed": true}.
func NestedMapToFlatMap(nested map[string]any) map[string]any {
	flat := make(map[string]any)
	flattenNestedMap(nested, "", flat)
	return flat
}

func flattenNestedMap(nested map[string]any, prefix string, flat map[string]any) {
	for key, value := range nested {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}
		if sub, ok := value.(map[string]any); ok {
			flattenNestedMap(sub, fullKey, flat)
		} else {
			flat[fullKey] = value
		}
	}
}

// MergedConfigToNestedJSON converts a flat dot-notation config map into a sorted nested JSON string.
// Returns "{}" for nil or empty maps.
func MergedConfigToNestedJSON(flat map[string]any) string {
	if len(flat) == 0 {
		return "{}"
	}
	nested := FlatMapToNestedMap(flat)
	data, err := json.Marshal(nested)
	if err != nil {
		return "{}"
	}
	return string(data)
}
