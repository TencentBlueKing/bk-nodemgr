/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package wmix provides a client for executing commands and uploading files on Windows machines using WMI.
package wmix

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
)

// DefaultTimeout defines the default timeout for WMI operations.
const DefaultTimeout = 1 * time.Minute

// IClient defines the interface for ssh client.
type IClient interface {
	// UploadFile upload file to remote.
	UploadFile(ctx context.Context, scrFilePath, dstPath string) (string, string, error)

	// RunCommand run command on remote.
	RunCommand(ctx context.Context, cmd string) (string, string, error)
}

// Config this is the config for wmi handler.
type Config struct {
	Domain     string
	IP         string
	User       string
	Password   string
	AuthMethod AuthMethod
	Timeout    time.Duration
	Logger     logger.ILogger
}

const (
	// WMIEnvPassword defines the env key for password.
	WMIEnvPassword = "WMI_PASSWORD"
)

// getTarget returns the target string.
func (conf *Config) getTarget() string {
	str := conf.User

	if conf.Domain != "" {
		str = conf.Domain + "/" + str
	}

	switch conf.AuthMethod {
	case AuthMethodPassword:
		str = str + "@" + conf.IP
	case AuthMethodNone:
		str = "-no-pass " + str + "@" + conf.IP
	default:
		str = "-no-pass " + str + "@" + conf.IP // default to no password
	}

	return str
}

// getEnvs return the envs.
func (conf *Config) getEnvs() []string {
	envs := os.Environ()

	switch conf.AuthMethod {
	case AuthMethodPassword:
		envs = append(envs, fmt.Sprintf("%s=%s", WMIEnvPassword, conf.Password))
	default:
	}

	return envs
}

// AuthMethod defines the auth method type.
type AuthMethod string

const (
	// AuthMethodNone this defines the none auth method.
	AuthMethodNone = "none"

	// AuthMethodPassword this defines the password auth method.
	AuthMethodPassword = "password"
)

// Validate validate the auth method.
func (auth AuthMethod) Validate() error {
	switch auth {
	case AuthMethodNone, AuthMethodPassword:
		return nil
	default:
		return fmt.Errorf("invalid auth method, method(%s)", auth)
	}
}

// Validate validate the config.
func (conf *Config) Validate() error {
	if conf.IP == "" {
		return errors.New("ip is empty")
	}

	if conf.User == "" {
		return errors.New("user is empty")
	}

	if conf.Logger == nil {
		return errors.New("logger is empty")
	}

	if err := conf.AuthMethod.Validate(); err != nil {
		return err
	}

	switch conf.AuthMethod {
	case AuthMethodNone:
		if len(conf.Password) > 0 {
			return errors.New("auth mode is none,but private_key or password is not empty")
		}
	case AuthMethodPassword:
		if len(conf.Password) == 0 {
			return errors.New("password is empty")
		}
	}

	return nil
}

// Client is the wmiexec executor.
type Client struct {
	target  string
	envs    []string
	timeout time.Duration
	logger  logger.ILogger
}

// NewClient creates a Client instance.
func NewClient(config *Config) (*Client, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	client := &Client{
		target:  config.getTarget(),
		envs:    config.getEnvs(),
		timeout: 1 * time.Second,
		logger:  config.Logger,
	}

	if config.Timeout > 0 {
		client.timeout = config.Timeout
	}

	return client, nil
}

// RunCommand a command on the target host.
func (client *Client) RunCommand(ctx context.Context, command string) (string, string, error) {
	if command == "" {
		return "", "", errors.New("command is empty")
	}

	args := []string{
		client.target,
		command,
	}

	tCtx, cancel := context.WithTimeout(ctx, client.timeout)
	defer cancel()

	return wmiRunCmd(tCtx, args, client.envs)
}

// RunSilentCommand run a command on the target host without outputting.
func (client *Client) RunSilentCommand(ctx context.Context, command string) (string, string, error) {
	if command == "" {
		return "", "", errors.New("command is empty")
	}

	args := []string{
		"-silentcommand",
		client.target,
		command,
	}

	tCtx, cancel := context.WithTimeout(ctx, client.timeout)
	defer cancel()

	return wmiRunCmd(tCtx, args, client.envs)
}

// UploadFile upload the file to the target host.
func (client *Client) UploadFile(ctx context.Context, srcFilePath, dstDirPath string) (string, string, error) {
	command := fmt.Sprintf("lput %s %s", srcFilePath, dstDirPath)

	stdOut, stdErr, err := client.RunCommand(ctx, command)
	if err != nil {
		return "", "", fmt.Errorf("upload file failed, srcFilePath(%s), dstDirPath(%s), err: %w",
			srcFilePath, dstDirPath, err)
	}

	return stdOut, stdErr, nil
}
