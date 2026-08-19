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
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func TestSyncAgentStateHostFieldSelection(t *testing.T) {
	legacySelection := syncAgentStateHostFieldSelection(false)
	if !legacySelection.HostID || !legacySelection.AgentID {
		t.Fatalf("legacy selection should include host id and agent id")
	}
	if legacySelection.NodeRole || legacySelection.NodeStatus || legacySelection.NodeVersion || legacySelection.NodeGeneration {
		t.Fatalf("legacy selection should not include current state fields")
	}

	compareSelection := syncAgentStateHostFieldSelection(true)
	if !compareSelection.HostID || !compareSelection.AgentID || !compareSelection.NodeRole || !compareSelection.NodeStatus ||
		!compareSelection.NodeVersion || !compareSelection.NodeGeneration {
		t.Fatalf("compare selection should include host id, agent id and all current state fields")
	}
}

func TestNewSyncAgentStateHostIDAgentID(t *testing.T) {
	host := &types.Host{
		HostID: 100,
		Dynamic: &types.HostDynamic{
			AgentID:        "agent-100",
			NodeRole:       types.NodeRoleAgent,
			NodeStatus:     types.NodeStatusRunning,
			NodeVersion:    "2.0.1",
			NodeGeneration: types.Generation2,
		},
	}

	legacyHost := newSyncAgentStateHostIDAgentID(host, false)
	if legacyHost.HostID != host.HostID || legacyHost.AgentID != host.Dynamic.AgentID {
		t.Fatalf("legacy host = %#v, want host_id %d agent_id %s", legacyHost, host.HostID, host.Dynamic.AgentID)
	}
	if legacyHost.CurrentNodeRole != nil || legacyHost.CurrentNodeStatus != nil || legacyHost.CurrentNodeVersion != nil ||
		legacyHost.CurrentNodeGeneration != nil {
		t.Fatalf("legacy host should not carry current state snapshot")
	}

	compareHost := newSyncAgentStateHostIDAgentID(host, true)
	if compareHost.CurrentNodeRole == nil || *compareHost.CurrentNodeRole != host.Dynamic.NodeRole {
		t.Fatalf("current node role snapshot = %v, want %s", compareHost.CurrentNodeRole, host.Dynamic.NodeRole)
	}
	if compareHost.CurrentNodeStatus == nil || *compareHost.CurrentNodeStatus != host.Dynamic.NodeStatus {
		t.Fatalf("current node status snapshot = %v, want %s", compareHost.CurrentNodeStatus, host.Dynamic.NodeStatus)
	}
	if compareHost.CurrentNodeVersion == nil || *compareHost.CurrentNodeVersion != host.Dynamic.NodeVersion {
		t.Fatalf("current node version snapshot = %v, want %s", compareHost.CurrentNodeVersion, host.Dynamic.NodeVersion)
	}
	if compareHost.CurrentNodeGeneration == nil || *compareHost.CurrentNodeGeneration != host.Dynamic.NodeGeneration {
		t.Fatalf("current node generation snapshot = %v, want %d", compareHost.CurrentNodeGeneration,
			host.Dynamic.NodeGeneration)
	}
}
