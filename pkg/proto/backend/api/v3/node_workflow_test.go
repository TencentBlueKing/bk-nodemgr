/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package v3

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

func buildMinimalOperationInstanceData(privateData map[string]any) *operation.InstanceData {
	return &operation.InstanceData{
		InstanceBriefData: operation.InstanceBriefData{
			Metadata: &operation.InstanceMetadata{},
		},
		ActionInstanceDataMap: map[string]*action.InstanceData{
			"action-1": {
				DisplayNameZh: "测试动作",
				DisplayNameEn: "Test Action",
				Messages:      nil,
				PrivateData:   privateData,
				Lifecycle:     &action.Lifecycle{},
			},
		},
	}
}

// TestNodeWorkflowOperationInstanceLogGetResp_ConvertResultFromTypes_WithSubWorkflowRefs verifies
// that sub_workflow_refs is populated from JSON-encoded private data.
func TestNodeWorkflowOperationInstanceLogGetResp_ConvertResultFromTypes_WithSubWorkflowRefs(t *testing.T) {
	resp := &NodeWorkflowOperationInstanceLogGetResp{}
	resp.ConvertResultFromTypes(buildMinimalOperationInstanceData(map[string]any{
		types.PDKeySubWorkflowRefs: `[{"workflow_id":"wf-plugin-123","workflow_domain":"plugin"}]`,
	}))

	assert.NotNil(t, resp.Data)
	actionData := resp.Data.OperInstLogs["action-1"]
	assert.NotNil(t, actionData)
	assert.Len(t, actionData.SubWorkflowRefs, 1)
	assert.Equal(t, "wf-plugin-123", actionData.SubWorkflowRefs[0].WorkflowId)
	assert.Equal(t, string(types.WorkflowDomainPlugin), actionData.SubWorkflowRefs[0].WorkflowDomain)
}

func TestNodeWorkflowOperationInstanceLogGetResp_ConvertResultFromTypes_SubWorkflowInvalidJSON(t *testing.T) {
	resp := &NodeWorkflowOperationInstanceLogGetResp{}
	resp.ConvertResultFromTypes(buildMinimalOperationInstanceData(map[string]any{
		types.PDKeySubWorkflowRefs: `{"workflow_id":"wf-bson"`,
	}))

	require.NotNil(t, resp.Data)
	act := resp.Data.OperInstLogs["action-1"]
	require.NotNil(t, act)
	assert.Empty(t, act.SubWorkflowRefs)
}

func TestNodeWorkflowOperationInstanceLogGetResp_ConvertResultFromTypes_SubWorkflowUnexpectedType(t *testing.T) {
	resp := &NodeWorkflowOperationInstanceLogGetResp{}
	resp.ConvertResultFromTypes(buildMinimalOperationInstanceData(map[string]any{
		types.PDKeySubWorkflowRefs: []any{"unexpected"},
	}))

	require.NotNil(t, resp.Data)
	act := resp.Data.OperInstLogs["action-1"]
	require.NotNil(t, act)
	assert.Empty(t, act.SubWorkflowRefs)
}

// TestNodeWorkflowOperationInstanceLogGetResp_ConvertResultFromTypes_NoSubWorkflowRefs verifies
// that sub_workflow_refs is empty when action PrivateData contains no known keys.
func TestNodeWorkflowOperationInstanceLogGetResp_ConvertResultFromTypes_NoSubWorkflowRefs(t *testing.T) {
	resp := &NodeWorkflowOperationInstanceLogGetResp{}
	resp.ConvertResultFromTypes(buildMinimalOperationInstanceData(map[string]any{
		"some_other_key": "irrelevant",
	}))

	assert.NotNil(t, resp.Data)
	actionData := resp.Data.OperInstLogs["action-1"]
	assert.NotNil(t, actionData)
	assert.Empty(t, actionData.SubWorkflowRefs)
}

// TestNodeWorkflowOperationInstanceLogGetResp_ConvertResultFromTypes_EmptyPrivateData verifies
// that sub_workflow_refs is empty when PrivateData is nil or empty.
func TestNodeWorkflowOperationInstanceLogGetResp_ConvertResultFromTypes_EmptyPrivateData(t *testing.T) {
	resp := &NodeWorkflowOperationInstanceLogGetResp{}
	resp.ConvertResultFromTypes(buildMinimalOperationInstanceData(nil))

	assert.NotNil(t, resp.Data)
	actionData := resp.Data.OperInstLogs["action-1"]
	assert.NotNil(t, actionData)
	assert.Empty(t, actionData.SubWorkflowRefs)
}

// TestNodeWorkflowOperationInstanceLogGetResp_BFFRoundTrip_PreservesSubWorkflowRefs verifies that proto →
// operation.InstanceData → proto preserves JSON-encoded sub_workflow_refs (application service uses this path).
func TestNodeWorkflowOperationInstanceLogGetResp_BFFRoundTrip_PreservesSubWorkflowRefs(t *testing.T) {
	backendResp := &NodeWorkflowOperationInstanceLogGetResp{
		Data: &NodeWorkflowOperationInstanceLogGetResp_Data{
			OperInstLogs: map[string]*WorkflowActionData{
				"install_pre_ordered_plugins": {
					DisplayNameZh: "安装预置插件",
					DisplayNameEn: "Install Pre-ordered Plugins",
					LifeCycle: &WorkflowLifeCycle{
						State:      "success",
						CreateTime: 1,
						StartTime:  2,
						EndTime:    3,
					},
					Message: &WorkflowActionMessage{Logs: nil},
					SubWorkflowRefs: []*SubWorkflowRef{{
						WorkflowId:     "plugin-wf-abc",
						WorkflowDomain: string(types.WorkflowDomainPlugin),
					}},
				},
			},
			ExtraExecutionLogs: &WorkflowActionMessage{Logs: nil},
		},
	}

	inst := backendResp.ConvertWorkflowOperationInstanceLogToTypes()
	require.NotNil(t, inst)
	act := inst.ActionInstanceDataMap["install_pre_ordered_plugins"]
	require.NotNil(t, act)
	assert.Equal(t, `[{"workflow_id":"plugin-wf-abc","workflow_domain":"plugin"}]`,
		act.PrivateData[types.PDKeySubWorkflowRefs])

	out := &NodeWorkflowOperationInstanceLogGetResp{}
	out.ConvertResultFromTypes(inst)
	require.NotNil(t, out.Data)
	actOut := out.Data.OperInstLogs["install_pre_ordered_plugins"]
	require.NotNil(t, actOut)
	require.Len(t, actOut.SubWorkflowRefs, 1)
	assert.Equal(t, "plugin-wf-abc", actOut.SubWorkflowRefs[0].WorkflowId)
	assert.Equal(t, string(types.WorkflowDomainPlugin), actOut.SubWorkflowRefs[0].WorkflowDomain)
}
