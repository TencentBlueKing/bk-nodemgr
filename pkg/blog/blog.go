/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package blog provides a logging wrapper for glog.
package blog

import (
	"log"
	"sync"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/blog/glog"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
)

// LogConfig is the configuration for initializing logs.
type LogConfig glog.LogConfig

// NewLogConfig creates a new LogConfig instance.
func NewLogConfig() LogConfig {
	return LogConfig{
		LogDir:       "./logs",
		LogMaxSizeMB: 200,
		LogMaxNum:    10,

		ToStdErr:        true,
		AlsoToStdErr:    true,
		Level:           "info",
		StdErrThreshold: "4",
		VModule:         "",
		TraceLocation:   "",
	}
}

const writerLoggerDepth = 2

// WriterInfo serves as a bridge between the standard log package and the glog package.
type WriterInfo struct{}

// Write implements the io.Writer interface.
func (writer WriterInfo) Write(data []byte) (n int, err error) {
	glog.InfoDepth(writerLoggerDepth, string(data))
	return len(data), nil
}

// WriterDebug serves as a bridge between the standard log package and the glog package.
type WriterDebug struct{}

// Write implements the io.Writer interface.
func (writer WriterDebug) Write(data []byte) (n int, err error) {
	glog.DebugDepth(writerLoggerDepth, string(data))
	return len(data), nil
}

// WriterError serves as a bridge between the standard log package and the glog package.
type WriterError struct{}

// Write implements the io.Writer interface.
func (writer WriterError) Write(data []byte) (n int, err error) {
	glog.ErrorDepth(writerLoggerDepth, string(data))
	return len(data), nil
}

var once sync.Once

// delayDefault is the default delay time to flush logs.
const delayDefault = 5 * time.Second

// InitLogs initializes logs the way we want for blog.
func InitLogs(logConfig LogConfig) {
	glog.InitLogs(glog.LogConfig(logConfig))

	once.Do(func() {
		log.SetOutput(WriterInfo{})
		log.SetFlags(0)
		go func() {
			tick := time.Tick(delayDefault)

			for range tick {
				glog.Flush()
			}
		}()
	})
}

// CloseLogs flush the rest logs in buffer.
func CloseLogs() {
	glog.Flush()
}

var (
	// Debug prints logs like fmt.Print in debug level.
	Debug = glog.Debug
	// Debugf prints logs like fmt.Printf in debug level.
	Debugf = glog.Debugf
	// Debugw prints logs with key(values) in debug level.
	Debugw = glog.Debugw

	// Info prints logs like fmt.Print in info level.
	Info = glog.Info
	// Infof prints logs like fmt.Printf in info level.
	Infof = glog.Infof
	// Infow prints logs with key(values) in info level.
	Infow = glog.Infow

	// Warn prints logs like fmt.Print in warn level.
	Warn = glog.Warning
	// Warnf prints logs like fmt.Printf in warn level.
	Warnf = glog.Warningf
	// Warnw prints logs with key(values) in warn level.
	Warnw = glog.Warningw

	// Error prints	logs like fmt.Print in error level.
	Error = glog.Error
	// Errorf prints logs like fmt.Printf in error level.
	Errorf = glog.Errorf
	// Errorw prints logs with key(values) in error level.
	Errorw = glog.Errorw
)

// SetLevel set the logging level.
func SetLevel(level string) {
	glog.SetLevel(level)
}

var _ logger.ILogger = &GlobalLogger{}

// GlobalLogger serves as a bridge between the standard log package and the glog package.
type GlobalLogger struct{}

const globalLoggerDepth = 1

// Debug ...
func (l GlobalLogger) Debug(args ...interface{}) {
	glog.DebugDepth(globalLoggerDepth, args...)
}

// Debugf ...
func (l GlobalLogger) Debugf(format string, args ...interface{}) {
	glog.DebugDepthf(globalLoggerDepth, format, args...)
}

// Debugw ...
func (l GlobalLogger) Debugw(args ...interface{}) {
	glog.DebugDepthw(globalLoggerDepth, args...)
}

// DebugCtxf prints logs with context in debug level.
func (l GlobalLogger) DebugCtxf(ctx contextx.IContext, format string, args ...interface{}) {
	glog.DebugDepthf(globalLoggerDepth, logger.FormatWithCtx(ctx, format), args...)
}

// Info ...
func (l GlobalLogger) Info(args ...interface{}) {
	glog.InfoDepth(globalLoggerDepth, args...)
}

// Infof ...
func (l GlobalLogger) Infof(format string, args ...interface{}) {
	glog.InfoDepthf(globalLoggerDepth, format, args...)
}

// Infow ...
func (l GlobalLogger) Infow(args ...interface{}) {
	glog.InfoDepthw(globalLoggerDepth, args...)
}

// InfoCtxf prints logs with context in info level.
func (l GlobalLogger) InfoCtxf(ctx contextx.IContext, format string, args ...interface{}) {
	glog.InfoDepthf(globalLoggerDepth, logger.FormatWithCtx(ctx, format), args...)
}

// Warn ...
func (l GlobalLogger) Warn(args ...interface{}) {
	glog.WarningDepth(globalLoggerDepth, args...)
}

// Warnf ...
func (l GlobalLogger) Warnf(format string, args ...interface{}) {
	glog.WarningDepthf(globalLoggerDepth, format, args...)
}

// Warnw ...
func (l GlobalLogger) Warnw(args ...interface{}) {
	glog.WarningDepthw(globalLoggerDepth, args...)
}

// WarnCtxf prints logs with context in warn level.
func (l GlobalLogger) WarnCtxf(ctx contextx.IContext, format string, args ...interface{}) {
	glog.WarningDepthf(globalLoggerDepth, logger.FormatWithCtx(ctx, format), args...)
}

// Error ...
func (l GlobalLogger) Error(args ...interface{}) {
	glog.ErrorDepth(globalLoggerDepth, args...)
}

// Errorf ...
func (l GlobalLogger) Errorf(format string, args ...interface{}) {
	glog.ErrorDepthf(globalLoggerDepth, format, args...)
}

// Errorw ...
func (l GlobalLogger) Errorw(args ...interface{}) {
	glog.ErrorDepthw(globalLoggerDepth, args...)
}

// ErrorCtxf prints logs with context in error level.
func (l GlobalLogger) ErrorCtxf(ctx contextx.IContext, format string, args ...interface{}) {
	glog.ErrorDepthf(globalLoggerDepth, logger.FormatWithCtx(ctx, format), args...)
}
