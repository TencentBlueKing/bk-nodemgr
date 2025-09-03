/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package nodeinstall

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/nodeinstall/utils"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/credit"
	nodedeployment "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node-deployment"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/release"
	"github.com/TencentBlueKing/bk-nodemgr/internal/relay/relayconstant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/creditvault"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/deployconstant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	protoRelay "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/relay"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/relayhandler"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/wmix"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNamePagentDetectInfoByWMI defines the action name.
	ActionNamePagentDetectInfoByWMI = "pagent_detect_info_by_wmi"
)

// NewActionPagentDetectInfoByWMI get a new action.
func NewActionPagentDetectInfoByWMI(
	logger logger.ILogger,

	storageActionInstance workflow.IStorageActionInstance,
	storageNodeDeployment nodedeployment.IStorageNodeDeployment,
	storageRelease release.IStorage,
	storageHostCredit credit.IStorageHostCredit,
	passwordVault creditvault.IHostPasswordVault,

	proxyMessager relayhandler.IServerMessager,
) action.Definition {

	return &actionPagentDetectInfoByWMI{
		logger: logger,

		storageHostCredit:     storageHostCredit,
		storageActionInstance: storageActionInstance,
		storageNodeDeployment: storageNodeDeployment,
		storageRelease:        storageRelease,

		passwordVault: passwordVault,

		proxyMessager: proxyMessager,
	}
}

// ActParamPagentDetectInfoByWMI ...
type ActParamPagentDetectInfoByWMI struct {
	Token    string `json:"token"`
	Operator string `json:"operator"`
}

type actionPagentDetectInfoByWMI struct {
	logger logger.ILogger

	storageNodeDeployment nodedeployment.IStorageNodeDeployment
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
// nolint: perfsprint,funlen,fnsize,gocognit,nestif
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (act *actionPagentDetectInfoByWMI) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActParamDetectInfoByWMI)
	err = conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		err = fmt.Errorf("failed to convert param: %w", err)

		return err
	}

	// initialize standard data.
	std := utils.NewNodeActionStandarder(act.storageNodeDeployment)
	if err = std.Initialize(ctx, param.NodeActionStandardParam); err != nil {
		return err
	}
	defer func() {
		if storeErr := std.Save(); storeErr != nil {
			err = errors.Join(storeErr, err)
		}
	}()

	// get wmi credit.
	credit := utils.NewCreditHandler(act.storageHostCredit, act.passwordVault)
	cMethod, cKey, err := credit.GetWMICredit(std)
	if err != nil {
		return fmt.Errorf("failed to get wmi credit: %w", err)
	}

	// send detect info request to relay.
	if err := act.detectInfo(ctx, std.DeployInfo(), cMethod, cKey, &std.DeployInfo().RelayInfo); err != nil {
		return err
	}
	ctx.Data.LogI(fmt.Sprintf("detect info by wmi send to relay.relay-host-id(%d)", std.DeployInfo().RelayInfo.HostID))

	// wait for relay report detect result.
	osType, cpuArch, err := act.waitForRelayReportDetect(ctx)
	if err != nil {
		return err
	}
	ctx.Data.LogI(fmt.Sprintf("detected os-type(%s), cpu-arch(%s)", osType, cpuArch))

	deployConstant, err := deployconstant.GetDeployConf(std.DeployInfo().Host.Dynamic.NodeGeneration, osType)
	if err != nil {
		return fmt.Errorf("failed to get deploy constant: %w", err)
	}

	// installer workdir priority: user specified in info > deploy constant default > connected dir.
	if std.DeployInfo().InstallerWorkDir == "" {
		std.DeployInfo().InstallerWorkDir = deployConstant.WorkDir
	}

	if std.DeployInfo().InstallerWorkDir == "" {
		std.DeployInfo().InstallerWorkDir = windowsDefaultInstallerWorkDir
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
				ctx.Data.LogI(fmt.Sprintf("user select, using target version. version(%s)", std.DeployInfo().Host.Dynamic.NodeVersion))

				break
			}
		}
	} else if std.DeployInfo().Host.Dynamic.NodeVersion == "" {
		// we'll automatically use the system information to select the default version,
		// when NodeVersion is empty.
		std.DeployInfo().Host.Dynamic.NodeVersion, err = autoSelectVersion(ctx.Ctx, CheckAndSelectVersionParam{
			daoRelease:  act.storageRelease,
			ReleaseType: releaseType,
			Generation:  std.DeployInfo().Host.Dynamic.NodeGeneration,
			OSType:      std.DeployInfo().Host.Dynamic.NodeOsType,
			CPUArch:     std.DeployInfo().Host.Dynamic.NodeCPUArch,
		})
		if err != nil {
			return err
		}
		ctx.Data.LogI(fmt.Sprintf("auto select, using system default version. version(%s)", std.DeployInfo().Host.Dynamic.NodeVersion))
	}

	err = checkVersionAvailability(
		ctx.Ctx, CheckAndSelectVersionParam{
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

func (act *actionPagentDetectInfoByWMI) detectInfo(ctx *action.InstanceContext,
	info *types.DeploymentInfo, cMethod wmix.AuthMethod, cKey string, relayHost *types.RelayInfo) error {

	detectInfoEvent := protoRelay.DetectInfoByWMIReq{
		ActionName: ctx.Data.Name,
		OperInstID: ctx.Data.OperationInstanceID,
		IP:         info.Host.Dynamic.LoginIP,
		Port:       info.Host.Dynamic.LoginPort,
		User:       info.Host.Dynamic.LoginUser,
		LoginMode:  string(cMethod),
		Password:   cKey,
	}

	data, err := json.Marshal(detectInfoEvent)
	if err != nil {
		act.logger.Errorf("failed to marshal data: %v", err)
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	errCh := act.proxyMessager.PushToClient(ctx.Ctx,
		protoRelay.ServerPushEventTypeDetectInfoByWMI, data, relayHost.AgentID)
	select {
	case err := <-errCh:
		if err != nil {
			act.logger.Errorf("detect info by WMI failed: %v", err)
			return fmt.Errorf("detect info by WMI failed: %w", err)
		}
	case <-time.After(queryClientTimeout):
		act.logger.Errorf("wait client timed out after (%s)", queryClientTimeout)
		return errors.New("wait client timed out")
	}

	return nil
}

// waitForRelayReportDetect wait for relay to report the detect result.
// nolint: gocognit
func (act *actionPagentDetectInfoByWMI) waitForRelayReportDetect(
	ctx *action.InstanceContext) (criteria.OSType, criteria.CPUArch, error) {

	timeoutCtx, cancel := context.WithTimeout(ctx.Ctx, waitForRelayReportTimeout)
	defer cancel()

	ticker := time.NewTicker(waitForRelayReportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			return "", "", fmt.Errorf("wait for relay report detect result timed out. oper_inst_id(%s), action_name(%s)",
				ctx.Data.OperationInstanceID, ctx.Data.Name)

		case <-ticker.C:
			privateData, err := act.storageActionInstance.GetActionInstancePrivateData(
				timeoutCtx, ctx.Data.OperationInstanceID, ctx.Data.Name)
			if err != nil {
				act.logger.Warnf("get private data failed, retrying. oper_inst_id(%s), action_name(%s): %v",
					ctx.Data.OperationInstanceID, ctx.Data.Name, err)

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

			osType, err := platform.NormalizeOS(osTypeStr)
			if err != nil {
				return "", "", fmt.Errorf("failed to detect info: %w", err)
			}

			cpuArch, err := platform.NormalizeArch(cpuArchStr)
			if err != nil {
				return "", "", fmt.Errorf("failed to detect info: %w", err)
			}

			return osType, cpuArch, nil
		}
	}
}
