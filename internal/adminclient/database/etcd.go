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

package database

import (
	"fmt"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/internal/adminclient/configfile"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	"github.com/spf13/cobra"
)

// newEtcdCMD creates the etcd command proxying to etcdctl with connection
// parameters from the service config file.
func newEtcdCMD(flags *configfile.Flags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "etcd [-f config] -- <etcdctl-subcommand> [args...]",
		Short: "proxy etcdctl with connection parameters from config",
		Example: `  bk-nodemgr-adminclient db etcd -- member list
  bk-nodemgr-adminclient db etcd -- endpoint health
  bk-nodemgr-adminclient db etcd -f /bk-nodemgr/etc/backend_conf.yaml -- get /bk-nodemgr/some/key`,
		// etcdctl flags must follow "--" so cobra does not try to parse them.
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}

			conf, path, err := loadConfig(flags, sectionEtcd)
			if err != nil {
				return err
			}
			if err := conf.Etcd.Validate(); err != nil {
				return fmt.Errorf("failed to validate etcd config in %s: %w", path, err)
			}

			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Using config: %s\n", path)

			etcdctlArgs := buildEtcdctlArgs(&conf.Etcd, args)

			// pass the credential via environment so it does not appear in the process
			// argument list (same approach as REDISCLI_AUTH).
			var extraEnv []string
			if conf.Etcd.Username != "" && conf.Etcd.Password != "" {
				extraEnv = append(extraEnv, "ETCDCTL_USER="+conf.Etcd.Username+":"+conf.Etcd.Password)
			}

			return execCommand("etcdctl", etcdctlArgs, extraEnv)
		},
	}

	return cmd
}

// buildEtcdctlArgs prepends connection parameters before the user supplied etcdctl args.
// The credential is not included here; it is passed via the ETCDCTL_USER environment variable.
func buildEtcdctlArgs(conf *config.Etcd, args []string) []string {
	var result []string

	if len(conf.Endpoints) > 0 {
		result = append(result, "--endpoints="+strings.Join(conf.Endpoints, ","))
	}
	if fileExists(conf.TLS.CAFile) {
		result = append(result, "--cacert="+conf.TLS.CAFile)
	}
	if fileExists(conf.TLS.CertFile) {
		result = append(result, "--cert="+conf.TLS.CertFile)
	}
	if fileExists(conf.TLS.KeyFile) {
		result = append(result, "--key="+conf.TLS.KeyFile)
	}

	return append(result, args...)
}
