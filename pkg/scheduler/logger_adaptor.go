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

// Package scheduler ...
package scheduler

import (
	"fmt"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
)

// LoggerAdapter adapts a logger.ILogger to the cron.Logger interface.
type LoggerAdapter struct {
}

// Info logs an informational message with additional context.
func (la LoggerAdapter) Info(msg string, keysAndValues ...interface{}) {
	var formatMsg string
	if len(keysAndValues) > 0 {
		formatMsg = fmt.Sprintf(formatString(len(keysAndValues)),
			append([]interface{}{msg}, formatTimes(keysAndValues)...)...)
	} else {
		formatMsg = msg
	}

	logger.G.Sys().Debug("scheduler task running: %s", formatMsg)
}

// Error logs an error message with additional context.
func (la LoggerAdapter) Error(err error, msg string, keysAndValues ...interface{}) {
	var formatMsg string
	if len(keysAndValues) > 0 {
		formatMsg = fmt.Sprintf(fmt.Sprintf(formatString(len(keysAndValues)),
			append([]interface{}{msg}, formatTimes(keysAndValues)...)...), "err(%v)", err)
	} else {
		formatMsg = fmt.Sprintf("%s, err(%v)", msg, err)
	}

	logger.G.Sys().Error("scheduler task run error in cron: %s", formatMsg)
}

// formatString returns a logfmt-like format string for the number of
// key/values.
func formatString(numKeysAndValues int) string {
	var sb strings.Builder
	sb.WriteString("%s")
	if numKeysAndValues > 0 {
		sb.WriteString(", ")
	}
	for i := 0; i < numKeysAndValues/2; i++ {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString("%v(%v)")
	}

	return sb.String()
}

// formatTimes formats any time.Time values as RFC3339.
func formatTimes(keysAndValues []any) []any {
	formattedArgs := make([]any, 0, len(keysAndValues))
	for _, arg := range keysAndValues {
		if t, ok := arg.(time.Time); ok {
			arg = t.Format(time.RFC3339)
		}
		formattedArgs = append(formattedArgs, arg)
	}

	return formattedArgs
}
