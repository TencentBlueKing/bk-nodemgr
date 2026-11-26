//go:build linux

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
	"testing"
	"time"
)

func TestServerConcurrentRenderAndRecovery(t *testing.T) {
	s := &server{startPort: DefaultStartPort, endPort: DefaultEndPort}
	t.Cleanup(func() {
		if s.process != nil {
			s.process.client.http.CloseIdleConnections()
			s.process.stop()
		}
	})
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	first, err := s.getClient(ctx)
	if err != nil {
		t.Fatal(err)
	}
	const tasks = 300
	var wg sync.WaitGroup
	start := time.Now()
	for i := range tasks {
		wg.Go(func() {
			for range 6 {
				c, err := s.getClient(ctx)
				if err != nil {
					t.Error(err)
					return
				}
				if c != first {
					t.Error("server was not reused")
					return
				}
				got, err := c.render(ctx, "Hello {{ name }}!", map[string]any{"name": i})
				if err != nil || got != fmt.Sprintf("Hello %d!", i) {
					t.Errorf("render task %d: got %q, error %v", i, got, err)
					return
				}
			}
		})
	}
	wg.Wait()
	t.Logf("%d tasks, 6 renders each: %s", tasks, time.Since(start))
	s.process.stop()
	second, err := s.getClient(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatal("server was not replaced after exit")
	}
	got, err := second.render(ctx, "{{ name }}", map[string]any{"name": "recovered"})
	if err != nil || got != "recovered" {
		t.Fatalf("render after restart: %q, %v", got, err)
	}
}
