/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package upload provides the upload storage interface.
package upload

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IStorage defines the interface of upload storage.
type IStorage interface {
	basestorage.Interface

	IDaoAgent
	IDaoServer
	IDaoProxy
	IDaoCert
	IDaoBinTool
	IDaoPluginBinTool
	IDaoPluginV2
	IDaoExternalPluginV2
	IDaoPluginV3
}

// IDaoAgent defines the interface of upload storage.
type IDaoAgent interface {
	// GetAgentUpload gets a upload by upload-id.
	GetAgentUpload(nCtx contextx.IContext, uploadID string) (*types.Upload, error)

	// CreateAgentUpload creates a upload.
	CreateAgentUpload(nCtx contextx.IContext, up *types.Upload) (string, error)

	// DeleteAgentUpload deletes a upload by upload-id.
	DeleteAgentUpload(nCtx contextx.IContext, uploadID string) error
}

// IDaoServer defines the interface of upload storage.
type IDaoServer interface {
	// GetServerUpload gets a upload by upload-id.
	GetServerUpload(nCtx contextx.IContext, uploadID string) (*types.Upload, error)

	// CreateServerUpload creates a upload.
	CreateServerUpload(nCtx contextx.IContext, up *types.Upload) (string, error)

	// DeleteServerUpload deletes a upload by upload-id.
	DeleteServerUpload(nCtx contextx.IContext, uploadID string) error
}

// IDaoProxy defines the interface of upload storage.
type IDaoProxy interface {
	// GetProxyUpload gets a upload by upload-id.
	GetProxyUpload(nCtx contextx.IContext, uploadID string) (*types.Upload, error)

	// CreateProxyUpload creates a upload.
	CreateProxyUpload(nCtx contextx.IContext, up *types.Upload) (string, error)

	// DeleteProxyUpload deletes a upload by upload-id.
	DeleteProxyUpload(nCtx contextx.IContext, uploadID string) error
}

// IDaoCert defines the interface of upload storage.
type IDaoCert interface {
	// GetCertUpload gets a upload by upload-id.
	GetCertUpload(nCtx contextx.IContext, uploadID string) (*types.Upload, error)

	// CreateCertUpload creates a upload.
	CreateCertUpload(nCtx contextx.IContext, up *types.Upload) (string, error)

	// DeleteCertUpload deletes a upload by upload-id.
	DeleteCertUpload(nCtx contextx.IContext, uploadID string) error
}

// IDaoBinTool defines the interface of upload storage.
type IDaoBinTool interface {
	// GetBinToolUpload gets a upload by upload-id.
	GetBinToolUpload(nCtx contextx.IContext, uploadID string) (*types.Upload, error)

	// CreateBinToolUpload creates a upload.
	CreateBinToolUpload(nCtx contextx.IContext, up *types.Upload) (string, error)

	// DeleteBinToolUpload deletes a upload by upload-id.
	DeleteBinToolUpload(nCtx contextx.IContext, uploadID string) error
}

// IDaoPluginBinTool defines the interface of upload storage.
type IDaoPluginBinTool interface {
	// GetPluginBinToolUpload gets an upload by upload-id.
	GetPluginBinToolUpload(nCtx contextx.IContext, uploadID string) (*types.Upload, error)

	// CreatePluginBinToolUpload creates an upload.
	CreatePluginBinToolUpload(nCtx contextx.IContext, up *types.Upload) (string, error)

	// DeletePluginBinToolUpload deletes an upload by upload-id.
	DeletePluginBinToolUpload(nCtx contextx.IContext, uploadID string) error
}

// IDaoPluginV2 defines the interface of upload storage.
type IDaoPluginV2 interface {
	// GetPluginV2Upload gets a upload by upload-id.
	GetPluginV2Upload(nCtx contextx.IContext, uploadID string) (*types.Upload, error)

	// CreatePluginV2Upload creates a upload.
	CreatePluginV2Upload(nCtx contextx.IContext, up *types.Upload) (string, error)

	// DeletePluginV2Upload deletes a upload by upload-id.
	DeletePluginV2Upload(nCtx contextx.IContext, uploadID string) error
}

// IDaoExternalPluginV2 defines the interface of upload storage.
type IDaoExternalPluginV2 interface {
	// GetExternalPluginV2Upload gets a upload by upload-id.
	GetExternalPluginV2Upload(nCtx contextx.IContext, uploadID string) (*types.Upload, error)

	// CreateExternalPluginV2Upload creates a upload.
	CreateExternalPluginV2Upload(nCtx contextx.IContext, up *types.Upload) (string, error)

	// DeleteExternalPluginV2Upload deletes a upload by upload-id.
	DeleteExternalPluginV2Upload(nCtx contextx.IContext, uploadID string) error
}

// IDaoPluginV3 defines the interface of upload storage.
type IDaoPluginV3 interface {
	// GetPluginV3Upload gets a upload by upload-id.
	GetPluginV3Upload(nCtx contextx.IContext, uploadID string) (*types.Upload, error)

	// CreatePluginV3Upload creates a upload.
	CreatePluginV3Upload(nCtx contextx.IContext, up *types.Upload) (string, error)

	// DeletePluginV3Upload deletes a upload by upload-id.
	DeletePluginV3Upload(nCtx contextx.IContext, uploadID string) error
}
