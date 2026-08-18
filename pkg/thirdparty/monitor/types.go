/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package monitor

import (
	"fmt"

	apigwclient "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/client"
)

const codeOK = 200

// Config defines monitor gateway config.
type Config struct {
	VirtualUserConfig apigwclient.VirtualUserConfig
}

// Validate validates monitor gateway config.
func (conf *Config) Validate() error {
	if err := conf.VirtualUserConfig.Validate(); err != nil {
		return fmt.Errorf("failed to validate monitor client config: %w", err)
	}

	return nil
}

// BaseBroker describes monitor gateway response wrapper.
type BaseBroker[T any] struct {
	Result  *bool  `json:"result"`
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

// IsFailed checks whether the response failed.
func (resp *BaseBroker[T]) IsFailed() error {
	if resp.Code != codeOK {
		return fmt.Errorf("code(%d), msg(%s)", resp.Code, resp.Message)
	}

	if resp.Result != nil && !*resp.Result {
		return fmt.Errorf("code(%d), msg(%s)", resp.Code, resp.Message)
	}

	return nil
}

// GetOrCreateAgentEventDataIDResp describes monitor event data-id response data.
type GetOrCreateAgentEventDataIDResp struct {
	BkDataID int64 `json:"bk_data_id"`
}
