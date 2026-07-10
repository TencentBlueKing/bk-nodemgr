/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package backendadmin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/globalsettings"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	restdiscovery "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/discovery"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/ssl"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

type testTraceService struct{}

func (testTraceService) TracerProvider() trace.TracerProvider {
	return noop.NewTracerProvider()
}

func (testTraceService) ServiceName() string {
	return "backendadmin_test"
}

func (testTraceService) Shutdown(_ context.Context) error {
	return nil
}

func (testTraceService) TracerPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
}

func TestHandlerGetNetworkUnitSegmentRulesReturnsTypedConfig(t *testing.T) {
	h := &Handler{cli: &cli{}}
	h.cli.getNetworkUnitSegmentRulesFn = func(contextx.IContext) (types.NetworkUnitSegmentRuleConfig, error) {
		return types.NetworkUnitSegmentRuleConfig{
			"0": {
				Rules: []types.NetworkUnitSegmentRule{{
					CIDRs:         []string{"9.135.144.0/24"},
					NetworkUnitID: 1,
				}},
			},
		}, nil
	}

	cfg, err := h.GetNetworkUnitSegmentRules(contextx.New(context.Background(), contextx.WithTenantID("t")))

	require.NoError(t, err)
	require.Contains(t, cfg, "0")
	require.Len(t, cfg["0"].Rules, 1)
	assert.Equal(t, []string{"9.135.144.0/24"}, cfg["0"].Rules[0].CIDRs)
	assert.Equal(t, int64(1), cfg["0"].Rules[0].NetworkUnitID)
}

func TestHandlerUpsertNetworkUnitSegmentRulesUsesFixedSettingName(t *testing.T) {
	h := &Handler{cli: &cli{}}
	var gotCfg types.NetworkUnitSegmentRuleConfig
	h.cli.upsertNetworkUnitSegmentRulesFn = func(_ contextx.IContext, cfg types.NetworkUnitSegmentRuleConfig) error {
		gotCfg = cfg
		return nil
	}

	err := h.UpsertNetworkUnitSegmentRules(contextx.New(context.Background(), contextx.WithTenantID("t")), types.NetworkUnitSegmentRuleConfig{
		"0": {
			Rules: []types.NetworkUnitSegmentRule{{
				CIDRs:         []string{"0.0.0.0/0"},
				NetworkUnitID: 0,
			}},
		},
	})

	require.NoError(t, err)
	require.Contains(t, gotCfg, "0")
	assert.Equal(t, int64(0), gotCfg["0"].Rules[0].NetworkUnitID)
}

func TestHandlerGetNetworkUnitSegmentRulesViaHTTP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		require.Equal(t, http.MethodPost, req.Method)
		require.Equal(t, "/admin/globalsettings/get", req.URL.Path)
		require.NotEmpty(t, req.Header.Get("X-Bk-Tenant-Id"))
		require.NotEmpty(t, req.Header.Get("X-Bknodemgr-Authorization"))

		var body map[string]string
		require.NoError(t, json.NewDecoder(req.Body).Decode(&body))
		assert.Equal(t, globalsettings.NetworkUnitSegmentRules, body["setting_name"])

		rw.Header().Set("Content-Type", "application/json")
		_, _ = rw.Write([]byte(`{"code":0,"message":"OK","request_id":"rid","data":{"value":"{\"0\":{\"rules\":[{\"cidrs\":[\"9.135.144.0/24\"],\"bk_networkunit_id\":1}]}}"}}`))
	}))
	defer server.Close()

	h := newHTTPTestHandler(t, server.URL)

	cfg, err := h.GetNetworkUnitSegmentRules(contextx.New(context.Background(), contextx.WithTenantID("t")))

	require.NoError(t, err)
	require.Contains(t, cfg, "0")
	require.Len(t, cfg["0"].Rules, 1)
	assert.Equal(t, []string{"9.135.144.0/24"}, cfg["0"].Rules[0].CIDRs)
	assert.Equal(t, int64(1), cfg["0"].Rules[0].NetworkUnitID)
}

func TestHandlerUpsertNetworkUnitSegmentRulesViaHTTP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		require.Equal(t, http.MethodPost, req.Method)
		require.Equal(t, "/admin/globalsettings/upsertmany", req.URL.Path)

		var body struct {
			Settings []struct {
				SettingName string `json:"setting_name"`
				Value       string `json:"value"`
			} `json:"settings"`
		}
		require.NoError(t, json.NewDecoder(req.Body).Decode(&body))
		require.Len(t, body.Settings, 1)
		assert.Equal(t, globalsettings.NetworkUnitSegmentRules, body.Settings[0].SettingName)
		assert.Contains(t, body.Settings[0].Value, `"bk_networkunit_id":0`)

		rw.Header().Set("Content-Type", "application/json")
		_, _ = rw.Write([]byte(`{"code":0,"message":"OK","request_id":"rid","data":{}}`))
	}))
	defer server.Close()

	h := newHTTPTestHandler(t, server.URL)

	err := h.UpsertNetworkUnitSegmentRules(contextx.New(context.Background(), contextx.WithTenantID("t")), types.NetworkUnitSegmentRuleConfig{
		"0": {
			Rules: []types.NetworkUnitSegmentRule{{
				CIDRs:         []string{"0.0.0.0/0"},
				NetworkUnitID: 0,
			}},
		},
	})

	require.NoError(t, err)
}

func newHTTPTestHandler(t *testing.T, endpoint string) *Handler {
	t.Helper()

	httpClient, err := client.NewHTTPClient(&ssl.TLSConfig{InsecureSkipVerify: true})
	require.NoError(t, err)

	h, err := New(&client.Capability{
		Name:                 "backendadmin-test",
		HTTPClient:           httpClient,
		Discover:             restdiscovery.NewDiscovery("backendadmin-test", []string{endpoint}),
		ToleranceLatencyTime: client.ToleranceLatencyTimeDefault,
		MetricOpts:           client.MetricOption{},
		TraceSvc:             testTraceService{},
	}, &Config{
		RestJWTSecret:          "test-secret",
		RestJWTTokenExpiration: time.Hour,
	})
	require.NoError(t, err)

	return h
}
