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

// Package release provides the release storage interface.
package release

import (
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/release"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	// StorageName defines the storage name.
	StorageName = "release"

	metricOperationUpsertManyReleaseAgent = "upsert_many_release_agent"
	metricOperationExistReleaseAgent      = "exist_release_agent"
	metricOperationGetReleaseAgent        = "get_release_agent"

	metricOperationUpsertManyReleaseProxy = "upsert_many_release_proxy"
	metricOperationExistReleaseProxy      = "exist_release_proxy"
	metricOperationGetReleaseProxy        = "get_release_proxy"

	metricOperationGetReleaseCert    = "get_release_cert"
	metricOperationExistReleaseCert  = "exist_release_cert"
	metricOperationUpsertReleaseCert = "upsert_release_cert"
	metricOperationDeleteReleaseCert = "delete_release_cert"

	metricOperationUpsertReleaseBinTool = "upsert_release_bintool"
	metricOperationDeleteReleaseBinTool = "delete_release_bintool"
	metricOperationGetReleaseBinTool    = "get_release_bintool"
	metricOperationExistReleaseBinTool  = "exist_release_bintool"

	metricOperationUpsertReleasePluginBinTool = "upsert_release_plugin_bintool"
	metricOperationDeleteReleasePluginBinTool = "delete_release_plugin_bintool"
	metricOperationGetReleasePluginBinTool    = "get_release_plugin_bintool"
	metricOperationExistReleasePluginBinTool  = "exist_release_plugin_bintool"

	metricOperationExistReleasePlugin      = "exist_release_plugin"
	metricOperationUpsertManyReleasePlugin = "upsert_many_release_plugin"
	metricOperationGetReleasePlugin        = "get_release_plugin"
)

// NewStorage creates a new release storage.
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

	daoRelease release.IHandler
}

func (s *Storage) initDao() error {
	s.daoRelease = release.New(s.Database)

	return nil
}

func (s *Storage) check() error {
	if s.daoRelease == nil {
		return errors.New("dao release is nil")
	}

	return nil
}

// ===============================================================================
// ReleaseAgent Related Interface
// ===============================================================================

// UpsertManyReleaseAgent upsert many release.
func (s *Storage) UpsertManyReleaseAgent(nCtx contextx.IContext, releaseAgents []*types.ReleaseAgent) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricOperationUpsertManyReleaseAgent, func(nCtx contextx.IContext) error {
		if err = s.upsertManyReleaseAgent(nCtx, releaseAgents); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// ExistReleaseAgent checks if release agent exists.
func (s *Storage) ExistReleaseAgent(nCtx contextx.IContext, gen types.Generation, version string, plats ...platform.Platform) (bool, error) {
	var (
		result bool
		err    error
	)

	err = s.WrapFn(nCtx, metricOperationExistReleaseAgent, func(nCtx contextx.IContext) error {
		result, err = s.existReleaseAgent(nCtx, gen, version, plats...)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return false, err
	}

	return result, nil
}

// GetReleaseAgent gets release by generation, type, platform and version.
func (s *Storage) GetReleaseAgent(nCtx contextx.IContext, gen types.Generation, plat platform.Platform, version string) (*types.ReleaseAgent, error) {
	var (
		data *types.ReleaseAgent
		err  error
	)

	err = s.WrapFn(nCtx, metricOperationGetReleaseAgent, func(nCtx contextx.IContext) error {
		data, err = s.getReleaseAgent(nCtx, gen, plat, version)
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

// ===============================================================================
// ReleaseProxy Related Interface
// ===============================================================================

// GetReleaseProxy gets release by generation, type, platform and version.
func (s *Storage) GetReleaseProxy(nCtx contextx.IContext, gen types.Generation, plat platform.Platform, version string) (*types.ReleaseProxy, error) {
	var (
		data *types.ReleaseProxy
		err  error
	)

	err = s.WrapFn(nCtx, metricOperationGetReleaseProxy, func(nCtx contextx.IContext) error {
		data, err = s.getReleaseProxy(nCtx, gen, plat, version)
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

// UpsertManyReleaseProxy upsert many release.
func (s *Storage) UpsertManyReleaseProxy(nCtx contextx.IContext, releaseProxys []*types.ReleaseProxy) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricOperationUpsertManyReleaseProxy, func(nCtx contextx.IContext) error {
		if err = s.upsertManyReleaseProxy(nCtx, releaseProxys); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// ExistReleaseProxy checks if release proxy exists.
func (s *Storage) ExistReleaseProxy(nCtx contextx.IContext, gen types.Generation, version string, plats ...platform.Platform) (bool, error) {
	var (
		result bool
		err    error
	)

	err = s.WrapFn(nCtx, metricOperationExistReleaseProxy, func(nCtx contextx.IContext) error {
		result, err = s.existReleaseProxy(nCtx, gen, version, plats...)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return false, err
	}

	return result, nil
}

// ===============================================================================
// ReleaseCert Related Interface
// ===============================================================================

// GetReleaseCert gets release cert.
func (s *Storage) GetReleaseCert(nCtx contextx.IContext) (*types.ReleaseCert, error) {
	var (
		data *types.ReleaseCert
		err  error
	)

	err = s.WrapFn(nCtx, metricOperationGetReleaseCert, func(nCtx contextx.IContext) error {
		data, err = s.getReleaseCert(nCtx)
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

// ExistReleaseCert checks if release cert exists.
func (s *Storage) ExistReleaseCert(nCtx contextx.IContext) (bool, error) {
	var (
		result bool
		err    error
	)

	err = s.WrapFn(nCtx, metricOperationExistReleaseCert, func(nCtx contextx.IContext) error {
		result, err = s.existReleaseCert(nCtx)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return false, err
	}

	return result, nil
}

// UpsertReleaseCert upserts release cert.
func (s *Storage) UpsertReleaseCert(nCtx contextx.IContext, cert types.ReleaseCert) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricOperationUpsertReleaseCert, func(nCtx contextx.IContext) error {
		if err = s.upsertReleaseCert(nCtx, cert); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// DeleteReleaseCert deletes release cert.
func (s *Storage) DeleteReleaseCert(nCtx contextx.IContext, fileName string) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricOperationDeleteReleaseCert, func(nCtx contextx.IContext) error {
		if err = s.deleteReleaseCert(nCtx, fileName); err != nil {
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
// ReleaseBinTool Related Interface
// ===============================================================================

// UpsertReleaseBinTool upserts release bintool.
func (s *Storage) UpsertReleaseBinTool(nCtx contextx.IContext, bintool types.ReleaseBinTool) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricOperationUpsertReleaseBinTool, func(nCtx contextx.IContext) error {
		if err = s.upsertReleaseBinTool(nCtx, bintool); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// DeleteReleaseBinTool deletes release bintool.
func (s *Storage) DeleteReleaseBinTool(nCtx contextx.IContext, gen types.Generation, fileName string) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricOperationDeleteReleaseBinTool, func(nCtx contextx.IContext) error {
		if err = s.deleteReleaseBinTool(nCtx, gen, fileName); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// GetReleaseBinTool gets release bintool.
func (s *Storage) GetReleaseBinTool(nCtx contextx.IContext, gen types.Generation) (*types.ReleaseBinTool, error) {
	var (
		data *types.ReleaseBinTool
		err  error
	)

	err = s.WrapFn(nCtx, metricOperationGetReleaseBinTool, func(nCtx contextx.IContext) error {
		data, err = s.getReleaseBinTool(nCtx, gen)
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

// ExistReleaseBinTool checks if release bintool exists.
func (s *Storage) ExistReleaseBinTool(nCtx contextx.IContext, gen types.Generation) (bool, error) {
	var (
		result bool
		err    error
	)

	err = s.WrapFn(nCtx, metricOperationExistReleaseBinTool, func(nCtx contextx.IContext) error {
		result, err = s.existReleaseBinTool(nCtx, gen)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return false, err
	}

	return result, nil
}

// ===============================================================================
// ReleasePluginBinTool Related Interface
// ===============================================================================

// UpsertReleasePluginBinTool upserts release plugin bintool.
func (s *Storage) UpsertReleasePluginBinTool(nCtx contextx.IContext, pluginBinTool types.ReleasePluginBinTool) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricOperationUpsertReleasePluginBinTool, func(nCtx contextx.IContext) error {
		if err = s.upsertReleasePluginBinTool(nCtx, pluginBinTool); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// DeleteReleasePluginBinTool deletes release plugin bintool.
func (s *Storage) DeleteReleasePluginBinTool(nCtx contextx.IContext, gen types.Generation, fileName string) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricOperationDeleteReleasePluginBinTool, func(nCtx contextx.IContext) error {
		if err = s.deleteReleasePluginBinTool(nCtx, gen, fileName); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// GetReleasePluginBinTool gets release plugin bintool.
func (s *Storage) GetReleasePluginBinTool(nCtx contextx.IContext, gen types.Generation, name string) (*types.ReleasePluginBinTool, error) {
	var (
		data *types.ReleasePluginBinTool
		err  error
	)

	err = s.WrapFn(nCtx, metricOperationGetReleasePluginBinTool, func(nCtx contextx.IContext) error {
		data, err = s.getReleasePluginBinTool(nCtx, gen, name)
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

// ExistReleasePluginBinTool checks if release plugin bintool exists.
func (s *Storage) ExistReleasePluginBinTool(nCtx contextx.IContext, gen types.Generation) (bool, error) {
	var (
		result bool
		err    error
	)

	err = s.WrapFn(nCtx, metricOperationExistReleasePluginBinTool, func(nCtx contextx.IContext) error {
		result, err = s.existReleasePluginBinTool(nCtx, gen)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return false, err
	}

	return result, nil
}

// ===============================================================================
// ReleasePlugin Related Interface
// ===============================================================================

// ExistReleasePlugin checks if release plugin exists.
func (s *Storage) ExistReleasePlugin(nCtx contextx.IContext, pluginName string, version string, plats ...platform.Platform) (bool, error) {
	var (
		result bool
		err    error
	)

	err = s.WrapFn(nCtx, metricOperationExistReleasePlugin, func(nCtx contextx.IContext) error {
		result, err = s.existReleasePlugin(nCtx, pluginName, version, plats...)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return false, err
	}

	return result, nil
}

// UpsertManyReleasePlugin upsert many release.
func (s *Storage) UpsertManyReleasePlugin(nCtx contextx.IContext, releasePlugins []*types.ReleasePlugin) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricOperationUpsertManyReleasePlugin, func(nCtx contextx.IContext) error {
		if err = s.upsertManyReleasePlugin(nCtx, releasePlugins); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// GetReleasePlugin gets release plugin.
func (s *Storage) GetReleasePlugin(nCtx contextx.IContext, pluginName string, gen types.Generation, plat platform.Platform, version string) (
	*types.ReleasePlugin, error) {

	var (
		data *types.ReleasePlugin
		err  error
	)

	err = s.WrapFn(nCtx, metricOperationGetReleasePlugin, func(nCtx contextx.IContext) error {
		data, err = s.getReleasePlugin(nCtx, pluginName, gen, plat, version)
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
