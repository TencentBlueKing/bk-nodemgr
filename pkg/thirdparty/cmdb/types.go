package cmdb

import (
	"fmt"
	"time"
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

	BKCloudName string `json:"bk_cloud_name"`
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

	BKCloudID   int64  `json:"bk_cloud_id"`
	BKCloudName string `json:"bk_cloud_name"`
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
	BKSetID   int64     `json:"bk_set_id"`
	BKSetName string    `json:"bk_set_name"`
	Module    []*Module `json:"module"`
}

// Module describe the module info define by cmdb.
type Module struct {
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
type FindModuleBatchResp []*Module
