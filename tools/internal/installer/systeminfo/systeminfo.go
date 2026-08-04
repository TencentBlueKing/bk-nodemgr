// Package systeminfo logs best-effort target machine diagnostics for installer startup.
package systeminfo

import "github.com/TencentBlueKing/bk-nodemgr/tools/pkg/logger"

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

		logger.Debugf(step, fieldLogFormat, field.key, value)
	}
}
