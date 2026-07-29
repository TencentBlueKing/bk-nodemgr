package syncdata

import (
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func TestShouldSyncAgentStateHost(t *testing.T) {
	nodeRole := types.NodeRoleAgent
	nodeStatus := types.NodeStatusRunning
	nodeVersion := "2.0.1"
	nodeGeneration := types.Generation2
	agentState := &types.AgentState{
		NodeRole:       nodeRole,
		NodeStatus:     nodeStatus,
		Version:        nodeVersion,
		NodeGeneration: nodeGeneration,
	}

	tests := []struct {
		name                string
		host                *HostIDAgentID
		compareCurrentState bool
		expected            bool
	}{
		{
			name:                "legacy mode always syncs",
			host:                &HostIDAgentID{},
			compareCurrentState: false,
			expected:            true,
		},
		{
			name: "complete equal snapshot skips sync",
			host: &HostIDAgentID{
				CurrentNodeRole:       &nodeRole,
				CurrentNodeStatus:     &nodeStatus,
				CurrentNodeVersion:    &nodeVersion,
				CurrentNodeGeneration: &nodeGeneration,
			},
			compareCurrentState: true,
			expected:            false,
		},
		{
			name: "complete changed snapshot syncs",
			host: &HostIDAgentID{
				CurrentNodeRole:       &nodeRole,
				CurrentNodeStatus:     &nodeStatus,
				CurrentNodeVersion:    ptr("2.0.0"),
				CurrentNodeGeneration: &nodeGeneration,
			},
			compareCurrentState: true,
			expected:            true,
		},
		{
			name: "incomplete snapshot preserves legacy sync",
			host: &HostIDAgentID{
				CurrentNodeRole:       &nodeRole,
				CurrentNodeStatus:     &nodeStatus,
				CurrentNodeGeneration: &nodeGeneration,
			},
			compareCurrentState: true,
			expected:            true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := shouldSyncAgentStateHost(tt.host, agentState, tt.compareCurrentState)
			if actual != tt.expected {
				t.Fatalf("shouldSyncAgentStateHost() = %v, want %v", actual, tt.expected)
			}
		})
	}
}

func ptr[T any](value T) *T {
	return &value
}
