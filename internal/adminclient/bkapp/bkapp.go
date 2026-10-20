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

// Package bkapp provides the bkapp command group: prints the bk app code/secret
// decrypted from the kms global envelope, for install/upgrade jobs that need the
// app credentials without reading plaintext from values or config files.
package bkapp

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/internal/adminclient/configfile"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	"github.com/spf13/cobra"
)

// sectionKMS is the top-level config section used for auto-discovery.
const sectionKMS = "kms"

// NewBkappCMD creates the bkapp command group.
func NewBkappCMD() *cobra.Command {
	flags := &configfile.Flags{}

	cmd := &cobra.Command{
		Use:   "bkapp",
		Short: "print the bk app code/secret from the kms global envelope",
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

// newValueCMD creates a leaf command printing one field of the bk app credential pair.
// The value is the only stdout output, so callers can capture it directly:
// APP_CODE=$(bk-nodemgr-adminclient bkapp code -f /bk-nodemgr/etc/backend_conf.yaml).
func newValueCMD(flags *configfile.Flags, use, short string, pick func(appCode, appSecret string) string) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		// runtime failures must not dump usage into the job logs that capture this
		// command's output; argument errors are already rejected by cobra.NoArgs.
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			appCode, appSecret, err := loadBKAppIDSecret(flags)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), pick(appCode, appSecret))

			return nil
		},
	}
}

// loadBKAppIDSecret resolves the config file and loads the bk app credential pair
// decrypted from the kms global envelope.
func loadBKAppIDSecret(flags *configfile.Flags) (string, string, error) {
	path, err := configfile.Resolve(flags.Path, flags.Dir, sectionKMS)
	if err != nil {
		return "", "", err
	}

	conf := &config.AdminClientBKApp{}
	if err := conf.LoadFromFile(path); err != nil {
		return "", "", fmt.Errorf("failed to load config file(%s): %w", path, err)
	}

	return conf.AppCode, conf.AppSecret, nil
}
