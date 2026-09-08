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
	"io"
	"path"
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
)

var _ fileiface.FileGroup = (*upstreamFileGroup)(nil)

type fileGroupEnsurer interface {
	EnsureFileGroup(nCtx contextx.IContext, groupPath string) (fileiface.FileGroup, error)
}

type upstreamFileGroup struct {
	ensurer  fileGroupEnsurer
	basePath string
	groups   sync.Map
}

type tenantFileGroup struct {
	mu    sync.Mutex
	group fileiface.FileGroup
}

// NewUpstreamFileGroup creates a tenant-aware BKRepo file group.
func NewUpstreamFileGroup(ensurer fileGroupEnsurer, basePath string) fileiface.FileGroup {
	return &upstreamFileGroup{
		ensurer:  ensurer,
		basePath: basePath,
	}
}

// Name returns the tenant-independent logical name of the file group.
func (group *upstreamFileGroup) Name() string {
	return path.Base(group.basePath)
}

// AbsDirs returns the tenant-independent logical directories of the file group.
func (group *upstreamFileGroup) AbsDirs() []string {
	return fileiface.ConvertAbsPathToAbsDirs(group.basePath)
}

// SubGroups returns the tenant's subgroups.
func (group *upstreamFileGroup) SubGroups(nCtx contextx.IContext) ([]fileiface.FileGroup, error) {
	tenantGroup, err := group.resolve(nCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve upstream file group, base-path(%s): %w", group.basePath, err)
	}

	return tenantGroup.SubGroups(nCtx)
}

// AllFiles returns all files in the tenant's file group.
func (group *upstreamFileGroup) AllFiles(nCtx contextx.IContext) ([]fileiface.File, error) {
	tenantGroup, err := group.resolve(nCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve upstream file group, base-path(%s): %w", group.basePath, err)
	}

	return tenantGroup.AllFiles(nCtx)
}

// GetFile returns a file from the tenant's file group.
func (group *upstreamFileGroup) GetFile(nCtx contextx.IContext, name string) (fileiface.File, error) {
	tenantGroup, err := group.resolve(nCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve upstream file group, base-path(%s): %w", group.basePath, err)
	}

	return tenantGroup.GetFile(nCtx, name)
}

// Store stores a file in the tenant's file group.
func (group *upstreamFileGroup) Store(nCtx contextx.IContext, info fileiface.FileInfo, reader io.ReadCloser, overwrite bool) error {
	tenantGroup, err := group.resolve(nCtx)
	if err != nil {
		return fmt.Errorf("failed to resolve upstream file group, base-path(%s): %w", group.basePath, err)
	}

	return tenantGroup.Store(nCtx, info, reader, overwrite)
}

func (group *upstreamFileGroup) resolve(nCtx contextx.IContext) (fileiface.FileGroup, error) {
	if nCtx == nil {
		return nil, errors.New("context is nil")
	}

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	tenantID := nCtx.TenantID()
	actual, _ := group.groups.LoadOrStore(tenantID, &tenantFileGroup{})
	entry, ok := actual.(*tenantFileGroup)
	if !ok {
		return nil, fmt.Errorf("invalid upstream file group cache entry for tenant %s: %T", tenantID, actual)
	}

	entry.mu.Lock()
	defer entry.mu.Unlock()

	if entry.group != nil {
		return entry.group, nil
	}
	if err := nCtx.Err(); err != nil {
		return nil, err
	}

	tenantGroup, err := group.ensurer.EnsureFileGroup(nCtx, path.Join("/", tenantID, group.basePath))
	if err != nil {
		return nil, err
	}

	entry.group = tenantGroup

	return entry.group, nil
}
