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

// Package common defines workflow common things.
package common

import (
	"fmt"
	"time"
)

// MessageHandler defines the interface for handling messages in action instances.
type MessageHandler interface {
	// SetMessage sets a message to the instance.
	SetMessage(zhMsg, enMsg, level string)
}

// Message describes the single message in action instance.
type Message struct {
	Time   time.Time
	TextZh string // Chinese content
	TextEn string // English content
	Level  string
}

// LogBuilder is a chainable log builder for bilingual logging.
type LogBuilder struct {
	msgHandler MessageHandler
	zhMsg      string
	enMsg      string
	zhSet      bool
	enSet      bool
}

// NewLogBuilder creates a new LogBuilder with the given MessageHandler.
func NewLogBuilder(handler MessageHandler) *LogBuilder {
	return &LogBuilder{
		msgHandler: handler,
	}
}

// Zh sets the Chinese log message.
func (builder *LogBuilder) Zh(format string, args ...any) *LogBuilder {
	builder.zhMsg = fmt.Sprintf(format, args...)
	builder.zhSet = true

	return builder
}

// En sets the English log message.
func (builder *LogBuilder) En(format string, args ...any) *LogBuilder {
	builder.enMsg = fmt.Sprintf(format, args...)
	builder.enSet = true

	return builder
}

// log is the internal method that handles auto-fill logic and appends the message.
func (builder *LogBuilder) log(level string) {
	zhMsg, enMsg := builder.zhMsg, builder.enMsg

	// Auto-fill rules based on flags
	switch {
	case builder.zhSet && !builder.enSet:
		enMsg = zhMsg
	case !builder.zhSet && builder.enSet:
		zhMsg = enMsg
	case !builder.zhSet && !builder.enSet:
		zhMsg, enMsg = level, level
	}

	builder.msgHandler.SetMessage(zhMsg, enMsg, level)
}

// Info logs at INFO level.
func (builder *LogBuilder) Info() { builder.log("INFO") }

// Warn logs at WARN level.
func (builder *LogBuilder) Warn() { builder.log("WARN") }

// Error logs at ERROR level.
func (builder *LogBuilder) Error() { builder.log("ERROR") }
