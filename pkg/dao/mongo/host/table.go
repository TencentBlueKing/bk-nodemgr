/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package host ...
package host

import (
	"fmt"
	"strconv"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
)

// TableName host table name.
func TableName(tenantID string) string {
	return fmt.Sprintf("host_%s", tenantID)
}

// HostStatic represents a host static information.
// nolint: revive
type HostStatic struct {
	BizID         int64    `json:"biz_id" bson:"biz_id"`
	SetID         int64    `json:"set_id" bson:"set_id"`
	ModuleID      int64    `json:"module_id" bson:"module_id"`
	NetworkAreaID int64    `json:"networkarea_id" bson:"networkarea_id"`
	HostName      string   `json:"host_name" bson:"host_name"`
	DeptName      string   `json:"dept_name" bson:"dept_name"`
	InnerIPList   []string `json:"inner_ip_list" bson:"inner_ip_list"`
	InnerIPV6List []string `json:"inner_ipv6_list" bson:"inner_ipv6_list"`
	OuterIPList   []string `json:"outer_ip_list" bson:"outer_ip_list"`
	OuterIPV6List []string `json:"outer_ipv6_list" bson:"outer_ipv6_list"`
	Mac           string   `json:"mac" bson:"mac"`
	Operator      string   `json:"operator" bson:"operator"`
	OSType        string   `json:"os_type" bson:"os_type"`
	OSTypeCCID    string   `json:"os_type_ccid" bson:"os_type_ccid"`
	Arch          string   `json:"arch" bson:"arch"`
	Addressing    string   `json:"addressing" bson:"addressing"`
	RegionID      string   `json:"region_id" bson:"region_id"`
	CityID        string   `json:"city_id" bson:"city_id"`
	SyncedAgentID string   `json:"synced_agent_id" bson:"synced_agent_id"`
	CPUNum        float64  `json:"cpu_num" bson:"cpu_num"`
	MemCap        float64  `json:"mem_cap" bson:"mem_cap"`
}

// HostDynamic represents a host dynamic information.
// nolint: revive
type HostDynamic struct {
	NodeRole                 string   `json:"node_role" bson:"node_role"`
	NodeStatus               string   `json:"node_status" bson:"node_status"`
	NodeVersion              string   `json:"node_version" bson:"node_version"`
	NodeGeneration           int64    `json:"node_generation" bson:"node_generation"`
	NodeCPUArch              string   `json:"node_cpu_arch" bson:"node_cpu_arch"`
	NodeOsType               string   `json:"node_os_type" bson:"node_os_type"`
	AgentID                  string   `json:"agent_id" bson:"agent_id"`
	NetworkUnitID            int64    `json:"networkunit_id" bson:"networkunit_id"`
	ProxyAccessDisabled      bool     `json:"proxy_access_disabled" bson:"proxy_access_disabled"`
	ProxyTags                []string `json:"proxy_tags" bson:"proxy_tags"`
	ProxyInstallOriginUnitID int64    `json:"proxy_install_origin_unit_id" bson:"proxy_install_origin_unit_id"`
	ProxyClusterPort         int64    `json:"proxy_cluster_port" bson:"proxy_cluster_port"`
	ProxyDataPort            int64    `json:"proxy_data_port" bson:"proxy_data_port"`
	ProxyFilePort            int64    `json:"proxy_file_port" bson:"proxy_file_port"`
	LoginIP                  string   `json:"login_ip" bson:"login_ip"`
	LoginPort                int64    `json:"login_port" bson:"login_port"`
	LoginUser                string   `json:"login_user" bson:"login_user"`
	LoginMode                string   `json:"login_mode" bson:"login_mode"`
	LoginCreditID            string   `json:"login_credit_id" bson:"login_credit_id"`
	ExportIP                 string   `json:"export_ip" bson:"export_ip"`
	ExportIPV6               string   `json:"export_ipv6" bson:"export_ipv6"`
	AdvertiseIP              string   `json:"advertise_ip" bson:"advertise_ip"`
	AdvertiseIPV6            string   `json:"advertise_ipv6" bson:"advertise_ipv6"`
	RelayDownloadPort        int64    `json:"relay_download_port" bson:"relay_download_port"`
	RelayCallbackPort        int64    `json:"relay_callback_port" bson:"relay_callback_port"`
	ConnCycleTime            string   `json:"conn_cycle_time" bson:"conn_cycle_time"`
}

var _ base.IData = &Host{}

// Host represents a host.
type Host struct {
	HostID   int64  `json:"host_id" bson:"host_id"`
	TenantID string `json:"tenant_id" bson:"tenant_id"`

	Static  *HostStatic  `json:"static" bson:"static"`
	Dynamic *HostDynamic `json:"dynamic" bson:"dynamic"`

	OperationUpdatedAt *time.Time `json:"operation_updated_at,omitempty" bson:"operation_updated_at,omitempty"`
}

// UniqueFields unique fields of the table.
func (h *Host) UniqueFields() []string {
	return []string{FieldKeyHostID}
}

// UniqueKey unique key of the table.
func (h *Host) UniqueKey() string {
	return strconv.FormatInt(h.HostID, 10)
}

// TableHost represents the complete db structures of a host.
type TableHost base.TableBroker[*Host]
