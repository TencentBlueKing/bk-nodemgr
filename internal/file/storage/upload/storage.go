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

	metricOperationGetAgentUpload         = "get_agent_upload"
	metricOperationCreateAgentUpload      = "create_agent_upload"
	metricOperationDeleteAgentUpload      = "delete_agent_upload"
	metricOperationDistinctAgentSavedName = "distinct_agent_saved_name"

	metricOperationGetServerUpload         = "get_server_upload"
	metricOperationCreateServerUpload      = "create_server_upload"
	metricOperationDeleteServerUpload      = "delete_server_upload"
	metricOperationDistinctServerSavedName = "distinct_server_saved_name"

	metricOperationGetProxyUpload         = "get_proxy_upload"
	metricOperationCreateProxyUpload      = "create_proxy_upload"
	metricOperationDeleteProxyUpload      = "delete_proxy_upload"
	metricOperationDistinctProxySavedName = "distinct_proxy_saved_name"

	metricOperationGetCertUpload         = "get_cert_upload"
	metricOperationCreateCertUpload      = "create_cert_upload"
	metricOperationDeleteCertUpload      = "delete_cert_upload"
	metricOperationDistinctCertSavedName = "distinct_cert_saved_name"

	metricOperationGetBinToolUpload         = "get_bintool_upload"
	metricOperationCreateBinToolUpload      = "create_bintool_upload"
	metricOperationDeleteBinToolUpload      = "delete_bintool_upload"
	metricOperationDistinctBinToolSavedName = "distinct_bintool_saved_name"

	metricOperationGetPluginBinToolUpload         = "get_plugin_bintool_upload"
	metricOperationCreatePluginBinToolUpload      = "create_plugin_bintool_upload"
	metricOperationDeletePluginBinToolUpload      = "delete_plugin_bintool_upload"
	metricOperationDistinctPluginBinToolSavedName = "distinct_plugin_bintool_saved_name"

	metricOperationGetPluginV2Upload         = "get_plugin_v2_upload"
	metricOperationCreatePluginV2Upload      = "create_plugin_v2_upload"
	metricOperationDeletePluginV2Upload      = "delete_plugin_v2_upload"
	metricOperationDistinctPluginV2SavedName = "distinct_plugin_v2_saved_name"

	metricOperationGetExternalPluginV2Upload         = "get_external_plugin_v2_upload"
	metricOperationCreateExternalPluginV2Upload      = "create_external_plugin_v2_upload"
	metricOperationDeleteExternalPluginV2Upload      = "delete_external_plugin_v2_upload"
	metricOperationDistinctExternalPluginV2SavedName = "distinct_external_plugin_v2_saved_name"

	metricOperationGetPluginV3Upload         = "get_plugin_v3_upload"
	metricOperationCreatePluginV3Upload      = "create_plugin_v3_upload"
	metricOperationDeletePluginV3Upload      = "delete_plugin_v3_upload"
	metricOperationDistinctPluginV3SavedName = "distinct_plugin_v3_saved_name"
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

// DistinctAgentSavedName distincts saved names of agent uploads.
func (s *Storage) DistinctAgentSavedName(nCtx contextx.IContext, names []string) ([]string, error) {
	var (
		bound []string
		err   error
	)

	err = s.WrapFn(nCtx, metricOperationDistinctAgentSavedName, func(nCtx contextx.IContext) error {
		bound, err = s.distinctAgentSavedName(nCtx, names)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return bound, nil
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

// DistinctServerSavedName distincts saved names of server uploads.
func (s *Storage) DistinctServerSavedName(nCtx contextx.IContext, names []string) ([]string, error) {
	var (
		bound []string
		err   error
	)

	err = s.WrapFn(nCtx, metricOperationDistinctServerSavedName, func(nCtx contextx.IContext) error {
		bound, err = s.distinctServerSavedName(nCtx, names)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return bound, nil
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

// DistinctProxySavedName distincts saved names of proxy uploads.
func (s *Storage) DistinctProxySavedName(nCtx contextx.IContext, names []string) ([]string, error) {
	var (
		bound []string
		err   error
	)

	err = s.WrapFn(nCtx, metricOperationDistinctProxySavedName, func(nCtx contextx.IContext) error {
		bound, err = s.distinctProxySavedName(nCtx, names)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return bound, nil
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

// DistinctCertSavedName distincts saved names of cert uploads.
func (s *Storage) DistinctCertSavedName(nCtx contextx.IContext, names []string) ([]string, error) {
	var (
		bound []string
		err   error
	)

	err = s.WrapFn(nCtx, metricOperationDistinctCertSavedName, func(nCtx contextx.IContext) error {
		bound, err = s.distinctCertSavedName(nCtx, names)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return bound, nil
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

// DistinctBinToolSavedName distincts saved names of bintool uploads.
func (s *Storage) DistinctBinToolSavedName(nCtx contextx.IContext, names []string) ([]string, error) {
	var (
		bound []string
		err   error
	)

	err = s.WrapFn(nCtx, metricOperationDistinctBinToolSavedName, func(nCtx contextx.IContext) error {
		bound, err = s.distinctBinToolSavedName(nCtx, names)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return bound, nil
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

// DistinctPluginBinToolSavedName distincts saved names of plugin bintool uploads.
func (s *Storage) DistinctPluginBinToolSavedName(nCtx contextx.IContext, names []string) ([]string, error) {
	var (
		bound []string
		err   error
	)

	err = s.WrapFn(nCtx, metricOperationDistinctPluginBinToolSavedName, func(nCtx contextx.IContext) error {
		bound, err = s.distinctPluginBinToolSavedName(nCtx, names)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return bound, nil
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

// DistinctPluginV2SavedName distincts saved names of plugin v2 uploads.
func (s *Storage) DistinctPluginV2SavedName(nCtx contextx.IContext, names []string) ([]string, error) {
	var (
		bound []string
		err   error
	)

	err = s.WrapFn(nCtx, metricOperationDistinctPluginV2SavedName, func(nCtx contextx.IContext) error {
		bound, err = s.distinctPluginV2SavedName(nCtx, names)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return bound, nil
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

// DistinctExternalPluginV2SavedName distincts saved names of external plugin v2 uploads.
func (s *Storage) DistinctExternalPluginV2SavedName(nCtx contextx.IContext, names []string) ([]string, error) {
	var (
		bound []string
		err   error
	)

	err = s.WrapFn(nCtx, metricOperationDistinctExternalPluginV2SavedName, func(nCtx contextx.IContext) error {
		bound, err = s.distinctExternalPluginV2SavedName(nCtx, names)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return bound, nil
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

// DistinctPluginV3SavedName distincts saved names of plugin v3 uploads.
func (s *Storage) DistinctPluginV3SavedName(nCtx contextx.IContext, names []string) ([]string, error) {
	var (
		bound []string
		err   error
	)

	err = s.WrapFn(nCtx, metricOperationDistinctPluginV3SavedName, func(nCtx contextx.IContext) error {
		bound, err = s.distinctPluginV3SavedName(nCtx, names)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return bound, nil
}
