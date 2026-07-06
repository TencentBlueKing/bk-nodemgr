package node

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func TestActionDetectInfoByWindowsAutoMetadata(t *testing.T) {
	act := NewActionDetectInfoByWindowsAuto(&Capability{})

	if got := act.Name(); got != ActionNameDetectInfoByWindowsAuto {
		t.Fatalf("Name() = %q, want %q", got, ActionNameDetectInfoByWindowsAuto)
	}
	if ActionNameDetectInfoByWindowsAuto != "detect_info_by_windows_auto" {
		t.Fatalf("ActionNameDetectInfoByWindowsAuto = %q, want detect_info_by_windows_auto", ActionNameDetectInfoByWindowsAuto)
	}
	if act.DisplayNameZh() == "" || act.DisplayNameEn() == "" || act.Description() == "" {
		t.Fatalf("metadata must not be empty")
	}
	if act.Timeout() != time.Minute {
		t.Fatalf("Timeout() = %s, want %s", act.Timeout(), time.Minute)
	}
}

func TestBuildWindowsAutoDetectPrivateData(t *testing.T) {
	sshErr := errors.New("ssh unavailable")
	tests := []struct {
		name   string
		result windowsAutoDetectResult
		want   map[string]any
	}{
		{
			name: "ssh success stores method and profile",
			result: windowsAutoDetectResult{
				method:  windowsAutoDetectMethodSSH,
				profile: windowsSSHProfileNative,
			},
			want: map[string]any{
				types.PDKeyWindowsAutoDetectMethod: string(windowsAutoDetectMethodSSH),
				types.PDKeyWindowsSSHProfile:       windowsSSHProfileNative,
			},
		},
		{
			name: "wmi fallback stores method and ssh error",
			result: windowsAutoDetectResult{
				method: windowsAutoDetectMethodWMI,
				sshErr: sshErr,
			},
			want: map[string]any{
				types.PDKeyWindowsAutoDetectMethod:   string(windowsAutoDetectMethodWMI),
				types.PDKeyWindowsAutoSSHDetectError: sshErr.Error(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildWindowsAutoDetectPrivateData(tt.result)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("buildWindowsAutoDetectPrivateData() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
