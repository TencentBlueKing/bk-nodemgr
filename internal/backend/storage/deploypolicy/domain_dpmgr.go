/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package deploypolicy

import (
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	daoDeployPolicy "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/deploypolicy"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func (s *Storage) discoverEnabledPoliciesBySpecifyPlugin(nCtx contextx.IContext, param *types.SpecifyPluginParam) ([]*types.DeployPolicy, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	if param == nil {
		return nil, base.ErrInvalidParam(errors.New("specify plugin param is nil"))
	}

	opts := convSpecifyPluginParamToOptions(param)
	// we only find the enabled policies.
	opts = append(opts, daoDeployPolicy.WithEnabled(true))
	policies, _, err := s.daoDeployPolicy.List(nCtx, types.UnlimitedPage(), opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to list deploy policies: %w", err)
	}

	return policies, nil
}

func convSpecifyPluginParamToOptions(param *types.SpecifyPluginParam) []daoDeployPolicy.OptFn {
	opts := make([]daoDeployPolicy.OptFn, 0)

	if param == nil {
		return opts
	}

	opts = append(opts, daoDeployPolicy.WithDeploySpecType(types.DeploySpecTypeSpecifyPlugin))

	if param.PluginName != "" {
		opts = append(opts, daoDeployPolicy.WithDeploySpecPluginName(param.PluginName))
	}

	return opts
}

func (s *Storage) refreshExecuteInfo(nCtx contextx.IContext, deployPolicy ...*types.DeployPolicy) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if len(deployPolicy) == 0 {
		return base.ErrInvalidParam(errors.New("deploy policy list is empty"))
	}

	if err := s.daoDeployPolicy.RefreshExecuteInfo(nCtx, deployPolicy...); err != nil {
		return fmt.Errorf("failed to refresh execute info: %w", err)
	}

	return nil
}
