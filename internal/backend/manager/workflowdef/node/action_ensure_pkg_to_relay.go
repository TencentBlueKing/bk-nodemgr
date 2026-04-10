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
	nodeStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/release"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/internal/relay/relayconstant"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/tool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoRelay "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/relay"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/relayhandler"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/file"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameEnsurePkgToRelay defines the action name.
	ActionNameEnsurePkgToRelay = "ensure_pkg_to_relay"

	waitForRelayReportInterval = 3 * time.Second
	waitForRelayReportTimeout  = 30 * time.Second
)

// NewActionEnsurePkgToRelay get a new action.
func NewActionEnsurePkgToRelay(capability *Capability) action.Definition {
	return &actionEnsurePkgToRelay{
		installerFileGroup: capability.InstallerFileGroup,

		storageRelease:        capability.StorageRelease,
		storageActionInstance: capability.StorageWorkflow,
		storageNodeDeployment: capability.StorageNode,
		storageHost:           capability.StorageTopo,

		fileHandler:   capability.FileHandler,
		proxyMessager: capability.ProxyMessager,
	}
}

// ActParamEnsurePkgToRelay ...
type ActParamEnsurePkgToRelay struct {
	nodeUtils.NodeActionStandardParam `json:",inline"`
}

type actionEnsurePkgToRelay struct {
	fileHandler   file.IHandler
	proxyMessager relayhandler.IServerMessager

	installerFileGroup fileiface.FileGroup

	storageRelease        release.IStorage
	storageActionInstance workflow.IStorageActionInstance
	storageNodeDeployment nodeStg.IDaoNodeDeployment
	storageHost           topoStg.IStorageHost
}

// Name returns the name of the action.
func (act *actionEnsurePkgToRelay) Name() string {
	return ActionNameEnsurePkgToRelay
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionEnsurePkgToRelay) DisplayNameZh() string {
	return "确保安装包同步到中继节点"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionEnsurePkgToRelay) DisplayNameEn() string {
	return "Ensure Package to Relay"
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
// nolint: perfsprint,funlen
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (act *actionEnsurePkgToRelay) Do(ctx *action.InstanceContext) error {
	param := new(ActParamEnsurePkgToRelay)
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

	// TODO: if pkg not required, no need to get the info.
	releasePkg, installerPkg, err := act.getRequiredPackages(std)
	if err != nil {
		return err
	}

	filesToProcess := act.determineFilesToProcess(std, releasePkg, installerPkg)
	if len(filesToProcess) == 0 {
		std.InstanceData().Log().
			Zh("无需处理安装包, 所有传输已被选项禁用").
			En("no packages to process, all transfers disabled by options").
			Info()

		return nil
	}

	relayInfo, err := std.GetSelectedRelay()
	if err != nil {
		return fmt.Errorf("failed to get selected relay: %w", err)
	}

	if err := act.queryRelayPackageState(std, filesToProcess, relayInfo); err != nil {
		return fmt.Errorf("failed to query relay package state: %w", err)
	}

	pkgStates, storageDir, err := act.waitForRelayReportFile(std, filesToProcess)
	if err != nil {
		return fmt.Errorf("failed to wait for relay report: %w", err)
	}

	if storageDir == "" {
		std.InstanceData().Log().
			Zh("获取 relay 存储目录失败, 存储目录为空").
			En("failed to get relay storage dir. storage dir is empty").
			Error()

		return errors.New("failed to get relay storage dir. storage dir is empty")
	}
	std.DeployInfo().RelayInfo.PackageDestDir = storageDir

	transferredPkgs, err := act.transferMissingPackages(std, releasePkg, installerPkg, pkgStates)
	if err != nil {
		return err
	}

	// only notify relay if there are packages to transfer.
	if len(transferredPkgs) > 0 {
		if err := act.notifyRelayToReceivePackage(std, transferredPkgs, relayInfo); err != nil {
			return fmt.Errorf("failed to notify transfer completion: %w", err)
		}
		if err := act.waitForRelayReportStorage(std); err != nil {
			return fmt.Errorf("failed to wait for relay report storage: %w", err)
		}
	}

	std.InstanceData().Log().
		Zh("所有安装包处理成功").
		En("all packages processed successfully").
		Info()

	return nil
}

func (act *actionEnsurePkgToRelay) getRequiredPackages(
	std *nodeUtils.NodeActionStandarder) (*types.Release, fileiface.File, error) {

	releasePkg, err := act.getReleasePackageInfo(std.Context(), std)
	if err != nil {
		return nil, nil, fmt.Errorf("get release package: %w", err)
	}

	installPkg, err := act.getInstallerFile(std.Context(), std)
	if err != nil {
		return nil, nil, fmt.Errorf("get installer package: %w", err)
	}

	return releasePkg, installPkg, nil
}

func (act *actionEnsurePkgToRelay) determineFilesToProcess(
	std *nodeUtils.NodeActionStandarder, releasePkg *types.Release, installPkg fileiface.File) []protoRelay.FileInfo {

	files := make([]protoRelay.FileInfo, 0)

	if !std.DeployInfo().TransferOptions.SelectDownloads || std.DeployInfo().TransferOptions.EnableReleasePackage {
		files = append(files, protoRelay.FileInfo{
			FileName: releasePkg.FileName,
			FileMD5:  releasePkg.MD5,
		})
	}

	if !std.DeployInfo().TransferOptions.SelectDownloads || std.DeployInfo().TransferOptions.EnableInstaller {
		files = append(files, protoRelay.FileInfo{
			FileName: installPkg.Info().Name,
			FileMD5:  installPkg.Info().MD5,
		})
	}

	return files
}

func (act *actionEnsurePkgToRelay) queryRelayPackageState(std *nodeUtils.NodeActionStandarder,
	files []protoRelay.FileInfo, relayInfo *types.RelayInfo) error {

	event := protoRelay.CheckPkgStateReq{
		ActionName: std.InstanceData().Name,
		OperInstID: std.InstanceData().OperationInstanceID,
		FileList:   files,
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	errCh := act.proxyMessager.PushToClient(std.Context(),
		protoRelay.ServerPushEventTypeCheckPkgState, data, relayInfo.AgentID)
	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("failed to push to relay. agent-id(%s): %w", relayInfo.AgentID, err)
		}

		return nil
	case <-std.Context().Done():
		return fmt.Errorf("context cancelled. agent-id(%s): %w", relayInfo.AgentID, std.Context().Err())
	case <-time.After(queryClientTimeout):
		return fmt.Errorf("wait client timed out. agent-id(%s)", relayInfo.AgentID)
	}
}

// nolint: gocognit
func (act *actionEnsurePkgToRelay) waitForRelayReportFile(
	std *nodeUtils.NodeActionStandarder, files []protoRelay.FileInfo) (map[string]bool, string, error) {

	results := make(map[string]bool, len(files))
	for _, file := range files {
		results[file.FileName] = false
	}

	fileStorageDir := ""
	completedCount := 0

	timeoutCtx, cancel := contextx.WithTimeout(contextx.From(std.Context()), waitForRelayReportTimeout)
	defer cancel()

	ticker := time.NewTicker(waitForRelayReportInterval)
	defer ticker.Stop()

	for completedCount < len(files) {
		select {
		case <-timeoutCtx.Done():
			return results, fileStorageDir, fmt.Errorf("wait for relay report timed out. oper_inst_id(%s), action_name(%s)",
				std.InstanceData().OperationInstanceID, std.InstanceData().Name)

		case <-ticker.C:
			privateData, err := act.storageActionInstance.GetActionInstancePrivateData(
				timeoutCtx, std.InstanceData().OperationInstanceID, std.InstanceData().Name)
			if err != nil {
				return results, fileStorageDir, fmt.Errorf("failed to get private data: %w", err)
			}

			fileStateRaw, exists := privateData[relayconstant.FileStateKey]
			if !exists {
				continue
			}

			fileState, ok := fileStateRaw.(map[string]any)
			if !ok {
				return results, fileStorageDir, errors.New("unexpected type for file state")
			}

			if fileStorageDir == "" {
				if storageDirRaw, exists := fileState[relayconstant.FileStateStorageKey]; exists {
					if storageDir, ok := storageDirRaw.(string); ok && storageDir != "" {
						fileStorageDir = storageDir
						std.InstanceData().Log().
							Zh("relay 存储目录已设置, 目录(%s)", fileStorageDir).
							En("relay storage dir set. dir(%s)", fileStorageDir).
							Info()
					}
				}
			}

			for fileName, state := range fileState {
				stateStr, ok := state.(string)
				if !ok {
					return results, fileStorageDir, errors.New("unexpected type for file state")
				}

				completedCount++
				if stateStr == string(relayconstant.RelayReportPkgComplete) {
					results[fileName] = true
					std.InstanceData().Log().
						Zh("包状态完成, 文件名(%s)", fileName).
						En("package state complete. file-name(%s)", fileName).
						Info()

					continue
				}
				std.InstanceData().Log().
					Zh("包状态未完成, 文件名(%s)", fileName).
					En("package state incomplete. file-name(%s)", fileName).
					Info()
			}
		}
	}

	return results, fileStorageDir, nil
}

func (act *actionEnsurePkgToRelay) getReleasePackageInfo(nCtx contextx.IContext, std *nodeUtils.NodeActionStandarder) (*types.Release, error) {
	var release *types.Release
	switch std.DeployInfo().Host.Dynamic.NodeRole {
	case types.NodeRoleAgent:
		r, err := act.storageRelease.GetReleaseAgent(nCtx, types.ReleaseAgentKey{
			Generation: std.DeployInfo().Host.Dynamic.NodeGeneration,
			Platform:   platfmt.Platform{OS: std.DeployInfo().Host.Dynamic.NodeOsType, Arch: std.DeployInfo().Host.Dynamic.NodeCPUArch},
			Version:    std.DeployInfo().Host.Dynamic.NodeVersion,
		})
		if err != nil {
			return nil, err
		}

		release = &r.Release

	case types.NodeRoleProxy:
		r, err := act.storageRelease.GetReleaseProxy(nCtx, types.ReleaseProxyKey{
			Generation: std.DeployInfo().Host.Dynamic.NodeGeneration,
			Platform:   platfmt.Platform{OS: std.DeployInfo().Host.Dynamic.NodeOsType, Arch: std.DeployInfo().Host.Dynamic.NodeCPUArch},
			Version:    std.DeployInfo().Host.Dynamic.NodeVersion,
		})
		if err != nil {
			return nil, err
		}

		release = &r.Release

	default:
		return nil, fmt.Errorf("invalid node role. role(%s)", std.DeployInfo().Host.Dynamic.NodeRole)
	}

	std.InstanceData().Log().
		Zh("获取发布包信息, 文件名(%s)", release.FileName).
		En("get release package info. file-name(%s)", release.FileName).
		Info()

	return release, nil
}

func (act *actionEnsurePkgToRelay) getInstallerFile(
	nCtx contextx.IContext, std *nodeUtils.NodeActionStandarder) (fileiface.File, error) {

	toolName, err := tool.FormatInstallerName(std.DeployInfo().Host.Dynamic.NodeOsType, std.DeployInfo().Host.Dynamic.NodeCPUArch)
	if err != nil {
		return nil, fmt.Errorf("failed to format tools name: %w", err)
	}

	installPkgInfo, err := act.installerFileGroup.GetFile(nCtx, toolName)
	if err != nil {
		return nil, fmt.Errorf("failed to get file. installer-name(%s): %w", toolName, err)
	}

	std.InstanceData().Log().
		Zh("获取安装器文件, 文件名(%s)", installPkgInfo.Info().Name).
		En("get installer file. file-name(%s)", installPkgInfo.Info().Name).
		Info()

	return installPkgInfo, nil
}

func (act *actionEnsurePkgToRelay) transferMissingPackages(
	std *nodeUtils.NodeActionStandarder,
	releasePkg *types.Release,
	installPkg fileiface.File,
	pkgStates map[string]bool) ([]string, error) {

	var transferred []string
	gp := gopool.NewPool()

	if state, exists := pkgStates[releasePkg.FileName]; exists && !state {
		gp.Go(func() error {
			if err := act.transferReleasePkg(std, releasePkg.Type); err != nil {
				return fmt.Errorf("failed to transfer release package: %w", err)
			}
			transferred = append(transferred, releasePkg.FileName)

			return nil
		})
	} else {
		std.InstanceData().Log().
			Zh("发布包已存在, 文件名(%s)", releasePkg.FileName).
			En("release package exists. file-name(%s)", releasePkg.FileName).
			Info()
	}

	installPkgName := installPkg.Info().Name
	if state, exists := pkgStates[installPkgName]; exists && !state {
		gp.Go(func() error {
			if err := act.transferInstaller(std); err != nil {
				return fmt.Errorf("failed to transfer installer package: %w", err)
			}
			transferred = append(transferred, installPkgName)

			return nil
		})
	} else {
		std.InstanceData().Log().
			Zh("安装器包已存在, 文件名(%s)", installPkgName).
			En("installer package exists. file-name(%s)", installPkgName).
			Info()
	}

	if err := gp.Wait(); err != nil {
		return nil, err
	}

	return transferred, nil
}

func (act *actionEnsurePkgToRelay) transferReleasePkg(std *nodeUtils.NodeActionStandarder,
	rt types.ReleaseType) error {

	std.InstanceData().Log().
		Zh("正在传输发布包, relay-host-id(%d)", std.DeployInfo().RelayInfo.HostID).
		En("transferring release package. relay-host-id(%d)", std.DeployInfo().RelayInfo.HostID).
		Info()

	transferHandler, err := act.fileHandler.LaunchTransferNode(std.Context(),
		std.DeployInfo().Host.Dynamic.NodeGeneration,
		rt,
		platfmt.Platform{
			OS:   std.DeployInfo().Host.Dynamic.NodeOsType,
			Arch: std.DeployInfo().Host.Dynamic.NodeCPUArch,
		},
		std.DeployInfo().Host.Dynamic.NodeVersion,
		std.DeployInfo().RelayInfo.PackageDestDir,
		&types.Host{HostID: std.DeployInfo().RelayInfo.HostID})
	if err != nil {
		return fmt.Errorf("failed to launch transfer release. relay-host-id(%d): %w", std.DeployInfo().RelayInfo.HostID, err)
	}

	std.InstanceData().Log().
		Zh("已启动传输发布包, task-id(%s), relay-host-id(%d)", transferHandler.GetTaskID(), std.DeployInfo().RelayInfo.HostID).
		En("launched transfer release. task-id(%s), relay-host-id(%d)", transferHandler.GetTaskID(), std.DeployInfo().RelayInfo.HostID).
		Info()

	result, err := transferHandler.WaitUntilDone(std.Context())
	if err != nil {
		return fmt.Errorf("failed to wait until transfer release done. task-id(%s), relay-host-id(%d): %w",
			transferHandler.GetTaskID(), std.DeployInfo().RelayInfo.HostID, err)
	}

	if !result.Terminated {
		return fmt.Errorf("transfer release not terminated. task-id(%s), relay-host-id(%d)",
			transferHandler.GetTaskID(), std.DeployInfo().RelayInfo.HostID)
	}

	if result.ErrorCode != 0 {
		return fmt.Errorf("transfer release failed. task-id(%s), relay-host-id(%d), err-code(%d), err-msg(%s)",
			transferHandler.GetTaskID(), std.DeployInfo().RelayInfo.HostID, result.ErrorCode, result.ErrorMessage)
	}

	std.InstanceData().Log().
		Zh("传输发布包完成, task-id(%s), relay-host-id(%d)", transferHandler.GetTaskID(), std.DeployInfo().RelayInfo.HostID).
		En("transfer release done. task-id(%s), relay-host-id(%d)", transferHandler.GetTaskID(), std.DeployInfo().RelayInfo.HostID).
		Info()

	return nil
}

func (act *actionEnsurePkgToRelay) transferInstaller(
	std *nodeUtils.NodeActionStandarder) error {

	std.InstanceData().Log().
		Zh("正在传输安装器包, relay-host-id(%d)", std.DeployInfo().RelayInfo.HostID).
		En("transferring installer package. relay-host-id(%d)", std.DeployInfo().RelayInfo.HostID).
		Info()

	transferHandler, err := act.fileHandler.LaunchTransferInstaller(std.Context(),
		std.DeployInfo().Host.Dynamic.NodeGeneration,
		platfmt.Platform{
			OS:   std.DeployInfo().Host.Dynamic.NodeOsType,
			Arch: std.DeployInfo().Host.Dynamic.NodeCPUArch,
		},
		std.DeployInfo().RelayInfo.PackageDestDir,
		&types.Host{HostID: std.DeployInfo().RelayInfo.HostID})
	if err != nil {
		return fmt.Errorf("failed to launch transfer installer. relay-host-id(%d): %w", std.DeployInfo().RelayInfo.HostID, err)
	}

	std.InstanceData().Log().
		Zh("已启动传输安装器, task-id(%s), relay-host-id(%d)", transferHandler.GetTaskID(), std.DeployInfo().RelayInfo.HostID).
		En("launched transfer installer. task-id(%s), relay-host-id(%d)", transferHandler.GetTaskID(), std.DeployInfo().RelayInfo.HostID).
		Info()

	result, err := transferHandler.WaitUntilDone(std.Context())
	if err != nil {
		return fmt.Errorf("failed to wait until transfer installer done. task-id(%s), relay-host-id(%d): %w",
			transferHandler.GetTaskID(), std.DeployInfo().RelayInfo.HostID, err)
	}

	if !result.Terminated {
		return fmt.Errorf("transfer installer not terminated. task-id(%s), relay-host-id(%d)",
			transferHandler.GetTaskID(), std.DeployInfo().RelayInfo.HostID)
	}

	if result.ErrorCode != 0 {
		return fmt.Errorf("failed to transfer installer. task-id(%s), relay-host-id(%d), err-code(%d), err-msg(%s)",
			transferHandler.GetTaskID(), std.DeployInfo().RelayInfo.HostID, result.ErrorCode, result.ErrorMessage)
	}

	std.InstanceData().Log().
		Zh("传输安装器完成, task-id(%s), relay-host-id(%d)", transferHandler.GetTaskID(), std.DeployInfo().RelayInfo.HostID).
		En("transfer installer done. task-id(%s), relay-host-id(%d)", transferHandler.GetTaskID(), std.DeployInfo().RelayInfo.HostID).
		Info()

	return nil
}

func (act *actionEnsurePkgToRelay) notifyRelayToReceivePackage(
	std *nodeUtils.NodeActionStandarder, pkgNames []string, relayInfo *types.RelayInfo) error {

	event := protoRelay.NotifyReceiveReq{
		ActionName: std.InstanceData().Name,
		OperInstID: std.InstanceData().OperationInstanceID,
		PkgName:    pkgNames}
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	errCh := act.proxyMessager.PushToClient(std.Context(),
		protoRelay.ServerPushEventTypeNotifyReceive, data, relayInfo.AgentID)
	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("failed to notify relay to receive. agent-id(%s): %w", relayInfo.AgentID, err)
		}

		return nil
	case <-std.Context().Done():
		return fmt.Errorf("context cancelled. agent-id(%s): %w", relayInfo.AgentID, std.Context().Err())
	case <-time.After(queryClientTimeout):
		return fmt.Errorf("wait client timed out. agent-id(%s)", relayInfo.AgentID)
	}
}

func (act *actionEnsurePkgToRelay) waitForRelayReportStorage(
	std *nodeUtils.NodeActionStandarder) error {

	timeoutCtx, cancel := contextx.WithTimeout(contextx.From(std.Context()), waitForRelayReportTimeout)
	defer cancel()

	ticker := time.NewTicker(waitForRelayReportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			return fmt.Errorf("wait for relay report storage result timed out. oper_inst_id(%s), action_name(%s)",
				std.InstanceData().OperationInstanceID, std.InstanceData().Name)

		case <-ticker.C:
			privateData, err := act.storageActionInstance.GetActionInstancePrivateData(
				timeoutCtx, std.InstanceData().OperationInstanceID, std.InstanceData().Name)
			if err != nil {
				logger.G.Sys().Ctx(timeoutCtx).
					WithErr(err).
					With("oper-inst-id", std.InstanceData().OperationInstanceID, "action-name", std.InstanceData().Name).
					Error("failed to get private data")

				continue
			}

			relayStorageResultRaw, exists := privateData[relayconstant.StorageResultKey]
			if !exists {
				continue
			}

			relayStorageResult, ok := relayStorageResultRaw.(map[string]any)
			if !ok {
				return errors.New("unexpected type for relay storage result")
			}

			errMsgRaw := relayStorageResult[relayconstant.StorageResultErrMsgKey]
			errMsg, ok := errMsgRaw.(string)
			if !ok {
				return errors.New("unexpected type for relay storage result message")
			}

			std.InstanceData().Log().
				Zh("等待 relay 报告存储结果完成").
				En("wait for relay report storage result done").
				Info()

			if errMsg == "" {
				return nil
			}

			return errors.New(errMsg)
		}
	}
}
