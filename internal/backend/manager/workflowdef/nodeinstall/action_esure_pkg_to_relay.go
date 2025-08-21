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
	relayReportKey "github.com/TencentBlueKing/bk-nodemgr/internal/relay/constance"
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

	queryRelayTimeout          = 3 * time.Second
	waitForRelayReportInterval = 2 * time.Second
	waitForRelayReportTimeout  = 10 * time.Second
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
	return 5 * time.Minute // nolint: mnd
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

	// TODO: if pkg not required, no need to get the info.
	releasePkg, installerPkg, err := act.getRequiredPackages(ctx, info)
	if err != nil {
		return err
	}

	filesToProcess := act.determineFilesToProcess(info, releasePkg, installerPkg)
	if len(filesToProcess) == 0 {
		ctx.Data.LogI("no packages to process, all transfers disabled by options")
		return nil
	}

	if err := act.queryRelayPackageState(ctx, filesToProcess, info.RelayInfo.AgentID); err != nil {
		return fmt.Errorf("query relay state failed: %w", err)
	}

	pkgStates, storageDir, err := act.waitForRelayReportFile(ctx, filesToProcess)
	if err != nil {
		return fmt.Errorf("wait for relay report failed: %w", err)
	}

	if storageDir == "" {
		ctx.Data.LogE("failed to get relay storage dir. storage dir is empty")
		return errors.New("failed to get relay storage dir. storage dir is empty")
	}
	info.RelayInfo.PackageDestDir = storageDir

	transferredPkgs, err := act.transferMissingPackages(ctx, info, releasePkg, installerPkg, pkgStates)
	if err != nil {
		return err
	}

	// only notify relay if there are packages to transfer.
	if len(transferredPkgs) > 0 {
		if err := act.notifyRelayToReceivePackage(ctx, transferredPkgs, &info.RelayInfo); err != nil {
			return fmt.Errorf("notify transfer completion failed: %w", err)
		}
		if err := act.waitForRelayReportStorage(ctx); err != nil {
			return fmt.Errorf("wait for relay report storage failed: %w", err)
		}
	}

	ctx.Data.LogI("all packages processed successfully")

	return nil
}

func (act *actionEnsurePkgToRelay) getRequiredPackages(
	ctx *action.InstanceContext, info *types.DeploymentInfo) (*types.Release, fileiface.File, error) {

	releasePkg, err := act.getReleasePackageInfo(ctx.Ctx, info)
	if err != nil {
		ctx.Data.LogE(fmt.Sprintf("get release package failed: %v", err))
		return nil, nil, fmt.Errorf("get release package: %w", err)
	}
	ctx.Data.LogI(fmt.Sprintf("get release package info. file-name(%s)", releasePkg.FileName))

	installPkg, err := act.getInstallerFile(ctx.Ctx, info)
	if err != nil {
		ctx.Data.LogE(fmt.Sprintf("get installer package failed: %v", err))
		return nil, nil, fmt.Errorf("get installer package: %w", err)
	}
	ctx.Data.LogI(fmt.Sprintf("get installer package info. file-name(%s)", installPkg.Info().Name))

	return releasePkg, installPkg, nil
}

func (act *actionEnsurePkgToRelay) determineFilesToProcess(
	info *types.DeploymentInfo, releasePkg *types.Release, installPkg fileiface.File) []protoRelay.FileInfo {

	files := make([]protoRelay.FileInfo, 0)

	if !info.TransferOptions.SelectDownloads || info.TransferOptions.EnableReleasePackage {
		files = append(files, protoRelay.FileInfo{
			FileName: releasePkg.FileName,
			FileMD5:  releasePkg.MD5,
		})
	}

	if !info.TransferOptions.SelectDownloads || info.TransferOptions.EnableInstaller {
		files = append(files, protoRelay.FileInfo{
			FileName: installPkg.Info().Name,
			FileMD5:  installPkg.Info().MD5,
		})
	}

	return files
}

func (act *actionEnsurePkgToRelay) queryRelayPackageState(
	ctx *action.InstanceContext, files []protoRelay.FileInfo, agentID string) error {

	event := protoRelay.CheckPkgStateReq{
		ActionName: ctx.Data.Name,
		OperInstID: ctx.Data.OperationInstanceID,
		FileList:   files,
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event failed: %w", err)
	}
	errCh := act.proxyMessager.PushToClient(ctx.Ctx,
		protoRelay.ServerPushEventTypeCheckPkgState, data, agentID)

	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("push to relay failed: %w", err)
		}
	case <-time.After(queryRelayTimeout):
		return errors.New("relay response timeout")
	}

	ctx.Data.LogI("package state query sent to relay.")

	return nil
}

func (act *actionEnsurePkgToRelay) waitForRelayReportFile(
	ctx *action.InstanceContext, files []protoRelay.FileInfo) (map[string]bool, string, error) {

	results := make(map[string]bool, len(files))
	for _, file := range files {
		results[file.FileName] = false
	}

	fileStorageDir := ""
	completedCount := 0

	timeoutCtx, cancel := context.WithTimeout(ctx.Ctx, waitForRelayReportTimeout)
	defer cancel()

	ticker := time.NewTicker(waitForRelayReportInterval)
	defer ticker.Stop()

	for completedCount < len(files) {
		select {
		case <-timeoutCtx.Done():
			return results, fileStorageDir, fmt.Errorf("wait for relay report timed out. oper_inst_id(%s), action_name(%s)",
				ctx.Data.OperationInstanceID, ctx.Data.Name)

		case <-ticker.C:
			privateData, err := act.storageActionInstance.GetActionInstancePrivateData(
				timeoutCtx, ctx.Data.OperationInstanceID, ctx.Data.Name)
			if err != nil {
				act.logger.Warnf("get private data failed, retrying. oper_inst_id(%s), action_name(%s): %v",
					ctx.Data.OperationInstanceID, ctx.Data.Name, err)

				continue
			}

			fileStateRaw, exists := privateData[relayReportKey.FileStateKey]
			if !exists {
				continue
			}

			fileState, ok := fileStateRaw.(map[string]any)
			if !ok {
				return results, fileStorageDir, errors.New("unexpected type for file state")
			}

			if fileStorageDir == "" {
				if storageDirRaw, exists := fileState[relayReportKey.FileStateStorageKey]; exists {
					if storageDir, ok := storageDirRaw.(string); ok && storageDir != "" {
						fileStorageDir = storageDir
						ctx.Data.LogI(fmt.Sprintf("relay storage dir set. dir(%s)", fileStorageDir))
					}
				}
			}

			for fileName, state := range fileState {
				stateStr, ok := state.(string)
				if !ok {
					return results, fileStorageDir, errors.New("unexpected type for file state")
				}

				if stateStr == string(protoRelay.RelayReportPkgComplete) {
					results[fileName] = true
					ctx.Data.LogI(fmt.Sprintf("package state complete. file-name(%s)", fileName))
				} else if stateStr == string(protoRelay.RelayReportPkgInComplete) {
					results[fileName] = false
					ctx.Data.LogI(fmt.Sprintf("package state incomplete. file-name(%s)", fileName))
				}
				completedCount++
			}
		}
	}

	return results, fileStorageDir, nil
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

func (act *actionEnsurePkgToRelay) transferMissingPackages(
	ctx *action.InstanceContext,
	info *types.DeploymentInfo,
	releasePkg *types.Release,
	installPkg fileiface.File,
	pkgStates map[string]bool) ([]string, error) {

	var transferred []string
	gp := gopool.NewPool()

	if state, exists := pkgStates[releasePkg.FileName]; exists && !state {
		gp.Go(func() error {
			if err := act.transferReleasePkg(ctx, types.ReleaseTypeAgent, &info.Host, &info.RelayInfo); err != nil {
				ctx.Data.LogE(fmt.Sprintf("failed to transfer release package: %v", err))
				return fmt.Errorf("failed to transfer release package: %w", err)
			}
			transferred = append(transferred, releasePkg.FileName)

			return nil
		})
	} else {
		ctx.Data.LogI(fmt.Sprintf("release package exists. file-name(%s)", releasePkg.FileName))
	}

	installPkgName := installPkg.Info().Name
	if state, exists := pkgStates[installPkgName]; exists && !state {
		gp.Go(func() error {
			if err := act.transferInstaller(ctx, &info.Host, &info.RelayInfo); err != nil {
				ctx.Data.LogE(fmt.Sprintf("failed to transfer installer package: %v", err))
				return fmt.Errorf("failed to transfer installer package: %w", err)
			}
			transferred = append(transferred, installPkgName)

			return nil
		})
	} else {
		ctx.Data.LogI(fmt.Sprintf("installer package exists. file-name(%s)", installPkgName))
	}

	if err := gp.Wait(); err != nil {
		return nil, err
	}

	return transferred, nil
}

func (act *actionEnsurePkgToRelay) transferReleasePkg(ctx *action.InstanceContext,
	rt types.ReleaseType, info *types.Host, relayInfo *types.RelayInfo) error {

	act.logger.Infof("transfer release. host-id(%d)", relayInfo.HostID)
	ctx.Data.LogI(fmt.Sprintf("transferring release package. relay-host-id(%d)", relayInfo.HostID))

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
	act.logger.InfoCtxf(ctx.Ctx, "launched transfer release. task-id(%s), relay-host-id(%d)",
		transferHandler.GetTaskID(), relayInfo.HostID)

	result, err := transferHandler.WaitUntilDone(ctx.Ctx)
	if err != nil {
		return fmt.Errorf("failed to wait until transfer release done. task-id(%s), relay-host-id(%d): %w",
			transferHandler.GetTaskID(), relayInfo.HostID, err)
	}

	if !result.Terminated {
		return fmt.Errorf("transfer release not terminated. task-id(%s), relay-host-id(%d)",
			transferHandler.GetTaskID(), relayInfo.HostID)
	}

	if result.ErrorCode != 0 {
		return fmt.Errorf("transfer release failed. task-id(%s), relay-host-id(%d), err-code(%d), err-msg(%s)",
			transferHandler.GetTaskID(), relayInfo.HostID, result.ErrorCode, result.ErrorMessage)
	}

	ctx.Data.LogI(fmt.Sprintf("transfer release done. task-id(%s), relay-host-id(%d)",
		transferHandler.GetTaskID(), relayInfo.HostID))
	act.logger.InfoCtxf(ctx.Ctx, "transfer release done. task-id(%s), relay-host-id(%d)",
		transferHandler.GetTaskID(), relayInfo.HostID)

	return nil
}

func (act *actionEnsurePkgToRelay) transferInstaller(ctx *action.InstanceContext,
	info *types.Host, relayInfo *types.RelayInfo) error {

	act.logger.Infof("transfer installer. relay-host-id(%d)", relayInfo.HostID)
	ctx.Data.LogI(fmt.Sprintf("transferring installer package. relay-host-id(%d)", relayInfo.HostID))

	transferHandler, err := act.fileHandler.LaunchTransferInstaller(ctx.Ctx,
		types.Generation2,
		platform.Platform{
			OS:   info.Dynamic.NodeOsType,
			Arch: info.Dynamic.NodeCPUArch,
		},
		relayInfo.PackageDestDir,
		&types.Host{HostID: relayInfo.HostID})
	if err != nil {
		ctx.Data.LogE(fmt.Sprintf("failed to launch transfer installer. relay-host-id(%d): %v", relayInfo.HostID, err))
		return fmt.Errorf("failed to launch transfer installer. relay-host-id(%d): %w", relayInfo.HostID, err)
	}

	ctx.Data.LogI(fmt.Sprintf("launched transfer installer. task-id(%s), relay-host-id(%d)",
		transferHandler.GetTaskID(), relayInfo.HostID))
	act.logger.InfoCtxf(ctx.Ctx, "launched transfer installer. task-id(%s), relay-host-id(%d)",
		transferHandler.GetTaskID(), relayInfo.HostID)

	result, err := transferHandler.WaitUntilDone(ctx.Ctx)
	if err != nil {
		return fmt.Errorf("failed to wait until transfer installer done. task-id(%s), relay-host-id(%d): %w",
			transferHandler.GetTaskID(), relayInfo.HostID, err)
	}

	if !result.Terminated {
		return fmt.Errorf("transfer installer not terminated. task-id(%s), relay-host-id(%d)",
			transferHandler.GetTaskID(), relayInfo.HostID)
	}

	if result.ErrorCode != 0 {
		return fmt.Errorf("transfer installer failed. task-id(%s), relay-host-id(%d), err-code(%d), err-msg(%s)",
			transferHandler.GetTaskID(), relayInfo.HostID, result.ErrorCode, result.ErrorMessage)
	}

	ctx.Data.LogI(fmt.Sprintf("transfer installer done. task-id(%s), relay-host-id(%d)",
		transferHandler.GetTaskID(), relayInfo.HostID))
	act.logger.InfoCtxf(ctx.Ctx, "transfer installer done. task-id(%s), relay-host-id(%d)",
		transferHandler.GetTaskID(), relayInfo.HostID)

	return nil
}

func (act *actionEnsurePkgToRelay) notifyRelayToReceivePackage(
	ctx *action.InstanceContext, pkgNames []string, relayInfo *types.RelayInfo) error {

	event := protoRelay.NotifyReceiveReq{
		ActionName: ctx.Data.Name,
		OperInstID: ctx.Data.OperationInstanceID,
		PkgName:    pkgNames}
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event failed: %w", err)
	}

	errCh := act.proxyMessager.PushToClient(ctx.Ctx,
		protoRelay.ServerPushEventTypeNotifyReceive, data, relayInfo.AgentID)

	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("notify completion failed: %w", err)
		}
	case <-ctx.Ctx.Done():
		return ctx.Ctx.Err()
	}

	act.logger.Infof("transfer completion notified for packages. pkg-names(%v)", pkgNames)

	return nil
}

func (act *actionEnsurePkgToRelay) waitForRelayReportStorage(
	ctx *action.InstanceContext) error {

	timeoutCtx, cancel := context.WithTimeout(ctx.Ctx, waitForRelayReportTimeout)
	defer cancel()

	ticker := time.NewTicker(waitForRelayReportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			return fmt.Errorf("wait for relay report storage result timed out. oper_inst_id(%s), action_name(%s)",
				ctx.Data.OperationInstanceID, ctx.Data.Name)

		case <-ticker.C:
			privateData, err := act.storageActionInstance.GetActionInstancePrivateData(
				timeoutCtx, ctx.Data.OperationInstanceID, ctx.Data.Name)
			if err != nil {
				act.logger.Warnf("get private data failed, retrying. oper_inst_id(%s), action_name(%s): %v",
					ctx.Data.OperationInstanceID, ctx.Data.Name, err)

				continue
			}

			relayStorageResultRaw, exists := privateData[relayReportKey.StorageResultKey]
			if !exists {
				continue
			}

			relayStorageResult, ok := relayStorageResultRaw.(map[string]any)
			if !ok {
				return errors.New("unexpected type for relay ")
			}

			errMsgRaw := relayStorageResult[relayReportKey.StorageResultMsgKey]
			errMsg, ok := errMsgRaw.(string)
			if !ok {
				return errors.New("unexpected type for file state")
			}

			if errMsg == "" {
				return nil
			}

			return errors.New(errMsg)
		}
	}
}
