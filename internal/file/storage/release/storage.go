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
	"fmt"

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

	metricOperationUpsertManyReleaseAgent      = "upsert_many_release_agent"
	metricOperationExistReleaseAgent           = "exist_release_agent"
	metricOperationGetReleaseAgent             = "get_release_agent"
	metricOperationEnableReleaseAgent          = "enable_release_agent"
	metricOperationDisableReleaseAgent         = "disable_release_agent"
	metricOperationSetAsDefaultReleaseAgent    = "set_as_default_release_agent"
	metricOperationCancelAsDefaultReleaseAgent = "cancel_as_default_release_agent"
	metricOperationDeleteReleaseAgent          = "delete_release_agent"
	metricOperationSetReleaseAgentLabelsMany   = "set_release_agent_labels_many"
	metricOperationListReleaseAgent            = "list_release_agent"
	metricOperationCountReleaseAgent           = "count_release_agent"
	metricOperationDistinctReleaseAgent        = "distinct_release_agent"

	metricOperationUpsertManyReleaseProxy      = "upsert_many_release_proxy"
	metricOperationExistReleaseProxy           = "exist_release_proxy"
	metricOperationGetReleaseProxy             = "get_release_proxy"
	metricOperationEnableReleaseProxy          = "enable_release_proxy"
	metricOperationDisableReleaseProxy         = "disable_release_proxy"
	metricOperationSetAsDefaultReleaseProxy    = "set_as_default_release_proxy"
	metricOperationCancelAsDefaultReleaseProxy = "cancel_as_default_release_proxy"
	metricOperationDeleteReleaseProxy          = "delete_release_proxy"
	metricOperationSetReleaseProxyLabelsMany   = "set_release_proxy_labels_many"
	metricOperationListReleaseProxy            = "list_release_proxy"
	metricOperationCountReleaseProxy           = "count_release_proxy"
	metricOperationDistinctReleaseProxy        = "distinct_release_proxy"

	metricOperationGetReleaseCert    = "get_release_cert"
	metricOperationExistReleaseCert  = "exist_release_cert"
	metricOperationUpsertReleaseCert = "upsert_release_cert"
	metricOperationDeleteReleaseCert = "delete_release_cert"
	metricOperationListReleaseCert   = "list_release_cert"

	metricOperationUpsertReleaseBinTool = "upsert_release_bintool"
	metricOperationDeleteReleaseBinTool = "delete_release_bintool"
	metricOperationGetReleaseBinTool    = "get_release_bintool"
	metricOperationExistReleaseBinTool  = "exist_release_bintool"
	metricOperationListReleaseBinTool   = "list_release_bintool"

	metricOperationUpsertReleasePluginBinTool       = "upsert_release_plugin_bintool"
	metricOperationDeleteReleasePluginBinTool       = "delete_release_plugin_bintool"
	metricOperationGetReleasePluginBinTool          = "get_release_plugin_bintool"
	metricOperationExistReleasePluginBinTool        = "exist_release_plugin_bintool"
	metricOperationListReleasePluginBinTool         = "list_release_plugin_bintool"
	metricOperationDistinctNameReleasePluginBinTool = "distinct_name_release_plugin_bintool"

	metricOperationExistReleasePlugin             = "exist_release_plugin"
	metricOperationUpsertManyReleasePlugin        = "upsert_many_release_plugin"
	metricOperationGetReleasePlugin               = "get_release_plugin"
	metricOperationListReleasePlugin              = "list_release_plugin"
	metricOperationEnableReleasePlugin            = "enable_release_plugin"
	metricOperationDisableReleasePlugin           = "disable_release_plugin"
	metricOperationSetAsDefaultReleasePlugin      = "set_as_default_release_plugin"
	metricOperationCancelAsDefaultReleasePlugin   = "cancel_as_default_release_plugin"
	metricOperationDeleteReleasePlugin            = "delete_release_plugin"
	metricOperationSetHiddenReleasePlugin         = "set_hidden_release_plugin"
	metricOperationCancelHiddenReleasePlugin      = "cancel_hidden_release_plugin"
	metricOperationCountReleasePlugin             = "count_release_plugin"
	metricOperationDistinctReleasePlugin          = "distinct_release_plugin"
	metricOperationGetReleasePluginDefaultVersion = "get_release_plugin_default_version"
	metricOperationDistinctNameReleasePlugin      = "distinct_name_release_plugin"
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

// EnableReleaseAgent enables an agent release.
func (s *Storage) EnableReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error {
	return s.WrapFn(nCtx, metricOperationEnableReleaseAgent, func(nCtx contextx.IContext) error {
		return s.enableRelease(nCtx, types.ReleaseTypeAgent, key.Generation, key.Platform, key.Version, types.ReleaseNameAgent)
	})
}

// DisableReleaseAgent disables an agent release.
func (s *Storage) DisableReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error {
	return s.WrapFn(nCtx, metricOperationDisableReleaseAgent, func(nCtx contextx.IContext) error {
		return s.disableRelease(nCtx, types.ReleaseTypeAgent, key.Generation, key.Platform, key.Version, types.ReleaseNameAgent)
	})
}

// SetAsDefaultReleaseAgent sets an agent release as default.
func (s *Storage) SetAsDefaultReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error {
	return s.WrapFn(nCtx, metricOperationSetAsDefaultReleaseAgent, func(nCtx contextx.IContext) error {
		return s.setAsDefaultRelease(nCtx, types.ReleaseTypeAgent, key.Generation, key.Platform, key.Version, types.ReleaseNameAgent)
	})
}

// CancelAsDefaultReleaseAgent cancels an agent release default.
func (s *Storage) CancelAsDefaultReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error {
	return s.WrapFn(nCtx, metricOperationCancelAsDefaultReleaseAgent, func(nCtx contextx.IContext) error {
		return s.cancelAsDefaultRelease(nCtx, types.ReleaseTypeAgent, key.Generation, key.Platform, key.Version, types.ReleaseNameAgent)
	})
}

// DeleteReleaseAgent deletes release metadata identified by its key.
func (s *Storage) DeleteReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error {
	return s.WrapFn(nCtx, metricOperationDeleteReleaseAgent, func(nCtx contextx.IContext) error {
		return s.deleteRelease(nCtx, types.ReleaseTypeAgent, key.Generation, key.Platform, key.Version, types.ReleaseNameAgent)
	})
}

// SetReleaseAgentLabelsMany updates labels for matching releases.
func (s *Storage) SetReleaseAgentLabelsMany(nCtx contextx.IContext, labels []string, conditions ...*types.ReleaseCondition) error {
	return s.WrapFn(nCtx, metricOperationSetReleaseAgentLabelsMany, func(nCtx contextx.IContext) error {
		return s.setReleaseLabelsMany(nCtx, types.ReleaseTypeAgent, labels, conditions...)
	})
}

// ListReleaseAgent lists agent releases by page and conditions.
func (s *Storage) ListReleaseAgent(nCtx contextx.IContext, page types.Page,
	conditions ...*types.ReleaseCondition) ([]*types.ReleaseAgent, int64, error) {

	var (
		results []*types.ReleaseAgent
		num     int64
	)

	err := s.WrapFn(nCtx, metricOperationListReleaseAgent, func(nCtx contextx.IContext) error {
		var err error
		if results, num, err = s.listReleaseAgent(nCtx, page, conditions...); err != nil {
			logger.G.Sys().WithErr(err).Error("failed to list release agent")

			return fmt.Errorf("failed to list release agent: %w", err)
		}

		return nil
	})

	return results, num, err
}

// CountReleaseAgent counts agent releases by conditions.
func (s *Storage) CountReleaseAgent(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) (int64, error) {
	var num int64
	err := s.WrapFn(nCtx, metricOperationCountReleaseAgent, func(nCtx contextx.IContext) error {
		var err error
		if num, err = s.countRelease(nCtx, types.ReleaseTypeAgent, conditions...); err != nil {
			logger.G.Sys().WithErr(err).Error("failed to count release agent")

			return fmt.Errorf("failed to count release agent: %w", err)
		}

		return nil
	})

	return num, err
}

// DistinctReleaseAgent gets agent releases distinct.
func (s *Storage) DistinctReleaseAgent(nCtx contextx.IContext, fields types.ReleaseDistinctField,
	conditions ...*types.ReleaseCondition) (*types.ReleaseDistinctResult, error) {

	var (
		data *types.ReleaseDistinctResult
	)

	err := s.WrapFn(nCtx, metricOperationDistinctReleaseAgent, func(nCtx contextx.IContext) error {
		var err error
		if data, err = s.distinctRelease(nCtx, types.ReleaseTypeAgent, fields, conditions...); err != nil {
			logger.G.Sys().WithErr(err).Error("failed to distinct release agent")

			return fmt.Errorf("failed to distinct release agent: %w", err)
		}

		return nil
	})

	return data, err
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

// EnableReleaseProxy enables a proxy release.
func (s *Storage) EnableReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error {
	return s.WrapFn(nCtx, metricOperationEnableReleaseProxy, func(nCtx contextx.IContext) error {
		return s.enableRelease(nCtx, types.ReleaseTypeProxy, key.Generation, key.Platform, key.Version, types.ReleaseNameProxy)
	})
}

// DisableReleaseProxy disables a proxy release.
func (s *Storage) DisableReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error {
	return s.WrapFn(nCtx, metricOperationDisableReleaseProxy, func(nCtx contextx.IContext) error {
		return s.disableRelease(nCtx, types.ReleaseTypeProxy, key.Generation, key.Platform, key.Version, types.ReleaseNameProxy)
	})
}

// SetAsDefaultReleaseProxy sets a proxy release as default.
func (s *Storage) SetAsDefaultReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error {
	return s.WrapFn(nCtx, metricOperationSetAsDefaultReleaseProxy, func(nCtx contextx.IContext) error {
		return s.setAsDefaultRelease(nCtx, types.ReleaseTypeProxy, key.Generation, key.Platform, key.Version, types.ReleaseNameProxy)
	})
}

// CancelAsDefaultReleaseProxy cancels a proxy release default.
func (s *Storage) CancelAsDefaultReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error {
	return s.WrapFn(nCtx, metricOperationCancelAsDefaultReleaseProxy, func(nCtx contextx.IContext) error {
		return s.cancelAsDefaultRelease(nCtx, types.ReleaseTypeProxy, key.Generation, key.Platform, key.Version, types.ReleaseNameProxy)
	})
}

// DeleteReleaseProxy deletes release metadata identified by its key.
func (s *Storage) DeleteReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error {
	return s.WrapFn(nCtx, metricOperationDeleteReleaseProxy, func(nCtx contextx.IContext) error {
		return s.deleteRelease(nCtx, types.ReleaseTypeProxy, key.Generation, key.Platform, key.Version, types.ReleaseNameProxy)
	})
}

// SetReleaseProxyLabelsMany updates labels for matching releases.
func (s *Storage) SetReleaseProxyLabelsMany(nCtx contextx.IContext, labels []string, conditions ...*types.ReleaseCondition) error {
	return s.WrapFn(nCtx, metricOperationSetReleaseProxyLabelsMany, func(nCtx contextx.IContext) error {
		return s.setReleaseLabelsMany(nCtx, types.ReleaseTypeProxy, labels, conditions...)
	})
}

// ListReleaseProxy lists proxy releases by page and conditions.
func (s *Storage) ListReleaseProxy(nCtx contextx.IContext, page types.Page,
	conditions ...*types.ReleaseCondition) ([]*types.ReleaseProxy, int64, error) {

	var (
		results []*types.ReleaseProxy
		num     int64
	)

	err := s.WrapFn(nCtx, metricOperationListReleaseProxy, func(nCtx contextx.IContext) error {
		var err error
		if results, num, err = s.listReleaseProxy(nCtx, page, conditions...); err != nil {
			logger.G.Sys().WithErr(err).Error("failed to list release proxy")

			return fmt.Errorf("failed to list release proxy: %w", err)
		}

		return nil
	})

	return results, num, err
}

// CountReleaseProxy counts proxy releases by conditions.
func (s *Storage) CountReleaseProxy(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) (int64, error) {
	var num int64
	err := s.WrapFn(nCtx, metricOperationCountReleaseProxy, func(nCtx contextx.IContext) error {
		var err error
		if num, err = s.countRelease(nCtx, types.ReleaseTypeProxy, conditions...); err != nil {
			logger.G.Sys().WithErr(err).Error("failed to count release proxy")

			return fmt.Errorf("failed to count release proxy: %w", err)
		}

		return nil
	})

	return num, err
}

// DistinctReleaseProxy gets proxy releases distinct.
func (s *Storage) DistinctReleaseProxy(nCtx contextx.IContext, fields types.ReleaseDistinctField,
	conditions ...*types.ReleaseCondition) (*types.ReleaseDistinctResult, error) {

	var (
		data *types.ReleaseDistinctResult
	)

	err := s.WrapFn(nCtx, metricOperationDistinctReleaseProxy, func(nCtx contextx.IContext) error {
		var err error
		if data, err = s.distinctRelease(nCtx, types.ReleaseTypeProxy, fields, conditions...); err != nil {
			logger.G.Sys().WithErr(err).Error("failed to distinct release proxy")

			return fmt.Errorf("failed to distinct release proxy: %w", err)
		}

		return nil
	})

	return data, err
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

// DeleteReleaseCert deletes release metadata identified by its key.
func (s *Storage) DeleteReleaseCert(nCtx contextx.IContext, key types.ReleaseCertKey) error {
	return s.WrapFn(nCtx, metricOperationDeleteReleaseCert, func(nCtx contextx.IContext) error {
		return s.deleteRelease(nCtx, types.ReleaseTypeCert, key.Generation, platform.UnknownPlatform(), types.ReleaseVersionCert, types.ReleaseNameCert)
	})
}

// ListReleaseCert lists cert releases by page and conditions.
func (s *Storage) ListReleaseCert(nCtx contextx.IContext, page types.Page,
	conditions ...*types.ReleaseCondition) ([]*types.ReleaseCert, int64, error) {

	var (
		results []*types.ReleaseCert
		num     int64
	)

	err := s.WrapFn(nCtx, metricOperationListReleaseCert, func(nCtx contextx.IContext) error {
		var err error
		results, num, err = s.listReleaseCert(nCtx, page, conditions...)

		return err
	})

	return results, num, err
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

// DeleteReleaseBinTool deletes release metadata identified by its key.
func (s *Storage) DeleteReleaseBinTool(nCtx contextx.IContext, key types.ReleaseBinToolKey) error {
	return s.WrapFn(nCtx, metricOperationDeleteReleaseBinTool, func(nCtx contextx.IContext) error {
		return s.deleteRelease(
			nCtx, types.ReleaseTypeBinTool, key.Generation, platform.UnknownPlatform(), types.ReleaseVersionBinTool, types.ReleaseNameBinTool)
	})
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

// ListReleaseBinTool lists bintool releases by page and conditions.
func (s *Storage) ListReleaseBinTool(nCtx contextx.IContext, page types.Page,
	conditions ...*types.ReleaseCondition) ([]*types.ReleaseBinTool, int64, error) {

	var (
		results []*types.ReleaseBinTool
		num     int64
	)

	err := s.WrapFn(nCtx, metricOperationListReleaseBinTool, func(nCtx contextx.IContext) error {
		var err error
		results, num, err = s.listReleaseBinTool(nCtx, page, conditions...)

		return err
	})

	return results, num, err
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

// DeleteReleasePluginBinTool deletes release metadata identified by its key.
func (s *Storage) DeleteReleasePluginBinTool(nCtx contextx.IContext, key types.ReleasePluginBinToolKey) error {
	return s.WrapFn(nCtx, metricOperationDeleteReleasePluginBinTool, func(nCtx contextx.IContext) error {
		return s.deleteRelease(
			nCtx, types.ReleaseTypePluginBinTool, key.Generation, platform.UnknownPlatform(), types.ReleaseVersionPluginBinTool, key.Name)
	})
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

// ListReleasePluginBinTool lists plugin bintool release by page and conditions.
func (s *Storage) ListReleasePluginBinTool(nCtx contextx.IContext, page types.Page,
	conditions ...*types.ReleaseCondition) ([]*types.ReleasePluginBinTool, int64, error) {

	var (
		results []*types.ReleasePluginBinTool
		num     int64
	)

	err := s.WrapFn(nCtx, metricOperationListReleasePluginBinTool, func(nCtx contextx.IContext) error {
		var err error
		results, num, err = s.listReleasePluginBinTool(nCtx, page, conditions...)

		return err
	})

	return results, num, err
}

// ===============================================================================
// ReleasePlugin Related Interface
// ===============================================================================

// ListReleasePlugin lists plugin releases by page and conditions.
func (s *Storage) ListReleasePlugin(nCtx contextx.IContext, page types.Page, conditions ...*types.ReleaseCondition) (
	[]*types.ReleasePlugin, int64, error) {

	var (
		releases []*types.ReleasePlugin
		total    int64
		err      error
	)
	err = s.WrapFn(nCtx, metricOperationListReleasePlugin, func(nCtx contextx.IContext) error {
		var err error
		releases, total, err = s.listReleasePlugin(nCtx, page, conditions...)

		return err
	})

	return releases, total, err
}

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

// EnableReleasePlugin enables a plugin release.
func (s *Storage) EnableReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error {
	return s.WrapFn(nCtx, metricOperationEnableReleasePlugin, func(nCtx contextx.IContext) error {
		return s.enableRelease(nCtx, types.ReleaseTypePlugin, key.Generation, key.Platform, key.Version, key.Name)
	})
}

// DisableReleasePlugin disables a plugin release.
func (s *Storage) DisableReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error {
	return s.WrapFn(nCtx, metricOperationDisableReleasePlugin, func(nCtx contextx.IContext) error {
		return s.disableRelease(nCtx, types.ReleaseTypePlugin, key.Generation, key.Platform, key.Version, key.Name)
	})
}

// SetAsDefaultReleasePlugin sets a plugin release as default.
func (s *Storage) SetAsDefaultReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error {
	return s.WrapFn(nCtx, metricOperationSetAsDefaultReleasePlugin, func(nCtx contextx.IContext) error {
		return s.setAsDefaultRelease(nCtx, types.ReleaseTypePlugin, key.Generation, key.Platform, key.Version, key.Name)
	})
}

// CancelAsDefaultReleasePlugin cancels a plugin release default.
func (s *Storage) CancelAsDefaultReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error {
	return s.WrapFn(nCtx, metricOperationCancelAsDefaultReleasePlugin, func(nCtx contextx.IContext) error {
		return s.cancelAsDefaultRelease(nCtx, types.ReleaseTypePlugin, key.Generation, key.Platform, key.Version, key.Name)
	})
}

// DeleteReleasePlugin deletes release metadata identified by its key.
func (s *Storage) DeleteReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error {
	return s.WrapFn(nCtx, metricOperationDeleteReleasePlugin, func(nCtx contextx.IContext) error {
		return s.deleteRelease(nCtx, types.ReleaseTypePlugin, key.Generation, key.Platform, key.Version, key.Name)
	})
}

// SetHiddenReleasePlugin hides a plugin release.
func (s *Storage) SetHiddenReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error {
	return s.WrapFn(nCtx, metricOperationSetHiddenReleasePlugin, func(nCtx contextx.IContext) error {
		return s.setHiddenReleasePlugin(nCtx, key, true)
	})
}

// CancelHiddenReleasePlugin unhides a plugin release.
func (s *Storage) CancelHiddenReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error {
	return s.WrapFn(nCtx, metricOperationCancelHiddenReleasePlugin, func(nCtx contextx.IContext) error {
		return s.setHiddenReleasePlugin(nCtx, key, false)
	})
}

// CountReleasePlugin counts plugin release by conditions.
func (s *Storage) CountReleasePlugin(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) (int64, error) {
	var num int64
	err := s.WrapFn(nCtx, metricOperationCountReleasePlugin, func(nCtx contextx.IContext) error {
		var err error
		num, err = s.countRelease(nCtx, types.ReleaseTypePlugin, conditions...)

		return err
	})

	return num, err
}

// DistinctReleasePlugin gets plugin releases distinct.
func (s *Storage) DistinctReleasePlugin(nCtx contextx.IContext, fields types.ReleaseDistinctField, conditions ...*types.ReleaseCondition) (
	*types.ReleaseDistinctResult, error) {

	var (
		result *types.ReleaseDistinctResult
	)

	err := s.WrapFn(nCtx, metricOperationDistinctReleasePlugin, func(nCtx contextx.IContext) error {
		var err error
		if result, err = s.distinctRelease(nCtx, types.ReleaseTypePlugin, fields, conditions...); err != nil {
			logger.G.Sys().WithErr(err).Error("failed to distinct release plugin")

			return fmt.Errorf("failed to distinct release plugin: %w", err)
		}

		return nil
	})

	return result, err
}

// GetReleasePluginDefaultVersion gets release plugin default version by name, generation and platform.
func (s *Storage) GetReleasePluginDefaultVersion(nCtx contextx.IContext, name string, gen types.Generation, plat platform.Platform) (string, error) {
	var (
		version string
		err     error
	)

	err = s.WrapFn(nCtx, metricOperationGetReleasePluginDefaultVersion, func(nCtx contextx.IContext) error {
		version, err = s.getReleaseDefaultVersion(nCtx, types.ReleaseTypePlugin, gen, plat, name)
		if err != nil {
			return err
		}

		return nil
	})

	return version, err
}

// DistinctNameReleasePlugin gets distinct plugin release names.
func (s *Storage) DistinctNameReleasePlugin(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) ([]string, error) {
	var names []string
	err := s.WrapFn(nCtx, metricOperationDistinctNameReleasePlugin, func(nCtx contextx.IContext) error {
		var err error
		opts, err := convertReleaseConditionsToOptions(conditions...)
		if err != nil {
			return fmt.Errorf("failed to convert release conditions to options: %w", err)
		}
		if names, err = s.daoRelease.DistinctName(nCtx, types.ReleaseTypePlugin, opts...); err != nil {
			logger.G.Sys().WithErr(err).Error("failed to distinct name release plugin")

			return fmt.Errorf("failed to distinct name release plugin: %w", err)
		}

		return nil
	})

	return names, err
}

// DistinctNameReleasePluginBinTool gets distinct plugin bintool release names.
func (s *Storage) DistinctNameReleasePluginBinTool(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) ([]string, error) {
	var names []string
	err := s.WrapFn(nCtx, metricOperationDistinctNameReleasePluginBinTool, func(nCtx contextx.IContext) error {
		var err error
		opts, err := convertReleaseConditionsToOptions(conditions...)
		if err != nil {
			return fmt.Errorf("failed to convert release conditions to options: %w", err)
		}
		if names, err = s.daoRelease.DistinctName(nCtx, types.ReleaseTypePluginBinTool, opts...); err != nil {
			logger.G.Sys().WithErr(err).Error("failed to distinct name release plugin bintool")

			return fmt.Errorf("failed to distinct name release plugin bintool: %w", err)
		}

		return nil
	})

	return names, err
}
