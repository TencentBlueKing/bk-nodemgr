/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package bizeventdataidconf provides backend storage for business event data-id configs.
package bizeventdataidconf

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IStorage defines business event data-id config storage operations.
type IStorage interface {
	basestorage.Interface
	IDaoBizEventDataIDConf
}

// IDaoBizEventDataIDConf defines single-table business event data-id config operations.
type IDaoBizEventDataIDConf interface {
	// GetBizEventDataIDConf gets business event data-id config by business ID.
	GetBizEventDataIDConf(nCtx contextx.IContext, bkBizID int64) (types.BizEventDataIDConf, bool, error)

	// UpdateBizAgentBaseAlarmEventDataID updates agent base alarm event data-id by business ID.
	UpdateBizAgentBaseAlarmEventDataID(nCtx contextx.IContext, bkBizID, eventDataID int64) error
}
