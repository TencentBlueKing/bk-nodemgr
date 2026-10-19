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

package tracing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTransportInjectsInstallerTraceID(t *testing.T) {
	Init()
	t.Cleanup(func() {
		if err := Shutdown(context.Background()); err != nil {
			t.Fatalf("shutdown tracer provider: %v", err)
		}
	})

	traceID := TraceID()
	if traceID == "" {
		t.Fatal("expected installer trace ID")
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		traceparent := req.Header.Get("Traceparent")
		parts := strings.Split(traceparent, "-")
		if len(parts) != 4 || parts[1] != traceID || parts[3] != "01" {
			t.Errorf("unexpected traceparent %q, trace ID %q", traceparent, traceID)
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	client := &http.Client{Transport: NewTransport(http.DefaultTransport)}
	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("send request: %v", err)
	}
	_ = resp.Body.Close()
}
