/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package syncdata

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operengine"
	"time"
)

// OperDefNameSyncHost sync specific biz's host from cmdb.
const OperDefNameSyncHost = "oper_def_sync_host"

// OperSyncHost the params of OperSyncHost
type OperSyncHost struct {
	TenantID string `json:"tenant_id"`
	BizID    int64  `json:"biz_id"`
}

// OperDef the operdef of OperSyncHost.
func (oper *OperSyncHost) OperDef() operengine.OperDefSnapshot {
	return operengine.OperDefSnapshot{
		OperDefName: OperDefNameSyncHost,
		ActionNames: []string{ActionNameSyncHostFromCMDB},
	}
}

// Param the param of OperSyncHost.
func (oper *OperSyncHost) Param() operengine.OperInstParam {
	return operengine.OperInstParam{
		Timeout:     10 * time.Minute, // nolint: mnd
		InitContent: conv.StructToMapIgnoreError(oper),
	}
}
