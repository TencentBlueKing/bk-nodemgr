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

package deploypolicy

import (
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
)

const tableNamePrefix = "deploypolicy"

// TableName deploy policy table name.
func TableName(tenantID string) string {
	return fmt.Sprintf("%s_%s", tableNamePrefix, tenantID)
}

var _ base.IData = &DeployPolicy{}

// DeployPolicy represents the table of deploy policy deployment.
// DeployPolicyID should be the unique key.
type DeployPolicy struct {
	TenantID       string    `json:"tenant_id" bson:"tenant_id"`
	DeployPolicyID int64     `json:"deploy_policy_id" bson:"deploy_policy_id"`
	DsuID          int64     `json:"dsu_id" bson:"dsu_id"`
	Meta           Meta      `json:"meta" bson:"meta"`
	Specs          []*Spec   `json:"specs" bson:"specs"`
	Scopes         []*Scope  `json:"scopes" bson:"scopes"`
	Enabled        bool      `json:"enabled" bson:"enabled"`
	EnsureAbsent   bool      `json:"ensure_absent" bson:"ensure_absent"`
	LifeCycle      LifeCycle `json:"life_cycle" bson:"life_cycle"`
	Operator       string    `json:"operator" bson:"operator"`
}

// UniqueFields unique fields of the table.
func (deploy *DeployPolicy) UniqueFields() []string {
	return []string{FieldKeyDeployPolicyID}
}

// UniqueKey unique key of the table.
func (deploy *DeployPolicy) UniqueKey() string {
	return fmt.Sprintf("%d", deploy.DeployPolicyID)
}

// Table represent the complete db structures of deploy policy deployment.
type Table base.TableBroker[*DeployPolicy]

// Meta represents the meta of deploy policy.
type Meta struct {
	Name        string `json:"name" bson:"name"`
	Description string `json:"description" bson:"description"`
}

// Spec represents the spec of deploy policy.
type Spec struct {
	Type                                    string                                       `json:"type" bson:"type"`
	ParamSpecifyAgent                       *SpecParamSpecifyAgent                       `json:"param_specify_agent,omitempty" bson:"param_specify_agent,omitempty"`
	ParamSpecifyProxy                       *SpecParamSpecifyProxy                       `json:"param_specify_proxy,omitempty" bson:"param_specify_proxy,omitempty"`
	ParamSpecifyPlugin                      *SpecParamSpecifyPlugin                      `json:"param_specify_plugin,omitempty" bson:"param_specify_plugin,omitempty"`
	ParamSpecifyPluginPkg                   *SpecParamSpecifyPluginPkg                   `json:"param_specify_plugin_pkg,omitempty" bson:"param_specify_plugin_pkg,omitempty"`
	ParamProjectPluginPkgToHosts            *SpecParamProjectPluginPkgToHosts            `json:"param_project_plugin_pkg_to_hosts,omitempty" bson:"param_project_plugin_pkg_to_hosts,omitempty"`                         //nolint:lll
	ParamProjectPluginConfigTemplateToHosts *SpecParamProjectPluginConfigTemplateToHosts `json:"param_project_plugin_config_template_to_hosts,omitempty" bson:"param_project_plugin_config_template_to_hosts,omitempty"` //nolint:lll
	ParamSpecifyPluginSubConfig             *SpecParamSpecifyPluginSubConfig             `json:"param_specify_plugin_sub_config,omitempty" bson:"param_specify_plugin_sub_config,omitempty"`                             //nolint:lll
	ParamSpecifyPluginSubConfigTemplate     *SpecParamSpecifyPluginSubConfigTemplate     `json:"param_specify_plugin_sub_config_template,omitempty" bson:"param_specify_plugin_sub_config_template,omitempty"`           //nolint:lll
}

// SpecParamSpecifyAgent represents the parameter for specify agent spec.
type SpecParamSpecifyAgent struct {
	NodeVersion string `json:"node_version" bson:"node_version"`
}

// SpecParamSpecifyProxy represents the parameter for specify proxy spec.
type SpecParamSpecifyProxy struct {
	NodeVersion string `json:"node_version" bson:"node_version"`
}

// SpecParamSpecifyPlugin represents the parameter for specify plugin spec.
type SpecParamSpecifyPlugin struct {
	PluginName          string         `json:"plugin_name" bson:"plugin_name"`
	Version             string         `json:"version" bson:"version"`
	CustomConfigContext map[string]any `json:"custom_config_context,omitempty" bson:"custom_config_context,omitempty"`
}

// SpecParamSpecifyPluginPkg represents the parameter for specify plugin pkg spec.
type SpecParamSpecifyPluginPkg struct {
	PluginPkgName       string         `json:"plugin_pkg_name" bson:"plugin_pkg_name"`
	Version             string         `json:"version" bson:"version"`
	CustomConfigContext map[string]any `json:"custom_config_context,omitempty" bson:"custom_config_context,omitempty"`
}

// SpecParamProjectPluginPkgToHosts represents the parameter for project plugin pkg to hosts spec.
type SpecParamProjectPluginPkgToHosts struct {
	PluginPkgName       string         `json:"plugin_pkg_name" bson:"plugin_pkg_name"`
	Version             string         `json:"version" bson:"version"`
	CustomConfigContext map[string]any `json:"custom_config_context,omitempty" bson:"custom_config_context,omitempty"`
	PlacementHostIDs    []int64        `json:"placement_host_ids" bson:"placement_host_ids"`
}

// SpecParamProjectPluginConfigTemplateToHosts represents the parameter for project plugin config template to hosts spec.
type SpecParamProjectPluginConfigTemplateToHosts struct {
	PluginName          string                    `json:"plugin_name" bson:"plugin_name"`
	ConfigFilesDetail   []*SpecPluginConfigDetail `json:"config_files_detail,omitempty" bson:"config_files_detail,omitempty"`
	CustomConfigContext map[string]any            `json:"custom_config_context,omitempty" bson:"custom_config_context,omitempty"`
}

// SpecParamSpecifyPluginSubConfig represents the parameter for specify plugin sub config spec.
type SpecParamSpecifyPluginSubConfig struct {
	PluginName          string                    `json:"plugin_name" bson:"plugin_name"`
	ConfigFilesDetail   []*SpecPluginConfigDetail `json:"config_files_detail,omitempty" bson:"config_files_detail,omitempty"`
	CustomConfigContext map[string]any            `json:"custom_config_context,omitempty" bson:"custom_config_context,omitempty"`
}

// SpecParamSpecifyPluginSubConfigTemplate represents the parameter for specify plugin sub config template spec.
type SpecParamSpecifyPluginSubConfigTemplate struct {
	PluginName          string                    `json:"plugin_name" bson:"plugin_name"`
	ConfigFilesDetail   []*SpecPluginConfigDetail `json:"config_files_detail,omitempty" bson:"config_files_detail,omitempty"`
	CustomConfigContext map[string]any            `json:"custom_config_context,omitempty" bson:"custom_config_context,omitempty"`
}

// SpecPluginConfigDetail represents the plugin config detail in database.
type SpecPluginConfigDetail struct {
	Name         string `json:"name" bson:"name"`
	TemplateName string `json:"template_name" bson:"template_name"`
	Content      string `json:"content" bson:"content"`
	IsMainConfig bool   `json:"is_main_config" bson:"is_main_config"`
}

// Scope represents the scope of deploy policy.
type Scope struct {
	Type                 string                `json:"type" bson:"type"`
	ScopeServiceTemplate *ScopeServiceTemplate `json:"scope_service_template,omitempty" bson:"scope_service_template,omitempty"`
	ScopeSetTemplate     *ScopeSetTemplate     `json:"scope_set_template,omitempty" bson:"scope_set_template,omitempty"`
	ScopeInstance        *ScopeInstance        `json:"scope_instance,omitempty" bson:"scope_instance,omitempty"`
	ScopeTopo            *ScopeTopo            `json:"scope_topo,omitempty" bson:"scope_topo,omitempty"`
	ScopeDynamicGroup    *ScopeDynamicGroup    `json:"scope_dynamic_group,omitempty" bson:"scope_dynamic_group,omitempty"`
}

// ScopeServiceTemplate represents the scope of service template.
type ScopeServiceTemplate struct {
	Granularity        string        `json:"granularity" bson:"granularity"`
	BizID              int64         `json:"biz_id" bson:"biz_id"`
	Filter             *TargetFilter `json:"filter" bson:"filter"`
	ServiceTemplateIDs []int64       `json:"service_template_ids" bson:"service_template_ids"`
	ModuleIDs          []int64       `json:"module_ids" bson:"module_ids"`
}

// ScopeSetTemplate represents the scope of set template.
type ScopeSetTemplate struct {
	Granularity    string        `json:"granularity" bson:"granularity"`
	BizID          int64         `json:"biz_id" bson:"biz_id"`
	Filter         *TargetFilter `json:"filter" bson:"filter"`
	SetTemplateIDs []int64       `json:"set_template_ids" bson:"set_template_ids"`
	SetIDs         []int64       `json:"set_ids" bson:"set_ids"`
}

// ScopeInstance represents the scope of instance.
type ScopeInstance struct {
	Granularity string        `json:"granularity" bson:"granularity"`
	BizID       int64         `json:"biz_id" bson:"biz_id"`
	Filter      *TargetFilter `json:"filter" bson:"filter"`
	InstanceIDs []int64       `json:"instance_ids" bson:"instance_ids"`
}

// ScopeTopo represents the scope of topo.
type ScopeTopo struct {
	Granularity string           `json:"granularity" bson:"granularity"`
	BizID       int64            `json:"biz_id" bson:"biz_id"`
	Filter      *TargetFilter    `json:"filter" bson:"filter"`
	Paths       []*ScopeTopoNode `json:"paths" bson:"paths"`
}

// ScopeTopoNode represents the topo node in scope topo.
type ScopeTopoNode struct {
	TopoObjID  string `json:"topo_obj_id" bson:"topo_obj_id"`
	TopoInstID int64  `json:"topo_inst_id" bson:"topo_inst_id"`
}

// ScopeDynamicGroup represents the scope of dynamic group.
type ScopeDynamicGroup struct {
	Granularity     string        `json:"granularity" bson:"granularity"`
	BizID           int64         `json:"biz_id" bson:"biz_id"`
	Filter          *TargetFilter `json:"filter" bson:"filter"`
	DynamicGroupIDs []string      `json:"dynamic_group_ids" bson:"dynamic_group_ids"`
}

// LifeCycle represents the life cycle of deploy policy.
type LifeCycle struct {
	CreatedAt  time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" bson:"updated_at"`
	ExecutedAt time.Time `json:"executed_at" bson:"executed_at"`
}

// TargetFilter represents the target filter of deploy policy.
type TargetFilter struct {
}
