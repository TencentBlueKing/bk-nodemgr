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

package wmix

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestServerConcurrentReuseAndRecovery(t *testing.T) {
	srv := &server{startPort: DefaultStartPort, endPort: DefaultEndPort}
	t.Cleanup(func() {
		if srv.process != nil {
			srv.process.client.http.CloseIdleConnections()
			srv.process.stop()
		}
	})
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	first, err := srv.getClient(ctx)
	if err != nil {
		t.Fatal(err)
	}
	unauthorized := *first
	unauthorized.token = "invalid-token"
	err = unauthorized.request(ctx, http.MethodGet, "/healthz", nil, &struct{}{})
	if err != nil {
		t.Log("embedded server rejected an invalid token")
	} else {
		t.Fatal("embedded server accepted an invalid token")
	}
	var wg sync.WaitGroup
	for range 300 {
		wg.Go(func() {
			cli, err := srv.getClient(ctx)
			if err != nil {
				t.Error(err)
				return
			}
			if cli != first {
				t.Error("embedded WMI process was not reused")
				return
			}
			if err := cli.request(ctx, http.MethodGet, "/healthz", nil, &struct{}{}); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	srv.process.stop()
	second, err := srv.getClient(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if second == first {
		t.Fatal("embedded WMI process was not replaced after exit")
	}
	if err := second.request(ctx, http.MethodGet, "/healthz", nil, &struct{}{}); err != nil {
		t.Fatal(err)
	}
}
