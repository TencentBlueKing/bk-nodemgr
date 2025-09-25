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

// Package main describes the logger example usages.
// nolint: mnd
package main

import (
	"context"
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
)

func main() {
	logger.Init(logger.Config{
		LogDir:       "./logs",
		LogMaxSizeMB: 1,
		LogMaxNum:    3,

		ToStdErr:     false,
		AlsoToStdErr: true,
		Level:        logger.LevelDebug,
	})
	defer logger.G.Flush()

	// generates a request context.
	ctx := contextx.New(context.Background(), contextx.WithValues(map[string]any{"a": "ok", "b": 3, "c": false}))

	// business logs.
	logger.G.Biz().Ctx(ctx).Debug("debug message")
	logger.G.Biz().Ctx(ctx).Info("received a request %s", "install/agent")
	logger.G.Biz().Ctx(ctx).WithErr(errors.New("invalid params")).Error("failed to decode params")
	logger.G.Biz().Ctx(ctx).
		With("host-id", 123).With("state", "unknown").
		Warn("got unexpected data")
	logger.G.Biz().Ctx(ctx).With("host-id", 123, "state", "unknown").
		Warn("got unexpected data")
	logger.G.Biz().Ctx(ctx).With("wrong-key").Info("try to print a wrong key-value pair")

	// system logs.
	logger.G.Sys().Debug("debug message")
	logger.G.Sys().Info("brings up a backend thread %d", 123)
	logger.G.Sys().WithErr(errors.New("internal error")).Error("failed to launch task")
	logger.G.Sys().Ctx(ctx).With("host-id", 123).With("state", "unknown").
		Warn("got unexpected data")
	logger.G.Sys().Ctx(ctx).With("host-id", 123, "state", "unknown").
		Warn("got unexpected data")
	logger.G.Sys().Ctx(ctx).With("wrong-key").Info("try to print a wrong key-value pair")
}
