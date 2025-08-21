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

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/credit"
	nodedeployment "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node-deployment"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/release"
	relayReportKey "github.com/TencentBlueKing/bk-nodemgr/internal/relay/constance"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/creditvault"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/deployconstant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	protoRelay "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/relay"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/relayhandler"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNamePagentDetectInfoBySSH defines the action name.
	ActionNamePagentDetectInfoBySSH = "pagent_detect_info_by_ssh"

	// queryClientTimeout defines the query client timeout.
	queryClientTimeout = 3 * time.Second

	waitForClientReportTimeout  = 10 * time.Second
	waitForClientReportInterval = 2 * time.Second
)

// NewActionPagentDetectInfoBySSH get a new action.
func NewActionPagentDetectInfoBySSH(
	logger logger.Logger,

	storageActionInstance workflow.IStorageActionInstance,
	storageNodeDeployment nodedeployment.IStorageNodeDeployment,
	storageRelease release.IStorage,
	storageHostCredit credit.IStorageHostCredit,
	passwordVault creditvault.IHostPasswordVault,

	proxyMessager relayhandler.IServerMessager,
) action.Definition {

	return &actionPagentDetectInfoBySSH{
		logger: logger,

		storageHostCredit:     storageHostCredit,
		storageActionInstance: storageActionInstance,
		storageNodeDeployment: storageNodeDeployment,
		storageRelease:        storageRelease,

		passwordVault: passwordVault,

		proxyMessager: proxyMessager,
	}
}

// ActParamPagentDetectInfoBySSH ...
type ActParamPagentDetectInfoBySSH struct {
	Token    string `json:"token"`
	Operator string `json:"operator"`
}

type actionPagentDetectInfoBySSH struct {
	logger logger.Logger

	storageNodeDeployment nodedeployment.IStorageNodeDeployment
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
		time.Sleep(5 * time.Second) // nolint: mnd
	}
}

// Do this func define what the action will do.
// To ensure readability, this action uses fmt.Sprintf to concatenate characters.
// nolint: perfsprint,funlen,fnsize,gocognit
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (act *actionPagentDetectInfoBySSH) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActParamDetectInfoBySSH)
	err = conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		err = fmt.Errorf("failed to convert param: %w", err)

		return err
	}

	info, err := act.storageNodeDeployment.GetInfo(ctx.Ctx, param.Token)
	if err != nil {
		return err
	}

	defer func() {
		if storeErr := act.storageNodeDeployment.UpdateInfo(ctx.Ctx, param.Token, info); storeErr != nil {
			err = errors.Join(storeErr, err)
		}
	}()

	password, err := act.QueryPassword(ctx.Ctx, param.Operator, act.storageHostCredit, act.passwordVault, info)
	if err != nil {
		return err
	}

	relayHost := &info.RelayInfo

	if err := act.detectInfo(ctx, info, password, relayHost); err != nil {
		return err
	}
	ctx.Data.LogI(fmt.Sprintf("detect info by ssh send to relay.relay-host-id(%d)", relayHost.HostID))

	osType, cpuArch, connectedDir, err := act.waitForRelayReportDetect(ctx)
	if err != nil {
		return err
	}
	ctx.Data.LogI(fmt.Sprintf("detected os-type(%s), cpu-arch(%s), connected-dir(%s)", osType, cpuArch, connectedDir))

	deployConstant, err := deployconstant.GetDeployConf(info.Host.Dynamic.NodeGeneration, osType)
	if err != nil {
		return fmt.Errorf("failed to get deploy constant: %w", err)
	}

	// installer workdir priority: user specified in info > deploy constant default > connected dir.
	if info.InstallerWorkDir == "" {
		info.InstallerWorkDir = deployConstant.WorkDir
	}

	// this is a fallback strategy, if system has no specified workdir, use connected dir.
	if info.InstallerWorkDir == "" {
		info.InstallerWorkDir = connectedDir
	}

	info.Host.Dynamic.NodeOsType = osType
	info.Host.Dynamic.NodeCPUArch = cpuArch

	releaseType, err := types.ConvertNodeRoleToReleaseType(info.Host.Dynamic.NodeRole)
	if err != nil {
		ctx.Data.LogE(fmt.Sprintf("failed to convert node role to release type: %v", err))
		return err
	}

	if len(info.TargetVersion) > 0 {
		for _, v := range info.TargetVersion {
			if info.Host.Dynamic.NodeOsType == v.OsType && info.Host.Dynamic.NodeCPUArch == v.CPUArch {
				// you can guarantee that there are no duplicates in the TargetVersion.
				info.Host.Dynamic.NodeVersion = v.Version
				ctx.Data.LogI(fmt.Sprintf("user select, using target version. version(%s)", info.Host.Dynamic.NodeVersion))
				break
			}
		}
	} else {
		// we'll automatically use the system information to select the default version,
		// when NodeVersion is empty.
		if info.Host.Dynamic.NodeVersion == "" {
			info.Host.Dynamic.NodeVersion, err = autoSelectVersion(ctx.Ctx, CheckAndSelectVersionParam{
				daoRelease:  act.storageRelease,
				ReleaseType: releaseType,
				Generation:  info.Host.Dynamic.NodeGeneration,
				OSType:      info.Host.Dynamic.NodeOsType,
				CPUArch:     info.Host.Dynamic.NodeCPUArch,
			})
			if err != nil {
				return err
			}
			ctx.Data.LogI(fmt.Sprintf("auto select, using system default version. version(%s)", info.Host.Dynamic.NodeVersion))
		}
	}

	err = checkVersionAvailability(
		ctx.Ctx, CheckAndSelectVersionParam{
			daoRelease:  act.storageRelease,
			ReleaseType: releaseType,
			Generation:  info.Host.Dynamic.NodeGeneration,
			OSType:      info.Host.Dynamic.NodeOsType,
			CPUArch:     info.Host.Dynamic.NodeCPUArch,
			Version:     info.Host.Dynamic.NodeVersion,
		})
	if err != nil {
		return err
	}

	return nil
}

func (act *actionPagentDetectInfoBySSH) detectInfo(ctx *action.InstanceContext,
	info *types.DeploymentInfo, password string, relayHost *types.RelayInfo) error {

	detectInfoEvent := protoRelay.DetectInfoBySSHReq{
		ActionName: ctx.Data.Name,
		OperInstID: ctx.Data.OperationInstanceID,
		IP:         info.LoginInfo.IP,
		Port:       info.LoginInfo.Port,
		User:       info.LoginInfo.User,
		LoginMode:  string(info.LoginInfo.Mode),
		Password:   password,
	}

	data, err := json.Marshal(detectInfoEvent)
	if err != nil {
		act.logger.Errorf("failed to marshal data: %v", err)
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	errCh := act.proxyMessager.PushToClient(ctx.Ctx,
		protoRelay.ServerPushEventTypeDetectInfoBySSH, data, relayHost.AgentID)
	select {
	case err := <-errCh:
		if err != nil {
			act.logger.Errorf("detect info by ssh failed: %v", err)
			return fmt.Errorf("detect info by ssh failed: %w", err)
		}
	case <-time.After(queryClientTimeout):
		act.logger.Errorf("wait client timed out after (%s)", queryClientTimeout)
		return errors.New("wait client timed out")
	}

	return nil
}

func (act *actionPagentDetectInfoBySSH) QueryPassword(
	ctx context.Context,
	operator string,
	storageHostCredit credit.IStorageHostCredit,
	passwordVault creditvault.IHostPasswordVault,
	info *types.DeploymentInfo) (string, error) {

	switch info.LoginInfo.Mode {
	case types.LoginModePassword:
		passwd, err := storageHostCredit.LoadHostCredit(
			ctx,
			info.Host.Static.NetworkAreaID,
			info.LoginInfo.IP,
			info.LoginInfo.User,
			types.LoginModePassword)
		if err != nil {
			return "", fmt.Errorf("failed to load password from storageHostCredit storage: %w", err)
		}

		return string(passwd), nil

	case types.LoginModeKeyFile:
		privateKey, err := storageHostCredit.LoadHostCredit(
			ctx,
			info.Host.Static.NetworkAreaID,
			info.LoginInfo.IP,
			info.LoginInfo.User,
			types.LoginModeKeyFile)
		if err != nil {
			return "", fmt.Errorf("failed to load private key from storageHostCredit storage: %w", err)
		}

		return string(privateKey), nil
	case types.LoginModePasswordVault:
		passwd, err := passwordVault.LoadPassword(
			ctx,
			operator,
			info.Host.Static.NetworkAreaID,
			info.LoginInfo.IP,
			info.LoginInfo.User)
		if err != nil {
			return "", fmt.Errorf("failed to load password from password vault: %w", err)
		}

		return string(passwd), nil
	default:
		return "", fmt.Errorf("unsupported login mode, mode(%s)", info.LoginInfo.Mode)
	}
}

func (act *actionPagentDetectInfoBySSH) waitForRelayReportDetect(
	ctx *action.InstanceContext) (criteria.OSType, criteria.CPUArch, string, error) {

	timeoutCtx, cancel := context.WithTimeout(ctx.Ctx, waitForRelayReportTimeout)
	defer cancel()

	ticker := time.NewTicker(waitForRelayReportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			return "", "", "", fmt.Errorf("wait for relay report detect result timed out. oper_inst_id(%s), action_name(%s)",
				ctx.Data.OperationInstanceID, ctx.Data.Name)

		case <-ticker.C:
			privateData, err := act.storageActionInstance.GetActionInstancePrivateData(
				timeoutCtx, ctx.Data.OperationInstanceID, ctx.Data.Name)
			if err != nil {
				act.logger.Warnf("get private data failed, retrying. oper_inst_id(%s), action_name(%s): %v",
					ctx.Data.OperationInstanceID, ctx.Data.Name, err)

				continue
			}

			relayDetectResultRaw, exists := privateData[relayReportKey.DetectResultKey]
			if !exists {
				continue
			}

			relayStorageResult, ok := relayDetectResultRaw.(map[string]any)
			if !ok {
				return "", "", "", errors.New("unexpected type for relay ")
			}

			errMsgRaw := relayStorageResult[relayReportKey.DetectResultErrMsgKey]
			errMsg, ok := errMsgRaw.(string)
			if !ok {
				return "", "", "", errors.New("unexpected type for file state")
			}

			if errMsg != "" {
				return "", "", "", errors.New(errMsg)
			}

			// TODO: must check.
			osTypeStr := relayStorageResult[relayReportKey.DetectResultOsTypeKey].(string)
			cpuArchStr := relayStorageResult[relayReportKey.DetectResultCPUArchKey].(string)
			connerctionDir := relayStorageResult[relayReportKey.DetectResultConnectionDirKey].(string)

			osType, err := platform.NormalizeOS(osTypeStr)
			if err != nil {
				return "", "", "", fmt.Errorf("failed to detect info, err: %w", err)
			}

			cpuArch, err := platform.NormalizeArch(cpuArchStr)
			if err != nil {
				return "", "", "", fmt.Errorf("failed to detect info, err: %w", err)
			}

			return osType, cpuArch, connerctionDir, nil
		}
	}
}
