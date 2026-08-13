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
	"strings"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

func TestFindPartialRetryStartIndex(t *testing.T) {
	actionNames := []string{"a1", "a2", "a3", "a4"}

	newAction := func(name string, index int, state action.State) *action.InstanceData {
		return &action.InstanceData{
			Name:      name,
			Index:     index,
			Lifecycle: &action.Lifecycle{State: state},
		}
	}

	newPrevInstance := func(state operation.State, actions map[string]*action.InstanceData) *operation.InstanceData {
		return &operation.InstanceData{
			InstanceBriefData: operation.InstanceBriefData{
				Metadata: &operation.InstanceMetadata{
					OperationInstanceID: "oper-inst:prev",
				},
				Lifecycle: &operation.Lifecycle{State: state},
			},
			ActionInstanceDataMap: actions,
		}
	}

	tests := []struct {
		name       string
		operState  operation.State
		actions    map[string]*action.InstanceData
		retryStart map[string]bool
		wantIndex  int
		wantErr    string
	}{
		{
			// operation-level failure (e.g. extra execution) before any action runs.
			name:      "operation failed with all actions pending falls back to 0",
			operState: operation.StateFailed,
			actions: map[string]*action.InstanceData{
				"a1": newAction("a1", 0, action.StatePending),
				"a2": newAction("a2", 1, action.StatePending),
				"a3": newAction("a3", 2, action.StatePending),
				"a4": newAction("a4", 3, action.StatePending),
			},
			wantIndex: 0,
		},
		{
			// regression: action-level failure still resolves a normal retry start point.
			name:      "operation failed with failed action resolves start point",
			operState: operation.StateFailed,
			actions: map[string]*action.InstanceData{
				"a1": newAction("a1", 0, action.StatePending),
				"a2": newAction("a2", 1, action.StateFailed),
				"a3": newAction("a3", 2, action.StatePending),
				"a4": newAction("a4", 3, action.StatePending),
			},
			retryStart: map[string]bool{"a2": true},
			wantIndex:  1,
		},
		{
			// operation-level failure after some actions already succeeded.
			name:      "operation failed with partial success falls back to 0",
			operState: operation.StateFailed,
			actions: map[string]*action.InstanceData{
				"a1": newAction("a1", 0, action.StateSuccess),
				"a2": newAction("a2", 1, action.StateSuccess),
				"a3": newAction("a3", 2, action.StatePending),
				"a4": newAction("a4", 3, action.StatePending),
			},
			wantIndex: 0,
		},
		{
			// regression: genuinely successful instance still errors.
			name:      "operation succeeded with all actions succeeded errors",
			operState: operation.StateSuccess,
			actions: map[string]*action.InstanceData{
				"a1": newAction("a1", 0, action.StateSuccess),
				"a2": newAction("a2", 1, action.StateSuccess),
				"a3": newAction("a3", 2, action.StateSuccess),
				"a4": newAction("a4", 3, action.StateSuccess),
			},
			wantErr: "all actions succeeded, nothing to retry",
		},
		{
			// regression: terminated action is a hard error and not retryable.
			name:      "operation failed with terminated action errors",
			operState: operation.StateFailed,
			actions: map[string]*action.InstanceData{
				"a1": newAction("a1", 0, action.StatePending),
				"a2": newAction("a2", 1, action.StateTerminated),
				"a3": newAction("a3", 2, action.StatePending),
				"a4": newAction("a4", 3, action.StatePending),
			},
			wantErr: "is terminated, can not partial retry",
		},
		{
			// timeout operation without any action anchor falls back to 0.
			name:      "operation timeout with no failed action falls back to 0",
			operState: operation.StateTimeout,
			actions: map[string]*action.InstanceData{
				"a1": newAction("a1", 0, action.StatePending),
				"a2": newAction("a2", 1, action.StatePending),
				"a3": newAction("a3", 2, action.StatePending),
				"a4": newAction("a4", 3, action.StatePending),
			},
			wantIndex: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctl := &controller{
				oper: &operation.Operation{
					InstanceIDs: []string{"oper-inst:prev"},
					Definition: &operation.DefinitionSnapshot{
						SnapshotName:           "test-operation",
						SnapshotActionDefNames: actionNames,
					},
					Param: operation.Param{
						RetryStartPoint: tt.retryStart,
					},
				},
			}

			gotIndex, err := ctl.findPartialRetryStartIndex(newPrevInstance(tt.operState, tt.actions))
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("findPartialRetryStartIndex() error = nil, want error containing %q", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("findPartialRetryStartIndex() error = %q, want containing %q", err.Error(), tt.wantErr)
				}

				return
			}

			if err != nil {
				t.Fatalf("findPartialRetryStartIndex() error = %v, want nil", err)
			}
			if gotIndex != tt.wantIndex {
				t.Fatalf("findPartialRetryStartIndex() index = %d, want %d", gotIndex, tt.wantIndex)
			}
		})
	}
}
