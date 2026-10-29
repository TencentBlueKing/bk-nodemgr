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

// Package jwttoken provides the jwt_token command: prints a jwt token for one service
// http server, signed with the service's jwt symmetric key. The key is decrypted from
// the kms global envelope when kms is enabled, or read from the plaintext server
// section of the service config file otherwise, so install/upgrade scripts can call
// authenticated internal APIs without knowing where the key comes from.
package jwttoken

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/internal/adminclient/configfile"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	restheader "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/header"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/spf13/cobra"
)

const (
	// sectionKMS is the top-level config section used for auto-discovery.
	sectionKMS = "kms"
	// serviceServerArgCount is the positional argument count of jwt_token: service + server.
	serviceServerArgCount = 2
)

// jwt server names.
const (
	serverInfo     = "info"
	serverAdmin    = "admin"
	serverBasic    = "basic"
	serverCallback = "callback"
	serverProxy    = "proxy"
	serverDownload = "download"
	serverExport   = "export"
)

// jwtServiceConfigs maps each service name to its valid http servers and the loader
// that reads the jwt key for one server from the service config file. Only symmetric
// keys are supported: the rest-server auth identity verifies tokens with
// JWTServerConfig.SymmetricKey.
// nolint: gochecknoglobals
var jwtServiceConfigs = map[string]struct {
	servers []string
	loadKey func(path, server string) (string, error)
}{
	"application": {
		servers: []string{serverInfo, serverAdmin, serverBasic},
		loadKey: loadApplicationJWTKey,
	},
	"backend": {
		servers: []string{serverInfo, serverAdmin, serverBasic, serverCallback, serverProxy},
		loadKey: loadBackendJWTKey,
	},
	"file": {
		servers: []string{serverInfo, serverAdmin, serverBasic, serverDownload, serverExport},
		loadKey: loadFileJWTKey,
	},
}

// NewJWTTokenCMD creates the jwt_token command.
func NewJWTTokenCMD() *cobra.Command {
	flags := &configfile.Flags{}
	var loginName, bkUserName, jwtKey string

	cmd := &cobra.Command{
		Use:   "jwt_token <service> <server>",
		Short: "print a jwt token for a service http server, signed with the service's jwt symmetric key",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(serviceServerArgCount)(cmd, args); err != nil {
				return err
			}

			return validateServiceServer(args[0], args[1])
		},
		// runtime failures must not dump usage into the job logs that capture this
		// command's output; argument errors are already rejected by Args validation.
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := loadJWTToken(flags, args[0], args[1], loginName, bkUserName, jwtKey)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), token)

			return nil
		},
	}
	configfile.AddPersistentFlags(cmd, flags)
	cmd.Flags().StringVar(&loginName, "login-name", "admin", "login name written into the jwt claims")
	cmd.Flags().StringVar(&bkUserName, "bk-username", "admin", "bk username written into the jwt claims")
	cmd.Flags().StringVar(&jwtKey, "jwt-key", "",
		"sign with this key directly and skip config file parsing; for manual troubleshooting only, install scripts must not use it")

	return cmd
}

// loadJWTToken signs a token for the service server pair. With an explicit jwt key it
// signs directly; otherwise it resolves the config file and loads the key from the kms
// global envelope (kms enabled) or the plaintext server sections (kms disabled).
func loadJWTToken(flags *configfile.Flags, service, server, loginName, bkUserName, jwtKey string) (string, error) {
	key := jwtKey
	if key == "" {
		loadedKey, err := loadJWTKeyFromConfig(flags, service, server)
		if err != nil {
			return "", err
		}
		key = loadedKey
	}

	token, err := restheader.NewNodeMgrAuthorizationManager(key).Generate(loginName, bkUserName)
	if err != nil {
		return "", err
	}

	return token, nil
}

// loadJWTKeyFromConfig resolves the config file and loads the jwt key for the service
// server pair from it.
func loadJWTKeyFromConfig(flags *configfile.Flags, service, server string) (string, error) {
	if err := validateServiceServer(service, server); err != nil {
		return "", err
	}

	path, err := configfile.Resolve(flags.Path, flags.Dir, sectionKMS)
	if err != nil {
		return "", err
	}

	key, err := jwtServiceConfigs[service].loadKey(path, server)
	if err != nil {
		return "", err
	}

	if key == "" {
		return "", fmt.Errorf("jwt symmetric key for %s %s is empty in config file(%s) or its kms global envelope",
			service, server, path)
	}

	return key, nil
}

// loadApplicationJWTKey loads the jwt key for one application server from the config
// file. The server name is validated before the call; an empty key means the config
// has none.
func loadApplicationJWTKey(path, server string) (string, error) {
	conf := &config.AdminClientApplication{}
	if err := conf.LoadFromFile(path); err != nil {
		return "", fmt.Errorf("failed to load config file(%s): %w", path, err)
	}

	switch server {
	case serverInfo:
		return conf.InfoServer.JWTServerConfig.SymmetricKey, nil
	case serverAdmin:
		return conf.AdminServer.JWTServerConfig.SymmetricKey, nil
	case serverBasic:
		return conf.BasicServer.JWTServerConfig.SymmetricKey, nil
	}

	return "", nil
}

// loadBackendJWTKey loads the jwt key for one backend server from the config file.
func loadBackendJWTKey(path, server string) (string, error) {
	conf := &config.AdminClientBackend{}
	if err := conf.LoadFromFile(path); err != nil {
		return "", fmt.Errorf("failed to load config file(%s): %w", path, err)
	}

	switch server {
	case serverInfo:
		return conf.InfoServer.JWTServerConfig.SymmetricKey, nil
	case serverAdmin:
		return conf.AdminServer.JWTServerConfig.SymmetricKey, nil
	case serverBasic:
		return conf.BasicServer.JWTServerConfig.SymmetricKey, nil
	case serverCallback:
		return conf.CallbackServer.JWTServerConfig.SymmetricKey, nil
	case serverProxy:
		return conf.ProxyServer.JWTServerConfig.SymmetricKey, nil
	}

	return "", nil
}

// loadFileJWTKey loads the jwt key for one file server from the config file.
func loadFileJWTKey(path, server string) (string, error) {
	conf := &config.AdminClientFile{}
	if err := conf.LoadFromFile(path); err != nil {
		return "", fmt.Errorf("failed to load config file(%s): %w", path, err)
	}

	switch server {
	case serverInfo:
		return conf.InfoServer.JWTServerConfig.SymmetricKey, nil
	case serverAdmin:
		return conf.AdminServer.JWTServerConfig.SymmetricKey, nil
	case serverBasic:
		return conf.BasicServer.JWTServerConfig.SymmetricKey, nil
	case serverDownload:
		return conf.DownloadServer.JWTServerConfig.SymmetricKey, nil
	case serverExport:
		return conf.ExportServer.JWTServerConfig.SymmetricKey, nil
	}

	return "", nil
}

// validateServiceServer validates the service server pair against the known combos.
func validateServiceServer(service, server string) error {
	entry, ok := jwtServiceConfigs[service]
	if !ok {
		return fmt.Errorf("unknown service '%s', available services: %s", service, joinSortedKeys(jwtServiceConfigs))
	}

	if !slices.Contains(entry.servers, server) {
		servers := slices.Clone(entry.servers)
		sort.Strings(servers)

		return fmt.Errorf("unknown server '%s' for service '%s', available servers: %s",
			server, service, strings.Join(servers, ", "))
	}

	return nil
}

// joinSortedKeys joins map keys in sorted order for stable error messages.
func joinSortedKeys[V any](source map[string]V) string {
	keys := conv.MapKeyToSlice(source)
	sort.Strings(keys)

	return strings.Join(keys, ", ")
}
