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

package release

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/markcache"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/cache"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const releaseCacheExpiration = time.Minute

// releaseGroup identifies the entity group of cached release queries. All query keys
// of one group share a single per-tenant invalidation marker.
type releaseGroup string

const (
	releaseGroupAgent         releaseGroup = "agent"
	releaseGroupProxy         releaseGroup = "proxy"
	releaseGroupPlugin        releaseGroup = "plugin"
	releaseGroupCert          releaseGroup = "cert"
	releaseGroupBinTool       releaseGroup = "bintool"
	releaseGroupPluginBinTool releaseGroup = "pluginbintool"
)

// releaseMarkerKey builds the invalidation marker key for a tenant and release group.
func releaseMarkerKey(tenantID string, group releaseGroup) string {
	return fmt.Sprintf("bknm:file:release:inv:%s:%s", tenantID, group)
}

// CachedStorage decorates IStorage: release reads are served from a process-local
// cache, and release writes invalidate the shared group markers so every file
// service pod refreshes its cached entries.
type CachedStorage struct {
	IStorage

	cache *markcache.Cache
	keys  cache.IKeyGenerator
}

var _ IStorage = (*CachedStorage)(nil)

// NewCachedStorage creates a caching decorator over the given release storage.
// A nil marker store keeps plain TTL-only caching.
func NewCachedStorage(inner IStorage, markers markcache.MarkerStore) *CachedStorage {
	return &CachedStorage{
		IStorage: inner,
		cache:    markcache.New(cache.NewMemoryCache(releaseCacheExpiration), markers),
		keys:     cache.NewDefaultKeyGenerator("file:release"),
	}
}

// listResult is the cached value shape for list queries.
type listResult[T any] struct {
	Items []T   `json:"items"`
	Total int64 `json:"total"`
}

// cacheKey builds the cache key for a release query method and its arguments.
func (cs *CachedStorage) cacheKey(tenantID, method string, args ...any) (string, error) {
	data, err := json.Marshal(args)
	if err != nil {
		return "", fmt.Errorf("failed to marshal release cache key args: %w", err)
	}

	return cs.keys.Generate(method, tenantID, base64.RawURLEncoding.EncodeToString(data)), nil
}

// cachedQuery serves a release query from the cache or loads it, refreshing entries
// once the group marker has been invalidated by a write on any pod.
func cachedQuery[T any](
	cs *CachedStorage, nCtx contextx.IContext, group releaseGroup, method string,
	load func() (T, error), keyArgs ...any,
) (T, error) {

	var zero T

	cacheKey, err := cs.cacheKey(nCtx.TenantID(), method, keyArgs...)
	if err != nil {
		return zero, err
	}

	data, err := cs.cache.GetOrLoad(nCtx, cacheKey, releaseMarkerKey(nCtx.TenantID(), group),
		func(context.Context) ([]byte, error) {
			result, err := load()
			if err != nil {
				return nil, err
			}

			return json.Marshal(result)
		})
	if err != nil {
		return zero, err
	}

	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		return zero, fmt.Errorf("failed to unmarshal cached release query. method(%s): %w", method, err)
	}

	return result, nil
}

// invalidate bumps the shared markers of the given groups after a successful write.
// The write has already committed, so invalidation failures only degrade freshness
// back to the TTL window and are logged instead of returned.
func (cs *CachedStorage) invalidate(nCtx contextx.IContext, groups ...releaseGroup) {
	keys := make([]string, 0, len(groups))
	for _, group := range groups {
		keys = append(keys, releaseMarkerKey(nCtx.TenantID(), group))
	}

	if err := cs.cache.Invalidate(nCtx, keys...); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to invalidate release cache markers after write")
	}
}

// ListReleaseAgent lists cached agent releases.
func (cs *CachedStorage) ListReleaseAgent(nCtx contextx.IContext, page types.Page, conditions ...*types.ReleaseCondition) (
	[]*types.ReleaseAgent, int64, error) {

	result, err := cachedQuery(cs, nCtx, releaseGroupAgent, "agent-list",
		func() (listResult[*types.ReleaseAgent], error) {
			items, total, err := cs.IStorage.ListReleaseAgent(nCtx, page, conditions...)
			if err != nil {
				return listResult[*types.ReleaseAgent]{}, err
			}

			return listResult[*types.ReleaseAgent]{Items: items, Total: total}, nil
		}, page, conditions)
	if err != nil {
		return nil, 0, err
	}

	return result.Items, result.Total, nil
}

// CountReleaseAgent counts cached release entries.
func (cs *CachedStorage) CountReleaseAgent(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) (int64, error) {
	return cachedQuery(cs, nCtx, releaseGroupAgent, "agent-count",
		func() (int64, error) { return cs.IStorage.CountReleaseAgent(nCtx, conditions...) }, conditions)
}

// GetReleaseAgent gets a cached release entry.
func (cs *CachedStorage) GetReleaseAgent(
	nCtx contextx.IContext, gen types.Generation, plat platfmt.Platform, version string,
) (*types.ReleaseAgent, error) {

	return cachedQuery(cs, nCtx, releaseGroupAgent, "agent",
		func() (*types.ReleaseAgent, error) { return cs.IStorage.GetReleaseAgent(nCtx, gen, plat, version) }, gen, plat, version)
}

// DistinctReleaseAgent gets cached distinct release fields.
func (cs *CachedStorage) DistinctReleaseAgent(nCtx contextx.IContext, fields types.ReleaseDistinctField, conditions ...*types.ReleaseCondition) (
	*types.ReleaseDistinctResult, error) {

	return cachedQuery(cs, nCtx, releaseGroupAgent, "agent-distinct",
		func() (*types.ReleaseDistinctResult, error) {
			return cs.IStorage.DistinctReleaseAgent(nCtx, fields, conditions...)
		}, fields, conditions)
}

// ListReleaseProxy lists cached proxy releases.
func (cs *CachedStorage) ListReleaseProxy(nCtx contextx.IContext, page types.Page, conditions ...*types.ReleaseCondition) (
	[]*types.ReleaseProxy, int64, error) {

	result, err := cachedQuery(cs, nCtx, releaseGroupProxy, "proxy-list",
		func() (listResult[*types.ReleaseProxy], error) {
			items, total, err := cs.IStorage.ListReleaseProxy(nCtx, page, conditions...)
			if err != nil {
				return listResult[*types.ReleaseProxy]{}, err
			}

			return listResult[*types.ReleaseProxy]{Items: items, Total: total}, nil
		}, page, conditions)
	if err != nil {
		return nil, 0, err
	}

	return result.Items, result.Total, nil
}

// CountReleaseProxy counts cached release entries.
func (cs *CachedStorage) CountReleaseProxy(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) (int64, error) {
	return cachedQuery(cs, nCtx, releaseGroupProxy, "proxy-count",
		func() (int64, error) { return cs.IStorage.CountReleaseProxy(nCtx, conditions...) }, conditions)
}

// GetReleaseProxy gets a cached release entry.
func (cs *CachedStorage) GetReleaseProxy(
	nCtx contextx.IContext, gen types.Generation, plat platfmt.Platform, version string,
) (*types.ReleaseProxy, error) {

	return cachedQuery(cs, nCtx, releaseGroupProxy, "proxy",
		func() (*types.ReleaseProxy, error) { return cs.IStorage.GetReleaseProxy(nCtx, gen, plat, version) }, gen, plat, version)
}

// DistinctReleaseProxy gets cached distinct release fields.
func (cs *CachedStorage) DistinctReleaseProxy(nCtx contextx.IContext, fields types.ReleaseDistinctField, conditions ...*types.ReleaseCondition) (
	*types.ReleaseDistinctResult, error) {

	return cachedQuery(cs, nCtx, releaseGroupProxy, "proxy-distinct",
		func() (*types.ReleaseDistinctResult, error) {
			return cs.IStorage.DistinctReleaseProxy(nCtx, fields, conditions...)
		}, fields, conditions)
}

// ListReleasePlugin lists cached plugin releases.
func (cs *CachedStorage) ListReleasePlugin(nCtx contextx.IContext, page types.Page, conditions ...*types.ReleaseCondition) (
	[]*types.ReleasePlugin, int64, error) {

	result, err := cachedQuery(cs, nCtx, releaseGroupPlugin, "plugin-list",
		func() (listResult[*types.ReleasePlugin], error) {
			items, total, err := cs.IStorage.ListReleasePlugin(nCtx, page, conditions...)
			if err != nil {
				return listResult[*types.ReleasePlugin]{}, err
			}

			return listResult[*types.ReleasePlugin]{Items: items, Total: total}, nil
		}, page, conditions)
	if err != nil {
		return nil, 0, err
	}

	return result.Items, result.Total, nil
}

// CountReleasePlugin counts cached release entries.
func (cs *CachedStorage) CountReleasePlugin(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) (int64, error) {
	return cachedQuery(cs, nCtx, releaseGroupPlugin, "plugin-count",
		func() (int64, error) { return cs.IStorage.CountReleasePlugin(nCtx, conditions...) }, conditions)
}

// DistinctReleasePlugin gets cached distinct release fields.
func (cs *CachedStorage) DistinctReleasePlugin(nCtx contextx.IContext, fields types.ReleaseDistinctField, conditions ...*types.ReleaseCondition) (
	*types.ReleaseDistinctResult, error) {

	return cachedQuery(cs, nCtx, releaseGroupPlugin, "plugin-distinct",
		func() (*types.ReleaseDistinctResult, error) {
			return cs.IStorage.DistinctReleasePlugin(nCtx, fields, conditions...)
		}, fields, conditions)
}

// GetReleasePlugin gets a cached release entry.
func (cs *CachedStorage) GetReleasePlugin(
	nCtx contextx.IContext, name string, gen types.Generation, plat platfmt.Platform, version string,
) (*types.ReleasePlugin, error) {

	return cachedQuery(cs, nCtx, releaseGroupPlugin, "plugin",
		func() (*types.ReleasePlugin, error) {
			return cs.IStorage.GetReleasePlugin(nCtx, name, gen, plat, version)
		}, name, gen, plat, version)
}

// GetReleasePluginDefaultVersion gets a cached release entry.
func (cs *CachedStorage) GetReleasePluginDefaultVersion(
	nCtx contextx.IContext, name string, gen types.Generation, plat platfmt.Platform,
) (string, error) {

	return cachedQuery(cs, nCtx, releaseGroupPlugin, "plugin-default-version",
		func() (string, error) { return cs.IStorage.GetReleasePluginDefaultVersion(nCtx, name, gen, plat) }, name, gen, plat)
}

// DistinctNameReleasePlugin gets cached distinct release names.
func (cs *CachedStorage) DistinctNameReleasePlugin(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) ([]string, error) {
	return cachedQuery(cs, nCtx, releaseGroupPlugin, "plugin-distinct-name",
		func() ([]string, error) { return cs.IStorage.DistinctNameReleasePlugin(nCtx, conditions...) }, conditions)
}

// ExistReleasePlugin checks cached plugin release existence.
func (cs *CachedStorage) ExistReleasePlugin(nCtx contextx.IContext, pluginName string, version string, plats ...platfmt.Platform) (bool, error) {
	return cachedQuery(cs, nCtx, releaseGroupPlugin, "plugin-exist",
		func() (bool, error) { return cs.IStorage.ExistReleasePlugin(nCtx, pluginName, version, plats...) }, pluginName, version, plats)
}

// ListReleaseCert lists cached cert releases.
func (cs *CachedStorage) ListReleaseCert(nCtx contextx.IContext, page types.Page, conditions ...*types.ReleaseCondition) (
	[]*types.ReleaseCert, int64, error) {

	result, err := cachedQuery(cs, nCtx, releaseGroupCert, "cert-list",
		func() (listResult[*types.ReleaseCert], error) {
			items, total, err := cs.IStorage.ListReleaseCert(nCtx, page, conditions...)
			if err != nil {
				return listResult[*types.ReleaseCert]{}, err
			}

			return listResult[*types.ReleaseCert]{Items: items, Total: total}, nil
		}, page, conditions)
	if err != nil {
		return nil, 0, err
	}

	return result.Items, result.Total, nil
}

// ListReleaseBinTool lists cached bintool releases.
func (cs *CachedStorage) ListReleaseBinTool(nCtx contextx.IContext, page types.Page, conditions ...*types.ReleaseCondition) (
	[]*types.ReleaseBinTool, int64, error) {

	result, err := cachedQuery(cs, nCtx, releaseGroupBinTool, "bintool-list",
		func() (listResult[*types.ReleaseBinTool], error) {
			items, total, err := cs.IStorage.ListReleaseBinTool(nCtx, page, conditions...)
			if err != nil {
				return listResult[*types.ReleaseBinTool]{}, err
			}

			return listResult[*types.ReleaseBinTool]{Items: items, Total: total}, nil
		}, page, conditions)
	if err != nil {
		return nil, 0, err
	}

	return result.Items, result.Total, nil
}

// ListReleasePluginBinTool lists cached plugin bintool releases.
func (cs *CachedStorage) ListReleasePluginBinTool(nCtx contextx.IContext, page types.Page, conditions ...*types.ReleaseCondition) (
	[]*types.ReleasePluginBinTool, int64, error) {

	result, err := cachedQuery(cs, nCtx, releaseGroupPluginBinTool, "plugin-bintool-list",
		func() (listResult[*types.ReleasePluginBinTool], error) {
			items, total, err := cs.IStorage.ListReleasePluginBinTool(nCtx, page, conditions...)
			if err != nil {
				return listResult[*types.ReleasePluginBinTool]{}, err
			}

			return listResult[*types.ReleasePluginBinTool]{Items: items, Total: total}, nil
		}, page, conditions)
	if err != nil {
		return nil, 0, err
	}

	return result.Items, result.Total, nil
}

// DistinctNameReleasePluginBinTool gets cached distinct release names.
func (cs *CachedStorage) DistinctNameReleasePluginBinTool(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) ([]string, error) {
	return cachedQuery(cs, nCtx, releaseGroupPluginBinTool, "plugin-bintool-distinct-name",
		func() ([]string, error) { return cs.IStorage.DistinctNameReleasePluginBinTool(nCtx, conditions...) }, conditions)
}

// EnableReleaseAgent enables an agent release and invalidates the group marker.
func (cs *CachedStorage) EnableReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error {
	if err := cs.IStorage.EnableReleaseAgent(nCtx, key); err != nil {
		return err
	}

	cs.invalidate(nCtx, releaseGroupAgent)

	return nil
}

// DisableReleaseAgent disables an agent release and invalidates the group marker.
func (cs *CachedStorage) DisableReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error {
	if err := cs.IStorage.DisableReleaseAgent(nCtx, key); err != nil {
		return err
	}

	cs.invalidate(nCtx, releaseGroupAgent)

	return nil
}

// SetAsDefaultReleaseAgent sets an agent release as default and invalidates the group marker.
func (cs *CachedStorage) SetAsDefaultReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error {
	if err := cs.IStorage.SetAsDefaultReleaseAgent(nCtx, key); err != nil {
		return err
	}

	cs.invalidate(nCtx, releaseGroupAgent)

	return nil
}

// CancelAsDefaultReleaseAgent cancels an agent release default and invalidates the group marker.
func (cs *CachedStorage) CancelAsDefaultReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error {
	if err := cs.IStorage.CancelAsDefaultReleaseAgent(nCtx, key); err != nil {
		return err
	}

	cs.invalidate(nCtx, releaseGroupAgent)

	return nil
}

// DeleteReleaseAgent deletes an agent release and invalidates the group marker.
func (cs *CachedStorage) DeleteReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error {
	if err := cs.IStorage.DeleteReleaseAgent(nCtx, key); err != nil {
		return err
	}

	cs.invalidate(nCtx, releaseGroupAgent)

	return nil
}

// SetReleaseAgentLabelsMany replaces labels for matching releases and invalidates the group marker.
func (cs *CachedStorage) SetReleaseAgentLabelsMany(nCtx contextx.IContext, labels []string, conditions ...*types.ReleaseCondition) error {
	if err := cs.IStorage.SetReleaseAgentLabelsMany(nCtx, labels, conditions...); err != nil {
		return err
	}

	cs.invalidate(nCtx, releaseGroupAgent)

	return nil
}

// UpsertManyReleaseAgent upserts release entries and invalidates the group marker.
func (cs *CachedStorage) UpsertManyReleaseAgent(nCtx contextx.IContext, releaseAgents []*types.ReleaseAgent) error {
	if err := cs.IStorage.UpsertManyReleaseAgent(nCtx, releaseAgents); err != nil {
		return err
	}

	cs.invalidate(nCtx, releaseGroupAgent)

	return nil
}

// EnableReleaseProxy enables a proxy release and invalidates the group marker.
func (cs *CachedStorage) EnableReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error {
	if err := cs.IStorage.EnableReleaseProxy(nCtx, key); err != nil {
		return err
	}

	cs.invalidate(nCtx, releaseGroupProxy)

	return nil
}

// DisableReleaseProxy disables a proxy release and invalidates the group marker.
func (cs *CachedStorage) DisableReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error {
	if err := cs.IStorage.DisableReleaseProxy(nCtx, key); err != nil {
		return err
	}

	cs.invalidate(nCtx, releaseGroupProxy)

	return nil
}

// SetAsDefaultReleaseProxy sets a proxy release as default and invalidates the group marker.
func (cs *CachedStorage) SetAsDefaultReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error {
	if err := cs.IStorage.SetAsDefaultReleaseProxy(nCtx, key); err != nil {
		return err
	}

	cs.invalidate(nCtx, releaseGroupProxy)

	return nil
}

// CancelAsDefaultReleaseProxy cancels a proxy release default and invalidates the group marker.
func (cs *CachedStorage) CancelAsDefaultReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error {
	if err := cs.IStorage.CancelAsDefaultReleaseProxy(nCtx, key); err != nil {
		return err
	}

	cs.invalidate(nCtx, releaseGroupProxy)

	return nil
}

// DeleteReleaseProxy deletes a proxy release and invalidates the group marker.
func (cs *CachedStorage) DeleteReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error {
	if err := cs.IStorage.DeleteReleaseProxy(nCtx, key); err != nil {
		return err
	}

	cs.invalidate(nCtx, releaseGroupProxy)

	return nil
}

// SetReleaseProxyLabelsMany replaces labels for matching releases and invalidates the group marker.
func (cs *CachedStorage) SetReleaseProxyLabelsMany(nCtx contextx.IContext, labels []string, conditions ...*types.ReleaseCondition) error {
	if err := cs.IStorage.SetReleaseProxyLabelsMany(nCtx, labels, conditions...); err != nil {
		return err
	}

	cs.invalidate(nCtx, releaseGroupProxy)

	return nil
}

// UpsertManyReleaseProxy upserts release entries and invalidates the group marker.
func (cs *CachedStorage) UpsertManyReleaseProxy(nCtx contextx.IContext, releaseProxys []*types.ReleaseProxy) error {
	if err := cs.IStorage.UpsertManyReleaseProxy(nCtx, releaseProxys); err != nil {
		return err
	}

	cs.invalidate(nCtx, releaseGroupProxy)

	return nil
}

// EnableReleasePlugin enables a plugin release and invalidates the group marker.
func (cs *CachedStorage) EnableReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error {
	if err := cs.IStorage.EnableReleasePlugin(nCtx, key); err != nil {
		return err
	}

	cs.invalidate(nCtx, releaseGroupPlugin)

	return nil
}

// DisableReleasePlugin disables a plugin release and invalidates the group marker.
func (cs *CachedStorage) DisableReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error {
	if err := cs.IStorage.DisableReleasePlugin(nCtx, key); err != nil {
		return err
	}

	cs.invalidate(nCtx, releaseGroupPlugin)

	return nil
}

// SetAsDefaultReleasePlugin sets a plugin release as default and invalidates the group marker.
func (cs *CachedStorage) SetAsDefaultReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error {
	if err := cs.IStorage.SetAsDefaultReleasePlugin(nCtx, key); err != nil {
		return err
	}

	cs.invalidate(nCtx, releaseGroupPlugin)

	return nil
}

// CancelAsDefaultReleasePlugin cancels a plugin release default and invalidates the group marker.
func (cs *CachedStorage) CancelAsDefaultReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error {
	if err := cs.IStorage.CancelAsDefaultReleasePlugin(nCtx, key); err != nil {
		return err
	}

	cs.invalidate(nCtx, releaseGroupPlugin)

	return nil
}

// SetHiddenReleasePlugin hides a plugin release and invalidates the group marker.
func (cs *CachedStorage) SetHiddenReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error {
	if err := cs.IStorage.SetHiddenReleasePlugin(nCtx, key); err != nil {
		return err
	}

	cs.invalidate(nCtx, releaseGroupPlugin)

	return nil
}

// CancelHiddenReleasePlugin unhides a plugin release and invalidates the group marker.
func (cs *CachedStorage) CancelHiddenReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error {
	if err := cs.IStorage.CancelHiddenReleasePlugin(nCtx, key); err != nil {
		return err
	}

	cs.invalidate(nCtx, releaseGroupPlugin)

	return nil
}

// DeleteReleasePlugin deletes a plugin release and invalidates the group marker.
func (cs *CachedStorage) DeleteReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error {
	if err := cs.IStorage.DeleteReleasePlugin(nCtx, key); err != nil {
		return err
	}

	cs.invalidate(nCtx, releaseGroupPlugin)

	return nil
}

// UpsertManyReleasePlugin upserts release entries and invalidates the group marker.
func (cs *CachedStorage) UpsertManyReleasePlugin(nCtx contextx.IContext, releasePlugins []*types.ReleasePlugin) error {
	if err := cs.IStorage.UpsertManyReleasePlugin(nCtx, releasePlugins); err != nil {
		return err
	}

	cs.invalidate(nCtx, releaseGroupPlugin)

	return nil
}

// DeleteReleaseCert deletes a cert release and invalidates the group marker.
func (cs *CachedStorage) DeleteReleaseCert(nCtx contextx.IContext, key types.ReleaseCertKey) error {
	if err := cs.IStorage.DeleteReleaseCert(nCtx, key); err != nil {
		return err
	}

	cs.invalidate(nCtx, releaseGroupCert)

	return nil
}

// UpsertReleaseCert upserts release entries and invalidates the group marker.
func (cs *CachedStorage) UpsertReleaseCert(nCtx contextx.IContext, cert types.ReleaseCert) error {
	if err := cs.IStorage.UpsertReleaseCert(nCtx, cert); err != nil {
		return err
	}

	cs.invalidate(nCtx, releaseGroupCert)

	return nil
}

// DeleteReleaseBinTool deletes a bintool release and invalidates the group marker.
func (cs *CachedStorage) DeleteReleaseBinTool(nCtx contextx.IContext, key types.ReleaseBinToolKey) error {
	if err := cs.IStorage.DeleteReleaseBinTool(nCtx, key); err != nil {
		return err
	}

	cs.invalidate(nCtx, releaseGroupBinTool)

	return nil
}

// UpsertReleaseBinTool upserts release entries and invalidates the group marker.
func (cs *CachedStorage) UpsertReleaseBinTool(nCtx contextx.IContext, bintool types.ReleaseBinTool) error {
	if err := cs.IStorage.UpsertReleaseBinTool(nCtx, bintool); err != nil {
		return err
	}

	cs.invalidate(nCtx, releaseGroupBinTool)

	return nil
}

// DeleteReleasePluginBinTool deletes a plugin bintool release and invalidates the group marker.
func (cs *CachedStorage) DeleteReleasePluginBinTool(nCtx contextx.IContext, key types.ReleasePluginBinToolKey) error {
	if err := cs.IStorage.DeleteReleasePluginBinTool(nCtx, key); err != nil {
		return err
	}

	cs.invalidate(nCtx, releaseGroupPluginBinTool)

	return nil
}

// UpsertReleasePluginBinTool upserts release entries and invalidates the group marker.
func (cs *CachedStorage) UpsertReleasePluginBinTool(nCtx contextx.IContext, pluginBinTool types.ReleasePluginBinTool) error {
	if err := cs.IStorage.UpsertReleasePluginBinTool(nCtx, pluginBinTool); err != nil {
		return err
	}

	cs.invalidate(nCtx, releaseGroupPluginBinTool)

	return nil
}
