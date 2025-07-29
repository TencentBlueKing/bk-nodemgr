# thirdparty

## 目录说明

此目录专门用于管理所有第三方服务的集成代码，确保外部依赖的统一管理和维护。

## 包结构规范

每个第三方服务包（pkg）采用统一的文件结构，具体组成如下（可根据实际需求适当调整）：

- `README.md` - 包的详细说明文档，包含使用方法、配置说明和注意事项
- `pkg_name.go` - 与包同名的核心文件，封装对第三方原始 API 的直接调用逻辑
- `types.go` - 定义第三方接口交互所需的数据结构，包括请求参数（request）和响应数据（response）等类型
- `handler.go` - 业务逻辑适配层，主要职责包括：
    - 将第三方系统的业务概念和实现细节封装在当前包内，避免外部系统概念泄露到业务代码中
    - 提供系统内部数据类型与第三方 API 数据类型之间的转换功能
    - 提供场景化的接口，而非提供原始 API 接口

## 注意事项

1. 应使用 handler 内部定义的 interface，而不是直接使用 Handler 这一具体实现，确保依赖抽象而不是具体实现，提高代码的可测试性和可扩展性
2. 所有第三方服务的错误处理应在 handler 层进行统一封装，将外部错误转换为系统内部的标准错误格式

## 示例

### handler.go

```go
/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package bkoa

import (
	"context"

	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
)

// IHandler defines the handler interface
type IHandler interface {
	Verify(ctx context.Context, bkTicket string) error
}

// Handler the Handler of cmdb.
type Handler struct {
	cli    *cli
	logger logger.Logger
}

// OptionFn ...
type OptionFn func(*Handler)

// WithLogger this func will set the logger of the Handler.
func WithLogger(logger logger.Logger) OptionFn {
	return func(s *Handler) {
		s.logger = logger
	}
}

// New initialize a new cmdb Handler.
func New(c *restclient.Capability, conf *Config, opts ...OptionFn) (Handler, error) {
	cli, err := newClient(c, conf)
	if err != nil {
		return nil, err
	}

	h := &Handler{
		cli:    cli,
		logger: logger.LoggerDefault{},
	}

	for _, opt := range opts {
		opt(h)
	}

	return h, nil
}

```