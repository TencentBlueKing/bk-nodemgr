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
// either the explicit -f/--file path, or auto-discovery under --conf-dir by a
// top-level section anchor (a shared section like etcd, or the section that
// uniquely identifies the target service).
package configfile

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/spf13/cobra"
)

const (
	// DefaultDir is the default directory of service config files.
	DefaultDir = "/bk-nodemgr/etc"
)

// serviceMarkers maps each service name to the top-level config section that
// uniquely identifies its config file. The markers hold in both deployment
// families: the helm configmaps and the legacy bk-nodemgr-<service>.yml
// templates, so discovery does not depend on file names.
// nolint: gochecknoglobals
var serviceMarkers = map[string]string{
	"application": "front",
	"backend":     "callbackServer",
	"file":        "downloadServer",
}

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

// Resolve returns the explicit flag path when set, otherwise the first config
// file containing the target top-level section.
func Resolve(path, dir, section string) (string, error) {
	if path != "" {
		return path, nil
	}

	matches, err := discover(dir, section)
	if err != nil {
		return "", err
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("no config file containing '%s:' found in %s/, use -f to specify manually", section, dir)
	}

	return matches[0], nil
}

// ResolveForService returns the explicit flag path when set, otherwise the config
// file of the named service: the single yaml file under dir containing the
// service-identifying top-level section (see serviceMarkers). Multiple matches
// are an error: the resolved file feeds credentials, so guessing is unsafe.
func ResolveForService(path, dir, service string) (string, error) {
	if path != "" {
		return path, nil
	}

	marker, ok := serviceMarkers[service]
	if !ok {
		return "", fmt.Errorf("unknown service '%s', available services: %s",
			service, strings.Join(conv.MapKeyToSlice(serviceMarkers), ", "))
	}

	matches, err := discover(dir, marker)
	if err != nil {
		return "", err
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("no config file for service '%s' found in %s/ (no yaml file contains the '%s:' section), "+
			"use -f to specify manually", service, dir, marker)
	}
	if len(matches) > 1 {
		return "", fmt.Errorf("multiple config files for service '%s' found in %s/: %s, use -f to specify manually",
			service, dir, strings.Join(matches, ", "))
	}

	return matches[0], nil
}

// discover returns all config files under dir containing the target top-level
// section. Both yaml and yml extensions are scanned, covering the helm
// *_conf.yaml files and the legacy bk-nodemgr-<service>.yml files alike.
func discover(dir, section string) ([]string, error) {
	var files []string
	for _, extension := range []string{"*.yaml", "*.yml"} {
		matched, err := filepath.Glob(filepath.Join(dir, extension))
		if err != nil {
			return nil, fmt.Errorf("failed to glob config files in %s: %w", dir, err)
		}
		files = append(files, matched...)
	}

	sectionPattern := regexp.MustCompile(`(?m)^` + section + `:`)
	var matches []string
	for _, file := range files {
		content, err := os.ReadFile(file) // nolint: gosec
		if err != nil {
			// an unreadable file is treated as not containing the section.
			continue
		}
		if sectionPattern.Match(content) {
			matches = append(matches, file)
		}
	}

	return matches, nil
}
