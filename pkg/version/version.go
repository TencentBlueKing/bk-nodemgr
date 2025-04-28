/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package version provides version info.
package version

import (
	"fmt"
)

const (
	// LOGO is nodemgr logo.
	LOGO = `
======================================================================================
                                                                                      
    _   _           _       __  __                                                    
   | \ | |         | |     |  \/  |                                                   
   |  \| | ___   __| | ___ |_/\/_| __ _ _ __                                         
   | . \ |/ _ \ / _\ |/ _ \| |\/| |/ _\ | '__|                                        
   | |\  | (_) | (_| |  __/| |  | | (_| | |                                           
   |_| \_|\___/ \__,_|\___||_|  |_|\__, |_|                                           
                                     __/ |                                            
                                    |___/                                             
                                                                                       
======================================================================================`
)

var (
	// VERSION is version info.
	VERSION = "debug"

	// BUILDTIME  build time.
	BUILDTIME = "unknown"

	// GITHASH git hash for release.
	GITHASH = "unknown"
)

// ShowVersion shows the version info.
func ShowVersion() {
	fmt.Println(FormatVersion())
}

// FormatVersion returns service's version.
func FormatVersion() string {
	return fmt.Sprintf("\nVersion: %s\nBuildTime: %s\nGitHash: %s\n", VERSION, BUILDTIME, GITHASH)
}

// GetStartInfo returns start info that includes version and logo.
func GetStartInfo() string {
	startInfo := fmt.Sprintf("%s\n\n%s\n", LOGO, FormatVersion())
	return startInfo
}

// Version ...
func Version() *SysVersion {
	return &SysVersion{
		Version: VERSION,
		Hash:    GITHASH,
		Time:    BUILDTIME,
	}
}

// SysVersion describe a binary version
type SysVersion struct {
	Version string `json:"version"`
	Hash    string `json:"hash"`
	Time    string `json:"time"`
}
