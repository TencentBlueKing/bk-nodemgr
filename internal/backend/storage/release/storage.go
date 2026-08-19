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
	daoPackageEvent "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/packageevent"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/release"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	// StorageName defines the storage name.
	StorageName = "release"

	metricOperationListReleaseAgent                 = "list_release_agent"
	metricOperationCountReleaseAgent                = "count_release_agent"
	metricOperationGetReleaseAgent                  = "get_release_agent"
	metricOperationDistinctReleaseAgent             = "distinct_release_agent"
	metricOperationDeleteReleaseAgent               = "delete_release_agent"
	metricOperationSetReleaseAgentLabelsMany        = "set_release_agent_labels_many"
	metricOperationEnableReleaseAgent               = "enable_release_agent"
	metricOperationDisableReleaseAgent              = "disable_release_agent"
	metricOperationSetAsDefaultReleaseAgent         = "set_as_default_release_agent"
	metricOperationCancelAsDefaultReleaseAgent      = "cancel_as_default_release_agent"
	metricOperationListReleaseProxy                 = "list_release_proxy"
	metricOperationCountReleaseProxy                = "count_release_proxy"
	metricOperationGetReleaseProxy                  = "get_release_proxy"
	metricOperationDistinctReleaseProxy             = "distinct_release_proxy"
	metricOperationDeleteReleaseProxy               = "delete_release_proxy"
	metricOperationSetReleaseProxyLabelsMany        = "set_release_proxy_labels_many"
	metricOperationEnableReleaseProxy               = "enable_release_proxy"
	metricOperationDisableReleaseProxy              = "disable_release_proxy"
	metricOperationSetAsDefaultReleaseProxy         = "set_as_default_release_proxy"
	metricOperationCancelAsDefaultReleaseProxy      = "cancel_as_default_release_proxy"
	metricOperationListReleasePlugin                = "list_release_plugin"
	metricOperationCountReleasePlugin               = "count_release_plugin"
	metricOperationDistinctReleasePlugin            = "distinct_release_plugin"
	metricOperationGetReleasePlugin                 = "get_release_plugin"
	metricOperationDeleteReleasePlugin              = "delete_release_plugin"
	metricOperationEnableReleasePlugin              = "enable_release_plugin"
	metricOperationDisableReleasePlugin             = "disable_release_plugin"
	metricOperationSetHiddenReleasePlugin           = "set_hidden_release_plugin"
	metricOperationSetAsDefaultReleasePlugin        = "set_as_default_release_plugin"
	metricOperationCancelAsDefaultReleasePlugin     = "cancel_as_default_release_plugin"
	metricOperationExistReleasePlugin               = "exist_release_plugin"
	metricOperationGetReleasePluginDefaultVersion   = "get_release_plugin_default_version"
	metricOperationListReleaseCert                  = "list_release_cert"
	metricOperationCountReleaseCert                 = "count_release_cert"
	metricOperationGetReleaseCert                   = "get_release_cert"
	metricOperationDeleteReleaseCert                = "delete_release_cert"
	metricOperationListReleaseBinTool               = "list_release_bintool"
	metricOperationCountReleaseBinTool              = "count_release_bintool"
	metricOperationGetReleaseBinTool                = "get_release_bintool"
	metricOperationDeleteReleaseBinTool             = "delete_release_bintool"
	metricOperationListReleasePluginBinTool         = "list_release_plugin_bintool"
	metricOperationCountReleasePluginBinTool        = "count_release_plugin_bintool"
	metricOperationGetReleasePluginBinTool          = "get_release_plugin_bintool"
	metricOperationDeleteReleasePluginBinTool       = "delete_release_plugin_bintool"
	metricOperationCountPackageEvent                = "count_package_event"
	metricOperationListPackageEvent                 = "list_package_event"
	metricOperationCreateManyPackageEvent           = "create_many_package_event"
	metricOperationDistinctPackageEvent             = "distinct_package_event"
	metricOperationDistinctNameReleasePlugin        = "distinct_name_release_plugin"
	metricOperationDistinctNameReleasePluginBinTool = "distinct_name_release_plugin_bintool"
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

	daoRelease      release.IHandler
	daoPackageEvent daoPackageEvent.IHandler
}

func (s *Storage) initDao() error {
	s.daoRelease = release.New(s.Database)
	s.daoPackageEvent = daoPackageEvent.New(s.Database)

	return nil
}

func (s *Storage) check() error {
	if s.daoRelease == nil {
		return errors.New("dao release is nil")
	}

	if s.daoPackageEvent == nil {
		return errors.New("dao package event is nil")
	}

	return nil
}

// ==================== IAgent Methods ====================

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

// GetReleaseAgent gets agent release.
func (s *Storage) GetReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) (*types.ReleaseAgent, error) {
	var (
		releaseAgent *types.ReleaseAgent
	)

	err := s.WrapFn(nCtx, metricOperationGetReleaseAgent, func(nCtx contextx.IContext) error {
		var err error
		releaseAgent, err = s.getReleaseAgent(nCtx, key.Generation, key.Platform, key.Version)
		if err != nil {
			logger.G.Sys().WithErr(err).With("key", key).Error("failed to get release agent")

			return fmt.Errorf("failed to get release agent: %w", err)
		}

		return nil
	})

	return releaseAgent, err
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

// DeleteReleaseAgent deletes agent release.
func (s *Storage) DeleteReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error {
	return s.WrapFn(nCtx, metricOperationDeleteReleaseAgent, func(nCtx contextx.IContext) error {
		err := s.deleteRelease(nCtx, types.ReleaseTypeAgent, key.Generation, key.Platform, key.Version, types.ReleaseNameAgent)
		if err != nil {
			logger.G.Sys().WithErr(err).With("key", key).Error("failed to delete release agent")

			return fmt.Errorf("failed to delete release agent: %w", err)
		}

		return nil
	})
}

// SetReleaseAgentLabelsMany sets many agent release labels.
func (s *Storage) SetReleaseAgentLabelsMany(nCtx contextx.IContext, labels []string, conditions ...*types.ReleaseCondition) error {
	return s.WrapFn(nCtx, metricOperationSetReleaseAgentLabelsMany, func(nCtx contextx.IContext) error {
		err := s.setReleaseLabelsMany(nCtx, types.ReleaseTypeAgent, labels, conditions...)
		if err != nil {
			logger.G.Sys().WithErr(err).Error("failed to set many release agent labels")

			return fmt.Errorf("failed to set many release agent labels: %w", err)
		}

		return nil
	})
}

// EnableReleaseAgent enables agent release active.
func (s *Storage) EnableReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error {
	return s.WrapFn(nCtx, metricOperationEnableReleaseAgent, func(nCtx contextx.IContext) error {
		err := s.enableRelease(nCtx, types.ReleaseTypeAgent, key.Generation, key.Platform, key.Version, types.ReleaseNameAgent)
		if err != nil {
			logger.G.Sys().WithErr(err).With("key", key).Error("failed to enable release agent")

			return fmt.Errorf("failed to enable release agent: %w", err)
		}

		return nil
	})
}

// DisableReleaseAgent disables agent release disactive.
func (s *Storage) DisableReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error {
	return s.WrapFn(nCtx, metricOperationDisableReleaseAgent, func(nCtx contextx.IContext) error {
		err := s.disableRelease(nCtx, types.ReleaseTypeAgent, key.Generation, key.Platform, key.Version, types.ReleaseNameAgent)
		if err != nil {
			logger.G.Sys().WithErr(err).With("key", key).Error("failed to disable release agent")

			return fmt.Errorf("failed to disable release agent: %w", err)
		}

		return nil
	})
}

// SetAsDefaultReleaseAgent sets the agent release as default.
func (s *Storage) SetAsDefaultReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error {
	return s.WrapFn(nCtx, metricOperationSetAsDefaultReleaseAgent, func(nCtx contextx.IContext) error {
		err := s.setAsDefaultRelease(nCtx, types.ReleaseTypeAgent, key.Generation, key.Platform, key.Version, types.ReleaseNameAgent)
		if err != nil {
			logger.G.Sys().WithErr(err).With("key", key).Error("failed to set release agent as default")

			return fmt.Errorf("failed to set release agent as default: %w", err)
		}

		return nil
	})
}

// CancelAsDefaultReleaseAgent cancels the agent release as default.
func (s *Storage) CancelAsDefaultReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error {
	return s.WrapFn(nCtx, metricOperationCancelAsDefaultReleaseAgent, func(nCtx contextx.IContext) error {
		err := s.cancelAsDefaultRelease(nCtx, types.ReleaseTypeAgent, key.Generation, key.Platform, key.Version, types.ReleaseNameAgent)
		if err != nil {
			logger.G.Sys().WithErr(err).With("key", key).Error("failed to cancel release agent as default")

			return fmt.Errorf("failed to cancel release agent as default: %w", err)
		}

		return nil
	})
}

// ==================== IProxy Methods ====================

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

// GetReleaseProxy gets proxy release.
func (s *Storage) GetReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) (*types.ReleaseProxy, error) {
	var (
		releaseProxy *types.ReleaseProxy
	)

	err := s.WrapFn(nCtx, metricOperationGetReleaseProxy, func(nCtx contextx.IContext) error {
		var err error
		releaseProxy, err = s.getReleaseProxy(nCtx, key.Generation, key.Platform, key.Version)
		if err != nil {
			logger.G.Sys().WithErr(err).With("key", key).Error("failed to get release proxy")

			return fmt.Errorf("failed to get release proxy: %w", err)
		}

		return nil
	})

	return releaseProxy, err
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

// DeleteReleaseProxy deletes proxy release.
func (s *Storage) DeleteReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error {
	return s.WrapFn(nCtx, metricOperationDeleteReleaseProxy, func(nCtx contextx.IContext) error {
		err := s.deleteRelease(nCtx, types.ReleaseTypeProxy, key.Generation, key.Platform, key.Version, types.ReleaseNameProxy)
		if err != nil {
			logger.G.Sys().WithErr(err).With("key", key).Error("failed to delete release proxy")

			return fmt.Errorf("failed to delete release proxy: %w", err)
		}

		return nil
	})
}

// SetReleaseProxyLabelsMany sets many proxy releases labels.
func (s *Storage) SetReleaseProxyLabelsMany(nCtx contextx.IContext, labels []string, conditions ...*types.ReleaseCondition) error {
	return s.WrapFn(nCtx, metricOperationSetReleaseProxyLabelsMany, func(nCtx contextx.IContext) error {
		err := s.setReleaseLabelsMany(nCtx, types.ReleaseTypeProxy, labels, conditions...)
		if err != nil {
			logger.G.Sys().WithErr(err).Error("failed to set many release proxy labels")

			return fmt.Errorf("failed to set many release proxy labels: %w", err)
		}

		return nil
	})
}

// EnableReleaseProxy enables proxy release active.
func (s *Storage) EnableReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error {
	return s.WrapFn(nCtx, metricOperationEnableReleaseProxy, func(nCtx contextx.IContext) error {
		err := s.enableRelease(nCtx, types.ReleaseTypeProxy, key.Generation, key.Platform, key.Version, types.ReleaseNameProxy)
		if err != nil {
			logger.G.Sys().WithErr(err).With("key", key).Error("failed to enable release proxy")

			return fmt.Errorf("failed to enable release proxy: %w", err)
		}

		return nil
	})
}

// DisableReleaseProxy disables proxy release disactive.
func (s *Storage) DisableReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error {
	return s.WrapFn(nCtx, metricOperationDisableReleaseProxy, func(nCtx contextx.IContext) error {
		err := s.disableRelease(nCtx, types.ReleaseTypeProxy, key.Generation, key.Platform, key.Version, types.ReleaseNameProxy)
		if err != nil {
			logger.G.Sys().WithErr(err).With("key", key).Error("failed to disable release proxy")

			return fmt.Errorf("failed to disable release proxy: %w", err)
		}

		return nil
	})
}

// SetAsDefaultReleaseProxy sets the proxy release as default.
func (s *Storage) SetAsDefaultReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error {
	return s.WrapFn(nCtx, metricOperationSetAsDefaultReleaseProxy, func(nCtx contextx.IContext) error {
		err := s.setAsDefaultRelease(nCtx, types.ReleaseTypeProxy, key.Generation, key.Platform, key.Version, types.ReleaseNameProxy)
		if err != nil {
			logger.G.Sys().WithErr(err).With("key", key).Error("failed to set release proxy as default")

			return fmt.Errorf("failed to set release proxy as default: %w", err)
		}

		return nil
	})
}

// CancelAsDefaultReleaseProxy cancels the proxy release as default.
func (s *Storage) CancelAsDefaultReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error {
	return s.WrapFn(nCtx, metricOperationCancelAsDefaultReleaseProxy, func(nCtx contextx.IContext) error {
		err := s.cancelAsDefaultRelease(nCtx, types.ReleaseTypeProxy, key.Generation, key.Platform, key.Version, types.ReleaseNameProxy)
		if err != nil {
			logger.G.Sys().WithErr(err).With("key", key).Error("failed to cancel release proxy as default")

			return fmt.Errorf("failed to cancel release proxy as default: %w", err)
		}

		return nil
	})
}

// ==================== IPlugin Methods ====================

// ListReleasePlugin lists plugin releases by page and conditions.
func (s *Storage) ListReleasePlugin(nCtx contextx.IContext, page types.Page,
	conditions ...*types.ReleaseCondition) ([]*types.ReleasePlugin, int64, error) {

	var (
		results []*types.ReleasePlugin
		num     int64
	)

	err := s.WrapFn(nCtx, metricOperationListReleasePlugin, func(nCtx contextx.IContext) error {
		var err error
		results, num, err = s.listReleasePlugin(nCtx, page, conditions...)

		return err
	})

	return results, num, err
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

// GetReleasePlugin gets plugin release.
func (s *Storage) GetReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) (*types.ReleasePlugin, error) {
	var releasePlugin *types.ReleasePlugin
	err := s.WrapFn(nCtx, metricOperationGetReleasePlugin, func(nCtx contextx.IContext) error {
		var err error
		releasePlugin, err = s.getReleasePlugin(nCtx, key.Name, key.Generation, key.Platform, key.Version)

		return err
	})

	return releasePlugin, err
}

// DeleteReleasePlugin deletes plugin release.
func (s *Storage) DeleteReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error {
	return s.WrapFn(nCtx, metricOperationDeleteReleasePlugin, func(nCtx contextx.IContext) error {
		return s.deleteRelease(nCtx, types.ReleaseTypePlugin, key.Generation, key.Platform, key.Version, key.Name)
	})
}

// EnableReleasePlugin enables plugin release active.
func (s *Storage) EnableReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error {
	return s.WrapFn(nCtx, metricOperationEnableReleasePlugin, func(nCtx contextx.IContext) error {
		return s.enableRelease(nCtx, types.ReleaseTypePlugin, key.Generation, key.Platform, key.Version, key.Name)
	})
}

// DisableReleasePlugin disables plugin release disactive.
func (s *Storage) DisableReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error {
	return s.WrapFn(nCtx, metricOperationDisableReleasePlugin, func(nCtx contextx.IContext) error {
		return s.disableRelease(nCtx, types.ReleaseTypePlugin, key.Generation, key.Platform, key.Version, key.Name)
	})
}

// SetHiddenReleasePlugin sets the plugin release hidden state.
func (s *Storage) SetHiddenReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey, isHidden bool) error {
	return s.WrapFn(nCtx, metricOperationSetHiddenReleasePlugin, func(nCtx contextx.IContext) error {
		return s.setReleaseHidden(nCtx, types.ReleaseTypePlugin, key.Generation, key.Platform, key.Version, key.Name, isHidden)
	})
}

// SetAsDefaultReleasePlugin sets the plugin release as default.
func (s *Storage) SetAsDefaultReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error {
	return s.WrapFn(nCtx, metricOperationSetAsDefaultReleasePlugin, func(nCtx contextx.IContext) error {
		return s.setAsDefaultRelease(nCtx, types.ReleaseTypePlugin, key.Generation, key.Platform, key.Version, key.Name)
	})
}

// CancelAsDefaultReleasePlugin cancels the plugin release as default.
func (s *Storage) CancelAsDefaultReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error {
	return s.WrapFn(nCtx, metricOperationCancelAsDefaultReleasePlugin, func(nCtx contextx.IContext) error {
		return s.cancelAsDefaultRelease(nCtx, types.ReleaseTypePlugin, key.Generation, key.Platform, key.Version, key.Name)
	})
}

// ExistReleasePlugin exist plugin release.
func (s *Storage) ExistReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) (bool, error) {
	var exist bool
	err := s.WrapFn(nCtx, metricOperationExistReleasePlugin, func(nCtx contextx.IContext) error {
		var err error
		exist, err = s.existRelease(nCtx, types.ReleaseTypePlugin, key.Generation, key.Platform, key.Version, key.Name)

		return err
	})

	return exist, err
}

// GetReleasePluginDefaultVersion gets release plugin default version by name, generation and platform.
func (s *Storage) GetReleasePluginDefaultVersion(nCtx contextx.IContext, name string, gen types.Generation, plat platfmt.Platform) (string, error) {
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

// ==================== ICert Methods ====================

// ListReleaseCert lists cert releases by page and conditions.
func (s *Storage) ListReleaseCert(nCtx contextx.IContext, page types.Page,
	conditions ...*types.ReleaseCondition) ([]*types.ReleaseCert, int64, error) {

	var (
		results []*types.ReleaseCert
		num     int64
	)

	err := s.WrapFn(nCtx, metricOperationListReleaseCert, func(nCtx contextx.IContext) error {
		rls, n, err := s.listRelease(nCtx, types.ReleaseTypeCert, page, conditions...)
		if err != nil {
			logger.G.Sys().WithErr(err).Error("failed to list release cert")

			return fmt.Errorf("failed to list release cert: %w", err)
		}
		num = n

		results = make([]*types.ReleaseCert, len(rls))
		for idx, r := range rls {
			results[idx] = &types.ReleaseCert{
				Release: *r,
			}
		}

		return nil
	})

	return results, num, err
}

// CountReleaseCert counts cert release by conditions.
func (s *Storage) CountReleaseCert(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) (int64, error) {
	var num int64
	err := s.WrapFn(nCtx, metricOperationCountReleaseCert, func(nCtx contextx.IContext) error {
		var err error
		if num, err = s.countRelease(nCtx, types.ReleaseTypeCert, conditions...); err != nil {
			logger.G.Sys().WithErr(err).Error("failed to count release cert")

			return fmt.Errorf("failed to count release cert: %w", err)
		}

		return nil
	})

	return num, err
}

// GetReleaseCert gets cert release.
func (s *Storage) GetReleaseCert(nCtx contextx.IContext, key types.ReleaseCertKey) (*types.ReleaseCert, error) {
	var releaseCert *types.ReleaseCert
	err := s.WrapFn(nCtx, metricOperationGetReleaseCert, func(nCtx contextx.IContext) error {
		rls, err := s.getRelease(
			nCtx, types.ReleaseTypeCert, key.Generation, platfmt.UnknownPlatform(), types.ReleaseVersionCert, types.ReleaseNameCert)
		if err != nil {
			logger.G.Sys().WithErr(err).With("key", key).Error("failed to get release cert")

			return fmt.Errorf("failed to get release cert: %w", err)
		}

		releaseCert = &types.ReleaseCert{
			Release: *rls,
		}

		return nil
	})

	return releaseCert, err
}

// DeleteReleaseCert deletes cert release.
func (s *Storage) DeleteReleaseCert(nCtx contextx.IContext, key types.ReleaseCertKey) error {
	return s.WrapFn(nCtx, metricOperationDeleteReleaseCert, func(nCtx contextx.IContext) error {
		if err := s.deleteRelease(
			nCtx, types.ReleaseTypeCert, key.Generation, platfmt.UnknownPlatform(), types.ReleaseVersionCert, types.ReleaseNameCert); err != nil {
			logger.G.Sys().WithErr(err).With("key", key).Error("failed to delete release cert")

			return fmt.Errorf("failed to delete release cert: %w", err)
		}

		return nil
	})
}

// ==================== IBinTool Methods ====================

// ListReleaseBinTool lists bintool releases by page and conditions.
func (s *Storage) ListReleaseBinTool(nCtx contextx.IContext, page types.Page,
	conditions ...*types.ReleaseCondition) ([]*types.ReleaseBinTool, int64, error) {

	var (
		results []*types.ReleaseBinTool
		num     int64
	)

	err := s.WrapFn(nCtx, metricOperationListReleaseBinTool, func(nCtx contextx.IContext) error {
		rls, n, err := s.listRelease(nCtx, types.ReleaseTypeBinTool, page, conditions...)
		if err != nil {
			logger.G.Sys().WithErr(err).Error("failed to list release bintool")

			return fmt.Errorf("failed to list release bintool: %w", err)
		}
		num = n

		results = make([]*types.ReleaseBinTool, len(rls))
		for idx, r := range rls {
			results[idx] = &types.ReleaseBinTool{
				Release: *r,
			}
		}

		return nil
	})

	return results, num, err
}

// CountReleaseBinTool counts bintool release by conditions.
func (s *Storage) CountReleaseBinTool(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) (int64, error) {
	var num int64
	err := s.WrapFn(nCtx, metricOperationCountReleaseBinTool, func(nCtx contextx.IContext) error {
		var err error
		if num, err = s.countRelease(nCtx, types.ReleaseTypeBinTool, conditions...); err != nil {
			logger.G.Sys().WithErr(err).Error("failed to count release bintool")

			return fmt.Errorf("failed to count release bintool: %w", err)
		}

		return nil
	})

	return num, err
}

// GetReleaseBinTool gets bintool release.
func (s *Storage) GetReleaseBinTool(nCtx contextx.IContext, key types.ReleaseBinToolKey) (*types.ReleaseBinTool, error) {
	var releaseBinTool *types.ReleaseBinTool
	err := s.WrapFn(nCtx, metricOperationGetReleaseBinTool, func(nCtx contextx.IContext) error {
		rls, err := s.getRelease(
			nCtx, types.ReleaseTypeBinTool, key.Generation, platfmt.UnknownPlatform(), types.ReleaseVersionBinTool, types.ReleaseNameBinTool)
		if err != nil {
			logger.G.Sys().WithErr(err).With("key", key).Error("failed to get release bintool")

			return fmt.Errorf("failed to get release bintool: %w", err)
		}

		releaseBinTool = &types.ReleaseBinTool{
			Release: *rls,
		}

		return nil
	})

	return releaseBinTool, err
}

// DeleteReleaseBinTool deletes bintool release.
func (s *Storage) DeleteReleaseBinTool(nCtx contextx.IContext, key types.ReleaseBinToolKey) error {
	return s.WrapFn(nCtx, metricOperationDeleteReleaseBinTool, func(nCtx contextx.IContext) error {
		if err := s.deleteRelease(
			nCtx,
			types.ReleaseTypeBinTool, key.Generation, platfmt.UnknownPlatform(), types.ReleaseVersionBinTool, types.ReleaseNameBinTool); err != nil {
			logger.G.Sys().WithErr(err).With("key", key).Error("failed to delete release bintool")

			return fmt.Errorf("failed to delete release bintool: %w", err)
		}

		return nil
	})
}

// ==================== IPluginBinTool Methods ====================

// ListReleasePluginBinTool lists plugin bintool release by page and conditions.
func (s *Storage) ListReleasePluginBinTool(nCtx contextx.IContext, page types.Page,
	conditions ...*types.ReleaseCondition) ([]*types.ReleasePluginBinTool, int64, error) {

	var (
		results []*types.ReleasePluginBinTool
		num     int64
	)

	err := s.WrapFn(nCtx, metricOperationListReleasePluginBinTool, func(nCtx contextx.IContext) error {
		rls, n, err := s.listRelease(nCtx, types.ReleaseTypePluginBinTool, page, conditions...)
		if err != nil {
			logger.G.Sys().WithErr(err).Error("failed to list release plugin bintool")

			return fmt.Errorf("failed to list release plugin bintool: %w", err)
		}
		num = n

		results = make([]*types.ReleasePluginBinTool, len(rls))
		for idx, r := range rls {
			results[idx] = &types.ReleasePluginBinTool{
				Release: *r,
			}
		}

		return nil
	})

	return results, num, err
}

// CountReleasePluginBinTool counts plugin bintool release by conditions.
func (s *Storage) CountReleasePluginBinTool(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) (int64, error) {
	var num int64
	err := s.WrapFn(nCtx, metricOperationCountReleasePluginBinTool, func(nCtx contextx.IContext) error {
		var err error
		if num, err = s.countRelease(nCtx, types.ReleaseTypePluginBinTool, conditions...); err != nil {
			logger.G.Sys().WithErr(err).Error("failed to count release plugin bintool")

			return fmt.Errorf("failed to count release plugin bintool: %w", err)
		}

		return nil
	})

	return num, err
}

// GetReleasePluginBinTool gets plugin bintool release.
func (s *Storage) GetReleasePluginBinTool(nCtx contextx.IContext, key types.ReleasePluginBinToolKey) (*types.ReleasePluginBinTool, error) {
	var releasePluginBinTool *types.ReleasePluginBinTool
	err := s.WrapFn(nCtx, metricOperationGetReleasePluginBinTool, func(nCtx contextx.IContext) error {
		rls, err := s.getRelease(
			nCtx, types.ReleaseTypePluginBinTool, key.Generation, platfmt.UnknownPlatform(), types.ReleaseVersionPluginBinTool, key.Name)
		if err != nil {
			logger.G.Sys().WithErr(err).With("key", key).Error("failed to get release plugin bintool")

			return fmt.Errorf("failed to get release plugin bintool: %w", err)
		}

		releasePluginBinTool = &types.ReleasePluginBinTool{
			Release: *rls,
		}

		return nil
	})

	return releasePluginBinTool, err
}

// DeleteReleasePluginBinTool deletes plugin bintool release.
func (s *Storage) DeleteReleasePluginBinTool(nCtx contextx.IContext, key types.ReleasePluginBinToolKey) error {
	return s.WrapFn(nCtx, metricOperationDeleteReleasePluginBinTool, func(nCtx contextx.IContext) error {
		if err := s.deleteRelease(
			nCtx, types.ReleaseTypePluginBinTool, key.Generation, platfmt.UnknownPlatform(), types.ReleaseVersionPluginBinTool, key.Name); err != nil {
			logger.G.Sys().WithErr(err).With("key", key).Error("failed to delete release plugin bintool")

			return fmt.Errorf("failed to delete release plugin bintool: %w", err)
		}

		return nil
	})
}

// ==================== IPackageEvent Methods ====================

// CountPackageEvent counts package events by conditions.
func (s *Storage) CountPackageEvent(nCtx contextx.IContext, conditions ...*types.PackageEventCondition) (int64, error) {
	var num int64
	err := s.WrapFn(nCtx, metricOperationCountPackageEvent, func(nCtx contextx.IContext) error {
		var err error
		if num, err = s.countPakcageEvent(nCtx, conditions...); err != nil {
			logger.G.Sys().WithErr(err).Error("failed to get count package event")

			return fmt.Errorf("failed to count package event: %w", err)
		}

		return nil
	})

	return num, err
}

// ListPackageEvent lists package events by page and conditions.
func (s *Storage) ListPackageEvent(nCtx contextx.IContext, page types.Page, conditions ...*types.PackageEventCondition) (
	[]*types.PackageEvent, int64, error) {

	var (
		results []*types.PackageEvent
		num     int64
	)

	err := s.WrapFn(nCtx, metricOperationListPackageEvent, func(nCtx contextx.IContext) error {
		var err error
		if results, num, err = s.listPackageEvent(nCtx, page, conditions...); err != nil {
			logger.G.Sys().WithErr(err).Error("failed to list package event")

			return fmt.Errorf("failed to list package event: %w", err)
		}

		return nil
	})

	return results, num, err
}

// CreateManyPackageEvent creates multiple package events.
func (s *Storage) CreateManyPackageEvent(nCtx contextx.IContext, events ...*types.PackageEvent) error {
	return s.WrapFn(nCtx, metricOperationCreateManyPackageEvent, func(nCtx contextx.IContext) error {
		err := s.createManyPackageEvent(nCtx, events...)
		if err != nil {
			logger.G.Sys().WithErr(err).Error("failed to create many package event")

			return fmt.Errorf("failed to create many package event: %w", err)
		}

		return nil
	})
}

// DistinctPackageEvent distincts package event fields.
func (s *Storage) DistinctPackageEvent(
	nCtx contextx.IContext, request types.PackageEventDistinctRequest, conditions ...*types.PackageEventCondition) (
	*types.PackageEventDistinctResult, error) {

	var (
		data *types.PackageEventDistinctResult
	)

	err := s.WrapFn(nCtx, metricOperationDistinctPackageEvent, func(nCtx contextx.IContext) error {
		var err error
		if data, err = s.distinctPackageEvent(nCtx, request, conditions...); err != nil {
			logger.G.Sys().WithErr(err).Error("failed to distinct package event")

			return fmt.Errorf("failed to distinct package event: %w", err)
		}

		return nil
	})

	return data, err
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
