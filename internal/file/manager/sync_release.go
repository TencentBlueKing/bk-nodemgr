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

package manager

import (
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// SyncSharedReleases copies all shared system releases to the target tenant.
func (m *Manager) SyncSharedReleases(nCtx contextx.IContext) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}
	if nCtx.TenantID() == tenant.SystemTenantID {
		return errors.New("system tenant cannot synchronize shared releases")
	}

	syncStages := []func(contextx.IContext) error{
		m.syncSharedReleaseCert,
		m.syncSharedReleaseBinTool,
		m.syncSharedReleasePluginBinTool,
		m.syncSharedReleaseAgent,
		m.syncSharedReleaseProxy,
		m.syncSharedReleasePlugin,
	}
	syncErrs := make([]error, 0)
	for _, syncStage := range syncStages {
		if err := nCtx.Err(); err != nil {
			return err
		}
		if err := syncStage(nCtx); err != nil {
			syncErrs = append(syncErrs, err)
		}
	}

	return errors.Join(syncErrs...)
}

func (m *Manager) syncSharedReleaseCert(nCtx contextx.IContext) error {
	systemCtx := contextx.From(nCtx, contextx.WithTenantID(tenant.SystemTenantID))
	releases, _, err := m.storageRelease.ListReleaseCert(systemCtx, types.UnlimitedPage(),
		&types.ReleaseCondition{ExactInclude: &types.ReleaseExactFields{IsShared: []bool{true}}})
	if err != nil {
		return fmt.Errorf("failed to list shared cert releases: %w", err)
	}
	syncErrs := make([]error, 0)
	for _, source := range releases {
		if err := nCtx.Err(); err != nil {
			return err
		}
		if err := m.syncSharedRelease(nCtx, &source.Release, func(release types.Release) error {
			synced := *source
			synced.Release = release

			return m.storageRelease.UpsertReleaseCert(nCtx, synced)
		}); err != nil {
			syncErrs = append(syncErrs, err)
		}
	}

	return errors.Join(syncErrs...)
}

func (m *Manager) syncSharedReleaseBinTool(nCtx contextx.IContext) error {
	systemCtx := contextx.From(nCtx, contextx.WithTenantID(tenant.SystemTenantID))
	releases, _, err := m.storageRelease.ListReleaseBinTool(systemCtx, types.UnlimitedPage(),
		&types.ReleaseCondition{ExactInclude: &types.ReleaseExactFields{IsShared: []bool{true}}})
	if err != nil {
		return fmt.Errorf("failed to list shared bintool releases: %w", err)
	}
	syncErrs := make([]error, 0)
	for _, source := range releases {
		if err := nCtx.Err(); err != nil {
			return err
		}
		if err := m.syncSharedRelease(nCtx, &source.Release, func(release types.Release) error {
			synced := *source
			synced.Release = release

			return m.storageRelease.UpsertReleaseBinTool(nCtx, synced)
		}); err != nil {
			syncErrs = append(syncErrs, err)
		}
	}

	return errors.Join(syncErrs...)
}

func (m *Manager) syncSharedReleasePluginBinTool(nCtx contextx.IContext) error {
	systemCtx := contextx.From(nCtx, contextx.WithTenantID(tenant.SystemTenantID))
	releases, _, err := m.storageRelease.ListReleasePluginBinTool(systemCtx, types.UnlimitedPage(),
		&types.ReleaseCondition{ExactInclude: &types.ReleaseExactFields{IsShared: []bool{true}}})
	if err != nil {
		return fmt.Errorf("failed to list shared pluginbintool releases: %w", err)
	}
	syncErrs := make([]error, 0)
	for _, source := range releases {
		if err := nCtx.Err(); err != nil {
			return err
		}
		if err := m.syncSharedRelease(nCtx, &source.Release, func(release types.Release) error {
			synced := *source
			synced.Release = release

			return m.storageRelease.UpsertReleasePluginBinTool(nCtx, synced)
		}); err != nil {
			syncErrs = append(syncErrs, err)
		}
	}

	return errors.Join(syncErrs...)
}

func (m *Manager) syncSharedReleaseAgent(nCtx contextx.IContext) error {
	systemCtx := contextx.From(nCtx, contextx.WithTenantID(tenant.SystemTenantID))
	releases, _, err := m.storageRelease.ListReleaseAgent(systemCtx, types.UnlimitedPage(),
		&types.ReleaseCondition{ExactInclude: &types.ReleaseExactFields{IsShared: []bool{true}}})
	if err != nil {
		return fmt.Errorf("failed to list shared agent releases: %w", err)
	}
	syncErrs := make([]error, 0)
	for _, source := range releases {
		if err := nCtx.Err(); err != nil {
			return err
		}
		if err := m.syncSharedRelease(nCtx, &source.Release, func(release types.Release) error {
			synced := *source
			synced.Release = release

			return m.storageRelease.UpsertManyReleaseAgent(nCtx, []*types.ReleaseAgent{&synced})
		}); err != nil {
			syncErrs = append(syncErrs, err)
		}
	}

	return errors.Join(syncErrs...)
}

func (m *Manager) syncSharedReleaseProxy(nCtx contextx.IContext) error {
	systemCtx := contextx.From(nCtx, contextx.WithTenantID(tenant.SystemTenantID))
	releases, _, err := m.storageRelease.ListReleaseProxy(systemCtx, types.UnlimitedPage(),
		&types.ReleaseCondition{ExactInclude: &types.ReleaseExactFields{IsShared: []bool{true}}})
	if err != nil {
		return fmt.Errorf("failed to list shared proxy releases: %w", err)
	}
	syncErrs := make([]error, 0)
	for _, source := range releases {
		if err := nCtx.Err(); err != nil {
			return err
		}
		if err := m.syncSharedRelease(nCtx, &source.Release, func(release types.Release) error {
			synced := *source
			synced.Release = release

			return m.storageRelease.UpsertManyReleaseProxy(nCtx, []*types.ReleaseProxy{&synced})
		}); err != nil {
			syncErrs = append(syncErrs, err)
		}
	}

	return errors.Join(syncErrs...)
}

func (m *Manager) syncSharedReleasePlugin(nCtx contextx.IContext) error {
	systemCtx := contextx.From(nCtx, contextx.WithTenantID(tenant.SystemTenantID))
	releases, _, err := m.storageRelease.ListReleasePlugin(systemCtx, types.UnlimitedPage(),
		&types.ReleaseCondition{ExactInclude: &types.ReleaseExactFields{IsShared: []bool{true}}})
	if err != nil {
		return fmt.Errorf("failed to list shared plugin releases: %w", err)
	}
	syncErrs := make([]error, 0)
	for _, source := range releases {
		if err := nCtx.Err(); err != nil {
			return err
		}
		if err := m.syncSharedRelease(nCtx, &source.Release, func(release types.Release) error {
			synced := *source
			synced.Release = release

			return m.storageRelease.UpsertManyReleasePlugin(nCtx, []*types.ReleasePlugin{&synced})
		}); err != nil {
			syncErrs = append(syncErrs, err)
		}
	}

	return errors.Join(syncErrs...)
}

func (m *Manager) syncSharedRelease(nCtx contextx.IContext, source *types.Release, save func(types.Release) error) error {
	target, existed, err := m.getSyncTargetRelease(nCtx, source)
	if err != nil {
		return fmt.Errorf("failed to get target release(%s): %w", source.FileName, err)
	}

	// Preserve tenant-owned releases and skip synchronized releases with unchanged content.
	if existed && (!target.IsSynced || target.MD5 == source.MD5) {
		return nil
	}

	syncedRelease := *source
	if !existed {
		syncedRelease.AsDefault, err = m.shouldSyncReleaseAsDefault(nCtx, source)
		if err != nil {
			return fmt.Errorf("failed to check target default release(%s): %w", source.FileName, err)
		}
	}
	if existed {
		syncedRelease = *target
		syncedRelease.FileName = source.FileName
		syncedRelease.MD5 = source.MD5
		syncedRelease.AdditionInfo = source.AdditionInfo
	}
	syncedRelease.IsShared = false
	syncedRelease.IsSynced = true

	group, err := m.releaseFileGroup(source.Type)
	if err != nil {
		return err
	}

	if err := group.Copy(nCtx, source.FileName, group, ".", true); err != nil {
		return fmt.Errorf("failed to copy shared release(%s): %w", source.FileName, err)
	}

	if err := save(syncedRelease); err != nil {
		return fmt.Errorf("failed to save synchronized release(%s): %w", source.FileName, err)
	}

	return nil
}

func (m *Manager) shouldSyncReleaseAsDefault(nCtx contextx.IContext, source *types.Release) (bool, error) {
	if !source.AsDefault {
		return false, nil
	}

	condition := &types.ReleaseCondition{ExactInclude: &types.ReleaseExactFields{
		Name:       []string{source.Name},
		Generation: []types.Generation{source.Generation},
		Platform:   []platfmt.Platform{source.Platform},
		AsDefault:  []bool{true},
	}}
	var (
		count int64
		err   error
	)
	switch source.Type {
	case types.ReleaseTypeAgent:
		count, err = m.storageRelease.CountReleaseAgent(nCtx, condition)
	case types.ReleaseTypeProxy:
		count, err = m.storageRelease.CountReleaseProxy(nCtx, condition)
	case types.ReleaseTypePlugin:
		count, err = m.storageRelease.CountReleasePlugin(nCtx, condition)
	default:
		// Cert and tool releases have a single record per package key, checked before this call.
		return true, nil
	}
	if err != nil {
		return false, err
	}

	return count == 0, nil
}

// getSyncTargetRelease returns a nil release and false when the target does not exist.
// nolint: gocognit,gocyclo,cyclop
func (m *Manager) getSyncTargetRelease(nCtx contextx.IContext, source *types.Release) (*types.Release, bool, error) {
	switch source.Type {
	case types.ReleaseTypeAgent:
		existed, err := m.storageRelease.ExistReleaseAgent(nCtx, source.Generation, source.Version, source.Platform)
		if err != nil {
			return nil, false, err
		}
		if !existed {
			return nil, false, nil
		}

		release, err := m.storageRelease.GetReleaseAgent(nCtx, source.Generation, source.Platform, source.Version)
		if err != nil {
			return nil, false, err
		}

		return &release.Release, true, nil
	case types.ReleaseTypeProxy:
		existed, err := m.storageRelease.ExistReleaseProxy(nCtx, source.Generation, source.Version, source.Platform)
		if err != nil {
			return nil, false, err
		}
		if !existed {
			return nil, false, nil
		}

		release, err := m.storageRelease.GetReleaseProxy(nCtx, source.Generation, source.Platform, source.Version)
		if err != nil {
			return nil, false, err
		}

		return &release.Release, true, nil
	case types.ReleaseTypePlugin:
		existed, err := m.storageRelease.ExistReleasePlugin(nCtx, source.Name, source.Version, source.Platform)
		if err != nil {
			return nil, false, err
		}
		if !existed {
			return nil, false, nil
		}

		release, err := m.storageRelease.GetReleasePlugin(nCtx, source.Name, source.Generation, source.Platform, source.Version)
		if err != nil {
			return nil, false, err
		}

		return &release.Release, true, nil
	case types.ReleaseTypeCert:
		existed, err := m.storageRelease.ExistReleaseCert(nCtx)
		if err != nil {
			return nil, false, err
		}
		if !existed {
			return nil, false, nil
		}

		release, err := m.storageRelease.GetReleaseCert(nCtx)
		if err != nil {
			return nil, false, err
		}

		return &release.Release, true, nil
	case types.ReleaseTypeBinTool:
		existed, err := m.storageRelease.ExistReleaseBinTool(nCtx, source.Generation)
		if err != nil {
			return nil, false, err
		}
		if !existed {
			return nil, false, nil
		}

		release, err := m.storageRelease.GetReleaseBinTool(nCtx, source.Generation)
		if err != nil {
			return nil, false, err
		}

		return &release.Release, true, nil
	case types.ReleaseTypePluginBinTool:
		existed, err := m.storageRelease.ExistReleasePluginBinTool(nCtx, source.Generation, source.Name)
		if err != nil {
			return nil, false, err
		}
		if !existed {
			return nil, false, nil
		}

		release, err := m.storageRelease.GetReleasePluginBinTool(nCtx, source.Generation, source.Name)
		if err != nil {
			return nil, false, err
		}

		return &release.Release, true, nil
	default:
		return nil, false, fmt.Errorf("unsupported shared release type: %s", source.Type)
	}
}

func (m *Manager) releaseFileGroup(releaseType types.ReleaseType) (fileiface.FileGroup, error) {
	switch releaseType {
	case types.ReleaseTypeAgent:
		return m.upstreamReleaseAgent, nil
	case types.ReleaseTypeProxy:
		return m.upstreamReleaseProxy, nil
	case types.ReleaseTypePlugin:
		return m.upstreamReleasePlugin, nil
	case types.ReleaseTypeCert:
		return m.upstreamReleaseCert, nil
	case types.ReleaseTypeBinTool:
		return m.upstreamReleaseBinTool, nil
	case types.ReleaseTypePluginBinTool:
		return m.upstreamReleasePluginBinTool, nil
	default:
		return nil, fmt.Errorf("unsupported shared release type: %s", releaseType)
	}
}
