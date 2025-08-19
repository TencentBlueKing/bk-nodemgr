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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/creditvault"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/deployconstant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/relayhandler"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNamePagentDetectInfoBySSH defines the action name.
	ActionNamePagentDetectInfoByWMI = "pagent_detect_info_by_wmi"
)

// NewActionDetectInfoBySSH get a new action.
func NewActionPagentDetectInfoByWMI(
	logger logger.Logger,
	storageNodeDeployment nodedeployment.IStorageNodeDeployment,
	storageRelease release.IStorage,
	storageHostCredit credit.IStorageHostCredit,
	passwordVault creditvault.IHostPasswordVault,
	proxyMessager relayhandler.IServerMessager,
) action.Definition {

	return &actionPagentDetectInfoByWMI{
		logger: logger,

		storageHostCredit:     storageHostCredit,
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
	logger                logger.Logger
	storageHostCredit     credit.IStorageHostCredit
	storageNodeDeployment nodedeployment.IStorageNodeDeployment
	storageRelease        release.IStorage
	passwordVault         creditvault.IHostPasswordVault
	proxyMessager         relayhandler.IServerMessager
}

// Name returns the name of the action.
func (act *actionPagentDetectInfoByWMI) Name() string {
	return ActionNameDetectInfoBySSH
}

// Version returns the version of the action.
func (act *actionPagentDetectInfoByWMI) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionPagentDetectInfoByWMI) Description() string {
	return "Let relay use wmi to connect to the target machine, detect the target machine's OS type and CPU architecture, " +
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
// nolint: perfsprint,funlen,fnsize
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (act *actionPagentDetectInfoByWMI) Do(ctx *action.InstanceContext) (err error) {
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

	osType, cpuArch, connectedDir, err := act.detectInfo(ctx, client)
	if err != nil {
		return err
	}

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

func (act *actionPagentDetectInfoByWMI) detectInfo(ctx *action.InstanceContext,
	files []protoRelay.FileInfo, agentID string) error {

	checkPkgEvent := protoRelay.CheckPkgStateReq{
		ActionName: ctx.Data.Name,
		OperInstID: ctx.Data.OperationInstanceID,
		FileList:   files,
	}

	data, err := json.Marshal(checkPkgEvent)
	if err != nil {
		act.logger.Errorf("failed to marshal data: %v", err)
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	errCh := act.proxyMessager.PushToClient(ctx.Ctx,
		protoRelay.ServerPushEventTypeCheckPkgState, data, agentID)
	select {
	case err := <-errCh:
		if err != nil {
			act.logger.Errorf("query client pkg state failed: %v", err)
			return fmt.Errorf("query client pkg state failed: %w", err)
		}
	case <-time.After(QueryClientTimeout):
		act.logger.Errorf("wait client timed out after (%s)", QueryClientTimeout)
		return errors.New("wait client timed out")
	}

	return nil
}

// waitForPkgsState waits for relay client to report package states.
func (act *actionPagentDetectInfoByWMI) waitForPkgsState(ctx *action.InstanceContext, pkgNames []string) (
	map[string]bool, string, error) {

	timeoutCtx, cancel := context.WithTimeout(ctx.Ctx, waitForClientReportTimeout)
	defer cancel()

	results := make(map[string]bool, len(pkgNames))
	for _, name := range pkgNames {
		results[name] = false
	}
	completedCount := 0
	fileStorageDir := ""

	ticker := time.NewTicker(waitForClientReportInterval)
	defer ticker.Stop()

	for completedCount < len(pkgNames) {
		select {
		case <-timeoutCtx.Done():
			return results, fileStorageDir, fmt.Errorf("timeout waiting for pkg states. oper_inst_id(%s), action_name(%s)",
				ctx.Data.OperationInstanceID, ctx.Data.Name)
		case <-ticker.C:
			privateDataMap, err := act.storageActionInstance.GetActionInstancePrivateData(timeoutCtx,
				ctx.Data.OperationInstanceID, ctx.Data.Name)
			if err != nil {
				act.logger.Warnf("get private data failed, retrying. oper_inst_id(%s), action_name(%s): %v",
					ctx.Data.OperationInstanceID, ctx.Data.Name, err)

				continue
			}

			fileStateRaw, exists := privateDataMap[relayFileStateKey]
			if !exists {
				continue
			}

			fileState, ok := fileStateRaw.(map[string]any)
			if !ok {
				return results, fileStorageDir, errors.New("unexpected type for file state")
			}

		}
	}

	return results, fileStorageDir, nil
}
