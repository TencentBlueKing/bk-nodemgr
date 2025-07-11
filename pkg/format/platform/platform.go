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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
)

// StandardOSMap this map is used to convert various OS names to Go standard GOOS values..
func StandardOSMap() map[string]criteria.OSType {
	return map[string]criteria.OSType{
		"aix":       criteria.OSAix,
		"android":   criteria.OSAndroid,
		"darwin":    criteria.OSDarwin,
		"dragonfly": criteria.OSDragonfly,
		"freebsd":   criteria.OSFreebsd,
		"hurd":      criteria.OSHurd,
		"illumos":   criteria.OSIllumos,
		"ios":       criteria.OSIos,
		"js":        criteria.OSJs,
		"linux":     criteria.OSLinux,
		"macos":     criteria.OSDarwin,
		"mac os":    criteria.OSDarwin,
		"mac os x":  criteria.OSDarwin,
		"mac":       criteria.OSDarwin,
		"netbsd":    criteria.OSNetbsd,
		"openbsd":   criteria.OSOpenbsd,
		"plan9":     criteria.OSPlan9,
		"solaris":   criteria.OSSolaris,
		"wasip1":    criteria.OSWasip1,
		"windows":   criteria.OSWindows,
		"win":       criteria.OSWindows,
		"win32":     criteria.OSWindows,
		"win64":     criteria.OSWindows,
		"wsl":       criteria.OSLinux,
		"zos":       criteria.OSZos,
		// normalize some common Windows environments to "windows"
		"msys":     criteria.OSWindows,
		"msys2":    criteria.OSWindows,
		"mingw":    criteria.OSWindows,
		"mingw32":  criteria.OSWindows,
		"mingw64":  criteria.OSWindows,
		"cygwin":   criteria.OSWindows,
		"git bash": criteria.OSWindows,
		// normalize some common Linux distributions to "linux"
		"ubuntu":   criteria.OSLinux,
		"debian":   criteria.OSLinux,
		"centos":   criteria.OSLinux,
		"fedora":   criteria.OSLinux,
		"rhel":     criteria.OSLinux,
		"redhat":   criteria.OSLinux,
		"alpine":   criteria.OSLinux,
		"arch":     criteria.OSLinux,
		"manjaro":  criteria.OSLinux,
		"gentoo":   criteria.OSLinux,
		"opensuse": criteria.OSLinux,
	}
}

// NormalizeOS convert various OS names to Go standard GOOS values.
// if unable to recognize, return original input and error.
func NormalizeOS(osName string) (criteria.OSType, error) {
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

	return "", fmt.Errorf("unknown OS: %s", osName)
}

// StandardArchMap this map is used to convert various architecture names to Go standard GOARCH values.
func StandardArchMap() map[string]criteria.CPUArch {
	return map[string]criteria.CPUArch{
		"386":      criteria.CPUArch386,
		"aarch64":  criteria.CPUArchArm64,
		"amd64":    criteria.CPUArchAmd64,
		"arm":      criteria.CPUArchArm,
		"arm64":    criteria.CPUArchArm64,
		"armv5":    criteria.CPUArchArm,
		"armv6":    criteria.CPUArchArm,
		"armv7":    criteria.CPUArchArm,
		"armv8":    criteria.CPUArchArm64,
		"i386":     criteria.CPUArch386,
		"i486":     criteria.CPUArch386,
		"i586":     criteria.CPUArch386,
		"i686":     criteria.CPUArch386,
		"ia32":     criteria.CPUArch386,
		"loong64":  criteria.CPUArchLoong64,
		"mips":     criteria.CPUArchMips,
		"mips64":   criteria.CPUArchMips64,
		"mips64le": criteria.CPUArchMips64le,
		"mipsle":   criteria.CPUArchMipsle,
		"ppc":      criteria.CPUArchPpc,
		"ppc64":    criteria.CPUArchPpc64,
		"ppc64le":  criteria.CPUArchPpc64le,
		"riscv":    criteria.CPUArchRiscv,
		"riscv64":  criteria.CPUArchRiscv64,
		"s390":     criteria.CPUArchS390,
		"s390x":    criteria.CPUArchS390x,
		"sparc":    criteria.CPUArchSparc,
		"sparc64":  criteria.CPUArchSparc64,
		"wasm":     criteria.CPUArchWasm,
		"x64":      criteria.CPUArchAmd64,
		"x86":      criteria.CPUArch386,
		"x86-64":   criteria.CPUArchAmd64,
		"x86_64":   criteria.CPUArchAmd64,
	}
}

// NormalizeArch convert various architecture names to Go standard GOARCH values.
// if unable to recognize, return original input and error.
func NormalizeArch(arch string) (criteria.CPUArch, error) {
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

	return "", fmt.Errorf("unknown architecture: %s", arch)
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
	OS   criteria.OSType
	Arch criteria.CPUArch
}

// EmptyPlatform returns an empty platform.
func EmptyPlatform() Platform {
	return Platform{}
}

// Normalize standardize the platform into go standard naming.
func Normalize(osName, arch string) (Platform, error) {
	normalizedOS, osErr := NormalizeOS(osName)
	normalizedArch, archErr := NormalizeArch(arch)

	var combinedErr error

	switch {
	case osErr != nil && archErr != nil:
		combinedErr = fmt.Errorf("both OS and CPUArch errors: %v, %v", osErr, archErr)
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
