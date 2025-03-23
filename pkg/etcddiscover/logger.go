/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package etcddiscover

import (
	"log"
)

// Logger logger in discover.
type Logger interface {
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

// defaultLogger default logger logs to stdout.
type defaultLogger struct{}

// Infof logs info messages.
func (d defaultLogger) Infof(format string, args ...interface{}) {
	log.Printf(format, args...)
}

// Errorf logs error messages.
func (d defaultLogger) Errorf(format string, args ...interface{}) {
	log.Printf(format, args...)
}
