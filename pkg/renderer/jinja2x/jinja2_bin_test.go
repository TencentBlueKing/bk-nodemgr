//go:build linux

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

package jinja2x

import (
	"errors"
	"os"
	"sync"
	"testing"
)

func TestJinja2ExecBinaryPathFirstCreate(t *testing.T) {
	resetJinja2ExecBin(t)

	path, err := jinja2ExecBinaryPath()
	if err != nil {
		t.Fatalf("jinja2ExecBinaryPath() error = %v", err)
	}
	assertExecutablePath(t, path)
}

func TestJinja2ExecBinaryPathCachedReuse(t *testing.T) {
	resetJinja2ExecBin(t)

	path1, err := jinja2ExecBinaryPath()
	if err != nil {
		t.Fatalf("first jinja2ExecBinaryPath() error = %v", err)
	}
	path2, err := jinja2ExecBinaryPath()
	if err != nil {
		t.Fatalf("second jinja2ExecBinaryPath() error = %v", err)
	}
	if path1 != path2 {
		t.Fatalf("cached path mismatch: first %q, second %q", path1, path2)
	}
}

func TestJinja2ExecBinaryPathDeleteAndRecreate(t *testing.T) {
	resetJinja2ExecBin(t)

	path1, err := jinja2ExecBinaryPath()
	if err != nil {
		t.Fatalf("first jinja2ExecBinaryPath() error = %v", err)
	}
	if err := os.Remove(path1); err != nil {
		t.Fatalf("remove cached binary %q: %v", path1, err)
	}

	path2, err := jinja2ExecBinaryPath()
	if err != nil {
		t.Fatalf("second jinja2ExecBinaryPath() error = %v", err)
	}
	assertExecutablePath(t, path2)
}

func TestJinja2ExecBinaryPathCleanupAfterDeletion(t *testing.T) {
	resetJinja2ExecBin(t)

	path, err := jinja2ExecBinaryPath()
	if err != nil {
		t.Fatalf("jinja2ExecBinaryPath() error = %v", err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatalf("remove cached binary %q: %v", path, err)
	}
	jinja2ExecBin.cleanup()
}

func TestJinja2ExecBinaryPathConcurrent(t *testing.T) {
	resetJinja2ExecBin(t)

	path, err := jinja2ExecBinaryPath()
	if err != nil {
		t.Fatalf("initial jinja2ExecBinaryPath() error = %v", err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatalf("remove cached binary %q: %v", path, err)
	}

	const goroutines = 8
	results := make(chan struct {
		path string
		err  error
	}, goroutines)

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for range goroutines {
		go func() {
			defer wg.Done()
			path, err := jinja2ExecBinaryPath()
			results <- struct {
				path string
				err  error
			}{path: path, err: err}
		}()
	}
	wg.Wait()
	close(results)

	for result := range results {
		if result.err != nil {
			t.Fatalf("concurrent jinja2ExecBinaryPath() error = %v", result.err)
		}
		assertExecutablePath(t, result.path)
	}
}

func resetJinja2ExecBin(t *testing.T) {
	t.Helper()

	savedBinaryPath := jinja2ExecBin.binaryPath
	savedCleanup := jinja2ExecBin.cleanup
	t.Cleanup(func() {
		jinja2ExecBin.mu.Lock()
		cleanup := jinja2ExecBin.cleanup
		jinja2ExecBin.binaryPath = savedBinaryPath
		jinja2ExecBin.cleanup = savedCleanup
		jinja2ExecBin.mu.Unlock()

		if cleanup != nil {
			cleanup()
		}
	})
	jinja2ExecBin = struct {
		binaryPath string
		mu         sync.Mutex
		cleanup    func()
	}{}
}

func assertExecutablePath(t *testing.T, path string) {
	t.Helper()

	if path == "" {
		t.Fatal("path is empty")
	}
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			t.Fatalf("path %q does not exist", path)
		}
		t.Fatalf("stat path %q: %v", path, err)
	}
	if info.Mode().Perm()&0100 == 0 {
		t.Fatalf("path %q mode %v has no owner executable bit", path, info.Mode().Perm())
	}
}
