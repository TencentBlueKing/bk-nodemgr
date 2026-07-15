/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package syncdata

import "github.com/TencentBlueKing/bk-nodemgr/pkg/types"

func fillDefaultAdvertiseIP(host, dbHost *types.Host) bool {
	if host == nil || host.Dynamic == nil || host.Static == nil {
		return false
	}

	if dbHost != nil && dbHost.Dynamic != nil {
		if dbHost.Dynamic.AdvertiseIP != "" {
			host.Dynamic.AdvertiseIP = dbHost.Dynamic.AdvertiseIP
		}
		if dbHost.Dynamic.AdvertiseIPV6 != "" {
			host.Dynamic.AdvertiseIPV6 = dbHost.Dynamic.AdvertiseIPV6
		}
	}

	needUpdate := false
	if host.Dynamic.AdvertiseIP == "" && len(host.Static.InnerIPList) > 0 {
		host.Dynamic.AdvertiseIP = host.Static.InnerIPList[0]
		needUpdate = true
	} else if dbHost != nil && host.Dynamic.AdvertiseIP != "" &&
		(dbHost.Dynamic == nil || dbHost.Dynamic.AdvertiseIP == "") {

		needUpdate = true
	}

	if host.Dynamic.AdvertiseIPV6 == "" && len(host.Static.InnerIPV6List) > 0 {
		host.Dynamic.AdvertiseIPV6 = host.Static.InnerIPV6List[0]
		needUpdate = true
	} else if dbHost != nil && host.Dynamic.AdvertiseIPV6 != "" &&
		(dbHost.Dynamic == nil || dbHost.Dynamic.AdvertiseIPV6 == "") {

		needUpdate = true
	}

	return needUpdate
}

func fillDefaultAdvertiseIPs(hosts, dbData []*types.Host) []*types.Host {
	dbHostMap := make(map[int64]*types.Host, len(dbData))
	for _, host := range dbData {
		dbHostMap[host.HostID] = host
	}

	backfillHosts := make([]*types.Host, 0)
	for _, host := range hosts {
		if fillDefaultAdvertiseIP(host, dbHostMap[host.HostID]) {
			backfillHosts = append(backfillHosts, host)
		}
	}

	return backfillHosts
}
