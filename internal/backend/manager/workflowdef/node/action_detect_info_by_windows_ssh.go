package node

import (
	"fmt"
	"strings"

	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	windowsSSHProfileCygwin = "ssh_cygwin"
	windowsSSHProfileNative = "ssh_native"

	windowsNativeArchCommand = "powershell -NoProfile -NonInteractive -ExecutionPolicy Bypass -Command \"$env:PROCESSOR_ARCHITECTURE\""
)

type windowsSSHDetectResult struct {
	osType  criteria.OSType
	cpuArch criteria.CPUArch
	profile string
}

type windowsSSHCommandRunner interface {
	RunCommand(string) (string, string, error)
}

func detectWindowsSSHInfo(data *action.InstanceData, runner windowsSSHCommandRunner) (windowsSSHDetectResult, error) {
	result := windowsSSHDetectResult{osType: criteria.OSWindows}

	osTypeStr, _, err := runner.RunCommand("uname -s")
	if err == nil && isCygwinUname(osTypeStr) {
		result.profile = windowsSSHProfileCygwin
		return detectWindowsSSHArch(data, runner, result, "uname -m")
	}

	result.profile = windowsSSHProfileNative
	return detectWindowsSSHArch(data, runner, result, windowsNativeArchCommand)
}

func detectWindowsSSHArch(
	data *action.InstanceData, runner windowsSSHCommandRunner, result windowsSSHDetectResult, command string,
) (windowsSSHDetectResult, error) {
	cpuArchStr, _, err := runner.RunCommand(command)
	if err != nil {
		return windowsSSHDetectResult{}, fmt.Errorf("failed to run %s: %w", command, err)
	}

	cpuArch, err := platfmt.NormalizeArch(strings.TrimSpace(cpuArchStr))
	if err != nil {
		return windowsSSHDetectResult{}, fmt.Errorf("failed to detect windows ssh cpu arch: %w", err)
	}
	result.cpuArch = cpuArch

	data.Log().
		Zh("主机操作系统类型(%s), Windows SSH Profile(%s)", result.osType, result.profile).
		En("host-os-type(%s), windows-ssh-profile(%s)", result.osType, result.profile).
		Info()
	data.Log().
		Zh("主机CPU架构(%s)", result.cpuArch).
		En("host-cpu-arch(%s)", result.cpuArch).
		Info()

	return result, nil
}

func isCygwinUname(osTypeStr string) bool {
	return strings.Contains(strings.ToUpper(strings.TrimSpace(osTypeStr)), "CYGWIN")
}
