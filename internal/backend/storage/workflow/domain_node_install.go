/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package workflow

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/operation"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	workoper "github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

// listOperationByNodeWorkflowOperationCondition lists operation by node workflow condition.
func (s *Storage) listOperationByNodeWorkflowOperationCondition(
	nCtx contextx.IContext, page types.Page, conditions ...*types.NodeWorkflowOperationCondition) (
	[]*workoper.Operation, int64, error) {

	opts := make([]operation.OptFn, 0)
	for _, condition := range conditions {
		if condition == nil {
			continue
		}

		if condition.ExactInclude != nil {
			opts = append(opts,
				operation.WithTriggerID(condition.ExactInclude.TriggerID),
				operation.WithBizID(condition.ExactInclude.BizID...),
				operation.WithNetworkAreaID(condition.ExactInclude.NetworkAreaID...),
				operation.WithIPv4(condition.ExactInclude.InnerIP...),
				operation.WithIPv6(condition.ExactInclude.InnerIPv6...),
			)
		}
	}

	return s.daoOperation.List(nCtx, page, opts...)
}
