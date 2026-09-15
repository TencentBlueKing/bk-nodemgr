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

package v3_test

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func TestDeployPolicyWorkflowConditionsRoundTrip(t *testing.T) {
	t.Parallel()
	condition := &types.DeployPolicyWorkflowCondition{
		OperateTimeRange: &types.TimeRange{StartTime: time.Unix(100, 0), EndTime: time.Unix(200, 0)},
		ExactInclude: &types.DeployPolicyWorkflowExactConditions{
			WorkflowID:     []string{"parent"},
			DeployPolicyID: []int64{17},
			Status:         []types.DeployPolicyWorkflowStatus{types.DeployPolicyWorkflowStatusRunning},
			Operator:       []string{"operator"},
		},
	}
	req := new(protoBackend.DeployPolicyWorkflowListReq)
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		t.Fatal(err)
	}
	got, err := req.ConvertConditionsToTypes()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, condition) {
		t.Fatalf("conditions = %#v, want %#v", got, condition)
	}

	t.Run("unsupported exclusion", func(t *testing.T) {
		condition := &types.DeployPolicyWorkflowCondition{ExactExclude: &types.DeployPolicyWorkflowExactConditions{}}
		if err := req.ConvertConditionsFromTypes(condition); err != nil {
			return
		}
		t.Fatal("expected unsupported exclusion to fail")
	})
	t.Run("reversed time range", func(t *testing.T) {
		req.OperateTimeRange = &protoBackend.TimeRange{StartTimestampSec: 200, EndTimestampSec: 100}
		if _, err := req.ConvertConditionsToTypes(); err != nil {
			return
		}
		t.Fatal("expected reversed time range to fail")
	})
}

func TestDeployPolicyWorkflowResponseContract(t *testing.T) {
	t.Parallel()
	for _, finished := range []bool{false, true} {
		name := "running"
		if finished {
			name = "finished"
		}
		t.Run(name, func(t *testing.T) {
			workflow := &types.DeployPolicyWorkflow{
				WorkflowID: "parent", TriggerID: "trigger", DeployPolicyID: 17, Operator: "operator",
				OperateTime: time.UnixMilli(123456), Status: types.DeployPolicyWorkflowStatusRunning,
				Children: []types.DeployPolicyWorkflowChild{
					{WorkflowID: "node-child", WorkflowDomain: types.WorkflowDomainNode},
					{WorkflowID: "plugin-child", WorkflowDomain: types.WorkflowDomainPlugin},
				},
			}
			if finished {
				workflow.FinishTime = time.UnixMilli(234567)
				workflow.Status = types.DeployPolicyWorkflowStatusSuccess
			}
			resp := new(protoBackend.DeployPolicyWorkflowListResp)
			resp.ConvertDeployPolicyWorkflowsFromTypes(1, []*types.DeployPolicyWorkflow{workflow})
			items, total := resp.ConvertDeployPolicyWorkflowsToTypes()
			if total != 1 || !reflect.DeepEqual(items, []*types.DeployPolicyWorkflow{workflow}) {
				t.Fatalf("response round trip = %#v, total %d", items, total)
			}
		})
	}
	resp := new(protoBackend.DeployPolicyWorkflowListResp)
	resp.ConvertDeployPolicyWorkflowsFromTypes(7, nil)
	body, err := json.Marshal(resp.GetData())
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != `{"total":7,"items":[]}` {
		t.Fatalf("count response = %s", body)
	}
	execute := new(protoBackend.DeployPolicyExecuteResp)
	execute.ConvertWorkflowID("parent")
	body, err = json.Marshal(execute.GetData())
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != `{"workflow_id":"parent"}` {
		t.Fatalf("execute response = %s", body)
	}
}
