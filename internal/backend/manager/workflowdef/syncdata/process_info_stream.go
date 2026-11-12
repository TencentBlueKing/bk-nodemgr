/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package syncdata

import (
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// collectProcessInfoDeltas persists each confirmed result without waiting for unfinished targets.
// A failed query is not evidence that a process is no longer controlled by GSE.
func collectProcessInfoDeltas(
	nCtx contextx.IContext, results <-chan gse.ProcessInfoResult, agentIDHostIDMap map[string]int64,
	save func(...*types.ProcessInfoDelta) error,
) ([]*types.ProcessInfoDelta, error) {

	deltas := make([]*types.ProcessInfoDelta, 0)
	queryErrors := make([]error, 0)
	for result := range results {
		if result.Err != nil {
			queryErrors = append(queryErrors, fmt.Errorf("failed to query plugin(%s) on agent(%s): %w",
				result.PluginName, result.AgentID, result.Err))

			continue
		}
		confirmed := make([]*types.ProcessInfoDelta, 0, len(result.ProcessInfos))
		for _, info := range result.ProcessInfos {
			confirmed = append(confirmed, &types.ProcessInfoDelta{
				HostID: agentIDHostIDMap[result.AgentID], PluginName: result.PluginName, ProcessInfo: info,
			})
		}
		if err := batchHandleProcessInfoDeltas(nCtx, confirmed, save); err != nil {
			queryErrors = append(queryErrors, fmt.Errorf("failed to save plugin(%s) on agent(%s): %w",
				result.PluginName, result.AgentID, err))

			continue
		}
		deltas = append(deltas, confirmed...)
	}

	return deltas, errors.Join(queryErrors...)
}
