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

package precheck

import (
	"context"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/utils"
)

// PortPolicy port policy.
type PortPolicy struct {
	BindIP  string            `json:"bind_ip"`
	Port    uint64            `json:"port"`
	Network utils.NetworkType `json:"network"`
}

// Validate validate port policy.
func (policy *PortPolicy) Validate() error {
	if policy.BindIP == "" {
		return fmt.Errorf("invalid bind_ip: %s", policy.BindIP)
	}

	if policy.Port == 0 {
		return fmt.Errorf("invalid port: %d", policy.Port)
	}

	switch policy.Network {
	case utils.NetTCP, utils.NetTCP4, utils.NetTCP6, utils.NetUDP, utils.NetUDP4, utils.NetUDP6:
		return nil

	default:
		return fmt.Errorf("invalid network: %s", policy.Network.String())
	}
}

// CheckPortPolicies check port policies.
func CheckPortPolicies(ctx context.Context, polices []PortPolicy) error {
	gp := gopool.NewPool()
	for idx := range polices {
		policy := &polices[idx]
		gp.Go(func() error {
			switch policy.Network {
			case utils.NetTCP, utils.NetTCP4, utils.NetTCP6:
				idle, err := utils.CheckTCPPortIdle(ctx, policy.Network.String(), policy.BindIP, policy.Port)
				if err != nil || !idle {
					return fmt.Errorf("network is not idle. network(%+v), err(%v)", policy, err)
				}

				return nil

			case utils.NetUDP, utils.NetUDP4, utils.NetUDP6:
				idle, err := utils.CheckUDPPortIdle(ctx, policy.Network.String(), policy.BindIP, policy.Port)
				if err != nil || !idle {
					return fmt.Errorf("network is not idle. network(%+v), err(%v)", policy, err)
				}

				return nil

			default:
				return fmt.Errorf("invalid network: %s", policy.Network.String())
			}
		})
	}

	if err := gp.Wait(); err != nil {
		return err
	}

	return nil
}
