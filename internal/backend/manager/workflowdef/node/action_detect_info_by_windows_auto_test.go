package node

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
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

func TestDetectWindowsAutoInfo(t *testing.T) {
	sshErr := errors.New("ssh unavailable")
	wmiErr := errors.New("wmi unavailable")

	tests := []struct {
		name       string
		sshResult  windowsSSHDetectResult
		sshErr     error
		wmiResult  windowsWMIDetectResult
		wmiErr     error
		want       windowsAutoDetectResult
		wantErr    bool
		wantSSHRun int
		wantWMIRun int
	}{
		{
			name: "ssh success skips wmi",
			sshResult: windowsSSHDetectResult{
				osType:  criteria.OSWindows,
				cpuArch: criteria.CPUArchAmd64,
				profile: windowsSSHProfileNative,
			},
			want: windowsAutoDetectResult{
				osType:  criteria.OSWindows,
				cpuArch: criteria.CPUArchAmd64,
				method:  windowsAutoDetectMethodSSH,
				profile: windowsSSHProfileNative,
			},
			wantSSHRun: 1,
		},
		{
			name:   "ssh failure falls back to wmi",
			sshErr: sshErr,
			wmiResult: windowsWMIDetectResult{
				osType:  criteria.OSWindows,
				cpuArch: criteria.CPUArchAmd64,
			},
			want: windowsAutoDetectResult{
				osType:  criteria.OSWindows,
				cpuArch: criteria.CPUArchAmd64,
				method:  windowsAutoDetectMethodWMI,
				sshErr:  sshErr,
			},
			wantSSHRun: 1,
			wantWMIRun: 1,
		},
		{
			name:       "both channels fail",
			sshErr:     sshErr,
			wmiErr:     wmiErr,
			wantErr:    true,
			wantSSHRun: 1,
			wantWMIRun: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sshRun := 0
			wmiRun := 0

			got, err := detectWindowsAutoInfo(
				func() (windowsSSHDetectResult, error) {
					sshRun++
					return tt.sshResult, tt.sshErr
				},
				func() (windowsWMIDetectResult, error) {
					wmiRun++
					return tt.wmiResult, tt.wmiErr
				},
			)
			if (err != nil) != tt.wantErr {
				t.Fatalf("detectWindowsAutoInfo() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if got.osType != tt.want.osType || got.cpuArch != tt.want.cpuArch || got.method != tt.want.method || got.profile != tt.want.profile || got.sshErr != tt.want.sshErr {
					t.Fatalf("detectWindowsAutoInfo() = %+v, want %+v", got, tt.want)
				}
			}
			if sshRun != tt.wantSSHRun || wmiRun != tt.wantWMIRun {
				t.Fatalf("probe runs = ssh:%d wmi:%d, want ssh:%d wmi:%d", sshRun, wmiRun, tt.wantSSHRun, tt.wantWMIRun)
			}
		})
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
