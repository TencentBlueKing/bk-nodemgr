/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package v3

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Validate check body.
func (x *TopoGraphNodeGetReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *TopoGraphNodeGetReq) AutoConvert() {
}

// ConvertGrapthNodeInfoFromTypes convert graph node info.
func (x *TopoGraphNodeGetResp) ConvertGrapthNodeInfoFromTypes(graphNodeInfos map[int64]*types.GraphNodeInfo) {
	graphNodeItems := make([]*GraphNodeInfo, len(graphNodeInfos))
	idx := 0
	for networkUnitID, graphNodeInfo := range graphNodeInfos {
		info := &GraphNodeInfo{
			BkNetworkunitId: networkUnitID,
			RunningProxy:    graphNodeInfo.RunningProxy,
			TotalProxy:      graphNodeInfo.TotalProxy,
			RunningAgent:    graphNodeInfo.RunningAgent,
			TotalAgent:      graphNodeInfo.TotalAgent,
			IsHealthy:       graphNodeInfo.IsHealthy,
			CycleTimes: conv.SliceToSlice(graphNodeInfo.CycleTimes, func(item types.CycleTime) *CycleTime {
				return &CycleTime{
					BkHostId:            &item.HostID,
					BkHostInneripList:   item.InnerIP,
					BkHostInneripV6List: item.InnerIPV6,
					BkAgentId:           &item.AgentID,
					Time:                item.Time,
				}
			}),
		}

		graphNodeItems[idx] = info
		idx++
	}

	x.Data = &TopoGraphNodeGetResp_Data{
		GraphNodeInfo: graphNodeItems,
	}
}

// ConvertResultToTypes converts the response to types.
func (x *TopoGraphNodeGetResp) ConvertResultToTypes() []*types.GraphNodeInfo {
	data := x.GetData()
	if data == nil {
		return nil
	}

	items := data.GetGraphNodeInfo()
	result := make([]*types.GraphNodeInfo, len(items))
	for idx, item := range items {
		result[idx] = &types.GraphNodeInfo{
			NetworkUnitID: item.GetBkNetworkunitId(),
			RunningProxy:  item.GetRunningProxy(),
			TotalProxy:    item.GetTotalProxy(),
			RunningAgent:  item.GetRunningAgent(),
			TotalAgent:    item.GetTotalAgent(),
			IsHealthy:     item.GetIsHealthy(),
			CycleTimes: conv.SliceToSlice(item.GetCycleTimes(), func(item *CycleTime) types.CycleTime {
				return types.CycleTime{
					HostID:    item.GetBkHostId(),
					InnerIP:   item.GetBkHostInneripList(),
					InnerIPV6: item.GetBkHostInneripV6List(),
					AgentID:   item.GetBkAgentId(),
					Time:      item.GetTime(),
				}
			}),
		}
	}

	return result
}
