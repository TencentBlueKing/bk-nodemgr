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
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/batchexecutor"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	// cleanupRetentionMonths is how long an unbound artifact is kept before deletion.
	cleanupRetentionMonths = 2

	cleanupBatchSize    = 200
	cleanupBatchTimeout = 5 * time.Minute

	cleanupParentName = ".."
)

// ICleanup provides on-demand cleanup of unbound package files.
type ICleanup interface {
	// CleanUploads removes old unbound upload files in one category for the request tenant.
	CleanUploads(nCtx contextx.IContext, category types.UploadCategory) error

	// CleanReleases removes old unbound release files of one type for the request tenant.
	CleanReleases(nCtx contextx.IContext, releaseType types.ReleaseType) error

	// CleanExports removes old unbound export files for the request tenant.
	CleanExports(nCtx contextx.IContext) error
}

// CleanUploads removes old unbound upload files in one category for the request tenant.
func (m *Manager) CleanUploads(nCtx contextx.IContext, category types.UploadCategory) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return fmt.Errorf("failed to clean uploads: %w", err)
	}

	var (
		group    fileiface.FileGroup
		bindings cleanupBindingsFn
	)

	switch category {
	case types.UploadCategoryOriginAgent:
		group = m.upstreamOriginAgent
		bindings = m.storageUpload.DistinctAgentSavedName
	case types.UploadCategoryOriginServer:
		group = m.upstreamOriginServer
		bindings = m.storageUpload.DistinctServerSavedName
	case types.UploadCategoryOriginProxy:
		group = m.upstreamOriginProxy
		bindings = m.storageUpload.DistinctProxySavedName
	case types.UploadCategoryOriginCert:
		group = m.upstreamOriginCert
		bindings = m.storageUpload.DistinctCertSavedName
	case types.UploadCategoryOriginBinTool:
		group = m.upstreamOriginBinTool
		bindings = m.storageUpload.DistinctBinToolSavedName
	case types.UploadCategoryOriginPluginBinTool:
		group = m.upstreamOriginPluginBinTool
		bindings = m.storageUpload.DistinctPluginBinToolSavedName
	case types.UploadCategoryOriginPluginV2:
		group = m.upstreamOriginPluginV2
		bindings = m.storageUpload.DistinctPluginV2SavedName
	case types.UploadCategoryOriginExternalPluginV2:
		group = m.upstreamOriginExternalPluginV2
		bindings = m.storageUpload.DistinctExternalPluginV2SavedName
	case types.UploadCategoryOriginPluginV3:
		group = m.upstreamOriginPluginV3
		bindings = m.storageUpload.DistinctPluginV3SavedName
	default:
		return fmt.Errorf("unsupported upload category: %q", category)
	}

	if err := m.cleanUnboundFiles(nCtx, group, bindings); err != nil {
		return fmt.Errorf("failed to clean uploads: %w", err)
	}

	return nil
}

// CleanReleases removes old unbound release files of one type for the request tenant.
func (m *Manager) CleanReleases(nCtx contextx.IContext, releaseType types.ReleaseType) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return fmt.Errorf("failed to clean releases: %w", err)
	}

	var (
		group    fileiface.FileGroup
		distinct func(contextx.IContext, types.ReleaseDistinctField, ...*types.ReleaseCondition) (*types.ReleaseDistinctResult, error)
	)
	switch releaseType {
	case types.ReleaseTypeAgent:
		group = m.upstreamReleaseAgent
		distinct = m.storageRelease.DistinctReleaseAgent
	case types.ReleaseTypeProxy:
		group = m.upstreamReleaseProxy
		distinct = m.storageRelease.DistinctReleaseProxy
	case types.ReleaseTypePlugin:
		group = m.upstreamReleasePlugin
		distinct = m.storageRelease.DistinctReleasePlugin
	default:
		return fmt.Errorf("unsupported release type: %q", releaseType)
	}

	bindings := func(nCtx contextx.IContext, names []string) ([]string, error) {
		fields := types.ReleaseDistinctField{FileName: true}
		condition := &types.ReleaseCondition{ExactInclude: &types.ReleaseExactFields{FileName: names}}

		result, err := distinct(nCtx, fields, condition)
		if err != nil {
			return nil, fmt.Errorf("failed to distinct %s release file names: %w", releaseType, err)
		}

		return result.FileName, nil
	}
	if err := m.cleanUnboundFiles(nCtx, group, bindings); err != nil {
		return fmt.Errorf("failed to clean releases: %w", err)
	}

	return nil
}

// CleanExports removes old unbound export files for the request tenant.
func (m *Manager) CleanExports(nCtx contextx.IContext) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return fmt.Errorf("failed to clean exports: %w", err)
	}

	bindings := func(nCtx contextx.IContext, names []string) ([]string, error) {
		return m.storagePackageExport.DistinctStorageKey(nCtx,
			&types.PackageExportCondition{ExactInclude: &types.PackageExportExactFields{
				TenantID: []string{nCtx.TenantID()}, StorageKey: names,
			}})
	}
	if err := m.cleanUnboundFiles(nCtx, m.upstreamExport, bindings); err != nil {
		return fmt.Errorf("failed to clean exports: %w", err)
	}

	return nil
}

// cleanupBindingsFn returns the subset of names still referenced by live records.
type cleanupBindingsFn func(nCtx contextx.IContext, names []string) ([]string, error)

// cleanUnboundFiles removes files older than the retention cutoff whose names have no live binding.
// nolint:gocognit // Linear enumerate-verify-remove flow; the complexity is the required safety guards.
func (m *Manager) cleanUnboundFiles(nCtx contextx.IContext, group fileiface.FileGroup, bindings cleanupBindingsFn) error {
	tenantID := nCtx.TenantID()
	groupPath := strings.Join(group.AbsDirs(), "/")
	cutoff := time.Now().AddDate(0, -cleanupRetentionMonths, 0)

	// Finish enumeration before deleting: offset pagination must not shift mid-scan.
	files, err := group.AllFiles(nCtx)
	if err != nil {
		return fmt.Errorf("failed to enumerate files, tenant-id(%s), group-path(%s): %w", tenantID, groupPath, err)
	}
	candidates := make([]fileiface.FileInfo, 0, len(files))
	for _, file := range files {
		info := file.Info()
		if !validCleanupFileInfo(info) {
			logger.G.Biz(nCtx).
				With("tenant-id", tenantID, "file-group", groupPath, "file", info.Name).
				Warn("artifact retained: unknown timestamp or invalid file name")

			continue
		}
		if info.ModTime.Before(cutoff) {
			candidates = append(candidates, info)
		}
	}

	err = batchexecutor.Execute(nCtx, candidates, func(nCtx contextx.IContext, batch []fileiface.FileInfo) error {
		names := conv.SliceToSlice(batch, func(file fileiface.FileInfo) string { return file.Name })
		references, err := bindings(nCtx, names)
		if err != nil {
			return fmt.Errorf("failed to verify bindings, tenant-id(%s), group-path(%s): %w", tenantID, groupPath, err)
		}
		for _, file := range batch {
			if slices.Contains(references, file.Name) {
				continue
			}

			// Verify the scanned file is still old and unchanged before removing it.
			current, err := group.GetFile(nCtx, file.Name)
			if err != nil {
				return fmt.Errorf("failed to get file, tenant-id(%s), group-path(%s), file-name(%s): %w", tenantID, groupPath, file.Name, err)
			}
			currentInfo := current.Info()
			if currentInfo.Name != file.Name || !validCleanupFileInfo(currentInfo) ||
				!currentInfo.ModTime.Equal(file.ModTime) || !currentInfo.ModTime.Before(cutoff) {

				continue
			}

			// Bindings are checked once per batch: re-uploads produce new artifact names and
			// same-name overwrites change ModTime (rejected above), so only an out-of-band
			// record change during the run could invalidate the check; that race is accepted.
			if err := group.Remove(nCtx, file.Name); err != nil {
				return fmt.Errorf("failed to remove file, tenant-id(%s), group-path(%s), file-name(%s): %w", tenantID, groupPath, file.Name, err)
			}
			logger.G.Biz(nCtx).
				With("tenant-id", tenantID, "file-group", groupPath, "file", file.Name, "mtime", file.ModTime).
				Info("removed unbound artifact")
		}

		return nil
	}, batchexecutor.WithBatchSize(cleanupBatchSize), batchexecutor.WithTimeout(cleanupBatchTimeout))
	if err != nil {
		return fmt.Errorf("failed to cleanup, tenant-id(%s), group-path(%s): %w", tenantID, groupPath, err)
	}

	return nil
}

// validCleanupFileInfo reports whether the name and timestamp are safe to evaluate for cleanup.
func validCleanupFileInfo(info fileiface.FileInfo) bool {
	return info.Name != "" && info.Name != "." &&
		info.Name != cleanupParentName &&
		!strings.ContainsAny(info.Name, "/\\") &&
		!info.ModTime.IsZero()
}
