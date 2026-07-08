// Package systeminfo logs best-effort target machine diagnostics for installer startup.
package systeminfo

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/logger"
)

const (
	unknownValue    = "unknown"
	startMessage    = "installer initial target machine information:"
	fieldLogFormat  = "target machine info: %s=%s"
	fieldWarnFormat = "failed to collect target machine info field(%s): %v"
)

type targetInfoField struct {
	key     string
	collect func() (string, error)
}

var targetInfoFields = []targetInfoField{
	{key: "goos", collect: collectGOOS},
	{key: "goarch", collect: collectGOARCH},
	{key: "hostname", collect: collectHostname},
	{key: "current_time", collect: collectCurrentTime},
	{key: "timezone", collect: collectTimezone},
	{key: "timezone_offset", collect: collectTimezoneOffset},
	{key: "system_version", collect: collectSystemVersion},
	{key: "kernel_version", collect: collectKernelVersion},
}

// LogInitialTargetInfo logs best-effort target machine information for installer startup diagnostics.
func LogInitialTargetInfo(step logger.Step) {
	logger.Infof(step, startMessage)

	for _, field := range targetInfoFields {
		value, err := field.collect()
		if err != nil {
			logger.Warnf(step, fieldWarnFormat, field.key, err)
			value = unknownValue
		}
		if value == "" {
			value = unknownValue
		}

		logger.Infof(step, fieldLogFormat, field.key, value)
	}
}

func collectGOOS() (string, error) {
	return runtime.GOOS, nil
}

func collectGOARCH() (string, error) {
	return runtime.GOARCH, nil
}

func collectHostname() (string, error) {
	return os.Hostname()
}

func collectCurrentTime() (string, error) {
	return time.Now().Format(time.RFC3339), nil
}

func collectTimezone() (string, error) {
	name, _ := time.Now().Zone()
	return name, nil
}

func collectTimezoneOffset() (string, error) {
	_, offset := time.Now().Zone()
	return formatTimezoneOffset(offset), nil
}

func collectSystemVersion() (string, error) {
	return "", nil
}

func collectKernelVersion() (string, error) {
	return "", nil
}

func formatTimezoneOffset(offsetSeconds int) string {
	sign := "+"
	if offsetSeconds < 0 {
		sign = "-"
		offsetSeconds = -offsetSeconds
	}

	hours := offsetSeconds / 3600
	minutes := offsetSeconds % 3600 / 60
	return fmt.Sprintf("%s%02d:%02d", sign, hours, minutes)
}
