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

// Package database provides CLI commands for database (etcd/redis/mongodb) health checks
// and client proxies. All commands load the database sections of a service config file,
// with kms secret override applied when kms is enabled.
package database

import (
	"crypto/tls"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	"github.com/spf13/cobra"
)

// configFlags carries the config flags shared by all database subcommands.
type configFlags struct {
	path    string
	confDir string
}

// NewDatabaseCMD creates the db command grouping database connectivity checks
// and client proxies.
func NewDatabaseCMD() *cobra.Command {
	flags := &configFlags{}

	cmd := &cobra.Command{
		Use:     "database",
		Aliases: []string{"db"},
		Short:   "database (etcd/redis/mongodb) connectivity checks and client proxies",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	cmd.PersistentFlags().StringVarP(&flags.path, "file", "f", "",
		"path of service config file (default: auto-discover in --conf-dir)")
	cmd.PersistentFlags().StringVar(&flags.confDir, "conf-dir", defaultConfDir,
		"directory of service config files for auto-discovery")

	cmd.AddCommand(newCheckCMD(flags))
	cmd.AddCommand(newEtcdCMD(flags))
	cmd.AddCommand(newRedisCMD(flags))
	cmd.AddCommand(newMongoCMD(flags))

	return cmd
}

// storage section names in service config files.
const (
	sectionEtcd    = "etcd"
	sectionRedis   = "redis"
	sectionMongoDB = "mongodb"
)

const (
	// defaultConfDir is the default directory of service config files.
	defaultConfDir = "/bk-nodemgr/etc"
	// confFileSuffix is the file name suffix of service config files.
	confFileSuffix = "_conf.yaml"
)

// loadConfig resolves the config path (flag value or auto-discovery) and loads the database config.
// It returns the loaded config and the resolved path for user-facing messages.
func loadConfig(configPath, confDir, section string) (*config.AdminClientDatabase, string, error) {
	path := configPath
	if path == "" {
		discovered, err := discoverConfig(confDir, section)
		if err != nil {
			return nil, "", err
		}
		path = discovered
	}

	conf := &config.AdminClientDatabase{}
	if err := conf.LoadFromFile(path); err != nil {
		return nil, "", fmt.Errorf("failed to load config file(%s): %w", path, err)
	}

	return conf, path, nil
}

// discoverConfig finds the first conf file containing the target section.
func discoverConfig(confDir, section string) (string, error) {
	matches, err := filepath.Glob(filepath.Join(confDir, "*"+confFileSuffix))
	if err != nil {
		return "", fmt.Errorf("failed to glob config files in %s: %w", confDir, err)
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

	return "", fmt.Errorf("no config file containing '%s:' found in %s/, use -f to specify manually", section, confDir)
}

// newClientTLSConf builds client side tls config from a storage tls section.
// TLS is enabled only when ca/cert/key are all set, mirroring the service side behavior
// (pkg/discover/etcddiscover initTLS, backend service mongo client construction).
func newClientTLSConf(conf config.TLSConfig) (*tls.Config, error) {
	if !conf.IsEnabled() {
		// nolint: nilnil // nil tls.Config explicitly means TLS is not configured.
		return nil, nil
	}

	// ToSSLConfig resolves inline content or file paths into cert material.
	sslConf, err := conf.ToSSLConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create tls config: %w", err)
	}

	clientConf, err := sslConf.NewClientTLSConf()
	if err != nil {
		return nil, fmt.Errorf("failed to create tls config: %w", err)
	}

	return clientConf, nil
}
