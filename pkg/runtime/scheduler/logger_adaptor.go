/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package scheduler ...
package scheduler

import (
	"fmt"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
)

// LoggerAdapter adapts a logger.Logger to the cron.Logger interface.
type LoggerAdapter struct {
	Logger logger.Logger
}

// Info logs an informational message with additional context.
// cron v3's Info messsage is useless, so we can ignore it.
func (la LoggerAdapter) Info(_ string, _ ...interface{}) {}

// Error logs an error message with additional context.
func (la LoggerAdapter) Error(err error, msg string, keysAndValues ...interface{}) {
	var formatMsg string
	if len(keysAndValues) > 0 {
		formatMsg = fmt.Sprintf(fmt.Sprintf(formatString(len(keysAndValues)),
			append([]interface{}{msg}, formatTimes(keysAndValues)...)...), "err(%v)", err)
	} else {
		formatMsg = fmt.Sprintf("%s, err(%v)", msg, err)
	}
	la.Logger.Errorf("scheduler task run error in cron, %s", formatMsg)
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
