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
	RequirementFileID string
	PluginName        string
	Name              string
	TargetFileName    string // User-chosen filename used when distributing the file.
	FileName          string // Actual FileGroup object name (UUID), not the distribution filename.
	OS                []criteria.OSType
	Arch              []criteria.CPUArch
	Version           []string
	MD5               string
	Size              int64
	UpdatedAt         time.Time
	Operator          string
}

// PluginRequirementFileKey identifies a named requirement within a plugin and tenant.
type PluginRequirementFileKey struct {
	PluginName string
	Name       string
}

// Validate checks the structural validity of the business identity.
func (key PluginRequirementFileKey) Validate() error {
	if strings.TrimSpace(key.PluginName) == "" || strings.ContainsFunc(key.PluginName, unicode.IsControl) {
		return errors.New("plugin requirement plugin name must not be blank or contain control characters")
	}
	if err := ValidatePluginRequirementName(key.Name); err != nil {
		return fmt.Errorf("invalid plugin requirement name: %w", err)
	}

	return nil
}

// GetUniqueKey returns the business identity independently of the stable record ID.
func (record *PluginRequirementFileRecord) GetUniqueKey() PluginRequirementFileKey {
	return PluginRequirementFileKey{PluginName: record.PluginName, Name: record.Name}
}

// ValidatePluginRequirementName validates a user-chosen dependency name, not a filename.
func ValidatePluginRequirementName(name string) error {
	if strings.TrimSpace(name) == "" || strings.ContainsFunc(name, unicode.IsControl) {
		return errors.New("plugin requirement name must not be blank or contain control characters")
	}

	return nil
}

// PluginRequirementFileScope describes applicable plugin package dimensions.
// Empty dimensions are unrestricted; values within a dimension use OR and dimensions use AND.
type PluginRequirementFileScope struct {
	OS      []criteria.OSType
	Arch    []criteria.CPUArch
	Version []string
}

// Validate checks that every supplied dimension value is concrete.
func (scope PluginRequirementFileScope) Validate() error {
	for _, os := range scope.OS {
		if os == criteria.OSUnknown {
			return errors.New("plugin requirement os must be concrete")
		}
		if err := os.Validate(); err != nil {
			return fmt.Errorf("invalid plugin requirement os: %w", err)
		}
	}
	for _, arch := range scope.Arch {
		if arch == criteria.CPUArchUnknown {
			return errors.New("plugin requirement arch must be concrete")
		}
		if err := arch.Validate(); err != nil {
			return fmt.Errorf("invalid plugin requirement arch: %w", err)
		}
	}
	for _, version := range scope.Version {
		if strings.TrimSpace(version) == "" || strings.EqualFold(strings.TrimSpace(version), "all") {
			return errors.New("plugin requirement version must be concrete")
		}
	}

	return nil
}

// PluginRequirementFileExactFields selects administrative file metadata by exact values.
type PluginRequirementFileExactFields struct {
	RequirementFileID []string
	PluginName        []string
	Name              []string
	TargetFileName    []string
	OS                []criteria.OSType
	Arch              []criteria.CPUArch
	Version           []string
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
