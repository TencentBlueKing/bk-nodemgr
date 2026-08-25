//go:build integration

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

package cmdb

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	restdiscovery "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/discovery"
	apigwclient "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/client"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tracing"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/testsuite/support"
)

const cmdbIntegrationReadyTimeout = 2 * time.Minute

func TestIntegrationCMDBSearchBusiness(t *testing.T) {
	target := newIntegrationTarget(t)
	handler := newIntegrationHandler(t, target)
	nCtx := contextx.New(
		context.Background(),
		contextx.WithTenantID(target.TenantID),
		contextx.WithBKUsername(target.VirtualUser),
	)

	businesses, err := handler.SearchBusiness(nCtx, types.Page{Offset: 0, Limit: 1})
	if err != nil {
		t.Fatalf("search business: %v", err)
	}

	if len(businesses) == 0 {
		return
	}
	if businesses[0] == nil {
		t.Fatal("first business is nil")
	}
	if businesses[0].TenantID != target.TenantID {
		t.Fatalf("business tenant id = %q, want %q", businesses[0].TenantID, target.TenantID)
	}
}

func newIntegrationTarget(t *testing.T) support.CMDBTarget {
	t.Helper()

	return support.RequireCMDBTarget(t)
}

func waitForCMDBReady(
	nCtx contextx.IContext, searchFn func(contextx.IContext) error,
) error {
	readyCtx, cancel := contextx.WithTimeout(nCtx, cmdbIntegrationReadyTimeout)
	defer cancel()

	for {
		err := searchFn(readyCtx)
		if isCMDBDiscoveryNotReady(err) {
			timer := time.NewTimer(500 * time.Millisecond)
			select {
			case <-readyCtx.Done():
				timer.Stop()

				return readyCtx.Err()
			case <-timer.C:
			}

			continue
		}
		if err != nil {
			return err
		}

		return nil
	}
}

func newIntegrationContext(target support.CMDBTarget) contextx.IContext {
	return contextx.New(
		context.Background(),
		contextx.WithTenantID(target.TenantID),
		contextx.WithBKUsername(target.VirtualUser),
	)
}

func isCMDBDiscoveryNotReady(err error) bool {
	readinessErr := new(cmdbReadinessError)
	return errors.As(err, &readinessErr) && (readinessErr.code == 1199037 || readinessErr.code == 1199002)
}

func newIntegrationHandler(t *testing.T, target support.CMDBTarget) IHandler {
	t.Helper()

	clientCap := newIntegrationCapability(t, target)
	waitForIntegrationCMDB(t, target, clientCap.HTTPClient)

	handler, err := New(clientCap, newIntegrationConfig(target))
	if err != nil {
		t.Fatalf("create cmdb handler: %v", err)
	}

	return handler
}

func newIntegrationClient(t *testing.T, target support.CMDBTarget) *cli {
	t.Helper()

	clientCap := newIntegrationCapability(t, target)
	waitForIntegrationCMDB(t, target, clientCap.HTTPClient)

	client, err := newClient(clientCap, newIntegrationConfig(target))
	if err != nil {
		t.Fatalf("create cmdb client: %v", err)
	}

	return client
}

func waitForIntegrationCMDB(t *testing.T, target support.CMDBTarget, client restclient.HTTPClient) {
	t.Helper()

	if err := waitForCMDBReady(newIntegrationContext(target), func(nCtx contextx.IContext) error {
		return probeIntegrationCMDB(nCtx, target, client)
	}); err != nil {
		t.Fatalf("wait for cmdb client ready: %v", err)
	}
}

type cmdbReadinessResponse struct {
	Result  bool   `json:"result"`
	Code    int    `json:"bk_error_code"`
	Message string `json:"bk_error_msg"`
}

type cmdbReadinessError struct {
	code    int
	message string
}

func (e *cmdbReadinessError) Error() string {
	return fmt.Sprintf("cmdb readiness failed: code(%d), msg(%s)", e.code, e.message)
}

func probeIntegrationCMDB(
	nCtx contextx.IContext, target support.CMDBTarget, client restclient.HTTPClient,
) error {
	body, err := json.Marshal(&SearchBusinessReq{Page: Page{Start: 0, Limit: 1}})
	if err != nil {
		return fmt.Errorf("marshal CMDB readiness request: %w", err)
	}

	endpoint := fmt.Sprintf("%s/api/v3/biz/search/%s", target.Endpoint, url.PathEscape(target.SupplierAccount))
	req, err := http.NewRequestWithContext(nCtx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create CMDB readiness request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Bk-Tenant-Id", target.TenantID)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("send CMDB readiness request: %w", err)
	}
	defer resp.Body.Close()

	result := new(cmdbReadinessResponse)
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("decode CMDB readiness response: %w", err)
	}
	if !result.Result {
		return &cmdbReadinessError{code: result.Code, message: result.Message}
	}

	return nil
}

func newIntegrationCapability(t *testing.T, target support.CMDBTarget) *restclient.Capability {
	t.Helper()

	baseClient, err := restclient.NewHTTPClient(nil)
	if err != nil {
		t.Fatal(err)
	}
	baseClient.CheckRedirect = checkSameOriginRedirect
	httpClient := &legacyHeaderHTTPClient{
		client:          baseClient,
		supplierAccount: target.SupplierAccount,
		virtualUser:     target.VirtualUser,
		appCode:         target.AppCode,
	}

	return &restclient.Capability{
		Name:                 "cmdb-integration",
		HTTPClient:           httpClient,
		Discover:             restdiscovery.NewDiscovery("cmdb-integration", []string{target.Endpoint}),
		ToleranceLatencyTime: restclient.ToleranceLatencyTimeDefault,
		MetricOpts:           restclient.MetricOption{},
		TraceSvc:             newIntegrationTraceService(t),
	}
}

func checkSameOriginRedirect(req *http.Request, via []*http.Request) error {
	if len(via) == 0 || sameOrigin(req.URL, via[0].URL) {
		return nil
	}

	return fmt.Errorf("cmdb integration redirect blocked from %s to %s", via[0].URL.String(), req.URL.String())
}

func sameOrigin(targetURL, sourceURL *url.URL) bool {
	return targetURL.Scheme == sourceURL.Scheme && targetURL.Host == sourceURL.Host
}

func newIntegrationConfig(target support.CMDBTarget) *Config {
	appCode := target.AppCode
	if appCode == "" {
		appCode = "bk-nodemgr-integration"
	}
	appConfig := apigwclient.NewAppConfig([]string{target.Endpoint}, appCode, "legacy-direct-apiserver")

	return &Config{
		SupplierAccount: target.SupplierAccount,
		APIGWUserConfig: apigwclient.UserConfig{
			AppConfig: appConfig,
			AuthMode:  apigwclient.AuthModeUn,
			LoginName: target.VirtualUser,
		},
	}
}

type legacyHeaderHTTPClient struct {
	client          restclient.HTTPClient
	supplierAccount string
	virtualUser     string
	appCode         string
}

func (c *legacyHeaderHTTPClient) Do(req *http.Request) (*http.Response, error) {
	reqWithHeaders := req.Clone(req.Context())
	reqWithHeaders.Header = req.Header.Clone()
	reqWithHeaders.Header.Set("BK_User", c.virtualUser)
	reqWithHeaders.Header.Set("HTTP_BK_SUPPLIER_ACCOUNT", c.supplierAccount)
	reqWithHeaders.Header.Set("HTTP_BLUEKING_SUPPLIER_ID", c.supplierAccount)
	if c.appCode != "" {
		reqWithHeaders.Header.Set("Bk-App-Code", c.appCode)
	}

	return c.client.Do(reqWithHeaders)
}

func newIntegrationTraceService(t *testing.T) tracing.IService {
	t.Helper()

	traceSvc, err := tracing.G().NewService(tracing.ServiceConfig{
		ServiceName: "cmdb-integration",
		SampleRate:  0,
	})
	if err != nil {
		t.Fatal(err)
	}

	return traceSvc
}
