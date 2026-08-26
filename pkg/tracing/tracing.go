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

// Package tracing provides OpenTelemetry-based distributed tracing support for multiple services.
// This package is designed to support multiple services within the same program
// while also registering a process-level OpenTelemetry fallback provider.
package tracing

import (
	"fmt"
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"go.opentelemetry.io/otel"
)

// nolint: gochecknoglobals
var globalHandler struct {
	sync.Once
	IHandler
}

// Init init the global tracing globalHandler.
func Init(conf Config) error {
	var err error
	globalHandler.Do(func() {
		handler, initErr := New(contextx.Background(), conf)
		if initErr != nil {
			err = fmt.Errorf("failed to init tracing globalHandler, err: %w", initErr)
			return
		}

		globalService, initErr := handler.initGlobalService()
		if initErr != nil {
			err = fmt.Errorf("failed to init tracing global service, err: %w", initErr)
			return
		}

		otel.SetTextMapPropagator(globalService.TracerPropagator())
		otel.SetTracerProvider(globalService.TracerProvider())
		globalHandler.IHandler = handler
	})

	return err
}

// G get the global tracing globalHandler.
// nolint: contextcheck
func G() IHandler {
	if globalHandler.IHandler == nil {
		_ = Init(DefaultConfig())
	}

	return globalHandler.IHandler
}
