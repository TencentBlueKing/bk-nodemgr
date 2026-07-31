package node

import (
	"errors"
	"testing"
	"time"
)

var errWindowsAutoTest = errors.New("windows auto test error")

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
	if act.Timeout() != 10*time.Minute {
		t.Fatalf("Timeout() = %s, want %s", act.Timeout(), 10*time.Minute)
	}
}

func TestActionInstallNodeByWindowsAutoMetadata(t *testing.T) {
	act := NewActionInstallNodeByWindowsAuto(&Capability{})

	if got := act.Name(); got != ActionNameInstallNodeByWindowsAuto {
		t.Fatalf("Name() = %q, want %q", got, ActionNameInstallNodeByWindowsAuto)
	}
	if ActionNameInstallNodeByWindowsAuto != "install_node_by_windows_auto" {
		t.Fatalf("ActionNameInstallNodeByWindowsAuto = %q, want install_node_by_windows_auto", ActionNameInstallNodeByWindowsAuto)
	}
	if act.DisplayNameZh() == "" || act.DisplayNameEn() == "" || act.Description() == "" {
		t.Fatalf("metadata must not be empty")
	}
	if act.Timeout() != 5*time.Minute {
		t.Fatalf("Timeout() = %s, want %s", act.Timeout(), 5*time.Minute)
	}
	if act.MaxRetryCount() != 0 {
		t.Fatalf("MaxRetryCount() = %d, want 0", act.MaxRetryCount())
	}
}

func TestWindowsAutoShouldFallback(t *testing.T) {
	tests := []struct {
		name           string
		sshLaunchReady bool
		sshErr         error
		want           bool
	}{
		{
			name:           "fallback before launch ready with ssh error",
			sshLaunchReady: false,
			sshErr:         errWindowsAutoTest,
			want:           true,
		},
		{
			name:           "no fallback after launch ready with ssh error",
			sshLaunchReady: true,
			sshErr:         errWindowsAutoTest,
			want:           false,
		},
		{
			name:           "no fallback without ssh error",
			sshLaunchReady: false,
			sshErr:         nil,
			want:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := windowsAutoShouldFallback(tt.sshLaunchReady, tt.sshErr); got != tt.want {
				t.Fatalf("windowsAutoShouldFallback() = %v, want %v", got, tt.want)
			}
		})
	}
}
