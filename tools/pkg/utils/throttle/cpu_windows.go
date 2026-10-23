//go:build windows

package throttle

import (
	"syscall"
	"unsafe"
)

// readProcCPU fills usage[0] with user CPU time and usage[1] with kernel CPU time
// of the current process, in nanoseconds.
func readProcCPU(usage *[2]int64) {
	var creationTime, exitTime, kernelTime, userTime syscall.Filetime

	handle, err := syscall.GetCurrentProcess()
	if err != nil {
		return
	}

	kernel32 := syscall.MustLoadDLL("kernel32.dll")
	proc := kernel32.MustFindProc("GetProcessTimes")

	ret, _, _ := proc.Call(
		uintptr(handle),
		uintptr(unsafe.Pointer(&creationTime)),
		uintptr(unsafe.Pointer(&exitTime)),
		uintptr(unsafe.Pointer(&kernelTime)),
		uintptr(unsafe.Pointer(&userTime)),
	)
	if ret == 0 {
		return
	}

	// FILETIME is in 100-nanosecond intervals
	usage[0] = fileTimeToNano(userTime)
	usage[1] = fileTimeToNano(kernelTime)
}

func fileTimeToNano(ft syscall.Filetime) int64 {
	// FILETIME is in 100-nanosecond intervals; combine then convert.
	ticks := int64(ft.HighDateTime)<<32 | int64(ft.LowDateTime)
	return ticks * 100
}
