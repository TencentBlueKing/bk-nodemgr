/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package node

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/deployconstant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func TestActionRenderNodeDeployment_ResolveNodeEventDataIDConf(t *testing.T) {
	defaultConf := deployconstant.NodeEventDataIDConf{
		AgentBaseAlarmEventDataID: 1000,
		TaskProcEventDataID:       2000,
	}

	tests := []struct {
		name            string
		deployInfo      *types.DeploymentInfo
		bizConf         *types.BizEventDataIDConf
		monitorDataID   int64
		monitorFound    bool
		expected        deployconstant.NodeEventDataIDConf
		expectedMonitor bool
		expectedWrite   bool
		expectedErrText string
	}{
		{
			name:       "uses local business config first",
			deployInfo: deploymentInfoWithBizID(10),
			bizConf: &types.BizEventDataIDConf{
				BizID:                     10,
				AgentBaseAlarmEventDataID: ptrInt64(3001),
				TaskProcEventDataID:       ptrInt64(3002),
			},
			expected: deployconstant.NodeEventDataIDConf{
				AgentBaseAlarmEventDataID: 3001,
				TaskProcEventDataID:       3002,
			},
		},
		{
			name:          "gets task process data id from monitor and writes it back",
			deployInfo:    deploymentInfoWithBizID(20),
			bizConf:       &types.BizEventDataIDConf{BizID: 20, AgentBaseAlarmEventDataID: ptrInt64(4001)},
			monitorDataID: 4002,
			monitorFound:  true,
			expected: deployconstant.NodeEventDataIDConf{
				AgentBaseAlarmEventDataID: 4001,
				TaskProcEventDataID:       4002,
			},
			expectedMonitor: true,
			expectedWrite:   true,
		},
		{
			name:            "uses defaults when monitor is disabled",
			deployInfo:      deploymentInfoWithBizID(30),
			monitorFound:    false,
			expected:        defaultConf,
			expectedMonitor: true,
		},
		{
			name:            "rejects missing business id",
			deployInfo:      deploymentInfoWithBizID(0),
			expectedErrText: "deployment biz id must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &fakeEventDataIDConfStorage{conf: tt.bizConf}
			monitor := &fakeMonitorHandler{dataID: tt.monitorDataID, found: tt.monitorFound}
			act := &actionRenderNodeDeployment{
				storageBizEventDataIDConf:  storage,
				monitorHandler:             monitor,
				defaultNodeEventDataIDConf: defaultConf,
			}

			conf, err := act.resolveNodeEventDataIDConf(contextx.Background(), tt.deployInfo)
			if tt.expectedErrText != "" {
				require.ErrorContains(t, err, tt.expectedErrText)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected, conf)
			require.Equal(t, tt.expectedMonitor, monitor.called)
			require.Equal(t, tt.expectedWrite, storage.updateCalled)
			if tt.expectedWrite {
				require.Equal(t, tt.deployInfo.Host.Static.BizID, storage.updatedBizID)
				require.Equal(t, tt.expected.TaskProcEventDataID, storage.updatedEventDataID)
			}
		})
	}
}

type fakeEventDataIDConfStorage struct {
	conf               *types.BizEventDataIDConf
	updateCalled       bool
	updatedBizID       int64
	updatedEventDataID int64
}

func (s *fakeEventDataIDConfStorage) Start(contextx.IContext) error {
	return nil
}

func (s *fakeEventDataIDConfStorage) CheckHealthz() error {
	return nil
}

func (s *fakeEventDataIDConfStorage) Terminate() error {
	return nil
}

func (s *fakeEventDataIDConfStorage) GetBizEventDataIDConf(
	contextx.IContext,
	int64,
) (types.BizEventDataIDConf, bool, error) {
	if s.conf == nil {
		return types.BizEventDataIDConf{}, false, nil
	}

	return *s.conf, true, nil
}

func (s *fakeEventDataIDConfStorage) UpdateBizTaskProcEventDataID(
	_ contextx.IContext,
	bkBizID int64,
	eventDataID int64,
) error {
	s.updateCalled = true
	s.updatedBizID = bkBizID
	s.updatedEventDataID = eventDataID
	return nil
}

type fakeMonitorHandler struct {
	dataID int64
	found  bool
	called bool
}

func (h *fakeMonitorHandler) GetOrCreateAgentEventDataID(contextx.IContext, int64) (int64, bool, error) {
	h.called = true
	return h.dataID, h.found, nil
}

func deploymentInfoWithBizID(bkBizID int64) *types.DeploymentInfo {
	return &types.DeploymentInfo{
		Host: types.Host{
			Static: &types.HostStatic{BizID: bkBizID},
		},
	}
}

func ptrInt64(value int64) *int64 {
	return &value
}
