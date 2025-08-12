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

	nodedeployment "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node-deployment"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/release"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/tool"
	protoRelay "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/relay"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/relayhandler"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/file"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameEnsurePkgToRelay defines the action name.
	ActionNameEnsurePkgToRelay = "ensure_pkg_to_relay"

	// QueryClientTimeout defines the query client timeout.
	QueryClientTimeout = 3 * time.Second

	waitForClientReportTimeout  = 10 * time.Second
	waitForClientReportInterval = 2 * time.Second
)

// NewActionEnsurePkgToRelay get a new action.
func NewActionEnsurePkgToRelay(
	installerFileGroup fileiface.FileGroup,
	storageRelease release.IStorage,
	storageActionInstance workflow.IStorageActionInstance,
	storageNodeDeployment nodedeployment.IStorageNodeDeployment,
	fileHandler file.IHandler,
	proxyMessager relayhandler.IServerMessager,
	logger logger.Logger) action.Definition {

	return &actionEnsurePkgToRelay{
		installerFileGroup: installerFileGroup,

		storageRelease:        storageRelease,
		storageActionInstance: storageActionInstance,
		storageNodeDeployment: storageNodeDeployment,

		fileHandler:   fileHandler,
		proxyMessager: proxyMessager,

		logger: logger,
	}
}

// ActParamEnsurePkgToRelay ...
type ActParamEnsurePkgToRelay struct {
	Token    string `json:"token"`
	Operator string `json:"operator"`
}

type actionEnsurePkgToRelay struct {
	fileHandler   file.IHandler
	proxyMessager relayhandler.IServerMessager

	installerFileGroup fileiface.FileGroup

	storageRelease        release.IStorage
	storageActionInstance workflow.IStorageActionInstance
	storageNodeDeployment nodedeployment.IStorageNodeDeployment

	logger logger.Logger
}

// Name returns the name of the action.
func (act *actionEnsurePkgToRelay) Name() string {
	return ActionNameDetectInfoByWMI
}

// Version returns the version of the action.
func (act *actionEnsurePkgToRelay) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionEnsurePkgToRelay) Description() string {
	return "Use wmi to connect to the target machine, transfer files through sftp, and execute the installation command"
}

// Timeout returns the timeout of the action.
func (act *actionEnsurePkgToRelay) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionEnsurePkgToRelay) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionEnsurePkgToRelay) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionEnsurePkgToRelay) DelayFn() func() {
	return func() {
		time.Sleep(5 * time.Second) // nolint: mnd
	}
}

// Do this func define what the action will do.
// To ensure readability, this action uses fmt.Sprintf to concatenate characters.
// nolint: perfsprint,funlen,fnsize
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (act *actionEnsurePkgToRelay) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActParamEnsurePkgToRelay)
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

	releasePkgInfo, err := act.getReleasePackageInfo(ctx.Ctx, info)
	if err != nil {
		err = fmt.Errorf("failed to get release package info: %w", err)

		return err
	}

	installPkgInfo, err := act.getInstallerFile(ctx.Ctx, info)
	if err != nil {
		err = fmt.Errorf("failed to get installer file: %w", err)
	}

	relayInfo := &info.RelayInfo
	gp := gopool.NewPool()

	gp.Go(func() error {
		return act.askClientPkgPresence(ctx, releasePkgInfo.FileName, releasePkgInfo.MD5, relayInfo.AgentID)
	})

	gp.Go(func() error {
		return act.askClientPkgPresence(ctx, installPkgInfo.Info().Name, installPkgInfo.Info().MD5, relayInfo.AgentID)
	})

	if err := gp.Wait(); err != nil {
		act.logger.ErrorCtxf(ctx.Ctx,
			"failed to ask client pkg presence. host-id(%d): %s", info.Host.HostID, err)

		return err
	}

	// if pkg ready return directly. or transfer pkg.
	var releaseTransferred, installerTransferred bool

	// TODO: need to add pkg dest dir.
	gp.Go(func() error {
		if err := act.waitForPkgState(ctx, releasePkgInfo.FileName); err != nil {
			act.logger.Infof("wait for release package state. file(%s)", releasePkgInfo.FileName)
			if transferErr := act.transferReleasePkg(
				ctx.Ctx, types.ReleaseTypeAgent, &info.Host, relayInfo); transferErr != nil {
				return fmt.Errorf("failed to transfer release pkg: %w", transferErr)
			}
			releaseTransferred = true
		}

		return nil
	})

	gp.Go(func() error {
		if err := act.waitForPkgState(ctx, installPkgInfo.Info().Name); err != nil {
			act.logger.Infof("wait for release package state. file(%s)", installPkgInfo.Info().Name)
			if transferErr := act.transferInstaller(ctx.Ctx, &info.Host, relayInfo); transferErr != nil {
				return fmt.Errorf("failed to transfer installer pkg: %w", transferErr)
			}
			installerTransferred = true
		}

		return nil
	})

	if err := gp.Wait(); err != nil {
		return err
	}

	if releaseTransferred {
		if err := act.sendTransferPkgCompleteToClient(ctx.Ctx, releasePkgInfo.FileName, relayInfo); err != nil {
			return fmt.Errorf("failed to send transfer pkg completion for release to client: %w", err)
		}
	}

	if installerTransferred {
		if err := act.sendTransferPkgCompleteToClient(ctx.Ctx, installPkgInfo.Info().Name, relayInfo); err != nil {
			return fmt.Errorf("failed to send transfer pkg completion for installer to client: %w", err)
		}
	}

	return nil
}

func (act *actionEnsurePkgToRelay) getReleasePackageInfo(
	ctx context.Context, info *types.DeploymentInfo) (*types.Release, error) {

	cond := &types.ReleaseCondition{
		ExactInclude: &types.ReleaseExactFields{
			Type:       []types.ReleaseType{types.ReleaseTypeAgent},
			Generation: []types.Generation{info.Host.Dynamic.NodeGeneration},
			Enabled:    []bool{true},
			Version:    []string{info.Host.Dynamic.NodeVersion},
		},
	}
	releases, _, err := act.storageRelease.ListRelease(ctx, types.UnlimitedPage(), cond)
	if err != nil {
		return nil, err
	}

	if len(releases) != 1 {
		return nil, errors.New("no release found")
	}

	return releases[0], nil
}

func (act *actionEnsurePkgToRelay) getInstallerFile(
	ctx context.Context, info *types.DeploymentInfo) (fileiface.File, error) {

	toolName, err := tool.FormatInstallerName(info.Host.Dynamic.NodeOsType, info.Host.Dynamic.NodeCPUArch)
	if err != nil {
		return nil, fmt.Errorf("failed to format tools name: %w", err)
	}

	installPkgInfo, err := act.installerFileGroup.GetFile(ctx, toolName)
	if err != nil {
		return nil, fmt.Errorf("failed to get file. installer name(%s): %w", toolName, err)
	}

	return installPkgInfo, nil
}

func (act *actionEnsurePkgToRelay) askClientPkgPresence(ctx *action.InstanceContext,
	pkgName, mD5 string, agentID string) error {

	checkPkgEvent := protoRelay.CheckPkgStateReq{
		ActionName: ctx.Data.Name,
		OperInstID: ctx.Data.OperationInstanceID,
		PkgName:    pkgName,
		MD5:        mD5,
	}

	data, err := json.Marshal(checkPkgEvent)
	if err != nil {
		act.logger.Errorf("failed to marshal data: %v", err)
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
		act.logger.Errorf("client push req timed out after (%s)", QueryClientTimeout)
		return errors.New("client push req timed out")
	}

	return nil
}

func (act *actionEnsurePkgToRelay) waitForPkgState(ctx *action.InstanceContext, pkgName string) error {
	timeoutCtx, cancel := context.WithTimeout(ctx.Ctx, waitForClientReportTimeout)
	defer cancel()

	for {
		select {
		case <-timeoutCtx.Done():
			return fmt.Errorf("timeout waiting for pkg state. oper_inst_id(%s), action_name(%s)",
				ctx.Data.OperationInstanceID, ctx.Data.Name)
		default:
		}

		privateDataMap, err := act.storageActionInstance.GetActionInstancePrivateData(
			timeoutCtx,
			ctx.Data.OperationInstanceID,
			ctx.Data.Name,
		)
		if err != nil {
			act.logger.Errorf("get private data failed, retrying. oper_inst_id(%s), action_name(%s): %v",
				ctx.Data.OperationInstanceID, ctx.Data.Name, err)

			return fmt.Errorf("get private data failed, retrying. oper_inst_id(%s), action_name(%s): %w",
				ctx.Data.OperationInstanceID, ctx.Data.Name, err)
		}

		pkgState, exists := privateDataMap[pkgName]
		if !exists {
			time.Sleep(waitForClientReportInterval)
			continue
		}

		switch pkgState {
		case string(protoRelay.ClientReportSignalPkgComplete):
			act.logger.Infof("wait for pkg state operation succeed. oper_inst_id(%s), action_name(%s), state(%s)",
				ctx.Data.OperationInstanceID, ctx.Data.Name)

			return nil

		case string(protoRelay.ClientReportSignalPkgUnComplete):
			return fmt.Errorf("wait for pkg state operation succeed. oper_inst_id(%s), action_name(%s), state(%s)",
				ctx.Data.OperationInstanceID, ctx.Data.Name, pkgState)

		default:
			return fmt.Errorf("wait for pkg state operation failed. oper_inst_id(%s), action_name(%s), state(%s)",
				pkgState, ctx.Data.OperationInstanceID, ctx.Data.Name)
		}
	}
}

func (act *actionEnsurePkgToRelay) transferReleasePkg(ctx context.Context,
	rt types.ReleaseType, info *types.Host, relayInfo *types.RelayInfo) error {

	transferHandler, err := act.fileHandler.LaunchTransferRelease(ctx,
		info.Dynamic.NodeGeneration,
		rt,
		platform.Platform{
			OS:   info.Dynamic.NodeOsType,
			Arch: info.Dynamic.NodeCPUArch,
		},
		info.Dynamic.NodeVersion,
		relayInfo.PackageDestDir,
		&types.Host{HostID: relayInfo.HostID})
	if err != nil {
		return fmt.Errorf("failed to launch transfer release. host-id(%d): %w", relayInfo.HostID, err)
	}

	act.logger.InfoCtxf(ctx, "launched transfer release. task-id(%s), host-id(%d)",
		transferHandler.GetTaskID(), relayInfo.HostID)

	result, err := transferHandler.WaitUntilDone(ctx)
	if err != nil {
		return fmt.Errorf("failed to wait until transfer release done. task-id(%s), host-id(%d): %w",
			transferHandler.GetTaskID(), relayInfo.HostID, err)
	}

	if !result.Terminated {
		return fmt.Errorf("transfer release not terminated. task-id(%s), host-id(%d)",
			transferHandler.GetTaskID(), relayInfo.HostID)
	}

	if result.ErrorCode != 0 {
		return fmt.Errorf("transfer release failed. task-id(%s), host-id(%d), err-code(%d), err-msg(%s)",
			transferHandler.GetTaskID(), relayInfo.HostID, result.ErrorCode, result.ErrorMessage)
	}

	act.logger.InfoCtxf(ctx, "transfer release done. task-id(%s), host-id(%d)",
		transferHandler.GetTaskID(), relayInfo.HostID)

	return nil
}

func (act *actionEnsurePkgToRelay) transferInstaller(ctx context.Context,
	info *types.Host, relayInfo *types.RelayInfo) error {

	transferHandler, err := act.fileHandler.LaunchTransferInstaller(ctx,
		types.Generation2,
		platform.Platform{
			OS:   info.Dynamic.NodeOsType,
			Arch: info.Dynamic.NodeCPUArch,
		},
		relayInfo.PackageDestDir,
		&types.Host{HostID: relayInfo.HostID})
	if err != nil {
		return fmt.Errorf("failed to launch transfer installer. host-id(%d): %w", info.HostID, err)
	}

	act.logger.InfoCtxf(ctx, "launched transfer installer. task-id(%s), host-id(%d)",
		transferHandler.GetTaskID(), info.HostID)

	result, err := transferHandler.WaitUntilDone(ctx)
	if err != nil {
		return fmt.Errorf("failed to wait until transfer installer done. task-id(%s), host-id(%d): %w",
			transferHandler.GetTaskID(), info.HostID, err)
	}

	if !result.Terminated {
		return fmt.Errorf("transfer installer not terminated. task-id(%s), host-id(%d)",
			transferHandler.GetTaskID(), info.HostID)
	}

	if result.ErrorCode != 0 {
		return fmt.Errorf("transfer installer failed. task-id(%s), host-id(%d), err-code(%d), err-msg(%s)",
			transferHandler.GetTaskID(), info.HostID, result.ErrorCode, result.ErrorMessage)
	}

	act.logger.InfoCtxf(ctx, "transfer installer done. task-id(%s), host-id(%d)",
		transferHandler.GetTaskID(), info.HostID)

	return nil
}

func (act *actionEnsurePkgToRelay) sendTransferPkgCompleteToClient(
	ctx context.Context, pkgName string, relayInfo *types.RelayInfo) error {

	transferPkgCompleteEvent := protoRelay.TransferPkgCompleteReq{
		PackageDestDir: relayInfo.PackageDestDir,
		PkgName:        pkgName,
	}
	data, err := json.Marshal(transferPkgCompleteEvent)
	if err != nil {
		act.logger.Errorf("failed to marshal data: %v", err)
	}

	errCh := act.proxyMessager.PushToClient(ctx,
		protoRelay.ServerPushEventTypeTransferPkgComplete, data, relayInfo.AgentID)

	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("failed to send transfer pkg complete to client. agent-id(%s): %w",
				relayInfo.AgentID, err)
		}

	case <-ctx.Done():
		return fmt.Errorf("failed to send transfer pkg complete to client. agent-id(%s): %w",
			relayInfo.AgentID, ctx.Err())
	}

	return nil
}
