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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
)

// TableName host table name.
func TableName(tenantID string) string {
	return fmt.Sprintf("host_%s", tenantID)
}

// HostStatic represents a host static information.
type HostStatic struct {
	BizID         int64  `json:"biz_id" bson:"biz_id"`
	NetworkAreaID int64  `json:"networkarea_id" bson:"networkarea_id"`
	HostName      string `json:"host_name" bson:"host_name"`
	DeptName      string `json:"dept_name" bson:"dept_name"`
	InnerIP       string `json:"inner_ip" bson:"inner_ip"`
	InnerIPV6     string `json:"inner_ipv6" bson:"inner_ipv6"`
	OuterIP       string `json:"outer_ip" bson:"outer_ip"`
	OuterIPV6     string `json:"outer_ipv6" bson:"outer_ipv6"`
	Mac           string `json:"mac" bson:"mac"`
	OSType        string `json:"os_type" bson:"os_type"`
}

// HostDynamic represents a host dynamic information.
type HostDynamic struct {
	NodeRole       string `json:"node_role" bson:"node_role"`
	NodeStatus     string `json:"node_status" bson:"node_status"`
	NodeVersion    string `json:"node_version" bson:"node_version"`
	NodeGeneration int64  `json:"node_generation" bson:"node_generation"`
	AgentID        string `json:"agent_id" bson:"agent_id"`
	NetworkUnitID  int64  `json:"networkunit_id" bson:"networkunit_id"`
}

// Host represents a host.
type Host struct {
	HostID   int64  `json:"host_id" bson:"host_id"`
	TenantID string `json:"tenant_id" bson:"tenant_id"`

	Static  *HostStatic  `json:"static" bson:"static"`
	Dynamic *HostDynamic `json:"dynamic" bson:"dynamic"`
}

// UniqueKey unique key of the table.
func (h *Host) UniqueKey() string {
	return strconv.FormatInt(h.HostID, 10)
}

// TableHost represents the complete db structures of a host.
type TableHost base.TableBroker[*Host]
