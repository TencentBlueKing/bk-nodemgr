package syncdata

import (
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
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

func TestHostIDAgentIDWorkflowPayloadCompatibility(t *testing.T) {
	legacyHost := &HostIDAgentID{
		HostID:  100,
		AgentID: "agent-100",
	}
	legacyPayload, err := conv.StructToMap(legacyHost)
	if err != nil {
		t.Fatalf("failed to convert legacy host to map: %v", err)
	}
	if _, ok := legacyPayload["current_node_role"]; ok {
		t.Fatalf("legacy payload should omit current_node_role")
	}
	if _, ok := legacyPayload["current_node_status"]; ok {
		t.Fatalf("legacy payload should omit current_node_status")
	}
	if _, ok := legacyPayload["current_node_version"]; ok {
		t.Fatalf("legacy payload should omit current_node_version")
	}
	if _, ok := legacyPayload["current_node_generation"]; ok {
		t.Fatalf("legacy payload should omit current_node_generation")
	}

	nodeRole := types.NodeRoleAgent
	nodeStatus := types.NodeStatusRunning
	nodeVersion := "2.0.1"
	nodeGeneration := types.Generation2
	currentHost := &HostIDAgentID{
		HostID:                100,
		AgentID:               "agent-100",
		CurrentNodeRole:       &nodeRole,
		CurrentNodeStatus:     &nodeStatus,
		CurrentNodeVersion:    &nodeVersion,
		CurrentNodeGeneration: &nodeGeneration,
	}
	currentPayload, err := conv.StructToMap(currentHost)
	if err != nil {
		t.Fatalf("failed to convert current host to map: %v", err)
	}

	var decoded HostIDAgentID
	if err = conv.MapToStruct(currentPayload, &decoded); err != nil {
		t.Fatalf("failed to convert current host payload to struct: %v", err)
	}
	if decoded.CurrentNodeRole == nil || *decoded.CurrentNodeRole != nodeRole {
		t.Fatalf("decoded current node role = %v, want %s", decoded.CurrentNodeRole, nodeRole)
	}
	if decoded.CurrentNodeStatus == nil || *decoded.CurrentNodeStatus != nodeStatus {
		t.Fatalf("decoded current node status = %v, want %s", decoded.CurrentNodeStatus, nodeStatus)
	}
	if decoded.CurrentNodeVersion == nil || *decoded.CurrentNodeVersion != nodeVersion {
		t.Fatalf("decoded current node version = %v, want %s", decoded.CurrentNodeVersion, nodeVersion)
	}
	if decoded.CurrentNodeGeneration == nil || *decoded.CurrentNodeGeneration != nodeGeneration {
		t.Fatalf("decoded current node generation = %v, want %d", decoded.CurrentNodeGeneration, nodeGeneration)
	}
}
