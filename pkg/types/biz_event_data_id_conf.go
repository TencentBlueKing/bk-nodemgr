/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package types

import "fmt"

// BizEventDataIDConf defines event data-id configuration scoped by business.
type BizEventDataIDConf struct {
	BizID                     int64
	AgentBaseAlarmEventDataID *int64
	TaskProcEventDataID       *int64
}

// Validate validates structural event data-id values.
func (conf BizEventDataIDConf) Validate() error {
	if conf.BizID <= 0 {
		return fmt.Errorf("bk_biz_id must be positive, got %d", conf.BizID)
	}

	if conf.AgentBaseAlarmEventDataID != nil {
		if *conf.AgentBaseAlarmEventDataID <= 0 {
			return fmt.Errorf("agentBaseAlarmEventDataID must be positive, got %d", *conf.AgentBaseAlarmEventDataID)
		}
	}

	if conf.TaskProcEventDataID != nil {
		if *conf.TaskProcEventDataID <= 0 {
			return fmt.Errorf("taskProcEventDataID must be positive, got %d", *conf.TaskProcEventDataID)
		}
	}

	return nil
}
