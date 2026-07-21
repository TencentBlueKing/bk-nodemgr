/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package action

import (
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/retrier"
)

// DefaultBackoffDelayFn returns the default backoff delay function.
func DefaultBackoffDelayFn(def Definition, attempt int) func() {
	return func() {
		delay := retrier.CalculateExpoDelay(attempt, retrier.ExpoBackoffDelayOpts{
			BaseDelay:     def.Timeout() / time.Duration(def.MaxRetryCount()+1) / 2,
			MaxDelay:      def.Timeout() / time.Duration(def.MaxRetryCount()+1),
			JitterPercent: 0.2, // nolint: mnd
		})
		time.Sleep(delay)
	}
}
