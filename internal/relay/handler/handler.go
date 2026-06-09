/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package handler provides the handler for the relay client.
package handler

import (
	"path/filepath"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/relay/file"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/relayhandler"
)

const (
	// ReportPrivateDataTimeout defines the report private data timeout.
	ReportPrivateDataTimeout = 3 * time.Second
	storageTmpDirName        = "transfer-file"
)

// IHandler defines a relay client handler.
type IHandler interface {
	// CheckPkgStats checks the package stats.
	CheckPkgStats(nCtx contextx.IContext, payload []byte)
	// StoragePkg stores the package.
	StoragePkg(nCtx contextx.IContext, payload []byte)
	// DetectInfoBySSH detects the node info by ssh.
	DetectInfoBySSH(nCtx contextx.IContext, payload []byte)
	// InstallPagentBySSH installs the pagent by ssh.
	InstallPagentBySSH(nCtx contextx.IContext, payload []byte)
	// DetectInfoByWMI detects the node info by wmi.
	DetectInfoByWMI(nCtx contextx.IContext, payload []byte)
	// InstallPagentByWMI installs the pagent by wmi.
	InstallPagentByWMI(nCtx contextx.IContext, payload []byte)
}

// handler is a relay client handler.
type handler struct {
	storageTmpDir string
	storageFS     workspaceFS

	fileManager file.IFileManager
	client      relayhandler.IClientMessager
}

// NewClientHandler creates a new client handler.
func NewClientHandler(
	fm file.IFileManager,
	client relayhandler.IClientMessager,
	conf *config.RelayService,
) IHandler {
	storageTmpDir := filepath.Join(conf.RelayWorkspaceFileGroup.FullPath, storageTmpDirName)

	return &handler{
		fileManager:   fm,
		client:        client,
		storageTmpDir: storageTmpDir,
		storageFS:     newWorkspaceFS(storageTmpDir),
	}
}
