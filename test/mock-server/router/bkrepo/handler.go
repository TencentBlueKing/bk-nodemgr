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

// Package bkrepo provides the BKRepo mock API handler.
package bkrepo

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/test/mock-server/router/common"
	"github.com/gin-gonic/gin"
)

// handler defines the BKRepo mock API handler.
type handler struct {
	rg    *gin.RouterGroup
	store *storage
}

// newHandler creates a new BKRepo handler.
func newHandler(rg *gin.RouterGroup, store *storage) *handler {
	return &handler{
		rg:    rg,
		store: store,
	}
}

// QueryNodeInfo handles node detail query request.
func (h *handler) QueryNodeInfo(gCtx *gin.Context) {
	var params PathParams
	if err := common.BindURI(gCtx, &params); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to bind uri parameters")
		respondError(gCtx, CodeInvalidParameter, fmt.Sprintf("invalid uri parameters: %v", err))

		return
	}

	// query node info.
	resp, err := h.store.QueryNodeInfo(params.Project, params.Repo, params.Path)
	if err != nil {
		if isNodeNotFound(err) {
			logger.G.Sys().WithErr(err).
				With("project", params.Project,
					"repo", params.Repo,
					"path", params.Path).
				Warn("node not found")
			respondError(gCtx, CodeNodeNotFound, fmt.Sprintf("node not found: %v", err))

			return
		}

		logger.G.Sys().WithErr(err).
			With("project", params.Project,
				"repo", params.Repo,
				"path", params.Path).
			Error("failed to query node info")
		respondError(gCtx, CodeServerError, fmt.Sprintf("failed to query node info: %v", err))

		return
	}

	respondSuccess(gCtx, resp)
}

// MkDir handles mkdir request.
func (h *handler) MkDir(gCtx *gin.Context) {
	var params PathParams
	if err := common.BindURI(gCtx, &params); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to bind uri parameters")
		respondError(gCtx, CodeInvalidParameter, fmt.Sprintf("invalid uri parameters: %v", err))

		return
	}

	// create directory.
	resp, err := h.store.CreateDir(params.Project, params.Repo, params.Path)
	if err != nil {
		logger.G.Sys().WithErr(err).
			With("project", params.Project,
				"repo", params.Repo,
				"path", params.Path).
			Error("failed to create directory")
		respondError(gCtx, CodeServerError, fmt.Sprintf("failed to create directory: %v", err))

		return
	}

	respondSuccess(gCtx, resp)
}

// UploadFile handles file upload request.
func (h *handler) UploadFile(gCtx *gin.Context) {
	var params PathParams
	if err := common.BindURI(gCtx, &params); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to bind uri parameters")
		respondError(gCtx, CodeInvalidParameter, fmt.Sprintf("invalid uri parameters: %v", err))

		return
	}

	// parse overwrite header.
	overwrite, err := common.GetHeaderBool(gCtx, HeaderOverwrite)
	if err != nil {
		logger.G.Sys().WithErr(err).With("header", HeaderOverwrite).Error("invalid overwrite header")
		respondError(gCtx, CodeInvalidParameter, fmt.Sprintf("invalid header: %v", err))

		return
	}

	// ensure request body is closed after upload.
	defer func() {
		if err := gCtx.Request.Body.Close(); err != nil {
			logger.G.Sys().WithErr(err).Error("failed to close request body")
		}
	}()

	// upload file.
	resp, err := h.store.UploadFile(params.Project, params.Repo, params.Path, gCtx.Request.Body, overwrite)
	if err != nil {
		logger.G.Sys().WithErr(err).
			With("project", params.Project,
				"repo", params.Repo,
				"path", params.Path,
				"overwrite", overwrite).
			Error("failed to upload file")

		respondError(gCtx, CodeServerError, fmt.Sprintf("failed to upload file: %v", err))

		return
	}

	respondSuccess(gCtx, resp)
}

// DownloadFile handles file download request.
func (h *handler) DownloadFile(gCtx *gin.Context) {
	var params PathParams
	if err := common.BindURI(gCtx, &params); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to bind uri parameters")
		respondError(gCtx, CodeInvalidParameter, fmt.Sprintf("invalid uri parameters: %v", err))

		return
	}

	// download file.
	reader, info, err := h.store.DownloadFile(params.Project, params.Repo, params.Path)
	if err != nil {
		logger.G.Sys().WithErr(err).
			With("project", params.Project,
				"repo", params.Repo,
				"path", params.Path).
			Error("failed to download file")

		if isNodeNotFound(err) {
			respondError(gCtx, CodeNodeNotFound, fmt.Sprintf("node not found: %v", err))

			return
		}

		respondError(gCtx, CodeServerError, fmt.Sprintf("failed to download file: %v", err))

		return
	}
	defer func() {
		err := reader.Close()
		if err != nil {
			logger.G.Sys().WithErr(err).Error("failed to close file")
		}
	}()

	respondFile(gCtx, reader, info.Size())
}

// ListNode handles node list request with pagination.
func (h *handler) ListNode(gCtx *gin.Context) {
	var params ListNodeParams
	if err := common.BindURI(gCtx, &params); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to bind uri parameters")
		respondError(gCtx, CodeInvalidParameter, fmt.Sprintf("invalid uri parameters: %v", err))

		return
	}

	if err := common.BindQuery(gCtx, &params); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to bind query parameters")
		respondError(gCtx, CodeInvalidParameter, fmt.Sprintf("invalid query parameters: %v", err))

		return
	}

	// list nodes.
	resp, err := h.store.ListNode(params.Project, params.Repo, params.Path, params.PageNum, params.PageSize)
	if err != nil {
		logger.G.Sys().WithErr(err).
			With("project", params.Project,
				"repo", params.Repo,
				"path", params.Path,
				"pageNum", params.PageNum,
				"pageSize", params.PageSize).
			Error("failed to list nodes")

		if isNodeNotFound(err) {
			respondError(gCtx, CodeNodeNotFound, fmt.Sprintf("node not found: %v", err))

			return
		}

		respondError(gCtx, CodeServerError, fmt.Sprintf("failed to list nodes: %v", err))

		return
	}

	respondSuccess(gCtx, resp)
}
