/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package workflowdef ...
package workflowdef

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/crypter"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/retrier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operengine"
	"golang.org/x/crypto/ssh"
)

// NewActionSshHostExecCmd ...
func NewActionSshHostExecCmd(crypter crypter.Crypter, logger logger.Logger) operengine.ActionDef {
	return &sshHostExecCmd{
		logger:  logger,
		crypter: crypter,
	}
}

// SshHostExecCmdParam ...
type SshHostExecCmdParam struct {
	IP         string   `json:"ip"`
	Port       int      `json:"port"`
	User       string   `json:"user"`
	Password   []byte   `json:"passwd"`
	Network    string   `json:"network"`
	Commands   []string `json:"command"`
	SSHTimeout int      `json:"ssh_timeout"`
}

// getUser get param user, if empty, use root.
func (s *SshHostExecCmdParam) getUser() string {
	if s.User == "" {
		return "root"
	}

	return s.User
}

// getCmds get cmd and check it.
func (s *SshHostExecCmdParam) getCmds() ([]string, error) {
	for _, cmd := range s.Commands {
		dangerousCommands := []string{"rm -rf", "rm -fr", "sudo", "su", "rm -", "shutdown", "halt", "poweroff"}
		for _, dc := range dangerousCommands {
			if strings.Contains(cmd, dc) {
				return nil, fmt.Errorf("command contains dangerous element, element(%s)", dc)
			}
		}
	}

	return s.Commands, nil
}

func (s *SshHostExecCmdParam) getAddr() string {
	return fmt.Sprintf("%s:%d", s.IP, s.Port)
}

// getNetwork get param network, if empty, use tcp.
func (s *SshHostExecCmdParam) getNetwork() string {
	switch s.Network {
	case "tcp":
		return "tcp"
	case "udp":
		return "udp"
	default:
		return "tcp"
	}
}

// syncHostFromCMDB ...
type sshHostExecCmd struct {
	logger  logger.Logger
	crypter crypter.Crypter
}

// Name returns the name of the action.
func (s *sshHostExecCmd) Name() string {
	return SshHostExecCmd
}

// Version returns the version of the action.
func (s *sshHostExecCmd) Version() string {
	return "v1"
}

// Description returns the description of the action.
func (s *sshHostExecCmd) Description() string {
	return "use ssh to exec cmd on host"
}

// Timeout returns the timeout of the action.
func (s *sshHostExecCmd) Timeout() time.Duration {
	return 2 * time.Minute
}

// Tags returns the tags of the action.
func (s *sshHostExecCmd) Tags() []operengine.ActionTag {
	return []operengine.ActionTag{}
}

// MaxRetryCount don't allow auto retry.
func (s *sshHostExecCmd) MaxRetryCount() uint {
	return 0
}

// DelayFn don't allow auto retry.
func (s *sshHostExecCmd) DelayFn() func() {
	return func() {}
}

// Do this func define what the action will do.
func (s *sshHostExecCmd) Do(ctx *operengine.ActionInstContext) error {
	param := new(SshHostExecCmdParam)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	passwd, err := s.crypter.Decrypt(param.Password)
	if err != nil {
		return err
	}

	config := &ssh.ClientConfig{
		User: param.getUser(),
		Auth: []ssh.AuthMethod{
			ssh.Password(string(passwd)),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
		BannerCallback: func(message string) error {
			s.logger.Infof("ssh banner: %s", message)

			return nil
		},
		Timeout: s.getTimeout(),
	}

	// because of the network may be unstable, so we need to retry.
	retrier := retrier.NewExpoBackoff(retrier.ExpoBackoffOpts{
		MaxRetries:    5,
		BaseDelay:     1 * time.Second,
		MaxDelay:      3 * time.Second,
		JitterPercent: 0.2,
		Logger:        s.logger,
	})

	var client *ssh.Client
	err = retrier.Do(ctx.Ctx, func(attempt int) error {
		client, err = ssh.Dial(param.getNetwork(), param.getAddr(), config)
		if err != nil {
			s.logger.Errorf("failed to connect to host, host(%s), err: %v", param.getAddr(), err)
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session, err: %v", err)
	}
	defer session.Close()

	s.logger.Infof("successfully connected to host, addr(%s)", param.getAddr())
	ctx.Data.Log(fmt.Sprintf("start to exec cmd on host, addr(%s)", param.getAddr()))

	cmds, err := param.getCmds()
	coutputs := make([]string, 0)
	if err != nil {
		return fmt.Errorf("failed to get cmd, err: %v", err)
	}

	for _, cmd := range cmds {
		output, err := session.CombinedOutput(cmd)
		if err != nil {
			return fmt.Errorf("failed to exec cmd, err: %v", err)
		}

		ctx.Data.Log(fmt.Sprintf("exec cmd, cmd(%s), output(%s)", cmd, output))
		s.logger.Infof("exec cmd on host, host(%s), cmd(%s), output(%s)", param.getAddr(), cmd, output)
	}

	ctx.Data.Content["ssh_output"] = coutputs

	return nil
}

// getTimeout get timeout, if not set, use default value.
func (s *sshHostExecCmd) getTimeout() time.Duration {
	if s.Timeout() == 0 {
		return 10 * time.Second
	}

	return s.Timeout()
}
