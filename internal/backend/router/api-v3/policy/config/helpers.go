/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package config

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func (h *handler) getConfigPolicy(nCtx contextx.IContext, configpolicyID []int64) ([]*types.ConfigPolicy, error) {
	// get the config policy info.
	configpolicies, _, err := h.storageConfigPolicy.ListConfigPolicy(nCtx, types.UnlimitedPage(),
		&types.ConfigPolicyCondition{
			ExactInclude: &types.ConfigPolicyExactFields{
				ConfigPolicyID: configpolicyID,
			},
		})
	if err != nil {
		return nil, fmt.Errorf("failed to get config policy: %w", err)
	}

	if len(configpolicies) == 0 {
		return nil, fmt.Errorf("config policy not found")
	}

	return configpolicies, nil
}
