/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package throttle

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"fmt"
	"io"
	"math"
	"runtime"
	"testing"
	"time"
)

// ─── helpers ────────────────────────────────────────────────────────

// makeTgz creates an in-memory .tgz archive containing a single file with the
// given payload.
func makeTgz(t *testing.T, payload []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	hdr := &tar.Header{
		Name:    "data.bin",
		Size:    int64(len(payload)),
		Mode:    0644,
		ModTime: time.Now(),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("tar write header: %v", err)
	}
	if _, err := tw.Write(payload); err != nil {
		t.Fatalf("tar write body: %v", err)
	}
	_ = tw.Close()
	_ = gw.Close()
	return buf.Bytes()
}

// ─── global variable + Copy() tests ─────────────────────────────────

func TestCopy_UsesGlobalCPULimit(t *testing.T) {
	payload := make([]byte, 2*1024*1024) // 2 MB
	if _, err := rand.Read(payload); err != nil {
		t.Fatalf("rand.Read: %v", err)
	}

	t.Run("global_unset_no_throttle", func(t *testing.T) {
		CPULimit = 0
		src := bytes.NewReader(payload)
		var dst bytes.Buffer

		n, err := Copy(&dst, src)
		if err != nil {
			t.Fatalf("Copy error: %v", err)
		}
		if n != int64(len(payload)) {
			t.Errorf("written bytes = %d, want %d", n, len(payload))
		}
		if !bytes.Equal(dst.Bytes(), payload) {
			t.Error("copied data does not match source")
		}
	})

	t.Run("global_set_throttle", func(t *testing.T) {
		CPULimit = 0.5
		defer func() { CPULimit = 0 }()

		src := bytes.NewReader(payload)
		var dst bytes.Buffer

		n, err := Copy(&dst, src)
		if err != nil {
			t.Fatalf("Copy error: %v", err)
		}
		if n != int64(len(payload)) {
			t.Errorf("written bytes = %d, want %d", n, len(payload))
		}
		if !bytes.Equal(dst.Bytes(), payload) {
			t.Error("copied data does not match source")
		}
	})
}

// ─── core tests ─────────────────────────────────────────────────────

func TestThrottledCopy_DataIntegrity(t *testing.T) {
	payload := make([]byte, 2*1024*1024) // 2 MB random data
	if _, err := rand.Read(payload); err != nil {
		t.Fatalf("rand.Read: %v", err)
	}

	tests := []struct {
		name     string
		cpuLimit float64
	}{
		{"no_throttle_negative", -1},
		{"no_throttle_zero", 0},
		{"no_throttle_full", 1.0},
		{"throttle_50pct", 0.5},
		{"throttle_30pct", 0.3},
		{"throttle_80pct", 0.8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := bytes.NewReader(payload)
			var dst bytes.Buffer

			n, err := ThrottledCopy(&dst, src, tt.cpuLimit)
			if err != nil {
				t.Fatalf("ThrottledCopy error: %v", err)
			}
			if n != int64(len(payload)) {
				t.Errorf("written bytes = %d, want %d", n, len(payload))
			}
			if !bytes.Equal(dst.Bytes(), payload) {
				t.Error("copied data does not match source")
			}
		})
	}
}

func TestThrottledCopy_GzipDecompress(t *testing.T) {
	payload := make([]byte, 4*1024*1024) // 4 MB
	if _, err := rand.Read(payload); err != nil {
		t.Fatalf("rand.Read: %v", err)
	}
	tgzData := makeTgz(t, payload)

	gr, err := gzip.NewReader(bytes.NewReader(tgzData))
	if err != nil {
		t.Fatalf("gzip.NewReader: %v", err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	if _, err := tr.Next(); err != nil {
		t.Fatalf("tar next: %v", err)
	}

	var dst bytes.Buffer
	n, err := ThrottledCopy(&dst, tr, 0.5)
	if err != nil {
		t.Fatalf("ThrottledCopy error: %v", err)
	}
	if n != int64(len(payload)) {
		t.Errorf("written bytes = %d, want %d", n, len(payload))
	}
	if !bytes.Equal(dst.Bytes(), payload) {
		t.Error("decompressed data does not match original")
	}
}

func TestThrottledCopy_GzipCompress(t *testing.T) {
	payload := bytes.Repeat([]byte("hello world, this is compressible data\n"), 100000)

	var compressedBuf bytes.Buffer
	gw := gzip.NewWriter(&compressedBuf)

	_, err := ThrottledCopy(gw, bytes.NewReader(payload), 0.5)
	if err != nil {
		t.Fatalf("ThrottledCopy compress error: %v", err)
	}
	if err := gw.Close(); err != nil {
		t.Fatalf("gzip close: %v", err)
	}

	gr, err := gzip.NewReader(&compressedBuf)
	if err != nil {
		t.Fatalf("gzip.NewReader: %v", err)
	}
	defer gr.Close()

	decompressed, err := io.ReadAll(gr)
	if err != nil {
		t.Fatalf("decompress readall: %v", err)
	}
	if !bytes.Equal(decompressed, payload) {
		t.Error("round-trip compressed data does not match original")
	}
}

func TestThrottledCopy_EmptyReader(t *testing.T) {
	var dst bytes.Buffer
	n, err := ThrottledCopy(&dst, bytes.NewReader(nil), 0.5)
	if err != nil {
		t.Fatalf("ThrottledCopy error: %v", err)
	}
	if n != 0 {
		t.Errorf("written bytes = %d, want 0", n)
	}
}

// ─── CPU usage verification ────────────────────────────────────────

// TestThrottledCopy_CPUUsage is the key test: it verifies that the duty-cycle
// mechanism actually reduces CPU consumption in a measurable way.
//
// Strategy:
//  1. Run gzip decompression WITHOUT throttle and measure CPU time.
//  2. Run gzip decompression WITH throttle (cpuLimit=0.5) and measure CPU time.
//  3. Assert that the throttled run uses significantly less CPU time relative to wall time.
func TestThrottledCopy_CPUUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipped in short mode")
	}

	runtime.GOMAXPROCS(1) // simulate single-core

	payload := bytes.Repeat([]byte("abcdefghijklmnopqrstuvwxyz0123456789\n"), 3000000) // ~111 MB
	tgzData := makeTgz(t, payload)

	type result struct {
		wallTime time.Duration
		cpuTime  time.Duration
		cpuRatio float64
	}

	runDecompress := func(cpuLimit float64) result {
		gr, err := gzip.NewReader(bytes.NewReader(tgzData))
		if err != nil {
			t.Fatalf("gzip.NewReader: %v", err)
		}
		defer gr.Close()
		tr := tar.NewReader(gr)
		if _, err := tr.Next(); err != nil {
			t.Fatalf("tar next: %v", err)
		}

		runtime.GC() // reduce GC interference

		cpuBefore := processCPUTime()
		wallStart := time.Now()

		var dst bytes.Buffer
		if _, err := ThrottledCopy(&dst, tr, cpuLimit); err != nil {
			t.Fatalf("ThrottledCopy: %v", err)
		}

		wallElapsed := time.Since(wallStart)
		cpuElapsed := processCPUTime() - cpuBefore

		ratio := float64(cpuElapsed) / float64(wallElapsed)

		return result{
			wallTime: wallElapsed,
			cpuTime:  cpuElapsed,
			cpuRatio: ratio,
		}
	}

	// --- run without throttle (cpuLimit=1.0 → full speed) ---
	baseline := runDecompress(1.0)
	t.Logf("baseline  (no throttle): wall=%v, cpu=%v, ratio=%.2f",
		baseline.wallTime, baseline.cpuTime, baseline.cpuRatio)

	// --- run with throttle at 50% ---
	throttled50 := runDecompress(0.5)
	t.Logf("throttled (cpuLimit=0.5): wall=%v, cpu=%v, ratio=%.2f",
		throttled50.wallTime, throttled50.cpuTime, throttled50.cpuRatio)

	// --- run with throttle at 30% ---
	throttled30 := runDecompress(0.3)
	t.Logf("throttled (cpuLimit=0.3): wall=%v, cpu=%v, ratio=%.2f",
		throttled30.wallTime, throttled30.cpuTime, throttled30.cpuRatio)

	// --- assertions ---

	// 1. Throttled wall time must be longer than baseline.
	if throttled50.wallTime <= baseline.wallTime {
		t.Errorf("throttled(0.5) wall time (%v) should be > baseline (%v)",
			throttled50.wallTime, baseline.wallTime)
	}

	// 2. Throttled CPU ratio should be close to the target.
	//    Allow ±0.15 tolerance for OS scheduling jitter.
	assertCPURatio(t, "cpuLimit=0.5", throttled50.cpuRatio, 0.5, 0.15)
	assertCPURatio(t, "cpuLimit=0.3", throttled30.cpuRatio, 0.3, 0.15)

	// 3. Throttle at 0.3 must take longer than throttle at 0.5.
	if throttled30.wallTime <= throttled50.wallTime {
		t.Errorf("throttled(0.3) wall time (%v) should be > throttled(0.5) (%v)",
			throttled30.wallTime, throttled50.wallTime)
	}

	// --- summary ---
	fmt.Printf("\n=== CPU Throttle Test Summary ===\n")
	fmt.Printf("  %-20s wall=%-12v cpu=%-12v ratio=%.2f\n",
		"baseline:", baseline.wallTime, baseline.cpuTime, baseline.cpuRatio)
	fmt.Printf("  %-20s wall=%-12v cpu=%-12v ratio=%.2f (target=0.50)\n",
		"throttled(0.5):", throttled50.wallTime, throttled50.cpuTime, throttled50.cpuRatio)
	fmt.Printf("  %-20s wall=%-12v cpu=%-12v ratio=%.2f (target=0.30)\n",
		"throttled(0.3):", throttled30.wallTime, throttled30.cpuTime, throttled30.cpuRatio)
	fmt.Printf("================================\n\n")
}

func assertCPURatio(t *testing.T, label string, actual, target, tolerance float64) {
	t.Helper()
	if math.Abs(actual-target) > tolerance {
		t.Errorf("%s: CPU ratio = %.2f, want %.2f ± %.2f", label, actual, target, tolerance)
	}
}

// ─── benchmark ──────────────────────────────────────────────────────

func BenchmarkThrottledCopy_NoThrottle(b *testing.B) {
	payload := bytes.Repeat([]byte("benchmark data\n"), 100000) // ~1.5 MB
	for i := 0; i < b.N; i++ {
		var dst bytes.Buffer
		_, _ = ThrottledCopy(&dst, bytes.NewReader(payload), 0)
	}
}

func BenchmarkThrottledCopy_50Pct(b *testing.B) {
	payload := bytes.Repeat([]byte("benchmark data\n"), 100000)
	for i := 0; i < b.N; i++ {
		var dst bytes.Buffer
		_, _ = ThrottledCopy(&dst, bytes.NewReader(payload), 0.5)
	}
}
