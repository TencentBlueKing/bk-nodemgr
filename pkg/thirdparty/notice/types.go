/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package notice

import (
	"fmt"

	apigwclient "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/client"
)

// Config the config of notice.
type Config struct {
	APIGWUserConfig apigwclient.UserConfig
}

// Validate configures the config.
func (conf *Config) Validate() error {
	if err := conf.APIGWUserConfig.Validate(); err != nil {
		return fmt.Errorf("failed to validate notice client config: %v", err)
	}

	return nil
}

const (
	// codeOK define the success code.
	codeOK = 0
)

// BaseBroker describe the base broker.
type BaseBroker[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

// IsFailed check the response is ok.
func (resp *BaseBroker[T]) IsFailed() error {
	if resp.Code != codeOK {
		return fmt.Errorf("code(%d), msg(%s)", resp.Code, resp.Message)
	}

	return nil
}
