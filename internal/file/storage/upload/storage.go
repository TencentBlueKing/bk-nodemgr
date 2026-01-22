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
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/upload"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	// StorageName defines the storage name.
	StorageName = "upload"

	metricOperationGetAgentUpload               = "get_agent_upload"
	metricOperationCreateAgentUpload            = "create_agent_upload"
	metricOperationDeleteAgentUpload            = "delete_agent_upload"
	metricOperationGetServerUpload              = "get_server_upload"
	metricOperationCreateServerUpload           = "create_server_upload"
	metricOperationDeleteServerUpload           = "delete_server_upload"
	metricOperationGetProxyUpload               = "get_proxy_upload"
	metricOperationCreateProxyUpload            = "create_proxy_upload"
	metricOperationDeleteProxyUpload            = "delete_proxy_upload"
	metricOperationGetCertUpload                = "get_cert_upload"
	metricOperationCreateCertUpload             = "create_cert_upload"
	metricOperationDeleteCertUpload             = "delete_cert_upload"
	metricOperationGetBinToolUpload             = "get_bintool_upload"
	metricOperationCreateBinToolUpload          = "create_bintool_upload"
	metricOperationDeleteBinToolUpload          = "delete_bintool_upload"
	metricOperationGetPluginBinToolUpload       = "get_plugin_bintool_upload"
	metricOperationCreatePluginBinToolUpload    = "create_plugin_bintool_upload"
	metricOperationDeletePluginBinToolUpload    = "delete_plugin_bintool_upload"
	metricOperationGetPluginV2Upload            = "get_plugin_v2_upload"
	metricOperationCreatePluginV2Upload         = "create_plugin_v2_upload"
	metricOperationDeletePluginV2Upload         = "delete_plugin_v2_upload"
	metricOperationGetExternalPluginV2Upload    = "get_external_plugin_v2_upload"
	metricOperationCreateExternalPluginV2Upload = "create_external_plugin_v2_upload"
	metricOperationDeleteExternalPluginV2Upload = "delete_external_plugin_v2_upload"
	metricOperationGetPluginV3Upload            = "get_plugin_v3_upload"
	metricOperationCreatePluginV3Upload         = "create_plugin_v3_upload"
	metricOperationDeletePluginV3Upload         = "delete_plugin_v3_upload"
)

// NewStorage creates a new upload storage.
func NewStorage(client *mongo.Client, database string) (*Storage, error) {
	if client == nil {
		return nil, errors.New("mongo client is nil")
	}

	s := &Storage{
		Storage: basestorage.Storage{
			Name:     StorageName,
			Database: client.Database(database),
		},
	}
	err := basestorage.InitStorage(&s.Storage,
		basestorage.WithStartFunc(s.initDao),
		basestorage.WithCheckFunc(s.check))
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to new storage")

		return nil, err
	}

	return s, nil
}

var _ IStorage = &Storage{}

// Storage implements IStorage.
type Storage struct {
	basestorage.Storage

	daoUpload upload.IHandler
}

func (s *Storage) initDao() error {
	s.daoUpload = upload.New(s.Database)

	return nil
}

func (s *Storage) check() error {
	if s.daoUpload == nil {
		return errors.New("dao upload is nil")
	}

	return nil
}

// ===============================================================================
// UploadAgent Related Interface
// ===============================================================================

// GetAgentUpload gets a upload by upload-id.
func (s *Storage) GetAgentUpload(nCtx contextx.IContext, uploadID string) (*types.Upload, error) {
	var (
		data *types.Upload
		err  error
	)

	err = s.WrapFn(nCtx, metricOperationGetAgentUpload, func(nCtx contextx.IContext) error {
		data, err = s.getAgentUpload(nCtx, uploadID)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return data, nil
}

// CreateAgentUpload creates a upload.
func (s *Storage) CreateAgentUpload(nCtx contextx.IContext, up *types.Upload) (string, error) {
	var (
		uploadID string
		err      error
	)

	err = s.WrapFn(nCtx, metricOperationCreateAgentUpload, func(nCtx contextx.IContext) error {
		uploadID, err = s.createAgentUpload(nCtx, up)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	return uploadID, nil
}

// DeleteAgentUpload deletes a upload by upload-id.
func (s *Storage) DeleteAgentUpload(nCtx contextx.IContext, uploadID string) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricOperationDeleteAgentUpload, func(nCtx contextx.IContext) error {
		if err = s.deleteAgentUpload(nCtx, uploadID); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// ===============================================================================
// UploadServer Related Interface
// ===============================================================================

// GetServerUpload gets a upload by upload-id.
func (s *Storage) GetServerUpload(nCtx contextx.IContext, uploadID string) (*types.Upload, error) {
	var (
		data *types.Upload
		err  error
	)

	err = s.WrapFn(nCtx, metricOperationGetServerUpload, func(nCtx contextx.IContext) error {
		data, err = s.getServerUpload(nCtx, uploadID)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return data, nil
}

// CreateServerUpload creates a upload.
func (s *Storage) CreateServerUpload(nCtx contextx.IContext, up *types.Upload) (string, error) {
	var (
		uploadID string
		err      error
	)

	err = s.WrapFn(nCtx, metricOperationCreateServerUpload, func(nCtx contextx.IContext) error {
		uploadID, err = s.createServerUpload(nCtx, up)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	return uploadID, nil
}

// DeleteServerUpload deletes a upload by upload-id.
func (s *Storage) DeleteServerUpload(nCtx contextx.IContext, uploadID string) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricOperationDeleteServerUpload, func(nCtx contextx.IContext) error {
		if err = s.deleteServerUpload(nCtx, uploadID); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// ===============================================================================
// UploadProxy Related Interface
// ===============================================================================

// GetProxyUpload gets a upload by upload-id.
func (s *Storage) GetProxyUpload(nCtx contextx.IContext, uploadID string) (*types.Upload, error) {
	var (
		data *types.Upload
		err  error
	)

	err = s.WrapFn(nCtx, metricOperationGetProxyUpload, func(nCtx contextx.IContext) error {
		data, err = s.getProxyUpload(nCtx, uploadID)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return data, nil
}

// CreateProxyUpload creates a upload.
func (s *Storage) CreateProxyUpload(nCtx contextx.IContext, up *types.Upload) (string, error) {
	var (
		uploadID string
		err      error
	)

	err = s.WrapFn(nCtx, metricOperationCreateProxyUpload, func(nCtx contextx.IContext) error {
		uploadID, err = s.createProxyUpload(nCtx, up)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	return uploadID, nil
}

// DeleteProxyUpload deletes a upload by upload-id.
func (s *Storage) DeleteProxyUpload(nCtx contextx.IContext, uploadID string) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricOperationDeleteProxyUpload, func(nCtx contextx.IContext) error {
		if err = s.deleteProxyUpload(nCtx, uploadID); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// ===============================================================================
// UploadCert Related Interface
// ===============================================================================

// GetCertUpload gets a upload by upload-id.
func (s *Storage) GetCertUpload(nCtx contextx.IContext, uploadID string) (*types.Upload, error) {
	var (
		data *types.Upload
		err  error
	)

	err = s.WrapFn(nCtx, metricOperationGetCertUpload, func(nCtx contextx.IContext) error {
		data, err = s.getCertUpload(nCtx, uploadID)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return data, nil
}

// CreateCertUpload creates a upload.
func (s *Storage) CreateCertUpload(nCtx contextx.IContext, up *types.Upload) (string, error) {
	var (
		uploadID string
		err      error
	)

	err = s.WrapFn(nCtx, metricOperationCreateCertUpload, func(nCtx contextx.IContext) error {
		uploadID, err = s.createCertUpload(nCtx, up)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	return uploadID, nil
}

// DeleteCertUpload deletes a upload by upload-id.
func (s *Storage) DeleteCertUpload(nCtx contextx.IContext, uploadID string) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricOperationDeleteCertUpload, func(nCtx contextx.IContext) error {
		if err = s.deleteCertUpload(nCtx, uploadID); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// ===============================================================================
// UploadBinTool Related Interface
// ===============================================================================

// GetBinToolUpload gets a upload by upload-id.
func (s *Storage) GetBinToolUpload(nCtx contextx.IContext, uploadID string) (*types.Upload, error) {
	var (
		data *types.Upload
		err  error
	)

	err = s.WrapFn(nCtx, metricOperationGetBinToolUpload, func(nCtx contextx.IContext) error {
		data, err = s.getBinToolUpload(nCtx, uploadID)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return data, nil
}

// CreateBinToolUpload creates a upload.
func (s *Storage) CreateBinToolUpload(nCtx contextx.IContext, up *types.Upload) (string, error) {
	var (
		uploadID string
		err      error
	)

	err = s.WrapFn(nCtx, metricOperationCreateBinToolUpload, func(nCtx contextx.IContext) error {
		uploadID, err = s.createBinToolUpload(nCtx, up)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	return uploadID, nil
}

// DeleteBinToolUpload deletes a upload by upload-id.
func (s *Storage) DeleteBinToolUpload(nCtx contextx.IContext, uploadID string) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricOperationDeleteBinToolUpload, func(nCtx contextx.IContext) error {
		if err = s.deleteBinToolUpload(nCtx, uploadID); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// ===============================================================================
// UploadPluginBinTool Related Interface
// ===============================================================================

// GetPluginBinToolUpload gets an upload by upload-id.
func (s *Storage) GetPluginBinToolUpload(nCtx contextx.IContext, uploadID string) (*types.Upload, error) {
	var (
		data *types.Upload
		err  error
	)

	err = s.WrapFn(nCtx, metricOperationGetPluginBinToolUpload, func(nCtx contextx.IContext) error {
		data, err = s.getPluginBinToolUpload(nCtx, uploadID)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return data, nil
}

// CreatePluginBinToolUpload creates an upload.
func (s *Storage) CreatePluginBinToolUpload(nCtx contextx.IContext, up *types.Upload) (string, error) {
	var (
		uploadID string
		err      error
	)

	err = s.WrapFn(nCtx, metricOperationCreatePluginBinToolUpload, func(nCtx contextx.IContext) error {
		uploadID, err = s.createPluginBinToolUpload(nCtx, up)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	return uploadID, nil
}

// DeletePluginBinToolUpload deletes an upload by upload-id.
func (s *Storage) DeletePluginBinToolUpload(nCtx contextx.IContext, uploadID string) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricOperationDeletePluginBinToolUpload, func(nCtx contextx.IContext) error {
		if err = s.deletePluginBinToolUpload(nCtx, uploadID); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// ===============================================================================
// UploadPluginV2 Related Interface
// ===============================================================================

// GetPluginV2Upload gets a upload by upload-id.
func (s *Storage) GetPluginV2Upload(nCtx contextx.IContext, uploadID string) (*types.Upload, error) {
	var (
		data *types.Upload
		err  error
	)

	err = s.WrapFn(nCtx, metricOperationGetPluginV2Upload, func(nCtx contextx.IContext) error {
		data, err = s.getPluginV2Upload(nCtx, uploadID)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return data, nil
}

// CreatePluginV2Upload creates a upload.
func (s *Storage) CreatePluginV2Upload(nCtx contextx.IContext, up *types.Upload) (string, error) {
	var (
		uploadID string
		err      error
	)

	err = s.WrapFn(nCtx, metricOperationCreatePluginV2Upload, func(nCtx contextx.IContext) error {
		uploadID, err = s.createPluginV2Upload(nCtx, up)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	return uploadID, nil
}

// DeletePluginV2Upload deletes a upload by upload-id.
func (s *Storage) DeletePluginV2Upload(nCtx contextx.IContext, uploadID string) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricOperationDeletePluginV2Upload, func(nCtx contextx.IContext) error {
		if err = s.deletePluginV2Upload(nCtx, uploadID); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// ===============================================================================
// UploadExternalPluginV2 Related Interface
// ===============================================================================

// GetExternalPluginV2Upload gets a upload by upload-id.
func (s *Storage) GetExternalPluginV2Upload(nCtx contextx.IContext, uploadID string) (*types.Upload, error) {
	var (
		data *types.Upload
		err  error
	)

	err = s.WrapFn(nCtx, metricOperationGetExternalPluginV2Upload, func(nCtx contextx.IContext) error {
		data, err = s.getExternalPluginV2Upload(nCtx, uploadID)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return data, nil
}

// CreateExternalPluginV2Upload creates a upload.
func (s *Storage) CreateExternalPluginV2Upload(nCtx contextx.IContext, up *types.Upload) (string, error) {
	var (
		uploadID string
		err      error
	)

	err = s.WrapFn(nCtx, metricOperationCreateExternalPluginV2Upload, func(nCtx contextx.IContext) error {
		uploadID, err = s.createExternalPluginV2Upload(nCtx, up)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	return uploadID, nil
}

// DeleteExternalPluginV2Upload deletes a upload by upload-id.
func (s *Storage) DeleteExternalPluginV2Upload(nCtx contextx.IContext, uploadID string) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricOperationDeleteExternalPluginV2Upload, func(nCtx contextx.IContext) error {
		if err = s.deleteExternalPluginV2Upload(nCtx, uploadID); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// ===============================================================================
// UploadPluginV3 Related Interface
// ===============================================================================

// GetPluginV3Upload gets a upload by upload-id.
func (s *Storage) GetPluginV3Upload(nCtx contextx.IContext, uploadID string) (*types.Upload, error) {
	var (
		data *types.Upload
		err  error
	)

	err = s.WrapFn(nCtx, metricOperationGetPluginV3Upload, func(nCtx contextx.IContext) error {
		data, err = s.getPluginV3Upload(nCtx, uploadID)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return data, nil
}

// CreatePluginV3Upload creates a upload.
func (s *Storage) CreatePluginV3Upload(nCtx contextx.IContext, up *types.Upload) (string, error) {
	var (
		uploadID string
		err      error
	)

	err = s.WrapFn(nCtx, metricOperationCreatePluginV3Upload, func(nCtx contextx.IContext) error {
		uploadID, err = s.createPluginV3Upload(nCtx, up)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	return uploadID, nil
}

// DeletePluginV3Upload deletes a upload by upload-id.
func (s *Storage) DeletePluginV3Upload(nCtx contextx.IContext, uploadID string) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricOperationDeletePluginV3Upload, func(nCtx contextx.IContext) error {
		if err = s.deletePluginV3Upload(nCtx, uploadID); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
