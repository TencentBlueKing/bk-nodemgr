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
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/crypter"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/sshx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

// NewActionSshHostExecCmd ...
func NewActionSshHostExecCmd(crypter crypter.Crypter, logger logger.Logger) action.Definition {
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
func (s *sshHostExecCmd) Tags() []action.Tag {
	return []action.Tag{}
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
func (s *sshHostExecCmd) Do(ctx *action.InstanceContext) error {
	param := new(SshHostExecCmdParam)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	passwd, err := s.crypter.Decrypt(param.Password)
	if err != nil {
		return err
	}

	client, err := sshx.NewClient(ctx.Ctx, &sshx.Config{
		Network:  sshx.Network(param.getNetwork()),
		IP:       param.IP,
		Port:     param.Port,
		User:     param.getUser(),
		Password: string(passwd),
		Logger:   s.logger,
	}, s.getTimeout())
	if err != nil {
		return fmt.Errorf("failed to connect to host, err: %v", err)
	}
	defer client.Close()

	s.logger.Infof("successfully connected to host, addr(%s)", param.getAddr())
	ctx.Data.Log(fmt.Sprintf("start to exec cmd on host, addr(%s)", param.getAddr()))

	cmds, err := param.getCmds()
	if err != nil {
		return fmt.Errorf("failed to get cmd, err: %v", err)
	}

	outputMap := make(map[string]string)
	for _, cmd := range cmds {
		output, err := client.RunCommand(cmd)
		if err != nil {
			return fmt.Errorf("failed to exec cmd, cmd(%s), output(%s), err: %v", cmd, output, err)
		}

		ctx.Data.Log(fmt.Sprintf("exec cmd, cmd(%s), output(%s)", cmd, output))
		s.logger.Infof("exec cmd on host, host(%s), cmd(%s), output(%s)", param.getAddr(), cmd, output)
		outputMap[cmd] = output
	}

	ctx.Data.Content["ssh_output"] = outputMap

	return nil
}

// getTimeout get timeout, if not set, use default value.
func (s *sshHostExecCmd) getTimeout() time.Duration {
	if s.Timeout() == 0 {
		return 10 * time.Second
	}

	return s.Timeout()
}
