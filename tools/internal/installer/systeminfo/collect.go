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

package systeminfo

import (
	"fmt"
	"os"
	"runtime"
	"time"
)

func collectGOOS() (string, error) {
	return runtime.GOOS, nil
}

func collectGOARCH() (string, error) {
	return runtime.GOARCH, nil
}

func collectHostname() (string, error) {
	return os.Hostname()
}

func collectCurrentTime() (string, error) {
	return time.Now().Format(time.RFC3339), nil
}

func collectTimezone() (string, error) {
	name, _ := time.Now().Zone()
	return name, nil
}

func collectTimezoneOffset() (string, error) {
	_, offset := time.Now().Zone()
	return formatTimezoneOffset(offset), nil
}

func formatTimezoneOffset(offsetSeconds int) string {
	sign := "+"
	if offsetSeconds < 0 {
		sign = "-"
		offsetSeconds = -offsetSeconds
	}

	hours := offsetSeconds / 3600
	minutes := offsetSeconds % 3600 / 60
	return fmt.Sprintf("%s%02d:%02d", sign, hours, minutes)
}
