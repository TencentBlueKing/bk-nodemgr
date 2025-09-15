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
	"context"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/host"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/networkunit"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// GetNetworkUnitByAreaIDs list network unit by area id.
// nolint: nonamedreturns
func (s *Storage) GetNetworkUnitByAreaIDs(ctx context.Context, networkAreaIDs []int64) (
	results []*types.NetworkUnit, err error) {

	opts := make([]networkunit.OptFn, 0)

	opts = append(opts, networkunit.WithNetworkAreaID(networkAreaIDs...))

	if results, _, err = s.daoNetworkUnit.List(ctx, types.UnlimitedPage(), opts...); err != nil {
		return nil, err
	}

	return results, nil
}

// GetHostsByAreaAndIP get hosts by area and ip.
// nolint: nonamedreturns
func (s *Storage) GetHostsByAreaAndIP(ctx context.Context,
	networkAreaID int64, ip string) (results []*types.Host, err error) {

	opts := make([]host.OptFn, 0)

	opts = append(opts, host.WithNetworkAreaID(networkAreaID))
	opts = append(opts, host.WithStaticInnerIP(ip))

	if results, _, err = s.daoHost.List(ctx, types.UnlimitedPage(), opts...); err != nil {
		return nil, err
	}

	return results, nil
}
