/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package node

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	nodeUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/node/utils"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/credit"
	nodeStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/release"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/internal/relay/relayconstant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/creditvault"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/deployconstant"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	protoRelay "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/relay"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/relayhandler"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/sshx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNamePagentDetectInfoBySSH defines the action name.
	ActionNamePagentDetectInfoBySSH = "pagent_detect_info_by_ssh"

	// queryClientTimeout defines the query client timeout.
	queryClientTimeout = 30 * time.Second

	// actionRetryDelay defines the delay before retrying the action when it fails.
	actionRetryDelay = 5 * time.Second
)

// NewActionPagentDetectInfoBySSH get a new action.
func NewActionPagentDetectInfoBySSH(capability *Capability) action.Definition {
	return &actionPagentDetectInfoBySSH{
		storageHostCredit:     capability.StorageHostCredit,
		storageActionInstance: capability.StorageWorkflow,
		storageNodeDeployment: capability.StorageNode,
		storageHost:           capability.StorageTopo,
		storageRelease:        capability.StorageRelease,

		passwordVault: capability.HostPasswordVault,

		proxyMessager: capability.ProxyMessager,
	}
}

// ActParamPagentDetectInfoBySSH ...
type ActParamPagentDetectInfoBySSH struct {
	nodeUtils.NodeActionStandardParam `json:",inline"`
}

type actionPagentDetectInfoBySSH struct {
	storageNodeDeployment nodeStg.IDaoNodeDeployment
	storageHost           topoStg.IStorageHost
	storageActionInstance workflow.IStorageActionInstance
	storageRelease        release.IStorage
	storageHostCredit     credit.IStorageHostCredit

	passwordVault creditvault.IHostPasswordVault

	proxyMessager relayhandler.IServerMessager
}

// Name returns the name of the action.
func (act *actionPagentDetectInfoBySSH) Name() string {
	return ActionNamePagentDetectInfoBySSH
}

// Version returns the version of the action.
func (act *actionPagentDetectInfoBySSH) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionPagentDetectInfoBySSH) Description() string {
	return "Let relay use ssh to connect to the target machine, detect the target machine's OS type and CPU architecture, " +
		"then use these info detect pkg version."
}

// Timeout returns the timeout of the action.
func (act *actionPagentDetectInfoBySSH) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionPagentDetectInfoBySSH) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionPagentDetectInfoBySSH) MaxRetryCount() uint {
	return 5 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionPagentDetectInfoBySSH) DelayFn() func() {
	return func() {
		time.Sleep(actionRetryDelay)
	}
}

// Do this func define what the action will do.
// To ensure readability, this action uses fmt.Sprintf to concatenate characters.
// nolint: perfsprint,funlen,gocognit,nestif
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (act *actionPagentDetectInfoBySSH) Do(ctx *action.InstanceContext) error {
	param := new(ActParamPagentDetectInfoBySSH)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	// initialize standard data.
	std := nodeUtils.NewNodeActionStandarder(act.storageNodeDeployment, act.storageHost)
	if err = std.Initialize(ctx, param.NodeActionStandardParam); err != nil {
		return err
	}
	defer func() {
		if storeErr := std.Save(); storeErr != nil {
			err = errors.Join(storeErr, err)
		}
	}()

	// get ssh credit.
	credit := nodeUtils.NewCreditHandler(act.storageHostCredit, act.passwordVault)
	cMethod, cKey, err := credit.GetSSHCredit(std)
	if err != nil {
		return fmt.Errorf("failed to get ssh credit: %w", err)
	}

	// send detect info request to relay.
	if err := act.notifyRelayToDetect(std, cMethod, cKey); err != nil {
		return err
	}

	// wait for relay report detect result.
	osType, cpuArch, connectedDir, err := act.waitForRelayReportDetect(std)
	if err != nil {
		return err
	}

	// get deploy constant.
	deployConstant, err := deployconstant.GetNodeDeployConf(std.DeployInfo().Host.Dynamic.NodeGeneration, osType)
	if err != nil {
		return fmt.Errorf("failed to get deploy constant: %w", err)
	}

	// installer workdir priority: user specified in info > deploy constant default > connected dir.
	if std.DeployInfo().InstallerWorkDir == "" {
		std.DeployInfo().InstallerWorkDir = deployConstant.WorkDir
	}

	// this is a fallback strategy, if system has no specified workdir, use connected dir.
	if std.DeployInfo().InstallerWorkDir == "" {
		std.DeployInfo().InstallerWorkDir = connectedDir
	}

	std.DeployInfo().Host.Dynamic.NodeOsType = osType
	std.DeployInfo().Host.Dynamic.NodeCPUArch = cpuArch

	releaseType, err := types.ConvertNodeRoleToReleaseType(std.DeployInfo().Host.Dynamic.NodeRole)
	if err != nil {
		std.InstanceData().LogE(fmt.Sprintf("failed to convert node role to release type: %v", err))
		return err
	}

	if len(std.DeployInfo().TargetVersion) > 0 {
		for _, v := range std.DeployInfo().TargetVersion {
			if std.DeployInfo().Host.Dynamic.NodeOsType == v.OsType && std.DeployInfo().Host.Dynamic.NodeCPUArch == v.CPUArch {
				// you can guarantee that there are no duplicates in the TargetVersion.
				std.DeployInfo().Host.Dynamic.NodeVersion = v.Version
				std.InstanceData().LogI(fmt.Sprintf("user select, using target version. version(%s)", std.DeployInfo().Host.Dynamic.NodeVersion))

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
		std.InstanceData().LogI(fmt.Sprintf("auto select, using system default version. version(%s)", std.DeployInfo().Host.Dynamic.NodeVersion))
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

	return nil
}

func (act *actionPagentDetectInfoBySSH) notifyRelayToDetect(
	std *nodeUtils.NodeActionStandarder, cMethod sshx.AuthMethod, cKey string) error {

	detectInfoEvent := protoRelay.DetectInfoBySSHReq{
		ActionName: std.InstanceData().Name,
		OperInstID: std.InstanceData().OperationInstanceID,
		IP:         std.DeployInfo().Host.Dynamic.LoginIP,
		Port:       std.DeployInfo().Host.Dynamic.LoginPort,
		User:       std.DeployInfo().Host.Dynamic.LoginUser,
		LoginMode:  string(cMethod),
		Password:   cKey,
	}

	data, err := json.Marshal(detectInfoEvent)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// Get multiple relay infos for retry
	relayInfos, err := std.GetRelayInfos()
	if err != nil {
		return err
	}
	if len(relayInfos) == 0 {
		return fmt.Errorf("no relay info selected")
	}

	// Try each relay sequentially until one succeeds
	return act.notifyRelayToDetectMultiRelay(std, data, relayInfos)
}

// notifyRelayToDetectSingle sends detect info to a single relay.
func (act *actionPagentDetectInfoBySSH) notifyRelayToDetectSingle(
	std *nodeUtils.NodeActionStandarder, data []byte, relayInfo *types.RelayInfo) error {

	if relayInfo == nil || relayInfo.AgentID == "" {
		return fmt.Errorf("relay info has no agent id")
	}

	errCh := act.proxyMessager.PushToClient(std.Context(),
		protoRelay.ServerPushEventTypeDetectInfoBySSH, data, relayInfo.AgentID)
	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("detect info by ssh failed. agent-id(%s): %w", relayInfo.AgentID, err)
		}

		return nil
	case <-time.After(queryClientTimeout):
		return fmt.Errorf("wait client timed out. agent-id(%s)", relayInfo.AgentID)
	}
}

// notifyRelayToDetectMultiRelay tries each relay sequentially until one succeeds.
func (act *actionPagentDetectInfoBySSH) notifyRelayToDetectMultiRelay(
	std *nodeUtils.NodeActionStandarder, data []byte, relayInfos []*types.RelayInfo) error {

	var lastErr error
	for i, relayInfo := range relayInfos {
		if relayInfo == nil || relayInfo.AgentID == "" {
			std.InstanceData().LogW(fmt.Sprintf("relay info at index %d has no agent id, trying next", i))
			continue
		}

		std.InstanceData().LogI(fmt.Sprintf("attempting to send detect info to relay, index(%d/%d), agent-id(%s)",
			i+1, len(relayInfos), relayInfo.AgentID))

		err := act.notifyRelayToDetectSingle(std, data, relayInfo)
		if err == nil {
			std.InstanceData().LogI(fmt.Sprintf("detect info by ssh sent to relay successfully, agent-id(%s)", relayInfo.AgentID))
			return nil
		}

		std.InstanceData().LogW(fmt.Sprintf("failed to send detect info to relay, index(%d/%d), agent-id(%s): %v",
			i+1, len(relayInfos), relayInfo.AgentID, err))
		lastErr = err
	}

	// All relays failed
	return fmt.Errorf("failed to send detect info to all relay(s). count(%d): %w", len(relayInfos), lastErr)
}

// nolint: gocognit
func (act *actionPagentDetectInfoBySSH) waitForRelayReportDetect(
	std *nodeUtils.NodeActionStandarder) (criteria.OSType, criteria.CPUArch, string, error) {

	timeoutCtx, cancel := contextx.WithTimeout(contextx.From(std.Context()), waitForRelayReportTimeout)
	defer cancel()

	ticker := time.NewTicker(waitForRelayReportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			return "", "", "", fmt.Errorf("wait for relay report detect result timed out. oper_inst_id(%s), action_name(%s)",
				std.InstanceData().OperationInstanceID, std.InstanceData().Name)

		case <-ticker.C:
			privateData, err := act.storageActionInstance.GetActionInstancePrivateData(
				timeoutCtx, std.InstanceData().OperationInstanceID, std.InstanceData().Name)
			if err != nil {
				continue
			}

			relayDetectResultRaw, exists := privateData[relayconstant.DetectResultKey]
			if !exists {
				continue
			}

			relayDetectResult, ok := relayDetectResultRaw.(map[string]any)
			if !ok {
				return "", "", "", errors.New("unexpected type for relay detect result")
			}

			errMsgRaw := relayDetectResult[relayconstant.DetectResultErrMsgKey]
			errMsg, ok := errMsgRaw.(string)
			if !ok {
				return "", "", "", errors.New("unexpected type for relay detect result error message")
			}

			if errMsg != "" {
				return "", "", "", errors.New(errMsg)
			}

			osTypeStr, osTypeOk := relayDetectResult[relayconstant.DetectResultOsTypeKey].(string)
			cpuArchStr, cpuArchOk := relayDetectResult[relayconstant.DetectResultCPUArchKey].(string)
			connectionDir, connerctionDirOk := relayDetectResult[relayconstant.DetectResultConnectionDirKey].(string)

			if !osTypeOk || !cpuArchOk || !connerctionDirOk {
				return "", "", "", errors.New("incomplete relay detect result")
			}

			osType, err := platfmt.NormalizeOS(osTypeStr)
			if err != nil {
				return "", "", "", fmt.Errorf("failed to detect info: %w", err)
			}

			cpuArch, err := platfmt.NormalizeArch(cpuArchStr)
			if err != nil {
				return "", "", "", fmt.Errorf("failed to detect info: %w", err)
			}

			std.InstanceData().LogI(fmt.Sprintf("wait for relay report detect result successfully. os-type(%s), cpu-arch(%s), connected-dir(%s)",
				osType, cpuArch, connectionDir))

			return osType, cpuArch, connectionDir, nil
		}
	}
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionPagentDetectInfoBySSH) DisplayNameZh() string { return act.Name() }

// DisplayNameEn returns the English display name of the action.
func (act *actionPagentDetectInfoBySSH) DisplayNameEn() string { return act.Name() }
