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

package cmdb

import (
	"fmt"
	"sync"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
)

const (
	// CCNoBusinessID describe the default business id.
	CCNoBusinessID = 0

	// CCResourcePoolBusinessID describe the resource pool business id.
	CCResourcePoolBusinessID = 1

	// CCPageSizeLimit describe the max page size.
	CCPageSizeLimit = 500

	// CCHostIDBatchSize describe the batch size when use host id as param.
	CCHostIDBatchSize = 500
)

type ccField string

// nolint: gochecknoglobals
var ccHostFieldsInstance = struct {
	fields []ccField
	once   sync.Once
}{}

// ccHostFields describe the default search host fields.
func ccHostFields() []ccField {
	ccHostFieldsInstance.once.Do(func() {
		filedMAp, _ := conv.StructToMap(HostInfo{})
		fields := conv.MapKeyToSlice(filedMAp)
		for _, field := range fields {
			ccHostFieldsInstance.fields = append(ccHostFieldsInstance.fields, ccField(field))
		}
	})

	return ccHostFieldsInstance.fields
}

const (
	ccFieldBKInnerIP          ccField = "bk_host_innerip"
	ccFieldBKHostID           ccField = "bk_host_id"
	ccFieldBKHostName         ccField = "bk_host_name"
	ccFieldBKCloudID          ccField = "bk_cloud_id"
	ccFieldBKAddressing       ccField = "bk_addressing"
	ccFieldBKSetID            ccField = "bk_set_id"
	ccFieldSetTemplateID      ccField = "set_template_id"
	ccFieldBKSetName          ccField = "bk_set_name"
	ccFieldBKBizID            ccField = "bk_biz_id"
	ccFieldOpsConsoleHostID   ccField = "ops_console_host_id"
	ccFieldOpsOutBandType     ccField = "ops_out_band_type"
	ccFieldOpsOutBandProtocol ccField = "ops_out_band_protocol"
	ccFieldOpsBMCIP           ccField = "ops_bmc_ip"
	ccFieldOpsBMCPort         ccField = "ops_bmc_port"
)

const (
	// CCInvalidID represents an invalid CMDB ID (typically 0).
	CCInvalidID int64 = 0
)

// Page describe the page data in request.
type Page struct {
	Start int    `json:"start"`
	Limit int    `json:"limit"`
	Sort  string `json:"sort,omitempty"`
}

// HostInfo describe the information of single host.
type HostInfo struct {
	// 主机ID
	BKHostID int64 `json:"bk_host_id"`
	// IDC城市ID
	IdcCityID string `json:"idc_city_id"`
	// 主机名称
	BKHostName string `json:"bk_host_name"`
	// 寻址方式
	BKAddressing string `json:"bk_addressing"`
	// GSE Agent ID
	BKAgentID string `json:"bk_agent_id"`
	// 管控区域
	BKCloudID int64 `json:"bk_cloud_id"`
	// 主要维护人
	Operator string `json:"operator"`
	// 内网IPv4
	BKHostInnerIPV4 string `json:"bk_host_innerip"`
	// 外网IPV4
	BKHostOuterIPV4 string `json:"bk_host_outerip"`
	// 内网IPv6
	BKHostInnerIPV6 string `json:"bk_host_innerip_v6"`
	// 外网IPV6
	BKHostOuterIPV6 string `json:"bk_host_outerip_v6"`
	// 运维部门
	DeptName string `json:"dept_name"`
	// 操作系统类型
	BKOSType string `json:"bk_os_type"`
	// IDC区域ID
	BKIDCAreaID int64 `json:"bk_idc_area_id"`
	// CPU逻辑核心数
	BKCpu float64 `json:"bk_cpu"`
	// 内网Mac 地址
	BKMac string `json:"bk_mac"`
	// 内存容量
	BKMem float64 `json:"bk_mem"`
	// CPU架构
	BKCpuArchitecture string `json:"bk_cpu_architecture"`
	// 主控机Host-ID
	OpsConsoleHostID int64 `json:"ops_console_host_id"`
	// 带外设备类型
	OpsOutBandType string `json:"ops_out_band_type"`
	// 带外设备协议
	OpsOutBandProtocol string `json:"ops_out_band_protocol"`
	// 带外管理IP
	OpsBMCIP string `json:"ops_bmc_ip"`
	// 带外管理端口
	OpsBMCPort int64 `json:"ops_bmc_port"`
}

// BusinessInfo describe the information of single business.
type BusinessInfo struct {
	BKBizID           int64  `json:"bk_biz_id"`
	BKBizName         string `json:"bk_biz_name"`
	BKBizMaintainer   string `json:"bk_biz_maintainer"`
	BKBizProducer     string `json:"bk_biz_producer"`
	BKBizDeveloper    string `json:"bk_biz_developer"`
	BKBizTester       string `json:"bk_biz_tester"`
	TimeZone          string `json:"time_zone"`
	Language          string `json:"language"`
	BKSupplierAccount string `json:"bk_supplier_account"`
	CreateTime        string `json:"create_time"`
	LastTime          string `json:"last_time"`

	// default field describes business type.
	Default     int    `json:"default"`
	Operator    string `json:"operator"`
	LifeCycle   string `json:"life_cycle"`
	BKCreatedAt string `json:"bk_created_at"`
	BKUpdatedAt string `json:"bk_updated_at"`
	BKCreatedBy string `json:"bk_created_by"`
}

// ObjectInfo describe the information of single object.
type ObjectInfo struct {
	BKObjectID        string `json:"bk_obj_id"`
	BKObjectName      string `json:"bk_obj_name"`
	BKSupplierAccount string `json:"bk_supplier_account"`
}

// RespCommon describe the common part of response data.
type RespCommon struct {
	Result     bool        `json:"result"`
	Code       int         `json:"code"`
	Message    string      `json:"message"`
	Permission *Permission `json:"permission"`
}

// Permission describe the permission of user.
type Permission struct {
	SystemID   string              `json:"system_id"`
	SystemName string              `json:"system_name"`
	Actions    []PermissionActions `json:"actions"`
}

// PermissionActions describe the actions of permission.
type PermissionActions struct {
	ID                   string                                  `json:"id"`
	Name                 string                                  `json:"name"`
	RelatedResourceTypes []PermissionActionsRelatedResourceTypes `json:"related_resource_types"`
}

// PermissionActionsRelatedResourceTypes describe the related resource types of permission actions.
type PermissionActionsRelatedResourceTypes struct {
	SystemID   string                                            `json:"system_id"`
	SystemName string                                            `json:"system_name"`
	Type       string                                            `json:"type"`
	TypeName   string                                            `json:"type_name"`
	Instances  [][]PermissionActionsRelatedResourceTypesInstance `json:"instances"`
}

// PermissionActionsRelatedResourceTypesInstance describe the instance of permission actions.
type PermissionActionsRelatedResourceTypesInstance struct {
	Type     string `json:"type"`
	TypeName string `json:"type_name"`
	ID       string `json:"id"`
	Name     string `json:"name"`
}

// BaseBroker describe the base broker.
type BaseBroker[T any] struct {
	RespCommon
	Data T `json:"data"`
}

const (
	// codeOK define the success code.
	codeOK = 0

	// codeNoPermission define the no permission code.
	// notice: !!! not all api will return correct response code.
	codeNoPermission = 9900403
)

// IsFailed check the response is ok.
func (resp *BaseBroker[T]) IsFailed() error {
	switch {
	case resp.Result && resp.Code == codeOK:
		return nil
	case (!resp.Result && resp.Code == codeNoPermission) || resp.Permission != nil:
		return fmt.Errorf("no permission, please check your permission, permission(%v)", resp.Permission)
	default:
		return fmt.Errorf("result(%v), code(%d) , msg(%s)", resp.Result, resp.Code, resp.Message)
	}
}

type filterCondition string

const (
	// hostPropertyFilterConditionAnd describe the and condition.
	hostPropertyFilterConditionAnd filterCondition = "AND"

	// hostPropertyFilterConditionOr describe the or condition.
	// nolint: unused
	hostPropertyFilterConditionOr filterCondition = "OR"
)

// HostPropertyFilter describe the host property filter.
type HostPropertyFilter struct {
	Condition filterCondition   `json:"condition"`
	Rules     []*FieldCondition `json:"rules"`
}

// ListBizHostsReq describe the request data of list_biz_hosts.
type ListBizHostsReq struct {
	// response page settings.
	Page Page `json:"page"`

	// biz id of this request.
	BKBizID int64 `json:"bk_biz_id"`

	// expected response fields.
	Fields []ccField `json:"fields"`

	// host property filter.
	HostPropertyFilter *HostPropertyFilter `json:"host_property_filter,omitempty"`
}

// ListBizHostsResp describe the response data of list_biz_hosts.
type ListBizHostsResp struct {
	Count int         `json:"count"`
	Info  []*HostInfo `json:"info"`
}

// SearchBusinessReq describe the request data of search_business.
type SearchBusinessReq struct {
	// Page ...
	Page Page `json:"page"`

	// Fields ...
	Fields []string `json:"fields"`
}

// SearchBusinessResp describe the response data of search_business.
type SearchBusinessResp struct {
	Count int             `json:"count"`
	Info  []*BusinessInfo `json:"info"`
}

// SearchCloudAreaReq describe the request data of search_cloud_area.
type SearchCloudAreaReq struct {
	Page Page `json:"page"`
}

// SearchCloudAreaResp describe the response data of search_cloud_area.
type SearchCloudAreaResp struct {
	Count int          `json:"count"`
	Info  []*CloudArea `json:"info"`
}

// CreateCloudAreaReq describe the request data of create_cloud_area.
type CreateCloudAreaReq struct {
	BKCloudName   string `json:"bk_cloud_name"`
	BKCloudVendor string `json:"bk_cloud_vendor"`
}

// CreateCloudAreaResp describe the response data of create_cloud_area.
type CreateCloudAreaResp struct {
	Created struct {
		ID          int64 `json:"id"`
		OriginIndex int64 `json:"origin_index"`
	} `json:"created"`
}

// UpdateCloudAreaReq describe the request data of update_cloud_area.
type UpdateCloudAreaReq struct {
	BKCloudID     int64  `json:"bk_cloud_id"`
	BKCloudName   string `json:"bk_cloud_name"`
	BKCloudVendor string `json:"bk_cloud_vendor"`
}

// UpdateCloudAreaResp describe the response data of update_cloud_area.
type UpdateCloudAreaResp string

// DeleteCloudAreaReq describe the request data of delete_cloud_area.
type DeleteCloudAreaReq struct {
	BKCloudID int64 `json:"bk_cloud_id"`
}

// DeleteCloudAreaResp describe the response data of delete_cloud_area.
type DeleteCloudAreaResp string

// UpdateHostCloudAreaFieldReq describe the request data of update_host_cloud_area_field.
type UpdateHostCloudAreaFieldReq struct {
	BKCloudID int64   `json:"bk_cloud_id"`
	BKBizID   int64   `json:"bk_biz_id"`
	BKHostIDs []int64 `json:"bk_host_ids"`
}

// UpdateHostCloudAreaFieldResp describe the response data of update_host_cloud_area_field.
type UpdateHostCloudAreaFieldResp string

// CloudArea cloud area info.
type CloudArea struct {
	BKCloudID         int64     `json:"bk_cloud_id"`
	BKCloudName       string    `json:"bk_cloud_name"`
	BKCloudVendor     string    `json:"bk_cloud_vendor"`
	BKSupplierAccount string    `json:"bk_supplier_account"`
	CreateTime        time.Time `json:"create_time"`
	LastTime          time.Time `json:"last_time"`
}

// SearchBizInstTopoReq describe the request data of search_biz_inst_topo.
type SearchBizInstTopoReq struct {
	// biz id of this request.
	BKBizID int64 `json:"bk_biz_id"`
}

// SearchBizInstTopoResp describe the response data of search_biz_inst_topo.
type SearchBizInstTopoResp []*BizInstTopo

// BizInstTopo describes the business instance topology structure defined by CMDB.
type BizInstTopo struct {
	Default    int            `json:"default"`
	BKInstID   int64          `json:"bk_inst_id"`
	BKInstName string         `json:"bk_inst_name"`
	BKObjID    string         `json:"bk_obj_id"`
	BKObjName  string         `json:"bk_obj_name"`
	Children   []*BizInstTopo `json:"child"`
}

// GetBizInternalModuleReq describe the request data of get_biz_internal_module.
type GetBizInternalModuleReq struct {
	// biz id of this request.
	BKBizID int64 `json:"bk_biz_id"`
}

// GetBizInternalModuleResp describe the response data of get_biz_internal_module.
type GetBizInternalModuleResp struct {
	BKSetID   int64         `json:"bk_set_id"`
	BKSetName string        `json:"bk_set_name"`
	Module    []*ModuleInfo `json:"module"`
}

// ModuleInfo describe the module info define by cmdb.
type ModuleInfo struct {
	BKModuleID        int64     `json:"bk_module_id"`
	BKModuleName      string    `json:"bk_module_name"`
	Default           int64     `json:"default"`
	CreateTime        time.Time `json:"create_time"`
	BKSetID           int64     `json:"bk_set_id"`
	BKBakOperator     string    `json:"bk_bak_operator"`
	BKBizID           int64     `json:"bk_biz_id"`
	BKModuleType      string    `json:"bk_module_type"`
	BKParentID        int64     `json:"bk_parent_id"`
	BKSupplierAccount string    `json:"bk_supplier_account"`
	LastTime          time.Time `json:"last_time"`
	HostApplyEnabled  bool      `json:"host_apply_enabled"`
	Operator          string    `json:"operator"`
	ServiceCategoryID int64     `json:"service_category_id"`
	ServiceTemplateID int64     `json:"service_template_id"`
	SetTemplateID     int64     `json:"set_template_id"`
	BKCreatedAt       time.Time `json:"bk_created_at"`
	BKUpdatedAt       time.Time `json:"bk_updated_at"`
	BKCreatedBy       string    `json:"bk_created_by"`
}

// FindTopoNodePathsReq describe the request data of find_topo_node_paths.
type FindTopoNodePathsReq struct {
	// biz id of this request.
	BKBizID int64 `json:"bk_biz_id"`

	// nodes of this request, list of business topo instance node information to be queried
	BKNodes []*Node `json:"bk_nodes"`
}

// Node describe the node info define by cmdb.
type Node struct {
	BKObjID    string `json:"bk_obj_id"`
	BKInstID   int64  `json:"bk_inst_id"`
	BKInstName string `json:"bk_inst_name"`
}

// FindTopoNodePathsResp describe the response data of find_topo_node_paths.
type FindTopoNodePathsResp []*NodePaths

// NodePaths describe the node paths info define by cmdb.
type NodePaths struct {
	Node    `json:",inline"`
	BKPaths [][]*Node `json:"bk_paths"`
}

// FindModuleBatchReq describe the request data of find_module_batch.
type FindModuleBatchReq struct {
	// biz id of this request.
	BKBizID int64 `json:"bk_biz_id"`

	// ids of this request, module id list
	BKIDs []int64 `json:"bk_ids"`

	// fields of this request, module attribute list
	Fields []string `json:"fields"`
}

// FindModuleBatchResp describe the response data of find_module_batch.
type FindModuleBatchResp []*ModuleInfo

// ObjectAttributeInfo describe the object attribute info define by cmdb.
type ObjectAttributeInfo struct {
	ID                int64  `json:"id"`
	BKBizID           int64  `json:"bk_biz_id"`
	BKObjID           string `json:"bk_obj_id"`
	BKPropertyID      string `json:"bk_property_id"`
	BKPropertyName    string `json:"bk_property_name"`
	BKPropertyType    string `json:"bk_property_type"`
	BKPropertyGroup   string `json:"bk_property_group"`
	Creator           string `json:"creator"`
	Unit              string `json:"unit"`
	Placeholder       string `json:"placeholder"`
	Editable          bool   `json:"editable"`
	IsRequired        bool   `json:"isrequired"`
	IsReadOnly        bool   `json:"isreadonly"`
	IsOnly            bool   `json:"isonly"`
	IsPre             bool   `json:"ispre"`
	Option            any    `json:"option"`
	Description       string `json:"description"`
	BKSupplierAccount string `json:"bk_supplier_account"`
	BKAsstObjID       string `json:"bk_asst_obj_id"`
	CreateTime        string `json:"create_time"`
	LastTime          string `json:"last_time"`
}

// SearchObjectReq describe the request data of search_object.
type SearchObjectReq struct {
	BKObjID string `json:"bk_obj_id"`
}

// SearchObjectResp describe the response data of search_object.
type SearchObjectResp []*ObjectInfo

// SearchObjectAttributeReq describe the request data of search_object_attribute.
type SearchObjectAttributeReq struct {
	BKBizID int64  `json:"bk_biz_id"`
	BKObjID string `json:"bk_obj_id"`
}

// SearchObjectAttributeResp describe the response data of search_object_attribute.
type SearchObjectAttributeResp []*ObjectAttributeInfo

// CreateBizCustomFieldReq describe the request data of create_biz_custom_field.
type CreateBizCustomFieldReq struct {
	BKObjID         string `json:"bk_obj_id"`
	BKBizID         int64  `json:"bk_biz_id"`
	BKPropertyID    string `json:"bk_property_id"`
	BKPropertyName  string `json:"bk_property_name"`
	BKPropertyType  string `json:"bk_property_type"`
	BKPropertyGroup string `json:"bk_property_group,omitempty"`
	Creator         string `json:"creator,omitempty"`
	Unit            string `json:"unit,omitempty"`
	Placeholder     string `json:"placeholder,omitempty"`
	Editable        bool   `json:"editable,omitempty"`
	IsRequired      bool   `json:"isrequired,omitempty"`
	IsReadOnly      bool   `json:"isreadonly,omitempty"`
	IsOnly          bool   `json:"isonly,omitempty"`
	IsPre           bool   `json:"ispre,omitempty"`
	Option          any    `json:"option,omitempty"`
	Default         any    `json:"default,omitempty"`
	Description     string `json:"description,omitempty"`
	BKAsstObjID     string `json:"bk_asst_obj_id,omitempty"`
}

// CreateBizCustomFieldResp describe the response data of create_biz_custom_field.
type CreateBizCustomFieldResp *ObjectAttributeInfo

// EnumOption describe the enum option info define by cmdb.
type EnumOption struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// NumberOption describe the number option info define by cmdb.
type NumberOption struct {
	Min string `json:"min"`
	Max string `json:"max"`
}

// HostAgentIDInfo describe the host agent id info define by cmdb.
type HostAgentIDInfo struct {
	BKHostID  int64  `json:"bk_host_id"`
	BKAgentID string `json:"bk_agent_id"`
}

// BindHostAgentReq describe the request data of bind_host_agent.
type BindHostAgentReq struct {
	List []*HostAgentIDInfo `json:"list"`
}

// BindHostAgentResp describe the response data of bind_host_agent.
type BindHostAgentResp string

// UnbindHostAgentReq describe the request data of unbind_host_agent.
type UnbindHostAgentReq BindHostAgentReq

// UnbindHostAgentResp describe the response data of unbind_host_agent.
type UnbindHostAgentResp string

// CreateHostInfo describe the host info to be created.
type CreateHostInfo struct {
	BKCloudID         int64  `json:"bk_cloud_id"`
	BKHostInnerIP     string `json:"bk_host_innerip"`
	BKHostInnerIPV6   string `json:"bk_host_innerip_v6"`
	BKHostOuterIP     string `json:"bk_host_outerip"`
	BKHostOuterIPV6   string `json:"bk_host_outerip_v6"`
	BKOSType          string `json:"bk_os_type"`
	BKCpuArchitecture string `json:"bk_cpu_architecture"`
	BKAddressing      string `json:"bk_addressing,omitempty"`
}

// AddHostToBusinessIdleReq describe the request data of add_host_to_business_idle.
type AddHostToBusinessIdleReq struct {
	BKBizID    int64             `json:"bk_biz_id"`
	BKHostList []*CreateHostInfo `json:"bk_host_list"`
}

// AddHostToBusinessIdleResp describe the response data of add_host_to_business_idle.
type AddHostToBusinessIdleResp struct {
	BKHostIDs []int64 `json:"bk_host_ids"`
}

// PushHostIdentifierReq describe the request data of push_host_identifier.
type PushHostIdentifierReq struct {
	BKHostIDs []int64 `json:"bk_host_ids"`
}

// PushHostIdentifierResp describe the response data of push_host_identifier.
type PushHostIdentifierResp struct {
	TaskID    string `json:"task_id"`
	HostInfos []struct {
		BKHostID       int64  `json:"bk_host_id"`
		Identification string `json:"identification"`
	} `json:"host_infos"`
}

// FindHostIdentifierPushResultReq describe the request data of find_host_identifier_push_result.
type FindHostIdentifierPushResultReq struct {
	TaskID string `json:"task_id"`
}

// FindHostIdentifierPushResultResp describe the response data of find_host_identifier_push_result.
type FindHostIdentifierPushResultResp struct {
	SuccessList []int64 `json:"success_list"`
	FailedList  []int64 `json:"failed_list"`
	PendingList []int64 `json:"pending_list"`
}

// AddHostToResourcePoolReq describe the request data of add_host_to_resource_pool.
type AddHostToResourcePoolReq struct {
	HostInfo []*CreateHostInfo `json:"host_info"`
}

// AddHostToResourcePoolResp describe the response data of add_host_to_resource_pool.
type AddHostToResourcePoolResp struct {
	Success []struct {
		Index    int64 `json:"index"`
		BKHostID int64 `json:"bk_host_id"`
	} `json:"success"`
	Error []struct {
		Index        int64  `json:"index"`
		ErrorMessage string `json:"error_message"`
	} `json:"error"`
}

// ListResourcePoolHostsReq describe the request data of list_resource_pool_hosts.
type ListResourcePoolHostsReq struct {
	Fields []ccField `json:"fields"`
	Page   Page      `json:"page"`
}

// ListResourcePoolHostsResp describe the response data of list_resource_pool_hosts.
type ListResourcePoolHostsResp struct {
	Count int         `json:"count"`
	Info  []*HostInfo `json:"info"`
}

type ruleOperator string

const (
	ruleOperatorEqual          ruleOperator = "euqal"
	ruleOperatorNotEqual       ruleOperator = "not_equal"
	ruleOperatorIn             ruleOperator = "in"
	ruleOperatorNotIn          ruleOperator = "not_in"
	ruleOperatorLess           ruleOperator = "less"
	ruleOperatorLessOREqual    ruleOperator = "less_or_equal"
	ruleOperatorGreater        ruleOperator = "greater"
	ruleOperatorGreaterOREqual ruleOperator = "greater_or_equal"
	ruleOperatorBetween        ruleOperator = "between"
	ruleOperatorNotBetween     ruleOperator = "not_between"
)

// Validate validate the rule operator.
func (operator ruleOperator) Validate() error {
	switch operator {
	case ruleOperatorEqual, ruleOperatorNotEqual, ruleOperatorIn, ruleOperatorNotIn, ruleOperatorLess, ruleOperatorLessOREqual,
		ruleOperatorGreater, ruleOperatorGreaterOREqual, ruleOperatorBetween, ruleOperatorNotBetween:
		return nil
	default:
		return fmt.Errorf("unsupported rule operator: %s", operator)
	}
}

// FieldCondition describe the general field condition structure defined by cmdb.
type FieldCondition struct {
	Field    ccField      `json:"field"`
	Operator ruleOperator `json:"operator"`
	Value    any          `json:"value"`
}

// DynamicGroupCondition describe the dynamic group condition define by cmdb.
type DynamicGroupCondition struct {
	BKObjID   string            `json:"bk_obj_id"`
	Condition []*FieldCondition `json:"condition"`
}

// DynamicGroupInfo describe the dynamic group info define by cmdb.
type DynamicGroupInfo struct {
	ID      string `json:"id"`
	BKBizID int64  `json:"bk_biz_id"`
	BKObjID string `json:"bk_obj_id,omitempty"`
	Name    string `json:"name,omitempty"`
	Info    struct {
		Condition         []*DynamicGroupCondition `json:"condition"`
		VariableCondition []*DynamicGroupCondition `json:"variable_condition"`
	} `json:"info,omitempty"`
}

// CreateDynamicGroupReq describe the request data of create_dynamic_group.
type CreateDynamicGroupReq struct {
	DynamicGroupInfo `json:",inline"`
}

// CreateDynamicGroupResp describe the response data of create_dynamic_group.
type CreateDynamicGroupResp struct {
	ID string `json:"id"`
}

// ExecuteDynamicGroupReq describe the request data of execute_dynamic_group.
type ExecuteDynamicGroupReq struct {
	BKBizID        int64     `json:"bk_biz_id"`
	ID             string    `json:"id"`
	Fields         []ccField `json:"fields"`
	DisableCounter bool      `json:"disable_counter"`
	Page           Page      `json:"page"`
}

// ExecuteDynamicGroupResp describe the response data of execute_dynamic_groupwhen dynamic group type is host.
type ExecuteDynamicGroupResp struct {
	Count int              `json:"count"`
	Info  []map[string]any `json:"info"`
}

// SetInfo describe the set info define by cmdb.
type SetInfo struct {
	BKSetID            int64     `json:"bk_set_id"`
	BKSetName          string    `json:"bk_set_name"`
	BKSetDesc          string    `json:"bk_set_desc"`
	BKSetEnv           string    `json:"bk_set_env"`
	BKBizID            int64     `json:"bk_biz_id"`
	BKCapacity         int64     `json:"bk_capacity"`
	BKParentID         int64     `json:"bk_parent_id"`
	Description        string    `json:"description"`
	SetTemplateID      int64     `json:"set_template_id"`
	SetTemplateVersion int64     `json:"set_template_version"`
	BKServiceStatus    string    `json:"bk_service_status"`
	BKSupplierAccount  string    `json:"bk_supplier_account"`
	Default            int64     `json:"default"`
	CreateTime         time.Time `json:"create_time"`
	LastTime           time.Time `json:"last_time"`
	BKCreateAt         time.Time `json:"bk_create_at"`
	BKUpdatedAt        time.Time `json:"bk_updated_at"`
}

// SearchDynamicGroupReq describe the request data of search_dynamic_group.
type SearchDynamicGroupReq struct {
	BKBizID        int64 `json:"bk_biz_id"`
	DisableCounter bool  `json:"disable_counter,omitempty"`
	Condition      struct {
		Name string `json:"name,omitempty"`
	} `json:"condition,omitempty"`
	Page Page `json:"page"`
}

// SearchDynamicGroupResp describe the response data of search_dynamic_group.
type SearchDynamicGroupResp struct {
	Count int                 `json:"count"`
	Info  []*DynamicGroupInfo `json:"info"`
}

// DeleteDynamicGroupReq describe the request data of delete_dynamic_group.
type DeleteDynamicGroupReq struct {
	ID      string `json:"id"`
	BKBizID int64  `json:"bk_biz_id"`
}

// DeleteDynamicGroupResp describe the response data of delete_dynamic_group.
type DeleteDynamicGroupResp string

// GetDynamicGroupReq describe the request data of get_dynamic_group.
type GetDynamicGroupReq struct {
	ID      string `json:"id"`
	BKBizID int64  `json:"bk_biz_id"`
}

// GetDynamicGroupResp describe the response data of get_dynamic_group.
type GetDynamicGroupResp struct {
	DynamicGroupInfo `json:",inline"`
	CreateUser       string    `json:"create_user"`
	ModifyUseer      string    `json:"modify_user"`
	LastTime         time.Time `json:"last_time"`
	CreateTime       time.Time `json:"create_time"`
}

// UpdateDynamicGroupReq describe the request data of update_dynamic_group.
type UpdateDynamicGroupReq struct {
	DynamicGroupInfo `json:",inline"`
}

// UpdateDynamicGroupResp describe the response data of update_dynamic_group.
type UpdateDynamicGroupResp string

// ListHostsWithoutBusinessReq describe the request data of list_host_without_business.
type ListHostsWithoutBusinessReq struct {
	Fields             []ccField           `json:"fields"`
	Page               Page                `json:"page"`
	HostPropertyFilter *HostPropertyFilter `json:"host_property_filter"`
}

// ListHostsWithoutBusinessResp describe the response data of list_host_without_business.
type ListHostsWithoutBusinessResp struct {
	Count int         `json:"count"`
	Info  []*HostInfo `json:"info"`
}

// ListServiceTemplateReq describe the request data of list_service_template.
type ListServiceTemplateReq struct {
	BKBizID             int64   `json:"bk_biz_id"`
	ServiceCategoryID   int64   `json:"service_category_id,omitempty"`
	ServiceTemplateName string  `json:"search,omitempty"`
	ServiceTemplateIDs  []int64 `json:"service_template_ids,omitempty"`
	IsExact             bool    `json:"is_exact,omitempty"`
	Page                Page    `json:"page"`
}

// ServiceTemplateInfo describe the service template info define by cmdb.
type ServiceTemplateInfo struct {
	BKBizID             int64     `json:"bk_biz_id"`
	ID                  int64     `json:"id"`
	ServiceTemplateName string    `json:"name"`
	ServiceCategoryID   int64     `json:"service_category_id"`
	Creator             string    `json:"creator"`
	Modifier            string    `json:"modifier"`
	CreateTime          time.Time `json:"create_time"`
	LastTime            time.Time `json:"last_time"`
	BKSupplierAccount   string    `json:"bk_supplier_account"`
	HostApplyEnabled    bool      `json:"host_apply_enabled"`
}

// ListServiceTemplateResp describe the response data of list_service_template.
type ListServiceTemplateResp struct {
	Count int                    `json:"count"`
	Info  []*ServiceTemplateInfo `json:"info"`
}

// KeyCondition describe the key condition define by cmdb.
type KeyCondition struct {
	Key      string `json:"key"`
	Values   any    `json:"values"`
	Operator string `json:"operator"`
}

// ListServiceInstanceReq describe the request data of list_service_instance.
type ListServiceInstanceReq struct {
	BKBizID                  int64           `json:"bk_biz_id"`
	BKModuleID               int64           `json:"bk_module_id"`
	BKHostIDs                []int64         `json:"bk_host_ids"`
	ServiceInstanceFuzzyName string          `json:"search_key"`
	Selectors                []*KeyCondition `json:"selectors"`
	Page                     Page            `json:"page"`
}

// ServiceInstanceInfo describe the service instance info define by cmdb.
type ServiceInstanceInfo struct {
	BKBizID             int64     `json:"bk_biz_id"`
	ID                  int64     `json:"id"`
	ServiceInstanceName string    `json:"name"`
	BKHostID            int64     `json:"bk_host_id"`
	BKModuleID          int64     `json:"bk_module_id"`
	Creator             string    `json:"creator"`
	Modifier            string    `json:"modifier"`
	CreateTime          time.Time `json:"create_time"`
	LastTime            time.Time `json:"last_time"`
	BKSupplierAccount   string    `json:"bk_supplier_account"`
}

// ListServiceInstanceResp describe the response data of list_service_instance.
type ListServiceInstanceResp struct {
	Count int                    `json:"count"`
	Info  []*ServiceInstanceInfo `json:"info"`
}

// ListProcessInstanceReq describe the request data of list_process_instance.
type ListProcessInstanceReq struct {
	BKBizID           int64 `json:"bk_biz_id"`
	ServiceInstanceID int64 `json:"service_instance_id"`
}

// ProcessProperty describe the process property define by cmdb.
type ProcessProperty struct {
	AutoStart         bool      `json:"auto_start"`
	BKBizID           int64     `json:"bk_biz_id"`
	BKFuncName        string    `json:"bk_func_name"`
	BKProcessID       int64     `json:"bk_process_id"`
	BKProcessName     string    `json:"bk_process_name"`
	BKStartParamRegex string    `json:"bk_start_param_regex"`
	BKSupplierAccount string    `json:"bk_supplier_account"`
	CreateTime        time.Time `json:"create_time"`
	LastTime          time.Time `json:"last_time"`
	Description       string    `json:"description"`
	FaceStopCMD       string    `json:"face_stop_cmd"`
	PidFile           string    `json:"pid_file"`
	Priority          int64     `json:"priority"`
	ProcNum           int64     `json:"proc_num"`
	ReloadCMD         string    `json:"reload_cmd"`
	RestartCMD        string    `json:"restart_cmd"`
	StartCMD          string    `json:"start_cmd"`
	StopCMD           string    `json:"stop_cmd"`
	Timeout           int64     `json:"timeout"`
	User              string    `json:"user"`
	WorkPath          string    `json:"work_path"`
	BKCreateAt        string    `json:"bk_created_at"`
	BKCreateBy        string    `json:"bk_created_by"`
	BKUpdateAt        string    `json:"bk_updated_at"`
	BKUpdateBy        string    `json:"bk_updated_by"`
	BindInfo          []struct {
		Enable        bool   `json:"enable"`
		IP            string `json:"ip"`
		Port          string `json:"port"`
		Protocol      string `json:"protocol"`
		TemplateRowID int64  `json:"template_row_id"`
	} `json:"bind_info"`
}

// ProcessInstanceInfo describe the process instance info define by cmdb.
type ProcessInstanceInfo struct {
	Property ProcessProperty `json:"property"`
	Relation struct {
		BKBizID           int64  `json:"bk_biz_id"`
		BKProcessID       int64  `json:"bk_process_id"`
		ServiceInstanceID int64  `json:"service_instance_id"`
		ProcessTemplateID int64  `json:"process_template_id"`
		BKHostID          int64  `json:"bk_host_id"`
		BKSupplierAccount string `json:"bk_supplier_account"`
	} `json:"relation"`
}

// ListProcessInstanceResp describe the response data of list_process_instance.
type ListProcessInstanceResp []*ProcessInstanceInfo

// ListProcTemplateReq describe the request data of list_proc_template.
type ListProcTemplateReq struct {
	BKBizID            int64   `json:"bk_biz_id"`
	ServiceTemplateID  int64   `json:"service_template_id"`
	ProcessTemplateIDs []int64 `json:"process_template_ids"`
}

// ProcessTemplateInfo describe the process template info define by cmdb.
type ProcessTemplateInfo struct {
	ID                    int64     `json:"id"`
	BKProcessTemplateName string    `json:"bk_process_name"`
	BKBizID               int64     `json:"bk_biz_id"`
	ServiceTemplateID     int64     `json:"service_template_id"`
	Creator               string    `json:"creator"`
	Modifier              string    `json:"modifier"`
	CreateTime            time.Time `json:"create_time"`
	LastTime              time.Time `json:"last_time"`
	BKSupplierAccount     string    `json:"bk_supplier_account"`
	Property              struct {
		AutoStart struct {
			Value          bool `json:"value"`
			AsDefaultValue bool `json:"as_default_value"`
		} `json:"auto_start"`
		BKBizID struct {
			Value          int64 `json:"value"`
			AsDefaultValue bool  `json:"as_default_value"`
		} `json:"bk_biz_id"`
		BKFuncName struct {
			Value          string `json:"value"`
			AsDefaultValue bool   `json:"as_default_value"`
		} `json:"bk_func_name"`
		BKProcessID struct {
			Value          int64 `json:"value"`
			AsDefaultValue bool  `json:"as_default_value"`
		} `json:"bk_process_id"`
		BKProcessName struct {
			Value          string `json:"value"`
			AsDefaultValue bool   `json:"as_default_value"`
		} `json:"bk_process_name"`
		BKStartParamRegex struct {
			Value          string `json:"value"`
			AsDefaultValue bool   `json:"as_default_value"`
		} `json:"bk_start_param_regex"`
		Description struct {
			Value          string `json:"value"`
			AsDefaultValue bool   `json:"as_default_value"`
		} `json:"description"`
		FaceStopCMD struct {
			Value          string `json:"value"`
			AsDefaultValue bool   `json:"as_default_value"`
		} `json:"face_stop_cmd"`
		PidFile struct {
			Value          string `json:"value"`
			AsDefaultValue bool   `json:"as_default_value"`
		} `json:"pid_file"`
		Priority struct {
			Value          int64 `json:"value"`
			AsDefaultValue bool  `json:"as_default_value"`
		} `json:"priority"`
		ProcNum struct {
			Value          int64 `json:"value"`
			AsDefaultValue bool  `json:"as_default_value"`
		} `json:"proc_num"`
		ReloadCMD struct {
			Value          string `json:"value"`
			AsDefaultValue bool   `json:"as_default_value"`
		} `json:"reload_cmd"`
		RestartCMD struct {
			Value          string `json:"value"`
			AsDefaultValue bool   `json:"as_default_value"`
		} `json:"restart_cmd"`
		StartCMD struct {
			Value          string `json:"value"`
			AsDefaultValue bool   `json:"as_default_value"`
		} `json:"start_cmd"`
		StopCMD struct {
			Value          string `json:"value"`
			AsDefaultValue bool   `json:"as_default_value"`
		} `json:"stop_cmd"`
		Timeout struct {
			Value          int64 `json:"value"`
			AsDefaultValue bool  `json:"as_default_value"`
		} `json:"timeout"`
		User struct {
			Value          string `json:"value"`
			AsDefaultValue bool   `json:"as_default_value"`
		} `json:"user"`
		WorkPath struct {
			Value          string `json:"value"`
			AsDefaultValue bool   `json:"as_default_value"`
		} `json:"work_path"`
		BindInfo struct {
			Value []struct {
				Enable struct {
					Value          bool `json:"value"`
					AsDefaultValue bool `json:"as_default_value"`
				} `json:"enable"`
				IP struct {
					Value          string `json:"value"`
					AsDefaultValue bool   `json:"as_default_value"`
				} `json:"ip"`
				Port struct {
					Value          string `json:"value"`
					AsDefaultValue bool   `json:"as_default_value"`
				} `json:"port"`
				Protocol struct {
					Value          string `json:"value"`
					AsDefaultValue bool   `json:"as_default_value"`
				} `json:"protocol"`
				TemplateRowID int64 `json:"row_id"`
			}
			AsDefaultValue bool `json:"as_default_value"`
		}
	}
}

// ListProcTemplateResp describe the response data of list_proc_template.
type ListProcTemplateResp struct {
	Count int                    `json:"count"`
	Info  []*ProcessTemplateInfo `json:"info"`
}

// FindSetBatchReq describe the request data of find_set_batch.
type FindSetBatchReq struct {
	BKBizID int64    `json:"bk_biz_id"`
	BKIDs   []int64  `json:"bk_ids"`
	Fields  []string `json:"fields"`
}

// FindSetBatchResp describe the response data of find_set_batch.
type FindSetBatchResp []*SetInfo

// SearchSetReq describe the request data of search_set.
type SearchSetReq struct {
	BKBizID int64    `json:"bk_biz_id"`
	Fields  []string `json:"fields"`
	Page    Page     `json:"page"`
}

// SearchSetResp describe the response data of search_set.
type SearchSetResp struct {
	Count int        `json:"count"`
	Info  []*SetInfo `json:"info"`
}

// SearchModuleReq describe the request data of search_module.
type SearchModuleReq struct {
	// tenant id of this request.
	TenantID string   `json:"-"`
	BKBizID  int64    `json:"bk_biz_id"`
	BKSetID  int64    `json:"bk_set_id"`
	Fields   []string `json:"fields"`
	Page     Page     `json:"page"`
}

// SearchModuleResp describe the response data of search_module.
type SearchModuleResp struct {
	Count int           `json:"count"`
	Info  []*ModuleInfo `json:"info"`
}

// FindHostTopoRelationReq describe the request data of find_host_topo_relation.
type FindHostTopoRelationReq struct {
	BKBizID     int64   `json:"bk_biz_id"`
	BKSetIDs    []int64 `json:"bk_set_ids,omitempty"`
	BKModuleIDs []int64 `json:"bk_module_ids,omitempty"`
	BKHostIDs   []int64 `json:"bk_host_ids,omitempty"`
	Page        Page    `json:"page"`
}

// HostTopoRelation describe the host topo relation define by cmdb.
type HostTopoRelation struct {
	BKBizID           int64  `json:"bk_biz_id"`
	BKHostID          int64  `json:"bk_host_id"`
	BKModuleID        int64  `json:"bk_module_id"`
	BKSetID           int64  `json:"bk_set_id"`
	BKSupplierAccount string `json:"bk_supplier_account"`
}

// FindHostTopoRelationResp describe the response data of find_host_topo_relation.
type FindHostTopoRelationResp struct {
	Count int                 `json:"count"`
	Data  []*HostTopoRelation `json:"data"`
	Page  Page                `json:"page"`
}

// FindHostBizRelationsReq describe the request data of find_host_biz_relations.
type FindHostBizRelationsReq struct {
	BKBizID  int64   `json:"bk_biz_id,omitempty"`
	BKHostID []int64 `json:"bk_host_id"`
}

// FindHostBizRelationsResp describe the response data of find_host_biz_relations.
type FindHostBizRelationsResp = []*HostTopoRelation

// FindHostByServiceTemplateReq describe the request data of find_host_by_service_template.
type FindHostByServiceTemplateReq struct {
	BKBizID              int64     `json:"bk_biz_id"`
	BKServiceTemplateIDs []int64   `json:"bk_service_template_ids"`
	BKModuleIDs          []int64   `json:"bk_module_ids"`
	Fields               []ccField `json:"fields"`
	Page                 Page      `json:"page"`
}

// FindHostByServiceTemplateResp describe the response data of find_host_by_service_template.
type FindHostByServiceTemplateResp struct {
	Count int         `json:"count"`
	Info  []*HostInfo `json:"info"`
}

// FindHostBySetTemplateReq describe the request data of find_host_by_set_template.
type FindHostBySetTemplateReq struct {
	BKBizID          int64     `json:"bk_biz_id"`
	BKSetTemplateIDs []int64   `json:"bk_set_template_ids"`
	BKSetIDs         []int64   `json:"bk_set_ids,omitempty"`
	Fields           []ccField `json:"fields"`
	Page             Page      `json:"page"`
}

// FindHostBySetTemplateResp describe the response data of find_host_by_set_template.
type FindHostBySetTemplateResp struct {
	Count int         `json:"count"`
	Info  []*HostInfo `json:"info"`
}

// FindHostByTopoReq describe the request data of find_host_by_topo.
type FindHostByTopoReq struct {
	BKBizID  int64     `json:"bk_biz_id"`
	BKObjID  string    `json:"bk_obj_id"`
	BKInstID int64     `json:"bk_inst_id"`
	Fields   []ccField `json:"fields"`
	Page     Page      `json:"page"`
}

// FindHostByTopoResp describe the response data of find_host_by_topo.
type FindHostByTopoResp struct {
	Count int         `json:"count"`
	Info  []*HostInfo `json:"info"`
}

// FindHostRelationsWithTopoReq describe the request data of find_host_relations_with_topo.
type FindHostRelationsWithTopoReq struct {
	BKBizID   int64    `json:"bk_biz_id"`
	BKObjID   string   `json:"bk_obj_id"`
	BKInstIDs []int64  `json:"bk_inst_ids"`
	Fields    []string `json:"fields"`
	Page      Page     `json:"page"`
}

// FindHostRelationsWithTopoResp describe the response data of find_host_relations_with_topo.
type FindHostRelationsWithTopoResp struct {
	Count int                 `json:"count"`
	Info  []*HostTopoRelation `json:"info"`
}

// ListServiceInstanceDetailReq describe the request data of list_service_instance_detail.
// Reference: https://github.com/TencentBlueKing/bk-cmdb/blob/master/docs/apidoc/apigw/open/zh/list_service_instance_detail.md
type ListServiceInstanceDetailReq struct {
	BKBizID              int64           `json:"bk_biz_id"`
	BKModuleID           int64           `json:"bk_module_id,omitempty"`
	BKHostList           []int64         `json:"bk_host_list,omitempty"`
	BKServiceTemplateIDs []int64         `json:"bk_service_template_ids,omitempty"`
	BKSetTemplateIDs     []int64         `json:"bk_set_template_ids,omitempty"`
	ServiceInstanceIDs   []int64         `json:"service_instance_ids,omitempty"`
	Selectors            []*KeyCondition `json:"selectors,omitempty"`
	Page                 Page            `json:"page"`
}

// ServiceInstanceDetailInfo describe the service instance detail info define by cmdb.
type ServiceInstanceDetailInfo struct {
	ID                int64     `json:"id"`
	Name              string    `json:"name"`
	ServiceTemplateID int64     `json:"service_template_id"`
	BKBizID           int64     `json:"bk_biz_id"`
	BKHostID          int64     `json:"bk_host_id"`
	BKModuleID        int64     `json:"bk_module_id"`
	Creator           string    `json:"creator"`
	Modifier          string    `json:"modifier"`
	CreateTime        time.Time `json:"create_time"`
	LastTime          time.Time `json:"last_time"`
	BKSupplierAccount string    `json:"bk_supplier_account"`
	ServiceCategoryID int64     `json:"service_category_id"`
	ProcessInstances  []struct {
		Process struct {
			AutoStart         bool      `json:"auto_start"`
			BKBizID           int64     `json:"bk_biz_id"`
			BKFuncName        string    `json:"bk_func_name"`
			BKProcessID       int64     `json:"bk_process_id"`
			BKProcessName     string    `json:"bk_process_name"`
			BKStartParamRegex string    `json:"bk_start_param_regex"`
			BKSupplierAccount string    `json:"bk_supplier_account"`
			CreateTime        time.Time `json:"create_time"`
			LastTime          time.Time `json:"last_time"`
			Description       string    `json:"description"`
			FaceStopCMD       string    `json:"face_stop_cmd"`
			PidFile           string    `json:"pid_file"`
			Priority          int64     `json:"priority"`
			ProcNum           int64     `json:"proc_num"`
			ReloadCMD         string    `json:"reload_cmd"`
			RestartCMD        string    `json:"restart_cmd"`
			StartCMD          string    `json:"start_cmd"`
			StopCMD           string    `json:"stop_cmd"`
			Timeout           int64     `json:"timeout"`
			User              string    `json:"user"`
			WorkPath          string    `json:"work_path"`
			BKCreateAt        string    `json:"bk_created_at"`
			BKCreateBy        string    `json:"bk_created_by"`
			BKUpdateAt        string    `json:"bk_updated_at"`
			BKUpdateBy        string    `json:"bk_updated_by"`
			BindInfo          []struct {
				Enable        bool   `json:"enable"`
				IP            string `json:"ip"`
				Port          string `json:"port"`
				Protocol      string `json:"protocol"`
				TemplateRowID int64  `json:"template_row_id"`
			} `json:"bind_info"`
		} `json:"process"`
		Relation struct {
			BKBizID           int64  `json:"bk_biz_id"`
			BKProcessID       int64  `json:"bk_process_id"`
			ServiceInstanceID int64  `json:"service_instance_id"`
			ProcessTemplateID int64  `json:"process_template_id"`
			BKHostID          int64  `json:"bk_host_id"`
			BKSupplierAccount string `json:"bk_supplier_account"`
		}
	} `json:"process_instances"`
}

// ListServiceInstanceDetailResp describe the response data of list_service_instance_detail.
type ListServiceInstanceDetailResp struct {
	Count int                          `json:"count"`
	Info  []*ServiceInstanceDetailInfo `json:"info"`
}

// GetMainlineObjectTopoReq describe the request data of get_mainline_object_topo.
type GetMainlineObjectTopoReq struct{}

// MainlineObjectTopo describe the mainline object topo define by cmdb.
type MainlineObjectTopo struct {
	BKObjID           string `json:"bk_obj_id"`
	BKObjName         string `json:"bk_obj_name"`
	BKSupplierAccount string `json:"bk_supplier_account"`
	BKNextObj         string `json:"bk_next_obj"`
	BKNextName        string `json:"bk_next_name"`
	BKPreObjID        string `json:"bk_pre_obj_id"`
	BKPreObjName      string `json:"bk_pre_obj_name"`
}

// GetMainlineObjectTopoResp describe the response data of get_mainline_object_topo.
type GetMainlineObjectTopoResp []*MainlineObjectTopo

// ListBizHostsTopoReq describe the request data of list_biz_hosts_topo.
type ListBizHostsTopoReq struct {
	//	tenant id of this request.
	TenantID string `json:"-"`

	BKBizID int64     `json:"bk_biz_id"`
	Fields  []ccField `json:"fields"`
	Page    Page      `json:"page"`
}

// HostTopo describe the topo node define by cmdb.
type HostTopo struct {
	BKSetID   int64  `json:"bk_set_id"`
	BKSetName string `json:"bk_set_name"`
	Module    []struct {
		BKModuleID   int64  `json:"bk_module_id"`
		BKModuleName string `json:"bk_module_name"`
	} `json:"module"`
}

// ListBizHostsTopoResp describe the response data of list_biz_hosts_topo.
type ListBizHostsTopoResp struct {
	Count int `json:"count"`
	Info  []struct {
		Host *HostInfo   `json:"host"`
		Topo []*HostTopo `json:"topo"`
	} `json:"info"`
}

// ListServiceInstanceByHostReq describe the request data of list_service_instance_by_host.
type ListServiceInstanceByHostReq struct {
	BKBizID  int64 `json:"bk_biz_id"`
	BKHostID int64 `json:"bk_host_id"`
	Page     Page  `json:"page"`
}

// ListServiceInstanceByHostResp describe the response data of list_service_instance_by_host.
type ListServiceInstanceByHostResp struct {
	Count int                    `json:"count"`
	Info  []*ServiceInstanceInfo `json:"info"`
}

// ListServiceInstanceBySetTemplateReq describe the request data of list_service_instance_by_set_template.
type ListServiceInstanceBySetTemplateReq struct {
	BKBizID       int64 `json:"bk_biz_id"`
	SetTemplateID int64 `json:"set_template_id"`
	Page          Page  `json:"page"`
}

// ListServiceInstanceBySetTemplateResp describe the response data of list_service_instance_by_set_template.
type ListServiceInstanceBySetTemplateResp struct {
	Count int                    `json:"count"`
	Info  []*ServiceInstanceInfo `json:"info"`
}

// ListSetTemplateReq describe the request data of list_set_template.
type ListSetTemplateReq struct {
	BKBizID        int64   `json:"bk_biz_id"`
	SetTemplateIDs []int64 `json:"set_template_ids,omitempty"`
	Page           Page    `json:"page"`
}

// ListSetTemplateResp describe the response data of list_set_template.
type ListSetTemplateResp struct {
	Count int                    `json:"count"`
	Info  []*ServiceInstanceInfo `json:"info"`
}

// UpdateHostProperties describe the host properties to be updated.
type UpdateHostProperties struct {
	// Properties describe the mutable host properties.
	// bk_cloud_id is intentionally excluded because it must not be modified here.
	Properties struct {
		// BKHostName matches HostInfo.BKHostName.
		BKHostName *string `json:"bk_host_name,omitempty"`
		// Operator matches HostInfo.Operator.
		Operator *string `json:"operator,omitempty"`
		// BKComment is the host comment field.
		BKComment *string `json:"bk_comment,omitempty"`
		// BKIspName is the ISP name field.
		BKIspName *string `json:"bk_isp_name,omitempty"`
		// OpsConsoleHostID matches HostInfo.OpsConsoleHostID.
		OpsConsoleHostID *int64 `json:"ops_console_host_id,omitempty"`
		// OpsOutBandType matches HostInfo.OpsOutBandType.
		OpsOutBandType *string `json:"ops_out_band_type,omitempty"`
		// OpsOutBandProtocol matches HostInfo.OpsOutBandProtocol.
		OpsOutBandProtocol *string `json:"ops_out_band_protocol,omitempty"`
		// OpsBMCIP matches HostInfo.OpsBMCIP.
		OpsBMCIP *string `json:"ops_bmc_ip,omitempty"`
		// OpsBMCPort matches HostInfo.OpsBMCPort.
		OpsBMCPort *int64 `json:"ops_bmc_port,omitempty"`
	} `json:"properties"`
	BKHostID int64 `json:"bk_host_id"`
}

// BatchUpdateHostReq describe the request data of batch_update_host.
type BatchUpdateHostReq struct {
	Update []*UpdateHostProperties `json:"update"`
}

// BatchUpdateHostResp describe the response data of batch_update_host.
type BatchUpdateHostResp string

// FindHostServiceTemplateReq describe the request data of find_host_service_template.
type FindHostServiceTemplateReq struct {
	BKHostID []int64 `json:"bk_host_id"`
}

// HostServiceTemplate describe the host service template define by cmdb.
type HostServiceTemplate struct {
	BKHostID          int64   `json:"bk_host_id"`
	ServiceTemplateID []int64 `json:"service_template_id"`
}

// FindHostServiceTemplateResp describe the response data of find_host_service_template.
type FindHostServiceTemplateResp []*HostServiceTemplate

// ResourceWatchReq describe the request data of resource_watch.
type ResourceWatchReq struct {
	BKResource   string    `json:"bk_resource"`
	BKEventTypes []string  `json:"bk_event_types,omitempty"`
	BKFields     []ccField `json:"bk_fields,omitempty"`
	BKStartFrom  int64     `json:"bk_start_from,omitempty"`
	BKCursor     string    `json:"bk_cursor,omitempty"`
}

// ResourceWatchResp describe the response data of resource_watch.
type ResourceWatchResp struct {
	BKWatched bool              `json:"bk_watched"`
	BKEvents  []*map[string]any `json:"bk_events"`
}

// EventInfo describe the event info define by cmdb.
type EventInfo[T any] struct {
	BKCursor    string `json:"bk_cursor,omitempty"`
	BKResource  string `json:"bk_resource"`
	BKEventType string `json:"bk_event_type,omitempty"`
	BKDetail    T      `json:"bk_detail"`
}

// HostEventInfo describe the host event info define by cmdb.
type HostEventInfo = EventInfo[*HostInfo]

// HostRelationEventInfo describe the host relation event info define by cmdb.
type HostRelationEventInfo = EventInfo[*HostTopoRelation]

// GetBizBriefCacheTopoReq describe the request data of get_biz_brief_cache_topo.
type GetBizBriefCacheTopoReq struct {
	BKBizID int64 `json:"bk_biz_id"`
}

// GetBizBriefCacheTopoResp describe the response data of get_biz_brief_cache_topo.
type GetBizBriefCacheTopoResp struct {
	Biz   bizBriefCacheTopoBiz     `json:"biz"`
	Idle  []*bizBriefCacheTopoNode `json:"idle"`
	Nodes []*bizBriefCacheTopoNode `json:"nds"`
}

// bizBriefCacheTopoBiz describes the business information in the brief topology response.
type bizBriefCacheTopoBiz struct {
	ID                int64  `json:"id"`
	Name              string `json:"nm"`
	Default           int    `json:"dft"`
	BKSupplierAccount string `json:"bk_supplier_account"`
}

// bizBriefCacheTopoNode describes a node in the brief topology response.
type bizBriefCacheTopoNode struct {
	Obj     string                   `json:"obj"`
	ID      int64                    `json:"id"`
	Name    string                   `json:"nm"`
	Default int                      `json:"dft"`
	Nodes   []*bizBriefCacheTopoNode `json:"nds"`
}

const (
	// TopoNodeObjIDBiz topo node object id for biz.
	TopoNodeObjIDBiz = "biz"
	// TopoNodeObjIDHost topo node object id for host.
	TopoNodeObjIDHost = "host"
	// TopoNodeObjIDSet topo node object id for set.
	TopoNodeObjIDSet = "set"
	// TopoNodeObjIDModule topo node object id for module.
	TopoNodeObjIDModule = "module"
)

// IsMainlineObject checks if the given object ID is a mainline object in the CMDB topology.
func IsMainlineObject(objID string) bool {
	switch objID {
	case TopoNodeObjIDBiz, TopoNodeObjIDSet, TopoNodeObjIDModule:
		return true
	default:
		return false
	}
}
