//go:build !windows

package systeminfo

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

const (
	osReleasePath     = "/etc/os-release"
	kernelReleasePath = "/proc/sys/kernel/osrelease"
)

func collectSystemVersion() (string, error) {
	content, err := readOptionalFile(osReleasePath)
	if err != nil || content == "" {
		return "", err
	}

	return parseOSReleasePrettyName(content), nil
}

func collectKernelVersion() (string, error) {
	if runtime.GOOS != "linux" {
		return "", nil
	}

	content, err := readOptionalFile(kernelReleasePath)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(content), nil
}

func readOptionalFile(path string) (string, error) {
	content, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}

		return "", err
	}

	return string(content), nil
}

func parseOSReleasePrettyName(content string) string {
	for _, line := range strings.Split(content, "\n") {
		key, value, ok := strings.Cut(line, "=")
		if !ok || strings.TrimSpace(key) != "PRETTY_NAME" {
			continue
		}

		return unquoteOSReleaseValue(value)
	}

	return ""
}

func unquoteOSReleaseValue(value string) string {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		return ""
	}

	unquotedValue, err := strconv.Unquote(trimmedValue)
	if err != nil {
		return trimmedValue
	}

	return strings.TrimSpace(unquotedValue)
}
