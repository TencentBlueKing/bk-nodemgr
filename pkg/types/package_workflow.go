/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package types

import (
	"fmt"
	"time"
)

// PackageWorkflow stores the lifecycle data of a package workflow.
type PackageWorkflow struct {
	TenantID    string
	WorkflowID  string
	TriggerID   string
	Type        PackageWorkflowType
	Operator    string
	OperateTime time.Time
	FinishTime  time.Time
	Status      PackageWorkflowStatus
}

// PackageWorkflowType defines the type of a package workflow.
type PackageWorkflowType string

const (
	// PackageWorkflowTypeImport defines the package import workflow type.
	PackageWorkflowTypeImport PackageWorkflowType = "import_package"
)

// Validate validates the package workflow type.
func (typ PackageWorkflowType) Validate() error {
	switch typ {
	case PackageWorkflowTypeImport:
		return nil
	default:
		return fmt.Errorf("package workflow type is invalid, type(%s)", typ)
	}
}

// PackageWorkflowStatus defines the status of a package workflow.
type PackageWorkflowStatus string

const (
	// PackageWorkflowStatusRunning indicates the workflow is running.
	PackageWorkflowStatusRunning PackageWorkflowStatus = "running"

	// PackageWorkflowStatusSuccess indicates the workflow is successful.
	PackageWorkflowStatusSuccess PackageWorkflowStatus = "success"

	// PackageWorkflowStatusFailed indicates the workflow failed.
	PackageWorkflowStatusFailed PackageWorkflowStatus = "failed"

	// PackageWorkflowStatusPartialFailed indicates part of workflow operations failed.
	PackageWorkflowStatusPartialFailed PackageWorkflowStatus = "partial_failed"
)

// Validate validates the package workflow status.
func (status PackageWorkflowStatus) Validate() error {
	switch status {
	case PackageWorkflowStatusRunning, PackageWorkflowStatusSuccess,
		PackageWorkflowStatusFailed, PackageWorkflowStatusPartialFailed:
		return nil
	default:
		return fmt.Errorf("package workflow status is invalid, status(%s)", status)
	}
}

// GetFinishedPackageWorkflowStatus returns all finished package workflow statuses.
func GetFinishedPackageWorkflowStatus() []PackageWorkflowStatus {
	return []PackageWorkflowStatus{
		PackageWorkflowStatusSuccess,
		PackageWorkflowStatusFailed,
		PackageWorkflowStatusPartialFailed,
	}
}
