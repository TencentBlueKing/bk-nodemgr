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
	"net"
	"net/url"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	"github.com/spf13/cobra"
)

const (
	mongoCommandLogin  = "login"
	mongoCommandStatus = "status"

	defaultMongoPort = "27017"
)

// mongoStatusScript prints a brief server status summary.
const mongoStatusScript = `
var status = db.serverStatus();
print('Server OK: ' + status.ok);
print('Version: ' + status.version);
print('Uptime: ' + status.uptime + 's');
try {
  var rsStatus = rs.status();
  print('ReplicaSet: ' + rsStatus.set + ' (members: ' + rsStatus.members.length + ')');
} catch(e) { print('ReplicaSet: N/A (standalone)'); }
`

// newMongoCMD creates the mongodb command proxying to the mongo shell with connection
// parameters from the service config file. "mongo" is kept as an alias.
func newMongoCMD(flags *configFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "mongodb [-f config] [login|status]",
		Aliases: []string{"mongo"},
		Short:   "connect to mongodb via mongo shell",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			command := mongoCommandLogin
			if len(args) > 0 {
				command = args[0]
			}
			if command != mongoCommandLogin && command != mongoCommandStatus {
				return fmt.Errorf("unknown command '%s', available commands: login, status", command)
			}

			conf, path, err := loadConfig(flags.path, flags.confDir, sectionMongoDB)
			if err != nil {
				return err
			}
			if err := conf.MongoDB.Validate(); err != nil {
				return fmt.Errorf("failed to validate mongodb config in %s: %w", path, err)
			}

			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Using config: %s\n", path)

			return runMongoShell(cmd, &conf.MongoDB, command)
		},
	}

	return cmd
}

// runMongoShell executes the login or status command against the configured mongodb.
func runMongoShell(cmd *cobra.Command, conf *config.MongoDB, command string) error {
	connURI := buildMongoConnURI(conf)
	args := append([]string{connURI}, mongoTLSArgs(conf)...)

	// NOTE: the connection uri contains the credential, do not echo the command line.

	switch command {
	case mongoCommandStatus:
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Checking MongoDB status: %s\n", strings.Join(conf.Hosts, ","))

		return execCommand("mongo", append(args, "--quiet", "--eval", mongoStatusScript), nil)
	default:
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Connecting to MongoDB: %s (db: %s)\n", strings.Join(conf.Hosts, ","), conf.Database)

		return execCommand("mongo", args, nil)
	}
}

// buildMongoConnURI builds the mongodb connection uri. Hosts get the default port when
// absent; credentials are percent-encoded via net/url (url.QueryEscape would encode a
// space as "+", which the mongo uri parser does not decode back).
// MongoDB.Validate already guarantees username/password/database are non-empty.
func buildMongoConnURI(conf *config.MongoDB) string {
	hosts := make([]string, 0, len(conf.Hosts))
	for _, host := range conf.Hosts {
		h, p := splitHostPort(host, defaultMongoPort)
		hosts = append(hosts, net.JoinHostPort(strings.Trim(h, "[]"), p))
	}

	connURI := "mongodb://" + url.UserPassword(conf.Username, conf.Password).String() + "@"
	connURI += strings.Join(hosts, ",") + "/" + conf.Database

	queryParams := make(url.Values)
	if conf.AuthSource != "" {
		queryParams.Set("authSource", conf.AuthSource)
	}
	if conf.AuthMechanism != "" {
		queryParams.Set("authMechanism", conf.AuthMechanism)
	}
	if len(queryParams) > 0 {
		connURI += "?" + queryParams.Encode()
	}

	return connURI
}

// mongoTLSArgs builds mongo shell tls arguments for existing cert files.
func mongoTLSArgs(conf *config.MongoDB) []string {
	var args []string
	if fileExists(conf.TLS.CAFile) {
		args = append(args, "--tls", "--tlsCAFile", conf.TLS.CAFile)
	}
	if fileExists(conf.TLS.CertFile) {
		args = append(args, "--tlsCertificateKeyFile", conf.TLS.CertFile)
	}

	return args
}
