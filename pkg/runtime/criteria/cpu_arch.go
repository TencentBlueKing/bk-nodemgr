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
