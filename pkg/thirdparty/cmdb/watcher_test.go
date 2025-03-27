/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package cmdb

import (
	"context"
	"testing"
	"time"
)

func testWatcher(t *testing.T) IWatcher {
	client := testClient(t)

	watcher, err := client.NewWatcher("test")
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}

	if err := watcher.Start(context.Background()); err != nil {
		t.Fatalf("failed to start watcher: %v", err)
	}

	return watcher
}

// TestWatcher_WatchHost tests the WatchHost method of the Watcher struct.
func TestWatcher_WatchHostRelation(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "normal",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := testWatcher(t)

			got, err := w.WatchHostRelation()
			if (err != nil) != tt.wantErr {
				t.Errorf("WatchHostRelation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for i := 0; i < 60; i++ {
				select {
				case event := <-got:
					t.Logf("got event, event: %v", event)
				default:
					time.Sleep(1 * time.Second)
				}
			}
		})
	}
}
