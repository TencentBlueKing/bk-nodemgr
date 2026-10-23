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

// Package configfile resolves service config files for adminclient commands:
// either the explicit -f/--file path, or auto-discovery of the first *_conf.yaml
// containing a given top-level section under --conf-dir.
package configfile

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/spf13/cobra"
)

const (
	// DefaultDir is the default directory of service config files.
	DefaultDir = "/bk-nodemgr/etc"
	// fileSuffix is the file name suffix of service config files.
	fileSuffix = "_conf.yaml"
)

// Flags carries the config file flags shared by adminclient command groups.
type Flags struct {
	Path string
	Dir  string
}

// AddPersistentFlags registers -f/--file and --conf-dir as persistent flags on cmd.
func AddPersistentFlags(cmd *cobra.Command, flags *Flags) {
	cmd.PersistentFlags().StringVarP(&flags.Path, "file", "f", "",
		"path of service config file (default: auto-discover in --conf-dir)")
	cmd.PersistentFlags().StringVar(&flags.Dir, "conf-dir", DefaultDir,
		"directory of service config files for auto-discovery")
}

// Resolve returns the explicit flag path when set, otherwise the first conf file
// containing the target top-level section.
func Resolve(path, dir, section string) (string, error) {
	if path != "" {
		return path, nil
	}

	return discover(dir, section)
}

// discover finds the first conf file containing the target section.
func discover(dir, section string) (string, error) {
	matches, err := filepath.Glob(filepath.Join(dir, "*"+fileSuffix))
	if err != nil {
		return "", fmt.Errorf("failed to glob config files in %s: %w", dir, err)
	}

	sectionPattern := regexp.MustCompile(`(?m)^` + section + `:`)
	for _, file := range matches {
		content, err := os.ReadFile(file) // nolint: gosec
		if err != nil {
			// an unreadable file is treated as not containing the section.
			continue
		}
		if sectionPattern.Match(content) {
			return file, nil
		}
	}

	return "", fmt.Errorf("no config file containing '%s:' found in %s/, use -f to specify manually", section, dir)
}
