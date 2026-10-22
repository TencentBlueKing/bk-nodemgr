//go:build windows

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

package firewall

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/node"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/logger"
)

func runFirewall(ctx context.Context, args StepArgs) error {
	features, err := parseFeatures(args.Features)
	if err != nil {
		return err
	}

	remoteIPs, err := parseRemoteIPs(args.RemoteIPs)
	if err != nil {
		return err
	}

	rules, err := buildRules(args.DeployEnv, features, remoteIPs)
	if err != nil {
		return err
	}

	return applyFirewallRules(ctx, rules, len(features) == 0 && len(remoteIPs) > 0)
}

func applyFirewallRules(ctx context.Context, rules []firewallRule, warnWithoutFeatures bool) error {
	if warnWithoutFeatures {
		logger.Warn(node.StepGeneral, "firewall remote IPs were provided without firewall features, skipping firewall rules")
		return nil
	}

	for _, rule := range rules {
		if err := applyFirewallRule(ctx, rule); err != nil {
			logger.Errorf(node.StepGeneral, "failed to configure firewall rule(%s): %v", rule.Name, err)
			return fmt.Errorf("failed to configure firewall rule %s: %w", rule.Name, err)
		}
	}

	if len(rules) > 0 {
		logger.Info(node.StepGeneral, "successfully configured Windows firewall rules")
	}
	return nil
}

func applyFirewallRule(ctx context.Context, rule firewallRule) error {
	show := exec.CommandContext(ctx, "netsh", "advfirewall", "firewall", "show", "rule", "name="+rule.Name)
	showOutput, err := show.CombinedOutput()
	if err != nil {
		if !strings.Contains(strings.ToLower(string(showOutput)), "no rules match") {
			return fmt.Errorf("failed to check existing rule: %w, output(%s)", err, showOutput)
		}
	} else {
		deleteCmd := exec.CommandContext(ctx, "netsh", "advfirewall", "firewall", "delete", "rule", "name="+rule.Name)
		if output, deleteErr := deleteCmd.CombinedOutput(); deleteErr != nil {
			return fmt.Errorf("failed to delete existing rule: %w, output(%s)", deleteErr, output)
		}
	}

	addArgs := []string{
		"advfirewall", "firewall", "add", "rule",
		"name=" + rule.Name,
		"dir=" + rule.Direction,
		"remoteip=" + rule.RemoteIPs,
		"protocol=" + rule.Protocol,
		rule.PortType + "=" + rule.Ports,
		"profile=public",
		"enable=yes",
		"action=allow",
	}
	addCmd := exec.CommandContext(ctx, "netsh", addArgs...)
	if output, addErr := addCmd.CombinedOutput(); addErr != nil {
		return fmt.Errorf("failed to add rule: %w, output(%s)", addErr, output)
	}

	return nil
}
