/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package pkg

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

// PackagePluginV3Import launches package import workflow.
func (h *handler) PackagePluginV3Import(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PackageImportPluginV3PkgReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to import package, failed to decode request body")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	deployment := h.generatePackageImportDeployment(req.GetFilename(), req.GetDownloadUrl(), req.GetMd5())
	workflowID, err := h.pkgMgrIface.LaunchPackageImportPluginV3Pkg(rCtx, types.PackageImportParam{
		Type:               types.PackageWorkflowTypeImport,
		PackageDeployments: []*types.PackageDeployment{deployment},
		Operator:           rCtx.BKUsername(),
	})
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to import package, failed to launch workflow")

		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	logger.G.Biz(rCtx).With("workflow-id", workflowID).Info("launched package import workflow")

	resp := &protoBackend.PackageImportPluginV3PkgResp{
		Data: &protoBackend.PackageImportPluginV3PkgResp_Data{
			WorkflowId: workflowID,
		},
	}

	return resp.GetData(), nil
}

// PackagePluginV2Import launches package import workflow for an official v2 plugin package.
func (h *handler) PackagePluginV2Import(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PackageImportPluginV2PkgReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to import package, failed to decode request body")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	deployment := h.generatePackageImportDeployment(req.GetFilename(), req.GetDownloadUrl(), req.GetMd5())
	workflowID, err := h.pkgMgrIface.LaunchPackageImportPluginV2Pkg(rCtx, types.PackageImportParam{
		Type:               types.PackageWorkflowTypeImport,
		PackageDeployments: []*types.PackageDeployment{deployment},
		Operator:           rCtx.BKUsername(),
	})
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to import package, failed to launch workflow")

		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	logger.G.Biz(rCtx).With("workflow-id", workflowID).Info("launched package import workflow")

	resp := &protoBackend.PackageImportPluginV2PkgResp{
		Data: &protoBackend.PackageImportPluginV2PkgResp_Data{
			WorkflowId: workflowID,
		},
	}

	return resp.GetData(), nil
}

// PackageExternalPluginV2Import launches package import workflow for an external v2 plugin package.
func (h *handler) PackageExternalPluginV2Import(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PackageImportExternalPluginV2PkgReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to import package, failed to decode request body")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	workflowID, err := h.pkgMgrIface.LaunchPackageImportExternalPluginV2Pkg(rCtx, types.PackageImportParam{
		Type: types.PackageWorkflowTypeImport,
		PackageDeployments: []*types.PackageDeployment{
			h.generatePackageImportDeployment(req.GetFilename(), req.GetDownloadUrl(), req.GetMd5()),
		},
		Operator: rCtx.BKUsername(),
	})
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to import package, failed to launch workflow")

		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	logger.G.Biz(rCtx).With("workflow-id", workflowID).Info("launched package import workflow")

	resp := &protoBackend.PackageImportExternalPluginV2PkgResp{
		Data: &protoBackend.PackageImportExternalPluginV2PkgResp_Data{
			WorkflowId: workflowID,
		},
	}

	return resp.GetData(), nil
}

func (h *handler) generatePackageImportDeployment(filename, downloadURL, md5 string) *types.PackageDeployment {
	deployment := types.NewPackageDeployment(&types.PackageDeploymentInfo{
		ImportPluginPkgOptions: types.PackageImportPluginPkgOptions{
			FileSourceType: types.FileSourceTypeDownload,
			FileSource:     downloadURL,
			FileName:       filename,
			MD5:            md5,
		},
	})

	return deployment
}

// PackageImportResult gets the package import result.
func (h *handler) PackageImportResult(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PackageImportResultReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get package import result, failed to decode request body")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	packageWorkflow, err := h.daoPackageWorkflow.GetPackageWorkflow(rCtx, req.GetWorkflowId())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get package workflow")

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}
	operations, _, err := h.daoWorkflow.ListOperation(rCtx, types.UnlimitedPage(), &types.OperationCondition{
		ExactInclude: &types.OperationExactFields{
			TriggerID: []string{packageWorkflow.TriggerID},
		},
	})
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list package workflow operations")

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	instances := make([]*operation.InstanceData, 0, len(operations))
	for _, oper := range operations {
		operInstID := oper.GetLastInstanceID()
		if operInstID == "" {
			continue
		}

		instance, err := h.daoWorkflow.GetOperationInstanceFullData(rCtx, operInstID)
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to get package workflow operation instance")

			return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
		}
		instances = append(instances, instance)
	}

	resp := new(protoBackend.PackageImportResultResp)
	resp.ConvertResultFromTypes(packageWorkflow, instances)

	return resp.GetData(), nil
}
