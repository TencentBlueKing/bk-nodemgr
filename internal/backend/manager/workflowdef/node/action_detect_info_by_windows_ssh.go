package node

import (
	"errors"
	"fmt"
	"strings"
	"time"

	nodeUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/node/utils"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/credit"
	nodeStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/release"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	workflowStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/creditvault"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/sshx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameDetectInfoByWindowsSSH defines the action name.
	ActionNameDetectInfoByWindowsSSH = "detect_info_by_windows_ssh"

	windowsSSHProfileCygwin = "ssh_cygwin"
	windowsSSHProfileNative = "ssh_native"

	windowsNativeArchCommand = "powershell -NoProfile -NonInteractive -ExecutionPolicy Bypass -Command \"$env:PROCESSOR_ARCHITECTURE\""
)

// NewActionDetectInfoByWindowsSSH get a new action.
func NewActionDetectInfoByWindowsSSH(capability *Capability) action.Definition {
	return &actionDetectInfoByWindowsSSH{
		storageHostCredit:     capability.StorageHostCredit,
		storageNodeDeployment: capability.StorageNode,
		storageHost:           capability.StorageTopo,
		storageRelease:        capability.StorageRelease,
		storageActionInstance: capability.StorageWorkflow,
		passwordVault:         capability.HostPasswordVault,
	}
}

// ActParamDetectInfoByWindowsSSH defines the action parameter.
type ActParamDetectInfoByWindowsSSH struct {
	nodeUtils.NodeActionStandardParam `json:",inline"`
}

type actionDetectInfoByWindowsSSH struct {
	storageHostCredit     credit.IStorageHostCredit
	storageNodeDeployment nodeStg.IDaoNodeDeployment
	storageHost           topoStg.IStorageHost
	storageRelease        release.IStorage
	storageActionInstance workflowStg.IStorage
	passwordVault         creditvault.IHostPasswordVault
}

// Name returns the name of the action.
func (act *actionDetectInfoByWindowsSSH) Name() string {
	return ActionNameDetectInfoByWindowsSSH
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionDetectInfoByWindowsSSH) DisplayNameZh() string {
	return "通过 Windows SSH 探测主机信息"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionDetectInfoByWindowsSSH) DisplayNameEn() string {
	return "Detect Host Info via Windows SSH"
}

// Version returns the version of the action.
func (act *actionDetectInfoByWindowsSSH) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionDetectInfoByWindowsSSH) Description() string {
	return "Use Windows SSH to connect to the target machine, detect host info, and select package version."
}

// Timeout returns the timeout of the action.
func (act *actionDetectInfoByWindowsSSH) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionDetectInfoByWindowsSSH) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionDetectInfoByWindowsSSH) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionDetectInfoByWindowsSSH) DelayFn() func() {
	return func() {
		time.Sleep(5 * time.Second) // nolint: mnd
	}
}

// Do this func define what the action will do.
// nolint: perfsprint,funlen,gocognit
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (act *actionDetectInfoByWindowsSSH) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActParamDetectInfoByWindowsSSH)
	err = conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return fmt.Errorf("failed to convert param: %w", err)
	}

	std := nodeUtils.NewNodeActionStandarder(act.storageNodeDeployment, act.storageHost)
	if err = std.Initialize(ctx, param.NodeActionStandardParam); err != nil {
		return err
	}
	defer func() {
		if storeErr := std.Save(); storeErr != nil {
			err = errors.Join(storeErr, err)
		}
	}()

	credit := nodeUtils.NewCreditHandler(act.storageHostCredit, act.passwordVault)
	cMethod, cKey, err := credit.GetSSHCredit(std)
	if err != nil {
		return fmt.Errorf("failed to get ssh credit: %w", err)
	}

	std.InstanceData().Log().
		Zh("凭证获取成功, 认证方式(%s), 凭证长度(%d)", cMethod, len(cKey)).
		En("credit loaded, auth-method(%s), credential-length(%d)", cMethod, len(cKey)).
		Info()

	client, err := sshx.NewClient(std.Context(), &sshx.Config{
		Network:    sshx.NetworkTCP,
		IP:         std.DeployInfo().Host.Dynamic.LoginIP,
		Port:       int(std.DeployInfo().Host.Dynamic.LoginPort),
		User:       std.DeployInfo().Host.Dynamic.LoginUser,
		AuthMethod: cMethod,
		Password: func() string {
			if cMethod == sshx.AuthMethodPassword {
				return cKey
			}

			return ""
		}(),
		PrivateKey: func() []byte {
			if cMethod == sshx.AuthMethodPrivateKey {
				return []byte(cKey)
			}

			return nil
		}(),
	}, sshx.DefaultTimeout)
	if err != nil {
		return fmt.Errorf("failed to generate new ssh client: %w", err)
	}
	defer func() {
		_ = client.Close()
	}()

	result, err := detectWindowsSSHInfo(std.InstanceData(), client)
	if err != nil {
		return err
	}

	std.DeployInfo().Host.Dynamic.NodeOsType = result.osType
	std.DeployInfo().Host.Dynamic.NodeCPUArch = result.cpuArch

	releaseType, err := types.ConvertNodeRoleToReleaseType(std.DeployInfo().Host.Dynamic.NodeRole)
	if err != nil {
		return err
	}
	if len(std.DeployInfo().TargetVersion) > 0 {
		for _, v := range std.DeployInfo().TargetVersion {
			if std.DeployInfo().Host.Dynamic.NodeOsType == v.OsType && std.DeployInfo().Host.Dynamic.NodeCPUArch == v.CPUArch {
				// you can guarantee that there are no duplicates in the TargetVersion.
				std.DeployInfo().Host.Dynamic.NodeVersion = v.Version
				std.InstanceData().Log().
					Zh("用户选择, 使用目标版本 version(%s)", std.DeployInfo().Host.Dynamic.NodeVersion).
					En("user select, using target version. version(%s)", std.DeployInfo().Host.Dynamic.NodeVersion).
					Info()

				break
			}
		}
	} else if std.DeployInfo().Host.Dynamic.NodeVersion == "" {
		// we'll automatically use the system information to select the default version,
		// when NodeVersion is empty.
		std.DeployInfo().Host.Dynamic.NodeVersion, err = autoSelectVersion(std.Context(), CheckAndSelectVersionParam{
			daoRelease:  act.storageRelease,
			ReleaseType: releaseType,
			Generation:  std.DeployInfo().Host.Dynamic.NodeGeneration,
			OSType:      std.DeployInfo().Host.Dynamic.NodeOsType,
			CPUArch:     std.DeployInfo().Host.Dynamic.NodeCPUArch,
		})
		if err != nil {
			return err
		}
		std.InstanceData().Log().
			Zh("自动选择, 使用系统默认版本. version(%s)", std.DeployInfo().Host.Dynamic.NodeVersion).
			En("auto select, using system default version. version(%s)", std.DeployInfo().Host.Dynamic.NodeVersion).
			Info()
	}

	err = checkVersionAvailability(
		std.Context(), CheckAndSelectVersionParam{
			daoRelease:  act.storageRelease,
			ReleaseType: releaseType,
			Generation:  std.DeployInfo().Host.Dynamic.NodeGeneration,
			OSType:      std.DeployInfo().Host.Dynamic.NodeOsType,
			CPUArch:     std.DeployInfo().Host.Dynamic.NodeCPUArch,
			Version:     std.DeployInfo().Host.Dynamic.NodeVersion,
		})
	if err != nil {
		return err
	}

	if err = act.storageActionInstance.UpsertActionInstancePrivateData(
		std.Context(),
		std.InstanceData().OperationInstanceID,
		ActionNameDetectInfoByWindowsSSH,
		map[string]any{types.PDKeyWindowsSSHProfile: result.profile},
	); err != nil {
		return fmt.Errorf("failed to save windows ssh profile to private data: %w", err)
	}

	return nil
}

type windowsSSHDetectResult struct {
	osType  criteria.OSType
	cpuArch criteria.CPUArch
	profile string
}

type windowsSSHCommandRunner interface {
	RunCommand(cmd string) (string, string, error)
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
