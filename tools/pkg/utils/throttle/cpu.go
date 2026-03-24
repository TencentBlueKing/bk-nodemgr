/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package throttle

import "time"

// processCPUTime returns user+system CPU time of the current process.
// It relies on platform-specific readProcCPU (getrusage on Unix,
// GetProcessTimes on Windows). Returns 0 if the underlying syscall fails.
func processCPUTime() time.Duration {
	var usage [2]int64
	readProcCPU(&usage)
	return time.Duration(usage[0]+usage[1]) * time.Nanosecond
}
