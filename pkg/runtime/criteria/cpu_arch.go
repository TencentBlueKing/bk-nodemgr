/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package criteria

import "fmt"

// CPUArch define the cpu architecture.
type CPUArch string

const (
	// CPUArch386 this defines the cpu arch of 386.
	CPUArch386 CPUArch = "386"

	// CPUArchArm this defines the cpu arch of arm.
	CPUArchArm CPUArch = "arm"

	// CPUArchArm64 this defines the cpu arch of arm64.
	CPUArchArm64 CPUArch = "arm64"

	// CPUArchAmd64 this defines the cpu arch of amd64.
	CPUArchAmd64 CPUArch = "amd64"

	// CPUArchLoong64 this defines the cpu arch of loong64.
	CPUArchLoong64 CPUArch = "loong64"

	// CPUArchMips this defines the cpu arch of mips.
	CPUArchMips CPUArch = "mips"

	// CPUArchMipsle this defines the cpu arch of mipsle.
	CPUArchMipsle CPUArch = "mipsle"

	// CPUArchMips64 this defines the cpu arch of mips64.
	CPUArchMips64 CPUArch = "mips64"

	// CPUArchMips64le this defines the cpu arch of mips64le.
	CPUArchMips64le CPUArch = "mips64le"

	// CPUArchPpc this defines the cpu arch of ppc.
	CPUArchPpc CPUArch = "ppc"

	// CPUArchPpc64 this defines the cpu arch of ppc64.
	CPUArchPpc64 CPUArch = "ppc64"

	// CPUArchPpc64le this defines the cpu arch of ppc64le.
	CPUArchPpc64le CPUArch = "ppc64le"

	// CPUArchRiscv this defines the cpu arch of riscv.
	CPUArchRiscv CPUArch = "riscv"

	// CPUArchRiscv64 this defines the cpu arch of riscv64.
	CPUArchRiscv64 CPUArch = "riscv64"

	// CPUArchS390 this defines the cpu arch of s390.
	CPUArchS390 CPUArch = "s390"

	// CPUArchS390x this defines the cpu arch of s390x.
	CPUArchS390x CPUArch = "s390x"

	// CPUArchSparc this defines the cpu arch of sparc.
	CPUArchSparc CPUArch = "sparc"

	// CPUArchSparc64 this defines the cpu arch of sparc64.
	CPUArchSparc64 CPUArch = "sparc64"

	// CPUArchWasm this defines the cpu arch of wasm.
	CPUArchWasm CPUArch = "wasm"

	// CPUArchUnknown this defines the cpu arch of unknown.
	CPUArchUnknown CPUArch = "unknown"
)

// Validate checks if the cpu arch is valid.
func (arch CPUArch) Validate() error {
	switch arch {
	case CPUArch386, CPUArchArm, CPUArchArm64, CPUArchAmd64, CPUArchLoong64, CPUArchMips,
		CPUArchMipsle, CPUArchMips64, CPUArchMips64le, CPUArchPpc, CPUArchPpc64, CPUArchPpc64le,
		CPUArchRiscv, CPUArchRiscv64, CPUArchS390, CPUArchS390x, CPUArchSparc, CPUArchSparc64,
		CPUArchWasm, CPUArchUnknown:
		return nil
	default:
		return fmt.Errorf("invalid cpu arch: %s", arch)
	}
}

// String returns the string representation of the CPUArch.
func (arch CPUArch) String() string {
	return string(arch)
}

// ToPkgArch convert the CPUArch to the package architecture name.
func (arch CPUArch) ToPkgArch() string {
	pkgArch := map[CPUArch]string{
		CPUArch386:      "x86",
		CPUArchAmd64:    "x86_64",
		CPUArchArm:      "arm",
		CPUArchArm64:    "aarch64",
		CPUArchLoong64:  "loong64",
		CPUArchMips:     "mips",
		CPUArchMipsle:   "mipsle",
		CPUArchMips64:   "mips64",
		CPUArchMips64le: "mips64le",
		CPUArchPpc:      "powerpc",
		CPUArchPpc64:    "ppc64",
		CPUArchPpc64le:  "ppc64le",
		CPUArchRiscv:    "riscv",
		CPUArchRiscv64:  "riscv64",
		CPUArchS390:     "s390",
		CPUArchS390x:    "s390x",
		CPUArchSparc:    "sparc",
		CPUArchSparc64:  "sparc64",
		CPUArchWasm:     "wasm",
		CPUArchUnknown:  "unknown",
	}

	if val, ok := pkgArch[arch]; ok {
		return val
	}

	return "unknown"
}
