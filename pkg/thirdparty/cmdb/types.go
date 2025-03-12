package cmdb

import (
	"fmt"
	"time"
)

const (
	// DefaultBusinessID describe the default business id.
	DefaultBusinessID = 0
)

// Page describe the page data in request.
type Page struct {
	Start int    `json:"start"`
	Limit int    `json:"limit"`
	Sort  string `json:"sort"`
}

// HostInfo describe the information of single host.
type HostInfo struct {
	// 主机ID
	BKHostID int64 `json:"bk_host_id"`
	// 机房ID
	IdcID int64 `json:"idc_id"`
	// 机房
	IdcName string `json:"idc_name"`
	// 机房管理单元ID
	IdcUnitID int64 `json:"idc_unit_id"`
	// 机房管理单元
	IdcUnitName string `json:"idc_unit_name"`
	// IDC城市ID
	IdcCityID string `json:"idc_city_id"`
	// IDC城市名
	IdcCityName string `json:"idc_city_name"`
	// 主机名称
	BKHostName string `json:"bk_host_name"`
	// 备注
	BKComment string `json:"bk_comment"`
	// 带外管理方式
	BKManageType string `json:"bk_manage_type"`
	// 内网交换机端口
	InnerSwitchPort string `json:"inner_switch_port"`
	// 寻址方式
	BKAddressing string `json:"bk_addressing"`
	// ServiceIdArr
	BKSvcIdArr string `json:"bk_svc_id_arr"`
	// CC逻辑区域ID
	BKLogicZone string `json:"bk_logic_zone"`
	// 网络结构版本
	NetStructName string `json:"net_struct_name"`
	// 固资编号
	BKAssetID string `json:"bk_asset_id"`
	// OuterEquipID
	BKOuterEquipID string `json:"bk_outer_equip_id"`
	// 标准设备类型ID
	BKSvrTypeID int64 `json:"bk_svr_type_id"`
	// GSE Agent ID
	BKAgentID string `json:"bk_agent_id"`
	// 设备编号
	SvrDeviceID string `json:"svr_device_id"`
	// 标准设备类型(腾讯机型代号)
	BKSvrDeviceClsName string `json:"bk_svr_device_cls_name"`
	// 启用时间
	SvrFirstTime time.Time `json:"svr_first_time"`
	// Module名称
	ModuleName string `json:"module_name"`
	// 存放机架
	Rack string `json:"rack"`
	// CC逻辑区域ID
	BKLogicZoneID string `json:"bk_logic_zone_id"`
	// 所属运营商
	BKIspName string `json:"bk_isp_name"`
	// 域名名称
	Domain string `json:"domain"`
	// 操作系统名称
	BKOsName string `json:"bk_os_name"`
	// 带外管理卡名称
	SrvOutBandManageType string `json:"srv_out_band_manage_type"`
	// 套餐计费起始时间
	BillingStartTime time.Time `json:"billing_start_time"`
	// 腾讯机型版本号
	BKStrVersion string `json:"bk_str_version"`
	// 硬件备注
	HardMemo string `json:"hard_memo"`
	// 区域ID
	BKIdcAreaID int64 `json:"bk_idc_area_id"`
	// 子ZoneID
	SubZoneID string `json:"sub_zone_id"`
	// 子Zone
	SubZone string `json:"sub_zone"`
	// 内网机房网络模块
	BKInnerNetIdc string `json:"bk_inner_net_idc"`
	// 服务器型号
	SvrTypeName string `json:"svr_type_name"`
	// 云主机实例ID
	BKCloudInstID string `json:"bk_cloud_inst_id"`
	// 交换机外网IP
	BKOuterSwitchIp string `json:"bk_outer_switch_ip"`
	// 管控区域
	BKCloudID int64 `json:"bk_cloud_id"`
	// 公司cmdbSvrID
	SvrID int64 `json:"svr_id"`
	// 交换机内网IP
	BKInnerSwitchIp string `json:"bk_inner_switch_ip"`
	// 重要级别
	SrvImportantLevel string `json:"srv_important_level"`
	// 入库时间
	SvrInputTime time.Time `json:"svr_input_time"`
	// 蓝鲸OWNER ID
	BKSupplierAccount string `json:"bk_supplier_account"`
	// 云可用区（Zone）
	BKCloudZone string `json:"bk_cloud_zone"`
	// 网络结构版本ID
	NetStructID int64 `json:"net_struct_id"`
	// 设备类型ID
	SvrDeviceTypeID int64 `json:"svr_device_type_id"`
	// 操作系统版本
	BKOsVersion string `json:"bk_os_version"`
	// 小组名称
	GroupName string `json:"group_name"`
	// 设备类型
	SvrDeviceTypeName string `json:"svr_device_type_name"`
	// 外网网段
	OuterNetworkSegment string `json:"outer_network_segment"`
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
	// RAID
	RaidName string `json:"raid_name"`
	// 运维部门
	DeptName string `json:"dept_name"`
	// 云子网ID
	BKCloudSubnetID string `json:"bk_cloud_subnet_id"`
	// 存放机架ID
	RackID string `json:"rack_id"`
	// RAID ID
	RaidID string `json:"raid_id"`
	// 逻辑区域ID
	LogicDomain string `json:"logic_domain"`
	// SLA级别
	BKSla string `json:"bk_sla"`
	// 云厂商
	BKCloudVendor string `json:"bk_cloud_vendor"`
	// 操作系统类型
	BKOsType string `json:"bk_os_type"`
	// 逻辑区域ID
	LogicDomainID string `json:"logic_domain_id"`
	// SCM设备类型
	SvrDeviceClass string `json:"svr_device_class"`
	// 网络设备ID
	NetDeviceID string `json:"net_device_id"`
	// 备份维护人
	BKBakOperator string `json:"bk_bak_operator"`
	// 实例计费模式
	InstanceChargeType string `json:"instance_charge_type"`
	// 云地域（Region）
	BKCloudRegion string `json:"bk_cloud_region"`
	// 区域
	BKIdcArea string `json:"bk_idc_area"`
	// 所属产品
	BKProduct string `json:"bk_product"`
	// 外网运营商
	BKIpOperName string `json:"bk_ip_oper_name"`
	// 服务器来源类型ID
	BKSvrSourceTypeID string `json:"bk_svr_source_type_id"`
	// 状态
	SrvStatus string `json:"srv_status"`
	// PositionName
	BKPositionName string `json:"bk_position_name"`
	// 云VPCID
	BKCloudVpcID string `json:"bk_cloud_vpc_id"`
	// ServiceArr
	BKServiceArr string `json:"bk_service_arr"`
	// 分级级别名称
	ClassifyLevelName string `json:"classify_level_name"`
	// 外网交换机端口
	OuterSwitchPort string `json:"outer_switch_port"`
	// 母机固资号
	BKSvrOwnerAssetID string `json:"bk_svr_owner_asset_id"`
	// 内网网段
	InnerNetworkSegment string `json:"inner_network_segment"`
	// 是否固资
	IsSpecial bool `json:"is_special"`
	// InnerEquipID
	BKInnerEquipID string `json:"bk_inner_equip_id"`
	// Zone名称
	BKZoneName string `json:"bk_zone_name"`
	// 设备SN
	BKSn string `json:"bk_sn"`
	// 套餐计费过期时间
	BillingExpireTime time.Time `json:"billing_expire_time"`
	// CPU逻辑核心数
	BKCpu *float64 `json:"bk_cpu"`
	// CPU型号
	BKCpuModule string `json:"bk_cpu_module"`
	// 操作系统位数
	BKOsBit string `json:"bk_os_bit"`
	// 内网Mac 地址
	BKMac string `json:"bk_mac"`
	// 内存容量
	BKMem *float64 `json:"bk_mem"`
	// CPU架构
	BKCpuArchitecture string `json:"bk_cpu_architecture"`
	// 外网MAC地址
	BKOuterMac string `json:"bk_outer_mac"`
	// IsVirtual
	BKIsVirtual bool `json:"bk_is_virtual"`
	// 录入方式
	ImportFrom string `json:"import_from"`
	// CLB_VIP
	ClbVip string `json:"clb_vip"`
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

type ObjectInfo struct {
	BKObjectID        string `json:"bk_obj_id"`
	BKObjectName      string `json:"bk_obj_name"`
	BKSupplierAccount string `json:"bk_supplier_account"`
}

// RespCommon describe the common part of response data.
type RespCommon struct {
	Result  bool   `json:"result"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// BaseBroker describe the base broker.
type BaseBroker[T any] struct {
	RespCommon
	Data T `json:"data"`
}

// CodeOK define the success code.
const CodeOK = 0

// IsFailed check the response is ok.
func (resp *BaseBroker[T]) IsFailed() error {
	if resp.Result != true || resp.Code != CodeOK {
		return fmt.Errorf("result(%v), code(%d) , msg(%s) ", resp.Result, resp.Code, resp.Message)
	}

	return nil
}

// ListBizHostsReq describe the request data of list_biz_hosts.
type ListBizHostsReq struct {
	// tenant id of this request.
	TenantID string `json:"-"`

	// response page settings.
	Page Page `json:"page"`

	// biz id of this request.
	BKBizID int64 `json:"bk_biz_id"`

	// expected response fields.
	Fields []string `json:"fields"`
}

// ListBizHostsResp describe the response data of list_biz_hosts.
type ListBizHostsResp struct {
	Count int         `json:"count"`
	Info  []*HostInfo `json:"info"`
}

// SearchBusinessReq describe the request data of search_business.
type SearchBusinessReq struct {
	// tenant id of this request.
	TenantID string `json:"-"`

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
	// tenant id of this request.
	TenantID string `json:"-"`

	Page Page `json:"page"`
}

// SearchCloudAreaResp describe the response data of search_cloud_area.
type SearchCloudAreaResp struct {
	Count int          `json:"count"`
	Info  []*CloudArea `json:"info"`
}

// CreateCloudAreaReq describe the request data of create_cloud_area.
type CreateCloudAreaReq struct {
	// tenant id of this request.
	TenantID string `json:"-"`

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
	// tenant id of this request.
	TenantID string `json:"-"`

	BKCloudID     int64  `json:"bk_cloud_id"`
	BKCloudName   string `json:"bk_cloud_name"`
	BKCloudVendor string `json:"bk_cloud_vendor"`
}

// UpdateCloudAreaResp describe the response data of update_cloud_area.
type UpdateCloudAreaResp string

// DeleteCloudAreaReq describe the request data of delete_cloud_area.
type DeleteCloudAreaReq struct {
	// tenant id of this request.
	TenantID string `json:"-"`

	BKCloudID int64 `json:"bk_cloud_id"`
}

// DeleteCloudAreaResp describe the response data of delete_cloud_area.
type DeleteCloudAreaResp string

// UpdateHostCloudAreaFieldReq describe the request data of update_host_cloud_area_field.
type UpdateHostCloudAreaFieldReq struct {
	// tenant id of this request.
	TenantID string `json:"-"`

	BKCloudID int64   `json:"bk_cloud_id"`
	BKBizID   int64   `json:"bk_biz_id"`
	BKHostIDs []int64 `json:"bk_host_ids"`
}

// UpdateHostCloudAreaFieldResp describe the response data of update_host_cloud_area_field.
type UpdateHostCloudAreaFieldResp string

// CloudArea cloud area info
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
	// tenant id of this request.
	TenantID string `json:"-"`

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
	// tenant id of this request.
	TenantID string `json:"-"`

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
	// tenant id of this request.
	TenantID string `json:"-"`

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
	// tenant id of this request.
	TenantID string `json:"-"`

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
	ID                  int64  `json:"id"`
	BKBizID             int64  `json:"bk_biz_id"`
	BKPropertyID        string `json:"bk_property_id"`
	BKPropertyName      string `json:"bk_property_name"`
	BKPropertyGroup     string `json:"bk_property_group"`
	BKPropertyGroupType string `json:"bk_property_type"`
	Creator             string `json:"creator"`
	Unit                string `json:"unit"`
	Placeholder         string `json:"placeholder"`
	Editable            bool   `json:"editable"`
	IsRequired          bool   `json:"isrequired"`
	IsReadOnly          bool   `json:"isreadonly"`
	IsOnly              bool   `json:"isonly"`
	IsPre               bool   `json:"ispre"`
	Option              any    `json:"option"`
	Description         string `json:"description"`
	BKSupplierAccount   string `json:"bk_supplier_account"`
	BKAsstObjID         string `json:"bk_asst_obj_id"`
	CreateTime          string `json:"create_time"`
	LastTime            string `json:"last_time"`
}

// SearchObjectAttributeReq describe the request data of search_object_attribute.
type SearchObjectAttributeReq struct {
	// tenant id of this request.
	TenantID string `json:"-"`

	BKBizID int64  `json:"bk_biz_id"`
	BKObjID string `json:"bk_obj_id"`
}

// SearchObjectAttributeResp describe the response data of search_object_attribute.
type SearchObjectAttributeResp []*ObjectAttributeInfo

// EnumOption describe the enum option info define by cmdb.
type EnumOption struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// FieldCondition describe the general field condition structure defined by cmdb.
type FieldCondition struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    any    `json:"value"`
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
	BKObjID string `json:"bk_obj_id"`
	Name    string `json:"name"`
	Info    struct {
		Condition         []*DynamicGroupCondition `json:"condition"`
		VariableCondition []*DynamicGroupCondition `json:"variable_condition"`
	} `json:"info"`
	CreateUser  string    `json:"create_user"`
	ModifyUseer string    `json:"modify_user"`
	LastTime    time.Time `json:"last_time"`
	CreateTime  time.Time `json:"create_time"`
}

// CreateDynamicGroupReq describe the request data of create_dynamic_group.
type CreateDynamicGroupReq struct {
	// tenant id of this request.
	TenantID string `json:"-"`

	DynamicGroupInfo `json:",inline"`
}

// CreateDynamicGroupResp describe the response data of create_dynamic_group.
type CreateDynamicGroupResp struct {
	ID string `json:"id"`
}

// ExecuteDynamicGroupReq describe the request data of execute_dynamic_group.
type ExecuteDynamicGroupReq struct {
	// tenant id of this request.
	TenantID string `json:"-"`

	BKBizID        int64    `json:"bk_biz_id"`
	ID             string   `json:"id"`
	Fields         []string `json:"fields"`
	DisableCounter bool     `json:"disable_counter"`
	Page           Page     `json:"page"`
}

// ExecuteHostDynamicGroupResp describe the response data of execute_dynamic_group when dynamic group type is host.
type ExecuteHostDynamicGroupResp struct {
	Count int         `json:"count"`
	Info  []*HostInfo `json:"info"`
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

// ExecuteSetDynamicGroupResp describe the response data of execute_dynamic_group when dynamic group type is set.
type ExecuteSetDynamicGroupResp struct {
	Count int        `json:"count"`
	Info  []*SetInfo `json:"info"`
}

// SearchDynamicGroupReq describe the request data of search_dynamic_group.
type SearchDynamicGroupReq struct {
	// tenant id of this request.
	TenantID string `json:"-"`

	BKBizID        int64 `json:"bk_biz_id"`
	DisableCounter bool  `json:"disable_counter"`
	Condition      struct {
		Name string `json:"name"`
	} `json:"condition"`
	Page Page `json:"page"`
}

// SearchDynamicGroupResp describe the response data of search_dynamic_group.
type SearchDynamicGroupResp struct {
	Count int                 `json:"count"`
	Info  []*DynamicGroupInfo `json:"info"`
}

// DeleteDynamicGroupReq describe the request data of delete_dynamic_group.
type DeleteDynamicGroupReq struct {
	// tenant id of this request.
	TenantID string `json:"-"`

	ID      string `json:"id"`
	BKBizID int64  `json:"bk_biz_id"`
}

// DeleteDynamicGroupResp describe the response data of delete_dynamic_group.
type DeleteDynamicGroupResp string

// GetDynamicGroupReq describe the request data of get_dynamic_group.
type GetDynamicGroupReq struct {
	// tenant id of this request.
	TenantID string `json:"-"`

	ID      string `json:"id"`
	BKBizID int64  `json:"bk_biz_id"`
}

// GetDynamicGroupResp describe the response data of get_dynamic_group.
type GetDynamicGroupResp struct {
	DynamicGroupInfo `json:",inline"`
}

// UpdateDynamicGroupReq describe the request data of update_dynamic_group.
type UpdateDynamicGroupReq struct {
	// tenant id of this request.
	TenantID string `json:"-"`

	DynamicGroupInfo `json:",inline"`
}

// UpdateDynamicGroupResp describe the response data of update_dynamic_group.
type UpdateDynamicGroupResp string

// ListHostsWithoutBusinessReq describe the request data of list_host_without_business.
type ListHostsWithoutBusinessReq struct {
	// tenant id of this request.
	TenantID string `json:"-"`

	Fields []string `json:"fields"`
	Page   Page     `json:"page"`
}

// ListHostsWithoutBusinessResp describe the response data of list_host_without_business.
type ListHostsWithoutBusinessResp struct {
	Count int         `json:"count"`
	Info  []*HostInfo `json:"info"`
}

// ListServiceTemplateReq describe the request data of list_service_template.
type ListServiceTemplateReq struct {
	// tenant id of this request.
	TenantID string `json:"-"`

	BKBizID             int64   `json:"bk_biz_id"`
	ServiceCategoryID   int64   `json:"service_category_id"`
	ServiceTemplateName string  `json:"search"`
	ServiceTemplateIDs  []int64 `json:"service_template_ids"`
	IsExact             bool    `json:"is_exact"`
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
	// tenant id of this request.
	TenantID string `json:"-"`

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
	// tenant id of this request.
	TenantID string `json:"-"`

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
	// tenant id of this request.
	TenantID string `json:"-"`

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
	// tenant id of this request.
	TenantID string `json:"-"`

	BKBizID int64    `json:"bk_biz_id"`
	BKIDs   []int64  `json:"bk_ids"`
	Fields  []string `json:"fields"`
}

// FindSetBatchResp describe the response data of find_set_batch.
type FindSetBatchResp []*SetInfo

// SearchSetReq describe the request data of search_set.
type SearchSetReq struct {
	// tenant id of this request.
	TenantID string `json:"-"`

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
	// tenant id of this request.
	TenantID string `json:"-"`

	BKBizID     int64   `json:"bk_biz_id"`
	BKSetIDs    []int64 `json:"bk_set_ids"`
	BKModuleIDs []int64 `json:"bk_module_ids"`
	BKHostIDs   []int64 `json:"bk_host_ids"`
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
	// tenant id of this request.
	TenantID string `json:"-"`

	BKBizID  int64   `json:"bk_biz_id"`
	BKHostID []int64 `json:"bk_host_id"`
}

// FindHostBizRelationsResp describe the response data of find_host_biz_relations.
type FindHostBizRelationsResp []*HostTopoRelation

// FindHostByServiceTemplateReq describe the request data of find_host_by_service_template.
type FindHostByServiceTemplateReq struct {
	// tenant id of this request.
	TenantID string `json:"-"`

	BKBizID              int64    `json:"bk_biz_id"`
	BKServiceTemplateIDs []int64  `json:"bk_service_template_ids"`
	BKModuleIDs          []int64  `json:"bk_module_ids"`
	Fields               []string `json:"fields"`
	Page                 Page     `json:"page"`
}

// FindHostByServiceTemplateResp describe the response data of find_host_by_service_template.
type FindHostByServiceTemplateResp struct {
	Count int         `json:"count"`
	Info  []*HostInfo `json:"info"`
}

// FindHostBySetTemplateReq describe the request data of find_host_by_set_template.
type FindHostBySetTemplateReq struct {
	// tenant id of this request.
	TenantID string `json:"-"`

	BKBizID          int64    `json:"bk_biz_id"`
	BKSetTemplateIDs []int64  `json:"bk_set_template_ids"`
	BKSetIDs         []int64  `json:"bk_set_ids"`
	Fields           []string `json:"fields"`
	Page             Page     `json:"page"`
}

// FindHostBySetTemplateResp describe the response data of find_host_by_set_template.
type FindHostBySetTemplateResp struct {
	Count int         `json:"count"`
	Info  []*HostInfo `json:"info"`
}

// FindHostByTopoReq describe the request data of find_host_by_topo.
type FindHostByTopoReq struct {
	// tenant id of this request.
	TenantID string `json:"-"`

	BKBizID  int64    `json:"bk_biz_id"`
	BKObjID  string   `json:"bk_obj_id"`
	BKInstID int64    `json:"bk_inst_id"`
	Fields   []string `json:"fields"`
	Page     Page     `json:"page"`
}

// FindHostByTopoResp describe the response data of find_host_by_topo.
type FindHostByTopoResp struct {
	Count int         `json:"count"`
	Info  []*HostInfo `json:"info"`
}

// FindHostRelationsWithTopoReq describe the request data of find_host_relations_with_topo.
type FindHostRelationsWithTopoReq struct {
	// tenant id of this request.
	TenantID string `json:"-"`

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

// ListServiceInstanceDetailReq describe the request data of list_service_instance_detail .
type ListServiceInstanceDetailReq struct {
	// tenant id of this request.
	TenantID string `json:"-"`

	BKBizID int64 `json:"bk_biz_id"`
	Page    Page  `json:"page"`
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
type GetMainlineObjectTopoReq struct {
	// tenant id of this request.
	TenantID string `json:"-"`
}

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

	BKBizID int64    `json:"bk_biz_id"`
	Fields  []string `json:"fields"`
	Page    Page     `json:"page"`
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
	// tenant id of this request.
	TenantID string `json:"-"`

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
	// tenant id of this request.
	TenantID string `json:"-"`

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
	// tenant id of this request.
	TenantID string `json:"-"`

	BKBizID        int64   `json:"bk_biz_id"`
	SetTemplateIDs []int64 `json:"set_template_ids"`
	Page           Page    `json:"page"`
}

// ListSetTemplateResp describe the response data of list_set_template.
type ListSetTemplateResp struct {
	Count int                    `json:"count"`
	Info  []*ServiceInstanceInfo `json:"info"`
}

// UpdateHostProperties describe the host properties to be updated.
type UpdateHostProperties struct {
	Properties struct {
		BKHostName string `json:"bk_host_name,omitempty"`
		Operator   string `json:"operator,omitempty"`
		BKComment  string `json:"bk_comment,omitempty"`
		BKIspName  string `json:"bk_isp_name,omitempty"`
	} `json:"properties"`
	BKHostID int64 `json:"bk_host_id"`
}

// BatchUpdateHostReq describe the request data of batch_update_host.
type BatchUpdateHostReq struct {
	// tenant id of this request.
	TenantID string `json:"-"`

	Update []*UpdateHostProperties `json:"update"`
}

// BatchUpdateHostResp describe the response data of batch_update_host.
type BatchUpdateHostResp string

// HostAgentIDInfo describe the host agent id info define by cmdb.
type HostAgentIDInfo struct {
	BKHostID  int64  `json:"bk_host_id"`
	BKAgentID string `json:"bk_agent_id"`
}

// BindHostAgentReq describe the request data of bind_host_agent.
type BindHostAgentReq struct {
	// tenant id of this request.
	TenantID string `json:"-"`

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
	BKAddressing      string `json:"bk_addressing"`
}

// AddHostToBusinessIdleReq describe the request data of add_host_to_business_idle.
type AddHostToBusinessIdleReq struct {
	// tenant id of this request.
	TenantID string `json:"-"`

	BKBizID    int64             `json:"bk_biz_id"`
	BKHostList []*CreateHostInfo `json:"bk_host_list"`
}

// AddHostToBusinessIdleResp describe the response data of add_host_to_business_idle.
type AddHostToBusinessIdleResp struct {
	BKHostIDs []int64 `json:"bk_host_ids"`
}

// PushHostIdentifierReq describe the request data of push_host_identifier.
type PushHostIdentifierReq struct {
	// tenant id of this request.
	TenantID string `json:"-"`

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

// ResourceWatchReq describe the request data of resource_watch.
type ResourceWatchReq struct {
	// tenant id of this request.
	TenantID string `json:"-"`

	BKEventTypes []string `json:"bk_event_types"`
	BKResource   string   `json:"bk_resource"`
	BKFields     []string `json:"bk_fields"`
	BKStartFrom  int64    `json:"bk_start_from"`
	BKCursor     string   `json:"bk_cursor"`
}

// HostEvent describe the event info define by cmdb.
type HostEvent struct {
	BKCursor    string    `json:"bk_cursor"`
	BKResource  string    `json:"bk_resource"`
	BKEventType string    `json:"bk_event_type"`
	BKDetail    *HostInfo `json:"bk_detail"`
}

// HostRelationEvent describe the event info define by cmdb.
type HostRelationEvent struct {
	BKCursor    string            `json:"bk_cursor"`
	BKResource  string            `json:"bk_resource"`
	BKEventType string            `json:"bk_event_type"`
	BKDetail    *HostTopoRelation `json:"bk_detail"`
}

// ProcessEvent describe the event info define by cmdb.
type ProcessEvent struct {
	BKCursor    string           `json:"bk_cursor"`
	BKResource  string           `json:"bk_resource"`
	BKEventType string           `json:"bk_event_type"`
	BKDetail    *ProcessProperty `json:"bk_detail"`
}

// ResourceWatchResp describe the response data of resource_watch.
type ResourceWatchResp[T any] struct {
	BKWatched bool `json:"bk_watched"`
	BKEvents  []*T `json:"bk_events"`
}
