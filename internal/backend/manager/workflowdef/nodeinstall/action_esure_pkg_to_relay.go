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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/nodepkg"
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

	// relayFileStateKey defines the relay file state key.
	relayFileStateKey = "relay_file_state"

	// relayFileStorageKey defines the relay file storage tmp dir key.
	relayFileStateStorageKey = "relay_file_state_storage_dir"

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
	return ActionNameEnsurePkgToRelay
}

// Version returns the version of the action.
func (act *actionEnsurePkgToRelay) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionEnsurePkgToRelay) Description() string {
	return "query the relay machine to check whether the package exists. " +
		"if not, pass the install and release packages to relay."
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
		ctx.Data.LogE(fmt.Sprintf("failed to get release package info: %s", err))
		err = fmt.Errorf("failed to get release package info: %w", err)

		return err
	}
	ctx.Data.LogI(fmt.Sprintf("get release package. package name(%s)", releasePkgInfo.FileName))

	installPkgInfo, err := act.getInstallerFile(ctx.Ctx, info)
	if err != nil {
		ctx.Data.LogE(fmt.Sprintf("failed to get installer file: %s", err))
		err = fmt.Errorf("failed to get installer file: %w", err)
	}
	ctx.Data.LogI(fmt.Sprintf("get installer package. package name(%s)", installPkgInfo.Info().Name))

	relayInfo := &info.RelayInfo

	filesToCheck := []protoRelay.FileInfo{}
	pkgNames := []string{}

	if !info.TransferOptions.SelectDownloads || info.TransferOptions.EnableReleasePackage {
		filesToCheck = append(filesToCheck, protoRelay.FileInfo{
			FileName: releasePkgInfo.FileName,
			FileMD5:  releasePkgInfo.MD5,
		})
		pkgNames = append(pkgNames, releasePkgInfo.FileName)
	}

	if !info.TransferOptions.SelectDownloads || info.TransferOptions.EnableInstaller {
		filesToCheck = append(filesToCheck, protoRelay.FileInfo{
			FileName: installPkgInfo.Info().Name,
			FileMD5:  installPkgInfo.Info().MD5,
		})
		pkgNames = append(pkgNames, installPkgInfo.Info().Name)
	}

	if len(filesToCheck) == 0 {
		ctx.Data.LogI("no packages to process, all transfers disabled by options")
		return nil
	}

	if err := act.askClientPkgState(ctx, filesToCheck, relayInfo.AgentID); err != nil {
		ctx.Data.LogE(fmt.Sprintf("failed to ask client pkg presence. host-id(%d): %s", info.Host.HostID, err))
		act.logger.ErrorCtxf(ctx.Ctx, "failed to ask client pkg presence. host-id(%d): %s", info.Host.HostID, err)

		return err
	}

	ctx.Data.LogI(fmt.Sprintf("ask client package state success. wait relay send pkg state. relay-host-id(%d)",
		info.RelayInfo.HostID))

	pkgStates, storageTmpDir, err := act.waitForPkgsState(ctx, pkgNames)
	if err != nil {
		ctx.Data.LogE(fmt.Sprintf("failed to wait for package states: %v", err))
		act.logger.ErrorCtxf(ctx.Ctx, "failed to wait for package states: %v", err)

		return fmt.Errorf("failed to wait for package states: %w", err)
	}
	if storageTmpDir != "" {
		ctx.Data.LogE("failed to wait for package states. get empty storage dir")
		act.logger.ErrorCtxf(ctx.Ctx, "failed to wait for package states. get empty storage dir")

		return errors.New("failed to wait for package states. get empty storage dir")
	}
	info.RelayInfo.PackageDestDir = relayInfo.PackageDestDir

	var transferredPkgs []string
	gp := gopool.NewPool()

	if !pkgStates[releasePkgInfo.FileName] {
		transferredPkgs = append(transferredPkgs, installPkgInfo.Info().Name)
		gp.Go(func() error {
			ctx.Data.LogI(fmt.Sprintf("start transfer release package. file-name(%s)", releasePkgInfo.FileName))
			if transferErr := act.transferReleasePkg(
				ctx, types.ReleaseTypeAgent, &info.Host, relayInfo); transferErr != nil {
				ctx.Data.LogE(fmt.Sprintf("failed to transfer release package: %v", transferErr))

				return fmt.Errorf("failed to transfer release package: %w", transferErr)
			}
			ctx.Data.LogI(fmt.Sprintf("transfer release package success. file-name(%s)", releasePkgInfo.FileName))

			return nil
		})
	}
	ctx.Data.LogI(fmt.Sprintf("release package already exists. skip transfer. file-name(%s)", releasePkgInfo.FileName))

	if !pkgStates[installPkgInfo.Info().Name] {
		transferredPkgs = append(transferredPkgs, installPkgInfo.Info().Name)
		gp.Go(func() error {
			ctx.Data.LogI(fmt.Sprintf("start transfer installer package. file-name(%s)", installPkgInfo.Info().Name))
			if transferErr := act.transferInstaller(ctx, &info.Host, relayInfo); transferErr != nil {
				ctx.Data.LogE(fmt.Sprintf("failed to transfer installer package: %v", transferErr))
				return fmt.Errorf("failed to transfer installer package: %w", transferErr)
			}

			ctx.Data.LogI(fmt.Sprintf("transfer installer package success. file-name(%s)", installPkgInfo.Info().Name))

			return nil
		})
	}
	ctx.Data.LogI(fmt.Sprintf("installer package already exists. skip transfer. file-name(%s)",
		installPkgInfo.Info().Name))

	if len(transferredPkgs) > 0 {
		if err := gp.Wait(); err != nil {
			return err
		}

		if err := act.sendTransferPkgCompleteToClient(ctx.Ctx, transferredPkgs, relayInfo); err != nil {
			ctx.Data.LogE(fmt.Sprintf("failed to send transfer pkg completion to relay. relay-id(%d)", info.RelayInfo.HostID))

			return fmt.Errorf("failed to send transfer package completion to relay: %w", err)
		}
		ctx.Data.LogI(fmt.Sprintf("send transfer pkg completion to client success to relay. package(%v). relay-id(%d",
			transferredPkgs, info.RelayInfo.HostID))
	}
	ctx.Data.LogI("all packages already complete")

	return nil
}

func (act *actionEnsurePkgToRelay) getReleasePackageInfo(
	ctx context.Context, info *types.DeploymentInfo) (*types.Release, error) {

	FileName, err := nodepkg.FormatPkgName(
		types.Generation2,
		types.ReleaseTypeAgent,
		platform.Platform{OS: info.Host.Dynamic.NodeOsType, Arch: info.Host.Dynamic.NodeCPUArch},
		info.Host.Dynamic.NodeVersion,
	)
	if err != nil {
		return nil, err
	}
	act.logger.Infof("get release package info. file-name(%s)", FileName)

	cond := &types.ReleaseCondition{
		ExactInclude: &types.ReleaseExactFields{
			FileName: []string{FileName},
		},
	}
	releases, _, err := act.storageRelease.ListRelease(ctx, types.UnlimitedPage(), cond)
	if err != nil {
		return nil, err
	}

	if len(releases) == 0 {
		return nil, fmt.Errorf("release package not found. file-name(%s)", FileName)
	}

	if len(releases) > 1 {
		return nil, fmt.Errorf("release package not unique. file-name(%s)", FileName)
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
		return nil, fmt.Errorf("failed to get file. installer-name(%s): %w", toolName, err)
	}

	return installPkgInfo, nil
}

func (act *actionEnsurePkgToRelay) askClientPkgState(ctx *action.InstanceContext,
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
func (act *actionEnsurePkgToRelay) waitForPkgsState(ctx *action.InstanceContext, pkgNames []string) (
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

			// get relay report storage directory path.
			storageDir, err := act.extractStorageDir(fileState)
			if err != nil {
				return results, fileStorageDir, err
			}
			fileStorageDir = storageDir

			// update file state results.
			count := act.updateResultsWithState(fileState, results, ctx)
			completedCount += count
		}
	}

	return results, fileStorageDir, nil
}

func (act *actionEnsurePkgToRelay) extractStorageDir(fileState map[string]any) (string, error) {
	storageDirRaw, exists := fileState[relayFileStateStorageKey]
	if !exists {
		act.logger.Warnf("Storage dir key '%s' not found in file state", relayFileStateStorageKey)
		return "", fmt.Errorf("storage dir key '%s' not found in file state", relayFileStateStorageKey)
	}

	storageDir, ok := storageDirRaw.(string)
	if !ok {
		act.logger.Errorf("unexpected type for storage dir: %T, expected string", storageDirRaw)
		return "", errors.New("invalid storage dir type")
	}

	return storageDir, nil
}

func (act *actionEnsurePkgToRelay) updateResultsWithState(fileState map[string]any,
	results map[string]bool, ctx *action.InstanceContext) int {

	completedThisRound := 0

	for pkgName, state := range fileState {
		// Skip already completed packages
		if completed, exists := results[pkgName]; exists && completed {
			continue
		}

		switch state {
		case string(protoRelay.ClientReportPkgComplete):
			ctx.Data.LogI(fmt.Sprintf("package state complete. pkg-name(%s), oper_inst_id(%s), action_name(%s)",
				pkgName, ctx.Data.OperationInstanceID, ctx.Data.Name))
			results[pkgName] = true
		case string(protoRelay.ClientReportPkgInComplete):
			ctx.Data.LogI(fmt.Sprintf("package state incomplete. pkg-name(%s), oper_inst_id(%s), action_name(%s)",
				pkgName, ctx.Data.OperationInstanceID, ctx.Data.Name))
			results[pkgName] = false
		default:
			continue
		}
		completedThisRound++
	}

	return completedThisRound
}

func (act *actionEnsurePkgToRelay) transferReleasePkg(ctx *action.InstanceContext,
	rt types.ReleaseType, info *types.Host, relayInfo *types.RelayInfo) error {

	act.logger.Infof("transfer release. host-id(%d)", relayInfo.HostID)
	ctx.Data.LogI(fmt.Sprintf("transfer release to relay. relay-host-id(%d)", relayInfo.HostID))

	transferHandler, err := act.fileHandler.LaunchTransferRelease(ctx.Ctx,
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
		ctx.Data.LogE(fmt.Sprintf("failed to launch transfer release. relay-host-id(%d): %v", relayInfo.HostID, err))
		return fmt.Errorf("failed to launch transfer release. relay-host-id(%d): %w", relayInfo.HostID, err)
	}

	ctx.Data.LogI(fmt.Sprintf("launched transfer release. task-id(%s), relay-host-id(%d)",
		transferHandler.GetTaskID(), relayInfo.HostID))
	act.logger.InfoCtxf(ctx.Ctx, "launched transfer release. task-id(%s), host-id(%d)",
		transferHandler.GetTaskID(), relayInfo.HostID)

	result, err := transferHandler.WaitUntilDone(ctx.Ctx)
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

	ctx.Data.LogI(fmt.Sprintf("transfer release done. task-id(%s), relay-host-id(%d)",
		transferHandler.GetTaskID(), relayInfo.HostID))
	act.logger.InfoCtxf(ctx.Ctx, "transfer release done. task-id(%s), host-id(%d)",
		transferHandler.GetTaskID(), relayInfo.HostID)

	return nil
}

func (act *actionEnsurePkgToRelay) transferInstaller(ctx *action.InstanceContext,
	info *types.Host, relayInfo *types.RelayInfo) error {

	act.logger.Infof("transfer installer. host-id(%d)", info.HostID)
	ctx.Data.LogI(fmt.Sprintf("transfer installer to relay. relay-host-id(%d)", relayInfo.HostID))

	transferHandler, err := act.fileHandler.LaunchTransferInstaller(ctx.Ctx,
		types.Generation2,
		platform.Platform{
			OS:   info.Dynamic.NodeOsType,
			Arch: info.Dynamic.NodeCPUArch,
		},
		relayInfo.PackageDestDir,
		&types.Host{HostID: relayInfo.HostID})
	if err != nil {
		ctx.Data.LogE(fmt.Sprintf("failed to launch transfer installer. host-id(%d): %v", info.HostID, err))
		return fmt.Errorf("failed to launch transfer installer. host-id(%d): %w", info.HostID, err)
	}

	ctx.Data.LogI(fmt.Sprintf("launched transfer installer. task-id(%s), relay-host-id(%d)",
		transferHandler.GetTaskID(), relayInfo.HostID))
	act.logger.InfoCtxf(ctx.Ctx, "launched transfer installer. task-id(%s), host-id(%d)",
		transferHandler.GetTaskID(), info.HostID)

	result, err := transferHandler.WaitUntilDone(ctx.Ctx)
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

	ctx.Data.LogI(fmt.Sprintf("transfer installer done. task-id(%s), relay-host-id(%d)",
		transferHandler.GetTaskID(), relayInfo.HostID))
	act.logger.InfoCtxf(ctx.Ctx, "transfer installer done. task-id(%s), host-id(%d)",
		transferHandler.GetTaskID(), info.HostID)

	return nil
}

func (act *actionEnsurePkgToRelay) sendTransferPkgCompleteToClient(
	ctx context.Context, pkgNames []string, relayInfo *types.RelayInfo) error {

	act.logger.Infof("send transfer pkg complete to client. agent-id(%s), pkg-names(%v)", relayInfo.AgentID, pkgNames)

	transferPkgCompleteEvent := protoRelay.TransferPkgCompleteReq{
		PkgName: pkgNames,
	}
	data, err := json.Marshal(transferPkgCompleteEvent)
	if err != nil {
		act.logger.Errorf("failed to marshal data: %v", err)
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	errCh := act.proxyMessager.PushToClient(ctx,
		protoRelay.ServerPushEventTypeTransferPkgComplete, data, relayInfo.AgentID)

	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("failed to send transfer pkg complete to client. agent-id(%s): %w",
				relayInfo.AgentID, err)
		}
		act.logger.Infof("successfully sent transfer pkg complete for packages. pkg-names(%v)", pkgNames)

	case <-ctx.Done():
		return fmt.Errorf("failed to send transfer pkg complete to client. agent-id(%s): %w",
			relayInfo.AgentID, ctx.Err())
	}

	return nil
}
