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
// http server, signed with the symmetric key decrypted from the kms global envelope, so
// install/upgrade scripts can call authenticated internal APIs without reading plaintext
// jwt keys from values or config files.
package jwttoken

import (
	"fmt"
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

// jwtKeyPickers maps each service and http server name to its decrypted jwt symmetric
// key in the adminclient config. Only symmetric keys are supported: the rest-server
// auth identity verifies tokens with JWTServerConfig.SymmetricKey.
// nolint: gochecknoglobals
var jwtKeyPickers = map[string]map[string]func(config.AdminClientJWT) string{
	"application": {
		"info":  func(conf config.AdminClientJWT) string { return conf.ApplicationInfo.SymmetricKey },
		"admin": func(conf config.AdminClientJWT) string { return conf.ApplicationAdmin.SymmetricKey },
		"basic": func(conf config.AdminClientJWT) string { return conf.ApplicationBasic.SymmetricKey },
	},
	"backend": {
		"info":     func(conf config.AdminClientJWT) string { return conf.BackendInfo.SymmetricKey },
		"admin":    func(conf config.AdminClientJWT) string { return conf.BackendAdmin.SymmetricKey },
		"basic":    func(conf config.AdminClientJWT) string { return conf.BackendBasic.SymmetricKey },
		"callback": func(conf config.AdminClientJWT) string { return conf.BackendCallback.SymmetricKey },
		"proxy":    func(conf config.AdminClientJWT) string { return conf.BackendProxy.SymmetricKey },
	},
	"file": {
		"info":     func(conf config.AdminClientJWT) string { return conf.FileInfo.SymmetricKey },
		"admin":    func(conf config.AdminClientJWT) string { return conf.FileAdmin.SymmetricKey },
		"basic":    func(conf config.AdminClientJWT) string { return conf.FileBasic.SymmetricKey },
		"download": func(conf config.AdminClientJWT) string { return conf.FileDownload.SymmetricKey },
		"export":   func(conf config.AdminClientJWT) string { return conf.FileExport.SymmetricKey },
	},
}

// NewJWTTokenCMD creates the jwt_token command.
func NewJWTTokenCMD() *cobra.Command {
	flags := &configfile.Flags{}
	var loginName, bkUserName string

	cmd := &cobra.Command{
		Use:   "jwt_token <service> <server>",
		Short: "print a jwt token for a service http server, signed with the kms global envelope key",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(serviceServerArgCount)(cmd, args); err != nil {
				return err
			}
			_, err := pickJWTKey(args[0], args[1])

			return err
		},
		// runtime failures must not dump usage into the job logs that capture this
		// command's output; argument errors are already rejected by Args validation.
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := loadJWTToken(flags, args[0], args[1], loginName, bkUserName)
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

	return cmd
}

// loadJWTToken resolves the config file, decrypts the jwt key for the service server
// pair from the kms global envelope, and signs a token with it.
func loadJWTToken(flags *configfile.Flags, service, server, loginName, bkUserName string) (string, error) {
	pick, err := pickJWTKey(service, server)
	if err != nil {
		return "", err
	}

	path, err := configfile.Resolve(flags.Path, flags.Dir, sectionKMS)
	if err != nil {
		return "", err
	}

	conf := &config.AdminClientJWT{}
	if err := conf.LoadFromFile(path); err != nil {
		return "", fmt.Errorf("failed to load config file(%s): %w", path, err)
	}

	key := pick(*conf)
	if key == "" {
		return "", fmt.Errorf("jwt symmetric key for %s %s is empty in the kms global envelope", service, server)
	}

	token, err := restheader.NewNodeMgrAuthorizationManager(key).Generate(loginName, bkUserName)
	if err != nil {
		return "", err
	}

	return token, nil
}

// pickJWTKey returns the jwt key picker for the service server pair.
func pickJWTKey(service, server string) (func(config.AdminClientJWT) string, error) {
	servers, ok := jwtKeyPickers[service]
	if !ok {
		return nil, fmt.Errorf("unknown service '%s', available services: %s", service, joinSortedKeys(jwtKeyPickers))
	}

	pick, ok := servers[server]
	if !ok {
		return nil, fmt.Errorf("unknown server '%s' for service '%s', available servers: %s",
			server, service, joinSortedKeys(servers))
	}

	return pick, nil
}

// joinSortedKeys joins map keys in sorted order for stable error messages.
func joinSortedKeys[V any](source map[string]V) string {
	keys := conv.MapKeyToSlice(source)
	sort.Strings(keys)

	return strings.Join(keys, ", ")
}
