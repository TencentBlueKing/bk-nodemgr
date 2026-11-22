/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package v3

import "github.com/TencentBlueKing/bk-nodemgr/pkg/cloudsg"

// Validate validates the pool list request.
func (x *TopoCloudSecurityGroupPoolListReq) Validate() error {
	return nil
}

// AutoConvert converts request defaults.
func (x *TopoCloudSecurityGroupPoolListReq) AutoConvert() {}

// ConvertPoolsFromTypes converts public pool summaries to proto.
func (x *TopoCloudSecurityGroupPoolListResp) ConvertPoolsFromTypes(pools []cloudsg.CloudSecurityGroupPoolSummary) {
	items := make([]*CloudSecurityGroupPool, len(pools))
	for idx, pool := range pools {
		items[idx] = &CloudSecurityGroupPool{
			PoolId:   new(string),
			PoolName: new(string),
		}
		*items[idx].PoolId = pool.PoolID
		*items[idx].PoolName = pool.PoolName
	}

	x.Data = &TopoCloudSecurityGroupPoolListResp_Data{Items: items}
}

// ConvertPoolsToTypes converts proto pool summaries to types.
func (x *TopoCloudSecurityGroupPoolListResp) ConvertPoolsToTypes() []cloudsg.CloudSecurityGroupPoolSummary {
	items := x.GetData().GetItems()
	pools := make([]cloudsg.CloudSecurityGroupPoolSummary, len(items))
	for idx, item := range items {
		pools[idx] = cloudsg.CloudSecurityGroupPoolSummary{
			PoolID:   item.GetPoolId(),
			PoolName: item.GetPoolName(),
		}
	}

	return pools
}
