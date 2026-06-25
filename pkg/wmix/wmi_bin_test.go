//go:build linux

package wmix

import (
	"errors"
	"os"
	"sync"
	"testing"
)

func resetWmiBin(t *testing.T) {
	t.Helper()

	savedPath := wmiBin.binaryPath
	savedCleanup := wmiBin.cleanup
	t.Cleanup(func() {
		wmiBin.mu.Lock()
		defer wmiBin.mu.Unlock()
		if wmiBin.cleanup != nil {
			wmiBin.cleanup()
		}
		wmiBin.binaryPath = savedPath
		wmiBin.cleanup = savedCleanup
	})
	wmiBin.mu.Lock()
	wmiBin.binaryPath = ""
	wmiBin.cleanup = nil
	wmiBin.mu.Unlock()
}

func TestWmiBinaryPathFirstCreate(t *testing.T) {
	resetWmiBin(t)

	path, err := wmiBinaryPath()
	if err != nil {
		t.Fatalf("wmiBinaryPath() error = %v", err)
	}
	if path == "" {
		t.Fatal("wmiBinaryPath() returned empty path")
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("os.Stat(%q) error = %v", path, err)
	}
	if info.Mode().Perm()&0100 == 0 {
		t.Fatalf("binary mode %v does not include owner executable bit", info.Mode().Perm())
	}
}

func TestWmiBinaryPathCachedReuse(t *testing.T) {
	resetWmiBin(t)

	path1, err := wmiBinaryPath()
	if err != nil {
		t.Fatalf("first wmiBinaryPath() error = %v", err)
	}
	path2, err := wmiBinaryPath()
	if err != nil {
		t.Fatalf("second wmiBinaryPath() error = %v", err)
	}
	if path1 != path2 {
		t.Fatalf("second path = %q, want cached path %q", path2, path1)
	}
}

func TestWmiBinaryPathDeleteAndRecreate(t *testing.T) {
	resetWmiBin(t)

	path1, err := wmiBinaryPath()
	if err != nil {
		t.Fatalf("first wmiBinaryPath() error = %v", err)
	}
	if err := os.Remove(path1); err != nil {
		t.Fatalf("os.Remove(%q) error = %v", path1, err)
	}

	path2, err := wmiBinaryPath()
	if err != nil {
		t.Fatalf("second wmiBinaryPath() error = %v", err)
	}
	info, err := os.Stat(path2)
	if err != nil {
		t.Fatalf("os.Stat(%q) error = %v", path2, err)
	}
	if info.Mode().Perm()&0100 == 0 {
		t.Fatalf("binary mode %v does not include owner executable bit", info.Mode().Perm())
	}
}

func TestWmiBinaryPathCleanupAfterDeletion(t *testing.T) {
	resetWmiBin(t)

	path, err := wmiBinaryPath()
	if err != nil {
		t.Fatalf("wmiBinaryPath() error = %v", err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatalf("os.Remove(%q) error = %v", path, err)
	}
	wmiBin.cleanup()
}

func TestWmiBinaryPathConcurrent(t *testing.T) {
	resetWmiBin(t)

	path, err := wmiBinaryPath()
	if err != nil {
		t.Fatalf("wmiBinaryPath() error = %v", err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatalf("os.Remove(%q) error = %v", path, err)
	}

	const goroutineCount = 8
	paths := make(chan string, goroutineCount)
	errs := make(chan error, goroutineCount)
	var wg sync.WaitGroup
	wg.Add(goroutineCount)
	for range goroutineCount {
		go func() {
			defer wg.Done()

			path, err := wmiBinaryPath()
			if err != nil {
				errs <- err

				return
			}
			paths <- path
		}()
	}
	wg.Wait()
	close(paths)
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent wmiBinaryPath() error = %v", err)
		}
	}
	for path := range paths {
		if path == "" {
			t.Fatal("concurrent wmiBinaryPath() returned empty path")
		}
		if _, err := os.Stat(path); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				t.Fatalf("concurrent path %q does not exist", path)
			}
			t.Fatalf("os.Stat(%q) error = %v", path, err)
		}
	}
}
