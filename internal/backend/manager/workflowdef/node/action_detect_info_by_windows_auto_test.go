package node

import (
	"testing"
	"time"
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
	if act.Timeout() != 10*time.Minute {
		t.Fatalf("Timeout() = %s, want %s", act.Timeout(), 10*time.Minute)
	}
}
