/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package deploypolicy this package provide deploy policy relate action definition.
package deploypolicy

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/dpmgr"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/deploypolicy"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
)

// Capability encapsulates the various capabilities the service supports.
type Capability struct {
	DPMgr               dpmgr.IHandler
	StorageDeployPolicy deploypolicy.IStorage
	WorkflowCtl         workflow.IController
}
