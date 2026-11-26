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

package jinja2x

import (
	"context"
	"fmt"
	"sync"
)

const (
	// DefaultStartPort is the first port available to the embedded renderer.
	DefaultStartPort = 28150
	// DefaultEndPort is the last port available to the embedded renderer.
	DefaultEndPort = 28159
)

type server struct {
	mu                 sync.Mutex
	startPort, endPort int
	process            *serverProcess
	closed             bool
}

type serverProcess struct {
	client *client
	done   chan struct{}
	stop   func()
}

// All handlers in the same Go process share one embedded server.
// nolint: gochecknoglobals
var defaultServer = &server{startPort: DefaultStartPort, endPort: DefaultEndPort}

// SetPortRange configures the embedded renderer before serving requests.
func SetPortRange(start, end int) error {
	if start < 1 || end > 65535 || start > end {
		return fmt.Errorf("invalid jinja2 port range: %d-%d", start, end)
	}
	defaultServer.mu.Lock()
	defer defaultServer.mu.Unlock()
	if defaultServer.process != nil {
		return fmt.Errorf("jinja2 server already started")
	}
	defaultServer.startPort, defaultServer.endPort = start, end

	return nil
}

func (srv *server) getClient(ctx context.Context) (*client, error) {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	if srv.closed {
		return nil, fmt.Errorf("jinja2 server is closed")
	}
	if srv.process != nil {
		select {
		case <-srv.process.done:
			srv.process.client.http.CloseIdleConnections()
			srv.process = nil
		default:
			return srv.process.client, nil
		}
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	process, err := startServer(ctx, srv.startPort, srv.endPort)
	if err != nil {
		return nil, err
	}
	srv.process = process

	return process.client, nil
}

// Close stops the shared embedded server during service shutdown.
func Close() {
	defaultServer.mu.Lock()
	defer defaultServer.mu.Unlock()
	defaultServer.closed = true
	if defaultServer.process != nil {
		defaultServer.process.client.http.CloseIdleConnections()
		defaultServer.process.stop()
		defaultServer.process = nil
	}
}
