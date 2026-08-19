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

// Package systeminfo logs best-effort target machine diagnostics for installer startup.
package systeminfo

import "github.com/TencentBlueKing/bk-nodemgr/tools/pkg/logger"

const (
	unknownValue    = "unknown"
	startMessage    = "installer initial target machine information:"
	fieldLogFormat  = "target machine info: %s=%s"
	fieldWarnFormat = "failed to collect target machine info field(%s): %v"
)

type targetInfoField struct {
	key     string
	collect func() (string, error)
}

var targetInfoFields = []targetInfoField{
	{key: "goos", collect: collectGOOS},
	{key: "goarch", collect: collectGOARCH},
	{key: "hostname", collect: collectHostname},
	{key: "current_time", collect: collectCurrentTime},
	{key: "timezone", collect: collectTimezone},
	{key: "timezone_offset", collect: collectTimezoneOffset},
	{key: "system_version", collect: collectSystemVersion},
	{key: "kernel_version", collect: collectKernelVersion},
}

// LogInitialTargetInfo logs best-effort target machine information for installer startup diagnostics.
func LogInitialTargetInfo(step logger.Step) {
	logger.Infof(step, startMessage)

	for _, field := range targetInfoFields {
		value, err := field.collect()
		if err != nil {
			logger.Warnf(step, fieldWarnFormat, field.key, err)
			value = unknownValue
		}
		if value == "" {
			value = unknownValue
		}

		logger.Debugf(step, fieldLogFormat, field.key, value)
	}
}
