/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package deployconstant

import (
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNodeDeployConf_GetEventDataIDConf(t *testing.T) {
	conf := NodeDeployConf{
		EventDataIDConfs: map[string]NodeEventDataIDConf{
			"tenant-a": {
				AgentBaseAlarmEventDataID: 1001,
				TaskProcEventDataID:       1002,
			},
		},
	}

	got, err := conf.GetEventDataIDConf("tenant-a")
	require.NoError(t, err)
	assert.Equal(t, int64(1001), got.AgentBaseAlarmEventDataID)
	assert.Equal(t, int64(1002), got.TaskProcEventDataID)

	_, err = conf.GetEventDataIDConf("tenant-b")
	assert.ErrorContains(t, err, "event data-id conf not found for tenant")

	_, err = (NodeDeployConf{}).GetEventDataIDConf("tenant-a")
	assert.ErrorContains(t, err, "event data-id conf not found for tenant")
}

func TestNodeDeployConf_ValidateEventDataIDConfs(t *testing.T) {
	conf := NodeDeployConf{
		DeployConf: DeployConf{
			Generation:    types.Generation2,
			OsType:        criteria.OSLinux,
			BaseDeployDir: "/usr/local/",
			BaseWorkDir:   "/tmp/bknm/",
		},
		EventDataIDConfs: map[string]NodeEventDataIDConf{
			"tenant-a": {
				AgentBaseAlarmEventDataID: 1001,
				TaskProcEventDataID:       1002,
			},
		},
	}
	require.NoError(t, conf.Validate())

	conf.EventDataIDConfs["tenant-a"] = NodeEventDataIDConf{
		AgentBaseAlarmEventDataID: 0,
		TaskProcEventDataID:       1002,
	}
	assert.ErrorContains(t, conf.Validate(), "invalid event data-id conf for tenantID tenant-a")
}
