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
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"path"
	"runtime/debug"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/retrier"
	"github.com/pkg/sftp"
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

// AuthMethod defines the auth method type.
type AuthMethod string

const (
	// AuthMethodNone this defines the none auth method.
	AuthMethodNone = "none"

	// AuthMethodPassword this defines the password auth method.
	AuthMethodPassword = "password"

	// AuthMethodPrivateKey this defines the private key auth method.
	AuthMethodPrivateKey = "private_key"
)

// Validate validate the auth method.
func (auth AuthMethod) Validate() error {
	switch auth {
	case AuthMethodNone, AuthMethodPassword, AuthMethodPrivateKey:
		return nil
	default:
		return fmt.Errorf("invalid auth method, method(%s)", auth)
	}
}

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
	Network    Network
	IP         string
	Port       int
	User       string
	Password   string
	AuthMethod AuthMethod
	PrivateKey []byte
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

	if err := conf.AuthMethod.Validate(); err != nil {
		return err
	}

	switch conf.AuthMethod {
	case AuthMethodNone:
		if len(conf.PrivateKey) > 0 || len(conf.Password) > 0 {
			return errors.New("auth mode is none,but private_key or password is not empty")
		}
	case AuthMethodPrivateKey:
		if len(conf.PrivateKey) == 0 {
			return errors.New("private_key is empty")
		}
	case AuthMethodPassword:
		if len(conf.Password) == 0 {
			return errors.New("password is empty")
		}
	}

	return nil
}

// getAddr get the addr.
func (conf *Config) getAddr() string {
	return fmt.Sprintf("%s:%d", conf.IP, conf.Port)
}

// DefaultTimeout is the default timeout for building the connection.
const DefaultTimeout = 30 * time.Second

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
		HostKeyCallback: func(_ string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
		BannerCallback: func(message string) error {
			logger.G.Sys().Warn("ssh banner: %s", message)

			return nil
		},
		Timeout: timeout,
	}

	switch config.AuthMethod {
	case AuthMethodPassword:
		sshConf.Auth = []ssh.AuthMethod{
			ssh.Password(config.Password),
			ssh.KeyboardInteractive(func(_, _ string, questions []string, _ []bool) ([]string, error) {
				answers := make([]string, len(questions))
				for i := range questions {
					answers[i] = config.Password
				}

				return answers, nil
			}),
		}
	case AuthMethodPrivateKey:
		signer, err := ssh.ParsePrivateKey(config.PrivateKey)
		if err != nil {
			return nil, err
		}

		sshConf.Auth = []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		}

		// internal don't known host key.
		// nolint: gosec
		sshConf.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	default:
		sshConf.Auth = []ssh.AuthMethod{}
	}

	client := &Client{}

	// because of the network may be unstable, so we need to retry.
	backoff := retrier.NewExpoBackoff(retrier.ExpoBackoffOpts{
		MaxRetries:    5,
		BaseDelay:     1 * time.Second,
		MaxDelay:      3 * time.Second,
		JitterPercent: 0.2,
	})
	err := backoff.Do(ctx, func(_ int) error {
		var dialErr error
		client.sshClient, dialErr = ssh.Dial(string(config.Network), config.getAddr(), sshConf)
		if dialErr != nil {
			logger.G.Sys().WithErr(dialErr).With("host", config.getAddr()).Error("failed to connect to host")

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
}

// RunCommand run command, returning stdout and stderr separately.
func (cli *Client) RunCommand(cmd string) (_, _ string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("failed to run command, cmd(%s): %v", cmd, r)

			logger.G.Sys().With("command", cmd, "recover", r, "stack", debug.Stack()).Error("failed to run command, got panic")
		}
	}()

	session, err := cli.sshClient.NewSession()
	if err != nil {
		return "", "", err
	}
	defer func() { _ = session.Close() }()

	var stdoutBuf, stderrBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Stderr = &stderrBuf

	if err := session.Run(cmd); err != nil {
		return stdoutBuf.String(), stderrBuf.String(),
			fmt.Errorf("failed to run command, cmd(%s), stdout(%s), stderr(%s): %w",
				cmd, stdoutBuf.String(), stderrBuf.String(), err)
	}

	return stdoutBuf.String(), stderrBuf.String(), nil
}

// Close close the ssh client.
func (cli *Client) Close() error {
	return cli.sshClient.Close()
}

// TransferFile transfer file.
func (cli *Client) TransferFile(file io.ReadCloser, destPath string) error {
	defer func() { _ = file.Close() }()

	sftpClient, err := sftp.NewClient(cli.sshClient)
	if err != nil {
		return fmt.Errorf("failed to create sftp client: %w", err)
	}

	destDir := path.Dir(destPath)
	if err := sftpClient.MkdirAll(destDir); err != nil {
		return fmt.Errorf("failed to create dir: %w", err)
	}

	destFile, err := sftpClient.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() { _ = destFile.Close() }()

	if _, err := io.Copy(destFile, file); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}
