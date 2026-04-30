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

import "time"

const (
	// syncDataCleanPolicyTimeoutMultiplier defines how long completed sync data operations are kept.
	syncDataCleanPolicyTimeoutMultiplier = 3.0

	// syncDataCleanPolicyHoursPerDay defines the hour-to-day conversion for trigger clean policy.
	syncDataCleanPolicyHoursPerDay = 24.0
)

func syncDataCleanPolicyMaxDays(timeout time.Duration) float64 {
	// This is trigger retention after the trigger becomes inactive, derived from generated operation timeout.
	return timeout.Hours() * syncDataCleanPolicyTimeoutMultiplier / syncDataCleanPolicyHoursPerDay
}
