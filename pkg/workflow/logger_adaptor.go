/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package workflow ...
package workflow

import (
	"github.com/RichardKnop/logging"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
)

type loggerAdaptor struct {
}

func newLoggerAdaptor() logging.LoggerInterface {
	return &loggerAdaptor{}
}

// Print this is implement of logging.LoggerInterface.
func (l *loggerAdaptor) Print(args ...interface{}) {
	logger.G.Sys().Debug("%v", args)
}

// Printf this is implement of logging.LoggerInterface.
func (l *loggerAdaptor) Printf(s string, args ...interface{}) {
	logger.G.Sys().Debug(s, args...)
}

// Println this is implement of logging.LoggerInterface.
func (l *loggerAdaptor) Println(args ...interface{}) {
	logger.G.Sys().Debug("%v", args)
}

// Fatal this is implement of logging.LoggerInterface.
func (l *loggerAdaptor) Fatal(args ...interface{}) {
	logger.G.Sys().Error("%v", args)
}

// Fatalf this is implement of logging.LoggerInterface.
func (l *loggerAdaptor) Fatalf(s string, args ...interface{}) {
	logger.G.Sys().Error(s, args...)
}

// Fatalln this is implement of logging.LoggerInterface.
func (l *loggerAdaptor) Fatalln(args ...interface{}) {
	logger.G.Sys().Error("%v", args)
}

// Panic this is implement of logging.LoggerInterface.
func (l *loggerAdaptor) Panic(args ...interface{}) {
	logger.G.Sys().Error("%v", args)
}

// Panicf this is implement of logging.LoggerInterface.
func (l *loggerAdaptor) Panicf(s string, args ...interface{}) {
	logger.G.Sys().Error(s, args...)
}

// Panicln this is implement of logging.LoggerInterface.
func (l *loggerAdaptor) Panicln(args ...interface{}) {
	logger.G.Sys().Error("%v", args)
}
