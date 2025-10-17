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

// OSType define the os type.
type OSType string

const (
	// OSAix this defines the os type of aix.
	OSAix OSType = "aix"

	// OSAix6 this defines the os type of aix6.
	OSAix6 OSType = "aix6"

	// OSAix7 this defines the os type of aix7.
	OSAix7 OSType = "aix7"

	// OSAndroid this defines the os type of android.
	OSAndroid OSType = "android"

	// OSDarwin this defines the os type of darwin.
	OSDarwin OSType = "darwin"

	// OSDragonfly this defines the os type of dragonfly.
	OSDragonfly OSType = "dragonfly"

	// OSFreebsd this defines the os type of freebsd.
	OSFreebsd OSType = "freebsd"

	// OSHurd this defines the os type of hurd.
	OSHurd OSType = "hurd"

	// OSIllumos this defines the os type of illumos.
	OSIllumos OSType = "illumos"

	// OSIos this defines the os type of ios.
	OSIos OSType = "ios"

	// OSJs this defines the os type of js.
	OSJs OSType = "js"

	// OSLinux this defines the os type of linux.
	OSLinux OSType = "linux"

	// OSNetbsd this defines the os type of netbsd.
	OSNetbsd OSType = "netbsd"

	// OSOpenbsd this defines the os type of openbsd.
	OSOpenbsd OSType = "openbsd"

	// OSPlan9 this defines the os type of plan9.
	OSPlan9 OSType = "plan9"

	// OSSolaris this defines the os type of solaris.
	OSSolaris OSType = "solaris"

	// OSWasip1 this defines the os type of wasip1.
	OSWasip1 OSType = "wasip1"

	// OSWindows this defines the os type of windows.
	OSWindows OSType = "windows"

	// OSZos this defines the os type of zos.
	OSZos OSType = "zos"

	// OSUnknown this defines the os type of unknown.
	OSUnknown OSType = "unknown"
)

// Validate checks if the os type is valid.
func (os OSType) Validate() error {
	switch os {
	case OSAix, OSAix6, OSAix7, OSAndroid, OSDarwin, OSDragonfly, OSFreebsd, OSHurd,
		OSIllumos, OSIos, OSJs, OSLinux, OSNetbsd, OSOpenbsd,
		OSPlan9, OSSolaris, OSWasip1, OSWindows, OSZos, OSUnknown:
		return nil
	default:
		return fmt.Errorf("invalid os type: %s", os)
	}
}

// String converts the OSType to its string representation.
func (os OSType) String() string {
	return string(os)
}

// StringListToOSTypeList converts a string list to a os type list.
func StringListToOSTypeList(stringList []string) ([]OSType, error) {
	data := make([]OSType, len(stringList))
	for idx, osType := range stringList {
		if err := OSType(osType).Validate(); err != nil {
			return nil, err
		}
		data[idx] = OSType(osType)
	}

	return data, nil
}

// OSTypeListToStringList converts a os type list to a string list.
func OSTypeListToStringList(osTypeList []OSType) []string {
	data := make([]string, len(osTypeList))
	for idx, osType := range osTypeList {
		data[idx] = string(osType)
	}

	return data
}
