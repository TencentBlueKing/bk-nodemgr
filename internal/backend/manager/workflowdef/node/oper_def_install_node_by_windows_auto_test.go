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
