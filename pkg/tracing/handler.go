/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package tracing

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

var _ IHandler = &Handler{}

// Handler manages multiple service svcs.
type Handler struct {
	mu       sync.RWMutex
	svcs     map[string]IService
	conf     Config
	exporter sdkTrace.SpanExporter
}

// New creates a new tracer manager.
func New(nCtx contextx.IContext, conf Config) (*Handler, error) {
	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	exp, err := newExporter(nCtx, conf.Exporter)
	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %w", err)
	}

	h := &Handler{
		conf:     conf,
		svcs:     make(map[string]IService),
		exporter: exp,
	}

	return h, nil
}

// NewService creates a new service tracer with the given configuration.
func (h *Handler) NewService(config ServiceConfig) (IService, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	// Check if tracer already exists for this service
	if existingTracer, exists := h.svcs[config.ServiceName]; exists {
		return existingTracer, nil
	}

	// Validate configuration
	if config.ServiceName == "" {
		return nil, fmt.Errorf("service name is required")
	}

	// Create resource with service information
	res, err := resource.New(contextx.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(config.ServiceName),
			semconv.ServiceNamespaceKey.String(h.conf.Namespace),
			semconv.ServiceInstanceIDKey.String(h.conf.InstanceID),
			semconv.ServiceVersionKey.String(h.conf.Version),
			semconv.DeploymentEnvironmentKey.String(h.conf.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	res, err = resource.Merge(resource.Default(), res)
	if err != nil {
		return nil, fmt.Errorf("failed to merge resource: %w", err)
	}

	// Create tracer tracerProvider with sampling
	tracerProvider := sdkTrace.NewTracerProvider(
		sdkTrace.WithBatcher(h.exporter),
		sdkTrace.WithResource(res),
		sdkTrace.WithSampler(sdkTrace.TraceIDRatioBased(config.SampleRate)),
	)

	tracerPropagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)

	// Create service tracer instance
	service := &Service{
		serviceName:      config.ServiceName,
		tracerProvider:   tracerProvider,
		tracerPropagator: tracerPropagator,
		shutdown:         tracerProvider.Shutdown,
	}

	// Store in manager
	h.svcs[config.ServiceName] = service

	return service, nil
}

// GetService retrieves an existing service tracer.
func (h *Handler) GetService(serviceName string) (IService, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	tracer, exists := h.svcs[serviceName]
	if !exists {
		return nil, fmt.Errorf("no tracer found for service: %s", serviceName)
	}

	return tracer, nil
}

// ShutdownAll shuts down all registered service svcs.
func (h *Handler) ShutdownAll(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	var errs error
	for serviceName, serviceTracer := range h.svcs {
		if err := serviceTracer.Shutdown(ctx); err != nil {
			errs = errors.Join(errs, fmt.Errorf("failed to shutdown tracer for service, service(%s): %w", serviceName, err))
		}
	}

	// Clear all svcs
	h.svcs = make(map[string]IService)

	return errs
}

// ListServices returns a list of all registered service names.
func (h *Handler) ListServices() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	services := make([]string, 0, len(h.svcs))
	for serviceName := range h.svcs {
		services = append(services, serviceName)
	}

	return services
}

// RemoveService removes a service tracer from the manager and shuts it down.
func (h *Handler) RemoveService(ctx context.Context, serviceName string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	serviceTracer, exists := h.svcs[serviceName]
	if !exists {
		return fmt.Errorf("no tracer found for service: %s", serviceName)
	}

	if err := serviceTracer.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown tracer for service %s: %w", serviceName, err)
	}

	delete(h.svcs, serviceName)

	return nil
}

// newExporter creates the appropriate exporter based on configuration.
func newExporter(nCtx contextx.IContext, config ExporterConfig) (sdkTrace.SpanExporter, error) {
	switch config.ExporterType {
	case ExporterTypeStdout:
		return stdouttrace.New(stdouttrace.WithPrettyPrint())

	case ExporterTypeOTLP:
		if config.OTLPConfig == nil {
			return nil, fmt.Errorf("OTLP conf is required for OTLP exporter")
		}

		var opts []otlptracegrpc.Option

		if config.OTLPConfig.Insecure {
			opts = append(opts, otlptracegrpc.WithInsecure())
		}

		if config.OTLPConfig.Endpoint != "" {
			opts = append(opts, otlptracegrpc.WithEndpoint(config.OTLPConfig.Endpoint))
		}

		if len(config.OTLPConfig.Headers) > 0 {
			opts = append(opts, otlptracegrpc.WithHeaders(config.OTLPConfig.Headers))
		}

		exp, err := otlptrace.New(nCtx, otlptracegrpc.NewClient(opts...))
		if err != nil {
			return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
		}

		return exp, nil
	default:
		return nil, fmt.Errorf("unsupported exporter type: %s", config.ExporterType)
	}
}
