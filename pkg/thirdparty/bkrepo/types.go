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

package bkrepo

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// CodeOK is the code of success.
const CodeOK = 0

var (
	errInvalidContext  = errors.New("invalid context")
	errEmptyPathOrName = errors.New("empty path or name")
	errNodeNotFound    = errors.New("node not found")
)

const (
	// BKRepo's node detail API actually returns 251010 for missing nodes, unlike
	// the documented "resource not found" code 250108. Keep the actual API code.
	// See https://github.com/TencentBlueKing/bk-nodemgr/issues/3575.
	errNumNodeNotFound = 251010
)

// RespCommon describes the common response.
type RespCommon struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	TraceID string `json:"traceId"`
}

// BaseBroker describe the base broker.
type BaseBroker[T any] struct {
	RespCommon
	Data T `json:"data"`
}

// IsFailed returns the code of response.
func (resp *BaseBroker[T]) IsFailed() error {
	if resp.Code != CodeOK {
		switch resp.Code {
		case errNumNodeNotFound:
			return errors.Join(errNodeNotFound, fmt.Errorf("code(%d) , msg(%s) ", resp.Code, resp.Message))

		default:
			return fmt.Errorf("code(%d) , msg(%s) ", resp.Code, resp.Message)
		}
	}

	return nil
}

// DownloadFileReq describe the download file request.
type DownloadFileReq struct {
	Path string
}

// DownloadFileResp describes the download file response.
type DownloadFileResp struct {
	Data io.ReadCloser
}

// UploadFileReq describe the upload file request.
type UploadFileReq struct {
	Path   string
	Info   UploadFileInfo
	Reader io.Reader
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
		header.Set("X-BKREPO-SHA256", info.Sha256) // nolint:canonicalheader
	}

	if info.Md5 != "" {
		header.Set("X-BKREPO-MD5", info.Md5) // nolint:canonicalheader
	}

	header.Set("X-BKREPO-OVERWRITE", strconv.FormatBool(info.Overwrite))   // nolint:canonicalheader
	header.Set("X-BKREPO-EXPIRES", strconv.FormatInt(info.ExpireDays, 10)) // nolint:canonicalheader

	if len(info.Meta) > 0 {
		metas := []string{}
		for k, v := range info.Meta {
			metas = append(metas, fmt.Sprintf("%s=%s", k, v))
		}

		header.Set("X-BKREPO-META", strings.Join(metas, "&")) // nolint:canonicalheader
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
	ProjectID        string         `json:"projectId"`
	RepoName         string         `json:"repoName"`
}

// QueryNodeInfoReq describe the query node info request.
type QueryNodeInfoReq struct {
	Path string
}

// QueryNodeInfoResp describe the query node info response.
type QueryNodeInfoResp struct {
	NodeInfo NodeInfo `json:"nodeInfo"`
}

// NodeInfo describe the node info.
type NodeInfo struct {
	ProjectID        string         `json:"projectId"`
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
	Path     string
	PageNum  int
	PageSize int
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
	ID           string         `json:"id"`
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
	ProjectID    string         `json:"projectId"`
	RepoName     string         `json:"repoName"`
}

// MkdirReq describe the mkdir request.
type MkdirReq struct {
	Path string
}

// MkdirResp describe the mkdir response.
type MkdirResp struct {
}

// CopyNodeReq describes the request for copying a node.
type CopyNodeReq struct {
	SrcProjectID  string `json:"srcProjectId"`
	SrcRepoName   string `json:"srcRepoName"`
	SrcFullPath   string `json:"srcFullPath"`
	DestProjectID string `json:"destProjectId"`
	DestRepoName  string `json:"destRepoName"`
	DestFullPath  string `json:"destFullPath"`
	Overwrite     bool   `json:"overwrite"`
}

// CopyNodeResp describes the response for copying a node.
type CopyNodeResp struct {
}

// DeleteNodeReq describes the request for deleting a node.
type DeleteNodeReq struct {
	Path string
}

// DeleteNodeResp describes the response for deleting a node.
type DeleteNodeResp struct {
	DeletedNumber int    `json:"deletedNumber"`
	DeletedSize   int64  `json:"deletedSize"`
	DeletedTime   string `json:"deletedTime"`
}
