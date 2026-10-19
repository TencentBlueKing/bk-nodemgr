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
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	"github.com/spf13/cobra"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/redis/go-redis/v9"
	"github.com/redis/go-redis/v9/maintnotifications"
)

const (
	// checkDialTimeout is the dial timeout of database clients used by check.
	checkDialTimeout = 5 * time.Second
	// checkTimeout is the overall timeout of a single check attempt.
	checkTimeout = 5 * time.Second
)

// newCheckCMD creates the check command for database connectivity checks.
// Each check is a single attempt: success exits 0, failure exits non-zero.
// Retry is left to the caller (e.g. the retry loop in the k8s initContainer).
func newCheckCMD(flags *configFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check",
		Short: "check database connectivity (single attempt, non-zero exit on failure)",
		// the command gates initContainer success; an unknown component must fail loudly
		// instead of printing help and exiting 0.
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "etcd",
		Short: "check etcd connectivity",
		Args:  cobra.NoArgs,
		// Runtime failures must not dump usage into initContainer logs on every retry.
		// Argument and flag validation still report errors before running the check.
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			conf, path, err := loadConfig(flags.path, flags.confDir, sectionEtcd)
			if err != nil {
				return err
			}
			if err := conf.Etcd.Validate(); err != nil {
				return fmt.Errorf("failed to validate etcd config in %s: %w", path, err)
			}

			return checkEtcd(cmd.Context(), cmd.OutOrStdout(), &conf.Etcd)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "redis",
		Short: "check redis connectivity",
		Args:  cobra.NoArgs,
		// Runtime failures must not dump usage into initContainer logs on every retry.
		// Argument and flag validation still report errors before running the check.
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			conf, path, err := loadConfig(flags.path, flags.confDir, sectionRedis)
			if err != nil {
				return err
			}
			if err := conf.Redis.Validate(); err != nil {
				return fmt.Errorf("failed to validate redis config in %s: %w", path, err)
			}

			return checkRedis(cmd.Context(), cmd.OutOrStdout(), &conf.Redis)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "mongodb",
		Short: "check mongodb connectivity",
		Args:  cobra.NoArgs,
		// Runtime failures must not dump usage into initContainer logs on every retry.
		// Argument and flag validation still report errors before running the check.
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			conf, path, err := loadConfig(flags.path, flags.confDir, sectionMongoDB)
			if err != nil {
				return err
			}
			if err := conf.MongoDB.Validate(); err != nil {
				return fmt.Errorf("failed to validate mongodb config in %s: %w", path, err)
			}

			return checkMongoDB(cmd.Context(), cmd.OutOrStdout(), &conf.MongoDB)
		},
	})

	return cmd
}

// checkEtcd succeeds when at least one endpoint responds to a status request,
// so a partially available etcd cluster does not block service startup.
func checkEtcd(ctx context.Context, out io.Writer, conf *config.Etcd) error {
	tlsConf, err := newClientTLSConf(conf.TLS)
	if err != nil {
		return err
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   conf.Endpoints,
		Username:    conf.Username,
		Password:    conf.Password,
		DialTimeout: checkDialTimeout,
		TLS:         tlsConf,
	})
	if err != nil {
		return fmt.Errorf("failed to create etcd client: %w", err)
	}
	defer func() { _ = cli.Close() }()

	// NOTE: clientv3 forces WaitForReady(true) in its per-call options, so a refused
	// endpoint can only fail at its own deadline; errors are joined per endpoint to
	// keep the failure reason of each endpoint visible.
	var errs []error
	for _, endpoint := range conf.Endpoints {
		// each endpoint gets its own timeout so a black-holed endpoint
		// does not consume the budget of the remaining ones.
		epCtx, epCancel := context.WithTimeout(ctx, checkTimeout)
		_, err := cli.Status(epCtx, endpoint)
		epCancel()
		if err != nil {
			errs = append(errs, fmt.Errorf("endpoint %s: %w", endpoint, err))
			continue
		}
		_, _ = fmt.Fprintf(out, "check etcd: OK (endpoint %s)\n", endpoint)

		return nil
	}

	return fmt.Errorf("failed to check etcd: %w", errors.Join(errs...))
}

// checkRedis pings redis with the same client construction as the backend service
// (internal/backend/service newRedisClient).
func checkRedis(ctx context.Context, out io.Writer, conf *config.Redis) error {
	tlsConf, err := newClientTLSConf(conf.TLS)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, checkTimeout)
	defer cancel()

	// maxRetriesDisabled keeps the single-attempt semantics of check: go-redis retries
	// are left to the caller-side retry loop (e.g. k8s initContainer restart).
	const maxRetriesDisabled = -1

	cli := redis.NewUniversalClient(&redis.UniversalOptions{
		IsClusterMode: conf.Type == config.RedisTypeCluster,
		Addrs:         conf.Addrs,
		MasterName:    conf.MasterName,
		Password:      conf.Password,
		DB:            conf.DB,
		DialTimeout:   checkDialTimeout,
		MaxRetries:    maxRetriesDisabled,
		TLSConfig:     tlsConf,
		// maintenance notifications (Redis Cloud endpoint migration push) are irrelevant
		// to a one-shot probe; auto mode would log a misleading handshake error against
		// servers without the CLIENT MAINT_NOTIFICATIONS command.
		MaintNotificationsConfig: &maintnotifications.Config{Mode: maintnotifications.ModeDisabled},
	})
	defer func() { _ = cli.Close() }()

	if err := cli.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to ping redis(type %s, addrs %v): %w", conf.Type, conf.Addrs, err)
	}

	_, _ = fmt.Fprintf(out, "check redis: OK (type %s, addrs %v)\n", conf.Type, conf.Addrs)

	return nil
}

// checkMongoDB pings mongodb with the configured credential.
func checkMongoDB(ctx context.Context, out io.Writer, conf *config.MongoDB) error {
	tlsConf, err := newClientTLSConf(conf.TLS)
	if err != nil {
		return err
	}

	// server selection is the blocking step against unreachable hosts; the ping itself
	// is bounded by the check timeout context below.
	serverSelectionTimeout := checkTimeout
	clientOptions := &mongoOptions.ClientOptions{
		Hosts: conf.Hosts,
		Auth: &mongoOptions.Credential{
			AuthMechanism: conf.AuthMechanism,
			AuthSource:    conf.AuthSource,
			Username:      conf.Username,
			Password:      conf.Password,
			PasswordSet:   true,
		},
		TLSConfig:              tlsConf,
		ServerSelectionTimeout: &serverSelectionTimeout,
	}
	if conf.ReplicaSet != "" {
		clientOptions.ReplicaSet = &conf.ReplicaSet
	}

	cli, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to create mongodb client: %w", err)
	}
	// nolint: contextcheck // disconnect must outlive the check timeout context.
	defer func() { _ = cli.Disconnect(context.Background()) }()

	ctx, cancel := context.WithTimeout(ctx, checkTimeout)
	defer cancel()

	if err := cli.Ping(ctx, readpref.Primary()); err != nil {
		return fmt.Errorf("failed to ping mongodb(%v): %w", conf.Hosts, err)
	}

	_, _ = fmt.Fprintf(out, "check mongodb: OK (hosts %v)\n", conf.Hosts)

	return nil
}
