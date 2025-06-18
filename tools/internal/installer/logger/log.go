/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package logger ...
package logger

import (
	"fmt"
	"log"
	"sync/atomic"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/constant"
)

// LogLevel defines the log severity levels.
type LogLevel int32

const (
	// LevelDebug captures detailed development information.
	LevelDebug LogLevel = iota
	// LevelInfo captures normal application behavior.
	LevelInfo
	// LevelWarn captures potentially harmful situations.
	LevelWarn
	// LevelError captures error events.
	LevelError
	// LevelOff turns off all logging.
	LevelOff
)

// Internal mapping of level names for output formatting.
func levelNames(level LogLevel) string {
	switch level {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Using atomic operations to ensure thread safety.
var currentLevel atomic.Int32 // nolint: gochecknoglobals

// Pre-compute aligned format string for better performance.
const logFormat = "| %-5s | %-15s | %s"

// init default log level.
func init() { // nolint: gochecknoinits
	currentLevel.Store(int32(LevelInfo))
}

// SetLevel changes the current logging level
// Parameter l must be a valid LogLevel constant.
func SetLevel(l LogLevel) error {
	if l < LevelDebug || l > LevelOff {
		return fmt.Errorf("invalid log level: %d, must be between %d and %d", l, LevelDebug, LevelOff)
	}
	currentLevel.Store(int32(l))

	return nil
}

// GetLevel returns the current logging level.
func GetLevel() LogLevel {
	return LogLevel(currentLevel.Load())
}

// Internal unified logging method to reduce code duplication.
func logInternal(level LogLevel, step constant.Step, msg string) {
	if LogLevel(currentLevel.Load()) <= level {
		// Ensure fixed width for each field to improve readability
		log.Print(fmt.Sprintf(logFormat, levelNames(level), step, msg))
	}
}

// Debug logs messages at debug level.
func Debug(step constant.Step, args ...interface{}) {
	logInternal(LevelDebug, step, fmt.Sprint(args...))
}

// Debugf logs formatted messages at debug level.
func Debugf(step constant.Step, format string, args ...interface{}) {
	logInternal(LevelDebug, step, fmt.Sprintf(format, args...))
}

// Info logs messages at info level.
func Info(step constant.Step, args ...interface{}) {
	logInternal(LevelInfo, step, fmt.Sprint(args...))
}

// Infof logs formatted messages at info level.
func Infof(step constant.Step, format string, args ...interface{}) {
	logInternal(LevelInfo, step, fmt.Sprintf(format, args...))
}

// Warn logs messages at warning level.
func Warn(step constant.Step, args ...interface{}) {
	logInternal(LevelWarn, step, fmt.Sprint(args...))
}

// Warnf logs formatted messages at warning level.
func Warnf(step constant.Step, format string, args ...interface{}) {
	logInternal(LevelWarn, step, fmt.Sprintf(format, args...))
}

// Error logs messages at error level.
func Error(step constant.Step, args ...interface{}) {
	logInternal(LevelError, step, fmt.Sprint(args...))
}

// Errorf logs formatted messages at error level.
func Errorf(step constant.Step, format string, args ...interface{}) {
	logInternal(LevelError, step, fmt.Sprintf(format, args...))
}
