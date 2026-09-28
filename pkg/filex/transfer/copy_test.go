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

package transfer

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
)

func TestCopyFileClosesSourceStream(t *testing.T) {
	storeErr := errors.New("store failed")
	closeErr := errors.New("close failed")
	cases := []struct {
		name     string
		storeErr error
		closeErr error
		wantErrs []error
	}{
		{name: "success"},
		{name: "store failure", storeErr: storeErr, wantErrs: []error{storeErr}},
		{name: "store cancellation", storeErr: context.Canceled, wantErrs: []error{context.Canceled}},
		{name: "close failure", closeErr: closeErr, wantErrs: []error{closeErr}},
		{name: "store and close failures", storeErr: storeErr, closeErr: closeErr, wantErrs: []error{storeErr, closeErr}},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := contextx.Background()
			stream := &trackedReadCloser{closeErr: tt.closeErr}
			source := &testFileGroup{file: &mockFile{content: func(contextx.IContext) (io.ReadCloser, error) {
				return stream, nil
			}}}
			destination := &testFileGroup{store: func(_ contextx.IContext, _ fileiface.FileInfo, _ io.ReadCloser, _ bool) error {
				return tt.storeErr
			}}

			err := copyFile(ctx, source, "source", destination, "destination", false)
			if stream.closeCalls != 1 {
				t.Fatalf("source Close called %d times, want 1", stream.closeCalls)
			}
			for _, want := range tt.wantErrs {
				if !errors.Is(err, want) {
					t.Errorf("copyFile error %v does not contain %v", err, want)
				}
			}
			if len(tt.wantErrs) == 0 && err != nil {
				t.Fatalf("copyFile returned unexpected error: %v", err)
			}
		})
	}

	t.Run("panic unwinding", func(t *testing.T) {
		stream := &trackedReadCloser{}
		source := &testFileGroup{file: &mockFile{content: func(contextx.IContext) (io.ReadCloser, error) { return stream, nil }}}
		destination := &testFileGroup{store: func(contextx.IContext, fileiface.FileInfo, io.ReadCloser, bool) error { panic("store panic") }}
		func() {
			defer func() {
				if recover() == nil {
					t.Fatal("expected Store panic")
				}
			}()
			_ = copyFile(contextx.New(context.Background()), source, "source", destination, "destination", false)
		}()
		if stream.closeCalls != 1 {
			t.Fatalf("source Close called %d times during panic unwinding, want 1", stream.closeCalls)
		}
	})
}

type trackedReadCloser struct {
	closeCalls int
	closeErr   error
}

func (*trackedReadCloser) Read([]byte) (int, error) { return 0, io.EOF }

func (reader *trackedReadCloser) Close() error {
	reader.closeCalls++
	return reader.closeErr
}

type mockFile struct {
	content func(contextx.IContext) (io.ReadCloser, error)
}

func (*mockFile) FileObject() fileiface.FileObject { return fileiface.LocalFile }
func (*mockFile) Info() fileiface.FileInfo         { return fileiface.FileInfo{Name: "source"} }
func (*mockFile) AbsDirs() []string                { return nil }
func (file *mockFile) Content(ctx contextx.IContext) (io.ReadCloser, error) {
	return file.content(ctx)
}

type testFileGroup struct {
	file  fileiface.File
	store func(contextx.IContext, fileiface.FileInfo, io.ReadCloser, bool) error
}

func (*testFileGroup) Name() string                                  { return "mock" }
func (*testFileGroup) AbsDirs() []string                             { return nil }
func (*testFileGroup) IsDir(contextx.IContext, string) (bool, error) { return false, nil }
func (group *testFileGroup) GetSubGroup(contextx.IContext, string) (fileiface.FileGroup, error) {
	return group, nil
}
func (group *testFileGroup) EnsureSubGroup(contextx.IContext, string) (fileiface.FileGroup, error) {
	return group, nil
}
func (*testFileGroup) SubGroups(contextx.IContext) ([]fileiface.FileGroup, error) {
	return nil, errors.New("unexpected SubGroups")
}
func (*testFileGroup) AllFiles(contextx.IContext) ([]fileiface.File, error) {
	return nil, errors.New("unexpected AllFiles")
}
func (group *testFileGroup) GetFile(contextx.IContext, string) (fileiface.File, error) {
	return group.file, nil
}
func (group *testFileGroup) Store(ctx contextx.IContext, info fileiface.FileInfo, reader io.ReadCloser, overwrite bool) error {
	return group.store(ctx, info, reader, overwrite)
}
func (*testFileGroup) Copy(contextx.IContext, string, fileiface.FileGroup, string, bool) error {
	return errors.New("unexpected Copy")
}
func (*testFileGroup) Remove(contextx.IContext, string) error {
	return errors.New("unexpected Remove")
}
