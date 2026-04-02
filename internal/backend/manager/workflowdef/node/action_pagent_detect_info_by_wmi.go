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
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	protoRelay "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/relay"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/relayhandler"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNamePagentDetectInfoByWMI defines the action name.
	ActionNamePagentDetectInfoByWMI = "pagent_detect_info_by_wmi"

	// queryClientTimeoutWMI defines the query client timeout for WMI action.
	queryClientTimeoutWMI = 30 * time.Second

	// waitForRelayReportTimeoutWMI defines the timeout for waiting relay report in WMI action.
	waitForRelayReportTimeoutWMI = 30 * time.Second

	// waitForRelayReportIntervalWMI defines the interval for checking relay report in WMI action.
	waitForRelayReportIntervalWMI = 3 * time.Second
)

// NewActionPagentDetectInfoByWMI get a new action.
func NewActionPagentDetectInfoByWMI(capability *Capability) action.Definition {
	return &actionPagentDetectInfoByWMI{
		storageHostCredit:     capability.StorageHostCredit,
		storageActionInstance: capability.StorageWorkflow,
		storageNodeDeployment: capability.StorageNode,
		storageHost:           capability.StorageTopo,
		storageRelease:        capability.StorageRelease,
		passwordVault:         capability.HostPasswordVault,
		proxyMessager:         capability.ProxyMessager,
	}
}

// ActParamPagentDetectInfoByWMI ...
type ActParamPagentDetectInfoByWMI struct {
	nodeUtils.NodeActionStandardParam `json:",inline"`
}

type actionPagentDetectInfoByWMI struct {
	storageNodeDeployment nodeStg.IDaoNodeDeployment
	storageHost           topoStg.IStorageHost
	storageActionInstance workflow.IStorageActionInstance
	storageRelease        release.IStorage
	storageHostCredit     credit.IStorageHostCredit

	passwordVault creditvault.IHostPasswordVault

	proxyMessager relayhandler.IServerMessager
}

// Name returns the name of the action.
func (act *actionPagentDetectInfoByWMI) Name() string {
	return ActionNamePagentDetectInfoByWMI
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionPagentDetectInfoByWMI) DisplayNameZh() string {
	return "通过 WMI 探测 P-Agent 信息"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionPagentDetectInfoByWMI) DisplayNameEn() string {
	return "Detect P-Agent Info via WMI"
}

// Version returns the version of the action.
func (act *actionPagentDetectInfoByWMI) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionPagentDetectInfoByWMI) Description() string {
	return "Let relay use WMI to connect to the target machine, detect the target machine's OS type and CPU architecture, " +
		"then use these info detect pkg version."
}

// Timeout returns the timeout of the action.
func (act *actionPagentDetectInfoByWMI) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionPagentDetectInfoByWMI) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionPagentDetectInfoByWMI) MaxRetryCount() uint {
	return 5 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionPagentDetectInfoByWMI) DelayFn() func() {
	return func() {
		time.Sleep(5 * time.Second) // nolint: mnd
	}
}

// Do this func define what the action will do.
// To ensure readability, this action uses fmt.Sprintf to concatenate characters.
// nolint: perfsprint,funlen,gocognit,nestif
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (act *actionPagentDetectInfoByWMI) Do(ctx *action.InstanceContext) error {
	param := new(ActParamPagentDetectInfoByWMI)
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

	// get wmi credit.
	credit := nodeUtils.NewCreditHandler(act.storageHostCredit, act.passwordVault)
	_, cKey, err := credit.GetWMICredit(std)
	if err != nil {
		return fmt.Errorf("failed to get wmi credit: %w", err)
	}

	// send detect info request to relay.
	if err := act.notifyRelayToDetect(std, cKey); err != nil {
		return err
	}

	// wait for relay report detect result.
	osType, cpuArch, err := act.waitForRelayReportDetect(std)
	if err != nil {
		return err
	}

	std.DeployInfo().Host.Dynamic.NodeOsType = osType
	std.DeployInfo().Host.Dynamic.NodeCPUArch = cpuArch

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
					Zh("用户选择, 使用目标版本. version(%s)", std.DeployInfo().Host.Dynamic.NodeVersion).
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

	return nil
}

func (act *actionPagentDetectInfoByWMI) notifyRelayToDetect(
	std *nodeUtils.NodeActionStandarder, cKey string) error {

	detectInfoEvent := protoRelay.DetectInfoByWMIReq{
		ActionName: std.InstanceData().Name,
		OperInstID: std.InstanceData().OperationInstanceID,
		IP:         std.DeployInfo().Host.Dynamic.LoginIP,
		Port:       std.DeployInfo().Host.Dynamic.LoginPort,
		User:       std.DeployInfo().Host.Dynamic.LoginUser,
		LoginMode:  string(std.DeployInfo().Host.Dynamic.LoginMode),
		Password:   cKey,
	}

	data, err := json.Marshal(detectInfoEvent)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	relayInfo, err := std.GetSelectedRelay()
	if err != nil {
		return fmt.Errorf("failed to get selected relay: %w", err)
	}

	errCh := act.proxyMessager.PushToClient(std.Context(),
		protoRelay.ServerPushEventTypeDetectInfoByWMI, data, relayInfo.AgentID)
	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("failed to detect info by wmi. agent-id(%s): %w", relayInfo.AgentID, err)
		}

		return nil
	case <-std.Context().Done():
		return fmt.Errorf("context cancelled. agent-id(%s): %w", relayInfo.AgentID, std.Context().Err())
	case <-time.After(queryClientTimeoutWMI):
		return fmt.Errorf("wait client timed out. agent-id(%s)", relayInfo.AgentID)
	}
}

// waitForRelayReportDetect wait for relay to report the detect result.
// nolint: gocognit
func (act *actionPagentDetectInfoByWMI) waitForRelayReportDetect(
	std *nodeUtils.NodeActionStandarder) (criteria.OSType, criteria.CPUArch, error) {

	timeoutCtx, cancel := contextx.WithTimeout(contextx.From(std.Context()), waitForRelayReportTimeoutWMI)
	defer cancel()

	ticker := time.NewTicker(waitForRelayReportIntervalWMI)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			return "", "", fmt.Errorf("wait for relay report detect result timed out. oper_inst_id(%s), action_name(%s)",
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
				return "", "", errors.New("unexpected type for relay detect result")
			}

			errMsgRaw := relayDetectResult[relayconstant.DetectResultErrMsgKey]
			errMsg, ok := errMsgRaw.(string)
			if !ok {
				return "", "", errors.New("unexpected type for relay detect result error message")
			}

			if errMsg != "" {
				return "", "", errors.New(errMsg)
			}

			osTypeStr, osTypeOk := relayDetectResult[relayconstant.DetectResultOsTypeKey].(string)
			cpuArchStr, cpuArchOk := relayDetectResult[relayconstant.DetectResultCPUArchKey].(string)

			if !osTypeOk || !cpuArchOk {
				return "", "", errors.New("incomplete relay detect result")
			}

			osType, err := platfmt.NormalizeOS(osTypeStr)
			if err != nil {
				return "", "", fmt.Errorf("failed to detect info: %w", err)
			}

			cpuArch, err := platfmt.NormalizeArch(cpuArchStr)
			if err != nil {
				return "", "", fmt.Errorf("failed to detect info: %w", err)
			}

			std.InstanceData().Log().
				Zh("等待 relay 报告探测结果成功, os-type(%s), cpu-arch(%s)", osType, cpuArch).
				En("wait for relay report detect result successfully, os-type(%s), cpu-arch(%s)", osType, cpuArch).
				Info()

			return osType, cpuArch, nil
		}
	}
}
