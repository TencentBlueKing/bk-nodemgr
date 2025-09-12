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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/networkunit"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IDomainNodeInstall defines the Storage interface for domain node install.
type IDomainNodeInstall interface {
	// ListHostByConditions lists hosts by page and conditions.
	ListHostByConditions(ctx context.Context, page types.Page, conditions ...*types.HostCondition) ([]*types.Host, int64, error)

	ListNetworkUnitByConditions(ctx context.Context, page types.Page, conditions ...*types.NetworkUnitCondition) (
		results []*types.NetworkUnit, num int64, err error)
}

// ListHostByConditions lists hosts by page and conditions.
// nolint: nonamedreturns
func (s *Storage) ListHostByConditions(ctx context.Context, page types.Page, conditions ...*types.HostCondition) (
	results []*types.Host, mun int64, err error) {

	// record metric.
	metric := s.metric().Start("list_host_by_conditions")
	defer metric.End(err)

	opts := convertHostConditionsToOptions(conditions...)

	if results, mun, err = s.daoHost.List(ctx, page, opts...); err != nil {
		return nil, 0, err
	}

	return results, mun, nil
}

// ListNetworkUnitByConditions lists networkunit by page and conditions.
// nolint: nonamedreturns
func (s *Storage) ListNetworkUnitByConditions(ctx context.Context, page types.Page, conditions ...*types.NetworkUnitCondition) (
	results []*types.NetworkUnit, num int64, err error) {

	// record metric.
	metric := s.metric().Start("list_networkunit")
	defer metric.End(err)

	opts := make([]networkunit.OptFn, 0)
	for _, condition := range conditions {
		if condition == nil {
			continue
		}

		if condition.ExactInclude != nil {
			opts = append(opts,
				networkunit.WithNetworkUnitID(condition.ExactInclude.NetworkUnitID...),
				networkunit.WithNetworkAreaID(condition.ExactInclude.NetworkAreaID...),
			)
		}

		if condition.ExactExclude != nil {
			opts = append(opts,
				networkunit.WithoutNetworkUnitID(condition.ExactExclude.NetworkUnitID...),
				networkunit.WithoutNetworkAreaID(condition.ExactExclude.NetworkAreaID...),
			)
		}
	}

	if results, num, err = s.daoNetworkUnit.List(ctx, page, opts...); err != nil {
		return nil, 0, err
	}

	return results, num, nil
}
