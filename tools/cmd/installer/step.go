/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package main

import (
	"context"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/checkdeploy"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/datareporter"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/filedownloader"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/nodeinstaller"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/precheck"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/startnode"
)

func stepInstallNode(ctx context.Context) error {
	step := nodeinstaller.NewStep(nodeinstaller.StepArgs{
		AgentID:           GetNodeAgentID(),
		ReRegisterAgentID: GetReRegisterAgentID() && GetReinstall(),
		SetupDirPath:      GetSetupDir(),
		PkgPath:           GetGsePkgPath(),
		SrcConfigDir:      GetTmpConfigDir(),
		// notice: It is now available, but there is no open configuration.
		Overwrite: false,
	})
	agentID, err := step.Run(ctx)
	if err != nil {
		return fmt.Errorf("install step failed, err: %v", err)
	}

	if err := SetNodeAgentID(agentID); err != nil {
		return fmt.Errorf("set node agent id failed, err: %v", err)
	}

	return nil
}

func stepCheckDeploy(ctx context.Context) error {
	step := checkdeploy.NewStep(checkdeploy.StepArgs{
		RunDir:    GetRunDir(),
		NodeRole:  GetNodeRole(),
		DeployEnv: GetDeployEnv(),
	})
	if err := step.Run(ctx); err != nil {
		return fmt.Errorf("check deploy step failed, err: %w", err)
	}

	return nil
}

func stepStartNode(ctx context.Context) error {
	step := startnode.NewStep(startnode.StepArgs{
		GseCtlPath: GetGseCtlPath(),
	})
	if err := step.Run(ctx); err != nil {
		return fmt.Errorf("start node step failed, err: %w", err)
	}

	return nil
}

func stepDownloadFiles(ctx context.Context) error {
	step := filedownloader.NewStep(filedownloader.StepArgs{
		DownloadPoint:        GetDownloadEndPoint(),
		CallbackEndpoint:     GetCallBackEndpoint(),
		PkgGeneration:        GetNodePkgGeneration(),
		PkgPath:              GetGsePkgPath(),
		PkgVersion:           GetNodePkgVersion(),
		NodeRole:             GetNodeRole(),
		Token:                GetToken(),
		TmpAgentConfPath:     GetTmpAgentConfPath(),
		TmpFileProxyConfPath: GetTmpFileProxyConfPath(),
		TmpDataProxyConfPath: GetTmpDataProxyConfPath(),
		CheckListPath:        GetPreCheckFilePath(),
	})
	if err := step.Run(ctx); err != nil {
		return fmt.Errorf("download files step failed, err: %w", err)
	}

	return nil
}

func stepPreCheck(ctx context.Context) error {
	step := precheck.NewStep(precheck.StepArgs{
		PreCheckListPath: GetPreCheckFilePath(),
		SetupDirPath:     GetSetupDir(),
	})

	if err := step.Run(ctx); err != nil {
		return fmt.Errorf("precheck step failed, err: %w", err)
	}

	return nil
}

func stepReportData(ctx context.Context) error {
	step := datareporter.NewStep(datareporter.StepArgs{
		Token:            GetToken(),
		AgentID:          GetNodeAgentID(),
		CallbackEndpoint: GetCallBackEndpoint(),
	})
	if err := step.Run(ctx); err != nil {
		return fmt.Errorf("report data step failed, err: %w", err)
	}

	return nil
}
