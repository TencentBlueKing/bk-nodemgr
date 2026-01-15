/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package precheck

import (
	"context"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/utils"
)

// NetworkPolicy network policy.
type NetworkPolicy struct {
	Host    string `json:"host"`
	Port    uint64 `json:"port"`
	Network string `json:"network"`
}

// Validate validate network policy.
func (policy *NetworkPolicy) Validate() error {
	if policy.Host == "" {
		return fmt.Errorf("invalid host: %s", policy.Host)
	}

	if policy.Port == 0 {
		return fmt.Errorf("invalid port: %d", policy.Port)
	}

	switch policy.Network {
	case utils.NetTCP4.String(), utils.NetTCP6.String(), utils.NetUDP4.String(), utils.NetUDP6.String():
		return nil
	default:
		return fmt.Errorf("invalid network: %s", policy.Network)
	}
}

// CheckNetworkPolicies check network policies.
func CheckNetworkPolicies(ctx context.Context, policies []NetworkPolicy) error {
	gp := gopool.NewPool()
	for idx := range policies {
		policy := &policies[idx]
		gp.Go(func() error {
			switch policy.Network {
			case utils.NetTCP4.String():
				idle, err := utils.CheckTCP4PortIdle(ctx, policy.Port)
				if err != nil || !idle {
					return fmt.Errorf("port %d is not idle", policy.Port)
				}

				return nil
			case utils.NetTCP6.String():
				idle, err := utils.CheckTCP6PortIdle(ctx, policy.Port)
				if err != nil || !idle {
					return fmt.Errorf("port %d is not idle", policy.Port)
				}

				return nil
			case utils.NetUDP4.String(), utils.NetUDP6.String():
				return fmt.Errorf("not support network: %s", policy.Network)
			default:
				return fmt.Errorf("invalid network: %s", policy.Network)
			}
		})
	}

	if err := gp.Wait(); err != nil {
		return err
	}

	return nil
}
