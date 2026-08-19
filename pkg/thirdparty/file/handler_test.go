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

package file

import (
	"os"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	restdiscovery "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/discovery"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/ssl"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/joho/godotenv"
)

var (
	agentFileName = ""
	proxyFileName = ""
)

func initParams(t *testing.T) {
	err := godotenv.Load(".env")
	if err != nil {
		t.Fatal(err)
	}

	agentFileName = os.Getenv("AGENT_FILE_NAME")
	proxyFileName = os.Getenv("PROXY_FILE_NAME")
}

// testClient ...
func testClient(t *testing.T) IHandler {
	err := godotenv.Load(".env")
	if err != nil {
		t.Fatal(err)
	}

	httpClient, err := restclient.NewHTTPClient(&ssl.TLSConfig{
		InsecureSkipVerify: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	clientCap := &restclient.Capability{
		HTTPClient:           httpClient,
		Discover:             restdiscovery.NewDiscovery("apigateway", []string{os.Getenv("BK_APIGW_ENDPOINT")}),
		ToleranceLatencyTime: restclient.ToleranceLatencyTimeDefault,
		MetricOpts:           restclient.MetricOption{},
	}

	h, err := New(clientCap, &Config{
		RestJwtSecret: "test",
	})
	if err != nil {
		t.Fatal(err)
	}

	return h
}

// Test_handler_UploadOriginAgent tests the UploadOriginAgent method.
func Test_handler_UploadOriginAgent(t *testing.T) {
	initParams(t)

	file, err := os.Open(agentFileName)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	data, err := testClient(t).UploadOriginAgent(contextx.Background(), agentFileName, file, types.Generation2, true)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("data: %+#v", data)
}

// Test_handler_UploadOriginProxy tests the UploadOriginProxy method.
func Test_handler_UploadOriginProxy(t *testing.T) {
	initParams(t)

	file, err := os.Open(proxyFileName)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	data, err := testClient(t).UploadOriginProxy(contextx.Background(), proxyFileName, file, types.Generation2, true)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("data: %+#v", data)
}
