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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestHTTPClientConcurrentRequests(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/execute" || r.Header.Get("Authorization") != "test-token" {
			t.Error("unexpected request method, path or token")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		var request struct {
			Args     []string `json:"args"`
			Password string   `json:"password"`
			Timeout  float64  `json:"timeout"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(request.Args, []string{"user@host", "whoami"}) || request.Timeout <= 0 {
			t.Error("command arguments or timeout were not preserved")
		}
		if err := json.NewEncoder(w).Encode(map[string]string{"stdout": request.Password, "stderr": ""}); err != nil {
			t.Error(err)
		}
	}))
	defer srv.Close()
	cli, err := newHTTPClient(DefaultStartPort, "test-token")
	if err != nil {
		t.Fatal(err)
	}
	endpoint, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	cli.url = *endpoint
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	var wg sync.WaitGroup
	for i := range 300 {
		wg.Go(func() {
			password := string(rune('a' + i))
			stdout, stderr, err := cli.execute(ctx, []string{"user@host", "whoami"}, []string{WMIEnvPassword + "=old", WMIEnvPassword + "=" + password})
			if err != nil || stdout != password || stderr != "" {
				t.Errorf("request %d: stdout %q, stderr %q, error %v", i, stdout, stderr, err)
			}
		})
	}
	wg.Wait()
}

func TestHTTPClientDoesNotRetryExecution(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()
	cli, err := newHTTPClient(DefaultStartPort, "test-token")
	if err != nil {
		t.Fatal(err)
	}
	endpoint, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	cli.url = *endpoint
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	_, _, err = cli.execute(ctx, []string{"user@host", "whoami"}, nil)
	if err != nil {
		if calls.Load() != 1 {
			t.Fatalf("execution was retried: %d requests", calls.Load())
		}

		return
	}
	t.Fatal("expected an HTTP error")
}
