/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package proxy

import (
	"fmt"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// validateHostNetworkUnit checks that all hosts have a valid network unit (NetworkUnitID >= 0).
// Returns an error listing all invalid hosts if any are found.
func validateHostNetworkUnit(hosts map[int64]*types.Host) error {
	var invalid []string
	for _, host := range hosts {
		if host.Dynamic.NetworkUnitID < 0 {
			invalid = append(invalid, fmt.Sprintf("host-id(%d) networkunit-id(%d)", host.HostID, host.Dynamic.NetworkUnitID))
		}
	}

	if len(invalid) == 0 {
		return nil
	}

	return fmt.Errorf("hosts have no valid network unit. %s", strings.Join(invalid, ", "))
}
