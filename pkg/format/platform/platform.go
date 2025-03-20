/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package platform ...
package platform

import (
	"errors"
	"fmt"
	"strings"
)

// StandardOSMap this map is used to convert various OS names to Go standard GOOS values..
func StandardOSMap() map[string]string {
	return map[string]string{
		"aix":       "aix",
		"android":   "android",
		"darwin":    "darwin",
		"dragonfly": "dragonfly",
		"freebsd":   "freebsd",
		"hurd":      "hurd",
		"illumos":   "illumos",
		"ios":       "ios",
		"js":        "js",
		"linux":     "linux",
		"macos":     "darwin",
		"mac os":    "darwin",
		"mac os x":  "darwin",
		"mac":       "darwin",
		"netbsd":    "netbsd",
		"openbsd":   "openbsd",
		"plan9":     "plan9",
		"solaris":   "solaris",
		"wasip1":    "wasip1",
		"windows":   "windows",
		"win":       "windows",
		"win32":     "windows",
		"win64":     "windows",
		"wsl":       "linux",
		"zos":       "zos",
		// normalize some common Windows environments to "windows"
		"msys":     "windows",
		"msys2":    "windows",
		"mingw":    "windows",
		"mingw32":  "windows",
		"mingw64":  "windows",
		"cygwin":   "windows",
		"git bash": "windows",
		// normalize some common Linux distributions to "linux"
		"ubuntu":   "linux",
		"debian":   "linux",
		"centos":   "linux",
		"fedora":   "linux",
		"rhel":     "linux",
		"redhat":   "linux",
		"alpine":   "linux",
		"arch":     "linux",
		"manjaro":  "linux",
		"gentoo":   "linux",
		"opensuse": "linux",
	}
}

// NormalizeOS convert various OS names to Go standard GOOS values.
// if unable to recognize, return original input and error.
func NormalizeOS(osName string) (string, error) {
	if osName == "" {
		return "", errors.New("empty OS name")
	}

	normalized := strings.ToLower(strings.TrimSpace(osName))

	osMap := StandardOSMap()

	if goos, exists := osMap[normalized]; exists {
		return goos, nil
	}

	// try to match with common patterns
	for key, value := range osMap {
		if strings.Contains(normalized, key) {
			// cover windows and win32/win64
			if key == "win" && strings.Contains(normalized, "windows") {
				continue
			}

			return value, nil
		}
	}

	return osName, fmt.Errorf("unknown OS: %s", osName)
}

// StandardArchMap this map is used to convert various architecture names to Go standard GOARCH values.
func StandardArchMap() map[string]string {
	return map[string]string{
		"386":      "386",
		"aarch64":  "arm64",
		"amd64":    "amd64",
		"arm":      "arm",
		"arm64":    "arm64",
		"armv5":    "arm",
		"armv6":    "arm",
		"armv7":    "arm",
		"armv8":    "arm64",
		"i386":     "386",
		"i486":     "386",
		"i586":     "386",
		"i686":     "386",
		"ia32":     "386",
		"ia64":     "ia64",
		"loong64":  "loong64",
		"mips":     "mips",
		"mips64":   "mips64",
		"mips64le": "mips64le",
		"mipsle":   "mipsle",
		"ppc":      "ppc",
		"ppc64":    "ppc64",
		"ppc64le":  "ppc64le",
		"riscv":    "riscv",
		"riscv64":  "riscv64",
		"s390":     "s390",
		"s390x":    "s390x",
		"sparc":    "sparc",
		"sparc64":  "sparc64",
		"wasm":     "wasm",
		"x64":      "amd64",
		"x86":      "386",
		"x86-64":   "amd64",
		"x86_64":   "amd64",
	}
}

// NormalizeArch convert various architecture names to Go standard GOARCH values.
// if unable to recognize, return original input and error.
func NormalizeArch(arch string) (string, error) {
	if arch == "" {
		return "", errors.New("empty architecture name")
	}

	normalized := strings.ToLower(strings.TrimSpace(arch))

	archMap := StandardArchMap()

	// try to match with common patterns.
	if goarch, exists := archMap[normalized]; exists {
		return goarch, nil
	}

	// try to match with common patterns
	if strings.HasPrefix(normalized, "armv") {
		ver := normalized[4:]
		// over the armv8, we think it is arm64
		if v, err := parseArmVersion(ver); err == nil && v >= 8 {
			return "arm64", nil
		}

		return "arm", nil
	}

	return arch, fmt.Errorf("unknown architecture: %s", arch)
}

// parseArmVersion try to extract version number from ARM version string.
func parseArmVersion(version string) (int, error) {
	// find the first character of the version
	for i, c := range version {
		if c >= '0' && c <= '9' {
			// convert numeric characters to int
			return int(c - '0'), nil
		}
		if i > 0 {
			// if the number is not found immediately, the error will be returned
			break
		}
	}

	return 0, fmt.Errorf("unable to parse ARM version, version(%s)", version)
}

// Platform represents a platform identifier.
type Platform struct {
	OS   string
	Arch string
}

// Normalize standardize the platform into go standard naming.
func Normalize(osName, arch string) (Platform, error) {
	normalizedOS, osErr := NormalizeOS(osName)
	normalizedArch, archErr := NormalizeArch(arch)

	var combinedErr error

	switch {
	case osErr != nil && archErr != nil:
		combinedErr = fmt.Errorf("both OS and Arch errors: %v, %v", osErr, archErr)
	case osErr != nil:
		combinedErr = osErr
	case archErr != nil:
		combinedErr = archErr
	default:
		// no error
	}

	return Platform{
		OS:   normalizedOS,
		Arch: normalizedArch,
	}, combinedErr
}

// String returns the string representation of the platform.
func (p Platform) String() string {
	return fmt.Sprintf("%s/%s", p.OS, p.Arch)
}

// ValidPlatforms is a map of valid platform combinations supported by Go.
func ValidPlatforms() map[string]bool {
	return map[string]bool{
		"aix/ppc64":       true,
		"android/386":     true,
		"android/amd64":   true,
		"android/arm":     true,
		"android/arm64":   true,
		"darwin/amd64":    true,
		"darwin/arm64":    true,
		"dragonfly/amd64": true,
		"freebsd/386":     true,
		"freebsd/amd64":   true,
		"freebsd/arm":     true,
		"freebsd/arm64":   true,
		"freebsd/riscv64": true,
		"illumos/amd64":   true,
		"ios/amd64":       true,
		"ios/arm64":       true,
		"js/wasm":         true,
		"linux/386":       true,
		"linux/amd64":     true,
		"linux/arm":       true,
		"linux/arm64":     true,
		"linux/loong64":   true,
		"linux/mips":      true,
		"linux/mips64":    true,
		"linux/mips64le":  true,
		"linux/mipsle":    true,
		"linux/ppc64":     true,
		"linux/ppc64le":   true,
		"linux/riscv64":   true,
		"linux/s390x":     true,
		"netbsd/386":      true,
		"netbsd/amd64":    true,
		"netbsd/arm":      true,
		"netbsd/arm64":    true,
		"openbsd/386":     true,
		"openbsd/amd64":   true,
		"openbsd/arm":     true,
		"openbsd/arm64":   true,
		"openbsd/mips64":  true,
		"openbsd/ppc64":   true,
		"openbsd/riscv64": true,
		"plan9/386":       true,
		"plan9/amd64":     true,
		"plan9/arm":       true,
		"solaris/amd64":   true,
		"wasip1/wasm":     true,
		"windows/386":     true,
		"windows/amd64":   true,
		"windows/arm":     true,
		"windows/arm64":   true,
	}
}

// Validate Verify that the platform configuration is in a valid combination supported by Go.
func (p Platform) Validate() bool {
	platformKey := strings.ToLower(p.String())
	return ValidPlatforms()[platformKey]
}
