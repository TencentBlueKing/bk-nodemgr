/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package host

const (
	// FieldKeyHostID the host id field key.
	FieldKeyHostID = "data.host_id"

	// FieldKeyTenantID the tenant id field key.
	FieldKeyTenantID = "data.tenant_id"

	// Static fields.

	// FieldKeyStatic the static field key.
	FieldKeyStatic = "data.static"

	// FieldKeyStaticBizID the static biz id field key.
	FieldKeyStaticBizID = "data.static.biz_id"

	// FieldKeyStaticSetID the static set id field key.
	FieldKeyStaticSetID = "data.static.set_id"

	// FieldKeyStaticModuleID the static module id field key.
	FieldKeyStaticModuleID = "data.static.module_id"

	// FieldKeyStaticTopo the static topo field key.
	FieldKeyStaticTopo = "data.static.topo"

	// FieldKeyStaticTopoSetID the static topo set id field key.
	FieldKeyStaticTopoSetID = "data.static.topo.bk_set_id"

	// FieldKeyStaticTopoModuleID the static topo module id field key.
	FieldKeyStaticTopoModuleID = "data.static.topo.bk_module_id"

	// FieldKeyStaticNetworkAreaID the static networkarea id field key.
	FieldKeyStaticNetworkAreaID = "data.static.networkarea_id"

	// FieldKeyStaticRegionID the static region id field key.
	FieldKeyStaticRegionID = "data.static.region_id"

	// FieldKeyStaticCityID the static city id field key.
	FieldKeyStaticCityID = "data.static.city_id"

	// FieldKeyStaticInnerIPList the static inner ip list field key.
	FieldKeyStaticInnerIPList = "data.static.inner_ip_list"

	// FieldKeyStaticInnerIPV6List the static inner ipv6 list field key.
	FieldKeyStaticInnerIPV6List = "data.static.inner_ipv6_list"

	// FieldKeyStaticOuterIPList the static outer ip list field key.
	FieldKeyStaticOuterIPList = "data.static.outer_ip_list"

	// FieldKeyStaticOuterIPV6List the static outer ipv6 list field key.
	FieldKeyStaticOuterIPV6List = "data.static.outer_ipv6_list"

	// FieldKeyStaticAddressing the static addressing field key.
	FieldKeyStaticAddressing = "data.static.addressing"

	// FieldKeyStaticDeptName the static dept name field key.
	FieldKeyStaticDeptName = "data.static.dept_name"

	// FieldKeyStaticOperator the static operator field key.
	FieldKeyStaticOperator = "data.static.operator"

	// FieldKeyStaticMac the static mac field key.
	FieldKeyStaticMac = "data.static.mac"

	// FieldKeyStaticOSTypeCCID the static os type cc id field key.
	FieldKeyStaticOSTypeCCID = "data.static.os_type_ccid"

	// FieldKeyStaticOSType the static os type field key.
	FieldKeyStaticOSType = "data.static.os_type"

	// FieldKeyStaticArch the static arch field key.
	FieldKeyStaticArch = "data.static.arch"

	// FieldKeyStaticHostName the static hostname field key.
	FieldKeyStaticHostName = "data.static.host_name"

	// FieldKeyStaticCPUNum the static cpu num field key.
	FieldKeyStaticCPUNum = "data.static.cpu_num"

	// FieldKeyStaticMemCap the static mem cap field key.
	FieldKeyStaticMemCap = "data.static.mem_cap"

	// FieldKeyStaticSyncedAgentID the static synced agent id field key.
	FieldKeyStaticSyncedAgentID = "data.static.synced_agent_id"

	// FieldKeyStaticSyncedOpsConsoleHostID the static synced ops console host id field key.
	FieldKeyStaticSyncedOpsConsoleHostID = "data.static.synced_ops_console_host_id"

	// FieldKeyStaticSyncedOpsOutBandType the static synced ops out-band type field key.
	FieldKeyStaticSyncedOpsOutBandType = "data.static.synced_ops_out_band_type"

	// FieldKeyStaticSyncedOpsOutBandProtocol the static synced ops out-band protocol field key.
	FieldKeyStaticSyncedOpsOutBandProtocol = "data.static.synced_ops_out_band_protocol"

	// FieldKeyStaticSyncedOpsBMCIP the static synced ops bmc ip field key.
	FieldKeyStaticSyncedOpsBMCIP = "data.static.synced_ops_bmc_ip"

	// FieldKeyStaticSyncedOpsBMCPort the static synced ops bmc port field key.
	FieldKeyStaticSyncedOpsBMCPort = "data.static.synced_ops_bmc_port"

	// Dynamic fields.

	// FieldKeyDynamic the dynamic field key.
	FieldKeyDynamic = "data.dynamic"

	// FieldKeyDynamicNodeRole the dynamic node role field key.
	FieldKeyDynamicNodeRole = "data.dynamic.node_role"

	// FieldKeyDynamicNodeStatus the dynamic node status field key.
	FieldKeyDynamicNodeStatus = "data.dynamic.node_status"

	// FieldKeyDynamicNodeVersion the dynamic node version field key.
	FieldKeyDynamicNodeVersion = "data.dynamic.node_version"

	// FieldKeyDynamicNodeGeneration the dynamic node generation field key.
	FieldKeyDynamicNodeGeneration = "data.dynamic.node_generation"

	// FieldKeyDynamicNodeCPUArch the dynamic node cpu arch field key.
	FieldKeyDynamicNodeCPUArch = "data.dynamic.node_cpu_arch"

	// FieldKeyDynamicNodeOsType the dynamic node os type field key.
	FieldKeyDynamicNodeOsType = "data.dynamic.node_os_type"

	// FieldKeyDynamicAgentID the dynamic agent id field key.
	FieldKeyDynamicAgentID = "data.dynamic.agent_id"

	// FieldKeyDynamicNetworkUnitID the dynamic networkunit id field key.
	FieldKeyDynamicNetworkUnitID = "data.dynamic.networkunit_id"

	// FieldKeyDynamicLoginIP the dynamic login ip field key.
	FieldKeyDynamicLoginIP = "data.dynamic.login_ip"

	// FieldKeyDynamicLoginPort the dynamic login port field key.
	FieldKeyDynamicLoginPort = "data.dynamic.login_port"

	// FieldKeyDynamicLoginUser the dynamic login user field key.
	FieldKeyDynamicLoginUser = "data.dynamic.login_user"

	// FieldKeyDynamicLoginMode the dynamic login mode field key.
	FieldKeyDynamicLoginMode = "data.dynamic.login_mode"

	// FieldKeyDynamicLoginCreditID the dynamic login credit id field key.
	FieldKeyDynamicLoginCreditID = "data.dynamic.login_credit_id"

	// FieldKeyDynamicExportIP the dynamic export ip field key.
	FieldKeyDynamicExportIP = "data.dynamic.export_ip"

	// FieldKeyDynamicExportIPV6 the dynamic export ipv6 field key.
	FieldKeyDynamicExportIPV6 = "data.dynamic.export_ipv6"

	// FieldKeyDynamicAdvertiseIP the dynamic advertise ip field key.
	FieldKeyDynamicAdvertiseIP = "data.dynamic.advertise_ip"

	// FieldKeyDynamicAdvertiseIPV6 the dynamic advertise ip field key.
	FieldKeyDynamicAdvertiseIPV6 = "data.dynamic.advertise_ipv6"

	// FieldKeyDynamicProxyTags the dynamic proxy tag field key.
	FieldKeyDynamicProxyTags = "data.dynamic.proxy_tags"

	// FieldKeyDynamicProxyClusterPort the dynamic proxy cluster port field key.
	FieldKeyDynamicProxyClusterPort = "data.dynamic.proxy_cluster_port"

	// FieldKeyDynamicProxyFilePort the dynamic proxy file port field key.
	FieldKeyDynamicProxyFilePort = "data.dynamic.proxy_file_port"

	// FieldKeyDynamicProxyDataPort the dynamic proxy data port field key.
	FieldKeyDynamicProxyDataPort = "data.dynamic.proxy_data_port"

	// FieldKeyDynamicProxyAccessDisabled the dynamic proxy access disabled field key.
	FieldKeyDynamicProxyAccessDisabled = "data.dynamic.proxy_access_disabled"

	// FieldKeyDynamicRelayDownloadPort the dynamic relay download port field key.
	FieldKeyDynamicRelayDownloadPort = "data.dynamic.relay_download_port"

	// FieldKeyDynamicRelayCallbackPort the dynamic relay callback port field key.
	FieldKeyDynamicRelayCallbackPort = "data.dynamic.relay_callback_port"

	// FieldKeyDynamicProxyInstallOriginUnitID the dynamic proxy install origin network unit id field key.
	FieldKeyDynamicProxyInstallOriginUnitID = "data.dynamic.proxy_install_origin_unit_id"

	// FieldKeyDynamicConnCycleTime the dynamic conn cycle time field key.
	FieldKeyDynamicConnCycleTime = "data.dynamic.conn_cycle_time"

	// FieldKeyDynamicOpsConsoleHostID the dynamic ops console host id field key.
	FieldKeyDynamicOpsConsoleHostID = "data.dynamic.ops_console_host_id"

	// FieldKeyDynamicOpsOutBandType the dynamic ops out-band type field key.
	FieldKeyDynamicOpsOutBandType = "data.dynamic.ops_out_band_type"

	// FieldKeyDynamicOpsOutBandProtocol the dynamic ops out-band protocol field key.
	FieldKeyDynamicOpsOutBandProtocol = "data.dynamic.ops_out_band_protocol"

	// FieldKeyDynamicOpsBMCIP the dynamic ops bmc ip field key.
	FieldKeyDynamicOpsBMCIP = "data.dynamic.ops_bmc_ip"

	// FieldKeyDynamicOpsBMCPort the dynamic ops bmc port field key.
	FieldKeyDynamicOpsBMCPort = "data.dynamic.ops_bmc_port"

	// FieldKeyOperationUpdatedAt the operation updated at field key for business-operation sort ordering.
	FieldKeyOperationUpdatedAt = "data.operation_updated_at"
)
