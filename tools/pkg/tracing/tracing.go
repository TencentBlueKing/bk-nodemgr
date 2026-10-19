/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

// Package tracing provides the installer-wide trace context and HTTP propagation.
package tracing

import (
	"context"
	"net/http"
	"sync"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

const (
	// SampleRate keeps every installer trace sampled for log correlation.
	SampleRate = 1.0

	instrumentationName = "github.com/TencentBlueKing/bk-nodemgr/tools"
)

var (
	stateMu    sync.RWMutex
	tracerStop func(context.Context) error
	rootCtx    context.Context
	rootSpan   trace.Span
)

// Init initializes the installer-wide tracer and wraps the default HTTP transport.
// No exporter is configured, so spans are used only to generate and propagate IDs.
func Init() {
	stateMu.Lock()
	defer stateMu.Unlock()

	if rootCtx != nil {
		return
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(SampleRate)),
	)
	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	rootCtx, rootSpan = otel.Tracer(instrumentationName).Start(context.Background(), "installer")
	tracerStop = tracerProvider.Shutdown
	http.DefaultTransport = NewTransport(http.DefaultTransport)
}

// Shutdown ends the installer root span and shuts down the tracer provider.
func Shutdown(ctx context.Context) error {
	stateMu.Lock()
	defer stateMu.Unlock()

	if rootCtx == nil {
		return nil
	}

	rootSpan.End()
	rootCtx = nil
	rootSpan = nil

	if tracerStop == nil {
		return nil
	}

	return tracerStop(ctx)
}

// TraceID returns the trace ID shared by HTTP requests made during this CLI run.
func TraceID() string {
	stateMu.RLock()
	defer stateMu.RUnlock()

	if rootCtx == nil {
		return ""
	}

	spanContext := trace.SpanContextFromContext(rootCtx)
	if !spanContext.IsValid() {
		return ""
	}

	return spanContext.TraceID().String()
}

// NewTransport wraps an HTTP transport with client instrumentation and trace propagation.
func NewTransport(base http.RoundTripper) http.RoundTripper {
	return &transport{
		base: otelhttp.NewTransport(base),
	}
}

type transport struct {
	base http.RoundTripper
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	stateMu.RLock()
	ctx := rootCtx
	stateMu.RUnlock()

	if ctx != nil && !trace.SpanContextFromContext(req.Context()).IsValid() {
		req = req.WithContext(ctx)
	}

	return t.base.RoundTrip(req)
}
