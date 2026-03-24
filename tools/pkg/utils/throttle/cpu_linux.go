//go:build linux || darwin || freebsd || aix

package throttle

import (
	"syscall"
	"time"
)

// readProcCPU fills usage[0] with user CPU time and usage[1] with system CPU time
// of the current process, in nanoseconds.
func readProcCPU(usage *[2]int64) {
	var ru syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &ru); err != nil {
		return
	}
	usage[0] = time.Duration(ru.Utime.Nano()).Nanoseconds()
	usage[1] = time.Duration(ru.Stime.Nano()).Nanoseconds()
}
