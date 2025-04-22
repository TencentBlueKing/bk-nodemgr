/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package bkrepo

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// CodeOK is the code of success.
const CodeOK = 0

// RespCommon describes the common response.
type RespCommon struct {
	Code    int    `json:"code"`
	Message string `json:"messaage"`
	TraceId string `json:"traceId"`
}

// BaseBroker describe the base broker.
type BaseBroker[T any] struct {
	RespCommon
	Data T `json:"data"`
}

// IsFailed returns the code of response.
func (resp *BaseBroker[T]) IsFailed() error {
	if resp.Code != CodeOK {
		return fmt.Errorf("code(%d) , msg(%s) ", resp.Code, resp.Message)
	}

	return nil
}

// DownloadFileReq describe the download file request.
type DownloadFileReq struct {
	ProjectID string
	RepoName  string
	Path      string
}

// DownloadFileResp describe the download file response.
type DownloadFileResp struct {
	Data []byte
}

// UploadFileReq describe the upload file request.
type UploadFileReq struct {
	ProjectId string
	RepoName  string
	Path      string
	Info      UploadFileInfo
	Data      []byte
}

// UploadFileInfo upload file info.
type UploadFileInfo struct {
	Sha256     string
	Md5        string
	Overwrite  bool
	ExpireDays int64
	Meta       map[string]string
}

// BindHeader bind http header.
func (info UploadFileInfo) BindHeader(header http.Header) http.Header {
	if info.Sha256 != "" {
		header.Set("X-BKREPO-SHA256", info.Sha256)
	}

	if info.Md5 != "" {
		header.Set("X-BKREPO-MD5", info.Md5)
	}

	header.Set("X-BKREPO-OVERWRITE", strconv.FormatBool(info.Overwrite))
	header.Set("X-BKREPO-EXPIRES", strconv.FormatInt(info.ExpireDays, 10))

	if len(info.Meta) > 0 {
		metas := []string{}
		for k, v := range info.Meta {
			metas = append(metas, fmt.Sprintf("%s=%s", k, v))
		}

		header.Set("X-BKREPO-META", strings.Join(metas, "&"))
	}

	return header
}

// UploadFileResp describe the upload file response.
type UploadFileResp struct {
	CreatedBy        string         `json:"createdBy"`
	CreatedDate      string         `json:"createdDate"`
	LastModifiedBy   string         `json:"lastModifiedBy"`
	LastModifiedDate string         `json:"lastModifiedDate"`
	Folder           bool           `json:"folder"`
	Path             string         `json:"path"`
	Name             string         `json:"name"`
	FullPath         string         `json:"fullPath"`
	Size             int64          `json:"size"`
	Sha256           string         `json:"sha256"`
	Md5              string         `json:"md5"`
	Metadata         map[string]any `json:"metadata"`
	ProjectId        string         `json:"projectId"`
	RepoName         string         `json:"repoName"`
}

// QueryNodeInfoReq describe the query node info request.
type QueryNodeInfoReq struct {
	ProjectId string
	RepoName  string
	Path      string
}

// QueryNodeInfoResp describe the query node info response.
type QueryNodeInfoResp struct {
	NodeInfo NodeInfo `json:"nodeInfo"`
}

// NodeInfo describe the node info.
type NodeInfo struct {
	ProjectId        string         `json:"projectId"`
	RepoName         string         `json:"repoName"`
	Path             string         `json:"path"`
	Name             string         `json:"name"`
	FullPath         string         `json:"fullPath"`
	Folder           bool           `json:"folder"`
	Size             int            `json:"size"`
	Sha256           string         `json:"sha256"`
	Md5              string         `json:"md5"`
	StageTag         string         `json:"stageTag"`
	Metadata         map[string]any `json:"metadata"`
	NodeMetadata     []NodeMetadata `json:"nodeMetadata"`
	CreatedBy        string         `json:"createdBy"`
	CreatedDate      string         `json:"createdDate"`
	LastModifiedBy   string         `json:"lastModifiedBy"`
	LastModifiedDate string         `json:"lastModifiedDate"`
}

// NodeMetadata node metadata.
type NodeMetadata struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Description string `json:"description"`
	System      bool   `json:"system"`
	Link        string `json:"link"`
}

// ListNodeReq describe the list node request.
type ListNodeReq struct {
	ProjectId string
	RepoName  string
	Path      string
	PageNum   int
	PageSize  int
}

// ListNodeResp describe the list node response.
type ListNodeResp struct {
	PageNumber   int          `json:"pageNumber"`
	PageSize     int          `json:"pageSize"`
	TotalRecords int          `json:"totalRecords"`
	TotalPages   int          `json:"totalPages"`
	Records      []NodeRecord `json:"records"`
	Count        int          `json:"count"`
	Page         int          `json:"page"`
}

// NodeRecord describe the node record.
type NodeRecord struct {
	Id           string         `json:"id"`
	Folder       bool           `json:"folder"`
	Path         string         `json:"path"`
	Name         string         `json:"name"`
	FullPath     string         `json:"fullPath"`
	Size         int            `json:"size"`
	NodeNum      int            `json:"nodeNum"`
	Sha256       string         `json:"sha256"`
	Md5          string         `json:"md5"`
	Metadata     map[string]any `json:"metadata"`
	NodeMetadata []NodeMetadata `json:"nodeMetadata"`
	ProjectId    string         `json:"projectId"`
	RepoName     string         `json:"repoName"`
}
