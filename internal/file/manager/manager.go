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

// Package manager provides the file manager.
package manager

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/file/storage/packageevent"
	"github.com/TencentBlueKing/bk-nodemgr/internal/file/storage/packageexport"
	"github.com/TencentBlueKing/bk-nodemgr/internal/file/storage/release"
	"github.com/TencentBlueKing/bk-nodemgr/internal/file/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/internal/file/storage/upload"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/downloader"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/filecache"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/local"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/goasync"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/token"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/google/uuid"
)

const (
	asyncPoolNum  = 10
	asyncPoolSize = 100
)

// IManager defines the file manager interface.
// nolint: interfacebloat
type IManager interface {
	// Start starts the manager
	Start(ctx context.Context) error

	IInstaller
	IAgent
	IProxy
	IServer
	ICert
	IBinTool
	IPluginBinTool
	IPluginV2
	IExternalPluginV2
	IPluginV3
	IExport
	IPackageEvent

	// EnsureNodeToLocal ensure the node pkg to local.
	// returns file, local-file-dir, error.
	EnsureNodeToLocal(ctx contextx.IContext, rt types.ReleaseType, gen types.Generation, plat platfmt.Platform, version string) (
		fileiface.File, string, error)

	// EnsurePluginToLocal ensure the plugin pkg to local.
	// returns file, local-file-dir, error.
	EnsurePluginToLocal(ctx contextx.IContext, name string, gen types.Generation, plat platfmt.Platform, version string) (
		fileiface.File, string, error)

	// LaunchTransferNode launch transfer node pkg.
	LaunchTransferNode(ctx contextx.IContext,
		gen types.Generation,
		rt types.ReleaseType,
		plat platfmt.Platform,
		version string,
		dstDir string,
		dstHost *types.Host) (types.ISimpleTransferHandler, error)

	// LaunchTransferPlugin launch transfer plugin pkg.
	LaunchTransferPlugin(nCtx contextx.IContext,
		name string,
		gen types.Generation,
		plat platfmt.Platform,
		version string,
		dstDir string,
		dstHost *types.Host) (types.ISimpleTransferHandler, error)

	// LaunchTransferInstaller launch transfer installer.
	LaunchTransferInstaller(nCtx contextx.IContext,
		plat platfmt.Platform,
		dstDir string,
		dstHost *types.Host) (types.ISimpleTransferHandler, error)

	// QueryTransfer query transfer.
	// return upload result, download result and error.
	QueryTransfer(nCtx contextx.IContext, taskID string) (
		*types.SimpleTransferResult, *types.SimpleTransferResult, error)

	// DownloadRemoteFile downloads, verifies, and caches a remote file.
	DownloadRemoteFile(nCtx contextx.IContext, filename, downloadURL, expectedMD5 string) (fileiface.File, error)

	// SyncSharedReleases copies shared system releases to the target tenant.
	SyncSharedReleases(nCtx contextx.IContext) error
}

// New returns a new file manager.
func New(opts ...OptionFn) *Manager {
	goAsyncPool, _ := goasync.NewHandler(goasync.HandlerOption{
		PoolNum:               asyncPoolNum,
		PerPoolSize:           asyncPoolSize,
		LoadBalancingStrategy: goasync.LoadBalancingStrategyLeastFirst,
	})

	manager := &Manager{goAsyncPool: goAsyncPool}

	for _, opt := range opts {
		opt(manager)
	}

	return manager
}

// OptionFn is an option function for ProviderEtcd.
type OptionFn func(manager *Manager)

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

// WithUpstreamOriginProxyFileGroup sets the upstream file group.
func WithUpstreamOriginProxyFileGroup(fileGroup fileiface.FileGroup) OptionFn {
	return func(manager *Manager) {
		manager.upstreamOriginProxy = fileGroup
	}
}

// WithUpstreamOriginPluginV2FileGroup sets the upstream file group.
func WithUpstreamOriginPluginV2FileGroup(fileGroup fileiface.FileGroup) OptionFn {
	return func(manager *Manager) {
		manager.upstreamOriginPluginV2 = fileGroup
	}
}

// WithUpstreamOriginExternalPluginV2FileGroup sets the upstream file group.
func WithUpstreamOriginExternalPluginV2FileGroup(fileGroup fileiface.FileGroup) OptionFn {
	return func(manager *Manager) {
		manager.upstreamOriginExternalPluginV2 = fileGroup
	}
}

// WithUpstreamOriginPluginV3FileGroup sets the upstream file group.
func WithUpstreamOriginPluginV3FileGroup(fileGroup fileiface.FileGroup) OptionFn {
	return func(manager *Manager) {
		manager.upstreamOriginPluginV3 = fileGroup
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

// WithUpstreamOriginPluginBinToolV2FileGroup sets the upstream file group.
func WithUpstreamOriginPluginBinToolV2FileGroup(fileGroup fileiface.FileGroup) OptionFn {
	return func(manager *Manager) {
		manager.upstreamOriginPluginBinTool = fileGroup
	}
}

// WithUpstreamReleaseAgentFileGroup sets the upstream file group.
func WithUpstreamReleaseAgentFileGroup(fileGroup fileiface.FileGroup) OptionFn {
	return func(manager *Manager) {
		manager.upstreamReleaseAgent = fileGroup
	}
}

// WithUpstreamReleasePluginFileGroup sets the upstream file group.
func WithUpstreamReleasePluginFileGroup(fileGroup fileiface.FileGroup) OptionFn {
	return func(manager *Manager) {
		manager.upstreamReleasePlugin = fileGroup
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

// WithUpstreamReleasePluginBinToolFileGroup sets the upstream file group.
func WithUpstreamReleasePluginBinToolFileGroup(fileGroup fileiface.FileGroup) OptionFn {
	return func(manager *Manager) {
		manager.upstreamReleasePluginBinTool = fileGroup
	}
}

// WithTempFileGroup sets the temp file group.
func WithTempFileGroup(fileGroup fileiface.FileGroup) OptionFn {
	return func(manager *Manager) {
		manager.tempFileGroup = fileGroup
	}
}

// WithTempFileExpiration sets the idle duration after which a temp file becomes
// eligible for deletion by the temp file GC. A non-positive value disables the GC.
func WithTempFileExpiration(d time.Duration) OptionFn {
	return func(manager *Manager) {
		manager.tempFileExpiration = d
	}
}

// WithTempFileGCInterval sets the sleep period between two temp file GC runs.
// A non-positive value disables the GC.
func WithTempFileGCInterval(d time.Duration) OptionFn {
	return func(manager *Manager) {
		manager.tempFileGCInterval = d
	}
}

// WithInstallerFileGroup sets the installer file group.
func WithInstallerFileGroup(fileGroup fileiface.FileGroup) OptionFn {
	return func(manager *Manager) {
		manager.installerFileGroup = fileGroup
	}
}

// WithFileCache sets the IFileCache instance used by ensureReleaseToLocal.
func WithFileCache(fc filecache.IFileCache) OptionFn {
	return func(manager *Manager) {
		manager.fileCache = fc
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

// WithStorageEvent sets the storage for event.
func WithStorageEvent(storageEvent packageevent.IStorage) OptionFn {
	return func(manager *Manager) {
		manager.storageEvent = storageEvent
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

// WithMount sets the mount host dir.
func WithMount(mountHostDir, mountContainerDir string) OptionFn {
	return func(manager *Manager) {
		manager.mountHostDir = mountHostDir
		manager.mountContainerDir = mountContainerDir
	}
}

// WithGSEHandler sets the gse handler.
func WithGSEHandler(gseHander gse.IHandler) OptionFn {
	return func(manager *Manager) {
		manager.gseHandler = gseHander
	}
}

// WithDownloader sets the downloader used to fetch remote packages.
func WithDownloader(dl downloader.IHandler) OptionFn {
	return func(manager *Manager) {
		manager.downloader = dl
	}
}

// WithUpstreamExportFileGroup sets the upstream export file group.
func WithUpstreamExportFileGroup(fileGroup fileiface.FileGroup) OptionFn {
	return func(manager *Manager) {
		manager.upstreamExport = fileGroup
	}
}

// WithStoragePackageExport sets the storage for package export records.
func WithStoragePackageExport(storagePackageExport packageexport.IStorage) OptionFn {
	return func(manager *Manager) {
		manager.storagePackageExport = storagePackageExport
	}
}

// WithTokenGenerator sets the generator used to validate export download tokens.
func WithTokenGenerator(generator *token.Generator) OptionFn {
	return func(manager *Manager) {
		manager.tokenGenerator = generator
	}
}

// WithExportServerPublicBaseURL sets the public base URL used in download addresses.
func WithExportServerPublicBaseURL(baseURL *url.URL) OptionFn {
	return func(manager *Manager) {
		manager.exportServerPublicBaseURL = baseURL
	}
}

var _ IManager = &Manager{}

// Manager provides the file manager.
type Manager struct {
	// upstream file group is regarded as the file source.
	upstreamOriginAgent            fileiface.FileGroup
	upstreamOriginServer           fileiface.FileGroup
	upstreamOriginProxy            fileiface.FileGroup
	upstreamOriginCert             fileiface.FileGroup
	upstreamOriginBinTool          fileiface.FileGroup
	upstreamOriginPluginBinTool    fileiface.FileGroup
	upstreamOriginPluginV2         fileiface.FileGroup
	upstreamOriginExternalPluginV2 fileiface.FileGroup
	upstreamOriginPluginV3         fileiface.FileGroup
	upstreamReleaseAgent           fileiface.FileGroup
	upstreamReleaseProxy           fileiface.FileGroup
	upstreamReleaseCert            fileiface.FileGroup
	upstreamReleaseBinTool         fileiface.FileGroup
	upstreamReleasePluginBinTool   fileiface.FileGroup
	upstreamReleasePlugin          fileiface.FileGroup
	upstreamExport                 fileiface.FileGroup

	// tokenGenerator authenticates package export download tokens.
	tokenGenerator *token.Generator

	// exportServerPublicBaseURL is the public base URL used in download addresses.
	exportServerPublicBaseURL *url.URL

	// installter file group.
	installerFileGroup fileiface.FileGroup

	// temp file group is regarded as the file temp.
	tempFileGroup fileiface.FileGroup

	// tempFileExpiration is the idle duration after which a temp file becomes
	// eligible for deletion by the temp file GC.
	tempFileExpiration time.Duration

	// tempFileGCInterval is the sleep period between two temp file GC runs.
	tempFileGCInterval time.Duration

	// fileCache is the generic local file cache used by ensureReleaseToLocal.
	fileCache filecache.IFileCache

	// host inner ip.
	hostAdvertiseIPV4 string
	hostAdvertiseIPV6 string

	// in container mount settings.
	mountHostDir      string
	mountContainerDir string

	// gse handler.
	gseHandler gse.IHandler

	// downloader fetches remote packages with the configured security boundary.
	downloader downloader.IHandler

	// storages.
	storageUpload        upload.IStorage
	storageRelease       release.IStorage
	storageTopo          topo.IStorage
	storageEvent         packageevent.IStorage
	storagePackageExport packageexport.IStorage
	goAsyncPool          goasync.IHandler
}

// Start starts the manager.
// nolint:gocognit,gocyclo,cyclop,funlen
func (m *Manager) Start(ctx context.Context) error {
	if m.upstreamOriginAgent == nil {
		return errors.New("invalid upstream origin agent")
	}

	if m.upstreamOriginProxy == nil {
		return errors.New("invalid upstream origin proxy")
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

	if m.upstreamOriginPluginBinTool == nil {
		return errors.New("invalid upstream origin bin tool v2")
	}

	if m.upstreamOriginPluginV2 == nil {
		return errors.New("invalid upstream origin plugin v2")
	}

	if m.upstreamOriginExternalPluginV2 == nil {
		return errors.New("invalid upstream origin external plugin v2")
	}

	if m.upstreamOriginPluginV3 == nil {
		return errors.New("invalid upstream origin plugin v3")
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

	if m.upstreamReleasePlugin == nil {
		return errors.New("invalid upstream release plugin")
	}

	if m.upstreamReleasePluginBinTool == nil {
		return errors.New("invalid upstream release plugin bin tool v2")
	}

	if m.upstreamExport == nil {
		return errors.New("invalid upstream export")
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

	if m.storageEvent == nil {
		return errors.New("invalid storage event")
	}

	if m.storagePackageExport == nil {
		return errors.New("invalid storage package export")
	}

	if m.tokenGenerator == nil {
		return errors.New("invalid token generator")
	}

	if m.installerFileGroup == nil {
		return errors.New("invalid installer file group")
	}

	if m.tempFileGroup == nil {
		return errors.New("invalid temp file group")
	}

	if m.fileCache == nil {
		return errors.New("invalid file cache")
	}

	if m.gseHandler == nil {
		return errors.New("invalid gse handler")
	}

	if m.downloader == nil {
		return errors.New("invalid downloader")
	}

	if m.exportServerPublicBaseURL == nil {
		return errors.New("invalid export server public base URL")
	}

	// start temp file GC if configured. Disabled when either knob is non-positive.
	if m.tempFileExpiration > 0 && m.tempFileGCInterval > 0 {
		go m.runTempFileGC(ctx)
	}

	return nil
}

// tempFileSuffix is the filename suffix used by all temp files produced by
// saveTempFile / createTempFile. The temp file GC matches files by:
//   - filename has this suffix.
//   - the stem before the suffix parses as a UUID.
//
// Keep newTempFileName and isTempFileName in sync: changing the format requires
// updating both.
const tempFileSuffix = ".tgz"

// newTempFileName returns a fresh temp file name following the project convention.
func newTempFileName() string {
	return uuid.NewString() + tempFileSuffix
}

// isTempFileName reports whether name was produced by newTempFileName.
// Used by the temp file GC to avoid touching unrelated files in the temp dir.
func isTempFileName(name string) bool {
	stem, ok := strings.CutSuffix(name, tempFileSuffix)
	if !ok {
		return false
	}
	if _, err := uuid.Parse(stem); err != nil {
		return false
	}

	return true
}

func (m *Manager) saveTempFile(nCtx contextx.IContext, file io.ReadCloser) (string, error) {
	tempFileName := newTempFileName()

	err := m.tempFileGroup.Store(nCtx, fileiface.FileInfo{Name: tempFileName}, file, true)
	if err != nil {
		return "", err
	}

	return tempFileName, nil
}

func (m *Manager) getTempFile(nCtx contextx.IContext, tempFileName string) (io.ReadCloser, error) {
	fileToCheck, err := m.tempFileGroup.GetFile(nCtx, tempFileName)
	if err != nil {
		return nil, err
	}

	return fileToCheck.Content(nCtx)
}

func (m *Manager) createTempFile(nCtx contextx.IContext) (string, error) {
	return m.saveTempFile(nCtx, io.NopCloser(strings.NewReader("")))
}

func (m *Manager) openTempFile(_ contextx.IContext, tempFileName string) (io.ReadWriteCloser, error) {
	// nolint: gosec, mnd
	file, err := os.OpenFile(
		local.GetLocalFileGroupAbsFilePath(m.tempFileGroup, tempFileName),
		os.O_RDWR|os.O_TRUNC,
		0644,
	)
	if err != nil {
		return nil, err
	}

	return filex.OnceReadWriteCloser(file), nil
}

// runTempFileGC periodically scans the temp file directory and deletes entries that
//  1. match the project temp file naming convention (see isTempFileName), AND
//  2. have not been modified for at least tempFileExpiration.
//
// Modification time is used as a proxy for "last access" — temp files are write-once
// (UUID-named, never overwritten), and mtime is stable across filesystems that may
// mount with noatime. The GC swallows all errors as warnings: cleanup failures must
// never affect serving.
//
// The goroutine exits when ctx is canceled. Capability.Start passes svc.ctx, which
// is canceled by Service.GracefulShutdown.
func (m *Manager) runTempFileGC(ctx context.Context) {
	ticker := time.NewTicker(m.tempFileGCInterval)
	defer ticker.Stop()

	// run once immediately to clean up residue from a previous crash.
	m.gcTempFilesOnce()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.gcTempFilesOnce()
		}
	}
}

// gcTempFilesOnce performs a single GC sweep over the temp file directory.
func (m *Manager) gcTempFilesOnce() {
	dir := local.GetLocalFileGroupAbsDirPath(m.tempFileGroup)

	entries, err := os.ReadDir(dir)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			logger.G.Sys().WithErr(err).With("dir", dir).Warn("temp file GC: failed to read temp dir")
		}

		return
	}

	cutoff := time.Now().Add(-m.tempFileExpiration)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !isTempFileName(name) {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			// file may have been deleted between ReadDir and Info; skip silently.
			continue
		}
		if !info.ModTime().Before(cutoff) {
			continue
		}
		p := filepath.Join(dir, name)
		if err := os.Remove(p); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				logger.G.Sys().WithErr(err).With("file", p).Warn("temp file GC: failed to remove stale temp file")
			}

			continue
		}

		logger.G.Sys().With("file", p, "mtime", info.ModTime().Format(time.RFC3339)).
			Info("temp file GC: removed stale temp file")
	}
}

func (m *Manager) wrapOriginPackageName(name string) string {
	return name + "-" + time.Now().Format("0102150405")
}

func (m *Manager) fetchReleaseCertToLocal(ctx contextx.IContext) (_ fileiface.File, retErr error) {
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
	defer func() {
		if errClose := content.Close(); errClose != nil {
			retErr = errors.Join(retErr, errClose)
		}
	}()

	localFileName, err := m.saveTempFile(ctx, content)
	if err != nil {
		return nil, fmt.Errorf("failed to save release cert to temp file: %w", err)
	}

	return m.tempFileGroup.GetFile(ctx, localFileName)
}

func (m *Manager) fetchReleaseBinToolToLocal(ctx contextx.IContext) (_ fileiface.File, retErr error) {
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
	defer func() {
		if errClose := content.Close(); errClose != nil {
			retErr = errors.Join(retErr, errClose)
		}
	}()

	localFileName, err := m.saveTempFile(ctx, content)
	if err != nil {
		return nil, fmt.Errorf("failed to save release bintool to temp file: %w", err)
	}

	return m.tempFileGroup.GetFile(ctx, localFileName)
}

func (m *Manager) fetchReleasePluginBinToolToLocal(ctx contextx.IContext, name string) (_ fileiface.File, retErr error) {
	// get plugin bintool.
	pluginBinTool, err := m.storageRelease.GetReleasePluginBinTool(ctx, types.Generation2, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get release plugin bintool: %w", err)
	}

	file, err := m.upstreamReleasePluginBinTool.GetFile(ctx, pluginBinTool.FileName)
	if err != nil {
		return nil, fmt.Errorf("failed to get upstream release plugin bintool file: %w", err)
	}

	content, err := file.Content(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get upstream release plugin bintool content: %w", err)
	}
	defer func() {
		if errClose := content.Close(); errClose != nil {
			retErr = errors.Join(retErr, errClose)
		}
	}()

	localFileName, err := m.saveTempFile(ctx, content)
	if err != nil {
		return nil, fmt.Errorf("failed to save release plugin bintool to temp file: %w", err)
	}

	return m.tempFileGroup.GetFile(ctx, localFileName)
}

func (m *Manager) fetchReleaseAgentLocal(ctx contextx.IContext, plat platfmt.Platform, version string) (_ fileiface.File, retErr error) {
	// get agent.
	agent, err := m.storageRelease.GetReleaseAgent(ctx, types.Generation2, plat, version)
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
	defer func() {
		if errClose := content.Close(); errClose != nil {
			retErr = errors.Join(retErr, errClose)
		}
	}()

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
