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
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Validate check body.
func (x *TopoGraphGetReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *TopoGraphGetReq) AutoConvert() {
}

// ConvertNetworkUnitsToTypes convert networkunits from types to proto.
func (x *TopoGraphGetResp) ConvertNetworkUnitsToTypes(networkUnits []*types.NetworkUnit) {
	unitItems := make([]*NetworkUnitGraph, len(networkUnits))
	for idx, networkUnit := range networkUnits {
		item := newEmptyNetworkUnitGraph()
		*item.TenantId = networkUnit.TenantID
		*item.BkNetworkunitId = networkUnit.ID
		*item.BkNetworkunitName = networkUnit.Name
		*item.BkNetworkareaId = networkUnit.NetworkAreaID
		item.Accesspoints = networkUnit.AccessPoints

		unitItems[idx] = item
	}

	x.Data = &TopoGraphGetResp_Data{
		Networkunit: unitItems,
		Links:       convertLinks(networkUnits),
	}
}

// Validate check body.
func (x *TopoGraphNodeCountReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *TopoGraphNodeCountReq) AutoConvert() {
}

// NetworkUnitInfo describes the informations in one networkunit.
type NetworkUnitInfo struct {
	Proxy int64
	Agent int64
}

// ConvertNetworkUnitInfosFromTypes convert networkunitinfos from types to proto.
func (x *TopoGraphNodeCountResp) ConvertNetworkUnitInfosFromTypes(networkUnitInfos map[int64]*NetworkUnitInfo) {
	items := make([]*TopoGraphNodeCountResp_NodeInfo, len(networkUnitInfos))
	idx := 0
	for networkUnitID, networkUnitInfo := range networkUnitInfos {
		items[idx] = &TopoGraphNodeCountResp_NodeInfo{
			BkNetworkunitId: networkUnitID,
			Proxy:           networkUnitInfo.Proxy,
			Agent:           networkUnitInfo.Agent,
		}
		idx++
	}

	x.Data = &TopoGraphNodeCountResp_Data{
		Networkunits: items,
	}
}

func convertLinks(networkUnits []*types.NetworkUnit) []*LinkGraph {
	links := make([]*LinkGraph, 0)
	for _, networkunit := range networkUnits {
		validLinks := make([]*LinkGraph, 0)

		for _, unitLink := range []struct {
			l *types.Link
			s string
		}{
			{
				l: networkunit.Links.Cluster,
				s: "cluster",
			},
			{
				l: networkunit.Links.File,
				s: "file",
			},
			{
				l: networkunit.Links.Data,
				s: "data",
			},
		} {
			if unitLink.l == nil {
				continue
			}

			targetNetworkUnitID, err := findNetworkUnitIDByAccessPointID(unitLink.l.AccessPointID, networkUnits)
			if err != nil {
				continue
			}

			found := false
			for _, validLink := range validLinks {
				if targetNetworkUnitID == validLink.GetTargetNetworkunitId() {
					validLink.Channel = append(validLink.Channel, unitLink.s)

					found = true

					break
				}
			}
			if !found {
				validLink := newEmptyLinkGraph()
				*validLink.SourceNetworkunitId = networkunit.ID
				*validLink.TargetNetworkunitId = targetNetworkUnitID
				*validLink.TargetAccesspointId = unitLink.l.AccessPointID
				validLink.Channel = []string{unitLink.s}

				validLinks = append(validLinks, validLink)
			}
		}

		links = append(links, validLinks...)
	}

	return links
}

func findNetworkUnitIDByAccessPointID(accessPointID int64, networkUnits []*types.NetworkUnit) (int64, error) {
	for _, networkUnit := range networkUnits {
		for _, apID := range networkUnit.AccessPoints {
			if apID == accessPointID {
				return networkUnit.ID, nil
			}
		}
	}

	return -1, errors.New("network unit not found")
}

func newEmptyLinkGraph() *LinkGraph {
	return &LinkGraph{
		SourceNetworkunitId: new(int64),
		TargetNetworkunitId: new(int64),
		TargetAccesspointId: new(int64),
		Channel:             make([]string, 0),
	}
}

func newEmptyNetworkUnitGraph() *NetworkUnitGraph {
	return &NetworkUnitGraph{
		TenantId:          new(string),
		BkNetworkunitId:   new(int64),
		BkNetworkunitName: new(string),
		BkNetworkareaId:   new(int64),
		Accesspoints:      make([]int64, 0),
	}
}
