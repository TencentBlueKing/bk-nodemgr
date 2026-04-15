/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/adminclient"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	restdiscovery "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/discovery"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/ssl"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/backendadmin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

func main() {
	rootCMD := adminclient.NewRootCMD(newBackendAdminHandler)
	if err := rootCMD.Execute(); err != nil {
		fmt.Printf("failed to execute cmd: %v\n", err)
		os.Exit(1)
	}
}

func newBackendAdminHandler(configPath string) (backendadmin.IHandler, error) {
	conf := config.NewBackendService()
	if err := conf.LoadFromFile(configPath); err != nil {
		return nil, fmt.Errorf("failed to load config file(%s): %w", configPath, err)
	}
	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	httpClient, err := restclient.NewHTTPClient(&ssl.TLSConfig{InsecureSkipVerify: true})
	if err != nil {
		return nil, fmt.Errorf("failed to create http client: %w", err)
	}

	endpoint := "http://" + net.JoinHostPort(conf.AdminServer.AdvertiseIPV4, strconv.Itoa(conf.AdminServer.Port))
	h, err := backendadmin.New(&restclient.Capability{
		Name:                 "backend-adminclient",
		HTTPClient:           httpClient,
		Discover:             restdiscovery.NewDiscovery("backendadmin", []string{endpoint}),
		ToleranceLatencyTime: restclient.ToleranceLatencyTimeDefault,
		MetricOpts:           restclient.MetricOption{},
		TraceSvc:             noopTraceService{},
	}, &backendadmin.Config{
		RestJWTSecret:          conf.AdminServer.JWTServerConfig.SymmetricKey,
		RestJWTTokenExpiration: 24 * time.Hour,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create backendadmin handler: %w", err)
	}

	return h, nil
}

type noopTraceService struct{}

func (noopTraceService) TracerProvider() trace.TracerProvider { return noop.NewTracerProvider() }
func (noopTraceService) ServiceName() string                  { return "adminclient" }
func (noopTraceService) Shutdown(_ context.Context) error     { return nil }
func (noopTraceService) TracerPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
}
