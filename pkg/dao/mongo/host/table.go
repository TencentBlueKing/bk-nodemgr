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

// Host represents a host.
type Host struct {
	TenantID      string `json:"tenant_id" bson:"tenant_id"`
	NetworkAreaID int64  `json:"network_area_id" bson:"network_area_id"`
	BizID         int64  `json:"biz_id" bson:"biz_id"`
	HostID        int64  `json:"host_id" bson:"host_id"`
	InnerIP       string `json:"inner_ip" bson:"inner_ip"`
	Mac           string `json:"mac" bson:"mac"`
	OSType        string `json:"os_type" bson:"os_type"`
}

// UniqueKey unique key of the table.
func (h *Host) UniqueKey() string {
	return strconv.FormatInt(h.HostID, 10)
}

// TableHost represents the complete db structures of a host.
type TableHost base.TableBroker[*Host]
