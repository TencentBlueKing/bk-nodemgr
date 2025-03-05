/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package sshx ...
package sshx

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/retrier"
	"golang.org/x/crypto/ssh"
)

// Network defines the network type.
type Network string

const (
	// NetworkTCP is tcp network.
	NetworkTCP Network = "tcp"

	// NetworkTcp6 is tcp6 network.
	NetworkTcp6 Network = "tcp6"

	// NetworkUdp is udp network.
	NetworkUdp Network = "udp"

	// NetworkUdp6 is udp6 network.
	NetworkUdp6 Network = "udp6"
)

// Validate validate the network.
func (net Network) Validate() error {
	switch net {
	case NetworkTCP, NetworkTcp6, NetworkUdp, NetworkUdp6:
	default:
		return fmt.Errorf("invalid network, network(%s)", net)
	}

	return nil
}

// Config this is the config for sshx handler.
type Config struct {
	Network  Network
	IP       string
	Port     int
	User     string
	Password string
	Logger   logger.Logger
}

// Validate validate the config.
func (conf *Config) Validate() error {
	if err := conf.Network.Validate(); err != nil {
		return err
	}

	if conf.IP == "" {
		return fmt.Errorf("ip is empty")
	}

	if conf.Port == 0 {
		return fmt.Errorf("port is empty")
	}

	if conf.User == "" {
		return fmt.Errorf("user is empty")
	}

	if conf.Logger == nil {
		return fmt.Errorf("logger is empty")
	}

	return nil
}

// getAddr get the addr.
func (conf *Config) getAddr() string {
	return fmt.Sprintf("%s:%d", conf.IP, conf.Port)
}

// NewClient new a ssh client.
// timeout is the timeout for building the connection。
func NewClient(ctx context.Context, config *Config, timeout time.Duration) (*Client, error) {
	if ctx == nil {
		return nil, fmt.Errorf("ctx is nil")
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	sshConf := &ssh.ClientConfig{
		User: config.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(config.Password),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
		BannerCallback: func(message string) error {
			config.Logger.Warnf("ssh banner: %s", message)

			return nil
		},
		Timeout: timeout,
	}

	client := &Client{
		logger: config.Logger,
	}

	// because of the network may be unstable, so we need to retry.
	backoff := retrier.NewExpoBackoff(retrier.ExpoBackoffOpts{
		MaxRetries:    5,
		BaseDelay:     1 * time.Second,
		MaxDelay:      3 * time.Second,
		JitterPercent: 0.2,
		Logger:        config.Logger,
	})
	err := backoff.Do(ctx, func(attempt int) error {
		var dialErr error
		client.sshClient, dialErr = ssh.Dial(string(config.Network), config.getAddr(), sshConf)
		if dialErr != nil {
			client.logger.Errorf("failed to connect to host, host(%s), err: %v", config.getAddr(), dialErr)
			return dialErr
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return client, nil
}

// Client this is a ssh client.
type Client struct {
	sshClient *ssh.Client
	logger    logger.Logger
}

// RunCommand run command.
func (cli *Client) RunCommand(cmd string) (outStr string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("failed to run command, cmd(%s), err: %v", cmd, r)
			cli.logger.Errorf("failed to run command, cmd(%s), err: %v", cmd, r)
		}
	}()

	session, err := cli.sshClient.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return "", err
	}

	return string(output), nil
}

// Close close the ssh client.
func (cli *Client) Close() error {
	return cli.sshClient.Close()
}
