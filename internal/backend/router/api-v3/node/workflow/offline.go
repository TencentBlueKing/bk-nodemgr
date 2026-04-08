/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package workflow

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/node"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/node/utils"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/nodeconfig"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/nodepkg"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/tool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/installer"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/system"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

// offlineInstallMetadata is the structure of metadata.json in the offline package.
type offlineInstallMetadata struct {
	InstanceID  string `json:"instance_id"`
	AgentID     string `json:"agent_id"`
	TargetIP    string `json:"target_ip"`
	NodeVersion string `json:"node_version"`
	Generation  int64  `json:"generation"`
	CreatedAt   string `json:"created_at"`
}

// GetOfflineInstallInfo returns the offline install info for the given operation.
// nolint: funlen, gocognit, gocyclo, cyclop
func (h *handler) GetOfflineInstallInfo(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.NodeWorkflowOperationOfflineInstallInfoGetReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get offline install info, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if err := req.Validate(); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get offline install info, invalid request")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	// load operation and validate it is an offline install operation.
	oper, err := h.storageWorkflow.GetOperation(rCtx, req.GetOperationId())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get offline install info, failed to get operation")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	if oper.Definition.Name() != node.OperDefNameInstallProxyByOffline {
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter,
			fmt.Errorf("operation(%s) is not an offline install operation, got def(%s)",
				req.GetOperationId(), oper.Definition.Name()))
	}

	// extract token and load node deployment data.
	param := new(utils.NodeActionStandardParam)
	if err = conv.MapToStruct(oper.Param.InitContent, param); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get offline install info, failed to extract token")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	nodeConf, err := h.daoNodeDeployment.GetNodeDeploymentNodeConf(rCtx, param.Token)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get offline install info, failed to get node conf")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	deployInfo, err := h.daoNodeDeployment.GetNodeDeploymentInfo(rCtx, param.Token)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get offline install info, failed to get deployment info")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	// validate WaitOfflineManualInstall action is in running state.
	lastInstID := oper.GetLastInstanceID()
	if lastInstID == "" {
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter,
			fmt.Errorf("operation(%s) has no instance", req.GetOperationId()))
	}

	lifecycle, err := h.storageWorkflow.GetActionInstanceLifecycle(
		rCtx, lastInstID, node.ActionNameWaitOfflineManualInstall)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get offline install info, failed to get action lifecycle")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	if lifecycle.State != action.StateRunning {
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter,
			fmt.Errorf("operation(%s) is not waiting for offline install, current state: %s",
				req.GetOperationId(), lifecycle.State))
	}

	// render GSE config files.
	configKeyToFilename := map[string]string{
		types.ConfigKeyAgent: installer.OfflineGseAgentConfFileName,
		types.ConfigKeyFile:  installer.OfflineGseFileProxyConfFileName,
		types.ConfigKeyData:  installer.OfflineGseDataProxyConfFileName,
	}

	configs := make(map[string]string, len(configKeyToFilename))
	for key, filename := range configKeyToFilename {
		rendered, renderErr := nodeconfig.RenderNodeConfig(key, nodeConf)
		if renderErr != nil {
			logger.G.Biz(rCtx).WithErr(renderErr).Error("failed to render node config. key: %s", key)
			return nil, resterrf.ErrWrap(resterrf.InvalidParameter, renderErr)
		}

		jsonBytes, marshalErr := json.Marshal(rendered)
		if marshalErr != nil {
			return nil, resterrf.ErrWrap(resterrf.InvalidParameter, marshalErr)
		}

		configs[filename] = string(jsonBytes)
	}

	// build precheck.json content.
	checkList, err := nodeconfig.BuildCheckList(deployInfo, nodeConf)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get offline install info, failed to build check list")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	precheckBytes, err := json.Marshal(checkList)
	if err != nil {
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	installerFileName, err := tool.FormatInstallerName(
		deployInfo.Host.Dynamic.NodeOsType, deployInfo.Host.Dynamic.NodeCPUArch)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get offline install info, failed to format installer file name")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	// build install.sh content.
	installScript := buildOfflineInstallScript(
		deployInfo, deployInfo.InstallerRuntime.BaseWorkDir, deployInfo.BaseRuntime.BaseDeployDir, param.Token, lastInstID, installerFileName)

	// build metadata.json content.
	targetIP := ""
	if len(deployInfo.Host.Static.InnerIPList) > 0 {
		targetIP = deployInfo.Host.Static.InnerIPList[0]
	}

	metadata := offlineInstallMetadata{
		InstanceID:  lastInstID,
		AgentID:     string(deployInfo.Host.Dynamic.AgentID),
		TargetIP:    targetIP,
		NodeVersion: string(deployInfo.Host.Dynamic.NodeVersion),
		Generation:  int64(deployInfo.Host.Dynamic.NodeGeneration),
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
	}

	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	// format release package name.
	releaseType, err := types.ConvertNodeRoleToReleaseType(deployInfo.Host.Dynamic.NodeRole)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get offline install info, failed to convert node role")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	osType := deployInfo.Host.Dynamic.NodeOsType
	cpuArch := deployInfo.Host.Dynamic.NodeCPUArch
	version := string(deployInfo.Host.Dynamic.NodeVersion)
	generation := deployInfo.Host.Dynamic.NodeGeneration

	// pkgFileName is computed only to verify the release package is available
	// (FormatPkgFileName validates generation/platform/version).
	plat := platfmt.NewPlatform(osType, cpuArch)
	if _, err = nodepkg.FormatPkgFileName(generation, releaseType, plat, version); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get offline install info, failed to format package name")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	packageName := buildOfflinePackageStem(deployInfo)

	return &protoBackend.NodeWorkflowOperationOfflineInstallInfoGetResp_Data{
		InstallScript: installScript,
		Metadata:      string(metadataBytes),
		Configs:       configs,
		Precheck:      string(precheckBytes),
		ReleaseInfo: &protoBackend.NodeWorkflowOperationOfflineInstallInfoGetResp_ReleaseInfo{
			Generation: int64(generation),
			OsType:     string(osType),
			CpuArch:    string(cpuArch),
			Version:    version,
		},
		InstallerInfo: &protoBackend.NodeWorkflowOperationOfflineInstallInfoGetResp_InstallerInfo{
			Generation: int64(generation),
			OsType:     string(osType),
			CpuArch:    string(cpuArch),
		},
		PackageName: packageName,
	}, nil
}

// SubmitOfflineInstallResult submits the offline install result for the given operation.
func (h *handler) SubmitOfflineInstallResult(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.NodeWorkflowOperationOfflineInstallResultSubmitReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to submit offline install result, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	// load operation and validate it is an offline install operation.
	oper, err := h.storageWorkflow.GetOperation(rCtx, req.GetOperationId())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to submit offline install result, failed to get operation")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	if oper.Definition.Name() != node.OperDefNameInstallProxyByOffline {
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter,
			fmt.Errorf("operation(%s) is not an offline install operation, got def(%s)",
				req.GetOperationId(), oper.Definition.Name()))
	}

	// validate WaitOfflineManualInstall action is in running state.
	lastInstID := oper.GetLastInstanceID()
	if lastInstID == "" {
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter,
			fmt.Errorf("operation(%s) has no instance", req.GetOperationId()))
	}

	lifecycle, err := h.storageWorkflow.GetActionInstanceLifecycle(
		rCtx, lastInstID, node.ActionNameWaitOfflineManualInstall)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to submit offline install result, failed to get action lifecycle")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	if lifecycle.State != action.StateRunning {
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter,
			fmt.Errorf("operation(%s) is not waiting for offline install, current state: %s",
				req.GetOperationId(), lifecycle.State))
	}

	// upsert action private data to unblock the waiting action.
	if err = h.storageWorkflow.UpsertActionInstancePrivateData(
		rCtx,
		lastInstID,
		node.ActionNameWaitOfflineManualInstall,
		map[string]any{
			types.PDKeyOfflineInstallResult: req.GetResultData(),
		},
	); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to submit offline install result, failed to upsert private data")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	return nil, nil //nolint:nilnil
}

// buildOfflineInstallScript builds the offline install.sh shell script content.
func buildOfflineInstallScript(
	deployInfo *types.DeploymentInfo,
	baseWorkDir string,
	baseDeployDir string,
	deployToken string,
	operInstID string,
	installerFileName string,
) string {

	deployEnv := system.GetEnv()
	// dataDir mirrors installer's persistentVars.DataDir = baseWorkDir/deployEnv/data.
	// Files from the extracted package's data/ must be placed here before the installer runs.
	dataDir := fmt.Sprintf("%s/%s/data", baseWorkDir, deployEnv)

	args := []string{
		fmt.Sprintf("--deploy_env %s", deployEnv),
		fmt.Sprintf("--generation %d", deployInfo.Host.Dynamic.NodeGeneration),
		fmt.Sprintf("--node_role %s", deployInfo.Host.Dynamic.NodeRole),
		fmt.Sprintf("--base_work_dir %s", baseWorkDir),
		fmt.Sprintf("--base_deploy_dir %s", baseDeployDir),
		fmt.Sprintf("--deploy_token %s", deployToken),
		fmt.Sprintf("--node_version %s", deployInfo.Host.Dynamic.NodeVersion),
		fmt.Sprintf("--oper_inst_id %s", operInstID),
		"--skip_callback",
		"--skip_download",
		// Match manual install: stream installer progress to the shell (action_install_node_by_manual.buildCMD).
		"--log_to_std",
	}

	if deployInfo.Host.Dynamic.AgentID != "" && !isWindows(deployInfo.Host.Dynamic.NodeOsType) {
		args = append(args, fmt.Sprintf("--agent_id %s", deployInfo.Host.Dynamic.AgentID))
	}

	// Installer expects: release package, precheck.json, and configs in DataDir.
	// Copy all content from the package's data/ directory into the expected location.
	lines := []string{
		"#!/bin/bash",
		`set -e`,
		`SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"`,
		fmt.Sprintf(`DATA_DIR="%s"`, dataDir),
		`mkdir -p "${DATA_DIR}"`,
		fmt.Sprintf(`cp -rf "${SCRIPT_DIR}/%s/." "${DATA_DIR}/"`, installer.OfflinePkgRelPathData),
		fmt.Sprintf(`chmod +x "${SCRIPT_DIR}/%s"`, installerFileName),
		fmt.Sprintf(`"${SCRIPT_DIR}/%s" `, installerFileName) + installer.NodeCmdFullInstall + ` \`,
	}

	for i, arg := range args {
		if i < len(args)-1 {
			lines = append(lines, "  "+arg+` \`)
		} else {
			lines = append(lines, "  "+arg)
		}
	}

	// After the installer completes (--skip_callback), results are written to installer.data.json.
	// Print the file content so the user can copy and paste it into the management portal.
	lines = append(lines,
		fmt.Sprintf(`echo "--- %s ---"`, installer.DataFileName),
		fmt.Sprintf(`cat "%s/%s"`, dataDir, installer.DataFileName),
	)

	return strings.Join(lines, "\n") + "\n"
}

func isWindows(osType criteria.OSType) bool {
	return osType == criteria.OSWindows
}

// buildOfflinePackageStem returns the top-level directory / download basename stem for the offline bundle:
// {prefix}-{network_area_id}-{ip_slugs}, where IPv4 dots and IPv6 colons become '-'.
func buildOfflinePackageStem(deployInfo *types.DeploymentInfo) string {
	networkAreaID := deployInfo.Host.Static.NetworkAreaID
	ip := ""
	if len(deployInfo.Host.Static.InnerIPList) > 0 {
		ip = deployInfo.Host.Static.InnerIPList[0]
	}
	ipSlug := strings.ReplaceAll(strings.ReplaceAll(ip, ".", "-"), ":", "-")
	if ipSlug == "" {
		ipSlug = "unknown"
	}

	return fmt.Sprintf("%s-%d-%s", installer.OfflinePackageNamePrefix, networkAreaID, ipSlug)
}
