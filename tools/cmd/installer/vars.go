/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

// ToolPrefix tool prefix.
// this prefix is used to identify the file or directory created by this tool.
const ToolPrefix = "installer_data"

// logFilePath log file path.
// nolint: gochecknoglobals
var logFilePath = struct {
	sync.Once
	filePath string
}{}

// GetLogFilePath get log file.
func GetLogFilePath() string {
	logFilePath.Do(func() {
		logFilePath.filePath = filepath.Join(GetTmpDir(), "logs", "installer.log")
	})

	return logFilePath.filePath
}

// SetLogFilePath set log file path.
func SetLogFilePath(filePath string) error {
	var err error
	logFilePath.Do(func() {
		if filePath == "" {
			err = errors.New("log file path is empty")
			return
		}

		filePath, err = filepath.Abs(filePath)
		if err != nil {
			err = fmt.Errorf("set log file path failed, err: %v", err)
			return
		}

		logFilePath.filePath = filePath
	})
	if err != nil {
		return err
	}

	return nil
}

const gsePkgPrefix = "gse_pkg_"
const gsePkgExt = ".tar.gz"
const gseSep = "_"

// GetGsePkgName get gse agent pkg name.
func GetGsePkgName() string {
	gsePkg.Do(func() {
		gsePkg.pkgName = strings.Join(
			[]string{
				string(node.nodeRole),
				runtime.GOOS,
				runtime.GOARCH,
				strconv.Itoa(node.generation),
				node.version,
			}, gseSep)
		gsePkg.pkgName = fmt.Sprintf("%s%s%s", gsePkgPrefix, gsePkg.pkgName, gsePkgExt)
	})

	return gsePkg.pkgName
}

// nolint:gochecknoglobals
var (
	gsePkg = struct {
		sync.Once
		pkgName string
	}{}
)

// GetGsePkgPath get gse pkg path.
func GetGsePkgPath() string {
	return filepath.Join(GetTmpDir(), GetGsePkgName())
}

// SetPkgName set gse agent pkg name.
func SetPkgName(pkgName string) error {
	var err error
	gsePkg.Do(func() {
		if pkgName == "" {
			err = errors.New("pkg name is empty")
			return
		}

		gsePkg.pkgName = pkgName
	})
	if err != nil {
		return err
	}

	return nil
}

// nolint:gochecknoglobals
var setupDirPath = struct {
	sync.Once
	dirPath string
}{}

// GetSetupDir get setup dir.
func GetSetupDir() string {
	setupDirPath.Do(func() {
		setupDirPath.dirPath = filepath.Join(node.gseRoot, string(node.nodeRole))
	})

	return setupDirPath.dirPath
}

// SetSetupDir set setup dir.
func SetSetupDir(dirPath string) error {
	var err error
	setupDirPath.Do(func() {
		if dirPath == "" {
			err = errors.New("dir path is empty")
			return
		}

		dirPath, err = filepath.Abs(dirPath)
		if err != nil {
			return
		}

		setupDirPath.dirPath = filepath.Clean(filepath.FromSlash(dirPath))
	})
	if err != nil {
		return fmt.Errorf("set setup dir failed, err: %w", err)
	}

	return nil
}

// nolint:gochecknoglobals
var runDirPath = struct {
	sync.Once
	dirPath string
}{}

// GetRunDir get run dir.
func GetRunDir() string {
	runDirPath.Do(func() {
		runDirPath.dirPath = filepath.Join(GetSetupDir(), "bin", "run")
	})

	return runDirPath.dirPath
}

// SetRunDir set run dir.
func SetRunDir(dirPath string) error {
	var err error
	runDirPath.Do(func() {
		if dirPath == "" {
			err = errors.New("dir path is empty")
			return
		}

		dirPath, err = filepath.Abs(dirPath)
		if err != nil {
			return
		}

		runDirPath.dirPath = filepath.Clean(filepath.FromSlash(dirPath))
	})
	if err != nil {
		return fmt.Errorf("set run dir failed, err: %w", err)
	}

	return nil
}

// nolint: gochecknoglobals
var workspaceDir = struct {
	sync.Once
	dirPath string
}{}

// SetWorkspaceDir workspace tmp dir.
func SetWorkspaceDir(dir string) error {
	var err error
	workspaceDir.Do(func() {
		dir, err = filepath.Abs(dir)
		if err != nil {
			return
		}

		workspaceDir.dirPath = filepath.Clean(filepath.FromSlash(dir))
	})
	if err != nil {
		return fmt.Errorf("set workspace dir failed, err: %w", err)
	}

	return nil
}

// GetTmpDir get tmp dir.
func GetTmpDir() string {
	return filepath.Join(workspaceDir.dirPath, ToolPrefix)
}

// GetTmpConfigDir get tmp config dir.
func GetTmpConfigDir() string {
	return filepath.Join(GetTmpDir(), "configs")
}

const baseNameAgent = "gse_agent"

// GetGseAgentName get gse agent name.
func GetGseAgentName() string {
	osType := runtime.GOOS
	switch osType {
	case "windows":
		return baseNameAgent + ".exe"
	default:
		return baseNameAgent
	}
}

// nolint: gochecknoglobals
var gseAgentPath = struct {
	sync.Once
	filePath string
}{}

// GetGseAgentPath get gse agent path.
func GetGseAgentPath() string {
	gseAgentPath.Do(func() {
		gseAgentPath.filePath = filepath.Join(GetSetupDir(), "bin", GetGseAgentName())
	})

	return gseAgentPath.filePath
}

const baseNameAgentCtl = "gsectl"

// GetGseAgentCtlName gse agent ctl name.
func GetGseAgentCtlName() string {
	osType := runtime.GOOS
	switch osType {
	case "windows":
		return baseNameAgentCtl + ".bat"
	default:
		return baseNameAgentCtl
	}
}

// nolint: gochecknoglobals
var gseAgentCtlPath = struct {
	sync.Once
	filePath string
}{}

// GetGseCtlPath get gse agent ctl path.
func GetGseCtlPath() string {
	gseAgentCtlPath.Do(func() {
		gseAgentCtlPath.filePath = filepath.Join(GetSetupDir(), "bin", GetGseAgentCtlName())
	})

	return gseAgentCtlPath.filePath
}

// nolint: gochecknoglobals
var preCheckFilePath = struct {
	sync.Once
	filePath string
}{}

// GetPreCheckFilePath get preCheck file path.
func GetPreCheckFilePath() string {
	preCheckFilePath.Do(func() {
		preCheckFilePath.filePath = filepath.Join(GetTmpDir(), "precheck.json")
	})

	return preCheckFilePath.filePath
}

// SetPreCheckFilePath set preCheck file path.
func SetPreCheckFilePath(filePath string) error {
	var err error
	preCheckFilePath.Do(func() {
		filePath, err = filepath.Abs(filepath.Clean(filePath))
		if err != nil {
			return
		}

		_, err = os.Stat(filePath)
		if err != nil {
			return
		}

		preCheckFilePath.filePath = filePath
	})
	if err != nil {
		return fmt.Errorf("set precheck file path failed, err: %w", err)
	}

	return nil
}

// nolint: gochecknoglobals
var token = struct {
	sync.Once
	token string
}{}

// GetToken get token.
func GetToken() string {
	return token.token
}

// SetToken set token.
func SetToken(tokenStr string) error {
	var err error
	token.Do(func() {
		if tokenStr == "" {
			err = errors.New("token is empty")
			return
		}

		token.token = tokenStr
	})
	if err != nil {
		return fmt.Errorf("set token failed, err: %w", err)
	}

	return nil
}
