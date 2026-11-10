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

package gse

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	restdiscovery "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/discovery"
	apigwclient "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/client"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

type processStreamHTTPClient func(*http.Request) (*http.Response, error)

func (fn processStreamHTTPClient) Do(req *http.Request) (*http.Response, error) { return fn(req) }

type processStreamTraceService struct{}

func (processStreamTraceService) TracerProvider() trace.TracerProvider {
	return noop.NewTracerProvider()
}
func (processStreamTraceService) ServiceName() string            { return "gse-stream-test" }
func (processStreamTraceService) Shutdown(context.Context) error { return nil }
func (processStreamTraceService) TracerPropagator() propagation.TextMapPropagator {
	return propagation.TraceContext{}
}

func newProcessStreamHandler(t *testing.T, fn processStreamHTTPClient) IHandlerProc {
	t.Helper()
	h, err := New(&restclient.Capability{
		Name: "gse-stream-test", HTTPClient: fn,
		Discover: restdiscovery.NewDiscovery("gse-stream-test", []string{"http://gse.test"}),
		TraceSvc: processStreamTraceService{},
	}, &Config{VirtualUserConfig: apigwclient.VirtualUserConfig{
		AppConfig: apigwclient.NewAppConfig([]string{"http://gse.test"}, "test", "test"),
		AuthMode:  apigwclient.AuthModeAt, AccessToken: "test",
	}})
	if err != nil {
		t.Fatal(err)
	}

	return h.NewHandlerProc()
}

func processStreamResponse(t *testing.T, data any) (*http.Response, error) {
	t.Helper()
	body, err := json.Marshal(map[string]any{"result": true, "code": 0, "data": data})
	if err != nil {
		t.Fatal(err)
	}

	return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header),
		Body: io.NopCloser(strings.NewReader(string(body)))}, nil
}

func TestQueryMultiProcessInfoManyStream(t *testing.T) {
	const goodKey = "good:bk-nodemgr:plugin"
	const otherKey = "other:bk-nodemgr:plugin"
	good := procOperateResult{Content: `{"process":[{"instance":[{"processName":"plugin","pid":123,"version":" v1 ","isAuto":true}]}]}`}
	tests := []struct {
		name       string
		poll       func(int) getProcOperateResultV2Resp
		submitFail bool
		cancel     bool
		wantErrors int
	}{
		{name: "completed results survive running peer timeout", wantErrors: 1,
			poll: func(int) getProcOperateResultV2Resp {
				return getProcOperateResultV2Resp{goodKey: good, otherKey: {ErrorCode: procOperateResultCodeRunning}}
			}},
		{name: "completed results are not repeated", poll: func(attempt int) getProcOperateResultV2Resp {
			if attempt == 1 {
				return getProcOperateResultV2Resp{goodKey: good, otherKey: {ErrorCode: procOperateResultCodeRunning}}
			}
			return getProcOperateResultV2Resp{goodKey: good, otherKey: good}
		}},
		{name: "empty response stays pending", wantErrors: 2, poll: func(int) getProcOperateResultV2Resp {
			return getProcOperateResultV2Resp{}
		}},
		{name: "missing target stays pending", wantErrors: 1, poll: func(int) getProcOperateResultV2Resp {
			return getProcOperateResultV2Resp{goodKey: good}
		}},
		{name: "terminal failure preserves successful peer", wantErrors: 1, poll: func(int) getProcOperateResultV2Resp {
			return getProcOperateResultV2Resp{goodKey: good, otherKey: {ErrorCode: 100, ErrorMsg: "query failed"}}
		}},
		{name: "malformed content only fails its target", wantErrors: 1, poll: func(int) getProcOperateResultV2Resp {
			return getProcOperateResultV2Resp{goodKey: good, otherKey: {Content: "invalid json"}}
		}},
		{name: "submission failure is observable", submitFail: true, wantErrors: 2},
		{name: "cancellation closes channel", cancel: true, wantErrors: 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				polls := 0
				h := newProcessStreamHandler(t, func(req *http.Request) (*http.Response, error) {
					if strings.HasSuffix(req.URL.Path, "operate_proc_multi") {
						if tt.submitFail {
							return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header),
								Body: io.NopCloser(strings.NewReader(`{"result":false,"code":1,"message":"submission failed"}`))}, nil
						}
						return processStreamResponse(t, operateProcMultiResp{TaskID: "task"})
					}
					polls++
					return processStreamResponse(t, tt.poll(polls))
				})
				nCtx, cancel := contextx.WithCancel(contextx.New(context.Background()))
				defer cancel()
				if tt.cancel {
					cancel()
				}
				seen := make(map[string]bool)
				failures := 0
				started := time.Now()
				for result := range h.QueryMultiProcessInfoManyStream(nCtx, &types.ProcessAgentGroup{
					PluginName: "plugin", ProcessName: "program", AgentIDList: []string{"good", "other", "good"},
				}) {
					if seen[result.AgentID] {
						t.Fatalf("duplicate result: %s", result.AgentID)
					}
					seen[result.AgentID] = true
					if result.PluginName != "plugin" {
						t.Fatalf("unexpected plugin: %s", result.PluginName)
					}
					if result.Err != nil {
						failures++
						continue
					}
					if result.AgentID == "good" && time.Since(started) >= time.Minute {
						t.Fatal("successful result was delayed until timeout")
					}
					if len(result.ProcessInfos) != 1 || result.ProcessInfos[0].Pid != 123 || result.ProcessInfos[0].Version != "v1" {
						t.Fatalf("unexpected process info: %+v", result)
					}
				}
				if len(seen) != 2 || failures != tt.wantErrors {
					t.Fatalf("got %d results, %d errors; want 2 results, %d errors", len(seen), failures, tt.wantErrors)
				}
			})
		})
	}
}

func TestQueryMultiProcessInfoManyStreamSubmitsAllBatches(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var submitted atomic.Int32
		h := newProcessStreamHandler(t, func(req *http.Request) (*http.Response, error) {
			if strings.HasSuffix(req.URL.Path, "operate_proc_multi") {
				return processStreamResponse(t, operateProcMultiResp{TaskID: fmt.Sprint(submitted.Add(1))})
			}
			return processStreamResponse(t, getProcOperateResultV2Resp{})
		})
		agents := make([]string, queryMultiProcessInfoPageSize+1)
		for i := range agents {
			agents[i] = fmt.Sprint(i)
		}
		results := h.QueryMultiProcessInfoManyStream(contextx.New(context.Background()), &types.ProcessAgentGroup{
			PluginName: "plugin", ProcessName: "program", AgentIDList: agents,
		})
		// The producer must finish even when the consumer does not drain until after the timeout.
		time.Sleep(2 * time.Minute)
		count := 0
		for result := range results {
			count++
			if result.Err != nil {
				continue
			}
			t.Fatal("empty response was treated as success")
		}
		if count != len(agents) || submitted.Load() != 2 {
			t.Fatalf("got %d results and %d submissions, want %d results and 2 submissions", count, submitted.Load(), len(agents))
		}
	})
}

func TestQueryMultiProcessInfoManyStreamSlowSubmission(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var submitted atomic.Int32
		h := newProcessStreamHandler(t, func(req *http.Request) (*http.Response, error) {
			if strings.HasSuffix(req.URL.Path, "operate_proc_multi") {
				if submitted.Add(1) > 1 {
					<-req.Context().Done()

					return nil, req.Context().Err()
				}

				return processStreamResponse(t, operateProcMultiResp{TaskID: "first"})
			}

			return processStreamResponse(t, getProcOperateResultV2Resp{
				"0:bk-nodemgr:plugin": {Content: `{"process":[{"instance":[{"pid":123}]}]}`},
			})
		})
		agents := make([]string, queryMultiProcessInfoPageSize+1)
		for i := range agents {
			agents[i] = fmt.Sprint(i)
		}
		started := time.Now()
		var successes int
		for result := range h.QueryMultiProcessInfoManyStream(contextx.New(context.Background()), &types.ProcessAgentGroup{
			PluginName: "plugin", ProcessName: "program", AgentIDList: agents,
		}) {
			if result.Err != nil {
				continue
			}
			if time.Since(started) >= time.Minute {
				t.Fatal("later submission delayed an already completed result until timeout")
			}
			successes++
		}
		if successes != 1 {
			t.Fatalf("got %d successful results, want 1", successes)
		}
	})
}
