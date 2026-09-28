//go:build darwin || linux

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

package local

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/stretchr/testify/require"
)

func TestLocalDir_AllFilesRejectsFIFOAndSymlinkToFIFO(t *testing.T) {
	for _, linked := range []bool{false, true} {
		t.Run(fmt.Sprint(linked), func(t *testing.T) {
			root := t.TempDir()
			pipe := filepath.Join(root, "pipe")
			if linked {
				pipe = filepath.Join(t.TempDir(), "pipe")
			}
			require.NoError(t, syscall.Mkfifo(pipe, 0600))
			if linked {
				require.NoError(t, os.Symlink(pipe, filepath.Join(root, "alias")))
			}
			group := newTestLocalDir(t, root)
			name := "pipe"
			if linked {
				name = "alias"
			}
			_, err := group.IsDir(contextx.Background(), name)
			require.ErrorContains(t, err, "unsupported file type")
			_, err = group.AllFiles(contextx.Background())
			require.ErrorContains(t, err, "unsupported file type")
		})
	}
}
