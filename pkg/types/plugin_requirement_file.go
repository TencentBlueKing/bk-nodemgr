/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package types

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
)

// PluginRequirementFileRecord stores the latest metadata of an independent requirement file.
type PluginRequirementFileRecord struct {
	ID        string
	Name      string
	SavedName string
	OS        criteria.OSType
	Arch      criteria.CPUArch
	Version   string
	MD5       string
	Size      int64
	UpdatedAt time.Time
	Operator  string
}

// PluginRequirementScopeAll applies a requirement file to every value of a dimension.
const PluginRequirementScopeAll = "all"

// PluginRequirementFileScope describes the applicable plugin package dimensions.
type PluginRequirementFileScope struct {
	OS      criteria.OSType
	Arch    criteria.CPUArch
	Version string
}

// Validate checks concrete OS and architecture values and an explicit package version or all.
func (scope PluginRequirementFileScope) Validate() error {
	if scope.OS == criteria.OSUnknown || scope.Arch == criteria.CPUArchUnknown {
		return errors.New("plugin requirement scope must use concrete values or all")
	}
	if scope.OS != PluginRequirementScopeAll {
		if err := scope.OS.Validate(); err != nil {
			return fmt.Errorf("invalid plugin requirement os: %w", err)
		}
	}
	if scope.Arch != PluginRequirementScopeAll {
		if err := scope.Arch.Validate(); err != nil {
			return fmt.Errorf("invalid plugin requirement arch: %w", err)
		}
	}
	if scope.Version == "" {
		return errors.New("plugin requirement version is required")
	}

	return nil
}

// PluginRequirementFileExactFields selects administrative file metadata by exact values.
type PluginRequirementFileExactFields struct {
	ID      []string
	Name    []string
	OS      []criteria.OSType
	Arch    []criteria.CPUArch
	Version []string
}

// PluginRequirementFileCondition contains administrative filters, not applicability matching rules.
type PluginRequirementFileCondition struct {
	ExactInclude *PluginRequirementFileExactFields
}

// ValidatePluginRequirementFileName validates a single portable requirement filename.
func ValidatePluginRequirementFileName(name string) error {
	if name == "" || name == "." || name == ".." {
		return errors.New("plugin requirement filename must not be empty or a directory reference")
	}
	if strings.ContainsAny(name, `/\:<>"|?*`) || strings.ContainsFunc(name, unicode.IsControl) {
		return errors.New("plugin requirement filename contains invalid characters")
	}
	if strings.HasSuffix(name, " ") || strings.HasSuffix(name, ".") {
		return errors.New("plugin requirement filename must not end with a space or dot")
	}
	if isReservedPluginRequirementFileName(name) {
		return errors.New("plugin requirement filename is a reserved device name")
	}

	return nil
}

func isReservedPluginRequirementFileName(name string) bool {
	baseName, _, _ := strings.Cut(name, ".")
	baseName = strings.ToUpper(baseName)
	switch baseName {
	case "CON", "PRN", "AUX", "NUL":
		return true
	}

	return len(baseName) == 4 && (strings.HasPrefix(baseName, "COM") || strings.HasPrefix(baseName, "LPT")) &&
		baseName[3] >= '1' && baseName[3] <= '9'
}
