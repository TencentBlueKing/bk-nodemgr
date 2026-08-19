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

package proxy

import (
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"testing"
)

func TestValidateUpgradeHostNetworkUnit_UsesTargetUnitWhenCurrentInvalid(t *testing.T) {
	hostID := int64(100)
	targetUnitID := int64(10)
	hosts := map[int64]*types.Host{
		hostID: &types.Host{
			HostID: hostID,
			Static: &types.HostStatic{BizID: 3},
			Dynamic: &types.HostDynamic{
				NetworkUnitID: -1,
			},
		},
	}

	err := validateUpgradeHostNetworkUnit([]*protoBackend.NodeProxyUpgradeReq_Host{
		{
			BkHostId:        hostID,
			BkNetworkunitId: &targetUnitID,
		},
	}, hosts)
	if err != nil {
		t.Fatalf("expected target network unit to satisfy validation, got %v", err)
	}
}
