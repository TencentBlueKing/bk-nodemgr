/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package file provides precheck tests that verify init packages are uploaded and published.
package file

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"

	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	protoFile "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/file/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	"github.com/TencentBlueKing/bk-nodemgr/test"
	"github.com/TencentBlueKing/bk-nodemgr/test/helper"
	"github.com/stretchr/testify/require"
)

const defaultGeneration = 2

func TestMain(m *testing.M) {
	test.ParseFlagsAndLoadEnv()
	os.Exit(m.Run())
}

// getCertReleaseListURL returns the cert release list API URL.
func getCertReleaseListURL() string {
	return helper.GetBackendBasicURL() + "/package/release/cert/list"
}

// getBinToolReleaseListURL returns the bintool release list API URL.
func getBinToolReleaseListURL() string {
	return helper.GetBackendBasicURL() + "/package/release/bintool/list"
}

// getAgentReleaseListURL returns the agent release list API URL.
func getAgentReleaseListURL() string {
	return helper.GetBackendBasicURL() + "/package/release/agent/list"
}

// getPluginBinToolReleaseListURL returns the plugin bintool release list API URL.
func getPluginBinToolReleaseListURL() string {
	return helper.GetBackendBasicURL() + "/package/release/plugin_bintool/list"
}

// getPluginReleaseListURL returns the plugin release list API URL.
func getPluginReleaseListURL() string {
	return helper.GetBackendBasicURL() + "/package/release/plugin/list"
}

// getUploadOriginCertURL returns the cert origin upload API URL.
func getUploadOriginCertURL() string {
	return helper.GetFileBasicURL() + "/upload/origin/cert"
}

// getUploadOriginBinToolURL returns the bintool origin upload API URL.
func getUploadOriginBinToolURL() string {
	return helper.GetFileBasicURL() + "/upload/origin/bintool"
}

// getUploadOriginAgentURL returns the agent origin upload API URL.
func getUploadOriginAgentURL() string {
	return helper.GetFileBasicURL() + "/upload/origin/agent"
}

// getUploadOriginPluginBinToolURL returns the plugin bintool origin upload API URL.
func getUploadOriginPluginBinToolURL() string {
	return helper.GetFileBasicURL() + "/upload/origin/plugin_bintool"
}

// getUploadOriginPluginURL returns the plugin origin upload API URL.
func getUploadOriginPluginURL() string {
	return helper.GetFileBasicURL() + "/upload/origin/v2/plugin"
}

// getPublishReleaseCertURL returns the cert release publish API URL.
func getPublishReleaseCertURL() string {
	return helper.GetFileBasicURL() + "/publish/release/cert"
}

// getPublishReleaseBinToolURL returns the bintool release publish API URL.
func getPublishReleaseBinToolURL() string {
	return helper.GetFileBasicURL() + "/publish/release/bintool"
}

// getPublishReleaseAgentURL returns the agent release publish API URL.
func getPublishReleaseAgentURL() string {
	return helper.GetFileBasicURL() + "/publish/release/agent"
}

// getPublishReleasePluginBinToolURL returns the plugin bintool release publish API URL.
func getPublishReleasePluginBinToolURL() string {
	return helper.GetFileBasicURL() + "/publish/release/plugin_bintool"
}

// getPublishReleasePluginURL returns the plugin release publish API URL.
func getPublishReleasePluginURL() string {
	return helper.GetFileBasicURL() + "/publish/release/v2/plugin"
}

// newDefaultPage creates a default page request.
func newDefaultPage() *protoBackend.Page {
	return &protoBackend.Page{Offset: 0, Limit: 100}
}

// newCertReleaseListReq creates a default cert release list request.
func newCertReleaseListReq() *protoBackend.PackageReleaseCertListReq {
	return &protoBackend.PackageReleaseCertListReq{Generation: defaultGeneration}
}

// newBinToolReleaseListReq creates a default bintool release list request.
func newBinToolReleaseListReq() *protoBackend.PackageReleaseBinToolListReq {
	return &protoBackend.PackageReleaseBinToolListReq{Generation: defaultGeneration}
}

// newAgentReleaseListReq creates a default agent release list request.
func newAgentReleaseListReq() *protoBackend.PackageReleaseAgentListReq {
	return &protoBackend.PackageReleaseAgentListReq{
		Generation: defaultGeneration,
		Page:       newDefaultPage(),
	}
}

// newPluginBinToolReleaseListReq creates a default plugin bintool release list request.
func newPluginBinToolReleaseListReq() *protoBackend.PackageReleasePluginBinToolListReq {
	return &protoBackend.PackageReleasePluginBinToolListReq{Generation: defaultGeneration}
}

// newPluginReleaseListReq creates a default plugin release list request.
func newPluginReleaseListReq() *protoBackend.PackageReleasePluginListReq {
	return &protoBackend.PackageReleasePluginListReq{
		Generation: defaultGeneration,
		Page:       newDefaultPage(),
	}
}

// TestInitPackages uploads, publishes and verifies that init packages are available.
func TestInitPackages(t *testing.T) {
	fileEnv := test.Env.Precheck.File

	tests := []struct {
		name             string
		filePath         string
		uploadAndPublish func(t *testing.T, filePath string)
		verifyRelease    func(t *testing.T)
	}{
		{
			name:             "CertReleaseAvailable",
			filePath:         fileEnv.OriginCertPath,
			uploadAndPublish: uploadAndPublishCert,
			verifyRelease:    verifyCertRelease,
		},
		{
			name:             "BinToolReleaseAvailable",
			filePath:         fileEnv.OriginBinToolPath,
			uploadAndPublish: uploadAndPublishBinTool,
			verifyRelease:    verifyBinToolRelease,
		},
		{
			name:             "AgentReleaseAvailable",
			filePath:         fileEnv.OriginAgentPath,
			uploadAndPublish: uploadAndPublishAgent,
			verifyRelease:    verifyAgentRelease,
		},
		{
			name:             "PluginBinToolReleaseAvailable",
			filePath:         fileEnv.OriginPluginBinToolPath,
			uploadAndPublish: uploadAndPublishPluginBinTool,
			verifyRelease:    verifyPluginBinToolRelease,
		},
		{
			name:             "PluginReleaseAvailable",
			filePath:         fileEnv.OriginPluginPath,
			uploadAndPublish: uploadAndPublishPlugin,
			verifyRelease:    verifyPluginRelease,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotEmpty(t, tt.filePath, "%s file path not configured", tt.name)
			require.FileExists(t, tt.filePath)

			tt.uploadAndPublish(t, tt.filePath)
			tt.verifyRelease(t)
		})
	}
}

func uploadAndPublishCert(t *testing.T, filePath string) {
	t.Helper()

	// upload origin cert
	reqBody, err := json.Marshal(&protoFile.UploadOriginCertReq{Overwrite: true})
	require.NoError(t, err)

	resp := helper.SendUploadHTTPRequest(t, getUploadOriginCertURL(), filePath, reqBody)

	var uploadResp protoFile.UploadOriginCertResp
	helper.ParseResponse(t, resp, &uploadResp)
	require.Equal(t, int32(errf.OK), uploadResp.Code)
	require.NotNil(t, uploadResp.Data)
	require.NotEmpty(t, uploadResp.Data.GetUploadId())

	// publish release cert
	reqBody, err = json.Marshal(&protoFile.PublishReleaseCertReq{UploadId: uploadResp.Data.GetUploadId()})
	require.NoError(t, err)

	resp = helper.SendHTTPRequest(t, http.MethodPost, getPublishReleaseCertURL(), reqBody)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func uploadAndPublishBinTool(t *testing.T, filePath string) {
	t.Helper()

	// upload origin bin tool
	reqBody, err := json.Marshal(&protoFile.UploadOriginBinToolReq{Generation: defaultGeneration, Overwrite: true})
	require.NoError(t, err)

	resp := helper.SendUploadHTTPRequest(t, getUploadOriginBinToolURL(), filePath, reqBody)

	var uploadResp protoFile.UploadOriginBinToolResp
	helper.ParseResponse(t, resp, &uploadResp)
	require.Equal(t, int32(errf.OK), uploadResp.Code)
	require.NotNil(t, uploadResp.Data)
	require.NotEmpty(t, uploadResp.Data.GetUploadId())

	// publish release bin tool
	reqBody, err = json.Marshal(&protoFile.PublishReleaseBinToolReq{UploadId: uploadResp.Data.GetUploadId()})
	require.NoError(t, err)

	resp = helper.SendHTTPRequest(t, http.MethodPost, getPublishReleaseBinToolURL(), reqBody)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func uploadAndPublishAgent(t *testing.T, filePath string) {
	t.Helper()

	// upload origin agent
	reqBody, err := json.Marshal(&protoFile.UploadOriginAgentReq{Generation: defaultGeneration, Overwrite: true})
	require.NoError(t, err)

	resp := helper.SendUploadHTTPRequest(t, getUploadOriginAgentURL(), filePath, reqBody)

	var uploadResp protoFile.UploadOriginAgentResp
	helper.ParseResponse(t, resp, &uploadResp)
	require.Equal(t, int32(errf.OK), uploadResp.Code)
	require.NotNil(t, uploadResp.Data)
	require.NotEmpty(t, uploadResp.Data.GetUploadId())

	// publish release agent
	reqBody, err = json.Marshal(&protoFile.PublishReleaseAgentReq{UploadId: uploadResp.Data.GetUploadId()})
	require.NoError(t, err)

	resp = helper.SendHTTPRequest(t, http.MethodPost, getPublishReleaseAgentURL(), reqBody)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func verifyCertRelease(t *testing.T) {
	t.Helper()

	reqBody, err := json.Marshal(newCertReleaseListReq())
	require.NoError(t, err)

	resp := helper.SendHTTPRequest(t, http.MethodPost, getCertReleaseListURL(), reqBody)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, resp.Body)

	var releaseListResp protoBackend.PackageReleaseCertListResp
	helper.ParseResponse(t, resp, &releaseListResp)
	require.Equal(t, int32(errf.OK), releaseListResp.Code)
	require.NotNil(t, releaseListResp.Data)
	require.Greater(t, releaseListResp.Data.Total, int64(0))
}

func verifyBinToolRelease(t *testing.T) {
	t.Helper()

	reqBody, err := json.Marshal(newBinToolReleaseListReq())
	require.NoError(t, err)

	resp := helper.SendHTTPRequest(t, http.MethodPost, getBinToolReleaseListURL(), reqBody)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, resp.Body)

	var releaseListResp protoBackend.PackageReleaseBinToolListResp
	helper.ParseResponse(t, resp, &releaseListResp)
	require.Equal(t, int32(errf.OK), releaseListResp.Code)
	require.NotNil(t, releaseListResp.Data)
	require.Greater(t, releaseListResp.Data.Total, int64(0))
}

func verifyAgentRelease(t *testing.T) {
	t.Helper()

	reqBody, err := json.Marshal(newAgentReleaseListReq())
	require.NoError(t, err)

	resp := helper.SendHTTPRequest(t, http.MethodPost, getAgentReleaseListURL(), reqBody)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, resp.Body)

	var releaseListResp protoBackend.PackageReleaseAgentListResp
	helper.ParseResponse(t, resp, &releaseListResp)
	require.Equal(t, int32(errf.OK), releaseListResp.Code)
	require.NotNil(t, releaseListResp.Data)
	require.Greater(t, releaseListResp.Data.Total, int64(0))
}

func uploadAndPublishPluginBinTool(t *testing.T, filePath string) {
	t.Helper()

	// upload origin plugin bintool
	reqBody, err := json.Marshal(&protoFile.UploadOriginPluginBinToolReq{Overwrite: true})
	require.NoError(t, err)

	resp := helper.SendUploadHTTPRequest(t, getUploadOriginPluginBinToolURL(), filePath, reqBody)

	var uploadResp protoFile.UploadOriginPluginBinToolResp
	helper.ParseResponse(t, resp, &uploadResp)
	require.Equal(t, int32(errf.OK), uploadResp.Code)
	require.NotNil(t, uploadResp.Data)
	require.NotEmpty(t, uploadResp.Data.GetUploadId())

	// publish release plugin bintool
	reqBody, err = json.Marshal(&protoFile.PublishReleasePluginBinToolReq{UploadId: uploadResp.Data.GetUploadId()})
	require.NoError(t, err)

	resp = helper.SendHTTPRequest(t, http.MethodPost, getPublishReleasePluginBinToolURL(), reqBody)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func uploadAndPublishPlugin(t *testing.T, filePath string) {
	t.Helper()

	// upload origin plugin
	reqBody, err := json.Marshal(&protoFile.UploadOriginPluginV2Req{Overwrite: true})
	require.NoError(t, err)

	resp := helper.SendUploadHTTPRequest(t, getUploadOriginPluginURL(), filePath, reqBody)

	var uploadResp protoFile.UploadOriginPluginV2Resp
	helper.ParseResponse(t, resp, &uploadResp)
	require.Equal(t, int32(errf.OK), uploadResp.Code)
	require.NotNil(t, uploadResp.Data)
	require.NotEmpty(t, uploadResp.Data.GetUploadId())

	// publish release plugin
	reqBody, err = json.Marshal(&protoFile.PublishReleasePluginV2Req{UploadId: uploadResp.Data.GetUploadId()})
	require.NoError(t, err)

	resp = helper.SendHTTPRequest(t, http.MethodPost, getPublishReleasePluginURL(), reqBody)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func verifyPluginBinToolRelease(t *testing.T) {
	t.Helper()

	reqBody, err := json.Marshal(newPluginBinToolReleaseListReq())
	require.NoError(t, err)

	resp := helper.SendHTTPRequest(t, http.MethodPost, getPluginBinToolReleaseListURL(), reqBody)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, resp.Body)

	var releaseListResp protoBackend.PackageReleasePluginBinToolListResp
	helper.ParseResponse(t, resp, &releaseListResp)
	require.Equal(t, int32(errf.OK), releaseListResp.Code)
	require.NotNil(t, releaseListResp.Data)
	require.Greater(t, releaseListResp.Data.Total, int64(0))
}

func verifyPluginRelease(t *testing.T) {
	t.Helper()

	reqBody, err := json.Marshal(newPluginReleaseListReq())
	require.NoError(t, err)

	resp := helper.SendHTTPRequest(t, http.MethodPost, getPluginReleaseListURL(), reqBody)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, resp.Body)

	var releaseListResp protoBackend.PackageReleasePluginListResp
	helper.ParseResponse(t, resp, &releaseListResp)
	require.Equal(t, int32(errf.OK), releaseListResp.Code)
	require.NotNil(t, releaseListResp.Data)
	require.Greater(t, releaseListResp.Data.Total, int64(0))
}
