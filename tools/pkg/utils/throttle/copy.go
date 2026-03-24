/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package throttle provides CPU-aware IO copy utilities.
package throttle

import (
	"io"
	"time"
)

const (
	defaultBufSize = 32 * 1024 // same as io.Copy internal buffer

	// throttleWindow is the minimum accumulated work time before a sleep is triggered.
	// Avoids micro-sleeps that have high overhead relative to actual work.
	throttleWindow = 10 * time.Millisecond

	// maxSleepPerWindow caps the sleep duration for a single throttle window
	// to guard against measurement anomalies. With a 10ms window this supports
	// cpuLimit values down to ~0.01 (1%) without clipping.
	maxSleepPerWindow = time.Second
)

// CPULimit is the global CPU usage limit for throttled IO operations.
// Value range (0.0, 1.0]: 0.5 means ~50% CPU usage.
// Values <= 0 or >= 1.0 disable throttling entirely.
var CPULimit float64

// Copy copies from src to dst using the global CPULimit setting.
// This is the primary entry point for callers that rely on the CLI-configured limit.
func Copy(dst io.Writer, src io.Reader) (int64, error) {
	return ThrottledCopy(dst, src, CPULimit)
}

// ThrottledCopy copies from src to dst while limiting CPU utilisation.
//
// It uses a closed-loop feedback controller: actual process CPU time (via
// getrusage / GetProcessTimes) is measured each throttle window and compared
// against the target ratio to compute the precise sleep duration. This makes
// the throttle accurate under cgroup CPU limits where wall-clock time and CPU
// time diverge. If the platform syscall is unavailable, it falls back to
// open-loop wall-clock duty-cycle control.
//
// cpuLimit controls the target CPU ratio (0.0, 1.0].
//   - cpuLimit <= 0 or >= 1.0: no throttling, delegates to io.Copy.
//   - cpuLimit  = 0.5: ~50 % CPU usage.
//   - cpuLimit  = 0.3: ~30 % CPU usage.
func ThrottledCopy(dst io.Writer, src io.Reader, cpuLimit float64) (int64, error) {
	if cpuLimit <= 0 || cpuLimit >= 1.0 {
		return io.Copy(dst, src)
	}

	buf := make([]byte, defaultBufSize)
	wallSleepRatio := 1.0/cpuLimit - 1.0

	var written int64
	var accWork time.Duration

	// Probe CPU-time feedback: use a sentinel so we can distinguish "syscall
	// unavailable" (sentinel unchanged) from "process has 0 CPU time" (overwritten
	// with 0).
	cpuSnap := [2]int64{-1, -1}
	readProcCPU(&cpuSnap)
	hasFeedback := cpuSnap[0] >= 0
	cpuBase := cpuSnap[0] + cpuSnap[1]
	wallBase := time.Now()

	for {
		start := time.Now()

		nr, readErr := src.Read(buf)
		if nr > 0 {
			nw, writeErr := dst.Write(buf[:nr])
			written += int64(nw)
			if writeErr != nil {
				return written, writeErr
			}
			if nw != nr {
				return written, io.ErrShortWrite
			}
		}

		accWork += time.Since(start)

		if accWork >= throttleWindow {
			var sleepDur time.Duration

			if hasFeedback {
				// Closed-loop: derive sleep from actual CPU consumption.
				//   target: cpuDelta / (wallDelta + sleep) = cpuLimit
				//   =>      sleep = cpuDelta/cpuLimit − wallDelta
				readProcCPU(&cpuSnap)
				cpuDelta := time.Duration(cpuSnap[0]+cpuSnap[1]-cpuBase) * time.Nanosecond
				wallDelta := time.Since(wallBase)
				sleepDur = time.Duration(float64(cpuDelta)/cpuLimit) - wallDelta
			} else {
				// Open-loop fallback: assume CPU time ≈ wall time.
				sleepDur = time.Duration(float64(accWork) * wallSleepRatio)
			}

			if sleepDur > maxSleepPerWindow {
				sleepDur = maxSleepPerWindow
			}
			if sleepDur > 0 {
				time.Sleep(sleepDur)
			}

			if hasFeedback {
				readProcCPU(&cpuSnap)
				cpuBase = cpuSnap[0] + cpuSnap[1]
				wallBase = time.Now()
			}
			accWork = 0
		}

		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return written, readErr
		}
	}

	return written, nil
}
