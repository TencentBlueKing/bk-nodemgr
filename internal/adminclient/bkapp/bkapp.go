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

// Package bkapp provides the bkapp command group: prints a named bk app
// code/secret for install/upgrade jobs. The credential source (kms global
// envelope, plaintext config sections, or job-injected env) is resolved inside
// pkg/config, so callers never perceive whether kms is enabled.
package bkapp

import (
	"fmt"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/internal/adminclient/configfile"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/spf13/cobra"
)

// serviceBackend is the service whose config file holds the bk app credentials.
const serviceBackend = "backend"

// bkapp credential names.
const (
	credAPIGWSync = "apigw-sync"
	credIAMV3     = "iamv3"
	credIAMV4     = "iamv4"
)

// bkAppCredential binds a credential name to its picker in the resolved
// adminclient config and its plaintext source (used in the empty-credential
// error; the kms source is always the global envelope). The name carries the
// consumer's credential selection semantics; the source difference (kms
// envelope vs plaintext) is already resolved by
// config.AdminClientBKApp.LoadFromFile.
type bkAppCredential struct {
	pick            func(conf *config.AdminClientBKApp) (string, string)
	plaintextSource string
}

// nolint: gochecknoglobals
var bkAppCredentials = map[string]bkAppCredential{
	credAPIGWSync: {
		pick: func(conf *config.AdminClientBKApp) (string, string) {
			return conf.Default.AppCode, conf.Default.AppSecret
		},
		plaintextSource: "the BK_APP_CODE/BK_APP_SECRET env",
	},
	credIAMV3: {
		pick: func(conf *config.AdminClientBKApp) (string, string) {
			return conf.IAMV3.AppCode, conf.IAMV3.AppSecret
		},
		plaintextSource: "the iamV3 config section",
	},
	credIAMV4: {
		pick: func(conf *config.AdminClientBKApp) (string, string) {
			return conf.IAMV4.AppCode, conf.IAMV4.AppSecret
		},
		plaintextSource: "the iamV4 config section",
	},
}

// NewBkappCMD creates the bkapp command group.
func NewBkappCMD() *cobra.Command {
	flags := &configfile.Flags{}

	cmd := &cobra.Command{
		Use:   "bkapp",
		Short: "print a named bk app code/secret resolved from the service config",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	configfile.AddPersistentFlags(cmd, flags)

	cmd.AddCommand(newValueCMD(flags, "code", "print the bk app code",
		func(appCode, _ string) string { return appCode }))
	cmd.AddCommand(newValueCMD(flags, "secret", "print the bk app secret",
		func(_, appSecret string) string { return appSecret }))

	return cmd
}

// newValueCMD creates a leaf command printing one field of the named bk app
// credential pair. The value is the only stdout output, so callers can capture it
// directly:
// APP_CODE=$(bk-nodemgr-adminclient bkapp code iamv3 -f /bk-nodemgr/etc/backend_conf.yaml).
func newValueCMD(flags *configfile.Flags, use, short string, pick func(appCode, appSecret string) string) *cobra.Command {
	return &cobra.Command{
		Use:   use + " <name>",
		Short: short,
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return err
			}

			return validateCredName(args[0])
		},
		// runtime failures must not dump usage into the job logs that capture this
		// command's output; argument errors are already rejected by Args validation.
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			appCode, appSecret, err := loadBKAppIDSecret(flags, args[0])
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), pick(appCode, appSecret))

			return nil
		},
	}
}

// loadBKAppIDSecret resolves the backend config file and loads the named bk app
// credential pair from it.
func loadBKAppIDSecret(flags *configfile.Flags, name string) (string, string, error) {
	path, err := configfile.ResolveForService(flags.Path, flags.Dir, serviceBackend)
	if err != nil {
		return "", "", err
	}

	conf := &config.AdminClientBKApp{}
	if err := conf.LoadFromFile(path); err != nil {
		return "", "", fmt.Errorf("failed to load config file(%s): %w", path, err)
	}

	cred := bkAppCredentials[name]
	appCode, appSecret := cred.pick(conf)
	if appCode == "" || appSecret == "" {
		source := cred.plaintextSource
		if conf.KMS.Enabled {
			source = "the kms global envelope"
		}

		return "", "", fmt.Errorf("bk app credential '%s' is empty in %s (config file %s)", name, source, path)
	}

	return appCode, appSecret, nil
}

// validateCredName validates the credential name against the known names.
func validateCredName(name string) error {
	if _, ok := bkAppCredentials[name]; !ok {
		return fmt.Errorf("unknown credential name '%s', available names: %s",
			name, strings.Join(conv.MapKeyToSlice(bkAppCredentials), ", "))
	}

	return nil
}
