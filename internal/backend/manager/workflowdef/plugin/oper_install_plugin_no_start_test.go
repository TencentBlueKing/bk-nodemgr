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

package plugin

import (
	"reflect"
	"testing"
	"time"

	pluginUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/plugin/utils"
)

func TestOperInstallPluginNoStartMetadata(t *testing.T) {
	oper := NewOperInstallPluginNoStart(OperParamInstallPluginNoStart{
		PluginActionStandardParam: pluginUtils.PluginActionStandardParam{
			Token:    "test-token",
			TenantID: "test-tenant",
			Operator: "tester",
		},
	})

	if got := oper.Name(); got != OperDefNameInstallPluginNoStart {
		t.Fatalf("Name() = %q, want %q", got, OperDefNameInstallPluginNoStart)
	}

	wantActions := []string{
		ActionNameTryStopProcess,
		ActionNameUpsertProcess,
		ActionNameVerifyPluginAvailability,
		ActionNameInjectPluginCustomDeployConfig,
		ActionNameInjectMainConfigDetail,
		ActionNameRenderPluginDeployment,
		ActionNameAllocatePluginBindAddress,
		ActionNameEnsureAndUpdatePluginConfigDetails,
		ActionNameRenderPluginConfig,
		ActionNameTransferPluginPkgToNode,
		ActionNameInstallPlugin,
		ActionNameWaitPluginInstallerComplete,
		ActionNameMarkPluginProcessStopped,
		ActionNameUpdateProcess,
	}

	if got := oper.ActionDefNames(); !reflect.DeepEqual(got, wantActions) {
		t.Fatalf("ActionDefNames() = %+v, want %+v", got, wantActions)
	}

	// the operation only delivers the plugin, so it must not contain the actions that start the
	// process or handle the plugin compatibility.
	forbiddenActions := map[string]struct{}{
		ActionNameCheckV2PluginNoAlive:                  {},
		ActionNameOverwritePluginConfigForCompatibility: {},
		ActionNameStartProcess:                          {},
		ActionNameCheckPluginProcessAlive:               {},
	}
	for _, actionName := range oper.ActionDefNames() {
		if _, ok := forbiddenActions[actionName]; ok {
			t.Fatalf("action(%s) must not be in the operation(install_plugin_no_start)", actionName)
		}
	}

	params := oper.DefaultParameters()
	if params.Timeout != 10*time.Minute {
		t.Fatalf("Timeout = %s, want %s", params.Timeout, 10*time.Minute)
	}

	if params.RetryStartPoint[ActionNameWaitPluginInstallerComplete] {
		t.Fatalf("wait action must not be a retry start point: %+v", params.RetryStartPoint)
	}

	if !params.RetryStartPoint[ActionNameMarkPluginProcessStopped] {
		t.Fatalf("mark stopped action must be a retry start point: %+v", params.RetryStartPoint)
	}

	for _, actionName := range wantActions {
		if _, ok := params.RetryStartPoint[actionName]; !ok {
			t.Fatalf("action(%s) must be configured in retry start point: %+v", actionName, params.RetryStartPoint)
		}
	}

	if got := oper.ExtraExecutionName(); got != "" {
		t.Fatalf("ExtraExecutionName() = %q, want empty", got)
	}
}
