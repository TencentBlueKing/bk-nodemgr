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
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/file/storage/release"
	"github.com/TencentBlueKing/bk-nodemgr/internal/file/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/internal/file/storage/upload"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/local"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/google/uuid"
)

// IManager defines the file manager interface.
// nolint: interfacebloat
type IManager interface {
	// Start starts the manager
	Start(ctx context.Context) error

	// UploadOriginAgent uploads the origin agent.
	UploadOriginAgent(ctx context.Context, pkgFile io.ReadCloser) (*types.OriginPkgDetail, error)

	// UploadOriginServer uploads the origin server.
	UploadOriginServer(ctx context.Context, pkgFile io.ReadCloser) (*types.OriginPkgDetail, error)

	// UploadOriginCert uploads the origin cert.
	UploadOriginCert(ctx context.Context, certFileName string, certFile io.ReadCloser) (*types.OriginCertPkgDetail, error)

	// UploadOriginBinTool upload origin bintool package.
	UploadOriginBinTool(ctx context.Context, binToolFile io.ReadCloser) (
		*types.OriginBinToolPkgDetail, error)

	// PublishReleaseAgent generates release agent by upload-id.
	PublishReleaseAgent(ctx context.Context, uploadID string) error

	// PublishReleaseProxy generates release proxy by upload-id.
	PublishReleaseProxy(ctx context.Context, uploadID string) error

	// PublishReleaseCert generates release cert by upload-id.
	PublishReleaseCert(ctx context.Context, uploadID string) error

	// PublishReleaseBinTool generate release bintool package.
	PublishReleaseBinTool(ctx context.Context, uploadID string) error

	// EnsureFileToLocal ensure the file to local.
	// returns file, local-file-dir, error.
	EnsureFileToLocal(ctx context.Context,
		gen types.Generation,
		rt types.ReleaseType,
		plat platform.Platform,
		version string) (fileiface.File, string, error)

	// EnsureReleaseToLocal ensure the release to local.
	// returns file, local-file-dir, error.
	EnsureReleaseToLocal(ctx context.Context, release *types.Release) (fileiface.File, string, error)

	// LaunchTransferRelease launch transfer release.
	LaunchTransferRelease(ctx context.Context,
		gen types.Generation,
		rt types.ReleaseType,
		plat platform.Platform,
		version string,
		dstDir string,
		dstHost *types.Host) (types.ISimpleTransferHandler, error)

	// LaunchTransferInstaller launch transfer installer.
	LaunchTransferInstaller(ctx context.Context,
		plat platform.Platform,
		dstDir string,
		dstHost *types.Host) (types.ISimpleTransferHandler, error)

	// QueryTransfer query transfer.
	// return upload result, download result and error.
	QueryTransfer(ctx context.Context, taskID string) (
		*types.SimpleTransferResult, *types.SimpleTransferResult, error)
}

// New returns a new file manager.
func New(opts ...OptionFn) *Manager {
	manager := &Manager{
		localFilePool: &localFilePool{
			files: map[string]*localFile{},
		},
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
func WithLogger(logger logger.ILogger) OptionFn {
	return func(manager *Manager) {
		manager.logger = logger
	}
}

// WithUpstreamOriginServerFileGroup sets the upstream file group.
func WithUpstreamOriginServerFileGroup(fileGroup fileiface.FileGroup) OptionFn {
	return func(manager *Manager) {
		manager.upstreamOriginServer = fileGroup
	}
}

// WithUpstreamOriginAgentFileGroup sets the upstream file group.
func WithUpstreamOriginAgentFileGroup(fileGroup fileiface.FileGroup) OptionFn {
	return func(manager *Manager) {
		manager.upstreamOriginAgent = fileGroup
	}
}

// WithUpstreamOriginCertFileGroup sets the upstream file group.
func WithUpstreamOriginCertFileGroup(fileGroup fileiface.FileGroup) OptionFn {
	return func(manager *Manager) {
		manager.upstreamOriginCert = fileGroup
	}
}

// WithUpstreamOriginBinToolFileGroup sets the upstream file group.
func WithUpstreamOriginBinToolFileGroup(fileGroup fileiface.FileGroup) OptionFn {
	return func(manager *Manager) {
		manager.upstreamOriginBinTool = fileGroup
	}
}

// WithUpstreamReleaseAgentFileGroup sets the upstream file group.
func WithUpstreamReleaseAgentFileGroup(fileGroup fileiface.FileGroup) OptionFn {
	return func(manager *Manager) {
		manager.upstreamReleaseAgent = fileGroup
	}
}

// WithUpstreamReleaseProxyFileGroup sets the upstream file group.
func WithUpstreamReleaseProxyFileGroup(fileGroup fileiface.FileGroup) OptionFn {
	return func(manager *Manager) {
		manager.upstreamReleaseProxy = fileGroup
	}
}

// WithUpstreamReleaseCertFileGroup sets the upstream file group.
func WithUpstreamReleaseCertFileGroup(fileGroup fileiface.FileGroup) OptionFn {
	return func(manager *Manager) {
		manager.upstreamReleaseCert = fileGroup
	}
}

// WithUpstreamReleaseBinToolFileGroup sets the upstream file group.
func WithUpstreamReleaseBinToolFileGroup(fileGroup fileiface.FileGroup) OptionFn {
	return func(manager *Manager) {
		manager.upstreamReleaseBinTool = fileGroup
	}
}

// WithTempFileGroup sets the temp file group.
func WithTempFileGroup(fileGroup fileiface.FileGroup) OptionFn {
	return func(manager *Manager) {
		manager.tempFileGroup = fileGroup
	}
}

// WithInstallerFileGroup sets the installer file group.
func WithInstallerFileGroup(fileGroup fileiface.FileGroup) OptionFn {
	return func(manager *Manager) {
		manager.installerFileGroup = fileGroup
	}
}

// WithCacheFileGroup sets the cache file group.
func WithCacheFileGroup(fileGroup fileiface.FileGroup) OptionFn {
	return func(manager *Manager) {
		manager.cacheFileGroup = fileGroup
	}
}

// WithStorageUpload sets the storage for upload.
func WithStorageUpload(storageUpload upload.IStorage) OptionFn {
	return func(manager *Manager) {
		manager.storageUpload = storageUpload
	}
}

// WithStorageRelease sets the storage for release.
func WithStorageRelease(storageRelease release.IStorage) OptionFn {
	return func(manager *Manager) {
		manager.storageRelease = storageRelease
	}
}

// WithStorageTopo sets the storage for topo.
func WithStorageTopo(storageTopo topo.IStorage) OptionFn {
	return func(manager *Manager) {
		manager.storageTopo = storageTopo
	}
}

// WithAdvertiseIPV4 sets the host advertise ipv4.
func WithAdvertiseIPV4(ipv4 string) OptionFn {
	return func(manager *Manager) {
		manager.hostAdvertiseIPV4 = ipv4
	}
}

// WithAdvertiseIPV6 sets the host advertise ipv4.
func WithAdvertiseIPV6(ipv6 string) OptionFn {
	return func(manager *Manager) {
		manager.hostAdvertiseIPV6 = ipv6
	}
}

// WithInContainer sets the in container.
func WithInContainer(inContainer bool) OptionFn {
	return func(manager *Manager) {
		manager.inContainer = inContainer
	}
}

// WithGSEHandler sets the gse handler.
func WithGSEHandler(gseHander gse.IHandler) OptionFn {
	return func(manager *Manager) {
		manager.gseHandler = gseHander
	}
}

// Manager provides the file manager.
type Manager struct {
	// upstream file group is regarded as the file source.
	upstreamOriginAgent    fileiface.FileGroup
	upstreamOriginServer   fileiface.FileGroup
	upstreamOriginCert     fileiface.FileGroup
	upstreamOriginBinTool  fileiface.FileGroup
	upstreamReleaseAgent   fileiface.FileGroup
	upstreamReleaseProxy   fileiface.FileGroup
	upstreamReleaseCert    fileiface.FileGroup
	upstreamReleaseBinTool fileiface.FileGroup

	// cache file group.
	cacheFileGroup fileiface.FileGroup

	// installter file group.
	installerFileGroup fileiface.FileGroup

	// temp file group is regarded as the file temp.
	tempFileGroup fileiface.FileGroup

	// local file pool.
	localFilePool *localFilePool

	// host inner ip.
	hostAdvertiseIPV4 string
	hostAdvertiseIPV6 string

	// in container.
	inContainer bool

	// gse handler.
	gseHandler gse.IHandler

	// storages.
	storageUpload  upload.IStorage
	storageRelease release.IStorage
	storageTopo    topo.IStorage

	// logger.
	logger logger.ILogger
}

// Start starts the manager.
func (m *Manager) Start(_ context.Context) error {
	if m.upstreamOriginAgent == nil {
		return errors.New("invalid upstream origin agent")
	}

	if m.upstreamOriginServer == nil {
		return errors.New("invalid upstream origin server")
	}

	if m.upstreamOriginCert == nil {
		return errors.New("invalid upstream origin cert")
	}

	if m.upstreamOriginBinTool == nil {
		return errors.New("invalid upstream origin bintool")
	}

	if m.upstreamReleaseAgent == nil {
		return errors.New("invalid upstream release agent")
	}

	if m.upstreamReleaseProxy == nil {
		return errors.New("invalid upstream release proxy")
	}

	if m.upstreamReleaseCert == nil {
		return errors.New("invalid upstream release cert")
	}

	if m.upstreamReleaseBinTool == nil {
		return errors.New("invalid upstream release bintool")
	}

	if m.storageUpload == nil {
		return errors.New("invalid storage upload")
	}

	if m.storageRelease == nil {
		return errors.New("invalid storage release")
	}

	if m.storageTopo == nil {
		return errors.New("invalid storage topo")
	}

	m.logger.Infof("started manager")

	return nil
}

func (m *Manager) saveTempFile(ctx context.Context, file io.ReadCloser) (string, error) {
	tempFileName := uuid.NewString() + ".tgz"

	err := m.tempFileGroup.Store(ctx, fileiface.FileInfo{Name: tempFileName}, file, true)
	if err != nil {
		return "", err
	}

	return tempFileName, nil
}

func (m *Manager) getTempFile(ctx context.Context, tempFileName string) (io.ReadCloser, error) {
	fileToCheck, err := m.tempFileGroup.GetFile(ctx, tempFileName)
	if err != nil {
		return nil, err
	}

	return fileToCheck.Content(ctx)
}

func (m *Manager) createTempFile(ctx context.Context) (string, error) {
	return m.saveTempFile(ctx, io.NopCloser(strings.NewReader("")))
}

func (m *Manager) openTempFile(_ context.Context, tempFileName string) (io.ReadWriteCloser, error) {
	// nolint: gosec, mnd
	return os.OpenFile(
		local.GetLocalFileGroupAbsFilePath(m.tempFileGroup, tempFileName),
		os.O_RDWR|os.O_TRUNC,
		0644,
	)
}

func (m *Manager) wrapOriginPackageName(name string) string {
	return name + "-" + time.Now().Format("0102150405")
}

func (m *Manager) fetchReleaseCertToLocal(ctx context.Context) (fileiface.File, error) {
	// get cert.
	cert, err := m.storageRelease.GetReleaseCert(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get release cert: %w", err)
	}

	file, err := m.upstreamReleaseCert.GetFile(ctx, cert.FileName)
	if err != nil {
		return nil, fmt.Errorf("failed to get upstream release cert file. err: %w", err)
	}

	content, err := file.Content(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get upstream release cert content: %w", err)
	}

	localFileName, err := m.saveTempFile(ctx, content)
	if err != nil {
		return nil, fmt.Errorf("failed to save release cert to temp file: %w", err)
	}

	return m.tempFileGroup.GetFile(ctx, localFileName)
}

func (m *Manager) fetchReleaseBinToolToLocal(ctx context.Context) (fileiface.File, error) {
	// get bintool.
	bintool, err := m.storageRelease.GetReleaseBinTool(ctx, types.Generation2)
	if err != nil {
		return nil, fmt.Errorf("failed to get release bintool: %w", err)
	}

	file, err := m.upstreamReleaseBinTool.GetFile(ctx, bintool.FileName)
	if err != nil {
		return nil, fmt.Errorf("failed to get upstream release bintool file: %w", err)
	}

	content, err := file.Content(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get upstream release bintool content: %w", err)
	}

	localFileName, err := m.saveTempFile(ctx, content)
	if err != nil {
		return nil, fmt.Errorf("failed to save release bintool to temp file: %w", err)
	}

	return m.tempFileGroup.GetFile(ctx, localFileName)
}

func (m *Manager) fetchReleaseAgentLocal(
	ctx context.Context, plat platform.Platform, version string) (fileiface.File, error) {

	// get agent.
	agent, err := m.storageRelease.GetRelease(ctx,
		types.Generation2, types.ReleaseTypeAgent, plat, version)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent release. platform(%s), version(%s): %w", plat.String(),
			version, err)
	}

	file, err := m.upstreamReleaseAgent.GetFile(ctx, agent.FileName)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent file from upstream. file(%s), platform(%s), version(%s): %w",
			agent.FileName, plat.String(), version, err)
	}

	content, err := file.Content(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get upstream release agent content: %w", err)
	}

	localFileName, err := m.saveTempFile(ctx, content)
	if err != nil {
		return nil, fmt.Errorf("failed to save release agent to temp file: %w", err)
	}

	localFile, err := m.tempFileGroup.GetFile(ctx, localFileName)
	if err != nil {
		return nil, fmt.Errorf("failed to get temp file: %w", err)
	}

	return localFile, nil
}

func parseEnvFile(r io.Reader) (map[string]any, error) {
	result := make(map[string]any)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if len(line) == 0 || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}

		index := strings.Index(line, "=")
		if index < 0 {
			continue
		}

		// key should be in format __key__
		key := "__" + strings.TrimSpace(line[:index]) + "__"
		value := strings.TrimSpace(line[index+1:])

		// get value according to the type.
		//
		// string value.
		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			result[key] = value[1 : len(value)-1]
			continue
		}

		// bool value.
		if value == "true" {
			result[key] = true
			continue
		}
		if value == "false" {
			result[key] = false
			continue
		}

		// integer value.
		if v, err := strconv.ParseInt(value, 10, 0); err == nil {
			result[key] = v
			continue
		}

		// default is string value.
		result[key] = value
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan env file: %w", err)
	}
	return result, nil
}
