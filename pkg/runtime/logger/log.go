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
	"runtime"
	"strings"
)

// Logger is the logger interface.
type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Debugw(args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Infow(args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Warnw(args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Errorw(args ...interface{})
}

// ANSI color codes.
const (
	reset   = "\033[0m"
	red     = "\033[31m"
	green   = "\033[32m"
	yellow  = "\033[33m"
	blue    = "\033[34m"
	magenta = "\033[35m"
)

// LoggerDefault is the default logger.
type LoggerDefault struct{}

// skip is the number of stack frames to skip when computing the caller's filename and line number.
const skip = 2

// getCallerInfo returns the filename and line number of the caller.
func getCallerInfo() string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return ""
	}
	// Trim the full path to just the filename
	shortFile := file
	if lastSepIndex := strings.LastIndexByte(file, '/'); lastSepIndex >= 0 {
		shortFile = file[lastSepIndex+1:]
	}

	return fmt.Sprintf("%s:%d", shortFile, line)
}

func debugPrefix(caller string) string {
	return fmt.Sprintf("%s[DEBUG]%s %s ", blue, reset, caller)
}

// Debug ...
func (logger LoggerDefault) Debug(args ...interface{}) {
	caller := getCallerInfo()
	log.Print(debugPrefix(caller), fmt.Sprint(args...))
}

// Debugf ...
func (logger LoggerDefault) Debugf(format string, args ...interface{}) {
	caller := getCallerInfo()
	log.Printf(debugPrefix(caller)+format, args...)
}

// Debugw ...
func (logger LoggerDefault) Debugw(args ...interface{}) {
	caller := getCallerInfo()
	log.Print(debugPrefix(caller), fmt.Sprint(args...))
}

func infoPrefix(caller string) string {
	return fmt.Sprintf("%s[DeploymentInfo]%s %s ", green, magenta, caller)
}

// Info ...
func (logger LoggerDefault) Info(args ...interface{}) {
	caller := getCallerInfo()
	log.Print(infoPrefix(caller), fmt.Sprint(args...))
}

// Infof ...
func (logger LoggerDefault) Infof(format string, args ...interface{}) {
	caller := getCallerInfo()
	log.Printf(infoPrefix(caller)+format, args...)
}

// Infow ...
func (logger LoggerDefault) Infow(args ...interface{}) {
	caller := getCallerInfo()
	log.Print(infoPrefix(caller), fmt.Sprint(args...))
}

func warnPrefix(caller string) string {
	return fmt.Sprintf("%s[WARN]%s %s ", yellow, reset, caller)
}

// Warn ...
func (logger LoggerDefault) Warn(args ...interface{}) {
	caller := getCallerInfo()
	log.Print(warnPrefix(caller), fmt.Sprint(args...))
}

// Warnf ...
func (logger LoggerDefault) Warnf(format string, args ...interface{}) {
	caller := getCallerInfo()
	log.Printf(warnPrefix(caller)+format, args...)
}

// Warnw ...
func (logger LoggerDefault) Warnw(args ...interface{}) {
	caller := getCallerInfo()
	log.Print(warnPrefix(caller), fmt.Sprint(args...))
}

func errorPrefix(caller string) string {
	return fmt.Sprintf("%s[ERROR]%s %s ", red, reset, caller)
}

// Error ...
func (logger LoggerDefault) Error(args ...interface{}) {
	caller := getCallerInfo()
	log.Print(errorPrefix(caller), fmt.Sprint(args...))
}

// Errorf ...
func (logger LoggerDefault) Errorf(format string, args ...interface{}) {
	caller := getCallerInfo()
	log.Printf(errorPrefix(caller)+format, args...)
}

// Errorw ...
func (logger LoggerDefault) Errorw(args ...interface{}) {
	caller := getCallerInfo()
	log.Print(errorPrefix(caller), fmt.Sprint(args...))
}
