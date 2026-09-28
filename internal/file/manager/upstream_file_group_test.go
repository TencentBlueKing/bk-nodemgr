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
	"context"
	"errors"
	"io"
	"path"
	"strings"
	"sync"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
	"github.com/stretchr/testify/require"
)

type fakeUpstreamHandler struct {
	mu           sync.Mutex
	ensurePaths  []string
	ensureErrs   []error
	nextID       int
	beforeEnsure func(string)
}

func (handler *fakeUpstreamHandler) EnsureFileGroup(_ contextx.IContext, groupPath string) (fileiface.FileGroup, error) {
	if handler.beforeEnsure != nil {
		handler.beforeEnsure(groupPath)
	}
	handler.mu.Lock()
	defer handler.mu.Unlock()

	handler.ensurePaths = append(handler.ensurePaths, groupPath)
	if len(handler.ensureErrs) > 0 {
		err := handler.ensureErrs[0]
		handler.ensureErrs = handler.ensureErrs[1:]
		if err != nil {
			return nil, err
		}
	}

	handler.nextID++
	return &fakeTenantFileGroup{id: handler.nextID, name: groupPath}, nil
}

func (handler *fakeUpstreamHandler) paths() []string {
	handler.mu.Lock()
	defer handler.mu.Unlock()

	return append([]string(nil), handler.ensurePaths...)
}

type fakeTenantFileGroup struct {
	id                  int
	name                string
	getFileNames        []string
	storedFileInfo      []fileiface.FileInfo
	storedOverwrite     []bool
	copyCalls           []fakeCopyCall
	isDirCalls          []fakeDirectoryCall
	getSubGroupCalls    []fakeDirectoryCall
	ensureSubGroupCalls []fakeDirectoryCall
	isDirResult         bool
	directoryErr        error
}

type fakeDirectoryCall struct {
	ctx          contextx.IContext
	relativePath string
}

type fakeCopyCall struct {
	srcPath   string
	destGroup fileiface.FileGroup
	destPath  string
	overwrite bool
}

func (group *fakeTenantFileGroup) Name() string {
	return group.name
}

func (group *fakeTenantFileGroup) AbsDirs() []string {
	return fileiface.ConvertAbsPathToAbsDirs(group.name)
}

func (group *fakeTenantFileGroup) SubGroups(contextx.IContext) ([]fileiface.FileGroup, error) {
	return []fileiface.FileGroup{group}, nil
}

func (group *fakeTenantFileGroup) IsDir(nCtx contextx.IContext, relativePath string) (bool, error) {
	group.isDirCalls = append(group.isDirCalls, fakeDirectoryCall{ctx: nCtx, relativePath: relativePath})
	if group.directoryErr != nil {
		return false, group.directoryErr
	}
	return relativePath == "." || group.isDirResult, nil
}

func (group *fakeTenantFileGroup) GetSubGroup(nCtx contextx.IContext, relativePath string) (fileiface.FileGroup, error) {
	group.getSubGroupCalls = append(group.getSubGroupCalls, fakeDirectoryCall{ctx: nCtx, relativePath: relativePath})
	if group.directoryErr != nil {
		return nil, group.directoryErr
	}
	if relativePath == "." {
		return group, nil
	}
	return &fakeTenantFileGroup{name: path.Join(group.name, relativePath)}, nil
}

func (group *fakeTenantFileGroup) EnsureSubGroup(nCtx contextx.IContext, relativePath string) (fileiface.FileGroup, error) {
	group.ensureSubGroupCalls = append(group.ensureSubGroupCalls, fakeDirectoryCall{ctx: nCtx, relativePath: relativePath})
	if group.directoryErr != nil {
		return nil, group.directoryErr
	}
	if relativePath == "." {
		return group, nil
	}
	return &fakeTenantFileGroup{name: path.Join(group.name, relativePath)}, nil
}

func (group *fakeTenantFileGroup) AllFiles(contextx.IContext) ([]fileiface.File, error) {
	return []fileiface.File{fakeUpstreamFile{name: "all.txt"}}, nil
}

func (group *fakeTenantFileGroup) GetFile(_ contextx.IContext, name string) (fileiface.File, error) {
	group.getFileNames = append(group.getFileNames, name)
	return fakeUpstreamFile{name: name}, nil
}

func (group *fakeTenantFileGroup) Store(_ contextx.IContext, info fileiface.FileInfo, _ io.ReadCloser, overwrite bool) error {
	group.storedFileInfo = append(group.storedFileInfo, info)
	group.storedOverwrite = append(group.storedOverwrite, overwrite)
	return nil
}

func (group *fakeTenantFileGroup) Copy(
	_ contextx.IContext,
	srcPath string,
	destGroup fileiface.FileGroup,
	destPath string,
	overwrite bool,
) error {
	group.copyCalls = append(group.copyCalls, fakeCopyCall{
		srcPath:   srcPath,
		destGroup: destGroup,
		destPath:  destPath,
		overwrite: overwrite,
	})

	return nil
}

func (group *fakeTenantFileGroup) Remove(_ contextx.IContext, _ string) error {
	return nil
}

type fakeUpstreamFile struct {
	name string
}

func (file fakeUpstreamFile) FileObject() fileiface.FileObject {
	return fileiface.RemoteFile
}

func (file fakeUpstreamFile) Info() fileiface.FileInfo {
	return fileiface.FileInfo{Name: file.name}
}

func (file fakeUpstreamFile) AbsDirs() []string {
	return nil
}

func (file fakeUpstreamFile) Content(contextx.IContext) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader(file.name)), nil
}

func TestUpstreamFileGroupResolvesAndCachesByTenant(t *testing.T) {
	handler := new(fakeUpstreamHandler)
	group := &upstreamFileGroup{ensurer: handler, basePath: "release/proxy"}
	systemCtx := contextx.New(t.Context(), contextx.WithTenantID("system"))
	otherCtx := contextx.New(t.Context(), contextx.WithTenantID("other"))

	first, err := group.resolve(systemCtx)
	require.NoError(t, err)
	second, err := group.resolve(systemCtx)
	require.NoError(t, err)
	other, err := group.resolve(otherCtx)
	require.NoError(t, err)

	require.Same(t, first, second)
	require.NotSame(t, first, other)
	require.Equal(t, []string{"/system/release/proxy", "/other/release/proxy"}, handler.paths())
	require.Equal(t, "proxy", group.Name())
	require.Equal(t, []string{"release", "proxy"}, group.AbsDirs())
}

func TestUpstreamFileGroupRejectsMissingTenant(t *testing.T) {
	handler := new(fakeUpstreamHandler)
	group := &upstreamFileGroup{ensurer: handler, basePath: "origin/agent"}

	_, err := group.GetFile(nil, "agent.tgz")
	require.ErrorContains(t, err, "context is nil")

	_, err = group.GetFile(contextx.New(t.Context()), "agent.tgz")
	require.ErrorContains(t, err, "tenant-id not found")
	require.Empty(t, handler.paths())
}

func TestUpstreamFileGroupDoesNotCacheEnsureFailure(t *testing.T) {
	ensureErr := errors.New("bkrepo unavailable")
	handler := &fakeUpstreamHandler{ensureErrs: []error{ensureErr}}
	group := &upstreamFileGroup{ensurer: handler, basePath: "origin/server"}
	nCtx := contextx.New(t.Context(), contextx.WithTenantID("system"))

	_, err := group.resolve(nCtx)
	require.ErrorIs(t, err, ensureErr)

	resolved, err := group.resolve(nCtx)
	require.NoError(t, err)
	require.NotNil(t, resolved)
	require.Equal(t, []string{"/system/origin/server", "/system/origin/server"}, handler.paths())
}

func TestUpstreamFileGroupSkipsCanceledInitialization(t *testing.T) {
	handler := new(fakeUpstreamHandler)
	group := &upstreamFileGroup{ensurer: handler, basePath: "export"}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	nCtx := contextx.New(ctx, contextx.WithTenantID("system"))

	_, err := group.resolve(nCtx)
	require.ErrorIs(t, err, context.Canceled)
	require.Empty(t, handler.paths())

	_, err = group.resolve(contextx.New(t.Context(), contextx.WithTenantID("system")))
	require.NoError(t, err)
	require.Equal(t, []string{"/system/export"}, handler.paths())
}

func TestUpstreamFileGroupDelegatesOperations(t *testing.T) {
	handler := new(fakeUpstreamHandler)
	group := &upstreamFileGroup{ensurer: handler, basePath: "origin/v3/plugin"}
	nCtx := contextx.New(t.Context(), contextx.WithTenantID("system"))

	file, err := group.GetFile(nCtx, "plugin.tgz")
	require.NoError(t, err)
	require.Equal(t, "plugin.tgz", file.Info().Name)

	info := fileiface.FileInfo{Name: "new-plugin.tgz"}
	err = group.Store(nCtx, info, io.NopCloser(strings.NewReader("content")), true)
	require.NoError(t, err)

	subGroups, err := group.SubGroups(nCtx)
	require.NoError(t, err)
	require.Len(t, subGroups, 1)

	files, err := group.AllFiles(nCtx)
	require.NoError(t, err)
	require.Equal(t, "all.txt", files[0].Info().Name)

	resolved, err := group.resolve(nCtx)
	require.NoError(t, err)
	tenantGroup, ok := resolved.(*fakeTenantFileGroup)
	require.True(t, ok)
	require.Equal(t, []string{"plugin.tgz"}, tenantGroup.getFileNames)
	require.Equal(t, []fileiface.FileInfo{info}, tenantGroup.storedFileInfo)
	require.Equal(t, []bool{true}, tenantGroup.storedOverwrite)
	require.Equal(t, []string{"/system/origin/v3/plugin"}, handler.paths())
}

func TestUpstreamFileGroupDelegatesDirectoryOperationsByTenant(t *testing.T) {
	handler := new(fakeUpstreamHandler)
	group := &upstreamFileGroup{ensurer: handler, basePath: "origin/v3/plugin"}
	for _, tenantID := range []string{"system", "tenant-a"} {
		nCtx := contextx.New(t.Context(), contextx.WithTenantID(tenantID))
		resolved, err := group.resolve(nCtx)
		require.NoError(t, err)
		tenantGroup, ok := resolved.(*fakeTenantFileGroup)
		require.True(t, ok)
		tenantGroup.isDirResult = true

		isDir, err := group.IsDir(nCtx, ".")
		require.NoError(t, err)
		require.True(t, isDir)
		root, err := group.GetSubGroup(nCtx, ".")
		require.NoError(t, err)
		require.Same(t, tenantGroup, root)
		root, err = group.EnsureSubGroup(nCtx, ".")
		require.NoError(t, err)
		require.Same(t, tenantGroup, root)

		isDir, err = group.IsDir(nCtx, "nested/child")
		require.NoError(t, err)
		require.True(t, isDir)
		subGroup, err := group.GetSubGroup(nCtx, "nested/child")
		require.NoError(t, err)
		require.Equal(t, path.Join(tenantGroup.name, "nested/child"), subGroup.Name())
		subGroup, err = group.EnsureSubGroup(nCtx, "nested/child")
		require.NoError(t, err)
		require.Equal(t, path.Join(tenantGroup.name, "nested/child"), subGroup.Name())

		require.Equal(t, []fakeDirectoryCall{
			{ctx: nCtx, relativePath: "."},
			{ctx: nCtx, relativePath: "nested/child"},
		}, tenantGroup.isDirCalls)
		require.Equal(t, []fakeDirectoryCall{
			{ctx: nCtx, relativePath: "."},
			{ctx: nCtx, relativePath: "nested/child"},
		}, tenantGroup.getSubGroupCalls)
		require.Equal(t, []fakeDirectoryCall{
			{ctx: nCtx, relativePath: "."},
			{ctx: nCtx, relativePath: "nested/child"},
		}, tenantGroup.ensureSubGroupCalls)
	}
	require.Equal(t, []string{"/system/origin/v3/plugin", "/tenant-a/origin/v3/plugin"}, handler.paths())
}

func TestUpstreamFileGroupDirectoryOperationsPreserveErrors(t *testing.T) {
	backendErr := errors.New("backend unavailable")
	operations := []struct {
		name string
		call func(fileiface.FileGroup, contextx.IContext) error
	}{
		{name: "IsDir", call: func(group fileiface.FileGroup, nCtx contextx.IContext) error {
			_, err := group.IsDir(nCtx, ".")
			return err
		}},
		{name: "GetSubGroup", call: func(group fileiface.FileGroup, nCtx contextx.IContext) error {
			_, err := group.GetSubGroup(nCtx, ".")
			return err
		}},
		{name: "EnsureSubGroup", call: func(group fileiface.FileGroup, nCtx contextx.IContext) error {
			_, err := group.EnsureSubGroup(nCtx, ".")
			return err
		}},
	}
	for _, operation := range operations {
		t.Run(operation.name, func(t *testing.T) {
			handler := &fakeUpstreamHandler{ensureErrs: []error{backendErr}}
			group := &upstreamFileGroup{ensurer: handler, basePath: "origin/v3/plugin"}
			nCtx := contextx.New(t.Context(), contextx.WithTenantID("tenant-a"))
			require.ErrorIs(t, operation.call(group, nCtx), backendErr)

			resolved, err := group.resolve(nCtx)
			require.NoError(t, err)
			tenantGroup, ok := resolved.(*fakeTenantFileGroup)
			require.True(t, ok)
			tenantGroup.directoryErr = backendErr
			require.ErrorIs(t, operation.call(group, nCtx), backendErr)
			require.ErrorContains(t, operation.call(group, nil), "context is nil")
		})
	}
}

func TestUpstreamFileGroupCopiesFromSystemToRequestTenant(t *testing.T) {
	handler := new(fakeUpstreamHandler)
	group := &upstreamFileGroup{ensurer: handler, basePath: "release/plugin"}
	nCtx := contextx.New(t.Context(), contextx.WithTenantID("tenant-a"))

	err := group.Copy(nCtx, "source.tgz", group, ".", false)
	require.NoError(t, err)

	systemGroup, err := group.resolve(contextx.From(nCtx, contextx.WithTenantID(tenant.SystemTenantID)))
	require.NoError(t, err)
	destinationGroup, err := group.resolve(nCtx)
	require.NoError(t, err)

	systemFileGroup, ok := systemGroup.(*fakeTenantFileGroup)
	require.True(t, ok)
	require.Equal(t, []fakeCopyCall{{
		srcPath:   "source.tgz",
		destGroup: destinationGroup,
		destPath:  ".",
		overwrite: false,
	}}, systemFileGroup.copyCalls)
	require.Equal(t, []string{"/system/release/plugin", "/tenant-a/release/plugin"}, handler.paths())
}

func TestUpstreamFileGroupConcurrentResolveInitializesOnce(t *testing.T) {
	const workerCount = 32

	started := make(chan struct{})
	release := make(chan struct{})
	var firstEnsure sync.Once
	handler := &fakeUpstreamHandler{beforeEnsure: func(string) {
		firstEnsure.Do(func() { close(started) })
		<-release
	}}
	group := &upstreamFileGroup{ensurer: handler, basePath: "release/plugin"}
	nCtx := contextx.New(t.Context(), contextx.WithTenantID("system"))
	results := make(chan fileiface.FileGroup, workerCount)
	errs := make(chan error, workerCount)
	var workers sync.WaitGroup
	var ready sync.WaitGroup
	ready.Add(workerCount)
	start := make(chan struct{})

	for range workerCount {
		workers.Go(func() {
			ready.Done()
			<-start
			resolved, err := group.resolve(nCtx)
			if err != nil {
				errs <- err
				return
			}
			results <- resolved
		})
	}

	ready.Wait()
	close(start)
	<-started
	close(release)
	workers.Wait()
	close(results)
	close(errs)
	require.Empty(t, errs)

	var winner fileiface.FileGroup
	for resolved := range results {
		if winner == nil {
			winner = resolved
			continue
		}
		require.Same(t, winner, resolved)
	}
	require.NotNil(t, winner)
	require.Equal(t, []string{"/system/release/plugin"}, handler.paths())
}

func TestUpstreamFileGroupInitializesTenantsIndependently(t *testing.T) {
	started := make(chan string, 2)
	release := make(chan struct{})
	handler := &fakeUpstreamHandler{beforeEnsure: func(groupPath string) {
		started <- groupPath
		<-release
	}}
	group := &upstreamFileGroup{ensurer: handler, basePath: "export"}
	errs := make(chan error, 2)
	for _, tenantID := range []string{"system", "other"} {
		go func() {
			nCtx := contextx.New(t.Context(), contextx.WithTenantID(tenantID))
			_, err := group.resolve(nCtx)
			errs <- err
		}()
	}

	// Both tenants must enter EnsureFileGroup before either initialization completes.
	paths := []string{<-started, <-started}
	close(release)
	require.NoError(t, <-errs)
	require.NoError(t, <-errs)
	require.ElementsMatch(t, []string{"/system/export", "/other/export"}, paths)
}
