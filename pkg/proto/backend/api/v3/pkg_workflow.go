/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package v3

import (
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

// Validate checks the plugin package export request.
func (x *PackageExportPluginReq) Validate() error {
	if conv.IsEmpty(strings.TrimSpace(x.GetPluginPkgName())) {
		return errors.New("plugin_pkg_name is required")
	}
	if conv.IsEmpty(strings.TrimSpace(x.GetPluginPkgVersion())) {
		return errors.New("plugin_pkg_version is required")
	}
	return nil
}

// AutoConvert converts the plugin package export request.
func (x *PackageExportPluginReq) AutoConvert() {}

// Validate checks the plugin package export result request.
func (x *PackageExportResultReq) Validate() error {
	if conv.IsEmpty(strings.TrimSpace(x.GetWorkflowId())) {
		return errors.New("workflow_id is required")
	}
	return nil
}

// AutoConvert converts the plugin package export result request.
func (x *PackageExportResultReq) AutoConvert() {}

// ConvertResultFromTypes converts an export result to the API response.
func (x *PackageExportResultResp) ConvertResultFromTypes(workflow *types.PackageWorkflow, downloadURL string, downloadURLExpiredAt time.Time) {
	data := &PackageExportResultResp_Data{}
	if workflow != nil {
		data.Status = string(workflow.Status)
		data.IsFinish = slices.Contains(types.GetFinishedPackageWorkflowStatus(), workflow.Status)
	}
	if downloadURL != "" {
		data.DownloadUrl = &downloadURL
		expiredAt := downloadURLExpiredAt.UnixMilli()
		data.DownloadUrlExpiredAt = &expiredAt
	}
	x.Data = data
}

// Validate check body.
func (x *PackageImportPluginV3PkgReq) Validate() error {
	if conv.IsEmpty(x.GetFilename()) {
		return errors.New("filename is required")
	}

	if conv.IsEmpty(strings.TrimSpace(x.GetDownloadUrl())) {
		return errors.New("download_url is required")
	}

	if conv.IsEmpty(x.GetMd5()) {
		return errors.New("md5 is required")
	}

	return nil
}

// AutoConvert auto convert.
func (x *PackageImportPluginV3PkgReq) AutoConvert() {
}

// Validate checks the plugin v2 package import request body.
func (x *PackageImportPluginV2PkgReq) Validate() error {
	if conv.IsEmpty(x.GetFilename()) {
		return errors.New("filename is required")
	}

	if conv.IsEmpty(strings.TrimSpace(x.GetDownloadUrl())) {
		return errors.New("download_url is required")
	}

	if conv.IsEmpty(x.GetMd5()) {
		return errors.New("md5 is required")
	}

	return nil
}

// AutoConvert auto converts the plugin v2 package import request.
func (x *PackageImportPluginV2PkgReq) AutoConvert() {
}

// Validate checks the external plugin v2 package import request body.
func (x *PackageImportExternalPluginV2PkgReq) Validate() error {
	if conv.IsEmpty(x.GetFilename()) {
		return errors.New("filename is required")
	}

	if conv.IsEmpty(strings.TrimSpace(x.GetDownloadUrl())) {
		return errors.New("download_url is required")
	}

	if conv.IsEmpty(x.GetMd5()) {
		return errors.New("md5 is required")
	}

	return nil
}

// AutoConvert auto converts the external plugin v2 package import request.
func (x *PackageImportExternalPluginV2PkgReq) AutoConvert() {
}

// Validate checks the request body.
func (x *PackageImportResultReq) Validate() error {
	if conv.IsEmpty(x.GetWorkflowId()) {
		return errors.New("workflow_id is required")
	}

	return nil
}

// AutoConvert auto converts the request body.
func (x *PackageImportResultReq) AutoConvert() {
}

// ConvertResultFromTypes converts a package workflow and its operation instances to the response.
func (x *PackageImportResultResp) ConvertResultFromTypes(
	workflow *types.PackageWorkflow, results []*operation.InstanceData) {

	data := &PackageImportResultResp_Data{}
	if workflow != nil {
		data.Status = string(workflow.Status)
		data.IsFinish = slices.Contains(types.GetFinishedPackageWorkflowStatus(), workflow.Status)
	}

	operations := make([]*PackageImportResultResp_OperationResult, 0, len(results))
	for _, result := range results {
		if result == nil {
			continue
		}

		operations = append(operations, &PackageImportResultResp_OperationResult{
			OperationId:        result.Metadata.OperationID,
			LastInstanceId:     result.Metadata.OperationInstanceID,
			OperInstLogs:       convertActionInstanceLogs(result),
			ExtraExecutionLogs: convertExtraExecutionLogs(result),
		})
	}
	data.Operations = operations
	x.Data = data
}

// convertActionInstanceLogs converts action instance logs from the operation instance data.
func convertActionInstanceLogs(result *operation.InstanceData) map[string]*WorkflowActionData {
	operInstLogs := make(map[string]*WorkflowActionData, len(result.ActionInstanceDataMap))
	for actionID, actionData := range result.ActionInstanceDataMap {
		messages := make([]*WorkflowActionMessage_Message, len(actionData.Messages))
		for idx, message := range actionData.Messages {
			messages[idx] = &WorkflowActionMessage_Message{
				Time:   message.Time.UnixMilli(),
				Level:  message.Level,
				TextZh: message.TextZh,
				TextEn: message.TextEn,
			}
		}

		operInstLogs[actionID] = &WorkflowActionData{
			LifeCycle: &WorkflowLifeCycle{
				State:      string(actionData.Lifecycle.State),
				CreateTime: timeToUnixMilli(actionData.Lifecycle.CreatedAt),
				StartTime:  timeToUnixMilli(actionData.Lifecycle.StartedAt),
				EndTime:    timeToUnixMilli(actionData.Lifecycle.EndedAt),
			},
			Message:         &WorkflowActionMessage{Logs: messages},
			DisplayNameZh:   actionData.DisplayNameZh,
			DisplayNameEn:   actionData.DisplayNameEn,
			SubWorkflowRefs: subWorkflowRefsFromPrivateData(actionData.PrivateData),
		}
	}

	return operInstLogs
}

// convertExtraExecutionLogs converts extra execution messages from the operation instance data.
func convertExtraExecutionLogs(result *operation.InstanceData) *WorkflowActionMessage {
	extraExecutionLogs := &WorkflowActionMessage{
		Logs: make([]*WorkflowActionMessage_Message, len(result.Metadata.ExtraExecutionMessages)),
	}
	for idx, message := range result.Metadata.ExtraExecutionMessages {
		extraExecutionLogs.Logs[idx] = &WorkflowActionMessage_Message{
			Time:   message.Time.UnixMilli(),
			Level:  message.Level,
			TextZh: message.TextZh,
			TextEn: message.TextEn,
		}
	}

	return extraExecutionLogs
}
