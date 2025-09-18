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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	workoper "github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

// IStorage defines the interface of schedule workflow storage.
type IStorage interface {
	basestorage.Interface
	workflow.IStorageTrigger
	workflow.IStorageActionInstance
	workflow.IStorageOperation
	workflow.IStorageOperationInstance
	workflow.IStorageSchedule

	IDomainNodeInstall
}

// IDomainNodeInstall defines the interface for domain node installation related operations.
type IDomainNodeInstall interface {
	// ListOperationByNodeWorkflowOperationCondition lists operations by condition with pagination support.
	ListOperationByNodeWorkflowOperationCondition(
		ctx contextx.IContext, page types.Page, condition ...*types.NodeWorkflowOperationCondition) (
		[]*workoper.Operation, int64, error)
}
