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

// IPluginRequirement manages file resources independently of plugin declarations.
type IPluginRequirement interface {
	// CreatePluginRequirement stores a new dependency file and its record.
	CreatePluginRequirement(nCtx contextx.IContext, name, contentBase64 string) (*types.PluginRequirementFileRecord, error)
	// UpdatePluginRequirement replaces content while preserving the file record identity.
	UpdatePluginRequirement(nCtx contextx.IContext, id, name, contentBase64 string) (*types.PluginRequirementFileRecord, error)
	// GetPluginRequirement returns the current tenant's file record.
	GetPluginRequirement(nCtx contextx.IContext, id string) (*types.PluginRequirementFileRecord, error)
	// ReadPluginRequirement resolves the current content to a local cached file.
	ReadPluginRequirement(nCtx contextx.IContext, id string) (fileiface.File, error)
	// LaunchTransferPluginRequirement transfers a snapshot of the current file to a host.
	LaunchTransferPluginRequirement(nCtx contextx.IContext, id, targetName, targetDir string,
		targetHost *types.Host) (types.ISimpleTransferHandler, error)
}

// CreatePluginRequirement stores content before making its file record visible.
func (m *Manager) CreatePluginRequirement(nCtx contextx.IContext, name, contentBase64 string) (*types.PluginRequirementFileRecord, error) {
	if nCtx == nil {
		return nil, fmt.Errorf("context is nil")
	}
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, fmt.Errorf("invalid tenant: %w", err)
	}

	record, err := m.storePluginRequirement(nCtx, name, contentBase64)
	if err != nil {
		return nil, err
	}
	record.ID = uuid.NewString()
	if err := m.storagePluginRequirement.CreatePluginRequirement(nCtx, record); err != nil {
		return nil, fmt.Errorf("failed to create plugin requirement record: %w", err)
	}

	return record, nil
}

// UpdatePluginRequirement keeps the old record unchanged until the new object is stored.
func (m *Manager) UpdatePluginRequirement(nCtx contextx.IContext, id, name, contentBase64 string) (*types.PluginRequirementFileRecord, error) {
	if _, err := m.GetPluginRequirement(nCtx, id); err != nil {
		return nil, fmt.Errorf("failed to get plugin requirement before update: %w", err)
	}

	record, err := m.storePluginRequirement(nCtx, name, contentBase64)
	if err != nil {
		return nil, err
	}
	record.ID = id
	if err := m.storagePluginRequirement.UpdatePluginRequirement(nCtx, record); err != nil {
		return nil, fmt.Errorf("failed to update plugin requirement record: %w", err)
	}

	return record, nil
}

func (m *Manager) storePluginRequirement(nCtx contextx.IContext, name, contentBase64 string) (*types.PluginRequirementFileRecord, error) {
	if err := types.ValidatePluginRequirementFileName(name); err != nil {
		return nil, fmt.Errorf("invalid plugin requirement name: %w", err)
	}
	// Decode completely before writing so malformed input cannot leave a partial object.
	content, err := base64.StdEncoding.Strict().DecodeString(contentBase64)
	if err != nil {
		return nil, fmt.Errorf("invalid plugin requirement base64 content: %w", err)
	}

	savedName := uuid.NewString()
	reader := io.NopCloser(bytes.NewReader(content))
	defer func() { _ = reader.Close() }()
	if err := m.upstreamPluginRequirement.Store(nCtx, fileiface.FileInfo{Name: savedName}, reader, false); err != nil {
		return nil, fmt.Errorf("failed to store plugin requirement: %w", err)
	}
	file, err := m.upstreamPluginRequirement.GetFile(nCtx, savedName)
	if err != nil {
		return nil, fmt.Errorf("failed to get stored plugin requirement: %w", err)
	}
	info := file.Info()
	checksum := md5.Sum(content) // nolint:gosec // Match the existing file cache checksum contract.
	if info.MD5 != hex.EncodeToString(checksum[:]) || info.Size != int64(len(content)) {
		return nil, fmt.Errorf("stored plugin requirement metadata is invalid")
	}

	return &types.PluginRequirementFileRecord{
		Name: name, SavedName: savedName, MD5: info.MD5, Size: info.Size,
		UpdatedAt: time.Now(), Operator: nCtx.BKUsername(),
	}, nil
}

// GetPluginRequirement resolves a file record only within the current tenant.
func (m *Manager) GetPluginRequirement(nCtx contextx.IContext, id string) (*types.PluginRequirementFileRecord, error) {
	if nCtx == nil {
		return nil, fmt.Errorf("context is nil")
	}
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, fmt.Errorf("invalid tenant: %w", err)
	}
	if err := uuid.Validate(id); err != nil {
		return nil, fmt.Errorf("invalid plugin requirement id: %w", err)
	}

	record, err := m.storagePluginRequirement.GetPluginRequirement(nCtx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get plugin requirement: %w", err)
	}

	return record, nil
}

// ReadPluginRequirement returns the current file without exposing its storage path.
func (m *Manager) ReadPluginRequirement(nCtx contextx.IContext, id string) (fileiface.File, error) {
	record, err := m.GetPluginRequirement(nCtx, id)
	if err != nil {
		return nil, err
	}
	file, _, err := m.ensurePluginRequirementToLocal(nCtx, record, record.Name)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (m *Manager) ensurePluginRequirementToLocal(nCtx contextx.IContext, record *types.PluginRequirementFileRecord,
	name string) (fileiface.File, string, error) {

	if err := types.ValidatePluginRequirementFileName(name); err != nil {
		return nil, "", fmt.Errorf("invalid plugin requirement target name: %w", err)
	}
	fetch := func(ctx contextx.IContext) (io.ReadCloser, error) {
		file, err := m.upstreamPluginRequirement.GetFile(ctx, record.SavedName)
		if err != nil {
			return nil, fmt.Errorf("failed to get plugin requirement object: %w", err)
		}

		return file.Content(ctx)
	}

	return m.fileCache.GetOrFetch(nCtx, name, record.MD5, fetch)
}

// LaunchTransferPluginRequirement transfers a resolved snapshot using the existing GSE task lifecycle.
func (m *Manager) LaunchTransferPluginRequirement(nCtx contextx.IContext, id, targetName, targetDir string,
	targetHost *types.Host) (types.ISimpleTransferHandler, error) {

	if targetHost == nil || targetHost.Static == nil || targetHost.Dynamic == nil || targetDir == "" {
		return nil, fmt.Errorf("target host and target directory are required")
	}
	record, err := m.GetPluginRequirement(nCtx, id)
	if err != nil {
		return nil, err
	}
	if targetHost.TenantID != nCtx.TenantID() {
		return nil, fmt.Errorf("plugin requirement target host belongs to a different tenant")
	}
	file, dir, err := m.ensurePluginRequirementToLocal(nCtx, record, targetName)
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
