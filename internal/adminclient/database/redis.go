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
	"io"
	"net"
	"os/exec"
	"strconv"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/internal/adminclient/configfile"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	"github.com/spf13/cobra"
)

const (
	redisCommandLogin  = "login"
	redisCommandStatus = "status"

	defaultRedisPort = "6379"
	// redisCLIAuthEnvName passes the password via environment, avoiding exposure in ps output.
	redisCLIAuthEnvName = "REDISCLI_AUTH"
)

// newRedisCMD creates the redis command proxying to redis-cli with connection
// parameters from the service config file.
func newRedisCMD(flags *configfile.Flags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "redis [-f config] [login|status]",
		Short: "connect to redis via redis-cli (supported types: standalone, cluster, sentinel)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			command := redisCommandLogin
			if len(args) > 0 {
				command = args[0]
			}
			if command != redisCommandLogin && command != redisCommandStatus {
				return fmt.Errorf("unknown command '%s', available commands: login, status", command)
			}

			conf, path, err := loadConfig(flags, sectionRedis)
			if err != nil {
				return err
			}
			if err := conf.Redis.Validate(); err != nil {
				return fmt.Errorf("failed to validate redis config in %s: %w", path, err)
			}

			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Using config: %s\n", path)

			return runRedisCLI(cmd, &conf.Redis, command)
		},
	}

	return cmd
}

// runRedisCLI executes the login or status command against the configured redis.
func runRedisCLI(cmd *cobra.Command, conf *config.Redis, command string) error {
	tlsArgs := redisTLSArgs(conf)

	// the password is passed via environment so it does not appear in ps output;
	// Redis.Validate already guarantees it is non-empty.
	extraEnv := []string{redisCLIAuthEnvName + "=" + conf.Password}

	addr, err := resolveRedisAddr(conf, tlsArgs, extraEnv, cmd.ErrOrStderr())
	if err != nil {
		return err
	}

	host, port := splitHostPort(addr, defaultRedisPort)

	switch command {
	case redisCommandStatus:
		return redisStatus(cmd, conf, host, port, tlsArgs, extraEnv)
	default:
		return redisLogin(cmd, conf, host, port, tlsArgs, extraEnv)
	}
}

// redisLogin starts an interactive redis-cli session.
func redisLogin(cmd *cobra.Command, conf *config.Redis, host, port string, tlsArgs, extraEnv []string) error {
	args := redisBaseArgs(conf, host, port, tlsArgs)

	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Connecting to Redis %s: %s:%s\n", conf.Type, host, port)

	return execCommand("redis-cli", args, extraEnv)
}

// redisStatus checks redis connectivity and prints brief server info.
func redisStatus(cmd *cobra.Command, conf *config.Redis, host, port string, tlsArgs, extraEnv []string) error {
	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Checking Redis %s: %s:%s\n", conf.Type, host, port)

	// -e makes server-side errors fail status checks without changing interactive login.
	args := append(redisBaseArgs(conf, host, port, tlsArgs), "-e")

	if conf.Type == config.RedisTypeCluster {
		return execRedisCLI(cmd, args, extraEnv, "CLUSTER", "INFO")
	}

	if err := execRedisCLI(cmd, args, extraEnv, "PING"); err != nil {
		return err
	}

	output, err := outputRedisCLI(args, extraEnv, "INFO", "server")
	if err != nil {
		return err
	}

	for _, line := range strings.Split(output, "\n") {
		if isRedisInfoSummaryLine(line) {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), strings.TrimRight(line, "\r"))
		}
	}

	return nil
}

// redisBaseArgs builds the shared redis-cli connection args:
// -c for cluster mode, -h/-p for the target address, -n for the db outside cluster mode.
func redisBaseArgs(conf *config.Redis, host, port string, tlsArgs []string) []string {
	var args []string
	if conf.Type == config.RedisTypeCluster {
		args = append(args, "-c")
	}
	args = append(args, "-h", host, "-p", port)
	if conf.Type != config.RedisTypeCluster {
		args = append(args, "-n", strconv.Itoa(conf.DB))
	}

	return append(args, tlsArgs...)
}

// isRedisInfoSummaryLine reports whether the line belongs to the brief server summary.
func isRedisInfoSummaryLine(line string) bool {
	switch {
	case strings.HasPrefix(line, "redis_version:"):
		return true
	case strings.HasPrefix(line, "uptime_in_seconds:"):
		return true
	case strings.HasPrefix(line, "connected_clients:"):
		return true
	default:
		return false
	}
}

// resolveRedisAddr resolves the address to connect to. In sentinel mode the master
// address is resolved via the first sentinel, other modes use the first configured addr.
func resolveRedisAddr(conf *config.Redis, tlsArgs, extraEnv []string, errOut io.Writer) (string, error) {
	if conf.Type != config.RedisTypeSentinel {
		return conf.Addrs[0], nil
	}

	sentinel := conf.Addrs[0]
	_, _ = fmt.Fprintf(errOut, "Resolving master '%s' via sentinel: %s\n", conf.MasterName, sentinel)

	host, port := splitHostPort(sentinel, defaultRedisPort)
	args := append([]string{"-h", host, "-p", port}, tlsArgs...)

	output, err := outputRedisCLI(args, extraEnv, "SENTINEL", "get-master-addr-by-name", conf.MasterName)
	if err != nil {
		return "", fmt.Errorf("could not resolve master '%s' from sentinel %s: %w", conf.MasterName, sentinel, err)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	// sentinel get-master-addr-by-name outputs two lines: host and port.
	const sentinelMasterOutputLines = 2
	if len(lines) < sentinelMasterOutputLines {
		return "", fmt.Errorf("could not resolve master '%s' from sentinel %s: unexpected output %q",
			conf.MasterName, sentinel, output)
	}

	return net.JoinHostPort(strings.TrimRight(lines[0], "\r"), strings.TrimRight(lines[1], "\r")), nil
}

// redisTLSArgs builds redis-cli tls arguments for existing cert files.
func redisTLSArgs(conf *config.Redis) []string {
	var args []string
	if fileExists(conf.TLS.CAFile) {
		args = append(args, "--tls", "--cacert", conf.TLS.CAFile)
	}
	if fileExists(conf.TLS.CertFile) {
		args = append(args, "--cert", conf.TLS.CertFile)
	}
	if fileExists(conf.TLS.KeyFile) {
		args = append(args, "--key", conf.TLS.KeyFile)
	}

	return args
}

// execRedisCLI runs redis-cli with the given base args plus command args, streaming output.
func execRedisCLI(cmd *cobra.Command, args []string, extraEnv []string, cmdArgs ...string) error {
	cli := exec.Command("redis-cli", append(args, cmdArgs...)...) // nolint: gosec
	cli.Stdout = cmd.OutOrStdout()
	cli.Stderr = cmd.ErrOrStderr()
	cli.Env = mergeEnv(extraEnv)

	return cli.Run()
}

// outputRedisCLI runs redis-cli and returns its stdout.
func outputRedisCLI(args []string, extraEnv []string, cmdArgs ...string) (string, error) {
	cli := exec.Command("redis-cli", append(args, cmdArgs...)...) // nolint: gosec
	cli.Env = mergeEnv(extraEnv)

	output, err := cli.Output()
	if err != nil {
		return "", err
	}

	return string(output), nil
}
