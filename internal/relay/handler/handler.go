/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package handler ...
package handler

import (
	"context"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/relay/file"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/relayhandler"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
)

const (
	// ReportPrivateDataTimeout defines the report private data timeout.
	ReportPrivateDataTimeout = 3 * time.Second
)

// IHandler defines a relay client handler.
type IHandler interface {
	CheckPkgStats(ctx context.Context, payload []byte)
	StoragePkg(ctx context.Context, payload []byte)
	DetectInfoBySSH(ctx context.Context, payload []byte)
	InstallPagentBySSH(ctx context.Context, payload []byte)
}

// handler is a relay client handler.
type handler struct {
	storageTmpDir string

	fileManager file.IFileManager
	client      relayhandler.IClientMessager

	callbackSvcIP   string
	callbackSvcPort int

	fileSvcIP   string
	fileSvcPort int

	logger logger.ILogger
}

// NewClientHandler creates a new file handler.
func NewClientHandler(fm file.IFileManager, client relayhandler.IClientMessager,
	logger logger.ILogger, storageTmpDir string,
	callbackSvcIP string, callbackSvcPort int, fileSvcIP string, fileSvcPort int) IHandler {

	return &handler{
		storageTmpDir:   storageTmpDir,
		fileManager:     fm,
		client:          client,
		callbackSvcIP:   callbackSvcIP,
		callbackSvcPort: callbackSvcPort,
		fileSvcIP:       fileSvcIP,
		fileSvcPort:     fileSvcPort,
		logger:          logger,
	}
}
