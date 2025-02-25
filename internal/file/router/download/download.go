/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package download is the file download router.
package download

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/TencentBlueKing/bk-nodemgr/internal/file/options"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/gin-gonic/gin"
)

type handler struct {
	rg             *gin.RouterGroup
	agentFileGroup iface.FileGroup
	logger         logger.Logger
}

func newHandler(rg *gin.RouterGroup, opt *options.Capability) *handler {
	return &handler{
		// this is a sub router, so we can use some special middleware in it and not affect the father router.
		rg:             rg.Group("/download"),
		agentFileGroup: opt.AgentFileGroup,
		logger:         opt.Logger,
	}
}

// Load enables web router into gin.Engine.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	h.rg.GET("/agent/release/:filename", rest.FileHandler(h.ReleaseAgent))
}

// ReleaseAgent ...
func (h *handler) ReleaseAgent(ctx *rest.Context) (*rest.FileResponse, error) {
	filename := ctx.Param("filename")
	if filename == "" {
		return nil, errf.ErrWrap(errf.InvalidParameter, errors.New("filename is empty"))
	}

	file, err := h.agentFileGroup.GetFile(filename)
	if err != nil {
		return nil, errf.ErrWrap(errf.InvalidParameter, fmt.Errorf("get file failed, err: %w", err))
	}

	reader, err := file.Content()
	if err != nil {
		return nil, errf.ErrWrap(errf.InvalidParameter, fmt.Errorf("get file content failed, err: %w", err))
	}

	return &rest.FileResponse{
		Data:        reader,
		Size:        file.Info().Size,
		FilePath:    filepath.Join(".", file.Name()),
		FileName:    file.Name(),
		ContentType: "application/octet-stream",
	}, nil
}
