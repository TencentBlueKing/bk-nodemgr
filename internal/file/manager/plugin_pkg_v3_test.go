/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package manager

import (
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/stretchr/testify/require"
)

func TestBuildPluginV3PkgControllerDebugCmd(t *testing.T) {
	controller := buildPluginV3PkgController(platform.NewPlatform(criteria.OSLinux, criteria.CPUArchAmd64), &PluginV3Definition{
		Control: PluginV3Control{
			StartCmd: "start.sh",
			StopCmd:  "stop.sh",
			DebugCmd: "./bk-nodemgr-relay -c ./etc/bk-nodemgr-relay.conf",
		},
	})

	require.Equal(t, "bin/bk-nodemgr-relay -c ./etc/bk-nodemgr-relay.conf", controller.DebugCmd)
}

func TestBuildPluginV3PkgControllerEmptyDebugCmd(t *testing.T) {
	controller := buildPluginV3PkgController(platform.NewPlatform(criteria.OSLinux, criteria.CPUArchAmd64), &PluginV3Definition{})

	require.Empty(t, controller.DebugCmd)
}
