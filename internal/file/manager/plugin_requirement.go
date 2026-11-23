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
	"bytes"
	"crypto/md5" // nolint:gosec // File integrity checksum, not a security primitive.
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/google/uuid"
)

// IPluginRequirement manages plugin-associated dependency files independently of deployment declarations.
type IPluginRequirement interface {
	// CreatePluginRequirement stores a named plugin dependency and its applicability.
	CreatePluginRequirement(nCtx contextx.IContext, key types.PluginRequirementFileKey,
		targetFileName, contentBase64 string, scope types.PluginRequirementFileScope) (*types.PluginRequirementFileRecord, error)
	// UpdatePluginRequirementScope changes applicability without accessing file content.
	UpdatePluginRequirementScope(nCtx contextx.IContext, requirementFileID string,
		scope types.PluginRequirementFileScope) (*types.PluginRequirementFileRecord, error)
	// ListPluginRequirements lists administrative file metadata with pagination.
	ListPluginRequirements(nCtx contextx.IContext, page types.Page,
		condition *types.PluginRequirementFileCondition) ([]*types.PluginRequirementFileRecord, int64, error)
	// UpdatePluginRequirement replaces content and names without changing plugin ownership or applicability.
	UpdatePluginRequirement(nCtx contextx.IContext, requirementFileID, name, targetFileName,
		contentBase64 string) (*types.PluginRequirementFileRecord, error)
	// GetPluginRequirement returns the current tenant's file record.
	GetPluginRequirement(nCtx contextx.IContext, requirementFileID string) (*types.PluginRequirementFileRecord, error)
	// ReadPluginRequirement resolves the current content to a local cached file.
	ReadPluginRequirement(nCtx contextx.IContext, requirementFileID string) (fileiface.File, error)
	// LaunchTransferPluginRequirement transfers the file using its registered filename.
	LaunchTransferPluginRequirement(nCtx contextx.IContext, requirementFileID, targetDir string,
		targetHost *types.Host) (types.ISimpleTransferHandler, error)
}

// CreatePluginRequirement stores content before making its named plugin dependency visible.
func (m *Manager) CreatePluginRequirement(nCtx contextx.IContext, key types.PluginRequirementFileKey,
	targetFileName, contentBase64 string, scope types.PluginRequirementFileScope) (*types.PluginRequirementFileRecord, error) {

	if nCtx == nil {
		return nil, fmt.Errorf("context is nil")
	}
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, fmt.Errorf("invalid tenant: %w", err)
	}
	if err := key.Validate(); err != nil {
		return nil, fmt.Errorf("invalid plugin requirement key: %w", err)
	}
	if err := scope.Validate(); err != nil {
		return nil, fmt.Errorf("invalid plugin requirement scope: %w", err)
	}

	record, err := m.storePluginRequirement(nCtx, targetFileName, contentBase64)
	if err != nil {
		return nil, err
	}
	record.RequirementFileID = uuid.NewString()
	record.PluginName, record.Name = key.PluginName, key.Name
	record.OS, record.Arch, record.Version = scope.OS, scope.Arch, scope.Version
	if err := m.storagePluginRequirement.CreatePluginRequirement(nCtx, record); err != nil {
		return nil, fmt.Errorf("failed to create plugin requirement record: %w", err)
	}

	return record, nil
}

// UpdatePluginRequirement keeps the old record unchanged until the new object is stored.
func (m *Manager) UpdatePluginRequirement(nCtx contextx.IContext, requirementFileID, name, targetFileName,
	contentBase64 string) (*types.PluginRequirementFileRecord, error) {

	if err := types.ValidatePluginRequirementName(name); err != nil {
		return nil, fmt.Errorf("invalid plugin requirement name: %w", err)
	}
	if _, err := m.GetPluginRequirement(nCtx, requirementFileID); err != nil {
		return nil, fmt.Errorf("failed to get plugin requirement before update: %w", err)
	}

	record, err := m.storePluginRequirement(nCtx, targetFileName, contentBase64)
	if err != nil {
		return nil, err
	}
	record.RequirementFileID, record.Name = requirementFileID, name
	if err := m.storagePluginRequirement.UpdatePluginRequirement(nCtx, record); err != nil {
		return nil, fmt.Errorf("failed to update plugin requirement record: %w", err)
	}

	return m.GetPluginRequirement(nCtx, requirementFileID)
}

// UpdatePluginRequirementScope changes only applicability and preserves the physical file reference.
func (m *Manager) UpdatePluginRequirementScope(nCtx contextx.IContext, requirementFileID string,
	scope types.PluginRequirementFileScope) (*types.PluginRequirementFileRecord, error) {

	if nCtx == nil {
		return nil, fmt.Errorf("context is nil")
	}
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, fmt.Errorf("invalid tenant: %w", err)
	}
	if err := uuid.Validate(requirementFileID); err != nil {
		return nil, fmt.Errorf("invalid requirement_file_id: %w", err)
	}
	if err := scope.Validate(); err != nil {
		return nil, fmt.Errorf("invalid plugin requirement scope: %w", err)
	}
	record := &types.PluginRequirementFileRecord{
		RequirementFileID: requirementFileID, OS: scope.OS, Arch: scope.Arch, Version: scope.Version,
		UpdatedAt: time.Now(), Operator: nCtx.BKUsername(),
	}
	if err := m.storagePluginRequirement.UpdatePluginRequirementScope(nCtx, record); err != nil {
		return nil, fmt.Errorf("failed to update plugin requirement scope: %w", err)
	}

	return m.GetPluginRequirement(nCtx, requirementFileID)
}

// ListPluginRequirements returns administrative matches without applicability expansion.
func (m *Manager) ListPluginRequirements(nCtx contextx.IContext, page types.Page,
	condition *types.PluginRequirementFileCondition) ([]*types.PluginRequirementFileRecord, int64, error) {

	if nCtx == nil {
		return nil, 0, fmt.Errorf("context is nil")
	}
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, 0, fmt.Errorf("invalid tenant: %w", err)
	}
	if err := page.Validate(); err != nil {
		return nil, 0, fmt.Errorf("invalid plugin requirement page: %w", err)
	}
	records, total, err := m.storagePluginRequirement.ListPluginRequirements(nCtx, page, condition)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list plugin requirements: %w", err)
	}

	return records, total, nil
}

func (m *Manager) storePluginRequirement(nCtx contextx.IContext, targetFileName, contentBase64 string) (*types.PluginRequirementFileRecord, error) {
	if err := types.ValidatePluginRequirementFileName(targetFileName); err != nil {
		return nil, fmt.Errorf("invalid plugin requirement filename: %w", err)
	}
	// Decode completely before writing so malformed input cannot leave a partial object.
	content, err := base64.StdEncoding.Strict().DecodeString(contentBase64)
	if err != nil {
		return nil, fmt.Errorf("invalid plugin requirement base64 content: %w", err)
	}

	fileName := uuid.NewString()
	reader := io.NopCloser(bytes.NewReader(content))
	defer func() { _ = reader.Close() }()
	if err := m.upstreamPluginRequirement.Store(nCtx, fileiface.FileInfo{Name: fileName}, reader, false); err != nil {
		return nil, fmt.Errorf("failed to store plugin requirement: %w", err)
	}
	file, err := m.upstreamPluginRequirement.GetFile(nCtx, fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to get stored plugin requirement: %w", err)
	}
	info := file.Info()
	checksum := md5.Sum(content) // nolint:gosec // Match the existing file cache checksum contract.
	if info.MD5 != hex.EncodeToString(checksum[:]) || info.Size != int64(len(content)) {
		return nil, fmt.Errorf("stored plugin requirement metadata is invalid")
	}

	return &types.PluginRequirementFileRecord{
		FileName: fileName, TargetFileName: targetFileName, MD5: info.MD5, Size: info.Size,
		UpdatedAt: time.Now(), Operator: nCtx.BKUsername(),
	}, nil
}

// GetPluginRequirement resolves a file record only within the current tenant.
func (m *Manager) GetPluginRequirement(nCtx contextx.IContext, requirementFileID string) (*types.PluginRequirementFileRecord, error) {
	if nCtx == nil {
		return nil, fmt.Errorf("context is nil")
	}
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, fmt.Errorf("invalid tenant: %w", err)
	}
	if err := uuid.Validate(requirementFileID); err != nil {
		return nil, fmt.Errorf("invalid requirement_file_id: %w", err)
	}

	record, err := m.storagePluginRequirement.GetPluginRequirement(nCtx, requirementFileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get plugin requirement: %w", err)
	}

	return record, nil
}

// ReadPluginRequirement returns the target filename without exposing its storage filename.
func (m *Manager) ReadPluginRequirement(nCtx contextx.IContext, requirementFileID string) (fileiface.File, error) {
	record, err := m.GetPluginRequirement(nCtx, requirementFileID)
	if err != nil {
		return nil, err
	}
	file, _, err := m.ensurePluginRequirementToLocal(nCtx, record)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (m *Manager) ensurePluginRequirementToLocal(nCtx contextx.IContext,
	record *types.PluginRequirementFileRecord) (fileiface.File, string, error) {

	if err := types.ValidatePluginRequirementFileName(record.TargetFileName); err != nil {
		return nil, "", fmt.Errorf("invalid plugin requirement filename: %w", err)
	}
	fetch := func(ctx contextx.IContext) (io.ReadCloser, error) {
		file, err := m.upstreamPluginRequirement.GetFile(ctx, record.FileName)
		if err != nil {
			return nil, fmt.Errorf("failed to get plugin requirement object: %w", err)
		}

		return file.Content(ctx)
	}

	return m.fileCache.GetOrFetch(nCtx, record.TargetFileName, record.MD5, fetch)
}

// LaunchTransferPluginRequirement transfers a resolved snapshot using its registered filename.
func (m *Manager) LaunchTransferPluginRequirement(nCtx contextx.IContext, requirementFileID, targetDir string,
	targetHost *types.Host) (types.ISimpleTransferHandler, error) {

	if targetHost == nil || targetHost.Static == nil || targetHost.Dynamic == nil || targetDir == "" {
		return nil, fmt.Errorf("target host and target directory are required")
	}
	record, err := m.GetPluginRequirement(nCtx, requirementFileID)
	if err != nil {
		return nil, err
	}
	if targetHost.TenantID != nCtx.TenantID() {
		return nil, fmt.Errorf("plugin requirement target host belongs to a different tenant")
	}
	file, dir, err := m.ensurePluginRequirementToLocal(nCtx, record)
	if err != nil {
		return nil, fmt.Errorf("failed to cache plugin requirement: %w", err)
	}
	tf, err := m.transferPkg(nCtx, filepath.Join(dir, file.Info().Name), targetDir, targetHost)
	if err != nil {
		return nil, fmt.Errorf("failed to transfer plugin requirement: %w", err)
	}
	tf.fileInfo = file.Info()

	return tf, nil
}
