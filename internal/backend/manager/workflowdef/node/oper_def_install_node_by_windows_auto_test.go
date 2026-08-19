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

package node

import (
	"reflect"
	"testing"
	"time"
)

func TestOperInstallNodeByWindowsAutoMetadata(t *testing.T) {
	oper := NewOperInstallNodeByWindowsAuto(OperParamInstallNodeByWindowsAuto{
		Token:    "test-token",
		Operator: "tester",
	})

	if got := oper.Name(); got != OperDefNameInstallNodeByWindowsAuto {
		t.Fatalf("Name() = %q, want %q", got, OperDefNameInstallNodeByWindowsAuto)
	}

	wantActions := []string{
		ActionNameTryReuseAgentID,
		ActionNameUpsertHostToCMDB,
		ActionNameDetectInfoByWindowsAuto,
		ActionNameInjectNodeCustomDeployConfig,
		ActionNameRenderNodeDeployment,
		ActionNameInstallNodeByWindowsAuto,
		ActionNameWaitInstallerComplete,
		ActionNameWaitGseReady,
		ActionNameSyncNodeInfo,
		ActionNameBindAgentHostRel,
		ActionNamePushHostIdentifier,
		ActionNameUpdateHost,
		ActionNameInstallPreOrderedPlugins,
	}
	if got := oper.ActionDefNames(); !reflect.DeepEqual(got, wantActions) {
		t.Fatalf("ActionDefNames() = %+v, want %+v", got, wantActions)
	}

	params := oper.DefaultParameters()
	if params.Timeout != 10*time.Minute {
		t.Fatalf("Timeout = %s, want %s", params.Timeout, 10*time.Minute)
	}
	if params.RetryStartPoint[ActionNameWaitInstallerComplete] || params.RetryStartPoint[ActionNameWaitGseReady] {
		t.Fatalf("wait actions must not be retry start points: %+v", params.RetryStartPoint)
	}
	if !params.RetryStartPoint[ActionNameInstallNodeByWindowsAuto] {
		t.Fatalf("install action must be a retry start point: %+v", params.RetryStartPoint)
	}
	if got := oper.ExtraExecutionName(); got != OperExtraExecutionNameLockAndUnlockHost {
		t.Fatalf("ExtraExecutionName() = %q, want %q", got, OperExtraExecutionNameLockAndUnlockHost)
	}
}
