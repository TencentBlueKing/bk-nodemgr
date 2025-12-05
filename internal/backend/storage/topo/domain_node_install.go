/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package topo

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/host"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/networkunit"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func (s *Storage) getHostsByAreaAndInnerIP(nCtx contextx.IContext,
	networkAreaID int64, innerip string) ([]*types.Host, error) {

	opts := make([]host.OptFn, 0)
	opts = append(opts,
		host.WithStaticNetworkAreaID(networkAreaID),
		host.WithStaticInnerIPList(innerip),
	)

	results, _, err := s.daoHost.List(nCtx, types.UnlimitedPage(), opts...)
	if err != nil {
		return nil, fmt.Errorf("list hosts failed. area-id(%d), innerip(%s): %w", networkAreaID, innerip, err)
	}

	return results, nil
}

func (s *Storage) existDedicatedInstallerProxyHost(nCtx contextx.IContext, networkUnitID int64) (bool, error) {
	opts := make([]host.OptFn, 0)
	opts = append(opts,
		host.WithDynamicNetworkUnitID(networkUnitID),
		host.WithDynamicNodeRole(types.NodeRoleProxy),
		host.WithDynamicNodeStatus(types.NodeStatusRunning),
		host.WithDynamicProxyTags(types.ProxyTagDedicatedInstaller),
	)

	exist, err := s.daoHost.Exist(nCtx, opts...)
	if err != nil {
		return false, fmt.Errorf("exist dedicated installer proxy host failed. unit-id(%d): %w", networkUnitID, err)
	}

	return exist, nil
}

func (s *Storage) getNetworkUnitByIDs(nCtx contextx.IContext, networkUnitIDs []int64) ([]*types.NetworkUnit, error) {
	opts := make([]networkunit.OptFn, 0)
	opts = append(opts, networkunit.WithNetworkUnitID(networkUnitIDs...))

	results, _, err := s.daoNetworkUnit.List(nCtx, types.UnlimitedPage(), opts...)
	if err != nil {
		return nil, fmt.Errorf("list networkunits failed. unit-ids(%v): %w", networkUnitIDs, err)
	}

	return results, nil
}
