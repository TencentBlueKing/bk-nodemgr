/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package manager provides the file manager.
package manager

import (
	"context"
	"errors"
	"io"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/release"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/google/uuid"
)

// IManager defines the file manager interface.
type IManager interface {
	// Start starts the manager
	Start(ctx context.Context) error

	// UploadOriginAgent uploads the origin agent.
	UploadOriginAgent(ctx context.Context, gen types.Generation, pkgFile io.ReadCloser) (iface.FileInfo, error)

	// UploadOriginServer uploads the origin server.
	UploadOriginServer(ctx context.Context, gen types.Generation, pkgFile io.ReadCloser) (iface.FileInfo, error)
}

// New returns a new file manager.
func New(opts ...OptionFn) *Manager {
	manager := &Manager{
		logger: logger.LoggerDefault{},
	}

	for _, opt := range opts {
		opt(manager)
	}

	return manager
}

// OptionFn is an option function for ProviderEtcd.
type OptionFn func(manager *Manager)

// WithLogger sets the logger.
func WithLogger(logger logger.Logger) OptionFn {
	return func(manager *Manager) {
		manager.logger = logger
	}
}

// WithUpstreamOriginServerFileGroup sets the upstream file group.
func WithUpstreamOriginServerFileGroup(fileGroup iface.FileGroup) OptionFn {
	return func(manager *Manager) {
		manager.upstreamOriginServer = fileGroup
	}
}

// WithUpstreamOriginAgentFileGroup sets the upstream file group.
func WithUpstreamOriginAgentFileGroup(fileGroup iface.FileGroup) OptionFn {
	return func(manager *Manager) {
		manager.upstreamOriginAgent = fileGroup
	}
}

// WithLocalTempFileGroup sets the local file group.
func WithLocalTempFileGroup(fileGroup iface.FileGroup) OptionFn {
	return func(manager *Manager) {
		manager.temp = fileGroup
	}
}

// WithLocalFileGroup sets the local file group.
func WithLocalFileGroup(fileGroup iface.FileGroup) OptionFn {
	return func(manager *Manager) {
		manager.local = fileGroup
	}
}

// WithDaoRelease sets the dao for release.
func WithDaoRelease(daoRelease release.IHandler) OptionFn {
	return func(manager *Manager) {
		manager.daoRelease = daoRelease
	}
}

// Manager provides the file manager.
type Manager struct {
	// upstream file group is regarded as the file source.
	upstreamOriginAgent  iface.FileGroup
	upstreamOriginServer iface.FileGroup

	// local file group is regarded as the file cache.
	local iface.FileGroup

	// temp file group is regarded as the file temp.
	temp iface.FileGroup

	// dao for release.
	daoRelease release.IHandler

	// logger.
	logger logger.Logger
}

// Start starts the manager.
func (m *Manager) Start(_ context.Context) error {
	if m.daoRelease == nil {
		return errors.New("invalid dao release")
	}

	if m.upstreamOriginAgent == nil {
		return errors.New("invalid upstream origin agent")
	}

	if m.upstreamOriginServer == nil {
		return errors.New("invalid upstream origin server")
	}

	m.logger.Infof("started manager")

	return nil
}

func (m *Manager) saveTempFile(ctx context.Context, file io.ReadCloser) (string, error) {
	tempFileName := uuid.NewString() + ".tgz"

	err := m.temp.Store(ctx, iface.FileInfo{Name: tempFileName}, file, true)
	if err != nil {
		return "", err
	}

	return tempFileName, nil
}

func (m *Manager) getTempFile(ctx context.Context, tempFileName string) (io.ReadCloser, error) {
	fileToCheck, err := m.temp.GetFile(ctx, tempFileName)
	if err != nil {
		return nil, err
	}

	return fileToCheck.Content(ctx)
}
