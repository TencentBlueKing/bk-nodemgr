package systeminfo

import (
	"fmt"
	"os"
	"runtime"
	"time"
)

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
