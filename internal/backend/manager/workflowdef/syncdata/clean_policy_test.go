/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package syncdata

import (
	"math"
	"testing"
	"time"
)

func TestSyncDataCleanPolicyMaxDaysUsesFractionalDays(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		timeout time.Duration
	}{
		{name: "sync agent info", timeout: OperDefNameSyncAgentInfoTimeout},
		{name: "sync agent state", timeout: OperDefNameSyncAgentStateTimeout},
		{name: "sync alive plugin process info", timeout: OperDefNameSyncAlivePluginProcessInfoTimeout},
		{name: "sync host", timeout: OperDefNameSyncHostTimeout},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			maxDays := syncDataCleanPolicyMaxDays(test.timeout)
			if maxDays <= 0 {
				t.Fatalf("expected positive clean policy max days, got %f", maxDays)
			}

			if maxDays >= 1 {
				t.Fatalf("expected sub-day clean policy max days, got %f", maxDays)
			}

			expectedMaxDays := test.timeout.Hours() * syncDataCleanPolicyTimeoutMultiplier / syncDataCleanPolicyHoursPerDay
			if math.Abs(maxDays-expectedMaxDays) > 0.000001 {
				t.Fatalf("expected clean policy max days %f, got %f", expectedMaxDays, maxDays)
			}
		})
	}
}

func TestSyncDataOperationCleanPolicyTimeouts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		operation  string
		maxDays    float64
		wantMaxDay float64
	}{
		{
			name:       "sync agent info",
			operation:  OperDefNameSyncAgentInfo,
			maxDays:    syncDataCleanPolicyMaxDays(OperDefNameSyncAgentInfoTimeout),
			wantMaxDay: syncDataCleanPolicyMaxDays(NewOperSyncAgentInfo(OperParamSyncAgentInfo{}).DefaultParameters().Timeout),
		},
		{
			name:       "sync agent state",
			operation:  OperDefNameSyncAgentState,
			maxDays:    syncDataCleanPolicyMaxDays(OperDefNameSyncAgentStateTimeout),
			wantMaxDay: syncDataCleanPolicyMaxDays(NewOperSyncAgentState(OperParamSyncAgentState{}).DefaultParameters().Timeout),
		},
		{
			name:       "sync alive plugin process info",
			operation:  OperDefNameSyncAlivePluginProcessInfo,
			maxDays:    syncDataCleanPolicyMaxDays(OperDefNameSyncAlivePluginProcessInfoTimeout),
			wantMaxDay: syncDataCleanPolicyMaxDays(NewOperSyncAlivePluginProcessInfo(OperParamSyncAlivePluginProcessInfo{}).DefaultParameters().Timeout),
		},
		{
			name:       "sync host",
			operation:  OperDefNameSyncHost,
			maxDays:    syncDataCleanPolicyMaxDays(OperDefNameSyncHostTimeout),
			wantMaxDay: syncDataCleanPolicyMaxDays(NewOperSyncHost(OperParamSyncHost{}).DefaultParameters().Timeout),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if math.Abs(test.maxDays-test.wantMaxDay) > 0.000001 {
				t.Fatalf("expected clean policy for operation %s to be %f, got %f",
					test.operation, test.wantMaxDay, test.maxDays)
			}
		})
	}
}
